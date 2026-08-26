package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
	"aDex-UI/internal/utils"
)

// Service handles configuration management
type Service struct {
	platform *utils.FeatureDetection
	config   *Config
	mu       sync.RWMutex
	configPath string
}

// NewService creates a new config service instance
func NewService() *Service {
	platform := utils.DetectPlatform()
	configDir := getConfigDir(platform)

	service := &Service{
		platform:   platform,
		configPath: filepath.Join(configDir, "config.yaml"),
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		// If we can't create config directory, use in-memory config
		service.config = service.getDefaultConfig()
		return service
	}

	// Load existing config or create default
	if err := service.loadConfig(); err != nil {
		service.config = service.getDefaultConfig()
	}

	return service
}

// getConfigDir returns the appropriate config directory for the platform
func getConfigDir(platform *utils.FeatureDetection) string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "aDex-UI")
	}

	home := platform.GetHomeDirectory()
	switch platform.Platform.OS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "aDex-UI")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "aDex-UI")
	default:
		return filepath.Join(home, ".config", "aDex-UI")
	}
}

// Config represents application configuration
type Config struct {
	AppName    string                 `json:"app_name"`
	Version    string                 `json:"version"`
	Settings   map[string]interface{} `json:"settings"`
	Theme      ThemeConfig            `json:"theme"`
	Terminal   TerminalConfig         `json:"terminal"`
	System     SystemConfig           `json:"system"`
	Filesystem FilesystemConfig       `json:"filesystem"`
	Audio      AudioConfig            `json:"audio"`
}

// ThemeConfig represents theme configuration
type ThemeConfig struct {
	Name        string `json:"name"`
	IsDark      bool   `json:"is_dark"`
	AccentColor string `json:"accent_color"`
}

// TerminalConfig represents terminal configuration
type TerminalConfig struct {
	Shell       string `json:"shell"`
	FontSize    int    `json:"font_size"`
	FontFamily  string `json:"font_family"`
	Opacity     int    `json:"opacity"`
	Scrollback  int    `json:"scrollback"`
}

// SystemConfig represents system monitoring configuration
type SystemConfig struct {
	UpdateInterval    int  `json:"update_interval"`
	EnableMonitoring  bool `json:"enable_monitoring"`
	ShowProcesses     bool `json:"show_processes"`
	MaxProcesses      int  `json:"max_processes"`
}

// FilesystemConfig represents filesystem configuration
type FilesystemConfig struct {
	ShowHiddenFiles bool     `json:"show_hidden_files"`
	DefaultPath     string   `json:"default_path"`
	Bookmarks       []string `json:"bookmarks"`
}

// AudioConfig represents audio configuration
type AudioConfig struct {
	EnableNotifications bool `json:"enable_notifications"`
	DefaultVolume       int  `json:"default_volume"`
	EnableEqualizer     bool `json:"enable_equalizer"`
}

// loadConfig loads configuration from the default path
func (s *Service) loadConfig() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	s.config = &config
	return nil
}

// getDefaultConfig returns the default configuration
func (s *Service) getDefaultConfig() *Config {
	return &Config{
		AppName:  "aDex-UI",
		Version:  "1.0.0",
		Settings: make(map[string]interface{}),
		Theme: ThemeConfig{
			Name:        "Default Dark",
			IsDark:      true,
			AccentColor: "#00ff00",
		},
		Terminal: TerminalConfig{
			Shell:      s.getDefaultShell(),
			FontSize:   14,
			FontFamily: "Consolas",
			Opacity:    90,
			Scrollback: 10000,
		},
		System: SystemConfig{
			UpdateInterval:   5,
			EnableMonitoring: true,
			ShowProcesses:    true,
			MaxProcesses:     20,
		},
		Filesystem: FilesystemConfig{
			ShowHiddenFiles: false,
			DefaultPath:     s.platform.GetHomeDirectory(),
			Bookmarks:       []string{},
		},
		Audio: AudioConfig{
			EnableNotifications: true,
			DefaultVolume:       70,
			EnableEqualizer:     false,
		},
	}
}

// getDefaultShell returns the default shell for the platform
func (s *Service) getDefaultShell() string {
	switch s.platform.Platform.OS {
	case "windows":
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			return "powershell.exe"
		}
		return "cmd.exe"
	case "darwin":
		if _, err := exec.LookPath("zsh"); err == nil {
			return "zsh"
		}
		return "bash"
	default:
		if shell := os.Getenv("SHELL"); shell != "" {
			return shell
		}
		return "bash"
	}
}

// LoadConfig loads configuration from file
func (s *Service) LoadConfig(ctx context.Context, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	s.mu.Lock()
	s.config = &config
	s.mu.Unlock()

	return &config, nil
}

// SaveConfig saves configuration to file
func (s *Service) SaveConfig(ctx context.Context, config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	s.mu.Lock()
	s.config = config
	s.mu.Unlock()

	return nil
}

// GetSetting retrieves a specific setting value
func (s *Service) GetSetting(ctx context.Context, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Check if it's a top-level setting
	switch key {
	case "app_name":
		return s.config.AppName, nil
	case "version":
		return s.config.Version, nil
	default:
		// Check in general settings
		if value, exists := s.config.Settings[key]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("setting not found: %s", key)
	}
}

// SetSetting updates a specific setting value
func (s *Service) SetSetting(ctx context.Context, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return fmt.Errorf("config not loaded")
	}

	// Update general settings
	if s.config.Settings == nil {
		s.config.Settings = make(map[string]interface{})
	}
	s.config.Settings[key] = value

	// Auto-save to default path
	return s.saveCurrentConfig()
}

// ResetToDefaults resets configuration to default values
func (s *Service) ResetToDefaults(ctx context.Context) (*Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = s.getDefaultConfig()

	// Save the default config
	if err := s.saveCurrentConfig(); err != nil {
		return nil, err
	}

	return s.config, nil
}

// GetConfig returns the current configuration
func (s *Service) GetConfig(ctx context.Context) (*Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Return a copy to prevent external modification
	configCopy := *s.config
	return &configCopy, nil
}

// SaveCurrentConfig saves the current configuration to the default path
func (s *Service) saveCurrentConfig() error {
	data, err := yaml.Marshal(s.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// ExportConfig exports configuration to specified format
func (s *Service) ExportConfig(ctx context.Context, format string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	switch format {
	case "json":
		return json.MarshalIndent(s.config, "", "  ")
	case "yaml":
		return yaml.Marshal(s.config)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ImportConfig imports configuration from data
func (s *Service) ImportConfig(ctx context.Context, data []byte, format string) (*Config, error) {
	var config Config

	switch format {
	case "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}

	s.mu.Lock()
	s.config = &config
	s.mu.Unlock()

	// Save imported config
	if err := s.saveCurrentConfig(); err != nil {
		return nil, err
	}

	return &config, nil
}
