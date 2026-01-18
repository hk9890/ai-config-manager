package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
)

func TestManagerInit(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	err := manager.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify directories were created
	commandsPath := filepath.Join(tmpDir, "commands")
	if _, err := os.Stat(commandsPath); err != nil {
		t.Errorf("commands directory was not created: %v", err)
	}

	skillsPath := filepath.Join(tmpDir, "skills")
	if _, err := os.Stat(skillsPath); err != nil {
		t.Errorf("skills directory was not created: %v", err)
	}
}

func TestAddCommand(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Create a test command
	testCmd := filepath.Join(tmpDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---

# Test Command

This is a test command.
`
	if err := os.WriteFile(testCmd, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	err := manager.AddCommand(testCmd)
	if err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	// Verify command was added
	destPath := manager.GetPath("test-cmd", resource.Command)
	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("Command was not added to repository: %v", err)
	}

	// Verify we can load it
	res, err := resource.LoadCommand(destPath)
	if err != nil {
		t.Errorf("Failed to load added command: %v", err)
	}
	if res.Name != "test-cmd" {
		t.Errorf("Command name = %v, want test-cmd", res.Name)
	}
}

func TestAddCommandConflict(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Create a test command
	testCmd := filepath.Join(tmpDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---

# Test Command
`
	if err := os.WriteFile(testCmd, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command first time - should succeed
	if err := manager.AddCommand(testCmd); err != nil {
		t.Fatalf("First AddCommand() error = %v", err)
	}

	// Add the same command again - should fail
	err := manager.AddCommand(testCmd)
	if err == nil {
		t.Error("AddCommand() expected error for duplicate command, got nil")
	}
}

func TestAddSkill(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Create a test skill
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: A test skill
---

# Test Skill

This is a test skill.
`
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add the skill
	err := manager.AddSkill(skillDir)
	if err != nil {
		t.Fatalf("AddSkill() error = %v", err)
	}

	// Verify skill was added
	destPath := manager.GetPath("test-skill", resource.Skill)
	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("Skill was not added to repository: %v", err)
	}

	// Verify SKILL.md exists
	destSkillMd := filepath.Join(destPath, "SKILL.md")
	if _, err := os.Stat(destSkillMd); err != nil {
		t.Errorf("SKILL.md was not copied: %v", err)
	}

	// Verify we can load it
	res, err := resource.LoadSkill(destPath)
	if err != nil {
		t.Errorf("Failed to load added skill: %v", err)
	}
	if res.Name != "test-skill" {
		t.Errorf("Skill name = %v, want test-skill", res.Name)
	}
}

func TestAddSkillWithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Create a test skill with subdirectories
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0755); err != nil {
		t.Fatalf("Failed to create references directory: %v", err)
	}

	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: A test skill with subdirectories
---

# Test Skill
`
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create test files in subdirectories
	scriptPath := filepath.Join(skillDir, "scripts", "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	refPath := filepath.Join(skillDir, "references", "README.md")
	if err := os.WriteFile(refPath, []byte("# Reference"), 0644); err != nil {
		t.Fatalf("Failed to create reference: %v", err)
	}

	// Add the skill
	err := manager.AddSkill(skillDir)
	if err != nil {
		t.Fatalf("AddSkill() error = %v", err)
	}

	// Verify subdirectories were copied
	destPath := manager.GetPath("test-skill", resource.Skill)
	destScriptPath := filepath.Join(destPath, "scripts", "test.sh")
	if _, err := os.Stat(destScriptPath); err != nil {
		t.Errorf("Script was not copied: %v", err)
	}

	destRefPath := filepath.Join(destPath, "references", "README.md")
	if _, err := os.Stat(destRefPath); err != nil {
		t.Errorf("Reference was not copied: %v", err)
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Add a command
	testCmd := filepath.Join(tmpDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---

# Test Command
`
	if err := os.WriteFile(testCmd, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}
	if err := manager.AddCommand(testCmd); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	// Add a skill
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: A test skill
---

# Test Skill
`
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}
	if err := manager.AddSkill(skillDir); err != nil {
		t.Fatalf("AddSkill() error = %v", err)
	}

	// List all resources
	resources, err := manager.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("List() returned %d resources, want 2", len(resources))
	}

	// List only commands
	cmdType := resource.Command
	commands, err := manager.List(&cmdType)
	if err != nil {
		t.Fatalf("List(Command) error = %v", err)
	}
	if len(commands) != 1 {
		t.Errorf("List(Command) returned %d resources, want 1", len(commands))
	}
	if commands[0].Type != resource.Command {
		t.Errorf("List(Command) returned non-command resource")
	}

	// List only skills
	skillType := resource.Skill
	skills, err := manager.List(&skillType)
	if err != nil {
		t.Fatalf("List(Skill) error = %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("List(Skill) returned %d resources, want 1", len(skills))
	}
	if skills[0].Type != resource.Skill {
		t.Errorf("List(Skill) returned non-skill resource")
	}
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Add a command
	testCmd := filepath.Join(tmpDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---

# Test Command
`
	if err := os.WriteFile(testCmd, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}
	if err := manager.AddCommand(testCmd); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	// Get the command
	res, err := manager.Get("test-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.Name != "test-cmd" {
		t.Errorf("Get() name = %v, want test-cmd", res.Name)
	}
	if res.Type != resource.Command {
		t.Errorf("Get() type = %v, want command", res.Type)
	}
}

func TestGetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	_, err := manager.Get("nonexistent", resource.Command)
	if err == nil {
		t.Error("Get() expected error for nonexistent resource, got nil")
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Add a command
	testCmd := filepath.Join(tmpDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---

# Test Command
`
	if err := os.WriteFile(testCmd, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}
	if err := manager.AddCommand(testCmd); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	// Verify it exists
	destPath := manager.GetPath("test-cmd", resource.Command)
	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("Command was not added: %v", err)
	}

	// Remove it
	err := manager.Remove("test-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(destPath); err == nil {
		t.Error("Command still exists after removal")
	}
}

func TestRemoveNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	err := manager.Remove("nonexistent", resource.Command)
	if err == nil {
		t.Error("Remove() expected error for nonexistent resource, got nil")
	}
}

func TestGetPath(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManagerWithPath(tmpDir)

	// Test command path
	cmdPath := manager.GetPath("test-cmd", resource.Command)
	expectedCmdPath := filepath.Join(tmpDir, "commands", "test-cmd.md")
	if cmdPath != expectedCmdPath {
		t.Errorf("GetPath(command) = %v, want %v", cmdPath, expectedCmdPath)
	}

	// Test skill path
	skillPath := manager.GetPath("test-skill", resource.Skill)
	expectedSkillPath := filepath.Join(tmpDir, "skills", "test-skill")
	if skillPath != expectedSkillPath {
		t.Errorf("GetPath(skill) = %v, want %v", skillPath, expectedSkillPath)
	}
}
