package system

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"aDex-UI/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

// CPUMonitor handles CPU metrics collection
type CPUMonitor struct {
	lastCPUTimes      []cpu.TimesStat
	lastCollectionTime time.Time
	isInitialized      bool
}

// NewCPUMonitor creates a new CPU monitor instance
func NewCPUMonitor() *CPUMonitor {
	return &CPUMonitor{
		isInitialized: false,
	}
}

// GetCPUMetrics collects current CPU metrics
func (cm *CPUMonitor) GetCPUMetrics(ctx context.Context) (*models.CPUMetrics, error) {
	// Get CPU usage percentages
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get per-core usage
	cpuPercentPerCore, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU usage: %w", err)
	}

	// Get CPU times
	cpuTimes, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %w", err)
	}

	// Get per-core times
	cpuTimesPerCore, err := cpu.TimesWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU times: %w", err)
	}

	// Get CPU info
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get host info for additional CPU details
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		// Host info is not critical, continue without it
		hostInfo = &host.InfoStat{}
	}

	// Build metrics
	metrics := &models.CPUMetrics{
		Timestamp: time.Now(),
	}

	// Overall CPU usage
	if len(cpuPercent) > 0 {
		metrics.UsagePercent = cpuPercent[0]
	}

	// Per-core usage
	metrics.PerCoreUsage = cpuPercentPerCore
	metrics.Cores = len(cpuPercentPerCore)

	// CPU info
	if len(cpuInfo) > 0 {
		info := cpuInfo[0]
		metrics.Model = info.ModelName
		metrics.Vendor = info.VendorID
		metrics.Frequency = float64(info.Mhz) // Current frequency in MHz
		metrics.FrequencyMax = float64(info.Mhz) // Max frequency (same as current for now)
	}

	// Get load averages (Unix-like systems only)
	if runtime.GOOS != "windows" {
		loadAvg, err := getLoadAverage(ctx)
		if err == nil {
			metrics.LoadAverage = loadAvg
		}
	}

	// Store times for next calculation
	cm.lastCPUTimes = cpuTimes
	cm.lastCollectionTime = metrics.Timestamp
	cm.isInitialized = true

	return metrics, nil
}

// GetCPUInfo returns detailed CPU information
func (cm *CPUMonitor) GetCPUInfo(ctx context.Context) ([]cpu.InfoStat, error) {
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}
	return cpuInfo, nil
}

// GetCPUTimes returns CPU time statistics
func (cm *CPUMonitor) GetCPUTimes(ctx context.Context) ([]cpu.TimesStat, error) {
	cpuTimes, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %w", err)
	}
	return cpuTimes, nil
}

// GetCPUTimesPerCore returns per-core CPU time statistics
func (cm *CPUMonitor) GetCPUTimesPerCore(ctx context.Context) ([]cpu.TimesStat, error) {
	cpuTimes, err := cpu.TimesWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU times: %w", err)
	}
	return cpuTimes, nil
}

// GetCPUCount returns the number of CPU cores
func (cm *CPUMonitor) GetCPUCount(ctx context.Context) (int, error) {
	cores, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU count: %w", err)
	}
	return cores, nil
}

// GetPhysicalCPUCount returns the number of physical CPU cores
func (cm *CPUMonitor) GetPhysicalCPUCount(ctx context.Context) (int, error) {
	cores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return 0, fmt.Errorf("failed to get physical CPU count: %w", err)
	}
	return cores, nil
}

// GetCPUFrequency returns current CPU frequencies
func (cm *CPUMonitor) GetCPUFrequency(ctx context.Context) (map[string]float64, error) {
	// Get current frequency for each core
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU frequency: %w", err)
	}

	frequencies := make(map[string]float64)
	for _, info := range cpuInfo {
		if info.Mhz > 0 {
			frequencies[fmt.Sprintf("cpu%d", info.CPU)] = float64(info.Mhz)
		}
	}

	return frequencies, nil
}

// IsCPUHighUsage checks if CPU usage is above the threshold
func (cm *CPUMonitor) IsCPUHighUsage(threshold float64) (bool, error) {
	if !cm.isInitialized {
		return false, fmt.Errorf("CPU monitor not initialized")
	}

	// We need to get current usage to check against threshold
	cpuPercent, err := cpu.Percent(context.Background(), 0, false)
	if err != nil {
		return false, fmt.Errorf("failed to get current CPU usage: %w", err)
	}

	if len(cpuPercent) > 0 {
		return cpuPercent[0] > threshold, nil
	}

	return false, fmt.Errorf("no CPU usage data available")
}

// GetCPUUsageHistory returns historical CPU usage data
// This would need to be implemented with a time-series database or in-memory storage
func (cm *CPUMonitor) GetCPUUsageHistory(duration time.Duration) ([]models.CPUMetrics, error) {
	// Placeholder implementation
	// In a real implementation, this would query a time-series database
	// or return data from an in-memory circular buffer
	return []models.CPUMetrics{}, nil
}

// CalculateCPUUsage calculates CPU usage percentage from time differences
func (cm *CPUMonitor) CalculateCPUUsage(current, previous cpu.TimesStat) float64 {
	if previous.Idle+previous.Iowait+previous.User+previous.System+previous.Nice+
		previous.Irq+previous.Softirq+previous.Steal == 0 {
		return 0
	}

	if current.Idle+current.Iowait+current.User+current.System+current.Nice+
		current.Irq+current.Softirq+current.Steal == 0 {
		return 0
	}

	prevIdle := previous.Idle + previous.Iowait
	prevNonIdle := previous.User + previous.System + previous.Nice +
		previous.Irq + previous.Softirq + previous.Steal
	prevTotal := prevIdle + prevNonIdle

	currentIdle := current.Idle + current.Iowait
	currentNonIdle := current.User + current.System + current.Nice +
		current.Irq + current.Softirq + current.Steal
	currentTotal := currentIdle + currentNonIdle

	totalDiff := currentTotal - prevTotal
	idleDiff := currentIdle - prevIdle

	if totalDiff == 0 {
		return 0
	}

	cpuPercentage := ((totalDiff - idleDiff) / totalDiff) * 100

	if cpuPercentage < 0 {
		return 0
	}
	if cpuPercentage > 100 {
		return 100
	}

	return cpuPercentage
}

// getLoadAverage retrieves system load averages (Unix-like systems only)
func getLoadAverage(ctx context.Context) ([]float64, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("load averages not available on Windows")
	}

	// This would need platform-specific implementation
	// For now, return empty slice
	return []float64{}, nil
}

// GetCPUTemperature returns CPU temperature if available
func (cm *CPUMonitor) GetCPUTemperature(ctx context.Context) (float64, error) {
	// This would need platform-specific temperature monitoring
	// For now, return 0 with an error indicating not implemented
	return 0, fmt.Errorf("CPU temperature monitoring not implemented")
}

// GetCPUThermalState returns thermal state information if available
func (cm *CPUMonitor) GetCPUThermalState(ctx context.Context) (string, error) {
	// This would need platform-specific thermal monitoring
	// For now, return empty string with an error indicating not implemented
	return "", fmt.Errorf("CPU thermal state monitoring not implemented")
}

// ValidateCPUMetrics validates CPU metrics for consistency
func (cm *CPUMonitor) ValidateCPUMetrics(metrics *models.CPUMetrics) error {
	if metrics == nil {
		return fmt.Errorf("CPU metrics is nil")
	}

	// Validate usage percentage
	if metrics.UsagePercent < 0 || metrics.UsagePercent > 100 {
		return fmt.Errorf("invalid CPU usage percentage: %f", metrics.UsagePercent)
	}

	// Validate core count
	if metrics.Cores <= 0 {
		return fmt.Errorf("invalid CPU core count: %d", metrics.Cores)
	}

	// Validate per-core usage
	if len(metrics.PerCoreUsage) != metrics.Cores {
		return fmt.Errorf("per-core usage length (%d) does not match core count (%d)",
			len(metrics.PerCoreUsage), metrics.Cores)
	}

	for i, usage := range metrics.PerCoreUsage {
		if usage < 0 || usage > 100 {
			return fmt.Errorf("invalid CPU usage for core %d: %f", i, usage)
		}
	}

	// Validate frequency
	if metrics.Frequency < 0 {
		return fmt.Errorf("invalid CPU frequency: %f", metrics.Frequency)
	}

	if metrics.FrequencyMax < 0 {
		return fmt.Errorf("invalid CPU max frequency: %f", metrics.FrequencyMax)
	}

	// Validate model and vendor
	if strings.TrimSpace(metrics.Model) == "" {
		return fmt.Errorf("CPU model name is empty")
	}

	if strings.TrimSpace(metrics.Vendor) == "" {
		return fmt.Errorf("CPU vendor is empty")
	}

	// Validate timestamp
	if metrics.Timestamp.IsZero() {
		return fmt.Errorf("CPU metrics timestamp is zero")
	}

	return nil
}

// GetCPUStatistics returns additional CPU statistics
func (cm *CPUMonitor) GetCPUStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get CPU count
	logicalCores, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get logical CPU count: %w", err)
	}
	stats["logical_cores"] = logicalCores

	// Get physical CPU count
	physicalCores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get physical CPU count: %w", err)
	}
	stats["physical_cores"] = physicalCores

	// Check if hyperthreading is enabled
	stats["hyperthreading_enabled"] = logicalCores > physicalCores

	// Get architecture
	stats["architecture"] = runtime.GOARCH

	// Get OS-specific CPU info
	if runtime.GOOS != "windows" {
		// Unix-like systems specific info
		loadAvg, err := getLoadAverage(ctx)
		if err == nil && len(loadAvg) >= 3 {
			stats["load_average_1min"] = loadAvg[0]
			stats["load_average_5min"] = loadAvg[1]
			stats["load_average_15min"] = loadAvg[2]
		}
	}

	// Get CPU cache info if available
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err == nil && len(cpuInfo) > 0 {
		info := cpuInfo[0]
		if info.CacheSize > 0 {
			stats["cache_size"] = info.CacheSize
		}
		if info.CoreID != "" {
			stats["core_id"] = info.CoreID
		}
		if info.PhysicalID != "" {
			stats["physical_id"] = info.PhysicalID
		}
	}

	return stats, nil
}

// GetCPUProcessAffinity returns CPU affinity information for a process
func (cm *CPUMonitor) GetCPUProcessAffinity(ctx context.Context, pid int) ([]int, error) {
	// This would need platform-specific implementation
	// For now, return all cores as the default affinity
	cpuCount, err := cm.GetCPUCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	affinity := make([]int, cpuCount)
	for i := 0; i < cpuCount; i++ {
		affinity[i] = i
	}

	return affinity, nil
}

// SetCPUProcessAffinity sets CPU affinity for a process
func (cm *CPUMonitor) SetCPUProcessAffinity(ctx context.Context, pid int, cores []int) error {
	// This would need platform-specific implementation
	return fmt.Errorf("CPU affinity setting not implemented")
}

// GetCPUPowerState returns power state information
func (cm *CPUMonitor) GetCPUPowerState(ctx context.Context) (map[string]interface{}, error) {
	// This would need platform-specific power management integration
	return map[string]interface{}{
		"performance_governor": "unknown",
		"power_management":     "unknown",
		"thermal_throttling":   false,
	}, nil
}

// Reset resets the CPU monitor state
func (cm *CPUMonitor) Reset() {
	cm.lastCPUTimes = nil
	cm.lastCollectionTime = time.Time{}
	cm.isInitialized = false
}