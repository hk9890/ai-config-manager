package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

// setupTestRepo creates a test repository with sample resources
func setupTestRepo(t *testing.T) (string, *repo.Manager) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create test resources
	testDir := t.TempDir()

	// Create command resources
	cmdFiles := map[string]string{
		"test-cmd.md": `---
description: Test command
---
# Test Command`,
		"pdf-extract.md": `---
description: Extract text from PDF
---
# PDF Extract`,
		"data-process.md": `---
description: Process data
---
# Data Process`,
	}

	for name, content := range cmdFiles {
		path := filepath.Join(testDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test command %s: %v", name, err)
		}
		if err := mgr.AddCommand(path, "file://"+path, "file"); err != nil {
			t.Fatalf("failed to add command %s: %v", name, err)
		}
	}

	// Create skill resources
	skillDirs := map[string]string{
		"test-skill": `---
description: Test skill
---
# Test Skill`,
		"pdf-processing": `---
description: Process PDF files
---
# PDF Processing`,
	}

	for name, content := range skillDirs {
		skillDir := filepath.Join(testDir, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("failed to create skill dir %s: %v", name, err)
		}
		skillMd := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillMd, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create skill file %s: %v", name, err)
		}
		if err := mgr.AddSkill(skillDir, "file://"+skillDir, "file"); err != nil {
			t.Fatalf("failed to add skill %s: %v", name, err)
		}
	}

	// Create agent resources
	agentFiles := map[string]string{
		"test-agent.md": `---
description: Test agent
---
# Test Agent`,
		"code-reviewer.md": `---
description: Reviews code quality
---
# Code Reviewer`,
	}

	for name, content := range agentFiles {
		path := filepath.Join(testDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test agent %s: %v", name, err)
		}
		if err := mgr.AddAgent(path, "file://"+path, "file"); err != nil {
			t.Fatalf("failed to add agent %s: %v", name, err)
		}
	}

	return repoDir, mgr
}

// filterResources applies pattern matching to resources (mimics list command logic)
func filterResources(t *testing.T, mgr *repo.Manager, patternStr string) []resource.Resource {
	t.Helper()

	// Parse pattern
	matcher, err := pattern.NewMatcher(patternStr)
	if err != nil {
		t.Fatalf("invalid pattern '%s': %v", patternStr, err)
	}

	// Get resource type filter if pattern specifies it
	resourceType, _, _ := pattern.ParsePattern(patternStr)
	var typeFilter *resource.ResourceType
	if resourceType != "" {
		typeFilter = &resourceType
	}

	// List resources with optional type filter
	resources, err := mgr.List(typeFilter)
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}

	// Apply pattern matching
	var filtered []resource.Resource
	for _, res := range resources {
		if matcher.Match(&res) {
			filtered = append(filtered, res)
		}
	}

	return filtered
}

func TestListCmd_NoPattern(t *testing.T) {
	_, mgr := setupTestRepo(t)

	// List all resources
	resources, err := mgr.List(nil)
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}

	// Should have all 7 resources
	if len(resources) != 7 {
		t.Errorf("expected 7 resources, got %d", len(resources))
	}

	// Check that all types are present
	var cmdCount, skillCount, agentCount int
	for _, res := range resources {
		if res.Type == resource.Command {
			cmdCount++
		} else if res.Type == resource.Skill {
			skillCount++
		} else if res.Type == resource.Agent {
			agentCount++
		}
	}

	if cmdCount != 3 {
		t.Errorf("expected 3 commands, got %d", cmdCount)
	}
	if skillCount != 2 {
		t.Errorf("expected 2 skills, got %d", skillCount)
	}
	if agentCount != 2 {
		t.Errorf("expected 2 agents, got %d", agentCount)
	}
}

func TestListCmd_TypeFilter(t *testing.T) {
	_, mgr := setupTestRepo(t)

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		expectedType  resource.ResourceType
	}{
		{
			name:          "filter skills only",
			pattern:       "skill/*",
			expectedCount: 2,
			expectedType:  resource.Skill,
		},
		{
			name:          "filter commands only",
			pattern:       "command/*",
			expectedCount: 3,
			expectedType:  resource.Command,
		},
		{
			name:          "filter agents only",
			pattern:       "agent/*",
			expectedCount: 2,
			expectedType:  resource.Agent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterResources(t, mgr, tt.pattern)

			if len(filtered) != tt.expectedCount {
				t.Errorf("expected %d resources, got %d", tt.expectedCount, len(filtered))
			}

			// All should be the expected type
			for _, res := range filtered {
				if res.Type != tt.expectedType {
					t.Errorf("expected all resources to be %s, got type: %s", tt.expectedType, res.Type)
				}
			}
		})
	}
}

func TestListCmd_NamePattern(t *testing.T) {
	_, mgr := setupTestRepo(t)

	tests := []struct {
		name          string
		pattern       string
		expectedNames []string
	}{
		{
			name:          "wildcard prefix",
			pattern:       "command/test*",
			expectedNames: []string{"test-cmd"},
		},
		{
			name:          "wildcard contains",
			pattern:       "*pdf*",
			expectedNames: []string{"pdf-extract", "pdf-processing"},
		},
		{
			name:          "wildcard suffix",
			pattern:       "agent/*-reviewer",
			expectedNames: []string{"code-reviewer"},
		},
		{
			name:          "exact match with type",
			pattern:       "skill/test-skill",
			expectedNames: []string{"test-skill"},
		},
		{
			name:          "wildcard matches multiple types",
			pattern:       "*test*",
			expectedNames: []string{"test-cmd", "test-skill", "test-agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterResources(t, mgr, tt.pattern)

			if len(filtered) != len(tt.expectedNames) {
				t.Errorf("expected %d resources, got %d", len(tt.expectedNames), len(filtered))
			}

			// Check each expected name is present
			foundNames := make(map[string]bool)
			for _, res := range filtered {
				foundNames[res.Name] = true
			}

			for _, expectedName := range tt.expectedNames {
				if !foundNames[expectedName] {
					t.Errorf("expected to find resource '%s', but it was not in results", expectedName)
				}
			}
		})
	}
}

func TestListCmd_NoMatches(t *testing.T) {
	_, mgr := setupTestRepo(t)

	filtered := filterResources(t, mgr, "command/nonexistent*")

	if len(filtered) != 0 {
		t.Errorf("expected 0 resources for non-matching pattern, got %d", len(filtered))
	}
}

func TestListCmd_EmptyRepository(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	resources, err := mgr.List(nil)
	if err != nil {
		t.Fatalf("failed to list resources: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources in empty repository, got %d", len(resources))
	}
}

func TestListCmd_InvalidPattern(t *testing.T) {
	_, _ = setupTestRepo(t)

	// Try to create matcher with invalid pattern
	_, err := pattern.NewMatcher("command/[invalid")
	if err == nil {
		t.Error("expected error for invalid pattern, got nil")
	}
}

func TestListCmd_TypeAndNameCombinations(t *testing.T) {
	_, mgr := setupTestRepo(t)

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		description   string
	}{
		{
			name:          "type with exact name",
			pattern:       "command/pdf-extract",
			expectedCount: 1,
			description:   "should match exact command name",
		},
		{
			name:          "type with prefix wildcard",
			pattern:       "skill/*-processing",
			expectedCount: 1,
			description:   "should match skill ending with -processing",
		},
		{
			name:          "type with suffix wildcard",
			pattern:       "command/data-*",
			expectedCount: 1,
			description:   "should match command starting with data-",
		},
		{
			name:          "type with contains wildcard",
			pattern:       "agent/*code*",
			expectedCount: 1,
			description:   "should match agent containing 'code'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterResources(t, mgr, tt.pattern)

			if len(filtered) != tt.expectedCount {
				t.Errorf("%s: expected %d resources, got %d", tt.description, tt.expectedCount, len(filtered))
			}
		})
	}
}

// TestGetInstalledTargets_Command tests target detection for command resources
func TestGetInstalledTargets_Command(t *testing.T) {
	// Create a temporary project directory
	projectDir := t.TempDir()

	// Create .claude/commands directory with a command
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create claude commands dir: %v", err)
	}

	// Create a test command file (symlink)
	repoFile := filepath.Join(t.TempDir(), "test-command.md")
	if err := os.WriteFile(repoFile, []byte("# Test Command"), 0644); err != nil {
		t.Fatalf("failed to create repo file: %v", err)
	}
	claudeCommandFile := filepath.Join(claudeCommandsDir, "test-command.md")
	if err := os.Symlink(repoFile, claudeCommandFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Create .opencode/commands directory with the same command
	opencodeCommandsDir := filepath.Join(projectDir, ".opencode", "commands")
	if err := os.MkdirAll(opencodeCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create opencode commands dir: %v", err)
	}
	opencodeCommandFile := filepath.Join(opencodeCommandsDir, "test-command.md")
	if err := os.Symlink(repoFile, opencodeCommandFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test: Command installed in claude and opencode
	targets := getInstalledTargets(projectDir, "test-command", resource.Command)
	if len(targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(targets))
	}

	// Verify both claude and opencode are in the list
	hasClaude := false
	hasOpencode := false
	for _, tool := range targets {
		if tool == tools.Claude {
			hasClaude = true
		}
		if tool == tools.OpenCode {
			hasOpencode = true
		}
	}
	if !hasClaude {
		t.Error("expected claude in targets")
	}
	if !hasOpencode {
		t.Error("expected opencode in targets")
	}

	// Copilot doesn't support commands, so it shouldn't be in the list
	for _, tool := range targets {
		if tool == tools.Copilot {
			t.Error("copilot should not be in targets (doesn't support commands)")
		}
	}
}

// TestGetInstalledTargets_Skill tests target detection for skill resources
func TestGetInstalledTargets_Skill(t *testing.T) {
	projectDir := t.TempDir()

	// Create skill directories
	repoSkillDir := filepath.Join(t.TempDir(), "test-skill")
	if err := os.MkdirAll(repoSkillDir, 0755); err != nil {
		t.Fatalf("failed to create repo skill dir: %v", err)
	}
	skillMd := filepath.Join(repoSkillDir, "SKILL.md")
	if err := os.WriteFile(skillMd, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("failed to create SKILL.md: %v", err)
	}

	// Install in all three tools (all support skills)
	for _, tool := range tools.AllTools() {
		toolInfo := tools.GetToolInfo(tool)
		skillsDir := filepath.Join(projectDir, toolInfo.SkillsDir)
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("failed to create skills dir for %s: %v", tool, err)
		}
		skillSymlink := filepath.Join(skillsDir, "test-skill")
		if err := os.Symlink(repoSkillDir, skillSymlink); err != nil {
			t.Fatalf("failed to create symlink for %s: %v", tool, err)
		}
	}

	// Test: Skill installed in all tools
	targets := getInstalledTargets(projectDir, "test-skill", resource.Skill)
	if len(targets) != 3 {
		t.Errorf("expected 3 targets (all tools support skills), got %d", len(targets))
	}

	// Verify all tools are present
	toolMap := make(map[tools.Tool]bool)
	for _, tool := range targets {
		toolMap[tool] = true
	}
	if !toolMap[tools.Claude] {
		t.Error("expected claude in targets")
	}
	if !toolMap[tools.OpenCode] {
		t.Error("expected opencode in targets")
	}
	if !toolMap[tools.Copilot] {
		t.Error("expected copilot in targets")
	}
}

// TestGetInstalledTargets_Agent tests target detection for agent resources
func TestGetInstalledTargets_Agent(t *testing.T) {
	projectDir := t.TempDir()

	// Create repo agent file
	repoAgentFile := filepath.Join(t.TempDir(), "test-agent.md")
	if err := os.WriteFile(repoAgentFile, []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("failed to create repo agent: %v", err)
	}

	// Install in claude only
	claudeAgentsDir := filepath.Join(projectDir, ".claude", "agents")
	if err := os.MkdirAll(claudeAgentsDir, 0755); err != nil {
		t.Fatalf("failed to create claude agents dir: %v", err)
	}
	claudeAgentFile := filepath.Join(claudeAgentsDir, "test-agent.md")
	if err := os.Symlink(repoAgentFile, claudeAgentFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test: Agent installed in claude only
	targets := getInstalledTargets(projectDir, "test-agent", resource.Agent)
	if len(targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(targets))
	}
	if len(targets) > 0 && targets[0] != tools.Claude {
		t.Errorf("expected claude, got %s", targets[0])
	}

	// Copilot doesn't support agents
	for _, tool := range targets {
		if tool == tools.Copilot {
			t.Error("copilot should not be in targets (doesn't support agents)")
		}
	}
}

// TestGetInstalledTargets_NotInstalled tests detection when resource is not installed
func TestGetInstalledTargets_NotInstalled(t *testing.T) {
	projectDir := t.TempDir()

	// Create empty tool directories
	for _, tool := range tools.AllTools() {
		toolInfo := tools.GetToolInfo(tool)
		if toolInfo.SupportsCommands {
			commandsDir := filepath.Join(projectDir, toolInfo.CommandsDir)
			if err := os.MkdirAll(commandsDir, 0755); err != nil {
				t.Fatalf("failed to create commands dir: %v", err)
			}
		}
	}

	// Test: Resource not installed anywhere
	targets := getInstalledTargets(projectDir, "nonexistent-command", resource.Command)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets for nonexistent resource, got %d", len(targets))
	}
}

// TestGetInstalledTargets_MissingDirectories tests graceful handling of missing directories
func TestGetInstalledTargets_MissingDirectories(t *testing.T) {
	projectDir := t.TempDir()

	// Don't create any tool directories
	// Test should handle missing directories gracefully (no error)

	targets := getInstalledTargets(projectDir, "test-command", resource.Command)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets when directories don't exist, got %d", len(targets))
	}
}

// TestGetInstalledTargets_BrokenSymlink tests handling of broken symlinks
func TestGetInstalledTargets_BrokenSymlink(t *testing.T) {
	projectDir := t.TempDir()

	// Create .claude/commands directory
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create claude commands dir: %v", err)
	}

	// Create a symlink pointing to a nonexistent file
	brokenSymlink := filepath.Join(claudeCommandsDir, "broken-command.md")
	if err := os.Symlink("/nonexistent/file.md", brokenSymlink); err != nil {
		t.Fatalf("failed to create broken symlink: %v", err)
	}

	// Test: Broken symlink should be ignored
	targets := getInstalledTargets(projectDir, "broken-command", resource.Command)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets for broken symlink, got %d", len(targets))
	}
}

// TestGetInstalledTargets_NestedCommand tests nested command detection
func TestGetInstalledTargets_NestedCommand(t *testing.T) {
	projectDir := t.TempDir()

	// Create repo file for nested command
	repoFile := filepath.Join(t.TempDir(), "deploy.md")
	if err := os.WriteFile(repoFile, []byte("# Deploy Command"), 0644); err != nil {
		t.Fatalf("failed to create repo file: %v", err)
	}

	// Create .claude/commands/api directory with nested command
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands", "api")
	if err := os.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create nested commands dir: %v", err)
	}
	claudeCommandFile := filepath.Join(claudeCommandsDir, "deploy.md")
	if err := os.Symlink(repoFile, claudeCommandFile); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Test: Nested command detection (name = "api/deploy")
	targets := getInstalledTargets(projectDir, "api/deploy", resource.Command)
	if len(targets) != 1 {
		t.Errorf("expected 1 target for nested command, got %d", len(targets))
	}
	if len(targets) > 0 && targets[0] != tools.Claude {
		t.Errorf("expected claude, got %s", targets[0])
	}
}

// TestGetInstalledTargets_RegularFile tests that regular files (non-symlinks) are detected
func TestGetInstalledTargets_RegularFile(t *testing.T) {
	projectDir := t.TempDir()

	// Create .claude/commands directory
	claudeCommandsDir := filepath.Join(projectDir, ".claude", "commands")
	if err := os.MkdirAll(claudeCommandsDir, 0755); err != nil {
		t.Fatalf("failed to create claude commands dir: %v", err)
	}

	// Create a regular file (not a symlink)
	regularFile := filepath.Join(claudeCommandsDir, "regular-command.md")
	if err := os.WriteFile(regularFile, []byte("# Regular Command"), 0644); err != nil {
		t.Fatalf("failed to create regular file: %v", err)
	}

	// Test: Regular file should also be detected (not just symlinks)
	targets := getInstalledTargets(projectDir, "regular-command", resource.Command)
	if len(targets) != 1 {
		t.Errorf("expected 1 target for regular file, got %d", len(targets))
	}
	if len(targets) > 0 && targets[0] != tools.Claude {
		t.Errorf("expected claude, got %s", targets[0])
	}
}

// TestGetInstalledTargets_AllTools tests that all tools are checked
func TestGetInstalledTargets_AllTools(t *testing.T) {
	// This test verifies that the function uses tools.AllTools() and doesn't hardcode tools

	projectDir := t.TempDir()

	// Create repo command file
	repoFile := filepath.Join(t.TempDir(), "test-command.md")
	if err := os.WriteFile(repoFile, []byte("# Test Command"), 0644); err != nil {
		t.Fatalf("failed to create repo file: %v", err)
	}

	// Install in all tools that support commands
	for _, tool := range tools.AllTools() {
		toolInfo := tools.GetToolInfo(tool)
		if !toolInfo.SupportsCommands {
			continue
		}
		commandsDir := filepath.Join(projectDir, toolInfo.CommandsDir)
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			t.Fatalf("failed to create commands dir for %s: %v", tool, err)
		}
		commandFile := filepath.Join(commandsDir, "test-command.md")
		if err := os.Symlink(repoFile, commandFile); err != nil {
			t.Fatalf("failed to create symlink for %s: %v", tool, err)
		}
	}

	// Test: All tools that support commands should be detected
	targets := getInstalledTargets(projectDir, "test-command", resource.Command)

	// Count how many tools support commands
	expectedCount := 0
	for _, tool := range tools.AllTools() {
		if tools.GetToolInfo(tool).SupportsCommands {
			expectedCount++
		}
	}

	if len(targets) != expectedCount {
		t.Errorf("expected %d targets (tools that support commands), got %d", expectedCount, len(targets))
	}
}
