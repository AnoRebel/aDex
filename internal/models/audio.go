package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// AudioEventType represents the type of audio event
type AudioEventType string

const (
	AudioEventTypeSystem       AudioEventType = "system"
	AudioEventTypeInteraction AudioEventType = "interaction"
	AudioEventTypeNotification AudioEventType = "notification"
	AudioTypeError            AudioEventType = "error"
	AudioEventTypeSuccess      AudioEventType = "success"
)

// AudioEvent represents a sound effect configuration
type AudioEvent struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	FilePath    string        `json:"file_path"`
	Duration    int           `json:"duration"`    // milliseconds
	Volume      int           `json:"volume"`      // 0-100
	Category    AudioEventType `json:"category"`
	Enabled     bool          `json:"enabled"`
	CreatedAt   time.Time     `json:"created_at"`
}

// AudioSettings represents user audio preferences
type AudioSettings struct {
	Enabled         bool          `json:"enabled"`
	Volume          int           `json:"volume"`           // 0-100
	Muted           bool          `json:"muted"`
	Soundpack       string        `json:"soundpack"`        // Name of soundpack
	GlobalVolume    float64       `json:"global_volume"`    // 0.0-1.0
	EffectsVolume   float64       `json:"effects_volume"`   // 0.0-1.0
	NotificationVolume float64   `json:"notification_volume"` // 0.0-1.0
	MuteInBackground bool          `json:"mute_in_background"`
	AutoPlay        bool          `json:"auto_play"`        // Auto-play on certain events
	EnabledEvents   []string      `json:"enabled_events"`   // Specific events to enable
	Version         string        `json:"version"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// AudioPlaybackStatus represents the current state of audio playback
type AudioPlaybackStatus struct {
	IsPlaying       bool          `json:"is_playing"`
	CurrentEvent     *AudioEvent   `json:"current_event,omitempty"`
	Volume           float64       `json:"volume"`
	Muted            bool          `json:"muted"`
	LoadedEvents     []string      `json:"loaded_events"`
	Error            string        `json:"error,omitempty"`
	SupportedFormats []string      `json:"supported_formats"`
}

// AudioEventMapping maps UI events to audio events
type AudioEventMapping struct {
	UIEvent      string    `json:"ui_event"`      // e.g., "terminal:command_entered"
	AudioEventID string    `json:"audio_event_id"`
	Conditions    []string  `json:"conditions"`    // Optional conditions
	Priority      int       `json:"priority"`       // Higher priority overrides others
	CooldownMs    int       `json:"cooldown_ms"`   // Minimum time between plays
	LastPlayed    time.Time `json:"last_played"`
}

// AudioStats represents audio usage statistics
type AudioStats struct {
	TotalPlays      int                    `json:"total_plays"`
	EventsPlayed     map[string]int         `json:"events_played"`
	CategoriesPlayed map[string]int         `json:"categories_played"`
	TotalPlaytime    time.Duration          `json:"total_playtime"`
	AverageVolume    float64                `json:"average_volume"`
	LastPlayed       time.Time             `json:"last_played"`
	MostPlayedEvent  string                `json:"most_played_event"`
	Settings         map[string]interface{} `json:"settings"`
	GeneratedAt      time.Time             `json:"generated_at"`
}

// Soundpack represents a collection of audio events
type Soundpack struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Events      []AudioEvent `json:"events"`
	Enabled     bool         `json:"enabled"`
	IsBuiltin   bool         `json:"is_builtin"`
	MinVersion  string       `json:"min_version"`  // Minimum app version
	MaxVersion  string       `json:"max_version"`  // Maximum app version (optional)
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// DefaultAudioSettings returns the default audio configuration
func DefaultAudioSettings() *AudioSettings {
	return &AudioSettings{
		Enabled:            false, // Disabled by default
		Volume:             50,
		Muted:              false,
		Soundpack:          "default",
		GlobalVolume:       0.5,
		EffectsVolume:      0.5,
		NotificationVolume: 0.3,
		MuteInBackground:  false,
		AutoPlay:           true,
		EnabledEvents: []string{
			"system:startup",
			"terminal:bell",
			"terminal:command_error",
			"ui:button_click",
			"ui:notification",
			"file:operation_complete",
			"network:connected",
		},
		Version: "1.0.0",
		UpdatedAt: time.Now(),
	}
}

// IsEventEnabled checks if a specific audio event is enabled
func (as *AudioSettings) IsEventEnabled(eventID string) bool {
	if !as.Enabled || as.Muted {
		return false
	}

	for _, enabledID := range as.EnabledEvents {
		if enabledID == eventID {
			return true
		}
	}

	return false
}

// UpdateEventToggles enables or disables specific audio events
func (as *AudioSettings) UpdateEventToggles(eventIDs []string, enabled bool) {
	if enabled {
		// Add events to enabled list if not already present
		enabledSet := make(map[string]bool)
		for _, id := range as.EnabledEvents {
			enabledSet[id] = true
		}

		for _, id := range eventIDs {
			if !enabledSet[id] {
				as.EnabledEvents = append(as.EnabledEvents, id)
				enabledSet[id] = true
			}
		}
	} else {
		// Remove events from enabled list
		var updatedEvents []string
		enabledSet := make(map[string]bool)
		for _, id := range eventIDs {
			enabledSet[id] = true
		}

		for _, id := range as.EnabledEvents {
			if !enabledSet[id] {
				updatedEvents = append(updatedEvents, id)
			}
		}

		as.EnabledEvents = updatedEvents
	}

	as.UpdatedAt = time.Now()
}

// ToJSON converts AudioSettings to JSON
func (as *AudioSettings) ToJSON() ([]byte, error) {
	return json.Marshal(as)
}

// FromJSON loads AudioSettings from JSON
func (as *AudioSettings) FromJSON(data []byte) error {
	return json.Unmarshal(data, as)
}

// Validate checks if audio settings are valid
func (as *AudioSettings) Validate() error {
	if as.Volume < 0 || as.Volume > 100 {
		return &ValidationError{Field: "volume", Message: "volume must be between 0 and 100"}
	}

	if as.GlobalVolume < 0 || as.GlobalVolume > 1 {
		return &ValidationError{Field: "global_volume", Message: "global_volume must be between 0.0 and 1.0"}
	}

	if as.EffectsVolume < 0 || as.EffectsVolume > 1 {
		return &ValidationError{Field: "effects_volume", Message: "effects_volume must be between 0.0 and 1.0"}
	}

	if as.NotificationVolume < 0 || as.NotificationVolume > 1 {
		return &ValidationError{Field: "notification_volume", Message: "notification_volume must be between 0.0 and 1.0"}
	}

	if as.Soundpack == "" {
		return &ValidationError{Field: "soundpack", Message: "soundpack cannot be empty"}
	}

	return nil
}

// CalculateEffectiveVolume returns the effective volume considering global and category volumes
func (as *AudioSettings) CalculateEffectiveVolume(category AudioEventType) float64 {
	if as.Muted {
		return 0
	}

	baseVolume := float64(as.Volume) / 100.0
	globalMultiplier := as.GlobalVolume

	var categoryMultiplier float64
	switch category {
	case AudioEventTypeInteraction:
		categoryMultiplier = as.EffectsVolume
	case AudioEventTypeNotification:
		categoryMultiplier = as.NotificationVolume
	default:
		categoryMultiplier = as.EffectsVolume
	}

	return baseVolume * globalMultiplier * categoryMultiplier
}

// GetStatsSummary returns a summary of audio statistics
func (as *AudioStats) GetStatsSummary() string {
	if as.TotalPlays == 0 {
		return "No audio events played yet"
	}

	return fmt.Sprintf(
		"Total plays: %d | Most played: %s | Playtime: %s",
		as.TotalPlays,
		as.MostPlayedEvent,
		as.TotalPlaytime.Round(time.Millisecond).String(),
	)
}

// BuiltinSoundpacks returns the list of built-in soundpacks
func BuiltinSoundpacks() []Soundpack {
	return []Soundpack{
		{
			ID:          "default",
			Name:        "Default",
			Description: "Default aDex-UI sound effects",
			Version:     "1.0.0",
			Author:      "aDex-UI Team",
			IsBuiltin:   true,
			MinVersion:  "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "minimal",
			Name:        "Minimal",
			Description: "Minimal sound effects with reduced noise",
			Version:     "1.0.0",
			Author:      "aDex-UI Team",
			IsBuiltin:   true,
			MinVersion:  "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "retro",
			Name:        "Retro",
			Description: "Retro 8-bit style sound effects",
			Version:     "1.0.0",
			Author:      "aDex-UI Team",
			IsBuiltin:   true,
			MinVersion:  "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cyberpunk",
			Name:        "Cyberpunk",
			Description: "Cyberpunk-themed electronic sound effects",
			Version:     "1.0.0",
			Author:      "aDex-UI Team",
			IsBuiltin:   true,
			MinVersion:  "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

// DefaultEventMappings returns the default UI event to audio event mappings
func DefaultEventMappings() []AudioEventMapping {
	return []AudioEventMapping{
		{
			UIEvent:      "system:startup",
			AudioEventID: "startup_sound",
			Priority:     10,
			CooldownMs:   0,
		},
		{
			UIEvent:      "terminal:bell",
			AudioEventID: "terminal_bell",
			Priority:     20,
			CooldownMs:   100,
		},
		{
			UIEvent:      "terminal:command_success",
			AudioEventID: "command_success",
			Priority:     15,
			CooldownMs:   200,
		},
		{
			UIEvent:      "terminal:command_error",
			AudioEventID: "command_error",
			Priority:     25,
			CooldownMs:   500,
		},
		{
			UIEvent:      "ui:button_click",
			AudioEventID: "button_click",
			Priority:     5,
			CooldownMs:   50,
		},
		{
			UIEvent:      "ui:menu_open",
			AudioEventID: "menu_open",
			Priority:     8,
			CooldownMs:   100,
		},
		{
			UIEvent:      "ui:notification",
			AudioEventID: "notification",
			Priority:     18,
			CooldownMs:   300,
		},
		{
			UIEvent:      "file:operation_complete",
			AudioEventID: "file_complete",
			Priority:     12,
			CooldownMs:   200,
		},
		{
			UIEvent:      "file:operation_error",
			AudioEventID: "file_error",
			Priority:     22,
			CooldownMs:   500,
		},
		{
			UIEvent:      "network:connected",
			AudioEventID: "network_connected",
			Priority:     14,
			CooldownMs:   1000,
		},
		{
			UIEvent:      "network:disconnected",
			AudioEventID: "network_disconnected",
			Priority:     20,
			CooldownMs:   1000,
		},
		{
			UIEvent:      "system:alert",
			AudioEventID: "system_alert",
			Priority:     30,
			CooldownMs:   2000,
		},
	}
}

// CanPlay checks if a mapping can be played based on cooldown
func (aem *AudioEventMapping) CanPlay() bool {
	if aem.CooldownMs <= 0 {
		return true
	}

	return time.Since(aem.LastPlayed) >= time.Duration(aem.CooldownMs)*time.Millisecond
}

// UpdateLastPlayed updates the last played timestamp
func (aem *AudioEventMapping) UpdateLastPlayed() {
	aem.LastPlayed = time.Now()
}