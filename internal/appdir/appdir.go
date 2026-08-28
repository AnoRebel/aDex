// Package appdir resolves the application's config and data directories.
//
// The name lived as a hardcoded "aDex-UI" string in six separate packages,
// so renaming meant editing every one and risking a mismatch where one
// service reads a different directory from another. It is defined once here.
package appdir

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Name is the directory name under the user's config and data roots.
const Name = "aDex"

// legacyName is the pre-rename directory. Its contents are migrated once, so
// an existing install keeps its settings, themes and lock passphrase rather
// than silently starting fresh.
const legacyName = "aDex-UI"

var migrateOnce sync.Once

// Config returns the application config directory, creating it if needed.
func Config() string {
	root, err := os.UserConfigDir()
	if err != nil {
		return Name
	}
	dir := filepath.Join(root, Name)
	migrateOnce.Do(func() { migrate(filepath.Join(root, legacyName), dir) })
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// Data returns the application data directory (themes and other bulk assets).
func Data() string {
	switch runtime.GOOS {
	case "windows":
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return filepath.Join(v, Name)
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", Name)
		}
	}
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return filepath.Join(v, Name)
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", Name)
	}
	return Name
}

// migrate moves a pre-rename directory to the new location, once.
//
// Deliberately conservative: it only acts when the old directory exists and
// the new one does not, so it can never overwrite current settings, and a
// failure is silent — an unmigratable directory means starting with defaults,
// which is far better than refusing to launch.
func migrate(old, current string) {
	if old == current {
		return
	}
	if _, err := os.Stat(current); err == nil {
		return // Already migrated, or a fresh install.
	}
	if _, err := os.Stat(old); err != nil {
		return // Nothing to migrate.
	}
	if err := os.Rename(old, current); err != nil {
		// A rename can fail across filesystems; leave the old directory in
		// place rather than half-copying it.
		return
	}
}
