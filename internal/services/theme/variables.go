package theme

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aDex-UI/internal/models"
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
	Variables     string            `json:"variables"`
	ThemeID       string            `json:"theme_id"`
	GeneratedAt   time.Time         `json:"generated_at"`
	Hash          string            `json:"hash"`
	VariableCount int               `json:"variable_count"`
	Metadata      map[string]string `json:"metadata"`
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

// computeThemeHash computes a hash for the theme for caching purposes
func (vg *VariableGenerator) computeThemeHash(theme *models.Theme) string {
	data, _ := json.Marshal(theme)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}

// GenerateCSSVariables generates CSS custom properties from a theme
func (vg *VariableGenerator) GenerateCSSVariables(theme *models.Theme) (*GeneratedCSS, error) {
	if theme == nil {
		return nil, fmt.Errorf("theme cannot be nil")
	}

	themeHash := vg.computeThemeHash(theme)

	// Check cache first
	if cached, exists := vg.variableCache[theme.ID]; exists {
		// Verify theme hasn't changed by comparing hashes
		if cached.Hash == themeHash {
			return cached, nil
		}
	}

	// Cache the theme
	vg.themeCache[theme.ID] = theme

	// Generate CSS variables
	css := &GeneratedCSS{
		ThemeID:       theme.ID,
		GeneratedAt:   time.Now(),
		Hash:          themeHash,
		VariableCount: 0,
		Metadata:      make(map[string]string),
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

	// Close root block
	builder.WriteString("}\n")

	css.Variables = builder.String()

	// Cache the result
	vg.variableCache[theme.ID] = css

	return css, nil
}

// generateColorVariables generates CSS color variables
func (vg *VariableGenerator) generateColorVariables(builder *strings.Builder, colors models.ThemeColors, css *GeneratedCSS) {
	// Primary colors
	vg.addColorVariable(builder, "primary", colors.Primary.Value, css)
	vg.addColorVariable(builder, "primary-hover", colors.PrimaryHover.Value, css)
	vg.addColorVariable(builder, "primary-active", colors.PrimaryActive.Value, css)

	// Secondary colors
	vg.addColorVariable(builder, "secondary", colors.Secondary.Value, css)
	vg.addColorVariable(builder, "secondary-hover", colors.SecondaryHover.Value, css)
	vg.addColorVariable(builder, "secondary-active", colors.SecondaryActive.Value, css)

	// Accent colors
	vg.addColorVariable(builder, "accent", colors.Accent.Value, css)
	vg.addColorVariable(builder, "accent-hover", colors.AccentHover.Value, css)
	vg.addColorVariable(builder, "accent-active", colors.AccentActive.Value, css)

	// Surface colors
	vg.addColorVariable(builder, "background", colors.Background.Value, css)
	vg.addColorVariable(builder, "surface", colors.Surface.Value, css)
	vg.addColorVariable(builder, "surface-hover", colors.SurfaceHover.Value, css)
	vg.addColorVariable(builder, "surface-border", colors.SurfaceBorder.Value, css)

	// Text colors
	vg.addColorVariable(builder, "text-primary", colors.TextPrimary.Value, css)
	vg.addColorVariable(builder, "text-secondary", colors.TextSecondary.Value, css)
	vg.addColorVariable(builder, "text-tertiary", colors.TextTertiary.Value, css)
	vg.addColorVariable(builder, "text-inverse", colors.TextInverse.Value, css)

	// Status colors
	vg.addColorVariable(builder, "success", colors.Success.Value, css)
	vg.addColorVariable(builder, "warning", colors.Warning.Value, css)
	vg.addColorVariable(builder, "error", colors.Error.Value, css)
	vg.addColorVariable(builder, "info", colors.Info.Value, css)

	// Special colors
	vg.addColorVariable(builder, "glitch", colors.Glitch.Value, css)
	vg.addColorVariable(builder, "cursor", colors.Cursor.Value, css)
	vg.addColorVariable(builder, "selection", colors.Selection.Value, css)

	// Terminal colors
	if colors.Terminal != nil {
		vg.addColorVariable(builder, "terminal-bg", colors.Terminal.Background.Value, css)
		vg.addColorVariable(builder, "terminal-fg", colors.Terminal.Foreground.Value, css)
		vg.addColorVariable(builder, "terminal-cursor", colors.Terminal.Cursor.Value, css)
		vg.addColorVariable(builder, "terminal-selection", colors.Terminal.Selection.Value, css)
		vg.addColorVariable(builder, "terminal-black", colors.Terminal.Black.Value, css)
		vg.addColorVariable(builder, "terminal-red", colors.Terminal.Red.Value, css)
		vg.addColorVariable(builder, "terminal-green", colors.Terminal.Green.Value, css)
		vg.addColorVariable(builder, "terminal-yellow", colors.Terminal.Yellow.Value, css)
		vg.addColorVariable(builder, "terminal-blue", colors.Terminal.Blue.Value, css)
		vg.addColorVariable(builder, "terminal-magenta", colors.Terminal.Magenta.Value, css)
		vg.addColorVariable(builder, "terminal-cyan", colors.Terminal.Cyan.Value, css)
		vg.addColorVariable(builder, "terminal-white", colors.Terminal.White.Value, css)
		vg.addColorVariable(builder, "terminal-bright-black", colors.Terminal.BrightBlack.Value, css)
		vg.addColorVariable(builder, "terminal-bright-red", colors.Terminal.BrightRed.Value, css)
		vg.addColorVariable(builder, "terminal-bright-green", colors.Terminal.BrightGreen.Value, css)
		vg.addColorVariable(builder, "terminal-bright-yellow", colors.Terminal.BrightYellow.Value, css)
		vg.addColorVariable(builder, "terminal-bright-blue", colors.Terminal.BrightBlue.Value, css)
		vg.addColorVariable(builder, "terminal-bright-magenta", colors.Terminal.BrightMagenta.Value, css)
		vg.addColorVariable(builder, "terminal-bright-cyan", colors.Terminal.BrightCyan.Value, css)
		vg.addColorVariable(builder, "terminal-bright-white", colors.Terminal.BrightWhite.Value, css)
	}
}

// addColorVariable adds a color CSS variable
func (vg *VariableGenerator) addColorVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	if value == "" {
		value = "transparent"
	}
	vg.addVariable(builder, name, value, css)
}

// addVariable adds a CSS variable to the builder
func (vg *VariableGenerator) addVariable(builder *strings.Builder, name, value string, css *GeneratedCSS) {
	varName := fmt.Sprintf("%s-%s", vg.variablePrefix, name)
	if vg.minify {
		builder.WriteString(fmt.Sprintf("%s:%s;", varName, value))
	} else {
		builder.WriteString(fmt.Sprintf("  %s: %s;\n", varName, value))
	}
	css.VariableCount++
}

// ClearCache clears all cached CSS variables
func (vg *VariableGenerator) ClearCache() {
	vg.variableCache = make(map[string]*GeneratedCSS)
	vg.themeCache = make(map[string]*models.Theme)
}

// GetCachedCSS returns cached CSS for a theme if available
func (vg *VariableGenerator) GetCachedCSS(themeID string) (*GeneratedCSS, bool) {
	css, exists := vg.variableCache[themeID]
	return css, exists
}

// GetCacheInfo returns information about the variable cache
func (vg *VariableGenerator) GetCacheInfo() map[string]interface{} {
	return map[string]interface{}{
		"cache_size":      len(vg.variableCache),
		"themes_cached":   len(vg.themeCache),
		"minify":          vg.minify,
		"variable_prefix": vg.variablePrefix,
	}
}
