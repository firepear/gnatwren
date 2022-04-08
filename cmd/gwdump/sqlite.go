package main

import (
	"log"
	"database/sql"
	"encoding/json"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

func dbSetup(dbloc string) (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", dbloc)
	return db, err
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

func dbGetOverview() (*map[string]data.AgentPayload, error) {
 	// make a map to hold the metrics
 	metrics := map[string]data.AgentPayload{}
	// and vars to hold each datum from the query
	var ts int64
	var host, mstr string

	// get most recent row for each host
	rows, err := db.Query("SELECT max(ts), host, data FROM current GROUP BY host ORDER BY host")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// iterate over rows, vivifying `data` and assembling metrics
	for rows.Next() {
		if err = rows.Scan(&ts, &host, &mstr); err != nil {
			return nil, err
		}
		var m data.AgentPayload
		err = json.Unmarshal([]byte(mstr), &m)
		metrics[host] = m
	}
	return &metrics, err
}

func dbGetCPUStats(duration string) (*map[int64]map[string]string, error) {
 	// map of temps (by timestamp, by host), to be returned
 	t := map[int64]map[string]string{}
	//
 	var rows *sql.Rows
	var err error

	switch duration {
	case "current":
		rows, err = db.Query("SELECT ts, host, json_extract(data, '$.Cpu.Temp') FROM current ORDER BY ts")
	case "hourly":
		rows, err = db.Query("SELECT ts, host, json_extract(data, '$.Cpu.Temp') FROM hourly ORDER BY ts")
	case "daily":
		rows, err = db.Query("SELECT ts, host, json_extract(data, '$.Cpu.Temp') FROM daily ORDER BY ts")
	}
	if err != nil {
		log.Printf("cpustats %s query failed: %s\n", duration, err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ts int64
		var host string
		var temp string
		if err = rows.Scan(&ts, &host, &temp); err != nil {
			return nil, err
		}

		if t[ts] == nil {
			t[ts] = map[string]string{}
		}
		t[ts][host] = temp

	}
	return &t, nil
}

// https://www.sqlite.org/sharedcache.html

