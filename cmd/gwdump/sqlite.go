package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

func dbSetup(dbloc string) (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", dbloc)
	return db, err
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
