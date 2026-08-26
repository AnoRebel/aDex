package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents the log level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger represents a logger instance
type Logger struct {
	level     LogLevel
	logger    *log.Logger
	file      *os.File
	useColors bool
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:     level,
		logger:    log.New(output, "", log.LstdFlags),
		useColors: output == os.Stdout || output == os.Stderr,
	}
}

// NewFileLogger creates a new file logger
func NewFileLogger(level LogLevel, filename string) (*Logger, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		level:  level,
		logger: log.New(file, "", log.LstdFlags),
		file:   file,
	}, nil
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetFlags sets the logger flags
func (l *Logger) SetFlags(flags int) {
	l.logger.SetFlags(flags)
}

// SetPrefix sets the logger prefix
func (l *Logger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log("DEBUG", msg, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		l.log("INFO", msg, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		l.log("WARN", msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		l.log("ERROR", msg, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	if l.level <= FATAL {
		l.log("FATAL", msg, args...)
		os.Exit(1)
	}
}

// log logs a message with the given level
func (l *Logger) log(level, msg string, args ...interface{}) {
	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = filepath.Base(file)
	}

	// Format message
	formattedMsg := fmt.Sprintf(msg, args...)

	// Add color if enabled
	if l.useColors {
		levelColor := l.getColorForLevel(level)
		resetColor := "\033[0m"
		formattedMsg = fmt.Sprintf("%s[%s]%s %s:%d - %s",
			levelColor, level, resetColor, file, line, formattedMsg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s:%d - %s", level, file, line, formattedMsg)
	}

	// Log the message
	l.logger.Println(formattedMsg)
}

// getColorForLevel returns the color code for the given log level
func (l *Logger) getColorForLevel(level string) string {
	switch level {
	case "DEBUG":
		return "\033[36m" // Cyan
	case "INFO":
		return "\033[32m" // Green
	case "WARN":
		return "\033[33m" // Yellow
	case "ERROR":
		return "\033[31m" // Red
	case "FATAL":
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // White
	}
}

// Close closes the logger and any associated files
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// ParseLogLevel parses a log level string
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(context map[string]interface{}) *Logger {
	prefix := ""
	for key, value := range context {
		prefix += fmt.Sprintf("%s=%v ", key, value)
	}

	newLogger := *l
	newLogger.logger = log.New(l.logger.Writer(), prefix, l.logger.Flags())
	return &newLogger
}

// WithField creates a new logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithContext(map[string]interface{}{key: value})
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration, clientIP string) {
	msg := fmt.Sprintf("%s %s %d %v %s", method, path, statusCode, duration, clientIP)

	if statusCode >= 500 {
		l.Error("%s", msg)
	} else if statusCode >= 400 {
		l.Warn("%s", msg)
	} else {
		l.Info("%s", msg)
	}
}

// LogError logs an error with stack trace
func (l *Logger) LogError(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}

	// Get stack trace
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	// Format error message
	errorMsg := fmt.Sprintf(msg, args...)
	fullMsg := fmt.Sprintf("%s: %v\nStack trace:\n%s", errorMsg, err, string(buf))

	l.Error("%s", fullMsg)
}

// Global logger instance
var DefaultLogger = NewLogger(INFO, os.Stdout)

// Convenience functions for global logger
func Debug(msg string, args ...interface{}) {
	DefaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	DefaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	DefaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	DefaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	DefaultLogger.Fatal(msg, args...)
}
