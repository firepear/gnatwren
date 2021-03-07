package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/hwmon"
	"github.com/firepear/gnatwren/internal/data"
)


var req = []byte("agentupdate ")


func gatherMetrics() ([]byte, error) {
	metrics := data.AgentPayload{}

	metrics.Host, _ = os.Hostname()
	metrics.Cpu = hwmon.Cpuinfo()
	metrics.Mem = hwmon.Meminfo()
	metrics.Ldavg = hwmon.Loadinfo()
	metrics.Upt = hwmon.Uptime()

	return json.Marshal(metrics)
}

func sendMetrics(pc *petrel.ClientConfig) {
	// TODO: don't die on errors, but store for later
	log.Printf("Sending data to Gather\n")
	sample, err := gatherMetrics()
	c, err := petrel.TCPClient(pc)
	if err != nil {
		log.Fatalf("can't initialize client: %s\n", err)
	}
	defer c.Quit()

	_, err = c.Dispatch(append(req, sample...))
	if err != nil {
		log.Fatalf("did not get successful response: %s\n", err)
	}
}


func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/agent.json", "Location of the gwagent config file")
	flag.Parse()
	config := data.AgentConfig{}
	content, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal(err)
	}
	// set up the things we need to pick our reporting intervals
	rand.Seed(time.Now().UnixNano())
	intlen := len(config.Intervals)
	// and the request we'll be making


        // set up client configuration and create client instance
        pconf := &petrel.ClientConfig{Addr: config.GatherAddr}

	// set up a channel to handle termination events
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// client event loop
	keepalive := true
        for keepalive {
                select {
                case <-time.After(time.Duration(config.Intervals[rand.Intn(intlen)]) * time.Second):
                        // this case selects one of our defined
                        // sampling periods and schedules an event for
                        // that many seconds in the future. if the
                        // event arrives, then we're still alive and
                        // we should report in.
			sendMetrics(pconf)
                case <-sigchan:
                        // we've trapped a signal from the OS. set
                        // keepalive to false and break out of our
                        // select
                        log.Println("OS signal received; shutting down")
                        keepalive = false
			break
                }
        }
}
