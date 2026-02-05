package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"aDex-UI/internal/logger"
	"golang.org/x/crypto/argon2"
)

// AuthService handles user authentication with argon2id password hashing
type AuthService struct {
	userStore *UserStore
	logger    *logger.Logger
	config    *AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Argon2id parameters
	Memory      uint32 `json:"memory"`      // Memory in KiB (default: 64MB)
	Iterations  uint32 `json:"iterations"`  // Number of iterations (default: 3)
	Parallelism uint8  `json:"parallelism"` // Degree of parallelism (default: 4)
	SaltLength  uint32 `json:"saltLength"`  // Salt length in bytes (default: 16)
	KeyLength   uint32 `json:"keyLength"`   // Derived key length (default: 32)

	// Authentication settings
	MaxLoginAttempts       int  `json:"maxLoginAttempts"`
	LockoutDurationMinutes int  `json:"lockoutDurationMinutes"`
	RequireStrongPassword  bool `json:"requireStrongPassword"`
	MinPasswordLength      int  `json:"minPasswordLength"`
	RequireUppercase       bool `json:"requireUppercase"`
	RequireLowercase       bool `json:"requireLowercase"`
	RequireNumbers         bool `json:"requireNumbers"`
	RequireSpecialChars    bool `json:"requireSpecialChars"`
}

// LoginAttempt tracks login attempts for rate limiting
type LoginAttempt struct {
	Username    string
	Attempts    int
	LastAttempt int64
	LockedUntil int64
}

// NewAuthService creates a new authentication service with argon2id
func NewAuthService(userStore *UserStore, logger *logger.Logger) *AuthService {
	return &AuthService{
		userStore: userStore,
		logger:    logger,
		config:    getDefaultAuthConfig(),
	}
}

// getDefaultAuthConfig returns secure default configuration for argon2id
func getDefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		// Argon2id parameters (OWASP recommendations)
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,         // 3 iterations
		Parallelism: 4,         // 4 parallel threads
		SaltLength:  16,        // 16 bytes salt
		KeyLength:   32,        // 32 bytes key

		// Authentication settings
		MaxLoginAttempts:       5,
		LockoutDurationMinutes: 15,
		RequireStrongPassword:  true,
		MinPasswordLength:      8,
		RequireUppercase:       true,
		RequireLowercase:       true,
		RequireNumbers:         true,
		RequireSpecialChars:    true,
	}
}

// HashPassword hashes a password using argon2id
func (a *AuthService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Validate password strength if required
	if a.config.RequireStrongPassword {
		if err := a.validatePasswordStrength(password); err != nil {
			return "", fmt.Errorf("password validation failed: %w", err)
		}
	}

	// Generate a cryptographically secure random salt
	salt := make([]byte, a.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		a.config.Iterations,
		a.config.Memory,
		a.config.Parallelism,
		a.config.KeyLength,
	)

	// Encode the hash and salt in a standard format
	// Format: $argon2id$v=19$m=65536,t=3,p=4$<base64_salt>$<base64_hash>
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.config.Memory,
		a.config.Iterations,
		a.config.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword verifies a password against an argon2id hash
func (a *AuthService) VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	// Verify algorithm
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	// Parse parameters
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute hash for the input password using the same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(hash)),
	)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

// Authenticate authenticates a user with username and password
func (a *AuthService) Authenticate(username, password string) (*User, error) {
	// Get user by username
	user, err := a.userStore.GetUserByUsername(username)
	if err != nil {
		a.logger.Warn("Authentication failed: user not found", map[string]interface{}{"username": username})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is enabled
	if !user.Enabled {
		a.logger.Warn("Authentication failed: user disabled", map[string]interface{}{"username": username})
		return nil, fmt.Errorf("user account is disabled")
	}

	// Verify password
	valid, err := a.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		a.logger.Error("Password verification error", err, map[string]interface{}{"username": username})
		return nil, fmt.Errorf("authentication failed")
	}

	if !valid {
		a.logger.Warn("Authentication failed: invalid password", map[string]interface{}{"username": username})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	if err := a.userStore.UpdateLastLogin(user.ID); err != nil {
		a.logger.Warn("Failed to update last login", map[string]interface{}{"error": err.Error(), "user_id": user.ID})
		// Don't fail authentication if we can't update login time
	}

	a.logger.Info("User authenticated successfully", map[string]interface{}{"username": username, "user_id": user.ID})
	return user, nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Get user
	user, err := a.userStore.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify old password
	valid, err := a.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !valid {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	if a.config.RequireStrongPassword {
		if err := a.validatePasswordStrength(newPassword); err != nil {
			return fmt.Errorf("new password validation failed: %w", err)
		}
	}

	// Hash new password
	newHash, err := a.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update user's password
	user.PasswordHash = newHash
	if err := a.userStore.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	a.logger.Info("Password changed successfully", map[string]interface{}{"user_id": userID})
	return nil
}

// ResetPassword resets a user's password (admin function)
func (a *AuthService) ResetPassword(userID, newPassword string) error {
	// Get user
	user, err := a.userStore.GetUser(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Validate new password strength
	if a.config.RequireStrongPassword {
		if err := a.validatePasswordStrength(newPassword); err != nil {
			return fmt.Errorf("password validation failed: %w", err)
		}
	}

	// Hash new password
	newHash, err := a.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	user.PasswordHash = newHash
	if err := a.userStore.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	a.logger.Info("Password reset successfully", map[string]interface{}{"user_id": userID})
	return nil
}

// validatePasswordStrength validates password strength requirements
func (a *AuthService) validatePasswordStrength(password string) error {
	if len(password) < a.config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", a.config.MinPasswordLength)
	}

	if a.config.RequireUppercase && !containsUppercase(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if a.config.RequireLowercase && !containsLowercase(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if a.config.RequireNumbers && !containsNumber(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if a.config.RequireSpecialChars && !containsSpecialChar(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// Helper functions for password validation

func containsUppercase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~"
	for _, c := range s {
		for _, special := range specialChars {
			if c == special {
				return true
			}
		}
	}
	return false
}
