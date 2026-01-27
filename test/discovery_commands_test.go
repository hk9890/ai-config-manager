package test

import (
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
)

func TestDiscoverCommands_StandardLocation(t *testing.T) {
	fixturePath := filepath.Join("..", "testdata", "repos", "commands-nested")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Should find build.md, test.md, nested/deploy.md
	if len(commands) < 3 {
		t.Errorf("Expected at least 3 commands, got %d", len(commands))
	}

	// Verify we found the expected commands
	// Note: deploy is nested, so its name is "nested/deploy"
	expectedNames := map[string]bool{
		"build":         true,
		"test":          true,
		"nested/deploy": true,
	}

	foundNames := make(map[string]bool)
	for _, cmd := range commands {
		foundNames[cmd.Name] = true
	}

	for name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("Expected to find command %q, but didn't", name)
		}
	}
}

func TestDiscoverCommands_NestedCommands(t *testing.T) {
	fixturePath := filepath.Join("..", "testdata", "repos", "commands-nested")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Verify nested command is found with full nested path
	foundNested := false
	for _, cmd := range commands {
		if cmd.Name == "nested/deploy" {
			foundNested = true
			if cmd.Description == "" {
				t.Errorf("Nested command has empty description")
			}
			break
		}
	}

	if !foundNested {
		t.Errorf("Nested command 'nested/deploy' not found")
	}
}

func TestDiscoverCommands_PriorityLocations(t *testing.T) {
	fixturePath := filepath.Join("..", "testdata", "repos", "dotdir-resources")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Should find commands in .claude/commands
	// Note: this fixture only has .claude/commands, not .opencode/commands
	if len(commands) < 1 {
		t.Errorf("Expected at least 1 command, got %d", len(commands))
	}

	// Verify we found the claude command
	foundClaude := false

	for _, cmd := range commands {
		if cmd.Name == "claude-cmd" {
			foundClaude = true
			if cmd.Description != "Claude-specific command" {
				t.Errorf("Wrong description for claude-cmd: %s", cmd.Description)
			}
		}
	}

	if !foundClaude {
		t.Error("Did not find command 'claude-cmd' from .claude/commands")
	}
}
