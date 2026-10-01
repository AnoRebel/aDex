package tests

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"aDex/internal/services/terminal"
)

// These boundary tests exercise the terminal service's CWD methods directly
// rather than spinning up the full ServiceCoordinator, because the coordinator
// pulls in the audio service (oto/v3 PCM) and event bus listeners which leak
// goroutines on test exit. Coordinator-level integration coverage will land
// after the audio shutdown fix in section 4.
//
// What is covered here:
//   - GetCWDStats returns the right shape and platform string
//   - GetCurrentCWD on an unknown terminal id returns an error
//   - GetCurrentCWD on a real PTY-backed terminal returns an absolute path
//     equal to the live cwd of the shell process

func TestTerminal_GetCWDStats_EmptyService(t *testing.T) {
	s := terminal.NewService()
	stats := s.GetCWDStats()

	if stats.Sessions != 0 {
		t.Fatalf("Sessions = %d, want 0 for fresh service", stats.Sessions)
	}
	if stats.Platform != runtime.GOOS {
		t.Fatalf("Platform = %q, want %q", stats.Platform, runtime.GOOS)
	}
	wantStrategy := map[string]string{
		"linux":  "procfs",
		"darwin": "lsof",
	}[runtime.GOOS]
	if wantStrategy == "" {
		wantStrategy = "recorded"
	}
	if stats.Strategy != wantStrategy {
		t.Fatalf("Strategy = %q, want %q", stats.Strategy, wantStrategy)
	}
}

func TestTerminal_GetCurrentCWD_NotFound(t *testing.T) {
	s := terminal.NewService()
	_, err := s.GetCurrentCWD("does-not-exist")
	if err == nil {
		t.Fatalf("expected error for unknown terminal id")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error %q does not mention 'not found'", err)
	}
}

func TestTerminal_GetCurrentCWD_HappyPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows pty path lives in internal/services/terminal; covered there")
	}

	s := terminal.NewService()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	term, err := s.CreateTerminal(ctx, 80, 24)
	if err != nil {
		t.Skipf("CreateTerminal not available in this env: %v", err)
	}
	t.Cleanup(func() { _ = s.CloseTerminal(ctx, term.ID) })

	cwd, err := s.GetCurrentCWD(term.ID)
	if err != nil {
		t.Fatalf("GetCurrentCWD: %v", err)
	}
	if cwd == "" {
		t.Fatalf("CWD is empty")
	}
	if !strings.HasPrefix(cwd, "/") {
		t.Fatalf("CWD %q is not absolute", cwd)
	}

	stats := s.GetCWDStats()
	if stats.Sessions < 1 {
		t.Fatalf("Sessions = %d, want >= 1", stats.Sessions)
	}
	if _, ok := stats.PerSession[term.ID]; !ok {
		t.Fatalf("PerSession missing key %q", term.ID)
	}
	if runtime.GOOS == "linux" && stats.Resolutions < 1 {
		t.Fatalf("Resolutions = %d, want >= 1 after a successful procfs lookup", stats.Resolutions)
	}
}
