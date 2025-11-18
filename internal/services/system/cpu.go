package system

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// Get sensor temperatures from the system
	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		// Try platform-specific methods as fallback
		return cm.getPlatformTemperature(ctx)
	}

	if len(temps) == 0 {
		return cm.getPlatformTemperature(ctx)
	}

	// Find CPU temperature from sensors
	// Common sensor names for CPU temperature
	cpuSensorNames := []string{
		"coretemp",
		"k10temp",
		"k8temp",
		"cpu_thermal",
		"cpu-thermal",
		"cpu",
		"Package id 0",
		"Core 0",
		"Tdie",
		"Tctl",
		"CPU",
	}

	var cpuTemp float64
	var found bool

	for _, temp := range temps {
		// Check if this is a CPU temperature sensor
		for _, name := range cpuSensorNames {
			if strings.Contains(strings.ToLower(temp.SensorKey), strings.ToLower(name)) {
				if temp.Temperature > cpuTemp {
					cpuTemp = temp.Temperature
					found = true
				}
				break
			}
		}
	}

	if !found {
		// If no specific CPU sensor found, use highest temperature as approximation
		for _, temp := range temps {
			if temp.Temperature > cpuTemp && temp.Temperature < 150 { // Sanity check
				cpuTemp = temp.Temperature
				found = true
			}
		}
	}

	if !found {
		return cm.getPlatformTemperature(ctx)
	}

	return cpuTemp, nil
}

// GetAllTemperatures returns all available temperature sensors
func (cm *CPUMonitor) GetAllTemperatures(ctx context.Context) (map[string]float64, error) {
	temps, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor temperatures: %w", err)
	}

	result := make(map[string]float64)
	for _, temp := range temps {
		result[temp.SensorKey] = temp.Temperature
	}

	return result, nil
}

// GetCPUThermalState returns thermal state information if available
func (cm *CPUMonitor) GetCPUThermalState(ctx context.Context) (string, error) {
	temp, err := cm.GetCPUTemperature(ctx)
	if err != nil {
		return "unknown", err
	}

	// Determine thermal state based on temperature
	// These thresholds are approximations for typical CPUs
	switch {
	case temp < 40:
		return "idle", nil
	case temp < 60:
		return "normal", nil
	case temp < 75:
		return "warm", nil
	case temp < 85:
		return "hot", nil
	case temp < 95:
		return "critical", nil
	default:
		return "emergency", nil
	}
}

// GetCPUTemperatureDetails returns detailed temperature information
func (cm *CPUMonitor) GetCPUTemperatureDetails(ctx context.Context) (*models.TemperatureInfo, error) {
	temp, err := cm.GetCPUTemperature(ctx)
	if err != nil {
		return nil, err
	}

	state, _ := cm.GetCPUThermalState(ctx)
	allTemps, _ := cm.GetAllTemperatures(ctx)

	// Find per-core temperatures if available
	coreTemps := make(map[string]float64)
	for key, val := range allTemps {
		if strings.Contains(strings.ToLower(key), "core") {
			coreTemps[key] = val
		}
	}

	return &models.TemperatureInfo{
		Current:         temp,
		State:           state,
		CoreTemperatures: coreTemps,
		Timestamp:       time.Now(),
	}, nil
}

// getPlatformTemperature attempts to get CPU temperature using platform-specific methods
func (cm *CPUMonitor) getPlatformTemperature(ctx context.Context) (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return cm.getLinuxTemperature()
	case "darwin":
		return cm.getMacOSTemperature()
	case "windows":
		return cm.getWindowsTemperature()
	default:
		return 0, fmt.Errorf("temperature monitoring not supported on %s", runtime.GOOS)
	}
}

// getLinuxTemperature reads CPU temperature from Linux sysfs
func (cm *CPUMonitor) getLinuxTemperature() (float64, error) {
	// Try common thermal zone paths
	thermalPaths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
		"/sys/class/hwmon/hwmon1/temp1_input",
		"/sys/devices/platform/coretemp.0/hwmon/hwmon*/temp1_input",
	}

	for _, path := range thermalPaths {
		// Handle glob patterns
		if strings.Contains(path, "*") {
			matches, err := filepath.Glob(path)
			if err != nil || len(matches) == 0 {
				continue
			}
			path = matches[0]
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Temperature is usually in millidegrees Celsius
		tempStr := strings.TrimSpace(string(data))
		temp, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			continue
		}

		// Convert from millidegrees to degrees if necessary
		if temp > 1000 {
			temp = temp / 1000
		}

		return temp, nil
	}

	return 0, fmt.Errorf("no temperature sensor found on Linux")
}

// getMacOSTemperature reads CPU temperature on macOS
func (cm *CPUMonitor) getMacOSTemperature() (float64, error) {
	// On macOS, temperature reading requires SMC access
	// which typically needs elevated permissions or a helper tool
	// For now, return an error as this requires additional native code
	return 0, fmt.Errorf("macOS temperature monitoring requires SMC access")
}

// getWindowsTemperature reads CPU temperature on Windows
func (cm *CPUMonitor) getWindowsTemperature() (float64, error) {
	// On Windows, temperature reading typically requires WMI or
	// hardware-specific APIs. The gopsutil library should handle this,
	// but if it fails, we return an error
	return 0, fmt.Errorf("Windows temperature monitoring requires WMI access")
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