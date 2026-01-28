package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDiscoverCommands_NestedCommandsDirectory tests discovery when 'commands' directory itself is nested
// This reproduces the bug where knowledge-base/dynatrace/commands/ is skipped
func TestDiscoverCommands_NestedCommandsDirectory(t *testing.T) {
	// Create temp directory structure: repo/subdir/commands/test.md
	tmpDir := t.TempDir()
	
	nestedCmdsDir := filepath.Join(tmpDir, "knowledge-base", "dynatrace", "commands", "dt")
	if err := os.MkdirAll(nestedCmdsDir, 0755); err != nil {
		t.Fatalf("Failed to create nested commands dir: %v", err)
	}
	
	cmdPath := filepath.Join(nestedCmdsDir, "test-cmd.md")
	cmdContent := `---
description: Test command in nested commands directory
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to write command file: %v", err)
	}
	
	// Discover from root (tmpDir)
	commands, err := DiscoverCommands(tmpDir, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}
	
	// Should find the command in knowledge-base/dynatrace/commands/dt/test-cmd.md
	if len(commands) == 0 {
		t.Fatalf("Expected to find command in nested 'commands' directory, but found none")
	}
	
	// Verify the command was found
	found := false
	for _, cmd := range commands {
		// Command should be named "dt/test-cmd" relative to the commands directory
		if cmd.Name == "dt/test-cmd" {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Command 'dt/test-cmd' not found. Found commands: %v", commands)
	}
}
