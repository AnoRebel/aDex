package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"aDex-UI/internal/models"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LevelTrace LogLevel = "trace"
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Logger    string                 `json:"logger,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Error     *models.AppError       `json:"error,omitempty"`
}

// LogFormat represents the output format for logs
type LogFormat string

const (
	FormatJSON LogFormat = "json"
	FormatText LogFormat = "text"
)

// Config represents logger configuration
type Config struct {
	Level      LogLevel    `json:"level"`
	Format     LogFormat   `json:"format"`
	Output     string      `json:"output"`
	File       string      `json:"file,omitempty"`
	MaxSize    int64       `json:"max_size,omitempty"`    // Max file size in bytes
	MaxBackups int         `json:"max_backups,omitempty"` // Max number of backup files
	MaxAge     int         `json:"max_age,omitempty"`     // Max age in days
	Compress   bool        `json:"compress,omitempty"`
	Console    bool        `json:"console"`
	AddSource  bool        `json:"add_source"`
}

// Logger represents a structured logger
type Logger struct {
	name       string
	config     Config
	writer     io.Writer
	mu         sync.RWMutex
	hooks      []Hook
	minLevel   LogLevel
	levelOrder map[LogLevel]int
}

// Hook defines an interface for log hooks
type Hook interface {
	Fire(entry *LogEntry) error
	Levels() []LogLevel
}

// DefaultHook provides basic hook functionality
type DefaultHook struct {
	levels []LogLevel
	fire   func(entry *LogEntry) error
}

func (h *DefaultHook) Fire(entry *LogEntry) error {
	return h.fire(entry)
}

func (h *DefaultHook) Levels() []LogLevel {
	return h.levels
}

// NewHook creates a new hook with specified levels and fire function
func NewHook(levels []LogLevel, fire func(entry *LogEntry) error) Hook {
	return &DefaultHook{
		levels: levels,
		fire:   fire,
	}
}

// New creates a new logger with the given configuration
func New(config Config) (*Logger, error) {
	logger := &Logger{
		config: config,
		hooks:  make([]Hook, 0),
		levelOrder: map[LogLevel]int{
			LevelTrace: 0,
			LevelDebug: 1,
			LevelInfo:  2,
			LevelWarn:  3,
			LevelError: 4,
			LevelFatal: 5,
		},
	}

	// Set minimum log level
	if _, ok := logger.levelOrder[config.Level]; ok {
		logger.minLevel = config.Level
	} else {
		logger.minLevel = LevelInfo
	}

	// Setup writer
	if err := logger.setupWriter(); err != nil {
		return nil, fmt.Errorf("failed to setup writer: %w", err)
	}

	return logger, nil
}

// NewLogger creates a new logger with a name
func NewLogger(name string, config Config) (*Logger, error) {
	logger, err := New(config)
	if err != nil {
		return nil, err
	}
	logger.name = name
	return logger, nil
}

// setupWriter configures the log output writer
func (l *Logger) setupWriter() error {
	var writers []io.Writer

	// Console output
	if l.config.Console || l.config.Output == "stdout" {
		writers = append(writers, os.Stdout)
	}

	// File output
	if l.config.File != "" {
		fileWriter, err := l.createFileWriter()
		if err != nil {
			return fmt.Errorf("failed to create file writer: %w", err)
		}
		writers = append(writers, fileWriter)
	}

	// Fallback to stdout if no writers configured
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	if len(writers) == 1 {
		l.writer = writers[0]
	} else {
		l.writer = io.MultiWriter(writers...)
	}

	return nil
}

// createFileWriter creates a file writer with rotation support
func (l *Logger) createFileWriter() (io.Writer, error) {
	// Ensure directory exists
	dir := filepath.Dir(l.config.File)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file
	file, err := os.OpenFile(l.config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// For now, return the file directly
	// TODO: Implement log rotation based on config
	return file, nil
}

// Log methods

// Trace logs a trace message
func (l *Logger) Trace(message string, context ...map[string]interface{}) {
	l.log(LevelTrace, message, nil, context...)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, context ...map[string]interface{}) {
	l.log(LevelDebug, message, nil, context...)
}

// Info logs an info message
func (l *Logger) Info(message string, context ...map[string]interface{}) {
	l.log(LevelInfo, message, nil, context...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, context ...map[string]interface{}) {
	l.log(LevelWarn, message, nil, context...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, context ...map[string]interface{}) {
	l.log(LevelError, message, err, context...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, context ...map[string]interface{}) {
	l.log(LevelFatal, message, err, context...)
	os.Exit(1)
}

// WithField adds a field to the context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// WithFields adds multiple fields to the context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// Create a new logger with additional default fields
	newLogger := &Logger{
		name:       l.name,
		config:     l.config,
		writer:     l.writer,
		mu:         sync.RWMutex{},
		hooks:      make([]Hook, len(l.hooks)),
		minLevel:   l.minLevel,
		levelOrder: l.levelOrder,
	}
	copy(newLogger.hooks, l.hooks)
	return newLogger
}

// WithError adds an error to the context
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err)
}

// AddHook adds a hook to the logger
func (l *Logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// GetLevel returns the current minimum log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.minLevel
}

// IsEnabled checks if the given log level is enabled
func (l *Logger) IsEnabled(level LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.levelOrder[level] >= l.levelOrder[l.minLevel]
}

// log is the core logging method
func (l *Logger) log(level LogLevel, message string, err error, context ...map[string]interface{}) {
	if !l.IsEnabled(level) {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Logger:    l.name,
	}

	// Add source information if configured
	if l.config.AddSource {
		if _, file, line, ok := runtime.Caller(3); ok {
			entry.Source = fmt.Sprintf("%s:%d", filepath.Base(file), line)
		}
	}

	// Add error information if provided
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			entry.Error = appErr
		} else {
			entry.Error = models.NewError(models.ErrInternal, err.Error(), err)
		}
	}

	// Merge context
	if len(context) > 0 {
		entry.Context = make(map[string]interface{})
		for _, ctx := range context {
			for k, v := range ctx {
				entry.Context[k] = v
			}
		}
	}

	// Write log entry
	l.writeEntry(entry)

	// Fire hooks
	l.fireHooks(entry)
}

// writeEntry writes the log entry to the configured writer
func (l *Logger) writeEntry(entry *LogEntry) {
	var output string

	switch l.config.Format {
	case FormatJSON:
		jsonData, err := json.Marshal(entry)
		if err != nil {
			output = fmt.Sprintf(`{"error":"failed to marshal log entry","message":"%s"}`, err.Error())
		} else {
			output = string(jsonData)
		}
	default: // FormatText
		output = l.formatText(entry)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintln(l.writer, output)
}

// formatText formats a log entry as text
func (l *Logger) formatText(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")

	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(string(entry.Level))))

	if entry.Logger != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Logger))
	}

	if entry.Source != "" {
		parts = append(parts, fmt.Sprintf("[%s]", entry.Source))
	}

	parts = append(parts, entry.Message)

	// Add error information
	if entry.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%s", entry.Error.Message))
	}

	// Add context
	if len(entry.Context) > 0 {
		var contextParts []string
		for k, v := range entry.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("context={%s}", strings.Join(contextParts, ", ")))
	}

	return strings.Join(parts, " ")
}

// fireHooks fires all registered hooks for the log entry
func (l *Logger) fireHooks(entry *LogEntry) {
	for _, hook := range l.hooks {
		if l.shouldFireHook(hook, entry.Level) {
			if err := hook.Fire(entry); err != nil {
				// Log hook errors to prevent logging failures
				log.Printf("Hook error: %v", err)
			}
		}
	}
}

// shouldFireHook checks if a hook should fire for the given log level
func (l *Logger) shouldFireHook(hook Hook, level LogLevel) bool {
	levels := hook.Levels()
	for _, hookLevel := range levels {
		if hookLevel == level {
			return true
		}
	}
	return false
}

// Close closes the logger and releases resources
func (l *Logger) Close() error {
	if closer, ok := l.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Global logger instance
var defaultLogger *Logger

// InitDefaultLogger initializes the default logger
func InitDefaultLogger(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefaultLogger returns the default logger
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		// Create a default logger if none exists
		defaultLogger, _ = New(Config{
			Level:   LevelInfo,
			Format:  FormatText,
			Console: true,
		})
	}
	return defaultLogger
}

// Global logging functions
func Trace(message string, context ...map[string]interface{}) {
	GetDefaultLogger().Trace(message, context...)
}

func Debug(message string, context ...map[string]interface{}) {
	GetDefaultLogger().Debug(message, context...)
}

func Info(message string, context ...map[string]interface{}) {
	GetDefaultLogger().Info(message, context...)
}

func Warn(message string, context ...map[string]interface{}) {
	GetDefaultLogger().Warn(message, context...)
}

func Error(message string, err error, context ...map[string]interface{}) {
	GetDefaultLogger().Error(message, err, context...)
}

func Fatal(message string, err error, context ...map[string]interface{}) {
	GetDefaultLogger().Fatal(message, err, context...)
}