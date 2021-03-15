package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	metrics.TS = time.Now().Unix()
	metrics.Cpu = hwmon.Cpuinfo()
	metrics.Mem = hwmon.Meminfo()
	metrics.Ldavg = hwmon.Loadinfo()
	metrics.Upt = hwmon.Uptime()

	return json.Marshal(metrics)
}

func sendMetrics(pc *petrel.ClientConfig) error {
	sample, err := gatherMetrics()
	c, err := petrel.TCPClient(pc)
	if err != nil {
		err = fmt.Errorf("can't initialize client: %w\n", err)
		return err
	}
	defer c.Quit()

	_, err = c.Dispatch(append(req, sample...))
	return err
}


func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/agent.json", "Location of the gwagent config file")
	flag.Parse()
	config := data.AgentConfig{}
	content, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatalf("can't read config: %s; bailing", err)
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatalf("can't parse config as JSON: %s; bailing", err)
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
                        // sampling periods and schedules a message
                        // for that many seconds in the future. when it
                        // arrives, metrics are gathered and reported
			err := sendMetrics(pconf)
			for err != nil {
				log.Printf("unsuccessful update: %s\n", err)
				//time.Sleep(2 * time.Second)
				//err = sendMetrics(pconf)
			}
                case <-sigchan:
                        // we've trapped a signal from the OS. set
                        // keepalive to false and break out of our
                        // select (AKA terminate)
                        log.Println("OS signal received; shutting down")
                        keepalive = false
			break
                }
        }
}
