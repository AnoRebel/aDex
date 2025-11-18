package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"aDex-UI/internal/logger"
	"aDex-UI/internal/services/security"
)

func TestArgon2idPasswordHashing(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	userStore, err := security.NewUserStore(storePath)
	require.NoError(t, err)

	authService := security.NewAuthService(userStore, testLogger)

	t.Run("Hash password with argon2id", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := authService.HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$argon2id$")
		assert.Contains(t, hash, "v=19")
	})

	t.Run("Verify correct password", func(t *testing.T) {
		password := "CorrectPassword123!"
		hash, err := authService.HashPassword(password)
		require.NoError(t, err)

		valid, err := authService.VerifyPassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Reject incorrect password", func(t *testing.T) {
		password := "CorrectPassword123!"
		wrongPassword := "WrongPassword456!"
		hash, err := authService.HashPassword(password)
		require.NoError(t, err)

		valid, err := authService.VerifyPassword(wrongPassword, hash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Different hashes for same password", func(t *testing.T) {
		password := "TestPassword123!"
		hash1, err := authService.HashPassword(password)
		require.NoError(t, err)

		hash2, err := authService.HashPassword(password)
		require.NoError(t, err)

		// Hashes should be different due to random salt
		assert.NotEqual(t, hash1, hash2)

		// But both should verify correctly
		valid1, err := authService.VerifyPassword(password, hash1)
		require.NoError(t, err)
		assert.True(t, valid1)

		valid2, err := authService.VerifyPassword(password, hash2)
		require.NoError(t, err)
		assert.True(t, valid2)
	})

	t.Run("Password strength validation", func(t *testing.T) {
		weakPasswords := []string{
			"weak",           // Too short
			"nouppernumber1", // No uppercase
			"NOLOWERNUMBER1", // No lowercase
			"NoNumbers!",     // No numbers
			"NoSpecial123",   // No special characters
		}

		for _, pwd := range weakPasswords {
			_, err := authService.HashPassword(pwd)
			assert.Error(t, err, "Password '%s' should be rejected", pwd)
		}

		strongPassword := "StrongPassword123!"
		_, err := authService.HashPassword(strongPassword)
		assert.NoError(t, err)
	})
}

func TestUserAuthentication(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	userStore, err := security.NewUserStore(storePath)
	require.NoError(t, err)

	authService := security.NewAuthService(userStore, testLogger)

	// Create a test user
	password := "TestPassword123!"
	hash, err := authService.HashPassword(password)
	require.NoError(t, err)

	testUser := &security.User{
		ID:           "user-001",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		DisplayName:  "Test User",
		Role:         security.RoleUser,
		Permissions:  []string{"read", "write"},
		Enabled:      true,
	}

	err = userStore.CreateUser(testUser)
	require.NoError(t, err)

	t.Run("Successful authentication", func(t *testing.T) {
		user, err := authService.Authenticate("testuser", password)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "user-001", user.ID)
	})

	t.Run("Failed authentication - wrong password", func(t *testing.T) {
		user, err := authService.Authenticate("testuser", "WrongPassword123!")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Failed authentication - user not found", func(t *testing.T) {
		user, err := authService.Authenticate("nonexistent", password)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Failed authentication - disabled user", func(t *testing.T) {
		// Disable the user
		testUser.Enabled = false
		err := userStore.UpdateUser(testUser)
		require.NoError(t, err)

		user, err := authService.Authenticate("testuser", password)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "disabled")

		// Re-enable for other tests
		testUser.Enabled = true
		err = userStore.UpdateUser(testUser)
		require.NoError(t, err)
	})

	t.Run("Update last login on successful auth", func(t *testing.T) {
		user, err := authService.Authenticate("testuser", password)
		require.NoError(t, err)

		// Verify last login was updated
		updatedUser, err := userStore.GetUser(user.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLoginAt)
		assert.Greater(t, updatedUser.LoginCount, 0)
	})
}

func TestPasswordChange(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	userStore, err := security.NewUserStore(storePath)
	require.NoError(t, err)

	authService := security.NewAuthService(userStore, testLogger)

	// Create a test user
	oldPassword := "OldPassword123!"
	hash, err := authService.HashPassword(oldPassword)
	require.NoError(t, err)

	testUser := &security.User{
		ID:           "user-002",
		Username:     "changeuser",
		Email:        "change@example.com",
		PasswordHash: hash,
		DisplayName:  "Change User",
		Role:         security.RoleUser,
		Enabled:      true,
	}

	err = userStore.CreateUser(testUser)
	require.NoError(t, err)

	t.Run("Successful password change", func(t *testing.T) {
		newPassword := "NewPassword456!"
		err := authService.ChangePassword(testUser.ID, oldPassword, newPassword)
		require.NoError(t, err)

		// Verify old password no longer works
		_, err = authService.Authenticate("changeuser", oldPassword)
		assert.Error(t, err)

		// Verify new password works
		user, err := authService.Authenticate("changeuser", newPassword)
		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("Failed password change - wrong old password", func(t *testing.T) {
		err := authService.ChangePassword(testUser.ID, "WrongOldPassword!", "NewPassword789!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect")
	})
}

func TestPasswordReset(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	userStore, err := security.NewUserStore(storePath)
	require.NoError(t, err)

	authService := security.NewAuthService(userStore, testLogger)

	// Create a test user
	password := "InitialPassword123!"
	hash, err := authService.HashPassword(password)
	require.NoError(t, err)

	testUser := &security.User{
		ID:           "user-003",
		Username:     "resetuser",
		Email:        "reset@example.com",
		PasswordHash: hash,
		DisplayName:  "Reset User",
		Role:         security.RoleUser,
		Enabled:      true,
	}

	err = userStore.CreateUser(testUser)
	require.NoError(t, err)

	t.Run("Admin password reset", func(t *testing.T) {
		newPassword := "AdminResetPassword456!"
		err := authService.ResetPassword(testUser.ID, newPassword)
		require.NoError(t, err)

		// Verify old password no longer works
		_, err = authService.Authenticate("resetuser", password)
		assert.Error(t, err)

		// Verify new password works
		user, err := authService.Authenticate("resetuser", newPassword)
		require.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("Reset with weak password fails", func(t *testing.T) {
		err := authService.ResetPassword(testUser.ID, "weak")
		assert.Error(t, err)
	})
}

func TestMultiUserSupport(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	userStore, err := security.NewUserStore(storePath)
	require.NoError(t, err)

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	authService := security.NewAuthService(userStore, testLogger)

	t.Run("Create multiple users", func(t *testing.T) {
		users := []struct {
			id       string
			username string
			role     security.UserRole
		}{
			{"admin-001", "admin", security.RoleAdmin},
			{"power-001", "poweruser", security.RolePowerUser},
			{"user-001", "regularuser", security.RoleUser},
			{"guest-001", "guestuser", security.RoleGuest},
		}

		for _, u := range users {
			hash, err := authService.HashPassword("Password123!")
			require.NoError(t, err)

			user := &security.User{
				ID:           u.id,
				Username:     u.username,
				Email:        u.username + "@example.com",
				PasswordHash: hash,
				DisplayName:  u.username,
				Role:         u.role,
				Enabled:      true,
			}

			err = userStore.CreateUser(user)
			require.NoError(t, err)
		}

		assert.Equal(t, 4, userStore.UserCount())
	})

	t.Run("Retrieve users by different methods", func(t *testing.T) {
		// By ID
		user, err := userStore.GetUser("admin-001")
		require.NoError(t, err)
		assert.Equal(t, "admin", user.Username)

		// By username
		user, err = userStore.GetUserByUsername("poweruser")
		require.NoError(t, err)
		assert.Equal(t, "power-001", user.ID)

		// By email
		user, err = userStore.GetUserByEmail("regularuser@example.com")
		require.NoError(t, err)
		assert.Equal(t, "user-001", user.ID)
	})

	t.Run("List all users", func(t *testing.T) {
		users := userStore.ListUsers()
		assert.Len(t, users, 4)
	})
}

func TestUserStorePersistence(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "users.json")

	testLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	require.NoError(t, err)

	authService := security.NewAuthService(nil, testLogger)

	// Create and populate first store
	{
		userStore, err := security.NewUserStore(storePath)
		require.NoError(t, err)

		hash, err := authService.HashPassword("Password123!")
		require.NoError(t, err)

		testUser := &security.User{
			ID:           "persist-001",
			Username:     "persistuser",
			Email:        "persist@example.com",
			PasswordHash: hash,
			DisplayName:  "Persist User",
			Role:         security.RoleUser,
			Enabled:      true,
		}

		err = userStore.CreateUser(testUser)
		require.NoError(t, err)
	}

	// Load second store from same file
	{
		userStore, err := security.NewUserStore(storePath)
		require.NoError(t, err)

		// Verify user was persisted
		user, err := userStore.GetUserByUsername("persistuser")
		require.NoError(t, err)
		assert.Equal(t, "persist-001", user.ID)
		assert.Equal(t, "persist@example.com", user.Email)
	}
}
