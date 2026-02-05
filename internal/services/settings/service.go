package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// Service manages application settings and configuration
type Service struct {
	logger     *logger.Logger
	configDir  string
	configFile string
	settings   *models.AppSettings
	migration  *Migration
}

// NewService creates a new settings service
func NewService(logger *logger.Logger) (*Service, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config directory: %w", err)
	}

	appConfigDir := filepath.Join(configDir, "aDex-UI")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	service := &Service{
		logger:     logger,
		configDir:  appConfigDir,
		configFile: filepath.Join(appConfigDir, "settings.json"),
		migration:  NewMigration(logger),
	}

	// Load or create default settings
	if err := service.loadSettings(); err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	return service, nil
}

// GetSettings returns the current application settings
func (s *Service) GetSettings() *models.AppSettings {
	s.settings.Lock()
	defer s.settings.Unlock()

	// Return a copy to prevent external modification
	copy := *s.settings
	return &copy
}

// UpdateSettings updates the application settings
func (s *Service) UpdateSettings(settings *models.AppSettings) error {
	if err := settings.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	s.settings.Lock()
	defer s.settings.Unlock()

	s.settings = settings
	s.settings.UpdatedAt = time.Now()

	return s.saveSettings()
}

// UpdateSettingsFunc updates settings using a function
func (s *Service) UpdateSettingsFunc(updateFn func(*models.AppSettings)) error {
	s.settings.Lock()
	defer s.settings.Unlock()

	updateFn(s.settings)
	s.settings.UpdatedAt = time.Now()

	if err := s.settings.Validate(); err != nil {
		return fmt.Errorf("invalid settings after update: %w", err)
	}

	return s.saveSettings()
}

// ResetToDefaults resets all settings to default values
func (s *Service) ResetToDefaults() error {
	s.settings.Lock()
	defer s.settings.Unlock()

	s.settings = models.DefaultAppSettings()
	return s.saveSettings()
}

// ExportSettings exports settings to a file
func (s *Service) ExportSettings(filename string) error {
	s.settings.Lock()
	defer s.settings.Unlock()

	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// ImportSettings imports settings from a file
func (s *Service) ImportSettings(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings models.AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	if err := settings.Validate(); err != nil {
		return fmt.Errorf("invalid imported settings: %w", err)
	}

	s.settings.Lock()
	defer s.settings.Unlock()

	s.settings = &settings
	s.settings.UpdatedAt = time.Now()

	return s.saveSettings()
}

// ImportFromLegacy imports settings from legacy eDEX-UI configuration
func (s *Service) ImportFromLegacy() error {
	s.logger.Info("Starting legacy configuration import")

	// Find legacy configurations
	configPaths, err := s.migration.FindLegacyConfigs()
	if err != nil {
		return fmt.Errorf("failed to find legacy configs: %w", err)
	}

	if len(configPaths) == 0 {
		s.logger.Info("No legacy eDEX-UI configurations found")
		return nil
	}

	// Import the first found configuration
	legacySettings, err := s.migration.ImportLegacyConfig(configPaths[0])
	if err != nil {
		return fmt.Errorf("failed to import legacy config: %w", err)
	}

	// Merge with current settings
	mergedSettings := s.mergeSettings(s.GetSettings(), legacySettings)

	// Update with merged settings
	if err := s.UpdateSettings(mergedSettings); err != nil {
		return fmt.Errorf("failed to update settings with legacy import: %w", err)
	}

	s.logger.Info("Successfully imported legacy configuration", map[string]interface{}{
		"legacy_version": mergedSettings.Legacy.Version,
		"imported_from":  configPaths[0],
	})

	return nil
}

// ExportToLegacy exports current settings in legacy eDEX-UI format
func (s *Service) ExportToLegacy(filename string) error {
	s.logger.Info("Exporting settings to legacy format")

	currentSettings := s.GetSettings()
	legacyConfig, err := s.migration.ExportLegacyData(currentSettings)
	if err != nil {
		return fmt.Errorf("failed to export to legacy format: %w", err)
	}

	data, err := json.MarshalIndent(legacyConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal legacy config: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// GetConfigDir returns the configuration directory path
func (s *Service) GetConfigDir() string {
	return s.configDir
}

// GetConfigFile returns the settings file path
func (s *Service) GetConfigFile() string {
	return s.configFile
}

// Cleanup removes temporary files and performs cleanup
func (s *Service) Cleanup() error {
	// Remove temporary files older than 24 hours
	tempDir := filepath.Join(s.configDir, "temp")
	if err := os.RemoveAll(tempDir); err != nil {
		s.logger.Warn("Failed to cleanup temp directory", map[string]interface{}{"error": err.Error()})
	}

	return nil
}

// loadSettings loads settings from file or creates defaults
func (s *Service) loadSettings() error {
	// Try to load existing settings
	if _, err := os.Stat(s.configFile); err == nil {
		return s.loadSettingsFromFile()
	}

	// No settings file exists, create defaults
	s.logger.Info("No settings file found, creating defaults")
	s.settings = models.DefaultAppSettings()
	return s.saveSettings()
}

// loadSettingsFromFile loads settings from the configuration file
func (s *Service) loadSettingsFromFile() error {
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings models.AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		// Try to recover by loading defaults and showing warning
		s.logger.Warn("Failed to unmarshal settings, loading defaults", map[string]interface{}{"error": err.Error()})
		s.settings = models.DefaultAppSettings()
		return s.saveSettings()
	}

	// Validate loaded settings
	if err := settings.Validate(); err != nil {
		s.logger.Warn("Invalid settings detected, fixing automatically", map[string]interface{}{"error": err.Error()})
		fixedSettings := models.DefaultAppSettings()
		s.mergeSettings(fixedSettings, &settings)
		s.settings = fixedSettings
		return s.saveSettings()
	}

	s.settings = &settings
	s.logger.Info("Loaded settings from file", map[string]interface{}{"file": s.configFile})
	return nil
}

// saveSettings saves settings to the configuration file
func (s *Service) saveSettings() error {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to temporary file first
	tempFile := s.configFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary settings file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, s.configFile); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temporary settings file: %w", err)
	}

	s.logger.Debug("Settings saved", map[string]interface{}{"file": s.configFile})
	return nil
}

// mergeSettings merges two settings objects, with new settings taking precedence
func (s *Service) mergeSettings(existing, imported *models.AppSettings) *models.AppSettings {
	merged := *existing // Start with existing as base

	// Only merge imported fields that are not default/empty

	// Terminal settings
	if imported.Terminal.Shell != "" && imported.Terminal.Shell != models.DefaultAppSettings().Terminal.Shell {
		merged.Terminal.Shell = imported.Terminal.Shell
	}
	if imported.Terminal.FontSize > 0 && imported.Terminal.FontSize != models.DefaultAppSettings().Terminal.FontSize {
		merged.Terminal.FontSize = imported.Terminal.FontSize
	}
	if imported.Terminal.FontFamily != "" && imported.Terminal.FontFamily != models.DefaultAppSettings().Terminal.FontFamily {
		merged.Terminal.FontFamily = imported.Terminal.FontFamily
	}

	// Audio settings
	if imported.Audio.Enabled != models.DefaultAppSettings().Audio.Enabled {
		merged.Audio.Enabled = imported.Audio.Enabled
	}
	if imported.Audio.Volume != models.DefaultAppSettings().Audio.Volume {
		merged.Audio.Volume = imported.Audio.Volume
	}
	if imported.Audio.SoundTheme != "" && imported.Audio.SoundTheme != models.DefaultAppSettings().Audio.SoundTheme {
		merged.Audio.SoundTheme = imported.Audio.SoundTheme
	}

	// Theme settings
	if imported.Theme.Name != "" && imported.Theme.Name != models.DefaultAppSettings().Theme.Name {
		merged.Theme.Name = imported.Theme.Name
	}
	if imported.Theme.Background != "" && imported.Theme.Background != models.DefaultAppSettings().Theme.Background {
		merged.Theme.Background = imported.Theme.Background
	}
	if imported.Theme.Transparency != 0 && imported.Theme.Transparency != models.DefaultAppSettings().Theme.Transparency {
		merged.Theme.Transparency = imported.Theme.Transparency
	}

	// Display settings
	if len(imported.Display.WindowSize) == 2 && imported.Display.WindowSize[0] > 0 {
		merged.Display.WindowSize = imported.Display.WindowSize
	}
	if imported.Display.ScaleFactor > 0 && imported.Display.ScaleFactor != models.DefaultAppSettings().Display.ScaleFactor {
		merged.Display.ScaleFactor = imported.Display.ScaleFactor
	}

	// Filesystem settings
	if imported.Filesystem.ShowHiddenFiles != models.DefaultAppSettings().Filesystem.ShowHiddenFiles {
		merged.Filesystem.ShowHiddenFiles = imported.Filesystem.ShowHiddenFiles
	}
	if imported.Filesystem.LastLocation != "" {
		merged.Filesystem.LastLocation = imported.Filesystem.LastLocation
	}

	// System settings
	if imported.System.ShowIP != models.DefaultAppSettings().System.ShowIP {
		merged.System.ShowIP = imported.System.ShowIP
	}
	if imported.System.NetworkInterface != "" && imported.System.NetworkInterface != models.DefaultAppSettings().System.NetworkInterface {
		merged.System.NetworkInterface = imported.System.NetworkInterface
	}

	// General settings
	if imported.General.LastCWD != "" {
		merged.General.LastCWD = imported.General.LastCWD
	}
	if imported.General.StartupCommand != "" && imported.General.StartupCommand != models.DefaultAppSettings().General.StartupCommand {
		merged.General.StartupCommand = imported.General.StartupCommand
	}

	// Preserve legacy information
	if imported.Legacy != nil {
		merged.Legacy = imported.Legacy
	}

	merged.UpdatedAt = time.Now()
	return &merged
}

// ValidateSettings checks if current settings are valid
func (s *Service) ValidateSettings() error {
	s.settings.Lock()
	defer s.settings.Unlock()

	return s.settings.Validate()
}

// CreateBackup creates a backup of current settings
func (s *Service) CreateBackup() error {
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(s.configDir, fmt.Sprintf("settings-backup-%s.json", timestamp))

	return s.ExportSettings(backupFile)
}

// ListBackups returns a list of available backup files
func (s *Service) ListBackups() ([]string, error) {
	pattern := filepath.Join(s.configDir, "settings-backup-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob backup files: %w", err)
	}

	return matches, nil
}

// RestoreBackup restores settings from a backup file
func (s *Service) RestoreBackup(backupFile string) error {
	if !filepath.HasPrefix(backupFile, s.configDir) {
		return fmt.Errorf("backup file must be in config directory")
	}

	return s.ImportSettings(backupFile)
}
