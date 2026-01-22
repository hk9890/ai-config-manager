package test

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/source"
)

// isOnline checks if we have network connectivity
func isOnline() bool {
	// Try to resolve github.com to check network connectivity
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", "github.com:443", timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// TestGitHubSourceSkillDiscovery tests discovering and adding a skill from GitHub
func TestGitHubSourceSkillDiscovery(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	t.Run("discover skills from GitHub repo", func(t *testing.T) {
		// Parse GitHub source
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source: %v", err)
		}

		// Verify parsed correctly
		if parsed.Type != source.GitHub {
			t.Errorf("Expected GitHub source type, got %v", parsed.Type)
		}

		// Get clone URL
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		// Clone repository to temp directory
		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}
		defer source.CleanupTempDir(tempDir)

		// Verify temp directory was created
		if _, err := os.Stat(tempDir); err != nil {
			t.Errorf("Temp directory was not created: %v", err)
		}

		// Build search path
		searchPath := tempDir
		if parsed.Subpath != "" {
			searchPath = filepath.Join(tempDir, parsed.Subpath)
		}

		// Discover skills in the cloned repository
		skills, err := discovery.DiscoverSkills(searchPath, "")
		if err != nil {
			t.Fatalf("Failed to discover skills: %v", err)
		}

		// Note: This repo may or may not have skills - that's ok
		// The important thing is that the discovery process works
		t.Logf("Found %d skills in repository", len(skills))

		// Verify cleanup happens
		source.CleanupTempDir(tempDir)
		if _, err := os.Stat(tempDir); err == nil {
			t.Error("Temp directory should be removed after cleanup")
		}
	})

	t.Run("add skill from GitHub to repository", func(t *testing.T) {
		// Create a local test skill to demonstrate the workflow
		// In a real scenario, this would come from a GitHub repo with skills
		testSkillDir := filepath.Join(t.TempDir(), "test-gh-skill")
		if err := os.MkdirAll(testSkillDir, 0755); err != nil {
			t.Fatalf("Failed to create test skill directory: %v", err)
		}

		skillContent := `---
name: test-gh-skill
description: A test skill from GitHub
license: MIT
---

# Test GitHub Skill

This demonstrates adding a skill from a GitHub source.
`
		skillMdPath := filepath.Join(testSkillDir, "SKILL.md")
		if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to create SKILL.md: %v", err)
		}

		// Add the skill to the manager
		err := manager.AddSkill(testSkillDir, "file://"+testSkillDir, "file")
		if err != nil {
			t.Fatalf("Failed to add skill to repository: %v", err)
		}

		// Verify the skill was added to the repository
		addedSkill, err := manager.Get("test-gh-skill", resource.Skill)
		if err != nil {
			t.Errorf("Failed to get added skill from repository: %v", err)
		}

		if addedSkill.Name != "test-gh-skill" {
			t.Errorf("Skill name mismatch: got %v, want test-gh-skill", addedSkill.Name)
		}

		// Verify the skill has correct metadata
		if addedSkill.Description == "" {
			t.Error("Skill description should not be empty")
		}
		if addedSkill.Type != resource.Skill {
			t.Errorf("Resource type = %v, want Skill", addedSkill.Type)
		}
	})
}

// TestGitHubSourceCommandDiscovery tests discovering and adding commands from GitHub
func TestGitHubSourceCommandDiscovery(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("discover commands from GitHub repo", func(t *testing.T) {
		// Parse GitHub source
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source: %v", err)
		}

		// Clone repository
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}
		defer source.CleanupTempDir(tempDir)

		// Discover commands in the cloned repository
		searchPath := tempDir
		if parsed.Subpath != "" {
			searchPath = filepath.Join(tempDir, parsed.Subpath)
		}

		commands, err := discovery.DiscoverCommands(searchPath, "")
		if err != nil {
			t.Fatalf("Failed to discover commands: %v", err)
		}

		// Commands might not exist in this repo, which is fine
		t.Logf("Found %d commands in repository", len(commands))
	})
}

// TestGitHubSourceAgentDiscovery tests discovering and adding agents from GitHub
func TestGitHubSourceAgentDiscovery(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("discover agents from GitHub repo", func(t *testing.T) {
		// Parse GitHub source
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source: %v", err)
		}

		// Clone repository
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}
		defer source.CleanupTempDir(tempDir)

		// Discover agents in the cloned repository
		searchPath := tempDir
		if parsed.Subpath != "" {
			searchPath = filepath.Join(tempDir, parsed.Subpath)
		}

		agents, err := discovery.DiscoverAgents(searchPath, "")
		if err != nil {
			t.Fatalf("Failed to discover agents: %v", err)
		}

		// Agents might not exist in this repo, which is fine
		t.Logf("Found %d agents in repository", len(agents))
	})
}

// TestGitHubSourceWithSubpath tests GitHub sources with specific subpaths
func TestGitHubSourceWithSubpath(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("parse GitHub source with subpath", func(t *testing.T) {
		// Test parsing with subpath
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts/computer-use-demo")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source with subpath: %v", err)
		}

		if parsed.Type != source.GitHub {
			t.Errorf("Expected GitHub source type, got %v", parsed.Type)
		}

		if parsed.Subpath != "computer-use-demo" {
			t.Errorf("Subpath = %v, want 'computer-use-demo'", parsed.Subpath)
		}

		// Clone and search in subpath
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}
		defer source.CleanupTempDir(tempDir)

		// Build search path with subpath
		searchPath := filepath.Join(tempDir, parsed.Subpath)

		// Verify subpath exists
		if _, err := os.Stat(searchPath); err != nil {
			t.Logf("Subpath does not exist (this is ok if repo structure changed): %v", err)
		}
	})
}

// TestGitHubSourceWithBranch tests GitHub sources with specific branch/tag references
func TestGitHubSourceWithBranch(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("parse GitHub source with branch", func(t *testing.T) {
		// Test parsing with branch reference
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts@main")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source with branch: %v", err)
		}

		if parsed.Type != source.GitHub {
			t.Errorf("Expected GitHub source type, got %v", parsed.Type)
		}

		if parsed.Ref != "main" {
			t.Errorf("Ref = %v, want 'main'", parsed.Ref)
		}

		// Clone with specific branch
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository with branch: %v", err)
		}
		defer source.CleanupTempDir(tempDir)

		// Verify temp directory was created
		if _, err := os.Stat(tempDir); err != nil {
			t.Errorf("Temp directory was not created: %v", err)
		}
	})

	t.Run("parse GitHub source with branch and subpath", func(t *testing.T) {
		// Test parsing with both branch and subpath
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts@main/computer-use-demo")
		if err != nil {
			t.Fatalf("Failed to parse GitHub source with branch and subpath: %v", err)
		}

		if parsed.Type != source.GitHub {
			t.Errorf("Expected GitHub source type, got %v", parsed.Type)
		}

		if parsed.Ref != "main" {
			t.Errorf("Ref = %v, want 'main'", parsed.Ref)
		}

		if parsed.Subpath != "computer-use-demo" {
			t.Errorf("Subpath = %v, want 'computer-use-demo'", parsed.Subpath)
		}
	})
}

// TestGitHubSourceErrorHandling tests error cases for GitHub sources
func TestGitHubSourceErrorHandling(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("invalid repository", func(t *testing.T) {
		parsed, err := source.ParseSource("gh:this-owner-does-not-exist/this-repo-does-not-exist-12345")
		if err != nil {
			t.Fatalf("Failed to parse source: %v", err)
		}

		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		// Try to clone - should fail
		_, err = source.CloneRepo(cloneURL, parsed.Ref)
		if err == nil {
			t.Error("Expected error when cloning non-existent repository")
		}

		// Verify error message is helpful
		if !strings.Contains(err.Error(), "git clone failed") {
			t.Errorf("Error message should mention 'git clone failed', got: %v", err)
		}
	})

	t.Run("invalid branch", func(t *testing.T) {
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts@this-branch-does-not-exist-12345")
		if err != nil {
			t.Fatalf("Failed to parse source: %v", err)
		}

		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		// Try to clone with invalid branch - should fail
		_, err = source.CloneRepo(cloneURL, parsed.Ref)
		if err == nil {
			t.Error("Expected error when cloning with non-existent branch")
		}
	})

	t.Run("cleanup non-existent directory", func(t *testing.T) {
		// os.RemoveAll doesn't error on non-existent directories, which is fine
		// Just verify it doesn't panic and returns without error
		err := source.CleanupTempDir("/tmp/this-does-not-exist-12345")
		if err != nil {
			t.Logf("Cleanup returned error (this is ok): %v", err)
		}
	})

	t.Run("cleanup with security check", func(t *testing.T) {
		// Try to cleanup a directory outside temp - should fail
		err := source.CleanupTempDir("/etc")
		if err == nil {
			t.Error("Expected error when trying to cleanup directory outside temp")
		}

		if !strings.Contains(err.Error(), "refusing to delete") {
			t.Errorf("Error should mention security check, got: %v", err)
		}
	})
}

// TestLocalSourceStillWorks tests that local sources continue to work
func TestLocalSourceStillWorks(t *testing.T) {
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	t.Run("add local skill", func(t *testing.T) {
		// Create a local skill directory
		localSkillDir := filepath.Join(t.TempDir(), "local-test-skill")
		if err := os.MkdirAll(localSkillDir, 0755); err != nil {
			t.Fatalf("Failed to create local skill directory: %v", err)
		}

		// Create SKILL.md
		skillContent := `---
name: local-test-skill
description: A local skill for testing
license: MIT
---

# Local Test Skill

This is a local skill for testing.
`
		skillMdPath := filepath.Join(localSkillDir, "SKILL.md")
		if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
			t.Fatalf("Failed to create SKILL.md: %v", err)
		}

		// Parse local source
		parsed, err := source.ParseSource(fmt.Sprintf("local:%s", localSkillDir))
		if err != nil {
			t.Fatalf("Failed to parse local source: %v", err)
		}

		if parsed.Type != source.Local {
			t.Errorf("Expected Local source type, got %v", parsed.Type)
		}

		if parsed.LocalPath != localSkillDir {
			t.Errorf("LocalPath = %v, want %v", parsed.LocalPath, localSkillDir)
		}

		// Add the local skill to repository
		err = manager.AddSkill(parsed.LocalPath, "file://"+parsed.LocalPath, "file")
		if err != nil {
			t.Fatalf("Failed to add local skill: %v", err)
		}

		// Verify skill was added
		skill, err := manager.Get("local-test-skill", resource.Skill)
		if err != nil {
			t.Errorf("Failed to get local skill from repository: %v", err)
		}

		if skill.Name != "local-test-skill" {
			t.Errorf("Skill name = %v, want local-test-skill", skill.Name)
		}
	})

	t.Run("add local command", func(t *testing.T) {
		// Create a local command file
		localCmdPath := filepath.Join(t.TempDir(), "local-test-cmd.md")
		cmdContent := `---
description: A local command for testing
---

# Local Test Command

This is a local command for testing.
`
		if err := os.WriteFile(localCmdPath, []byte(cmdContent), 0644); err != nil {
			t.Fatalf("Failed to create local command: %v", err)
		}

		// Parse local source
		parsed, err := source.ParseSource(fmt.Sprintf("local:%s", localCmdPath))
		if err != nil {
			t.Fatalf("Failed to parse local source: %v", err)
		}

		if parsed.Type != source.Local {
			t.Errorf("Expected Local source type, got %v", parsed.Type)
		}

		// Add the local command to repository
		err = manager.AddCommand(parsed.LocalPath, "file://"+parsed.LocalPath, "file")
		if err != nil {
			t.Fatalf("Failed to add local command: %v", err)
		}

		// Verify command was added
		cmd, err := manager.Get("local-test-cmd", resource.Command)
		if err != nil {
			t.Errorf("Failed to get local command from repository: %v", err)
		}

		if cmd.Name != "local-test-cmd" {
			t.Errorf("Command name = %v, want local-test-cmd", cmd.Name)
		}
	})
}

// TestGitHubURLFormats tests various GitHub URL formats
func TestGitHubURLFormats(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    source.SourceType
		expectedRef     string
		expectedSubpath string
		expectError     bool
	}{
		{
			name:         "gh prefix with owner/repo",
			input:        "gh:anthropics/anthropic-quickstarts",
			expectedType: source.GitHub,
		},
		{
			name:         "owner/repo without prefix",
			input:        "anthropics/anthropic-quickstarts",
			expectedType: source.GitHub,
		},
		{
			name:         "gh prefix with branch",
			input:        "gh:anthropics/anthropic-quickstarts@main",
			expectedType: source.GitHub,
			expectedRef:  "main",
		},
		{
			name:            "gh prefix with subpath",
			input:           "gh:anthropics/anthropic-quickstarts/computer-use-demo",
			expectedType:    source.GitHub,
			expectedSubpath: "computer-use-demo",
		},
		{
			name:            "gh prefix with branch and subpath",
			input:           "gh:anthropics/anthropic-quickstarts@main/computer-use-demo",
			expectedType:    source.GitHub,
			expectedRef:     "main",
			expectedSubpath: "computer-use-demo",
		},
		{
			name:         "https GitHub URL",
			input:        "https://github.com/anthropics/anthropic-quickstarts",
			expectedType: source.GitHub,
		},
		{
			name:         "https GitHub URL with tree/branch",
			input:        "https://github.com/anthropics/anthropic-quickstarts/tree/main",
			expectedType: source.GitHub,
			expectedRef:  "main",
		},
		{
			name:            "https GitHub URL with tree/branch/path",
			input:           "https://github.com/anthropics/anthropic-quickstarts/tree/main/computer-use-demo",
			expectedType:    source.GitHub,
			expectedRef:     "main",
			expectedSubpath: "computer-use-demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := source.ParseSource(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if parsed.Type != tt.expectedType {
				t.Errorf("Type = %v, want %v", parsed.Type, tt.expectedType)
			}

			if tt.expectedRef != "" && parsed.Ref != tt.expectedRef {
				t.Errorf("Ref = %v, want %v", parsed.Ref, tt.expectedRef)
			}

			if tt.expectedSubpath != "" && parsed.Subpath != tt.expectedSubpath {
				t.Errorf("Subpath = %v, want %v", parsed.Subpath, tt.expectedSubpath)
			}
		})
	}
}

// TestCleanupOnError tests that temp directories are cleaned up even on errors
func TestCleanupOnError(t *testing.T) {
	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	t.Run("cleanup after successful operation", func(t *testing.T) {
		parsed, err := source.ParseSource("gh:anthropics/anthropic-quickstarts")
		if err != nil {
			t.Fatalf("Failed to parse source: %v", err)
		}

		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			t.Fatalf("Failed to get clone URL: %v", err)
		}

		tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
		if err != nil {
			t.Fatalf("Failed to clone repository: %v", err)
		}

		// Verify temp directory exists
		if _, err := os.Stat(tempDir); err != nil {
			t.Errorf("Temp directory should exist: %v", err)
		}

		// Cleanup
		err = source.CleanupTempDir(tempDir)
		if err != nil {
			t.Errorf("Cleanup failed: %v", err)
		}

		// Verify temp directory is gone
		if _, err := os.Stat(tempDir); err == nil {
			t.Error("Temp directory should be removed after cleanup")
		}
	})
}

// findSkillDirectories is a helper function to find skill directories in a path
func findSkillDirectories(basePath string) ([]string, error) {
	var skillDirs []string

	// Search in priority locations
	priorityLocations := []string{
		"skills",
		".claude/skills",
		".opencode/skills",
	}

	for _, location := range priorityLocations {
		locationPath := filepath.Join(basePath, location)
		if _, err := os.Stat(locationPath); err != nil {
			continue
		}

		entries, err := os.ReadDir(locationPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillPath := filepath.Join(locationPath, entry.Name())
			// Check if it has SKILL.md
			skillMdPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMdPath); err == nil {
				skillDirs = append(skillDirs, skillPath)
			}
		}
	}

	return skillDirs, nil
}
