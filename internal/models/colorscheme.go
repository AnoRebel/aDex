package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// ColorScheme represents a terminal color theme
type ColorScheme struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"display_name"`
	Description string                `json:"description"`
	Colors      ColorSchemeColors     `json:"colors"`
	Font        ColorSchemeFont       `json:"font"`
	Cursor      ColorSchemeCursor     `json:"cursor"`
	Background  ColorSchemeBackground `json:"background"`
	Author      string                `json:"author"`
	Version     string                `json:"version"`
	IsBuiltIn   bool                  `json:"is_built_in"`
	IsDark      bool                  `json:"is_dark"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// ColorSchemeColors defines the color palette for a terminal theme
type ColorSchemeColors struct {
	Background        string            `json:"background"`
	Foreground        string            `json:"foreground"`
	Cursor            string            `json:"cursor"`
	CursorAccent      string            `json:"cursor_accent,omitempty"`
	Selection         string            `json:"selection"`
	SelectionForeground string          `json:"selection_foreground,omitempty"`
	Black             string            `json:"black"`
	Red               string            `json:"red"`
	Green             string            `json:"green"`
	Yellow            string            `json:"yellow"`
	Blue              string            `json:"blue"`
	Magenta           string            `json:"magenta"`
	Cyan              string            `json:"cyan"`
	White             string            `json:"white"`
	BrightBlack       string            `json:"bright_black"`
	BrightRed         string            `json:"bright_red"`
	BrightGreen       string            `json:"bright_green"`
	BrightYellow      string            `json:"bright_yellow"`
	BrightBlue        string            `json:"bright_blue"`
	BrightMagenta     string            `json:"bright_magenta"`
	BrightCyan        string            `json:"bright_cyan"`
	BrightWhite       string            `json:"bright_white"`
	ANSI              map[int]string    `json:"ansi,omitempty"`
}

// ColorSchemeFont defines font settings for a color scheme
type ColorSchemeFont struct {
	Family      string  `json:"family"`
	Size        int     `json:"size"`
	Weight      string  `json:"weight"`
	LineHeight  float64 `json:"line_height"`
	LetterSpacing *int  `json:"letter_spacing,omitempty"`
	Ligatures   bool    `json:"ligatures"`
	Antialias   bool    `json:"antialias"`
	Hinting     string  `json:"hinting"`
}

// ColorSchemeCursor defines cursor appearance for a color scheme
type ColorSchemeCursor struct {
	Style         string  `json:"style"`         // block, underline, bar
	Blink         bool    `json:"blink"`
	BlinkInterval *int    `json:"blink_interval,omitempty"`
	Width         *int    `json:"width,omitempty"`
	Color         string  `json:"color,omitempty"`
	Accent        string  `json:"accent,omitempty"`
}

// ColorSchemeBackground defines background settings for a color scheme
type ColorSchemeBackground struct {
	Type     string  `json:"type"`     // solid, gradient, image
	Value    string  `json:"value,omitempty"`
	Opacity  *int    `json:"opacity,omitempty"`    // 0-100
	Blur     *int    `json:"blur,omitempty"`      // 0-100
	Size     string  `json:"size,omitempty"`      // cover, contain, auto
	Position string  `json:"position,omitempty"`
}

// ColorSchemeConfig holds configuration for color scheme management
type ColorSchemeConfig struct {
	DefaultScheme string                   `json:"default_scheme"`
	UserSchemes   map[string]*ColorScheme   `json:"user_schemes"`
	EnabledSchemes []string                `json:"enabled_schemes"`
	AutoSwitch    bool                     `json:"auto_switch"`     // Switch based on system theme
	ImportPath    string                   `json:"import_path"`     // Path to import custom schemes
	ExportPath    string                   `json:"export_path"`     // Path to export schemes
}

// ColorSchemeImport represents a color scheme import configuration
type ColorSchemeImport struct {
	Source      string                 `json:"source"`        // file, url, clipboard
	Format      string                 `json:"format"`        // json, yaml, toml, ini
	Content     map[string]interface{} `json:"content"`
	Metadata    map[string]string      `json:"metadata"`
	Preview     bool                   `json:"preview"`       // Show preview before importing
	Overwrite   bool                   `json:"overwrite"`     // Overwrite existing scheme
}

// ColorSchemeExport represents a color scheme export configuration
type ColorSchemeExport struct {
	SchemeIDs   []string `json:"scheme_ids"`   // IDs of schemes to export
	Format      string   `json:"format"`       // json, yaml, toml, ini
	IncludeBuiltIn bool  `json:"include_built_in"`
	Compress    bool     `json:"compress"`     // Compress output file
	Preview     bool     `json:"preview"`      // Show preview before export
}

// ColorSchemeValidationResult represents the result of color scheme validation
type ColorSchemeValidationResult struct {
	Valid   bool                   `json:"valid"`
	Errors  []ColorSchemeError     `json:"errors"`
	Warnings []ColorSchemeWarning  `json:"warnings"`
	Scheme  *ColorScheme           `json:"scheme,omitempty"`
}

// ColorSchemeError represents an error in color scheme validation
type ColorSchemeError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ColorSchemeWarning represents a warning in color scheme validation
type ColorSchemeWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ColorSchemePreview represents a preview of a color scheme
type ColorSchemePreview struct {
	SchemeID  string                 `json:"scheme_id"`
	Name      string                 `json:"name"`
	Colors    map[string]string      `json:"colors"`
	Preview   string                 `json:"preview"`     // Base64 encoded preview image
	Sample    string                 `json:"sample"`      // Sample terminal output
	CreatedAt time.Time              `json:"created_at"`
}

// NewColorScheme creates a new color scheme with default values
func NewColorScheme(id, name string) *ColorScheme {
	now := time.Now()
	return &ColorScheme{
		ID:          id,
		Name:        name,
		DisplayName: name,
		Description: "",
		Colors: ColorSchemeColors{
			Background:  "#1e1e1e",
			Foreground:  "#f0f0f0",
			Cursor:      "#ffffff",
			Selection:   "rgba(255, 255, 255, 0.3)",
			Black:       "#000000",
			Red:         "#ff5555",
			Green:       "#50fa7b",
			Yellow:      "#f1fa8c",
			Blue:        "#bd93f9",
			Magenta:     "#ff79c6",
			Cyan:        "#8be9fd",
			White:       "#f8f8f2",
			BrightBlack: "#6272a4",
			BrightRed:   "#ff6e6e",
			BrightGreen: "#69ff94",
			BrightYellow:"#ffffa5",
			BrightBlue:  "#d6acff",
			BrightMagenta:"#ff92df",
			BrightCyan:  "#a4ffff",
			BrightWhite: "#ffffff",
		},
		Font: ColorSchemeFont{
			Family:     "JetBrains Mono",
			Size:       14,
			Weight:     "normal",
			LineHeight: 1.5,
			Ligatures:  true,
			Antialias:  true,
			Hinting:    "slight",
		},
		Cursor: ColorSchemeCursor{
			Style: "block",
			Blink: false,
		},
		Background: ColorSchemeBackground{
			Type:    "solid",
			Opacity: intPtr(100),
		},
		IsBuiltIn: false,
		IsDark:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Extensions: make(map[string]interface{}),
	}
}

// Validate validates a color scheme
func (cs *ColorScheme) Validate() *ColorSchemeValidationResult {
	result := &ColorSchemeValidationResult{
		Valid:    true,
		Errors:   make([]ColorSchemeError, 0),
		Warnings: make([]ColorSchemeWarning, 0),
		Scheme:   cs,
	}

	// Validate required fields
	if cs.ID == "" {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "id",
			Message: "ID is required",
			Code:    "required",
		})
		result.Valid = false
	}

	if cs.Name == "" {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "name",
			Message: "Name is required",
			Code:    "required",
		})
		result.Valid = false
	}

	// Validate colors
	validateColor := func(color, field string) {
		if color == "" {
			result.Errors = append(result.Errors, ColorSchemeError{
				Field:   field,
				Message: "Color is required",
				Code:    "required",
			})
			result.Valid = false
		} else if !isValidColor(color) {
			result.Errors = append(result.Errors, ColorSchemeError{
				Field:   field,
				Message: "Invalid color format",
				Code:    "invalid_format",
			})
			result.Valid = false
		}
	}

	validateColor(cs.Colors.Background, "colors.background")
	validateColor(cs.Colors.Foreground, "colors.foreground")
	validateColor(cs.Colors.Cursor, "colors.cursor")
	validateColor(cs.Colors.Selection, "colors.selection")

	// Validate font
	if cs.Font.Size <= 0 {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "font.size",
			Message: "Font size must be greater than 0",
			Code:    "invalid_value",
		})
		result.Valid = false
	}

	if cs.Font.LineHeight <= 0 {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "font.line_height",
			Message: "Line height must be greater than 0",
			Code:    "invalid_value",
		})
		result.Valid = false
	}

	// Validate cursor style
	validCursorStyles := []string{"block", "underline", "bar"}
	if !contains(validCursorStyles, cs.Cursor.Style) {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "cursor.style",
			Message: "Invalid cursor style",
			Code:    "invalid_value",
		})
		result.Valid = false
	}

	// Validate background type
	validBackgroundTypes := []string{"solid", "gradient", "image"}
	if !contains(validBackgroundTypes, cs.Background.Type) {
		result.Errors = append(result.Errors, ColorSchemeError{
			Field:   "background.type",
			Message: "Invalid background type",
			Code:    "invalid_value",
		})
		result.Valid = false
	}

	// Check for warnings
	if cs.Colors.Background == cs.Colors.Foreground {
		result.Warnings = append(result.Warnings, ColorSchemeWarning{
			Field:   "colors",
			Message: "Background and foreground colors are identical",
			Code:    "low_contrast",
		})
	}

	return result
}

// ToXtermTheme converts the color scheme to xterm.js theme format
func (cs *ColorScheme) ToXtermTheme() map[string]interface{} {
	theme := map[string]interface{}{
		"background": cs.Colors.Background,
		"foreground": cs.Colors.Foreground,
		"cursor":     cs.Colors.Cursor,
		"selection":  cs.Colors.Selection,
		"black":      cs.Colors.Black,
		"red":        cs.Colors.Red,
		"green":      cs.Colors.Green,
		"yellow":     cs.Colors.Yellow,
		"blue":       cs.Colors.Blue,
		"magenta":    cs.Colors.Magenta,
		"cyan":       cs.Colors.Cyan,
		"white":      cs.Colors.White,
		"brightBlack":    cs.Colors.BrightBlack,
		"brightRed":      cs.Colors.BrightRed,
		"brightGreen":    cs.Colors.BrightGreen,
		"brightYellow":   cs.Colors.BrightYellow,
		"brightBlue":     cs.Colors.BrightBlue,
		"brightMagenta":  cs.Colors.BrightMagenta,
		"brightCyan":     cs.Colors.BrightCyan,
		"brightWhite":    cs.Colors.BrightWhite,
	}

	// Add optional colors
	if cs.Colors.CursorAccent != "" {
		theme["cursorAccent"] = cs.Colors.CursorAccent
	}
	if cs.Colors.SelectionForeground != "" {
		theme["selectionForeground"] = cs.Colors.SelectionForeground
	}

	// Add ANSI colors if available
	if len(cs.Colors.ANSI) > 0 {
		for colorCode, color := range cs.Colors.ANSI {
			theme[fmt.Sprintf("ansi%d", colorCode)] = color
		}
	}

	return theme
}

// Clone creates a deep copy of the color scheme
func (cs *ColorScheme) Clone() *ColorScheme {
	clone := *cs

	// Deep copy maps
	if cs.Extensions != nil {
		clone.Extensions = make(map[string]interface{})
		for k, v := range cs.Extensions {
			clone.Extensions[k] = v
		}
	}

	if cs.Colors.ANSI != nil {
		clone.Colors.ANSI = make(map[int]string)
		for k, v := range cs.Colors.ANSI {
			clone.Colors.ANSI[k] = v
		}
	}

	return &clone
}

// ToJSON converts the color scheme to JSON
func (cs *ColorScheme) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cs, "", "  ")
}

// FromJSON creates a color scheme from JSON
func FromJSON(data []byte) (*ColorScheme, error) {
	var scheme ColorScheme
	err := json.Unmarshal(data, &scheme)
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func isValidColor(color string) bool {
	// Simple validation for hex colors and rgba colors
	if len(color) == 7 && color[0] == '#' {
		return true
	}
	if len(color) == 4 && color[0] == '#' {
		return true
	}
	if len(color) > 4 && color[:4] == "rgba" {
		return true
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}