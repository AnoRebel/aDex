package settings

import (
	"testing"

	"aDex-UI/internal/models"
)

// A persisted theme id the application does not ship must be repaired to the
// default rather than applied — this is what let Theme: "default" (never a
// real id) survive as the shipped default and leave the interface unstyled.
func TestValidateDisplay_RepairsUnknownTheme(t *testing.T) {
	for _, bad := range []string{"default", "cyberpunk", "", "does-not-exist"} {
		s := models.DefaultUISettings()
		s.Display.Theme = bad

		if !validateDisplay(s) {
			t.Fatalf("theme %q: expected a repair to be reported", bad)
		}
		if s.Display.Theme != defaultThemeID {
			t.Errorf("theme %q: got %q, want %q", bad, s.Display.Theme, defaultThemeID)
		}
	}
}

func TestValidateDisplay_KeepsValidTheme(t *testing.T) {
	for _, good := range []string{"tron", "matrix", "nord", "cyborg-focus"} {
		s := models.DefaultUISettings()
		s.Display.Theme = good
		s.Display.Layout = ""

		if validateDisplay(s) {
			t.Errorf("theme %q is valid; no repair should be reported", good)
		}
		if s.Display.Theme != good {
			t.Errorf("theme %q was modified to %q", good, s.Display.Theme)
		}
	}
}

// An empty layout is the documented "follow the active theme's layout"
// value, so it must not be treated as invalid.
func TestValidateDisplay_EmptyLayoutIsValid(t *testing.T) {
	s := models.DefaultUISettings()
	s.Display.Theme = defaultThemeID
	s.Display.Layout = ""

	if validateDisplay(s) {
		t.Error("empty layout means 'follow the theme'; it should not be repaired")
	}
}

func TestValidateDisplay_RepairsUnknownLayout(t *testing.T) {
	s := models.DefaultUISettings()
	s.Display.Theme = defaultThemeID
	s.Display.Layout = "not-a-layout"

	if !validateDisplay(s) {
		t.Fatal("expected a repair to be reported for an unknown layout")
	}
	if s.Display.Layout != "" {
		t.Errorf("got layout %q, want empty (follow the theme)", s.Display.Layout)
	}
}

// The shipped defaults must themselves be valid — this is the regression that
// started the whole investigation.
func TestDefaultUISettings_AreValid(t *testing.T) {
	s := models.DefaultUISettings()
	if validateDisplay(s) {
		t.Errorf("DefaultUISettings() is not valid: theme=%q layout=%q",
			s.Display.Theme, s.Display.Layout)
	}
}
