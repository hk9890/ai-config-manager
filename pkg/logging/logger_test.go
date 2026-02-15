package logging

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRepoLogger_Success(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v, want nil", err)
	}

	if logger == nil {
		t.Fatal("NewRepoLogger() returned nil logger")
	}

	// Verify logs directory was created
	logsDir := filepath.Join(tmpDir, "logs")
	info, err := os.Stat(logsDir)
	if err != nil {
		t.Fatalf("logs directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("logs path exists but is not a directory")
	}

	// Check directory permissions
	if info.Mode().Perm() != 0755 {
		t.Errorf("logs directory permissions = %o, want %o", info.Mode().Perm(), 0755)
	}

	// Verify log file was created
	logFile := filepath.Join(logsDir, "operations.log")
	fileInfo, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("operations.log not created: %v", err)
	}
	if fileInfo.IsDir() {
		t.Error("operations.log is a directory, want file")
	}

	// Check file permissions
	if fileInfo.Mode().Perm() != 0644 {
		t.Errorf("operations.log permissions = %o, want %o", fileInfo.Mode().Perm(), 0644)
	}
}

func TestNewRepoLogger_DirectoryAlreadyExists(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Pre-create logs directory
	logsDir := filepath.Join(tmpDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatalf("failed to pre-create logs directory: %v", err)
	}

	// Create logger - should use existing directory
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v, want nil", err)
	}

	if logger == nil {
		t.Fatal("NewRepoLogger() returned nil logger")
	}

	// Verify log file was created
	logFile := filepath.Join(logsDir, "operations.log")
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("operations.log not created in existing directory: %v", err)
	}
}

func TestNewRepoLogger_InvalidPath(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
	}{
		{
			name:     "invalid characters",
			repoPath: "/dev/null/invalid\x00path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewRepoLogger(tt.repoPath, slog.LevelDebug)
			if err == nil {
				t.Errorf("NewRepoLogger(%q) expected error, got nil", tt.repoPath)
			}
			if logger != nil {
				t.Errorf("NewRepoLogger(%q) returned non-nil logger on error", tt.repoPath)
			}
		})
	}
}

func TestNewRepoLogger_PermissionDenied(t *testing.T) {
	// Skip on systems where we can't test permissions (e.g., running as root)
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create read-only parent directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0444); err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}
	// Ensure cleanup can work by restoring permissions
	defer os.Chmod(readOnlyDir, 0755)

	// Try to create logger in read-only directory
	_, err := NewRepoLogger(readOnlyDir, slog.LevelDebug)
	if err == nil {
		t.Error("NewRepoLogger() expected error for permission denied, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to create logs directory") {
		t.Errorf("NewRepoLogger() error = %v, want error about creating logs directory", err)
	}
}

func TestNewRepoLogger_JSONFormat(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v, want nil", err)
	}

	// Write test log entries at different levels
	logger.Debug("debug message", "key1", "value1")
	logger.Info("info message", "key2", "value2")
	logger.Warn("warn message", "key3", "value3")
	logger.Error("error message", "key4", "value4")

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "operations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse log entries
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 log lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	expectedMessages := []string{"debug message", "info message", "warn message", "error message"}
	expectedKeys := []string{"key1", "key2", "key3", "key4"}
	expectedValues := []string{"value1", "value2", "value3", "value4"}

	for i, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d is not valid JSON: %v\nLine: %s", i, err, line)
			continue
		}

		// Verify standard slog fields exist
		if _, ok := entry["time"]; !ok {
			t.Errorf("line %d missing 'time' field", i)
		}
		if _, ok := entry["level"]; !ok {
			t.Errorf("line %d missing 'level' field", i)
		}
		if _, ok := entry["msg"]; !ok {
			t.Errorf("line %d missing 'msg' field", i)
		}

		// Verify level
		if level, ok := entry["level"].(string); !ok || level != expectedLevels[i] {
			t.Errorf("line %d level = %v, want %v", i, entry["level"], expectedLevels[i])
		}

		// Verify message
		if msg, ok := entry["msg"].(string); !ok || msg != expectedMessages[i] {
			t.Errorf("line %d msg = %v, want %v", i, entry["msg"], expectedMessages[i])
		}

		// Verify custom attributes
		if val, ok := entry[expectedKeys[i]].(string); !ok || val != expectedValues[i] {
			t.Errorf("line %d %s = %v, want %v", i, expectedKeys[i], entry[expectedKeys[i]], expectedValues[i])
		}
	}
}

func TestNewRepoLogger_AppendMode(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create first logger and write entry
	logger1, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() first call error = %v", err)
	}
	logger1.Info("first entry")

	// Create second logger and write entry
	logger2, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() second call error = %v", err)
	}
	logger2.Info("second entry")

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "operations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify both entries exist (not overwritten)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines (append mode), got %d", len(lines))
	}

	// Parse and verify messages
	var entry1, entry2 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("failed to parse first entry: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("failed to parse second entry: %v", err)
	}

	if entry1["msg"] != "first entry" {
		t.Errorf("first entry msg = %v, want 'first entry'", entry1["msg"])
	}
	if entry2["msg"] != "second entry" {
		t.Errorf("second entry msg = %v, want 'second entry'", entry2["msg"])
	}
}

func TestNewRepoLogger_DebugLevel(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v", err)
	}

	// Write debug level message
	logger.Debug("debug level message")

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "operations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify debug message was written
	if !strings.Contains(string(content), "debug level message") {
		t.Error("debug level message not found in log file - logger may not be using Debug level")
	}

	// Verify it's at DEBUG level
	var entry map[string]interface{}
	line := strings.TrimSpace(string(content))
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if level, ok := entry["level"].(string); !ok || level != "DEBUG" {
		t.Errorf("log level = %v, want DEBUG", entry["level"])
	}
}

func TestNewRepoLogger_NestedPath(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Use nested path that doesn't exist yet
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3")

	// Create logger - should create all parent directories
	logger, err := NewRepoLogger(nestedPath, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() with nested path error = %v", err)
	}

	if logger == nil {
		t.Fatal("NewRepoLogger() returned nil logger")
	}

	// Verify nested logs directory was created
	logsDir := filepath.Join(nestedPath, "logs")
	if _, err := os.Stat(logsDir); err != nil {
		t.Errorf("nested logs directory not created: %v", err)
	}

	// Verify log file works
	logger.Info("test message")
	logFile := filepath.Join(logsDir, "operations.log")
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("log file not created in nested path: %v", err)
	}
}

func TestNewRepoLogger_MultipleAttributes(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v", err)
	}

	// Write log with multiple attributes
	logger.Info("operation completed",
		"resource", "test-command",
		"action", "install",
		"duration_ms", 150,
		"success", true,
	)

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "operations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse log entry
	var entry map[string]interface{}
	line := strings.TrimSpace(string(content))
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	// Verify all attributes are present with correct types
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"msg", "operation completed"},
		{"resource", "test-command"},
		{"action", "install"},
		{"duration_ms", float64(150)}, // JSON numbers are float64
		{"success", true},
	}

	for _, tt := range tests {
		if val, ok := entry[tt.key]; !ok {
			t.Errorf("attribute %q not found in log entry", tt.key)
		} else if val != tt.expected {
			t.Errorf("attribute %q = %v (%T), want %v (%T)", tt.key, val, val, tt.expected, tt.expected)
		}
	}
}

func TestNewRepoLogger_WithContext(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v", err)
	}

	// Create logger with context (common slog pattern)
	contextLogger := logger.With("component", "repo", "version", "1.0.0")
	contextLogger.Info("test message", "extra", "data")

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "operations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse log entry
	var entry map[string]interface{}
	line := strings.TrimSpace(string(content))
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	// Verify context attributes and message attributes are present
	expectedAttrs := map[string]string{
		"component": "repo",
		"version":   "1.0.0",
		"extra":     "data",
	}

	for key, expectedVal := range expectedAttrs {
		if val, ok := entry[key].(string); !ok || val != expectedVal {
			t.Errorf("attribute %q = %v, want %v", key, entry[key], expectedVal)
		}
	}
}

func TestNewRepoLogger_LoggerType(t *testing.T) {
	// Create isolated temp directory
	tmpDir := t.TempDir()

	// Create logger
	logger, err := NewRepoLogger(tmpDir, slog.LevelDebug)
	if err != nil {
		t.Fatalf("NewRepoLogger() error = %v", err)
	}

	// Verify it returns *slog.Logger
	if _, ok := interface{}(logger).(*slog.Logger); !ok {
		t.Errorf("NewRepoLogger() returned %T, want *slog.Logger", logger)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel slog.Level
		wantErr   bool
	}{
		{
			name:      "debug lowercase",
			input:     "debug",
			wantLevel: slog.LevelDebug,
			wantErr:   false,
		},
		{
			name:      "debug uppercase",
			input:     "DEBUG",
			wantLevel: slog.LevelDebug,
			wantErr:   false,
		},
		{
			name:      "info lowercase",
			input:     "info",
			wantLevel: slog.LevelInfo,
			wantErr:   false,
		},
		{
			name:      "info uppercase",
			input:     "INFO",
			wantLevel: slog.LevelInfo,
			wantErr:   false,
		},
		{
			name:      "warn lowercase",
			input:     "warn",
			wantLevel: slog.LevelWarn,
			wantErr:   false,
		},
		{
			name:      "warn uppercase",
			input:     "WARN",
			wantLevel: slog.LevelWarn,
			wantErr:   false,
		},
		{
			name:      "warning lowercase",
			input:     "warning",
			wantLevel: slog.LevelWarn,
			wantErr:   false,
		},
		{
			name:      "warning uppercase",
			input:     "WARNING",
			wantLevel: slog.LevelWarn,
			wantErr:   false,
		},
		{
			name:      "error lowercase",
			input:     "error",
			wantLevel: slog.LevelError,
			wantErr:   false,
		},
		{
			name:      "error uppercase",
			input:     "ERROR",
			wantLevel: slog.LevelError,
			wantErr:   false,
		},
		{
			name:      "invalid level",
			input:     "invalid",
			wantLevel: slog.LevelInfo, // Returns default on error
			wantErr:   true,
		},
		{
			name:      "empty string",
			input:     "",
			wantLevel: slog.LevelInfo, // Returns default on error
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevel, err := ParseLogLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if gotLevel != tt.wantLevel {
				t.Errorf("ParseLogLevel(%q) = %v, want %v", tt.input, gotLevel, tt.wantLevel)
			}
		})
	}
}
