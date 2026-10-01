package font

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"aDex/internal/events"
	"aDex/internal/logger"
	"aDex/internal/models"
)

// Service manages font configurations and system fonts
type Service struct {
	configurations map[string]*models.FontConfiguration
	systemFonts    map[string]*models.SystemFont
	settings       *models.FontSettings
	eventBus       events.IEventBus
	logger         *logger.Logger
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	fontScanner    *FontScanner
	fontValidator  *FontValidator
	configPath     string
}

// NewService creates a new font service
func NewService(configPath string) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		configurations: make(map[string]*models.FontConfiguration),
		systemFonts:    make(map[string]*models.SystemFont),
		settings:       models.NewDefaultFontSettings(),
		ctx:            ctx,
		cancel:         cancel,
		configPath:     configPath,
		fontScanner:    NewFontScanner(),
		fontValidator:  NewFontValidator(),
	}

	// Initialize built-in font configurations
	service.initializeBuiltinConfigurations()

	return service
}

// Initialize initializes the font service
func (s *Service) Initialize(ctx context.Context) error {
	s.ctx = ctx
	s.logger = logger.GetDefaultLogger()
	s.logger.Info("Initializing font service")

	// Initialize event bus
	if s.eventBus == nil {
		s.eventBus = events.GetEventBus()
	}

	// Load existing configurations
	if err := s.loadConfigurations(); err != nil {
		s.logger.Warn("Failed to load font configurations, using defaults", map[string]interface{}{
			"error": err,
		})
	}

	// Load settings
	if err := s.loadSettings(); err != nil {
		s.logger.Warn("Failed to load font settings, using defaults", map[string]interface{}{
			"error": err,
		})
	}

	// Scan for system fonts if auto-detect is enabled
	if s.settings.AutoDetectFonts {
		go s.scanSystemFonts()
	}

	s.logger.Info("Font service initialized successfully")
	return nil
}

// SetEventBus sets the event bus for the service
func (s *Service) SetEventBus(bus events.IEventBus) {
	s.eventBus = bus
}

// GetEventBus returns the event bus
func (s *Service) GetEventBus() events.IEventBus {
	if s.eventBus == nil {
		return nil
	}
	return s.eventBus
}

// GetConfigurations returns all font configurations
func (s *Service) GetConfigurations() map[string]*models.FontConfiguration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*models.FontConfiguration)
	for id, config := range s.configurations {
		result[id] = config.Clone()
	}
	return result
}

// GetConfiguration returns a specific font configuration by ID
func (s *Service) GetConfiguration(id string) *models.FontConfiguration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if config, exists := s.configurations[id]; exists {
		return config.Clone()
	}
	return nil
}

// CreateConfiguration creates a new font configuration
func (s *Service) CreateConfiguration(config *models.FontConfiguration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.ID == "" {
		return models.NewError(models.ErrConfigValidate, "Configuration ID is required", nil)
	}

	if !config.IsValid() {
		return models.NewError(models.ErrConfigValidate, "Invalid font configuration", nil)
	}

	if _, exists := s.configurations[config.ID]; exists {
		return models.NewError(models.ErrConfigValidate, "Configuration already exists", nil)
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	s.configurations[config.ID] = config.Clone()

	// Save configurations
	if err := s.saveConfigurations(); err != nil {
		s.logger.Error("Failed to save font configurations", err)
	}

	// Publish event
	s.publishEvent("font.configuration.created", map[string]interface{}{
		"configuration_id": config.ID,
		"name":             config.Name,
		"family":           config.Family,
		"timestamp":        time.Now(),
	})

	s.logger.Info("Font configuration created", map[string]interface{}{
		"id":     config.ID,
		"name":   config.Name,
		"family": config.Family,
	})

	return nil
}

// UpdateConfiguration updates an existing font configuration
func (s *Service) UpdateConfiguration(config *models.FontConfiguration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.ID == "" {
		return models.NewError(models.ErrConfigValidate, "Configuration ID is required", nil)
	}

	if !config.IsValid() {
		return models.NewError(models.ErrConfigValidate, "Invalid font configuration", nil)
	}

	if _, exists := s.configurations[config.ID]; !exists {
		return models.NewError(models.ErrConfigValidate, "Configuration not found", nil)
	}

	// Don't allow modification of built-in configurations
	if existing, exists := s.configurations[config.ID]; exists && existing.IsBuiltIn {
		return models.NewError(models.ErrConfigValidate, "Cannot modify built-in configuration", nil)
	}

	config.UpdatedAt = time.Now()
	s.configurations[config.ID] = config.Clone()

	// Save configurations
	if err := s.saveConfigurations(); err != nil {
		s.logger.Error("Failed to save font configurations", err)
	}

	// Publish event
	s.publishEvent("font.configuration.updated", map[string]interface{}{
		"configuration_id": config.ID,
		"name":             config.Name,
		"family":           config.Family,
		"timestamp":        time.Now(),
	})

	s.logger.Info("Font configuration updated", map[string]interface{}{
		"id":     config.ID,
		"name":   config.Name,
		"family": config.Family,
	})

	return nil
}

// DeleteConfiguration deletes a font configuration
func (s *Service) DeleteConfiguration(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configurations[id]
	if !exists {
		return models.NewError(models.ErrConfigValidate, "Configuration not found", nil)
	}

	// Don't allow deletion of built-in configurations
	if config.IsBuiltIn {
		return models.NewError(models.ErrConfigValidate, "Cannot delete built-in configuration", nil)
	}

	// Don't allow deletion of the default configuration
	if s.settings.DefaultFontID == id {
		return models.NewError(models.ErrConfigValidate, "Cannot delete default font configuration", nil)
	}

	delete(s.configurations, id)

	// Save configurations
	if err := s.saveConfigurations(); err != nil {
		s.logger.Error("Failed to save font configurations", err)
	}

	// Publish event
	s.publishEvent("font.configuration.deleted", map[string]interface{}{
		"configuration_id": id,
		"name":             config.Name,
		"timestamp":        time.Now(),
	})

	s.logger.Info("Font configuration deleted", map[string]interface{}{
		"id":   id,
		"name": config.Name,
	})

	return nil
}

// GetSystemFonts returns all detected system fonts
func (s *Service) GetSystemFonts() map[string]*models.SystemFont {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*models.SystemFont)
	for id, font := range s.systemFonts {
		result[id] = font
	}
	return result
}

// GetMonospaceFonts returns monospace fonts suitable for terminal use
func (s *Service) GetMonospaceFonts() []*models.SystemFont {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var monospaceFonts []*models.SystemFont
	for _, font := range s.systemFonts {
		if font.IsMonospace && font.IsInstalled {
			monospaceFonts = append(monospaceFonts, font)
		}
	}

	// Sort by display name
	sort.Slice(monospaceFonts, func(i, j int) bool {
		return monospaceFonts[i].DisplayName < monospaceFonts[j].DisplayName
	})

	return monospaceFonts
}

// ScanSystemFonts scans the system for available fonts
func (s *Service) ScanSystemFonts() error {
	s.logger.Info("Starting system font scan")

	fonts, err := s.fontScanner.ScanSystemFonts(s.settings.FontDirectories)
	if err != nil {
		return models.NewError(models.ErrConfigValidate, "Failed to scan system fonts", err)
	}

	s.mu.Lock()
	// Clear existing system fonts
	s.systemFonts = make(map[string]*models.SystemFont)

	// Add scanned fonts
	for id, font := range fonts {
		s.systemFonts[id] = font
	}

	// Update last scan time
	s.settings.LastScanTime = time.Now()
	s.mu.Unlock()

	// Save settings
	if err := s.saveSettings(); err != nil {
		s.logger.Error("Failed to save font settings", err)
	}

	// Publish event
	s.publishEvent("font.system.scanned", map[string]interface{}{
		"font_count": len(fonts),
		"timestamp":  time.Now(),
	})

	s.logger.Info("System font scan completed", map[string]interface{}{
		"font_count": len(fonts),
	})

	return nil
}

// ImportFont imports a font from a file or URL
func (s *Service) ImportFont(req *models.FontImportRequest) (*models.FontImportResult, error) {
	result := &models.FontImportResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Validate request
	if req.Source == "" {
		result.Errors = append(result.Errors, "Source is required")
		return result, models.NewError(models.ErrConfigValidate, "Invalid import request", nil)
	}

	// Validate font file
	fontData, err := s.fontValidator.ValidateFontSource(req.Source)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Font validation failed: %v", err))
		return result, models.NewError(models.ErrConfigValidate, "Invalid font file", err)
	}

	// If validate only, return validation result
	if req.ValidateOnly {
		result.Success = true
		result.Warnings = fontData.Warnings
		return result, nil
	}

	// Import the font
	systemFont, err := s.fontScanner.ImportFont(req.Source, req)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Font import failed: %v", err))
		return result, models.NewError(models.ErrConfigValidate, "Failed to import font", err)
	}

	// Add to system fonts
	s.mu.Lock()
	s.systemFonts[systemFont.Family] = systemFont
	s.mu.Unlock()

	// Create font configuration if requested
	if req.Name != "" {
		config := models.NewFontConfiguration(
			fmt.Sprintf("imported-%s", strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))),
			req.Name,
			systemFont.Family,
		)
		config.Size = s.settings.FontSize
		config.LineHeight = s.settings.LineHeight
		config.LetterSpacing = s.settings.LetterSpacing
		config.Ligatures = s.settings.EnableLigatures
		config.Antialias = s.settings.EnableAntialias
		config.Hinting = s.settings.Hinting

		if err := s.CreateConfiguration(config); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Created font import but failed to create configuration: %v", err))
		} else {
			result.Configuration = config
		}
	}

	result.Success = true
	result.Font = systemFont
	result.Warnings = append(result.Warnings, fontData.Warnings...)

	// Publish event
	s.publishEvent("font.imported", map[string]interface{}{
		"font_family":  systemFont.Family,
		"display_name": systemFont.DisplayName,
		"source":       req.Source,
		"timestamp":    time.Now(),
	})

	s.logger.Info("Font imported successfully", map[string]interface{}{
		"family": systemFont.Family,
		"source": req.Source,
	})

	return result, nil
}

// ValidateFont validates a font configuration
func (s *Service) ValidateFont(config *models.FontConfiguration) *models.FontValidationResult {
	return s.fontValidator.ValidateConfiguration(config)
}

// GetFontMetrics returns metrics for a font
func (s *Service) GetFontMetrics(family string, size int) (*models.FontMetrics, error) {
	s.mu.RLock()
	font, exists := s.systemFonts[family]
	s.mu.RUnlock()

	if !exists {
		// Try to find the font in configurations
		s.mu.RLock()
		for _, config := range s.configurations {
			if config.Family == family {
				font = config.ToSystemFont()
				break
			}
		}
		s.mu.RUnlock()

		if font == nil {
			return nil, models.NewError(models.ErrConfigValidate, "Font not found", nil)
		}
	}

	// Get font metrics
	metrics, err := s.fontScanner.GetFontMetrics(font, size)
	if err != nil {
		return nil, models.NewError(models.ErrConfigValidate, "Failed to get font metrics", err)
	}

	return metrics, nil
}

// GetSettings returns current font settings
func (s *Service) GetSettings() *models.FontSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	settingsCopy := *s.settings
	if s.settings.FontFeatureSettings != nil {
		settingsCopy.FontFeatureSettings = make(map[string]bool)
		for k, v := range s.settings.FontFeatureSettings {
			settingsCopy.FontFeatureSettings[k] = v
		}
	}
	if s.settings.FontDirectories != nil {
		settingsCopy.FontDirectories = make([]string, len(s.settings.FontDirectories))
		copy(settingsCopy.FontDirectories, s.settings.FontDirectories)
	}

	return &settingsCopy
}

// UpdateSettings updates font settings
func (s *Service) UpdateSettings(settings *models.FontSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate settings
	if settings.DefaultFontID == "" {
		return models.NewError(models.ErrConfigValidate, "Default font ID is required", nil)
	}

	if settings.FontSize < 6 || settings.FontSize > 72 {
		return models.NewError(models.ErrConfigValidate, "Font size must be between 6 and 72", nil)
	}

	if settings.LineHeight < 0.8 || settings.LineHeight > 3.0 {
		return models.NewError(models.ErrConfigValidate, "Line height must be between 0.8 and 3.0", nil)
	}

	s.settings = settings

	// Save settings
	if err := s.saveSettings(); err != nil {
		s.logger.Error("Failed to save font settings", err)
	}

	// Publish event
	s.publishEvent("font.settings.updated", map[string]interface{}{
		"default_font_id": settings.DefaultFontID,
		"font_size":       settings.FontSize,
		"timestamp":       time.Now(),
	})

	s.logger.Info("Font settings updated", map[string]interface{}{
		"default_font": settings.DefaultFontID,
		"font_size":    settings.FontSize,
	})

	return nil
}

// GetDefaultConfiguration returns the default font configuration
func (s *Service) GetDefaultConfiguration() *models.FontConfiguration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try to find the configured default
	if config, exists := s.configurations[s.settings.DefaultFontID]; exists {
		return config.Clone()
	}

	// Fallback to built-in default
	for _, config := range s.configurations {
		if config.IsDefault {
			return config.Clone()
		}
	}

	// Create a default one if none exists
	defaultConfig := models.NewFontConfiguration("default-mono", "Default Monospace", "JetBrains Mono")
	defaultConfig.IsDefault = true
	defaultConfig.IsBuiltIn = true
	return defaultConfig
}

// SetDefaultConfiguration sets the default font configuration
func (s *Service) SetDefaultConfiguration(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configurations[id]; !exists {
		return models.NewError(models.ErrConfigValidate, "Configuration not found", nil)
	}

	oldDefault := s.settings.DefaultFontID
	s.settings.DefaultFontID = id

	// Update configuration flags
	for _, config := range s.configurations {
		config.IsDefault = (config.ID == id)
	}

	// Save settings
	if err := s.saveSettings(); err != nil {
		s.logger.Error("Failed to save font settings", err)
	}

	// Save configurations
	if err := s.saveConfigurations(); err != nil {
		s.logger.Error("Failed to save font configurations", err)
	}

	// Publish event
	s.publishEvent("font.default.changed", map[string]interface{}{
		"old_default": oldDefault,
		"new_default": id,
		"timestamp":   time.Now(),
	})

	s.logger.Info("Default font configuration changed", map[string]interface{}{
		"old_default": oldDefault,
		"new_default": id,
	})

	return nil
}

// Shutdown shuts down the font service
func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down font service")

	// Cancel context
	s.cancel()

	// Save final state
	s.mu.Lock()
	if err := s.saveConfigurations(); err != nil {
		s.logger.Error("Failed to save font configurations during shutdown", err)
	}
	if err := s.saveSettings(); err != nil {
		s.logger.Error("Failed to save font settings during shutdown", err)
	}
	s.mu.Unlock()

	s.logger.Info("Font service shutdown complete")
	return nil
}

// Private methods

func (s *Service) initializeBuiltinConfigurations() {
	builtins := []*models.FontConfiguration{
		{
			ID:          "jetbrains-mono",
			Name:        "JetBrains Mono",
			Description: "JetBrains Mono typeface for developers",
			Family:      "JetBrains Mono",
			Size:        14,
			Weight:      string(models.FontWeightNormal),
			LineHeight:  1.4,
			Ligatures:   true,
			Antialias:   true,
			Hinting:     models.FontHintingSlight,
			Features: map[string]bool{
				"liga": true,
				"dlig": true,
			},
			IsBuiltIn: true,
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "fira-code",
			Name:        "Fira Code",
			Description: "Fira Code monospace font with programming ligatures",
			Family:      "Fira Code",
			Size:        14,
			Weight:      string(models.FontWeightNormal),
			LineHeight:  1.4,
			Ligatures:   true,
			Antialias:   true,
			Hinting:     models.FontHintingSlight,
			Features: map[string]bool{
				"liga": true,
				"dlig": true,
				"zero": true,
			},
			IsBuiltIn: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "source-code-pro",
			Name:        "Source Code Pro",
			Description: "Source Code Pro by Adobe",
			Family:      "Source Code Pro",
			Size:        14,
			Weight:      string(models.FontWeightNormal),
			LineHeight:  1.4,
			Ligatures:   false,
			Antialias:   true,
			Hinting:     models.FontHintingSlight,
			IsBuiltIn:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cascadia-code",
			Name:        "Cascadia Code",
			Description: "Cascadia Code by Microsoft",
			Family:      "Cascadia Code",
			Size:        14,
			Weight:      string(models.FontWeightNormal),
			LineHeight:  1.4,
			Ligatures:   true,
			Antialias:   true,
			Hinting:     models.FontHintingSlight,
			Features: map[string]bool{
				"liga": true,
				"dlig": true,
				"zero": true,
			},
			IsBuiltIn: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, config := range builtins {
		s.configurations[config.ID] = config
	}
}

func (s *Service) scanSystemFonts() {
	if err := s.ScanSystemFonts(); err != nil {
		s.logger.Error("Failed to scan system fonts", err)
	}
}

func (s *Service) loadConfigurations() error {
	if s.configPath == "" {
		return nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, use defaults
		}
		return err
	}

	var configs map[string]*models.FontConfiguration
	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	s.mu.Lock()
	for id, config := range configs {
		s.configurations[id] = config
	}
	s.mu.Unlock()

	return nil
}

func (s *Service) saveConfigurations() error {
	if s.configPath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.configurations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

func (s *Service) loadSettings() error {
	settingsPath := strings.TrimSuffix(s.configPath, filepath.Ext(s.configPath)) + "-settings.json"

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, use defaults
		}
		return err
	}

	var settings models.FontSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	s.mu.Lock()
	s.settings = &settings
	s.mu.Unlock()

	return nil
}

func (s *Service) saveSettings() error {
	settingsPath := strings.TrimSuffix(s.configPath, filepath.Ext(s.configPath)) + "-settings.json"

	// Ensure directory exists
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

func (s *Service) publishEvent(eventType string, data map[string]interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(s.ctx, eventType, data, "font-service")
		if err != nil {
			s.logger.Error("Failed to publish event", err, map[string]interface{}{
				"event_type": eventType,
				"data":       data,
			})
		}
	}
}
