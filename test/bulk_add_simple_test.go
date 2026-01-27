package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBulkAddSimple tests basic bulk add functionality
func TestBulkAddSimple(t *testing.T) {
	// Create unique temp directories
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	repoPath := filepath.Join(xdgData, "ai-config", "repo")

	// Set XDG_DATA_HOME to control where the repo is created
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create one command using helper function
	cmdPath := createTestCommand(t, "test-cmd", "Test command")
	resourcesDir := filepath.Dir(filepath.Dir(cmdPath)) // Go up to parent of commands/

	// Run unified add
	output, err := runAimgr(t, "repo", "import", resourcesDir)
	if err != nil {
		t.Fatalf("Add failed: %v\nOutput: %s", err, output)
	}

	// Verify success message
	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Expected 1 added, got output: %s", output)
	}

	// Verify resource was added to repo
	if _, err := os.Stat(filepath.Join(repoPath, "commands", "test-cmd.md")); err != nil {
		t.Errorf("Resource not found in repo: %v", err)
	}
}
