package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"aDex/internal/services/coordinator"
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

	dir := IsolateAppDirs(t)

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

// TestMain keeps every test in this package away from the real config and
// data directories: services resolve them from the environment, so a test
// that forgets IsolateAppDirs would otherwise write into the user's own.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "adex-tests-")
	if err != nil {
		panic(err)
	}
	for _, k := range []string{"HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "APPDATA", "LOCALAPPDATA"} {
		_ = os.Setenv(k, dir)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// IsolateAppDirs points the app's config and data directories at a fresh
// temp dir for the test, so services never touch the real ones. Returns it.
func IsolateAppDirs(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, k := range []string{"HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "APPDATA", "LOCALAPPDATA"} {
		t.Setenv(k, dir)
	}
	return dir
}

// AssertNoErr fails the test if err is non-nil with a labelled message.
func AssertNoErr(t *testing.T, label string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", label, err)
	}
}
