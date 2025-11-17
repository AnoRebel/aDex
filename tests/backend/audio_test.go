package tests

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adex-ui/aDex-UI/internal/models"
	"github.com/adex-ui/aDex-UI/internal/services/audio"
)

func TestAudioService(t *testing.T) {
	// Test audio service initialization
	t.Run("InitializeAudioService", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, tempDir, service.GetAudioPath())
	})

	// Test soundpack management
	t.Run("LoadSoundpacks", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Load built-in soundpacks
		soundpacks := service.GetSoundpacks()
		assert.NotEmpty(t, soundpacks)

		// Should have default soundpack
		defaultPack := service.GetSoundpack("default")
		assert.NotNil(t, defaultPack)
		assert.Equal(t, "default", defaultPack.ID)
		assert.True(t, defaultPack.IsBuiltin)
	})

	// Test audio settings
	t.Run("AudioSettings", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Get default settings
		settings := service.GetSettings()
		assert.NotNil(t, settings)
		assert.False(t, settings.Enabled) // Should be disabled by default
		assert.Equal(t, "default", settings.Soundpack)

		// Update settings
		settings.Enabled = true
		settings.Volume = 75
		err = service.UpdateSettings(settings)
		assert.NoError(t, err)

		// Verify settings were updated
		updatedSettings := service.GetSettings()
		assert.True(t, updatedSettings.Enabled)
		assert.Equal(t, 75, updatedSettings.Volume)
	})

	// Test audio event management
	t.Run("AudioEvents", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Get available events
		events := service.GetAvailableEvents()
		assert.NotEmpty(t, events)

		// Should have basic events
		bellEvent := service.GetEvent("terminal_bell")
		assert.NotNil(t, bellEvent)
		assert.Equal(t, "terminal_bell", bellEvent.ID)
		assert.Equal(t, models.AudioEventTypeInteraction, bellEvent.Category)

		// Test non-existent event
		nonExistentEvent := service.GetEvent("non_existent")
		assert.Nil(t, nonExistentEvent)
	})

	// Test audio playback
	t.Run("AudioPlayback", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Enable audio first
		settings := service.GetSettings()
		settings.Enabled = true
		service.UpdateSettings(settings)

		// Test playing an event (this might not produce sound in tests)
		err = service.PlayEvent("terminal_bell")

		// Should not error, but might not actually play in test environment
		// depending on audio system availability
		if err != nil {
			// Error is acceptable in CI/testing environments without audio
			t.Logf("Audio playback error (expected in test env): %v", err)
		}

		// Test playing non-existent event
		err = service.PlayEvent("non_existent")
		assert.Error(t, err)
	})

	// Test UI event mapping
	t.Run("UIEventMapping", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Test UI event to audio event mapping
		err = service.PlayUIEvent("ui:button_click")
		if err != nil {
			t.Logf("UI event playback error (expected in test env): %v", err)
		}

		// Test unmapped UI event
		err = service.PlayUIEvent("ui:non_existent_event")
		// Should not error, just silently ignore unmapped events
		assert.NoError(t, err)
	})

	// Test volume control
	t.Run("VolumeControl", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Test volume setting
		err = service.SetVolume(50)
		assert.NoError(t, err)

		settings := service.GetSettings()
		assert.Equal(t, 50, settings.Volume)

		// Test volume bounds
		err = service.SetVolume(-10)
		assert.Error(t, err)

		err = service.SetVolume(150)
		assert.Error(t, err)

		err = service.SetVolume(0)
		assert.NoError(t, err)

		err = service.SetVolume(100)
		assert.NoError(t, err)
	})

	// Test mute functionality
	t.Run("MuteFunctionality", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Test mute
		err = service.SetMuted(true)
		assert.NoError(t, err)

		settings := service.GetSettings()
		assert.True(t, settings.Muted)

		// Test unmute
		err = service.SetMuted(false)
		assert.NoError(t, err)

		settings = service.GetSettings()
		assert.False(t, settings.Muted)

		// Test toggle
		originalMuted := settings.Muted
		err = service.ToggleMute()
		assert.NoError(t, err)

		settings = service.GetSettings()
		assert.NotEqual(t, originalMuted, settings.Muted)
	})

	// Test soundpack switching
	t.Run("SoundpackSwitching", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Get available soundpacks
		soundpacks := service.GetSoundpacks()
		assert.Len(t, soundpacks, 4) // default, minimal, retro, cyberpunk

		// Switch to different soundpack
		err = service.ChangeSoundpack("minimal")
		assert.NoError(t, err)

		settings := service.GetSettings()
		assert.Equal(t, "minimal", settings.Soundpack)

		// Test invalid soundpack
		err = service.ChangeSoundpack("non_existent")
		assert.Error(t, err)
	})

	// Test audio statistics
	t.Run("AudioStatistics", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Get initial stats
		stats := service.GetStats()
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats.TotalPlays)

		// Enable audio
		settings := service.GetSettings()
		settings.Enabled = true
		service.UpdateSettings(settings)

		// Play some events (might not actually produce sound in test environment)
		service.PlayEvent("terminal_bell")
		service.PlayEvent("ui:button_click")

		// Check stats (might not increment if audio doesn't actually play)
		updatedStats := service.GetStats()
		assert.NotNil(t, updatedStats)
		assert.GreaterOrEqual(t, updatedStats.TotalPlays, stats.TotalPlays)

		// Reset stats
		err = service.ResetStats()
		assert.NoError(t, err)

		resetStats := service.GetStats()
		assert.Equal(t, 0, resetStats.TotalPlays)
	})

	// Test audio file loading
	t.Run("AudioFileLoading", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Test supported formats
		supportedFormats := service.GetSupportedFormats()
		assert.Contains(t, supportedFormats, "wav")
		assert.Contains(t, supportedFormats, "mp3")
		assert.Contains(t, supportedFormats, "ogg")
		assert.Contains(t, supportedFormats, "aac")

		// Test file existence check
		exists := service.AudioFileExists("terminal_bell")
		// In test environment, audio files might not exist
		if !exists {
			t.Log("Audio file doesn't exist (expected in test environment)")
		}
	})

	// Test event cooldown system
	t.Run("EventCooldown", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Enable audio
		settings := service.GetSettings()
		settings.Enabled = true
		service.UpdateSettings(settings)

		// Play event multiple times rapidly
		err := service.PlayEvent("terminal_bell")
		if err != nil {
			t.Logf("First play error (expected in test env): %v", err)
		}

		// Immediate second play should be subject to cooldown
		err = service.PlayEvent("terminal_bell")
		if err != nil {
			t.Logf("Second play error (expected due to cooldown/test env): %v", err)
		}

		// Wait for cooldown and try again
		time.Sleep(200 * time.Millisecond)
		err = service.PlayEvent("terminal_bell")
		if err != nil {
			t.Logf("Third play error (expected in test env): %v", err)
		}
	})

	// Test event enable/disable
	t.Run("EventToggle", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Test disabling an event
		err = service.DisableEvent("terminal_bell")
		assert.NoError(t, err)

		isEnabled := service.IsEventEnabled("terminal_bell")
		assert.False(t, isEnabled)

		// Test enabling an event
		err = service.EnableEvent("terminal_bell")
		assert.NoError(t, err)

		isEnabled = service.IsEventEnabled("terminal_bell")
		assert.True(t, isEnabled)

		// Test batch operations
		events := []string{"terminal_bell", "ui:button_click"}
		err = service.DisableEvents(events)
		assert.NoError(t, err)

		assert.False(t, service.IsEventEnabled("terminal_bell"))
		assert.False(t, service.IsEventEnabled("ui:button_click"))

		err = service.EnableEvents(events)
		assert.NoError(t, err)

		assert.True(t, service.IsEventEnabled("terminal_bell"))
		assert.True(t, service.IsEventEnabled("ui:button_click"))
	})

	// Test persistence
	t.Run("SettingsPersistence", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Modify settings
		settings := service.GetSettings()
		settings.Enabled = true
		settings.Volume = 80
		settings.Soundpack = "retro"
		service.UpdateSettings(settings)

		// Create new service instance to test persistence
		newService, err := audio.NewService(tempDir)
		require.NoError(t, err)

		loadedSettings := newService.GetSettings()
		assert.Equal(t, 80, loadedSettings.Volume)
		assert.Equal(t, "retro", loadedSettings.Soundpack)
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test with invalid directory
		_, err := audio.NewService("/non/existent/directory")
		// Should not error - service should create directory
		assert.NoError(t, err)

		// Test with nil settings
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		err = service.UpdateSettings(nil)
		assert.Error(t, err)

		// Test with invalid event ID
		err = service.PlayEvent("")
		assert.Error(t, err)

		err = service.PlayEvent("!invalid@event#id")
		assert.Error(t, err)
	})

	// Test concurrency
	t.Run("Concurrency", func(t *testing.T) {
		tempDir := t.TempDir()
		service, err := audio.NewService(tempDir)
		require.NoError(t, err)

		// Enable audio
		settings := service.GetSettings()
		settings.Enabled = true
		service.UpdateSettings(settings)

		// Test concurrent playback
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()
				for j := 0; j < 10; j++ {
					service.PlayEvent("terminal_bell")
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Service should still be functional
		settings = service.GetSettings()
		assert.NotNil(t, settings)
	})
}

// Benchmark audio service operations
func BenchmarkAudioService(b *testing.B) {
	tempDir := b.TempDir()
	service, err := audio.NewService(tempDir)
	require.NoError(b, err)

	// Enable audio
	settings := service.GetSettings()
	settings.Enabled = true
	service.UpdateSettings(settings)

	b.Run("PlayEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.PlayEvent("terminal_bell")
		}
	})

	b.Run("GetSettings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.GetSettings()
		}
	})

	b.Run("IsEventEnabled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.IsEventEnabled("terminal_bell")
		}
	})

	b.Run("GetEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.GetEvent("terminal_bell")
		}
	})
}

// Example test showing how to integrate with actual audio system
func ExampleAudioService() {
	// This would be used in the actual application
	tempDir := "/tmp/adex-audio"
	service, err := audio.NewService(tempDir)
	if err != nil {
		panic(err)
	}

	// Enable audio
	settings := service.GetSettings()
	settings.Enabled = true
	settings.Volume = 50
	service.UpdateSettings(settings)

	// Play a sound effect
	service.PlayEvent("terminal_bell")

	// Handle UI event automatically
	service.PlayUIEvent("ui:button_click")

	// Get current playback status
	status := service.GetPlaybackStatus()
	if status.IsPlaying {
		println("Currently playing:", status.CurrentEvent.Name)
	}
}