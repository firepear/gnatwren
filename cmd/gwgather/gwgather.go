package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
	badger "github.com/dgraph-io/badger/v3"
)


var (
	// gwgather config
	config data.GatherConfig
	// global placeholder for the db conn
	db *badger.DB
	// the fake, empty response sent back to 'agentupdate'
	// requests
	fresp []byte
	// nodeStatus holds the last check-in time of nodes running
	// agents. mux is its lock
	nodeStatus = map[string][2]int64{}
	mux sync.RWMutex
)


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


func query (args [][]byte) ([]byte, error) {
	var q = data.Query{}
	err := json.Unmarshal(args[0], &q)
	if err != nil {
		return fresp, err
	}

	if q.Op == "status" {
		curMetrics, err := dbGetCurrentStats()
		respb, err := json.Marshal(curMetrics)
		return respb, err
	}
	return fresp, err
}



func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/gather.json", "Location of the gwgather config file")
	flag.Parse()

	configstr, err := os.ReadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(configstr, &config)
	if err != nil {
		log.Fatal(err)
	}


	// set up a channel to handle termination events
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)


	// configure the petrel server
	c := &petrel.ServerConfig{
                Sockname: config.BindAddr,
                Msglvl: petrel.Error,
		Timeout: 5,
        }
	// and instantiate it
	s, err := petrel.TCPServer(c)
        if err != nil {
                log.Printf("could not instantiate Server: %s\n", err)
                os.Exit(1)
        }
	log.Printf("gwagent server instantiated")
	// then register handler function(s)
	err = s.Register("agentupdate", "blob", agentUpdate)
        if err != nil {
                log.Printf("failed to register responder 'agentupdate': %s", err)
                os.Exit(1)
        }
	err = s.Register("query", "blob", query)
        if err != nil {
                log.Printf("failed to register responder 'query': %s", err)
                os.Exit(1)
        }


	// Open the Badger database
	options := badger.DefaultOptions(config.DB.Loc)
	options.Logger = nil
	db, err = badger.Open(options)
	if err != nil {
		log.Fatalf("badger: can't open db: %s", err)
	}
	defer db.Close()
	// GC the DB
	_ := db.RunValueLogGC(0.7)
	// and launch a ticker for future GC
	dbgctick := time.NewTicker(2700 * time.Second)
	defer dbcgctick.Stop()


	keepalive := true
        for keepalive {
                select {
                case msg := <-s.Msgr:
                        // handle messages from petrel
			switch msg.Code {
			case 199: // petrel quit
				log.Printf("petrel server has shut down. last Msg received was: %s", msg)
				keepalive = false
				break
			case 599: // petrel network error (listener socket died)
				s.Quit()
				keepalive = false
				break
			default:
				// anything else we'll log to the console
				log.Printf("petrel: %s", msg)
			}
		case <-dbgc.C:
			// DB garbage collection
			_ := db.RunValueLogGC(0.7)
		case <-sigchan:
                        // OS signal. tell petrel to shut down, then
                        // shut ourselves down
                        log.Println("OS signal received; shutting down")
                        s.Quit()
			keepalive = false
			break
                }
        }
}
