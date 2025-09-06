package data

// gwagent configuration
type AgentConfig struct {
	GatherAddr string `json:"gather_addr"`
	Active     bool   `json:"active"`
	Intervals  []int  `json:"intervals"`
	Stowdir    string `json:"workdir"`
	Gpu        string `json:"gpu"`
}

// gwgather configuration
type GatherConfig struct {
	BindAddr string           `json:"bind_addr"`
	DB       GatherDBConfig   `json:"db"`
	Files    GatherFileConfig `json:"files"`
	Log      GatherLogConfig  `json:"log"`
	UI       GatherUIConfig   `json:"ui"`
}

type GatherLogConfig struct {
	File  string `json:"file"`
	Level string `json:"level"`
}

type GatherDBConfig struct {
	Loc  string `json:"location"`
	Hrs  int64  `json:"hours_retained"`
	Days int64  `json:"days_retained"`
}

type GatherFileConfig struct {
	JsonLoc string `json:"json_location"`
	JsonInt int64  `json:"json_interval"`
}

type GatherUIConfig struct {
	Title       string `json:"title"`
	TempHiCpu   int64  `json:"temp_hi_cpu"`
	TempCritCpu int64  `json:"temp_crit_cpu"`
}

// AgentStatus is a repackaged AgentPayload that adds the most recent
// check-in time from gwgather
type AgentStatus struct {
	TS      int64
	Payload string
}

// AgentPayload represents one sample, as collected by gwagent.
type AgentPayload struct {
	Host  string
	Arch  string
	TS    int64
	Cpu   CPUdata
	Gpu   GPUdata
	Mem   [3]int
	Ldavg string
	Upt   string
}

// The CPUdata struct is used by internal/hwmon to report the data
// collected on a machine's CPU. Name is the CPU name as reported by
// the OS, and Cores is a map of core ids to speeds in MHz.
type CPUdata struct {
	Name  string
	Cores map[string]string
	Temp  float64
}

type GPUdata struct {
	Name    string
	TempCur string
	TempMax string
	Fan     string
	PowCur  string
	PowMax  string
}

var(
	Ver = "0.20.0"
)
