package main

import (
	"fmt"
	"github.com/firepear/gnatwren/internal/hwmon"
)

func main() {
	x := hwmon.Meminfo()
	fmt.Printf("Total memory: %5.2fG\n", float64(x[0])/1024/1024)
	fmt.Printf("Available   : %5.2f%%\n", float64(x[1])/float64(x[0])*100)
	fmt.Println(hwmon.Uptime())
	//fmt.Printf("Uptime      : %dd %02d:%02d:%02d\n", y[0], y[1], y[2], y[3])
	fmt.Println(hwmon.Cpuinfo())
}
