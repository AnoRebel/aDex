package system

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"aDex-UI/internal/models"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryMonitor handles memory metrics collection
type MemoryMonitor struct {
	lastVirtualMemory *mem.VirtualMemoryStat
	lastSwapMemory    *mem.SwapMemoryStat
	isInitialized     bool
}

// NewMemoryMonitor creates a new memory monitor instance
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		isInitialized: false,
	}
}

// GetMemoryMetrics collects current memory metrics
func (mm *MemoryMonitor) GetMemoryMetrics(ctx context.Context) (*models.MemoryMetrics, error) {
	// Get virtual memory information
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory: %w", err)
	}

	// Get swap memory information
	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		// Swap memory might not be available on all systems
		swapMem = &mem.SwapMemoryStat{}
	}

	// Get host info for additional memory details
	_, err = host.InfoWithContext(ctx)
	if err != nil {
		// Host info is not critical, continue without it
	}

	// Build metrics
	metrics := &models.MemoryMetrics{
		Total:        virtualMem.Total,
		Available:    virtualMem.Available,
		Used:         virtualMem.Used,
		Free:         virtualMem.Free,
		UsagePercent: virtualMem.UsedPercent,
		Cached:       virtualMem.Cached,
		Buffers:      virtualMem.Buffers,
		SwapTotal:    swapMem.Total,
		SwapUsed:     swapMem.Used,
		SwapFree:     swapMem.Free,
		SwapPercent:  swapMem.UsedPercent,
		Timestamp:    time.Now(),
	}

	// Store for next calculation
	mm.lastVirtualMemory = virtualMem
	mm.lastSwapMemory = swapMem
	mm.isInitialized = true

	return metrics, nil
}

// GetVirtualMemory returns detailed virtual memory information
func (mm *MemoryMonitor) GetVirtualMemory(ctx context.Context) (*mem.VirtualMemoryStat, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory: %w", err)
	}
	return virtualMem, nil
}

// GetSwapMemory returns detailed swap memory information
func (mm *MemoryMonitor) GetSwapMemory(ctx context.Context) (*mem.SwapMemoryStat, error) {
	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap memory: %w", err)
	}
	return swapMem, nil
}

// GetMemoryDevices returns information about memory devices
func (mm *MemoryMonitor) GetMemoryDevices(ctx context.Context) (map[string]interface{}, error) {
	// This would need platform-specific implementation
	// For now, return basic memory information
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory devices: %w", err)
	}

	devices := map[string]interface{}{
		"total_memory":     virtualMem.Total,
		"available_memory": virtualMem.Available,
		"used_memory":      virtualMem.Used,
		"free_memory":      virtualMem.Free,
		"memory_type":      "unknown",
		"speed":            "unknown",
	}

	return devices, nil
}

// IsMemoryHighUsage checks if memory usage is above the threshold
func (mm *MemoryMonitor) IsMemoryHighUsage(threshold float64) (bool, error) {
	if !mm.isInitialized {
		return false, fmt.Errorf("memory monitor not initialized")
	}

	virtualMem, err := mem.VirtualMemoryWithContext(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to get current memory usage: %w", err)
	}

	return virtualMem.UsedPercent > threshold, nil
}

// IsSwapHighUsage checks if swap usage is above the threshold
func (mm *MemoryMonitor) IsSwapHighUsage(threshold float64) (bool, error) {
	if !mm.isInitialized {
		return false, fmt.Errorf("memory monitor not initialized")
	}

	swapMem, err := mem.SwapMemoryWithContext(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to get current swap usage: %w", err)
	}

	// If swap is not available or total is 0, return false
	if swapMem.Total == 0 {
		return false, nil
	}

	return swapMem.UsedPercent > threshold, nil
}

// GetMemoryPressure returns memory pressure information
func (mm *MemoryMonitor) GetMemoryPressure(ctx context.Context) (map[string]interface{}, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory pressure: %w", err)
	}

	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		swapMem = &mem.SwapMemoryStat{}
	}

	// Calculate memory pressure indicators
	pressure := map[string]interface{}{
		"memory_usage_percent": virtualMem.UsedPercent,
		"available_memory_mb":  virtualMem.Available / 1024 / 1024,
		"swap_usage_percent":   swapMem.UsedPercent,
		"cache_mb":             virtualMem.Cached / 1024 / 1024,
		"buffers_mb":           virtualMem.Buffers / 1024 / 1024,
		"pressure_level":       mm.calculatePressureLevel(virtualMem.UsedPercent, swapMem.UsedPercent),
		"recommendation":       mm.getMemoryRecommendation(virtualMem.UsedPercent, swapMem.UsedPercent),
	}

	return pressure, nil
}

// calculatePressureLevel determines memory pressure level
func (mm *MemoryMonitor) calculatePressureLevel(memUsage, swapUsage float64) string {
	if swapUsage > 50 {
		return "critical"
	}
	if memUsage > 90 {
		return "high"
	}
	if memUsage > 75 {
		return "moderate"
	}
	if memUsage > 50 {
		return "low"
	}
	return "normal"
}

// getMemoryRecommendation provides recommendations based on memory usage
func (mm *MemoryMonitor) getMemoryRecommendation(memUsage, swapUsage float64) string {
	if swapUsage > 50 {
		return "Consider adding more RAM or closing memory-intensive applications"
	}
	if memUsage > 90 {
		return "Close unnecessary applications immediately"
	}
	if memUsage > 75 {
		return "Monitor memory usage closely"
	}
	if memUsage > 50 {
		return "Memory usage is moderate"
	}
	return "Memory usage is normal"
}

// GetMemoryUsageHistory returns historical memory usage data
// This would need to be implemented with a time-series database or in-memory storage
func (mm *MemoryMonitor) GetMemoryUsageHistory(duration time.Duration) ([]models.MemoryMetrics, error) {
	// Placeholder implementation
	// In a real implementation, this would query a time-series database
	// or return data from an in-memory circular buffer
	return []models.MemoryMetrics{}, nil
}

// GetMemoryStatistics returns additional memory statistics
func (mm *MemoryMonitor) GetMemoryStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get current memory info
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory statistics: %w", err)
	}

	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		swapMem = &mem.SwapMemoryStat{}
	}

	// Basic statistics
	stats["total_memory_mb"] = virtualMem.Total / 1024 / 1024
	stats["used_memory_mb"] = virtualMem.Used / 1024 / 1024
	stats["free_memory_mb"] = virtualMem.Free / 1024 / 1024
	stats["available_memory_mb"] = virtualMem.Available / 1024 / 1024
	stats["memory_usage_percent"] = virtualMem.UsedPercent
	stats["cached_memory_mb"] = virtualMem.Cached / 1024 / 1024
	stats["buffers_memory_mb"] = virtualMem.Buffers / 1024 / 1024

	// Swap statistics
	stats["total_swap_mb"] = swapMem.Total / 1024 / 1024
	stats["used_swap_mb"] = swapMem.Used / 1024 / 1024
	stats["free_swap_mb"] = swapMem.Free / 1024 / 1024
	stats["swap_usage_percent"] = swapMem.UsedPercent

	// Additional calculations
	stats["shared_memory_mb"] = virtualMem.Shared / 1024 / 1024
	stats["active_memory_mb"] = virtualMem.Active / 1024 / 1024
	stats["inactive_memory_mb"] = virtualMem.Inactive / 1024 / 1024

	// Memory efficiency indicators
	stats["cache_efficiency"] = float64(virtualMem.Cached) / float64(virtualMem.Total) * 100
	stats["buffer_efficiency"] = float64(virtualMem.Buffers) / float64(virtualMem.Total) * 100

	// OS-specific information
	stats["os"] = runtime.GOOS
	stats["arch"] = runtime.GOARCH

	// Runtime memory information
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats["go_heap_alloc_mb"] = m.HeapAlloc / 1024 / 1024
	stats["go_heap_sys_mb"] = m.HeapSys / 1024 / 1024
	stats["go_heap_idle_mb"] = m.HeapIdle / 1024 / 1024
	stats["go_heap_inuse_mb"] = m.HeapInuse / 1024 / 1024
	stats["go_gc_pause_total_ns"] = m.PauseTotalNs

	return stats, nil
}

// GetMemoryUsageByProcess returns memory usage breakdown by top processes
func (mm *MemoryMonitor) GetMemoryUsageByProcess(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// This would need process monitoring integration
	// For now, return empty slice
	return []map[string]interface{}{}, nil
}

// ValidateMemoryMetrics validates memory metrics for consistency
func (mm *MemoryMonitor) ValidateMemoryMetrics(metrics *models.MemoryMetrics) error {
	if metrics == nil {
		return fmt.Errorf("memory metrics is nil")
	}

	// Validate total memory
	if metrics.Total <= 0 {
		return fmt.Errorf("invalid total memory: %d", metrics.Total)
	}

	// Validate memory components (should add up correctly)
	if metrics.Used+metrics.Free > metrics.Total {
		return fmt.Errorf("used + free memory exceeds total: %d + %d > %d",
			metrics.Used, metrics.Free, metrics.Total)
	}

	// Validate available memory
	if metrics.Available > metrics.Total {
		return fmt.Errorf("available memory exceeds total: %d > %d",
			metrics.Available, metrics.Total)
	}

	// Validate usage percentage
	if metrics.UsagePercent < 0 || metrics.UsagePercent > 100 {
		return fmt.Errorf("invalid memory usage percentage: %f", metrics.UsagePercent)
	}

	// Validate swap memory
	if metrics.SwapTotal > 0 {
		if metrics.SwapUsed+metrics.SwapFree > metrics.SwapTotal {
			return fmt.Errorf("swap used + free exceeds total: %d + %d > %d",
				metrics.SwapUsed, metrics.SwapFree, metrics.SwapTotal)
		}

		if metrics.SwapPercent < 0 || metrics.SwapPercent > 100 {
			return fmt.Errorf("invalid swap usage percentage: %f", metrics.SwapPercent)
		}
	}

	// Validate timestamp
	if metrics.Timestamp.IsZero() {
		return fmt.Errorf("memory metrics timestamp is zero")
	}

	return nil
}

// GetMemoryInfo returns detailed memory information
func (mm *MemoryMonitor) GetMemoryInfo(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Get virtual memory
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get swap memory
	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		swapMem = &mem.SwapMemoryStat{}
	}

	// Basic memory information
	info["total_memory"] = virtualMem.Total
	info["available_memory"] = virtualMem.Available
	info["used_memory"] = virtualMem.Used
	info["free_memory"] = virtualMem.Free
	info["percent_used"] = virtualMem.UsedPercent
	info["cached"] = virtualMem.Cached
	info["buffers"] = virtualMem.Buffers
	info["shared"] = virtualMem.Shared
	info["active"] = virtualMem.Active
	info["inactive"] = virtualMem.Inactive

	// Swap information
	info["swap_total"] = swapMem.Total
	info["swap_used"] = swapMem.Used
	info["swap_free"] = swapMem.Free
	info["swap_percent"] = swapMem.UsedPercent

	// Memory type information
	info["memory_type"] = "unknown"  // Would need platform-specific detection
	info["memory_speed"] = "unknown" // Would need platform-specific detection

	// Additional platform-specific info
	if runtime.GOOS == "linux" {
		// Linux-specific memory info could be added here
		info["hugepages_total"] = "unknown"
		info["hugepages_free"] = "unknown"
		info["hugepages_rsvd"] = "unknown"
	}

	return info, nil
}

// GetMemoryWarningThresholds returns memory warning thresholds
func (mm *MemoryMonitor) GetMemoryWarningThresholds() map[string]float64 {
	return map[string]float64{
		"memory_warning":  75.0, // Memory usage warning threshold
		"memory_critical": 90.0, // Memory usage critical threshold
		"swap_warning":    25.0, // Swap usage warning threshold
		"swap_critical":   50.0, // Swap usage critical threshold
	}
}

// CheckMemoryWarnings checks if memory usage exceeds warning thresholds
func (mm *MemoryMonitor) CheckMemoryWarnings(ctx context.Context) (map[string]interface{}, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check memory warnings: %w", err)
	}

	swapMem, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		swapMem = &mem.SwapMemoryStat{}
	}

	thresholds := mm.GetMemoryWarningThresholds()
	warnings := map[string]interface{}{
		"has_warnings": false,
		"warnings":     []string{},
	}

	// Check memory warnings
	if virtualMem.UsedPercent >= thresholds["memory_critical"] {
		warnings["has_warnings"] = true
		warnings["warnings"] = append(warnings["warnings"].([]string), "Critical memory usage")
	} else if virtualMem.UsedPercent >= thresholds["memory_warning"] {
		warnings["has_warnings"] = true
		warnings["warnings"] = append(warnings["warnings"].([]string), "High memory usage")
	}

	// Check swap warnings
	if swapMem.Total > 0 {
		if swapMem.UsedPercent >= thresholds["swap_critical"] {
			warnings["has_warnings"] = true
			warnings["warnings"] = append(warnings["warnings"].([]string), "Critical swap usage")
		} else if swapMem.UsedPercent >= thresholds["swap_warning"] {
			warnings["has_warnings"] = true
			warnings["warnings"] = append(warnings["warnings"].([]string), "High swap usage")
		}
	}

	return warnings, nil
}

// Reset resets the memory monitor state
func (mm *MemoryMonitor) Reset() {
	mm.lastVirtualMemory = nil
	mm.lastSwapMemory = nil
	mm.isInitialized = false
}

// FormatBytes formats bytes into human readable format
func (mm *MemoryMonitor) FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetMemoryBreakdown returns a detailed breakdown of memory usage
func (mm *MemoryMonitor) GetMemoryBreakdown(ctx context.Context) (map[string]interface{}, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory breakdown: %w", err)
	}

	breakdown := map[string]interface{}{
		"total": map[string]interface{}{
			"bytes": virtualMem.Total,
			"human": mm.FormatBytes(virtualMem.Total),
		},
		"used": map[string]interface{}{
			"bytes":   virtualMem.Used,
			"human":   mm.FormatBytes(virtualMem.Used),
			"percent": virtualMem.UsedPercent,
		},
		"free": map[string]interface{}{
			"bytes": virtualMem.Free,
			"human": mm.FormatBytes(virtualMem.Free),
		},
		"available": map[string]interface{}{
			"bytes": virtualMem.Available,
			"human": mm.FormatBytes(virtualMem.Available),
		},
		"cached": map[string]interface{}{
			"bytes": virtualMem.Cached,
			"human": mm.FormatBytes(virtualMem.Cached),
		},
		"buffers": map[string]interface{}{
			"bytes": virtualMem.Buffers,
			"human": mm.FormatBytes(virtualMem.Buffers),
		},
		"shared": map[string]interface{}{
			"bytes": virtualMem.Shared,
			"human": mm.FormatBytes(virtualMem.Shared),
		},
		"active": map[string]interface{}{
			"bytes": virtualMem.Active,
			"human": mm.FormatBytes(virtualMem.Active),
		},
		"inactive": map[string]interface{}{
			"bytes": virtualMem.Inactive,
			"human": mm.FormatBytes(virtualMem.Inactive),
		},
	}

	return breakdown, nil
}
