package main

import (
	//"log"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/firepear/gnatwren/internal/data"
	badger "github.com/dgraph-io/badger/v3"
)


func dbUpdate(payload []byte, upd data.AgentPayload) error {
	// build key from the metrics timestamp and node name
	key := []byte(strconv.Itoa(int(upd.TS)))
	key = append(key, []byte(upd.Host)...)
	// execute a Set transaction
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, payload).WithTTL(24 * time.Hour)
		err := txn.SetEntry(e)
		if err != nil {
			// we treat key conflicts as no-ops
			if err.Error() == badger.ErrConflict.Error() {
				return nil
			}
		}
		return err
	})
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

	// iterate on nodeCopy to get the hostname and most recent
	// update time, which we need to build a key, to get the most
	// recent metrics
	var err error
	for k, v := range nodeCopy {
		// make key
		key := []byte(strconv.Itoa(int(v[1])))
		key = append(key, []byte(k)...)
		// lookup data
		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				return err
			}
			// we basically have a cursor at this point,
			// and call Value to vivify it. that data is
			// only accessible inside the function call,
			// however
			err = item.Value(func(val []byte) error {
				var m data.AgentStatus
				m.TS = v[0]
				err = json.Unmarshal(val, &m.Payload)
				metrics[k] = m
				return err
			})
			return err
		})
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
	tlimit := strconv.Itoa(int(time.Now().Unix()) - 3600)


	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			// skip keys that happened more than 1h ago
			if string(k) < tlimit {
				continue
			}
			err := item.Value(func(v []byte) error {
				err := json.Unmarshal(v, &m)
				if err != nil {
					return err
				}
				if t[m.TS] == nil {
					t[m.TS] = map[string]string{}
				}
				t[m.TS][m.Host] = fmt.Sprintf("%5.2f", m.Cpu.Temp)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return t, err
}

func dbGetDBStats() (data.DBStatus, error) {
	var dbs data.DBStatus

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if dbs.Count == 0 {
				dbs.Oldest = string(k)
			} else {
				dbs.Newest = string(k)
			}
			dbs.Count++
		}
		return nil
	})
	return dbs, err
}
