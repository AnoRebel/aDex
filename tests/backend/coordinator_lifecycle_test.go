package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"aDex/internal/services/coordinator"
	"aDex/internal/events"
)

// TestCoordinator_FullLifecycle proves Initialize → Shutdown completes
// cleanly within the test timeout. Before the audio + event-bus shutdown
// fix this test would hang for 30+ seconds while the oto/v3 PCM writer
// goroutine and the event-bus delivery goroutines kept running past
// `go test`'s wait.
func TestCoordinator_FullLifecycle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init path differs on windows")
	}

	sc := coordinator.NewServiceCoordinator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sc.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if !sc.IsStarted() {
		t.Fatalf("coordinator did not start")
	}

	// Add a live subscription on the global bus so coordinator's Shutdown
	// has work to drain. We can't reach the coordinator's own bus pointer
	// directly (the accessor is intentionally absent — see service.go's
	// "Bus accessor intentionally omitted." comment) and that's fine:
	// coordinator initialises with `events.GetEventBus()`, the same
	// singleton.
	bus := events.GetEventBus()
	sub, err := bus.Subscribe(ctx, []string{"system.info.updated", "terminal.output"},
		func(_ context.Context, _ events.Event) error {
			return nil
		})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	_ = sub

	done := make(chan error, 1)
	go func() { done <- sc.Shutdown() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("Shutdown timed out — audio or event bus likely leaking goroutines")
	}

	// Idempotent.
	if err := sc.Shutdown(); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

// TestCoordinator_GetTerminalCWD_ViaCoordinator covers the binding the
// frontend will call (replacing the old HTTP /api/terminal/.../cwd) when
// the full coordinator is wired.
func TestCoordinator_GetTerminalCWD_ViaCoordinator(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init path differs on windows")
	}

	sc := coordinator.NewServiceCoordinator()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sc.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(func() {
		done := make(chan error, 1)
		go func() { done <- sc.Shutdown() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Errorf("cleanup Shutdown timed out")
		}
	})

	// Unknown id should error cleanly.
	_, err := sc.GetTerminalCWD("nope")
	if err == nil {
		t.Fatalf("expected error for unknown terminal id")
	}

	// CWDStats should be queryable even with no sessions.
	stats, err := sc.GetCWDStats()
	if err != nil {
		t.Fatalf("GetCWDStats: %v", err)
	}
	if stats.Platform != runtime.GOOS {
		t.Fatalf("Platform = %q, want %q", stats.Platform, runtime.GOOS)
	}
}
