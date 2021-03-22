package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/firepear/petrel"
	"github.com/firepear/gnatwren/internal/data"
)

func dispatchQuery(c *petrel.Client, q data.Query, op string) []byte {
	q.Op = op
	var reqhead = []byte("query ")
	qj, err := json.Marshal(q)
	if err != nil {
		fmt.Printf("could not marshal request: %s\n", err)
		os.Exit(1)
	}
	resp, err := c.Dispatch(append(reqhead, qj...))
	if err != nil {
		fmt.Printf("did not get successful response: %s\n", err)
		os.Exit(1)
	}
	return resp
}

func printMetrics(resp []byte) {
	metrics := map[string]data.AgentStatus{}
	err := json.Unmarshal(resp, &metrics);
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
		checkin := metrics[hostname].TS
		hostdata := metrics[hostname].Payload
		fmt.Printf("%s  %s (%d threads)\n", hostname, hostdata.Cpu.Name, len(hostdata.Cpu.Cores))

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
		ts := time.Now().Unix() - checkin

		fmt.Printf("  SYS || Up %dd %02d:%02d:%02d  |  Ldavg %s  |  Chkd %ds ago\n", int(d), int(h), int(m), int(s), hostdata.Ldavg[2], ts)
		fmt.Printf("  CPU || Min/max/avg %d / %d / %d MHz  |  Temp %05.2fC\n", mincore, maxcore, avgcore, hostdata.Cpu.Temp)
		fmt.Printf("  MEM || Tot/free/avail %05.2fGB / %05.2f%% / %05.2f%%\n\n",
			float64(hostdata.Mem[0]) / 1024.0 / 1024.0,
			(float64(hostdata.Mem[1]) / float64(hostdata.Mem[0]) * 100),
			(float64(hostdata.Mem[2]) / float64(hostdata.Mem[0]) * 100))
	}
}


func printDBStatus (resp []byte) {
	metrics := data.DBStatus{}
	err := json.Unmarshal(resp, &metrics);
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		os.Exit(1)
	}

	nums := regexp.MustCompile(`[a-zA-Z]+`)
	newestInt, _ := strconv.Atoi(nums.Split(metrics.Newest, -1)[0])
	oldestInt, _ := strconv.Atoi(nums.Split(metrics.Oldest, -1)[0])
	keydiff := newestInt - oldestInt
	fmt.Printf("Number of rows currently existing: %d\n", metrics.Count)
	fmt.Printf("Oldest/newest key:                 %s / %s\n", metrics.Oldest, metrics.Newest)
	fmt.Printf("TS of oldest/newest key:           %s / %s\n", time.Unix(int64(oldestInt), 0), time.Unix(int64(newestInt), 0))
	fmt.Printf("Time diff btw keys:                %d secs (%3.2fh)\n", keydiff, (float64(keydiff) / 3600.0))
}



func main() {
	// set up configuration and create client instance
	conf := &petrel.ClientConfig{Addr: "localhost:11099"}
	c, err := petrel.TCPClient(conf)
	if err != nil {
		fmt.Printf("can't initialize client: %s\n", err)
	}
	defer c.Quit()

	var req = data.Query{}
	switch os.Args[1] {
	case "status":
		resp := dispatchQuery(c, req, os.Args[1])
		printMetrics(resp)
	case "dbstatus":
		resp := dispatchQuery(c, req, os.Args[1])
		printDBStatus(resp)
	default:
		fmt.Printf("bad query type: '%s'\n", os.Args[1])
	}
}
