package models

import (
	"time"
)

// SystemMetrics represents a complete snapshot of system metrics
type SystemMetrics struct {
	CPU         CPUMetrics         `json:"cpu"`
	Memory      MemoryMetrics      `json:"memory"`
	Processes   ProcessMetrics     `json:"processes"`
	Disks       DiskMetrics        `json:"disks"`
	Network     NetworkMetrics     `json:"network"`
	Temperature *TemperatureMetrics `json:"temperature,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

// CPUMetrics represents CPU utilization and information
type CPUMetrics struct {
	UsagePercent  float64   `json:"usagePercent"`  // Overall CPU usage percentage (0-100)
	Cores         int       `json:"cores"`         // Number of CPU cores
	Model         string    `json:"model"`         // CPU model name
	Vendor        string    `json:"vendor"`        // CPU vendor (Intel, AMD, etc.)
	Frequency     float64   `json:"frequency"`     // Current CPU frequency in MHz
	FrequencyMax  float64   `json:"frequencyMax"`  // Maximum CPU frequency in MHz
	PerCoreUsage  []float64 `json:"perCoreUsage"`  // Usage percentage per core
	LoadAverage   []float64 `json:"loadAverage"`   // 1min, 5min, 15min load averages (Unix-like systems)
	Timestamp     time.Time `json:"timestamp"`
}

// MemoryMetrics represents memory utilization
type MemoryMetrics struct {
	Total       uint64    `json:"total"`       // Total memory in bytes
	Available   uint64    `json:"available"`   // Available memory in bytes
	Used        uint64    `json:"used"`        // Used memory in bytes
	Free        uint64    `json:"free"`        // Free memory in bytes
	UsagePercent float64   `json:"usagePercent"` // Memory usage percentage (0-100)
	Cached      uint64    `json:"cached"`      // Cached memory in bytes (Unix-like systems)
	Buffers     uint64    `json:"buffers"`     // Buffer memory in bytes (Unix-like systems)
	SwapTotal   uint64    `json:"swapTotal"`   // Total swap space in bytes
	SwapUsed    uint64    `json:"swapUsed"`    // Used swap space in bytes
	SwapFree    uint64    `json:"swapFree"`    // Free swap space in bytes
	SwapPercent float64   `json:"swapPercent"` // Swap usage percentage (0-100)
	Timestamp   time.Time `json:"timestamp"`
}

// ProcessInfo represents information about a single process
type ProcessInfo struct {
	PID           int     `json:"pid"`           // Process ID
	PPID          int     `json:"ppid"`          // Parent Process ID
	Name          string  `json:"name"`          // Process name
	Command       string  `json:"command"`       // Full command line
	User          string  `json:"user"`          // Process owner
	Status        string  `json:"status"`        // Process status (Running, Sleeping, etc.)
	CPUPercent    float64 `json:"cpuPercent"`    // CPU usage percentage
	MemoryPercent float64 `json:"memoryPercent"` // Memory usage percentage
	MemoryRSS     uint64  `json:"memoryRSS"`     // Resident Set Size in bytes
	MemoryVMS     uint64  `json:"memoryVMS"`     // Virtual Memory Size in bytes
	CreateTime    time.Time `json:"createTime"`    // Process creation time
	NumThreads    int     `json:"numThreads"`    // Number of threads
	NumFDs        int     `json:"numFDs"`        // Number of file descriptors
	CWD           string  `json:"cwd"`           // Current working directory
	Executable    string  `json:"executable"`    // Executable path
}

// ProcessMetrics represents process information and statistics
type ProcessMetrics struct {
	Processes        []ProcessInfo `json:"processes"`        // List of processes (sorted by CPU usage)
	TotalProcesses   int           `json:"totalProcesses"`   // Total number of processes
	RunningProcesses int           `json:"runningProcesses"` // Number of running processes
	SleepingProcesses int          `json:"sleepingProcesses"` // Number of sleeping processes
	StoppedProcesses int          `json:"stoppedProcesses"` // Number of stopped processes
	ZombieProcesses  int          `json:"zombieProcesses"`  // Number of zombie processes
	Timestamp        time.Time     `json:"timestamp"`
}

// DiskInfo represents information about a disk/mount point
type DiskInfo struct {
	Device      string  `json:"device"`      // Device name (e.g., /dev/sda1)
	Mountpoint  string  `json:"mountpoint"`  // Mount point (e.g., /)
	FSType      string  `json:"fsType"`      // File system type (e.g., ext4, NTFS)
	Total       uint64  `json:"total"`       // Total space in bytes
	Used        uint64  `json:"used"`        // Used space in bytes
	Free        uint64  `json:"free"`        // Free space in bytes
	UsagePercent float64 `json:"usagePercent"` // Usage percentage (0-100)
	InodesTotal uint64  `json:"inodesTotal"` // Total inodes (Unix-like systems)
	InodesUsed  uint64  `json:"inodesUsed"`  // Used inodes (Unix-like systems)
	InodesFree  uint64  `json:"inodesFree"`  // Free inodes (Unix-like systems)
	ReadOnly    bool    `json:"readOnly"`    // Whether the file system is read-only
}

// DiskMetrics represents disk usage and I/O statistics
type DiskMetrics struct {
	Disks       []DiskInfo `json:"disks"`       // List of disk information
	TotalSpace  uint64     `json:"totalSpace"`  // Total space across all disks in bytes
	TotalUsed   uint64     `json:"totalUsed"`   // Total used space across all disks in bytes
	TotalFree   uint64     `json:"totalFree"`   // Total free space across all disks in bytes
	Timestamp   time.Time  `json:"timestamp"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name        string   `json:"name"`        // Interface name (e.g., eth0, wlan0)
	IsUp        bool     `json:"isUp"`        // Whether the interface is up
	BytesSent   uint64   `json:"bytesSent"`   // Total bytes sent
	BytesRecv   uint64   `json:"bytesRecv"`   // Total bytes received
	PacketsSent uint64   `json:"packetsSent"` // Total packets sent
	PacketsRecv uint64   `json:"packetsRecv"` // Total packets received
	Errin       uint64   `json:"errin"`       // Input errors
	Errout      uint64   `json:"errout"`      // Output errors
	Dropin      uint64   `json:"dropin"`      // Input packets dropped
	Dropout     uint64   `json:"dropout"`     // Output packets dropped
	IPAddresses []string `json:"ipAddresses"` // IP addresses assigned to this interface
	MAC         string   `json:"mac"`         // MAC address
	Speed       uint64   `json:"speed"`       // Interface speed in bits per second
	MTU         uint64   `json:"mtu"`         // Maximum transmission unit
}

// NetworkMetrics represents network interface statistics
type NetworkMetrics struct {
	Interfaces     []NetworkInterface `json:"interfaces"`     // List of network interfaces
	TotalBytesSent uint64             `json:"totalBytesSent"` // Total bytes sent across all interfaces
	TotalBytesRecv uint64             `json:"totalBytesRecv"` // Total bytes received across all interfaces
	Timestamp      time.Time          `json:"timestamp"`
}

// TemperatureSensor represents a temperature sensor
type TemperatureSensor struct {
	Name         string  `json:"name"`         // Sensor name (e.g., CPU, GPU, Motherboard)
	Temperature  float64 `json:"temperature"`  // Current temperature in Celsius
	Min          float64 `json:"min"`          // Minimum recorded temperature
	Max          float64 `json:"max"`          // Maximum recorded temperature
	Critical     float64 `json:"critical"`     // Critical temperature threshold
	High         float64 `json:"high"`         // High temperature threshold
	Unit         string  `json:"unit"`         // Temperature unit (typically Celsius)
	DevicePath   string  `json:"devicePath"`   // Device path (Linux sensors)
}

// TemperatureMetrics represents temperature sensor readings
type TemperatureMetrics struct {
	Sensors    []TemperatureSensor `json:"sensors"`    // List of temperature sensors
	Timestamp  time.Time           `json:"timestamp"`
}

// SystemInfo represents static system information
type SystemInfo struct {
	Hostname      string            `json:"hostname"`      // System hostname
	OS            string            `json:"os"`            // Operating system name
	OSVersion     string            `json:"osVersion"`     // OS version
	KernelVersion string            `json:"kernelVersion"` // Kernel version
	Architecture  string            `json:"architecture"`  // System architecture (x86_64, arm64, etc.)
	Uptime        time.Duration     `json:"uptime"`        // System uptime
	BootTime      time.Time         `json:"bootTime"`      // System boot time
	ProcessID     int               `json:"processId"`     // Current process ID
	Username      string            `json:"username"`      // Current username
	HomeDir       string            `json:"homeDir"`       // User home directory
	WorkingDir    string            `json:"workingDir"`    // Current working directory
	Environment   map[string]string `json:"environment"`   // Environment variables (selective)
}

// Helper functions for SystemMetrics

// GetHottestSensor returns the sensor with the highest temperature
func (tm *TemperatureMetrics) GetHottestSensor() *TemperatureSensor {
	if tm == nil || len(tm.Sensors) == 0 {
		return nil
	}

	var hottest *TemperatureSensor
	for i := range tm.Sensors {
		if hottest == nil || tm.Sensors[i].Temperature > hottest.Temperature {
			hottest = &tm.Sensors[i]
		}
	}
	return hottest
}

// GetColdestSensor returns the sensor with the lowest temperature
func (tm *TemperatureMetrics) GetColdestSensor() *TemperatureSensor {
	if tm == nil || len(tm.Sensors) == 0 {
		return nil
	}

	var coldest *TemperatureSensor
	for i := range tm.Sensors {
		if coldest == nil || tm.Sensors[i].Temperature < coldest.Temperature {
			coldest = &tm.Sensors[i]
		}
	}
	return coldest
}

// GetSensorByName returns a sensor by name
func (tm *TemperatureMetrics) GetSensorByName(name string) *TemperatureSensor {
	if tm == nil {
		return nil
	}

	for i := range tm.Sensors {
		if tm.Sensors[i].Name == name {
			return &tm.Sensors[i]
		}
	}
	return nil
}

// GetCriticalSensors returns all sensors that are above critical temperature
func (tm *TemperatureMetrics) GetCriticalSensors() []TemperatureSensor {
	if tm == nil {
		return nil
	}

	var critical []TemperatureSensor
	for _, sensor := range tm.Sensors {
		if sensor.Critical > 0 && sensor.Temperature >= sensor.Critical {
			critical = append(critical, sensor)
		}
	}
	return critical
}

// GetHighTempSensors returns all sensors that are above high temperature threshold
func (tm *TemperatureMetrics) GetHighTempSensors() []TemperatureSensor {
	if tm == nil {
		return nil
	}

	var high []TemperatureSensor
	for _, sensor := range tm.Sensors {
		if sensor.High > 0 && sensor.Temperature >= sensor.High {
			high = append(high, sensor)
		}
	}
	return high
}

// Helper functions for ProcessMetrics

// GetProcessByPID returns a process by its PID
func (pm *ProcessMetrics) GetProcessByPID(pid int) *ProcessInfo {
	if pm == nil {
		return nil
	}

	for i := range pm.Processes {
		if pm.Processes[i].PID == pid {
			return &pm.Processes[i]
		}
	}
	return nil
}

// GetProcessesByName returns all processes with the given name
func (pm *ProcessMetrics) GetProcessesByName(name string) []ProcessInfo {
	if pm == nil {
		return nil
	}

	var matches []ProcessInfo
	for _, proc := range pm.Processes {
		if proc.Name == name {
			matches = append(matches, proc)
		}
	}
	return matches
}

// GetTopCPUProcesses returns the top n processes by CPU usage
func (pm *ProcessMetrics) GetTopCPUProcesses(n int) []ProcessInfo {
	if pm == nil || n <= 0 {
		return nil
	}

	if n > len(pm.Processes) {
		n = len(pm.Processes)
	}

	return pm.Processes[:n]
}

// GetTopMemoryProcesses returns the top n processes by memory usage
func (pm *ProcessMetrics) GetTopMemoryProcesses(n int) []ProcessInfo {
	if pm == nil || n <= 0 {
		return nil
	}

	// Create a copy and sort by memory usage
	processes := make([]ProcessInfo, len(pm.Processes))
	copy(processes, pm.Processes)

	// Simple bubble sort for demonstration (replace with more efficient sort in production)
	for i := 0; i < len(processes)-1; i++ {
		for j := 0; j < len(processes)-i-1; j++ {
			if processes[j].MemoryPercent < processes[j+1].MemoryPercent {
				processes[j], processes[j+1] = processes[j+1], processes[j]
			}
		}
	}

	if n > len(processes) {
		n = len(processes)
	}

	return processes[:n]
}

// Helper functions for DiskMetrics

// GetDiskByMountpoint returns a disk by its mount point
func (dm *DiskMetrics) GetDiskByMountpoint(mountpoint string) *DiskInfo {
	if dm == nil {
		return nil
	}

	for i := range dm.Disks {
		if dm.Disks[i].Mountpoint == mountpoint {
			return &dm.Disks[i]
		}
	}
	return nil
}

// GetDiskByDevice returns a disk by its device name
func (dm *DiskMetrics) GetDiskByDevice(device string) *DiskInfo {
	if dm == nil {
		return nil
	}

	for i := range dm.Disks {
		if dm.Disks[i].Device == device {
			return &dm.Disks[i]
		}
	}
	return nil
}

// GetHighUsageDisks returns all disks with usage above the threshold
func (dm *DiskMetrics) GetHighUsageDisks(threshold float64) []DiskInfo {
	if dm == nil {
		return nil
	}

	var high []DiskInfo
	for _, disk := range dm.Disks {
		if disk.UsagePercent >= threshold {
			high = append(high, disk)
		}
	}
	return high
}

// Helper functions for NetworkMetrics

// GetInterfaceByName returns a network interface by name
func (nm *NetworkMetrics) GetInterfaceByName(name string) *NetworkInterface {
	if nm == nil {
		return nil
	}

	for i := range nm.Interfaces {
		if nm.Interfaces[i].Name == name {
			return &nm.Interfaces[i]
		}
	}
	return nil
}

// GetActiveInterfaces returns all interfaces that are up
func (nm *NetworkMetrics) GetActiveInterfaces() []NetworkInterface {
	if nm == nil {
		return nil
	}

	var active []NetworkInterface
	for _, iface := range nm.Interfaces {
		if iface.IsUp {
			active = append(active, iface)
		}
	}
	return active
}

// GetInterfacesByIP returns interfaces that have the given IP address
func (nm *NetworkMetrics) GetInterfacesByIP(ip string) []NetworkInterface {
	if nm == nil {
		return nil
	}

	var matches []NetworkInterface
	for _, iface := range nm.Interfaces {
		for _, ifaceIP := range iface.IPAddresses {
			if ifaceIP == ip {
				matches = append(matches, iface)
				break
			}
		}
	}
	return matches
}

// NewSystemMetrics creates a new SystemMetrics instance with current timestamp
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		Timestamp: time.Now(),
	}
}

// NewCPUMetrics creates a new CPUMetrics instance with current timestamp
func NewCPUMetrics() *CPUMetrics {
	return &CPUMetrics{
		Timestamp: time.Now(),
	}
}

// NewMemoryMetrics creates a new MemoryMetrics instance with current timestamp
func NewMemoryMetrics() *MemoryMetrics {
	return &MemoryMetrics{
		Timestamp: time.Now(),
	}
}

// NewProcessMetrics creates a new ProcessMetrics instance with current timestamp
func NewProcessMetrics() *ProcessMetrics {
	return &ProcessMetrics{
		Timestamp: time.Now(),
	}
}

// NewDiskMetrics creates a new DiskMetrics instance with current timestamp
func NewDiskMetrics() *DiskMetrics {
	return &DiskMetrics{
		Timestamp: time.Now(),
	}
}

// NewNetworkMetrics creates a new NetworkMetrics instance with current timestamp
func NewNetworkMetrics() *NetworkMetrics {
	return &NetworkMetrics{
		Timestamp: time.Now(),
	}
}

// NewTemperatureMetrics creates a new TemperatureMetrics instance with current timestamp
func NewTemperatureMetrics() *TemperatureMetrics {
	return &TemperatureMetrics{
		Timestamp: time.Now(),
	}
}