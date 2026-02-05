package theme

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aDex-UI/internal/models"
)

// LegacyThemeConverter handles conversion from legacy eDEX-UI themes to new format
type LegacyThemeConverter struct {
	// Configuration for conversion rules
	config *ConversionConfig
}

// ConversionConfig contains configuration for theme conversion
type ConversionConfig struct {
	StrictMode       bool                            `json:"strictMode"`
	PreserveOriginal bool                            `json:"preserveOriginal"`
	AutoFixColors    bool                            `json:"autoFixColors"`
	GeneratePreview  bool                            `json:"generatePreview"`
	ImportMetadata   bool                            `json:"importMetadata"`
	CategoryMapping  map[string]models.ThemeCategory `json:"categoryMapping"`
}

// DefaultConversionConfig returns default conversion configuration
func DefaultConversionConfig() *ConversionConfig {
	return &ConversionConfig{
		StrictMode:       false,
		PreserveOriginal: true,
		AutoFixColors:    true,
		GeneratePreview:  true,
		ImportMetadata:   true,
		CategoryMapping: map[string]models.ThemeCategory{
			"dark":     models.ThemeCategoryDark,
			"light":    models.ThemeCategoryLight,
			"colorful": models.ThemeCategoryColorful,
			"minimal":  models.ThemeCategoryMinimal,
			"retro":    models.ThemeCategoryRetro,
			"cyber":    models.ThemeCategoryCyberpunk,
			"custom":   models.ThemeCategoryCustom,
		},
	}
}

// NewLegacyThemeConverter creates a new theme converter
func NewLegacyThemeConverter(config *ConversionConfig) *LegacyThemeConverter {
	if config == nil {
		config = DefaultConversionConfig()
	}
	return &LegacyThemeConverter{
		config: config,
	}
}

// ConvertFromLegacy converts a legacy theme to the new format
func (c *LegacyThemeConverter) ConvertFromLegacy(legacy *models.LegacyTheme) (*models.Theme, error) {
	if legacy == nil {
		return nil, fmt.Errorf("legacy theme cannot be nil")
	}

	// Generate theme ID
	themeID := c.generateThemeID(legacy.Name)

	// Create new theme
	theme := models.NewTheme(themeID, legacy.Name)

	// Set basic metadata
	theme.Description = legacy.Description
	if legacy.Version != "" {
		theme.Version = legacy.Version
	}

	// Parse creation date if available
	if legacy.Created != "" {
		if createdAt, err := c.parseDate(legacy.Created); err == nil {
			theme.CreatedAt = createdAt
			theme.UpdatedAt = createdAt
		}
	}

	// Convert colors
	if err := c.convertColors(legacy.Colors, &theme.Colors); err != nil {
		if c.config.StrictMode {
			return nil, fmt.Errorf("color conversion failed: %w", err)
		}
		// Log warning for non-strict mode
	}

	// Convert fonts
	if err := c.convertFonts(legacy.Fonts, &theme.Fonts); err != nil {
		if c.config.StrictMode {
			return nil, fmt.Errorf("font conversion failed: %w", err)
		}
	}

	// Convert effects
	if err := c.convertEffects(legacy.Effects, &theme.Effects); err != nil {
		if c.config.StrictMode {
			return nil, fmt.Errorf("effects conversion failed: %w", err)
		}
	}

	// Convert settings
	if err := c.convertSettings(legacy.Settings, &theme.Settings); err != nil {
		if c.config.StrictMode {
			return nil, fmt.Errorf("settings conversion failed: %w", err)
		}
	}

	// Detect theme category
	theme.Category = c.detectThemeCategory(legacy)

	// Generate tags
	theme.Tags = c.generateTags(legacy)

	// Set as built-in if it's a known legacy theme
	theme.IsBuiltIn = c.isKnownLegacyTheme(legacy.Name)

	// Generate preview if requested
	if c.config.GeneratePreview {
		theme.Preview = c.generatePreview(theme)
	}

	return theme, nil
}

// ConvertFromLegacyJSON converts legacy theme from JSON data
func (c *LegacyThemeConverter) ConvertFromLegacyJSON(jsonData []byte) (*models.Theme, error) {
	var legacy models.LegacyTheme
	if err := json.Unmarshal(jsonData, &legacy); err != nil {
		return nil, fmt.Errorf("failed to parse legacy theme JSON: %w", err)
	}

	return c.ConvertFromLegacy(&legacy)
}

// BatchConvertThemes converts multiple legacy themes
func (c *LegacyThemeConverter) BatchConvertThemes(legacyThemes []*models.LegacyTheme) ([]*models.Theme, error) {
	var themes []*models.Theme
	var errors []error

	for i, legacy := range legacyThemes {
		theme, err := c.ConvertFromLegacy(legacy)
		if err != nil {
			errors = append(errors, fmt.Errorf("theme %d (%s): %w", i+1, legacy.Name, err))
			continue
		}
		themes = append(themes, theme)
	}

	if len(errors) > 0 && c.config.StrictMode {
		return nil, fmt.Errorf("batch conversion failed with %d errors: %v", len(errors), errors)
	}

	return themes, nil
}

// Helper methods

func (c *LegacyThemeConverter) generateThemeID(name string) string {
	// Convert name to a valid ID format
	id := strings.ToLower(name)
	id = regexp.MustCompile(`[^a-z0-9\-_]`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")

	// Ensure ID is not empty
	if id == "" {
		id = "theme-unknown"
	}

	return id
}

func (c *LegacyThemeConverter) parseDate(dateStr string) (time.Time, error) {
	// Try various date formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (c *LegacyThemeConverter) convertColors(legacyColors map[string]interface{}, colors *models.ThemeColors) error {
	// Map legacy color names to new theme structure
	colorMappings := map[string]func(models.ColorValue) error{
		"primary": func(cv models.ColorValue) error {
			colors.Primary = cv
			return nil
		},
		"primaryHover": func(cv models.ColorValue) error {
			colors.PrimaryHover = cv
			return nil
		},
		"primaryActive": func(cv models.ColorValue) error {
			colors.PrimaryActive = cv
			return nil
		},
		"secondary": func(cv models.ColorValue) error {
			colors.Secondary = cv
			return nil
		},
		"accent": func(cv models.ColorValue) error {
			colors.Accent = cv
			return nil
		},
		"background": func(cv models.ColorValue) error {
			colors.Background = cv
			return nil
		},
		"surface": func(cv models.ColorValue) error {
			colors.Surface = cv
			return nil
		},
		"text": func(cv models.ColorValue) error {
			colors.TextPrimary = cv
			return nil
		},
		"textSecondary": func(cv models.ColorValue) error {
			colors.TextSecondary = cv
			return nil
		},
		"success": func(cv models.ColorValue) error {
			colors.Success = cv
			return nil
		},
		"warning": func(cv models.ColorValue) error {
			colors.Warning = cv
			return nil
		},
		"error": func(cv models.ColorValue) error {
			colors.Error = cv
			return nil
		},
		"glitch": func(cv models.ColorValue) error {
			colors.Glitch = cv
			return nil
		},
		"cursor": func(cv models.ColorValue) error {
			colors.Cursor = cv
			return nil
		},
		"selection": func(cv models.ColorValue) error {
			colors.Selection = cv
			return nil
		},
	}

	for legacyName, value := range legacyColors {
		if converter, exists := colorMappings[legacyName]; exists {
			colorValue, err := c.convertColorValue(value)
			if err != nil {
				return fmt.Errorf("failed to convert color %s: %w", legacyName, err)
			}

			if err := converter(colorValue); err != nil {
				return fmt.Errorf("failed to set color %s: %w", legacyName, err)
			}
		}
	}

	return nil
}

func (c *LegacyThemeConverter) convertColorValue(value interface{}) (models.ColorValue, error) {
	var colorStr string

	switch v := value.(type) {
	case string:
		colorStr = v
	case float64:
		colorStr = c.hexFromInt(int(v))
	case int:
		colorStr = c.hexFromInt(v)
	default:
		return models.ColorValue{}, fmt.Errorf("unsupported color type: %T", v)
	}

	// Clean up the color string
	colorStr = strings.TrimSpace(colorStr)
	if !strings.HasPrefix(colorStr, "#") {
		colorStr = "#" + colorStr
	}

	// Validate and fix color if needed
	if !c.isValidColor(colorStr) {
		if c.config.AutoFixColors {
			colorStr = c.fixColor(colorStr)
		} else {
			return models.ColorValue{}, fmt.Errorf("invalid color format: %s", colorStr)
		}
	}

	// Determine color type
	colorType := c.detectColorType(colorStr)

	return models.ColorValue{
		Value: colorStr,
		Type:  colorType,
	}, nil
}

func (c *LegacyThemeConverter) convertFonts(legacyFonts map[string]interface{}, fonts *models.ThemeFonts) error {
	// Default fonts if not specified
	if len(legacyFonts) == 0 {
		// Use default fonts
		return nil
	}

	// Convert primary font
	if primary, exists := legacyFonts["primary"]; exists {
		fontConfig, err := c.convertFontConfig(primary)
		if err != nil {
			return fmt.Errorf("failed to convert primary font: %w", err)
		}
		fonts.Primary = fontConfig
	}

	// Convert monospace font
	if monospace, exists := legacyFonts["monospace"]; exists {
		fontConfig, err := c.convertFontConfig(monospace)
		if err != nil {
			return fmt.Errorf("failed to convert monospace font: %w", err)
		}
		fonts.Monospace = fontConfig
	}

	// Convert UI font
	if ui, exists := legacyFonts["ui"]; exists {
		fontConfig, err := c.convertFontConfig(ui)
		if err != nil {
			return fmt.Errorf("failed to convert UI font: %w", err)
		}
		fonts.UI = fontConfig
	}

	return nil
}

func (c *LegacyThemeConverter) convertFontConfig(font interface{}) (models.FontConfig, error) {
	var fontStr string

	switch v := font.(type) {
	case string:
		fontStr = v
	default:
		return models.FontConfig{}, fmt.Errorf("invalid font type: %T", v)
	}

	// Parse font string
	parts := strings.Split(fontStr, " ")
	if len(parts) < 2 {
		return models.FontConfig{
			Family: fontStr,
			Size:   "14px",
			Weight: "400",
			Style:  "normal",
		}, nil
	}

	// Extract font family (last part)
	family := strings.Join(parts[1:], " ")

	// Extract size (first part if it contains size info)
	size := "14px"
	weight := "400"
	style := "normal"

	// Parse size
	if sizeMatch := regexp.MustCompile(`(\d+(?:\.\d+)?(?:px|em|rem|%)?)`).FindString(parts[0]); sizeMatch != "" {
		size = sizeMatch
	}

	// Parse weight
	if weightMatch := regexp.MustCompile(`(?:bold|light|regular|medium|normal|\d{3}|1\d{3})`).FindString(parts[0]); weightMatch != "" {
		weightMap := map[string]string{
			"thin":    "100",
			"light":   "300",
			"regular": "400",
			"normal":  "400",
			"medium":  "500",
			"bold":    "700",
			"black":   "900",
		}
		if mapped, exists := weightMap[weightMatch]; exists {
			weight = mapped
		} else {
			// Try to parse as number
			if num, err := strconv.Atoi(weightMatch); err == nil {
				weight = fmt.Sprintf("%d", num)
			}
		}
	}

	// Parse style
	if strings.Contains(parts[0], "italic") {
		style = "italic"
	}

	return models.FontConfig{
		Family:   family,
		Size:     size,
		Weight:   weight,
		Style:    style,
		Variants: []string{"400", "500", "600", "700"},
		Features: []string{"liga", "kern"},
	}, nil
}

func (c *LegacyThemeConverter) convertEffects(legacyEffects map[string]interface{}, effects *models.ThemeEffects) error {
	// Convert border radius
	if borderRadius, exists := legacyEffects["borderRadius"]; exists {
		if borderStr := c.toString(borderRadius); borderStr != "" {
			effects.BorderRadius = borderStr
		}
	}

	// Convert box shadow
	if boxShadow, exists := legacyEffects["boxShadow"]; exists {
		if shadowStr := c.toString(boxShadow); shadowStr != "" {
			effects.BoxShadow = shadowStr
		}
	}

	// Convert text shadow
	if textShadow, exists := legacyEffects["textShadow"]; exists {
		if shadowStr := c.toString(textShadow); shadowStr != "" {
			effects.TextShadow = shadowStr
		}
	}

	// Convert glow effect
	if glow, exists := legacyEffects["glow"]; exists {
		if glowStr := c.toString(glow); glowStr != "" {
			effects.Glow = glowStr
		}
	}

	// Convert animation settings
	if animation, exists := legacyEffects["animation"]; exists {
		if animationMap, ok := animation.(map[string]interface{}); ok {
			animConfig := &models.AnimationConfig{
				Duration: "0.2s",
				Easing:   "ease",
				Delay:    "0s",
				Count:    "1",
			}

			if duration, exists := animationMap["duration"]; exists {
				animConfig.Duration = c.toString(duration)
			}
			if easing, exists := animationMap["easing"]; exists {
				animConfig.Easing = c.toString(easing)
			}
			if delay, exists := animationMap["delay"]; exists {
				animConfig.Delay = c.toString(delay)
			}
			if count, exists := animationMap["count"]; exists {
				animConfig.Count = c.toString(count)
			}

			effects.Animation = animConfig
		}
	}

	return nil
}

func (c *LegacyThemeConverter) convertSettings(legacySettings map[string]interface{}, settings *models.ThemeSettings) error {
	// Set default values
	*settings = models.NewDefaultSettings()

	// Override with legacy settings
	if autoSave, exists := legacySettings["autoSave"]; exists {
		if boolVal, ok := autoSave.(bool); ok {
			settings.AutoSave = boolVal
		}
	}

	if transitions, exists := legacySettings["transitions"]; exists {
		if boolVal, ok := transitions.(bool); ok {
			settings.Transitions = boolVal
		}
	}

	if reduceMotion, exists := legacySettings["reduceMotion"]; exists {
		if boolVal, ok := reduceMotion.(bool); ok {
			settings.ReduceMotion = boolVal
		}
	}

	if animations, exists := legacySettings["animations"]; exists {
		if boolVal, ok := animations.(bool); ok {
			settings.Animations = boolVal
		}
	}

	// Convert custom variables
	if variables, exists := legacySettings["variables"]; exists {
		if varMap, ok := variables.(map[string]interface{}); ok {
			settings.Variables = c.convertVariables(varMap)
		}
	}

	// Convert custom CSS
	if customCSS, exists := legacySettings["customCSS"]; exists {
		if cssStr := c.toString(customCSS); cssStr != "" {
			settings.CustomCSS = cssStr
		}
	}

	return nil
}

func (c *LegacyThemeConverter) convertVariables(legacyVars map[string]interface{}) models.ThemeVariables {
	variables := models.NewDefaultVariables()

	// Convert border width variables
	if borderVars, exists := legacyVars["borderWidth"]; exists {
		if varMap, ok := borderVars.(map[string]interface{}); ok {
			for name, value := range varMap {
				if strVal := c.toString(value); strVal != "" {
					variables.BorderWidth[name] = strVal
				}
			}
		}
	}

	// Convert spacing variables
	if spacingVars, exists := legacyVars["spacing"]; exists {
		if varMap, ok := spacingVars.(map[string]interface{}); ok {
			for name, value := range varMap {
				if strVal := c.toString(value); strVal != "" {
					variables.Spacing[name] = strVal
				}
			}
		}
	}

	// Convert border radius variables
	if radiusVars, exists := legacyVars["borderRadius"]; exists {
		if varMap, ok := radiusVars.(map[string]interface{}); ok {
			for name, value := range varMap {
				if strVal := c.toString(value); strVal != "" {
					variables.BorderRadius[name] = strVal
				}
			}
		}
	}

	// Convert custom variables
	if customVars, exists := legacyVars["custom"]; exists {
		if varMap, ok := customVars.(map[string]interface{}); ok {
			for name, value := range varMap {
				if strVal := c.toString(value); strVal != "" {
					variables.Custom[name] = strVal
				}
			}
		}
	}

	return variables
}

func (c *LegacyThemeConverter) detectThemeCategory(legacy *models.LegacyTheme) models.ThemeCategory {
	// Try to detect category from name or description
	searchText := strings.ToLower(legacy.Name + " " + legacy.Description)

	// Check for specific keywords
	categoryKeywords := map[models.ThemeCategory][]string{
		models.ThemeCategoryDark:      {"dark", "black", "night", "midnight"},
		models.ThemeCategoryLight:     {"light", "white", "day", "bright"},
		models.ThemeCategoryColorful:  {"colorful", "rainbow", "vibrant", "neon"},
		models.ThemeCategoryMinimal:   {"minimal", "simple", "clean", "plain"},
		models.ThemeCategoryCyberpunk: {"cyber", "neon", "hacker", "matrix"},
		models.ThemeCategoryRetro:     {"retro", "vintage", "classic", "old-school"},
		models.ThemeCategoryCustom:    {"custom"},
	}

	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(searchText, keyword) {
				return category
			}
		}
	}

	// Default to dark theme for eDEX-UI
	return models.ThemeCategoryDark
}

func (c *LegacyThemeConverter) generateTags(legacy *models.LegacyTheme) []string {
	var tags []string

	// Add category tag
	tags = append(tags, string(c.detectThemeCategory(legacy)))

	// Add style-based tags
	searchText := strings.ToLower(legacy.Name + " " + legacy.Description)

	styleKeywords := map[string][]string{
		"cyber":    {"cyber", "hacker", "matrix", "tech", "sci-fi"},
		"minimal":  {"minimal", "simple", "clean", "basic"},
		"retro":    {"retro", "vintage", "classic", "old-school"},
		"colorful": {"colorful", "rainbow", "vibrant", "neon"},
		"animated": {"animated", "dynamic", "interactive"},
	}

	for tag, keywords := range styleKeywords {
		for _, keyword := range keywords {
			if strings.Contains(searchText, keyword) {
				tags = append(tags, tag)
				break
			}
		}
	}

	// Add feature tags
	if legacy.Fonts != nil && len(legacy.Fonts) > 0 {
		tags = append(tags, "fonts")
	}

	if legacy.Colors != nil && len(legacy.Colors) > 5 {
		tags = append(tags, "custom-colors")
	}

	if legacy.Effects != nil && len(legacy.Effects) > 0 {
		tags = append(tags, "effects")
	}

	return tags
}

func (c *LegacyThemeConverter) isKnownLegacyTheme(name string) bool {
	knownThemes := []string{
		"default", "eDEX-UI", "dark", "light", "cyber", "retro", "minimal",
	}

	searchName := strings.ToLower(name)
	for _, known := range knownThemes {
		if strings.Contains(searchName, known) {
			return true
		}
	}
	return false
}

func (c *LegacyThemeConverter) generatePreview(theme *models.Theme) *models.ThemePreview {
	// Create a simple preview with theme colors
	colors := []string{
		theme.Colors.Primary.Value,
		theme.Colors.Background.Value,
		theme.Colors.Accent.Value,
		theme.Colors.Success.Value,
		theme.Colors.Error.Value,
	}

	return &models.ThemePreview{
		Thumbnail:  "",
		Colors:     colors,
		Layout:     "default",
		Components: []string{"header", "sidebar", "content"},
	}
}

// Helper utility methods

func (c *LegacyThemeConverter) toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (c *LegacyThemeConverter) hexFromInt(value int) string {
	return fmt.Sprintf("#%06x", value&0xFFFFFF)
}

func (c *LegacyThemeConverter) isValidColor(color string) bool {
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

	return false
}

func (c *LegacyThemeConverter) detectColorType(color string) string {
	if strings.HasPrefix(color, "#") {
		return "hex"
	}
	if strings.HasPrefix(color, "rgb") {
		return "rgb"
	}
	if strings.HasPrefix(color, "rgba") {
		return "rgba"
	}
	if strings.HasPrefix(color, "hsl") {
		return "hsl"
	}
	if strings.HasPrefix(color, "hsla") {
		return "hsla"
	}
	return "named"
}

func (c *LegacyThemeConverter) fixColor(color string) string {
	// Remove any whitespace
	color = strings.TrimSpace(color)

	// Ensure it starts with #
	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}

	// Convert 3-digit hex to 6-digit
	if len(color) == 4 && strings.HasPrefix(color, "#") {
		return fmt.Sprintf("#%c%c%c%c%c%c", color[1], color[1], color[2], color[2], color[3], color[3])
	}

	// Ensure valid hex length
	if len(color) > 7 {
		return "#" + color[1:7]
	}

	return color
}

// ConversionResult contains the result of a conversion operation
type ConversionResult struct {
	Theme    *models.Theme       `json:"theme"`
	Warnings []string            `json:"warnings"`
	Errors   []string            `json:"errors"`
	Success  bool                `json:"success"`
	Metadata *ConversionMetadata `json:"metadata"`
}

// ConversionMetadata contains metadata about the conversion process
type ConversionMetadata struct {
	OriginalName    string        `json:"originalName"`
	ConvertedAt     time.Time     `json:"convertedAt"`
	ProcessingTime  time.Duration `json:"processingTime"`
	FieldsConverted []string      `json:"fieldsConverted"`
	FieldsSkipped   []string      `json:"fieldsSkipped"`
}

// ConvertWithResult converts a legacy theme and returns detailed results
func (c *LegacyThemeConverter) ConvertWithResult(legacy *models.LegacyTheme) *ConversionResult {
	startTime := time.Now()

	result := &ConversionResult{
		Success: false,
		Metadata: &ConversionMetadata{
			OriginalName:    legacy.Name,
			ConvertedAt:     startTime,
			ProcessingTime:  0,
			FieldsConverted: []string{},
			FieldsSkipped:   []string{},
		},
	}

	// Attempt conversion
	theme, err := c.ConvertFromLegacy(legacy)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.Success = false
	} else {
		result.Theme = theme
		result.Success = true

		// Track which fields were converted
		if legacy.Colors != nil && len(legacy.Colors) > 0 {
			result.Metadata.FieldsConverted = append(result.Metadata.FieldsConverted, "colors")
		}
		if legacy.Fonts != nil && len(legacy.Fonts) > 0 {
			result.Metadata.FieldsConverted = append(result.Metadata.FieldsConverted, "fonts")
		}
		if legacy.Effects != nil && len(legacy.Effects) > 0 {
			result.Metadata.FieldsConverted = append(result.Metadata.FieldsConverted, "effects")
		}
		if legacy.Settings != nil && len(legacy.Settings) > 0 {
			result.Metadata.FieldsConverted = append(result.Metadata.FieldsConverted, "settings")
		}
	}

	result.Metadata.ProcessingTime = time.Since(startTime)
	result.Metadata.ConvertedAt = time.Now()

	return result
}
