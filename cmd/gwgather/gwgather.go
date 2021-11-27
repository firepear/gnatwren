package main

import (
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
)


var (
	// gwgather config
	config data.GatherConfig
	// nodeStatus holds the last check-in time of nodes running
	// agents. mux is its lock
	nodeStatus = map[string][2]int64{}
	mux sync.RWMutex
	// db handle
	db *sql.DB
	// terminates event loop when false
	keepalive = true
)

func exportJSON() error {
	machStats, err := dbGetCurrentStats()
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(config.Files.JsonLoc)
	sb.WriteString("/machines.json")
	machStatsj, _ := json.Marshal(machStats)
	err = os.WriteFile(sb.String(), machStatsj, 0644)
	if err != nil {
		return err
	}

	cpuTemps, err := dbGetCPUTemps()
	if err != nil {
		return err
	}
	sb.Reset()
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
                log.Printf("failed to register responder 'agentupdate': %s\n", err)
                os.Exit(1)
        }

	// initialize database
	db, err = dbSetup(config.DB.Loc)
	if err != nil {
		log.Fatalf("sqlite: can't init db: %s", err)
	}
	dbLoadNodeStatus()
	defer db.Close()
	// do an initial pruning, then launch ticker for hourly table
	// rollover (set ticker for 10 minutes, but routine is a no-op
	// unless enough time has passed)
	dbPruneMigrate()
	prunetick := time.NewTicker(600 * time.Second)
	defer prunetick.Stop()

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


	log.Println("gwagent server up and listening")

        for keepalive {
                select {
                case msg := <-petrel.Msgr:
                        // handle messages from petrel
			switch msg.Code {
			case 199: // petrel quit
				log.Printf("petrel server has shut down. last msg received was: %s\n", msg)
				keepalive = false
				break
			case 599: // petrel network error (listener socket died)
				petrel.Quit()
				keepalive = false
				break
			default:
				// anything else we'll log to the console
				log.Printf("petrel: %s\n", msg)
			}
		case <-jsontick.C:
			err := exportJSON()
			if err != nil {
				log.Printf("couldn't export to json: %s\n", err)
			}
		case <-prunetick.C:
			dbPruneMigrate()
		case <-sigchan:
                        // OS signal. tell petrel to shut down, then quit
                        log.Println("OS signal received; shutting down")
                        petrel.Quit()
			keepalive = false
			break
                }
        }
}
