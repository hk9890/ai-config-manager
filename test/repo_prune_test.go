package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// TestCLIRepoPruneBasic tests basic prune command execution
func TestCLIRepoPruneBasic(t *testing.T) {
	// Create isolated directories for this test
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "temp-cmd.md")
	cmdContent := `---
description: A temporary command
---
# Temp Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	addOutput, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v\nOutput: %s", err, addOutput)
	}

	// Delete the source file to create orphaned metadata
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to delete source file: %v", err)
	}

	// Test: aimgr repo prune --dry-run
	output, err := runAimgr(t, "repo", "prune", "--dry-run")
	if err != nil {
		t.Fatalf("Failed to run prune dry-run: %v\nOutput: %s", err, output)
	}

	// Should mention at least one orphaned entry
	if !strings.Contains(output, "Found") && !strings.Contains(output, "orphaned") {
		t.Errorf("Expected orphaned metadata message in output, got: %s", output)
	}

	// Should mention temp-cmd specifically
	if !strings.Contains(output, "temp-cmd") {
		t.Errorf("Expected 'temp-cmd' in output, got: %s", output)
	}

	// Should mention dry run
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("Expected '[DRY RUN]' in output, got: %s", output)
	}

	// Run prune again with --force to verify it actually removes
	output2, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune --force: %v\nOutput: %s", err, output2)
	}

	// Should show removal happened
	if !strings.Contains(output2, "Removed") {
		t.Errorf("Expected 'Removed' in output, got: %s", output2)
	}
}

// TestCLIRepoPruneForce tests force mode (no confirmation)
func TestCLIRepoPruneForce(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "force-cmd.md")
	cmdContent := `---
description: A command to test force mode
---
# Force Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the source file to create orphaned metadata
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to delete source file: %v", err)
	}

	// Test: aimgr repo prune --force (no stdin interaction)
	output, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune with force: %v\nOutput: %s", err, output)
	}

	// Should mention removal
	if !strings.Contains(output, "Removed") || !strings.Contains(output, "force-cmd") {
		t.Errorf("Expected removal message for 'force-cmd', got: %s", output)
	}

	// Verify metadata is actually removed
	manager := repo.NewManagerWithPath(repoDir)
	_, err = manager.GetMetadata("force-cmd", resource.Command)
	if err == nil {
		t.Errorf("Metadata should be removed after prune --force")
	}
}

// TestCLIRepoPruneNoOrphaned tests behavior when no orphaned metadata exists
func TestCLIRepoPruneNoOrphaned(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command (and keep the source file)
	cmdPath := filepath.Join(testDir, "valid-cmd.md")
	cmdContent := `---
description: A valid command
---
# Valid Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test: aimgr repo prune (no orphaned entries)
	output, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune: %v\nOutput: %s", err, output)
	}

	// Should indicate no orphaned metadata
	if !strings.Contains(output, "No orphaned metadata found") {
		t.Errorf("Expected 'No orphaned metadata found', got: %s", output)
	}
}

// TestCLIRepoPruneMultipleOrphaned tests pruning multiple orphaned entries
func TestCLIRepoPruneMultipleOrphaned(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add multiple commands
	for i := 1; i <= 3; i++ {
		cmdPath := filepath.Join(testDir, "multi-cmd-"+string(rune('0'+i))+".md")
		cmdContent := `---
description: Multi command test
---
# Multi Command
`
		if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
			t.Fatalf("Failed to create test command %d: %v", i, err)
		}

		_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
		if err != nil {
			t.Fatalf("Failed to add command %d: %v", i, err)
		}

		// Delete source file
		if err := os.Remove(cmdPath); err != nil {
			t.Fatalf("Failed to delete source file %d: %v", i, err)
		}
	}

	// Test: aimgr repo prune --dry-run
	output, err := runAimgr(t, "repo", "prune", "--dry-run")
	if err != nil {
		t.Fatalf("Failed to run prune dry-run: %v\nOutput: %s", err, output)
	}

	// Should mention at least 3 entries (may find orphans from other tests in shared env)
	if !strings.Contains(output, "multi-cmd-1") || !strings.Contains(output, "multi-cmd-2") || !strings.Contains(output, "multi-cmd-3") {
		t.Errorf("Expected all three multi-cmd entries in output, got: %s", output)
	}

	if !strings.Contains(output, "[DRY RUN]") || !strings.Contains(output, "Would remove") {
		t.Errorf("Expected dry-run removal message, got: %s", output)
	}

	// Actually prune with --force
	output, err = runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune --force: %v\nOutput: %s", err, output)
	}

	// Should show removal confirmation
	if !strings.Contains(output, "Removed") && !strings.Contains(output, "orphaned") {
		t.Errorf("Expected removal confirmation in output, got: %s", output)
	}

	// Verify the three test commands were removed
	manager := repo.NewManagerWithPath(repoDir)
	for i := 1; i <= 3; i++ {
		cmdName := "multi-cmd-" + string(rune('0'+i))
		_, err := manager.GetMetadata(cmdName, resource.Command)
		if err == nil {
			t.Errorf("Metadata for %s should be removed after prune", cmdName)
		}
	}
}

// TestCLIRepoPruneSkillOrphaned tests pruning orphaned skill metadata
func TestCLIRepoPruneSkillOrphaned(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test skill
	skillDir := filepath.Join(testDir, "prune-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: prune-skill
description: A skill to test pruning
---
# Prune Skill
`
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add the skill
	_, err := runAimgr(t, "repo", "add", "--force", skillDir)
	if err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}

	// Delete the skill directory to create orphaned metadata
	if err := os.RemoveAll(skillDir); err != nil {
		t.Fatalf("Failed to delete skill directory: %v", err)
	}

	// Test: aimgr repo prune --dry-run
	output, err := runAimgr(t, "repo", "prune", "--dry-run")
	if err != nil {
		t.Fatalf("Failed to run prune dry-run: %v\nOutput: %s", err, output)
	}

	// Should mention the orphaned skill
	if !strings.Contains(output, "prune-skill") {
		t.Errorf("Expected 'prune-skill' in output, got: %s", output)
	}

	if !strings.Contains(output, "skill") {
		t.Errorf("Expected type 'skill' in output, got: %s", output)
	}
}

// TestCLIRepoPruneAgentOrphaned tests pruning orphaned agent metadata
func TestCLIRepoPruneAgentOrphaned(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test agent
	agentPath := filepath.Join(testDir, "prune-agent.md")
	agentContent := `---
description: An agent to test pruning
---
# Prune Agent
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add the agent
	_, err := runAimgr(t, "repo", "add", "--force", agentPath)
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Delete the agent file to create orphaned metadata
	if err := os.Remove(agentPath); err != nil {
		t.Fatalf("Failed to delete agent file: %v", err)
	}

	// Test: aimgr repo prune --force
	output, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune: %v\nOutput: %s", err, output)
	}

	// Should mention the orphaned agent
	if !strings.Contains(output, "prune-agent") {
		t.Errorf("Expected 'prune-agent' in output, got: %s", output)
	}

	if !strings.Contains(output, "agent") {
		t.Errorf("Expected type 'agent' in output, got: %s", output)
	}

	// Verify metadata is removed
	manager := repo.NewManagerWithPath(repoDir)
	_, err = manager.GetMetadata("prune-agent", resource.Agent)
	if err == nil {
		t.Errorf("Metadata should be removed after prune")
	}
}

// TestCLIRepoPruneGitSource tests that git sources are NOT pruned
func TestCLIRepoPruneGitSource(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create a test command
	cmdPath := filepath.Join(testDir, "git-cmd.md")
	cmdContent := `---
description: A command from git
---
# Git Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add to repo first (to get it into the repo)
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.AddCommand(cmdPath, "https://github.com/test/repo", "git"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the source file
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to delete source file: %v", err)
	}

	// Manually update metadata to be a git source
	meta, err := manager.GetMetadata("git-cmd", resource.Command)
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	meta.SourceType = "git"
	meta.SourceURL = "https://github.com/test/repo"
	if err := metadata.Save(meta, repoDir); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Test: aimgr repo prune (should not find orphaned git sources)
	output, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune: %v\nOutput: %s", err, output)
	}

	// Should indicate no orphaned metadata (git sources are excluded)
	if !strings.Contains(output, "No orphaned metadata found") {
		t.Errorf("Expected 'No orphaned metadata found' (git sources should be skipped), got: %s", output)
	}

	// Verify metadata still exists
	_, err = manager.GetMetadata("git-cmd", resource.Command)
	if err != nil {
		t.Errorf("Git source metadata should NOT be pruned, got error: %v", err)
	}
}

// TestCLIRepoPruneHelp tests the help output
func TestCLIRepoPruneHelp(t *testing.T) {
	output, err := runAimgr(t, "repo", "prune", "--help")
	if err != nil {
		t.Fatalf("Failed to get help: %v", err)
	}

	// Check for key help sections
	expectedContent := []string{
		"Remove metadata entries", // Part of the description
		"--force",
		"--dry-run",
		"source paths no longer exist",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing '%s'\nOutput: %s", expected, output)
		}
	}
}

// TestCLIRepoPruneEmptyRepo tests prune on empty repository
func TestCLIRepoPruneEmptyRepo(t *testing.T) {
	repoDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Test: aimgr repo prune on empty repo
	output, err := runAimgr(t, "repo", "prune", "--force")
	if err != nil {
		t.Fatalf("Failed to run prune on empty repo: %v\nOutput: %s", err, output)
	}

	// Should indicate no orphaned metadata
	if !strings.Contains(output, "No orphaned metadata found") {
		t.Errorf("Expected 'No orphaned metadata found', got: %s", output)
	}
}

// TestCLIRepoPruneMixedSources tests pruning with mixed source types
func TestCLIRepoPruneMixedSources(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create commands with different source types
	// 1. Local source (file still exists) - should NOT be pruned
	validPath := filepath.Join(testDir, "valid-cmd.md")
	if err := os.WriteFile(validPath, []byte(`---
description: Valid command
---
# Valid
`), 0644); err != nil {
		t.Fatalf("Failed to create valid command: %v", err)
	}
	_, err := runAimgr(t, "repo", "add", "--force", validPath)
	if err != nil {
		t.Fatalf("Failed to add valid command: %v", err)
	}

	// 2. Local source (file deleted) - SHOULD be pruned
	orphanPath := filepath.Join(testDir, "orphan-cmd.md")
	if err := os.WriteFile(orphanPath, []byte(`---
description: Orphan command
---
# Orphan
`), 0644); err != nil {
		t.Fatalf("Failed to create orphan command: %v", err)
	}
	_, err = runAimgr(t, "repo", "add", "--force", orphanPath)
	if err != nil {
		t.Fatalf("Failed to add orphan command: %v", err)
	}
	if err := os.Remove(orphanPath); err != nil {
		t.Fatalf("Failed to delete orphan file: %v", err)
	}

	// Test: aimgr repo prune --dry-run
	output, err := runAimgr(t, "repo", "prune", "--dry-run")
	if err != nil {
		t.Fatalf("Failed to run prune dry-run: %v\nOutput: %s", err, output)
	}
	t.Logf("Prune dry-run output:\n%s", output)

	// Should only find the orphaned one
	if !strings.Contains(output, "orphan-cmd") {
		t.Errorf("Expected 'orphan-cmd' in output, got: %s", output)
	}

	if strings.Contains(output, "valid-cmd") {
		t.Errorf("Should NOT mention 'valid-cmd' (file still exists), got: %s", output)
	}

	if !strings.Contains(output, "Found 1") {
		t.Errorf("Expected 'Found 1' orphaned entry, got: %s", output)
	}
}
