package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log formats and writes a log message
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format the message
	message := fmt.Sprintf(format, args...)

	// Create log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	filename := filepath.Base(file)
	logEntry := fmt.Sprintf("[%s] %s %s:%d %s",
		timestamp, level.String(), filename, line, message)

	l.logger.Println(logEntry)

	// Exit on fatal
	if level == LogLevelFatal {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LogLevelFatal, format, args...)
}

// Global logger instance
var GlobalLogger = NewLogger(LogLevelInfo)

// Package-level logging functions
func Debug(format string, args ...interface{}) {
	GlobalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GlobalLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GlobalLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GlobalLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GlobalLogger.Fatal(format, args...)
}