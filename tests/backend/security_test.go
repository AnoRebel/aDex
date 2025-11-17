package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aDex-UI/internal/logger"
	"aDex-UI/internal/services/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityService(t *testing.T) {
	logger := logger.New(logger.Config{
		Level:  "debug",
		Output: "test",
	})

	service := security.NewService(logger)
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
		assert.Contains(t, result.Errors[0], "XSS")
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_SQLInjectionDetection", func(t *testing.T) {
		sqlInput := "'; DROP TABLE users; --"
		result := service.SanitizeInput(sqlInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "SQL injection")
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_PathTraversal", func(t *testing.T) {
		pathInput := "../../../etc/passwd"
		result := service.SanitizeInput(pathInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "Path traversal")
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_CommandInjection", func(t *testing.T) {
		cmdInput := "; rm -rf /"
		result := service.SanitizeInput(cmdInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "Command injection")
		assert.Greater(t, result.RiskLevel, 0)
	})

	t.Run("SanitizeInput_TooLong", func(t *testing.T) {
		longInput := string(make([]byte, 2000))
		result := service.SanitizeInput(longInput, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "too long")
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
		assert.Contains(t, result.Errors[0], "Protocol not allowed")
	})

	t.Run("ValidateURL_TooLong", func(t *testing.T) {
		longURL := "https://example.com/" + string(make([]byte, 3000))
		result := service.ValidateURL(longURL, nil)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
		assert.Contains(t, result.Errors[0], "too long")
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
		assert.Contains(t, result.Errors[0], "Reserved filename")
	})

	t.Run("GenerateCSRFToken", func(t *testing.T) {
		token, err := service.GenerateCSRFToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token.Token)
		assert.Greater(t, token.ExpiresAt, time.Now().Unix())
	})

	t.Run("ValidateCSRFToken", func(t *testing.T) {
		token, err := service.GenerateCSRFToken()
		require.NoError(t, err)

		// Valid token
		assert.True(t, service.ValidateCSRFToken(token.Token, token))

		// Invalid token
		assert.False(t, service.ValidateCSRFToken("invalid", token))

		// Expired token
		expiredToken := &security.CSRFToken{
			Token:     token.Token,
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		}
		assert.False(t, service.ValidateCSRFToken(token.Token, expiredToken))
	})

	t.Run("HashPassword", func(t *testing.T) {
		password := "securePassword123"
		hash, err := service.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		// Empty password
		_, err = service.HashPassword("")
		assert.Error(t, err)
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
		assert.Len(t, issues, 1)
		assert.Equal(t, "XSS", issues[0].Type)
		assert.Equal(t, "High", issues[0].Severity)
		assert.Equal(t, "test_input", issues[0].Location)
	})
}

func TestSecurityMiddleware(t *testing.T) {
	logger := logger.New(logger.Config{
		Level:  "debug",
		Output: "test",
	})

	securityService := security.NewService(logger)
	config := securityService.GetDefaultConfig()
	middleware := security.NewMiddleware(securityService, logger, config)

	t.Run("SecurityHeaders", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.SecurityMiddleware(handler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	})

	t.Run("RateLimiting", func(t *testing.T) {
		// Enable rate limiting
		config.RateLimitEnabled = true
		config.MaxRequestsPerMinute = 2

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.SecurityMiddleware(handler)

		// First request should pass
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.RemoteAddr = "127.0.0.1:12345"
		w1 := httptest.NewRecorder()
		wrapped.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request should pass
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.RemoteAddr = "127.0.0.1:12345"
		w2 := httptest.NewRecorder()
		wrapped.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Third request should be rate limited
		req3 := httptest.NewRequest(http.MethodGet, "/", nil)
		req3.RemoteAddr = "127.0.0.1:12345"
		w3 := httptest.NewRecorder()
		wrapped.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	})

	t.Run("SuspiciousUserAgent", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.SecurityMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", "sqlmap/1.0")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("InvalidQueryParameters", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.SecurityMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/?param=<script>", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
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

func TestInputValidationLevels(t *testing.T) {
	logger := logger.New(logger.Config{
		Level:  "debug",
		Output: "test",
	})

	service := security.NewService(logger)

	testCases := []struct {
		name     string
		level    security.SanitizationLevel
		input    string
		expected string
	}{
		{
			name:     "Minimal_HTML",
			level:    security.SanitizationLevelMinimal,
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "Standard_ControlChars",
			level:    security.SanitizationLevelStandard,
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "Strict_Suspicious",
			level:    security.SanitizationLevelStrict,
			input:    "javascript:alert('xss')",
			expected: ":alert(&#39;xss&#39;)",
		},
		{
			name:     "Paranoid_SpecialChars",
			level:    security.SanitizationLevelParanoid,
			input:    "Hello <>\"'&",
			expected: "Hello ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &security.SecurityConfig{
				SanitizationLevel: tc.level,
			}

			result := service.SanitizeInput(tc.input, config)
			assert.True(t, result.IsValid)
			assert.Equal(t, tc.expected, result.Sanitized)
		})
	}
}

// Benchmark tests
func BenchmarkSanitizeInput(b *testing.B) {
	logger := logger.New(logger.Config{
		Level:  "warn", // Minimal logging for benchmarks
		Output: "test",
	})

	service := security.NewService(logger)
	input := "Hello, World! This is a normal input string."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SanitizeInput(input, nil)
	}
}

func BenchmarkValidateURL(b *testing.B) {
	logger := logger.New(logger.Config{
		Level:  "warn",
		Output: "test",
	})

	service := security.NewService(logger)
	url := "https://example.com/path?param=value"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateURL(url, nil)
	}
}