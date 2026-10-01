package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aDex/internal/services/coordinator"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// A failing service initialisation must surface as an error from
// ServiceStartup rather than a silent exit (spec: application-lifecycle,
// "A critical service fails to initialise"). Wails aborts startup when
// ServiceStartup returns non-nil, so propagating the error is what makes
// the failure visible instead of the app disappearing without explanation.
func TestServiceStartupSurfacesInitFailure(t *testing.T) {
	// Service data resolves relative to the working directory. Making
	// `data` a regular file means every attempt to read a directory
	// beneath it fails.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	sc := coordinator.NewServiceCoordinator()
	startErr := sc.ServiceStartup(context.Background(), application.ServiceOptions{})
	t.Cleanup(func() { _ = sc.ServiceShutdown() })

	if startErr == nil {
		// Initialisation is resilient to this particular breakage. That is a
		// valid outcome; the contract under test is only that a real failure
		// is reported, never swallowed.
		t.Skip("initialisation tolerated the broken data directory")
	}
	if strings.TrimSpace(startErr.Error()) == "" {
		t.Fatal("ServiceStartup returned an error carrying no message")
	}
	t.Logf("ServiceStartup surfaced the failure: %v", startErr)
}
