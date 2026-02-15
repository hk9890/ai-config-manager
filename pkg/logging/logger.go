// Package logging provides structured JSON logging for repository operations.
//
// This package uses Go's standard library log/slog to provide structured
// logging to files without any console output. All logs are written in JSON
// format to {repoPath}/logs/operations.log for easy parsing and analysis.
//
// Example usage:
//
//	logger, err := logging.NewRepoLogger("/path/to/repo")
//	if err != nil {
//	    return fmt.Errorf("failed to create logger: %w", err)
//	}
//	logger.Info("operation completed", "resource", "my-command", "action", "install")
package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// NewRepoLogger creates a new structured JSON logger that writes to {repoPath}/logs/operations.log.
//
// The logger:
//   - Writes JSON formatted log entries (one per line)
//   - Creates the logs directory if it doesn't exist (permissions: 0755)
//   - Opens/creates the log file in append mode (permissions: 0644)
//   - Uses the specified level as minimum logging level
//   - Writes to file only (no console output)
//
// Returns an error if the log directory cannot be created or the log file
// cannot be opened for writing.
func NewRepoLogger(repoPath string, level slog.Level) (*slog.Logger, error) {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(repoPath, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open/create log file in append mode
	logFilePath := filepath.Join(logsDir, "operations.log")
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create JSON handler with specified level
	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: level,
	})

	// Create and return logger
	logger := slog.New(handler)
	return logger, nil
}

// ParseLogLevel parses a log level string and returns the corresponding slog.Level.
// Valid levels: "debug", "info", "warn", "error" (case-insensitive).
// Returns an error if the level string is invalid.
func ParseLogLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "debug", "DEBUG":
		return slog.LevelDebug, nil
	case "info", "INFO":
		return slog.LevelInfo, nil
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn, nil
	case "error", "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %q (valid levels: debug, info, warn, error)", levelStr)
	}
}
