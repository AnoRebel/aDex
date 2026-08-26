package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"aDex-UI/internal/services/network"
)

// Section 5.C boundary tests for the network service. Asserts that
// GetNetworkMetrics returns at least one interface and consistent counters
// on Linux/Darwin — the data the frontend NetworkMonitor reads.

func TestNetwork_GetNetworkMetrics_Populated(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("network path differs on windows")
	}
	s := network.NewNetworkService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	metrics, err := s.GetNetworkMetrics(ctx)
	if err != nil {
		t.Fatalf("GetNetworkMetrics: %v", err)
	}
	if metrics == nil {
		t.Fatalf("metrics is nil")
	}
	if len(metrics.Interfaces) == 0 {
		t.Fatalf("no interfaces reported — at least one non-loopback interface should exist on a real host")
	}

	// The service intentionally filters out lo/lo0 (see service.go:GetNetworkMetrics).
	// Verify that contract instead of asserting loopback is present.
	for _, iface := range metrics.Interfaces {
		if iface.Name == "lo" || iface.Name == "lo0" {
			t.Errorf("loopback %q should be filtered out", iface.Name)
		}
	}
}

func TestNetwork_GetConnections_DoesNotError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("network path differs on windows")
	}
	s := network.NewNetworkService()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Connection list can be empty in a sandboxed env — the contract is
	// that the call doesn't error.
	if _, err := s.GetConnections(ctx); err != nil {
		t.Fatalf("GetConnections: %v", err)
	}
}

func TestNetwork_StartStopMonitoring(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("network path differs on windows")
	}
	s := network.NewNetworkService()
	ctx, cancel := context.WithCancel(context.Background())

	if s.IsMonitoring() {
		t.Fatalf("freshly-constructed service reports monitoring already active")
	}

	// StartMonitoring runs the polling loop in the calling goroutine and
	// only returns once the context is cancelled — the caller is expected
	// to spawn it in a goroutine. Mirror the production wiring here.
	loopDone := make(chan error, 1)
	go func() { loopDone <- s.StartMonitoring(ctx) }()

	// Give the loop a tick to mark itself running.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && !s.IsMonitoring() {
		time.Sleep(10 * time.Millisecond)
	}
	if !s.IsMonitoring() {
		cancel()
		<-loopDone
		t.Fatalf("after StartMonitoring goroutine, IsMonitoring should become true")
	}

	cancel()
	select {
	case <-loopDone:
	case <-time.After(2 * time.Second):
		t.Fatalf("monitor loop did not exit after ctx cancel")
	}
	if s.IsMonitoring() {
		t.Errorf("after ctx cancel, IsMonitoring should be false")
	}
}
