package system

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemService provides unified system monitoring functionality
type SystemService struct {
	cpuMonitor     *CPUMonitor
	memoryMonitor  *MemoryMonitor
	processMonitor *ProcessMonitor
	eventBus       events.IEventBus
	config         *SystemServiceConfig
}

// SystemServiceConfig contains configuration for the system service
type SystemServiceConfig struct {
	RefreshInterval    time.Duration `json:"refreshInterval"`
	MaxProcesses       int           `json:"maxProcesses"`
	EnableTemperature  bool          `json:"enableTemperature"`
	EnableNetwork      bool          `json:"enableNetwork"`
	EnableDisk         bool          `json:"enableDisk"`
	EnableDetailedInfo bool          `json:"enableDetailedInfo"`
}

// DefaultSystemServiceConfig returns default configuration
func DefaultSystemServiceConfig() *SystemServiceConfig {
	return &SystemServiceConfig{
		RefreshInterval:    1000 * time.Millisecond,
		MaxProcesses:       100,
		EnableTemperature:  true,
		EnableNetwork:      true,
		EnableDisk:         true,
		EnableDetailedInfo: false, // Disabled by default for performance
	}
}

// NewSystemService creates a new system service instance
func NewSystemService() *SystemService {
	return NewSystemServiceWithConfig(DefaultSystemServiceConfig())
}

// NewSystemServiceWithConfig creates a new system service with custom configuration
func NewSystemServiceWithConfig(config *SystemServiceConfig) *SystemService {
	return &SystemService{
		cpuMonitor:     NewCPUMonitor(),
		memoryMonitor:  NewMemoryMonitor(),
		processMonitor: NewProcessMonitor(),
		config:         config,
	}
}

// SetEventBus sets the event bus for the service
func (s *SystemService) SetEventBus(eventBus events.IEventBus) {
	s.eventBus = eventBus
}

// GetCPUMetrics returns current CPU metrics
func (s *SystemService) GetCPUMetrics(ctx context.Context) (*models.CPUMetrics, error) {
	metrics, err := s.cpuMonitor.GetCPUMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.cpu.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetMemoryMetrics returns current memory metrics
func (s *SystemService) GetMemoryMetrics(ctx context.Context) (*models.MemoryMetrics, error) {
	metrics, err := s.memoryMonitor.GetMemoryMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metrics: %w", err)
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.memory.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetProcessMetrics returns current process metrics
func (s *SystemService) GetProcessMetrics(ctx context.Context, limit int) (*models.ProcessMetrics, error) {
	if limit <= 0 {
		limit = s.config.MaxProcesses
	}

	metrics, err := s.processMonitor.GetProcessMetrics(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get process metrics: %w", err)
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.processes.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetDiskMetrics returns current disk metrics
func (s *SystemService) GetDiskMetrics(ctx context.Context) (*models.DiskMetrics, error) {
	if !s.config.EnableDisk {
		return nil, fmt.Errorf("disk monitoring is disabled")
	}

	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var diskInfos []models.DiskInfo
	var totalSpace, totalUsed uint64

	for _, partition := range partitions {
		// Skip non-filesystem partitions
		if partition.Fstype == "" {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			// Skip partitions we can't get usage for
			continue
		}

		diskInfo := models.DiskInfo{
			Device:       partition.Device,
			Mountpoint:   partition.Mountpoint,
			FSType:       partition.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
			InodesTotal:  usage.InodesTotal,
			InodesUsed:   usage.InodesUsed,
			InodesFree:   usage.InodesFree,
			ReadOnly:     partition.Opts != nil && contains(partition.Opts, "ro"),
		}

		diskInfos = append(diskInfos, diskInfo)
		totalSpace += usage.Total
		totalUsed += usage.Used
	}

	metrics := &models.DiskMetrics{
		Disks:      diskInfos,
		TotalSpace: totalSpace,
		TotalUsed:  totalUsed,
		TotalFree:  totalSpace - totalUsed,
		Timestamp:  time.Now(),
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.disk.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetNetworkMetrics returns current network metrics
func (s *SystemService) GetNetworkMetrics(ctx context.Context) (*models.NetworkMetrics, error) {
	if !s.config.EnableNetwork {
		return nil, fmt.Errorf("network monitoring is disabled")
	}

	// Get interface info
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Get IO counters
	ioCounters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network IO counters: %w", err)
	}

	// Create a map for quick lookup of IO counters by interface name
	ioMap := make(map[string]net.IOCountersStat)
	for _, io := range ioCounters {
		ioMap[io.Name] = io
	}

	var networkInterfaces []models.NetworkInterface
	var totalSent, totalRecv uint64

	for _, iface := range interfaces {
		// Skip loopback interfaces
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}

		// Get addresses as strings
		var addrs []string
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}

		networkIface := models.NetworkInterface{
			Name:        iface.Name,
			IsUp:        len(iface.Flags) > 0, // Simplified check
			IPAddresses: addrs,
			MAC:         iface.HardwareAddr,
			MTU:         uint64(iface.MTU),
		}

		// Add IO stats if available
		if io, ok := ioMap[iface.Name]; ok {
			networkIface.BytesSent = io.BytesSent
			networkIface.BytesRecv = io.BytesRecv
			networkIface.PacketsSent = io.PacketsSent
			networkIface.PacketsRecv = io.PacketsRecv
			networkIface.Errin = io.Errin
			networkIface.Errout = io.Errout
			networkIface.Dropin = io.Dropin
			networkIface.Dropout = io.Dropout
			totalSent += io.BytesSent
			totalRecv += io.BytesRecv
		}

		networkInterfaces = append(networkInterfaces, networkIface)
	}

	metrics := &models.NetworkMetrics{
		Interfaces:     networkInterfaces,
		TotalBytesSent: totalSent,
		TotalBytesRecv: totalRecv,
		Timestamp:      time.Now(),
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.network.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetTemperatureMetrics returns current temperature metrics
func (s *SystemService) GetTemperatureMetrics(ctx context.Context) (*models.TemperatureMetrics, error) {
	if !s.config.EnableTemperature {
		return nil, fmt.Errorf("temperature monitoring is disabled")
	}

	// This would need platform-specific temperature monitoring
	// For now, return empty metrics
	metrics := &models.TemperatureMetrics{
		Sensors:   []models.TemperatureSensor{},
		Timestamp: time.Now(),
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.temperature.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetAllMetrics returns a complete snapshot of all system metrics
func (s *SystemService) GetAllMetrics(ctx context.Context) (*models.SystemMetrics, error) {
	// Get all metrics concurrently for better performance
	metrics := &models.SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU metrics
	cpuMetrics, err := s.GetCPUMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}
	metrics.CPU = *cpuMetrics

	// Memory metrics
	memoryMetrics, err := s.GetMemoryMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metrics: %w", err)
	}
	metrics.Memory = *memoryMetrics

	// Process metrics
	processMetrics, err := s.GetProcessMetrics(ctx, s.config.MaxProcesses)
	if err != nil {
		return nil, fmt.Errorf("failed to get process metrics: %w", err)
	}
	metrics.Processes = *processMetrics

	// Disk metrics (optional)
	if s.config.EnableDisk {
		diskMetrics, err := s.GetDiskMetrics(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get disk metrics: %w", err)
		}
		metrics.Disks = *diskMetrics
	}

	// Network metrics (optional)
	if s.config.EnableNetwork {
		networkMetrics, err := s.GetNetworkMetrics(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get network metrics: %w", err)
		}
		metrics.Network = *networkMetrics
	}

	// Temperature metrics (optional)
	if s.config.EnableTemperature {
		tempMetrics, err := s.GetTemperatureMetrics(ctx)
		if err != nil {
			// Temperature metrics are optional, don't fail if unavailable
			metrics.Temperature = nil
		} else {
			metrics.Temperature = tempMetrics
		}
	}

	// Emit comprehensive update event
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "system.metrics.updated", map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "system-service")
	}

	return metrics, nil
}

// GetSystemInfo returns static system information
func (s *SystemService) GetSystemInfo(ctx context.Context) (*models.SystemInfo, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// Get current user information
	var username, homeDir string
	if u, err := user.Current(); err == nil {
		username = u.Username
		homeDir = u.HomeDir
	}

	// Get current working directory
	workingDir, _ := os.Getwd()

	// Get process ID
	pid := os.Getpid()

	info := &models.SystemInfo{
		Hostname:      hostInfo.Hostname,
		OS:            hostInfo.OS,
		OSVersion:     hostInfo.PlatformVersion,
		KernelVersion: hostInfo.KernelVersion,
		Architecture:  hostInfo.KernelArch,
		Uptime:        time.Duration(hostInfo.Uptime) * time.Second,
		BootTime:      time.Unix(int64(hostInfo.BootTime), 0),
		ProcessID:     pid,
		Username:      username,
		HomeDir:       homeDir,
		WorkingDir:    workingDir,
		Environment:   make(map[string]string),
	}

	// Add selective environment variables
	if s.config.EnableDetailedInfo {
		envVars := []string{"PATH", "HOME", "USER", "SHELL", "TERM", "LANG"}
		for _, env := range envVars {
			if value := os.Getenv(env); value != "" {
				info.Environment[env] = value
			}
		}
	}

	return info, nil
}

// GetSystemStatistics returns additional system statistics
func (s *SystemService) GetSystemStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// CPU statistics
	cpuStats, err := s.cpuMonitor.GetCPUStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU statistics: %w", err)
	}
	stats["cpu"] = cpuStats

	// Memory statistics
	memStats, err := s.memoryMonitor.GetMemoryStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory statistics: %w", err)
	}
	stats["memory"] = memStats

	// Process statistics
	processStats, err := s.processMonitor.GetProcessStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get process statistics: %w", err)
	}
	stats["processes"] = processStats

	// System info
	stats["runtime"] = map[string]interface{}{
		"go_version":    runtime.Version(),
		"go_os":         runtime.GOOS,
		"go_arch":       runtime.GOARCH,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"num_cgo_call":  runtime.NumCgoCall(),
	}

	// Add timestamp
	stats["timestamp"] = time.Now()

	return stats, nil
}

// StartMonitoring starts continuous monitoring
func (s *SystemService) StartMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(s.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Collect metrics
			metrics, err := s.GetAllMetrics(ctx)
			if err != nil {
				// Log error but continue monitoring
				fmt.Printf("Error collecting metrics: %v\n", err)
				continue
			}

			// Check for warnings
			warnings := s.checkWarnings(metrics)
			if len(warnings) > 0 && s.eventBus != nil {
				s.eventBus.Publish(ctx, "system.warnings", map[string]interface{}{
					"warnings":  warnings,
					"timestamp": time.Now(),
				}, "system-service")
			}
		}
	}
}

// checkWarnings checks for system warnings
func (s *SystemService) checkWarnings(metrics *models.SystemMetrics) []string {
	var warnings []string

	// CPU warnings
	if metrics.CPU.UsagePercent > 90 {
		warnings = append(warnings, "High CPU usage detected")
	}

	// Memory warnings
	if metrics.Memory.UsagePercent > 90 {
		warnings = append(warnings, "High memory usage detected")
	}

	// Swap warnings
	if metrics.Memory.SwapPercent > 50 {
		warnings = append(warnings, "High swap usage detected")
	}

	// Disk warnings
	for _, disk := range metrics.Disks.Disks {
		if disk.UsagePercent > 90 {
			warnings = append(warnings, fmt.Sprintf("Low disk space on %s", disk.Mountpoint))
		}
	}

	// Temperature warnings
	if metrics.Temperature != nil {
		critical := metrics.Temperature.GetCriticalSensors()
		if len(critical) > 0 {
			warnings = append(warnings, "Critical temperature detected")
		}

		high := metrics.Temperature.GetHighTempSensors()
		if len(high) > 0 {
			warnings = append(warnings, "High temperature detected")
		}
	}

	return warnings
}

// ValidateConfiguration validates the service configuration
func (s *SystemService) ValidateConfiguration() error {
	if s.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if s.config.RefreshInterval <= 0 {
		return fmt.Errorf("refresh interval must be positive")
	}

	if s.config.MaxProcesses <= 0 {
		return fmt.Errorf("max processes must be positive")
	}

	if s.config.RefreshInterval < 100*time.Millisecond {
		return fmt.Errorf("refresh interval too frequent (minimum 100ms)")
	}

	if s.config.MaxProcesses > 10000 {
		return fmt.Errorf("max processes too high (maximum 10000)")
	}

	return nil
}

// UpdateConfiguration updates the service configuration
func (s *SystemService) UpdateConfiguration(config *SystemServiceConfig) error {
	if err := s.ValidateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	s.config = config
	return nil
}

// GetConfiguration returns the current configuration
func (s *SystemService) GetConfiguration() *SystemServiceConfig {
	return s.config
}

// Reset resets all monitors
func (s *SystemService) Reset() {
	s.cpuMonitor.Reset()
	s.memoryMonitor.Reset()
	s.processMonitor.Reset()
}

// Utility functions

// contains checks if a string slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
