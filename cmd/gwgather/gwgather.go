package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
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


func exportJSON() error {
	cpuTemps, err := dbGetCPUTemps()
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(config.Files.JsonLoc)
	sb.WriteString("/cputemps.json")
	cpuTempsj, _ := json.Marshal(cpuTemps)
	err = os.WriteFile(sb.String(), cpuTempsj, 0644)
	if err != nil {
		return err
	}
	return err
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

	// configure the petrel server
	pc := &petrel.ServerConfig{
                Sockname: config.BindAddr,
                Msglvl: petrel.Error,
		Timeout: 5,
        }
	// and instantiate it
	petrel, err := petrel.TCPServer(pc)
        if err != nil {
                log.Printf("could not instantiate Server: %s\n", err)
                os.Exit(1)
        }
	// then register handler function(s)
	err = petrel.Register("agentupdate", "blob", agentUpdate)
        if err != nil {
                log.Printf("failed to register responder 'agentupdate': %s", err)
                os.Exit(1)
        }
	err = petrel.Register("query", "blob", queryHandler)
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

	// do an initial export of data as it stands
	err = exportJSON()
	if err != nil {
		log.Printf("couldn't export to json: %s\n", err)
	}
	// then launch a ticker to export every 5 min
	jsontick := time.NewTicker(300 * time.Second)
	defer jsontick.Stop()

	// set up a channel to handle termination events
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)


	log.Printf("gwagent server up and listening")

	keepalive := true
        for keepalive {
                select {
                case msg := <-petrel.Msgr:
                        // handle messages from petrel
			switch msg.Code {
			case 199: // petrel quit
				log.Printf("petrel server has shut down. last msg received was: %s", msg)
				keepalive = false
				break
			case 599: // petrel network error (listener socket died)
				petrel.Quit()
				keepalive = false
				break
			default:
				// anything else we'll log to the console
				log.Printf("petrel: %s", msg)
			}
		case <-jsontick.C:
			err := exportJSON()
			if err != nil {
				log.Printf("couldn't export to json: %s\n", err)
			}
		case <-dbgctick.C:
			// DB garbage collection
			_ = db.RunValueLogGC(0.7)
		case <-sigchan:
                        // OS signal. tell petrel to shut down, then quit
                        log.Println("OS signal received; shutting down")
                        petrel.Quit()
			keepalive = false
			break
                }
        }
}
