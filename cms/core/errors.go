package core

import (
	"errors"
	"fmt"
)

// Error types for better error handling
var (
	// FileManager errors
	ErrFileNotFound      = errors.New("file not found")
	ErrDirectoryNotFound = errors.New("directory not found")
	ErrInvalidFilePath   = errors.New("invalid file path")
	ErrFileAlreadyExists = errors.New("file already exists")

	// Plugin errors
	ErrPluginFailed    = errors.New("plugin processing failed")
	ErrPluginNotFound  = errors.New("plugin not found")
	ErrInvalidPlugin   = errors.New("invalid plugin")

	// Router errors
	ErrRouteNotFound     = errors.New("route not found")
	ErrRouteExists       = errors.New("route already exists")
	ErrInvalidRoute      = errors.New("invalid route")

	// File watcher errors
	ErrWatcherNotRunning = errors.New("file watcher not running")
	ErrWatcherRunning    = errors.New("file watcher already running")

	// Configuration errors
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrConfigMissing = errors.New("configuration missing")

	// Security errors
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrForbidden      = errors.New("forbidden access")
	ErrRateLimited    = errors.New("rate limited")
	ErrInvalidInput   = errors.New("invalid input")
)

// FileManagerError wraps file manager related errors
type FileManagerError struct {
	Op   string
	Path string
	Err  error
}

func (e *FileManagerError) Error() string {
	return fmt.Sprintf("filemanager %s %s: %v", e.Op, e.Path, e.Err)
}

func (e *FileManagerError) Unwrap() error {
	return e.Err
}

// NewFileManagerError creates a new FileManagerError
func NewFileManagerError(op, path string, err error) *FileManagerError {
	return &FileManagerError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// PluginError wraps plugin related errors
type PluginError struct {
	Plugin string
	File   string
	Err    error
}

func (e *PluginError) Error() string {
	return fmt.Sprintf("plugin %s processing file %s: %v", e.Plugin, e.File, e.Err)
}

func (e *PluginError) Unwrap() error {
	return e.Err
}

// NewPluginError creates a new PluginError
func NewPluginError(plugin, file string, err error) *PluginError {
	return &PluginError{
		Plugin: plugin,
		File:   file,
		Err:    err,
	}
}

// RouterError wraps router related errors
type RouterError struct {
	Op    string
	Route string
	Err   error
}

func (e *RouterError) Error() string {
	return fmt.Sprintf("router %s %s: %v", e.Op, e.Route, e.Err)
}

func (e *RouterError) Unwrap() error {
	return e.Err
}

// NewRouterError creates a new RouterError
func NewRouterError(op, route string, err error) *RouterError {
	return &RouterError{
		Op:    op,
		Route: route,
		Err:   err,
	}
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s (value: %v): %s", e.Field, e.Value, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}