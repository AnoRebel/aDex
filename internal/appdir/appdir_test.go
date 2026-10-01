package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

// An existing install must keep its settings: the pre-rename directory is
// moved, not ignored, or the user silently starts fresh.
func TestMigrate_MovesLegacyDirectory(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, legacyName)
	current := filepath.Join(root, Name)

	if err := os.MkdirAll(old, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(old, "settings.json")
	if err := os.WriteFile(marker, []byte(`{"kept":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	migrate(old, current)

	data, err := os.ReadFile(filepath.Join(current, "settings.json"))
	if err != nil {
		t.Fatalf("settings did not survive the migration: %v", err)
	}
	if string(data) != `{"kept":true}` {
		t.Errorf("contents changed: %s", data)
	}
}

// Migration must never clobber a directory that is already in use.
func TestMigrate_DoesNotOverwriteExisting(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, legacyName)
	current := filepath.Join(root, Name)

	if err := os.MkdirAll(old, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(current, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "f"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(current, "f"), []byte("current"), 0o600); err != nil {
		t.Fatal(err)
	}

	migrate(old, current)

	data, _ := os.ReadFile(filepath.Join(current, "f"))
	if string(data) != "current" {
		t.Errorf("migration overwrote the in-use directory: got %q", data)
	}
}

// A fresh install has nothing to migrate and must not fail.
func TestMigrate_NoLegacyDirectoryIsFine(t *testing.T) {
	root := t.TempDir()
	migrate(filepath.Join(root, legacyName), filepath.Join(root, Name))
}
