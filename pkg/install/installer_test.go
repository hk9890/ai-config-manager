package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hans-m-leitner/ai-config-manager/pkg/repo"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
	"github.com/hans-m-leitner/ai-config-manager/pkg/tools"
)

func setupTestRepo(t *testing.T) (*repo.Manager, string) {
	t.Helper()

	// Create temp repo
	tmpDir := t.TempDir()
	manager := repo.NewManagerWithPath(tmpDir)

	// Add a test command
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
		t.Fatalf("Failed to add command to repo: %v", err)
	}

	// Add a test skill
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
		t.Fatalf("Failed to add skill to repo: %v", err)
	}

	return manager, tmpDir
}

func TestNewInstaller(t *testing.T) {
	tmpDir := t.TempDir()

	installer, err := NewInstaller(tmpDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	if installer.projectPath != tmpDir {
		t.Errorf("projectPath = %v, want %v", installer.projectPath, tmpDir)
	}

	// Should have Claude as target when no existing dirs
	targets := installer.GetTargetTools()
	if len(targets) != 1 || targets[0] != tools.Claude {
		t.Errorf("targetTools = %v, want [Claude]", targets)
	}
}

func TestDetectInstallTargets(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		defaultTool   tools.Tool
		expectedTools []tools.Tool
	}{
		{
			name:          "no existing folders - uses default",
			setupFunc:     func(dir string) error { return nil },
			defaultTool:   tools.Claude,
			expectedTools: []tools.Tool{tools.Claude},
		},
		{
			name: "one existing folder - claude",
			setupFunc: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, ".claude"), 0755)
			},
			defaultTool:   tools.OpenCode,
			expectedTools: []tools.Tool{tools.Claude},
		},
		{
			name: "one existing folder - opencode",
			setupFunc: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, ".opencode"), 0755)
			},
			defaultTool:   tools.Claude,
			expectedTools: []tools.Tool{tools.OpenCode},
		},
		{
			name: "multiple existing folders",
			setupFunc: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(dir, ".opencode"), 0755)
			},
			defaultTool:   tools.Copilot,
			expectedTools: []tools.Tool{tools.Claude, tools.OpenCode},
		},
		{
			name: "all three tools",
			setupFunc: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(dir, ".github", "skills"), 0755)
			},
			defaultTool:   tools.Claude,
			expectedTools: []tools.Tool{tools.Claude, tools.OpenCode, tools.Copilot},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			targets, err := DetectInstallTargets(tmpDir, tt.defaultTool)
			if err != nil {
				t.Fatalf("DetectInstallTargets() error = %v", err)
			}

			if len(targets) != len(tt.expectedTools) {
				t.Errorf("DetectInstallTargets() returned %d tools, want %d", len(targets), len(tt.expectedTools))
				t.Errorf("Got: %v, Want: %v", targets, tt.expectedTools)
				return
			}

			// Check each expected tool is present
			for _, expected := range tt.expectedTools {
				found := false
				for _, target := range targets {
					if target == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tool %v not found in targets %v", expected, targets)
				}
			}
		})
	}
}

func TestInstallCommand(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Install command
	err = installer.InstallCommand("test-cmd", manager)
	if err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// Verify symlink was created in .claude/commands
	symlinkPath := filepath.Join(projectDir, ".claude", "commands", "test-cmd.md")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Symlink not created: %v", err)
	}

	// Verify it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected symlink, got regular file")
	}

	// Verify symlink target is correct
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	expectedTarget := manager.GetPath("test-cmd", resource.Command)
	if target != expectedTarget {
		t.Errorf("Symlink target = %v, want %v", target, expectedTarget)
	}
}

func TestInstallCommandMultipleTools(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	// Create both .claude and .opencode directories
	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.MkdirAll(filepath.Join(projectDir, ".opencode"), 0755)

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Should target both tools
	targets := installer.GetTargetTools()
	if len(targets) != 2 {
		t.Fatalf("Expected 2 target tools, got %d", len(targets))
	}

	// Install command
	err = installer.InstallCommand("test-cmd", manager)
	if err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// Verify symlink was created in both directories
	claudeSymlink := filepath.Join(projectDir, ".claude", "commands", "test-cmd.md")
	opencodeSymlink := filepath.Join(projectDir, ".opencode", "commands", "test-cmd.md")

	for _, path := range []string{claudeSymlink, opencodeSymlink} {
		info, err := os.Lstat(path)
		if err != nil {
			t.Errorf("Symlink not created at %s: %v", path, err)
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("Expected symlink at %s, got regular file", path)
		}
	}
}

func TestInstallSkill(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Install skill
	err = installer.InstallSkill("test-skill", manager)
	if err != nil {
		t.Fatalf("InstallSkill() error = %v", err)
	}

	// Verify symlink was created in .claude/skills
	symlinkPath := filepath.Join(projectDir, ".claude", "skills", "test-skill")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Symlink not created: %v", err)
	}

	// Verify it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected symlink, got regular directory")
	}

	// Verify symlink target is correct
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	expectedTarget := manager.GetPath("test-skill", resource.Skill)
	if target != expectedTarget {
		t.Errorf("Symlink target = %v, want %v", target, expectedTarget)
	}
}

func TestUninstall(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Install command
	if err := installer.InstallCommand("test-cmd", manager); err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// Verify it's installed
	if !installer.IsInstalled("test-cmd", resource.Command) {
		t.Fatal("Command should be installed")
	}

	// Uninstall
	err = installer.Uninstall("test-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify it's uninstalled
	if installer.IsInstalled("test-cmd", resource.Command) {
		t.Error("Command should be uninstalled")
	}
}

func TestUninstallNotInstalled(t *testing.T) {
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Try to uninstall non-existent resource
	err = installer.Uninstall("nonexistent", resource.Command)
	if err == nil {
		t.Error("Uninstall() expected error for non-installed resource, got nil")
	}
}

func TestList(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Initially empty
	resources, err := installer.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("List() returned %d resources, want 0", len(resources))
	}

	// Install command and skill
	if err := installer.InstallCommand("test-cmd", manager); err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}
	if err := installer.InstallSkill("test-skill", manager); err != nil {
		t.Fatalf("InstallSkill() error = %v", err)
	}

	// List should return 2 resources
	resources, err = installer.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d resources, want 2", len(resources))
	}

	// Verify types
	commandFound := false
	skillFound := false
	for _, res := range resources {
		if res.Type == resource.Command && res.Name == "test-cmd" {
			commandFound = true
		}
		if res.Type == resource.Skill && res.Name == "test-skill" {
			skillFound = true
		}
	}

	if !commandFound {
		t.Error("List() did not return installed command")
	}
	if !skillFound {
		t.Error("List() did not return installed skill")
	}
}

func TestIsInstalled(t *testing.T) {
	manager, _ := setupTestRepo(t)
	projectDir := t.TempDir()

	installer, err := NewInstaller(projectDir, tools.Claude)
	if err != nil {
		t.Fatalf("NewInstaller() error = %v", err)
	}

	// Not installed initially
	if installer.IsInstalled("test-cmd", resource.Command) {
		t.Error("IsInstalled() returned true for non-installed command")
	}

	// Install command
	if err := installer.InstallCommand("test-cmd", manager); err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// Should be installed now
	if !installer.IsInstalled("test-cmd", resource.Command) {
		t.Error("IsInstalled() returned false for installed command")
	}

	// Other resource should not be installed
	if installer.IsInstalled("test-skill", resource.Skill) {
		t.Error("IsInstalled() returned true for non-installed skill")
	}
}
