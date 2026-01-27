package testutil

import (
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// GetFixturePath returns the absolute path to a test fixture file.
// The path is relative to the test/testdata directory.
//
// Example:
//
//	path := GetFixturePath("commands/test-command.md")
//	// Returns: /path/to/ai-config-manager/test/testdata/commands/test-command.md
func GetFixturePath(relativePath string) string {
	// Assuming fixtures are stored in test/testdata
	return filepath.Join("..", "testdata", relativePath)
}

// isGitAvailable checks if git is available in PATH
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// isOnline checks if network connectivity is available by attempting
// to connect to github.com
func isOnline() bool {
	// Try to resolve github.com to check network connectivity
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", "github.com:443", timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// SkipIfNoGit skips the test if Git is not available or if there is no
// network connectivity. This is useful for integration tests that require
// cloning remote Git repositories.
//
// Example:
//
//	func TestGitOperations(t *testing.T) {
//	    testutil.SkipIfNoGit(t)
//	    // Test code that requires Git and network
//	}
func SkipIfNoGit(t *testing.T) {
	t.Helper()

	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}
}
