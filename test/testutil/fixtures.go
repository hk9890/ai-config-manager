package testutil

import (
	"os/exec"
	"path/filepath"
	"testing"
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

// SkipIfNoGit skips the test if Git is not available.
// This is useful for integration tests that require Git operations.
//
// Example:
//
//	func TestGitOperations(t *testing.T) {
//	    testutil.SkipIfNoGit(t)
//	    // Test code that requires Git
//	}
func SkipIfNoGit(t *testing.T) {
	t.Helper()

	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}
}

// SetupIsolatedRepo creates an isolated repository for testing
// and sets AIMGR_REPO_PATH to prevent pollution of real repo.
// Returns the path to the isolated repository directory.
//
// Example:
//
//	func TestRepoOperations(t *testing.T) {
//	    repoDir := testutil.SetupIsolatedRepo(t)
//	    // Test code that uses repository
//	}
func SetupIsolatedRepo(t *testing.T) string {
	t.Helper()

	testDir := t.TempDir()
	repoDir := filepath.Join(testDir, "repo")
	t.Setenv("AIMGR_REPO_PATH", repoDir)
	return repoDir
}
