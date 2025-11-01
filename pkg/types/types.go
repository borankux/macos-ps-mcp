package types

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID       int32  `json:"pid"`
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Status    string `json:"status,omitempty"`
	User      string `json:"user,omitempty"`
	StartTime string `json:"start_time,omitempty"`
}

// WindowInfo represents information about an open window
type WindowInfo struct {
	Title    string `json:"title"`
	PID      int32  `json:"pid"`
	Process  string `json:"process"`
	AppName  string `json:"app_name,omitempty"`
	Geometry string `json:"geometry,omitempty"`
}

// PortInfo represents information about an open port
type PortInfo struct {
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	State    string `json:"state,omitempty"`
	LocalIP  string `json:"local_ip,omitempty"`
}

// ResourceUsage represents CPU and memory usage
type ResourceUsage struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	MemoryRSS     uint64  `json:"memory_rss"`   // Resident Set Size in bytes
	MemoryVMS     uint64  `json:"memory_vms"`   // Virtual Memory Size in bytes
	MemoryHuman   string  `json:"memory_human"` // Human readable memory
	CPUHuman      string  `json:"cpu_human"`    // Human readable CPU
	Threads       int32   `json:"threads,omitempty"`
	OpenFiles     int32   `json:"open_files,omitempty"`
}

// ServiceInfo represents a system service
type ServiceInfo struct {
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	PID           int32   `json:"pid,omitempty"`
	CPUPercent    float64 `json:"cpu_percent,omitempty"`
	MemoryPercent float32 `json:"memory_percent,omitempty"`
	MemoryHuman   string  `json:"memory_human,omitempty"`
	CPUHuman      string  `json:"cpu_human,omitempty"`
}

// Response types for MCP
type ProcessesResponse struct {
	Processes []ProcessInfo `json:"processes"`
	Count     int           `json:"count"`
}

type WindowsResponse struct {
	Windows []WindowInfo `json:"windows"`
	Count   int          `json:"count"`
}

type PortsResponse struct {
	Ports []PortInfo `json:"ports"`
	Count int        `json:"count"`
}

type ResourceResponse struct {
	Usage ResourceUsage `json:"usage"`
}

type ServicesResponse struct {
	Services []ServiceInfo `json:"services"`
	Count    int           `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
