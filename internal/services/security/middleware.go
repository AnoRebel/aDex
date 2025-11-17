package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adex-ui/aDex-UI/internal/logger"
)

// Middleware provides HTTP security middleware
type Middleware struct {
	securityService *Service
	logger          *logger.Logger
	config          *SecurityConfig
	rateLimiter     *RateLimiter
}

// RateLimiter provides request rate limiting
type RateLimiter struct {
	clients map[string]*ClientLimits
	config  *SecurityConfig
}

// ClientLimits tracks rate limits for a client
type ClientLimits struct {
	Requests      int
	ResetTime     time.Time
	LastReset     time.Time
	TokenCount    map[string]int
	TokenReset    time.Time
}

// SecurityHeaders adds security headers to HTTP responses
type SecurityHeaders struct {
	ContentSecurityPolicy string
	XFrameOptions         string
	XContentTypeOptions   string
	XSSProtection         string
	StrictTransportSecurity string
	ReferrerPolicy        string
}

// NewMiddleware creates a new security middleware
func NewMiddleware(securityService *Service, logger *logger.Logger, config *SecurityConfig) *Middleware {
	if config == nil {
		config = securityService.GetDefaultConfig()
	}

	return &Middleware{
		securityService: securityService,
		logger:          logger,
		config:          config,
		rateLimiter:     NewRateLimiter(config),
	}
}

// SecurityMiddleware returns an HTTP middleware function for security
func (m *Middleware) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply security headers
		m.addSecurityHeaders(w)

		// Rate limiting check
		if m.config.RateLimitEnabled {
			clientIP := m.getClientIP(r)
			if !m.rateLimiter.IsAllowed(clientIP) {
				m.logger.Warn("Rate limit exceeded", "client_ip", clientIP)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		// CSRF protection for state-changing methods
		if m.config.BlockCSRF && m.isStateChangingMethod(r.Method) {
			if !m.validateCSRFToken(r) {
				m.logger.Warn("CSRF token validation failed", "method", r.Method, "path", r.URL.Path)
				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}
		}

		// Input validation for query parameters
		if err := m.validateQueryParams(r); err != nil {
			m.logger.Warn("Invalid query parameters", "error", err, "path", r.URL.Path)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		// Check for suspicious user agent
		if m.isSuspiciousUserAgent(r.UserAgent()) {
			m.logger.Warn("Suspicious user agent blocked", "user_agent", r.UserAgent())
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// InputSanitizationMiddleware provides input sanitization for requests
func (m *Middleware) InputSanitizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sanitize form data
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if err := r.ParseForm(); err == nil {
				m.sanitizeFormValues(r)
			}
		}

		// Sanitize URL parameters
		m.sanitizeURLParams(r)

		next.ServeHTTP(w, r)
	})
}

// ValidateInputMiddleware validates input data
func (m *Middleware) ValidateInputMiddleware(inputType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get input data based on method and content type
			input := m.extractInputData(r)

			if input != "" {
				result := m.securityService.SanitizeInput(input, m.config)
				if !result.IsValid {
					m.logger.Warn("Invalid input detected",
						"input_type", inputType,
						"errors", result.Errors,
						"risk_level", result.RiskLevel)

					http.Error(w, fmt.Sprintf("Invalid input: %s", strings.Join(result.Errors, ", ")), http.StatusBadRequest)
					return
				}

				// Store sanitized input back in request context
				ctx := context.WithValue(r.Context(), "sanitized_input", result.Sanitized)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs security-relevant events
func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log suspicious activities
		if wrapped.statusCode >= 400 {
			m.logSecurityEvent(r, wrapped.statusCode, duration)
		}

		// Log high-value targets
		if m.isHighValueTarget(r.URL.Path) {
			m.logger.Info("High-value target accessed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"client_ip", m.getClientIP(r))
		}
	})
}

// Private helper methods

// addSecurityHeaders adds security headers to the response
func (m *Middleware) addSecurityHeaders(w http.ResponseWriter) {
	headers := m.getSecurityHeaders()

	if headers.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", headers.ContentSecurityPolicy)
	}
	w.Header().Set("X-Frame-Options", headers.XFrameOptions)
	w.Header().Set("X-Content-Type-Options", headers.XContentTypeOptions)
	w.Header().Set("X-XSS-Protection", headers.XSSProtection)
	w.Header().Set("Referrer-Policy", headers.ReferrerPolicy)

	// HSTS only for HTTPS
	if headers.StrictTransportSecurity != "" {
		w.Header().Set("Strict-Transport-Security", headers.StrictTransportSecurity)
	}

	// Additional security headers
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
}

// getSecurityHeaders returns configured security headers
func (m *Middleware) getSecurityHeaders() SecurityHeaders {
	return SecurityHeaders{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:;",
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		XSSProtection:         "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// getClientIP extracts the real client IP address
func (m *Middleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// isStateChangingMethod checks if the HTTP method changes state
func (m *Middleware) isStateChangingMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

// validateCSRFToken validates CSRF token for the request
func (m *Middleware) validateCSRFToken(r *http.Request) bool {
	// Get token from header
	headerToken := r.Header.Get("X-CSRF-Token")

	// Get token from form
	formToken := r.FormValue("csrf_token")

	// Use whichever is provided
	token := headerToken
	if token == "" {
		token = formToken
	}

	// In a real implementation, validate against session/cookie
	// This is a placeholder that should be replaced with proper session validation
	return token != ""
}

// validateQueryParams validates all query parameters
func (m *Middleware) validateQueryParams(r *http.Request) error {
	for key, values := range r.URL.Query() {
		for _, value := range values {
			result := m.securityService.SanitizeInput(value, m.config)
			if !result.IsValid {
				return fmt.Errorf("invalid parameter %s: %s", key, strings.Join(result.Errors, ", "))
			}
		}
	}
	return nil
}

// isSuspiciousUserAgent checks for suspicious user agents
func (m *Middleware) isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true
	}

	suspiciousPatterns := []string{
		"sqlmap", "nikto", "dirb", "nmap", "masscan", "zap",
		"burp", "scanner", "crawler", "bot", "spider",
	}

	ua := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}

	return false
}

// sanitizeFormValues sanitizes form values
func (m *Middleware) sanitizeFormValues(r *http.Request) {
	for key := range r.Form {
		values := r.Form[key]
		for i, value := range values {
			result := m.securityService.SanitizeInput(value, m.config)
			values[i] = result.Sanitized
		}
		r.Form[key] = values
	}
}

// sanitizeURLParams sanitizes URL parameters
func (m *Middleware) sanitizeURLParams(r *http.Request) {
	for key, values := range r.URL.Query() {
		for i, value := range values {
			result := m.securityService.SanitizeInput(value, m.config)
			values[i] = result.Sanitized
		}
		// Update query
		query := r.URL.Query()
		query.Set(key, values[0])
		r.URL.RawQuery = query.Encode()
	}
}

// extractInputData extracts input data from request
func (m *Middleware) extractInputData(r *http.Request) string {
	// Try to get input from various sources
	if r.Method == http.MethodGet {
		return r.URL.RawQuery
	}

	// For POST/PUT/PATCH, try form data
	if err := r.ParseForm(); err == nil {
		return r.Form.Encode()
	}

	return ""
}

// isHighValueTarget checks if the requested path is a high-value target
func (m *Middleware) isHighValueTarget(path string) bool {
	highValuePaths := []string{
		"/admin", "/api/admin", "/settings", "/config",
		"/users", "/auth", "/login", "/register",
	}

	for _, hvp := range highValuePaths {
		if strings.HasPrefix(path, hvp) {
			return true
		}
	}

	return false
}

// logSecurityEvent logs security-related events
func (m *Middleware) logSecurityEvent(r *http.Request, statusCode int, duration time.Duration) {
	level := "info"
	if statusCode >= 500 {
		level = "error"
	} else if statusCode >= 400 {
		level = "warn"
	}

	m.logger.Log(level, "Security event",
		"method", r.Method,
		"path", r.URL.Path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
		"client_ip", m.getClientIP(r),
		"user_agent", r.UserAgent())
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *SecurityConfig) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*ClientLimits),
		config:  config,
	}
}

// IsAllowed checks if a client is allowed to make a request
func (rl *RateLimiter) IsAllowed(clientID string) bool {
	now := time.Now()

	// Initialize client if not exists
	if _, exists := rl.clients[clientID]; !exists {
		rl.clients[clientID] = &ClientLimits{
			Requests:   0,
			ResetTime:  now.Add(time.Minute),
			LastReset:  now,
			TokenCount: make(map[string]int),
			TokenReset: now.Add(time.Hour),
		}
	}

	client := rl.clients[clientID]

	// Reset if window expired
	if now.After(client.ResetTime) {
		client.Requests = 0
		client.ResetTime = now.Add(time.Minute)
		client.LastReset = now
	}

	// Check rate limit
	if client.Requests >= rl.config.MaxRequestsPerMinute {
		return false
	}

	client.Requests++
	return true
}

// Cleanup removes old client entries
func (rl *RateLimiter) Cleanup() {
	now := time.Now()
	for clientID, client := range rl.clients {
		if now.Sub(client.LastReset) > 5*time.Minute {
			delete(rl.clients, clientID)
		}
	}
}