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

// Migration handles importing settings from legacy eDEX-UI configurations
type Migration struct {
	logger  *logger.Logger
	legacy  *LegacyConfig
	current *models.AppSettings
}

// LegacyConfig represents eDEX-UI configuration structure
type LegacyConfig struct {
	Version       string                  `json:"version"`
	General       LegacyGeneralConfig     `json:"general"`
	Terminal      LegacyTerminalConfig    `json:"terminal"`
	Display       LegacyDisplayConfig     `json:"display"`
	Network       LegacyNetworkConfig     `json:"network"`
	FileBrowser   LegacyFileBrowserConfig `json:"filebrowser"`
	Customization LegacyCustomization     `json:"customization"`
	Performance   LegacyPerformanceConfig `json:"performance"`
	Audio         LegacyAudioConfig       `json:"audio"`
}

// LegacyGeneralConfig represents general settings from legacy config
type LegacyGeneralConfig struct {
	LastCWD        string `json:"lastCWD"`
	StartupCommand string `json:"startupCommand"`
	ExitCommand    string `json:"exitCommand"`
	ExitAction     string `json:"exitAction"`
}

// LegacyTerminalConfig represents terminal settings from legacy config
type LegacyTerminalConfig struct {
	Shell          string   `json:"shell"`
	ShellArgs      []string `json:"shellArgs"`
	Profile        string   `json:"profile"`
	WorkingDir     string   `json:"workingDir"`
	CursorBlink    bool     `json:"cursorBlink"`
	CursorStyle    string   `json:"cursorStyle"`
	FontSize       int      `json:"fontSize"`
	FontFamily     string   `json:"fontFamily"`
	LineHeight     float64  `json:"lineHeight"`
	LetterSpacing  float64  `json:"letterSpacing"`
	WordSpacing    float64  `json:"wordSpacing"`
	Opacity        float64  `json:"opacity"`
	Background     string   `json:"background"`
	BackgroundBlur bool     `json:"backgroundBlur"`
}

// LegacyDisplayConfig represents display settings from legacy config
type LegacyDisplayConfig struct {
	WindowSize            [2]int  `json:"windowSize"`
	WindowPosition        [2]int  `json:"windowPosition"`
	WindowMaximized       bool    `json:"windowMaximized"`
	ScaleFactor           float64 `json:"scaleFactor"`
	UseWindowFrame        bool    `json:"useWindowFrame"`
	Fullscreen            bool    `json:"fullscreen"`
	ShowFPS               bool    `json:"showFPS"`
	ShowZoomControls      bool    `json:"showZoomControls"`
	NetworkGraph          bool    `json:"networkGraph"`
	TransparentBackground bool    `json:"transparentBackground"`
}

// LegacyNetworkConfig represents network settings from legacy config
type LegacyNetworkConfig struct {
	ShowIP           bool   `json:"showIP"`
	ShowDownload     bool   `json:"showDownload"`
	ShowUpload       bool   `json:"showUpload"`
	DownloadUnit     string `json:"downloadUnit"`
	UploadUnit       string `json:"uploadUnit"`
	NetworkInterface string `json:"networkInterface"`
}

// LegacyFileBrowserConfig represents file browser settings from legacy config
type LegacyFileBrowserConfig struct {
	ShowHiddenFiles bool   `json:"showHiddenFiles"`
	LastLocation    string `json:"lastLocation"`
	ViewMode        string `json:"viewMode"`
	SortBy          string `json:"sortBy"`
	SortOrder       string `json:"sortOrder"`
}

// LegacyCustomization represents theme customization from legacy config
type LegacyCustomization struct {
	ThemeName         string            `json:"themeName"`
	ThemeBackground   string            `json:"themeBackground"`
	ThemeTransparency float64           `json:"themeTransparency"`
	ThemeBlurRadius   float64           `json:"themeBlurRadius"`
	ThemeBorderColor  string            `json:"themeBorderColor"`
	ThemeCursorColor  string            `json:"themeCursorColor"`
	ThemeGridColor    string            `json:"themeGridColor"`
	FontSmoothing     bool              `json:"fontSmoothing"`
	CustomCSS         string            `json:"customCSS"`
	CustomCSSModules  []string          `json:"customCSSModules"`
	ThemeVariants     map[string]string `json:"themeVariants"`
}

// LegacyPerformanceConfig represents performance settings from legacy config
type LegacyPerformanceConfig struct {
	MaxCPULoad      int  `json:"maxCPULoad"`
	NetworkThrottle int  `json:"networkThrottle"`
	TerminalFPS     int  `json:"terminalFPS"`
	GPUAcceleration bool `json:"gpuAcceleration"`
	WebGLRenderer   bool `json:"webglRenderer"`
	Animations      bool `json:"animations"`
	Transparency    bool `json:"transparency"`
	Vsync           bool `json:"vsync"`
}

// LegacyAudioConfig represents audio settings from legacy config
type LegacyAudioConfig struct {
	Enabled      bool              `json:"enabled"`
	Volume       int               `json:"volume"`
	StartupSound bool              `json:"startupSound"`
	BellSound    bool              `json:"bellSound"`
	ErrorSound   bool              `json:"errorSound"`
	SoundTheme   string            `json:"soundTheme"`
	CustomSounds map[string]string `json:"customSounds"`
}

// NewMigration creates a new migration service
func NewMigration(logger *logger.Logger) *Migration {
	return &Migration{
		logger: logger,
	}
}

// FindLegacyConfigs searches for legacy eDEX-UI configuration files
func (m *Migration) FindLegacyConfigs() ([]string, error) {
	var configs []string

	// Common eDEX-UI configuration locations
	paths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "edex-ui"),
		filepath.Join(os.Getenv("APPDATA"), "eDEX-UI"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "eDEX-UI"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			configs = append(configs, path)
			m.logger.Info("Found legacy eDEX-UI config", map[string]interface{}{"path": path})
		}
	}

	return configs, nil
}

// ImportLegacyConfig imports settings from legacy configuration
func (m *Migration) ImportLegacyConfig(configPath string) (*models.AppSettings, error) {
	m.logger.Info("Importing legacy configuration", map[string]interface{}{"path": configPath})

	// Load legacy configuration
	legacy, err := m.loadLegacyConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load legacy config: %w", err)
	}

	// Convert to modern settings
	modern, err := m.convertLegacySettings(legacy)
	if err != nil {
		return nil, fmt.Errorf("failed to convert legacy settings: %w", err)
	}

	// Validate converted settings
	if err := modern.Validate(); err != nil {
		return nil, fmt.Errorf("invalid converted settings: %w", err)
	}

	m.logger.Info("Successfully imported legacy configuration", map[string]interface{}{
		"version": legacy.Version,
		"themes":  len(legacy.Customization.ThemeVariants),
	})

	return modern, nil
}

// loadLegacyConfig loads legacy eDEX-UI configuration from file
func (m *Migration) loadLegacyConfig(configPath string) (*LegacyConfig, error) {
	configFile := filepath.Join(configPath, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config LegacyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// convertLegacySettings converts legacy configuration to modern settings
func (m *Migration) convertLegacySettings(legacy *LegacyConfig) (*models.AppSettings, error) {
	settings := &models.AppSettings{
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	// Terminal settings
	settings.Terminal = convertTerminalSettings(legacy.Terminal)

	// Display settings
	settings.Display = convertDisplaySettings(legacy.Display)

	// Audio settings
	settings.Audio = convertAudioSettings(legacy.Audio)

	// System monitoring settings
	settings.System = convertSystemSettings(legacy.Network)

	// File browser settings
	settings.Filesystem = convertFileBrowserSettings(legacy.FileBrowser)

	// Theme settings
	settings.Theme = convertThemeSettings(legacy.Customization)

	// Performance settings
	settings.Performance = convertPerformanceSettings(legacy.Performance)

	// General settings
	settings.General = convertGeneralSettings(legacy.General)

	// Mark as migrated from legacy
	settings.Legacy = &models.LegacyInfo{
		Version:      legacy.Version,
		ImportedAt:   time.Now(),
		ImportedFrom: "eDEX-UI",
		HasAudio:     legacy.Audio.Enabled,
		HasTheme:     legacy.Customization.ThemeName != "",
	}

	return settings, nil
}

// convertTerminalSettings converts legacy terminal settings
func convertTerminalSettings(legacy LegacyTerminalConfig) *models.TerminalSettings {
	return &models.TerminalSettings{
		Shell:          legacy.Shell,
		ShellArgs:      legacy.ShellArgs,
		Profile:        legacy.Profile,
		WorkingDir:     legacy.WorkingDir,
		FontSize:       legacy.FontSize,
		FontFamily:     legacy.FontFamily,
		LineHeight:     legacy.LineHeight,
		LetterSpacing:  legacy.LetterSpacing,
		WordSpacing:    legacy.WordSpacing,
		Opacity:        legacy.Opacity,
		Background:     legacy.Background,
		CursorBlink:    legacy.CursorBlink,
		CursorStyle:    legacy.CursorStyle,
		BackgroundBlur: legacy.BackgroundBlur,
		// Modern additions
		MaxLines:           1000,
		Scrollback:         10000,
		ColorScheme:        "default",
		BellEnabled:        true,
		CopyOnSelect:       false,
		PasteOnMiddleClick: false,
	}
}

// convertDisplaySettings converts legacy display settings
func convertDisplaySettings(legacy LegacyDisplayConfig) *models.DisplaySettings {
	return &models.DisplaySettings{
		WindowSize:            legacy.WindowSize,
		WindowPosition:        legacy.WindowPosition,
		WindowMaximized:       legacy.WindowMaximized,
		ScaleFactor:           legacy.ScaleFactor,
		UseWindowFrame:        legacy.UseWindowFrame,
		Fullscreen:            legacy.Fullscreen,
		ShowFPS:               legacy.ShowFPS,
		ShowZoomControls:      legacy.ShowZoomControls,
		TransparentBackground: legacy.TransparentBackground,
		// Modern additions
		ShowMenuBar:      true,
		ShowStatusBar:    true,
		ShowTabBar:       true,
		ShowNetworkGraph: legacy.NetworkGraph,
		Animations:       true,
		Theme:            "",
		Language:         "en",
	}
}

// convertAudioSettings converts legacy audio settings
func convertAudioSettings(legacy LegacyAudioConfig) *models.AppAudioSettings {
	return &models.AppAudioSettings{
		Enabled:      legacy.Enabled,
		Volume:       legacy.Volume,
		StartupSound: legacy.StartupSound,
		BellSound:    legacy.BellSound,
		ErrorSound:   legacy.ErrorSound,
		SoundTheme:   legacy.SoundTheme,
		CustomSounds: legacy.CustomSounds,
		// Modern additions
		MasterVolume:       float64(legacy.Volume) / 100.0,
		EffectsVolume:      0.7,
		NotificationVolume: 0.5,
		MuteInBackground:   false,
		Soundpack:          "default",
	}
}

// convertSystemSettings converts legacy system monitoring settings
func convertSystemSettings(legacy LegacyNetworkConfig) *models.SystemSettings {
	return &models.SystemSettings{
		ShowIP:           legacy.ShowIP,
		ShowDownload:     legacy.ShowDownload,
		ShowUpload:       legacy.ShowUpload,
		DownloadUnit:     legacy.DownloadUnit,
		UploadUnit:       legacy.UploadUnit,
		NetworkInterface: legacy.NetworkInterface,
		// Modern additions
		ShowCPU:          true,
		ShowMemory:       true,
		ShowDisk:         true,
		ShowTemperature:  true,
		ShowProcesses:    true,
		UpdateInterval:   1000,
		HistoryLength:    60,
		EnableGraphs:     true,
		ShowNetworkGraph: false,
	}
}

// convertFileBrowserSettings converts legacy file browser settings
func convertFileBrowserSettings(legacy LegacyFileBrowserConfig) *models.FilesystemSettings {
	return &models.FilesystemSettings{
		ShowHiddenFiles: legacy.ShowHiddenFiles,
		LastLocation:    legacy.LastLocation,
		ViewMode:        legacy.ViewMode,
		SortBy:          legacy.SortBy,
		SortOrder:       legacy.SortOrder,
		// Modern additions
		FollowTerminalCWD: true,
		PathInjection:     true,
		ShowFileIcons:     true,
		ShowFileSize:      true,
		ShowModifiedDate:  true,
		ShowPermissions:   false,
		DoubleClickAction: "open",
		MaxHistoryItems:   50,
	}
}

// convertThemeSettings converts legacy theme settings
func convertThemeSettings(legacy LegacyCustomization) *models.AppThemeSettings {
	themeSettings := &models.AppThemeSettings{
		Name:             legacy.ThemeName,
		Background:       legacy.ThemeBackground,
		Transparency:     legacy.ThemeTransparency,
		BlurRadius:       legacy.ThemeBlurRadius,
		BorderColor:      legacy.ThemeBorderColor,
		CursorColor:      legacy.ThemeCursorColor,
		GridColor:        legacy.ThemeGridColor,
		FontSmoothing:    legacy.FontSmoothing,
		CustomCSS:        legacy.CustomCSS,
		CustomCSSModules: legacy.CustomCSSModules,
		// Modern additions
		AutoTheme:         false,
		DarkMode:          true,
		HighContrast:      false,
		ReducedMotion:     false,
		ColorBlindMode:    "none",
		IconTheme:         "default",
		WindowDecorations: "native",
	}

	// Convert theme variants to color scheme
	if len(legacy.ThemeVariants) > 0 {
		themeSettings.ColorScheme = legacy.ThemeVariants["colors"]
		themeSettings.AccentColors = []string{
			legacy.ThemeVariants["primary"],
			legacy.ThemeVariants["secondary"],
			legacy.ThemeVariants["tertiary"],
		}
	}

	return themeSettings
}

// convertPerformanceSettings converts legacy performance settings
func convertPerformanceSettings(legacy LegacyPerformanceConfig) *models.PerformanceSettings {
	return &models.PerformanceSettings{
		MaxCPULoad:      legacy.MaxCPULoad,
		NetworkThrottle: legacy.NetworkThrottle,
		TerminalFPS:     legacy.TerminalFPS,
		GPUAcceleration: legacy.GPUAcceleration,
		WebGLRenderer:   legacy.WebGLRenderer,
		Animations:      legacy.Animations,
		Transparency:    legacy.Transparency,
		Vsync:           legacy.Vsync,
		// Modern additions
		MaxMemory:          512 * 1024 * 1024, // 512MB
		CacheEnabled:       true,
		CacheSize:          100 * 1024 * 1024, // 100MB
		OptimizeForBattery: false,
		BackgroundThrottle: true,
		ReduceMotion:       false,
	}
}

// convertGeneralSettings converts legacy general settings
func convertGeneralSettings(legacy LegacyGeneralConfig) *models.GeneralSettings {
	return &models.GeneralSettings{
		LastCWD:        legacy.LastCWD,
		StartupCommand: legacy.StartupCommand,
		ExitCommand:    legacy.ExitCommand,
		ExitAction:     legacy.ExitAction,
		// Modern additions
		AutoSaveSettings:   true,
		CheckUpdates:       false,
		TelemetryEnabled:   false,
		Language:           "en",
		Timezone:           "auto",
		DateFormat:         "YYYY-MM-DD",
		TimeFormat:         "24h",
		AutoStart:          false,
		MinimizeToTray:     true,
		ShowInTaskbar:      true,
		RememberWindowSize: true,
	}
}

// ExportLegacyData exports current settings in legacy format
func (m *Migration) ExportLegacyData(settings *models.AppSettings) (*LegacyConfig, error) {
	legacy := &LegacyConfig{
		Version: "2.3.0", // eDEX-UI version this would be compatible with
	}

	// Convert back to legacy format
	legacy.Terminal = LegacyTerminalConfig{
		Shell:          settings.Terminal.Shell,
		ShellArgs:      settings.Terminal.ShellArgs,
		Profile:        settings.Terminal.Profile,
		WorkingDir:     settings.Terminal.WorkingDir,
		FontSize:       settings.Terminal.FontSize,
		FontFamily:     settings.Terminal.FontFamily,
		LineHeight:     settings.Terminal.LineHeight,
		LetterSpacing:  settings.Terminal.LetterSpacing,
		WordSpacing:    settings.Terminal.WordSpacing,
		Opacity:        settings.Terminal.Opacity,
		Background:     settings.Terminal.Background,
		CursorBlink:    settings.Terminal.CursorBlink,
		CursorStyle:    settings.Terminal.CursorStyle,
		BackgroundBlur: settings.Terminal.BackgroundBlur,
	}

	legacy.Display = LegacyDisplayConfig{
		WindowSize:            settings.Display.WindowSize,
		WindowPosition:        settings.Display.WindowPosition,
		WindowMaximized:       settings.Display.WindowMaximized,
		ScaleFactor:           settings.Display.ScaleFactor,
		UseWindowFrame:        settings.Display.UseWindowFrame,
		Fullscreen:            settings.Display.Fullscreen,
		ShowFPS:               settings.Display.ShowFPS,
		ShowZoomControls:      settings.Display.ShowZoomControls,
		NetworkGraph:          settings.Display.ShowNetworkGraph,
		TransparentBackground: settings.Display.TransparentBackground,
	}

	legacy.Audio = LegacyAudioConfig{
		Enabled:      settings.Audio.Enabled,
		Volume:       settings.Audio.Volume,
		StartupSound: settings.Audio.StartupSound,
		BellSound:    settings.Audio.BellSound,
		ErrorSound:   settings.Audio.ErrorSound,
		SoundTheme:   settings.Audio.SoundTheme,
		CustomSounds: settings.Audio.CustomSounds,
	}

	legacy.Network = LegacyNetworkConfig{
		ShowIP:           settings.System.ShowIP,
		ShowDownload:     settings.System.ShowDownload,
		ShowUpload:       settings.System.ShowUpload,
		DownloadUnit:     settings.System.DownloadUnit,
		UploadUnit:       settings.System.UploadUnit,
		NetworkInterface: settings.System.NetworkInterface,
	}

	legacy.FileBrowser = LegacyFileBrowserConfig{
		ShowHiddenFiles: settings.Filesystem.ShowHiddenFiles,
		LastLocation:    settings.Filesystem.LastLocation,
		ViewMode:        settings.Filesystem.ViewMode,
		SortBy:          settings.Filesystem.SortBy,
		SortOrder:       settings.Filesystem.SortOrder,
	}

	legacy.Customization = LegacyCustomization{
		ThemeName:         settings.Theme.Name,
		ThemeBackground:   settings.Theme.Background,
		ThemeTransparency: settings.Theme.Transparency,
		ThemeBlurRadius:   settings.Theme.BlurRadius,
		ThemeBorderColor:  settings.Theme.BorderColor,
		ThemeCursorColor:  settings.Theme.CursorColor,
		ThemeGridColor:    settings.Theme.GridColor,
		FontSmoothing:     settings.Theme.FontSmoothing,
		CustomCSS:         settings.Theme.CustomCSS,
		CustomCSSModules:  settings.Theme.CustomCSSModules,
		ThemeVariants:     make(map[string]string),
	}

	// Set theme variants
	if settings.Theme.ColorScheme != "" {
		legacy.Customization.ThemeVariants["colors"] = settings.Theme.ColorScheme
	}

	legacy.Performance = LegacyPerformanceConfig{
		MaxCPULoad:      settings.Performance.MaxCPULoad,
		NetworkThrottle: settings.Performance.NetworkThrottle,
		TerminalFPS:     settings.Performance.TerminalFPS,
		GPUAcceleration: settings.Performance.GPUAcceleration,
		WebGLRenderer:   settings.Performance.WebGLRenderer,
		Animations:      settings.Performance.Animations,
		Transparency:    settings.Performance.Transparency,
		Vsync:           settings.Performance.Vsync,
	}

	legacy.General = LegacyGeneralConfig{
		LastCWD:        settings.General.LastCWD,
		StartupCommand: settings.General.StartupCommand,
		ExitCommand:    settings.General.ExitCommand,
		ExitAction:     settings.General.ExitAction,
	}

	return legacy, nil
}

// ValidateLegacyConfig checks if legacy configuration is valid
func (m *Migration) ValidateLegacyConfig(config *LegacyConfig) error {
	if config == nil {
		return fmt.Errorf("legacy config is nil")
	}

	// Validate required fields
	if config.Version == "" {
		return fmt.Errorf("legacy config missing version")
	}

	// Validate terminal settings
	if config.Terminal.Shell == "" {
		return fmt.Errorf("terminal shell path is required")
	}

	// Validate audio settings
	if config.Audio.Volume < 0 || config.Audio.Volume > 100 {
		return fmt.Errorf("audio volume must be between 0 and 100")
	}

	// Validate display settings
	if config.Display.ScaleFactor <= 0 {
		return fmt.Errorf("display scale factor must be positive")
	}

	return nil
}

// CleanLegacyFiles removes legacy configuration files after successful migration
func (m *Migration) CleanLegacyFiles(configPath string) error {
	// Don't actually remove files by default, just log what would be cleaned
	m.logger.Info("Legacy migration complete", map[string]interface{}{"config": configPath})
	m.logger.Info("To clean legacy files, manually remove the configuration directory", nil)

	return nil
}
