package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// TestUpdateFromLocalDirectory_Skill tests updating a skill from a local directory
// This tests the fix for the bug where local directory sources were treated as if
// they were direct resource paths, causing "directory must contain SKILL.md" errors.
func TestUpdateFromLocalDirectory_Skill(t *testing.T) {
	// Create temporary directories for repo and source
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create a skill in the source directory
	skillDir := filepath.Join(sourceDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	// Write initial SKILL.md
	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	initialContent := `---
description: My test skill v1
---
# My Skill

Initial version.
`
	if err := os.WriteFile(skillMdPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Initialize repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Import skill from source directory (not the specific skill directory)
	// This simulates the real scenario where users import from a parent directory
	sourceURL := "file://" + sourceDir
	if err := manager.AddSkill(skillDir, sourceURL, "local"); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	// Verify skill was added
	skillPath := filepath.Join(repoDir, "skills", "my-skill", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("Failed to read imported skill: %v", err)
	}
	if !strings.Contains(string(content), "Initial version") {
		t.Error("Imported skill should contain initial content")
	}

	// Modify the skill in the source directory
	updatedContent := `---
description: My test skill v2
---
# My Skill

Updated version with new content.
`
	if err := os.WriteFile(skillMdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update source skill: %v", err)
	}

	// Load metadata
	meta, err := manager.GetMetadata("my-skill", resource.Skill)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Update from local source directory
	// This is where the bug occurred - it would fail with "directory must contain SKILL.md"
	skipped, updateErr := updateFromLocalSource(manager, "my-skill", resource.Skill, meta)
	if skipped {
		t.Errorf("Update should not be skipped")
	}
	if updateErr != nil {
		t.Fatalf("Failed to update skill: %v", updateErr)
	}

	// Verify the skill was updated with new content
	updatedSkillContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("Failed to read updated skill: %v", err)
	}
	contentStr := string(updatedSkillContent)
	if !strings.Contains(contentStr, "Updated version") {
		t.Error("Updated skill should contain new content")
	}
	if !strings.Contains(contentStr, "v2") {
		t.Error("Updated skill should contain v2 description")
	}
	if strings.Contains(contentStr, "Initial version") {
		t.Error("Updated skill should not contain old content")
	}
}

// TestUpdateFromLocalDirectory_Command tests updating a command from a local directory
func TestUpdateFromLocalDirectory_Command(t *testing.T) {
	// Create temporary directories for repo and source
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create a command in the source directory
	cmdPath := filepath.Join(sourceDir, "my-command.md")
	initialContent := `---
description: My test command v1
---
# My Command

Initial version.
`
	if err := os.WriteFile(cmdPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Initialize repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Import command from source directory
	sourceURL := "file://" + sourceDir
	if err := manager.AddCommand(cmdPath, sourceURL, "local"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Verify command was added
	commandPath := filepath.Join(repoDir, "commands", "my-command.md")
	content, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatalf("Failed to read imported command: %v", err)
	}
	if !strings.Contains(string(content), "Initial version") {
		t.Error("Imported command should contain initial content")
	}

	// Modify the command in the source directory
	updatedContent := `---
description: My test command v2
---
# My Command

Updated version with new content.
`
	if err := os.WriteFile(cmdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update source command: %v", err)
	}

	// Load metadata
	meta, err := manager.GetMetadata("my-command", resource.Command)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Update from local source directory
	skipped, updateErr := updateFromLocalSource(manager, "my-command", resource.Command, meta)
	if skipped {
		t.Errorf("Update should not be skipped")
	}
	if updateErr != nil {
		t.Fatalf("Failed to update command: %v", updateErr)
	}

	// Verify the command was updated with new content
	updatedCommandContent, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatalf("Failed to read updated command: %v", err)
	}
	contentStr := string(updatedCommandContent)
	if !strings.Contains(contentStr, "Updated version") {
		t.Error("Updated command should contain new content")
	}
	if !strings.Contains(contentStr, "v2") {
		t.Error("Updated command should contain v2 description")
	}
	if strings.Contains(contentStr, "Initial version") {
		t.Error("Updated command should not contain old content")
	}
}

// TestUpdateFromLocalDirectory_Agent tests updating an agent from a local directory
func TestUpdateFromLocalDirectory_Agent(t *testing.T) {
	// Create temporary directories for repo and source
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Create an agent in the source directory
	agentPath := filepath.Join(sourceDir, "my-agent.md")
	initialContent := `---
description: My test agent v1
---
# My Agent

Initial version.
`
	if err := os.WriteFile(agentPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Initialize repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Import agent from source directory
	sourceURL := "file://" + sourceDir
	if err := manager.AddAgent(agentPath, sourceURL, "local"); err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Verify agent was added
	agentRepoPath := filepath.Join(repoDir, "agents", "my-agent.md")
	content, err := os.ReadFile(agentRepoPath)
	if err != nil {
		t.Fatalf("Failed to read imported agent: %v", err)
	}
	if !strings.Contains(string(content), "Initial version") {
		t.Error("Imported agent should contain initial content")
	}

	// Modify the agent in the source directory
	updatedContent := `---
description: My test agent v2
---
# My Agent

Updated version with new content.
`
	if err := os.WriteFile(agentPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update source agent: %v", err)
	}

	// Load metadata
	meta, err := manager.GetMetadata("my-agent", resource.Agent)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Update from local source directory
	skipped, updateErr := updateFromLocalSource(manager, "my-agent", resource.Agent, meta)
	if skipped {
		t.Errorf("Update should not be skipped")
	}
	if updateErr != nil {
		t.Fatalf("Failed to update agent: %v", updateErr)
	}

	// Verify the agent was updated with new content
	updatedAgentContent, err := os.ReadFile(agentRepoPath)
	if err != nil {
		t.Fatalf("Failed to read updated agent: %v", err)
	}
	contentStr := string(updatedAgentContent)
	if !strings.Contains(contentStr, "Updated version") {
		t.Error("Updated agent should contain new content")
	}
	if !strings.Contains(contentStr, "v2") {
		t.Error("Updated agent should contain v2 description")
	}
	if strings.Contains(contentStr, "Initial version") {
		t.Error("Updated agent should not contain old content")
	}
}
