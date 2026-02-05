//go:build windows
// +build windows

package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// WindowsCWDTracker provides CWD tracking for Windows systems using detached processes
type WindowsCWDTracker struct {
	eventBus     events.IEventBus
	logger       *logger.Logger
	sessions     map[string]*WindowsCWDSession
	mutex        sync.RWMutex
	enabled      bool
	pollInterval time.Duration
	commandFile  string
	batchFile    string
}

// WindowsCWDSession tracks CWD for a Windows terminal session
type WindowsCWDSession struct {
	SessionID      string
	PID            int
	CurrentCWD     string
	PreviousCWD    string
	MonitorProcess *os.Process
	LastUpdate     time.Time
	IsTracking     bool
	ShellType      string
	mutex          sync.RWMutex
}

// WindowsCWDConfig holds configuration for Windows CWD tracking
type WindowsCWDConfig struct {
	Enabled           bool          `json:"enabled"`
	PollInterval      time.Duration `json:"poll_interval"`
	CreateHelperFiles bool          `json:"create_helper_files"`
	MaxRetries        int           `json:"max_retries"`
	Timeout           time.Duration `json:"timeout"`
}

// DefaultWindowsCWDConfig returns default configuration for Windows CWD tracking
func DefaultWindowsCWDConfig() *WindowsCWDConfig {
	return &WindowsCWDConfig{
		Enabled:           true,
		PollInterval:      1 * time.Second,
		CreateHelperFiles: true,
		MaxRetries:        3,
		Timeout:           30 * time.Second,
	}
}

// NewWindowsCWDTracker creates a new Windows CWD tracker
func NewWindowsCWDTracker(config *WindowsCWDConfig) *WindowsCWDTracker {
	if runtime.GOOS != "windows" {
		return &WindowsCWDTracker{
			enabled: false,
		}
	}

	tracker := &WindowsCWDTracker{
		enabled:      config.Enabled,
		pollInterval: config.PollInterval,
		sessions:     make(map[string]*WindowsCWDSession),
	}

	if config.CreateHelperFiles {
		tracker.setupHelperFiles()
	}

	return tracker
}

// Initialize initializes the Windows CWD tracker
func (t *WindowsCWDTracker) Initialize(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("Windows CWD tracker is only available on Windows")
	}

	t.logger = logger.GetDefaultLogger()
	t.logger.Info("Initializing Windows CWD tracker", map[string]interface{}{
		"enabled":       t.enabled,
		"poll_interval": t.pollInterval,
	})

	if t.enabled {
		// Start background monitoring
		go t.monitoringLoop(ctx)
	}

	return nil
}

// SetEventBus sets the event bus
func (t *WindowsCWDTracker) SetEventBus(eventBus events.IEventBus) {
	t.eventBus = eventBus
}

// setupHelperFiles creates helper batch and command files for CWD monitoring
func (t *WindowsCWDTracker) setupHelperFiles() {
	tempDir := os.TempDir()

	// Create a PowerShell script to get current working directory
	t.commandFile = filepath.Join(tempDir, "dex-cwd-tracker.ps1")
	psScript := `# PowerShell script to get current working directory of a process
param($ProcessId)

if (-not $ProcessId) {
    Write-Error "Process ID is required"
    exit 1
}

try {
    $process = Get-Process -Id $ProcessId -ErrorAction Stop
    if ($process.MainModule.FileName) {
        $cwd = (Get-Location).Path
        Write-Output $cwd
    } else {
        Write-Error "Process not found or inaccessible"
        exit 1
    }
} catch {
    Write-Error "Failed to get process information: $($_.Exception.Message)"
    exit 1
}
`

	if err := os.WriteFile(t.commandFile, []byte(psScript), 0644); err != nil {
		t.logger.Warn("Failed to create PowerShell helper script", map[string]interface{}{
			"file":  t.commandFile,
			"error": err.Error(),
		})
	}

	// Create a batch file wrapper
	t.batchFile = filepath.Join(tempDir, "dex-cwd-tracker.bat")
	batchScript := fmt.Sprintf(`@echo off
powershell -ExecutionPolicy Bypass -File "%s" -ProcessId %%1
`, t.commandFile)

	if err := os.WriteFile(t.batchFile, []byte(batchScript), 0644); err != nil {
		t.logger.Warn("Failed to create batch helper script", map[string]interface{}{
			"file":  t.batchFile,
			"error": err.Error(),
		})
	}
}

// AddSession adds a terminal session for Windows CWD tracking
func (t *WindowsCWDTracker) AddSession(sessionID string, pty *PTY, initialCWD string) error {
	if !t.enabled {
		return nil
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("Windows CWD tracking is only available on Windows")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, exists := t.sessions[sessionID]; exists {
		return fmt.Errorf("session %s is already being tracked", sessionID)
	}

	// Get the PID from PTY
	pid := 0
	if process := pty.GetProcess(); process != nil {
		pid = process.Pid
	}

	if pid == 0 {
		return fmt.Errorf("no process ID available for session %s", sessionID)
	}

	// Detect shell type
	shellType := t.detectWindowsShellType(pty)

	// Create monitor process for this session
	monitorProcess, err := t.createMonitorProcess(pid, sessionID)
	if err != nil {
		return fmt.Errorf("failed to create monitor process: %w", err)
	}

	session := &WindowsCWDSession{
		SessionID:      sessionID,
		PID:            pid,
		CurrentCWD:     initialCWD,
		PreviousCWD:    initialCWD,
		MonitorProcess: monitorProcess,
		LastUpdate:     time.Now(),
		IsTracking:     true,
		ShellType:      shellType,
	}

	t.sessions[sessionID] = session

	t.logger.Info("Added session to Windows CWD tracker", map[string]interface{}{
		"session_id":  sessionID,
		"pid":         pid,
		"initial_cwd": initialCWD,
		"shell_type":  shellType,
	})

	return nil
}

// RemoveSession removes a terminal session from Windows CWD tracking
func (t *WindowsCWDTracker) RemoveSession(sessionID string) {
	if runtime.GOOS != "windows" {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if session, exists := t.sessions[sessionID]; exists {
		session.IsTracking = false

		// Terminate monitor process
		if session.MonitorProcess != nil {
			if err := session.MonitorProcess.Kill(); err != nil {
				t.logger.Warn("Failed to terminate monitor process", err, map[string]interface{}{
					"session_id":  sessionID,
					"monitor_pid": session.MonitorProcess.Pid,
				})
			}
		}

		delete(t.sessions, sessionID)

		t.logger.Info("Removed session from Windows CWD tracker", map[string]interface{}{
			"session_id": sessionID,
			"final_cwd":  session.CurrentCWD,
		})
	}
}

// GetCurrentCWD returns the current working directory for a session
func (t *WindowsCWDTracker) GetCurrentCWD(sessionID string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("Windows CWD tracker is only available on Windows")
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

// createMonitorProcess creates a detached process to monitor the terminal's working directory
func (t *WindowsCWDTracker) createMonitorProcess(pid int, sessionID string) (*os.Process, error) {
	if t.batchFile == "" {
		return nil, fmt.Errorf("batch helper file not available")
	}

	// Create a detached process that will monitor the terminal
	cmd := exec.Command(t.batchFile, strconv.Itoa(pid))

	// Hide the window on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x08000000, // CREATE_NO_WINDOW
	}

	// Set up pipes to capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start monitor process: %w", err)
	}

	// Start a goroutine to read the output and update CWD
	go func() {
		defer stdout.Close()

		for {
			if !t.enabled {
				return
			}

			t.mutex.RLock()
			session, exists := t.sessions[sessionID]
			t.mutex.RUnlock()

			if !exists || !session.IsTracking {
				return
			}

			// Read current working directory from the monitor process
			cwd, err := t.getCurrentDirectoryFromProcess(pid)
			if err != nil {
				t.logger.Debug("Failed to get CWD from process", err, map[string]interface{}{
					"session_id": sessionID,
					"pid":        pid,
				})
				time.Sleep(t.pollInterval)
				continue
			}

			// Update CWD if it changed
			session.mutex.Lock()
			if cwd != "" && cwd != session.CurrentCWD {
				session.PreviousCWD = session.CurrentCWD
				session.CurrentCWD = cwd
				session.LastUpdate = time.Now()

				t.logger.Info("Working directory changed (Windows)", map[string]interface{}{
					"session_id":   sessionID,
					"previous_cwd": session.PreviousCWD,
					"new_cwd":      cwd,
				})

				// Publish directory change event
				t.publishDirectoryChanged(sessionID, cwd, session.PreviousCWD, "windows_monitor")
			}
			session.mutex.Unlock()

			time.Sleep(t.pollInterval)
		}
	}()

	return cmd.Process, nil
}

// getCurrentDirectoryFromProcess gets the current working directory of a Windows process
func (t *WindowsCWDTracker) getCurrentDirectoryFromProcess(pid int) (string, error) {
	// Method 1: Try using PowerShell to get the process working directory
	if t.commandFile != "" {
		cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", t.commandFile, "-ProcessId", strconv.Itoa(pid))
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}

		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			cwd := strings.TrimSpace(string(output))
			if cwd != "" && filepath.IsAbs(cwd) {
				return cwd, nil
			}
		}
	}

	// Method 2: Try using WMIC (Windows Management Instrumentation)
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "ExecutablePath")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			exePath := strings.TrimSpace(lines[1])
			if exePath != "" && filepath.IsAbs(exePath) {
				return filepath.Dir(exePath), nil
			}
		}
	}

	// Method 3: Try using tasklist
	cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		// Parse tasklist output to get image path
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, ".exe") {
				fields := strings.Split(line, ",")
				if len(fields) > 0 {
					imagePath := strings.Trim(fields[0], "\"")
					if filepath.IsAbs(imagePath) {
						return filepath.Dir(imagePath), nil
					}
				}
			}
		}
	}

	// Method 4: Last resort - try to inject and run a command
	return t.injectCWDQuery(pid)
}

// injectCWDQuery attempts to inject a CWD query command into the target process
func (t *WindowsCWDTracker) injectCWDQuery(pid int) (string, error) {
	// This is a simplified implementation
	// In a production environment, you might use Windows API calls or more sophisticated injection

	// Try to send a 'cd' command without arguments to get current directory
	// This is a fallback method and may not work in all cases

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Location | Select-Object -ExpandProperty Path"))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		cwd := strings.TrimSpace(string(output))
		if cwd != "" && filepath.IsAbs(cwd) {
			return cwd, nil
		}
	}

	return "", fmt.Errorf("failed to determine current working directory for PID %d", pid)
}

// detectWindowsShellType detects the shell type on Windows
func (t *WindowsCWDTracker) detectWindowsShellType(pty *PTY) string {
	if pty == nil || pty.cmd == nil {
		return "unknown"
	}

	shellPath := pty.cmd.Path
	shellName := strings.ToLower(filepath.Base(shellPath))

	switch {
	case strings.Contains(shellName, "cmd.exe"):
		return "cmd"
	case strings.Contains(shellName, "powershell.exe"):
		return "powershell"
	case strings.Contains(shellName, "pwsh.exe"):
		return "pwsh"
	case strings.Contains(shellName, "bash.exe"):
		return "bash"
	case strings.Contains(shellName, "wsl.exe"):
		return "wsl"
	case strings.Contains(shellName, "git-bash.exe"):
		return "gitbash"
	default:
		return "unknown"
	}
}

// monitoringLoop runs background monitoring for all sessions
func (t *WindowsCWDTracker) monitoringLoop(ctx context.Context) {
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
func (t *WindowsCWDTracker) performPeriodicChecks() {
	t.mutex.RLock()
	sessions := make([]*WindowsCWDSession, 0, len(t.sessions))
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

		// Check if the process is still running
		if !t.isProcessRunning(session.PID) {
			t.logger.Info("Process no longer running, removing session", map[string]interface{}{
				"session_id": session.SessionID,
				"pid":        session.PID,
			})
			t.RemoveSession(session.SessionID)
			continue
		}

		// Get current CWD
		cwd, err := t.getCurrentDirectoryFromProcess(session.PID)
		if err != nil {
			t.logger.Debug("Failed to get CWD during periodic check", err, map[string]interface{}{
				"session_id": session.SessionID,
				"pid":        session.PID,
			})
			continue
		}

		// Update CWD if it changed
		session.mutex.Lock()
		if cwd != "" && cwd != session.CurrentCWD {
			session.PreviousCWD = session.CurrentCWD
			session.CurrentCWD = cwd
			session.LastUpdate = time.Now()

			t.publishDirectoryChanged(session.SessionID, cwd, session.PreviousCWD, "periodic_check")
		}
		session.mutex.Unlock()
	}
}

// isProcessRunning checks if a process with the given PID is still running
func (t *WindowsCWDTracker) isProcessRunning(pid int) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), fmt.Sprintf("%d", pid))
}

// publishDirectoryChanged publishes a directory changed event
func (t *WindowsCWDTracker) publishDirectoryChanged(sessionID, newCWD, previousCWD, method string) {
	if t.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "directory_changed",
		CWD:       newCWD,
		Timestamp: time.Now(),
		Command:   method,
	}

	err := t.eventBus.Publish(context.Background(), "terminal.directory.changed", data, "windows-cwd-tracker")
	if err != nil {
		t.logger.Error("Failed to publish directory changed event", err, map[string]interface{}{
			"session_id":   sessionID,
			"new_cwd":      newCWD,
			"previous_cwd": previousCWD,
			"method":       method,
		})
	}
}

// GetStats returns tracking statistics
func (t *WindowsCWDTracker) GetStats() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	stats := map[string]interface{}{
		"enabled":         t.enabled,
		"total_sessions":  len(t.sessions),
		"active_sessions": 0,
		"platform":        runtime.GOOS,
		"method":          "windows_detached",
		"helper_files": map[string]interface{}{
			"powershell_script": t.commandFile,
			"batch_script":      t.batchFile,
		},
	}

	for _, session := range t.sessions {
		if session.IsTracking {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1
		}
	}

	return stats
}

// ProcessOutput processes terminal output for Windows sessions
func (t *WindowsCWDTracker) ProcessOutput(sessionID string, data []byte) {
	// For Windows, we primarily rely on process monitoring rather than output parsing
	// but we can still look for certain commands that might indicate directory changes

	if runtime.GOOS != "windows" || !t.enabled {
		return
	}

	t.mutex.RLock()
	session, exists := t.sessions[sessionID]
	t.mutex.RUnlock()

	if !exists || !session.IsTracking {
		return
	}

	// Look for cd commands in the output
	output := string(data)

	// Common Windows cd command patterns
	cdPatterns := []string{
		"cd /d ",
		"cd ",
		"pushd ",
		"popd ",
	}

	for _, pattern := range cdPatterns {
		if strings.Contains(strings.ToLower(output), pattern) {
			// Trigger a CWD check after a short delay
			go func() {
				time.Sleep(500 * time.Millisecond)
				if cwd, err := t.getCurrentDirectoryFromProcess(session.PID); err == nil {
					session.mutex.Lock()
					if cwd != "" && cwd != session.CurrentCWD {
						session.PreviousCWD = session.CurrentCWD
						session.CurrentCWD = cwd
						session.LastUpdate = time.Now()

						t.publishDirectoryChanged(sessionID, cwd, session.PreviousCWD, "output_trigger")
					}
					session.mutex.Unlock()
				}
			}()
			break
		}
	}
}

// Cleanup helper files
func (t *WindowsCWDTracker) cleanup() {
	if t.commandFile != "" {
		os.Remove(t.commandFile)
	}
	if t.batchFile != "" {
		os.Remove(t.batchFile)
	}
}

// Shutdown shuts down the Windows CWD tracker
func (t *WindowsCWDTracker) Shutdown(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	t.logger.Info("Shutting down Windows CWD tracker")

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Stop tracking all sessions and terminate monitor processes
	for sessionID, session := range t.sessions {
		session.IsTracking = false

		if session.MonitorProcess != nil {
			if err := session.MonitorProcess.Kill(); err != nil {
				t.logger.Warn("Failed to terminate monitor process", err, map[string]interface{}{
					"session_id":  sessionID,
					"monitor_pid": session.MonitorProcess.Pid,
				})
			}
		}

		t.logger.Debug("Stopped tracking session", map[string]interface{}{
			"session_id": sessionID,
			"final_cwd":  session.CurrentCWD,
		})
	}

	// Clear sessions
	t.sessions = make(map[string]*WindowsCWDSession)

	// Clean up helper files
	t.cleanup()

	t.logger.Info("Windows CWD tracker shutdown complete")
	return nil
}
