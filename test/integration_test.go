package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hans-m-leitner/ai-config-manager/pkg/config"
	"github.com/hans-m-leitner/ai-config-manager/pkg/install"
	"github.com/hans-m-leitner/ai-config-manager/pkg/repo"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
	"github.com/hans-m-leitner/ai-config-manager/pkg/tools"
)

// TestCompleteWorkflow tests the complete workflow: add -> list -> install -> remove
func TestCompleteWorkflow(t *testing.T) {
	// Create temporary directories for repo and project
	repoDir := t.TempDir()
	projectDir := t.TempDir()

	// Create test resources
	testCmdPath := filepath.Join(repoDir, "test-cmd.md")
	cmdContent := `---
description: A test command for integration testing
agent: test
model: test-model
---

# Test Command

This is a test command for integration testing.
`
	if err := os.WriteFile(testCmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	skillDir := filepath.Join(repoDir, "test-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: A test skill for integration testing
license: MIT
metadata:
  author: test
  version: "1.0.0"
---

# Test Skill

This is a test skill for integration testing.
`
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add a script to the skill
	scriptPath := filepath.Join(skillDir, "scripts", "test.sh")
	scriptContent := `#!/bin/bash
echo "Test script"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Step 1: Create manager and add resources
	t.Log("Step 1: Adding resources to repository")
	manager := repo.NewManagerWithPath(repoDir)

	if err := manager.AddCommand(testCmdPath); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	if err := manager.AddSkill(skillDir); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	// Step 2: List resources
	t.Log("Step 2: Listing resources")
	resources, err := manager.List(nil)
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	// Verify command exists
	cmdRes, err := manager.Get("test-cmd", resource.Command)
	if err != nil {
		t.Errorf("Failed to get command: %v", err)
	}
	if cmdRes.Name != "test-cmd" {
		t.Errorf("Command name = %v, want test-cmd", cmdRes.Name)
	}

	// Verify skill exists
	skillRes, err := manager.Get("test-skill", resource.Skill)
	if err != nil {
		t.Errorf("Failed to get skill: %v", err)
	}
	if skillRes.Name != "test-skill" {
		t.Errorf("Skill name = %v, want test-skill", skillRes.Name)
	}

	// Step 3: Install resources in project
	t.Log("Step 3: Installing resources")
	installer, err := install.NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	if err := installer.InstallCommand("test-cmd", manager); err != nil {
		t.Fatalf("Failed to install command: %v", err)
	}

	if err := installer.InstallSkill("test-skill", manager); err != nil {
		t.Fatalf("Failed to install skill: %v", err)
	}

	// Verify installations
	if !installer.IsInstalled("test-cmd", resource.Command) {
		t.Error("Command should be installed")
	}
	if !installer.IsInstalled("test-skill", resource.Skill) {
		t.Error("Skill should be installed")
	}

	// Verify symlinks were created correctly
	cmdSymlink := filepath.Join(projectDir, ".claude", "commands", "test-cmd.md")
	if _, err := os.Lstat(cmdSymlink); err != nil {
		t.Errorf("Command symlink not created: %v", err)
	}

	skillSymlink := filepath.Join(projectDir, ".claude", "skills", "test-skill")
	if _, err := os.Lstat(skillSymlink); err != nil {
		t.Errorf("Skill symlink not created: %v", err)
	}

	// Step 4: List installed resources
	t.Log("Step 4: Listing installed resources")
	installedResources, err := installer.List()
	if err != nil {
		t.Fatalf("Failed to list installed resources: %v", err)
	}

	if len(installedResources) != 2 {
		t.Errorf("Expected 2 installed resources, got %d", len(installedResources))
	}

	// Step 5: Uninstall resources
	t.Log("Step 5: Uninstalling resources")
	if err := installer.Uninstall("test-cmd", resource.Command); err != nil {
		t.Errorf("Failed to uninstall command: %v", err)
	}

	if err := installer.Uninstall("test-skill", resource.Skill); err != nil {
		t.Errorf("Failed to uninstall skill: %v", err)
	}

	// Verify uninstallations
	if installer.IsInstalled("test-cmd", resource.Command) {
		t.Error("Command should be uninstalled")
	}
	if installer.IsInstalled("test-skill", resource.Skill) {
		t.Error("Skill should be uninstalled")
	}

	// Step 6: Remove resources from repository
	t.Log("Step 6: Removing resources from repository")
	if err := manager.Remove("test-cmd", resource.Command); err != nil {
		t.Errorf("Failed to remove command: %v", err)
	}

	if err := manager.Remove("test-skill", resource.Skill); err != nil {
		t.Errorf("Failed to remove skill: %v", err)
	}

	// Verify removals
	_, err = manager.Get("test-cmd", resource.Command)
	if err == nil {
		t.Error("Command should be removed from repository")
	}

	_, err = manager.Get("test-skill", resource.Skill)
	if err == nil {
		t.Error("Skill should be removed from repository")
	}

	t.Log("Complete workflow test passed!")
}

// TestErrorCases tests various error scenarios
func TestErrorCases(t *testing.T) {
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	t.Run("add invalid command", func(t *testing.T) {
		// Try to add a non-existent file
		err := manager.AddCommand("/nonexistent/file.md")
		if err == nil {
			t.Error("Expected error for non-existent command, got nil")
		}
	})

	t.Run("add invalid skill", func(t *testing.T) {
		// Try to add a non-existent directory
		err := manager.AddSkill("/nonexistent/skill")
		if err == nil {
			t.Error("Expected error for non-existent skill, got nil")
		}
	})

	t.Run("get non-existent resource", func(t *testing.T) {
		_, err := manager.Get("nonexistent", resource.Command)
		if err == nil {
			t.Error("Expected error for non-existent resource, got nil")
		}
	})

	t.Run("remove non-existent resource", func(t *testing.T) {
		err := manager.Remove("nonexistent", resource.Command)
		if err == nil {
			t.Error("Expected error for non-existent resource, got nil")
		}
	})

	t.Run("install non-existent resource", func(t *testing.T) {
		projectDir := t.TempDir()
		installer, err := install.NewInstaller(projectDir, tools.Claude)
		if err != nil {
			t.Fatalf("Failed to create installer: %v", err)
		}

		err = installer.InstallCommand("nonexistent", manager)
		if err == nil {
			t.Error("Expected error for non-existent resource, got nil")
		}
	})

	t.Run("add duplicate resource", func(t *testing.T) {
		// Create a test command
		testCmdPath := filepath.Join(repoDir, "dup-cmd.md")
		cmdContent := `---
description: A duplicate command
---
# Duplicate
`
		if err := os.WriteFile(testCmdPath, []byte(cmdContent), 0644); err != nil {
			t.Fatalf("Failed to create test command: %v", err)
		}

		// Add it once
		if err := manager.AddCommand(testCmdPath); err != nil {
			t.Fatalf("Failed to add command: %v", err)
		}

		// Try to add again - should fail
		err := manager.AddCommand(testCmdPath)
		if err == nil {
			t.Error("Expected error for duplicate resource, got nil")
		}
	})
}

// TestMultiToolInstallation tests multi-tool installation scenarios
func TestMultiToolInstallation(t *testing.T) {
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create a test command for testing
	testCmdPath := filepath.Join(repoDir, "multi-test-cmd.md")
	cmdContent := `---
description: A test command for multi-tool testing
---
# Multi-Tool Test Command
`
	if err := os.WriteFile(testCmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}
	if err := manager.AddCommand(testCmdPath); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	t.Run("fresh project uses default tool", func(t *testing.T) {
		projectDir := t.TempDir()

		// Create installer with Claude as default
		installer, err := install.NewInstaller(projectDir, tools.Claude)
		if err != nil {
			t.Fatalf("Failed to create installer: %v", err)
		}

		// Install command
		if err := installer.InstallCommand("multi-test-cmd", manager); err != nil {
			t.Fatalf("Failed to install command: %v", err)
		}

		// Verify .claude/commands/ was created
		cmdSymlink := filepath.Join(projectDir, ".claude", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(cmdSymlink); err != nil {
			t.Errorf("Command symlink not created in .claude: %v", err)
		}

		// Verify NOT created in other directories
		opencodeSymlink := filepath.Join(projectDir, ".opencode", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(opencodeSymlink); err == nil {
			t.Error("Command symlink should not be created in .opencode when it doesn't exist")
		}
	})

	t.Run("existing opencode folder installs to opencode", func(t *testing.T) {
		projectDir := t.TempDir()

		// Create .opencode directory
		if err := os.MkdirAll(filepath.Join(projectDir, ".opencode"), 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// Create installer with Claude as default (should be ignored)
		installer, err := install.NewInstaller(projectDir, tools.Claude)
		if err != nil {
			t.Fatalf("Failed to create installer: %v", err)
		}

		// Install command
		if err := installer.InstallCommand("multi-test-cmd", manager); err != nil {
			t.Fatalf("Failed to install command: %v", err)
		}

		// Verify installed to .opencode
		opencodeSymlink := filepath.Join(projectDir, ".opencode", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(opencodeSymlink); err != nil {
			t.Errorf("Command symlink not created in .opencode: %v", err)
		}

		// Verify NOT created in .claude (since it doesn't exist)
		claudeSymlink := filepath.Join(projectDir, ".claude", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(claudeSymlink); err == nil {
			t.Error("Command symlink should not be created in .claude when it doesn't exist")
		}
	})

	t.Run("multiple existing folders install to all", func(t *testing.T) {
		projectDir := t.TempDir()

		// Create both .claude and .opencode directories
		if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755); err != nil {
			t.Fatalf("Failed to create .claude directory: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(projectDir, ".opencode"), 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// Create installer with any default (should be ignored)
		installer, err := install.NewInstaller(projectDir, tools.Copilot)
		if err != nil {
			t.Fatalf("Failed to create installer: %v", err)
		}

		// Verify it detected both tools
		targets := installer.GetTargetTools()
		if len(targets) != 2 {
			t.Errorf("Expected 2 target tools, got %d", len(targets))
		}

		// Install command
		if err := installer.InstallCommand("multi-test-cmd", manager); err != nil {
			t.Fatalf("Failed to install command: %v", err)
		}

		// Verify installed to BOTH .claude and .opencode
		claudeSymlink := filepath.Join(projectDir, ".claude", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(claudeSymlink); err != nil {
			t.Errorf("Command symlink not created in .claude: %v", err)
		}

		opencodeSymlink := filepath.Join(projectDir, ".opencode", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(opencodeSymlink); err != nil {
			t.Errorf("Command symlink not created in .opencode: %v", err)
		}

		// Verify both symlinks point to the same target
		claudeTarget, _ := os.Readlink(claudeSymlink)
		opencodeTarget, _ := os.Readlink(opencodeSymlink)
		if claudeTarget != opencodeTarget {
			t.Errorf("Symlinks should point to same target: %s != %s", claudeTarget, opencodeTarget)
		}
	})

	t.Run("copilot skills only no commands", func(t *testing.T) {
		projectDir := t.TempDir()

		// Create .github/skills directory
		if err := os.MkdirAll(filepath.Join(projectDir, ".github", "skills"), 0755); err != nil {
			t.Fatalf("Failed to create .github/skills directory: %v", err)
		}

		// Create installer - should detect Copilot
		installer, err := install.NewInstaller(projectDir, tools.Claude)
		if err != nil {
			t.Fatalf("Failed to create installer: %v", err)
		}

		// Verify it detected Copilot
		targets := installer.GetTargetTools()
		if len(targets) != 1 || targets[0] != tools.Copilot {
			t.Errorf("Expected Copilot as only target, got %v", targets)
		}

		// Try to install command - should work but not create anything
		// (Copilot doesn't support commands)
		if err := installer.InstallCommand("multi-test-cmd", manager); err != nil {
			t.Fatalf("Failed to install command: %v", err)
		}

		// Verify NOT created in .github (Copilot doesn't support commands)
		githubSymlink := filepath.Join(projectDir, ".github", "commands", "multi-test-cmd.md")
		if _, err := os.Lstat(githubSymlink); err == nil {
			t.Error("Command symlink should not be created for Copilot (commands not supported)")
		}
	})
}

// TestConfigIntegration tests the config functionality
func TestConfigIntegration(t *testing.T) {
	t.Run("load config with default tool", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a config file
		configPath := filepath.Join(tmpDir, ".ai-repo.yaml")
		configContent := "default-tool: opencode\n"
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// Load config
		cfg, err := config.Load(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify default tool
		if cfg.DefaultTool != "opencode" {
			t.Errorf("DefaultTool = %v, want opencode", cfg.DefaultTool)
		}

		// Get default tool as type
		tool, err := cfg.GetDefaultTool()
		if err != nil {
			t.Fatalf("Failed to get default tool: %v", err)
		}
		if tool != tools.OpenCode {
			t.Errorf("GetDefaultTool() = %v, want OpenCode", tool)
		}
	})

	t.Run("config defaults to claude", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Load config without file (should use defaults)
		cfg, err := config.Load(tmpDir)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Verify defaults to claude
		if cfg.DefaultTool != "claude" {
			t.Errorf("DefaultTool = %v, want claude", cfg.DefaultTool)
		}

		tool, err := cfg.GetDefaultTool()
		if err != nil {
			t.Fatalf("Failed to get default tool: %v", err)
		}
		if tool != tools.Claude {
			t.Errorf("GetDefaultTool() = %v, want Claude", tool)
		}
	})

	t.Run("config validates tool names", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a config file with invalid tool
		configPath := filepath.Join(tmpDir, ".ai-repo.yaml")
		configContent := "default-tool: invalid\n"
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// Load config - should fail validation
		_, err := config.Load(tmpDir)
		if err == nil {
			t.Error("Expected error for invalid tool name, got nil")
		}
	})
}

// TestWithExampleResources tests using the actual example resources
func TestWithExampleResources(t *testing.T) {
	// Skip if examples don't exist (e.g., in CI without examples)
	examplesDir := "../examples"
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("Examples directory not found, skipping test")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	t.Run("add sample command", func(t *testing.T) {
		sampleCmdPath := filepath.Join(examplesDir, "sample-command.md")
		if _, err := os.Stat(sampleCmdPath); err != nil {
			t.Skip("sample-command.md not found")
		}

		if err := manager.AddCommand(sampleCmdPath); err != nil {
			t.Errorf("Failed to add sample command: %v", err)
		}

		// Verify it was added
		res, err := manager.Get("sample-command", resource.Command)
		if err != nil {
			t.Errorf("Failed to get sample command: %v", err)
		}
		if res.Description == "" {
			t.Error("Sample command description is empty")
		}
	})

	t.Run("add sample skill", func(t *testing.T) {
		sampleSkillPath := filepath.Join(examplesDir, "sample-skill")
		if _, err := os.Stat(sampleSkillPath); err != nil {
			t.Skip("sample-skill directory not found")
		}

		if err := manager.AddSkill(sampleSkillPath); err != nil {
			t.Errorf("Failed to add sample skill: %v", err)
		}

		// Verify it was added
		res, err := manager.Get("sample-skill", resource.Skill)
		if err != nil {
			t.Errorf("Failed to get sample skill: %v", err)
		}
		if res.Description == "" {
			t.Error("Sample skill description is empty")
		}
		if res.License != "MIT" {
			t.Errorf("Sample skill license = %v, want MIT", res.License)
		}
	})
}
