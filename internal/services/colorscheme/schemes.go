package colorscheme

import (
	"time"

	"aDex-UI/internal/models"
)

// createDefaultDarkScheme creates the default dark color scheme
func (s *Service) createDefaultDarkScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("default-dark", "Default Dark")
	scheme.DisplayName = "Default Dark"
	scheme.Description = "Default dark theme for terminal"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#1e1e1e",
		Foreground:         "#f0f0f0",
		Cursor:             "#ffffff",
		CursorAccent:       "#000000",
		Selection:          "rgba(255, 255, 255, 0.3)",
		SelectionForeground: "#000000",
		Black:              "#000000",
		Red:                "#ff5555",
		Green:              "#50fa7b",
		Yellow:             "#f1fa8c",
		Blue:               "#bd93f9",
		Magenta:            "#ff79c6",
		Cyan:               "#8be9fd",
		White:              "#f8f8f2",
		BrightBlack:        "#6272a4",
		BrightRed:          "#ff6e6e",
		BrightGreen:        "#69ff94",
		BrightYellow:       "#ffffa5",
		BrightBlue:         "#d6acff",
		BrightMagenta:      "#ff92df",
		BrightCyan:         "#a4ffff",
		BrightWhite:        "#ffffff",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "eDEX-UI"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createDefaultLightScheme creates the default light color scheme
func (s *Service) createDefaultLightScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("default-light", "Default Light")
	scheme.DisplayName = "Default Light"
	scheme.Description = "Default light theme for terminal"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#ffffff",
		Foreground:         "#2e2e2e",
		Cursor:             "#000000",
		CursorAccent:       "#ffffff",
		Selection:          "rgba(0, 0, 0, 0.2)",
		SelectionForeground: "#ffffff",
		Black:              "#000000",
		Red:                "#cc0000",
		Green:              "#4e9a06",
		Yellow:             "#c4a000",
		Blue:               "#3465a4",
		Magenta:            "#75507b",
		Cyan:               "#06989a",
		White:              "#d3d7cf",
		BrightBlack:        "#555753",
		BrightRed:          "#ef2929",
		BrightGreen:        "#8ae234",
		BrightYellow:       "#fce94f",
		BrightBlue:         "#729fcf",
		BrightMagenta:      "#ad7fa8",
		BrightCyan:         "#34e2e2",
		BrightWhite:        "#eeeeec",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = false
	scheme.Author = "eDEX-UI"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createSolarizedDarkScheme creates the Solarized Dark color scheme
func (s *Service) createSolarizedDarkScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("solarized-dark", "Solarized Dark")
	scheme.DisplayName = "Solarized Dark"
	scheme.Description = "Solarized dark theme with careful color selection"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#002b36",
		Foreground:         "#839496",
		Cursor:             "#93a1a1",
		CursorAccent:       "#002b36",
		Selection:          "#073642",
		SelectionForeground: "#93a1a1",
		Black:              "#073642",
		Red:                "#dc322f",
		Green:              "#859900",
		Yellow:             "#b58900",
		Blue:               "#268bd2",
		Magenta:            "#d33682",
		Cyan:               "#2aa198",
		White:              "#eee8d5",
		BrightBlack:        "#002b36",
		BrightRed:          "#cb4b16",
		BrightGreen:        "#586e75",
		BrightYellow:       "#657b83",
		BrightBlue:         "#839496",
		BrightMagenta:      "#6c71c4",
		BrightCyan:         "#93a1a1",
		BrightWhite:        "#fdf6e3",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "Ethan Schoonover"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createSolarizedLightScheme creates the Solarized Light color scheme
func (s *Service) createSolarizedLightScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("solarized-light", "Solarized Light")
	scheme.DisplayName = "Solarized Light"
	scheme.Description = "Solarized light theme with careful color selection"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#fdf6e3",
		Foreground:         "#657b83",
		Cursor:             "#586e75",
		CursorAccent:       "#fdf6e3",
		Selection:          "#eee8d5",
		SelectionForeground: "#586e75",
		Black:              "#073642",
		Red:                "#dc322f",
		Green:              "#859900",
		Yellow:             "#b58900",
		Blue:               "#268bd2",
		Magenta:            "#d33682",
		Cyan:               "#2aa198",
		White:              "#eee8d5",
		BrightBlack:        "#002b36",
		BrightRed:          "#cb4b16",
		BrightGreen:        "#586e75",
		BrightYellow:       "#657b83",
		BrightBlue:         "#839496",
		BrightMagenta:      "#6c71c4",
		BrightCyan:         "#93a1a1",
		BrightWhite:        "#fdf6e3",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = false
	scheme.Author = "Ethan Schoonover"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createDraculaScheme creates the Dracula color scheme
func (s *Service) createDraculaScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("dracula", "Dracula")
	scheme.DisplayName = "Dracula"
	scheme.Description = "Dark theme inspired by the Dracula color palette"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#282a36",
		Foreground:         "#f8f8f2",
		Cursor:             "#f8f8f2",
		CursorAccent:       "#282a36",
		Selection:          "#44475a",
		SelectionForeground: "#f8f8f2",
		Black:              "#21222c",
		Red:                "#ff5555",
		Green:              "#50fa7b",
		Yellow:             "#f1fa8c",
		Blue:               "#bd93f9",
		Magenta:            "#ff79c6",
		Cyan:               "#8be9fd",
		White:              "#f8f8f2",
		BrightBlack:        "#6272a4",
		BrightRed:          "#ff6e6e",
		BrightGreen:        "#69ff94",
		BrightYellow:       "#ffffa5",
		BrightBlue:         "#d6acff",
		BrightMagenta:      "#ff92df",
		BrightCyan:         "#a4ffff",
		BrightWhite:        "#ffffff",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "Zeno Rocha"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createMonokaiScheme creates the Monokai color scheme
func (s *Service) createMonokaiScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("monokai", "Monokai")
	scheme.DisplayName = "Monokai"
	scheme.Description = "Classic Monokai theme for terminal"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#272822",
		Foreground:         "#f8f8f2",
		Cursor:             "#f8f8f0",
		CursorAccent:       "#272822",
		Selection:          "#49483e",
		SelectionForeground: "#f8f8f2",
		Black:              "#272822",
		Red:                "#f92672",
		Green:              "#a6e22e",
		Yellow:             "#f4bf75",
		Blue:               "#66d9ef",
		Magenta:            "#ae81ff",
		Cyan:               "#a1efe4",
		White:              "#f8f8f2",
		BrightBlack:        "#75715e",
		BrightRed:          "#f92672",
		BrightGreen:        "#a6e22e",
		BrightYellow:       "#f4bf75",
		BrightBlue:         "#66d9ef",
		BrightMagenta:      "#ae81ff",
		BrightCyan:         "#a1efe4",
		BrightWhite:        "#f9f8f5",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "Wimer Hazenberg"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createNordScheme creates the Nord color scheme
func (s *Service) createNordScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("nord", "Nord")
	scheme.DisplayName = "Nord"
	scheme.Description = "A polar night inspired color theme"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#2e3440",
		Foreground:         "#d8dee9",
		Cursor:             "#d8dee9",
		CursorAccent:       "#2e3440",
		Selection:          "#434c5e",
		SelectionForeground: "#d8dee9",
		Black:              "#3b4252",
		Red:                "#bf616a",
		Green:              "#a3be8c",
		Yellow:             "#ebcb8b",
		Blue:               "#5e81ac",
		Magenta:            "#b48ead",
		Cyan:               "#88c0d0",
		White:              "#e5e9f0",
		BrightBlack:        "#4c566a",
		BrightRed:          "#d08770",
		BrightGreen:        "#8fbcbb",
		BrightYellow:       "#ebcb8b",
		BrightBlue:         "#81a1c1",
		BrightMagenta:      "#b48ead",
		BrightCyan:         "#8fbcbb",
		BrightWhite:        "#eceff4",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "Arctic Ice Studio"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createGruvboxDarkScheme creates the Gruvbox Dark color scheme
func (s *Service) createGruvboxDarkScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("gruvbox-dark", "Gruvbox Dark")
	scheme.DisplayName = "Gruvbox Dark"
	scheme.Description = "Retro groove color scheme dark variant"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#282828",
		Foreground:         "#ebdbb2",
		Cursor:             "#ebdbb2",
		CursorAccent:       "#282828",
		Selection:          "#3c3836",
		SelectionForeground: "#ebdbb2",
		Black:              "#282828",
		Red:                "#cc241d",
		Green:              "#98971a",
		Yellow:             "#d79921",
		Blue:               "#458588",
		Magenta:            "#b16286",
		Cyan:               "#689d6a",
		White:              "#a89984",
		BrightBlack:        "#928374",
		BrightRed:          "#fb4934",
		BrightGreen:        "#b8bb26",
		BrightYellow:       "#fabd2f",
		BrightBlue:         "#83a598",
		BrightMagenta:      "#d3869b",
		BrightCyan:         "#8ec07c",
		BrightWhite:        "#ebdbb2",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = true
	scheme.Author = "Pavel Pertsev"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}

// createGruvboxLightScheme creates the Gruvbox Light color scheme
func (s *Service) createGruvboxLightScheme() *models.ColorScheme {
	scheme := models.NewColorScheme("gruvbox-light", "Gruvbox Light")
	scheme.DisplayName = "Gruvbox Light"
	scheme.Description = "Retro groove color scheme light variant"
	scheme.Colors = models.ColorSchemeColors{
		Background:         "#fbf1c7",
		Foreground:         "#3c3836",
		Cursor:             "#3c3836",
		CursorAccent:       "#fbf1c7",
		Selection:          "#ebdbb2",
		SelectionForeground: "#3c3836",
		Black:              "#fbf1c7",
		Red:                "#cc241d",
		Green:              "#98971a",
		Yellow:             "#d79921",
		Blue:               "#458588",
		Magenta:            "#b16286",
		Cyan:               "#689d6a",
		White:              "#7c6f64",
		BrightBlack:        "#928374",
		BrightRed:          "#9d0006",
		BrightGreen:        "#79740e",
		BrightYellow:       "#b57614",
		BrightBlue:         "#076678",
		BrightMagenta:      "#8f3f71",
		BrightCyan:         "#427b58",
		BrightWhite:        "#3c3836",
	}
	scheme.IsBuiltIn = true
	scheme.IsDark = false
	scheme.Author = "Pavel Pertsev"
	scheme.Version = "1.0.0"
	scheme.CreatedAt = time.Now()
	scheme.UpdatedAt = time.Now()
	return scheme
}