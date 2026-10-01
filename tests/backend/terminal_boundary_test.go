package tests

import (
	"runtime"
	"testing"
	"time"
)

// Section 5.A boundary tests — exercise every ServiceCoordinator method
// the frontend's terminal store calls, asserting both success and error
// paths. Skipped on Windows because PTY allocation differs there and the
// internal/services/terminal Windows path covers it separately.

func TestTerminal_CreateWriteResizeClose_Roundtrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init path differs on windows")
	}
	sc, _ := NewTestCoordinator(t)
	if !sc.IsStarted() {
		t.Fatalf("coordinator did not start")
	}

	term, err := sc.CreateTerminal(80, 24)
	if err != nil {
		t.Skipf("CreateTerminal not available in this env: %v", err)
	}
	if term.ID == "" {
		t.Fatalf("Terminal.ID empty")
	}
	if term.Width != 80 || term.Height != 24 {
		t.Fatalf("dimensions = %dx%d, want 80x24", term.Width, term.Height)
	}

	// Echo something through the PTY. We don't read it back here (that's
	// covered by the event-emit test); we just prove WriteToTerminal
	// accepts a string (frontend-facing wire type — Wails marshals JS
	// strings transparently for Go string params, while []byte would
	// require base64 encoding).
	if err := sc.WriteToTerminal(term.ID, "echo hi\n"); err != nil {
		t.Fatalf("WriteToTerminal: %v", err)
	}

	if err := sc.ResizeTerminal(term.ID, 100, 30); err != nil {
		t.Fatalf("ResizeTerminal: %v", err)
	}
	info, err := sc.GetTerminalInfo(term.ID)
	if err != nil {
		t.Fatalf("GetTerminalInfo: %v", err)
	}
	if info.Width != 100 || info.Height != 30 {
		t.Fatalf("after resize, dimensions = %dx%d, want 100x30", info.Width, info.Height)
	}

	ids, err := sc.ListTerminals()
	if err != nil {
		t.Fatalf("ListTerminals: %v", err)
	}
	if len(ids) != 1 || ids[0] != term.ID {
		t.Fatalf("ListTerminals = %v, want [%s]", ids, term.ID)
	}

	if err := sc.CloseTerminal(term.ID); err != nil {
		t.Fatalf("CloseTerminal: %v", err)
	}

	// After close, the terminal should be gone. Allow a brief moment for
	// the reader goroutine to exit and remove it from the map.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		ids, _ := sc.ListTerminals()
		if len(ids) == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	ids, _ = sc.ListTerminals()
	if len(ids) != 0 {
		t.Fatalf("after close, ListTerminals = %v, want []", ids)
	}
}

func TestTerminal_ErrorPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init path differs on windows")
	}
	sc, _ := NewTestCoordinator(t)

	// Operations on an unknown id surface errors rather than panicking.
	for _, tc := range []struct {
		name string
		op   func() error
	}{
		{"WriteToTerminal", func() error { return sc.WriteToTerminal("nope", "x") }},
		{"ResizeTerminal", func() error { return sc.ResizeTerminal("nope", 80, 24) }},
		{"CloseTerminal", func() error { return sc.CloseTerminal("nope") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.op(); err == nil {
				t.Fatalf("expected error for unknown terminal id")
			}
		})
	}

	t.Run("GetTerminalInfo", func(t *testing.T) {
		_, err := sc.GetTerminalInfo("nope")
		if err == nil {
			t.Fatalf("expected error for unknown terminal id")
		}
	})
}

// TestTerminal_DoubleCloseDoesNotPanic regression-tests the
// close-of-closed-channel panic that was fixed in section 1 by guarding
// terminal.Done with sync.Once. Without the fix, this test would panic
// during the second CloseTerminal call.
func TestTerminal_DoubleCloseDoesNotPanic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init path differs on windows")
	}
	sc, _ := NewTestCoordinator(t)

	term, err := sc.CreateTerminal(80, 24)
	if err != nil {
		t.Skipf("CreateTerminal not available in this env: %v", err)
	}

	if err := sc.CloseTerminal(term.ID); err != nil {
		t.Fatalf("first CloseTerminal: %v", err)
	}
	// The second call should fail cleanly (terminal not found) without
	// crashing the process.
	if err := sc.CloseTerminal(term.ID); err == nil {
		t.Fatalf("second CloseTerminal should return not-found error")
	}
}
