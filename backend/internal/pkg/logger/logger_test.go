package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		wantLevel slog.Level
	}{
		{
			name:      "development environment",
			env:       "development",
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "production environment",
			env:       "production",
			wantLevel: slog.LevelInfo,
		},
		{
			name:      "test environment",
			env:       "test",
			wantLevel: slog.LevelWarn,
		},
		{
			name:      "dev shorthand",
			env:       "dev",
			wantLevel: slog.LevelDebug,
		},
		{
			name:      "prod shorthand",
			env:       "prod",
			wantLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize(tt.env)
			if log == nil {
				t.Fatal("logger not initialized")
			}
		})
	}
}

func TestLoggingFunctions(t *testing.T) {
	// Initialize logger
	Initialize("production")

	// Test that functions don't panic
	t.Run("Debug log", func(t *testing.T) {
		Debug("debug message", "key", "value")
	})

	t.Run("Info log", func(t *testing.T) {
		Info("info message", "user_id", "123")
	})

	t.Run("Warn log", func(t *testing.T) {
		Warn("warning message", "attempt", 3)
	})

	t.Run("Error log", func(t *testing.T) {
		Error("error message", "error", "something went wrong")
	})
}

func TestLoggingWithContext(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Initialize("production")

	ctx := context.Background()
	InfoContext(ctx, "context info", "request_id", "req-123")
	ErrorContext(ctx, "context error", "error", "failed")

	// Read output
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Restore stdout
	os.Stdout = oldStdout

	// Verify JSON structure
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var logEntry map[string]any
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("failed to parse JSON log: %v", err)
		}
		if _, ok := logEntry["message"]; !ok {
			t.Error("log entry missing 'message' field")
		}
	}
}

func TestWith(t *testing.T) {
	Initialize("production")

	contextLogger := With("service", "test", "version", "1.0")
	if contextLogger == nil {
		t.Fatal("With() returned nil logger")
	}

	// Verify it's a valid logger
	if _, ok := any(contextLogger).(*slog.Logger); !ok {
		t.Error("With() did not return a *slog.Logger")
	}
}

func TestGetLogger(t *testing.T) {
	Initialize("production")

	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}

	if _, ok := any(logger).(*slog.Logger); !ok {
		t.Error("GetLogger() did not return a *slog.Logger")
	}
}

func TestJSONFormatting(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Initialize("production")
	Info("test message", "key1", "value1", "key2", 123)

	// Read output
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Restore stdout
	os.Stdout = oldStdout

	// Verify JSON format
	var logEntry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON log: %v\nOutput: %s", err, output)
	}

	// Verify fields
	if logEntry["message"] != "test message" {
		t.Errorf("expected message 'test message', got %v", logEntry["message"])
	}
	if logEntry["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got %v", logEntry["key1"])
	}
	if logEntry["key2"] != float64(123) {
		t.Errorf("expected key2 123, got %v", logEntry["key2"])
	}
}
