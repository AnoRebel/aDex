package events

// Common event types ------------------------------------------------------
const (
	// System events
	SystemInfoUpdated    = "system.info.updated"
	SystemAlert          = "system.alert"
	SystemProcessStarted = "system.process.started"
	SystemProcessEnded   = "system.process.ended"

	// Terminal events
	TerminalCreated = "terminal.created"
	TerminalClosed  = "terminal.closed"
	TerminalResized = "terminal.resized"
	TerminalOutput  = "terminal.output"
	TerminalInput   = "terminal.input"
	TerminalCommand = "terminal.command"

	// Filesystem events
	FileCreated       = "file.created"
	FileModified      = "file.modified"
	FileDeleted       = "file.deleted"
	FileMoved         = "file.moved"
	DirectoryCreated  = "directory.created"
	DirectoryDeleted  = "directory.deleted"
	DirectoryModified = "directory.modified"

	// Audio events
	AudioDeviceChanged  = "audio.device.changed"
	AudioVolumeChanged  = "audio.volume.changed"
	AudioSessionStarted = "audio.session.started"
	AudioSessionEnded   = "audio.session.ended"

	// Config events
	ConfigChanged = "config.changed"
	ConfigSaved   = "config.saved"
	ConfigLoaded  = "config.loaded"
	ConfigReset   = "config.reset"

	// Theme events
	ThemeChanged = "theme.changed"
	ThemeLoaded  = "theme.loaded"
	ThemeCreated = "theme.created"
	ThemeDeleted = "theme.deleted"

	// Application lifecycle
	AppStarted  = "app.started"
	AppStopped  = "app.stopped"
	AppShutdown = "app.shutdown"
	AppRestart  = "app.restart"

	// User events
	UserLogin              = "user.login"
	UserLogout             = "user.logout"
	UserPreferencesChanged = "user.preferences.changed"

	// Network events
	NetworkConnected    = "network.connected"
	NetworkDisconnected = "network.disconnected"
	NetworkError        = "network.error"
	NetworkUpdated      = "network.updated"

	// Error events
	ErrorOccurred = "error.occurred"
	PanicOccurred = "panic.occurred"
	WarningIssued = "warning.issued"
)

// Event data structs -----------------------------------------------------

type SystemInfoData struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   int64   `json:"network_io"`
}

type SystemAlertData struct {
	Type      string  `json:"type"`
	Resource  string  `json:"resource"`
	Threshold float64 `json:"threshold"`
	Current   float64 `json:"current"`
	Message   string  `json:"message"`
	Severity  string  `json:"severity"`
}

type TerminalData struct {
	SessionID string `json:"session_id"`
	Command   string `json:"command"`
	Output    string `json:"output"`
	Error     string `json:"error"`
	ExitCode  int    `json:"exit_code"`
}

type FileData struct {
	Path      string `json:"path"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
	Operation string `json:"operation"`
	Error     string `json:"error,omitempty"`
}

type AudioData struct {
	DeviceID string  `json:"device_id"`
	Volume   float64 `json:"volume"`
	Muted    bool    `json:"muted"`
	Type     string  `json:"type"`
}

type ConfigData struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Section  string      `json:"section"`
}

type ThemeData struct {
	ThemeID string `json:"theme_id"`
	Name    string `json:"name"`
	IsDark  bool   `json:"is_dark"`
}

type ErrorData struct {
	Error   string `json:"error"`
	Type    string `json:"type"`
	Context string `json:"context"`
	Stack   string `json:"stack"`
}

type NetworkData struct {
	Interface string `json:"interface"`
	Status    string `json:"status"`
	IP        string `json:"ip"`
	Error     string `json:"error,omitempty"`
}
