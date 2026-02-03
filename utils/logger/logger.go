package logger

import (
	"fmt"
	"log"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for the log level (optional, for CLI)
func (l LogLevel) Color() string {
	switch l {
	case DEBUG:
		return "\033[36m" // Cyan
	case INFO:
		return "\033[32m" // Green
	case WARN:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	default:
		return "\033[0m"
	}
}

const colorReset = "\033[0m"

// Logger wraps the standard logger with level-based logging
type Logger struct {
	logger      *log.Logger
	component   string
	minLevel    LogLevel
	useColors   bool
	development bool
}

// New creates a new Logger instance
func New(component string, minLevel LogLevel) *Logger {
	return &Logger{
		logger:      log.New(os.Stdout, "", log.LstdFlags),
		component:   component,
		minLevel:    minLevel,
		useColors:   false,
		development: false,
	}
}

// WithColors enables colored output (useful for CLI)
func (l *Logger) WithColors(enabled bool) *Logger {
	l.useColors = enabled
	return l
}

// WithDevelopmentMode enables development mode (more verbose)
func (l *Logger) WithDevelopmentMode(enabled bool) *Logger {
	l.development = enabled
	return l
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	message := fmt.Sprintf(format, args...)
	levelStr := level.String()

	if l.useColors {
		l.logger.Printf("%s[%s]%s %s: %s",
			level.Color(),
			levelStr,
			colorReset,
			l.component,
			message,
		)
	} else {
		l.logger.Printf("[%s] %s: %s", levelStr, l.component, message)
	}
}

// Debug logs a debug message (only in development mode)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.development {
		l.log(DEBUG, format, args...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// WithField returns a formatted string with key-value pair
func WithField(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}

// WithFields returns a formatted string with multiple key-value pairs
func WithFields(fields map[string]interface{}) string {
	result := ""
	for k, v := range fields {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s=%v", k, v)
	}
	return result
}
