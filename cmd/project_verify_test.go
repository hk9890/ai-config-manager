package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

func TestScanProjectIssues(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string, string) error
		expectedCount int
		expectedType  string
	}{
		{
			name: "detects broken symlinks",
			setupFunc: func(projectDir, repoDir string) error {
				claudeDir := filepath.Join(projectDir, ".claude", "commands")
				if err := os.MkdirAll(claudeDir, 0755); err != nil {
					return err
				}
				// Create symlink to non-existent target
				target := filepath.Join(repoDir, "commands", "missing-cmd")
				return os.Symlink(target, filepath.Join(claudeDir, "missing-cmd"))
			},
			expectedCount: 1,
			expectedType:  "broken",
		},
		{
			name: "detects symlinks pointing to wrong repo",
			setupFunc: func(projectDir, repoDir string) error {
				claudeDir := filepath.Join(projectDir, ".claude", "commands")
				if err := os.MkdirAll(claudeDir, 0755); err != nil {
					return err
				}
				// Create symlink to different repo (use a subdir of projectDir to avoid shared tmp issues)
				wrongRepo := filepath.Join(projectDir, "wrong-repo")
				wrongRepoCommands := filepath.Join(wrongRepo, "commands")
				if err := os.MkdirAll(wrongRepoCommands, 0755); err != nil {
					return err
				}
				target := filepath.Join(wrongRepoCommands, "test-cmd")
				if err := os.WriteFile(target, []byte("test"), 0644); err != nil {
					return err
				}
				return os.Symlink(target, filepath.Join(claudeDir, "test-cmd"))
			},
			expectedCount: 1,
			expectedType:  "wrong-repo",
		},
		{
			name: "no issues with valid symlinks",
			setupFunc: func(projectDir, repoDir string) error {
				claudeDir := filepath.Join(projectDir, ".claude", "commands")
				if err := os.MkdirAll(claudeDir, 0755); err != nil {
					return err
				}
				// Create valid symlink
				target := filepath.Join(repoDir, "commands", "test-cmd")
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(target, []byte("test"), 0644); err != nil {
					return err
				}
				return os.Symlink(target, filepath.Join(claudeDir, "test-cmd"))
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			repoDir := t.TempDir()

			// Setup test scenario
			if err := tt.setupFunc(projectDir, repoDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Detect tools
			detectedTools, err := tools.DetectExistingTools(projectDir)
			if err != nil {
				t.Fatalf("Failed to detect tools: %v", err)
			}

			// Scan for issues
			issues, err := scanProjectIssues(projectDir, detectedTools, repoDir)
			if err != nil {
				t.Fatalf("scanProjectIssues failed: %v", err)
			}

			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}

			if tt.expectedCount > 0 && len(issues) > 0 {
				if issues[0].IssueType != tt.expectedType {
					t.Errorf("Issue type = %v, want %v", issues[0].IssueType, tt.expectedType)
				}
			}
		})
	}
}

func TestVerifyDirectory(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string, string) error
		expectedCount int
	}{
		{
			name: "detects broken symlink",
			setupFunc: func(dir, repoPath string) error {
				target := filepath.Join(repoPath, "commands", "missing-cmd")
				return os.Symlink(target, filepath.Join(dir, "missing-cmd"))
			},
			expectedCount: 1,
		},
		{
			name: "valid symlink has no issues",
			setupFunc: func(dir, repoPath string) error {
				target := filepath.Join(repoPath, "commands", "test-cmd")
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(target, []byte("test"), 0644); err != nil {
					return err
				}
				return os.Symlink(target, filepath.Join(dir, "test-cmd"))
			},
			expectedCount: 0,
		},
		{
			name: "ignores regular files",
			setupFunc: func(dir, repoPath string) error {
				return os.WriteFile(filepath.Join(dir, "regular-file"), []byte("test"), 0644)
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			repoDir := t.TempDir()

			testDir := filepath.Join(dir, "commands")
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Setup test scenario
			if err := tt.setupFunc(testDir, repoDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Verify directory
			issues, err := verifyDirectory(testDir, "claude", repoDir)
			if err != nil {
				t.Fatalf("verifyDirectory failed: %v", err)
			}

			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}
		})
	}
}

func TestCheckManifestSync(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string, string) error
		expectedCount int
	}{
		{
			name: "detects resource in manifest but not installed",
			setupFunc: func(projectDir, repoDir string) error {
				// Create manifest with a resource
				m := &manifest.Manifest{
					Resources: []string{"skill/test-skill"},
				}
				manifestPath := filepath.Join(projectDir, manifest.ManifestFileName)
				return m.Save(manifestPath)
			},
			expectedCount: 1,
		},
		{
			name: "no issues when resource is installed",
			setupFunc: func(projectDir, repoDir string) error {
				// Create manifest
				m := &manifest.Manifest{
					Resources: []string{"skill/test-skill"},
				}
				manifestPath := filepath.Join(projectDir, manifest.ManifestFileName)
				if err := m.Save(manifestPath); err != nil {
					return err
				}

				// Create installed resource
				claudeDir := filepath.Join(projectDir, ".claude", "skills")
				if err := os.MkdirAll(claudeDir, 0755); err != nil {
					return err
				}

				// Create target in repo
				target := filepath.Join(repoDir, "skills", "test-skill")
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("test"), 0644); err != nil {
					return err
				}

				// Create symlink
				return os.Symlink(target, filepath.Join(claudeDir, "test-skill"))
			},
			expectedCount: 0,
		},
		{
			name: "no issues when no manifest exists",
			setupFunc: func(projectDir, repoDir string) error {
				// Don't create manifest
				return nil
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			repoDir := t.TempDir()

			// Setup test scenario
			if err := tt.setupFunc(projectDir, repoDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Detect tools
			detectedTools, err := tools.DetectExistingTools(projectDir)
			if err != nil {
				t.Fatalf("Failed to detect tools: %v", err)
			}

			// Check manifest sync
			issues, err := checkManifestSync(projectDir, detectedTools, repoDir)
			if err != nil {
				t.Fatalf("checkManifestSync failed: %v", err)
			}

			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}

			if tt.expectedCount > 0 && len(issues) > 0 {
				if issues[0].IssueType != "not-installed" {
					t.Errorf("Issue type = %v, want not-installed", issues[0].IssueType)
				}
			}
		})
	}
}

func TestProjectVerifyCommand(t *testing.T) {
	// Create temp directories
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Set AIMGR_REPO_PATH
	oldEnv := os.Getenv("AIMGR_REPO_PATH")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("AIMGR_REPO_PATH", oldEnv)
		} else {
			_ = os.Unsetenv("AIMGR_REPO_PATH")
		}
	}()
	_ = os.Setenv("AIMGR_REPO_PATH", repoDir)

	// Initialize repo
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Create tool directory with valid symlink
	claudeDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create tool directory: %v", err)
	}

	// Create repo command
	repoCommand := filepath.Join(repoDir, "commands", "test-cmd")
	if err := os.WriteFile(repoCommand, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create repo command: %v", err)
	}

	// Create symlink
	symlinkPath := filepath.Join(claudeDir, "test-cmd")
	if err := os.Symlink(repoCommand, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Change to project directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	// Run verify command
	err = projectVerifyCmd.RunE(projectVerifyCmd, []string{})
	if err != nil {
		t.Errorf("Verify command failed: %v", err)
	}
}

func TestDisplayVerifyIssues(t *testing.T) {
	issues := []VerifyIssue{
		{
			Resource:    "test-cmd",
			Tool:        "claude",
			IssueType:   "broken",
			Description: "Symlink target doesn't exist",
			Severity:    "error",
		},
		{
			Resource:    "test-skill",
			Tool:        "opencode",
			IssueType:   "wrong-repo",
			Description: "Points to wrong repo",
			Severity:    "warning",
		},
	}

	// Just verify it doesn't panic and returns no error
	err := displayVerifyIssues(issues, output.Table)
	if err != nil {
		t.Errorf("displayVerifyIssues failed: %v", err)
	}
}

func TestVerifyIssueTypes(t *testing.T) {
	issueTypes := []string{"broken", "wrong-repo", "not-installed", "orphaned", "unreadable"}

	for _, issueType := range issueTypes {
		t.Run(issueType, func(t *testing.T) {
			issue := VerifyIssue{
				Resource:    "test-resource",
				Tool:        "test-tool",
				IssueType:   issueType,
				Description: "Test issue",
			}

			if issue.IssueType != issueType {
				t.Errorf("IssueType = %v, want %v", issue.IssueType, issueType)
			}
		})
	}
}

// TestParseResourceFromIssue tests the helper that extracts resource type and name
// from a VerifyIssue based on its path.
func TestParseResourceFromIssue(t *testing.T) {
	tests := []struct {
		name         string
		issue        VerifyIssue
		expectedType resource.ResourceType
		expectedName string
	}{
		{
			name: "skill from opencode skills directory",
			issue: VerifyIssue{
				Resource: "my-skill",
				Path:     "/project/.opencode/skills/my-skill",
			},
			expectedType: resource.Skill,
			expectedName: "my-skill",
		},
		{
			name: "command from claude commands directory",
			issue: VerifyIssue{
				Resource: "my-cmd.md",
				Path:     "/project/.claude/commands/my-cmd.md",
			},
			expectedType: resource.Command,
			expectedName: "my-cmd",
		},
		{
			name: "agent from opencode agents directory",
			issue: VerifyIssue{
				Resource: "my-agent.md",
				Path:     "/project/.opencode/agents/my-agent.md",
			},
			expectedType: resource.Agent,
			expectedName: "my-agent",
		},
		{
			name: "command strips .md extension",
			issue: VerifyIssue{
				Resource: "test-command.md",
				Path:     "/project/.opencode/commands/test-command.md",
			},
			expectedType: resource.Command,
			expectedName: "test-command",
		},
		{
			name: "agent strips .md extension",
			issue: VerifyIssue{
				Resource: "debug-agent.md",
				Path:     "/project/.claude/agents/debug-agent.md",
			},
			expectedType: resource.Agent,
			expectedName: "debug-agent",
		},
		{
			name: "skill without .md extension kept as-is",
			issue: VerifyIssue{
				Resource: "pdf-processing",
				Path:     "/project/.claude/skills/pdf-processing",
			},
			expectedType: resource.Skill,
			expectedName: "pdf-processing",
		},
		{
			name: "fallback defaults to skill",
			issue: VerifyIssue{
				Resource: "unknown-resource",
				Path:     "/project/.unknown/something/unknown-resource",
			},
			expectedType: resource.Skill,
			expectedName: "unknown-resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resType, resName := parseResourceFromIssue(tt.issue)

			if resType != tt.expectedType {
				t.Errorf("parseResourceFromIssue() type = %v, want %v", resType, tt.expectedType)
			}
			if resName != tt.expectedName {
				t.Errorf("parseResourceFromIssue() name = %v, want %v", resName, tt.expectedName)
			}
		})
	}
}

// TestFixVerifyIssues_ReinstallsBrokenSkill tests that fixVerifyIssues()
// can reinstall a broken skill when the resource still exists in the repo.
func TestFixVerifyIssues_ReinstallsBrokenSkill(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Initialize repo
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create and add a skill to the repo
	tempSkillDir := t.TempDir()
	skillDir := filepath.Join(tempSkillDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte("---\ndescription: A test skill\n---\n\n# Test Skill\n\nA test skill"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := manager.AddSkill(skillDir, "file://"+skillDir, "file"); err != nil {
		t.Fatalf("Failed to add skill to repo: %v", err)
	}

	// Create tool directory with a broken symlink (simulating a broken installation)
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	brokenSymlink := filepath.Join(skillsDir, "test-skill")
	if err := os.Symlink("/nonexistent/old-repo/skills/test-skill", brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Verify the symlink is broken
	if _, err := os.Stat(brokenSymlink); err == nil {
		t.Fatal("Expected broken symlink, but target exists")
	}

	// Create a VerifyIssue for the broken symlink
	issues := []VerifyIssue{
		{
			Resource:    "test-skill",
			Tool:        "opencode",
			IssueType:   "broken",
			Description: "Symlink target doesn't exist",
			Path:        brokenSymlink,
			Severity:    "error",
		},
	}

	// Call fixVerifyIssues
	err := fixVerifyIssues(projectDir, issues, manager)
	if err != nil {
		t.Fatalf("fixVerifyIssues returned error: %v", err)
	}

	// Verify the symlink is now valid
	targetInfo, err := os.Stat(brokenSymlink)
	if err != nil {
		t.Fatalf("Symlink target is invalid after fix: %v", err)
	}
	if !targetInfo.IsDir() {
		t.Error("Expected symlink to point to a directory")
	}

	// Verify symlink points to the repo
	actualTarget, err := os.Readlink(brokenSymlink)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	expectedTarget := filepath.Join(repoDir, "skills", "test-skill")
	if actualTarget != expectedTarget {
		t.Errorf("Symlink points to wrong target: got %s, want %s", actualTarget, expectedTarget)
	}
}

// TestFixVerifyIssues_ReportsUnrecoverableResource tests that fixVerifyIssues()
// handles the case where a broken symlink's resource no longer exists in the repo.
func TestFixVerifyIssues_ReportsUnrecoverableResource(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Initialize repo (but don't add any resources)
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create tool directory with a broken symlink
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}
	brokenSymlink := filepath.Join(skillsDir, "gone-skill")
	if err := os.Symlink("/nonexistent/old-repo/skills/gone-skill", brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Create a VerifyIssue for the broken symlink
	issues := []VerifyIssue{
		{
			Resource:    "gone-skill",
			Tool:        "opencode",
			IssueType:   "broken",
			Description: "Symlink target doesn't exist",
			Path:        brokenSymlink,
			Severity:    "error",
		},
	}

	// Call fixVerifyIssues
	err := fixVerifyIssues(projectDir, issues, manager)
	if err != nil {
		t.Fatalf("fixVerifyIssues returned error: %v", err)
	}

	// Verify the broken symlink was removed
	if _, err := os.Lstat(brokenSymlink); err == nil {
		t.Error("Expected broken symlink to be removed, but it still exists")
	}
}

// TestFixVerifyIssues_InstallsNotInstalledResource tests that fixVerifyIssues()
// can install a resource that's in the manifest but not yet installed.
func TestFixVerifyIssues_InstallsNotInstalledResource(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Initialize repo
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create and add a skill to the repo
	tempSkillDir := t.TempDir()
	skillDir := filepath.Join(tempSkillDir, "missing-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\ndescription: A missing skill\n---\n\n# Missing Skill"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := manager.AddSkill(skillDir, "file://"+skillDir, "file"); err != nil {
		t.Fatalf("Failed to add skill to repo: %v", err)
	}

	// Create tool directory (so DetectExistingTools finds something)
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	// Create a VerifyIssue for a not-installed resource
	issues := []VerifyIssue{
		{
			Resource:    "skill/missing-skill",
			Tool:        "any",
			IssueType:   "not-installed",
			Description: "Listed in ai.package.yaml but not installed",
			Path:        filepath.Join(projectDir, "ai.package.yaml"),
			Severity:    "warning",
		},
	}

	// Call fixVerifyIssues
	err := fixVerifyIssues(projectDir, issues, manager)
	if err != nil {
		t.Fatalf("fixVerifyIssues returned error: %v", err)
	}

	// Verify the skill is now installed
	installedSymlink := filepath.Join(skillsDir, "missing-skill")
	info, err := os.Lstat(installedSymlink)
	if err != nil {
		t.Fatalf("Expected skill to be installed, but symlink doesn't exist: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected a symlink, got a regular file/directory")
	}

	// Verify the symlink target is valid
	if _, err := os.Stat(installedSymlink); err != nil {
		t.Errorf("Symlink target is invalid: %v", err)
	}
}

// TestFixVerifyIssues_WrongRepoReinstalls tests that fixVerifyIssues()
// can fix a "wrong-repo" issue by reinstalling from the correct repo.
func TestFixVerifyIssues_WrongRepoReinstalls(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Initialize repo
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create and add a skill to the repo
	tempSkillDir := t.TempDir()
	skillDir := filepath.Join(tempSkillDir, "fix-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\ndescription: A skill to fix\n---\n\n# Fix Skill"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := manager.AddSkill(skillDir, "file://"+skillDir, "file"); err != nil {
		t.Fatalf("Failed to add skill to repo: %v", err)
	}

	// Create tool directory with symlink pointing to wrong repo
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	// Create a "wrong repo" — a valid symlink target but in wrong location
	wrongRepo := t.TempDir()
	wrongSkillDir := filepath.Join(wrongRepo, "skills", "fix-skill")
	if err := os.MkdirAll(wrongSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create wrong repo skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wrongSkillDir, "SKILL.md"), []byte("wrong"), 0644); err != nil {
		t.Fatalf("Failed to write wrong SKILL.md: %v", err)
	}

	wrongSymlink := filepath.Join(skillsDir, "fix-skill")
	if err := os.Symlink(wrongSkillDir, wrongSymlink); err != nil {
		t.Fatalf("Failed to create wrong-repo symlink: %v", err)
	}

	// Create a VerifyIssue for the wrong-repo symlink
	issues := []VerifyIssue{
		{
			Resource:    "fix-skill",
			Tool:        "opencode",
			IssueType:   "wrong-repo",
			Description: "Points to wrong repo",
			Path:        wrongSymlink,
			Severity:    "warning",
		},
	}

	// Call fixVerifyIssues
	err := fixVerifyIssues(projectDir, issues, manager)
	if err != nil {
		t.Fatalf("fixVerifyIssues returned error: %v", err)
	}

	// Verify the symlink now points to the correct repo
	actualTarget, err := os.Readlink(wrongSymlink)
	if err != nil {
		t.Fatalf("Failed to read symlink after fix: %v", err)
	}
	expectedTarget := filepath.Join(repoDir, "skills", "fix-skill")
	if actualTarget != expectedTarget {
		t.Errorf("Symlink points to wrong target after fix: got %s, want %s", actualTarget, expectedTarget)
	}
}


// TestCheckManifestSync_PackageListsMissingResources verifies that when a
// package has missing resources, the issue description names each one.
func TestCheckManifestSync_PackageListsMissingResources(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Create a package definition in the repo with 3 resources
	pkg := &resource.Package{
		Name:        "test-pkg",
		Description: "A test package",
		Resources:   []string{"skill/installed-skill", "agent/missing-agent", "command/missing-cmd"},
	}
	if err := resource.SavePackage(pkg, repoDir); err != nil {
		t.Fatalf("Failed to save package: %v", err)
	}

	// Create manifest referencing the package
	m := &manifest.Manifest{
		Resources: []string{"package/test-pkg"},
	}
	manifestPath := filepath.Join(projectDir, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create tool directory and install only one of the three resources
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	// Create the installed skill target in repo
	repoSkillDir := filepath.Join(repoDir, "skills", "installed-skill")
	if err := os.MkdirAll(repoSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create repo skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoSkillDir, "SKILL.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := os.Symlink(repoSkillDir, filepath.Join(skillsDir, "installed-skill")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Also create agents dir (needed for opencode tool detection, but leave agent missing)
	agentsDir := filepath.Join(projectDir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Also create commands dir (needed for opencode tool detection, but leave command missing)
	commandsDir := filepath.Join(projectDir, ".opencode", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Detect tools
	detectedTools, err := tools.DetectExistingTools(projectDir)
	if err != nil {
		t.Fatalf("Failed to detect tools: %v", err)
	}

	// Check manifest sync
	issues, err := checkManifestSync(projectDir, detectedTools, repoDir)
	if err != nil {
		t.Fatalf("checkManifestSync failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d: %+v", len(issues), issues)
	}

	issue := issues[0]
	if issue.Resource != "package/test-pkg" {
		t.Errorf("Expected resource 'package/test-pkg', got %q", issue.Resource)
	}
	if issue.IssueType != "not-installed" {
		t.Errorf("Expected issue type 'not-installed', got %q", issue.IssueType)
	}

	// Verify the description includes the count and the missing resource names
	desc := issue.Description
	if !strings.Contains(desc, "2 resource(s) not installed") {
		t.Errorf("Expected description to contain '2 resource(s) not installed', got: %s", desc)
	}
	if !strings.Contains(desc, "agent/missing-agent") {
		t.Errorf("Expected description to contain 'agent/missing-agent', got: %s", desc)
	}
	if !strings.Contains(desc, "command/missing-cmd") {
		t.Errorf("Expected description to contain 'command/missing-cmd', got: %s", desc)
	}
	// Verify installed resource is NOT in the missing list
	if strings.Contains(desc, "installed-skill") {
		t.Errorf("Description should NOT contain 'installed-skill' (it's installed), got: %s", desc)
	}
}

// TestCheckManifestSync_PackageAllInstalledNoIssue verifies that a package
// with all resources installed produces no issues.
func TestCheckManifestSync_PackageAllInstalledNoIssue(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Create a package definition with one resource
	pkg := &resource.Package{
		Name:        "complete-pkg",
		Description: "A fully installed package",
		Resources:   []string{"skill/good-skill"},
	}
	if err := resource.SavePackage(pkg, repoDir); err != nil {
		t.Fatalf("Failed to save package: %v", err)
	}

	// Create manifest
	m := &manifest.Manifest{
		Resources: []string{"package/complete-pkg"},
	}
	if err := m.Save(filepath.Join(projectDir, manifest.ManifestFileName)); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Install the skill
	skillsDir := filepath.Join(projectDir, ".opencode", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	repoSkillDir := filepath.Join(repoDir, "skills", "good-skill")
	if err := os.MkdirAll(repoSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create repo skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoSkillDir, "SKILL.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write SKILL.md: %v", err)
	}
	if err := os.Symlink(repoSkillDir, filepath.Join(skillsDir, "good-skill")); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	detectedTools, err := tools.DetectExistingTools(projectDir)
	if err != nil {
		t.Fatalf("Failed to detect tools: %v", err)
	}

	issues, err := checkManifestSync(projectDir, detectedTools, repoDir)
	if err != nil {
		t.Fatalf("checkManifestSync failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for fully installed package, got %d: %+v", len(issues), issues)
	}
}
// TestCheckManifestSync_DetectsBrokenSymlinkAsNotInstalled verifies that
// checkManifestSync treats a broken symlink as "not installed" since os.Stat
// correctly follows the symlink and fails when the target doesn't exist.
func TestCheckManifestSync_DetectsBrokenSymlinkAsNotInstalled(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Create manifest referencing a skill
	m := &manifest.Manifest{
		Resources: []string{"skill/broken-skill"},
	}
	manifestPath := filepath.Join(projectDir, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create tool directory with a broken symlink (target doesn't exist)
	skillsDir := filepath.Join(projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	brokenTarget := filepath.Join(repoDir, "skills", "broken-skill")
	if err := os.Symlink(brokenTarget, filepath.Join(skillsDir, "broken-skill")); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Detect tools
	detectedTools, err := tools.DetectExistingTools(projectDir)
	if err != nil {
		t.Fatalf("Failed to detect tools: %v", err)
	}

	// Check manifest sync — should detect the broken symlink as "not-installed"
	issues, err := checkManifestSync(projectDir, detectedTools, repoDir)
	if err != nil {
		t.Fatalf("checkManifestSync failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d: %+v", len(issues), issues)
	}
	if issues[0].IssueType != "not-installed" {
		t.Errorf("Expected issue type 'not-installed', got %q", issues[0].IssueType)
	}
	if issues[0].Resource != "skill/broken-skill" {
		t.Errorf("Expected resource 'skill/broken-skill', got %q", issues[0].Resource)
	}
}

// TestDeduplicateIssues verifies that manifest "not-installed" issues are
// dropped when Phase 1 already reported the same resource (e.g., as "broken").
func TestDeduplicateIssues(t *testing.T) {
	tests := []struct {
		name           string
		existing       []VerifyIssue
		manifestIssues []VerifyIssue
		expectedCount  int
	}{
		{
			name: "removes duplicate broken+not-installed for same resource",
			existing: []VerifyIssue{
				{Resource: "my-skill", Tool: "claude", IssueType: "broken", Severity: "error"},
			},
			manifestIssues: []VerifyIssue{
				{Resource: "skill/my-skill", Tool: "any", IssueType: "not-installed", Severity: "warning"},
			},
			expectedCount: 1, // Only the "broken" issue
		},
		{
			name:     "keeps non-duplicate manifest issues",
			existing: []VerifyIssue{},
			manifestIssues: []VerifyIssue{
				{Resource: "skill/other-skill", Tool: "any", IssueType: "not-installed", Severity: "warning"},
			},
			expectedCount: 1, // The "not-installed" issue is kept
		},
		{
			name: "keeps non-not-installed manifest issues even if resource matches",
			existing: []VerifyIssue{
				{Resource: "my-skill", Tool: "claude", IssueType: "broken", Severity: "error"},
			},
			manifestIssues: []VerifyIssue{
				{Resource: "skill/my-skill", Tool: "any", IssueType: "orphaned", Severity: "warning"},
			},
			expectedCount: 2, // Both kept (orphaned is not filtered)
		},
		{
			name: "no manifest issues means no change",
			existing: []VerifyIssue{
				{Resource: "cmd", Tool: "opencode", IssueType: "broken", Severity: "error"},
			},
			manifestIssues: []VerifyIssue{},
			expectedCount:  1,
		},
		{
			name:           "both empty",
			existing:       []VerifyIssue{},
			manifestIssues: []VerifyIssue{},
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateIssues(tt.existing, tt.manifestIssues)
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d: %+v", tt.expectedCount, len(result), result)
			}
		})
	}
}

// TestBrokenSymlinkNoDuplicatesBetweenPhases verifies the end-to-end behavior:
// a broken symlink that is also in the manifest produces exactly one issue
// (the "broken" from Phase 1), not a duplicate "not-installed" from Phase 2.
func TestBrokenSymlinkNoDuplicatesBetweenPhases(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()

	// Create manifest referencing a skill
	m := &manifest.Manifest{
		Resources: []string{"skill/test-skill"},
	}
	manifestPath := filepath.Join(projectDir, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Create tool directory with a broken symlink
	skillsDir := filepath.Join(projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	brokenTarget := filepath.Join(repoDir, "skills", "test-skill")
	if err := os.Symlink(brokenTarget, filepath.Join(skillsDir, "test-skill")); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Detect tools
	detectedTools, err := tools.DetectExistingTools(projectDir)
	if err != nil {
		t.Fatalf("Failed to detect tools: %v", err)
	}

	// Phase 1: scan for symlink issues
	phase1Issues, err := scanProjectIssues(projectDir, detectedTools, repoDir)
	if err != nil {
		t.Fatalf("scanProjectIssues failed: %v", err)
	}

	// Phase 2: check manifest sync
	phase2Issues, err := checkManifestSync(projectDir, detectedTools, repoDir)
	if err != nil {
		t.Fatalf("checkManifestSync failed: %v", err)
	}

	// Deduplicate
	allIssues := deduplicateIssues(phase1Issues, phase2Issues)

	// Should have exactly 1 issue (the "broken" from Phase 1)
	if len(allIssues) != 1 {
		t.Fatalf("Expected exactly 1 issue after dedup, got %d: %+v", len(allIssues), allIssues)
	}
	if allIssues[0].IssueType != "broken" {
		t.Errorf("Expected issue type 'broken', got %q", allIssues[0].IssueType)
	}
	if allIssues[0].Resource != "test-skill" {
		t.Errorf("Expected resource 'test-skill', got %q", allIssues[0].Resource)
	}
}
