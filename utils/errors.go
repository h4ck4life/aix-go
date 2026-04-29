package utils

import "fmt"

// Exit codes matching the Node version
const (
	ExitSuccess         = 0
	ExitGeneralError    = 1
	ExitInvalidArgument = 2
	ExitFileNotFound    = 3
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

// ValidationError for input validation failures
type ValidationError struct {
	AixError
	Field   string
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

// IsAixError checks if an error is an AixError
func IsAixError(err error) bool {
	_, ok := err.(*AixError)
	return ok
}

// GetExitCode returns the exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var aixErr *AixError
	if AsAixError(err, &aixErr) {
		return aixErr.Code
	}

	return ExitGeneralError
}

// AsAixError unwraps to an AixError
func AsAixError(err error, target **AixError) bool {
	if err == nil {
		return false
	}

	// Try direct match
	if e, ok := err.(*AixError); ok {
		*target = e
		return true
	}

	// Try wrapped errors
	for {
		unwrapped := Unwrap(err)
		if unwrapped == nil {
			return false
		}
		if e, ok := unwrapped.(*AixError); ok {
			*target = e
			return true
		}
		err = unwrapped
	}
}

// Unwrap returns the wrapped error
func Unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}

	if e, ok := err.(unwrapper); ok {
		return e.Unwrap()
	}
	return nil
}
