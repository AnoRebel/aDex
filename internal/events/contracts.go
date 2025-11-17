package events

import (
	"context"
	"time"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data interface{}, source string) error
	Broadcast(ctx context.Context, data interface{}, source string) error
}

// EventSubscriber defines the interface for subscribing to events
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) (*EventSubscription, error)
	SubscribeOnce(ctx context.Context, eventType string, handler EventHandler) (*EventSubscription, error)
	Unsubscribe(subscriptionID string) error
}

// IEventBus defines the complete event system interface (alias to avoid conflicts)
type IEventBus interface {
	EventPublisher
	EventSubscriber
	GetSubscribers(eventType string) int
}

// ServiceEventEmitter defines the interface for services that emit events
type ServiceEventEmitter interface {
	SetEventBus(bus IEventBus)
	GetEventBus() IEventBus
}


// EventFilter determines whether an event should be processed
type EventFilter func(event Event) bool

// EventValidator validates event data before processing
type EventValidator func(event Event) error

// EventMiddleware allows for event processing pipelines
type EventMiddleware func(next EventHandler) EventHandler

// EventContract defines the structure for typed events
type EventContract struct {
	Type        string      `json:"type"`
	Version     string      `json:"version"`
	Schema      interface{} `json:"schema"`
	Validator   EventValidator
	Middlewares []EventMiddleware
}

// EventRegistry holds all registered event contracts
type EventRegistry struct {
	contracts map[string]*EventContract
	validators map[string]EventValidator
}

// NewEventRegistry creates a new event registry
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		contracts: make(map[string]*EventContract),
		validators: make(map[string]EventValidator),
	}
}

// Register registers a new event contract
func (er *EventRegistry) Register(contract *EventContract) {
	er.contracts[contract.Type] = contract
	if contract.Validator != nil {
		er.validators[contract.Type] = contract.Validator
	}
}

// Validate validates an event against its contract
func (er *EventRegistry) Validate(event Event) error {
	if validator, exists := er.validators[event.Type]; exists {
		return validator(event)
	}
	return nil
}

// GetContract retrieves an event contract
func (er *EventRegistry) GetContract(eventType string) (*EventContract, bool) {
	contract, exists := er.contracts[eventType]
	return contract, exists
}

// Common event contracts for the application
var (
	// System Event Contracts
	SystemInfoContract = &EventContract{
		Type:    SystemInfoUpdated,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"cpu_usage":    "float64",
			"memory_usage": "float64",
			"disk_usage":   "float64",
			"network_io":   "int64",
		},
		Validator: validateSystemInfo,
	}

	SystemAlertContract = &EventContract{
		Type:    SystemAlert,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"type":       "string",
			"resource":   "string",
			"threshold":  "float64",
			"current":    "float64",
			"message":    "string",
			"severity":   "string",
		},
		Validator: validateSystemAlert,
	}

	// Terminal Event Contracts
	TerminalCreatedContract = &EventContract{
		Type:    TerminalCreated,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"session_id": "string",
			"command":    "string",
			"cwd":        "string",
		},
		Validator: validateTerminalEvent,
	}

	TerminalOutputContract = &EventContract{
		Type:    TerminalOutput,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"session_id": "string",
			"output":     "string",
			"timestamp":  "time.Time",
		},
		Validator: validateTerminalOutput,
	}

	// Filesystem Event Contracts
	FileOperationContract = &EventContract{
		Type:    FileCreated,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"path":      "string",
			"type":      "string",
			"size":      "int64",
			"operation": "string",
		},
		Validator: validateFileOperation,
	}

	// Audio Event Contracts
	AudioVolumeContract = &EventContract{
		Type:    AudioVolumeChanged,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"device_id": "string",
			"volume":    "float64",
			"muted":     "bool",
		},
		Validator: validateAudioVolume,
	}

	// Configuration Event Contracts
	ConfigChangedContract = &EventContract{
		Type:    ConfigChanged,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"key":      "string",
			"old_value": "interface{}",
			"new_value": "interface{}",
			"section":  "string",
		},
		Validator: validateConfigChange,
	}

	// Theme Event Contracts
	ThemeChangedContract = &EventContract{
		Type:    ThemeChanged,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"theme_id": "string",
			"name":     "string",
			"is_dark":  "bool",
		},
		Validator: validateThemeChange,
	}

	// Network Event Contracts
	NetworkStatusContract = &EventContract{
		Type:    NetworkConnected,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"interface": "string",
			"status":    "string",
			"ip":        "string",
		},
		Validator: validateNetworkStatus,
	}

	// Error Event Contracts
	ErrorOccurredContract = &EventContract{
		Type:    ErrorOccurred,
		Version: "1.0.0",
		Schema: map[string]interface{}{
			"error":   "string",
			"type":    "string",
			"context": "string",
			"stack":   "string",
		},
		Validator: validateErrorOccurred,
	}
)

// Event validation functions
func validateSystemInfo(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateSystemAlert(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateTerminalEvent(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateTerminalOutput(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateFileOperation(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateAudioVolume(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateConfigChange(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateThemeChange(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateNetworkStatus(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateErrorOccurred(event Event) error {
	// Implementation would validate the structure and values
	return nil
}

// Service interfaces for event-driven architecture

// SystemService defines the interface for system monitoring services
type SystemService interface {
	ServiceEventEmitter
	StartMonitoring(ctx context.Context) error
	StopMonitoring() error
	GetSystemInfo() (*SystemInfoData, error)
}

// TerminalService defines the interface for terminal services
type TerminalService interface {
	ServiceEventEmitter
	CreateSession(options *TerminalOptions) (*TerminalSession, error)
	CloseSession(sessionID string) error
	WriteToSession(sessionID string, data []byte) error
	ResizeSession(sessionID string, rows, cols int) error
}

// FileService defines the interface for file system services
type FileService interface {
	ServiceEventEmitter
	WatchPath(path string) error
	StopWatch(path string) error
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
}

// AudioService defines the interface for audio services
type AudioService interface {
	ServiceEventEmitter
	GetDevices() ([]AudioDevice, error)
	SetVolume(deviceID string, volume float64) error
	GetVolume(deviceID string) (float64, bool, error)
}

// ConfigService defines the interface for configuration services
type ConfigService interface {
	ServiceEventEmitter
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Load() error
	Save() error
}

// ThemeService defines the interface for theme services
type ThemeService interface {
	ServiceEventEmitter
	GetThemes() ([]Theme, error)
	SetTheme(themeID string) error
	GetCurrentTheme() (*Theme, error)
}

// Supporting types

type TerminalOptions struct {
	Shell   string `json:"shell"`
	CWD     string `json:"cwd"`
	Env     []string `json:"env"`
	Rows    int    `json:"rows"`
	Cols    int    `json:"cols"`
}

type TerminalSession struct {
	ID       string `json:"id"`
	PID      int    `json:"pid"`
	Shell    string `json:"shell"`
	CWD      string `json:"cwd"`
	Active   bool   `json:"active"`
	Started  time.Time `json:"started"`
}

type AudioDevice struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Volume float64 `json:"volume"`
	Muted  bool   `json:"muted"`
}

type Theme struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	IsDark  bool              `json:"is_dark"`
	Colors  map[string]string `json:"colors"`
	Author  string            `json:"author"`
	Version string            `json:"version"`
}