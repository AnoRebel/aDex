package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"aDex-UI/internal/logger"
	"aDex-UI/internal/services/security"
)

func TestSessionManagement(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	testUser := &security.User{
		ID:       "user-session-001",
		Username: "sessionuser",
		Role:     security.RoleUser,
	}

	t.Run("Create session", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.100", "Mozilla/5.0")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, testUser.ID, session.UserID)
		assert.Equal(t, testUser.Username, session.Username)
		assert.Equal(t, "192.168.1.100", session.IPAddress)
		assert.NotZero(t, session.CreatedAt)
		assert.NotZero(t, session.ExpiresAt)
	})

	t.Run("Retrieve session", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.101", "Chrome")
		require.NoError(t, err)

		retrieved, err := sessionManager.GetSession(session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.UserID, retrieved.UserID)
	})

	t.Run("Validate active session", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.102", "Safari")
		require.NoError(t, err)

		validated, err := sessionManager.ValidateSession(session.ID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, validated.ID)
	})

	t.Run("Validate non-existent session", func(t *testing.T) {
		_, err := sessionManager.ValidateSession("non-existent-session-id")
		assert.Error(t, err)
	})
}

func TestSessionExpiration(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	// Set very short session duration for testing
	config := sessionManager.GetConfig()
	config.SessionDuration = 100 * time.Millisecond
	config.IdleTimeout = 50 * time.Millisecond
	sessionManager.UpdateConfig(config)

	testUser := &security.User{
		ID:       "user-expire-001",
		Username: "expireuser",
		Role:     security.RoleUser,
	}

	t.Run("Session expires after duration", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.200", "Test")
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should fail validation
		_, err = sessionManager.ValidateSession(session.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Session expires due to inactivity", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.201", "Test")
		require.NoError(t, err)

		// Wait for idle timeout
		time.Sleep(70 * time.Millisecond)

		// Should fail validation
		_, err = sessionManager.ValidateSession(session.ID)
		assert.Error(t, err)
	})
}

func TestSessionRefresh(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	testUser := &security.User{
		ID:       "user-refresh-001",
		Username: "refreshuser",
		Role:     security.RoleUser,
	}

	t.Run("Refresh session updates last seen", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.300", "Test")
		require.NoError(t, err)

		originalLastSeen := session.LastSeen

		time.Sleep(10 * time.Millisecond)

		err = sessionManager.RefreshSession(session.ID)
		require.NoError(t, err)

		// Get updated session
		updated, err := sessionManager.GetSession(session.ID)
		require.NoError(t, err)

		assert.True(t, updated.LastSeen.After(originalLastSeen))
	})
}

func TestSessionDestruction(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	testUser := &security.User{
		ID:       "user-destroy-001",
		Username: "destroyuser",
		Role:     security.RoleUser,
	}

	t.Run("Destroy single session", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.400", "Test")
		require.NoError(t, err)

		err = sessionManager.DestroySession(session.ID)
		require.NoError(t, err)

		// Session should no longer exist
		_, err = sessionManager.GetSession(session.ID)
		assert.Error(t, err)
	})

	t.Run("Destroy all user sessions", func(t *testing.T) {
		// Create multiple sessions
		for i := 0; i < 3; i++ {
			_, err := sessionManager.CreateSession(testUser, "192.168.1.500", "Test")
			require.NoError(t, err)
		}

		sessions := sessionManager.GetUserSessions(testUser.ID)
		assert.GreaterOrEqual(t, len(sessions), 3)

		err := sessionManager.DestroyUserSessions(testUser.ID)
		require.NoError(t, err)

		sessions = sessionManager.GetUserSessions(testUser.ID)
		assert.Len(t, sessions, 0)
	})
}

func TestMaxSessionsPerUser(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	// Set max sessions to 3
	config := sessionManager.GetConfig()
	config.MaxSessionsPerUser = 3
	sessionManager.UpdateConfig(config)

	testUser := &security.User{
		ID:       "user-max-001",
		Username: "maxuser",
		Role:     security.RoleUser,
	}

	t.Run("Limit concurrent sessions per user", func(t *testing.T) {
		// Create 5 sessions
		var sessions []*security.Session
		for i := 0; i < 5; i++ {
			session, err := sessionManager.CreateSession(testUser, "192.168.1.600", "Test")
			require.NoError(t, err)
			sessions = append(sessions, session)
		}

		// Should only have 3 active sessions
		activeSessions := sessionManager.GetUserSessions(testUser.ID)
		assert.Len(t, activeSessions, 3)

		// First sessions should be removed (oldest first)
		_, err := sessionManager.GetSession(sessions[0].ID)
		assert.Error(t, err)

		_, err = sessionManager.GetSession(sessions[1].ID)
		assert.Error(t, err)

		// Last 3 should still exist
		_, err = sessionManager.GetSession(sessions[2].ID)
		assert.NoError(t, err)
	})
}

func TestSessionMetadata(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	testUser := &security.User{
		ID:       "user-meta-001",
		Username: "metauser",
		Role:     security.RoleUser,
	}

	t.Run("Set and get session metadata", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.700", "Test")
		require.NoError(t, err)

		// Set metadata
		err = sessionManager.SetSessionMetadata(session.ID, "csrf_token", "test-csrf-token-123")
		require.NoError(t, err)

		err = sessionManager.SetSessionMetadata(session.ID, "custom_field", "custom_value")
		require.NoError(t, err)

		// Get metadata
		csrfToken, err := sessionManager.GetSessionMetadata(session.ID, "csrf_token")
		require.NoError(t, err)
		assert.Equal(t, "test-csrf-token-123", csrfToken)

		customField, err := sessionManager.GetSessionMetadata(session.ID, "custom_field")
		require.NoError(t, err)
		assert.Equal(t, "custom_value", customField)
	})

	t.Run("Get non-existent metadata key", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.701", "Test")
		require.NoError(t, err)

		_, err = sessionManager.GetSessionMetadata(session.ID, "non_existent")
		assert.Error(t, err)
	})
}

func TestSessionRenewal(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	// Set short duration for testing renewal
	config := sessionManager.GetConfig()
	config.SessionDuration = 200 * time.Millisecond
	config.EnableRenewal = true
	sessionManager.UpdateConfig(config)

	testUser := &security.User{
		ID:       "user-renew-001",
		Username: "renewuser",
		Role:     security.RoleUser,
	}

	t.Run("Session is renewed when close to expiration", func(t *testing.T) {
		session, err := sessionManager.CreateSession(testUser, "192.168.1.800", "Test")
		require.NoError(t, err)

		originalExpiry := session.ExpiresAt

		// Wait until close to expiration (75% of duration)
		time.Sleep(160 * time.Millisecond)

		// Validate should renew
		validated, err := sessionManager.ValidateSession(session.ID)
		require.NoError(t, err)

		// Expiry should be extended
		assert.True(t, validated.ExpiresAt.After(originalExpiry))
	})
}

func TestGetActiveSessions(t *testing.T) {
	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	sessionManager := security.NewSessionManager(testLogger)

	testUser := &security.User{
		ID:       "user-active-001",
		Username: "activeuser",
		Role:     security.RoleUser,
	}

	t.Run("Count active sessions", func(t *testing.T) {
		initialCount := sessionManager.GetActiveSessions()

		// Create 3 sessions
		for i := 0; i < 3; i++ {
			_, err := sessionManager.CreateSession(testUser, "192.168.1.900", "Test")
			require.NoError(t, err)
		}

		newCount := sessionManager.GetActiveSessions()
		assert.Equal(t, initialCount+3, newCount)
	})
}
