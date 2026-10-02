package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolateConfigDir points the application's config and data directories at a
// fresh temporary directory for the duration of the test.
//
// Setting XDG_CONFIG_HOME alone is not enough: os.UserConfigDir honours it on
// Linux and the BSDs, but NOT on macOS, where it returns
// ~/Library/Application Support, or on Windows, where it reads APPDATA. A
// test that sets only XDG_CONFIG_HOME therefore escapes its sandbox on those
// platforms and reads the developer's real themes — which is exactly how
// TestSaveAndLoadUserTheme_RoundTrip failed on macOS CI while passing on
// Linux. Set all of them.
func isolateConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("HOME", dir)    // macOS: ~/Library/Application Support
	t.Setenv("APPDATA", dir) // Windows
	t.Setenv("LOCALAPPDATA", dir)
	return dir
}

// The theme id becomes a filename, so it must not be able to escape the
// themes directory.
func TestSaveUserTheme_RejectsPathTraversal(t *testing.T) {
	for _, bad := range []string{"../evil", "..", ".", "a/b", "/etc/passwd"} {
		if err := SaveUserTheme(bad, map[string]interface{}{"id": bad}); err == nil {
			t.Errorf("id %q was accepted; it must be rejected", bad)
		}
	}
}

func TestSaveUserTheme_RejectsEmptyID(t *testing.T) {
	if err := SaveUserTheme("", map[string]interface{}{}); err == nil {
		t.Error("an empty id must be rejected")
	}
}

func TestDeleteUserTheme_RejectsPathTraversal(t *testing.T) {
	for _, bad := range []string{"../evil", "a/b", ""} {
		if err := DeleteUserTheme(bad); err == nil {
			t.Errorf("id %q was accepted for deletion; it must be rejected", bad)
		}
	}
}

// A missing directory is the normal case for a user with no themes.
func TestLoadUserThemes_MissingDirectoryIsNotAnError(t *testing.T) {
	isolateConfigDir(t)
	themes, err := LoadUserThemes()
	if err != nil {
		t.Fatalf("expected no error for a missing directory, got %v", err)
	}
	if len(themes) != 0 {
		t.Errorf("expected no themes, got %d", len(themes))
	}
}

// One malformed file must not hide the valid ones.
func TestLoadUserThemes_SkipsMalformedFiles(t *testing.T) {
	isolateConfigDir(t)
	dir := UserThemesPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "good.json"), []byte(`{"id":"good"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{not json`), 0o600); err != nil {
		t.Fatal(err)
	}

	themes, err := LoadUserThemes()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected the one valid theme, got %d", len(themes))
	}
	m := themes[0].(map[string]interface{})
	if m["id"] != "good" {
		t.Errorf("wrong theme survived: %v", m["id"])
	}
	if m["source"] != "user" {
		t.Errorf("user themes must be marked as such, got %v", m["source"])
	}
}

func TestSaveAndLoadUserTheme_RoundTrip(t *testing.T) {
	isolateConfigDir(t)
	if err := SaveUserTheme("mine", map[string]interface{}{"id": "mine", "displayName": "Mine"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	themes, err := LoadUserThemes()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(themes) != 1 || !strings.Contains(themes[0].(map[string]interface{})["displayName"].(string), "Mine") {
		t.Errorf("theme did not round-trip: %v", themes)
	}
}
