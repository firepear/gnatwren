package data

// gwagent configuration
type AgentConfig struct {
	GatherAddr string `json:"gather_addr"`
	Active     bool   `json:"active"`
	Intervals  []int  `json:"intervals"`
}


// gwgather configuration
type GatherConfig struct {
	BindAddr string            `json:"bind_addr"`
	Alerts   GatherAlertConfig `json:"alerts"`
	DB       GatherDBConfig    `json:"db"`
	Files    GatherFileConfig  `json:"files"`
}

type GatherAlertConfig struct {
	LateCheck int64 `json:"late_checkin"`
	OverTemp  int64 `json:"over_temp"`
}

type GatherDBConfig struct {
	Loc  string `json:"location"`
	Hrs  int64  `json:"hours_retained"`
	Days int64  `json:"days_retained"`
}

type GatherFileConfig struct {
	Enabled bool   `json:"enabled"`
	JsonLoc string `json:"json_location"`
	JsonInt int64  `json:"json_interval"`
}


// AgentStatus is a repackaged AgentPayload that adds the most recent
// check-in time from gwgather
type AgentStatus struct {
	TS int64
	Payload string
}

// AgentPayload represents one sample, as collected by gwagent.
type AgentPayload struct {
	Host string
	Arch string
	TS int64
	Cpu CPUdata
	Mem [3]int
	Ldavg string
	Upt string
}

// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name   string
	Cores  map[string]string
	Avgclk float64
	Temp   float64
}
