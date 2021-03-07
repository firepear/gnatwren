package data

// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name string
	Cores map[string]string
	Temp float64
}


type AgentConfig struct {
	GatherAddr string `json:"gather_addr"`
	Active     bool   `json:"active"`
	Intervals  []int  `json:"intervals"`
}

type GatherConfig struct {
	BindAddr string `json:"bind_addr"`
}


// AgentPayload represents one sample, as collected by gwagent.
type AgentPayload struct {
	Host string
	Cpu CPUdata
	Mem [3]int
	Ldavg [3]string
	Upt string
}
