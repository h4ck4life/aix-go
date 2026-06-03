package utils

import (
	"errors"
	"testing"
)

func TestWrapError(t *testing.T) {
	inner := errors.New("something failed")
	wrapped := WrapError("operation failed", inner)

	aixErr, ok := wrapped.(*AixError)
	if !ok {
		t.Fatal("WrapError should return *AixError")
	}
	if aixErr.Code != ExitGeneralError {
		t.Errorf("exit code = %d, want %d", aixErr.Code, ExitGeneralError)
	}
	if aixErr.Message != "operation failed" {
		t.Errorf("message = %q, want %q", aixErr.Message, "operation failed")
	}
	if aixErr.Cause != inner {
		t.Errorf("cause = %v, want %v", aixErr.Cause, inner)
	}
	if GetExitCode(wrapped) != ExitGeneralError {
		t.Errorf("GetExitCode = %d, want %d", GetExitCode(wrapped), ExitGeneralError)
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"nil error", nil, ExitSuccess},
		{"validation error", NewValidationError("field", "bad input"), ExitInvalidArgument},
		{"file not found error", NewFileNotFoundError("/path/to/file"), ExitFileNotFound},
		{"permission denied error", NewPermissionDeniedError("/path/to/file"), ExitPermissionDenied},
		{"token error", NewTokenError("bad token"), ExitGeneralError},
		{"wrapped error", WrapError("wrapped", errors.New("inner")), ExitGeneralError},
		{"bare error", errors.New("plain error"), ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExitCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("GetExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestIsAixError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"validation", NewValidationError("f", "msg"), true},
		{"file not found", NewFileNotFoundError("/x"), true},
		{"permission", NewPermissionDeniedError("/x"), true},
		{"token", NewTokenError("msg"), true},
		{"wrapped", WrapError("msg", errors.New("inner")), true},
		{"bare", errors.New("plain"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAixError(tt.err); got != tt.want {
				t.Errorf("IsAixError() = %v, want %v", got, tt.want)
			}
		})
	}
}
