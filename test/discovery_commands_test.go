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
	expectedNames := map[string]bool{
		"build":  true,
		"test":   true,
		"deploy": true,
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

	// Verify nested command is found
	foundNested := false
	for _, cmd := range commands {
		if cmd.Name == "deploy" {
			foundNested = true
			if cmd.Description == "" {
				t.Errorf("Nested command has empty description")
			}
			break
		}
	}

	if !foundNested {
		t.Errorf("Nested command 'deploy' not found")
	}
}

func TestDiscoverCommands_PriorityLocations(t *testing.T) {
	fixturePath := filepath.Join("..", "testdata", "repos", "dotdir-resources")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Should find command in .claude/commands/
	if len(commands) == 0 {
		t.Errorf("Expected to find commands in priority locations")
	}

	// Verify we found claude-cmd
	foundClaude := false
	for _, cmd := range commands {
		if cmd.Name == "claude-cmd" {
			foundClaude = true
			break
		}
	}

	if !foundClaude {
		t.Errorf("Expected to find claude-cmd in .claude/commands/")
	}
}

func TestDiscoverCommands_ExcludedFiles(t *testing.T) {
	// Verify README.md, SKILL.md are excluded
	fixturePath := filepath.Join("..", "testdata", "repos", "skills-standard")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Should NOT find SKILL.md or README.md as commands
	for _, cmd := range commands {
		if cmd.Name == "SKILL" || cmd.Name == "README" || cmd.Name == "readme" {
			t.Errorf("Excluded file found as command: %s", cmd.Name)
		}
	}
}

func TestDiscoverCommands_MixedResources(t *testing.T) {
	fixturePath := filepath.Join("..", "testdata", "repos", "mixed-resources")

	commands, err := discovery.DiscoverCommands(fixturePath, "")
	if err != nil {
		t.Fatalf("DiscoverCommands failed: %v", err)
	}

	// Should find only commands, not skills or agents
	if len(commands) == 0 {
		t.Errorf("Expected to find commands in mixed-resources")
	}

	// Verify we found cmd1
	foundCmd1 := false
	for _, cmd := range commands {
		if cmd.Name == "cmd1" {
			foundCmd1 = true
			break
		}
	}

	if !foundCmd1 {
		t.Errorf("Expected to find cmd1 in mixed-resources")
	}

	// Verify skills and agents are NOT returned as commands
	for _, cmd := range commands {
		// Check that agent files are not loaded as commands
		if cmd.Name == "agent1" {
			t.Errorf("Agent file incorrectly loaded as command")
		}
		// Check that SKILL.md is not loaded as command
		if cmd.Name == "SKILL" {
			t.Errorf("SKILL.md incorrectly loaded as command")
		}
	}
}
