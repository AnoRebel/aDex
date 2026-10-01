package settings

import (
	"aDex/internal/models"
)

// Catalogues of what the frontend actually ships. Kept here so an invalid
// persisted value is corrected once, at the persistence boundary, rather than
// being papered over by a frontend fallback on every start.
//
// This matters because the shipped Go default was once Theme: "default" — an
// id that has never existed in themes-index.json — so a fresh install
// persisted a value the theme loader could not resolve.
//
// Keep in sync with:
//   - frontend/app/assets/data/themes-index.json
//   - frontend/app/assets/css/layouts/*.css
var (
	validThemes = map[string]bool{
		"apollo": true, "apollo-notype": true, "blade": true,
		"chalkboard": true, "chalkboard-ligatures": true, "chalkboard-notype": true,
		"cyborg": true, "cyborg-focus": true, "interstellar": true,
		"matrix": true, "navy": true, "navy-disrupted": true, "navy-notype": true,
		"nord": true, "red": true, "tron": true, "tron-colorfilter": true,
		"tron-disrupted": true, "tron-fulltype": true, "tron-notype": true,
		"tron-typeleft": true,
	}

	validLayouts = map[string]bool{
		"colorfilter": true, "default": true, "disrupted": true,
		"fulltype": true, "notype": true, "typeleft": true,
		"terminal-focus": true,
	}
)

// defaultThemeID must match useAdexTheme's DEFAULT_THEME_ID.
const defaultThemeID = "tron"

// Globe rendering styles, matching the components pages/index.vue selects
// between: the stylised dot grid and the country-outline projection.
var validGlobeStyles = map[string]bool{
	"classic": true,
	"geo":     true,
}

const defaultGlobeStyle = "classic"

// validateDisplay repairs display settings that name something the
// application does not ship. Returns true when a value was changed, so the
// caller can persist the repair instead of re-reporting it every start.
//
// An empty layout is left as-is: it is the documented "follow the active
// theme's layout" value, not an invalid one.
func validateDisplay(s *models.UISettings) bool {
	if s == nil {
		return false
	}
	changed := false

	if !validThemes[s.Display.Theme] {
		s.Display.Theme = defaultThemeID
		changed = true
	}

	// An unset globe style means an older settings file written before the
	// choice existed; fill in the default rather than treating it as invalid.
	if s.Display.GlobeStyle == "" {
		s.Display.GlobeStyle = defaultGlobeStyle
		changed = true
	} else if !validGlobeStyles[s.Display.GlobeStyle] {
		s.Display.GlobeStyle = defaultGlobeStyle
		changed = true
	}

	// Only reject a layout that is neither built in NOR user-defined.
	// Custom layouts live in layouts.json and their ids are arbitrary, so
	// validating against the built-in list alone would silently erase a
	// user's own layout choice on every start.
	if s.Display.Layout != "" && !validLayouts[s.Display.Layout] && !isCustomLayoutID(s.Display.Layout) {
		s.Display.Layout = ""
		changed = true
	}

	return changed
}

// isCustomLayoutID reports whether the id matches a layout the user defined in
// layouts.json. Errors reading that file are treated as "not custom" — the
// caller then falls back to the default, which is the safe direction.
func isCustomLayoutID(id string) bool {
	entries, err := LoadCustomLayouts()
	if err != nil {
		return false
	}
	for _, e := range entries {
		m, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := m["id"].(string); ok && v == id {
			return true
		}
	}
	return false
}
