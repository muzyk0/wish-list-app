// Package logger provides structured logging using Go's standard log/slog package.
//
// Usage:
//
//	logger.Info("user logged in", "user_id", userID, "email", email)
//	logger.Error("failed to create item", "error", err, "item_id", itemID)
//	logger.Warn("retry attempt", "attempt", retryCount, "max_retries", maxRetries)
//
// The package automatically initializes a global logger with JSON formatting
// and appropriate log levels based on the environment.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var log *slog.Logger

// Initialize sets up the global logger with JSON formatting.
// Call this once during application startup.
func Initialize(env string) {
	level := slog.LevelInfo

	// Set log level based on environment
	switch strings.ToLower(env) {
	case "development", "dev":
		level = slog.LevelDebug
	case "production", "prod":
		level = slog.LevelInfo
	case "test":
		level = slog.LevelWarn
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize attribute names if needed
			// For example, rename "msg" to "message"
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	log = slog.New(handler)
	slog.SetDefault(log)
}

// Debug logs a debug-level message with optional key-value pairs.
func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

// DebugContext logs a debug-level message with context.
func DebugContext(ctx context.Context, msg string, args ...any) {
	log.DebugContext(ctx, msg, args...)
}

// Info logs an info-level message with optional key-value pairs.
func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

// InfoContext logs an info-level message with context.
func InfoContext(ctx context.Context, msg string, args ...any) {
	log.InfoContext(ctx, msg, args...)
}

// Warn logs a warning-level message with optional key-value pairs.
func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

// WarnContext logs a warning-level message with context.
func WarnContext(ctx context.Context, msg string, args ...any) {
	log.WarnContext(ctx, msg, args...)
}

// Error logs an error-level message with optional key-value pairs.
func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

// ErrorContext logs an error-level message with context.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	log.ErrorContext(ctx, msg, args...)
}

// With returns a new logger with the given key-value pairs added to all log entries.
// Useful for adding consistent context across multiple log calls.
//
// Example:
//
//	userLogger := logger.With("user_id", userID, "session_id", sessionID)
//	userLogger.Info("processing request")
//	userLogger.Error("request failed", "error", err)
func With(args ...any) *slog.Logger {
	return log.With(args...)
}

// GetLogger returns the underlying slog.Logger instance.
// Use this when you need direct access to slog.Logger methods.
func GetLogger() *slog.Logger {
	return log
}
