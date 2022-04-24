package hwmon

import (
	"bufio"
	"log"
	"os/exec"
	"strings"

	"github.com/firepear/gnatwren/internal/data"
)

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

func Gpuinfo(manu string) data.GPUdata {
	var gpudata data.GPUdata
	if manu == "nvidia" {
		GpuinfoNvidia(&gpudata)
	} else if manu == "amd" {
		GpuinfoAMD(&gpudata)
	}
	return gpudata
}

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
			gpudata.TempCur = v
		case "GPU Shutdown Temp":
			gpudata.TempMax = v
		case "Fan Speed":
			gpudata.Fan = v
		case "Power Draw":
			gpudata.PowCur = v
		case "Power Limit":
			gpudata.PowMax = v
		}
	}
	nvidiasmi.Wait()
}

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
}
