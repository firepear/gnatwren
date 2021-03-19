package main

import (
	//"log"
	"encoding/json"
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

func dbGetCurrentStats() (map[string]data.AgentPayload, error) {
	// copy the nodeStatus to minimize time it's locked
	nodeCopy := map[string]int64{}
	mux.RLock()
	for k, v := range nodeStatus {
		nodeCopy[k] = v
	}
	mux.RUnlock()
	// make a map to hold the metrics
	metrics := map[string]data.AgentPayload{}

	var err error
	for k, v := range nodeCopy {
		// make key
		key := []byte(strconv.Itoa(int(v)))
		key = append(key, []byte(k)...)
		// lookup data
		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				return err
			}

			err = item.Value(func(val []byte) error {
				var m data.AgentPayload
				err = json.Unmarshal(val, &m)
				metrics[k] = m
				//log.Printf("%s: %v", k, metrics[k])
				return err
			})
			return err
		})
	}
	//log.Println(metrics)
	return metrics, err
}
