package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/firepear/gnatwren/internal/hwmon"
	"github.com/firepear/gnatwren/internal/data"
	pc "github.com/firepear/petrel/client"
)


var (
	req = []byte("agentupdate ")
	nl = []byte("\n")
	mux = &sync.RWMutex{}
	// the directory and filename where we stow metrics that we
	// can't immediately report to gwgather
	stowdir = "/var/run/gnatwren"
	stow = fmt.Sprintf("%s/agent_metrics.log", stowdir)
	// the machine architecture
	arch = ""
	// cpu model name
	cpuname = ""
	// hostname
	hostname = ""
	// GPU manufacturer, and for non-Nvidia GPUs, the model name
	// and sysfs directory location
	gpumanu = ""
	gpuname = ""
	gpuloc = ""
)



func gatherMetrics() ([]byte, error) {
	metrics := data.AgentPayload{}

	metrics.Arch = arch
	metrics.Host = hostname
	metrics.TS = time.Now().Unix()
	metrics.Cpu = hwmon.Cpuinfo(cpuname)
	metrics.Gpu = hwmon.Gpuinfo(gpumanu, gpuname, gpuloc)
	metrics.Mem = hwmon.Meminfo()
	metrics.Ldavg = hwmon.Loadinfo()
	metrics.Upt = hwmon.Uptime()

	cpuname = metrics.Cpu.Name
	return json.Marshal(metrics)
}


func sendMetrics(pconf *pc.Config) {
	// get metrics for this run
	sample, err := gatherMetrics()
	// try to instantiate a petrel client
	c, err := pc.TCPClient(pconf)
	if err != nil {
		// on failure, stow metrics and return error
		log.Printf("can't initialize client: %s\n", err)
		err = stowMetrics(sample)
		if err != nil {
			log.Printf("metrics lost: %s\n", err)
			return
		}
		log.Printf("metrics stowed\n")
		return
	}
	defer c.Quit()

	// we have a client; send metrics to gwgather
	_, err = c.Dispatch(append(req, sample...))
	if err != nil {
		// on failure, stow metrics
		log.Printf("can't dispatch metrics: %s\n", err)
		err = stowMetrics(sample)
		if err != nil {
			log.Printf("metrics lost: %s\n", err)
			return
		}
		log.Printf("metrics stowed\n")
	}
}


func stowMetrics(m []byte) error {
	// create stowdir if it doesn't exist
	if _, err := os.Stat(stowdir); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(stowdir, 0755)
		if err != nil {
			return err
		}
	}
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


func sendUndeliveredMetrics(pconf *pc.Config, c chan error) {
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
	pet, err := pc.TCPClient(pconf)
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

	// read the file line-by-line and report the metrics inside
	petok := true
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := scanner.Bytes()
		// TODO handle the actual response, to know how not to count dupes
		_, err = pet.Dispatch(append(req, m...))
		if err != nil && ! errors.Is(err, io.EOF) {
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
		log.Printf("sent %d stowed metrics; done\n", sent)
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
        pconf := &pc.Config{Addr: config.GatherAddr}

	// set up a channel to handle termination events
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)


	// get machine architecture, hostname, and gpu maker (plus
	// other details, sometimes). these things are either
	// irritating or expensive to fetch, so we do it here, once
	arch = hwmon.Arch()
	hostname, _ = os.Hostname()
	gpumanu = hwmon.GpuManu()
	if gpumanu != "nvidia" {
		gpuname = hwmon.GpuName(gpumanu)
		gpuloc = hwmon.GpuSysfsLoc()
	}

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
