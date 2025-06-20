// Package errors provides enhanced error handling for GopherStrike
package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents the type of error
type ErrorType int

const (
	// Error types
	ValidationError ErrorType = iota
	NetworkError
	FileError
	SecurityError
	ConfigError
	SystemError
	UserError
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity int

const (
	// Severity levels
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// AppError represents an application error with context
type AppError struct {
	Type       ErrorType
	Severity   ErrorSeverity
	Message    string
	Details    string
	Err        error
	StackTrace string
	Context    map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows errors.Is and errors.As to work
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new application error
func New(errType ErrorType, severity ErrorSeverity, message string) *AppError {
	return &AppError{
		Type:       errType,
		Severity:   severity,
		Message:    message,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *AppError {
	if err == nil {
		return nil
	}
	
	// If it's already an AppError, preserve the original context
	if appErr, ok := err.(*AppError); ok {
		appErr.Message = fmt.Sprintf("%s: %s", message, appErr.Message)
		return appErr
	}
	
	return &AppError{
		Type:       errType,
		Severity:   SeverityMedium,
		Message:    message,
		Err:        err,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// WithSeverity sets the severity of the error
func (e *AppError) WithSeverity(severity ErrorSeverity) *AppError {
	e.Severity = severity
	return e
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	
	var builder strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			builder.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		}
		if !more {
			break
		}
	}
	
	return builder.String()
}

// Helper functions for common error types

// ValidationFailed creates a validation error
func ValidationFailed(field string, reason string) *AppError {
	return New(ValidationError, SeverityLow, fmt.Sprintf("Validation failed for %s: %s", field, reason)).
		WithContext("field", field).
		WithContext("reason", reason)
}

// NetworkFailed creates a network error
func NetworkFailed(operation string, err error) *AppError {
	return Wrap(err, NetworkError, fmt.Sprintf("Network operation failed: %s", operation)).
		WithContext("operation", operation)
}

// FileFailed creates a file operation error
func FileFailed(operation string, path string, err error) *AppError {
	return Wrap(err, FileError, fmt.Sprintf("File operation failed: %s", operation)).
		WithContext("operation", operation).
		WithContext("path", path)
}

// SecurityFailed creates a security error
func SecurityFailed(issue string) *AppError {
	return New(SecurityError, SeverityCritical, fmt.Sprintf("Security issue: %s", issue))
}

// ConfigFailed creates a configuration error
func ConfigFailed(setting string, reason string) *AppError {
	return New(ConfigError, SeverityHigh, fmt.Sprintf("Configuration error for %s: %s", setting, reason)).
		WithContext("setting", setting)
}

// SystemFailed creates a system error
func SystemFailed(operation string, err error) *AppError {
	return Wrap(err, SystemError, fmt.Sprintf("System operation failed: %s", operation)).
		WithContext("operation", operation).
		WithSeverity(SeverityHigh)
}

// UserInputError creates a user input error
func UserInputError(input string, reason string) *AppError {
	return New(UserError, SeverityLow, fmt.Sprintf("Invalid input: %s", reason)).
		WithContext("input", input).
		WithContext("reason", reason)
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logFunc      func(error)
	panicOnCrit  bool
	showStack    bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		logFunc:     defaultLog,
		panicOnCrit: false,
		showStack:   false,
	}
}

// SetLogFunction sets the logging function
func (h *ErrorHandler) SetLogFunction(f func(error)) {
	h.logFunc = f
}

// SetPanicOnCritical sets whether to panic on critical errors
func (h *ErrorHandler) SetPanicOnCritical(panic bool) {
	h.panicOnCrit = panic
}

// SetShowStackTrace sets whether to show stack traces
func (h *ErrorHandler) SetShowStackTrace(show bool) {
	h.showStack = show
}

// Handle processes an error
func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}
	
	// Log the error
	if h.logFunc != nil {
		h.logFunc(err)
	}
	
	// Check if it's an AppError
	if appErr, ok := err.(*AppError); ok {
		// Panic on critical errors if configured
		if h.panicOnCrit && appErr.Severity == SeverityCritical {
			panic(err)
		}
	}
}

// defaultLog is the default logging function
func defaultLog(err error) {
	fmt.Printf("Error: %v\n", err)
}

// IsType checks if an error is of a specific type
func IsType(err error, errType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errType
	}
	return false
}

// GetSeverity returns the severity of an error
func GetSeverity(err error) ErrorSeverity {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Severity
	}
	return SeverityMedium
}

// GetContext returns the context of an error
func GetContext(err error) map[string]interface{} {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Context
	}
	return nil
}