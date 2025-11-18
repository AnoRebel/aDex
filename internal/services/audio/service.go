package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
	"github.com/djherbis/atime"
)

// Service manages audio effects and sound playback
type Service struct {
	// Configuration
	config ServiceConfig

	// Audio state
	settings       *models.AudioSettings
	mu            sync.RWMutex
	soundpacks    map[string]*models.Soundpack
	audioEvents   map[string]*models.AudioEvent
	eventMappings []models.AudioEventMapping
	stats         *models.AudioStats

	// Playback management
	playbackQueue chan string
	isPlaying     bool
	currentEvent  *models.AudioEvent
	audioPlayers  map[string]Player

	// Event system
	eventBus *events.EventBus

	// Background processing
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// ServiceConfig holds configuration for the audio service
type ServiceConfig struct {
	AudioDirectory      string        `json:"audio_directory"`
	SoundpacksDirectory string        `json:"soundpacks_directory"`
	SupportedFormats    []string      `json:"supported_formats"`
	MaxConcurrentSounds int           `json:"max_concurrent_sounds"`
	DefaultVolume       float64       `json:"default_volume"`
	CacheSize           int           `json:"cache_size"`
	MaxEventQueue       int           `json:"max_event_queue"`
	EnableAnalytics     bool          `json:"enable_analytics"`
	AnalyticsRetention  time.Duration `json:"analytics_retention"`
}

// Player interface for audio playback
type Player interface {
	Play(event *models.AudioEvent, volume float64) error
	Stop() error
	IsPlaying() bool
	GetDuration() time.Duration
	SetVolume(volume float64) error
	Cleanup() error
}

// DefaultServiceConfig returns the default configuration for the audio service
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		AudioDirectory:      "audio",
		SoundpacksDirectory: "soundpacks",
		SupportedFormats:    []string{"mp3", "wav", "ogg", "flac"},
		MaxConcurrentSounds: 3,
		DefaultVolume:       0.5,
		CacheSize:           100,
		MaxEventQueue:       100,
		EnableAnalytics:     true,
		AnalyticsRetention:  30 * 24 * time.Hour, // 30 days
	}
}

// NewService creates a new audio service with the given configuration
func NewService(config ServiceConfig, eventBus *events.EventBus) *Service {
	service := &Service{
		config:       config,
		settings:     models.DefaultAudioSettings(),
		soundpacks:   make(map[string]*models.Soundpack),
		audioEvents:  make(map[string]*models.AudioEvent),
		eventMappings: models.DefaultEventMappings(),
		stats:        &models.AudioStats{
			EventsPlayed:      make(map[string]int),
			CategoriesPlayed: make(map[string]int),
			Settings:          make(map[string]interface{}),
			GeneratedAt:       time.Now(),
		},
		playbackQueue: make(chan string, config.MaxEventQueue),
		audioPlayers:  make(map[string]Player),
		eventBus:      eventBus,
		stopChan:      make(chan struct{}),
	}

	return service
}

// Initialize initializes the audio service
func (s *Service) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create audio directories if they don't exist
	if err := os.MkdirAll(s.config.AudioDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %w", err)
	}

	if err := os.MkdirAll(s.config.SoundpacksDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create soundpacks directory: %w", err)
	}

	// Load built-in soundpacks
	if err := s.loadBuiltinSoundpacks(); err != nil {
		return fmt.Errorf("failed to load built-in soundpacks: %w", err)
	}

	// Load custom soundpacks from directory
	if err := s.loadSoundpacksFromDirectory(); err != nil {
		return fmt.Errorf("failed to load soundpacks from directory: %w", err)
	}

	// Load settings from file
	if err := s.loadSettings(); err != nil {
		// If settings don't exist, create default ones
		if os.IsNotExist(err) {
			if err := s.saveSettings(); err != nil {
				return fmt.Errorf("failed to save default settings: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load settings: %w", err)
		}
	}

	// Load analytics data if analytics is enabled
	if s.config.EnableAnalytics {
		s.loadAnalytics()
	}

	// Start background audio processor
	s.wg.Add(1)
	go s.audioProcessor()

	// Start analytics processor if enabled
	if s.config.EnableAnalytics {
		s.wg.Add(1)
		go s.analyticsProcessor()
	}

	return nil
}

// loadBuiltinSoundpacks loads built-in soundpacks
func (s *Service) loadBuiltinSoundpacks() error {
	builtinSoundpacks := models.BuiltinSoundpacks()

	for _, soundpack := range builtinSoundpacks {
		// Add built-in audio events
		soundpack.Events = s.createBuiltinAudioEvents(soundpack.ID)
		s.soundpacks[soundpack.ID] = &soundpack

		// Add audio events to the main map
		for _, event := range soundpack.Events {
			s.audioEvents[event.ID] = &event
		}
	}

	return nil
}

// createBuiltinAudioEvents creates built-in audio events for a soundpack
func (s *Service) createBuiltinAudioEvents(soundpackID string) []models.AudioEvent {
	now := time.Now()
	events := []models.AudioEvent{}

	// Common audio events across all soundpacks
	baseEvents := []struct {
		id          string
		name        string
		description string
		category    models.AudioEventType
		filePath    string
		duration    int
	}{
		{"startup_sound", "Startup Sound", "Application startup sound", models.AudioEventTypeSystem, "startup.mp3", 500},
		{"shutdown_sound", "Shutdown Sound", "Application shutdown sound", models.AudioEventTypeSystem, "shutdown.mp3", 300},
		{"terminal_bell", "Terminal Bell", "Terminal bell notification", models.AudioEventTypeNotification, "bell.mp3", 200},
		{"command_success", "Command Success", "Command executed successfully", models.AudioEventTypeSuccess, "success.mp3", 150},
		{"command_error", "Command Error", "Command execution failed", models.AudioTypeError, "error.mp3", 400},
		{"button_click", "Button Click", "Button click sound", models.AudioEventTypeInteraction, "click.mp3", 50},
		{"menu_open", "Menu Open", "Menu opening sound", models.AudioEventTypeInteraction, "menu_open.mp3", 100},
		{"notification", "Notification", "System notification sound", models.AudioEventTypeNotification, "notification.mp3", 300},
		{"file_complete", "File Operation Complete", "File operation completed successfully", models.AudioEventTypeSuccess, "file_complete.mp3", 200},
		{"file_error", "File Operation Error", "File operation failed", models.AudioTypeError, "file_error.mp3", 350},
		{"network_connected", "Network Connected", "Network connection established", models.AudioEventTypeNotification, "network_connected.mp3", 250},
		{"network_disconnected", "Network Disconnected", "Network connection lost", models.AudioEventTypeNotification, "network_disconnected.mp3", 300},
		{"system_alert", "System Alert", "Important system alert", models.AudioEventTypeSystem, "alert.mp3", 600},
	}

	for _, event := range baseEvents {
		audioEvent := models.AudioEvent{
			ID:          event.id,
			Name:        event.name,
			Description: event.description,
			FilePath:    filepath.Join(s.config.AudioDirectory, soundpackID, event.filePath),
			Duration:    event.duration,
			Volume:      75,
			Category:    event.category,
			Enabled:     true,
			CreatedAt:   now,
		}
		events = append(events, audioEvent)
	}

	// Add soundpack-specific events
	switch soundpackID {
	case "minimal":
		for i := range events {
			events[i].Volume = 50 // Lower volume for minimal pack
		}
	case "retro":
		// Retro-specific events could be added here
		events = append(events, models.AudioEvent{
			ID:          "retro_powerup",
			Name:        "Retro Power Up",
			Description: "8-bit power up sound",
			FilePath:    filepath.Join(s.config.AudioDirectory, soundpackID, "powerup.mp3"),
			Duration:    400,
			Volume:      80,
			Category:    models.AudioEventTypeSystem,
			Enabled:     true,
			CreatedAt:   now,
		})
	case "cyberpunk":
		// Cyberpunk-specific events could be added here
		events = append(events, models.AudioEvent{
			ID:          "cyber_startup",
			Name:        "Cyber Startup",
			Description: "Cyberpunk startup sequence",
			FilePath:    filepath.Join(s.config.AudioDirectory, soundpackID, "cyber_startup.mp3"),
			Duration:    1200,
			Volume:      90,
			Category:    models.AudioEventTypeSystem,
			Enabled:     true,
			CreatedAt:   now,
		})
	}

	return events
}

// loadSoundpacksFromDirectory loads custom soundpacks from the directory
func (s *Service) loadSoundpacksFromDirectory() error {
	return filepath.WalkDir(s.config.SoundpacksDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !d.Name() {
			return nil
		}

		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		// Load soundpack from file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read soundpack file '%s': %w", path, err)
		}

		var soundpack models.Soundpack
		if err := json.Unmarshal(data, &soundpack); err != nil {
			return fmt.Errorf("failed to unmarshal soundpack from '%s': %w", path, err)
		}

		// Add soundpack
		s.soundpacks[soundpack.ID] = &soundpack

		// Add audio events to the main map
		for _, event := range soundpack.Events {
			s.audioEvents[event.ID] = &event
		}

		return nil
	})
}

// PlaySound plays a sound effect by ID
func (s *Service) PlaySound(soundID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if audio is enabled
	if !s.settings.Enabled || s.settings.Muted {
		return nil // Silently fail if audio is disabled
	}

	// Check if the sound event exists
	event, exists := s.audioEvents[soundID]
	if !exists {
		return fmt.Errorf("audio event '%s' not found", soundID)
	}

	// Check if the event is enabled
	if !event.Enabled {
		return nil
	}

	// Add to playback queue (non-blocking)
	select {
	case s.playbackQueue <- soundID:
		return nil
	default:
		return fmt.Errorf("audio queue is full")
	}
}

// PlayUIEvent plays a sound for a UI event
func (s *Service) PlayUIEvent(uiEvent string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find the audio event mapping for this UI event
	var highestPriority int = -1
	var selectedMapping *models.AudioEventMapping

	for i := range s.eventMappings {
		mapping := &s.eventMappings[i]
		if mapping.UIEvent == uiEvent && s.settings.IsEventEnabled(mapping.AudioEventID) {
			if mapping.Priority > highestPriority && mapping.CanPlay() {
				highestPriority = mapping.Priority
				selectedMapping = mapping
			}
		}
	}

	if selectedMapping == nil {
		return nil // No mapping found or none can play
	}

	// Update last played time
	selectedMapping.UpdateLastPlayed()

	// Play the sound
	return s.PlaySound(selectedMapping.AudioEventID)
}

// GetSettings returns the current audio settings
func (s *Service) GetSettings() *models.AudioSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	settingsCopy := *s.settings
	return &settingsCopy
}

// UpdateSettings updates the audio settings
func (s *Service) UpdateSettings(settings *models.AudioSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Validate settings
	if err := settings.Validate(); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings = settings
	s.settings.UpdatedAt = time.Now()

	// Save settings to file
	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Emit settings changed event
	if s.eventBus != nil {
		s.eventBus.Emit("audio:settings_changed", map[string]interface{}{
			"settings": settings,
			"timestamp": time.Now(),
		})
	}

	return nil
}

// SetVolume sets the master volume
func (s *Service) SetVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.Volume = volume
	s.settings.UpdatedAt = time.Now()

	// Save settings
	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Update all audio players if needed
	s.updatePlayersVolume()

	return nil
}

// SetMuted sets the muted state
func (s *Service) SetMuted(muted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.Muted = muted
	s.settings.UpdatedAt = time.Now()

	// Save settings
	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Update all audio players
	s.updatePlayersVolume()

	return nil
}

// ToggleMute toggles the muted state
func (s *Service) ToggleMute() error {
	s.mu.RLock()
	currentMuted := s.settings.Muted
	s.mu.RUnlock()

	return s.SetMuted(!currentMuted)
}

// GetSoundpacks returns all available soundpacks
func (s *Service) GetSoundpacks() []*models.Soundpack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	soundpacks := make([]*models.Soundpack, 0, len(s.soundpacks))
	for _, soundpack := range s.soundpacks {
		soundpacks = append(soundpacks, soundpack)
	}

	return soundpacks
}

// GetActiveSoundpack returns the currently active soundpack
func (s *Service) GetActiveSoundpack() *models.Soundpack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.soundpacks[s.settings.Soundpack]
}

// SetSoundpack sets the active soundpack
func (s *Service) SetSoundpack(soundpackID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.soundpacks[soundpackID]; !exists {
		return fmt.Errorf("soundpack '%s' not found", soundpackID)
	}

	s.settings.Soundpack = soundpackID
	s.settings.UpdatedAt = time.Now()

	// Save settings
	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// GetAudioEvents returns all audio events from the active soundpack
func (s *Service) GetAudioEvents() []*models.AudioEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	soundpack := s.soundpacks[s.settings.Soundpack]
	if soundpack == nil {
		return nil
	}

	// Return a copy of the events
	events := make([]*models.AudioEvent, len(soundpack.Events))
	for i, event := range soundpack.Events {
		eventCopy := event
		events[i] = &eventCopy
	}

	return events
}

// GetPlaybackStatus returns the current playback status
func (s *Service) GetPlaybackStatus() *models.AudioPlaybackStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loadedEvents := make([]string, 0, len(s.audioEvents))
	for id := range s.audioEvents {
		loadedEvents = append(loadedEvents, id)
	}

	return &models.AudioPlaybackStatus{
		IsPlaying:       s.isPlaying,
		CurrentEvent:     s.currentEvent,
		Volume:           float64(s.settings.Volume) / 100.0,
		Muted:            s.settings.Muted,
		LoadedEvents:     loadedEvents,
		SupportedFormats: s.config.SupportedFormats,
	}
}

// GetStats returns audio usage statistics
func (s *Service) GetStats() *models.AudioStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy of the stats
	statsCopy := *s.stats
	stats := &statsCopy

	// Add current settings
	stats.Settings = map[string]interface{}{
		"enabled":            s.settings.Enabled,
		"volume":             s.settings.Volume,
		"muted":              s.settings.Muted,
		"soundpack":          s.settings.Soundpack,
		"global_volume":      s.settings.GlobalVolume,
		"effects_volume":     s.settings.EffectsVolume,
		"notification_volume": s.settings.NotificationVolume,
		"auto_play":          s.settings.AutoPlay,
		"enabled_events_count": len(s.settings.EnabledEvents),
	}

	return stats
}

// ResetStats resets the audio statistics
func (s *Service) ResetStats() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats = &models.AudioStats{
		EventsPlayed:      make(map[string]int),
		CategoriesPlayed: make(map[string]int),
		Settings:          make(map[string]interface{}),
		GeneratedAt:       time.Now(),
	}

	// Save analytics
	if s.config.EnableAnalytics {
		return s.saveAnalytics()
	}

	return nil
}

// EnableAudio enables or disables audio
func (s *Service) EnableAudio(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.Enabled = enabled
	s.settings.UpdatedAt = time.Now()

	// Save settings
	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// audioProcessor processes the audio playback queue
func (s *Service) audioProcessor() {
	defer s.wg.Done()

	for {
		select {
		case soundID := <-s.playbackQueue:
			s.processAudioEvent(soundID)
		case <-s.stopChan:
			return
		}
	}
}

// processAudioEvent processes a single audio event
func (s *Service) processAudioEvent(soundID string) {
	s.mu.Lock()
	event, exists := s.audioEvents[soundID]
	if !exists {
		s.mu.Unlock()
		return
	}

	// Check if audio is enabled
	if !s.settings.Enabled || s.settings.Muted {
		s.mu.Unlock()
		return
	}

	s.isPlaying = true
	s.currentEvent = event
	s.mu.Unlock()

	// Update statistics
	if s.config.EnableAnalytics {
		s.updateStats(event)
	}

	// Create player if it doesn't exist
	player, exists := s.audioPlayers[soundID]
	if !exists {
		player = NewPlayer(event) // Implementation needed
		s.audioPlayers[soundID] = player
	}

	// Calculate effective volume
	effectiveVolume := s.settings.CalculateEffectiveVolume(event.Category)

	// Play the audio
	if err := player.Play(event, effectiveVolume); err != nil {
		// Log error but don't crash
		fmt.Printf("Failed to play audio '%s': %v\n", soundID, err)
	}

	s.mu.Lock()
	s.isPlaying = false
	s.currentEvent = nil
	s.mu.Unlock()
}

// updateStats updates audio statistics
func (s *Service) updateStats(event *models.AudioEvent) {
	s.stats.TotalPlays++
	s.stats.EventsPlayed[event.ID]++
	s.stats.CategoriesPlayed[string(event.Category)]
	s.stats.LastPlayed = time.Now()

	// Update most played event
	if s.stats.EventsPlayed[event.ID] > s.stats.EventsPlayed[s.stats.MostPlayedEvent] {
		s.stats.MostPlayedEvent = event.ID
	}

	// Update average volume (simplified calculation)
	totalEvents := len(s.stats.EventsPlayed)
	if totalEvents > 0 {
		s.stats.AverageVolume = (s.stats.AverageVolume*float64(totalEvents-1) + float64(event.Volume)/100.0) / float64(totalEvents)
	}
}

// updatePlayersVolume updates volume for all active audio players
func (s *Service) updatePlayersVolume() {
	// Implementation would update volume for all players
	// This is a simplified version - in practice you'd iterate through players
}

// saveSettings saves audio settings to file
func (s *Service) saveSettings() error {
	settingsPath := filepath.Join(s.config.AudioDirectory, "settings.json")
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// loadSettings loads audio settings from file
func (s *Service) loadSettings() error {
	settingsPath := filepath.Join(s.config.AudioDirectory, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var settings models.AudioSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	s.settings = &settings
	return nil
}

// analyticsProcessor processes analytics data periodically
func (s *Service) analyticsProcessor() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupAnalytics()
		case <-s.stopChan:
			return
		}
	}
}

// cleanupAnalytics removes old analytics data
func (s *Service) cleanupAnalytics() {
	// Remove analytics data older than retention period
	// Implementation would clean up old analytics data
}

// loadAnalytics loads analytics data from file
func (s *Service) loadAnalytics() {
	analyticsPath := filepath.Join(s.config.AudioDirectory, "analytics.json")
	data, err := os.ReadFile(analyticsPath)
	if err != nil {
		// Analytics file doesn't exist yet, that's OK
		return
	}

	var stats models.AudioStats
	if err := json.Unmarshal(data, &stats); err != nil {
		// Invalid analytics file, ignore it
		return
	}

	s.stats = &stats
}

// saveAnalytics saves analytics data to file
func (s *Service) saveAnalytics() error {
	analyticsPath := filepath.Join(s.config.AudioDirectory, "analytics.json")
	data, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	if err := os.WriteFile(analyticsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write analytics file: %w", err)
	}

	return nil
}

// Cleanup stops the audio service and cleans up resources
func (s *Service) Cleanup() error {
	// Stop background processors
	close(s.stopChan)
	s.wg.Wait()

	// Stop all audio players
	s.mu.Lock()
	for _, player := range s.audioPlayers {
		player.Stop()
		player.Cleanup()
	}
	s.audioPlayers = make(map[string]Player)
	s.mu.Unlock()

	// Save final analytics and settings
	if s.config.EnableAnalytics {
		if err := s.saveAnalytics(); err != nil {
			fmt.Printf("Warning: failed to save analytics: %v\n", err)
		}
	}

	if err := s.saveSettings(); err != nil {
		fmt.Printf("Warning: failed to save settings: %v\n", err)
	}

	return nil
}

// NewPlayer creates a new audio player (placeholder implementation)
func NewPlayer(event *models.AudioEvent) Player {
	// This would be implemented with actual audio playback logic
	// For now, return a no-op player
	return &NoOpPlayer{}
}

// NoOpPlayer is a placeholder player that does nothing
type NoOpPlayer struct{}

func (p *NoOpPlayer) Play(event *models.AudioEvent, volume float64) error {
	return nil
}

func (p *NoOpPlayer) Stop() error {
	return nil
}

func (p *NoOpPlayer) IsPlaying() bool {
	return false
}

func (p *NoOpPlayer) GetDuration() time.Duration {
	return 0
}

func (p *NoOpPlayer) SetVolume(volume float64) error {
	return nil
}

func (p *NoOpPlayer) Cleanup() error {
	return nil
}

// AudioSessionDetector handles detection of audio sessions and background state
type AudioSessionDetector struct {
	isAppInBackground    bool
	otherAppsPlaying     bool
	lastCheckTime        time.Time
	checkInterval        time.Duration
	mu                   sync.RWMutex
}

// NewAudioSessionDetector creates a new audio session detector
func NewAudioSessionDetector() *AudioSessionDetector {
	return &AudioSessionDetector{
		checkInterval: 1 * time.Second,
	}
}

// IsAppInBackground checks if the application is in the background
func (d *AudioSessionDetector) IsAppInBackground() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.isAppInBackground
}

// SetAppInBackground sets the background state of the application
func (d *AudioSessionDetector) SetAppInBackground(inBackground bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.isAppInBackground = inBackground
}

// IsOtherAppsPlaying checks if other applications are playing audio
func (d *AudioSessionDetector) IsOtherAppsPlaying() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.otherAppsPlaying
}

// CheckAudioSessions checks if other applications are playing audio
func (d *AudioSessionDetector) CheckAudioSessions(ctx context.Context) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if we need to update
	if time.Since(d.lastCheckTime) < d.checkInterval {
		return d.otherAppsPlaying, nil
	}

	d.lastCheckTime = time.Now()

	// Platform-specific audio session detection
	var playing bool
	var err error

	switch runtime.GOOS {
	case "linux":
		playing, err = d.checkLinuxAudioSessions(ctx)
	case "darwin":
		playing, err = d.checkMacOSAudioSessions(ctx)
	case "windows":
		playing, err = d.checkWindowsAudioSessions(ctx)
	default:
		return false, fmt.Errorf("audio session detection not supported on %s", runtime.GOOS)
	}

	if err == nil {
		d.otherAppsPlaying = playing
	}

	return playing, err
}

// checkLinuxAudioSessions checks for active audio sessions on Linux
func (d *AudioSessionDetector) checkLinuxAudioSessions(ctx context.Context) (bool, error) {
	// Try PulseAudio first
	cmd := exec.CommandContext(ctx, "pactl", "list", "sink-inputs", "short")
	output, err := cmd.Output()
	if err == nil {
		// If there are any sink inputs, audio is playing
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			return true, nil
		}
		return false, nil
	}

	// Try PipeWire
	cmd = exec.CommandContext(ctx, "pw-cli", "list-objects")
	output, err = cmd.Output()
	if err == nil {
		// Check for active streams
		if strings.Contains(string(output), "type: PipeWire:Interface:Node") {
			// Simplified check - in production would parse more carefully
			return true, nil
		}
		return false, nil
	}

	// Neither PulseAudio nor PipeWire available
	return false, fmt.Errorf("no audio system detected (tried PulseAudio and PipeWire)")
}

// checkMacOSAudioSessions checks for active audio sessions on macOS
func (d *AudioSessionDetector) checkMacOSAudioSessions(ctx context.Context) (bool, error) {
	// Use coreaudiod status or check for active audio streams
	// This is a simplified implementation - full implementation would use CoreAudio API

	// Check if any application is outputting audio using pmset
	cmd := exec.CommandContext(ctx, "pmset", "-g", "assertions")
	output, err := cmd.Output()
	if err == nil {
		// Check for audio-related assertions
		if strings.Contains(string(output), "PreventUserIdleSystemSleep") ||
		   strings.Contains(string(output), "Audio") {
			return true, nil
		}
	}

	return false, nil
}

// checkWindowsAudioSessions checks for active audio sessions on Windows
func (d *AudioSessionDetector) checkWindowsAudioSessions(ctx context.Context) (bool, error) {
	// On Windows, we would use WASAPI to check for active audio sessions
	// This requires CGO bindings to Windows APIs
	// For now, return an error indicating full implementation is needed

	// PowerShell command to check for active audio sessions
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		`Get-AudioDevice -PlaybackMute`)
	_, err := cmd.Output()
	if err != nil {
		// AudioDevice cmdlet not available, use alternative method
		// Check if any process has audio enabled (simplified)
		return false, fmt.Errorf("Windows audio session detection requires AudioDevice module")
	}

	return false, nil
}

// GetAudioSessionInfo returns information about current audio sessions
func (d *AudioSessionDetector) GetAudioSessionInfo() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"app_in_background":   d.isAppInBackground,
		"other_apps_playing":  d.otherAppsPlaying,
		"last_check_time":     d.lastCheckTime,
		"check_interval_ms":   d.checkInterval.Milliseconds(),
		"platform":            runtime.GOOS,
	}
}

// HandleBackgroundStateChange handles application background state changes
func (s *Service) HandleBackgroundStateChange(inBackground bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we should mute in background
	if s.settings.MuteInBackground {
		if inBackground && !s.settings.Muted {
			// Store current mute state and mute
			s.settings.Muted = true

			// Emit event
			if s.eventBus != nil {
				s.eventBus.Emit("audio:muted_background", map[string]interface{}{
					"muted":    true,
					"reason":   "app_in_background",
					"timestamp": time.Now(),
				})
			}
		} else if !inBackground && s.settings.Muted {
			// Restore audio when coming to foreground
			s.settings.Muted = false

			// Emit event
			if s.eventBus != nil {
				s.eventBus.Emit("audio:unmuted_foreground", map[string]interface{}{
					"muted":    false,
					"reason":   "app_in_foreground",
					"timestamp": time.Now(),
				})
			}
		}
	}

	return nil
}

// StartBackgroundMonitor starts monitoring for background state and audio sessions
func (s *Service) StartBackgroundMonitor(detector *AudioSessionDetector, checkInterval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check audio sessions
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = detector.CheckAudioSessions(ctx)
				cancel()

				// Handle background state
				if detector.IsAppInBackground() && s.settings.MuteInBackground {
					s.HandleBackgroundStateChange(true)
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}

// GetActiveAudioSessions returns a list of active audio session names (Linux only for now)
func GetActiveAudioSessions(ctx context.Context) ([]string, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("active session listing only supported on Linux")
	}

	cmd := exec.CommandContext(ctx, "pactl", "list", "sink-inputs")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list audio sessions: %w", err)
	}

	var sessions []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "application.name") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				sessions = append(sessions, name)
			}
		}
	}

	return sessions, nil
}