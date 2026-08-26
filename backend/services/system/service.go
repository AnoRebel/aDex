package system

import (
	"context"
	"errors"
	"fmt"
	stdnet "net"
	"os/user"
	"sort"
	"strconv"
	"time"

	"aDex-UI/backend/utils"
	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Service handles system monitoring and operations
type Service struct {
	platform     *utils.FeatureDetection
	isMonitoring bool
	stopChan     chan struct{}
}

// NewService creates a new system service instance
func NewService() *Service {
	return &Service{
		platform: utils.DetectPlatform(),
		stopChan: make(chan struct{}),
	}
}

// SystemInfo represents system information.
//
// JSON tags are explicit so Wails marshalling produces the field
// names the frontend reads (`platform`, `kernelVersion`, `architecture`,
// etc.). The two name pairs (OS/Platform, KernelVersion/Kernel) are
// duplicated under both JSON keys for compatibility with the older
// frontend reader that probed multiple shapes.
type SystemInfo struct {
	OS            string        `json:"platform"`
	Architecture  string        `json:"architecture"`
	KernelVersion string        `json:"kernelVersion"`
	Hostname      string        `json:"hostname"`
	Uptime        time.Duration `json:"uptime"`
	CPUUsage      float64       `json:"cpuUsage"`
	MemoryUsage   MemoryInfo    `json:"memoryUsage"`
	DiskUsage     []DiskInfo    `json:"diskUsage"`
	NetworkInfo   NetworkInfo   `json:"networkInfo"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Percent   float64
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Mountpoint string
	Total      uint64
	Used       uint64
	Free       uint64
	Percent    float64
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Interfaces []NetworkInterface
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name      string
	IPAddress string
	IsUp      bool
}

// GetSystemInfo retrieves current system information
func (s *Service) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	uptime := time.Duration(hostInfo.Uptime) * time.Second

	// Get network interfaces
	netInterfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var interfaces []NetworkInterface
	for _, iface := range netInterfaces {
		addrs := iface.Addrs
		var ipAddress string
		if len(addrs) > 0 {
			ipAddress = addrs[0].Addr
		}

		// Check if interface is up by checking if it has addresses
		isUp := len(addrs) > 0

		interfaces = append(interfaces, NetworkInterface{
			Name:      iface.Name,
			IPAddress: ipAddress,
			IsUp:      isUp,
		})
	}

	return &SystemInfo{
		OS:            hostInfo.OS,
		Architecture:  hostInfo.KernelArch,
		KernelVersion: hostInfo.KernelVersion,
		Hostname:      hostInfo.Hostname,
		Uptime:        uptime,
		NetworkInfo: NetworkInfo{
			Interfaces: interfaces,
		},
	}, nil
}

// CPUInfo represents CPU usage information including per-core data
type CPUInfo struct {
	Usage     float64   `json:"usage"`     // Overall CPU usage percentage
	Cores     []float64 `json:"cores"`     // Per-core usage percentages
	CoreCount int       `json:"coreCount"` // Number of logical cores
	ModelName string    `json:"modelName"` // CPU model name
	Frequency float64   `json:"frequency"` // CPU frequency in MHz
}

// GetCPUUsage retrieves current CPU usage (legacy - returns single value)
func (s *Service) GetCPUUsage(ctx context.Context) (float64, error) {
	percent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	if len(percent) > 0 {
		return percent[0], nil
	}
	return 0.0, nil
}

// GetCPUInfo retrieves detailed CPU information including per-core usage
func (s *Service) GetCPUInfo(ctx context.Context) (*CPUInfo, error) {
	// Get overall CPU usage
	overallPercent, err := cpu.PercentWithContext(ctx, time.Millisecond*500, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall CPU usage: %w", err)
	}

	// Get per-core CPU usage
	corePercent, err := cpu.PercentWithContext(ctx, time.Millisecond*500, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU usage: %w", err)
	}

	// Get CPU info (model, frequency, etc.)
	cpuInfos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		// Continue without CPU info - not critical
		cpuInfos = nil
	}

	// Get logical core count
	coreCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		coreCount = len(corePercent)
	}

	var modelName string
	var frequency float64
	if len(cpuInfos) > 0 {
		modelName = cpuInfos[0].ModelName
		frequency = cpuInfos[0].Mhz
	}

	var overallUsage float64
	if len(overallPercent) > 0 {
		overallUsage = overallPercent[0]
	}

	return &CPUInfo{
		Usage:     overallUsage,
		Cores:     corePercent,
		CoreCount: coreCount,
		ModelName: modelName,
		Frequency: frequency,
	}, nil
}

// TemperatureSensor mirrors a single entry from gopsutil's
// host.SensorsTemperatures, renamed into the camelCase shape the
// frontend's TemperatureMetrics type expects.
type TemperatureSensor struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
	High        float64 `json:"high"`
	Critical    float64 `json:"critical"`
}

// PowerInfo describes the system's power state. Source-of-truth is
// the canonical `distatus/battery` library (Linux 2.6.39+, macOS
// 10.10+, Windows XP+, *BSD, Solaris) — see
// https://github.com/distatus/battery. We translate its (possibly
// per-cell, possibly partial-error) output into a single normalised
// shape the frontend can render.
type PowerInfo struct {
	OnBattery bool   `json:"onBattery"`
	Percent   int    `json:"percent"`          // 0..100, -1 if unknown
	Status    string `json:"status,omitempty"` // Charging / Discharging / Full / Empty / Unknown
	Source    string `json:"source"`           // "AC" | "Battery"
}

// GetPowerInfo returns the current power source + battery percent.
// Defaults to {Source:"AC", Percent:-1} when the probe fails or the
// machine has no battery (desktop) — that's the right "we're plugged
// in" fallback for the eDex aesthetic.
//
// Edge cases handled (per distatus/battery open issues):
//   - Recent Apple Silicon Mac minis report a ghost battery with
//     Full == 0; we treat that as "no battery" and return AC.
//   - Some Linux kernels (Arch with newer power_now removal) return
//     a partial-error ErrPartial alongside usable Capacity values;
//     we accept the percent and ignore the error.
func (s *Service) GetPowerInfo(ctx context.Context) (*PowerInfo, error) {
	out := &PowerInfo{Source: "AC", Percent: -1}
	// Defensive recover — distatus #33 reports an open issue where
	// reading /sys/class/power_supply/BAT1/power_now panics on some
	// Arch kernels (the file exists but unreadable). The library has
	// since been hardened, but we still wrap the call so a panic in
	// any future upstream regression doesn't take the whole app down.
	var batteries []*battery.Battery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				batteries = nil
			}
		}()
		batteries, err = battery.GetAll()
	}()
	// `battery.ErrFatal` means the platform code blew up — fall back
	// to AC. A `battery.Errors` (slice) error is non-fatal: some
	// fields read fine, others didn't. Carry on with whatever we got.
	if err != nil {
		var fatal battery.ErrFatal
		if errors.As(err, &fatal) {
			return out, nil
		}
	}
	if len(batteries) == 0 {
		return out, nil
	}
	// Aggregate across all batteries (multi-cell laptops). Sum Full
	// + Current to compute one normalised percent; pick the dominant
	// state across cells.
	var total, current float64
	stateCounts := map[string]int{}
	for _, b := range batteries {
		if b == nil {
			continue
		}
		// Ghost-battery guard: M4 Mac mini reports Full == 0 with
		// non-zero adapter info. Skip cells that don't have a real
		// capacity reading.
		if b.Full <= 0 {
			continue
		}
		total += b.Full
		current += b.Current
		stateCounts[b.State.String()]++
	}
	if total == 0 {
		return out, nil
	}
	pct := int((current / total) * 100)
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	out.Percent = pct

	// Pick the most-common state. distatus reports per-cell, and on
	// most laptops there's only one battery anyway.
	dominant := ""
	maxCount := 0
	for st, c := range stateCounts {
		if c > maxCount {
			maxCount = c
			dominant = st
		}
	}
	out.Status = dominant
	switch dominant {
	case "Charging", "Full":
		out.Source = "AC"
	case "Discharging":
		out.Source = "Battery"
		out.OnBattery = true
	default:
		// "Empty" / "Unknown" — leave Source at the AC default.
	}
	return out, nil
}

// MeasureLatency does a TCP-handshake-based round-trip to the given
// host:port and returns the elapsed milliseconds. Defaults to
// 8.8.8.8:53 — a public DNS endpoint that's reachable from virtually
// every internet-connected host without ICMP/CAP_NET_RAW privileges
// (which a sandboxed Wails app rarely has). Returns -1 on timeout or
// any failure so the frontend can render "N/A" instead of stalling on
// "Measuring..." forever.
func (s *Service) MeasureLatency(ctx context.Context, target string) (float64, error) {
	if target == "" {
		target = "8.8.8.8:53"
	}
	timeout := 2 * time.Second
	if dl, ok := ctx.Deadline(); ok {
		if remaining := time.Until(dl); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	d := stdnet.Dialer{Timeout: timeout}
	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return -1, nil
	}
	defer conn.Close()
	return float64(time.Since(start).Microseconds()) / 1000.0, nil
}

// GetTemperatures returns all available hardware temperature sensors.
// Returns an empty slice (not error) on platforms / containers without
// sensor access — the frontend treats absent data as "no temp" and
// skips the panel row, which is preferable to showing an error toast
// every poll cycle.
func (s *Service) GetTemperatures(ctx context.Context) ([]TemperatureSensor, error) {
	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		// Sensor errors are common (Wayland flatpaks, locked-down
		// containers). Return empty rather than propagating so the UI
		// stays calm.
		return []TemperatureSensor{}, nil
	}
	out := make([]TemperatureSensor, 0, len(temps))
	for _, t := range temps {
		// Skip zero-reading entries (some ACPI sensors stream those).
		if t.Temperature <= 0 {
			continue
		}
		out = append(out, TemperatureSensor{
			Name:        t.SensorKey,
			Temperature: t.Temperature,
			High:        t.High,
			Critical:    t.Critical,
		})
	}
	return out, nil
}

// GetMemoryUsage retrieves current memory usage
func (s *Service) GetMemoryUsage(ctx context.Context) (*MemoryInfo, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}

	return &MemoryInfo{
		Total:     virtualMem.Total,
		Used:      virtualMem.Used,
		Available: virtualMem.Available,
		Percent:   virtualMem.UsedPercent,
	}, nil
}

// GetDiskUsage retrieves disk usage information
func (s *Service) GetDiskUsage(ctx context.Context) ([]DiskInfo, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var diskInfos []DiskInfo
	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue // Skip partitions we can't get usage for
		}

		diskInfos = append(diskInfos, DiskInfo{
			Mountpoint: partition.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			Percent:    usage.UsedPercent,
		})
	}

	return diskInfos, nil
}

// GetNetworkInfo retrieves network information
func (s *Service) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	netInterfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var interfaces []NetworkInterface
	for _, iface := range netInterfaces {
		addrs := iface.Addrs
		var ipAddress string
		if len(addrs) > 0 {
			ipAddress = addrs[0].Addr
		}

		// Check if interface is up by checking if it has addresses
		isUp := len(addrs) > 0

		interfaces = append(interfaces, NetworkInterface{
			Name:      iface.Name,
			IPAddress: ipAddress,
			IsUp:      isUp,
		})
	}

	return &NetworkInfo{
		Interfaces: interfaces,
	}, nil
}

// StartMonitoring starts system monitoring
func (s *Service) StartMonitoring(ctx context.Context, interval time.Duration) error {
	if s.isMonitoring {
		return fmt.Errorf("monitoring is already running")
	}

	s.isMonitoring = true
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		defer func() { s.isMonitoring = false }()

		for {
			select {
			case <-ticker.C:
				// This is where you would emit monitoring events
				// For now, we just collect data internally
				s.collectMetrics(ctx)

			case <-s.stopChan:
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// StopMonitoring stops system monitoring
func (s *Service) StopMonitoring() error {
	if !s.isMonitoring {
		return nil
	}

	close(s.stopChan)
	s.stopChan = make(chan struct{})
	s.isMonitoring = false

	return nil
}

// collectMetrics collects current system metrics
func (s *Service) collectMetrics(ctx context.Context) {
	// Collect CPU, memory, disk usage
	// This could emit events or store data for later retrieval
	// For now, this is a placeholder for monitoring logic
	_, _ = s.GetCPUUsage(ctx)
	_, _ = s.GetMemoryUsage(ctx)
	_, _ = s.GetDiskUsage(ctx)
}

// ProcessInfo represents process information
type ProcessInfo struct {
	PID           int     `json:"pid"`
	Name          string  `json:"name"`
	Command       string  `json:"command"`
	User          string  `json:"user"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent"`
	MemoryRSS     uint64  `json:"memoryRss"`
	NumThreads    int     `json:"numThreads"`
}

// GetTopProcesses returns top processes by CPU or memory usage
func (s *Service) GetTopProcesses(ctx context.Context, metric string, limit int) ([]ProcessInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var processInfos []ProcessInfo
	for _, p := range processes {
		info, err := s.getProcessInfo(ctx, p)
		if err != nil {
			continue
		}
		processInfos = append(processInfos, info)
	}

	// Sort by the specified metric
	switch metric {
	case "memory":
		sort.Slice(processInfos, func(i, j int) bool {
			return processInfos[i].MemoryPercent > processInfos[j].MemoryPercent
		})
	default: // default to CPU
		sort.Slice(processInfos, func(i, j int) bool {
			return processInfos[i].CPUPercent > processInfos[j].CPUPercent
		})
	}

	// Limit results
	if limit > 0 && len(processInfos) > limit {
		processInfos = processInfos[:limit]
	}

	return processInfos, nil
}

// getProcessInfo converts a gopsutil process to our model
func (s *Service) getProcessInfo(ctx context.Context, p *process.Process) (ProcessInfo, error) {
	pid := p.Pid

	name, err := p.NameWithContext(ctx)
	if err != nil {
		name = "unknown"
	}

	cmdline, err := p.CmdlineWithContext(ctx)
	if err != nil {
		cmdline = name
	}

	username := ""
	uids, err := p.UidsWithContext(ctx)
	if err == nil && len(uids) > 0 {
		if u, err := user.LookupId(strconv.Itoa(int(uids[0]))); err == nil {
			username = u.Username
		}
	}

	statusSlice, err := p.StatusWithContext(ctx)
	var status string
	if err != nil {
		status = "unknown"
	} else if len(statusSlice) > 0 {
		status = statusSlice[0]
	} else {
		status = "unknown"
	}

	cpuPercent, err := p.CPUPercentWithContext(ctx)
	if err != nil {
		cpuPercent = 0
	}

	memInfo, err := p.MemoryInfoWithContext(ctx)
	if err != nil {
		memInfo = &process.MemoryInfoStat{}
	}

	memPercent, err := p.MemoryPercentWithContext(ctx)
	if err != nil {
		memPercent = 0
	}

	numThreads, err := p.NumThreadsWithContext(ctx)
	if err != nil {
		numThreads = 0
	}

	return ProcessInfo{
		PID:           int(pid),
		Name:          name,
		Command:       cmdline,
		User:          username,
		Status:        status,
		CPUPercent:    cpuPercent,
		MemoryPercent: float64(memPercent),
		MemoryRSS:     memInfo.RSS,
		NumThreads:    int(numThreads),
	}, nil
}
