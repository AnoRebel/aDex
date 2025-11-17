package theme

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"adx/internal/models"
)

// VariableGenerator handles CSS custom property generation from themes
type VariableGenerator struct {
	// Cache for generated CSS variables
	variableCache map[string]*GeneratedCSS
	// Cache for processed themes
	themeCache map[string]*models.Theme
	// Whether to enable minification
	minify bool
	// Custom prefix for CSS variables
	variablePrefix string
}

// GeneratedCSS represents generated CSS with metadata
type GeneratedCSS struct {
	Variables    string            `json:"variables"`
	ThemeID      string            `json:"theme_id"`
	GeneratedAt  time.Time         `json:"generated_at"`
	Hash         string            `json:"hash"`
	VariableCount int              `json:"variable_count"`
	Metadata     map[string]string `json:"metadata"`
}

// NewVariableGenerator creates a new CSS variable generator
func NewVariableGenerator() *VariableGenerator {
	return &VariableGenerator{
		variableCache:  make(map[string]*GeneratedCSS),
		themeCache:     make(map[string]*models.Theme),
		minify:         false,
		variablePrefix: "--dex-theme",
	}
}

// SetMinify enables or disables CSS minification
func (vg *VariableGenerator) SetMinify(minify bool) {
	vg.minify = minify
	// Clear cache when minification setting changes
	vg.variableCache = make(map[string]*GeneratedCSS)
}

// SetVariablePrefix sets a custom prefix for CSS variables
func (vg *VariableGenerator) SetVariablePrefix(prefix string) {
	vg.variablePrefix = prefix
	// Clear cache when prefix changes
	vg.variableCache = make(map[string]*GeneratedCSS)
}

// GenerateCSSVariables generates CSS custom properties from a theme
func (vg *VariableGenerator) GenerateCSSVariables(theme *models.Theme) (*GeneratedCSS, error) {
	if theme == nil {
		return nil, fmt.Errorf("theme cannot be nil")
	}

	// Check cache first
	if cached, exists := vg.variableCache[theme.ID]; exists {
		// Verify theme hasn't changed by comparing hashes
		if cached.Hash == theme.Hash() {
			return cached, nil
		}
	}

	// Cache the theme
	vg.themeCache[theme.ID] = theme

	// Generate CSS variables
	css := &GeneratedCSS{
		ThemeID:      theme.ID,
		GeneratedAt:  time.Now(),
		Hash:         theme.Hash(),
		VariableCount: 0,
		Metadata:     make(map[string]string),
	}

	var builder strings.Builder

	// Add theme metadata comment
	if !vg.minify {
		builder.WriteString(fmt.Sprintf("/* Theme: %s - %s */\n", theme.Name, theme.Description))
		builder.WriteString(fmt.Sprintf("/* Generated: %s */\n\n", css.GeneratedAt.Format(time.RFC3339)))
	}

	// Generate root selector block
	builder.WriteString(":root {\n")

	// Color variables
	vg.generateColorVariables(&builder, theme.Colors, css)

	// Font variables
	vg.generateFontVariables(&builder, theme.Fonts, css)

	// Effect variables
	vg.generateEffectVariables(&builder, theme.Effects, css)

	// Settings variables
	vg.generateSettingsVariables(&builder, theme.Settings, css)

	// Close root block
	builder.WriteString("}\n")

	// Add utility classes if not minified
	if !vg.minify {
		vg.generateUtilityClasses(&builder, theme, css)
	}

	css.Variables = builder.String()

	// Cache the result
	vg.variableCache[theme.ID] = css

	return css, nil
}

// generateColorVariables generates CSS color variables
func (vg *VariableGenerator) generateColorVariables(builder *strings.Builder, colors models.ThemeColors, css *GeneratedCSS) {
	// Base colors
	vg.addColorVariable(builder, "background-primary", colors.Background.Primary, css)
	vg.addColorVariable(builder, "background-secondary", colors.Background.Secondary, css)
	vg.addColorVariable(builder, "background-tertiary", colors.Background.Tertiary, css)

	vg.addColorVariable(builder, "foreground-primary", colors.Foreground.Primary, css)
	vg.addColorVariable(builder, "foreground-secondary", colors.Foreground.Secondary, css)
	vg.addColorVariable(builder, "foreground-tertiary", colors.Foreground.Tertiary, css)

	// Accent colors
	vg.addColorVariable(builder, "accent-primary", colors.Accent.Primary, css)
	vg.addColorVariable(builder, "accent-secondary", colors.Accent.Secondary, css)

	// Status colors
	vg.addColorVariable(builder, "success", colors.Status.Success, css)
	vg.addColorVariable(builder, "warning", colors.Status.Warning, css)
	vg.addColorVariable(builder, "error", colors.Status.Error, css)
	vg.addColorVariable(builder, "info", colors.Status.Info, css)

	// Terminal colors
	for i, color := range colors.Terminal {
		vg.addColorVariable(builder, fmt.Sprintf("terminal-%d", i), color, css)
	}

	// UI component colors
	vg.addColorVariable(builder, "ui-button-bg", colors.UI.ButtonBackground, css)
	vg.addColorVariable(builder, "ui-button-fg", colors.UI.ButtonForeground, css)
	vg.addColorVariable(builder, "ui-button-hover", colors.UI.ButtonHover, css)
	vg.addColorVariable(builder, "ui-button-active", colors.UI.ButtonActive, css)

	vg.addColorVariable(builder, "ui-input-bg", colors.UI.InputBackground, css)
	vg.addColorVariable(builder, "ui-input-fg", colors.UI.InputForeground, css)
	vg.addColorVariable(builder, "ui-input-border", colors.UI.InputBorder, css)
	vg.addColorVariable(builder, "ui-input-focus", colors.UI.InputFocus, css)

	vg.addColorVariable(builder, "ui-border", colors.UI.Border, css)
	vg.addColorVariable(builder, "ui-shadow", colors.UI.Shadow, css)
}

// generateFontVariables generates CSS font variables
func (vg *VariableGenerator) generateFontVariables(builder *strings.Builder, fonts models.ThemeFonts, css *GeneratedCSS) {
	vg.addFontVariable(builder, "font-family-primary", fonts.Families.Primary, css)
	vg.addFontVariable(builder, "font-family-secondary", fonts.Families.Secondary, css)
	vg.addFontVariable(builder, "font-family-mono", fonts.Families.Monospace, css)

	vg.addSizeVariable(builder, "font-size-xs", fonts.Sizes.ExtraSmall, css)
	vg.addSizeVariable(builder, "font-size-sm", fonts.Sizes.Small, css)
	vg.addSizeVariable(builder, "font-size-base", fonts.Sizes.Base, css)
	vg.addSizeVariable(builder, "font-size-lg", fonts.Sizes.Large, css)
	vg.addSizeVariable(builder, "font-size-xl", fonts.Sizes.ExtraLarge, css)
	vg.addSizeVariable(builder, "font-size-2xl", fonts.Sizes.DoubleExtraLarge, css)

	vg.addWeightVariable(builder, "font-weight-light", fonts.Weights.Light, css)
	vg.addWeightVariable(builder, "font-weight-normal", fonts.Weights.Normal, css)
	vg.addWeightVariable(builder, "font-weight-medium", fonts.Weights.Medium, css)
	vg.addWeightVariable(builder, "font-weight-semibold", fonts.Weights.SemiBold, css)
	vg.addWeightVariable(builder, "font-weight-bold", fonts.Weights.Bold, css)

	vg.addSizeVariable(builder, "line-height-tight", fonts.LineHeights.Tight, css)
	vg.addSizeVariable(builder, "line-height-normal", fonts.LineHeights.Normal, css)
	vg.addSizeVariable(builder, "line-height-relaxed", fonts.LineHeights.Relaxed, css)

	vg.addSizeVariable(builder, "letter-spacing-tight", fonts.LetterSpacing.Tight, css)
	vg.addSizeVariable(builder, "letter-spacing-normal", fonts.LetterSpacing.Normal, css)
	vg.addSizeVariable(builder, "letter-spacing-wide", fonts.LetterSpacing.Wide, css)
}

// generateEffectVariables generates CSS effect variables
func (vg *VariableGenerator) generateEffectVariables(builder *strings.Builder, effects models.ThemeEffects, css *GeneratedCSS) {
	// Border radius
	vg.addSizeVariable(builder, "border-radius-sm", effects.BorderRadius.Small, css)
	vg.addSizeVariable(builder, "border-radius-md", effects.BorderRadius.Medium, css)
	vg.addSizeVariable(builder, "border-radius-lg", effects.BorderRadius.Large, css)
	vg.addSizeVariable(builder, "border-radius-xl", effects.BorderRadius.ExtraLarge, css)
	vg.addSizeVariable(builder, "border-radius-full", effects.BorderRadius.Full, css)

	// Shadows
	for i, shadow := range effects.Shadows {
		vg.addShadowVariable(builder, fmt.Sprintf("shadow-%d", i+1), shadow, css)
	}

	// Gradients
	for i, gradient := range effects.Gradients {
		vg.addGradientVariable(builder, fmt.Sprintf("gradient-%d", i+1), gradient, css)
	}

	// Animations
	vg.addAnimationVariable(builder, "animation-duration-fast", effects.Animations.Duration.Fast, css)
	vg.addAnimationVariable(builder, "animation-duration-normal", effects.Animations.Duration.Normal, css)
	vg.addAnimationVariable(builder, "animation-duration-slow", effects.Animations.Duration.Slow, css)

	vg.addAnimationVariable(builder, "animation-easing-linear", effects.Animations.Easing.Linear, css)
	vg.addAnimationVariable(builder, "animation-easing-ease-in", effects.Animations.Easing.EaseIn, css)
	vg.addAnimationVariable(builder, "animation-easing-ease-out", effects.Animations.Easing.EaseOut, css)
	vg.addAnimationVariable(builder, "animation-easing-ease-in-out", effects.Animations.Easing.EaseInOut, css)

	// Blur effects
	vg.addSizeVariable(builder, "blur-sm", effects.Blur.Small, css)
	vg.addSizeVariable(builder, "blur-md", effects.Blur.Medium, css)
	vg.addSizeVariable(builder, "blur-lg", effects.Blur.Large, css)
	vg.addSizeVariable(builder, "blur-xl", effects.Blur.ExtraLarge, css)
}

// generateSettingsVariables generates CSS settings variables
func (vg *VariableGenerator) generateSettingsVariables(builder *strings.Builder, settings models.ThemeSettings, css *GeneratedCSS) {
	vg.addOpacityVariable(builder, "opacity-disabled", settings.DisabledOpacity, css)
	vg.addOpacityVariable(builder, "opacity-hover", settings.HoverOpacity, css)
	vg.addOpacityVariable(builder, "opacity-active", settings.ActiveOpacity, css)

	vg.addTransitionVariable(builder, "transition-fast", settings.TransitionDuration.Fast, css)
	vg.addTransitionVariable(builder, "transition-normal", settings.TransitionDuration.Normal, css)
	vg.addTransitionVariable(builder, "transition-slow", settings.TransitionDuration.Slow, css)

	// Z-index scale
	for name, value := range settings.ZIndexScale {
		vg.addZIndexVariable(builder, fmt.Sprintf("z-%s", name), value, css)
	}
}

// Helper methods for adding CSS variables

func (vg *VariableGenerator) addColorVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "transparent"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addFontVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "inherit"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addSizeVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "0"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addWeightVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "400"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addShadowVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "none"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addGradientVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "transparent"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addAnimationVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "0s"
	}
	vg.addVariable(builder, name, value, css)
}

func (vg *VariableGenerator) addOpacityVariable(builder *strings.Builder, name string, value float32, css *GeneratedCSS) {
	vg.addVariable(builder, name, fmt.Sprintf("%.2f", value), css)
}

func (vg *VariableGenerator) addTransitionVariable(builder *strings.Builder, name string, value time.Duration, css *GeneratedCSS) {
	vg.addVariable(builder, name, fmt.Sprintf("%.0fms", float64(value.Milliseconds())), css)
}

func (vg *VariableGenerator) addZIndexVariable(builder *strings.Builder, name string, value int, css *GeneratedCSS) {
	vg.addVariable(builder, name, fmt.Sprintf("%d", value), css)
}

func (vg *VariableGenerator) addVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	indent := "  "
	if vg.minify {
		indent = ""
	}

	variableName := fmt.Sprintf("%s-%s", vg.variablePrefix, name)

	// Sanitize the variable name
	variableName = vg.sanitizeVariableName(variableName)

	// Sanitize the value
	value = vg.sanitizeValue(value)

	builder.WriteString(fmt.Sprintf("%s%s: %s;\n", indent, variableName, value))
	css.VariableCount++
}

// generateUtilityClasses generates utility CSS classes for common theme values
func (vg *VariableGenerator) generateUtilityClasses(builder *strings.Builder, theme *models.Theme, css *GeneratedCSS) {
	builder.WriteString("\n/* Utility Classes */\n")

	// Text color utilities
	builder.WriteString(".text-primary { color: var(--dex-theme-foreground-primary); }\n")
	builder.WriteString(".text-secondary { color: var(--dex-theme-foreground-secondary); }\n")
	builder.WriteString(".text-accent { color: var(--dex-theme-accent-primary); }\n")

	// Background utilities
	builder.WriteString(".bg-primary { background-color: var(--dex-theme-background-primary); }\n")
	builder.WriteString(".bg-secondary { background-color: var(--dex-theme-background-secondary); }\n")

	// Border utilities
	builder.WriteString(".border-themed { border-color: var(--dex-theme-ui-border); }\n")
	builder.WriteString(".border-radius-themed { border-radius: var(--dex-theme-border-radius-md); }\n")

	// Font utilities
	builder.WriteString(".font-primary { font-family: var(--dex-theme-font-family-primary); }\n")
	builder.WriteString(".font-mono { font-family: var(--dex-theme-font-family-mono); }\n")
}

// sanitizeVariableName ensures CSS variable names are valid
func (vg *VariableGenerator) sanitizeVariableName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	name = reg.ReplaceAllString(name, "-")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	return name
}

// sanitizeValue ensures CSS values are safe
func (vg *VariableGenerator) sanitizeValue(value string) string {
	// Remove potentially unsafe characters
	value = strings.ReplaceAll(value, "<", "")
	value = strings.ReplaceAll(value, ">", "")
	value = strings.ReplaceAll(value, "&", "")
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "'", "")

	return strings.TrimSpace(value)
}

// GenerateCSSForThemeID generates CSS variables for a theme by ID
func (vg *VariableGenerator) GenerateCSSForThemeID(themeID string) (*GeneratedCSS, error) {
	theme, exists := vg.themeCache[themeID]
	if !exists {
		return nil, fmt.Errorf("theme with ID '%s' not found in cache", themeID)
	}

	return vg.GenerateCSSVariables(theme)
}

// GetCachedCSS retrieves cached CSS if available
func (vg *VariableGenerator) GetCachedCSS(themeID string) (*GeneratedCSS, bool) {
	css, exists := vg.variableCache[themeID]
	return css, exists
}

// ClearCache clears the CSS variable cache
func (vg *VariableGenerator) ClearCache() {
	vg.variableCache = make(map[string]*GeneratedCSS)
	vg.themeCache = make(map[string]*models.Theme)
}

// GetCacheInfo returns information about the current cache state
func (vg *VariableGenerator) GetCacheInfo() map[string]interface{} {
	return map[string]interface{}{
		"cached_themes":      len(vg.themeCache),
		"cached_css_files":   len(vg.variableCache),
		"minify_enabled":     vg.minify,
		"variable_prefix":    vg.variablePrefix,
	}
}

// GenerateCSSVariablesBatch generates CSS variables for multiple themes
func (vg *VariableGenerator) GenerateCSSVariablesBatch(themes []*models.Theme) ([]*GeneratedCSS, error) {
	if len(themes) == 0 {
		return nil, fmt.Errorf("no themes provided")
	}

	results := make([]*GeneratedCSS, 0, len(themes))

	for _, theme := range themes {
		css, err := vg.GenerateCSSVariables(theme)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CSS for theme '%s': %w", theme.ID, err)
		}
		results = append(results, css)
	}

	return results, nil
}