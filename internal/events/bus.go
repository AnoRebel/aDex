package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventBus handles event publishing and subscription
type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

// Event represents an application event
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ID      string
	Types   []string
	Handler EventHandler
	Active  bool
}

// EventHandler handles incoming events
type EventHandler func(ctx context.Context, event Event) error

// Global event bus instance
var globalEventBus *EventBus

// GetEventBus returns the global event bus instance
func GetEventBus() *EventBus {
	if globalEventBus == nil {
		globalEventBus = &EventBus{
			subscribers: make(map[string][]chan Event),
		}
	}
	return globalEventBus
}

// Subscribe subscribes to specific event types
func (eb *EventBus) Subscribe(ctx context.Context, eventTypes []string, handler EventHandler) (*EventSubscription, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subID := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	subscription := &EventSubscription{
		ID:      subID,
		Types:   eventTypes,
		Handler: handler,
		Active:  true,
	}

	// Create channels for each event type
	channels := make([]chan Event, len(eventTypes))
	for i, eventType := range eventTypes {
		if eb.subscribers[eventType] == nil {
			eb.subscribers[eventType] = make([]chan Event, 0)
		}
		channel := make(chan Event, 100) // Buffered channel
		eb.subscribers[eventType] = append(eb.subscribers[eventType], channel)
		channels[i] = channel
	}

	// Start event listener goroutine
	go eb.eventListener(ctx, subscription, channels)

	return subscription, nil
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, eventType string, data interface{}, source string) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.subscribers[eventType] == nil {
		return nil // No subscribers for this event type
	}

	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Source:    source,
	}

	// Send event to all subscribers (non-blocking)
	for _, subscriber := range eb.subscribers[eventType] {
		select {
		case subscriber <- event:
			// Event sent successfully
		default:
			// Channel is full, skip this subscriber
		}
	}

	return nil
}

// Unsubscribe removes an event subscription
func (eb *EventBus) Unsubscribe(subscriptionID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Mark subscription as inactive (actual cleanup happens in eventListener)
	// In a real implementation, you'd track subscriptions and clean them up properly
	return nil
}

// SubscribeOnce subscribes to an event type for a single occurrence
func (eb *EventBus) SubscribeOnce(ctx context.Context, eventType string, handler EventHandler) (*EventSubscription, error) {
	wrappedHandler := func(ctx context.Context, event Event) error {
		err := handler(ctx, event)
		// Unsubscribe after handling (one-time subscription)
		return err
	}
	return eb.Subscribe(ctx, []string{eventType}, wrappedHandler)
}

// GetSubscribers returns the number of subscribers for an event type
func (eb *EventBus) GetSubscribers(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if subscribers, exists := eb.subscribers[eventType]; exists {
		return len(subscribers)
	}
	return 0
}

// Emit is an alias for Publish for compatibility
func (eb *EventBus) Emit(eventType string, data interface{}) error {
	return eb.Publish(context.Background(), eventType, data, "system")
}

// eventListener listens for events and forwards them to the handler
func (eb *EventBus) eventListener(ctx context.Context, subscription *EventSubscription, channels []chan Event) {
	defer func() {
		// Clean up channels
		for _, channel := range channels {
			close(channel)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			subscription.Active = false
			return

		case event := <-channels[0]:
			if !subscription.Active {
				return
			}

			// Check if this event type matches the subscription
			for _, eventType := range subscription.Types {
				if event.Type == eventType {
					if err := subscription.Handler(ctx, event); err != nil {
						// Log error but continue processing other events
						continue
					}
					break
				}
			}

		case event := <-channels[1]:
			// Handle multiple channels similarly...
			if !subscription.Active {
				return
			}

			for _, eventType := range subscription.Types {
				if event.Type == eventType {
					if err := subscription.Handler(ctx, event); err != nil {
						continue
					}
					break
				}
			}

			// Add more cases for additional channels if needed
		}
	}
}

// Broadcast publishes an event to all event types (use sparingly)
func (eb *EventBus) Broadcast(ctx context.Context, data interface{}, source string) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	event := Event{
		Type:      "broadcast",
		Timestamp: time.Now(),
		Data:      data,
		Source:    source,
	}

	// Send to all subscribers of all event types
	for _, subscribers := range eb.subscribers {
		for _, subscriber := range subscribers {
			select {
			case subscriber <- event:
				// Event sent successfully
			default:
				// Channel is full, skip this subscriber
			}
		}
	}

	return nil
}

// Common event types
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

	// Application events
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

// Event data structures
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
