package main

import (
	"log"
)

// the fake, empty response sent back to 'agentupdate'
// requests
var fresp []byte

func agentUpdate(args [][]byte) ([]byte, error) {
	// send data to the DB. we're in blob mode, so all data will
	// be in the zeroth byteslice
	err := dbUpdate(args[0])
	if err != nil {
		log.Printf("agentUpdate: db err: %s", err)
	}
	return fresp, err
}
