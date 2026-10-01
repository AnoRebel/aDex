package theme

import (
	"aDex-UI/internal/appdir"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"aDex-UI/internal/utils"
)

// Service handles theme management
type Service struct {
	platform     *utils.FeatureDetection
	themes       map[string]*Theme
	currentTheme string
	themesDir    string
	mu           sync.RWMutex
}

// NewService creates a new theme service instance
func NewService() *Service {
	platform := utils.DetectPlatform()
	themesDir := getThemesDir()

	service := &Service{
		platform:     platform,
		themes:       make(map[string]*Theme),
		currentTheme: "default-dark",
		themesDir:    themesDir,
	}

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		// If we can't create themes directory, use built-in themes only
		service.loadBuiltInThemes()
		return service
	}

	// Load built-in and custom themes
	service.loadBuiltInThemes()
	service.loadCustomThemes()

	return service
}

// getThemesDir returns the appropriate themes directory for the platform
func getThemesDir() string {
	// Per-platform branching lives in appdir now.
	return filepath.Join(appdir.Data(), "themes")
}

// loadBuiltInThemes loads the built-in themes
func (s *Service) loadBuiltInThemes() {
	// Default Dark Theme
	defaultDark := &Theme{
		ID:     "default-dark",
		Name:   "Default Dark",
		IsDark: true,
		Colors: map[string]string{
			"primary":        "#00ff00",
			"secondary":      "#00cc00",
			"accent":         "#00ff88",
			"background":     "#000000",
			"surface":        "#0a0a0a",
			"text":           "#00ff00",
			"text_secondary": "#00cc00",
			"error":          "#ff0000",
			"warning":        "#ffaa00",
			"success":        "#00ff00",
			"info":           "#00aaff",
		},
		Fonts: map[string]string{
			"monospace": "Consolas, Monaco, monospace",
			"sans":      "Arial, sans-serif",
		},
		Sizes: map[string]int{
			"small":  12,
			"medium": 14,
			"large":  16,
			"xlarge": 18,
		},
		Spacing: map[string]int{
			"xs": 4,
			"sm": 8,
			"md": 16,
			"lg": 24,
			"xl": 32,
		},
		BorderRadius: map[string]int{
			"small":  2,
			"medium": 4,
			"large":  8,
		},
		Shadows: map[string]string{
			"small":  "0 1px 3px rgba(0, 255, 0, 0.2)",
			"medium": "0 4px 6px rgba(0, 255, 0, 0.3)",
		},
	}
	s.themes["default-dark"] = defaultDark

	// Default Light Theme
	defaultLight := &Theme{
		ID:     "default-light",
		Name:   "Default Light",
		IsDark: false,
		Colors: map[string]string{
			"primary":        "#006600",
			"secondary":      "#004400",
			"accent":         "#008800",
			"background":     "#ffffff",
			"surface":        "#f5f5f5",
			"text":           "#000000",
			"text_secondary": "#333333",
			"error":          "#ff0000",
			"warning":        "#ff8800",
			"success":        "#00aa00",
			"info":           "#0066cc",
		},
		Fonts:        defaultDark.Fonts,
		Sizes:        defaultDark.Sizes,
		Spacing:      defaultDark.Spacing,
		BorderRadius: defaultDark.BorderRadius,
		Shadows: map[string]string{
			"small":  "0 1px 3px rgba(0, 0, 0, 0.2)",
			"medium": "0 4px 6px rgba(0, 0, 0, 0.3)",
		},
	}
	s.themes["default-light"] = defaultLight

	// Matrix Theme (inspired by original eDEX-UI)
	matrixTheme := &Theme{
		ID:     "matrix",
		Name:   "Matrix",
		IsDark: true,
		Colors: map[string]string{
			"primary":        "#00ff41",
			"secondary":      "#00cc33",
			"accent":         "#00ff88",
			"background":     "#000000",
			"surface":        "#0a0a0a",
			"text":           "#00ff41",
			"text_secondary": "#00cc33",
			"error":          "#ff0000",
			"warning":        "#ffaa00",
			"success":        "#00ff00",
			"info":           "#00aaff",
		},
		Fonts:        defaultDark.Fonts,
		Sizes:        defaultDark.Sizes,
		Spacing:      defaultDark.Spacing,
		BorderRadius: defaultDark.BorderRadius,
		Shadows:      defaultDark.Shadows,
	}
	s.themes["matrix"] = matrixTheme
}

// Theme represents a visual theme
type Theme struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	IsDark       bool              `json:"is_dark"`
	Colors       map[string]string `json:"colors"`
	Fonts        map[string]string `json:"fonts"`
	Sizes        map[string]int    `json:"sizes"`
	Spacing      map[string]int    `json:"spacing"`
	BorderRadius map[string]int    `json:"border_radius"`
	Shadows      map[string]string `json:"shadows"`
}

// ThemeColors represents color palette
type ThemeColors struct {
	Primary       string `json:"primary"`
	Secondary     string `json:"secondary"`
	Accent        string `json:"accent"`
	Background    string `json:"background"`
	Surface       string `json:"surface"`
	Text          string `json:"text"`
	TextSecondary string `json:"text_secondary"`
	Error         string `json:"error"`
	Warning       string `json:"warning"`
	Success       string `json:"success"`
	Info          string `json:"info"`
}

// loadCustomThemes loads custom themes from the themes directory
func (s *Service) loadCustomThemes() {
	entries, err := os.ReadDir(s.themesDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		themePath := filepath.Join(s.themesDir, entry.Name())
		if err := s.loadThemeFromFile(themePath); err != nil {
			// Log error but continue loading other themes
			continue
		}
	}
}

// loadThemeFromFile loads a theme from a JSON file
func (s *Service) loadThemeFromFile(themePath string) error {
	data, err := os.ReadFile(themePath)
	if err != nil {
		return err
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return err
	}

	s.themes[theme.ID] = &theme
	return nil
}

// GetAvailableThemes retrieves list of available themes
func (s *Service) GetAvailableThemes(ctx context.Context) ([]Theme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var themes []Theme
	for _, theme := range s.themes {
		themes = append(themes, *theme)
	}

	return themes, nil
}

// GetTheme retrieves a specific theme by ID
func (s *Service) GetTheme(ctx context.Context, themeID string) (*Theme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	theme, exists := s.themes[themeID]
	if !exists {
		return nil, fmt.Errorf("theme not found: %s", themeID)
	}

	// Return a copy to prevent external modification
	themeCopy := *theme
	return &themeCopy, nil
}

// GetCurrentTheme retrieves the currently active theme
func (s *Service) GetCurrentTheme(ctx context.Context) (*Theme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	theme, exists := s.themes[s.currentTheme]
	if !exists {
		// Fallback to default dark theme
		if defaultTheme, exists := s.themes["default-dark"]; exists {
			s.currentTheme = "default-dark"
			themeCopy := *defaultTheme
			return &themeCopy, nil
		}
		return nil, fmt.Errorf("no themes available")
	}

	// Return a copy
	themeCopy := *theme
	return &themeCopy, nil
}

// SetTheme sets the active theme
func (s *Service) SetTheme(ctx context.Context, themeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.themes[themeID]
	if !exists {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	s.currentTheme = themeID
	return nil
}

// CreateTheme creates a new custom theme
func (s *Service) CreateTheme(ctx context.Context, theme *Theme) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if theme.ID == "" {
		return fmt.Errorf("theme ID cannot be empty")
	}

	// Check if theme already exists
	if _, exists := s.themes[theme.ID]; exists {
		return fmt.Errorf("theme already exists: %s", theme.ID)
	}

	// Save to file
	themePath := filepath.Join(s.themesDir, theme.ID+".json")
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save theme file: %w", err)
	}

	// Add to themes
	s.themes[theme.ID] = theme
	return nil
}

// UpdateTheme updates an existing theme
func (s *Service) UpdateTheme(ctx context.Context, themeID string, theme *Theme) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't allow updating built-in themes
	if s.isBuiltinTheme(themeID) {
		return fmt.Errorf("cannot update built-in theme: %s", themeID)
	}

	// Check if theme exists
	if _, exists := s.themes[themeID]; !exists {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	// Save to file
	themePath := filepath.Join(s.themesDir, themeID+".json")
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save theme file: %w", err)
	}

	// Update theme
	s.themes[themeID] = theme
	return nil
}

// DeleteTheme deletes a theme
func (s *Service) DeleteTheme(ctx context.Context, themeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't allow deleting built-in themes
	if s.isBuiltinTheme(themeID) {
		return fmt.Errorf("cannot delete built-in theme: %s", themeID)
	}

	// Check if theme exists
	if _, exists := s.themes[themeID]; !exists {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	// If this is the current theme, switch to default
	if s.currentTheme == themeID {
		s.currentTheme = "default-dark"
	}

	// Delete file
	themePath := filepath.Join(s.themesDir, themeID+".json")
	os.Remove(themePath) // Ignore error if file doesn't exist

	// Remove from themes
	delete(s.themes, themeID)
	return nil
}

// isBuiltinTheme checks if a theme is a built-in theme
func (s *Service) isBuiltinTheme(themeID string) bool {
	builtinThemes := []string{"default-dark", "default-light", "matrix"}
	for _, builtin := range builtinThemes {
		if themeID == builtin {
			return true
		}
	}
	return false
}

// ExportTheme exports a theme to specified format
func (s *Service) ExportTheme(ctx context.Context, themeID string, format string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	theme, exists := s.themes[themeID]
	if !exists {
		return nil, fmt.Errorf("theme not found: %s", themeID)
	}

	switch format {
	case "json":
		return json.MarshalIndent(theme, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ImportTheme imports a theme from data
func (s *Service) ImportTheme(ctx context.Context, data []byte, format string) (*Theme, error) {
	var theme Theme

	switch format {
	case "json":
		if err := json.Unmarshal(data, &theme); err != nil {
			return nil, fmt.Errorf("failed to parse JSON theme: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}

	// Use CreateTheme to handle the import
	return &theme, s.CreateTheme(ctx, &theme)
}

// GenerateCSS generates CSS for a theme
func (s *Service) GenerateCSS(ctx context.Context, themeID string) (string, error) {
	theme, err := s.GetTheme(ctx, themeID)
	if err != nil {
		return "", err
	}

	var css strings.Builder

	// CSS Variables
	css.WriteString(":root {\n")
	for colorName, colorValue := range theme.Colors {
		css.WriteString(fmt.Sprintf("  --color-%s: %s;\n", colorName, colorValue))
	}
	for sizeName, sizeValue := range theme.Sizes {
		css.WriteString(fmt.Sprintf("  --size-%s: %dpx;\n", sizeName, sizeValue))
	}
	for spacingName, spacingValue := range theme.Spacing {
		css.WriteString(fmt.Sprintf("  --spacing-%s: %dpx;\n", spacingName, spacingValue))
	}
	css.WriteString("}\n\n")

	// Basic styling
	css.WriteString("body {\n")
	css.WriteString(fmt.Sprintf("  background-color: %s;\n", theme.Colors["background"]))
	css.WriteString(fmt.Sprintf("  color: %s;\n", theme.Colors["text"]))
	css.WriteString(fmt.Sprintf("  font-family: %s;\n", theme.Fonts["sans"]))
	css.WriteString("}\n")

	return css.String(), nil
}
