package backend

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"aDex-UI/internal/models"
	"aDex-UI/internal/services/system"
)

func TestSystemMetricsCollection(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()
	ctx := context.Background()

	// Test CPU metrics collection
	t.Run("CPU Metrics Collection", func(t *testing.T) {
		cpuMetrics, err := systemService.GetCPUMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cpuMetrics)

		// Verify CPU metrics structure
		assert.GreaterOrEqual(t, cpuMetrics.UsagePercent, 0.0)
		assert.LessOrEqual(t, cpuMetrics.UsagePercent, 100.0)
		assert.GreaterOrEqual(t, cpuMetrics.Cores, 1)
		assert.NotEmpty(t, cpuMetrics.Model)
		assert.NotEmpty(t, cpuMetrics.Vendor)

		// Test per-core metrics
		assert.Len(t, cpuMetrics.PerCoreUsage, int(cpuMetrics.Cores))
		for _, usage := range cpuMetrics.PerCoreUsage {
			assert.GreaterOrEqual(t, usage, 0.0)
			assert.LessOrEqual(t, usage, 100.0)
		}

		// Test frequency
		assert.Greater(t, cpuMetrics.Frequency, 0.0)
		assert.Greater(t, cpuMetrics.FrequencyMax, 0.0)
	})

	// Test memory metrics collection
	t.Run("Memory Metrics Collection", func(t *testing.T) {
		memMetrics, err := systemService.GetMemoryMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, memMetrics)

		// Verify memory metrics structure
		assert.Greater(t, memMetrics.Total, uint64(0))
		assert.GreaterOrEqual(t, memMetrics.Available, uint64(0))
		assert.GreaterOrEqual(t, memMetrics.Used, uint64(0))
		assert.GreaterOrEqual(t, memMetrics.Free, uint64(0))

		// Verify logical consistency
		assert.Equal(t, memMetrics.Total, memMetrics.Used+memMetrics.Free)
		assert.LessOrEqual(t, memMetrics.Used, memMetrics.Total)

		// Test usage percentage calculation
		expectedUsage := float64(memMetrics.Used) / float64(memMetrics.Total) * 100
		assert.InDelta(t, expectedUsage, memMetrics.UsagePercent, 1.0)

		// Test swap metrics (if available)
		if memMetrics.SwapTotal > 0 {
			assert.GreaterOrEqual(t, memMetrics.SwapUsed, uint64(0))
			assert.GreaterOrEqual(t, memMetrics.SwapFree, uint64(0))
			assert.Equal(t, memMetrics.SwapTotal, memMetrics.SwapUsed+memMetrics.SwapFree)
		}
	})

	// Test process metrics collection
	t.Run("Process Metrics Collection", func(t *testing.T) {
		procMetrics, err := systemService.GetProcessMetrics(ctx, 100) // Get top 100 processes
		require.NoError(t, err)
		assert.NotNil(t, procMetrics)

		// Should have at least some processes
		assert.Greater(t, len(procMetrics.Processes), 0)
		assert.LessOrEqual(t, len(procMetrics.Processes), 100)

		// Verify process metrics structure
		for _, proc := range procMetrics.Processes {
			assert.Greater(t, proc.PID, 0)
			assert.NotEmpty(t, proc.Name)
			assert.GreaterOrEqual(t, proc.CPUPercent, 0.0)
			assert.GreaterOrEqual(t, proc.MemoryPercent, 0.0)
			assert.Greater(t, proc.MemoryRSS, uint64(0))
			assert.Greater(t, proc.MemoryVMS, uint64(0))

			// CPU usage should be reasonable
			assert.LessOrEqual(t, proc.CPUPercent, 100.0*float64(runtime.NumCPU()))

			// Memory usage should be reasonable
			assert.LessOrEqual(t, proc.MemoryRSS, proc.MemoryVMS)
		}

		// Verify sorting (highest CPU usage first)
		for i := 1; i < len(procMetrics.Processes); i++ {
			assert.GreaterOrEqual(t, procMetrics.Processes[i-1].CPUPercent, procMetrics.Processes[i].CPUPercent)
		}

		// Verify total counts
		assert.Greater(t, procMetrics.TotalProcesses, 0)
		assert.GreaterOrEqual(t, procMetrics.RunningProcesses, 0)
		assert.GreaterOrEqual(t, procMetrics.SleepingProcesses, 0)
	})

	// Test disk metrics collection
	t.Run("Disk Metrics Collection", func(t *testing.T) {
		diskMetrics, err := systemService.GetDiskMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, diskMetrics)

		// Should have at least one disk
		assert.Greater(t, len(diskMetrics.Disks), 0)

		// Verify disk metrics structure
		for _, disk := range diskMetrics.Disks {
			assert.NotEmpty(t, disk.Device)
			assert.NotEmpty(t, disk.Mountpoint)
			assert.Greater(t, disk.Total, uint64(0))
			assert.GreaterOrEqual(t, disk.Used, uint64(0))
			assert.GreaterOrEqual(t, disk.Free, uint64(0))

			// Verify logical consistency
			assert.Equal(t, disk.Total, disk.Used+disk.Free)
			assert.LessOrEqual(t, disk.Used, disk.Total)

			// Test usage percentage calculation
			expectedUsage := float64(disk.Used) / float64(disk.Total) * 100
			assert.InDelta(t, expectedUsage, disk.UsagePercent, 1.0)

			// File system type should be present
			assert.NotEmpty(t, disk.FSType)
		}

		// Verify total calculations
		var totalSpace, totalUsed uint64
		for _, disk := range diskMetrics.Disks {
			totalSpace += disk.Total
			totalUsed += disk.Used
		}
		assert.Greater(t, diskMetrics.TotalSpace, uint64(0))
		assert.Greater(t, diskMetrics.TotalUsed, uint64(0))
		assert.Equal(t, diskMetrics.TotalSpace, diskMetrics.TotalUsed+diskMetrics.TotalFree)
	})

	// Test network metrics collection
	t.Run("Network Metrics Collection", func(t *testing.T) {
		netMetrics, err := systemService.GetNetworkMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, netMetrics)

		// Should have at least one network interface
		assert.Greater(t, len(netMetrics.Interfaces), 0)

		// Verify network metrics structure
		for _, iface := range netMetrics.Interfaces {
			assert.NotEmpty(t, iface.Name)
			assert.GreaterOrEqual(t, iface.BytesSent, uint64(0))
			assert.GreaterOrEqual(t, iface.BytesRecv, uint64(0))
			assert.GreaterOrEqual(t, iface.PacketsSent, uint64(0))
			assert.GreaterOrEqual(t, iface.PacketsRecv, uint64(0))

			// Test error counters
			assert.GreaterOrEqual(t, iface.Errin, uint64(0))
			assert.GreaterOrEqual(t, iface.Errout, uint64(0))
			assert.GreaterOrEqual(t, iface.Dropin, uint64(0))
			assert.GreaterOrEqual(t, iface.Dropout, uint64(0))

			// IP addresses should be valid if present
			if len(iface.IPAddresses) > 0 {
				for _, ip := range iface.IPAddresses {
					assert.NotEmpty(t, ip)
				}
			}
		}

		// Verify total calculations
		var totalSent, totalRecv uint64
		for _, iface := range netMetrics.Interfaces {
			totalSent += iface.BytesSent
			totalRecv += iface.BytesRecv
		}
		assert.GreaterOrEqual(t, netMetrics.TotalBytesSent, totalSent)
		assert.GreaterOrEqual(t, netMetrics.TotalBytesRecv, totalRecv)
	})

	// Test temperature metrics collection (if supported)
	t.Run("Temperature Metrics Collection", func(t *testing.T) {
		tempMetrics, err := systemService.GetTemperatureMetrics(ctx)

		// Temperature might not be supported on all systems
		if err != nil {
			t.Skipf("Temperature monitoring not supported: %v", err)
			return
		}

		assert.NotNil(t, tempMetrics)

		// Verify temperature metrics structure
		for _, sensor := range tempMetrics.Sensors {
			assert.NotEmpty(t, sensor.Name)
			assert.GreaterOrEqual(t, sensor.Temperature, float64(-273.15)) // Absolute zero
			assert.LessOrEqual(t, sensor.Temperature, float64(200)) // Reasonable upper bound

			// Critical and max temperatures should be logical
			if sensor.Critical > 0 {
				assert.Greater(t, sensor.Critical, sensor.Max)
			}
			if sensor.Max > 0 {
				assert.Greater(t, sensor.Max, sensor.Temperature)
			}
		}

		// Hottest and coldest sensors should be consistent
		if len(tempMetrics.Sensors) > 0 {
			hottest := tempMetrics.GetHottestSensor()
			coldest := tempMetrics.GetColdestSensor()
			assert.NotNil(t, hottest)
			assert.NotNil(t, coldest)
			assert.GreaterOrEqual(t, hottest.Temperature, coldest.Temperature)
		}
	})
}

func TestSystemMetricsRealTimeUpdates(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()
	ctx := context.Background()

	// Test that metrics update over time
	t.Run("Real-time Updates", func(t *testing.T) {
		// Get initial metrics
		initialCPU, err := systemService.GetCPUMetrics(ctx)
		require.NoError(t, err)

		initialMem, err := systemService.GetMemoryMetrics(ctx)
		require.NoError(t, err)

		// Wait a short time
		time.Sleep(100 * time.Millisecond)

		// Get updated metrics
		updatedCPU, err := systemService.GetCPUMetrics(ctx)
		require.NoError(t, err)

		updatedMem, err := systemService.GetMemoryMetrics(ctx)
		require.NoError(t, err)

		// Verify timestamps are different (indicating fresh data)
		assert.True(t, updatedCPU.Timestamp.After(initialCPU.Timestamp))
		assert.True(t, updatedMem.Timestamp.After(initialMem.Timestamp))

		// CPU usage should be a float and reasonably bounded
		assert.GreaterOrEqual(t, updatedCPU.UsagePercent, 0.0)
		assert.LessOrEqual(t, updatedCPU.UsagePercent, 100.0)

		// Memory total should be the same, but usage might change
		assert.Equal(t, initialMem.Total, updatedMem.Total)
		assert.GreaterOrEqual(t, updatedMem.UsagePercent, 0.0)
		assert.LessOrEqual(t, updatedMem.UsagePercent, 100.0)
	})
}

func TestSystemMetricsErrorHandling(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()

	// Test with cancelled context
	t.Run("Cancelled Context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := systemService.GetCPUMetrics(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	// Test with timeout
	t.Run("Timeout Context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// This might timeout depending on system responsiveness
		_, err := systemService.GetProcessMetrics(ctx, 1000)
		if err != nil {
			assert.Contains(t, err.Error(), "context deadline exceeded")
		}
	})
}

func TestSystemMetricsPerformance(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()
	ctx := context.Background()

	// Test performance requirements
	t.Run("Performance Requirements", func(t *testing.T) {
		// CPU metrics should be fast (<10ms)
		start := time.Now()
		_, err := systemService.GetCPUMetrics(ctx)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 10*time.Millisecond, "CPU metrics collection should be under 10ms")

		// Memory metrics should be fast (<5ms)
		start = time.Now()
		_, err = systemService.GetMemoryMetrics(ctx)
		duration = time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 5*time.Millisecond, "Memory metrics collection should be under 5ms")

		// Process metrics (top 100) should be reasonable (<50ms)
		start = time.Now()
		_, err = systemService.GetProcessMetrics(ctx, 100)
		duration = time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 50*time.Millisecond, "Process metrics collection should be under 50ms")
	})
}

func TestSystemMetricsConsistency(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()
	ctx := context.Background()

	// Test data consistency across multiple calls
	t.Run("Data Consistency", func(t *testing.T) {
		// Get metrics multiple times
		var samples []models.SystemMetrics
		for i := 0; i < 5; i++ {
			metrics, err := systemService.GetAllMetrics(ctx)
			require.NoError(t, err)
			samples = append(samples, metrics)
			time.Sleep(10 * time.Millisecond)
		}

		// CPU core count should be consistent
		for i := 1; i < len(samples); i++ {
			assert.Equal(t, samples[0].CPU.Cores, samples[i].CPU.Cores, "CPU core count should be consistent")
			assert.Equal(t, samples[0].CPU.Model, samples[i].CPU.Model, "CPU model should be consistent")
			assert.Equal(t, samples[0].CPU.Vendor, samples[i].CPU.Vendor, "CPU vendor should be consistent")
		}

		// Memory total should be consistent
		for i := 1; i < len(samples); i++ {
			assert.Equal(t, samples[0].Memory.Total, samples[i].Memory.Total, "Total memory should be consistent")
		}

		// Disk space should be consistent
		for i := 1; i < len(samples); i++ {
			assert.Equal(t, len(samples[0].Disks.Disks), len(samples[i].Disks.Disks), "Disk count should be consistent")
			for j := 0; j < len(samples[0].Disks.Disks); j++ {
				assert.Equal(t, samples[0].Disks.Disks[j].Total, samples[i].Disks.Disks[j].Total,
					"Disk total space should be consistent for %s", samples[0].Disks.Disks[j].Device)
			}
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkCPUMetricsCollection(b *testing.B) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := systemService.GetCPUMetrics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryMetricsCollection(b *testing.B) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := systemService.GetMemoryMetrics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessMetricsCollection(b *testing.B) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := systemService.GetProcessMetrics(ctx, 100)
		if err != nil {
			b.Fatal(err)
		}
	}
}