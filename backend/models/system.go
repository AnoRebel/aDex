package models

import "time"

// SystemInfo represents comprehensive system information
type SystemInfo struct {
	Hostname     string       `json:"hostname"`
	OS           OSInfo       `json:"os"`
	Hardware     HardwareInfo `json:"hardware"`
	Uptime       time.Duration `json:"uptime"`
	BootTime     time.Time    `json:"boot_time"`
	Timezone     string       `json:"timezone"`
	LoadAverage  []float64    `json:"load_average"`
	Processes    []Process    `json:"processes"`
	Services     []Service    `json:"services"`
}

// OSInfo represents operating system information
type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
	Distribution string `json:"distribution"`
}

// HardwareInfo represents hardware information
type HardwareInfo struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Storage []DiskInfo  `json:"storage"`
	Network []NICInfo   `json:"network"`
	GPU     []GPUInfo   `json:"gpu"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model     string    `json:"model"`
	Cores     int       `json:"cores"`
	Threads   int       `json:"threads"`
	Frequency float64   `json:"frequency"`
	Cache     CacheInfo `json:"cache"`
	Usage     CPUUsage  `json:"usage"`
}

// CPUUsage represents CPU usage metrics
type CPUUsage struct {
	Total    float64            `json:"total"`
	User     float64            `json:"user"`
	System   float64            `json:"system"`
	Idle     float64            `json:"idle"`
	IOWait   float64            `json:"io_wait"`
	PerCore  []float64          `json:"per_core"`
	Load     map[string]float64 `json:"load"`
}

// CacheInfo represents CPU cache information
type CacheInfo struct {
	L1 int `json:"l1"`
	L2 int `json:"l2"`
	L3 int `json:"l3"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total     uint64      `json:"total"`
	Used      uint64      `json:"used"`
	Free      uint64      `json:"free"`
	Available uint64      `json:"available"`
	Cached    uint64      `json:"cached"`
	Buffers   uint64      `json:"buffers"`
	Swap      SwapInfo    `json:"swap"`
	Usage     MemoryUsage `json:"usage"`
}

// SwapInfo represents swap memory information
type SwapInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// MemoryUsage represents memory usage metrics
type MemoryUsage struct {
	Percent    float64 `json:"percent"`
	AppsPercent float64 `json:"apps_percent"`
	WiredPercent float64 `json:"wired_percent"`
}

// DiskInfo represents disk/storage information
type DiskInfo struct {
	Device     string    `json:"device"`
	Mountpoint string    `json:"mountpoint"`
	Filesystem string    `json:"filesystem"`
	Total      uint64    `json:"total"`
	Used       uint64    `json:"used"`
	Free       uint64    `json:"free"`
	Percent    float64   `json:"percent"`
	ReadRate   float64   `json:"read_rate"`
	WriteRate  float64   `json:"write_rate"`
	IOStats    IOStats   `json:"io_stats"`
}

// IOStats represents I/O statistics
type IOStats struct {
	ReadOps   uint64 `json:"read_ops"`
	WriteOps  uint64 `json:"write_ops"`
	ReadBytes uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
}

// NICInfo represents network interface information
type NICInfo struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Status       string        `json:"status"`
	MAC          string        `json:"mac"`
	IPAddresses  []string      `json:"ip_addresses"`
	Speed        int64         `json:"speed"`
	Duplex       string        `json:"duplex"`
	MTU          int           `json:"mtu"`
	Stats        NetworkStats  `json:"stats"`
}

// NetworkStats represents network statistics
type NetworkStats struct {
	BytesReceived uint64 `json:"bytes_received"`
	BytesSent     uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent   uint64 `json:"packets_sent"`
	ErrorsIn      uint64 `json:"errors_in"`
	ErrorsOut     uint64 `json:"errors_out"`
	DropsIn       uint64 `json:"drops_in"`
	DropsOut      uint64 `json:"drops_out"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	Name     string  `json:"name"`
	Vendor   string  `json:"vendor"`
	Memory   uint64  `json:"memory"`
	Usage    float64 `json:"usage"`
	Temperature float64 `json:"temperature"`
	PowerUsage  float64 `json:"power_usage"`
}

// Process represents a system process
type Process struct {
	PID         int               `json:"pid"`
	Name        string            `json:"name"`
	CmdLine     string            `json:"cmdline"`
	State       string            `json:"state"`
	PPID        int               `json:"ppid"`
	UID         int               `json:"uid"`
	GID         int               `json:"gid"`
	CPUUsage    float64           `json:"cpu_usage"`
	MemoryUsage MemoryInfo        `json:"memory_usage"`
	StartTime   time.Time         `json:"start_time"`
	RunTime     time.Duration     `json:"run_time"`
	Threads     int               `json:"threads"`
	OpenFiles   int               `json:"open_files"`
	IOStats     ProcessIOStats    `json:"io_stats"`
	Children    []int             `json:"children"`
}

// ProcessIOStats represents process I/O statistics
type ProcessIOStats struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadOps    uint64 `json:"read_ops"`
	WriteOps   uint64 `json:"write_ops"`
}

// Service represents a system service
type Service struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	State       string    `json:"state"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Loaded      bool      `json:"loaded"`
	Active      bool      `json:"active"`
	Running     bool      `json:"running"`
	PID         int       `json:"pid"`
	StartTime   time.Time `json:"start_time"`
}

// SystemEvent represents system-related events
type SystemEvent struct {
	Type      string      `json:"type"`
	Source    string      `json:"source"`
	Message   string      `json:"message"`
	Severity  string      `json:"severity"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ResourceAlert represents a resource usage alert
type ResourceAlert struct {
	Type       string    `json:"type"`
	Resource   string    `json:"resource"`
	Threshold  float64   `json:"threshold"`
	CurrentValue float64 `json:"current_value"`
	Message    string    `json:"message"`
	Severity   string    `json:"severity"`
	Timestamp  time.Time `json:"timestamp"`
}
