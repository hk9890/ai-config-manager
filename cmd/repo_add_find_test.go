package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindAgentFile_SkipsToolDirectories tests that findAgentFile skips .claude, .opencode, .github directories
// This ensures that repo import correctly identifies source resources vs installed resources
func TestFindAgentFile_SkipsToolDirectories(t *testing.T) {
	// Create a temporary directory structure that mimics a project with tool installations
	tempDir := t.TempDir()

	// Create source agents directory
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a real source agent
	sourceAgentPath := filepath.Join(agentsDir, "my-agent.md")
	sourceAgentContent := `---
description: Real source agent
---
# My Agent

This is the real source agent.
`
	if err := os.WriteFile(sourceAgentPath, []byte(sourceAgentContent), 0644); err != nil {
		t.Fatalf("Failed to create source agent: %v", err)
	}

	// Create .claude directory with installed agents (should be skipped)
	claudeAgentsDir := filepath.Join(tempDir, ".claude", "agents")
	if err := os.MkdirAll(claudeAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude/agents directory: %v", err)
	}

	claudeAgentPath := filepath.Join(claudeAgentsDir, "my-agent.md")
	claudeAgentContent := `---
description: Installed agent in .claude
---
# My Agent

This is an installed agent (should be ignored).
`
	if err := os.WriteFile(claudeAgentPath, []byte(claudeAgentContent), 0644); err != nil {
		t.Fatalf("Failed to create .claude agent: %v", err)
	}

	// Create .opencode directory with installed agents (should be skipped)
	opencodeAgentsDir := filepath.Join(tempDir, ".opencode", "agents")
	if err := os.MkdirAll(opencodeAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create .opencode/agents directory: %v", err)
	}

	opencodeAgentPath := filepath.Join(opencodeAgentsDir, "my-agent.md")
	if err := os.WriteFile(opencodeAgentPath, []byte(claudeAgentContent), 0644); err != nil {
		t.Fatalf("Failed to create .opencode agent: %v", err)
	}

	// Test: findAgentFile should find the source agent, not the installed ones
	foundPath, err := findAgentFile(tempDir, "my-agent")
	if err != nil {
		t.Fatalf("findAgentFile failed: %v", err)
	}

	// Verify it found the source agent, not the installed one
	if foundPath != sourceAgentPath {
		t.Errorf("Expected to find source agent at %s, but got %s", sourceAgentPath, foundPath)
	}

	// Verify it's not from .claude or .opencode
	if filepath.Dir(filepath.Dir(foundPath)) == filepath.Join(tempDir, ".claude") {
		t.Error("findAgentFile should not return agents from .claude directory")
	}
	if filepath.Dir(filepath.Dir(foundPath)) == filepath.Join(tempDir, ".opencode") {
		t.Error("findAgentFile should not return agents from .opencode directory")
	}
}

// TestFindCommandFile_SkipsToolDirectories tests that findCommandFile skips tool directories
func TestFindCommandFile_SkipsToolDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create source commands directory
	commandsDir := filepath.Join(tempDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create a real source command
	sourceCommandPath := filepath.Join(commandsDir, "my-command.md")
	sourceCommandContent := `---
description: Real source command
---
# My Command

This is the real source command.
`
	if err := os.WriteFile(sourceCommandPath, []byte(sourceCommandContent), 0644); err != nil {
		t.Fatalf("Failed to create source command: %v", err)
	}

	// Create .claude directory with installed command (should be skipped)
	claudeCommandsDir := filepath.Join(tempDir, ".claude", "commands")
	if err := os.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude/commands directory: %v", err)
	}

	claudeCommandPath := filepath.Join(claudeCommandsDir, "my-command.md")
	if err := os.WriteFile(claudeCommandPath, []byte(sourceCommandContent), 0644); err != nil {
		t.Fatalf("Failed to create .claude command: %v", err)
	}

	// Test: findCommandFile should find the source command
	foundPath, err := findCommandFile(tempDir, "my-command")
	if err != nil {
		t.Fatalf("findCommandFile failed: %v", err)
	}

	if foundPath != sourceCommandPath {
		t.Errorf("Expected to find source command at %s, but got %s", sourceCommandPath, foundPath)
	}
}

// TestFindSkillDir_SkipsToolDirectories tests that findSkillDir skips tool directories
func TestFindSkillDir_SkipsToolDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create source skills directory
	skillsDir := filepath.Join(tempDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	// Create a real source skill
	sourceSkillDir := filepath.Join(skillsDir, "my-skill")
	if err := os.MkdirAll(sourceSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create source skill directory: %v", err)
	}

	sourceSkillMdPath := filepath.Join(sourceSkillDir, "SKILL.md")
	sourceSkillContent := `---
description: Real source skill
---
# My Skill

This is the real source skill.
`
	if err := os.WriteFile(sourceSkillMdPath, []byte(sourceSkillContent), 0644); err != nil {
		t.Fatalf("Failed to create source SKILL.md: %v", err)
	}

	// Create .claude directory with installed skill (should be skipped)
	claudeSkillsDir := filepath.Join(tempDir, ".claude", "skills", "my-skill")
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude/skills directory: %v", err)
	}

	claudeSkillMdPath := filepath.Join(claudeSkillsDir, "SKILL.md")
	if err := os.WriteFile(claudeSkillMdPath, []byte(sourceSkillContent), 0644); err != nil {
		t.Fatalf("Failed to create .claude SKILL.md: %v", err)
	}

	// Test: findSkillDir should find the source skill
	foundPath, err := findSkillDir(tempDir, "my-skill")
	if err != nil {
		t.Fatalf("findSkillDir failed: %v", err)
	}

	if foundPath != sourceSkillDir {
		t.Errorf("Expected to find source skill at %s, but got %s", sourceSkillDir, foundPath)
	}
}

// TestFindAgentFile_AfterRemove tests the specific scenario:
// When a source directory contains .claude/agents/, findAgentFile should only find
// the source agent and not any installed files
func TestFindAgentFile_AfterRemove(t *testing.T) {
	tempDir := t.TempDir()

	// Create source agent
	agentsDir := filepath.Join(tempDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	sourceAgentPath := filepath.Join(agentsDir, "my-agent.md")
	sourceAgentContent := `---
description: Real source agent
---
# My Agent

Source agent.
`
	if err := os.WriteFile(sourceAgentPath, []byte(sourceAgentContent), 0644); err != nil {
		t.Fatalf("Failed to create source agent: %v", err)
	}

	// Create .claude directory (but don't create the agent file - simulating it being deleted)
	claudeAgentsDir := filepath.Join(tempDir, ".claude", "agents")
	if err := os.MkdirAll(claudeAgentsDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude/agents directory: %v", err)
	}

	// Test: findAgentFile should successfully find the source agent
	// even though .claude/agents/ directory exists (but is empty)
	foundPath, err := findAgentFile(tempDir, "my-agent")
	if err != nil {
		t.Fatalf("findAgentFile failed: %v", err)
	}

	if foundPath != sourceAgentPath {
		t.Errorf("Expected to find source agent at %s, but got %s", sourceAgentPath, foundPath)
	}
}
