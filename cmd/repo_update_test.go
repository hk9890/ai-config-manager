package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestUpdateFromLocalSource_MissingSource(t *testing.T) {
	// Create a temporary repo
	repoDir := t.TempDir()

	// Create test command in temp source location
	sourceDir := t.TempDir()
	cmdPath := filepath.Join(sourceDir, "test-cmd.md")
	cmdContent := `---
description: Test command
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add command to repo
	manager := repo.NewManagerWithPath(repoDir)

	sourceURL := "file://" + cmdPath
	if err := manager.AddCommand(cmdPath, sourceURL, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the source file to simulate missing source
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to remove source file: %v", err)
	}

	// Load metadata
	meta, err := manager.GetMetadata("test-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Try to update from missing source
	skipped, updateErr := updateFromLocalSource(manager, "test-cmd", resource.Command, meta)

	// Verify that the update was skipped
	if !skipped {
		t.Errorf("Expected skipped=true for missing source, got false")
	}

	// Verify error message mentions prune
	if updateErr == nil {
		t.Fatal("Expected error for missing source, got nil")
	}

	errorMsg := updateErr.Error()
	if errorMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Check that error message is helpful
	expectedPhrases := []string{"source path no longer exists", "aimgr repo prune"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(errorMsg, phrase) {
			t.Errorf("Error message should contain '%s', got: %s", phrase, errorMsg)
		}
	}
}

func TestUpdateSingleResource_MissingSource(t *testing.T) {
	// Create a temporary repo
	repoDir := t.TempDir()

	// Create test command in temp source location
	sourceDir := t.TempDir()
	cmdPath := filepath.Join(sourceDir, "test-cmd2.md")
	cmdContent := `---
description: Test command 2
---
# Test Command 2
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add command to repo
	manager := repo.NewManagerWithPath(repoDir)

	sourceURL := "file://" + cmdPath
	if err := manager.AddCommand(cmdPath, sourceURL, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the source file to simulate missing source
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to remove source file: %v", err)
	}

	// Try to update the resource
	result := updateSingleResource(manager, "test-cmd2", resource.Command)

	// Verify the result
	if result.Success {
		t.Error("Expected Success=false for missing source")
	}

	if !result.Skipped {
		t.Error("Expected Skipped=true for missing source")
	}

	if result.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestUpdateFromLocalSource_ValidSource(t *testing.T) {
	// Create a temporary repo
	repoDir := t.TempDir()

	// Create test command in temp source location
	sourceDir := t.TempDir()
	cmdPath := filepath.Join(sourceDir, "valid-cmd.md")
	cmdContent := `---
description: Valid command
version: "1.0.0"
---
# Valid Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add command to repo
	manager := repo.NewManagerWithPath(repoDir)

	sourceURL := "file://" + cmdPath
	if err := manager.AddCommand(cmdPath, sourceURL, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Update the source file
	updatedContent := `---
description: Updated valid command
version: "2.0.0"
---
# Updated Valid Command
`
	if err := os.WriteFile(cmdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update test command: %v", err)
	}

	// Load metadata
	meta, err := manager.GetMetadata("valid-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Try to update from valid source
	skipped, updateErr := updateFromLocalSource(manager, "valid-cmd", resource.Command, meta)

	// Verify that the update was NOT skipped
	if skipped {
		t.Error("Expected skipped=false for valid source")
	}

	if updateErr != nil {
		t.Errorf("Expected no error for valid source, got: %v", updateErr)
	}

	// Verify the resource was actually updated
	updatedRes, err := resource.LoadCommand(filepath.Join(repoDir, "commands", "valid-cmd.md"))
	if err != nil {
		t.Fatalf("Failed to load updated command: %v", err)
	}

	if updatedRes.Description != "Updated valid command" {
		t.Errorf("Expected description 'Updated valid command', got '%s'", updatedRes.Description)
	}

	if updatedRes.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", updatedRes.Version)
	}
}

func TestDisplayUpdateSummary_WithSkipped(t *testing.T) {
	// Create results with success, failure, and skipped
	results := []UpdateResult{
		{Name: "success-cmd", Type: resource.Command, Success: true, Skipped: false, Message: "Updated successfully"},
		{Name: "skipped-cmd", Type: resource.Command, Success: false, Skipped: true, Message: "source path no longer exists"},
		{Name: "failed-cmd", Type: resource.Command, Success: false, Skipped: false, Message: "some error"},
	}

	// Note: This test just verifies the function doesn't crash
	// In a real scenario, we'd capture stdout to verify the output
	displayUpdateSummary(results)
}

func TestDisplayUpdateSummary_Empty(t *testing.T) {
	results := []UpdateResult{}
	displayUpdateSummary(results)
}
