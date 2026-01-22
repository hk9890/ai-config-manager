package cmd

import (
	"os"
	"path/filepath"
	"testing"

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
