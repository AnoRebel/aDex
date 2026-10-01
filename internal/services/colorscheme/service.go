package colorscheme

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aDex/internal/events"
	"aDex/internal/logger"
	"aDex/internal/models"
)

// Service manages terminal color schemes
type Service struct {
	schemes        map[string]*models.ColorScheme
	builtInSchemes map[string]*models.ColorScheme
	config         *models.ColorSchemeConfig
	eventBus       events.IEventBus
	logger         *logger.Logger
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	configPath     string
	schemesDir     string
}

// NewService creates a new color scheme service
func NewService(configPath, schemesDir string) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		schemes:        make(map[string]*models.ColorScheme),
		builtInSchemes: make(map[string]*models.ColorScheme),
		config: &models.ColorSchemeConfig{
			DefaultScheme:  "default-dark",
			UserSchemes:    make(map[string]*models.ColorScheme),
			EnabledSchemes: []string{"default-dark", "default-light", "solarized-dark", "solarized-light"},
			AutoSwitch:     false,
			ImportPath:     filepath.Join(schemesDir, "import"),
			ExportPath:     filepath.Join(schemesDir, "export"),
		},
		ctx:        ctx,
		cancel:     cancel,
		configPath: configPath,
		schemesDir: schemesDir,
	}

	return service
}

// Initialize initializes the color scheme service
func (s *Service) Initialize(ctx context.Context) error {
	s.ctx = ctx
	s.logger = logger.GetDefaultLogger()
	s.logger.Info("Initializing color scheme service")

	// Initialize event bus
	if s.eventBus == nil {
		s.eventBus = events.GetEventBus()
	}

	// Ensure directories exist
	if err := s.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Load built-in color schemes
	s.initializeBuiltInSchemes()

	// Load configuration
	if err := s.loadConfig(); err != nil {
		s.logger.Warn("Failed to load color scheme config, using defaults", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Load user color schemes
	if err := s.loadUserSchemes(); err != nil {
		s.logger.Warn("Failed to load user color schemes", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Merge all schemes
	s.mergeSchemes()

	s.logger.Info("Color scheme service initialized successfully", map[string]interface{}{
		"total_schemes":   len(s.schemes),
		"builtin_schemes": len(s.builtInSchemes),
		"user_schemes":    len(s.config.UserSchemes),
	})

	return nil
}

// SetEventBus sets the event bus for the service
func (s *Service) SetEventBus(bus events.IEventBus) {
	s.eventBus = bus
}

// GetEventBus returns the event bus
func (s *Service) GetEventBus() events.IEventBus {
	return s.eventBus
}

// GetSchemes returns all available color schemes
func (s *Service) GetSchemes() map[string]*models.ColorScheme {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*models.ColorScheme)
	for id, scheme := range s.schemes {
		result[id] = scheme.Clone()
	}
	return result
}

// GetScheme returns a specific color scheme by ID
func (s *Service) GetScheme(id string) *models.ColorScheme {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if scheme, exists := s.schemes[id]; exists {
		return scheme.Clone()
	}
	return nil
}

// CreateScheme creates a new color scheme
func (s *Service) CreateScheme(scheme *models.ColorScheme) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate scheme
	validation := scheme.Validate()
	if !validation.Valid {
		return fmt.Errorf("invalid color scheme: %v", validation.Errors)
	}

	// Check if scheme already exists
	if _, exists := s.schemes[scheme.ID]; exists {
		return fmt.Errorf("color scheme with ID '%s' already exists", scheme.ID)
	}

	// Set timestamps
	now := time.Now()
	scheme.CreatedAt = now
	scheme.UpdatedAt = now

	// Add to schemes
	s.schemes[scheme.ID] = scheme.Clone()

	// If not built-in, add to user schemes
	if !scheme.IsBuiltIn {
		s.config.UserSchemes[scheme.ID] = scheme.Clone()
		// Save to file
		if err := s.saveUserScheme(scheme); err != nil {
			s.logger.Error("Failed to save user scheme", err, map[string]interface{}{
				"scheme_id": scheme.ID,
			})
		}
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		s.logger.Error("Failed to save color scheme config", err)
	}

	// Publish event
	s.publishEvent("colorscheme.created", map[string]interface{}{
		"scheme_id":   scheme.ID,
		"scheme_name": scheme.Name,
		"is_built_in": scheme.IsBuiltIn,
		"timestamp":   time.Now(),
	})

	s.logger.Info("Color scheme created", map[string]interface{}{
		"scheme_id":   scheme.ID,
		"scheme_name": scheme.Name,
		"is_built_in": scheme.IsBuiltIn,
	})

	return nil
}

// UpdateScheme updates an existing color scheme
func (s *Service) UpdateScheme(scheme *models.ColorScheme) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate scheme
	validation := scheme.Validate()
	if !validation.Valid {
		return fmt.Errorf("invalid color scheme: %v", validation.Errors)
	}

	// Check if scheme exists
	existing, exists := s.schemes[scheme.ID]
	if !exists {
		return fmt.Errorf("color scheme with ID '%s' not found", scheme.ID)
	}

	// Don't allow updating built-in schemes directly
	if existing.IsBuiltIn {
		return fmt.Errorf("cannot update built-in color scheme '%s'", scheme.ID)
	}

	// Update timestamp
	scheme.UpdatedAt = time.Now()
	scheme.CreatedAt = existing.CreatedAt // Preserve creation time

	// Update scheme
	s.schemes[scheme.ID] = scheme.Clone()
	s.config.UserSchemes[scheme.ID] = scheme.Clone()

	// Save to file
	if err := s.saveUserScheme(scheme); err != nil {
		s.logger.Error("Failed to save user scheme", err, map[string]interface{}{
			"scheme_id": scheme.ID,
		})
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		s.logger.Error("Failed to save color scheme config", err)
	}

	// Publish event
	s.publishEvent("colorscheme.updated", map[string]interface{}{
		"scheme_id":   scheme.ID,
		"scheme_name": scheme.Name,
		"timestamp":   time.Now(),
	})

	s.logger.Info("Color scheme updated", map[string]interface{}{
		"scheme_id":   scheme.ID,
		"scheme_name": scheme.Name,
	})

	return nil
}

// DeleteScheme deletes a color scheme
func (s *Service) DeleteScheme(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scheme exists
	scheme, exists := s.schemes[id]
	if !exists {
		return fmt.Errorf("color scheme with ID '%s' not found", id)
	}

	// Don't allow deleting built-in schemes
	if scheme.IsBuiltIn {
		return fmt.Errorf("cannot delete built-in color scheme '%s'", id)
	}

	// Don't allow deleting the default scheme
	if s.config.DefaultScheme == id {
		return fmt.Errorf("cannot delete default color scheme '%s'", id)
	}

	// Remove from schemes
	delete(s.schemes, id)
	delete(s.config.UserSchemes, id)

	// Remove file
	schemePath := s.getUserSchemePath(id)
	if err := os.Remove(schemePath); err != nil && !os.IsNotExist(err) {
		s.logger.Error("Failed to delete user scheme file", err, map[string]interface{}{
			"scheme_id": id,
			"path":      schemePath,
		})
	}

	// Update enabled schemes list
	s.config.EnabledSchemes = removeFromStringSlice(s.config.EnabledSchemes, id)

	// Save configuration
	if err := s.saveConfig(); err != nil {
		s.logger.Error("Failed to save color scheme config", err)
	}

	// Publish event
	s.publishEvent("colorscheme.deleted", map[string]interface{}{
		"scheme_id":   id,
		"scheme_name": scheme.Name,
		"timestamp":   time.Now(),
	})

	s.logger.Info("Color scheme deleted", map[string]interface{}{
		"scheme_id":   id,
		"scheme_name": scheme.Name,
	})

	return nil
}

// SetDefaultScheme sets the default color scheme
func (s *Service) SetDefaultScheme(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scheme exists
	if _, exists := s.schemes[id]; !exists {
		return fmt.Errorf("color scheme with ID '%s' not found", id)
	}

	// Update default
	s.config.DefaultScheme = id

	// Save configuration
	if err := s.saveConfig(); err != nil {
		s.logger.Error("Failed to save color scheme config", err)
	}

	// Publish event
	s.publishEvent("colorscheme.default_changed", map[string]interface{}{
		"scheme_id": id,
		"timestamp": time.Now(),
	})

	s.logger.Info("Default color scheme changed", map[string]interface{}{
		"scheme_id": id,
	})

	return nil
}

// GetDefaultScheme returns the default color scheme
func (s *Service) GetDefaultScheme() *models.ColorScheme {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.GetScheme(s.config.DefaultScheme)
}

// GetConfig returns the color scheme configuration
func (s *Service) GetConfig() *models.ColorSchemeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	config := *s.config
	config.UserSchemes = make(map[string]*models.ColorScheme)
	for id, scheme := range s.config.UserSchemes {
		config.UserSchemes[id] = scheme.Clone()
	}
	return &config
}

// UpdateConfig updates the color scheme configuration
func (s *Service) UpdateConfig(config *models.ColorSchemeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate default scheme exists
	if _, exists := s.schemes[config.DefaultScheme]; !exists {
		return fmt.Errorf("default color scheme '%s' not found", config.DefaultScheme)
	}

	// Update config
	s.config = config

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Publish event
	s.publishEvent("colorscheme.config_updated", map[string]interface{}{
		"default_scheme": config.DefaultScheme,
		"auto_switch":    config.AutoSwitch,
		"timestamp":      time.Now(),
	})

	s.logger.Info("Color scheme configuration updated", map[string]interface{}{
		"default_scheme": config.DefaultScheme,
		"auto_switch":    config.AutoSwitch,
	})

	return nil
}

// ImportScheme imports a color scheme from the specified source
func (s *Service) ImportScheme(importConfig *models.ColorSchemeImport) (*models.ColorScheme, error) {
	// This would implement import functionality from various sources
	// For now, it's a placeholder
	s.logger.Info("Color scheme import requested", map[string]interface{}{
		"source": importConfig.Source,
		"format": importConfig.Format,
	})

	return nil, fmt.Errorf("import functionality not yet implemented")
}

// ExportScheme exports color schemes to the specified destination
func (s *Service) ExportScheme(exportConfig *models.ColorSchemeExport) error {
	// This would implement export functionality to various formats
	// For now, it's a placeholder
	s.logger.Info("Color scheme export requested", map[string]interface{}{
		"format":  exportConfig.Format,
		"schemes": exportConfig.SchemeIDs,
	})

	return fmt.Errorf("export functionality not yet implemented")
}

// ValidateScheme validates a color scheme
func (s *Service) ValidateScheme(scheme *models.ColorScheme) *models.ColorSchemeValidationResult {
	return scheme.Validate()
}

// GetSchemePreview generates a preview for a color scheme
func (s *Service) GetSchemePreview(id string) (*models.ColorSchemePreview, error) {
	s.mu.RLock()
	scheme, exists := s.schemes[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("color scheme with ID '%s' not found", id)
	}

	// Create preview
	preview := &models.ColorSchemePreview{
		SchemeID: id,
		Name:     scheme.Name,
		Colors: map[string]string{
			"background": scheme.Colors.Background,
			"foreground": scheme.Colors.Foreground,
			"cursor":     scheme.Colors.Cursor,
			"selection":  scheme.Colors.Selection,
			"black":      scheme.Colors.Black,
			"red":        scheme.Colors.Red,
			"green":      scheme.Colors.Green,
			"yellow":     scheme.Colors.Yellow,
			"blue":       scheme.Colors.Blue,
			"magenta":    scheme.Colors.Magenta,
			"cyan":       scheme.Colors.Cyan,
			"white":      scheme.Colors.White,
		},
		Sample:    s.generateSampleOutput(scheme),
		CreatedAt: time.Now(),
	}

	return preview, nil
}

// Shutdown shuts down the color scheme service
func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down color scheme service")

	// Cancel context
	s.cancel()

	// Save configuration
	if err := s.saveConfig(); err != nil {
		s.logger.Error("Failed to save color scheme config on shutdown", err)
	}

	s.logger.Info("Color scheme service shutdown complete")
	return nil
}

// Helper methods

func (s *Service) ensureDirectories() error {
	dirs := []string{
		s.schemesDir,
		s.config.ImportPath,
		s.config.ExportPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (s *Service) initializeBuiltInSchemes() {
	s.builtInSchemes = map[string]*models.ColorScheme{
		"default-dark":    s.createDefaultDarkScheme(),
		"default-light":   s.createDefaultLightScheme(),
		"solarized-dark":  s.createSolarizedDarkScheme(),
		"solarized-light": s.createSolarizedLightScheme(),
		"dracula":         s.createDraculaScheme(),
		"monokai":         s.createMonokaiScheme(),
		"nord":            s.createNordScheme(),
		"gruvbox-dark":    s.createGruvboxDarkScheme(),
		"gruvbox-light":   s.createGruvboxLightScheme(),
	}
}

func (s *Service) loadConfig() error {
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		// Config doesn't exist, create default
		return s.saveConfig()
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return json.Unmarshal(data, s.config)
}

func (s *Service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(s.configPath, data, 0644)
}

func (s *Service) loadUserSchemes() error {
	// Load all JSON files from the schemes directory
	files, err := filepath.Glob(filepath.Join(s.schemesDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob scheme files: %w", err)
	}

	for _, file := range files {
		if err := s.loadUserSchemeFile(file); err != nil {
			s.logger.Error("Failed to load user scheme file", err, map[string]interface{}{
				"file": file,
			})
		}
	}

	return nil
}

func (s *Service) loadUserSchemeFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read scheme file: %w", err)
	}

	var scheme models.ColorScheme
	if err := json.Unmarshal(data, &scheme); err != nil {
		return fmt.Errorf("failed to unmarshal scheme: %w", err)
	}

	// Mark as user scheme
	scheme.IsBuiltIn = false
	s.config.UserSchemes[scheme.ID] = &scheme

	return nil
}

func (s *Service) saveUserScheme(scheme *models.ColorScheme) error {
	data, err := json.MarshalIndent(scheme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scheme: %w", err)
	}

	schemePath := s.getUserSchemePath(scheme.ID)
	return os.WriteFile(schemePath, data, 0644)
}

func (s *Service) getUserSchemePath(id string) string {
	return filepath.Join(s.schemesDir, fmt.Sprintf("%s.json", id))
}

func (s *Service) mergeSchemes() {
	// Start with built-in schemes
	for id, scheme := range s.builtInSchemes {
		s.schemes[id] = scheme.Clone()
	}

	// Add user schemes
	for id, scheme := range s.config.UserSchemes {
		s.schemes[id] = scheme.Clone()
	}
}

func (s *Service) publishEvent(eventType string, data map[string]interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(s.ctx, eventType, data, "colorscheme-service")
		if err != nil {
			s.logger.Error("Failed to publish event", err, map[string]interface{}{
				"event_type": eventType,
				"data":       data,
			})
		}
	}
}

func (s *Service) generateSampleOutput(scheme *models.ColorScheme) string {
	// Generate sample terminal output showing the color scheme
	sample := fmt.Sprintf(`\x1b[38;2;%smWelcome to %s\x1b[0m

\x1b[38;2;%smNormal text\x1b[0m in \x1b[38;2;%sm%s\x1b[0m theme

Color palette:
 \x1b[38;2;%sm● Black\x1b[0m   \x1b[38;2;%sm● Red\x1b[0m    \x1b[38;2;%sm● Green\x1b[0m  \x1b[38;2;%sm● Yellow\x1b[0m
 \x1b[38;2;%sm● Blue\x1b[0m   \x1b[38;2;%sm● Magenta\x1b[0m \x1b[38;2;%sm● Cyan\x1b[0m   \x1b[38;2;%sm● White\x1b[0m

\x1b[1mBold text\x1b[0m and \x1b[3mitalic text\x1b[0m

$ ls -la
drwxr-xr-x  5 user group  160 Jan 15 10:30 \x1b[38;2;%smdirectory\x1b[0m
-rw-r--r--  1 user group 1024 Jan 15 10:30 \x1b[38;2;%sfile.txt\x1b[0m
-rwxr-xr-x  1 user group 2048 Jan 15 10:30 \x1b[38;2;%sexecutable\x1b[0m`,
		scheme.Colors.Foreground,
		scheme.Name,
		scheme.Colors.Foreground,
		scheme.Colors.Blue,
		scheme.Name,
		scheme.Colors.Black,
		scheme.Colors.Red,
		scheme.Colors.Green,
		scheme.Colors.Yellow,
		scheme.Colors.Blue,
		scheme.Colors.Magenta,
		scheme.Colors.Cyan,
		scheme.Colors.White,
		scheme.Colors.Blue,
		scheme.Colors.Green,
		scheme.Colors.Green,
	)

	return sample
}

func removeFromStringSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
