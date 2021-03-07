package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
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
	log.Printf("Updated %s: %v", upd.Host, curMetrics[upd.Host])
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



// msgHandler takes care of messages which arrive on the Server's Msgr
// channel. It accepts all messages, but only handles 599 (network
// error) and 199 (shutdown)
func msgHandler(s *petrel.Server, msgchan chan error) {
        var msg *petrel.Msg
        keepalive := true

        for keepalive {
                msg = <-s.Msgr
                switch msg.Code {
                case 599:
                        s.Quit()
                        keepalive = false
                        msgchan <- msg
                case 199:
                        keepalive = false
                        msgchan <- msg
		default:
                        // anything else we'll log to the console to
                        // show what's going on under the hood!
                        log.Println(msg)

		}
        }
}

func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/gather.json", "Location of the gwgather config file")
	flag.Parse()

	config := data.GatherConfig{}
	content, err := ioutil.ReadFile(configfile)
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
                Msglvl: petrel.All,
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
                log.Printf("failed to register responder 'gather': %s", err)
                os.Exit(1)
        }

	// create the shutdown channel
	msgchan := make(chan error, 1)
        go msgHandler(s, msgchan)

	keepalive := true
        for keepalive {
                select {
                case msg := <-msgchan:
                        // we've been handed a Msg over msgchan, which
                        // means that our Server has shut down.exit
                        // this loop, causing main() to terminate.
                        log.Printf("Handler has shut down. Last Msg received was: %s", msg)
                        keepalive = false
                        break
                case <-sigchan:
                        // we've trapped a signal from the OS. tell
                        // our Server to shut down, but don't exit the
                        // eventloop because we want to handle the
                        // Msgs which will be incoming -- including
                        // the one we'll get on msgchan once the
                        // Server has finished its work.
                        log.Println("OS signal received; shutting down")
                        s.Quit()
                }
        }
}
