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
	Loc string `json:"location"`
}

type GatherFileConfig struct {
	Enabled bool   `json:"enabled"`
	JsonLoc string `json:"json_location"`
	JsonInt int64  `json:"json_interval"`
}


// Query represents a request for information from gwquery to
// gwgather. Op is the query operation to be performed. Hosts is a
// list of hostnames for limiting by host. Tbegin and Tend are
// timestamps delineating the timespan that data is being requested
// for, with Tend meaning "now" when not specified/null
type Query struct {
	Op string      `json:"op"`
	Hosts []string `json:"hosts"`
	Tbegin int64   `json:"time_begin"`
	Tend int64     `json:"time_end"`
}


// AgentStatus is a repackaged AgentPayload that adds the most recent
// check-in time from gwgather
type AgentStatus struct {
	TS int64
	Payload AgentPayload
}

// AgentPayload represents one sample, as collected by gwagent.
type AgentPayload struct {
	Host string
	Arch string
	TS int64
	Cpu CPUdata
	Mem [3]int
	Ldavg [3]string
	Upt string
}

// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name string
	Cores map[string]string
	Temp float64
}

// DBStatus holds selected facts about the current state of the
// database
type DBStatus struct {
	Newest string
	Oldest string
	Count int
}
