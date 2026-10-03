package error

import (
	"aDex/internal/utils"
	"fmt"
	"runtime"
	"strings"
	"time"

	"aDex/internal/logger"
)

// Service provides comprehensive error handling and user-friendly messaging
type Service struct {
	logger    *logger.Logger
	config    *ErrorConfig
	templates map[string]ErrorTemplate
}

// ErrorConfig holds error handling configuration
type ErrorConfig struct {
	ShowStackTrace      bool          `json:"showStackTrace"`
	LogLevel            string        `json:"logLevel"`
	UserFriendlyErrors  bool          `json:"userFriendlyErrors"`
	ErrorReporting      bool          `json:"errorReporting"`
	ReportingInterval   time.Duration `json:"reportingInterval"`
	MaxErrorsInMemory   int           `json:"maxErrorsInMemory"`
	EnableRecovery      bool          `json:"enableRecovery"`
	AutoRetryAttempts   int           `json:"autoRetryAttempts"`
	RetryDelay          time.Duration `json:"retryDelay"`
	LocalizationEnabled bool          `json:"localizationEnabled"`
	DefaultLanguage     string        `json:"defaultLanguage"`
}

// ErrorTemplate defines how specific errors should be presented to users
type ErrorTemplate struct {
	Code          string            `json:"code"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Cause         string            `json:"cause"`
	Solution      string            `json:"solution"`
	Severity      string            `json:"severity"`      // "error", "warning", "info"
	Category      string            `json:"category"`      // "system", "user", "network", "file"
	Retryable     bool              `json:"retryable"`     // Can the user retry this operation
	RecoverySteps []string          `json:"recoverySteps"` // Steps the user can take
	Translations  map[string]string `json:"translations"`  // Language code -> translated message
}

// ErrorContext provides additional context for error handling
type ErrorContext struct {
	Operation   string                 `json:"operation"`
	Component   string                 `json:"component"`
	UserID      string                 `json:"userId,omitempty"`
	SessionID   string                 `json:"sessionId,omitempty"`
	RequestID   string                 `json:"requestId,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UserAgent   string                 `json:"userAgent,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  []StackFrame           `json:"stackTrace,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// AppError represents a structured application error
type AppError struct {
	ID          string         `json:"id"`
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	UserMessage string         `json:"userMessage"`
	Context     *ErrorContext  `json:"context"`
	Template    *ErrorTemplate `json:"template,omitempty"`
	Cause       error          `json:"cause,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Retryable   bool           `json:"retryable"`
	Recovered   bool           `json:"recovered"`
}

// Error implements the error interface for AppError
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// ErrorReport represents a collection of errors for reporting
type ErrorReport struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Version   string      `json:"version"`
	Errors    []*AppError `json:"errors"`
	Stats     ErrorStats  `json:"stats"`
	System    SystemInfo  `json:"system"`
}

// ErrorStats provides statistics about errors
type ErrorStats struct {
	TotalErrors      int            `json:"totalErrors"`
	ErrorsByCode     map[string]int `json:"errorsByCode"`
	ErrorsBySeverity map[string]int `json:"errorsBySeverity"`
	ErrorsByCategory map[string]int `json:"errorsByCategory"`
	RecoveredErrors  int            `json:"recoveredErrors"`
	RetryableErrors  int            `json:"retryableErrors"`
}

// SystemInfo provides system context for error reports
type SystemInfo struct {
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
	Memory       MemoryInfo        `json:"memory"`
	Environment  map[string]string `json:"environment"`
}

// MemoryInfo provides memory statistics
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"totalAlloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"numGC"`
}

// NewService creates a new error handling service
func NewService(logger *logger.Logger) *Service {
	service := &Service{
		logger:    logger,
		config:    getDefaultConfig(),
		templates: make(map[string]ErrorTemplate),
	}

	service.loadDefaultTemplates()
	service.startBackgroundTasks()

	return service
}

// NewError creates a new application error
func (s *Service) NewError(code, message string, cause error) *AppError {
	appError := &AppError{
		ID:        generateErrorID(),
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Cause:     cause,
	}

	// Find template for this error code
	if template, exists := s.templates[code]; exists {
		appError.Template = &template
		appError.UserMessage = template.Description
		appError.Retryable = template.Retryable
	} else {
		appError.UserMessage = message
		appError.Retryable = false
	}

	// Generate user-friendly message if enabled
	if s.config.UserFriendlyErrors {
		appError.UserMessage = s.generateUserFriendlyMessage(appError)
	}

	// Capture context
	appError.Context = s.captureErrorContext()

	return appError
}

// HandleError processes an error with optional context
func (s *Service) HandleError(err error, context *ErrorContext) *AppError {
	var appError *AppError

	if ae, ok := err.(*AppError); ok {
		appError = ae
	} else {
		appError = s.NewError("UNKNOWN", err.Error(), err)
	}

	// Merge context if provided
	if context != nil {
		if appError.Context == nil {
			appError.Context = context
		} else {
			// Merge metadata
			if context.Metadata != nil {
				if appError.Context.Metadata == nil {
					appError.Context.Metadata = make(map[string]interface{})
				}
				for k, v := range context.Metadata {
					appError.Context.Metadata[k] = v
				}
			}
			// Update other fields
			if context.Operation != "" {
				appError.Context.Operation = context.Operation
			}
			if context.Component != "" {
				appError.Context.Component = context.Component
			}
		}
	}

	// Log the error
	s.logError(appError)

	// Attempt recovery if enabled
	if s.config.EnableRecovery && appError.Retryable {
		appError.Recovered = s.attemptRecovery(appError)
	}

	return appError
}

// HandlePanic recovers from panic and converts to structured error
func (s *Service) HandlePanic() {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = fmt.Errorf("panic: %s", x)
		case error:
			err = x
		default:
			err = fmt.Errorf("panic: %v", x)
		}

		appError := s.NewError("PANIC", "Application panic occurred", err)
		appError.Context = s.captureErrorContext()

		s.logError(appError)
		s.logger.Error("Panic recovered", fmt.Errorf("%s", appError.Message), map[string]interface{}{"id": appError.ID})
	}
}

// GetErrorMessage returns the appropriate error message for the user
func (s *Service) GetErrorMessage(appError *AppError, language string) string {
	if appError.Template != nil {
		if translation, exists := appError.Template.Translations[language]; exists {
			return translation
		}
	}

	return appError.UserMessage
}

// GetRecoverySteps returns steps the user can take to recover from an error
func (s *Service) GetRecoverySteps(appError *AppError) []string {
	if appError.Template != nil {
		return appError.Template.RecoverySteps
	}

	return []string{
		"Please try the operation again.",
		"If the problem persists, restart the application.",
		"Contact support if the issue continues.",
	}
}

// CreateErrorReport generates a comprehensive error report
func (s *Service) CreateErrorReport() *ErrorReport {
	// This would typically collect errors from memory or database
	// For now, return a basic structure
	return &ErrorReport{
		ID:        generateReportID(),
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Errors:    []*AppError{},
		Stats: ErrorStats{
			TotalErrors:      0,
			ErrorsByCode:     make(map[string]int),
			ErrorsBySeverity: make(map[string]int),
			ErrorsByCategory: make(map[string]int),
			RecoveredErrors:  0,
			RetryableErrors:  0,
		},
		System: s.getSystemInfo(),
	}
}

// RegisterTemplate registers a custom error template
func (s *Service) RegisterTemplate(template ErrorTemplate) {
	s.templates[template.Code] = template
	s.logger.Debug("Error template registered", map[string]interface{}{"code": template.Code})
}

// SetConfig updates the error handling configuration
func (s *Service) SetConfig(config *ErrorConfig) {
	s.config = config
	s.logger.Info("Error handling configuration updated")
}

// Private helper methods

// getDefaultConfig returns the default error configuration
func getDefaultConfig() *ErrorConfig {
	return &ErrorConfig{
		ShowStackTrace:      false,
		LogLevel:            "error",
		UserFriendlyErrors:  true,
		ErrorReporting:      false,
		ReportingInterval:   time.Hour,
		MaxErrorsInMemory:   100,
		EnableRecovery:      true,
		AutoRetryAttempts:   3,
		RetryDelay:          time.Second * 2,
		LocalizationEnabled: true,
		DefaultLanguage:     "en",
	}
}

// loadDefaultTemplates loads built-in error templates
func (s *Service) loadDefaultTemplates() {
	templates := []ErrorTemplate{
		{
			Code:        "FILE_NOT_FOUND",
			Title:       "File Not Found",
			Description: "The requested file could not be found.",
			Cause:       "The file may have been moved, deleted, or never existed.",
			Solution:    "Check the file path and ensure the file exists.",
			Severity:    "error",
			Category:    "file",
			Retryable:   false,
			RecoverySteps: []string{
				"Verify the file path is correct",
				"Check if the file exists in the expected location",
				"Ensure you have permission to access the file",
			},
			Translations: map[string]string{
				"en": "The requested file could not be found.",
				"es": "El archivo solicitado no pudo ser encontrado.",
				"fr": "Le fichier demandé n'a pas pu être trouvé.",
			},
		},
		{
			Code:        "NETWORK_ERROR",
			Title:       "Network Connection Failed",
			Description: "Unable to connect to the network or server.",
			Cause:       "Network connection may be unavailable or the server is not responding.",
			Solution:    "Check your network connection and try again.",
			Severity:    "error",
			Category:    "network",
			Retryable:   true,
			RecoverySteps: []string{
				"Check your internet connection",
				"Verify the server is accessible",
				"Try again in a few moments",
				"Contact your network administrator if issues persist",
			},
			Translations: map[string]string{
				"en": "Unable to connect to the network or server.",
				"es": "No se puede conectar a la red o al servidor.",
				"fr": "Impossible de se connecter au réseau ou au serveur.",
			},
		},
		{
			Code:        "PERMISSION_DENIED",
			Title:       "Access Denied",
			Description: "You don't have permission to perform this action.",
			Cause:       "Your user account may not have the required privileges.",
			Solution:    "Contact your administrator for the necessary permissions.",
			Severity:    "error",
			Category:    "user",
			Retryable:   false,
			RecoverySteps: []string{
				"Log in with an account that has the required permissions",
				"Contact your administrator for access",
				"Check if the action is allowed for your user role",
			},
			Translations: map[string]string{
				"en": "You don't have permission to perform this action.",
				"es": "No tienes permiso para realizar esta acción.",
				"fr": "Vous n'avez pas la permission d'effectuer cette action.",
			},
		},
		{
			Code:        "VALIDATION_ERROR",
			Title:       "Invalid Input",
			Description: "The provided input is not valid.",
			Cause:       "The input may contain invalid characters or format.",
			Solution:    "Please check your input and try again.",
			Severity:    "warning",
			Category:    "user",
			Retryable:   true,
			RecoverySteps: []string{
				"Check your input for typos or errors",
				"Ensure all required fields are filled",
				"Follow the specified format requirements",
			},
			Translations: map[string]string{
				"en": "The provided input is not valid.",
				"es": "La entrada proporcionada no es válida.",
				"fr": "L'entrée fournie n'est pas valide.",
			},
		},
		{
			Code:        "SYSTEM_ERROR",
			Title:       "System Error",
			Description: "An unexpected system error occurred.",
			Cause:       "The system encountered an internal error.",
			Solution:    "Please try again or contact support if the issue persists.",
			Severity:    "error",
			Category:    "system",
			Retryable:   true,
			RecoverySteps: []string{
				"Restart the application",
				"Check if you have sufficient memory and disk space",
				"Update to the latest version if available",
				"Contact technical support",
			},
			Translations: map[string]string{
				"en": "An unexpected system error occurred.",
				"es": "Ocurrió un error inesperado del sistema.",
				"fr": "Une erreur système inattendue s'est produite.",
			},
		},
		{
			Code:        "TIMEOUT_ERROR",
			Title:       "Operation Timeout",
			Description: "The operation took too long to complete.",
			Cause:       "The operation may be too complex or the system is busy.",
			Solution:    "Try again or contact support if the issue persists.",
			Severity:    "warning",
			Category:    "system",
			Retryable:   true,
			RecoverySteps: []string{
				"Try the operation again",
				"Wait for system load to decrease",
				"Break down the operation into smaller steps",
			},
			Translations: map[string]string{
				"en": "The operation took too long to complete.",
				"es": "La operación tomó demasiado tiempo en completarse.",
				"fr": "L'opération a pris trop de temps à se terminer.",
			},
		},
	}

	for _, template := range templates {
		s.templates[template.Code] = template
	}
}

// captureErrorContext captures current execution context
func (s *Service) captureErrorContext() *ErrorContext {
	context := &ErrorContext{
		Timestamp:   time.Now(),
		Environment: make(map[string]string),
	}

	// Capture stack trace if enabled
	if s.config.ShowStackTrace {
		context.StackTrace = captureStackTrace()
	}

	// Add basic environment info
	context.Environment["go_version"] = runtime.Version()
	context.Environment["num_goroutines"] = fmt.Sprintf("%d", runtime.NumGoroutine())

	return context
}

// captureStackTrace captures the current call stack
func captureStackTrace() []StackFrame {
	var frames []StackFrame
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs) // Skip this function and the caller
	pcs = pcs[:n]

	for _, pc := range pcs {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
			Package:  extractPackageName(fn.Name()),
		})
	}

	return frames
}

// logError logs an error with appropriate level
func (s *Service) logError(appError *AppError) {
	ctx := map[string]interface{}{
		"id":      appError.ID,
		"code":    appError.Code,
		"message": appError.Message,
	}
	if appError.Context != nil {
		ctx["component"] = appError.Context.Component
		ctx["operation"] = appError.Context.Operation
	}

	switch s.config.LogLevel {
	case "debug":
		s.logger.Debug("Error occurred", ctx)
	case "info":
		s.logger.Info("Error occurred", ctx)
	case "warn":
		s.logger.Warn("Error occurred", ctx)
	case "error":
		s.logger.Error("Error occurred", appError.Cause, ctx)
	default:
		s.logger.Error("Error occurred", appError.Cause, ctx)
	}
}

// extractPackageName extracts the package name from a fully qualified function name
func extractPackageName(funcName string) string {
	// funcName is like "github.com/user/repo/pkg/subpkg.FuncName"
	// We want to extract "github.com/user/repo/pkg/subpkg"
	lastSlash := strings.LastIndex(funcName, "/")
	if lastSlash == -1 {
		return ""
	}
	afterSlash := funcName[lastSlash+1:]
	dotIndex := strings.Index(afterSlash, ".")
	if dotIndex == -1 {
		return funcName[:lastSlash+1+len(afterSlash)]
	}
	return funcName[:lastSlash+1+dotIndex]
}

// generateUserFriendlyMessage generates a user-friendly error message
func (s *Service) generateUserFriendlyMessage(appError *AppError) string {
	if appError.Template != nil {
		return appError.Template.Description
	}

	// Generate a generic user-friendly message
	switch appError.Code {
	case "UNKNOWN":
		return "An unexpected error occurred. Please try again."
	case "PANIC":
		return "The application encountered a serious error and needs to restart."
	default:
		return "Something went wrong. Please check your input and try again."
	}
}

// attemptRecovery attempts to recover from an error
func (s *Service) attemptRecovery(appError *AppError) bool {
	// This is a placeholder for recovery logic
	// In a real implementation, this would try specific recovery strategies
	// based on the error type and context

	s.logger.Info("Attempting recovery", map[string]interface{}{"error_id": appError.ID, "code": appError.Code})

	// Simulate recovery attempt
	// In practice, this might retry operations, reset state, etc.
	return false // Default to false for now
}

// getSystemInfo collects system information for error reporting
func (s *Service) getSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Memory: MemoryInfo{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		Environment: map[string]string{
			"go_version":     runtime.Version(),
			"num_cpu":        fmt.Sprintf("%d", runtime.NumCPU()),
			"num_goroutines": fmt.Sprintf("%d", runtime.NumGoroutine()),
		},
	}
}

// startBackgroundTasks starts background error handling tasks
func (s *Service) startBackgroundTasks() {
	// Start error reporting task if enabled
	if s.config.ErrorReporting {
		go s.errorReportingTask()
	}
}

// errorReportingTask periodically generates error reports
func (s *Service) errorReportingTask() {
	ticker := time.NewTicker(s.config.ReportingInterval)
	defer ticker.Stop()

	for range ticker.C {
		report := s.CreateErrorReport()
		s.logger.Debug("Error report generated", map[string]interface{}{
			"report_id":    report.ID,
			"total_errors": report.Stats.TotalErrors,
		})
	}
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return utils.UniqueID("ERR_")
}

// generateReportID generates a unique report ID
func generateReportID() string {
	return utils.UniqueID("RPT_")
}

// Helper function for Go 1.21+ strings.ContainsFold fallback
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
