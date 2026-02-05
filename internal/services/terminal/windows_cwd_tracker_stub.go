//go:build !windows
// +build !windows

package terminal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
)

// WindowsCWDTracker provides CWD tracking for Windows systems using detached processes
// This is a stub implementation for non-Windows platforms
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
	SessionID   string
	PID         int
	CurrentCWD  string
	PreviousCWD string
	LastUpdate  time.Time
	IsTracking  bool
	ShellType   string
	mutex       sync.RWMutex
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
		Enabled:           false,
		PollInterval:      1 * time.Second,
		CreateHelperFiles: false,
		MaxRetries:        3,
		Timeout:           30 * time.Second,
	}
}

// NewWindowsCWDTracker creates a new Windows CWD tracker (stub for non-Windows)
func NewWindowsCWDTracker(config *WindowsCWDConfig) *WindowsCWDTracker {
	return &WindowsCWDTracker{
		enabled:  false,
		sessions: make(map[string]*WindowsCWDSession),
	}
}

// Initialize initializes the Windows CWD tracker (stub for non-Windows)
func (t *WindowsCWDTracker) Initialize(ctx context.Context) error {
	return fmt.Errorf("Windows CWD tracker is only available on Windows")
}

// SetEventBus sets the event bus
func (t *WindowsCWDTracker) SetEventBus(eventBus events.IEventBus) {
	t.eventBus = eventBus
}

// AddSession adds a terminal session for Windows CWD tracking (stub for non-Windows)
func (t *WindowsCWDTracker) AddSession(sessionID string, pty *PTY, initialCWD string) error {
	return fmt.Errorf("Windows CWD tracking is only available on Windows")
}

// RemoveSession removes a terminal session from Windows CWD tracking (stub for non-Windows)
func (t *WindowsCWDTracker) RemoveSession(sessionID string) {
	// No-op on non-Windows
}

// GetCurrentCWD returns the current working directory for a session (stub for non-Windows)
func (t *WindowsCWDTracker) GetCurrentCWD(sessionID string) (string, error) {
	return "", fmt.Errorf("Windows CWD tracker is only available on Windows")
}

// GetStats returns tracking statistics
func (t *WindowsCWDTracker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         false,
		"total_sessions":  0,
		"active_sessions": 0,
		"platform":        "non-windows",
		"method":          "stub",
	}
}

// ProcessOutput processes terminal output for Windows sessions (stub for non-Windows)
func (t *WindowsCWDTracker) ProcessOutput(sessionID string, data []byte) {
	// No-op on non-Windows
}

// Shutdown shuts down the Windows CWD tracker (stub for non-Windows)
func (t *WindowsCWDTracker) Shutdown(ctx context.Context) error {
	return nil
}
