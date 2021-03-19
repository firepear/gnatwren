package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
)


// the fake, empty response sent back to 'agentupdate' requests
var fresp []byte
// temporary struct to hold last-reported metrics until a datastore is
// implemented
var curMetrics = map[string]data.AgentPayload{}
var mux = &sync.RWMutex{}


func agentUpdate(args [][]byte) ([]byte, error) {
	var upd = data.AgentPayload{}
	err := json.Unmarshal(args[0], &upd)
	if err != nil {
		return fresp, err
	}

	mux.Lock()
	defer mux.Unlock()
	curMetrics[upd.Host] = upd
	return fresp, err
}


func query (args [][]byte) ([]byte, error) {
	var q = data.Query{}
	err := json.Unmarshal(args[0], &q)
	if err != nil {
		return fresp, err
	}

	if q.Op == "status" {
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

	config := data.GatherConfig{}
	content, err := os.ReadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &config)
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

	// register our handler function(s)
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
