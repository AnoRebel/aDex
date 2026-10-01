package tests

import (
	"testing"

	"aDex/internal/logger"
	"aDex/internal/services/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestLogger(t *testing.T) *logger.Logger {
	testLogger, err := logger.New(logger.Config{
		Level:  logger.LevelDebug,
		Format: logger.FormatText,
		Output: "stdout",
	})
	require.NoError(t, err)
	return testLogger
}

func TestSecurityService(t *testing.T) {
	testLogger := createTestLogger(t)
	service := security.NewService(testLogger)

	t.Run("DefaultConfig", func(t *testing.T) {
		config := service.GetDefaultConfig()
		require.NotNil(t, config)
		assert.Equal(t, 1024, config.MaxInputLength)
		assert.True(t, config.BlockXSS)
		assert.True(t, config.BlockSQLInjection)
	})

	t.Run("SanitizeInput_ValidInput", func(t *testing.T) {
		result := service.SanitizeInput("Hello, World!", nil)
		assert.True(t, result.IsValid)
		assert.Equal(t, "Hello, World!", result.Sanitized)
		assert.Empty(t, result.Errors)
		assert.Equal(t, 0, result.RiskLevel)
	})

	t.Run("SanitizeInput_XSSDetection", func(t *testing.T) {
		xssInput := "<script>alert('xss')</script>"
		result := service.SanitizeInput(xssInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_SQLInjectionDetection", func(t *testing.T) {
		sqlInput := "'; DROP TABLE users; --"
		result := service.SanitizeInput(sqlInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_PathTraversal", func(t *testing.T) {
		pathInput := "../../../etc/passwd"
		result := service.SanitizeInput(pathInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_CommandInjection", func(t *testing.T) {
		cmdInput := "; rm -rf /"
		result := service.SanitizeInput(cmdInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_TooLong", func(t *testing.T) {
		longInput := string(make([]byte, 2000))
		result := service.SanitizeInput(longInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("ValidateURL_ValidURL", func(t *testing.T) {
		validURL := "https://example.com/path?param=value"
		result := service.ValidateURL(validURL, nil)
		assert.True(t, result.IsValid)
		assert.Equal(t, validURL, result.Sanitized)
		assert.Empty(t, result.Errors)
	})

	t.Run("ValidateURL_InvalidProtocol", func(t *testing.T) {
		invalidURL := "javascript:alert('xss')"
		result := service.ValidateURL(invalidURL, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("ValidateURL_TooLong", func(t *testing.T) {
		longURL := "https://example.com/" + string(make([]byte, 3000))
		result := service.ValidateURL(longURL, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("ValidateFilename_ValidFilename", func(t *testing.T) {
		validFilename := "document.txt"
		result := service.ValidateFilename(validFilename, nil)
		assert.True(t, result.IsValid)
		assert.Equal(t, validFilename, result.Sanitized)
		assert.Empty(t, result.Errors)
	})

	t.Run("ValidateFilename_DangerousCharacters", func(t *testing.T) {
		dangerousFilename := "../../../etc/passwd"
		result := service.ValidateFilename(dangerousFilename, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("ValidateFilename_ReservedName", func(t *testing.T) {
		reservedFilename := "CON.txt"
		result := service.ValidateFilename(reservedFilename, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("GenerateCSRFToken", func(t *testing.T) {
		token, err := service.GenerateCSRFToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token.Token)
		assert.Greater(t, token.ExpiresAt, int64(0))
	})

	t.Run("ValidateCSRFToken", func(t *testing.T) {
		token, err := service.GenerateCSRFToken()
		require.NoError(t, err)

		// Valid token
		assert.True(t, service.ValidateCSRFToken(token.Token, token))

		// Invalid token
		assert.False(t, service.ValidateCSRFToken("invalid", token))
	})

	t.Run("GenerateSecureToken", func(t *testing.T) {
		token, err := service.GenerateSecureToken(32)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 30)

		// Invalid length
		_, err = service.GenerateSecureToken(0)
		assert.Error(t, err)
	})

	t.Run("DetectSecurityIssues", func(t *testing.T) {
		xssInput := "<script>alert('xss')</script>"
		issues := service.DetectSecurityIssues(xssInput, "test_input")
		// May detect multiple issues (XSS and Command Injection)
		assert.GreaterOrEqual(t, len(issues), 1)
		// First should be XSS
		assert.Equal(t, "XSS", issues[0].Type)
		assert.Equal(t, "High", issues[0].Severity)
		assert.Equal(t, "test_input", issues[0].Location)
	})
}

func TestRateLimiter(t *testing.T) {
	config := &security.SecurityConfig{
		RateLimitEnabled:     true,
		MaxRequestsPerMinute: 3,
	}

	rateLimiter := security.NewRateLimiter(config)

	t.Run("InitialAllow", func(t *testing.T) {
		assert.True(t, rateLimiter.IsAllowed("client1"))
		assert.True(t, rateLimiter.IsAllowed("client1"))
		assert.True(t, rateLimiter.IsAllowed("client1"))
	})

	t.Run("RateLimitExceeded", func(t *testing.T) {
		assert.False(t, rateLimiter.IsAllowed("client1"))
	})

	t.Run("DifferentClients", func(t *testing.T) {
		assert.True(t, rateLimiter.IsAllowed("client2"))
		assert.True(t, rateLimiter.IsAllowed("client2"))
	})

	t.Run("Cleanup", func(t *testing.T) {
		rateLimiter.Cleanup()
		// Should not panic and should clean old entries
	})
}

// Benchmark tests
func BenchmarkSanitizeInput(b *testing.B) {
	testLogger, _ := logger.New(logger.Config{
		Level:  logger.LevelWarn,
		Format: logger.FormatText,
		Output: "stdout",
	})

	service := security.NewService(testLogger)
	input := "Hello, World! This is a normal input string."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SanitizeInput(input, nil)
	}
}

func BenchmarkValidateURL(b *testing.B) {
	testLogger, _ := logger.New(logger.Config{
		Level:  logger.LevelWarn,
		Format: logger.FormatText,
		Output: "stdout",
	})

	service := security.NewService(testLogger)
	url := "https://example.com/path?param=value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateURL(url, nil)
	}
}
