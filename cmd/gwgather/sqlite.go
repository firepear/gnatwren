package main

import (
	"database/sql"
	"encoding/json"
	//"fmt"
	"log"
	"time"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

func dbSetup(dbloc string) (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", dbloc)
	if err != nil {
		return db, err
	}
	// create tables and indices if needed
	stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS current (ts INTEGER, host TEXT, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS currentidx ON current(ts)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS hourly (ts INTEGER, host TEXT, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS hourlyidx ON hourly(ts)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS daily (ts INTEGER, host TEXT, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS dailyidx ON daily(ts)")
	stmt.Exec()
	return db, nil
}

func dbLoadNodeStatus() {
	// build nodeStatus from data in DB on startup, which lets exportJSON work
	//
	// first get a list of hostnames
	hosts := []string{}
	rows, err := db.Query(`SELECT DISTINCT host FROM
                                   (SELECT DISTINCT host FROM current
                                    UNION ALL
                                    SELECT DISTINCT host FROM hourly
                                    UNION ALL
                                    SELECT DISTINCT host FROM daily)
                               ORDER BY host`)
	if err != nil {
		return
	}
	for rows.Next() {
		var host string
		if err = rows.Scan(&host); err != nil {
			return
		}
		hosts = append(hosts, host)
	}

	// now, for each host, fetch the most recent timestamp and slug it into nodeStatus
	for _, host := range hosts {
		row := db.QueryRow("SELECT ts FROM current WHERE host = ? ORDER BY ts DESC LIMIT 1",
			host)
		var ts int64
		if err = row.Scan(&ts); err != nil {
			return
		}
		nodeStatus[host] = &[2]int64{ts, ts}
	}
}

func dbPruneMigrate() {
	var c int64
	// timestamp, one hour ago
	tlimit := time.Now().Unix() - 3600
	// if nothing newer than tlimit exists in the current table,
	// there haven't been DB updates in a while; do nothing
	row := db.QueryRow("SELECT count(ts) FROM current WHERE ts >= ?", tlimit)
	if err := row.Scan(&c); err != nil {
		return
	}
	if c == 0 {
		return
	}

	// get the newest timestamp from the hourly table
	c = 0
	row = db.QueryRow("SELECT ts FROM hourly ORDER BY ts DESC LIMIT 1")
	switch err := row.Scan(&c); err {
	case sql.ErrNoRows:
		// treat an empty table as a nil
		fallthrough
	case nil:
		//  do nothing if the most recent timestamp is
		//  less than 1h old AND NOT zero (empty table)
		if c >= tlimit && c != 0 {
			break
		}
		// for each host, grab the newest row from common --
		// since we test for the most recent row in hourly
		// being at least 1h old -- and copy to hourly
		for host := range nodeStatus {
			var (
				q  = "SELECT ts, data FROM current WHERE host = ? ORDER BY ts DESC LIMIT 1"
				ts int64
				d  string
			)
			row = db.QueryRow(q, host, tlimit)
			if err := row.Scan(&ts, &d); err != nil {
				log.Printf("didn't find current data for %s: err: %s\n", host, err)
				continue
			}
			stmt, err := db.Prepare("INSERT INTO hourly VALUES (?, ?, ?)")
			if err != nil {
				log.Printf("couldn't insert data for %s into hourly: %s\n", host, err)
				continue
			}
			stmt.Exec(ts, host, d)
		}
	default:
		log.Printf("err: %s\n", err)
		return
	}

	// prune current
	stmt, err := db.Prepare("DELETE FROM current WHERE ts < ?")
	if err != nil {
		log.Printf("db: can't prune current table: %s\n", err)
	}
	stmt.Exec(tlimit)

	// now find the most recent timestamp from the daily table
	c = 0
	tlimit = tlimit - 169200 // go back another 47h
	row = db.QueryRow("SELECT ts FROM daily ORDER BY ts DESC LIMIT 1")
	switch err := row.Scan(&c); err {
	case sql.ErrNoRows:
		fallthrough
	case nil:
		//  do nothing if the most recent timestamp is
		//  less than 48h old AND NOT zero (empty table)
		if c >= tlimit && c != 0 {
			break
		}
		// otherwise, copy most recent data for each host from hourly to daily
		for host := range nodeStatus {
			var (
				q  = "SELECT ts, data FROM hourly WHERE host = ? ORDER BY ts DESC LIMIT 1"
				ts int64
				d  string
			)
			row = db.QueryRow(q, host, tlimit)
			if err := row.Scan(&ts, &d); err != nil {
				log.Printf("didn't find hourly data for %s: err: %s\n", host, err)
				continue
			}
			stmt, err := db.Prepare("INSERT INTO daily VALUES (?, ?, ?)")
			if err != nil {
				log.Printf("couldn't insert data for %s into daily: %s\n", host, err)
				continue
			}
			stmt.Exec(ts, host, d)
		}
	default:
		log.Printf("err: %s\n", err)
		return
	}
	// prune hourly
	stmt, err = db.Prepare("DELETE FROM hourly WHERE ts < ?")
	if err != nil {
		log.Printf("db: can't prune hourly table: %s\n", err)
	}
	stmt.Exec(tlimit)
	// and daily
	tlimit = tlimit - 5011200 // go back another 58 days
	stmt, err = db.Prepare("DELETE FROM daily WHERE ts < ?")
	if err != nil {
		log.Printf("db: can't prune daily table: %s\n", err)
	}
	stmt.Exec(tlimit)
}

func dbUpdate(nodedata []byte) error {
	// vivify the update data
	var upd = data.AgentPayload{}
	err := json.Unmarshal(nodedata, &upd)
	if err != nil {
		log.Printf("agentUpdate: json unmarshal err: %s", err)
		return err
	}

	mux.Lock()
	// update nodeStatus. the first timestamp is now (check-in
	// rec'd time)
	checkin := time.Now().Unix()
	if _, ok := nodeStatus[upd.Host]; ok {
		nodeStatus[upd.Host][0] = checkin
	} else {
		nodeStatus[upd.Host] = &[2]int64{checkin, checkin}
	}
	// second timestamp is the hosts's last reporting time, which
	// can be in the past due to event playback). only update if
	// the event timestamp is newer than what we have
	if upd.TS > nodeStatus[upd.Host][1] {
		nodeStatus[upd.Host][1] = upd.TS
	}

	// insert payload
	stmt, err := db.Prepare("INSERT INTO current VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(upd.TS, upd.Host, string(nodedata))
	mux.Unlock()

	return err
}

// https://www.sqlite.org/sharedcache.html
