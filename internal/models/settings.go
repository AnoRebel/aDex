package models

import (
	"sync"
	"time"
)

// AppSettings represents the complete application configuration
type AppSettings struct {
	sync.RWMutex `json:"-"`

	// Metadata
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`

	// General application settings
	General *GeneralSettings `json:"general"`

	// Terminal settings
	Terminal *TerminalSettings `json:"terminal"`

	// Display and window settings
	Display *DisplaySettings `json:"display"`

	// Audio settings
	Audio *AppAudioSettings `json:"audio"`

	// System monitoring settings
	System *SystemSettings `json:"system"`

	// File browser settings
	Filesystem *FilesystemSettings `json:"filesystem"`

	// Theme settings
	Theme *AppThemeSettings `json:"theme"`

	// Performance settings
	Performance *PerformanceSettings `json:"performance"`

	// Network settings (moved from system)
	Network *NetworkSettings `json:"network"`

	// User preferences
	UserPreferences *UserPreferences `json:"userPreferences"`

	// Legacy migration info
	Legacy *LegacyInfo `json:"legacy,omitempty"`
}

// CopyValue returns a copy of the settings without the embedded mutex.
//
// `*s` cannot be assigned directly: AppSettings embeds sync.RWMutex, so a
// plain struct copy duplicates the lock along with its current state — which
// `go vet` flags, and which would hand callers a mutex that is already held
// if the copy is taken under lock.
//
// The pointer fields are shared with the original by design; this is the same
// shallow copy the previous `copy := *s.settings` produced, minus the lock.
// Callers that need to mutate a nested group must replace that group rather
// than write through it.
func (s *AppSettings) CopyValue() AppSettings {
	if s == nil {
		return AppSettings{}
	}
	return AppSettings{
		Version:         s.Version,
		UpdatedAt:       s.UpdatedAt,
		CreatedAt:       s.CreatedAt,
		General:         s.General,
		Terminal:        s.Terminal,
		Display:         s.Display,
		Audio:           s.Audio,
		System:          s.System,
		Filesystem:      s.Filesystem,
		Theme:           s.Theme,
		Performance:     s.Performance,
		Network:         s.Network,
		UserPreferences: s.UserPreferences,
		Legacy:          s.Legacy,
	}
}

// GeneralSettings contains general application settings
type GeneralSettings struct {
	LastCWD            string `json:"lastCWD"`
	StartupCommand     string `json:"startupCommand"`
	ExitCommand        string `json:"exitCommand"`
	ExitAction         string `json:"exitAction"`
	AutoSaveSettings   bool   `json:"autoSaveSettings"`
	CheckUpdates       bool   `json:"checkUpdates"`
	TelemetryEnabled   bool   `json:"telemetryEnabled"`
	Language           string `json:"language"`
	Timezone           string `json:"timezone"`
	DateFormat         string `json:"dateFormat"`
	TimeFormat         string `json:"timeFormat"`
	AutoStart          bool   `json:"autoStart"`
	MinimizeToTray     bool   `json:"minimizeToTray"`
	ShowInTaskbar      bool   `json:"showInTaskbar"`
	RememberWindowSize bool   `json:"rememberWindowSize"`
}

// TerminalSettings contains terminal emulator configuration
type TerminalSettings struct {
	Shell              string   `json:"shell"`
	ShellArgs          []string `json:"shellArgs"`
	Profile            string   `json:"profile"`
	WorkingDir         string   `json:"workingDir"`
	FontSize           int      `json:"fontSize"`
	FontFamily         string   `json:"fontFamily"`
	LineHeight         float64  `json:"lineHeight"`
	LetterSpacing      float64  `json:"letterSpacing"`
	WordSpacing        float64  `json:"wordSpacing"`
	Opacity            float64  `json:"opacity"`
	Background         string   `json:"background"`
	CursorBlink        bool     `json:"cursorBlink"`
	CursorStyle        string   `json:"cursorStyle"`
	BackgroundBlur     bool     `json:"backgroundBlur"`
	MaxLines           int      `json:"maxLines"`
	Scrollback         int      `json:"scrollback"`
	ColorScheme        string   `json:"colorScheme"`
	BellEnabled        bool     `json:"bellEnabled"`
	CopyOnSelect       bool     `json:"copyOnSelect"`
	PasteOnMiddleClick bool     `json:"pasteOnMiddleClick"`
}

// DisplaySettings contains display and window configuration
type DisplaySettings struct {
	WindowSize            [2]int  `json:"windowSize"`
	WindowPosition        [2]int  `json:"windowPosition"`
	WindowMaximized       bool    `json:"windowMaximized"`
	ScaleFactor           float64 `json:"scaleFactor"`
	UseWindowFrame        bool    `json:"useWindowFrame"`
	Fullscreen            bool    `json:"fullscreen"`
	ShowFPS               bool    `json:"showFPS"`
	ShowZoomControls      bool    `json:"showZoomControls"`
	TransparentBackground bool    `json:"transparentBackground"`
	ShowMenuBar           bool    `json:"showMenuBar"`
	ShowStatusBar         bool    `json:"showStatusBar"`
	ShowTabBar            bool    `json:"showTabBar"`
	ShowNetworkGraph      bool    `json:"showNetworkGraph"`
	Animations            bool    `json:"animations"`
	Theme                 string  `json:"theme"`
	Language              string  `json:"language"`
}

// AppAudioSettings contains audio configuration for app settings
// Note: This is used in AppSettings, separate from the more detailed AudioSettings in audio.go
type AppAudioSettings struct {
	Enabled            bool              `json:"enabled"`
	Volume             int               `json:"volume"`
	StartupSound       bool              `json:"startupSound"`
	BellSound          bool              `json:"bellSound"`
	ErrorSound         bool              `json:"errorSound"`
	SoundTheme         string            `json:"soundTheme"`
	CustomSounds       map[string]string `json:"customSounds"`
	MasterVolume       float64           `json:"masterVolume"`
	EffectsVolume      float64           `json:"effectsVolume"`
	NotificationVolume float64           `json:"notificationVolume"`
	MuteInBackground   bool              `json:"muteInBackground"`
	Soundpack          string            `json:"soundpack"`
}

// SystemSettings contains system monitoring configuration
type SystemSettings struct {
	ShowIP           bool   `json:"showIP"`
	ShowDownload     bool   `json:"showDownload"`
	ShowUpload       bool   `json:"showUpload"`
	DownloadUnit     string `json:"downloadUnit"`
	UploadUnit       string `json:"uploadUnit"`
	NetworkInterface string `json:"networkInterface"`
	ShowCPU          bool   `json:"showCPU"`
	ShowMemory       bool   `json:"showMemory"`
	ShowDisk         bool   `json:"showDisk"`
	ShowTemperature  bool   `json:"showTemperature"`
	ShowProcesses    bool   `json:"showProcesses"`
	UpdateInterval   int    `json:"updateInterval"` // milliseconds
	HistoryLength    int    `json:"historyLength"`
	EnableGraphs     bool   `json:"enableGraphs"`
	ShowNetworkGraph bool   `json:"showNetworkGraph"`
}

// NetworkSettings contains network-specific settings
type NetworkSettings struct {
	AutoDetect         bool           `json:"autoDetect"`
	PreferredInterface string         `json:"preferredInterface"`
	GeoIPEnabled       bool           `json:"geoIPEnabled"`
	ShowLatency        bool           `json:"showLatency"`
	ShowPacketLoss     bool           `json:"showPacketLoss"`
	RefreshRate        int            `json:"refreshRate"` // milliseconds
	ConnectionTimeout  int            `json:"connectionTimeout"`
	ProxySettings      *ProxySettings `json:"proxySettings,omitempty"`
}

// ProxySettings contains proxy configuration
type ProxySettings struct {
	Enabled bool   `json:"enabled"`
	HTTP    string `json:"http"`
	HTTPS   string `json:"https"`
	SOCKS   string `json:"socks"`
	Exclude string `json:"exclude"`
}

// FilesystemSettings contains file browser configuration
type FilesystemSettings struct {
	ShowHiddenFiles   bool   `json:"showHiddenFiles"`
	LastLocation      string `json:"lastLocation"`
	ViewMode          string `json:"viewMode"`  // "list", "grid", "tree"
	SortBy            string `json:"sortBy"`    // "name", "size", "modified"
	SortOrder         string `json:"sortOrder"` // "asc", "desc"
	FollowTerminalCWD bool   `json:"followTerminalCWD"`
	PathInjection     bool   `json:"pathInjection"`
	ShowFileIcons     bool   `json:"showFileIcons"`
	ShowFileSize      bool   `json:"showFileSize"`
	ShowModifiedDate  bool   `json:"showModifiedDate"`
	ShowPermissions   bool   `json:"showPermissions"`
	DoubleClickAction string `json:"doubleClickAction"` // "open", "properties", "rename"
	MaxHistoryItems   int    `json:"maxHistoryItems"`
}

// AppThemeSettings contains theme and appearance configuration for app settings
// Note: This is used in AppSettings, separate from the more detailed ThemeSettings in theme.go
type AppThemeSettings struct {
	Name              string   `json:"name"`
	Background        string   `json:"background"`
	Transparency      float64  `json:"transparency"`
	BlurRadius        float64  `json:"blurRadius"`
	BorderColor       string   `json:"borderColor"`
	CursorColor       string   `json:"cursorColor"`
	GridColor         string   `json:"gridColor"`
	FontSmoothing     bool     `json:"fontSmoothing"`
	CustomCSS         string   `json:"customCSS"`
	CustomCSSModules  []string `json:"customCSSModules"`
	AutoTheme         bool     `json:"autoTheme"`
	DarkMode          bool     `json:"darkMode"`
	HighContrast      bool     `json:"highContrast"`
	ReducedMotion     bool     `json:"reducedMotion"`
	ColorBlindMode    string   `json:"colorBlindMode"` // "none", "protanopia", "deuteranopia", "tritanopia"
	IconTheme         string   `json:"iconTheme"`
	WindowDecorations string   `json:"windowDecorations"` // "native", "frameless", "custom"
	ColorScheme       string   `json:"colorScheme"`
	AccentColors      []string `json:"accentColors"`
}

// PerformanceSettings contains performance optimization settings
type PerformanceSettings struct {
	MaxCPULoad         int   `json:"maxCPULoad"`
	NetworkThrottle    int   `json:"networkThrottle"` // KB/s
	TerminalFPS        int   `json:"terminalFPS"`
	GPUAcceleration    bool  `json:"gpuAcceleration"`
	WebGLRenderer      bool  `json:"webglRenderer"`
	Animations         bool  `json:"animations"`
	Transparency       bool  `json:"transparency"`
	Vsync              bool  `json:"vsync"`
	MaxMemory          int64 `json:"maxMemory"` // bytes
	CacheEnabled       bool  `json:"cacheEnabled"`
	CacheSize          int64 `json:"cacheSize"` // bytes
	OptimizeForBattery bool  `json:"optimizeForBattery"`
	BackgroundThrottle bool  `json:"backgroundThrottle"`
	ReduceMotion       bool  `json:"reduceMotion"`
}

// UserPreferences contains user-specific preferences
type UserPreferences struct {
	StartupBehavior   string `json:"startupBehavior"`   // "maximize", "last-state", "custom"
	NotificationLevel string `json:"notificationLevel"` // "all", "errors", "none"
	DataCollection    bool   `json:"dataCollection"`
	CrashReporting    bool   `json:"crashReporting"`
	CheckForUpdates   bool   `json:"checkForUpdates"`
	AutoUpdate        bool   `json:"autoUpdate"`
	Telemetry         bool   `json:"telemetry"`
}

// LegacyInfo contains information about legacy configuration import
type LegacyInfo struct {
	Version      string    `json:"version"`
	ImportedAt   time.Time `json:"importedAt"`
	ImportedFrom string    `json:"importedFrom"`
	HasAudio     bool      `json:"hasAudio"`
	HasTheme     bool      `json:"hasTheme"`
}

// DefaultAppSettings returns the default application configuration
func DefaultAppSettings() *AppSettings {
	now := time.Now()
	return &AppSettings{
		Version:   "1.0.0",
		UpdatedAt: now,
		CreatedAt: now,
		General: &GeneralSettings{
			LastCWD:            "",
			StartupCommand:     "",
			ExitCommand:        "exit",
			ExitAction:         "close",
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
		},
		Terminal: &TerminalSettings{
			Shell:              "/bin/bash",
			ShellArgs:          []string{},
			Profile:            "",
			WorkingDir:         "",
			FontSize:           14,
			FontFamily:         "JetBrains Mono",
			LineHeight:         1.5,
			LetterSpacing:      0.0,
			WordSpacing:        0.0,
			Opacity:            1.0,
			Background:         "transparent",
			CursorBlink:        true,
			CursorStyle:        "block",
			BackgroundBlur:     false,
			MaxLines:           1000,
			Scrollback:         10000,
			ColorScheme:        "default",
			BellEnabled:        true,
			CopyOnSelect:       false,
			PasteOnMiddleClick: false,
		},
		Display: &DisplaySettings{
			WindowSize:            [2]int{1200, 800},
			WindowPosition:        [2]int{100, 100},
			WindowMaximized:       false,
			ScaleFactor:           1.0,
			UseWindowFrame:        true,
			Fullscreen:            false,
			ShowFPS:               false,
			ShowZoomControls:      false,
			TransparentBackground: false,
			ShowMenuBar:           true,
			ShowStatusBar:         true,
			ShowTabBar:            true,
			ShowNetworkGraph:      true,
			Animations:            true,
			Theme:                 "default",
			Language:              "en",
		},
		Audio: &AppAudioSettings{
			Enabled:            false, // Disabled by default
			Volume:             50,
			StartupSound:       true,
			BellSound:          true,
			ErrorSound:         true,
			SoundTheme:         "default",
			CustomSounds:       make(map[string]string),
			MasterVolume:       0.5,
			EffectsVolume:      0.7,
			NotificationVolume: 0.5,
			MuteInBackground:   false,
			Soundpack:          "default",
		},
		System: &SystemSettings{
			ShowIP:           true,
			ShowDownload:     true,
			ShowUpload:       true,
			DownloadUnit:     "auto",
			UploadUnit:       "auto",
			NetworkInterface: "",
			ShowCPU:          true,
			ShowMemory:       true,
			ShowDisk:         true,
			ShowTemperature:  true,
			ShowProcesses:    true,
			UpdateInterval:   1000,
			HistoryLength:    60,
			EnableGraphs:     true,
			ShowNetworkGraph: true,
		},
		Network: &NetworkSettings{
			AutoDetect:         true,
			PreferredInterface: "",
			GeoIPEnabled:       false,
			ShowLatency:        false,
			ShowPacketLoss:     false,
			RefreshRate:        1000,
			ConnectionTimeout:  5000,
		},
		Filesystem: &FilesystemSettings{
			ShowHiddenFiles:   false,
			LastLocation:      "",
			ViewMode:          "list",
			SortBy:            "name",
			SortOrder:         "asc",
			FollowTerminalCWD: true,
			PathInjection:     true,
			ShowFileIcons:     true,
			ShowFileSize:      true,
			ShowModifiedDate:  true,
			ShowPermissions:   false,
			DoubleClickAction: "open",
			MaxHistoryItems:   50,
		},
		Theme: &AppThemeSettings{
			Name:              "default",
			Background:        "transparent",
			Transparency:      1.0,
			BlurRadius:        0.0,
			BorderColor:       "#333333",
			CursorColor:       "#00ff00",
			GridColor:         "#003300",
			FontSmoothing:     true,
			CustomCSS:         "",
			CustomCSSModules:  []string{},
			AutoTheme:         false,
			DarkMode:          true,
			HighContrast:      false,
			ReducedMotion:     false,
			ColorBlindMode:    "none",
			IconTheme:         "default",
			WindowDecorations: "native",
			ColorScheme:       "",
			AccentColors:      []string{},
		},
		Performance: &PerformanceSettings{
			MaxCPULoad:         30,
			NetworkThrottle:    1000,
			TerminalFPS:        60,
			GPUAcceleration:    true,
			WebGLRenderer:      true,
			Animations:         true,
			Transparency:       false,
			Vsync:              false,
			MaxMemory:          512 * 1024 * 1024, // 512MB
			CacheEnabled:       true,
			CacheSize:          100 * 1024 * 1024, // 100MB
			OptimizeForBattery: false,
			BackgroundThrottle: true,
			ReduceMotion:       false,
		},
		UserPreferences: &UserPreferences{
			StartupBehavior:   "last-state",
			NotificationLevel: "errors",
			DataCollection:    false,
			CrashReporting:    false,
			CheckForUpdates:   false,
			AutoUpdate:        false,
			Telemetry:         false,
		},
	}
}

// Validate checks if the settings are valid
func (s *AppSettings) Validate() error {
	// Validate general settings
	if s.General == nil {
		return SettingsValidationError{Field: "general", Message: "general settings cannot be nil"}
	}

	// Validate terminal settings
	if s.Terminal != nil {
		if s.Terminal.FontSize < 6 || s.Terminal.FontSize > 72 {
			return SettingsValidationError{Field: "terminal.fontSize", Message: "font size must be between 6 and 72"}
		}
		if s.Terminal.Opacity < 0 || s.Terminal.Opacity > 1 {
			return SettingsValidationError{Field: "terminal.opacity", Message: "opacity must be between 0 and 1"}
		}
	}

	// Validate audio settings
	if s.Audio != nil {
		if s.Audio.Volume < 0 || s.Audio.Volume > 100 {
			return SettingsValidationError{Field: "audio.volume", Message: "volume must be between 0 and 100"}
		}
		if s.Audio.MasterVolume < 0 || s.Audio.MasterVolume > 1 {
			return SettingsValidationError{Field: "audio.masterVolume", Message: "master volume must be between 0 and 1"}
		}
	}

	// Validate display settings
	if s.Display != nil {
		if s.Display.ScaleFactor <= 0 {
			return SettingsValidationError{Field: "display.scaleFactor", Message: "scale factor must be positive"}
		}
		if len(s.Display.WindowSize) != 2 || s.Display.WindowSize[0] <= 0 || s.Display.WindowSize[1] <= 0 {
			return SettingsValidationError{Field: "display.windowSize", Message: "window size must be [width, height] with positive values"}
		}
	}

	// Validate performance settings
	if s.Performance != nil {
		if s.Performance.MaxCPULoad < 0 || s.Performance.MaxCPULoad > 100 {
			return SettingsValidationError{Field: "performance.maxCPULoad", Message: "max CPU load must be between 0 and 100"}
		}
		if s.Performance.TerminalFPS < 1 || s.Performance.TerminalFPS > 120 {
			return SettingsValidationError{Field: "performance.terminalFPS", Message: "terminal FPS must be between 1 and 120"}
		}
	}

	return nil
}

// SettingsValidationError represents a settings validation error
// Note: This is distinct from the ValidationError in errors.go
type SettingsValidationError struct {
	Field   string
	Message string
}

func (e SettingsValidationError) Error() string {
	return "validation error in " + e.Field + ": " + e.Message
}
