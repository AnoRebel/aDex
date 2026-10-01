package font

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aDex/internal/models"
)

// FontValidator handles validation of font configurations and files
type FontValidator struct {
	knownFontFamilies map[string]bool
	validWeights      map[string]bool
	validHints        map[models.FontHinting]bool
}

// NewFontValidator creates a new font validator
func NewFontValidator() *FontValidator {
	return &FontValidator{
		knownFontFamilies: map[string]bool{
			"JetBrains Mono":   true,
			"Fira Code":        true,
			"Source Code Pro":  true,
			"Cascadia Code":    true,
			"IBM Plex Mono":    true,
			"Ubuntu Mono":      true,
			"Consolas":         true,
			"Monaco":           true,
			"Courier New":      true,
			"Lucida Console":   true,
			"monospace":        true,
			"Anonymous Pro":    true,
			"Inconsolata":      true,
			"Space Mono":       true,
			"Victor Mono":      true,
			"Iosevka":          true,
			"DejaVu Sans Mono": true,
			"Liberation Mono":  true,
		},
		validWeights: map[string]bool{
			string(models.FontWeightThin):       true,
			string(models.FontWeightExtraLight): true,
			string(models.FontWeightLight):      true,
			string(models.FontWeightNormal):     true,
			string(models.FontWeightMedium):     true,
			string(models.FontWeightSemiBold):   true,
			string(models.FontWeightBold):       true,
			string(models.FontWeightExtraBold):  true,
			string(models.FontWeightBlack):      true,
			"normal":                            true,
			"bold":                              true,
		},
		validHints: map[models.FontHinting]bool{
			models.FontHintingNone:   true,
			models.FontHintingSlight: true,
			models.FontHintingMedium: true,
			models.FontHintingFull:   true,
		},
	}
}

// ValidateConfiguration validates a font configuration
func (fv *FontValidator) ValidateConfiguration(config *models.FontConfiguration) *models.FontValidationResult {
	result := &models.FontValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
		Features:    make([]string, 0),
		Limitations: make([]string, 0),
	}

	// Validate required fields
	if config.Family == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font family is required")
	}

	if config.Name == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font name is required")
	}

	// Validate font family
	if config.Family != "" && !fv.knownFontFamilies[config.Family] {
		result.Warnings = append(result.Warnings, fmt.Sprintf("'%s' is not a known font family", config.Family))
		result.Suggestions = append(result.Suggestions, "Consider using a well-supported monospace font")
	}

	// Validate font size
	if config.Size < 6 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font size must be at least 6px")
	} else if config.Size > 72 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font size must not exceed 72px")
	} else if config.Size < 10 {
		result.Warnings = append(result.Warnings, "Font size below 10px may be difficult to read")
	} else if config.Size > 20 {
		result.Warnings = append(result.Warnings, "Font size above 20px may use excessive screen space")
	}

	// Validate line height
	if config.LineHeight < 0.8 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Line height must be at least 0.8")
	} else if config.LineHeight > 3.0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Line height must not exceed 3.0")
	} else if config.LineHeight < 1.0 {
		result.Warnings = append(result.Warnings, "Line height below 1.0 may cause overlapping lines")
	} else if config.LineHeight > 1.8 {
		result.Warnings = append(result.Warnings, "Line height above 1.8 may create excessive spacing")
	}

	// Validate letter spacing
	if config.LetterSpacing < -2 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Letter spacing must be at least -2px")
	} else if config.LetterSpacing > 10 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Letter spacing must not exceed 10px")
	} else if config.LetterSpacing < -0.5 {
		result.Warnings = append(result.Warnings, "Negative letter spacing may cause character overlap")
	} else if config.LetterSpacing > 2 {
		result.Warnings = append(result.Warnings, "High letter spacing may reduce readability")
	}

	// Validate font weight
	if config.Weight != "" && !fv.validWeights[config.Weight] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid font weight: %s", config.Weight))
		result.Suggestions = append(result.Suggestions, "Use standard font weights (100-900, normal, bold)")
	}

	// Validate hinting
	if !fv.validHints[config.Hinting] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid font hinting: %s", config.Hinting))
		result.Suggestions = append(result.Suggestions, "Use standard hinting values (none, slight, medium, full)")
	}

	// Validate features
	for feature := range config.Features {
		if !fv.isValidFontFeature(feature) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Unknown font feature: %s", feature))
		}
	}

	// Collect features and limitations
	if config.Ligatures {
		result.Features = append(result.Features, "Ligatures enabled")
		if !fv.supportsLigatures(config.Family) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("'%s' may not support ligatures", config.Family))
		}
	}

	if config.Antialias {
		result.Features = append(result.Features, "Antialiasing enabled")
	}

	switch config.Hinting {
	case models.FontHintingNone:
		result.Features = append(result.Features, "No font hinting")
		result.Limitations = append(result.Limitations, "Text may appear blurry on low-resolution displays")
	case models.FontHintingSlight:
		result.Features = append(result.Features, "Slight font hinting")
	case models.FontHintingMedium:
		result.Features = append(result.Features, "Medium font hinting")
	case models.FontHintingFull:
		result.Features = append(result.Features, "Full font hinting")
		result.Limitations = append(result.Limitations, "Text may appear too sharp on high-resolution displays")
	}

	// Performance considerations
	if config.Size > 16 && config.Ligatures {
		result.Warnings = append(result.Warnings, "Large fonts with ligatures may impact performance")
	}

	return result
}

// ValidateFontSource validates a font file from a source
func (fv *FontValidator) ValidateFontSource(source string) (*models.FontValidationResult, error) {
	result := &models.FontValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
		Features:    make([]string, 0),
		Limitations: make([]string, 0),
	}

	// Check if source is a file path
	if _, err := os.Stat(source); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Font file not found: %s", source))
		return result, err
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(source))
	if !fv.isValidFontFormat(ext) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported font format: %s", ext))
		result.Suggestions = append(result.Suggestions, "Use supported formats: TTF, OTF, WOFF, WOFF2")
		return result, nil
	}

	// Check file size
	info, err := os.Stat(source)
	if err == nil {
		const maxSize = 50 * 1024 * 1024 // 50MB
		if info.Size() > maxSize {
			result.Warnings = append(result.Warnings, "Font file is very large and may impact performance")
		} else if info.Size() < 1024 {
			result.Warnings = append(result.Warnings, "Font file is unusually small and may be incomplete")
		}
	}

	result.Features = append(result.Features, fmt.Sprintf("Valid font format: %s", strings.ToUpper(ext[1:])))

	return result, nil
}

// ValidateFontSettings validates font settings
func (fv *FontValidator) ValidateFontSettings(settings *models.FontSettings) *models.FontValidationResult {
	result := &models.FontValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
		Features:    make([]string, 0),
		Limitations: make([]string, 0),
	}

	// Validate default font ID
	if settings.DefaultFontID == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Default font ID is required")
	}

	// Validate font size
	if settings.FontSize < 6 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font size must be at least 6px")
	} else if settings.FontSize > 72 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Font size must not exceed 72px")
	}

	// Validate line height
	if settings.LineHeight < 0.8 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Line height must be at least 0.8")
	} else if settings.LineHeight > 3.0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Line height must not exceed 3.0")
	}

	// Validate letter spacing
	if settings.LetterSpacing < -2 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Letter spacing must be at least -2px")
	} else if settings.LetterSpacing > 10 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Letter spacing must not exceed 10px")
	}

	// Validate hinting
	if !fv.validHints[settings.Hinting] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid font hinting: %s", settings.Hinting))
	}

	// Validate font directories
	for _, dir := range settings.FontDirectories {
		if dir == "" {
			result.Warnings = append(result.Warnings, "Empty font directory path found")
		} else if !strings.HasPrefix(dir, "/") && !strings.HasPrefix(dir, "~/") {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Font directory path should be absolute: %s", dir))
		}
	}

	return result
}

// Check if a font feature is valid
func (fv *FontValidator) isValidFontFeature(feature string) bool {
	validFeatures := map[string]bool{
		"liga": true, // Standard ligatures
		"dlig": true, // Discretionary ligatures
		"hlig": true, // Historical ligatures
		"clig": true, // Contextual ligatures
		"zero": true, // Slashed zero
		"onum": true, // Old-style numerals
		"lnum": true, // Lining numerals
		"tnum": true, // Tabular numerals
		"pnum": true, // Proportional numerals
		"frac": true, // Fractions
		"afrc": true, // Alternative fractions
		"sups": true, // Superiors
		"subs": true, // Subscripts
		"ordn": true, // Ordinals
		"case": true, // Case-sensitive forms
		"cpsp": true, // Capital spacing
		"smcp": true, // Small capitals
		"c2sc": true, // Small capitals from capitals
		"ss01": true, // Stylistic set 1
		"ss02": true, // Stylistic set 2
		"ss03": true, // Stylistic set 3
		"ss04": true, // Stylistic set 4
		"ss05": true, // Stylistic set 5
		"ss06": true, // Stylistic set 6
		"ss07": true, // Stylistic set 7
		"ss08": true, // Stylistic set 8
		"ss09": true, // Stylistic set 9
		"ss10": true, // Stylistic set 10
		"ss11": true, // Stylistic set 11
		"ss12": true, // Stylistic set 12
		"ss13": true, // Stylistic set 13
		"ss14": true, // Stylistic set 14
		"ss15": true, // Stylistic set 15
		"ss16": true, // Stylistic set 16
		"ss17": true, // Stylistic set 17
		"ss18": true, // Stylistic set 18
		"ss19": true, // Stylistic set 19
		"ss20": true, // Stylistic set 20
		"kern": true, // Kerning
		"mark": true, // Mark positioning
		"mkmk": true, // Mark-to-mark positioning
		"calt": true, // Contextual alternates
		"hist": true, // Historical forms
		"salt": true, // Stylistic alternates
		"nalt": true, // Alternate annotation forms
		"dalt": true, // Alternate forms
	}

	return validFeatures[feature]
}

// Check if a font format is valid
func (fv *FontValidator) isValidFontFormat(ext string) bool {
	validFormats := map[string]bool{
		".ttf":   true,
		".otf":   true,
		".woff":  true,
		".woff2": true,
		".eot":   true,
		".svg":   true,
	}

	return validFormats[ext]
}

// Check if a font family supports ligatures
func (fv *FontValidator) supportsLigatures(family string) bool {
	ligatureFonts := map[string]bool{
		"Fira Code":       true,
		"JetBrains Mono":  true,
		"Cascadia Code":   true,
		"Iosevka":         true,
		"Victor Mono":     true,
		"Anonymous Pro":   true,
		"Inconsolata":     true,
		"Source Code Pro": false, // Limited ligature support
		"IBM Plex Mono":   false,
		"Ubuntu Mono":     false,
		"Consolas":        false,
		"Monaco":          false,
		"Courier New":     false,
		"Lucida Console":  false,
		"monospace":       false, // Generic fallback
	}

	return ligatureFonts[family]
}

// GetRecommendedSettings returns recommended font settings for different use cases
func (fv *FontValidator) GetRecommendedSettings(useCase string) *models.FontSettings {
	switch useCase {
	case "coding":
		return &models.FontSettings{
			DefaultFontID:   "jetbrains-mono",
			FallbackFont:    "monospace",
			FontSize:        14,
			LineHeight:      1.4,
			LetterSpacing:   0,
			EnableLigatures: true,
			EnableAntialias: true,
			Hinting:         models.FontHintingSlight,
			FontFeatureSettings: map[string]bool{
				"liga": true,
				"dlig": true,
				"zero": true,
			},
			AllowCustomFonts: false,
			AutoDetectFonts:  true,
		}
	case "accessibility":
		return &models.FontSettings{
			DefaultFontID:   "source-code-pro",
			FallbackFont:    "monospace",
			FontSize:        16,
			LineHeight:      1.6,
			LetterSpacing:   0.5,
			EnableLigatures: false, // Can be confusing for screen readers
			EnableAntialias: true,
			Hinting:         models.FontHintingMedium,
			FontFeatureSettings: map[string]bool{
				"kern": true,
			},
			AllowCustomFonts: false,
			AutoDetectFonts:  true,
		}
	case "performance":
		return &models.FontSettings{
			DefaultFontID:       "consolas",
			FallbackFont:        "monospace",
			FontSize:            12,
			LineHeight:          1.2,
			LetterSpacing:       0,
			EnableLigatures:     false,                  // Performance impact
			EnableAntialias:     false,                  // Performance impact
			Hinting:             models.FontHintingNone, // Performance impact
			FontFeatureSettings: map[string]bool{},
			AllowCustomFonts:    false,
			AutoDetectFonts:     false, // Skip scanning
		}
	default:
		return models.NewDefaultFontSettings()
	}
}

// SuggestFontImprovements provides suggestions for improving a font configuration
func (fv *FontValidator) SuggestFontImprovements(config *models.FontConfiguration) []string {
	suggestions := make([]string, 0)

	// Check for common improvements
	if config.Size < 12 {
		suggestions = append(suggestions, "Consider increasing font size to at least 12px for better readability")
	}

	if config.LineHeight < 1.2 {
		suggestions = append(suggestions, "Consider increasing line height to at least 1.2 for better readability")
	}

	if config.Ligatures && !fv.supportsLigatures(config.Family) {
		suggestions = append(suggestions, fmt.Sprintf("Consider using a font with better ligature support (e.g., Fira Code, JetBrains Mono)"))
	}

	if !config.Antialias {
		suggestions = append(suggestions, "Consider enabling antialiasing for smoother text rendering")
	}

	if config.Hinting == models.FontHintingNone && config.Size > 14 {
		suggestions = append(suggestions, "Consider using slight or medium hinting for larger fonts")
	}

	if config.LetterSpacing < 0 {
		suggestions = append(suggestions, "Consider positive letter spacing for better character separation")
	}

	return suggestions
}
