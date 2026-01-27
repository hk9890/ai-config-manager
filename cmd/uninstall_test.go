package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

func TestParseResourceArg(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantType    resource.ResourceType
		wantName    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid skill",
			arg:      "skill/my-skill",
			wantType: resource.Skill,
			wantName: "my-skill",
			wantErr:  false,
		},
		{
			name:     "valid command",
			arg:      "command/my-command",
			wantType: resource.Command,
			wantName: "my-command",
			wantErr:  false,
		},
		{
			name:     "valid agent",
			arg:      "agent/my-agent",
			wantType: resource.Agent,
			wantName: "my-agent",
			wantErr:  false,
		},
		{
			name:        "invalid format - no slash",
			arg:         "skill",
			wantErr:     true,
			errContains: "must be 'type/name'",
		},
		{
			name:        "invalid format - empty name",
			arg:         "skill/",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "invalid type",
			arg:         "invalid/name",
			wantErr:     true,
			errContains: "must be one of 'skill', 'command', or 'agent'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotName, err := parseResourceArg(tt.arg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseResourceArg() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("parseResourceArg() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("parseResourceArg() unexpected error = %v", err)
				return
			}

			if gotType != tt.wantType {
				t.Errorf("parseResourceArg() type = %v, want %v", gotType, tt.wantType)
			}
			if gotName != tt.wantName {
				t.Errorf("parseResourceArg() name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestProcessUninstall(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a temporary repo
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "commands"), 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Create a test project
	projectDir := filepath.Join(tempDir, "project")
	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}

	// Create a test command file in repo
	testCommandPath := filepath.Join(repoDir, "commands", "test-command.md")
	testContent := []byte("---\ndescription: Test command\n---\n# Test")
	if err := os.WriteFile(testCommandPath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test command: %v", err)
	}

	// Create symlink pointing to repo
	symlinkPath := filepath.Join(commandsDir, "test-command.md")
	if err := os.Symlink(testCommandPath, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test uninstalling the command
	result := processUninstall("command/test-command", projectDir, repoDir, []tools.Tool{tools.Claude})

	if !result.success {
		t.Errorf("processUninstall() failed: %v", result.message)
	}

	if len(result.toolsRemoved) != 1 {
		t.Errorf("processUninstall() removed from %d tools, want 1", len(result.toolsRemoved))
	}

	// Verify symlink was removed
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Errorf("processUninstall() symlink still exists after uninstall")
	}
}

func TestProcessUninstall_NotManaged(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a temporary repo
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Create a test project
	projectDir := filepath.Join(tempDir, "project")
	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}

	// Create a test command file OUTSIDE repo
	otherDir := filepath.Join(tempDir, "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("failed to create other dir: %v", err)
	}
	testCommandPath := filepath.Join(otherDir, "test-command.md")
	testContent := []byte("---\ndescription: Test command\n---\n# Test")
	if err := os.WriteFile(testCommandPath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test command: %v", err)
	}

	// Create symlink pointing to other location (not repo)
	symlinkPath := filepath.Join(commandsDir, "test-command.md")
	if err := os.Symlink(testCommandPath, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test uninstalling the command
	result := processUninstall("command/test-command", projectDir, repoDir, []tools.Tool{tools.Claude})

	// Should be skipped since it doesn't point to repo
	if !result.skipped {
		t.Errorf("processUninstall() expected skipped=true for non-managed symlink")
	}

	// Verify symlink was NOT removed
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Errorf("processUninstall() removed non-managed symlink, should have skipped it")
	}
}

func TestProcessUninstall_NotSymlink(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a temporary repo
	repoDir := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Create a test project
	projectDir := filepath.Join(tempDir, "project")
	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}

	// Create a regular file (not a symlink)
	filePath := filepath.Join(commandsDir, "test-command.md")
	testContent := []byte("---\ndescription: Test command\n---\n# Test")
	if err := os.WriteFile(filePath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test uninstalling the command
	result := processUninstall("command/test-command", projectDir, repoDir, []tools.Tool{tools.Claude})

	// Should be skipped since it's not a symlink
	if !result.skipped {
		t.Errorf("processUninstall() expected skipped=true for non-symlink file")
	}

	// Verify file was NOT removed
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("processUninstall() removed non-symlink file, should have skipped it")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test helpers for uninstall pattern tests

// setupTestRepoForUninstall creates a test repo and installs some resources
func setupTestRepoForUninstall(t *testing.T) (repoPath string, projectPath string, cleanup func()) {
	t.Helper()

	// Create temp repo
	repoDir := t.TempDir()

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(repoDir, "commands"), 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create skills dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "agents"), 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create test commands
	testCommands := []string{"test-command", "pdf-command", "review-command"}
	for _, name := range testCommands {
		content := []byte(fmt.Sprintf("---\ndescription: %s\n---\n# %s", name, name))
		path := filepath.Join(repoDir, "commands", name+".md")
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("failed to create command %s: %v", name, err)
		}
	}

	// Create test skills
	testSkills := []string{"pdf-processing", "pdf-extraction", "image-processing", "test-skill"}
	for _, name := range testSkills {
		skillDir := filepath.Join(repoDir, "skills", name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("failed to create skill dir %s: %v", name, err)
		}
		content := []byte(fmt.Sprintf("---\ndescription: %s\n---\n# %s", name, name))
		path := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("failed to create skill %s: %v", name, err)
		}
	}

	// Create test agents
	testAgents := []string{"code-reviewer", "test-agent", "doc-generator"}
	for _, name := range testAgents {
		content := []byte(fmt.Sprintf("---\ndescription: %s\n---\n# %s", name, name))
		path := filepath.Join(repoDir, "agents", name+".md")
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("failed to create agent %s: %v", name, err)
		}
	}

	// Create project directory
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "commands"), 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude", "agents"), 0755); err != nil {
		t.Fatalf("failed to create project dirs: %v", err)
	}

	cleanup = func() {
		// Cleanup is automatic with t.TempDir()
	}

	return repoDir, projectDir, cleanup
}

// installSymlink creates a symlink from project to repo
func installSymlink(t *testing.T, projectPath, repoPath string, resourceType resource.ResourceType, name string) {
	t.Helper()

	var srcPath, dstPath string
	switch resourceType {
	case resource.Command:
		srcPath = filepath.Join(repoPath, "commands", name+".md")
		dstPath = filepath.Join(projectPath, ".claude", "commands", name+".md")
	case resource.Skill:
		srcPath = filepath.Join(repoPath, "skills", name)
		dstPath = filepath.Join(projectPath, ".claude", "skills", name)
	case resource.Agent:
		srcPath = filepath.Join(repoPath, "agents", name+".md")
		dstPath = filepath.Join(projectPath, ".claude", "agents", name+".md")
	}

	if err := os.Symlink(srcPath, dstPath); err != nil {
		t.Fatalf("failed to create symlink for %s/%s: %v", resourceType, name, err)
	}
}

// Test Pattern Expansion for Uninstall

func TestExpandUninstallPattern_AllSkills(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install some skills
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-extraction")
	installSymlink(t, projectPath, repoPath, resource.Skill, "test-skill")

	// Expand "skill/*" pattern
	detectedTools := []tools.Tool{tools.Claude}
	matches, err := expandUninstallPattern(projectPath, "skill/*", detectedTools)
	if err != nil {
		t.Fatalf("expandUninstallPattern failed: %v", err)
	}

	// Should match 3 installed skills
	if len(matches) != 3 {
		t.Errorf("expandUninstallPattern(skill/*) returned %d matches, want 3", len(matches))
	}
}

func TestExpandUninstallPattern_PrefixMatch(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install some skills
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-extraction")
	installSymlink(t, projectPath, repoPath, resource.Skill, "test-skill")

	// Expand "skill/pdf*" pattern
	detectedTools := []tools.Tool{tools.Claude}
	matches, err := expandUninstallPattern(projectPath, "skill/pdf*", detectedTools)
	if err != nil {
		t.Fatalf("expandUninstallPattern failed: %v", err)
	}

	// Should match 2 pdf skills
	if len(matches) != 2 {
		t.Errorf("expandUninstallPattern(skill/pdf*) returned %d matches, want 2", len(matches))
	}

	// Verify matches
	expectedMatches := map[string]bool{
		"skill/pdf-processing": true,
		"skill/pdf-extraction": true,
	}
	for _, match := range matches {
		if !expectedMatches[match] {
			t.Errorf("unexpected match: %q", match)
		}
	}
}

func TestExpandUninstallPattern_AllTypes(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install resources with "test" in name across types
	installSymlink(t, projectPath, repoPath, resource.Command, "test-command")
	installSymlink(t, projectPath, repoPath, resource.Skill, "test-skill")
	installSymlink(t, projectPath, repoPath, resource.Agent, "test-agent")

	// Expand "*test*" pattern (matches across all types)
	detectedTools := []tools.Tool{tools.Claude}
	matches, err := expandUninstallPattern(projectPath, "*test*", detectedTools)
	if err != nil {
		t.Fatalf("expandUninstallPattern failed: %v", err)
	}

	// Should match all 3 test resources
	if len(matches) != 3 {
		t.Errorf("expandUninstallPattern(*test*) returned %d matches, want 3", len(matches))
		for _, m := range matches {
			t.Logf("  match: %s", m)
		}
	}
}

func TestExpandUninstallPattern_NoMatches(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install one skill
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")

	// Expand pattern that matches nothing
	detectedTools := []tools.Tool{tools.Claude}
	matches, err := expandUninstallPattern(projectPath, "skill/nomatch*", detectedTools)
	if err != nil {
		t.Fatalf("expandUninstallPattern failed: %v", err)
	}

	// Should return empty slice
	if len(matches) != 0 {
		t.Errorf("expandUninstallPattern(skill/nomatch*) returned %d matches, want 0", len(matches))
	}
}

func TestExpandUninstallPattern_ExactName(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install a skill
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")

	// Expand exact name (not a pattern)
	detectedTools := []tools.Tool{tools.Claude}
	matches, err := expandUninstallPattern(projectPath, "skill/pdf-processing", detectedTools)
	if err != nil {
		t.Fatalf("expandUninstallPattern failed: %v", err)
	}

	// Should return the exact name as-is
	if len(matches) != 1 {
		t.Fatalf("expandUninstallPattern(skill/pdf-processing) returned %d matches, want 1", len(matches))
	}
	if matches[0] != "skill/pdf-processing" {
		t.Errorf("expandUninstallPattern(skill/pdf-processing) = %q, want 'skill/pdf-processing'", matches[0])
	}
}

func TestScanToolDir_Commands(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install some commands
	installSymlink(t, projectPath, repoPath, resource.Command, "test-command")
	installSymlink(t, projectPath, repoPath, resource.Command, "pdf-command")

	// Create matcher for "*command"
	matcher, err := pattern.NewMatcher("*command")
	if err != nil {
		t.Fatalf("NewMatcher failed: %v", err)
	}

	// Scan .claude/commands directory
	matches := scanToolDir(projectPath, ".claude/commands", resource.Command, matcher)

	// Should find both commands
	if len(matches) != 2 {
		t.Errorf("scanToolDir returned %d matches, want 2", len(matches))
	}
}

func TestScanToolDir_Skills(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install skills
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")
	installSymlink(t, projectPath, repoPath, resource.Skill, "image-processing")

	// Create matcher for "*processing"
	matcher, err := pattern.NewMatcher("*processing")
	if err != nil {
		t.Fatalf("NewMatcher failed: %v", err)
	}

	// Scan .claude/skills directory
	matches := scanToolDir(projectPath, ".claude/skills", resource.Skill, matcher)

	// Should find both skills
	if len(matches) != 2 {
		t.Errorf("scanToolDir returned %d matches, want 2", len(matches))
	}
}

func TestScanToolDir_NonExistentDirectory(t *testing.T) {
	projectPath := t.TempDir()

	// Create matcher
	matcher, err := pattern.NewMatcher("*")
	if err != nil {
		t.Fatalf("NewMatcher failed: %v", err)
	}

	// Scan non-existent directory
	matches := scanToolDir(projectPath, ".claude/commands", resource.Command, matcher)

	// Should return nil (no matches)
	if matches != nil {
		t.Errorf("scanToolDir on non-existent dir returned %v, want nil", matches)
	}
}

func TestDeduplicateStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no duplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "all duplicates",
			input: []string{"a", "a", "a"},
			want:  []string{"a"},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateStrings(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("deduplicateStrings() returned %d items, want %d", len(got), len(tt.want))
			}
			// Check all expected items are present
			gotMap := make(map[string]bool)
			for _, s := range got {
				gotMap[s] = true
			}
			for _, s := range tt.want {
				if !gotMap[s] {
					t.Errorf("deduplicateStrings() missing expected item %q", s)
				}
			}
		})
	}
}

func TestUninstallAll(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Install multiple resources
	installSymlink(t, projectPath, repoPath, resource.Command, "test-command")
	installSymlink(t, projectPath, repoPath, resource.Command, "pdf-command")
	installSymlink(t, projectPath, repoPath, resource.Skill, "pdf-processing")
	installSymlink(t, projectPath, repoPath, resource.Skill, "test-skill")
	installSymlink(t, projectPath, repoPath, resource.Agent, "code-reviewer")

	// Run uninstallAll
	err := uninstallAll(projectPath, repoPath, []tools.Tool{tools.Claude})
	if err != nil {
		t.Fatalf("uninstallAll() failed: %v", err)
	}

	// Verify all symlinks were removed
	commandsDir := filepath.Join(projectPath, ".claude", "commands")
	skillsDir := filepath.Join(projectPath, ".claude", "skills")
	agentsDir := filepath.Join(projectPath, ".claude", "agents")

	// Check commands
	if _, err := os.Lstat(filepath.Join(commandsDir, "test-command.md")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove test-command.md")
	}
	if _, err := os.Lstat(filepath.Join(commandsDir, "pdf-command.md")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove pdf-command.md")
	}

	// Check skills
	if _, err := os.Lstat(filepath.Join(skillsDir, "pdf-processing")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove pdf-processing")
	}
	if _, err := os.Lstat(filepath.Join(skillsDir, "test-skill")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove test-skill")
	}

	// Check agents
	if _, err := os.Lstat(filepath.Join(agentsDir, "code-reviewer.md")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove code-reviewer.md")
	}
}

func TestUninstallAll_OnlyRemovesManagedSymlinks(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	commandsDir := filepath.Join(projectPath, ".claude", "commands")

	// Install a managed symlink
	installSymlink(t, projectPath, repoPath, resource.Command, "test-command")

	// Create an unmanaged symlink (points outside repo)
	otherDir := t.TempDir()
	unmanagedFile := filepath.Join(otherDir, "unmanaged.md")
	if err := os.WriteFile(unmanagedFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create unmanaged file: %v", err)
	}
	unmanagedSymlink := filepath.Join(commandsDir, "unmanaged.md")
	if err := os.Symlink(unmanagedFile, unmanagedSymlink); err != nil {
		t.Fatalf("failed to create unmanaged symlink: %v", err)
	}

	// Create a regular file (not a symlink)
	regularFile := filepath.Join(commandsDir, "regular.md")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create regular file: %v", err)
	}

	// Run uninstallAll
	err := uninstallAll(projectPath, repoPath, []tools.Tool{tools.Claude})
	if err != nil {
		t.Fatalf("uninstallAll() failed: %v", err)
	}

	// Verify managed symlink was removed
	if _, err := os.Lstat(filepath.Join(commandsDir, "test-command.md")); !os.IsNotExist(err) {
		t.Error("uninstallAll() did not remove managed symlink")
	}

	// Verify unmanaged symlink was NOT removed
	if _, err := os.Lstat(unmanagedSymlink); err != nil {
		t.Error("uninstallAll() removed unmanaged symlink, should have skipped it")
	}

	// Verify regular file was NOT removed
	if _, err := os.Stat(regularFile); err != nil {
		t.Error("uninstallAll() removed regular file, should have skipped it")
	}
}

func TestUninstallAll_EmptyDirectory(t *testing.T) {
	repoPath, projectPath, cleanup := setupTestRepoForUninstall(t)
	defer cleanup()

	// Run uninstallAll on empty directories
	err := uninstallAll(projectPath, repoPath, []tools.Tool{tools.Claude})
	if err != nil {
		t.Fatalf("uninstallAll() failed on empty directories: %v", err)
	}
}
