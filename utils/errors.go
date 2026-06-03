package utils

import (
	"errors"
	"fmt"
)

// Exit codes matching the Node version
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitInvalidArgument  = 2
	ExitFileNotFound     = 3
	ExitPermissionDenied = 4
)

// AixError is the base error type
type AixError struct {
	Code    int
	Message string
	Cause   error
}

func (e *AixError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AixError) Unwrap() error {
	return e.Cause
}

func (e *AixError) ExitCode() int {
	return e.Code
}

// coder is implemented by all aix error types
type coder interface {
	ExitCode() int
}

// ValidationError for input validation failures
type ValidationError struct {
	AixError
	Field string
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		AixError: AixError{
			Code:    ExitInvalidArgument,
			Message: message,
		},
		Field: field,
	}
}

// FileNotFoundError for missing files
type FileNotFoundError struct {
	AixError
	Path string
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string) *FileNotFoundError {
	return &FileNotFoundError{
		AixError: AixError{
			Code:    ExitFileNotFound,
			Message: fmt.Sprintf("file not found: %s", path),
		},
		Path: path,
	}
}

// PermissionDeniedError for permission issues
type PermissionDeniedError struct {
	AixError
	Path string
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(path string) *PermissionDeniedError {
	return &PermissionDeniedError{
		AixError: AixError{
			Code:    ExitPermissionDenied,
			Message: fmt.Sprintf("permission denied: %s", path),
		},
		Path: path,
	}
}

// TokenError for token-related failures
type TokenError struct {
	AixError
}

// NewTokenError creates a token error
func NewTokenError(message string) *TokenError {
	return &TokenError{
		AixError: AixError{
			Code:    ExitGeneralError,
			Message: message,
		},
	}
}

// WrapError wraps any error in a generic AixError with ExitGeneralError.
// Use this for errors that don't fit a specific error category.
func WrapError(message string, err error) error {
	return &AixError{
		Code:    ExitGeneralError,
		Message: message,
		Cause:   err,
	}
}

// IsAixError checks if an error is an AixError
func IsAixError(err error) bool {
	var c coder
	return errors.As(err, &c)
}

// GetExitCode returns the exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var c coder
	if errors.As(err, &c) {
		return c.ExitCode()
	}

	return ExitGeneralError
}

// AsAixError unwraps to an AixError
func AsAixError(err error, target **AixError) bool {
	if err == nil {
		return false
	}

	// Try each concrete type since embedded types are not assignable to *AixError
	switch e := err.(type) {
	case *AixError:
		*target = e
		return true
	case *ValidationError:
		*target = &e.AixError
		return true
	case *FileNotFoundError:
		*target = &e.AixError
		return true
	case *PermissionDeniedError:
		*target = &e.AixError
		return true
	case *TokenError:
		*target = &e.AixError
		return true
	}

	// Try wrapped errors
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return AsAixError(unwrapped, target)
	}

	return false
}

// Unwrap returns the wrapped error
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
