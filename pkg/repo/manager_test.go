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

func TestAddBulk(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(tmpDir string, manager *Manager) []string
		opts             BulkImportOptions
		wantAddedCount   int
		wantSkippedCount int
		wantFailedCount  int
		wantError        bool
	}{
		{
			name: "import multiple commands successfully",
			setup: func(tmpDir string, manager *Manager) []string {
				cmd1Path := filepath.Join(tmpDir, "cmd1.md")
				cmd2Path := filepath.Join(tmpDir, "cmd2.md")

				os.WriteFile(cmd1Path, []byte("---\ndescription: Command 1\n---\n"), 0644)
				os.WriteFile(cmd2Path, []byte("---\ndescription: Command 2\n---\n"), 0644)

				return []string{cmd1Path, cmd2Path}
			},
			opts:             BulkImportOptions{},
			wantAddedCount:   2,
			wantSkippedCount: 0,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "import multiple skills successfully",
			setup: func(tmpDir string, manager *Manager) []string {
				skill1Dir := filepath.Join(tmpDir, "skill1")
				skill2Dir := filepath.Join(tmpDir, "skill2")

				os.MkdirAll(skill1Dir, 0755)
				os.MkdirAll(skill2Dir, 0755)

				os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"),
					[]byte("---\nname: skill1\ndescription: Skill 1\n---\n"), 0644)
				os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"),
					[]byte("---\nname: skill2\ndescription: Skill 2\n---\n"), 0644)

				return []string{skill1Dir, skill2Dir}
			},
			opts:             BulkImportOptions{},
			wantAddedCount:   2,
			wantSkippedCount: 0,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "import mixed commands and skills",
			setup: func(tmpDir string, manager *Manager) []string {
				cmdPath := filepath.Join(tmpDir, "cmd.md")
				skillDir := filepath.Join(tmpDir, "skill")

				os.WriteFile(cmdPath, []byte("---\ndescription: Command\n---\n"), 0644)
				os.MkdirAll(skillDir, 0755)
				os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
					[]byte("---\nname: skill\ndescription: Skill\n---\n"), 0644)

				return []string{cmdPath, skillDir}
			},
			opts:             BulkImportOptions{},
			wantAddedCount:   2,
			wantSkippedCount: 0,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "conflict with skip existing",
			setup: func(tmpDir string, manager *Manager) []string {
				// Add first command
				cmd1Path := filepath.Join(tmpDir, "existing.md")
				os.WriteFile(cmd1Path, []byte("---\ndescription: Existing\n---\n"), 0644)
				manager.AddCommand(cmd1Path)

				// Try to add again
				cmd2Path := filepath.Join(tmpDir, "existing.md")
				os.WriteFile(cmd2Path, []byte("---\ndescription: Existing\n---\n"), 0644)

				return []string{cmd2Path}
			},
			opts:             BulkImportOptions{SkipExisting: true},
			wantAddedCount:   0,
			wantSkippedCount: 1,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "conflict with force",
			setup: func(tmpDir string, manager *Manager) []string {
				// Add first command
				cmd1Path := filepath.Join(tmpDir, "forced.md")
				os.WriteFile(cmd1Path, []byte("---\ndescription: Original\n---\n"), 0644)
				manager.AddCommand(cmd1Path)

				// Force overwrite
				cmd2Path := filepath.Join(tmpDir, "forced.md")
				os.WriteFile(cmd2Path, []byte("---\ndescription: Updated\n---\n"), 0644)

				return []string{cmd2Path}
			},
			opts:             BulkImportOptions{Force: true},
			wantAddedCount:   1,
			wantSkippedCount: 0,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "conflict without flags fails",
			setup: func(tmpDir string, manager *Manager) []string {
				// Add first command
				cmd1Path := filepath.Join(tmpDir, "conflict.md")
				os.WriteFile(cmd1Path, []byte("---\ndescription: Conflict\n---\n"), 0644)
				manager.AddCommand(cmd1Path)

				// Try to add again without flags
				cmd2Path := filepath.Join(tmpDir, "conflict.md")
				os.WriteFile(cmd2Path, []byte("---\ndescription: Conflict\n---\n"), 0644)

				return []string{cmd2Path}
			},
			opts:             BulkImportOptions{},
			wantAddedCount:   0,
			wantSkippedCount: 0,
			wantFailedCount:  1,
			wantError:        true,
		},
		{
			name: "dry run doesn't actually import",
			setup: func(tmpDir string, manager *Manager) []string {
				cmdPath := filepath.Join(tmpDir, "dryrun.md")
				os.WriteFile(cmdPath, []byte("---\ndescription: Dry run\n---\n"), 0644)
				return []string{cmdPath}
			},
			opts:             BulkImportOptions{DryRun: true},
			wantAddedCount:   1,
			wantSkippedCount: 0,
			wantFailedCount:  0,
			wantError:        false,
		},
		{
			name: "invalid resource fails",
			setup: func(tmpDir string, manager *Manager) []string {
				invalidPath := filepath.Join(tmpDir, "invalid.txt")
				os.WriteFile(invalidPath, []byte("not a resource"), 0644)
				return []string{invalidPath}
			},
			opts:             BulkImportOptions{},
			wantAddedCount:   0,
			wantSkippedCount: 0,
			wantFailedCount:  1,
			wantError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			repoPath := filepath.Join(tmpDir, "repo")
			manager := NewManagerWithPath(repoPath)

			sources := tt.setup(tmpDir, manager)

			result, err := manager.AddBulk(sources, tt.opts)
			if (err != nil) != tt.wantError {
				t.Errorf("AddBulk() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if len(result.Added) != tt.wantAddedCount {
				t.Errorf("AddBulk() added count = %v, want %v", len(result.Added), tt.wantAddedCount)
			}
			if len(result.Skipped) != tt.wantSkippedCount {
				t.Errorf("AddBulk() skipped count = %v, want %v", len(result.Skipped), tt.wantSkippedCount)
			}
			if len(result.Failed) != tt.wantFailedCount {
				t.Errorf("AddBulk() failed count = %v, want %v", len(result.Failed), tt.wantFailedCount)
			}

			// Verify dry run doesn't actually import
			if tt.opts.DryRun && tt.wantAddedCount > 0 {
				// Check that repo is empty
				resources, _ := manager.List(nil)
				if len(resources) != 0 {
					t.Errorf("DryRun should not import resources, but found %d resources", len(resources))
				}
			}
		})
	}
}
