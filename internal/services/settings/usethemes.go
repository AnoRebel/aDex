package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"aDex-UI/internal/appdir"
)

// User-authored themes.
//
// The 21 bundled themes are compiled into the binary, so users cannot add to
// them by dropping a file in. This reads a `themes/` directory under the
// config dir and hands the entries to the frontend, which merges them with the
// bundled catalogue — the same shape as custom layouts, so a user who has
// learned one already understands the other.

const userThemesDirName = "themes"

// UserThemesPath returns the directory user themes are read from, whether or
// not it exists. Shown in Settings so users know where to put a file.
func UserThemesPath() string {
	return filepath.Join(appdir.Config(), userThemesDirName)
}

// LoadUserThemes reads every *.json in the user themes directory.
//
// Returned as generic JSON: the frontend owns the theme schema and validates
// against it, so parsing into a Go struct here would duplicate that schema in
// a second place and let the two drift.
//
// A missing directory is the normal case and is not an error. One malformed
// file is skipped rather than failing the whole load, so a single bad theme
// cannot hide every good one.
func LoadUserThemes() ([]interface{}, error) {
	dir := UserThemesPath()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", dir, err)
	}

	// Sorted so the list is stable between runs rather than filesystem order.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	themes := make([]interface{}, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var theme map[string]interface{}
		if err := json.Unmarshal(data, &theme); err != nil {
			continue // Skip this one; a bad file must not hide the rest.
		}
		// Mark the origin so the UI can label it and refuse to overwrite a
		// bundled theme with a user one of the same id.
		theme["source"] = "user"
		themes = append(themes, theme)
	}
	return themes, nil
}

// SaveUserTheme writes a theme to the user themes directory, creating it if
// needed. The id becomes the filename, so it must be filesystem-safe.
func SaveUserTheme(id string, theme interface{}) error {
	if id == "" {
		return fmt.Errorf("theme id is required")
	}
	// Reject anything that could escape the directory.
	if filepath.Base(id) != id || id == "." || id == ".." {
		return fmt.Errorf("invalid theme id %q", id)
	}

	dir := UserThemesPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode theme: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, id+".json"), data, 0o600)
}

// DeleteUserTheme removes a user theme. Bundled themes are compiled in and
// cannot be removed this way.
func DeleteUserTheme(id string) error {
	if id == "" || filepath.Base(id) != id {
		return fmt.Errorf("invalid theme id %q", id)
	}
	err := os.Remove(filepath.Join(UserThemesPath(), id+".json"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
