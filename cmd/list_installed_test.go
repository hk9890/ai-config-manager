package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

func TestIsInstalledInTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create symlink for a command
	commandDir := filepath.Join(tmpDir, ".claude", "commands")
	if err := os.MkdirAll(commandDir, 0755); err != nil {
		t.Fatalf("failed to create command dir: %v", err)
	}

	// Create a dummy target file
	targetFile := filepath.Join(tmpDir, "target-command.md")
	if err := os.WriteFile(targetFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	symlinkPath := filepath.Join(commandDir, "test-command.md")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test: symlink should be detected
	if !isInstalledInTool(tmpDir, "test-command", resource.Command, tools.Claude) {
		t.Error("expected command to be installed in Claude, but it wasn't detected")
	}

	// Test: non-existent resource should not be detected
	if isInstalledInTool(tmpDir, "non-existent", resource.Command, tools.Claude) {
		t.Error("expected non-existent command to not be installed, but it was detected")
	}

	// Test: tool that doesn't support the resource type should return false
	if isInstalledInTool(tmpDir, "test-command", resource.Command, tools.Copilot) {
		t.Error("expected Copilot to not support commands, but resource was detected")
	}
}

func TestIsInstalledInTool_IgnoresRegularFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file (not a symlink)
	commandDir := filepath.Join(tmpDir, ".claude", "commands")
	if err := os.MkdirAll(commandDir, 0755); err != nil {
		t.Fatalf("failed to create command dir: %v", err)
	}

	regularFile := filepath.Join(commandDir, "regular-command.md")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create regular file: %v", err)
	}

	// Test: regular file should NOT be detected (only symlinks count)
	if isInstalledInTool(tmpDir, "regular-command", resource.Command, tools.Claude) {
		t.Error("expected regular file to be ignored, but it was detected")
	}
}

func TestBuildResourceInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create symlinks in both Claude and OpenCode
	claudeDir := filepath.Join(tmpDir, ".claude", "skills")
	opencodeDir := filepath.Join(tmpDir, ".opencode", "skills")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	// Create target directory
	targetDir := filepath.Join(tmpDir, "target-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	// Create symlinks
	claudeSymlink := filepath.Join(claudeDir, "test-skill")
	opencodeSymlink := filepath.Join(opencodeDir, "test-skill")
	if err := os.Symlink(targetDir, claudeSymlink); err != nil {
		t.Fatalf("failed to create claude symlink: %v", err)
	}
	if err := os.Symlink(targetDir, opencodeSymlink); err != nil {
		t.Fatalf("failed to create opencode symlink: %v", err)
	}

	// Create test resource
	resources := []resource.Resource{
		{
			Type:        resource.Skill,
			Name:        "test-skill",
			Description: "Test skill description",
			Version:     "1.0.0",
		},
	}

	detectedTools := []tools.Tool{tools.Claude, tools.OpenCode}
	infos := buildResourceInfo(resources, tmpDir, detectedTools)

	// Verify results
	if len(infos) != 1 {
		t.Fatalf("expected 1 resource info, got %d", len(infos))
	}

	info := infos[0]
	if info.Name != "test-skill" {
		t.Errorf("expected name 'test-skill', got '%s'", info.Name)
	}
	if len(info.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(info.Targets))
	}
	if info.Targets[0] != "claude" || info.Targets[1] != "opencode" {
		t.Errorf("expected targets [claude, opencode], got %v", info.Targets)
	}
}

func TestBuildResourceInfo_PartialInstallation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create symlink only in Claude (not OpenCode)
	claudeDir := filepath.Join(tmpDir, ".claude", "skills")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create command dir: %v", err)
	}

	targetDir := filepath.Join(tmpDir, "target-skill")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	claudeSymlink := filepath.Join(claudeDir, "test-skill")
	if err := os.Symlink(targetDir, claudeSymlink); err != nil {
		t.Fatalf("failed to create claude symlink: %v", err)
	}

	resources := []resource.Resource{
		{
			Type: resource.Skill,
			Name: "test-skill",
		},
	}

	detectedTools := []tools.Tool{tools.Claude, tools.OpenCode}
	infos := buildResourceInfo(resources, tmpDir, detectedTools)

	if len(infos) != 1 {
		t.Fatalf("expected 1 resource info, got %d", len(infos))
	}

	info := infos[0]
	if len(info.Targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(info.Targets))
	}
	if info.Targets[0] != "claude" {
		t.Errorf("expected target [claude], got %v", info.Targets)
	}
}

// TestListInstalled_PatternFiltering tests pattern matching functionality
func TestListInstalled_PatternFiltering(t *testing.T) {
	// Create mock resources for testing pattern matching
	testResources := []resource.Resource{
		{Type: resource.Skill, Name: "pdf-processing", Description: "PDF processing skill"},
		{Type: resource.Skill, Name: "test-runner", Description: "Test runner skill"},
		{Type: resource.Command, Name: "test-command", Description: "Test command"},
		{Type: resource.Agent, Name: "debug-agent", Description: "Debug agent"},
	}

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		expectedNames []string
		shouldError   bool
	}{
		{
			name:          "filter by skill type",
			pattern:       "skill/*",
			expectedCount: 2,
			expectedNames: []string{"pdf-processing", "test-runner"},
		},
		{
			name:          "filter by command type",
			pattern:       "command/*",
			expectedCount: 1,
			expectedNames: []string{"test-command"},
		},
		{
			name:          "filter by agent type",
			pattern:       "agent/*",
			expectedCount: 1,
			expectedNames: []string{"debug-agent"},
		},
		{
			name:          "filter by name prefix",
			pattern:       "test*",
			expectedCount: 2,
			expectedNames: []string{"test-runner", "test-command"},
		},
		{
			name:          "filter by name suffix",
			pattern:       "*agent",
			expectedCount: 1,
			expectedNames: []string{"debug-agent"},
		},
		{
			name:          "filter by name contains",
			pattern:       "*test*",
			expectedCount: 2,
			expectedNames: []string{"test-runner", "test-command"},
		},
		{
			name:          "combined type and name pattern",
			pattern:       "skill/pdf*",
			expectedCount: 1,
			expectedNames: []string{"pdf-processing"},
		},
		{
			name:          "no matches",
			pattern:       "skill/nonexistent*",
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Filter by pattern - this simulates the logic in list_installed.go
			var filtered []resource.Resource

			matcher, err := pattern.NewMatcher(tt.pattern)
			if err != nil {
				if !tt.shouldError {
					t.Fatalf("unexpected error creating matcher: %v", err)
				}
				return
			}

			for _, res := range testResources {
				if matcher.Match(&res) {
					filtered = append(filtered, res)
				}
			}

			// Verify count
			if len(filtered) != tt.expectedCount {
				t.Errorf("expected %d resources, got %d", tt.expectedCount, len(filtered))
			}

			// Verify names
			gotNames := make([]string, len(filtered))
			for i, res := range filtered {
				gotNames[i] = res.Name
			}

			if len(gotNames) != len(tt.expectedNames) {
				t.Errorf("expected names %v, got %v", tt.expectedNames, gotNames)
			} else {
				// Check that all expected names are present (order doesn't matter)
				expectedMap := make(map[string]bool)
				for _, name := range tt.expectedNames {
					expectedMap[name] = true
				}
				for _, name := range gotNames {
					if !expectedMap[name] {
						t.Errorf("unexpected resource name: %s", name)
					}
				}
			}
		})
	}
}

// TestListInstalled_OutputFormats tests buildResourceInfo function
func TestListInstalled_OutputFormats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test resources
	resources := []resource.Resource{
		{
			Type:        resource.Command,
			Name:        "test-cmd",
			Description: "Test command",
			Version:     "1.0.0",
		},
	}

	// Test building resource info (used by output formatters)
	infos := buildResourceInfo(resources, tmpDir, []tools.Tool{tools.Claude})
	if len(infos) != 1 {
		t.Fatalf("expected 1 resource info, got %d", len(infos))
	}

	// Verify resource info structure
	if infos[0].Name != "test-cmd" {
		t.Errorf("expected name 'test-cmd', got '%s'", infos[0].Name)
	}
	if infos[0].Type != resource.Command {
		t.Errorf("expected type 'command', got '%s'", infos[0].Type)
	}
	if infos[0].Description != "Test command" {
		t.Errorf("expected description 'Test command', got '%s'", infos[0].Description)
	}
	if infos[0].Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", infos[0].Version)
	}

	_ = bytes.Buffer{} // Avoid unused import error
}
