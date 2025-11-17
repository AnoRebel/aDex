package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// Service manages terminal sessions and operations
type Service struct {
	ptyManager   *PTYManager
	sessions     map[string]*models.TerminalSession
	sessionOrder []string
	eventBus     *events.EventBus
	logger       *logger.Logger
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	outputBuffer map[string][]byte
	history      map[string]*models.TerminalHistory
	profiles     map[string]*models.TerminalProfile
	config       *ServiceConfig
	cwdTracker   *CWDTracker
}

// ServiceConfig holds configuration for the terminal service
type ServiceConfig struct {
	MaxSessions        int           `json:"max_sessions"`
	DefaultShell       string        `json:"default_shell"`
	DefaultWorkingDir  string        `json:"default_working_dir"`
	HistorySize        int           `json:"history_size"`
	BufferSize         int           `json:"buffer_size"`
	SessionTimeout     time.Duration `json:"session_timeout"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	EnableBell         bool          `json:"enable_bell"`
	EnableNotifications bool         `json:"enable_notifications"`
	Theme              string        `json:"theme"`
	Font               string        `json:"font"`
	CWDTracking        *CWDTrackingConfig `json:"cwd_tracking,omitempty"`
}

// DefaultServiceConfig returns the default terminal service configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxSessions:        10,
		DefaultShell:       getDefaultShell(),
		DefaultWorkingDir:  "/tmp",
		HistorySize:        1000,
		BufferSize:         8192,
		SessionTimeout:     30 * time.Minute,
		CleanupInterval:    5 * time.Minute,
		EnableBell:         true,
		EnableNotifications: true,
		Theme:              "default-dark",
		Font:               "JetBrains Mono",
		CWDTracking:        DefaultCWDTrackingConfig(),
	}
}

// NewService creates a new terminal service
func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())

	config := DefaultServiceConfig()

	service := &Service{
		ptyManager:   NewPTYManager(),
		sessions:     make(map[string]*models.TerminalSession),
		sessionOrder: make([]string, 0),
		outputBuffer: make(map[string][]byte),
		history:      make(map[string]*models.TerminalHistory),
		profiles:     make(map[string]*models.TerminalProfile),
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		cwdTracker:   NewCWDTracker(config.CWDTracking),
	}

	// Initialize default profiles
	service.initializeDefaultProfiles()

	return service
}

// Initialize initializes the terminal service
func (s *Service) Initialize(ctx context.Context) error {
	s.ctx = ctx
	s.logger = logger.GetDefaultLogger()
	s.logger.Info("Initializing terminal service")

	// Initialize event bus
	if s.eventBus == nil {
		s.eventBus = events.GetEventBus()
	}

	// Initialize CWD tracker
	if s.cwdTracker != nil {
		s.cwdTracker.SetEventBus(s.eventBus)
		if err := s.cwdTracker.Initialize(ctx); err != nil {
			s.logger.Warn("Failed to initialize CWD tracker", err, nil)
		} else {
			s.logger.Info("CWD tracker initialized successfully")
		}
	}

	// Start cleanup goroutine
	go s.cleanupRoutine()

	s.logger.Info("Terminal service initialized successfully")
	return nil
}

// SetEventBus sets the event bus for the service
func (s *Service) SetEventBus(bus events.IEventBus) {
	s.eventBus = bus
	if s.cwdTracker != nil {
		s.cwdTracker.SetEventBus(bus)
	}
}

// GetEventBus returns the event bus
func (s *Service) GetEventBus() events.IEventBus {
	return s.eventBus
}

// CreateSession creates a new terminal session
func (s *Service) CreateSession(ctx context.Context, options *models.TerminalOptions) (*models.TerminalSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check session limit
	if len(s.sessions) >= s.config.MaxSessions {
		return nil, models.NewError(models.ErrTerminalCreate, "Maximum number of sessions reached", nil)
	}

	// Create session model
	session := models.NewTerminalSession(options)

	// Set defaults from config
	if session.Shell == "" {
		session.Shell = s.config.DefaultShell
	}
	if session.CWD == "" {
		session.CWD = s.config.DefaultWorkingDir
	}

	// Validate working directory
	if _, err := os.Stat(session.CWD); os.IsNotExist(err) {
		return nil, models.NewError(models.ErrTerminalCreate, fmt.Sprintf("Working directory does not exist: %s", session.CWD), err)
	}

	// Create PTY
	pty, err := s.ptyManager.CreatePTY(ctx, session.ID, options)
	if err != nil {
		return nil, models.NewError(models.ErrTerminalCreate, "Failed to create PTY", err)
	}

	// Set session PID if available
	if process := pty.GetProcess(); process != nil {
		session.PID = process.Pid
	}

	// Initialize session history
	s.history[session.ID] = &models.TerminalHistory{
		SessionID: session.ID,
		Commands:  make([]models.TerminalCommand, 0),
		MaxSize:   s.config.HistorySize,
		Current:   0,
	}

	// Initialize output buffer
	s.outputBuffer[session.ID] = make([]byte, 0, s.config.BufferSize)

	// Add session to collection
	s.sessions[session.ID] = session
	s.sessionOrder = append(s.sessionOrder, session.ID)

	// Add session to CWD tracker
	if s.cwdTracker != nil {
		if err := s.cwdTracker.AddSession(session.ID, pty, session.CWD); err != nil {
			s.logger.Warn("Failed to add session to CWD tracker", err, map[string]interface{}{
				"session_id": session.ID,
				"cwd":        session.CWD,
			})
		}
	}

	// Start output reader goroutine
	go s.readOutput(session.ID, pty)

	// Publish session created event
	s.publishEvent(models.TerminalCreated, map[string]interface{}{
		"session_id": session.ID,
		"shell":      session.Shell,
		"cwd":        session.CWD,
		"size":       session.Size,
		"created_at": session.CreatedAt,
	})

	s.logger.Info("Terminal session created", map[string]interface{}{
		"session_id": session.ID,
		"shell":      session.Shell,
		"cwd":        session.CWD,
	})

	return session, nil
}

// WriteToSession writes data to a terminal session
func (s *Service) WriteToSession(ctx context.Context, sessionID string, data []byte) error {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return models.NewError(models.ErrTerminalWrite, "Session not found", nil)
	}

	if !session.Active {
		return models.NewError(models.ErrTerminalWrite, "Session is not active", nil)
	}

	// Get PTY
	pty := s.ptyManager.GetPTY(sessionID)
	if pty == nil {
		return models.NewError(models.ErrTerminalWrite, "PTY not found for session", nil)
	}

	// Write data to PTY
	if err := pty.Write(data); err != nil {
		session.Active = false
		return models.NewError(models.ErrTerminalWrite, "Failed to write to terminal", err)
	}

	// Handle special characters
	s.handleSpecialCharacters(sessionID, data)

	// Update session activity
	session.LastSeen = time.Now()
	session.UpdatedAt = time.Now()

	// Publish input event
	s.publishEvent(models.TerminalInput, map[string]interface{}{
		"session_id": sessionID,
		"data":       string(data),
		"timestamp":  time.Now(),
	})

	return nil
}

// ResizeSession resizes a terminal session
func (s *Service) ResizeSession(ctx context.Context, sessionID string, rows, cols uint16) error {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return models.NewError(models.ErrTerminalResize, "Session not found", nil)
	}

	if !session.Active {
		return models.NewError(models.ErrTerminalResize, "Session is not active", nil)
	}

	// Get PTY
	pty := s.ptyManager.GetPTY(sessionID)
	if pty == nil {
		return models.NewError(models.ErrTerminalResize, "PTY not found for session", nil)
	}

	// Resize PTY
	if err := pty.SetSize(rows, cols); err != nil {
		return models.NewError(models.ErrTerminalResize, "Failed to resize terminal", err)
	}

	// Update session size
	session.Size = &models.TerminalSize{
		Rows: rows,
		Cols: cols,
	}
	session.UpdatedAt = time.Now()

	// Publish resize event
	s.publishEvent(models.TerminalResized, map[string]interface{}{
		"session_id": sessionID,
		"rows":       rows,
		"cols":       cols,
		"timestamp":  time.Now(),
	})

	s.logger.Info("Terminal session resized", map[string]interface{}{
		"session_id": sessionID,
		"rows":       rows,
		"cols":       cols,
	})

	return nil
}

// CloseSession closes a terminal session
func (s *Service) CloseSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.NewError(models.ErrTerminalCreate, "Session not found", nil)
	}

	// Mark session as inactive
	session.Active = false
	session.UpdatedAt = time.Now()

	// Close PTY
	s.ptyManager.RemovePTY(sessionID)

	// Remove from CWD tracker
	if s.cwdTracker != nil {
		s.cwdTracker.RemoveSession(sessionID)
	}

	// Remove from session order
	for i, id := range s.sessionOrder {
		if id == sessionID {
			s.sessionOrder = append(s.sessionOrder[:i], s.sessionOrder[i+1:]...)
			break
		}
	}

	// Clean up resources
	delete(s.sessions, sessionID)
	delete(s.outputBuffer, sessionID)
	delete(s.history, sessionID)

	// Publish close event
	s.publishEvent(models.TerminalClosed, map[string]interface{}{
		"session_id": sessionID,
		"timestamp":  time.Now(),
	})

	s.logger.Info("Terminal session closed", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

// GetSessions returns all active sessions
func (s *Service) GetSessions() []*models.TerminalSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.TerminalSession, 0, len(s.sessions))
	for _, sessionID := range s.sessionOrder {
		if session := s.sessions[sessionID]; session != nil {
			// Check if session is still active
			pty := s.ptyManager.GetPTY(sessionID)
			if pty != nil && pty.IsActive() {
				sessions = append(sessions, session)
			} else {
				// Clean up inactive session
				session.Active = false
				session.UpdatedAt = time.Now()
			}
		}
	}

	return sessions
}

// GetSession returns a specific session by ID
func (s *Service) GetSession(sessionID string) *models.TerminalSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessions[sessionID]
}

// GetSessionHistory returns the command history for a session
func (s *Service) GetSessionHistory(sessionID string) *models.TerminalHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.history[sessionID]
}

// SubscribeToOutput subscribes to terminal output for a session
func (s *Service) SubscribeToOutput(sessionID string, callback func([]byte)) {
	// This would be implemented using the event bus
	// For now, it's a placeholder
}

// readOutput reads output from a PTY and processes it
func (s *Service) readOutput(sessionID string, pty *PTY) {
	buffer := make([]byte, 1024)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			if !pty.IsActive() {
				return
			}

			n, err := pty.Read(buffer)
			if err != nil {
				if err == io.EOF {
					// Session ended
					s.mu.Lock()
					if session := s.sessions[sessionID]; session != nil {
						session.Active = false
						session.UpdatedAt = time.Now()
					}
					s.mu.Unlock()

					s.publishEvent(models.TerminalClosed, map[string]interface{}{
						"session_id": sessionID,
						"reason":     "process_exited",
						"timestamp":  time.Now(),
					})
					return
				}
				continue
			}

			if n > 0 {
				data := make([]byte, n)
				copy(data, buffer[:n])

				// Process output
				s.processOutput(sessionID, data)
			}
		}
	}
}

// processOutput processes terminal output data
func (s *Service) processOutput(sessionID string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Send output to CWD tracker for processing
	if s.cwdTracker != nil {
		s.cwdTracker.ProcessOutput(sessionID, data)
	}

	// Add to buffer
	if buffer, exists := s.outputBuffer[sessionID]; exists {
		buffer = append(buffer, data...)

		// Prevent buffer from growing too large
		if len(buffer) > s.config.BufferSize {
			// Keep only the latest data
			buffer = buffer[len(buffer)-s.config.BufferSize:]
		}

		s.outputBuffer[sessionID] = buffer
	}

	// Check for bell character
	if s.config.EnableBell {
		for _, b := range data {
			if b == 7 { // ASCII BEL
				s.handleBell(sessionID)
				break
			}
		}
	}

	// Publish output event
	s.publishEvent(models.TerminalOutput, map[string]interface{}{
		"session_id": sessionID,
		"data":       string(data),
		"timestamp":  time.Now(),
	})
}

// handleBell handles terminal bell notifications
func (s *Service) handleBell(sessionID string) {
	if s.config.EnableNotifications {
		// Publish bell event
		s.publishEvent("terminal.bell", map[string]interface{}{
			"session_id": sessionID,
			"timestamp":  time.Now(),
		})

		s.logger.Debug("Terminal bell", map[string]interface{}{
			"session_id": sessionID,
		})
	}
}

// handleSpecialCharacters handles special terminal characters
func (s *Service) handleSpecialCharacters(sessionID string, data []byte) {
	// Handle Ctrl+C (ASCII 3)
	for _, b := range data {
		if b == 3 {
			// This could be used to interrupt running commands
			break
		}
	}
}

// publishEvent publishes an event to the event bus
func (s *Service) publishEvent(eventType string, data map[string]interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(s.ctx, eventType, data, "terminal-service")
		if err != nil {
			s.logger.Error("Failed to publish event", err, map[string]interface{}{
				"event_type": eventType,
				"data":       data,
			})
		}
	}
}

// cleanupRoutine periodically cleans up inactive sessions
func (s *Service) cleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions removes inactive sessions
func (s *Service) cleanupInactiveSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	inactiveSessions := make([]string, 0)

	for sessionID, session := range s.sessions {
		if !session.Active || now.Sub(session.LastSeen) > s.config.SessionTimeout {
			inactiveSessions = append(inactiveSessions, sessionID)
		}
	}

	for _, sessionID := range inactiveSessions {
		s.logger.Info("Cleaning up inactive session", map[string]interface{}{
			"session_id": sessionID,
		})

		s.ptyManager.RemovePTY(sessionID)
		delete(s.sessions, sessionID)
		delete(s.outputBuffer, sessionID)
		delete(s.history, sessionID)

		// Remove from session order
		for i, id := range s.sessionOrder {
			if id == sessionID {
				s.sessionOrder = append(s.sessionOrder[:i], s.sessionOrder[i+1:]...)
				break
			}
		}
	}
}

// initializeDefaultProfiles initializes default terminal profiles
func (s *Service) initializeDefaultProfiles() {
	profiles := []*models.TerminalProfile{
		{
			ID:          "default",
			Name:        "Default",
			Description: "Default terminal profile",
			Shell:       s.config.DefaultShell,
			WorkingDir:  s.config.DefaultWorkingDir,
			IsDefault:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "bash",
			Name:        "Bash",
			Description: "Bash shell profile",
			Shell:       "/bin/bash",
			WorkingDir:  "/tmp",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "zsh",
			Name:        "Zsh",
			Description: "Zsh shell profile",
			Shell:       "/bin/zsh",
			WorkingDir:  "/tmp",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, profile := range profiles {
		s.profiles[profile.ID] = profile
	}
}

// GetProfiles returns all available terminal profiles
func (s *Service) GetProfiles() map[string]*models.TerminalProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*models.TerminalProfile)
	for id, profile := range s.profiles {
		result[id] = profile
	}
	return result
}

// GetProfile returns a specific profile by ID
func (s *Service) GetProfile(profileID string) *models.TerminalProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.profiles[profileID]
}

// CreateProfile creates a new terminal profile
func (s *Service) CreateProfile(profile *models.TerminalProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.profiles[profile.ID]; exists {
		return models.NewError(models.ErrConfigValidate, "Profile already exists", nil)
	}

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	s.profiles[profile.ID] = profile

	return nil
}

// UpdateProfile updates an existing terminal profile
func (s *Service) UpdateProfile(profile *models.TerminalProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.profiles[profile.ID]; !exists {
		return models.NewError(models.ErrConfigValidate, "Profile not found", nil)
	}

	profile.UpdatedAt = time.Now()
	s.profiles[profile.ID] = profile

	return nil
}

// DeleteProfile deletes a terminal profile
func (s *Service) DeleteProfile(profileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if profileID == "default" {
		return models.NewError(models.ErrConfigValidate, "Cannot delete default profile", nil)
	}

	delete(s.profiles, profileID)
	return nil
}

// GetCurrentCWD returns the current working directory for a terminal session
func (s *Service) GetCurrentCWD(sessionID string) (string, error) {
	if s.cwdTracker == nil {
		// Fallback to session data if CWD tracker is not available
		s.mu.RLock()
		session, exists := s.sessions[sessionID]
		s.mu.RUnlock()

		if !exists {
			return "", models.NewError(models.ErrTerminalNotFound, "Session not found", nil)
		}

		return session.CWD, nil
	}

	// Use CWD tracker for accurate current directory
	return s.cwdTracker.GetCurrentCWD(sessionID)
}

// GetCWDTrackerStats returns CWD tracking statistics
func (s *Service) GetCWDTrackerStats() map[string]interface{} {
	if s.cwdTracker == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "CWD tracker not initialized",
		}
	}

	return s.cwdTracker.GetStats()
}

// Shutdown shuts down the terminal service
func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down terminal service")

	// Cancel context
	s.cancel()

	// Close all sessions
	for sessionID := range s.sessions {
		s.CloseSession(ctx, sessionID)
	}

	// Cleanup PTY manager
	s.ptyManager.Cleanup()

	// Shutdown CWD tracker
	if s.cwdTracker != nil {
		if err := s.cwdTracker.Shutdown(ctx); err != nil {
			s.logger.Warn("Failed to shutdown CWD tracker", err, nil)
		}
	}

	s.logger.Info("Terminal service shutdown complete")
	return nil
}

// Helper function to get default shell
func getDefaultShell() string {
	// Try common shells in order of preference
	shells := []string{"/bin/bash", "/bin/zsh", "/bin/sh", "/usr/bin/bash"}

	for _, shell := range shells {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}

	// Fallback
	return "/bin/sh"
}