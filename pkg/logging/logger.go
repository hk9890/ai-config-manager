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
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// lazyLogWriter defers log directory/file creation until first write.
// This keeps manager construction read-only for commands that fail validation
// before any log event is emitted.
type lazyLogWriter struct {
	repoPath string

	mu   sync.Mutex
	file io.Writer
}

func newLazyLogWriter(repoPath string) *lazyLogWriter {
	return &lazyLogWriter{repoPath: repoPath}
}

func (w *lazyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		logsDir := filepath.Join(w.repoPath, "logs")
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return 0, fmt.Errorf("failed to create logs directory: %w", err)
		}

		logFilePath := filepath.Join(logsDir, "operations.log")
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, fmt.Errorf("failed to open log file: %w", err)
		}

		w.file = logFile
	}

	return w.file.Write(p)
}

// NewRepoLogger creates a new structured JSON logger that writes to {repoPath}/logs/operations.log.
//
// The logger:
//   - Writes JSON formatted log entries (one per line)
//   - Lazily creates the logs directory and file on first write
//   - Uses the specified level as minimum logging level
//   - Writes to file only (no console output)
//
// Returns an error if the repo path is inaccessible.
func NewRepoLogger(repoPath string, level slog.Level) (*slog.Logger, error) {
	// Verify the repo path exists before creating subdirectories
	info, err := os.Stat(repoPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("repo path does not exist: %s", repoPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access repo path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repo path is not a directory: %s", repoPath)
	}

	// Create JSON handler with specified level
	handler := slog.NewJSONHandler(newLazyLogWriter(repoPath), &slog.HandlerOptions{
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
