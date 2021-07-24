package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//"strconv"
	"time"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

func dbSetup() (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", "./.gnatwren.db")
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
	rows, err := db.Query("SELECT DISTINCT host FROM current ORDER BY host")
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
		nodeStatus[host] = [2]int64{ts, ts}
	}
}

func dbUpdate(payload []byte, upd data.AgentPayload) error {
 	// insert payload (we don't have to care about concurrency
 	// here; that's taken care of by a mutex in petrel.go)
 	stmt, err := db.Prepare("INSERT INTO current VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(upd.TS, upd.Host, string(payload))
	if err != nil {
		return err
	}
 	return err
}

func dbGetCurrentStats() (map[string]data.AgentStatus, error) {
	// copy the nodeStatus to minimize time it's locked
	nodeCopy := map[string][2]int64{}
	mux.RLock()
	for k, v := range nodeStatus {
		nodeCopy[k] = v
	}
	mux.RUnlock()

 	// make a map to hold the metrics
 	metrics := map[string]data.AgentStatus{}

	// loop over nodeCopy, getting the most recently inserted row
	// for each machine and adding it to metrics
	var err error
	for host, hostTs := range nodeCopy {
		row := db.QueryRow("SELECT data FROM current WHERE ts = ? AND host = ?", hostTs[1], host)
		var (
			d string
			m data.AgentStatus
		)
		if err = row.Scan(&d); err != nil {
			return nil, err
		}

		m.TS = hostTs[1]
		err = json.Unmarshal([]byte(d), &m.Payload)
		metrics[host] = m
	}
 	return metrics, err
}

func dbGetCPUTemps() (map[int64]map[string]string, error) {
 	// map of temps (by timestamp, by host), to be returned
 	t := map[int64]map[string]string{}
 	// json goes here
 	m := data.AgentPayload{}
 	// timestamp, as a string, one hour ago. we don't want
 	// anything older than this
 	tlimit := time.Now().Unix() - 3600

	rows, err := db.Query("SELECT data FROM current WHERE ts >= ?", tlimit)
	if err != nil {
		return t, err
	}
	for rows.Next() {
		var d string
		if err = rows.Scan(&d); err != nil {
			return t, err
		}
		err = json.Unmarshal([]byte(d), &m)
		if err != nil {
			return t, err
		}

		if t[m.TS] == nil {
			t[m.TS] = map[string]string{}
		}
		t[m.TS][m.Host] = fmt.Sprintf("%5.2f", m.Cpu.Temp)

	}
	return t, nil
}

func dbGetDBStats() (data.DBStatus, error) {
 	var dbs data.DBStatus

// 	err := db.View(func(txn *badger.Txn) error {
// 		it := txn.NewIterator(badger.DefaultIteratorOptions)
// 		defer it.Close()

// 		for it.Rewind(); it.Valid(); it.Next() {
// 			item := it.Item()
// 			k := item.Key()
// 			if dbs.Count == 0 {
// 				dbs.Oldest = string(k)
// 			} else {
// 				dbs.Newest = string(k)
// 			}
// 			dbs.Count++
// 		}
// 		return nil
// 	})
 	//return dbs, err
	return dbs, nil
}

// https://www.sqlite.org/sharedcache.html

