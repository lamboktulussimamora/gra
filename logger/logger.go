// Package logger provides logging functionality.
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	// DEBUG level for detailed debugging
	DEBUG LogLevel = iota
	// INFO level for general information
	INFO
	// WARN level for warnings
	WARN
	// ERROR level for errors
	ERROR
	// FATAL level for fatal errors
	FATAL
)

// Logger provides logging functionality
type Logger struct {
	level  LogLevel
	prefix string
	logger *log.Logger
}

var (
	defaultLogger *Logger
	once          sync.Once
	osExit        = os.Exit // Variable for overriding os.Exit in tests
)

// Get returns the default logger
func Get() *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			level:  INFO,
			prefix: "",
			logger: log.New(os.Stdout, "", log.LstdFlags),
		}
	})
	return defaultLogger
}

// New creates a new logger with the specified prefix
func New(prefix string) *Logger {
	return &Logger{
		level:  INFO,
		prefix: prefix,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetPrefix sets the log prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// log logs a message at the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	var levelStr string
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	case FATAL:
		levelStr = "FATAL"
	}

	prefix := ""
	if l.prefix != "" {
		prefix = "[" + l.prefix + "] "
	}

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s %s%s: %s", timestamp, prefix, levelStr, message)

	if level == FATAL {
		osExit(1)
	}
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, "%s", fmt.Sprint(args...))
}

// Debugf logs a formatted message at DEBUG level
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs a message at INFO level
func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, "%s", fmt.Sprint(args...))
}

// Infof logs a formatted message at INFO level
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(args ...interface{}) {
	l.log(WARN, "%s", fmt.Sprint(args...))
}

// Warnf logs a formatted message at WARN level
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, "%s", fmt.Sprint(args...))
}

// Errorf logs a formatted message at ERROR level
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a message at FATAL level and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.log(FATAL, "%s", fmt.Sprint(args...))
}

// Fatalf logs a formatted message at FATAL level and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}
