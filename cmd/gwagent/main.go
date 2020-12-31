package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)


// gw_meminfo scans the file /proc/meminfo and extracts the values for
// total and available memory, in kilobytes, and returns them in that
// order.
func gw_meminfo() [2]int {
	found := 0
	memtotal := 0
	memavail := 0

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if found == 2 {
			break
		}

		line := strings.Fields(scanner.Text())
		if line[0] == "MemTotal:" {
			memtotal, err = strconv.Atoi(line[1])
			found += 1
		} else if line[0] == "MemAvailable:" {
			memavail, err = strconv.Atoi(line[1])
			found += 1
		}
	}
	return [2]int{memtotal, memavail}
}

func gw_uptime() [4]int {
	content, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatal(err)
	}

	chunk, err := strconv.ParseFloat(strings.Fields(string(content))[0], 64)
	if err != nil {
		log.Fatalf("%v", err)
	}
	uptime := int(chunk)

	d :=  uptime / 86400
	uptime = uptime - (86400 * d)
	h := uptime / 3600
	uptime = uptime - (3600 * h)
	m := uptime / 60
	s := uptime - (60 * m)
	return [4]int{d, h, m, s}
}


func main() {
	x := gw_meminfo()
	fmt.Printf("Total memory: %5.2fG\n", float64(x[0])/1024/1024)
	fmt.Printf("Available   : %5.2f%%\n", float64(x[1])/float64(x[0])*100)
	y := gw_uptime()
	fmt.Printf("Uptime      : %dd %d:%d:%d\n", y[0], y[1], y[2], y[3])
}
