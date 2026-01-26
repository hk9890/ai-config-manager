package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFilterSkills tests filtering to import only skills
func TestFilterSkills(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create resources with mixed types
	resourcesDir := filepath.Join(testDir, "resources")

	// Create commands
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	cmd1 := []byte("---\ndescription: Test command 1\n---\n# Command 1")
	os.WriteFile(filepath.Join(commandsDir, "test-cmd1.md"), cmd1, 0644)

	// Create skills
	skillsDir := filepath.Join(resourcesDir, "skills")
	skill1Dir := filepath.Join(skillsDir, "test-skill1")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skill1 := []byte("---\nname: test-skill1\ndescription: Test skill 1\n---\n# Skill 1")
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), skill1, 0644)

	skill2Dir := filepath.Join(skillsDir, "pdf-skill")
	if err := os.MkdirAll(skill2Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skill2 := []byte("---\nname: pdf-skill\ndescription: PDF skill\n---\n# PDF Skill")
	os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), skill2, 0644)

	// Create agents
	agentsDir := filepath.Join(resourcesDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}
	agent1 := []byte("---\ndescription: Test agent 1\n---\n# Agent 1")
	os.WriteFile(filepath.Join(agentsDir, "test-agent1.md"), agent1, 0644)

	// Test: Filter to import only skills
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "skill/*")
	if err != nil {
		t.Fatalf("Failed to add with filter: %v\nOutput: %s", err, output)
	}

	// Verify output shows filtering
	if !strings.Contains(output, "Filter: skill/*") {
		t.Errorf("Output should mention filter, got: %s", output)
	}

	// Should show 2 skills added
	if !strings.Contains(output, "Summary: 2 added") {
		t.Errorf("Should add 2 skills, got: %s", output)
	}

	// Verify only skills are in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if !strings.Contains(listOutput, "test-skill1") {
		t.Errorf("Should contain test-skill1, got: %s", listOutput)
	}
	if !strings.Contains(listOutput, "pdf-skill") {
		t.Errorf("Should contain pdf-skill, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "test-cmd1") {
		t.Errorf("Should NOT contain test-cmd1, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "test-agent1") {
		t.Errorf("Should NOT contain test-agent1, got: %s", listOutput)
	}
}

// TestFilterPattern tests pattern matching filters
func TestFilterPattern(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create resources with various names
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Commands with different patterns
	cmd1 := []byte("---\ndescription: PDF command\n---\n# PDF Command")
	os.WriteFile(filepath.Join(commandsDir, "pdf-parser.md"), cmd1, 0644)

	cmd2 := []byte("---\ndescription: PDF utility\n---\n# PDF Utility")
	os.WriteFile(filepath.Join(commandsDir, "pdf-utils.md"), cmd2, 0644)

	cmd3 := []byte("---\ndescription: Image command\n---\n# Image Command")
	os.WriteFile(filepath.Join(commandsDir, "image-tool.md"), cmd3, 0644)

	// Test: Filter with pattern matching
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "pdf*")
	if err != nil {
		t.Fatalf("Failed to add with filter: %v\nOutput: %s", err, output)
	}

	// Should show 2 commands added (pdf-parser and pdf-utils)
	if !strings.Contains(output, "Summary: 2 added") {
		t.Errorf("Should add 2 pdf commands, got: %s", output)
	}

	// Verify only pdf commands are in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if !strings.Contains(listOutput, "pdf-parser") {
		t.Errorf("Should contain pdf-parser, got: %s", listOutput)
	}
	if !strings.Contains(listOutput, "pdf-utils") {
		t.Errorf("Should contain pdf-utils, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "image-tool") {
		t.Errorf("Should NOT contain image-tool, got: %s", listOutput)
	}
}

// TestFilterExactName tests exact name matching
func TestFilterExactName(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create multiple resources
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmd1 := []byte("---\ndescription: Exact command\n---\n# Exact")
	os.WriteFile(filepath.Join(commandsDir, "exact-cmd.md"), cmd1, 0644)

	cmd2 := []byte("---\ndescription: Another command\n---\n# Another")
	os.WriteFile(filepath.Join(commandsDir, "other-cmd.md"), cmd2, 0644)

	// Test: Filter with exact name
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "exact-cmd")
	if err != nil {
		t.Fatalf("Failed to add with filter: %v\nOutput: %s", err, output)
	}

	// Should show 1 command added
	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Should add 1 exact command, got: %s", output)
	}

	// Verify only exact-cmd is in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if !strings.Contains(listOutput, "exact-cmd") {
		t.Errorf("Should contain exact-cmd, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "other-cmd") {
		t.Errorf("Should NOT contain other-cmd, got: %s", listOutput)
	}
}

// TestFilterNoMatch tests filter that matches no resources
func TestFilterNoMatch(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create resources
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmd1 := []byte("---\ndescription: Test command\n---\n# Test")
	os.WriteFile(filepath.Join(commandsDir, "test-cmd.md"), cmd1, 0644)

	// Test: Filter with no matches
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "nomatch*")
	// Should not error, just show warning
	if err != nil {
		t.Fatalf("Should not error on no matches: %v\nOutput: %s", err, output)
	}

	// Should show warning about zero matches
	if !strings.Contains(output, "Warning") && !strings.Contains(output, "matched 0") {
		t.Errorf("Should show warning about zero matches, got: %s", output)
	}

	// Should show no resources added
	if !strings.Contains(output, "Summary: 0 added") && !strings.Contains(output, "0 resources") {
		t.Errorf("Should show 0 resources added, got: %s", output)
	}
}

// TestFilterWithDryRun tests combining filter with dry-run flag
func TestFilterWithDryRun(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create resources
	resourcesDir := filepath.Join(testDir, "resources")
	skillsDir := filepath.Join(resourcesDir, "skills")
	skill1Dir := filepath.Join(skillsDir, "test-skill")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skill1 := []byte("---\nname: test-skill\ndescription: Test skill\n---\n# Skill")
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), skill1, 0644)

	// Test: Filter with dry-run
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "skill/*", "--dry-run")
	if err != nil {
		t.Fatalf("Failed dry-run with filter: %v\nOutput: %s", err, output)
	}

	// Should show dry-run mode
	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("Should mention dry-run mode, got: %s", output)
	}

	// Should show filter
	if !strings.Contains(output, "Filter: skill/*") {
		t.Errorf("Should mention filter, got: %s", output)
	}

	// Verify nothing was actually added
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if strings.Contains(listOutput, "test-skill") {
		t.Errorf("Dry-run should not add resources, got: %s", listOutput)
	}
}

// TestFilterWithForce tests combining filter with force flag
func TestFilterWithForce(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create and add a command first
	cmd1Path := filepath.Join(testDir, "update-me.md")
	cmd1 := []byte("---\ndescription: Original version\n---\n# Original")
	os.WriteFile(cmd1Path, cmd1, 0644)

	_, err := runAimgr(t, "repo", "add", cmd1Path)
	if err != nil {
		t.Fatalf("Failed to add initial command: %v", err)
	}

	// Create resources directory with updated version
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmd2 := []byte("---\ndescription: Updated version\n---\n# Updated")
	os.WriteFile(filepath.Join(commandsDir, "update-me.md"), cmd2, 0644)

	cmd3 := []byte("---\ndescription: Another command\n---\n# Another")
	os.WriteFile(filepath.Join(commandsDir, "other-cmd.md"), cmd3, 0644)

	// Test: Filter with force to update only matching resource
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "update*", "--force")
	if err != nil {
		t.Fatalf("Failed to add with filter and force: %v\nOutput: %s", err, output)
	}

	// Should show 1 command added
	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Should add 1 command, got: %s", output)
	}

	// Verify repository contains updated version
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if !strings.Contains(listOutput, "update-me") {
		t.Errorf("Should contain update-me, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "other-cmd") {
		t.Errorf("Should NOT contain other-cmd (filtered out), got: %s", listOutput)
	}
}

// TestFilterWithSkipExisting tests combining filter with skip-existing flag
func TestFilterWithSkipExisting(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create and add a command first
	cmd1Path := filepath.Join(testDir, "existing.md")
	cmd1 := []byte("---\ndescription: Existing command\n---\n# Existing")
	os.WriteFile(cmd1Path, cmd1, 0644)

	_, err := runAimgr(t, "repo", "add", cmd1Path)
	if err != nil {
		t.Fatalf("Failed to add initial command: %v", err)
	}

	// Create resources directory
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmd2 := []byte("---\ndescription: Existing command copy\n---\n# Existing")
	os.WriteFile(filepath.Join(commandsDir, "existing.md"), cmd2, 0644)

	cmd3 := []byte("---\ndescription: New command\n---\n# New")
	os.WriteFile(filepath.Join(commandsDir, "new-cmd.md"), cmd3, 0644)

	// Test: Filter with skip-existing
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "command/*", "--skip-existing")
	if err != nil {
		t.Fatalf("Failed to add with filter and skip-existing: %v\nOutput: %s", err, output)
	}

	// Should show 1 added, 1 skipped
	if !strings.Contains(output, "1 added") {
		t.Errorf("Should show 1 added, got: %s", output)
	}
	if !strings.Contains(output, "1 skipped") {
		t.Errorf("Should show 1 skipped, got: %s", output)
	}
}

// TestFilterWildcardAll tests using * to match all resources
func TestFilterWildcardAll(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create mixed resources
	resourcesDir := filepath.Join(testDir, "resources")

	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	cmd1 := []byte("---\ndescription: Command 1\n---\n# Command")
	os.WriteFile(filepath.Join(commandsDir, "cmd1.md"), cmd1, 0644)

	skillsDir := filepath.Join(resourcesDir, "skills")
	skill1Dir := filepath.Join(skillsDir, "skill1")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skill1 := []byte("---\nname: skill1\ndescription: Skill 1\n---\n# Skill")
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), skill1, 0644)

	// Test: Wildcard all
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "*")
	if err != nil {
		t.Fatalf("Failed to add with wildcard: %v\nOutput: %s", err, output)
	}

	// Should add all resources
	if !strings.Contains(output, "Summary: 2 added") {
		t.Errorf("Should add all 2 resources, got: %s", output)
	}
}

// TestFilterTypeWithExactName tests combining type filter with exact name
func TestFilterTypeWithExactName(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create resources with same name but different types
	resourcesDir := filepath.Join(testDir, "resources")

	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	cmd1 := []byte("---\ndescription: Test command\n---\n# Test Command")
	os.WriteFile(filepath.Join(commandsDir, "test-util.md"), cmd1, 0644)

	skillsDir := filepath.Join(resourcesDir, "skills")
	skill1Dir := filepath.Join(skillsDir, "test-util")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}
	skill1 := []byte("---\nname: test-util\ndescription: Test skill\n---\n# Test Skill")
	os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), skill1, 0644)

	// Test: Filter by type and exact name
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "skill/test-util")
	if err != nil {
		t.Fatalf("Failed to add with type filter: %v\nOutput: %s", err, output)
	}

	// Should add only the skill
	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Should add only 1 skill, got: %s", output)
	}

	// Verify only skill is in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	// Count occurrences - should be exactly 1 (the skill)
	count := strings.Count(listOutput, "test-util")
	if count != 1 {
		t.Errorf("Should have exactly 1 test-util (skill), got %d occurrences in: %s", count, listOutput)
	}

	// Verify it's the skill type
	if !strings.Contains(listOutput, "skill") && !strings.Contains(listOutput, "Skill") {
		t.Errorf("Should be a skill type, got: %s", listOutput)
	}
}

// TestFilterAgents tests filtering agent resources
func TestFilterAgents(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")
	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create mixed resources
	resourcesDir := filepath.Join(testDir, "resources")

	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}
	cmd1 := []byte("---\ndescription: Command\n---\n# Command")
	os.WriteFile(filepath.Join(commandsDir, "test-cmd.md"), cmd1, 0644)

	agentsDir := filepath.Join(resourcesDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}
	agent1 := []byte("---\ndescription: Code reviewer agent\n---\n# Reviewer")
	os.WriteFile(filepath.Join(agentsDir, "code-reviewer.md"), agent1, 0644)

	agent2 := []byte("---\ndescription: Test agent\n---\n# Tester")
	os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), agent2, 0644)

	// Test: Filter agents only
	output, err := runAimgr(t, "repo", "add", resourcesDir, "--filter", "agent/*")
	if err != nil {
		t.Fatalf("Failed to add agents: %v\nOutput: %s", err, output)
	}

	// Should add 2 agents
	if !strings.Contains(output, "Summary: 2 added") {
		t.Errorf("Should add 2 agents, got: %s", output)
	}

	// Verify only agents in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if !strings.Contains(listOutput, "code-reviewer") {
		t.Errorf("Should contain code-reviewer, got: %s", listOutput)
	}
	if !strings.Contains(listOutput, "test-agent") {
		t.Errorf("Should contain test-agent, got: %s", listOutput)
	}
	if strings.Contains(listOutput, "test-cmd") {
		t.Errorf("Should NOT contain test-cmd, got: %s", listOutput)
	}
}
