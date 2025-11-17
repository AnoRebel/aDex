package backend

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"aDex-UI/internal/services/system"
)

func TestSystemMonitoringPerformanceRequirements(t *testing.T) {
	// Create system service
	systemService := system.NewSystemService()
	ctx := context.Background()

	// Performance targets from specification
	const (
		maxCPUUsage    = 3.0  // <3% CPU usage during monitoring
		maxMemoryUsage = 200 * 1024 * 1024 // <200MB memory
		minUpdateTime  = 1000 * time.Millisecond // 500-1000ms intervals
		maxUpdateTime  = 2000 * time.Millisecond // 500-1000ms intervals
	)

	t.Run("CPU Usage Under Load", func(t *testing.T) {
		// Monitor CPU usage while collecting metrics
		startTime := time.Now()
		var measurements []float64

		// Collect metrics for 10 seconds while measuring CPU usage
		endTime := startTime.Add(10 * time.Second)
		for time.Now().Before(endTime) {
			iterStart := time.Now()

			// Collect all metrics
			_, err := systemService.GetCPUMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetMemoryMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetProcessMetrics(ctx, 100)
			require.NoError(t, err)

			_, err = systemService.GetDiskMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetNetworkMetrics(ctx)
			require.NoError(t, err)

			iterDuration := time.Since(iterStart)
			measurements = append(measurements, float64(iterDuration.Nanoseconds()))

			// Small delay to simulate real monitoring interval
			time.Sleep(500 * time.Millisecond)
		}

		// Calculate average time per collection cycle
		var total time.Duration
		for _, dur := range measurements {
			total += time.Duration(int64(dur))
		}
		avgDuration := total / time.Duration(len(measurements))

		// Verify collection time is reasonable (should be much less than 100ms)
		assert.Less(t, avgDuration, 50*time.Millisecond, "Average metrics collection should be under 50ms")

		// Verify we can sustain the monitoring frequency
		assert.Less(t, avgDuration, minUpdateTime/10, "Collection should be fast enough for 1-second intervals")
	})

	t.Run("Memory Usage Stability", func(t *testing.T) {
		// Get initial memory usage
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Run monitoring for extended period
		startTime := time.Now()
		collectionCount := 0

		for time.Since(startTime) < 30*time.Second {
			// Collect all metrics
			_, err := systemService.GetCPUMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetMemoryMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetProcessMetrics(ctx, 50)
			require.NoError(t, err)

			collectionCount++

			// Sleep to simulate monitoring interval
			time.Sleep(1 * time.Second)
		}

		// Check final memory usage
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := m2.Alloc - m1.Alloc
		memoryPerCollection := float64(memoryIncrease) / float64(collectionCount)

		// Verify memory usage is reasonable
		assert.Less(t, memoryIncrease, maxMemoryUsage, "Total memory increase should be under 200MB")
		assert.Less(t, memoryPerCollection, 1024, "Memory per collection should be under 1KB")

		// Verify no memory leaks (memory growth should be minimal)
		assert.Less(t, memoryIncrease, uint64(50*1024*1024), "Memory growth should be under 50MB over 30 seconds")
	})

	t.Run("Concurrent Collection Performance", func(t *testing.T) {
		// Test multiple concurrent collectors
		numCollectors := 5
		collectionsPerCollector := 20

		results := make(chan time.Duration, numCollectors*collectionsPerCollector)

		startTime := time.Now()

		// Start concurrent collectors
		for i := 0; i < numCollectors; i++ {
			go func() {
				for j := 0; j < collectionsPerCollector; j++ {
					collectStart := time.Now()

					// Collect metrics
					_, err := systemService.GetCPUMetrics(ctx)
					if err != nil {
						t.Errorf("CPU metrics collection failed: %v", err)
						continue
					}

					_, err = systemService.GetMemoryMetrics(ctx)
					if err != nil {
						t.Errorf("Memory metrics collection failed: %v", err)
						continue
					}

					_, err = systemService.GetProcessMetrics(ctx, 50)
					if err != nil {
						t.Errorf("Process metrics collection failed: %v", err)
						continue
					}

					results <- time.Since(collectStart)
				}
			}()
		}

		// Wait for all collections to complete
		for i := 0; i < numCollectors*collectionsPerCollector; i++ {
			select {
			case duration := <-results:
				// Verify individual collection times
				assert.Less(t, duration, 100*time.Millisecond, "Individual collection should be under 100ms")
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent collection test timed out")
			}
		}

		totalTime := time.Since(startTime)
		totalCollections := numCollectors * collectionsPerCollector
		avgTime := totalTime / time.Duration(totalCollections)

		// Verify concurrent performance
		assert.Less(t, avgTime, 200*time.Millisecond, "Average concurrent collection should be under 200ms")
	})

	t.Run("Update Interval Compliance", func(t *testing.T) {
		// Test that we can maintain required update intervals
		requiredInterval := 1000 * time.Millisecond
		tolerance := 100 * time.Millisecond

		// Monitor for multiple intervals
		numIntervals := 10
		startTime := time.Now()
		var intervalDurations []time.Duration

		for i := 0; i < numIntervals; i++ {
			intervalStart := time.Now()

			// Collect all metrics
			_, err := systemService.GetAllMetrics(ctx)
			require.NoError(t, err)

			// Wait for next interval
			time.Sleep(requiredInterval)

			intervalDuration := time.Since(intervalStart)
			intervalDurations = append(intervalDurations, intervalDuration)
		}

		// Verify intervals are within tolerance
		for i, duration := range intervalDurations {
			assert.GreaterOrEqual(t, duration, requiredInterval-tolerance,
				"Interval %d should not be too short", i)
			assert.LessOrEqual(t, duration, requiredInterval+tolerance,
				"Interval %d should not be too long", i)
		}

		// Verify average interval accuracy
		var totalInterval time.Duration
		for _, duration := range intervalDurations {
			totalInterval += duration
		}
		avgInterval := totalInterval / time.Duration(len(intervalDurations))

		assert.InDelta(t, requiredInterval, avgInterval, float64(tolerance),
			"Average interval should be close to required interval")
	})

	t.Run("System Resource Impact", func(t *testing.T) {
		// Measure system impact during monitoring
		startTime := time.Now()
		monitoringDuration := 10 * time.Second

		// Get baseline system metrics
		baselineCPU, err := systemService.GetCPUMetrics(ctx)
		require.NoError(t, err)

		baselineMem, err := systemService.GetMemoryMetrics(ctx)
		require.NoError(t, err)

		// Run continuous monitoring
		collectionCount := 0
		for time.Since(startTime) < monitoringDuration {
			_, err := systemService.GetCPUMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetMemoryMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetProcessMetrics(ctx, 100)
			require.NoError(t, err)

			collectionCount++
			time.Sleep(500 * time.Millisecond)
		}

		// Get post-monitoring system metrics
		postCPU, err := systemService.GetCPUMetrics(ctx)
		require.NoError(t, err)

		postMem, err := systemService.GetMemoryMetrics(ctx)
		require.NoError(t, err)

		// Calculate impact
		cpuImpact := postCPU.UsagePercent - baselineCPU.UsagePercent
		memoryImpact := postMem.Used - baselineMem.Used

		// Verify impact is minimal
		assert.Less(t, cpuImpact, maxCPUUsage, "CPU usage impact should be under 3%")
		assert.Less(t, memoryImpact, uint64(100*1024*1024), "Memory impact should be under 100MB")

		// Verify collection frequency was maintained
		expectedCollections := int(monitoringDuration / (500 * time.Millisecond))
		assert.GreaterOrEqual(t, collectionCount, expectedCollections/2,
			"Should maintain reasonable collection frequency")
	})
}

func TestSystemMonitoringScalability(t *testing.T) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	t.Run("Large Process List Performance", func(t *testing.T) {
		// Test performance with large process lists
		processCounts := []int{100, 500, 1000, 2000}

		for _, count := range processCounts {
			t.Run(fmt.Sprintf("ProcessCount_%d", count), func(t *testing.T) {
				start := time.Now()

				_, err := systemService.GetProcessMetrics(ctx, count)
				require.NoError(t, err)

				duration := time.Since(start)

				// Performance should scale reasonably
				maxDuration := time.Duration(count) * time.Microsecond * 100 // 100μs per process
				assert.Less(t, duration, maxDuration,
					"Process collection with %d processes should be under %v", count, maxDuration)

				// Even with 2000 processes, should be under 200ms
				assert.Less(t, duration, 200*time.Millisecond,
					"Process collection should be under 200ms even for large lists")
			})
		}
	})

	t.Run("Frequency Stress Test", func(t *testing.T) {
		// Test high-frequency monitoring
		intervals := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			500 * time.Millisecond,
		}

		for _, interval := range intervals {
			t.Run(fmt.Sprintf("Interval_%v", interval), func(t *testing.T) {
				startTime := time.Now()
				testDuration := 5 * time.Second
				collectionCount := 0

				for time.Since(startTime) < testDuration {
					collectStart := time.Now()

					_, err := systemService.GetCPUMetrics(ctx)
					require.NoError(t, err)

					collectionCount++
					time.Sleep(interval)
				}

				// Calculate actual collection rate
				actualInterval := testDuration / time.Duration(collectionCount)

				// Should be able to maintain the requested interval
				assert.LessOrEqual(t, actualInterval, interval*2,
					"Should be able to maintain interval %v (actual: %v)", interval, actualInterval)
			})
		}
	})

	t.Run("Memory Scalability", func(t *testing.T) {
		// Test memory usage scales properly with data size
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Collect metrics with increasing process counts
		processCounts := []int{100, 500, 1000}
		for _, count := range processCounts {
			_, err := systemService.GetProcessMetrics(ctx, count)
			require.NoError(t, err)
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := m2.Alloc - m1.Alloc

		// Memory increase should be proportional to data size
		expectedMaxIncrease := uint64(10 * 1024 * 1024) // 10MB max for all tests
		assert.Less(t, memoryIncrease, expectedMaxIncrease,
			"Memory increase should be under 10MB for scalability test")
	})
}

func TestSystemMonitoringReliability(t *testing.T) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	t.Run("Long-running Stability", func(t *testing.T) {
		// Test stability over extended period
		testDuration := 60 * time.Second
		interval := 1 * time.Second

		startTime := time.Now()
		successCount := 0
		errorCount := 0
		var errors []error

		for time.Since(startTime) < testDuration {
			_, err := systemService.GetCPUMetrics(ctx)
			if err != nil {
				errorCount++
				errors = append(errors, err)
			} else {
				successCount++
			}

			_, err = systemService.GetMemoryMetrics(ctx)
			if err != nil {
				errorCount++
				errors = append(errors, err)
			}

			time.Sleep(interval)
		}

		totalCollections := successCount + errorCount
		successRate := float64(successCount) / float64(totalCollections) * 100

		// Should have very high success rate
		assert.Greater(t, successRate, 99.0, "Success rate should be over 99%")
		assert.Less(t, errorCount, 5, "Should have fewer than 5 errors in 60 seconds")

		if errorCount > 0 {
			t.Logf("Errors encountered: %v", errors)
		}
	})

	t.Run("Resource Cleanup", func(t *testing.T) {
		// Test that resources are properly cleaned up
		var m1, m2, m3 runtime.MemStats

		// Baseline
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Perform many collections
		for i := 0; i < 1000; i++ {
			_, err := systemService.GetCPUMetrics(ctx)
			require.NoError(t, err)

			_, err = systemService.GetMemoryMetrics(ctx)
			require.NoError(t, err)

			if i%100 == 0 {
				runtime.GC() // Periodic GC
			}
		}

		// After collections
		runtime.GC()
		runtime.ReadMemStats(&m2)

		// Force cleanup and wait
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
		runtime.GC()
		runtime.ReadMemStats(&m3)

		// Memory should be cleaned up
		memoryGrowth := m2.Alloc - m1.Alloc
		memoryAfterCleanup := m3.Alloc - m1.Alloc

		assert.Less(t, memoryAfterCleanup, memoryGrowth,
			"Memory should be cleaned up after GC")
		assert.Less(t, memoryAfterCleanup, uint64(5*1024*1024),
			"Final memory usage should be close to baseline")
	})
}

// Benchmark tests for performance regression detection
func BenchmarkSystemMetricsCollection(b *testing.B) {
	systemService := system.NewSystemService()
	ctx := context.Background()

	b.Run("AllMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := systemService.GetAllMetrics(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CPUMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := systemService.GetCPUMetrics(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MemoryMetrics", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := systemService.GetMemoryMetrics(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ProcessMetrics_100", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := systemService.GetProcessMetrics(ctx, 100)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ProcessMetrics_1000", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := systemService.GetProcessMetrics(ctx, 1000)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}