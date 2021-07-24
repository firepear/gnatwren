package main

import (
	"database/sql"
	//"encoding/json"
	//"fmt"
	//"strconv"
	//"time"

	"github.com/firepear/gnatwren/internal/data"
	_ "github.com/mattn/go-sqlite3"
)

func dbSetup() (*sql.DB, error) {
	// Open the database
	db, err := sql.Open("sqlite3", "./gnatwren.db")
	if err != nil {
		return db, err
	}
	// create tables and indices if needed
	stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS current (ts INTEGER, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS currentidx ON current")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS hourly (ts INTEGER, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS hourlyidx ON hourly")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS daily (ts INTEGER, data TEXT)")
	stmt.Exec()
	stmt, _ = db.Prepare("CREATE INDEX IF NOT EXISTS dailyidx ON daily")
	stmt.Exec()
	return db, nil
}

func dbUpdate(payload []byte, upd data.AgentPayload) error {
	// get timestamp
 	ts := upd.TS
 	// insert payload (we don't have to care about concurrency
 	// here; that's taken care of by a mutex in petrel.go)
 	stmt, err := db.Prepare("INSERT INTO current VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(ts, string(payload))
	if err != nil {
		return err
	}
 	return err
}

func dbGetCurrentStats() (map[string]data.AgentStatus, error) {
// 	// copy the nodeStatus to minimize time it's locked
// 	nodeCopy := map[string][2]int64{}
// 	mux.RLock()
// 	for k, v := range nodeStatus {
// 		nodeCopy[k] = v
// 	}
// 	mux.RUnlock()

 	// make a map to hold the metrics
 	metrics := map[string]data.AgentStatus{}

// 	// iterate on nodeCopy to get the hostname and most recent
// 	// update time, which we need to build a key, to get the most
// 	// recent metrics
 	var err error
// 	for k, v := range nodeCopy {
// 		// make key
// 		key := []byte(strconv.Itoa(int(v[1])))
// 		key = append(key, []byte(k)...)
// 		// lookup data
// 		err = db.View(func(txn *badger.Txn) error {
// 			item, err := txn.Get(key)
// 			if err != nil {
// 				return err
// 			}
// 			// we basically have a cursor at this point,
// 			// and call Value to vivify it. that data is
// 			// only accessible inside the function call,
// 			// however
// 			err = item.Value(func(val []byte) error {
// 				var m data.AgentStatus
// 				m.TS = v[0]
// 				err = json.Unmarshal(val, &m.Payload)
// 				metrics[k] = m
// 				return err
// 			})
// 			return err
// 		})
// 	}
 	return metrics, err
}

func dbGetCPUTemps() (map[int64]map[string]string, error) {
// 	// map of temps (by timestamp, by host), to be returned
 	t := map[int64]map[string]string{}
// 	// json goes here
// 	m := data.AgentPayload{}
// 	// timestamp, as a string, one hour ago. we don't want
// 	// anything older than this
// 	tlimit := strconv.Itoa(int(time.Now().Unix()) - 3600)


// 	err := db.View(func(txn *badger.Txn) error {
// 		opts := badger.DefaultIteratorOptions
// 		opts.PrefetchValues = false
// 		it := txn.NewIterator(opts)
// 		defer it.Close()

// 		for it.Rewind(); it.Valid(); it.Next() {
// 			item := it.Item()
// 			k := item.Key()
// 			// skip keys that happened more than 1h ago
// 			if string(k) < tlimit {
// 				continue
// 			}
// 			err := item.Value(func(v []byte) error {
// 				err := json.Unmarshal(v, &m)
// 				if err != nil {
// 					return err
// 				}
// 				if t[m.TS] == nil {
// 					t[m.TS] = map[string]string{}
// 				}
// 				t[m.TS][m.Host] = fmt.Sprintf("%5.2f", m.Cpu.Temp)
// 				return nil
// 			})
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})
	// 	return t, err
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

