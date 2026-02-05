package models

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Error types for different categories of errors
const (
	// General application errors
	ErrUnknown        = "unknown"
	ErrInternal       = "internal"
	ErrValidation     = "validation"
	ErrNotFound       = "not_found"
	ErrPermission     = "permission"
	ErrTimeout        = "timeout"
	ErrEventQueueFull = "event_queue_full"

	// System-related errors
	ErrSystemInfo = "system_info"
	ErrProcess    = "process"
	ErrMemory     = "memory"
	ErrDisk       = "disk"
	ErrNetwork    = "network"
	ErrAudio      = "audio"

	// Terminal-related errors
	ErrTerminalCreate = "terminal_create"
	ErrTerminalWrite  = "terminal_write"
	ErrTerminalRead   = "terminal_read"
	ErrTerminalResize = "terminal_resize"
	ErrPTY            = "pty"
	ErrShell          = "shell"

	// Filesystem-related errors
	ErrFileNotFound   = "file_not_found"
	ErrFilePermission = "file_permission"
	ErrFileCorrupted  = "file_corrupted"
	ErrDirectory      = "directory"
	ErrSymlink        = "symlink"
	ErrWatch          = "file_watch"

	// Configuration-related errors
	ErrConfigLoad     = "config_load"
	ErrConfigSave     = "config_save"
	ErrConfigParse    = "config_parse"
	ErrConfigValidate = "config_validate"

	// Theme-related errors
	ErrThemeLoad     = "theme_load"
	ErrThemeParse    = "theme_parse"
	ErrThemeValidate = "theme_validate"
	ErrThemeNotFound = "theme_not_found"

	// Service-related errors
	ErrServiceInit        = "service_init"
	ErrServiceStart       = "service_start"
	ErrServiceStop        = "service_stop"
	ErrServiceUnavailable = "service_unavailable"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityTrace   ErrorSeverity = "trace"
	SeverityDebug   ErrorSeverity = "debug"
	SeverityInfo    ErrorSeverity = "info"
	SeverityWarning ErrorSeverity = "warning"
	SeverityError   ErrorSeverity = "error"
	SeverityFatal   ErrorSeverity = "fatal"
)

// AppError represents a structured application error
type AppError struct {
	Type        string                 `json:"type"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Severity    ErrorSeverity          `json:"severity"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Source      string                 `json:"source,omitempty"`
	UserMessage string                 `json:"user_message,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  []StackFrame           `json:"stack_trace,omitempty"`
	Cause       error                  `json:"cause,omitempty"`
}

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Error interface implementation
func (e *AppError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("[%s:%s] %s (context: %v)", e.Type, e.Code, e.Message, e.Context)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// NewError creates a new application error
func NewError(errorType, message string, cause error) *AppError {
	return &AppError{
		Type:      errorType,
		Code:      generateErrorCode(errorType),
		Message:   message,
		Severity:  SeverityError,
		Retryable: false,
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

// NewErrorWithSeverity creates a new error with specified severity
func NewErrorWithSeverity(errorType, message string, severity ErrorSeverity, cause error) *AppError {
	return &AppError{
		Type:      errorType,
		Code:      generateErrorCode(errorType),
		Message:   message,
		Severity:  severity,
		Retryable: isRetryableError(errorType, severity),
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

// NewErrorWithContext creates a new error with context
func NewErrorWithContext(errorType, message string, context map[string]interface{}, cause error) *AppError {
	return &AppError{
		Type:      errorType,
		Code:      generateErrorCode(errorType),
		Message:   message,
		Severity:  SeverityError,
		Context:   context,
		Retryable: false,
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, add to the context
	if appErr, ok := err.(*AppError); ok {
		if appErr.Context == nil {
			appErr.Context = make(map[string]interface{})
		}
		appErr.Context["wrapped_message"] = message
		appErr.Context["wrapped_at"] = time.Now()
		return appErr
	}

	return NewError(errorType, message, err)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// Error implements the error interface for ValidationError
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve *ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// NewValidationError creates a validation error
func NewValidationError(field, value, message, rule string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Rule:    rule,
	}
}

// NewValidationErrors creates multiple validation errors
func NewValidationErrors(errors []ValidationError) *ValidationErrors {
	return &ValidationErrors{
		Errors: errors,
	}
}

// ServiceError represents a service-related error
type ServiceError struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Cause     error  `json:"cause,omitempty"`
}

func (se *ServiceError) Error() string {
	return fmt.Sprintf("service %s: %s during %s", se.Service, se.ErrorType, se.Operation)
}

func (se *ServiceError) Unwrap() error {
	return se.Cause
}

// NewServiceError creates a new service error
func NewServiceError(service, operation, errorType, message string, retryable bool, cause error) *ServiceError {
	return &ServiceError{
		Service:   service,
		Operation: operation,
		ErrorType: errorType,
		Message:   message,
		Retryable: retryable,
		Cause:     cause,
	}
}

// Helper functions

// generateErrorCode generates a unique error code
func generateErrorCode(errorType string) string {
	return fmt.Sprintf("%s_%d", errorType, time.Now().UnixNano())
}

// isRetryableError determines if an error is retryable based on type and severity
func isRetryableError(errorType string, severity ErrorSeverity) bool {
	retryableTypes := map[string]bool{
		ErrTimeout:        true,
		ErrNetwork:        true,
		ErrServiceStart:   true,
		ErrServiceStop:    true,
		ErrEventQueueFull: true,
	}

	retryableSeverities := map[ErrorSeverity]bool{
		SeverityWarning: true,
		SeverityError:   false,
		SeverityFatal:   false,
	}

	if retryable, ok := retryableTypes[errorType]; ok {
		return retryable
	}

	return retryableSeverities[severity]
}

// CaptureStackTrace captures the current stack trace
func CaptureStackTrace() []StackFrame {
	var frames []StackFrame
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc) // Skip this function and the caller
	if n == 0 {
		return frames
	}

	pc = pc[:n]
	frames = make([]StackFrame, 0, n)

	for _, pcValue := range pc {
		fn := runtime.FuncForPC(pcValue)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pcValue)
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	if svcErr, ok := err.(*ServiceError); ok {
		return svcErr.Retryable
	}
	return false
}

// GetErrorType returns the type of an error
func GetErrorType(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type
	}
	if svcErr, ok := err.(*ServiceError); ok {
		return svcErr.ErrorType
	}
	return ErrUnknown
}

// GetSeverity returns the severity of an error
func GetSeverity(err error) ErrorSeverity {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Severity
	}
	return SeverityError
}

// WithStackTrace adds stack trace to an error
func WithStackTrace(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		appErr.StackTrace = CaptureStackTrace()
		return appErr
	}

	return NewError(ErrInternal, err.Error(), err)
}

// WithContext adds context to an error
func WithContext(err error, context map[string]interface{}) *AppError {
	if appErr, ok := err.(*AppError); ok {
		if appErr.Context == nil {
			appErr.Context = make(map[string]interface{})
		}
		for k, v := range context {
			appErr.Context[k] = v
		}
		return appErr
	}

	return NewErrorWithContext(ErrInternal, err.Error(), context, err)
}

// Predefined errors for common situations
var (
	ErrTerminalNotInitialized = NewError(ErrTerminalCreate, "Terminal service not initialized", nil)
	ErrInvalidConfiguration   = NewError(ErrConfigValidate, "Invalid configuration provided", nil)
	ErrServiceUnavailableVar  = NewError(ErrServiceUnavailable, "Service is not available", nil)
	ErrPermissionDenied       = NewError(ErrPermission, "Permission denied", nil)
	ErrOperationTimeout       = NewError(ErrTimeout, "Operation timed out", nil)
)
