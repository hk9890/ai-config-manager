package repo

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNestedCommandImport verifies that commands from nested directories
// are stored preserving their nested structure, fixing bug ai-config-manager-k78x
func TestNestedCommandImport(t *testing.T) {
	// Create test repository
	repoPath := t.TempDir()
	mgr := NewManagerWithPath(repoPath)

	// Create test nested structure
	testDir := t.TempDir()
	nestedPath := filepath.Join(testDir, "commands", "api", "v2")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create test command
	cmdPath := filepath.Join(nestedPath, "deploy.md")
	content := `---
description: Deploy API v2
---
# Deploy
`
	if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	// Import the command
	sourceURL := "file://" + cmdPath
	if err := mgr.AddCommand(cmdPath, sourceURL, "file"); err != nil {
		t.Fatalf("Failed to import command: %v", err)
	}

	// Verify it's stored in nested structure
	expectedPath := filepath.Join(repoPath, "commands", "api", "v2", "deploy.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Command not stored in nested structure. Expected: %s", expectedPath)
		
		// Check if it was stored flat (the bug)
		flatPath := filepath.Join(repoPath, "commands", "deploy.md")
		if _, err := os.Stat(flatPath); err == nil {
			t.Errorf("Command was stored flat (bug!): %s", flatPath)
		}
	}

	// Test name conflict - import another deploy.md from different folder
	dbPath := filepath.Join(testDir, "commands", "db")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}

	cmdPath2 := filepath.Join(dbPath, "deploy.md")
	content2 := `---
description: Deploy database
---
# Deploy DB
`
	if err := os.WriteFile(cmdPath2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write command 2: %v", err)
	}

	// Import second command - should not conflict
	sourceURL2 := "file://" + cmdPath2
	if err := mgr.AddCommand(cmdPath2, sourceURL2, "file"); err != nil {
		t.Fatalf("Failed to import second command: %v", err)
	}

	// Verify both exist
	expectedPath2 := filepath.Join(repoPath, "commands", "db", "deploy.md")
	if _, err := os.Stat(expectedPath2); os.IsNotExist(err) {
		t.Errorf("Second command not stored: %s", expectedPath2)
	}

	// Both should exist
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("First command was overwritten!")
	}
}
