package settings

import (
	"aDex-UI/internal/appdir"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// customLayoutsFileName is the user-editable file describing custom panel
// arrangements. It sits beside the UI settings so there is one obvious place
// to look, and is intentionally a separate file: it is hand-edited, whereas
// adex-ui-settings.json is written by the application.
const customLayoutsFileName = "layouts.json"

// CustomLayoutsPath returns the path the file is read from, whether or not it
// exists. The settings panel shows this so users know where to create it.
func CustomLayoutsPath() string {
	return filepath.Join(appdir.Config(), customLayoutsFileName)
}

// LoadCustomLayouts reads the user's layout definitions.
//
// The contents are returned as generic JSON rather than a typed struct: the
// frontend owns the panel and region vocabulary, validates against it, and
// drops individual bad entries with a message the user can act on. Parsing
// here too would mean maintaining that vocabulary in two places.
//
// A missing file is not an error — it is the normal case for a user who has
// not defined any layouts.
func LoadCustomLayouts() ([]interface{}, error) {
	path := CustomLayoutsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Accept either a bare array or an object with a "layouts" key, so the
	// file can grow other top-level settings later without breaking.
	var asArray []interface{}
	if err := json.Unmarshal(data, &asArray); err == nil {
		return asArray, nil
	}

	var asObject struct {
		Layouts []interface{} `json:"layouts"`
	}
	if err := json.Unmarshal(data, &asObject); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if asObject.Layouts == nil {
		return []interface{}{}, nil
	}
	return asObject.Layouts, nil
}
