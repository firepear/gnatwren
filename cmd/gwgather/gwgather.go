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
	// nodeStatus holds the last check-in time of nodes running
	// agents. mux is its lock
	nodeStatus = map[string][2]int64{}
	mux sync.RWMutex
)


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
	err = s.Register("query", "blob", queryHandler)
        if err != nil {
                log.Printf("failed to register responder 'status': %s", err)
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
	_ = db.RunValueLogGC(0.7)
	// and launch a ticker for future GC
	dbgctick := time.NewTicker(2700 * time.Second)
	defer dbgctick.Stop()


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
		case <-dbgctick.C:
			// DB garbage collection
			_ = db.RunValueLogGC(0.7)
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
