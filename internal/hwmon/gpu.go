package hwmon

import (
	"os/exec"
	"strings"

	"github.com/firepear/gnatwren/internal/data"
)

func GpuManu() string {
	cmd := "/bin/env lspci -mm | grep VGA | grep -i nvidia"
	vga, err := exec.Command("/bin/env", "bash", "-c", cmd).Output()
	if err != nil {
		return "nvidia"
	}
	if strings.Contains(string(vga), "AMD") {
		return "amd"
	}
	return "intel"
}

func Gpuinfo(manu string) data.GPUdata {
	if manu == "nvidia" {
		gpudata := GpuinfoNvidia()
	}
	return gpudata
}

func GpuinfoNvidia() data.GPUdata {
}
