package main

import (
	"strconv"

	"github.com/firepear/gnatwren/internal/data"
	badger "github.com/dgraph-io/badger/v3"
)

func dbUpdate(payload []byte, upd data.AgentPayload) error {
	// Open the Badger database
	db, err := badger.Open(badger.DefaultOptions(config.DB.Loc))
	if err != nil {
		return err
	}
	defer db.Close()

	// Your code here…
	key := []byte(strconv.Itoa(int(upd.TS)))
	key = append(key, []byte(upd.Host)...)
	err = db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, payload)
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

