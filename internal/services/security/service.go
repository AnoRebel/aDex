package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"aDex-UI/internal/logger"
)

// Service provides security features including input sanitization and validation
type Service struct {
	logger *logger.Logger
}

// SanitizationLevel represents the level of input sanitization
type SanitizationLevel int

const (
	SanitizationLevelMinimal SanitizationLevel = iota
	SanitizationLevelStandard
	SanitizationLevelStrict
	SanitizationLevelParanoid
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	MaxInputLength       int                `json:"maxInputLength"`
	SanitizationLevel    SanitizationLevel  `json:"sanitizationLevel"`
	AllowHTMLTags        bool               `json:"allowHTMLTags"`
	AllowedTags          []string           `json:"allowedTags"`
	AllowedAttributes    []string           `json:"allowedAttributes"`
	BlockXSS             bool               `json:"blockXSS"`
	BlockSQLInjection    bool               `json:"blockSQLInjection"`
	BlockCSRF            bool               `json:"blockCSRF"`
	RateLimitEnabled     bool               `json:"rateLimitEnabled"`
	MaxRequestsPerMinute int                `json:"maxRequestsPerMinute"`
	InputValidation      ValidationConfig   `json:"inputValidation"`
}

// ValidationConfig holds input validation settings
type ValidationConfig struct {
	RequireSecureHeaders bool     `json:"requireSecureHeaders"`
	MaxURLLength        int      `json:"maxURLLength"`
	AllowedProtocols    []string `json:"allowedProtocols"`
	BlockReservedIPs    bool     `json:"blockReservedIPs"`
	MaxFilenameLength   int      `json:"maxFilenameLength"`
}

// ValidationResult contains the result of input validation
type ValidationResult struct {
	IsValid   bool     `json:"isValid"`
	Sanitized string   `json:"sanitized"`
	Warnings  []string `json:"warnings"`
	Errors    []string `json:"errors"`
	RiskLevel int     `json:"riskLevel"` // 0-10
}

// SecurityIssue represents a detected security issue
type SecurityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Location    string `json:"location"`
	Timestamp   int64  `json:"timestamp"`
}

// CSRFToken represents a CSRF protection token
type CSRFToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

// NewService creates a new security service
func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// GetDefaultConfig returns the default security configuration
func (s *Service) GetDefaultConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxInputLength:       1024,
		SanitizationLevel:    SanitizationLevelStandard,
		AllowHTMLTags:        false,
		AllowedTags:          []string{"b", "i", "em", "strong", "br"},
		AllowedAttributes:    []string{"class", "id"},
		BlockXSS:             true,
		BlockSQLInjection:    true,
		BlockCSRF:            true,
		RateLimitEnabled:     true,
		MaxRequestsPerMinute: 60,
		InputValidation: ValidationConfig{
			RequireSecureHeaders: true,
			MaxURLLength:        2048,
			AllowedProtocols:    []string{"http", "https", "ws", "wss"},
			BlockReservedIPs:    true,
			MaxFilenameLength:   255,
		},
	}
}

// SanitizeInput sanitizes user input based on the provided configuration
func (s *Service) SanitizeInput(input string, config *SecurityConfig) *ValidationResult {
	if config == nil {
		config = s.GetDefaultConfig()
	}

	result := &ValidationResult{
		IsValid:   true,
		Sanitized: input,
		Warnings:  []string{},
		Errors:    []string{},
		RiskLevel: 0,
	}

	// Check input length
	if len(input) > config.MaxInputLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Input too long: %d > %d", len(input), config.MaxInputLength))
		result.RiskLevel += 3
		input = input[:config.MaxInputLength]
	}

	// Apply sanitization based on level
	switch config.SanitizationLevel {
	case SanitizationLevelMinimal:
		result.Sanitized = s.sanitizeMinimal(input, config)
	case SanitizationLevelStandard:
		result.Sanitized = s.sanitizeStandard(input, config)
	case SanitizationLevelStrict:
		result.Sanitized = s.sanitizeStrict(input, config)
	case SanitizationLevelParanoid:
		result.Sanitized = s.sanitizeParanoid(input, config)
	}

	// Security checks
	if config.BlockXSS && s.detectXSS(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Potential XSS attack detected")
		result.RiskLevel += 5
		s.logger.Warn("XSS attempt blocked", "input", input[:min(len(input), 100)])
	}

	if config.BlockSQLInjection && s.detectSQLInjection(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Potential SQL injection detected")
		result.RiskLevel += 5
		s.logger.Warn("SQL injection attempt blocked", "input", input[:min(len(input), 100)])
	}

	// Path traversal detection
	if s.detectPathTraversal(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Path traversal attempt detected")
		result.RiskLevel += 4
		s.logger.Warn("Path traversal attempt blocked", "input", input[:min(len(input), 100)])
	}

	// Command injection detection
	if s.detectCommandInjection(input) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Command injection attempt detected")
		result.RiskLevel += 5
		s.logger.Warn("Command injection attempt blocked", "input", input[:min(len(input), 100)])
	}

	return result
}

// ValidateURL validates a URL for security
func (s *Service) ValidateURL(rawURL string, config *SecurityConfig) *ValidationResult {
	result := &ValidationResult{
		IsValid:   true,
		Sanitized: rawURL,
		Warnings:  []string{},
		Errors:    []string{},
		RiskLevel: 0,
	}

	if config == nil {
		config = s.GetDefaultConfig()
	}

	// Check URL length
	if len(rawURL) > config.InputValidation.MaxURLLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("URL too long: %d > %d", len(rawURL), config.InputValidation.MaxURLLength))
		result.RiskLevel += 2
		return result
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid URL format: %v", err))
		result.RiskLevel += 3
		return result
	}

	// Check protocol
	protocolAllowed := false
	for _, protocol := range config.InputValidation.AllowedProtocols {
		if parsedURL.Scheme == protocol {
			protocolAllowed = true
			break
		}
	}

	if !protocolAllowed {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Protocol not allowed: %s", parsedURL.Scheme))
		result.RiskLevel += 3
		return result
	}

	// Check for suspicious patterns
	if s.detectSuspiciousURL(rawURL) {
		result.IsValid = false
		result.Errors = append(result.Errors, "Suspicious URL pattern detected")
		result.RiskLevel += 4
		s.logger.Warn("Suspicious URL blocked", "url", rawURL)
	}

	result.Sanitized = parsedURL.String()
	return result
}

// ValidateFilename validates a filename for security
func (s *Service) ValidateFilename(filename string, config *SecurityConfig) *ValidationResult {
	result := &ValidationResult{
		IsValid:   true,
		Sanitized: filename,
		Warnings:  []string{},
		Errors:    []string{},
		RiskLevel: 0,
	}

	if config == nil {
		config = s.GetDefaultConfig()
	}

	// Check filename length
	if len(filename) > config.InputValidation.MaxFilenameLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Filename too long: %d > %d", len(filename), config.InputValidation.MaxFilenameLength))
		result.RiskLevel += 2
		return result
	}

	// Check for empty filename
	if strings.TrimSpace(filename) == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Filename cannot be empty")
		result.RiskLevel += 2
		return result
	}

	// Check for dangerous characters
	dangerousChars := []string{
		"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|",
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05",
	}

	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Dangerous character in filename: %s", char))
			result.RiskLevel += 3
			s.logger.Warn("Dangerous filename blocked", "filename", filename)
		}
	}

	// Check for reserved names (Windows)
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	base := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range reservedNames {
		if base == reserved {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Reserved filename: %s", base))
			result.RiskLevel += 2
		}
	}

	// Sanitize filename
	result.Sanitized = s.sanitizeFilename(filename)

	return result
}

// GenerateCSRFToken generates a new CSRF protection token
func (s *Service) GenerateCSRFToken() (*CSRFToken, error) {
	// Generate 32-byte random token
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	// Set expiration to 1 hour
	expiresAt := time.Now().Add(time.Hour).Unix()

	return &CSRFToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateCSRFToken validates a CSRF token
func (s *Service) ValidateCSRFToken(token string, expectedToken *CSRFToken) bool {
	if expectedToken == nil {
		return false
	}

	// Check expiration
	if time.Now().Unix() > expectedToken.ExpiresAt {
		return false
	}

	// Check token match
	return token == expectedToken.Token
}

// HashPassword securely hashes a password using argon2id
// This method is deprecated - use AuthService.HashPassword instead
func (s *Service) HashPassword(password string) (string, error) {
	// Create a temporary AuthService for backwards compatibility
	authService := NewAuthService(nil, s.logger)
	return authService.HashPassword(password)
}

// GenerateSecureToken generates a cryptographically secure random token
func (s *Service) GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// DetectSecurityIssues scans input for security issues
func (s *Service) DetectSecurityIssues(input string, inputType string) []SecurityIssue {
	var issues []SecurityIssue

	// XSS detection
	if s.detectXSS(input) {
		issues = append(issues, SecurityIssue{
			Type:        "XSS",
			Severity:    "High",
			Description: "Cross-site scripting attempt detected",
			Input:       input,
			Location:    inputType,
			Timestamp:   time.Now().Unix(),
		})
	}

	// SQL injection detection
	if s.detectSQLInjection(input) {
		issues = append(issues, SecurityIssue{
			Type:        "SQL Injection",
			Severity:    "High",
			Description: "SQL injection attempt detected",
			Input:       input,
			Location:    inputType,
			Timestamp:   time.Now().Unix(),
		})
	}

	// Path traversal detection
	if s.detectPathTraversal(input) {
		issues = append(issues, SecurityIssue{
			Type:        "Path Traversal",
			Severity:    "Medium",
			Description: "Path traversal attempt detected",
			Input:       input,
			Location:    inputType,
			Timestamp:   time.Now().Unix(),
		})
	}

	// Command injection detection
	if s.detectCommandInjection(input) {
		issues = append(issues, SecurityIssue{
			Type:        "Command Injection",
			Severity:    "High",
			Description: "Command injection attempt detected",
			Input:       input,
			Location:    inputType,
			Timestamp:   time.Now().Unix(),
		})
	}

	return issues
}

// Private helper methods

// sanitizeMinimal applies minimal sanitization
func (s *Service) sanitizeMinimal(input string, config *SecurityConfig) string {
	// Basic HTML escaping
	return html.EscapeString(input)
}

// sanitizeStandard applies standard sanitization
func (s *Service) sanitizeStandard(input string, config *SecurityConfig) string {
	// HTML escape + remove potentially dangerous characters
	sanitized := html.EscapeString(input)

	// Remove null bytes and control characters
	sanitized = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	return sanitized
}

// sanitizeStrict applies strict sanitization
func (s *Service) sanitizeStrict(input string, config *SecurityConfig) string {
	// Remove all HTML tags and escape special characters
	sanitized := html.EscapeString(input)

	// Remove suspicious patterns
	sanitized = regexp.MustCompile(`javascript:|vbscript:|data:|file:`).ReplaceAllString(sanitized, "")

	// Remove control characters
	sanitized = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	return sanitized
}

// sanitizeParanoid applies paranoid-level sanitization
func (s *Service) sanitizeParanoid(input string, config *SecurityConfig) string {
	// Allow only alphanumeric and basic punctuation
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9\s.,!?()@#$%^&*\-_=+\[\]{}'"` + "`]`).ReplaceAllString(input, "")
	return html.EscapeString(sanitized)
}

// detectXSS detects potential XSS attacks
func (s *Service) detectXSS(input string) bool {
	lower := strings.ToLower(input)

	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=",
		"onerror=", "onclick=", "onmouseover=", "onfocus=", "onblur=",
		"eval(", "expression(", "alert(", "confirm(", "prompt(",
		"<iframe", "<object", "<embed", "<link", "<meta", "<style",
		"@import", "behavior:", "binding:", "include-source:",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// detectSQLInjection detects potential SQL injection attacks
func (s *Service) detectSQLInjection(input string) bool {
	lower := strings.ToLower(input)

	sqlPatterns := []string{
		"union select", "drop table", "insert into", "delete from",
		"update set", "create table", "alter table", "exec(", "execute(",
		"sp_executesql", "xp_cmdshell", "--", "/*", "*/", "' or '1'='1",
		"' or 1=1", "' waitfor delay '", "sleep(", "benchmark(",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// detectPathTraversal detects path traversal attacks
func (s *Service) detectPathTraversal(input string) bool {
	pathPatterns := []string{
		"../", "..\\", "%2e%2e%2f", "%2e%2e\\", "..%2f", "..%5c",
		"/etc/passwd", "/etc/shadow", "/proc/", "/sys/", "/root/",
		"\\windows\\system32", "c:\\windows", "boot.ini",
	}

	lower := strings.ToLower(input)
	for _, pattern := range pathPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// detectCommandInjection detects command injection attacks
func (s *Service) detectCommandInjection(input string) bool {
	lower := strings.ToLower(input)

	commandPatterns := []string{
		";", "|", "&", "&&", "||", "`", "$(", "${", "&&", "||",
		">>", ">", "<", "2>&1", "/dev/null", "nc ", "netcat",
		"wget ", "curl ", "powershell", "cmd.exe", "/bin/sh",
		"/bin/bash", "python -c", "perl -e", "ruby -e",
	}

	for _, pattern := range commandPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// detectSuspiciousURL detects suspicious URL patterns
func (s *Service) detectSuspiciousURL(url string) bool {
	lower := strings.ToLower(url)

	suspiciousPatterns := []string{
		"javascript:", "vbscript:", "data:", "file:", "ftp:",
		"mailto:", "tel:", "sms:", "callto:", "tftp:",
		"ldap:", "gopher:", "mms:", "rtp:", "rtsp:",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.HasPrefix(lower, pattern) {
			return true
		}
	}

	return false
}

// sanitizeFilename sanitizes a filename
func (s *Service) sanitizeFilename(filename string) string {
	// Remove dangerous characters
	sanitized := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`).ReplaceAllString(filename, "_")

	// Trim leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")

	// Replace multiple dots with single dot
	sanitized = regexp.MustCompile(`\.+`).ReplaceAllString(sanitized, ".")

	return sanitized
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}