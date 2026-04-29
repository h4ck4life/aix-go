package utils

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

var (
	logger *slog.Logger
	debug  bool
)

// InitLogger initializes the structured logger
func InitLogger(isDebug bool) {
	debug = isDebug
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

// Log returns the logger instance
func Log() *slog.Logger {
	if logger == nil {
		InitLogger(false)
	}
	return logger
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if debug {
		Log().Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	Log().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Log().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Log().Error(msg, args...)
}

// Timer measures elapsed time for an operation
func Timer(name string) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		Debug(fmt.Sprintf("%s completed", name), "elapsed", elapsed)
	}
}
