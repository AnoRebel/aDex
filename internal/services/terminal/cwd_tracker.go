package terminal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// CWDTracker tracks current working directory changes in terminal sessions
type CWDTracker struct {
	eventBus       events.IEventBus
	logger         *logger.Logger
	sessions       map[string]*CWDTrackingSession
	mutex          sync.RWMutex
	enabled        bool
	pollInterval   time.Duration
	commandPattern *regexp.Regexp
	windowsTracker *WindowsCWDTracker
}

// CWDTrackingSession tracks CWD for a specific terminal session
type CWDTrackingSession struct {
	SessionID     string
	CurrentCWD    string
	PreviousCWD   string
	PTY           *PTY
	LastUpdate    time.Time
	IsTracking    bool
	CommandBuffer []byte
	PromptPattern *regexp.Regexp
	ShellType     string
	mutex         sync.RWMutex
}

// CWDTrackingConfig holds configuration for CWD tracking
type CWDTrackingConfig struct {
	Enabled          bool          `json:"enabled"`
	PollInterval     time.Duration `json:"poll_interval"`
	TrackCommands    bool          `json:"track_commands"`
	AutoDetectShell  bool          `json:"auto_detect_shell"`
	EnablePTYPolling bool          `json:"enable_pty_polling"`
}

// DefaultCWDTrackingConfig returns default configuration
func DefaultCWDTrackingConfig() *CWDTrackingConfig {
	return &CWDTrackingConfig{
		Enabled:          true,
		PollInterval:     500 * time.Millisecond,
		TrackCommands:    true,
		AutoDetectShell:  true,
		EnablePTYPolling: true,
	}
}

// NewCWDTracker creates a new CWD tracker
func NewCWDTracker(config *CWDTrackingConfig) *CWDTracker {
	if config == nil {
		config = DefaultCWDTrackingConfig()
	}

	tracker := &CWDTracker{
		enabled:      config.Enabled,
		pollInterval: config.PollInterval,
		sessions:     make(map[string]*CWDTrackingSession),
		// Command pattern matches common directory change commands
		commandPattern: regexp.MustCompile(`(?i)^\s*(cd|pushd|popd)\s+['"]?([^'"\s]+)['"]?\s*(?:&&|\|\||;|$)`),
	}

	// Platform-specific initialization
	switch runtime.GOOS {
	case "linux", "darwin":
		tracker.enabled = tracker.enabled && config.EnablePTYPolling
	case "windows":
		// Initialize Windows-specific tracker
		windowsConfig := DefaultWindowsCWDConfig()
		windowsConfig.PollInterval = config.PollInterval
		windowsConfig.Enabled = config.Enabled
		tracker.windowsTracker = NewWindowsCWDTracker(windowsConfig)
	default:
		// Other platforms not supported
		tracker.enabled = false
	}

	return tracker
}

// Initialize initializes the CWD tracker
func (t *CWDTracker) Initialize(ctx context.Context) error {
	t.logger = logger.GetDefaultLogger()
	t.logger.Info("Initializing CWD tracker", map[string]interface{}{
		"enabled":       t.enabled,
		"platform":      runtime.GOOS,
		"poll_interval": t.pollInterval,
	})

	// Initialize Windows tracker if available
	if t.windowsTracker != nil {
		if err := t.windowsTracker.Initialize(ctx); err != nil {
			t.logger.Warn("Failed to initialize Windows CWD tracker", map[string]interface{}{"error": err.Error()})
		} else {
			t.logger.Info("Windows CWD tracker initialized successfully", nil)
		}
	}

	if t.enabled {
		// Start background monitoring
		go t.monitoringLoop(ctx)
	}

	return nil
}

// SetEventBus sets the event bus
func (t *CWDTracker) SetEventBus(eventBus events.IEventBus) {
	t.eventBus = eventBus
	if t.windowsTracker != nil {
		t.windowsTracker.SetEventBus(eventBus)
	}
}

// AddSession adds a terminal session for CWD tracking
func (t *CWDTracker) AddSession(sessionID string, pty *PTY, initialCWD string) error {
	if !t.enabled {
		return nil
	}

	// On Windows, use the Windows-specific tracker
	if runtime.GOOS == "windows" && t.windowsTracker != nil {
		return t.windowsTracker.AddSession(sessionID, pty, initialCWD)
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.sessions[sessionID]; exists {
		return fmt.Errorf("session %s is already being tracked", sessionID)
	}

	// Detect shell type for better prompt detection
	shellType := t.detectShellType(pty)

	// Create prompt patterns based on shell type
	promptPattern := t.createPromptPattern(shellType)

	session := &CWDTrackingSession{
		SessionID:     sessionID,
		CurrentCWD:    initialCWD,
		PreviousCWD:   initialCWD,
		PTY:           pty,
		LastUpdate:    time.Now(),
		IsTracking:    true,
		CommandBuffer: make([]byte, 0),
		PromptPattern: promptPattern,
		ShellType:     shellType,
	}

	t.sessions[sessionID] = session

	t.logger.Info("Added session to CWD tracker", map[string]interface{}{
		"session_id":  sessionID,
		"initial_cwd": initialCWD,
		"shell_type":  shellType,
	})

	return nil
}

// RemoveSession removes a terminal session from CWD tracking
func (t *CWDTracker) RemoveSession(sessionID string) {
	// On Windows, use the Windows-specific tracker
	if runtime.GOOS == "windows" && t.windowsTracker != nil {
		t.windowsTracker.RemoveSession(sessionID)
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if session, exists := t.sessions[sessionID]; exists {
		session.IsTracking = false
		delete(t.sessions, sessionID)

		t.logger.Info("Removed session from CWD tracker", map[string]interface{}{
			"session_id": sessionID,
			"final_cwd":  session.CurrentCWD,
		})
	}
}

// GetCurrentCWD returns the current working directory for a session
func (t *CWDTracker) GetCurrentCWD(sessionID string) (string, error) {
	// On Windows, use the Windows-specific tracker
	if runtime.GOOS == "windows" && t.windowsTracker != nil {
		return t.windowsTracker.GetCurrentCWD(sessionID)
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	session, exists := t.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.CurrentCWD, nil
}

// ProcessOutput processes terminal output to detect directory changes
func (t *CWDTracker) ProcessOutput(sessionID string, data []byte) {
	if !t.enabled {
		return
	}

	// On Windows, delegate to the Windows-specific tracker
	if runtime.GOOS == "windows" && t.windowsTracker != nil {
		t.windowsTracker.ProcessOutput(sessionID, data)
		return
	}

	t.mutex.RLock()
	session, exists := t.sessions[sessionID]
	t.mutex.RUnlock()

	if !exists || !session.IsTracking {
		return
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	// Append data to command buffer
	session.CommandBuffer = append(session.CommandBuffer, data...)

	// Process command buffer for directory changes
	t.processCommandBuffer(session)

	// Update last activity
	session.LastUpdate = time.Now()
}

// processCommandBuffer processes the accumulated command buffer for directory changes
func (t *CWDTracker) processCommandBuffer(session *CWDTrackingSession) {
	data := string(session.CommandBuffer)

	// Look for complete commands (lines ending with newline)
	lines := strings.Split(data, "\n")

	// Keep the incomplete last line in buffer
	if len(lines) > 1 {
		// Process all complete lines
		for _, line := range lines[:len(lines)-1] {
			t.processCommandLine(session, line)
		}

		// Keep the incomplete line
		session.CommandBuffer = []byte(lines[len(lines)-1])
	}

	// Also check for directory change patterns without waiting for newline
	if t.commandPattern.Match(session.CommandBuffer) {
		t.processDirectoryChangeCommand(session, string(session.CommandBuffer))
		// Clear buffer after processing
		session.CommandBuffer = make([]byte, 0)
	}
}

// processCommandLine processes a single command line
func (t *CWDTracker) processCommandLine(session *CWDTrackingSession, line string) {
	line = strings.TrimSpace(line)

	// Skip empty lines and prompts
	if line == "" || session.PromptPattern.MatchString(line) {
		return
	}

	// Check for directory change commands
	if t.commandPattern.MatchString(line) {
		t.processDirectoryChangeCommand(session, line)
	}

	// Check for other directory-related commands
	t.processDirectoryRelatedCommands(session, line)
}

// processDirectoryChangeCommand processes cd/pushd/popd commands
func (t *CWDTracker) processDirectoryChangeCommand(session *CWDTrackingSession, commandLine string) {
	matches := t.commandPattern.FindStringSubmatch(commandLine)
	if len(matches) < 3 {
		return
	}

	command := strings.ToLower(matches[1])
	targetPath := matches[2]

	var newCWD string
	var err error

	// Calculate the new working directory
	switch command {
	case "cd":
		newCWD, err = t.resolvePath(session.CurrentCWD, targetPath)
	case "pushd":
		newCWD, err = t.resolvePath(session.CurrentCWD, targetPath)
		// pushd also adds to directory stack - would need additional handling for full implementation
	case "popd":
		// popd would need directory stack tracking - for now, try to resolve reasonably
		newCWD, err = t.resolvePath(session.CurrentCWD, targetPath)
	default:
		return
	}

	if err != nil {
		t.logger.Warn("Failed to resolve new working directory", map[string]interface{}{
			"error":       err.Error(),
			"session_id":  session.SessionID,
			"command":     commandLine,
			"target_path": targetPath,
			"current_cwd": session.CurrentCWD,
		})
		return
	}

	// Verify the directory exists
	if _, err := os.Stat(newCWD); os.IsNotExist(err) {
		t.logger.Debug("Target directory does not exist", map[string]interface{}{
			"session_id":  session.SessionID,
			"target_path": newCWD,
			"command":     commandLine,
		})
		return
	}

	// Update the working directory if it changed
	if newCWD != session.CurrentCWD {
		session.PreviousCWD = session.CurrentCWD
		session.CurrentCWD = newCWD
		session.LastUpdate = time.Now()

		t.logger.Info("Working directory changed", map[string]interface{}{
			"session_id":   session.SessionID,
			"previous_cwd": session.PreviousCWD,
			"new_cwd":      newCWD,
			"command":      commandLine,
		})

		// Publish directory change event
		t.publishDirectoryChanged(session.SessionID, newCWD, session.PreviousCWD, command)
	}
}

// processDirectoryRelatedCommands processes other directory-related commands
func (t *CWDTracker) processDirectoryRelatedCommands(session *CWDTrackingSession, line string) {
	// Handle other commands that might affect the current directory
	// For example: source .bashrc, . ~/.profile, etc.
	patterns := []struct {
		pattern *regexp.Regexp
		handler func(*CWDTrackingSession, string, []string)
	}{
		{
			pattern: regexp.MustCompile(`(?i)^\s*(source|\.)\s+['"]?([^'"\s]+)['"]?`),
			handler: t.handleSourceCommand,
		},
		{
			pattern: regexp.MustCompile(`(?i)^\s*exec\s+(['"]?[^'"\s]+['"]?)`),
			handler: t.handleExecCommand,
		},
	}

	for _, p := range patterns {
		if matches := p.pattern.FindStringSubmatch(line); len(matches) >= 2 {
			p.handler(session, line, matches)
			break
		}
	}
}

// handleSourceCommand handles source/. commands that might change the environment
func (t *CWDTracker) handleSourceCommand(session *CWDTrackingSession, line string, matches []string) {
	// Source commands might change directory indirectly through environment setup
	// For now, we'll trigger a CWD check after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		t.checkCWDViaPTY(session)
	}()
}

// handleExecCommand handles exec commands that replace the current shell
func (t *CWDTracker) handleExecCommand(session *CWDTrackingSession, line string, matches []string) {
	// exec replaces the current shell process, so we need to re-detect the shell
	go func() {
		time.Sleep(200 * time.Millisecond)
		if session.PTY != nil && session.PTY.IsActive() {
			newShellType := t.detectShellType(session.PTY)
			if newShellType != session.ShellType {
				session.ShellType = newShellType
				session.PromptPattern = t.createPromptPattern(newShellType)

				t.logger.Info("Shell type changed after exec", map[string]interface{}{
					"session_id": session.SessionID,
					"old_shell":  session.ShellType,
					"new_shell":  newShellType,
				})
			}
		}
	}()
}

// resolvePath resolves a path relative to the current working directory
func (t *CWDTracker) resolvePath(currentCWD, targetPath string) (string, error) {
	if targetPath == "" {
		return currentCWD, nil
	}

	// Handle special cases
	switch targetPath {
	case "~":
		home, err := os.UserHomeDir()
		if err != nil {
			return currentCWD, err
		}
		return home, nil
	case "-":
		// Return to previous directory (if we have one)
		return currentCWD, nil // Would need proper stack implementation
	case "..":
		return filepath.Dir(currentCWD), nil
	case ".":
		return currentCWD, nil
	}

	// Handle absolute paths
	if filepath.IsAbs(targetPath) {
		return filepath.Clean(targetPath), nil
	}

	// Handle relative paths
	return filepath.Clean(filepath.Join(currentCWD, targetPath)), nil
}

// detectShellType detects the shell type from PTY information
func (t *CWDTracker) detectShellType(pty *PTY) string {
	if pty == nil || pty.cmd == nil {
		return "unknown"
	}

	shellPath := pty.cmd.Path
	shellName := filepath.Base(shellPath)

	// Extract shell name without extension
	shellName = strings.TrimSuffix(shellName, filepath.Ext(shellName))

	switch strings.ToLower(shellName) {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "sh":
		return "sh"
	case "dash":
		return "dash"
	case "ksh":
		return "ksh"
	case "csh":
		return "csh"
	case "tcsh":
		return "tcsh"
	default:
		return "unknown"
	}
}

// createPromptPattern creates a regex pattern for detecting shell prompts
func (t *CWDTracker) createPromptPattern(shellType string) *regexp.Regexp {
	switch shellType {
	case "bash":
		// Bash prompt patterns: user@host:path$ or PS1 variations
		return regexp.MustCompile(`^[^$#\s]+@[^:\s]+:[^$#\s]+[$#]?\s*$`)
	case "zsh":
		// Zsh prompt patterns: user@host path% or custom prompts
		return regexp.MustCompile(`^[^%#\s]+@[^:\s]+\s+[^%#\s]+[%#]?\s*$`)
	case "fish":
		// Fish prompt patterns: user@hostname path>
		return regexp.MustCompile(`^[^>\s]+@[^>\s]+[^>]+>\s*$`)
	default:
		// Generic prompt pattern
		return regexp.MustCompile(`^[>$#]\s*$`)
	}
}

// checkCWDViaPTY checks the current working directory via PTY (Linux/macOS only)
func (t *CWDTracker) checkCWDViaPTY(session *CWDTrackingSession) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return
	}

	if session.PTY == nil || !session.PTY.IsActive() {
		return
	}

	// Send a pwd command to check current directory
	pwdCommand := []byte("pwd\n")

	// Write the command to the PTY
	if err := session.PTY.Write(pwdCommand); err != nil {
		t.logger.Warn("Failed to write pwd command to PTY", map[string]interface{}{
			"error":      err.Error(),
			"session_id": session.SessionID,
		})
		return
	}

	// The output will be captured by ProcessOutput and handled there
}

// monitoringLoop runs background monitoring for all sessions
func (t *CWDTracker) monitoringLoop(ctx context.Context) {
	if !t.enabled {
		return
	}

	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.performPeriodicChecks()
		}
	}
}

// performPeriodicChecks performs periodic CWD checks for all sessions
func (t *CWDTracker) performPeriodicChecks() {
	t.mutex.RLock()
	sessions := make([]*CWDTrackingSession, 0, len(t.sessions))
	for _, session := range t.sessions {
		if session.IsTracking {
			sessions = append(sessions, session)
		}
	}
	t.mutex.RUnlock()

	for _, session := range sessions {
		// Skip recent updates to avoid excessive polling
		if time.Since(session.LastUpdate) < t.pollInterval {
			continue
		}

		// For Linux/macOS, periodically verify CWD via PTY
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			go t.checkCWDViaPTY(session)
		}
	}
}

// publishDirectoryChanged publishes a directory changed event
func (t *CWDTracker) publishDirectoryChanged(sessionID, newCWD, previousCWD, command string) {
	if t.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "directory_changed",
		CWD:       newCWD,
		Timestamp: time.Now(),
		Command:   command,
	}

	err := t.eventBus.Publish(context.Background(), "terminal.directory.changed", data, "cwd-tracker")
	if err != nil {
		t.logger.Error("Failed to publish directory changed event", err, map[string]interface{}{
			"session_id":   sessionID,
			"new_cwd":      newCWD,
			"previous_cwd": previousCWD,
			"command":      command,
		})
	}
}

// GetStats returns tracking statistics
func (t *CWDTracker) GetStats() map[string]interface{} {
	// On Windows, delegate to the Windows-specific tracker
	if runtime.GOOS == "windows" && t.windowsTracker != nil {
		return t.windowsTracker.GetStats()
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	stats := map[string]interface{}{
		"enabled":         t.enabled,
		"total_sessions":  len(t.sessions),
		"active_sessions": 0,
		"platform":        runtime.GOOS,
		"method":          "pty_command_monitoring",
		"poll_interval":   t.pollInterval.String(),
	}

	for _, session := range t.sessions {
		if session.IsTracking {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1
		}
	}

	return stats
}

// Shutdown shuts down the CWD tracker
func (t *CWDTracker) Shutdown(ctx context.Context) error {
	t.logger.Info("Shutting down CWD tracker")

	// Shutdown Windows tracker if available
	if t.windowsTracker != nil {
		if err := t.windowsTracker.Shutdown(ctx); err != nil {
			t.logger.Warn("Failed to shutdown Windows CWD tracker", map[string]interface{}{"error": err.Error()})
		}
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Stop tracking all sessions
	for sessionID, session := range t.sessions {
		session.IsTracking = false
		t.logger.Debug("Stopped tracking session", map[string]interface{}{
			"session_id": sessionID,
			"final_cwd":  session.CurrentCWD,
		})
	}

	// Clear sessions
	t.sessions = make(map[string]*CWDTrackingSession)

	t.logger.Info("CWD tracker shutdown complete")
	return nil
}
