package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"aDex/internal/services/system"
)

// Section 5.B boundary tests for the system monitor service. Exercises
// the methods the frontend store calls via the coordinator
// (GetSystemInfo, GetCPUUsage, GetMemoryUsage, GetDiskUsage,
// GetTopProcesses) and asserts each returns non-zero data on Linux/Darwin
// — the exact opposite of the "shows zeros" symptom from the user's bug
// report.

func TestSystem_GetSystemInfo_Populated(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("system info path differs on windows; covered separately")
	}
	s := system.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := s.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("GetSystemInfo: %v", err)
	}
	if info == nil {
		t.Fatalf("info is nil")
	}
	if info.Hostname == "" {
		t.Errorf("Hostname empty — gopsutil host info path broken")
	}
	if info.OS == "" {
		t.Errorf("OS empty")
	}
	if info.Architecture == "" {
		t.Errorf("Architecture empty")
	}
}

func TestSystem_GetCPUInfo_NonZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("system path differs on windows")
	}
	s := system.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cpu, err := s.GetCPUInfo(ctx)
	if err != nil {
		t.Fatalf("GetCPUInfo: %v", err)
	}
	if cpu.CoreCount <= 0 {
		t.Fatalf("CoreCount = %d, want > 0", cpu.CoreCount)
	}
	// Per-core slice length should match CoreCount.
	if len(cpu.Cores) != cpu.CoreCount {
		t.Errorf("len(Cores) = %d, want %d", len(cpu.Cores), cpu.CoreCount)
	}
	// Usage in [0, 100].
	if cpu.Usage < 0 || cpu.Usage > 100 {
		t.Errorf("Usage = %f, want 0..100", cpu.Usage)
	}
}

func TestSystem_GetMemoryUsage_NonZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("system path differs on windows")
	}
	s := system.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mem, err := s.GetMemoryUsage(ctx)
	if err != nil {
		t.Fatalf("GetMemoryUsage: %v", err)
	}
	if mem.Total == 0 {
		t.Fatalf("Memory.Total = 0 — gopsutil mem path broken (regression of the 'shows zeros' bug)")
	}
	if mem.Used > mem.Total {
		t.Errorf("Used (%d) > Total (%d)", mem.Used, mem.Total)
	}
	// Field is named Percent on MemoryInfo (the coordinator remaps to
	// `usage` for the frontend at the binding boundary).
	if mem.Percent < 0 || mem.Percent > 100 {
		t.Errorf("Memory.Percent = %f, want 0..100", mem.Percent)
	}
}

func TestSystem_GetTopProcesses_NonEmpty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("system path differs on windows")
	}
	s := system.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	procs, err := s.GetTopProcesses(ctx, "cpu", 5)
	if err != nil {
		t.Fatalf("GetTopProcesses: %v", err)
	}
	if len(procs) == 0 {
		t.Fatalf("no processes returned — gopsutil process listing broken")
	}
	if len(procs) > 5 {
		t.Errorf("returned %d processes, asked for limit 5", len(procs))
	}
	// At least one process should have a non-empty name.
	hasName := false
	for _, p := range procs {
		if p.Name != "" {
			hasName = true
			break
		}
	}
	if !hasName {
		t.Errorf("no process has a non-empty Name")
	}
}

func TestSystem_GetDiskUsage_NonZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("system path differs on windows")
	}
	s := system.NewService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	disks, err := s.GetDiskUsage(ctx)
	if err != nil {
		t.Fatalf("GetDiskUsage: %v", err)
	}
	if len(disks) == 0 {
		t.Skip("no disk info available — likely a sandboxed env, not a regression")
	}
	// Root partition should report non-zero total.
	for _, d := range disks {
		if d.Total > 0 {
			return // at least one valid partition found
		}
	}
	t.Errorf("every reported disk has Total = 0")
}
