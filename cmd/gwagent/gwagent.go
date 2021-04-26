package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/hwmon"
	"github.com/firepear/gnatwren/internal/data"
)


var (
	req = []byte("agentupdate ")
	nl = []byte("\n")
	mux = &sync.RWMutex{}
	stow = "/var/run/gnatwren/agent_metrics.log"
	arch = ""
	hostname = ""
)



func gatherMetrics() ([]byte, error) {
	metrics := data.AgentPayload{}

	metrics.Arch = arch
	metrics.Host = hostname
	metrics.TS = time.Now().Unix()
	metrics.Cpu = hwmon.Cpuinfo()
	metrics.Mem = hwmon.Meminfo()
	metrics.Ldavg = hwmon.Loadinfo()
	metrics.Upt = hwmon.Uptime()

	return json.Marshal(metrics)
}


func sendMetrics(pc *petrel.ClientConfig) {
	// get metrics for this run
	sample, err := gatherMetrics()
	// try to instantiate a petrel client
	c, err := petrel.TCPClient(pc)
	if err != nil {
		// on failure, stow metrics and return error
		log.Printf("can't initialize client: %s\n", err)
		err = stowMetrics(sample)
		if err != nil {
			log.Printf("metrics lost: %s\n", err)
		}
		log.Printf("metrics stowed\n")
		return
	}
	defer c.Quit()

	// we have a client; send metrics to gwgather
	_, err = c.Dispatch(append(req, sample...))
	if err != nil {
		// on failure, stow metrics
		log.Printf("can't initialize client: %s\n", err)
		err = stowMetrics(sample)
		if err != nil {
			log.Printf("metrics lost: %s\n", err)
		}
		log.Printf("metrics stowed\n")
	}
}


func stowMetrics(m []byte) error {
	// get write lock on the mux, then open the file
	mux.Lock()
	defer mux.Unlock()
	f, err := os.OpenFile(stow, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("stow open failed: %s", err)
	}
	defer f.Close()

	// have file and lock: stow data
	_, err = f.Write(append(m, nl...))
	if err != nil {
		err = fmt.Errorf("stow write failed: %s", err)
	}
	return err
}


func sendUndeliveredMetrics(pc *petrel.ClientConfig, c chan error) {
	// if the stow file doesn't exist, there's nothing to do
	if _, err := os.Stat(stow); os.IsNotExist(err) {
		c <- nil
		return
	}

	// if it does, aquire lock on mux. this is not just to prevent
	// file access but to prevent this func from running multiple
	// times simultaneously
	mux.Lock()
	defer mux.Unlock()

	sent := 0
	// try to instantiate a petrel client
	pet, err := petrel.TCPClient(pc)
	if err != nil {
		c <- fmt.Errorf("found stowed metrics but can't connect: %s; deferring\n", err)
		return
	}
	defer pet.Quit()

	// open the stow file and try to send the contents to gather
	// (one metrics set per line)
	f, err := os.Open(stow)
	if err != nil {
		c <- fmt.Errorf("found stowed metrics but can't open: %s; deferring\n", err)
		return
	}
	defer f.Close()
	log.Printf("found stowed metrics\n")

	// read the file line-by-line and report the metrics inside
	petok := true
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := scanner.Bytes()
		// TODO handle the actual response, to know how not to count dupes
		_, err = pet.Dispatch(append(req, m...))
		if err != nil {
			log.Printf("sent %d metrics then hit a problem: %s\n", sent, err)
			petok = false
			break
		}
		sent++
	}

	// we've looped through the whole stow file. if petok is still
	// true then we sent everything and can clean up
	if petok {
		f.Close()
		os.Remove(stow)
		log.Printf("sent %d metrics; done\n", sent)
	}
	c <- nil
}


func main() {
	// find out where the gwagent config file is and read it in
	var configfile string
	flag.StringVar(&configfile, "config", "/etc/gnatwren/agent.json", "Location of the gwagent config file")
	flag.Parse()
	config := data.AgentConfig{}
	content, err := os.ReadFile(configfile)
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


	// get machine architecture and hostname, once
	arch = hwmon.Arch()
	hostname, _ = os.Hostname()

	// handle any saved metrics, synchronously, if we have them
	c := make(chan error)
	go sendUndeliveredMetrics(pconf, c)
	err = <-c
	if err != nil {
		log.Printf("%s\n", err)
	}

	// now setup a ticker for future stowage checks
	stowtick := time.NewTicker(90 * time.Second)
	defer stowtick.Stop()
	// and a timer for metrics (with a random duration from our
	// list of intervals)
	metrictick := time.NewTimer(time.Duration(config.Intervals[rand.Intn(intlen)]) * time.Second)

	// client event loop
	keepalive := true
        for keepalive {
                select {
                case <-metrictick.C:
                        // gather metrics, try to ship them, and set a
                        // new timer for this case
			sendMetrics(pconf)
			metrictick = time.NewTimer(time.Duration(config.Intervals[rand.Intn(intlen)]) * time.Second)
		case <-stowtick.C:
			// see if there are undelivered metrics and
			// try to deliver them
			c := make(chan error)
			go sendUndeliveredMetrics(pconf, c)
			err := <-c
			if err != nil {
				log.Printf("%s\n", err)
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
