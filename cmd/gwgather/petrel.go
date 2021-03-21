package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/firepear/gnatwren/internal/data"
)

// the fake, empty response sent back to 'agentupdate'
// requests
var fresp []byte

func agentUpdate(args [][]byte) ([]byte, error) {
	// vivify the update data
	var upd = data.AgentPayload{}
	err := json.Unmarshal(args[0], &upd)
	if err != nil {
		log.Printf("agentUpdate: json unmarshal err: %s", err)
		return fresp, err
	}

	// update nodeStatus
	newTS := [2]int64{}
	mux.Lock()
	// the first timestamp is now (check-in ts)
	newTS[0] = time.Now().Unix()
	// second timestamp is the hosts's reporting time (which can
	// be in the past due to event playback). only update if the
	// event timestamp is newer than what we have
	if upd.TS > nodeStatus[upd.Host][1] {
		newTS[1] = upd.TS
	} else {
		newTS[1] = nodeStatus[upd.Host][1]
	}
	nodeStatus[upd.Host] = newTS
	mux.Unlock()

	// send data to the DB
	err = dbUpdate(args[0], upd)
	if err != nil {
		log.Printf("agentUpdate: badgerdb err: %s", err)
	}
	return fresp, err
}


func queryHandler (args [][]byte) ([]byte, error) {
	var q = data.Query{}
	err := json.Unmarshal(args[0], &q)
	if err != nil {
		return fresp, err
	}

	switch q.Op {
	case "status":
		curMetrics, err := dbGetCurrentStats()
		respb, err := json.Marshal(curMetrics)
		return respb, err
		//case "dbstatus":
		//dbMetrics, err := dbGetDBStats()
		//respb, err := json.Marshal(curMetrics)
		//return respb, err
	}
	return fresp, err
}



