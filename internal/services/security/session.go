package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"aDex-UI/internal/logger"
)

// Session represents an authenticated user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"userId"`
	Username  string                 `json:"username"`
	Role      UserRole               `json:"role"`
	CreatedAt time.Time              `json:"createdAt"`
	ExpiresAt time.Time              `json:"expiresAt"`
	LastSeen  time.Time              `json:"lastSeen"`
	IPAddress string                 `json:"ipAddress"`
	UserAgent string                 `json:"userAgent"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SessionManager manages user sessions
type SessionManager struct {
	sessions      map[string]*Session // indexed by session ID
	byUserID      map[string][]*Session // sessions by user ID
	config        *SessionConfig
	mu            sync.RWMutex
	logger        *logger.Logger
}

// SessionConfig holds session configuration
type SessionConfig struct {
	SessionDuration    time.Duration `json:"sessionDuration"`    // How long sessions last
	IdleTimeout        time.Duration `json:"idleTimeout"`        // Inactivity timeout
	MaxSessionsPerUser int           `json:"maxSessionsPerUser"` // Max concurrent sessions
	EnableRenewal      bool          `json:"enableRenewal"`      // Allow session renewal
	SecureOnly         bool          `json:"secureOnly"`         // Require HTTPS
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger *logger.Logger) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		byUserID: make(map[string][]*Session),
		logger:   logger,
		config:   getDefaultSessionConfig(),
	}

	// Start background cleanup goroutine
	go sm.cleanupExpiredSessions()

	return sm
}

// getDefaultSessionConfig returns default session configuration
func getDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		SessionDuration:    24 * time.Hour, // 24 hours
		IdleTimeout:        2 * time.Hour,  // 2 hours of inactivity
		MaxSessionsPerUser: 5,              // Max 5 concurrent sessions
		EnableRenewal:      true,
		SecureOnly:         true,
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(user *User, ipAddress, userAgent string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check max sessions per user
	if sm.config.MaxSessionsPerUser > 0 {
		userSessions := sm.byUserID[user.ID]
		if len(userSessions) >= sm.config.MaxSessionsPerUser {
			// Remove oldest session
			sm.removeOldestSession(user.ID)
		}
	}

	// Generate secure session ID
	sessionID, err := sm.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.config.SessionDuration),
		LastSeen:  now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  make(map[string]interface{}),
	}

	// Store session
	sm.sessions[sessionID] = session
	sm.byUserID[user.ID] = append(sm.byUserID[user.ID], session)

	sm.logger.Info("Session created", "session_id", sessionID, "user_id", user.ID)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// ValidateSession validates a session and checks if it's expired
func (sm *SessionManager) ValidateSession(sessionID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	now := time.Now()

	// Check if session has expired
	if now.After(session.ExpiresAt) {
		sm.removeSessionLocked(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	// Check idle timeout
	if now.Sub(session.LastSeen) > sm.config.IdleTimeout {
		sm.removeSessionLocked(sessionID)
		return nil, fmt.Errorf("session idle timeout")
	}

	// Update last seen time
	session.LastSeen = now

	// Renew session if enabled and close to expiration
	if sm.config.EnableRenewal {
		timeUntilExpiry := session.ExpiresAt.Sub(now)
		if timeUntilExpiry < sm.config.SessionDuration/4 {
			// Renew for another full duration
			session.ExpiresAt = now.Add(sm.config.SessionDuration)
			sm.logger.Debug("Session renewed", "session_id", sessionID)
		}
	}

	return session, nil
}

// RefreshSession updates the last seen time for a session
func (sm *SessionManager) RefreshSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.LastSeen = time.Now()
	return nil
}

// DestroySession destroys a session (logout)
func (sm *SessionManager) DestroySession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	sm.logger.Info("Session destroyed", "session_id", sessionID, "user_id", session.UserID)
	sm.removeSessionLocked(sessionID)
	return nil
}

// DestroyUserSessions destroys all sessions for a user
func (sm *SessionManager) DestroyUserSessions(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessions := sm.byUserID[userID]
	for _, session := range sessions {
		sm.removeSessionLocked(session.ID)
	}

	sm.logger.Info("All user sessions destroyed", "user_id", userID, "count", len(sessions))
	return nil
}

// GetUserSessions returns all active sessions for a user
func (sm *SessionManager) GetUserSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := sm.byUserID[userID]
	result := make([]*Session, len(sessions))
	copy(result, sessions)
	return result
}

// GetActiveSessions returns count of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}

// SetSessionMetadata sets metadata for a session
func (sm *SessionManager) SetSessionMetadata(sessionID, key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	if session.Metadata == nil {
		session.Metadata = make(map[string]interface{})
	}

	session.Metadata[key] = value
	return nil
}

// GetSessionMetadata gets metadata from a session
func (sm *SessionManager) GetSessionMetadata(sessionID, key string) (interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	value, exists := session.Metadata[key]
	if !exists {
		return nil, fmt.Errorf("metadata key not found: %s", key)
	}

	return value, nil
}

// Private methods

// generateSessionID generates a cryptographically secure session ID
func (sm *SessionManager) generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// removeSessionLocked removes a session (must be called with lock held)
func (sm *SessionManager) removeSessionLocked(sessionID string) {
	session, exists := sm.sessions[sessionID]
	if !exists {
		return
	}

	// Remove from main map
	delete(sm.sessions, sessionID)

	// Remove from user's session list
	userSessions := sm.byUserID[session.UserID]
	for i, s := range userSessions {
		if s.ID == sessionID {
			sm.byUserID[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
			break
		}
	}

	// Clean up empty user session list
	if len(sm.byUserID[session.UserID]) == 0 {
		delete(sm.byUserID, session.UserID)
	}
}

// removeOldestSession removes the oldest session for a user
func (sm *SessionManager) removeOldestSession(userID string) {
	sessions := sm.byUserID[userID]
	if len(sessions) == 0 {
		return
	}

	// Find oldest session
	oldestIdx := 0
	oldestTime := sessions[0].CreatedAt

	for i, session := range sessions {
		if session.CreatedAt.Before(oldestTime) {
			oldestIdx = i
			oldestTime = session.CreatedAt
		}
	}

	sm.logger.Info("Removing oldest session due to max sessions limit",
		"user_id", userID,
		"session_id", sessions[oldestIdx].ID)

	sm.removeSessionLocked(sessions[oldestIdx].ID)
}

// cleanupExpiredSessions runs periodically to remove expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()

		now := time.Now()
		expiredCount := 0

		// Find and remove expired sessions
		for sessionID, session := range sm.sessions {
			if now.After(session.ExpiresAt) || now.Sub(session.LastSeen) > sm.config.IdleTimeout {
				sm.removeSessionLocked(sessionID)
				expiredCount++
			}
		}

		sm.mu.Unlock()

		if expiredCount > 0 {
			sm.logger.Debug("Cleaned up expired sessions", "count", expiredCount)
		}
	}
}

// UpdateConfig updates the session configuration
func (sm *SessionManager) UpdateConfig(config *SessionConfig) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.config = config
	sm.logger.Info("Session configuration updated")
}

// GetConfig returns the current session configuration
func (sm *SessionManager) GetConfig() *SessionConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.config
}
