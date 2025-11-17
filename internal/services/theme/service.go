package theme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"adx/internal/models"
)

// Service is the main theme service that coordinates all theme operations
type Service struct {
	// Core components
	converter      *LegacyThemeConverter
	variableGen    *VariableGenerator

	// Storage
	themes         map[string]*models.Theme
	themeDirectory string

	// Runtime state
	currentThemeID string
	defaultThemeID string

	// Synchronization
	mutex sync.RWMutex

	// Configuration
	config ServiceConfig

	// Event handlers
	eventHandlers map[string][]func(*models.Theme)
}

// ServiceConfig holds configuration for the theme service
type ServiceConfig struct {
	ThemeDirectory     string        `json:"theme_directory"`
	DefaultTheme       string        `json:"default_theme"`
	AutoSave           bool          `json:"auto_save"`
	AutoReload         bool          `json:"auto_reload"`
	CacheEnabled       bool          `json:"cache_enabled"`
	VariablePrefix     string        `json:"variable_prefix"`
	MinifyCSS          bool          `json:"minify_css"`
	WatchInterval      time.Duration `json:"watch_interval"`
	MaxCacheSize       int           `json:"max_cache_size"`
	EnableLegacyImport bool          `json:"enable_legacy_import"`
}

// DefaultServiceConfig returns the default configuration for the theme service
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ThemeDirectory:     "themes",
		DefaultTheme:       "default-dark",
		AutoSave:           true,
		AutoReload:         true,
		CacheEnabled:       true,
		VariablePrefix:     "--dex-theme",
		MinifyCSS:          false,
		WatchInterval:      time.Second * 5,
		MaxCacheSize:       100,
		EnableLegacyImport: true,
	}
}

// NewService creates a new theme service with the given configuration
func NewService(config ServiceConfig) *Service {
	service := &Service{
		converter:      NewLegacyThemeConverter(),
		variableGen:    NewVariableGenerator(),
		themes:         make(map[string]*models.Theme),
		themeDirectory: config.ThemeDirectory,
		currentThemeID: config.DefaultTheme,
		defaultThemeID: config.DefaultTheme,
		config:         config,
		eventHandlers:  make(map[string][]func(*models.Theme)),
	}

	// Configure variable generator
	service.variableGen.SetMinify(config.MinifyCSS)
	service.variableGen.SetVariablePrefix(config.VariablePrefix)

	return service
}

// Initialize initializes the theme service
func (s *Service) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create theme directory if it doesn't exist
	if err := os.MkdirAll(s.themeDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	// Load built-in themes
	if err := s.loadBuiltinThemes(); err != nil {
		return fmt.Errorf("failed to load built-in themes: %w", err)
	}

	// Load themes from directory
	if err := s.loadThemesFromDirectory(); err != nil {
		return fmt.Errorf("failed to load themes from directory: %w", err)
	}

	// Ensure default theme exists
	if _, exists := s.themes[s.defaultThemeID]; !exists {
		if err := s.createDefaultTheme(); err != nil {
			return fmt.Errorf("failed to create default theme: %w", err)
		}
	}

	// Set current theme to default if not already set
	if s.currentThemeID == "" {
		s.currentThemeID = s.defaultThemeID
	}

	// Start background watcher if auto-reload is enabled
	if s.config.AutoReload {
		go s.startThemeWatcher()
	}

	return nil
}

// loadBuiltinThemes loads built-in themes
func (s *Service) loadBuiltinThemes() error {
	builtinThemes := s.getBuiltinThemes()

	for _, theme := range builtinThemes {
		s.themes[theme.ID] = theme
	}

	return nil
}

// getBuiltinThemes returns the built-in themes
func (s *Service) getBuiltinThemes() []*models.Theme {
	return []*models.Theme{
		s.createDefaultDarkTheme(),
		s.createDefaultLightTheme(),
		s.createCyberpunkTheme(),
		s.createRetroTheme(),
	}
}

// createDefaultTheme creates the default theme if it doesn't exist
func (s *Service) createDefaultTheme() error {
	theme := s.createDefaultDarkTheme()
	s.themes[theme.ID] = theme

	if s.config.AutoSave {
		return s.SaveTheme(theme)
	}

	return nil
}

// createDefaultDarkTheme creates the default dark theme
func (s *Service) createDefaultDarkTheme() *models.Theme {
	theme := models.NewTheme("default-dark", "Default Dark")
	theme.Description = "Default dark theme for aDex-UI"
	theme.Author = "aDex-UI Team"
	theme.Version = "1.0.0"

	// Set up dark theme colors
	theme.Colors = models.ThemeColors{
		Background: models.ColorPalette{
			Primary:   "#0a0a0a",
			Secondary: "#1a1a1a",
			Tertiary:  "#2a2a2a",
		},
		Foreground: models.ColorPalette{
			Primary:   "#ffffff",
			Secondary: "#cccccc",
			Tertiary:  "#999999",
		},
		Accent: models.AccentColors{
			Primary:   "#00ff41",
			Secondary: "#00cc33",
		},
		Status: models.StatusColors{
			Success: "#00ff41",
			Warning: "#ffaa00",
			Error:   "#ff3333",
			Info:    "#00aaff",
		},
		Terminal: models.TerminalColors{
			"#000000", "#ff0000", "#00ff00", "#ffff00",
			"#0000ff", "#ff00ff", "#00ffff", "#ffffff",
			"#808080", "#ff8080", "#80ff80", "#ffff80",
			"#8080ff", "#ff80ff", "#80ffff", "#c0c0c0",
		},
		UI: models.UIColors{
			ButtonBackground:   "#1a1a1a",
			ButtonForeground:   "#ffffff",
			ButtonHover:        "#2a2a2a",
			ButtonActive:       "#00ff41",
			InputBackground:    "#0a0a0a",
			InputForeground:    "#ffffff",
			InputBorder:        "#333333",
			InputFocus:         "#00ff41",
			Border:             "#333333",
			Shadow:             "rgba(0, 0, 0, 0.5)",
		},
	}

	// Set up fonts
	theme.Fonts = models.DefaultThemeFonts()

	// Set up effects
	theme.Effects = models.DefaultThemeEffects()

	// Set up settings
	theme.Settings = models.DefaultThemeSettings()

	return theme
}

// createDefaultLightTheme creates the default light theme
func (s *Service) createDefaultLightTheme() *models.Theme {
	theme := models.NewTheme("default-light", "Default Light")
	theme.Description = "Default light theme for aDex-UI"
	theme.Author = "aDex-UI Team"
	theme.Version = "1.0.0"

	// Set up light theme colors (inverse of dark)
	theme.Colors = models.ThemeColors{
		Background: models.ColorPalette{
			Primary:   "#ffffff",
			Secondary: "#f5f5f5",
			Tertiary:  "#e0e0e0",
		},
		Foreground: models.ColorPalette{
			Primary:   "#000000",
			Secondary: "#333333",
			Tertiary:  "#666666",
		},
		Accent: models.AccentColors{
			Primary:   "#0066cc",
			Secondary: "#0052a3",
		},
		Status: models.StatusColors{
			Success: "#00aa44",
			Warning: "#ff8800",
			Error:   "#cc0000",
			Info:    "#0088cc",
		},
		Terminal: models.TerminalColors{
			"#ffffff", "#cc0000", "#00cc00", "#cccc00",
			"#0000cc", "#cc00cc", "#00cccc", "#000000",
			"#808080", "#ff8080", "#80ff80", "#ffff80",
			"#8080ff", "#ff80ff", "#80ffff", "#c0c0c0",
		},
		UI: models.UIColors{
			ButtonBackground:   "#f5f5f5",
			ButtonForeground:   "#000000",
			ButtonHover:        "#e0e0e0",
			ButtonActive:       "#0066cc",
			InputBackground:    "#ffffff",
			InputForeground:    "#000000",
			InputBorder:        "#cccccc",
			InputFocus:         "#0066cc",
			Border:             "#cccccc",
			Shadow:             "rgba(0, 0, 0, 0.1)",
		},
	}

	theme.Fonts = models.DefaultThemeFonts()
	theme.Effects = models.DefaultThemeEffects()
	theme.Settings = models.DefaultThemeSettings()

	return theme
}

// createCyberpunkTheme creates a cyberpunk-themed theme
func (s *Service) createCyberpunkTheme() *models.Theme {
	theme := models.NewTheme("cyberpunk", "Cyberpunk")
	theme.Description = "Cyberpunk theme with neon colors"
	theme.Author = "aDex-UI Team"
	theme.Version = "1.0.0"

	theme.Colors = models.ThemeColors{
		Background: models.ColorPalette{
			Primary:   "#0a0a0f",
			Secondary: "#1a0a2a",
			Tertiary:  "#2a1a3a",
		},
		Foreground: models.ColorPalette{
			Primary:   "#ff00ff",
			Secondary: "#00ffff",
			Tertiary:  "#ffff00",
		},
		Accent: models.AccentColors{
			Primary:   "#ff0080",
			Secondary: "#00ff80",
		},
		Status: models.StatusColors{
			Success: "#00ff80",
			Warning: "#ffaa00",
			Error:   "#ff0040",
			Info:    "#00aaff",
		},
		Terminal: models.TerminalColors{
			"#0a0a0f", "#ff0040", "#00ff80", "#ffff00",
			"#0080ff", "#ff00ff", "#00ffff", "#ffffff",
			"#80808f", "#ff8080", "#80ff80", "#ffff80",
			"#8080ff", "#ff80ff", "#80ffff", "#c0c0c0",
		},
		UI: models.UIColors{
			ButtonBackground:   "#1a0a2a",
			ButtonForeground:   "#00ffff",
			ButtonHover:        "#2a1a3a",
			ButtonActive:       "#ff0080",
			InputBackground:    "#0a0a0f",
			InputForeground:    "#00ffff",
			InputBorder:        "#ff0080",
			InputFocus:         "#00ff80",
			Border:             "#ff0080",
			Shadow:             "rgba(255, 0, 255, 0.3)",
		},
	}

	theme.Fonts = models.DefaultThemeFonts()
	theme.Effects = models.DefaultThemeEffects()
	theme.Settings = models.DefaultThemeSettings()

	return theme
}

// createRetroTheme creates a retro terminal theme
func (s *Service) createRetroTheme() *models.Theme {
	theme := models.NewTheme("retro", "Retro Terminal")
	theme.Description = "Retro green terminal theme"
	theme.Author = "aDex-UI Team"
	theme.Version = "1.0.0"

	theme.Colors = models.ThemeColors{
		Background: models.ColorPalette{
			Primary:   "#000000",
			Secondary: "#0a0a0a",
			Tertiary:  "#141414",
		},
		Foreground: models.ColorPalette{
			Primary:   "#00ff41",
			Secondary: "#00cc33",
			Tertiary:   "#009922",
		},
		Accent: models.AccentColors{
			Primary:   "#00ff41",
			Secondary: "#00cc33",
		},
		Status: models.StatusColors{
			Success: "#00ff41",
			Warning: "#ffaa00",
			Error:   "#ff3333",
			Info:    "#00aaff",
		},
		Terminal: models.TerminalColors{
			"#000000", "#cc0000", "#00cc00", "#cccc00",
			"#0000cc", "#cc00cc", "#00cccc", "#cccccc",
			"#808080", "#ff8080", "#80ff80", "#ffff80",
			"#8080ff", "#ff80ff", "#80ffff", "#ffffff",
		},
		UI: models.UIColors{
			ButtonBackground:   "#0a0a0a",
			ButtonForeground:   "#00ff41",
			ButtonHover:        "#141414",
			ButtonActive:       "#00ff41",
			InputBackground:    "#000000",
			InputForeground:    "#00ff41",
			InputBorder:        "#00ff41",
			InputFocus:         "#00ff41",
			Border:             "#00ff41",
			Shadow:             "rgba(0, 255, 65, 0.2)",
		},
	}

	theme.Fonts = models.DefaultThemeFonts()
	theme.Effects = models.DefaultThemeEffects()
	theme.Settings = models.DefaultThemeSettings()

	return theme
}

// Public API methods

// GetCurrentTheme returns the currently active theme
func (s *Service) GetCurrentTheme() (*models.Theme, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	theme, exists := s.themes[s.currentThemeID]
	if !exists {
		return nil, fmt.Errorf("current theme '%s' not found", s.currentThemeID)
	}

	return theme, nil
}

// SetCurrentTheme sets the currently active theme
func (s *Service) SetCurrentTheme(themeID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.themes[themeID]; !exists {
		return fmt.Errorf("theme '%s' not found", themeID)
	}

	oldThemeID := s.currentThemeID
	s.currentThemeID = themeID

	// Trigger theme change event
	s.triggerEvent("theme:changed", s.themes[themeID])

	// If auto-save is enabled, save the current theme preference
	if s.config.AutoSave {
		go s.saveCurrentThemePreference()
	}

	return nil
}

// GetTheme returns a theme by ID
func (s *Service) GetTheme(themeID string) (*models.Theme, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	theme, exists := s.themes[themeID]
	if !exists {
		return nil, fmt.Errorf("theme '%s' not found", themeID)
	}

	return theme, nil
}

// GetAllThemes returns all available themes
func (s *Service) GetAllThemes() []*models.Theme {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	themes := make([]*models.Theme, 0, len(s.themes))
	for _, theme := range s.themes {
		themes = append(themes, theme)
	}

	return themes
}

// SaveTheme saves a theme to disk
func (s *Service) SaveTheme(theme *models.Theme) error {
	if theme == nil {
		return fmt.Errorf("theme cannot be nil")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update theme in memory
	s.themes[theme.ID] = theme

	// Save to disk
	themePath := filepath.Join(s.themeDirectory, theme.ID+".json")
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Trigger save event
	s.triggerEvent("theme:saved", theme)

	return nil
}

// DeleteTheme deletes a theme
func (s *Service) DeleteTheme(themeID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if themeID == s.defaultThemeID {
		return fmt.Errorf("cannot delete default theme")
	}

	if themeID == s.currentThemeID {
		s.currentThemeID = s.defaultThemeID
	}

	// Remove from memory
	delete(s.themes, themeID)

	// Remove from disk
	themePath := filepath.Join(s.themeDirectory, themeID+".json")
	if err := os.Remove(themePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete theme file: %w", err)
	}

	// Trigger delete event
	s.triggerEvent("theme:deleted", nil)

	return nil
}

// ImportLegacyTheme imports a legacy eDEX-UI theme
func (s *Service) ImportLegacyTheme(legacyTheme *models.LegacyTheme) (*models.Theme, error) {
	if !s.config.EnableLegacyImport {
		return nil, fmt.Errorf("legacy theme import is disabled")
	}

	if legacyTheme == nil {
		return nil, fmt.Errorf("legacy theme cannot be nil")
	}

	// Convert legacy theme
	theme, err := s.converter.ConvertFromLegacy(legacyTheme)
	if err != nil {
		return nil, fmt.Errorf("failed to convert legacy theme: %w", err)
	}

	// Save the converted theme
	if err := s.SaveTheme(theme); err != nil {
		return nil, fmt.Errorf("failed to save converted theme: %w", err)
	}

	// Add to memory
	s.mutex.Lock()
	s.themes[theme.ID] = theme
	s.mutex.Unlock()

	// Trigger import event
	s.triggerEvent("theme:imported", theme)

	return theme, nil
}

// GenerateCSSVariables generates CSS variables for a theme
func (s *Service) GenerateCSSVariables(themeID string) (*GeneratedCSS, error) {
	s.mutex.RLock()
	theme, exists := s.themes[themeID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("theme '%s' not found", themeID)
	}

	return s.variableGen.GenerateCSSVariables(theme)
}

// GenerateCurrentCSSVariables generates CSS variables for the current theme
func (s *Service) GenerateCurrentCSSVariables() (*GeneratedCSS, error) {
	return s.GenerateCSSVariables(s.currentThemeID)
}

// Event handling methods

// Subscribe to theme events
func (s *Service) Subscribe(eventType string, handler func(*models.Theme)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.eventHandlers[eventType] == nil {
		s.eventHandlers[eventType] = make([]func(*models.Theme), 0)
	}

	s.eventHandlers[eventType] = append(s.eventHandlers[eventType], handler)
}

// triggerEvent triggers an event
func (s *Service) triggerEvent(eventType string, theme *models.Theme) {
	if handlers, exists := s.eventHandlers[eventType]; exists {
		for _, handler := range handlers {
			go handler(theme)
		}
	}
}

// Helper methods

// loadThemesFromDirectory loads themes from the theme directory
func (s *Service) loadThemesFromDirectory() error {
	if _, err := os.Stat(s.themeDirectory); os.IsNotExist(err) {
		return nil // Directory doesn't exist, that's okay
	}

	return filepath.WalkDir(s.themeDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Load theme from file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read theme file '%s': %w", path, err)
		}

		var theme models.Theme
		if err := json.Unmarshal(data, &theme); err != nil {
			return fmt.Errorf("failed to unmarshal theme from '%s': %w", path, err)
		}

		// Add to themes
		s.themes[theme.ID] = &theme

		return nil
	})
}

// startThemeWatcher starts a background watcher for theme changes
func (s *Service) startThemeWatcher() {
	ticker := time.NewTicker(s.config.WatchInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.reloadThemes()
	}
}

// reloadThemes reloads themes from disk
func (s *Service) reloadThemes() {
	if err := s.loadThemesFromDirectory(); err != nil {
		// Log error but don't crash
		fmt.Printf("Failed to reload themes: %v\n", err)
	}
}

// saveCurrentThemePreference saves the current theme preference
func (s *Service) saveCurrentThemePreference() {
	preferencePath := filepath.Join(s.themeDirectory, ".current-theme")
	data := []byte(s.currentThemeID)

	if err := os.WriteFile(preferencePath, data, 0644); err != nil {
		fmt.Printf("Failed to save theme preference: %v\n", err)
	}
}

// GetServiceStatus returns the current status of the theme service
func (s *Service) GetServiceStatus() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"current_theme_id":   s.currentThemeID,
		"default_theme_id":   s.defaultThemeID,
		"total_themes":       len(s.themes),
		"theme_directory":    s.themeDirectory,
		"auto_reload":        s.config.AutoReload,
		"cache_enabled":      s.config.CacheEnabled,
		"variable_cache_info": s.variableGen.GetCacheInfo(),
	}
}