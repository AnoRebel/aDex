package settings

import (
	"aDex-UI/internal/models"
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
	}
)

// defaultThemeID must match useAdexTheme's DEFAULT_THEME_ID.
const defaultThemeID = "tron"

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

	if s.Display.Layout != "" && !validLayouts[s.Display.Layout] {
		s.Display.Layout = ""
		changed = true
	}

	return changed
}
