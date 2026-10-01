package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aDex-UI/internal/services/coordinator"
)

// NewTestCoordinator builds a fresh ServiceCoordinator rooted at a temporary
// data directory, initialized via the public Initialize(ctx) entrypoint.
//
// Use this in boundary tests so each test gets a clean state and t.Cleanup()
// shuts the coordinator down. Returns the coordinator and the temp dir path
// in case the test wants to seed files into it.
func NewTestCoordinator(t *testing.T) (*coordinator.ServiceCoordinator, string) {
	t.Helper()
	// Note: previous versions of this helper set ADEX_DISABLE_AUDIO=1 to
	// dodge a hang in the oto/v3 PCM writer on Linux ALSA at process exit.
	// That bypass is no longer needed — audio.Service.Shutdown() now
	// suspends the oto context and the coordinator's Shutdown() drains
	// the event bus. Tests can run the full coordinator lifecycle.

	dir := t.TempDir()
	// Some services hard-code "./data" relative paths; cd into the temp dir
	// so they don't pollute the repo when tests run.
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}

	sc := coordinator.NewServiceCoordinator()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := sc.Initialize(ctx); err != nil {
		t.Fatalf("coordinator initialize: %v", err)
	}
	t.Cleanup(func() {
		// Bound the shutdown so a single hang doesn't take the whole
		// suite hostage — surfaces as a test error instead of a 10-minute
		// `go test` timeout.
		done := make(chan error, 1)
		go func() { done <- sc.Shutdown() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Errorf("coordinator Shutdown timed out after 5s")
		}
	})

	return sc, dir
}

// AssertNoErr fails the test if err is non-nil with a labelled message.
func AssertNoErr(t *testing.T, label string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", label, err)
	}
}
