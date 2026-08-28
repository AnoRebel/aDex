package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
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
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
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
