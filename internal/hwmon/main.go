package hwmon

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/firepear/gnatwren/internal/data"
)


// Cpuinfo scans /proc/cpuinfo and extracts values for the
// cpu name, Tdie temp, and the current speed of every core
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
	if len(procs) == 0 {
		procs = CpuinfoSysfs()
	}

	temp := Tempinfo()
	return data.CPUdata{Name: procname, Temp: (float64(temp) / 1000), Cores: procs}
}

// CpuinfoSysfs is the fallback function for gathering core speeds. It
// examines the /sys/devices/system/cpu/ subtree of sysfs
func CpuinfoSysfs() map[string]string {
	err := os.Chdir("/sys/devices/system/cpu")
	if err != nil {
		log.Fatalf("%v", err)
	}

	procs := map[string]string{}
	cpus, _ := filepath.Glob("cpu[0-9]*")
	for _, cpu := range cpus {
		// get just the number, to match Cpuinfo's naming
		cpunum := strings.Replace(cpu, "cpu", "", 1)
		// build a path to the freq file and slurp it
		path := fmt.Sprintf("/sys/devices/system/cpu/%s/cpufreq/scaling_cur_freq", cpu)
		freqb, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		// convert to match Cpuinfo's format
		freqi, _ := strconv.Atoi(strings.TrimSpace(string(freqb)))
		procs[cpunum] = fmt.Sprintf("%8.3f", (float64(freqi) / 1000))
	}
	return procs
}

func Loadinfo() [3]string {
	loadavg_b, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		log.Fatal(err)
	}
	loadavg := strings.Fields(string(loadavg_b))
	return [3]string{loadavg[0], loadavg[1], loadavg[2]}
}

// Meminfo scans the file /proc/meminfo and extracts the values for
// total and available memory, in kilobytes, and returns them in that
// order.
func Meminfo() [3]int {
	found := 0
	memdata := [3]int{}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if found == 3 {
			break
		}

		line := strings.Fields(scanner.Text())
		if line[0] == "MemTotal:" {
			memdata[0], err = strconv.Atoi(line[1])
			found += 1
		} else if line[0] == "MemFree:" {
			memdata[1], err = strconv.Atoi(line[1])
			found += 1
		} else if line[0] == "MemAvailable:" {
			memdata[2], err = strconv.Atoi(line[1])
			found += 1
		}
	}
	return memdata
}


// Tempinfo scans the /sys/class/hwmon tree, looking for a hwmonX
// subtree with a name of 'k10temp' (AMD) or 'cpu_thermal' (Raspberry
// Pi). It then examines the temp* files until it finds the one
// labelled 'Tdie', and checks its matching input to get the current
// CPU temperature (in millidegrees C). Returns -1 if no CPU temp can
// be found.
func Tempinfo() int {
	cputemp := -1

	// cd to the top of the tree
	err := os.Chdir("/sys/class/hwmon")
	if err != nil {
		log.Fatalf("%v", err)
	}
	// get a list of all hwmon dirs/symlinks
	hwmons, err := filepath.Glob("hwmon*")
	if len(hwmons) == 0 {
		return cputemp
	}

	for _, hwmon := range hwmons {
		// if cputemp has been set, then we found what we're
		// looking for and we're done
		if cputemp != -1 {
			break
		}

		// otherwise, build a path to the 'name' file in the
		// current hwmon dir and read its contents
		path := fmt.Sprintf("/sys/class/hwmon/%s/name", hwmon)
		name, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		if string(name) == "k10temp\n" {
			// we're only interested in "k10temp" on AMD
			// CPUs. build a list of the available temp
			// data source labels
			glob := fmt.Sprintf("/sys/class/hwmon/%s/temp?_label", hwmon)
			temps, _ := filepath.Glob(glob)
			if len(temps) == 0 {
				return cputemp
			}
			// and look at each of them
			for _, temp := range temps {
				label, err := ioutil.ReadFile(temp)
				if err != nil {
					log.Fatal(err)
				}
				labelstr := string(label)
				// we're only interested in the Tdie reading
				if labelstr != "Tdie\n" {
					continue
				}
				// when we find it, edit our path to point at
				// the temperature source value, and read it
				value, err := ioutil.ReadFile(strings.Replace(temp, "label", "input", 1))
				if err != nil {
					log.Fatal(err)
				}
				// it's []byte, so convert it to string, strip
				// the newline, and convert that to an
				// integer, which we will return
				cputemp, err = strconv.Atoi(strings.TrimSpace(string(value)))
				if err != nil {
					log.Fatal(err)
				}
				break
			}
		} else if string(name) == "cpu_thermal\n" {
			// Or 'cpu_thermal' on ARM. RPi doesn't have
			// labels, and there appears to only be one
			// input under the cpu_thermal entry. read it
			// and handle as above
			value, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/hwmon/%s/temp1_input", hwmon))
			if err != nil {
				log.Fatal(err)
			}
			cputemp, err = strconv.Atoi(strings.TrimSpace(string(value)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return cputemp
}


// Uptime reports the uptime count from /proc/uptime
func Uptime() string {
	content, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatal(err)
	}

	uptime := strings.Fields(string(content))[0]
	return uptime
}
