package models

import "time"

// AppConfig represents the main application configuration
type AppConfig struct {
	App struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Author      string `json:"author"`
		License     string `json:"license"`
		Homepage    string `json:"homepage"`
	} `json:"app"`
	
	Server struct {
		Host         string `json:"host"`
		Port         int    `json:"port"`
		ReadTimeout  int    `json:"read_timeout"`
		WriteTimeout int    `json:"write_timeout"`
		IdleTimeout  int    `json:"idle_timeout"`
		MaxHeaderBytes int  `json:"max_header_bytes"`
		EnableHTTPS  bool   `json:"enable_https"`
		CertFile     string `json:"cert_file"`
		KeyFile      string `json:"key_file"`
	} `json:"server"`
	
	Security struct {
		EnableAuth    bool     `json:"enable_auth"`
		AllowOrigins  []string `json:"allow_origins"`
		AllowMethods  []string `json:"allow_methods"`
		AllowHeaders  []string `json:"allow_headers"`
		MaxAge        int      `json:"max_age"`
		SecretKey     string   `json:"secret_key"`
		TokenExpiry   int      `json:"token_expiry"`
	} `json:"security"`
	
	Database struct {
		Type     string `json:"type"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
		SSLMode  string `json:"ssl_mode"`
	} `json:"database"`
	
	Logging struct {
		Level      string `json:"level"`
		Format     string `json:"format"`
		Output     string `json:"output"`
		MaxSize    int    `json:"max_size"`
		MaxBackups int    `json:"max_backups"`
		MaxAge     int    `json:"max_age"`
		Compress   bool   `json:"compress"`
	} `json:"logging"`
	
	Features struct {
		Terminal   bool `json:"terminal"`
		Filesystem bool `json:"filesystem"`
		System     bool `json:"system"`
		Audio      bool `json:"audio"`
		Theme      bool `json:"theme"`
		Plugins    bool `json:"plugins"`
	} `json:"features"`
	
	Plugins struct {
		Enabled  bool     `json:"enabled"`
		Paths    []string `json:"paths"`
		AutoLoad bool     `json:"auto_load"`
	} `json:"plugins"`
	
	LastModified time.Time `json:"last_modified"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	UserID      string                 `json:"user_id"`
	Theme       ThemePreferences       `json:"theme"`
	Terminal    TerminalPreferences    `json:"terminal"`
	Filesystem  FilesystemPreferences  `json:"filesystem"`
	System      SystemPreferences      `json:"system"`
	Audio       AudioPreferences       `json:"audio"`
	General     GeneralPreferences     `json:"general"`
	Keyboard    KeyboardPreferences    `json:"keyboard"`
	Notifications NotificationPreferences `json:"notifications"`
	LastModified time.Time             `json:"last_modified"`
}

// ThemePreferences represents theme-related preferences
type ThemePreferences struct {
	Name        string            `json:"name"`
	IsDark      bool              `json:"is_dark"`
	AccentColor string            `json:"accent_color"`
	FontFamily  string            `json:"font_family"`
	FontSize    int               `json:"font_size"`
	Colors      map[string]string `json:"colors"`
}

// TerminalPreferences represents terminal preferences
type TerminalPreferences struct {
	Shell           string   `json:"shell"`
	FontFamily      string   `json:"font_family"`
	FontSize        int      `json:"font_size"`
	Opacity         int      `json:"opacity"`
	Scrollback      int      `json:"scrollback"`
	BellEnabled     bool     `json:"bell_enabled"`
	CursorBlink     bool     `json:"cursor_blink"`
	Hotkeys         map[string]string `json:"hotkeys"`
	Profiles        []TerminalProfile `json:"profiles"`
	DefaultProfile  string   `json:"default_profile"`
}

// FilesystemPreferences represents filesystem preferences
type FilesystemPreferences struct {
	ShowHiddenFiles     bool     `json:"show_hidden_files"`
	ShowSystemFiles     bool     `json:"show_system_files"`
	DefaultView         string   `json:"default_view"`
	SortBy              string   `json:"sort_by"`
	SortOrder           string   `json:"sort_order"`
	Bookmarks           []string `json:"bookmarks"`
	RecentFiles         []string `json:"recent_files"`
	MaxRecentFiles      int      `json:"max_recent_files"`
	EnablePreview       bool     `json:"enable_preview"`
	MaxPreviewSize      int64    `json:"max_preview_size"`
	DoubleClickAction   string   `json:"double_click_action"`
	ConfirmDelete       bool     `json:"confirm_delete"`
	ShowFileSizeFormat  string   `json:"show_file_size_format"`
}

// SystemPreferences represents system monitoring preferences
type SystemPreferences struct {
	UpdateInterval     int    `json:"update_interval"`
	EnableMonitoring   bool   `json:"enable_monitoring"`
	ShowProcesses      bool   `json:"show_processes"`
	MaxProcesses       int    `json:"max_processes"`
	ShowServices       bool   `json:"show_services"`
	EnableAlerts       bool   `json:"enable_alerts"`
	CPUAlertThreshold  float64 `json:"cpu_alert_threshold"`
	MemoryAlertThreshold float64 `json:"memory_alert_threshold"`
	DiskAlertThreshold float64 `json:"disk_alert_threshold"`
	NotificationSound  bool   `json:"notification_sound"`
}

// AudioPreferences represents audio preferences
type AudioPreferences struct {
	EnableNotifications bool   `json:"enable_notifications"`
	DefaultVolume       int    `json:"default_volume"`
	EnableEqualizer     bool   `json:"enable_equalizer"`
	EqualizerPresets    []string `json:"equalizer_presets"`
	DefaultDevice       string `json:"default_device"`
	AudioFormat         string `json:"audio_format"`
	SampleRate          int    `json:"sample_rate"`
	Bitrate             int    `json:"bitrate"`
}

// GeneralPreferences represents general application preferences
type GeneralPreferences struct {
	Language       string `json:"language"`
	Timezone       string `json:"timezone"`
	Dateformat     string `json:"date_format"`
	TimeFormat     string `json:"time_format"`
	AutoSave       bool   `json:"auto_save"`
	AutoSaveInterval int  `json:"auto_save_interval"`
	CheckUpdates   bool   `json:"check_updates"`
	UpdateChannel  string `json:"update_channel"`
	Telemetry      bool   `json:"telemetry"`
	StartupBehavior string `json:"startup_behavior"`
}

// KeyboardPreferences represents keyboard preferences
type KeyboardPreferences struct {
	KeyBindings    map[string]string `json:"key_bindings"`
	RepeatDelay    int               `json:"repeat_delay"`
	RepeatRate     int               `json:"repeat_rate"`
	CapsLockBehavior string          `json:"caps_lock_behavior"`
}

// NotificationPreferences represents notification preferences
type NotificationPreferences struct {
	Enabled         bool     `json:"enabled"`
	ShowDesktop     bool     `json:"show_desktop"`
	PlaySound       bool     `json:"play_sound"`
	Position        string   `json:"position"`
	Duration        int      `json:"duration"`
	Types           []string `json:"types"`
	QuietHours      QuietHours `json:"quiet_hours"`
}

// QuietHours represents quiet hours configuration
type QuietHours struct {
	Enabled bool   `json:"enabled"`
	From    string `json:"from"`
	To      string `json:"to"`
}

// TerminalProfile represents a terminal profile

// ConfigChange represents a configuration change event
type ConfigChange struct {
	Key         string      `json:"key"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Type        string      `json:"type"`
	UserID      string      `json:"user_id"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ConfigValidation represents configuration validation result
type ConfigValidation struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Section string   `json:"section"`
}
