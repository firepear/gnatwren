package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
)

func main() {
	hosts := []string{"node01", "node02", "node03", "node04", "node05", "node06", "nodepi02", "nodepi03", "nodepi04", }
	addrs := map[string]string{
		"node01": "10.1.10.201:11099", "node02": "10.1.10.202:11099", "node03": "10.1.10.203:11099",
		"node04": "10.1.10.204:11099", "node05": "10.1.10.205:11099", "node06": "10.1.10.206:11099",
		"nodepi01": "10.1.10.221:11099", "nodepi02": "10.1.10.222:11099",
		"nodepi03": "10.1.10.223:11099", "nodepi04": "10.1.10.224:11099",
	}

	for _, host := range hosts {
		// set up configuration and create client instance
		conf := &petrel.ClientConfig{Addr: addrs[host]}
		c, err := petrel.TCPClient(conf)
		if err != nil {
			fmt.Printf("can't initialize client: %s\n", err)
			continue
		}
		defer c.Quit()

		// stitch together the non-option arguments into a request
		req := []byte("gather")

		// and dispatch it to the server!
		resp, err := c.Dispatch(req)
		if err != nil {
			fmt.Printf("did not get successful response: %s\n", err)
			os.Exit(1)
		}

		// print out what we got back
		fmt.Println(host)
		metrics := data.AgentPayload{}
		err = json.Unmarshal(resp, &metrics);
		if err != nil {
			fmt.Printf("could not unmarshal json: %s\n", err)
			os.Exit(1)
		}

		mincore, maxcore, avgcore, coretot := 0, 0, 0, 0
		for _, core := range metrics.Cpu.Cores {
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
		avgcore = coretot / len(metrics.Cpu.Cores)
		uptime_f, _ := strconv.ParseFloat(metrics.Upt, 64)
		uptime := int(uptime_f)
		d := uptime / 86400
		uptime = uptime - d * 86400
		h := uptime / 3600.0
		uptime = uptime - h * 3600
		m := uptime / 60
		s := uptime - m * 60

		fmt.Printf("%s\t%d threads\tTemp: %05.2fC\n", metrics.Cpu.Name, len(metrics.Cpu.Cores), metrics.Cpu.Temp)
		fmt.Printf("\tAvg / Max / Min clocks: %d / %d / %d MHz\n", avgcore, maxcore, mincore)
		fmt.Printf("\tMemory: %05.2fGB (%05.2f%% avail)\n", float64(metrics.Mem[0]) / 1024.0 / 1024.0, (float64(metrics.Mem[1]) / float64(metrics.Mem[0]) * 100))
		fmt.Printf("\tUptime: %dd %02d:%02d:%02d\n\n", int(d), int(h), int(m), int(s))
	}
}
