package main

import (
	"log"
)

// the fake, empty response sent back to 'agentupdate'
// requests
var fresp []byte

func agentUpdate(args []byte) (uint16, []byte, error) {
	// send data to the DB. we're in blob mode, so all data will
	// be in the zeroth byteslice
	err := dbUpdate(args)
	if err != nil {
		log.Printf("agentUpdate: db err: %s", err)
		return 500, fresp, err
	}
	return 200, fresp, err
}
