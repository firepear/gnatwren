package data

type AgentConfig struct {
	GatherAddr string `json:"gather_addr"`
	Active     bool   `json:"active"`
	Intervals  []int  `json:"intervals"`
}

type GatherConfig struct {
	BindAddr string `json:"bind_addr"`
}


// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name string
	Cores map[string]string
	Temp float64
}


// AgentPayload represents one sample, as collected by gwagent.
type AgentPayload struct {
	Host string
	TS int64
	Cpu CPUdata
	Mem [3]int
	Ldavg [3]string
	Upt string
}

// Query represents a request for information from gwquery to
// gwgather. Op is the query operation to be performed. Hosts is a
// list of hostnames for limiting by host. Tbegin and Tend are
// timestamps delineating the timespan that data is being requested
// for, with Tend meaning "now" when not specified.
type Query struct {
	Op string
	Hosts []string
	Tbegin int64
	Tend int64
}
