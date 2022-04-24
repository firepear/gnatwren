package hwmon

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	//"regexp"
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
	return "WIP"
}


// GpuSysfsLoc looks through the /sys directory tree to find the hwmon
// directory corresponding to the first GPU in a system. This is later
// used by `Gpuinfo()`.
func GpuSysfsLoc() string {
	return "WIP"
}


// Gpuinfo is a top-level function for gathering GPU data. It will
// call an appropriate child function, based on GPU manufacturer, to
// do the actual data gathering.
func Gpuinfo(manu string) data.GPUdata {
	var gpudata data.GPUdata
	if manu == "nvidia" {
		GpuinfoNvidia(&gpudata)
	} else if manu == "amd" {
		//gpudata.Name = gpuname
		GpuinfoAMD(&gpudata)
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
			gpudata.Name = v
		case "GPU Current Temp":
			gpudata.TempCur = strings.ReplaceAll(v, " ", "")
		case "GPU Shutdown Temp":
			gpudata.TempMax = strings.ReplaceAll(v, " ", "")
		case "Fan Speed":
			gpudata.Fan = strings.ReplaceAll(v, " ", "")
		case "Power Draw":
			chunks = strings.Split(v, ".")
			gpudata.PowCur = fmt.Sprintf("%sW", chunks[0])
		case "Power Limit":
			chunks = strings.Split(v, ".")
			gpudata.PowMax = fmt.Sprintf("%sW", chunks[0])
		}
	}
	nvidiasmi.Wait()
}


// GpuinfoAMD gathers GPU status data for AMD GPUs.
func GpuinfoAMD(gpudata *data.GPUdata) {
	// available data is at /sys/class/drm/card0/device/hwmon/hwmonN
	//
	// relevant files are:
	//   temp1_input, temp1_crit
	//   power1_average, power1_cap_max
	//   fan1_input, fan1_max
	//   device (to get PCI ID)
	//
	// find GPU name by reading /usr/share/hwdata/pci.ids
	//   scan until line matching /^1002/
	//   then scan for /\tPCIID/ # minus the leading '0x'
	/*var (
		tempCur string
		tempCrit string
		powerCur string
		powerMax string
		fanSpeed string
	)*/
}
