package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
)

func main() {
	// set up configuration and create client instance
	conf := &petrel.ClientConfig{Addr: "localhost:11099"}
	c, err := petrel.TCPClient(conf)
	if err != nil {
		fmt.Printf("can't initialize client: %s\n", err)
	}
	defer c.Quit()

	// stitch together a query
	var reqh = []byte("query ")
	var req = data.Query{}
	req.Op = "status"
	reqj, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("could not marshal request: %s\n", err)
		os.Exit(1)
	}

	// and dispatch it to the server!
	resp, err := c.Dispatch(append(reqh, reqj...))
	if err != nil {
		fmt.Printf("did not get successful response: %s\n", err)
		os.Exit(1)
	}

	// print out what we got back
	metrics := map[string]data.AgentPayload{}
	err = json.Unmarshal(resp, &metrics);
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		os.Exit(1)
	}

	var hosts []string
	for k := range metrics {
		hosts = append(hosts, k)
	}
	sort.Strings(hosts)

	for _, hostname := range hosts  {
		hostdata := metrics[hostname]
		fmt.Printf("%s  ||  %s (%d threads)\n", hostname, hostdata.Cpu.Name, len(hostdata.Cpu.Cores))
		mincore, maxcore, avgcore, coretot := 0, 0, 0, 0
		for _, core := range hostdata.Cpu.Cores {
			clock_f, _ := strconv.ParseFloat(core, 64)
			clock := int(clock_f)
			coretot += clock
			if mincore == 0 || clock < mincore {
				mincore = clock
			}
			if clock > maxcore {
				maxcore = clock
			}
		}
		avgcore = coretot / len(hostdata.Cpu.Cores)
		uptime_f, _ := strconv.ParseFloat(hostdata.Upt, 64)
		uptime := int(uptime_f)
		d := uptime / 86400
		uptime = uptime - d * 86400
		h := uptime / 3600.0
		uptime = uptime - h * 3600
		m := uptime / 60
		s := uptime - m * 60
		ts := time.Unix(hostdata.TS, 0).Format("2 Jan 15:04:05")

		fmt.Printf("  Up %dd %02d:%02d:%02d  ||  Ldavg %s %s %s  ||  %s\n", int(d), int(h), int(m), int(s),
			hostdata.Ldavg[0], hostdata.Ldavg[1], hostdata.Ldavg[2], ts)
		fmt.Printf("  Min/max/avg %d / %d / %d MHz  ||  Temp %05.2fC\n", mincore, maxcore, avgcore, hostdata.Cpu.Temp)
		fmt.Printf("  Mem tot/free/avail %05.2fGB / %05.2f%% / %05.2f%%\n\n",
			float64(hostdata.Mem[0]) / 1024.0 / 1024.0,
			(float64(hostdata.Mem[1]) / float64(hostdata.Mem[0]) * 100),
			(float64(hostdata.Mem[2]) / float64(hostdata.Mem[0]) * 100))
	}
}
