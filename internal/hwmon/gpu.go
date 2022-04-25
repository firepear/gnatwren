package hwmon

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/firepear/gnatwren/internal/data"
)

// GpuManu scrapes the output of `lspci` to determine the manufacturer
// of a machine's GPU.
func GpuManu() string {
	cmd := "/bin/env lspci -mm | grep VGA"
	vgabytes, err := exec.Command("/bin/env", "bash", "-c", cmd).Output()
	if err != nil {
		log.Println("Couldn't get GPU manu:", err)
		return "ERRNOGPUMANU"
	}
	vga := string(vgabytes)
	if strings.Contains(vga, "NVIDIA") {
		return "nvidia"
	}
	if strings.Contains(vga, "AMD") {
		return "amd"
	}
	return "intel"
}


// GpuName uses the pci.ids file to find the product name of a
// GPU. This is not needed for Nvidia GPUs.
func GpuName(manu string) string {
	var rmanu, rmodel *regexp.Regexp

	// pci.ids is organized by manufacturer, keying off
	// manufacturer id. manu section lines are the only lines
	// (other than comments) which do not have leading tabs. so
	// compile a regexp that lets us find the right section of the
	// file
	if manu == "amd" {
		rmanu, _ = regexp.Compile("^1002")
	}

	// the other thing we need before starting is to look up our
	// model id and construct a regexp from it
	modfile, err := os.Open("/sys/class/drm/card0/device/device")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer modfile.Close()
	scanner := bufio.NewScanner(modfile)
	// one line file, so scan just once, convert to string, and
	// trim hex marker
	scanner.Scan()
	modid := strings.TrimPrefix(scanner.Text(), "0x")
	// we'll be looking for a line that begins with a tab, then
	// our model id
	rmodel, _ = regexp.Compile(fmt.Sprintf("^\t%s", modid))

	// this variable flags when we have found the correct
	// manufacturer section
	foundmanu := false

	// open the file and start iterating over each line
	pcifile, err := os.Open("/usr/share/hwdata/pci.ids")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer pcifile.Close()
	scanner = bufio.NewScanner(pcifile)
	for scanner.Scan() {
		line := scanner.Text()
		// if we haven't gotten to our manu's section yet,
		// test to see if we're there now. if so, set the flag
		if !foundmanu {
			match := rmanu.MatchString(line)
			if match {
				foundmanu = true
			}
			continue
		}
		// if we make it down here, we're in the right
		// section. time to start looking for our card
		match := rmodel.MatchString(line)
		if match {
			// found it! extract the name + return it
			return strings.TrimPrefix(line, fmt.Sprintf("\t%s  ", modid))
		}
	}
	return "NONE"
}


// GpuSysfsLoc looks through the /sys directory tree to find the hwmon
// directory corresponding to the first GPU in a system. This is later
// used by `Gpuinfo()`.
func GpuSysfsLoc() string {
	gpus, err := filepath.Glob("/sys/class/drm/card0/device/hwmon/*")
	if err != nil {
		return "NONE"
	}
	return gpus[0]
}


// Gpuinfo is a top-level function for gathering GPU data. It will
// call an appropriate child function, based on GPU manufacturer, to
// do the actual data gathering.
func Gpuinfo(manu, name, loc string) data.GPUdata {
	var gpudata data.GPUdata
	if loc == "NONE" {
		return gpudata
	}

	if manu == "nvidia" {
		GpuinfoNvidia(&gpudata)
	} else if manu == "amd" {
		gpudata.Name = name
		GpuinfoAMD(&gpudata, loc)
	}
	return gpudata
}

// GpuinfoNvidia gethers GPU status data for Nvidia GPUs. Unlike other
// Gpuinfo* functions, which use sysfs, this function scrapes the
// output of `nvidia-smi -q`.
func GpuinfoNvidia(gpudata *data.GPUdata) {
	nvidiasmi := exec.Command("/usr/bin/nvidia-smi", "-q")
	stdout, _ := nvidiasmi.StdoutPipe()
	nvidiasmi.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		chunks := strings.Split(line, ":")
		if len(chunks) < 2 {
			continue
		}
		k := strings.TrimSpace(chunks[0])
		v := strings.TrimSpace(chunks[1])
		switch k {
		case "Product Name":
			gpudata.Name = strings.TrimPrefix(v, "NVIDIA ")
		case "GPU Current Temp":
			gpudata.TempCur = strings.ReplaceAll(v, " ", "")
		case "GPU Shutdown Temp":
			gpudata.TempMax = strings.ReplaceAll(v, " ", "")
		case "Fan Speed":
			if v == "N/A" {
				gpudata.Fan = v
			} else {
				gpudata.Fan = strings.ReplaceAll(v, " ", "")
			}
		case "Power Draw":
			gpudata.PowCur = strings.ReplaceAll(v, " ", "")
		case "Power Limit":
			gpudata.PowMax = strings.ReplaceAll(v, " ", "")
		}
	}
	nvidiasmi.Wait()
}


// GpuinfoAMD gathers GPU status data for AMD GPUs.
func GpuinfoAMD(gpudata *data.GPUdata, loc string) {
	log.Println(loc)
	//   power1_average, power1_cap_max
	//   fan1_input, fan1_max

	// temperature data
	file, err := os.Open(fmt.Sprintf("%s/temp1_input", loc))
	if err != nil {
		gpudata.TempCur = "N/A"
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		num, _ := strconv.Atoi(scanner.Text())
		// value is in millidegC
		gpudata.TempCur = fmt.Sprintf("%dC", num / 1000)
		file.Close()
	}
	file, err = os.Open(fmt.Sprintf("%s/temp1_crit", loc))
	if err != nil {
		gpudata.TempMax = "N/A"
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		num, _ := strconv.Atoi(scanner.Text())
		gpudata.TempMax = fmt.Sprintf("%dC", num / 1000)
		file.Close()
	}

	// power data
	file, err = os.Open(fmt.Sprintf("%s/power1_average", loc))
	if err != nil {
		gpudata.PowCur = "N/A"
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		num, _ := strconv.ParseFloat(scanner.Text(), 64)
		// value is in microW
		gpudata.PowCur = fmt.Sprintf("%.2fW", num / 1000000.0)
		file.Close()
	}
	file, err = os.Open(fmt.Sprintf("%s/power1_cap_max", loc))
	if err != nil {
		gpudata.PowMax = "N/A"
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		num, _ := strconv.ParseFloat(scanner.Text(), 64)
		gpudata.PowMax = fmt.Sprintf("%.2fW", num / 1000000.0)
		file.Close()
	}

	// fan data
	var fancur, fanmax int
	file, err = os.Open(fmt.Sprintf("%s/fan1_input", loc))
	if err != nil {
		fancur = -1
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		// value is in RPM, but we need an int to start with
		fancur, _ = strconv.Atoi(scanner.Text())
		file.Close()
	}
	file, err = os.Open(fmt.Sprintf("%s/fan1_max", loc))
	if err != nil {
		fanmax = -1
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		fancur, _ = strconv.Atoi(scanner.Text())
		file.Close()
	}
	// calculate % if we can, to match nvidia
	if fancur > -1 && fanmax > -1 {
		gpudata.Fan = fmt.Sprintf("%.0f%%", fancur * 100 / fanmax)
	} else if fancur > -1 {
		gpudata.Fan = fmt.Sprintf("%dRPM", fancur)
	} else {
		gpudata.Fan = "N/A"
	}
}
