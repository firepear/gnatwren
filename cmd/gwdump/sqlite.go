package main

import (
	"database/sql"
	"time"

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

func dbGetOverview() (*map[string]data.AgentStatus, error) {
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
		var m data.AgentStatus
		if err = row.Scan(&m.Payload); err != nil {
			// this used to barf on no data. now it
			// doesn't, but the right thing to do is
			// something more useful TODO
			continue
		}

		m.TS = hostTs[1]
		metrics[host] = m
	}
	return &metrics, err
}

func dbGetCPUStats(duration string) (*map[int64]map[string]string, error) {
 	// map of temps (by timestamp, by host), to be returned
 	t := map[int64]map[string]string{}
 	// timestamp, one hour ago. we don't want anything older than
 	// this
 	var tlimit int64
	switch duration {
	case "current":
		tlimit = time.Now().Unix() - 3600
	case "hourly":
		tlimit = time.Now().Unix() - 3600
	case "daily":
		tlimit = time.Now().Unix() - 3600
	}

	rows, err := db.Query("SELECT ts, host, json_extract(data, '$.Cpu.Temp') FROM ? WHERE ts >= ?", duration, tlimit)
	if err != nil {
		return nil, err
	}
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

