package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/repo"
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
				// Create symlink to different repo
				wrongRepo := filepath.Join(os.TempDir(), "wrong-repo")
				_ = os.MkdirAll(wrongRepo, 0755)
				target := filepath.Join(wrongRepo, "commands", "test-cmd")
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
			issues, err := verifyDirectory(testDir, "command", "claude", repoDir)
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
