package audio

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"aDex-UI/internal/utils"
)

// Service handles audio operations
type Service struct {
	platform       *utils.FeatureDetection
	otoCtx         *oto.Context
	readyChan      <-chan struct{}
	players        map[string]*oto.Player
	soundFiles     map[string]string
	sessions       map[string]*AudioSession
	isMonitoring   bool
	stopChan       chan struct{}
	mu             sync.RWMutex
	isInitialized  bool
}

// NewService creates a new audio service instance.
//
// Setting ADEX_DISABLE_AUDIO=1 in the environment skips PCM context creation,
// which is required for headless test runs where the oto/v3 PCM writer
// goroutine would otherwise outlive the test binary.
func NewService() *Service {
	service := &Service{
		platform:   utils.DetectPlatform(),
		players:    make(map[string]*oto.Player),
		soundFiles: make(map[string]string),
		sessions:   make(map[string]*AudioSession),
		stopChan:   make(chan struct{}),
	}

	if os.Getenv("ADEX_DISABLE_AUDIO") == "1" {
		service.isInitialized = false
		return service
	}

	// Try to initialize audio context
	if err := service.initializeAudio(); err != nil {
		// Audio not available, but service can still be used for file management
		service.isInitialized = false
	}

	return service
}

// initializeAudio sets up the OTO audio context
func (s *Service) initializeAudio() error {
	// Configure audio context options
	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatFloat32LE,
	}

	// Create audio context
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return fmt.Errorf("failed to create audio context: %w", err)
	}

	s.otoCtx = otoCtx
	s.readyChan = readyChan
	s.isInitialized = true

	// Start a goroutine to wait for audio to be ready
	go func() {
		<-readyChan
	}()

	return nil
}

// AudioDevice represents an audio device
type AudioDevice struct {
	ID          string
	Name        string
	Type        string // input, output, or both
	IsDefault   bool
	SampleRate  int
	Channels    int
	BufferSize  int
}

// AudioSession represents an active audio session
type AudioSession struct {
	ID         string
	DeviceID   string
	ProcessID  int
	ProcessName string
	State      string // playing, paused, stopped
	Volume     float64
}

// GetDevices retrieves available audio devices
func (s *Service) GetDevices(ctx context.Context) ([]AudioDevice, error) {
	if !s.platform.HasFeature("audio") {
		return nil, fmt.Errorf("audio not supported on this platform")
	}

	switch s.platform.Platform.OS {
	case "windows":
		return s.getWindowsDevices(ctx)
	case "darwin":
		return s.getMacOSDevices(ctx)
	default:
		return s.getLinuxDevices(ctx)
	}
}

// PlaySound plays a sound file using oto
func (s *Service) PlaySound(ctx context.Context, soundPath string) error {
	if !s.isInitialized {
		return fmt.Errorf("audio context not initialized")
	}

	// Check if audio context is ready
	select {
	case <-s.readyChan:
		// Ready to play
	case <-time.After(5 * time.Second):
		return fmt.Errorf("audio context not ready")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Open the sound file
	file, err := os.Open(soundPath)
	if err != nil {
		return fmt.Errorf("failed to open sound file: %w", err)
	}
	defer file.Close()

	// Create a new player for this sound
	player := s.otoCtx.NewPlayer(file)
	if player == nil {
		return fmt.Errorf("failed to create audio player")
	}

	// Store the player for management
	playerID := filepath.Base(soundPath)
	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	// Play the sound
	player.Play()

	// Start a goroutine to clean up when done
	go func() {
		for player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
		}
		player.Close()

		s.mu.Lock()
		delete(s.players, playerID)
		s.mu.Unlock()
	}()

	return nil
}

// PlaySoundFromBytes plays sound from byte data
func (s *Service) PlaySoundFromBytes(ctx context.Context, soundData []byte, format string) error {
	if !s.isInitialized {
		return fmt.Errorf("audio context not initialized")
	}

	// Check if audio context is ready
	select {
	case <-s.readyChan:
		// Ready to play
	case <-time.After(5 * time.Second):
		return fmt.Errorf("audio context not ready")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Create a reader from byte data
	reader := bytes.NewReader(soundData)

	// Create a new player
	player := s.otoCtx.NewPlayer(reader)
	if player == nil {
		return fmt.Errorf("failed to create audio player")
	}

	// Store the player
	playerID := fmt.Sprintf("byte-sound-%d", time.Now().UnixNano())
	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	// Play the sound
	player.Play()

	// Clean up when done
	go func() {
		for player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
		}
		player.Close()

		s.mu.Lock()
		delete(s.players, playerID)
		s.mu.Unlock()
	}()

	return nil
}

// StopAllSounds stops all currently playing sounds
func (s *Service) StopAllSounds() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, player := range s.players {
		player.Close()
		delete(s.players, id)
	}
}

// LoadSoundFile loads a sound file for later playback
func (s *Service) LoadSoundFile(name, filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("sound file not found: %w", err)
	}

	s.mu.Lock()
	s.soundFiles[name] = filePath
	s.mu.Unlock()

	return nil
}

// PlayLoadedSound plays a previously loaded sound
func (s *Service) PlayLoadedSound(ctx context.Context, soundName string) error {
	s.mu.RLock()
	filePath, exists := s.soundFiles[soundName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("sound not loaded: %s", soundName)
	}

	return s.PlaySound(ctx, filePath)
}

// GetActiveSessions retrieves active audio sessions
func (s *Service) GetActiveSessions(ctx context.Context) ([]AudioSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []AudioSession
	for _, session := range s.sessions {
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

// SetDeviceVolume sets the volume for a device
func (s *Service) SetDeviceVolume(ctx context.Context, deviceID string, volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	switch s.platform.Platform.OS {
	case "windows":
		return s.setWindowsVolume(ctx, deviceID, volume)
	case "darwin":
		return s.setMacOSVolume(ctx, deviceID, volume)
	default:
		return s.setLinuxVolume(ctx, deviceID, volume)
	}
}

// GetDeviceVolume gets the current volume for a device
func (s *Service) GetDeviceVolume(ctx context.Context, deviceID string) (float64, error) {
	switch s.platform.Platform.OS {
	case "windows":
		return s.getWindowsVolume(ctx, deviceID)
	case "darwin":
		return s.getMacOSVolume(ctx, deviceID)
	default:
		return s.getLinuxVolume(ctx, deviceID)
	}
}

// MuteDevice mutes/unmutes a device
func (s *Service) MuteDevice(ctx context.Context, deviceID string, muted bool) error {
	switch s.platform.Platform.OS {
	case "windows":
		return s.muteWindowsDevice(ctx, deviceID, muted)
	case "darwin":
		return s.muteMacOSDevice(ctx, deviceID, muted)
	default:
		return s.muteLinuxDevice(ctx, deviceID, muted)
	}
}

// SetDefaultDevice sets a device as default
func (s *Service) SetDefaultDevice(ctx context.Context, deviceID string) error {
	switch s.platform.Platform.OS {
	case "windows":
		return s.setWindowsDefaultDevice(ctx, deviceID)
	case "darwin":
		return s.setMacOSDefaultDevice(ctx, deviceID)
	default:
		return s.setLinuxDefaultDevice(ctx, deviceID)
	}
}

// StartMonitoring starts audio monitoring
func (s *Service) StartMonitoring(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isMonitoring {
		return fmt.Errorf("audio monitoring is already running")
	}

	s.isMonitoring = true
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer ticker.Stop()
		defer func() { s.isMonitoring = false }()

		for {
			select {
			case <-ticker.C:
				s.updateAudioSessions(ctx)

			case <-s.stopChan:
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// StopMonitoring stops audio monitoring
func (s *Service) StopMonitoring() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isMonitoring {
		return nil
	}

	close(s.stopChan)
	s.stopChan = make(chan struct{})
	s.isMonitoring = false

	return nil
}

// Shutdown stops monitoring, closes any active players, and suspends the
// oto PCM context so its CGO writer goroutine winds down. Without this,
// the goroutine outlives the test binary on Linux ALSA and `go test` hangs.
//
// Suspend() is the closest oto/v3 offers to a teardown — it stops the
// writei loop until Resume() is called. We don't Resume in the same
// process; a fresh NewService() will create a new context.
func (s *Service) Shutdown(ctx context.Context) error {
	// Stop the monitoring goroutine first (uses its own lock).
	_ = s.StopMonitoring()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.players {
		p.Close()
	}
	s.players = nil

	if s.otoCtx != nil {
		// Best-effort suspend; we don't fail Shutdown on a suspend error
		// because the PCM device may already be in an unrecoverable state.
		_ = s.otoCtx.Suspend()
		s.otoCtx = nil
	}
	s.isInitialized = false
	return nil
}

// Platform-specific device detection methods

// getWindowsDevices retrieves audio devices on Windows
func (s *Service) getWindowsDevices(ctx context.Context) ([]AudioDevice, error) {
	var devices []AudioDevice

	// Try to use PowerShell to get audio devices
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Get-WmiObject -Class Win32_SoundDevice | Select-Object Name, DeviceID | ConvertTo-Json")

	output, err := cmd.Output()
	if err != nil {
		// Fallback: return default devices
		return []AudioDevice{
			{
				ID:         "default",
				Name:       "Default Audio Device",
				Type:       "output",
				IsDefault:  true,
				SampleRate: 44100,
				Channels:   2,
				BufferSize: 1024,
			},
		}, nil
	}

	// Parse PowerShell output (simplified)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Name") {
			// Extract device name (simplified parsing)
			name := strings.TrimSpace(strings.Split(line, ":")[1])
			if name != "" {
				devices = append(devices, AudioDevice{
					ID:         fmt.Sprintf("windows-%d", len(devices)),
					Name:       strings.Trim(name, `",`),
					Type:       "output",
					IsDefault:  len(devices) == 0,
					SampleRate: 44100,
					Channels:   2,
					BufferSize: 1024,
				})
			}
		}
	}

	return devices, nil
}

// getMacOSDevices retrieves audio devices on macOS
func (s *Service) getMacOSDevices(ctx context.Context) ([]AudioDevice, error) {
	var devices []AudioDevice

	// Use system_profiler to get audio devices
	cmd := exec.CommandContext(ctx, "system_profiler", "SPAudioDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: use switchaudio-osx or return default
		return []AudioDevice{
			{
				ID:         "default",
				Name:       "Built-in Output",
				Type:       "output",
				IsDefault:  true,
				SampleRate: 44100,
				Channels:   2,
				BufferSize: 1024,
			},
		}, nil
	}

	// Parse system_profiler output (simplified)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "_name") && !strings.Contains(line, "Device") {
			name := strings.TrimSpace(strings.Split(line, ":")[1])
			if name != "" && name != `""` {
				devices = append(devices, AudioDevice{
					ID:         fmt.Sprintf("macos-%d", len(devices)),
					Name:       strings.Trim(name, `",`),
					Type:       "output",
					IsDefault:  len(devices) == 0,
					SampleRate: 44100,
					Channels:   2,
					BufferSize: 1024,
				})
			}
		}
	}

	return devices, nil
}

// getLinuxDevices retrieves audio devices on Linux
func (s *Service) getLinuxDevices(ctx context.Context) ([]AudioDevice, error) {
	// Try to use pactl (PulseAudio) first
	cmd := exec.CommandContext(ctx, "pactl", "list", "sinks")
	output, err := cmd.Output()
	if err == nil {
		return s.parsePulseAudioDevices(string(output)), nil
	}

	// Fallback to ALSA
	cmd = exec.CommandContext(ctx, "aplay", "-l")
	output, err = cmd.Output()
	if err == nil {
		return s.parseALSADevices(string(output)), nil
	}

	// Final fallback
	return []AudioDevice{
		{
			ID:         "default",
			Name:       "Default Audio Device",
			Type:       "output",
			IsDefault:  true,
			SampleRate: 44100,
			Channels:   2,
			BufferSize: 1024,
		},
	}, nil
}

// parsePulseAudioDevices parses pactl output
func (s *Service) parsePulseAudioDevices(output string) []AudioDevice {
	var devices []AudioDevice
	lines := strings.Split(output, "\n")

	var currentDevice *AudioDevice
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Sink #") {
			if currentDevice != nil {
				devices = append(devices, *currentDevice)
			}
			currentDevice = &AudioDevice{
				Type:       "output",
				SampleRate: 44100,
				Channels:   2,
				BufferSize: 1024,
			}
		} else if strings.HasPrefix(line, "Description:") && currentDevice != nil {
			currentDevice.Name = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			currentDevice.ID = fmt.Sprintf("pulse-%d", len(devices))
			currentDevice.IsDefault = len(devices) == 0
		}
	}

	if currentDevice != nil {
		devices = append(devices, *currentDevice)
	}

	return devices
}

// parseALSADevices parses aplay output
func (s *Service) parseALSADevices(output string) []AudioDevice {
	var devices []AudioDevice
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "card") && strings.Contains(line, "device") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				name := strings.Join(parts[2:], " ")
				name = strings.Trim(name, "[]")
				devices = append(devices, AudioDevice{
					ID:         fmt.Sprintf("alsa-%d", len(devices)),
					Name:       name,
					Type:       "output",
					IsDefault:  len(devices) == 0,
					SampleRate: 44100,
					Channels:   2,
					BufferSize: 1024,
				})
			}
		}
	}

	return devices
}

// Platform-specific volume control methods

// setWindowsVolume sets volume on Windows using PowerShell
func (s *Service) setWindowsVolume(ctx context.Context, deviceID string, volume float64) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("(New-Object -comObject WScript.Shell).SendKeys([char]175)"))
	return cmd.Run()
}

// setMacOSVolume sets volume on macOS using osascript
func (s *Service) setMacOSVolume(ctx context.Context, deviceID string, volume float64) error {
	volumePercent := int(volume * 100)
	cmd := exec.CommandContext(ctx, "osascript", "-e",
		fmt.Sprintf("set volume output volume %d", volumePercent))
	return cmd.Run()
}

// setLinuxVolume sets volume on Linux using amixer
func (s *Service) setLinuxVolume(ctx context.Context, deviceID string, volume float64) error {
	volumePercent := int(volume * 100)
	cmd := exec.CommandContext(ctx, "amixer", "sset", "Master", fmt.Sprintf("%d%%", volumePercent))
	return cmd.Run()
}

// getWindowsVolume gets current volume on Windows
func (s *Service) getWindowsVolume(ctx context.Context, deviceID string) (float64, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Get-AudioDevice -List | Where-Object {$_.Type -eq 'Playback'} | Select-Object -First 1 | Get-AudioDeviceVolume")

	output, err := cmd.Output()
	if err != nil {
		return 0.5, nil // Return default volume on error
	}

	// Parse output to get volume percentage (simplified)
	volumeStr := strings.TrimSpace(string(output))
	if volumeStr == "" {
		return 0.5, nil
	}

	// Default to 50% if parsing fails
	return 0.5, nil
}

// getMacOSVolume gets current volume on macOS
func (s *Service) getMacOSVolume(ctx context.Context, deviceID string) (float64, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", "output volume of (get volume settings)")

	output, err := cmd.Output()
	if err != nil {
		return 0.5, nil
	}

	volumeStr := strings.TrimSpace(string(output))
	volume := 0.5
	if volumeStr != "" {
		if vol, err := strconv.Atoi(volumeStr); err == nil {
			volume = float64(vol) / 100.0
		}
	}

	return volume, nil
}

// getLinuxVolume gets current volume on Linux
func (s *Service) getLinuxVolume(ctx context.Context, deviceID string) (float64, error) {
	cmd := exec.CommandContext(ctx, "amixer", "get", "Master")

	output, err := cmd.Output()
	if err != nil {
		return 0.5, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "%]") {
			start := strings.LastIndex(line, "[") + 1
			end := strings.LastIndex(line, "%]")
			if start > 0 && end > start {
				volumeStr := line[start:end]
				if volume, err := strconv.Atoi(volumeStr); err == nil {
					return float64(volume) / 100.0, nil
				}
			}
		}
	}

	return 0.5, nil
}

// muteWindowsDevice mutes/unmutes device on Windows
func (s *Service) muteWindowsDevice(ctx context.Context, deviceID string, muted bool) error {
	if muted {
		cmd := exec.CommandContext(ctx, "powershell", "-Command",
			"(New-Object -comObject WScript.Shell).SendKeys([char]173)")
		return cmd.Run()
	} else {
		cmd := exec.CommandContext(ctx, "powershell", "-Command",
			"(New-Object -comObject WScript.Shell).SendKeys([char]173)")
		return cmd.Run()
	}
}

// muteMacOSDevice mutes/unmutes device on macOS
func (s *Service) muteMacOSDevice(ctx context.Context, deviceID string, muted bool) error {
	if muted {
		cmd := exec.CommandContext(ctx, "osascript", "-e", "set volume with output muted")
		return cmd.Run()
	} else {
		cmd := exec.CommandContext(ctx, "osascript", "-e", "set volume without output muted")
		return cmd.Run()
	}
}

// muteLinuxDevice mutes/unmutes device on Linux
func (s *Service) muteLinuxDevice(ctx context.Context, deviceID string, muted bool) error {
	cmd := exec.CommandContext(ctx, "amixer", "sset", "Master", func() string {
		if muted {
			return "mute"
		}
		return "unmute"
	}())
	return cmd.Run()
}

// setWindowsDefaultDevice sets default device on Windows
func (s *Service) setWindowsDefaultDevice(ctx context.Context, deviceID string) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Set-AudioDevice -ID %s", deviceID))
	return cmd.Run()
}

// setMacOSDefaultDevice sets default device on macOS
func (s *Service) setMacOSDefaultDevice(ctx context.Context, deviceID string) error {
	// macOS default device switching is more complex and typically requires
	// using third-party tools or system preferences modifications
	return fmt.Errorf("setting default audio device on macOS is not supported")
}

// setLinuxDefaultDevice sets default device on Linux
func (s *Service) setLinuxDefaultDevice(ctx context.Context, deviceID string) error {
	cmd := exec.CommandContext(ctx, "pacmd", "set-default-sink", deviceID)
	return cmd.Run()
}

// updateAudioSessions updates the list of active audio sessions
func (s *Service) updateAudioSessions(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear current sessions
	s.sessions = make(map[string]*AudioSession)

	// Get current sessions based on platform
	switch s.platform.Platform.OS {
	case "windows":
		s.updateWindowsSessions(ctx)
	case "darwin":
		s.updateMacOSSessions(ctx)
	default:
		s.updateLinuxSessions(ctx)
	}
}

// updateWindowsSessions updates audio sessions on Windows
func (s *Service) updateWindowsSessions(ctx context.Context) {
	// Placeholder for Windows session detection
	// This would use Windows APIs to enumerate audio sessions
}

// updateMacOSSessions updates audio sessions on macOS
func (s *Service) updateMacOSSessions(ctx context.Context) {
	// Placeholder for macOS session detection
	// This would use CoreAudio APIs to enumerate audio sessions
}

// updateLinuxSessions updates audio sessions on Linux
func (s *Service) updateLinuxSessions(ctx context.Context) {
	// Placeholder for Linux session detection
	// This would use PulseAudio or ALSA APIs to enumerate audio sessions
}
