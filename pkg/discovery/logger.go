package discovery

import "log/slog"

// Package-level logger for discovery operations
var logger *slog.Logger

// SetLogger sets the logger for the discovery package.
// This should be called by the application during initialization.
func SetLogger(l *slog.Logger) {
	logger = l
}
