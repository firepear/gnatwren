package data

// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name string
	Cores map[string]string
	Temp float64
}
