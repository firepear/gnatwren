package hwmon

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/firepear/gnatwren/internal/data"
)


// Cpuinfo scans the file /proc/cpuinfo and extracts values for the
// cpu name and the current speed of every core
func Cpuinfo() data.CPUdata {
	procs := map[string]string{}
	procname := ""
	procnum := ""

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) == 0 {
			continue
		}

		if line[0] == "processor" {
			procnum = line[2]
		} else if line[0] == "model" && line[1] == "name" {
			procname = strings.Join(line[3:], " ")
		} else if line[1] == "MHz" {
			procs[procnum] = line[3]
		}
	}
	return data.CPUdata{Name: procname, Cores: procs}
}

// Meminfo scans the file /proc/meminfo and extracts the values for
// total and available memory, in kilobytes, and returns them in that
// order.
func Meminfo() [2]int {
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


// Tempinfo scans the /sys/class/hwmon tree, looking for a hwmonX
// subtree with a name of 'k10temp'. It then examines the temp* files
// until it finds the one labelled 'Tdie', and checks its matching
// input to get the current CPU temperature (in millidegrees C)
func Tempinfo() int {
	return 0
}


// Uptime reports the uptime count from /proc/uptime
func Uptime() int {
	content, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatal(err)
	}

	uptime, err := strconv.ParseFloat(strings.Fields(string(content))[0], 64)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return int(uptime)

	//d :=  uptime / 86400
	//uptime = uptime - (86400 * d)
	//h := uptime / 3600
	//uptime = uptime - (3600 * h)
	//m := uptime / 60
	//s := uptime - (60 * m)
	//return [4]int{d, h, m, s}
}
