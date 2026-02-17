package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestRepoVerifyCommand(t *testing.T) {
	tests := []struct {
		name         string
		setupRepo    func(*repo.Manager) error
		wantErrors   bool
		wantWarnings bool
	}{
		{
			name: "healthy repository with no issues",
			setupRepo: func(mgr *repo.Manager) error {
				// Create a command with proper metadata
				commandPath := filepath.Join(mgr.GetRepoPath(), "commands", "test-cmd")
				if err := os.WriteFile(commandPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
					return err
				}

				// Create metadata for the command
				meta := &metadata.ResourceMetadata{
					Name:       "test-cmd",
					Type:       resource.Command,
					SourceType: "local",
					SourceURL:  "file:///test/source",
				}
				return metadata.Save(meta, mgr.GetRepoPath(), "local")
			},
			wantErrors:   false,
			wantWarnings: false,
		},
		{
			name: "resource without metadata",
			setupRepo: func(mgr *repo.Manager) error {
				// Create a command without metadata
				commandPath := filepath.Join(mgr.GetRepoPath(), "commands", "no-meta-cmd")
				return os.WriteFile(commandPath, []byte("#!/bin/bash\necho test"), 0755)
			},
			wantErrors:   false,
			wantWarnings: true,
		},
		{
			name: "orphaned metadata without resource",
			setupRepo: func(mgr *repo.Manager) error {
				// Create metadata without corresponding resource
				meta := &metadata.ResourceMetadata{
					Name:       "orphaned-cmd",
					Type:       resource.Command,
					SourceType: "local",
					SourceURL:  "file:///test/source",
				}
				return metadata.Save(meta, mgr.GetRepoPath(), "local")
			},
			wantErrors:   true,
			wantWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test repository
			repoDir := t.TempDir()

			// Set AIMGR_REPO_PATH to use test directory
			oldEnv := os.Getenv("AIMGR_REPO_PATH")
			defer func() {
				if oldEnv != "" {
					_ = os.Setenv("AIMGR_REPO_PATH", oldEnv)
				} else {
					_ = os.Unsetenv("AIMGR_REPO_PATH")
				}
			}()
			_ = os.Setenv("AIMGR_REPO_PATH", repoDir)

			// Initialize repository
			manager := repo.NewManagerWithPath(repoDir)
			if err := manager.Init(); err != nil {
				t.Fatalf("Failed to initialize repo: %v", err)
			}

			// Setup test scenario
			if tt.setupRepo != nil {
				if err := tt.setupRepo(manager); err != nil {
					t.Fatalf("Failed to setup repo: %v", err)
				}
			}

			// Run verification
			result, err := verifyRepository(manager, false, nil)
			if err != nil {
				t.Fatalf("Verification failed: %v", err)
			}

			// Check expectations
			if result.HasErrors != tt.wantErrors {
				t.Errorf("HasErrors = %v, want %v", result.HasErrors, tt.wantErrors)
			}
			if result.HasWarnings != tt.wantWarnings {
				t.Errorf("HasWarnings = %v, want %v", result.HasWarnings, tt.wantWarnings)
			}
		})
	}
}

func TestRepoVerifyWithFix(t *testing.T) {
	// Create temp directory for test repository
	repoDir := t.TempDir()

	// Set AIMGR_REPO_PATH to use test directory
	oldEnv := os.Getenv("AIMGR_REPO_PATH")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("AIMGR_REPO_PATH", oldEnv)
		} else {
			_ = os.Unsetenv("AIMGR_REPO_PATH")
		}
	}()
	_ = os.Setenv("AIMGR_REPO_PATH", repoDir)

	// Initialize repository
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Create a resource without metadata
	commandPath := filepath.Join(repoDir, "commands", "test-cmd")
	if err := os.WriteFile(commandPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Run verification without fix - should have warnings
	result, err := verifyRepository(manager, false, nil)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if !result.HasWarnings {
		t.Error("Expected warnings for missing metadata")
	}

	// Run verification with fix
	result, err = verifyRepository(manager, true, nil)
	if err != nil {
		t.Fatalf("Verification with fix failed: %v", err)
	}

	// Verify metadata was created
	meta, err := manager.GetMetadata("test-cmd", resource.Command)
	if err != nil {
		t.Errorf("Metadata was not created by fix: %v", err)
	}

	if meta.Name != "test-cmd" {
		t.Errorf("Metadata name = %v, want test-cmd", meta.Name)
	}
}

func TestFormatResourceReference(t *testing.T) {
	tests := []struct {
		name         string
		resourceType resource.ResourceType
		resourceName string
		want         string
	}{
		{
			name:         "command reference",
			resourceType: resource.Command,
			resourceName: "test-cmd",
			want:         "command/test-cmd",
		},
		{
			name:         "skill reference",
			resourceType: resource.Skill,
			resourceName: "my-skill",
			want:         "skill/my-skill",
		},
		{
			name:         "agent reference",
			resourceType: resource.Agent,
			resourceName: "my-agent",
			want:         "agent/my-agent",
		},
		{
			name:         "package reference",
			resourceType: resource.PackageType,
			resourceName: "my-package",
			want:         "package/my-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatResourceReference(tt.resourceType, tt.resourceName)
			if got != tt.want {
				t.Errorf("formatResourceReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputVerifyResults(t *testing.T) {
	tests := []struct {
		name   string
		result *VerifyResult
		format output.Format
	}{
		{
			name: "empty result in table format",
			result: &VerifyResult{
				HasErrors:   false,
				HasWarnings: false,
			},
			format: output.Table,
		},
		{
			name: "result with issues in table format",
			result: &VerifyResult{
				ResourcesWithoutMetadata: []ResourceIssue{
					{Name: "test-cmd", Type: resource.Command, Path: "/test/path"},
				},
				HasErrors:   false,
				HasWarnings: true,
			},
			format: output.Table,
		},
		{
			name: "result in JSON format",
			result: &VerifyResult{
				HasErrors:   false,
				HasWarnings: false,
			},
			format: output.JSON,
		},
		{
			name: "result in YAML format",
			result: &VerifyResult{
				HasErrors:   false,
				HasWarnings: false,
			},
			format: output.YAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic or error
			err := outputVerifyResults(tt.result, tt.format, false)
			if err != nil {
				t.Errorf("outputVerifyResults() error = %v", err)
			}
		})
	}
}

func TestVerifyRepository_TypeMismatch(t *testing.T) {
	// Create temp directory for test repository
	repoDir := t.TempDir()

	// Set AIMGR_REPO_PATH to use test directory
	oldEnv := os.Getenv("AIMGR_REPO_PATH")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("AIMGR_REPO_PATH", oldEnv)
		} else {
			_ = os.Unsetenv("AIMGR_REPO_PATH")
		}
	}()
	_ = os.Setenv("AIMGR_REPO_PATH", repoDir)

	// Initialize repository
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Create a command
	commandPath := filepath.Join(repoDir, "commands", "test-cmd")
	if err := os.WriteFile(commandPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Create metadata with wrong type (skill instead of command)
	meta := &metadata.ResourceMetadata{
		Name:       "test-cmd",
		Type:       resource.Skill, // Wrong type!
		SourceType: "local",
		SourceURL:  "file:///test/source",
	}
	if err := metadata.Save(meta, repoDir, "local"); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Run verification
	result, err := verifyRepository(manager, false, nil)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	// Should have type mismatch error
	if !result.HasErrors {
		t.Error("Expected errors for type mismatch")
	}

	if len(result.TypeMismatches) == 0 {
		t.Error("Expected type mismatch to be detected")
	}
}
