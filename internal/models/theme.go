package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Theme represents a complete theme configuration
type Theme struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Author      string                `json:"author"`
	Version     string                `json:"version"`
	Colors      ThemeColors           `json:"colors"`
	Fonts       ThemeFonts            `json:"fonts"`
	Effects     ThemeEffects          `json:"effects"`
	Settings    ThemeSettings         `json:"settings"`
	Metadata    ThemeMetadata         `json:"metadata"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	IsBuiltIn   bool                  `json:"is_builtin"`
	IsActive    bool                  `json:"is_active"`
	Category    ThemeCategory         `json:"category"`
	Tags        []string              `json:"tags"`
	Preview     *ThemePreview         `json:"preview,omitempty"`
}

// ThemeColors contains all color definitions for a theme
type ThemeColors struct {
	// Primary colors
	Primary      ColorValue `json:"primary"`
	PrimaryHover ColorValue `json:"primaryHover"`
	PrimaryActive ColorValue `json:"primaryActive"`

	// Secondary colors
	Secondary      ColorValue `json:"secondary"`
	SecondaryHover ColorValue `json:"secondaryHover"`
	SecondaryActive ColorValue `json:"secondaryActive"`

	// Accent colors
	Accent      ColorValue `json:"accent"`
	AccentHover ColorValue `json:"accentHover"`
	AccentActive ColorValue `json:"accentActive"`

	// Surface colors
	Background   ColorValue `json:"background"`
	Surface      ColorValue `json:"surface"`
	SurfaceHover ColorValue `json:"surfaceHover"`
	SurfaceBorder ColorValue `json:"surfaceBorder"`

	// Text colors
	TextPrimary   ColorValue `json:"textPrimary"`
	TextSecondary ColorValue `json:"textSecondary"`
	TextTertiary  ColorValue `json:"textTertiary"`
	TextInverse   ColorValue `json:"textInverse"`

	// Status colors
	Success       ColorValue `json:"success"`
	Warning       ColorValue `json:"warning"`
	Error         ColorValue `json:"error"`
	Info          ColorValue `json:"info"`

	// Special colors
	Glitch       ColorValue `json:"glitch"`
	Cursor        ColorValue `json:"cursor"`
	Selection     ColorValue `json:"selection"`

	// Component specific colors
	Terminal     *TerminalColors `json:"terminal,omitempty"`
	Charts       *ChartColors     `json:"charts,omitempty"`
}

// ColorValue represents a color that can be hex, rgb, or named
type ColorValue struct {
	Value string `json:"value"`
	Type  string `json:"type"` // "hex", "rgb", "rgba", "hsl", "named"
}

// TerminalColors contains terminal-specific color definitions
type TerminalColors struct {
	Background ColorValue   `json:"background"`
	Foreground ColorValue   `json:"foreground"`
	Cursor     ColorValue   `json:"cursor"`
	Selection  ColorValue   `json:"selection"`
	Black      ColorValue   `json:"black"`
	Red        ColorValue   `json:"red"`
	Green      ColorValue   `json:"green"`
	Yellow     ColorValue   `json:"yellow"`
	Blue       ColorValue   `json:"blue"`
	Magenta    ColorValue   `json:"magenta"`
	Cyan       ColorValue   `json:"cyan"`
	White      ColorValue   `json:"white"`
	BrightBlack   ColorValue `json:"brightBlack"`
	BrightRed     ColorValue `json:"brightRed"`
	BrightGreen   ColorValue `json:"brightGreen"`
	BrightYellow  ColorValue `json:"brightYellow"`
	BrightBlue    ColorValue `json:"brightBlue"`
	BrightMagenta ColorValue `json:"brightMagenta"`
	BrightCyan    ColorValue `json:"brightCyan"`
	BrightWhite   ColorValue `json:"brightWhite"`
}

// ChartColors contains color definitions for charts
type ChartColors struct {
	Primary   []ColorValue `json:"primary"`
	Secondary []ColorValue `json:"secondary"`
	Gradient  []GradientStop `json:"gradient"`
	Grid      ColorValue    `json:"grid"`
	Axis      ColorValue    `json:"axis"`
}

// GradientStop represents a color stop in a gradient
type GradientStop struct {
	Color ColorValue `json:"color"`
	Stop  float64     `json:"stop"` // 0.0 to 1.0
}

// ThemeFonts contains font configuration
type ThemeFonts struct {
	Primary    FontConfig `json:"primary"`
	Monospace  FontConfig `json:"monospace"`
	Secondary  FontConfig `json:"secondary"`
	Display    FontConfig `json:"display"`
	UI         FontConfig `json:"ui"`
}

// FontConfig represents font configuration
type FontConfig struct {
	Family   string   `json:"family"`
	Size      string   `json:"size"`      // CSS font-size value
	Weight    string   `json:"weight"`    // CSS font-weight value
	Style     string   `json:"style"`     // normal, italic
	Variants  []string `json:"variants"`  // Available font variants
	Features  []string `json:"features"`  // OpenType features
}

// ThemeEffects contains visual effects configuration
type ThemeEffects struct {
	BorderRadius   string `json:"borderRadius"`
	BoxShadow      string `json:"boxShadow"`
	TextShadow     string `json:"textShadow"`
	Blur           string `json:"blur"`
	Glow           string `json:"glow"`
	Animation      *AnimationConfig `json:"animation,omitempty"`
}

// AnimationConfig contains animation settings
type AnimationConfig struct {
	Duration  string `json:"duration"`
	Easing    string `json:"easing"`
	Delay     string `json:"delay"`
	Count     string `json:"count"` // infinite, number
}

// ThemeSettings contains behavior settings
type ThemeSettings struct {
	AutoSave     bool          `json:"autoSave"`
	Transitions  bool          `json:"transitions"`
	ReduceMotion bool          `json:"reduceMotion"`
	HighContrast bool          `json:"highContrast"`
	Animations  bool          `json:"animations"`
	CustomCSS    string        `json:"customCSS"`
	Variables    ThemeVariables `json:"variables"`
}

// ThemeVariables contains custom CSS variables
type ThemeVariables struct {
	BorderWidth    map[string]string `json:"borderWidth"`
	Spacing         map[string]string `json:"spacing"`
	BorderRadius    map[string]string `json:"borderRadius"`
	Opacity         map[string]string `json:"opacity"`
	ZIndex          map[string]string `json:"zIndex"`
	Custom          map[string]string `json:"custom"`
}

// ThemeMetadata contains additional metadata
type ThemeMetadata struct {
	Screenshots     []string           `json:"screenshots"`
	DemoData        interface{}        `json:"demoData"`
	Compatibility    []string           `json:"compatibility"`
	Requirements    []string           `json:"requirements"`
	Limitations     []string           `json:"limitations"`
	License         string             `json:"license"`
	Repository      string             `json:"repository"`
	Documentation   string             `json:"documentation"`
	Changelog       string             `json:"changelog"`
}

// ThemePreview contains preview information
type ThemePreview struct {
	Thumbnail string  `json:"thumbnail"`
	Colors    []string `json:"colors"`
	Layout    string  `json:"layout"`
	Components []string `json:"components"`
}

// ThemeCategory represents theme categories
type ThemeCategory string

const (
	ThemeCategoryDark      ThemeCategory = "dark"
	ThemeCategoryLight     ThemeCategory = "light"
	ThemeCategoryColorful  ThemeCategory = "colorful"
	ThemeCategoryMinimal   ThemeCategory = "minimal"
	ThemeCategoryCyberpunk ThemeCategory = "cyberpunk"
	ThemeCategoryRetro     ThemeCategory = "retro"
	ThemeCategoryCustom    ThemeCategory = "custom"
	ThemeCategorySystem    ThemeCategory = "system"
)

// ThemePreset represents a predefined theme preset
type ThemePreset struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Theme       *Theme      `json:"theme"`
	Parent      string      `json:"parent"` // Parent theme ID
}

// ThemeValidationError represents a theme validation error
type ThemeValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value"`
}

// ThemeValidationResult contains validation results
type ThemeValidationResult struct {
	IsValid bool                    `json:"isValid"`
	Errors  []ThemeValidationError  `json:"errors"`
	Warnings []ThemeValidationError  `json:"warnings"`
}

// ThemeExportOptions contains options for theme export
type ThemeExportOptions struct {
	IncludePreview   bool     `json:"includePreview"`
	IncludeMetadata  bool     `json:"includeMetadata"`
	CompactFormat    bool     `json:"compactFormat"`
	MinifyCSS       bool     `json:"minifyCSS"`
	IncludeVariables bool     `json:"includeVariables"`
	ExportType       string  `json:"exportType"` // "json", "css", "both"
}

// ThemeImportOptions contains options for theme import
type ThemeImportOptions struct {
	Validate      bool   `json:"validate"`
	Overwrite     bool   `json:"overwrite"`
	GeneratePreview bool  `json:"generatePreview"`
	DetectType    bool   `json:"detectType"`
	FixColors     bool   `json:"fixColors"`
}

// LegacyTheme represents the old eDEX-UI theme format
type LegacyTheme struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Colors      map[string]interface{} `json:"colors"`
	Fonts       map[string]interface{} `json:"fonts"`
	Effects     map[string]interface{} `json:"effects"`
	Settings    map[string]interface{} `json:"settings"`
	Version     string                 `json:"version"`
	Created     string                 `json:"created"`
}

// Helper methods for Theme

// NewTheme creates a new theme with default values
func NewTheme(id, name string) *Theme {
	now := time.Now()
	return &Theme{
		ID:          id,
		Name:        name,
		Description: "",
		Author:      "Unknown",
		Version:     "1.0.0",
		Colors:      NewDefaultColors(),
		Fonts:       NewDefaultFonts(),
		Effects:     NewDefaultEffects(),
		Settings:    NewDefaultSettings(),
		Metadata:    NewDefaultMetadata(),
		CreatedAt:   now,
		UpdatedAt:   now,
		IsBuiltIn:   false,
		IsActive:    false,
		Category:    ThemeCategoryCustom,
		Tags:        []string{},
	}
}

// NewDefaultColors creates default color values
func NewDefaultColors() ThemeColors {
	return ThemeColors{
		Primary:       ColorValue{Value: "#0ea5e9", Type: "hex"},
		PrimaryHover:  ColorValue{Value: "#0284c7", Type: "hex"},
		PrimaryActive: ColorValue{Value: "#0369a1", Type: "hex"},
		Secondary:     ColorValue{Value: "#64748b", Type: "hex"},
		SecondaryHover: ColorValue{Value: "#475569", Type: "hex"},
		SecondaryActive: ColorValue{Value: "#334155", Type: "hex"},
		Accent:        ColorValue{Value: "#f59e0b", Type: "hex"},
		AccentHover:   ColorValue{Value: "#d97706", Type: "hex"},
		AccentActive:  ColorValue{Value: "#b45309", Type: "hex"},
		Background:    ColorValue{Value: "#0a0a0a", Type: "hex"},
		Surface:       ColorValue{Value: "#1a1a1a", Type: "hex"},
		SurfaceHover:  ColorValue{Value: "#2a2a2a", Type: "hex"},
		SurfaceBorder:  ColorValue{Value: "#374151", Type: "hex"},
		TextPrimary:   ColorValue{Value: "#f8fafc", Type: "hex"},
		TextSecondary: ColorValue{Value: "#cbd5e1", Type: "hex"},
		TextTertiary:  ColorValue{Value: "#94a3b8", Type: "hex"},
		TextInverse:   ColorValue{Value: "#1f2937", Type: "hex"},
		Success:       ColorValue{Value: "#10b981", Type: "hex"},
		Warning:       ColorValue{Value: "#f59e0b", Type: "hex"},
		Error:         ColorValue{Value: "#ef4444", Type: "hex"},
		Info:          ColorValue{Value: "#3b82f6", Type: "hex"},
		Glitch:       ColorValue{Value: "#00ff00", Type: "hex"},
		Cursor:        ColorValue{Value: "#ffffff", Type: "hex"},
		Selection:     ColorValue{Value: "#3b82f6", Type: "hex"},
	}
}

// NewDefaultFonts creates default font configuration
func NewDefaultFonts() ThemeFonts {
	return ThemeFonts{
		Primary: FontConfig{
			Family:   "Inter, system-ui, sans-serif",
			Size:      "14px",
			Weight:   "400",
			Style:     "normal",
			Variants:  []string{"100", "200", "300", "400", "500", "600", "700", "800", "900"},
			Features:  []string{"liga", "kern"},
		},
		Monospace: FontConfig{
			Family:   "Fira Code, Monaco, monospace",
			Size:      "13px",
			Weight:   "400",
			Style:     "normal",
			Variants:  []string{"300", "400", "500", "600", "700"},
			Features:  []string{"liga", "kern"},
		},
		Secondary: FontConfig{
			Family:   "JetBrains Mono, monospace",
			Size:      "12px",
			Weight:   "400",
			Style:     "normal",
			Variants:  []string{"100", "200", "300", "400", "500", "600", "700", "800"},
			Features:  []string{"liga", "kern"},
		},
		Display: FontConfig{
			Family:   "Orbitron, system-ui, sans-serif",
			Size:      "24px",
			Weight:   "700",
			Style:     "normal",
			Variants:  []string{"400", "700", "900"},
			Features:  []string{},
		},
		UI: FontConfig{
			Family:   "system-ui, sans-serif",
			Size:      "12px",
			Weight:   "500",
			Style:     "normal",
			Variants:  []string{"400", "500", "600", "700"},
			Features:  []string{},
		},
	}
}

// NewDefaultEffects creates default effects
func NewDefaultEffects() ThemeEffects {
	return ThemeEffects{
		BorderRadius: "8px",
		BoxShadow:    "0 4px 20px rgba(0, 0, 0, 0.5)",
		TextShadow:   "0 1px 2px rgba(0, 0, 0, 0.1)",
		Blur:         "2px",
		Glow:         "0 0 10px rgba(14, 165, 233, 0.3)",
		Animation: &AnimationConfig{
			Duration: "0.2s",
			Easing:   "ease",
			Delay:    "0s",
			Count:    "1",
		},
	}
}

// NewDefaultSettings creates default settings
func NewDefaultSettings() ThemeSettings {
	return ThemeSettings{
		AutoSave:     true,
		Transitions:  true,
		ReduceMotion: false,
		HighContrast: false,
		Animations:  true,
		CustomCSS:    "",
		Variables:    NewDefaultVariables(),
	}
}

// NewDefaultVariables creates default CSS variables
func NewDefaultVariables() ThemeVariables {
	return ThemeVariables{
		BorderWidth: map[string]string{
			"thin":  "1px",
			"normal": "2px",
			"thick":  "4px",
		},
		Spacing: map[string]string{
			"xs":   "4px",
			"sm":   "8px",
			"md":   "16px",
			"lg":   "24px",
			"xl":   "32px",
		},
		BorderRadius: map[string]string{
			"sm":  "4px",
			"md":  "8px",
			"lg":  "12px",
			"xl":  "16px",
			"full": "9999px",
		},
		Opacity: map[string]string{
			"hidden": "0",
			"light":  "0.3",
			"medium": "0.6",
			"strong": "0.9",
			"visible": "1",
		},
		ZIndex: map[string]string{
			"dropdown":   "1000",
			"modal":     "2000",
			"tooltip":   "3000",
			"notification": "4000",
		},
		Custom: make(map[string]string),
	}
}

// NewDefaultMetadata creates default metadata
func NewDefaultMetadata() ThemeMetadata {
	return ThemeMetadata{
	Compatibility: []string{"modern", "responsive"},
	Requirements: []string{"CSS3", "JavaScript ES6+"},
	License:      "MIT",
	Repository:   "",
	Documentation: "",
	Changelog:     "",
	}
}

// Validation methods

// Validate validates the theme and returns any errors
func (t *Theme) Validate() ThemeValidationResult {
	var errors []ThemeValidationError
	var warnings []ThemeValidationError

	// Validate required fields
	if t.ID == "" {
		errors = append(errors, ThemeValidationError{
			Field:   "id",
			Message: "Theme ID is required",
			Value:   t.ID,
		})
	}

	if t.Name == "" {
		errors = append(errors, ThemeValidationError{
			Field:   "name",
			Message: "Theme name is required",
			Value:   t.Name,
		})
	}

	// Validate colors
	if colorErrors := t.validateColors(); len(colorErrors) > 0 {
		errors = append(errors, colorErrors...)
	}

	// Validate fonts
	if fontErrors := t.validateFonts(); len(fontErrors) > 0 {
		errors = append(errors, fontErrors...)
	}

	// Validate category
	if !isValidCategory(t.Category) {
		warnings = append(warnings, ThemeValidationError{
			Field:   "category",
			Message: "Unknown theme category",
			Value:   t.Category,
		})
	}

	return ThemeValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// validateColors validates color values
func (t *Theme) validateColors() []ThemeValidationError {
	var errors []ThemeValidationError

	colorValidation := func(field string, color ColorValue) {
		if color.Value == "" {
			errors = append(errors, ThemeValidationError{
				Field:   field,
				Message: "Color value cannot be empty",
				Value:   color.Value,
			})
			return
		}

		if !isValidColor(color.Value) {
			errors = append(errors, ThemeValidationError{
				Field:   field,
				Message: "Invalid color format",
				Value:   color.Value,
			})
		}
	}

	// Validate all color fields
	colorValidation("colors.primary", t.Colors.Primary)
	colorValidation("colors.background", t.Colors.Background)
	colorValidation("colors.textPrimary", t.Colors.TextPrimary)
	colorValidation("colors.success", t.Colors.Success)
	colorValidation("colors.error", t.Colors.Error)

	return errors
}

// validateFonts validates font configurations
func (t *Theme) validateFonts() []ThemeValidationError {
	var errors []ThemeValidationError

	fontValidation := func(field string, font FontConfig) {
		if font.Family == "" {
			errors = append(errors, ThemeValidationError{
				Field:   field,
				Message: "Font family cannot be empty",
				Value:   font.Family,
			})
		}

		if font.Size == "" {
			errors = append(errors, ThemeValidationError{
				Field:   field,
				Message: "Font size cannot be empty",
				Value:   font.Size,
			})
		}

		if font.Weight == "" {
			errors = append(errors, ThemeValidationError{
				Field:   field,
				Message: "Font weight cannot be empty",
				Value:   font.Weight,
			})
		}
	}

	fontValidation("fonts.primary", t.Fonts.Primary)
	fontValidation("fonts.monospace", t.Fonts.Monospace)
	fontValidation("fonts.secondary", t.Fonts.Secondary)

	return errors
}

// Helper functions

// isValidColor checks if a color value is valid
func isValidColor(color string) bool {
	// Check hex colors
	if strings.HasPrefix(color, "#") {
		return regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`).MatchString(color)
	}

	// Check rgb/rgba colors
	if strings.HasPrefix(color, "rgb") {
		return regexp.MustCompile(`^rgba?\(\s*(\d+%?\s*,\s*){2,3}\s*[\d.]+\s*\)$`).MatchString(color)
	}

	// Check hsl/hsla colors
	if strings.HasPrefix(color, "hsl") {
		return regexp.MustCompile(`^hsla?\(\s*(\d+%?\s*,\s*){2,3}\s*[\d.]+\s*\)$`).MatchString(color)
	}

	// Check named colors
	namedColors := map[string]bool{
		"transparent": true, "inherit": true, "initial": true, "unset": true,
		"black": true, "white": true, "red": true, "green": true, "blue": true,
		// Add more named colors as needed
	}
	return namedColors[strings.ToLower(color)]
}

// isValidCategory checks if a theme category is valid
func isValidCategory(category ThemeCategory) bool {
	validCategories := map[ThemeCategory]bool{
		ThemeCategoryDark:      true,
		ThemeCategoryLight:     true,
		ThemeCategoryColorful:  true,
		ThemeCategoryMinimal:   true,
		ThemeCategoryCyberpunk: true,
		ThemeCategoryRetro:     true,
		ThemeCategoryCustom:    true,
		ThemeCategorySystem:    true,
	}
	return validCategories[category]
}

// ToJSON converts theme to JSON
func (t *Theme) ToJSON(options *ThemeExportOptions) ([]byte, error) {
	if options == nil {
		options = &ThemeExportOptions{
			IncludePreview:  true,
			IncludeMetadata: true,
			CompactFormat:   false,
		}
	}

	// Create export copy
	exportTheme := *t

	if !options.IncludePreview {
		exportTheme.Preview = nil
	}

	if !options.IncludeMetadata {
		exportTheme.Metadata = ThemeMetadata{}
	}

	if options.CompactFormat {
		return json.Marshal(exportTheme)
	}

	return json.MarshalIndent(exportTheme, "", "  ")
}

// FromJSON creates theme from JSON
func (t *Theme) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// Clone creates a deep copy of the theme
func (t *Theme) Clone() *Theme {
	data, err := json.Marshal(t)
	if err != nil {
		return nil
	}

	var clone Theme
	err = json.Unmarshal(data, &clone)
	if err != nil {
		return nil
	}

	return &clone
}

// UpdateTimestamp updates the theme's updated timestamp
func (t *Theme) UpdateTimestamp() {
	t.UpdatedAt = time.Now()
}

// GenerateCSS generates CSS variables for the theme
func (t *Theme) GenerateCSS() string {
	var css strings.Builder

	css.WriteString(":root {\n")

	// Color variables
	t.generateColorCSS(&css)

	// Font variables
	t.generateFontCSS(&css)

	// Effect variables
	t.generateEffectCSS(&css)

	// Custom variables
	t.generateVariableCSS(&css)

	css.WriteString("}\n")

	return css.String()
}

func (t *Theme) generateColorCSS(css *strings.Builder) {
	colors := map[string]ColorValue{
		"--color-primary":          t.Colors.Primary,
		"--color-primary-hover":     t.Colors.PrimaryHover,
		"--color-primary-active":    t.Colors.PrimaryActive,
		"--color-secondary":        t.Colors.Secondary,
		"--color-secondary-hover":   t.Colors.SecondaryHover,
		"--color-secondary-active":  t.Colors.SecondaryActive,
		"--color-accent":           t.Colors.Accent,
	"--color-accent-hover":      t.Colors.AccentHover,
		"--color-accent-active":     t.Colors.AccentActive,
		"--color-background":       t.Colors.Background,
		"--color-surface":          t.Colors.Surface,
		"--color-surface-hover":     t.Colors.SurfaceHover,
		"--color-surface-border":    t.Colors.SurfaceBorder,
		"--color-text-primary":     t.Colors.TextPrimary,
		"--color-text-secondary":   t.Colors.TextSecondary,
	"--color-text-tertiary":    t.Colors.TextTertiary,
	"--color-text-inverse":     t.Colors.TextInverse,
		"--color-success":          t.Colors.Success,
		"--color-warning":          t.Colors.Warning,
		"--color-error":            t.Colors.Error,
	"--color-info":             t.Colors.Info,
		"--color-glitch":           t.Colors.Glitch,
	"--color-cursor":           t.Colors.Cursor,
	"--color-selection":        t.Colors.Selection,
	}

	for name, color := range colors {
		fmt.Fprintf(css, "  %s: %s;\n", name, color.Value)
	}
}

func (t *Theme) generateFontCSS(css *strings.Builder) {
	fonts := map[string]FontConfig{
		"--font-primary-family":   t.Fonts.Primary,
		"--font-primary-size":     t.Fonts.Primary,
		"--font-primary-weight":   t.Fonts.Primary,
		"--font-primary-style":    t.Fonts.Primary,
		"--font-monospace-family":  t.Fonts.Monospace,
		"--font-monospace-size":    t.Fonts.Monospace,
	"--font-monospace-weight":  t.Fonts.Monospace,
		"--font-monospace-style":   t.Fonts.Monospace,
	}

	for name, font := range fonts {
		fmt.Fprintf(css, "  %s: %s;\n", name, font.Family)
		if name == "--font-primary-size" {
			fmt.Fprintf(css, "  --font-primary-weight: %s;\n", font.Weight)
			fmt.Fprintf(css, "  --font-primary-style: %s;\n", font.Style)
		}
	}
}

func (t *Theme) generateEffectCSS(css *strings.Builder) {
	effects := map[string]string{
		"--effect-border-radius": t.Effects.BorderRadius,
		"--effect-box-shadow":    t.Effects.BoxShadow,
		"--effect-text-shadow":   t.Effects.TextShadow,
		"--effect-blur":          t.Effects.Blur,
		"--effect-glow":          t.Effects.Glow,
	}

	for name, value := range effects {
		fmt.Fprintf(css, "  %s: %s;\n", name, value)
	}

	if t.Effects.Animation != nil {
		fmt.Fprintf(css, "  --animation-duration: %s;\n", t.Effects.Animation.Duration)
		fmt.Fprintf(css, "  --animation-easing: %s;\n", t.Effects.Animation.Easing)
		fmt.Fprintf(css, "  --animation-delay: %s;\n", t.Effects.Animation.Delay)
		fmt.Fprintf(css, "  --animation-count: %s;\n", t.Effects.Animation.Count)
	}
}

func (t *Theme) generateVariableCSS(css *strings.Builder) {
	variables := t.Settings.Variables

	for category, vars := range map[string]map[string]string{
		"border":   variables.BorderWidth,
		"spacing":  variables.Spacing,
		"radius":   variables.BorderRadius,
		"opacity":  variables.Opacity,
		"z-index":  variables.ZIndex,
	} {
		for name, value := range vars {
			varName := fmt.Sprintf("--var-%s-%s", category, name)
			fmt.Fprintf(css, "  %s: %s;\n", varName, value)
		}
	}

	for name, value := range variables.Custom {
		varName := fmt.Sprintf("--var-custom-%s", name)
		fmt.Fprintf(css, "  %s: %s;\n", varName, value)
	}
}

// String returns a string representation of the theme
func (t *Theme) String() string {
	return fmt.Sprintf("Theme{ID:%s, Name:%s, Category:%s}", t.ID, t.Name, t.Category)
}