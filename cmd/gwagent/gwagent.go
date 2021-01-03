package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/hwmon"
	"github.com/firepear/gnatwren/internal/data"
)

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

func gatherMetrics(args [][]byte) ([]byte, error) {
	metrics := data.AgentPayload{}

	metrics.Cpu = hwmon.Cpuinfo()
	metrics.Mem = hwmon.Meminfo()
	metrics.Upt = hwmon.Uptime()

	return json.Marshal(metrics)
}


func main() {
	var socket = flag.String("socket", "localhost:11099", "Addr:port to bind the socket to")
	flag.Parse()

	// set up a channel to handle termination events
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// configure the petrel server
	c := &petrel.ServerConfig{
                Sockname: *socket,
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
	err = s.Register("gather", "blob", gatherMetrics)
        if err != nil {
                log.Printf("failed to register responder 'gather': %s", err)
                os.Exit(1)
        }
	log.Printf("gather function registered; server ready")

	// create the 
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
