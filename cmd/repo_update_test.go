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

func TestRepoUpdateCmd_ExactPattern(t *testing.T) {
	// Create a temporary repo with test resources
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create test commands
	sourceDir := t.TempDir()
	cmd1Path := filepath.Join(sourceDir, "cmd1.md")
	cmd1Content := `---
description: Test command 1
version: "1.0.0"
---
# Command 1
`
	if err := os.WriteFile(cmd1Path, []byte(cmd1Content), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	cmd2Path := filepath.Join(sourceDir, "cmd2.md")
	cmd2Content := `---
description: Test command 2
version: "1.0.0"
---
# Command 2
`
	if err := os.WriteFile(cmd2Path, []byte(cmd2Content), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add commands to repo
	if err := manager.AddCommand(cmd1Path, "file://"+cmd1Path, "file"); err != nil {
		t.Fatalf("Failed to add cmd1: %v", err)
	}
	if err := manager.AddCommand(cmd2Path, "file://"+cmd2Path, "file"); err != nil {
		t.Fatalf("Failed to add cmd2: %v", err)
	}

	// Update cmd1 source file
	updatedCmd1 := `---
description: Updated command 1
version: "2.0.0"
---
# Updated Command 1
`
	if err := os.WriteFile(cmd1Path, []byte(updatedCmd1), 0644); err != nil {
		t.Fatalf("Failed to update cmd1: %v", err)
	}

	// Test updating with exact pattern
	var toUpdate []string
	matches, err := ExpandPattern(manager, "command/cmd1")
	if err != nil {
		t.Fatalf("Failed to expand pattern: %v", err)
	}
	toUpdate = append(toUpdate, matches...)

	if len(toUpdate) != 1 {
		t.Errorf("Expected 1 match, got %d", len(toUpdate))
	}
	if len(toUpdate) > 0 && toUpdate[0] != "command/cmd1" {
		t.Errorf("Expected 'command/cmd1', got '%s'", toUpdate[0])
	}

	// Update the resource
	resType, name, err := ParseResourceArg(toUpdate[0])
	if err != nil {
		t.Fatalf("Failed to parse resource arg: %v", err)
	}
	result := updateSingleResource(manager, name, resType)

	if !result.Success {
		t.Errorf("Expected successful update, got: %s", result.Message)
	}

	// Verify the resource was updated
	updatedRes, err := resource.LoadCommand(filepath.Join(repoDir, "commands", "cmd1.md"))
	if err != nil {
		t.Fatalf("Failed to load updated command: %v", err)
	}

	if updatedRes.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", updatedRes.Version)
	}
}

func TestRepoUpdateCmd_WildcardPattern(t *testing.T) {
	// Create a temporary repo with test resources
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create test commands
	sourceDir := t.TempDir()
	testCmd1Path := filepath.Join(sourceDir, "test-cmd1.md")
	testCmd1Content := `---
description: Test command 1
---
# Test Command 1
`
	if err := os.WriteFile(testCmd1Path, []byte(testCmd1Content), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	testCmd2Path := filepath.Join(sourceDir, "test-cmd2.md")
	testCmd2Content := `---
description: Test command 2
---
# Test Command 2
`
	if err := os.WriteFile(testCmd2Path, []byte(testCmd2Content), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	prodCmdPath := filepath.Join(sourceDir, "prod-cmd.md")
	prodCmdContent := `---
description: Production command
---
# Production Command
`
	if err := os.WriteFile(prodCmdPath, []byte(prodCmdContent), 0644); err != nil {
		t.Fatalf("Failed to create prod command: %v", err)
	}

	// Add commands to repo
	if err := manager.AddCommand(testCmd1Path, "file://"+testCmd1Path, "file"); err != nil {
		t.Fatalf("Failed to add test-cmd1: %v", err)
	}
	if err := manager.AddCommand(testCmd2Path, "file://"+testCmd2Path, "file"); err != nil {
		t.Fatalf("Failed to add test-cmd2: %v", err)
	}
	if err := manager.AddCommand(prodCmdPath, "file://"+prodCmdPath, "file"); err != nil {
		t.Fatalf("Failed to add prod-cmd: %v", err)
	}

	// Test updating with wildcard pattern
	matches, err := ExpandPattern(manager, "command/test*")
	if err != nil {
		t.Fatalf("Failed to expand pattern: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d: %v", len(matches), matches)
	}

	// Verify the matched resources
	matchedNames := make(map[string]bool)
	for _, match := range matches {
		_, name, err := ParseResourceArg(match)
		if err != nil {
			t.Fatalf("Failed to parse resource arg: %v", err)
		}
		matchedNames[name] = true
	}

	if !matchedNames["test-cmd1"] {
		t.Error("Expected 'test-cmd1' in matches")
	}
	if !matchedNames["test-cmd2"] {
		t.Error("Expected 'test-cmd2' in matches")
	}
	if matchedNames["prod-cmd"] {
		t.Error("Did not expect 'prod-cmd' in matches")
	}
}

func TestRepoUpdateCmd_TypeFilter(t *testing.T) {
	// Create a temporary repo with test resources
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create test skill
	skillDir := filepath.Join(t.TempDir(), "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
description: Test skill
---
# Test Skill
`
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	// Create test command
	cmdPath := filepath.Join(t.TempDir(), "test-cmd.md")
	cmdContent := `---
description: Test command
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	// Add resources to repo
	if err := manager.AddSkill(skillDir, "file://"+skillDir, "file"); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}
	if err := manager.AddCommand(cmdPath, "file://"+cmdPath, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test updating with type filter
	matches, err := ExpandPattern(manager, "skill/*")
	if err != nil {
		t.Fatalf("Failed to expand pattern: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d: %v", len(matches), matches)
	}

	if len(matches) > 0 {
		resType, name, err := ParseResourceArg(matches[0])
		if err != nil {
			t.Fatalf("Failed to parse resource arg: %v", err)
		}

		if resType != resource.Skill {
			t.Errorf("Expected resource type 'skill', got '%s'", resType)
		}
		if name != "test-skill" {
			t.Errorf("Expected name 'test-skill', got '%s'", name)
		}
	}
}

func TestRepoUpdateCmd_MultiplePatterns(t *testing.T) {
	// Create a temporary repo with test resources
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create test command
	cmdPath := filepath.Join(t.TempDir(), "cmd1.md")
	cmdContent := `---
description: Test command
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Create test agent
	agentPath := filepath.Join(t.TempDir(), "agent1.md")
	agentContent := `---
description: Test agent
---
# Test Agent
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add resources to repo
	if err := manager.AddCommand(cmdPath, "file://"+cmdPath, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}
	if err := manager.AddAgent(agentPath, "file://"+agentPath, "file"); err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Test with multiple patterns
	var toUpdate []string
	for _, pattern := range []string{"command/*", "agent/*"} {
		matches, err := ExpandPattern(manager, pattern)
		if err != nil {
			t.Fatalf("Failed to expand pattern '%s': %v", pattern, err)
		}
		toUpdate = append(toUpdate, matches...)
	}

	if len(toUpdate) != 2 {
		t.Errorf("Expected 2 matches, got %d: %v", len(toUpdate), toUpdate)
	}

	// Verify we have one command and one agent
	foundCommand := false
	foundAgent := false
	for _, match := range toUpdate {
		resType, _, err := ParseResourceArg(match)
		if err != nil {
			t.Fatalf("Failed to parse resource arg: %v", err)
		}
		if resType == resource.Command {
			foundCommand = true
		}
		if resType == resource.Agent {
			foundAgent = true
		}
	}

	if !foundCommand {
		t.Error("Expected to find command in matches")
	}
	if !foundAgent {
		t.Error("Expected to find agent in matches")
	}
}

func TestRepoUpdateCmd_DuplicatePatterns(t *testing.T) {
	// Create a temporary repo with test resources
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create test command
	cmdPath := filepath.Join(t.TempDir(), "test-cmd.md")
	cmdContent := `---
description: Test command
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Add command to repo
	if err := manager.AddCommand(cmdPath, "file://"+cmdPath, "file"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test with duplicate patterns
	var toUpdate []string
	for _, pattern := range []string{"command/test-cmd", "command/test*", "command/*"} {
		matches, err := ExpandPattern(manager, pattern)
		if err != nil {
			t.Fatalf("Failed to expand pattern '%s': %v", pattern, err)
		}
		toUpdate = append(toUpdate, matches...)
	}

	// Before deduplication, should have 3 entries
	if len(toUpdate) != 3 {
		t.Errorf("Expected 3 matches before dedup, got %d", len(toUpdate))
	}

	// Remove duplicates
	toUpdate = uniqueStrings(toUpdate)

	// After deduplication, should have 1 entry
	if len(toUpdate) != 1 {
		t.Errorf("Expected 1 match after dedup, got %d: %v", len(toUpdate), toUpdate)
	}

	if toUpdate[0] != "command/test-cmd" {
		t.Errorf("Expected 'command/test-cmd', got '%s'", toUpdate[0])
	}
}
