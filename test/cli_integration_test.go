package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to run aimgr CLI command
func runAimgr(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Build path to aimgr binary
	binPath := filepath.Join("..", "aimgr")

	cmd := exec.Command(binPath, args...)
	// Start with parent environment
	cmd.Env = os.Environ()

	// Override with test-specific AIMGR_REPO_PATH if set
	// NOTE: t.Setenv() only affects os.Getenv() in the test process,
	// not child processes, so we need to explicitly propagate it
	if repoPath := os.Getenv("AIMGR_REPO_PATH"); repoPath != "" {
		// Replace or add AIMGR_REPO_PATH in the environment
		found := false
		for i, env := range cmd.Env {
			if strings.HasPrefix(env, "AIMGR_REPO_PATH=") {
				cmd.Env[i] = "AIMGR_REPO_PATH=" + repoPath
				found = true
				break
			}
		}
		if !found {
			cmd.Env = append(cmd.Env, "AIMGR_REPO_PATH="+repoPath)
		}
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// TestCLIRepoAdd tests the 'aimgr repo add' command
func TestCLIRepoAdd(t *testing.T) {
	// Create temporary directories
	repoDir := t.TempDir()
	testDir := t.TempDir()

	// Set custom repo path
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command file
	cmdPath := filepath.Join(testDir, "test-cmd.md")
	cmdContent := `---
description: A test command
---
# Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Test: aimgr repo add (unified command)
	output, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "test-cmd") {
		t.Errorf("Output should mention command name, got: %s", output)
	}
}

// TestCLIRepoList tests the 'aimgr repo list' command
func TestCLIRepoList(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "list-test.md")
	cmdContent := `---
description: A command for list testing
---
# List Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test: aimgr repo list
	output, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "list-test") {
		t.Errorf("List output should contain 'list-test', got: %s", output)
	}
}

// TestCLIRepoShow tests the 'aimgr repo show' command
func TestCLIRepoShow(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command with metadata
	cmdPath := filepath.Join(testDir, "show-test.md")
	cmdContent := `---
description: A command for show testing
version: "1.0.0"
author: test-author
license: MIT
---
# Show Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test: aimgr repo show command/show-test
	output, err := runAimgr(t, "repo", "show", "command/show-test")
	if err != nil {
		t.Fatalf("Failed to show command: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected fields
	expectedStrings := []string{
		"show-test",
		"A command for show testing",
		"1.0.0",
		"test-author",
		"MIT",
		"Source:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Show output should contain '%s', got: %s", expected, output)
		}
	}
}

// TestCLIRepoShowSkill tests 'aimgr repo show skill'
func TestCLIRepoShowSkill(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test skill
	skillDir := filepath.Join(testDir, "show-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillContent := `---
name: show-skill
description: A skill for show testing
version: "2.0.0"
author: skill-author
license: Apache-2.0
---
# Show Skill
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add the skill
	addOutput, err := runAimgr(t, "repo", "import", "--force", skillDir)
	if err != nil {
		t.Fatalf("Failed to add skill: %v\nOutput: %s", err, addOutput)
	}

	// Test: aimgr repo show skill/show-skill
	output, err := runAimgr(t, "repo", "show", "skill/show-skill")
	if err != nil {
		t.Fatalf("Failed to show skill: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected fields
	expectedStrings := []string{
		"show-skill",
		"A skill for show testing",
		"2.0.0",
		"skill-author",
		"Apache-2.0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Show output should contain '%s', got: %s", expected, output)
		}
	}
}

// TestCLIRepoShowAgent tests 'aimgr repo show agent'
func TestCLIRepoShowAgent(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test agent
	agentPath := filepath.Join(testDir, "show-agent.md")
	agentContent := `---
description: An agent for show testing
type: helper
version: "1.5.0"
---
# Show Agent
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create test agent: %v", err)
	}

	// Add the agent
	_, err := runAimgr(t, "repo", "import", "--force", agentPath)
	if err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Test: aimgr repo show agent/show-agent
	output, err := runAimgr(t, "repo", "show", "agent/show-agent")
	if err != nil {
		t.Fatalf("Failed to show agent: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected fields
	expectedStrings := []string{
		"show-agent",
		"An agent for show testing",
		"1.5.0",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Show output should contain '%s', got: %s", expected, output)
		}
	}
}

// TestCLIRepoUpdate tests the 'aimgr repo update' command
func TestCLIRepoUpdate(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command
	cmdPath := filepath.Join(testDir, "update-test.md")
	cmdContent := `---
description: Original description
version: "1.0.0"
---
# Update Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Modify the source file
	updatedContent := `---
description: Updated description
version: "2.0.0"
---
# Update Test
`
	if err := os.WriteFile(cmdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update test command: %v", err)
	}

	// Test: aimgr repo update command/update-test
	output, err := runAimgr(t, "repo", "update", "command/update-test")
	if err != nil {
		t.Fatalf("Failed to update command: %v\nOutput: %s", err, output)
	}

	// Verify the resource was updated
	showOutput, err := runAimgr(t, "repo", "show", "command/update-test")
	if err != nil {
		t.Fatalf("Failed to show updated command: %v", err)
	}

	if !strings.Contains(showOutput, "Updated description") {
		t.Errorf("Command should have updated description, got: %s", showOutput)
	}
	if !strings.Contains(showOutput, "2.0.0") {
		t.Errorf("Command should have updated version, got: %s", showOutput)
	}
}

// TestCLIRepoUpdateAll tests 'aimgr repo update' (update all)
func TestCLIRepoUpdateAll(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command
	cmdPath := filepath.Join(testDir, "update-all-test.md")
	cmdContent := `---
description: Original
---
# Update All Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Modify the source file
	updatedContent := `---
description: Updated
---
# Update All Test
`
	if err := os.WriteFile(cmdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update test command: %v", err)
	}

	// Test: aimgr repo update (no args = update all)
	output, err := runAimgr(t, "repo", "update")
	if err != nil {
		t.Fatalf("Failed to update all: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "update-all-test") {
		t.Errorf("Update all output should mention the command, got: %s", output)
	}
}

// TestCLIInstall tests 'aimgr install' command
func TestCLIInstall(t *testing.T) {
	repoDir := t.TempDir()
	projectDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create .claude directory to trigger Claude detection
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create and add a test skill
	skillDir := filepath.Join(testDir, "install-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillContent := `---
name: install-skill
description: A skill for install testing
---
# Install Skill
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add to repository
	addOutput, err := runAimgr(t, "repo", "import", "--force", skillDir)
	if err != nil {
		t.Fatalf("Failed to add skill: %v\nOutput: %s", err, addOutput)
	}

	// Test: aimgr install skill/install-skill
	output, err := runAimgr(t, "install", "skill/install-skill", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("Failed to install skill: %v\nOutput: %s", err, output)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(projectDir, ".claude", "skills", "install-skill")
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Errorf("Symlink should be created at %s: %v", symlinkPath, err)
	}
}

// TestCLIInstallMultiple tests installing multiple resources at once
func TestCLIInstallMultiple(t *testing.T) {
	repoDir := t.TempDir()
	projectDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create .claude directory
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create test command
	cmdPath := filepath.Join(testDir, "multi-cmd.md")
	cmdContent := `---
description: Multi test command
---
# Multi Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Create test skill
	skillDir := filepath.Join(testDir, "multi-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillContent := `---
name: multi-skill
description: Multi test skill
---
# Multi Skill
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add both to repository
	addCmdOutput, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v\nOutput: %s", err, addCmdOutput)
	}

	addSkillOutput, err := runAimgr(t, "repo", "import", "--force", skillDir)
	if err != nil {
		t.Fatalf("Failed to add skill: %v\nOutput: %s", err, addSkillOutput)
	}

	// Test: install both resources
	output, err := runAimgr(t, "install", "command/multi-cmd", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("Failed to install command: %v\nOutput: %s", err, output)
	}

	output, err = runAimgr(t, "install", "skill/multi-skill", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("Failed to install skill: %v\nOutput: %s", err, output)
	}

	// Verify both symlinks were created
	cmdSymlink := filepath.Join(projectDir, ".claude", "commands", "multi-cmd.md")
	if _, err := os.Lstat(cmdSymlink); err != nil {
		t.Errorf("Command symlink should be created: %v", err)
	}

	skillSymlink := filepath.Join(projectDir, ".claude", "skills", "multi-skill")
	if _, err := os.Lstat(skillSymlink); err != nil {
		t.Errorf("Skill symlink should be created: %v", err)
	}
}

// TestCLIUninstall tests the 'aimgr uninstall' command
func TestCLIUninstall(t *testing.T) {
	repoDir := t.TempDir()
	projectDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create .claude directory
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create and add a test skill
	skillDir := filepath.Join(testDir, "uninstall-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillContent := `---
name: uninstall-skill
description: A skill for uninstall testing
---
# Uninstall Skill
`
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Add to repository
	addOutput, err := runAimgr(t, "repo", "import", "--force", skillDir)
	if err != nil {
		t.Fatalf("Failed to add skill: %v\nOutput: %s", err, addOutput)
	}

	// Install it
	installOutput, err := runAimgr(t, "install", "skill/uninstall-skill", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("Failed to install skill: %v\nOutput: %s", err, installOutput)
	}

	// Verify symlink exists
	symlinkPath := filepath.Join(projectDir, ".claude", "skills", "uninstall-skill")
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Fatalf("Symlink should exist before uninstall: %v", err)
	}

	// Test: aimgr uninstall skill/uninstall-skill
	output, err := runAimgr(t, "uninstall", "skill/uninstall-skill", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("Failed to uninstall: %v\nOutput: %s", err, output)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(symlinkPath); err == nil {
		t.Error("Symlink should be removed after uninstall")
	}
}

// TestCLIUninstallSkipsNonSymlinks tests uninstall safety features
func TestCLIUninstallSkipsNonSymlinks(t *testing.T) {
	projectDir := t.TempDir()

	// Create .claude directory with a regular file (not a symlink)
	claudeSkillsDir := filepath.Join(projectDir, ".claude", "skills")
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	// Create a regular file (not managed by aimgr)
	regularFile := filepath.Join(claudeSkillsDir, "regular-file")
	if err := os.MkdirAll(regularFile, 0755); err != nil {
		t.Fatalf("Failed to create regular directory: %v", err)
	}

	// Test: aimgr uninstall skill/regular-file should skip it
	output, err := runAimgr(t, "uninstall", "skill/regular-file", "--project-path", projectDir)

	// Should not error, but should skip
	if err != nil && !strings.Contains(output, "skipped") {
		t.Fatalf("Unexpected error: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(strings.ToLower(output), "skip") {
		t.Errorf("Output should indicate skipping, got: %s", output)
	}

	// Verify the file still exists
	if _, err := os.Stat(regularFile); err != nil {
		t.Error("Regular file should not be removed")
	}
}

// TestCLIUninstallSkipsExternalSymlinks tests uninstall only removes aimgr-managed symlinks
func TestCLIUninstallSkipsExternalSymlinks(t *testing.T) {
	projectDir := t.TempDir()
	externalDir := t.TempDir()

	// Create .claude directory
	claudeSkillsDir := filepath.Join(projectDir, ".claude", "skills")
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	// Create an external skill directory
	externalSkill := filepath.Join(externalDir, "external-skill")
	if err := os.MkdirAll(externalSkill, 0755); err != nil {
		t.Fatalf("Failed to create external skill: %v", err)
	}

	// Create a symlink pointing to external directory
	symlinkPath := filepath.Join(claudeSkillsDir, "external-skill")
	if err := os.Symlink(externalSkill, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Test: aimgr uninstall skill/external-skill should skip it (not from repo)
	output, err := runAimgr(t, "uninstall", "skill/external-skill", "--project-path", projectDir)

	// Should not error, but should skip
	if err != nil && !strings.Contains(output, "skipped") {
		t.Fatalf("Unexpected error: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(strings.ToLower(output), "skip") && !strings.Contains(strings.ToLower(output), "not managed") {
		t.Errorf("Output should indicate skipping external symlink, got: %s", output)
	}

	// Verify the symlink still exists
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Error("External symlink should not be removed")
	}
}

// TestCLIMetadataTracking tests that metadata is properly tracked
func TestCLIMetadataTracking(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command
	cmdPath := filepath.Join(testDir, "metadata-test.md")
	cmdContent := `---
description: Metadata test command
---
# Metadata Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Show the command to check metadata
	output, err := runAimgr(t, "repo", "show", "command/metadata-test")
	if err != nil {
		t.Fatalf("Failed to show command: %v\nOutput: %s", err, output)
	}

	// Verify metadata fields are present
	metadataFields := []string{
		"Source:",
		"Source Type:",
		"First Installed:",
		"Last Updated:",
	}

	for _, field := range metadataFields {
		if !strings.Contains(output, field) {
			t.Errorf("Metadata should contain '%s', got: %s", field, output)
		}
	}

	// Note: We cannot easily verify metadata file location in CLI tests
	// because the CLI uses XDG data home, not a custom repo path.
	// The metadata fields in the show output above verify that metadata is working.
}

// TestCLIMetadataUpdatedOnUpdate tests metadata timestamps are updated
func TestCLIMetadataUpdatedOnUpdate(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command
	cmdPath := filepath.Join(testDir, "meta-update-test.md")
	cmdContent := `---
description: Original
---
# Meta Update Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Get initial metadata
	output1, err := runAimgr(t, "repo", "show", "command/meta-update-test")
	if err != nil {
		t.Fatalf("Failed to show command: %v", err)
	}

	// Update the command
	updatedContent := `---
description: Updated
---
# Meta Update Test
`
	if err := os.WriteFile(cmdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to update test command: %v", err)
	}

	_, err = runAimgr(t, "repo", "update", "command/meta-update-test")
	if err != nil {
		t.Fatalf("Failed to update command: %v", err)
	}

	// Get updated metadata
	output2, err := runAimgr(t, "repo", "show", "command/meta-update-test")
	if err != nil {
		t.Fatalf("Failed to show updated command: %v", err)
	}

	// Verify Last Updated field changed
	if !strings.Contains(output2, "Last Updated:") {
		t.Error("Updated metadata should contain Last Updated field")
	}

	// Note: We can't easily verify the timestamp changed without parsing,
	// but we verify the field is present
	if output1 == output2 {
		t.Error("Metadata output should change after update")
	}
}

// TestCLIMetadataDeletedOnRemove tests metadata is deleted with resource
func TestCLIMetadataDeletedOnRemove(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create test command
	cmdPath := filepath.Join(testDir, "meta-remove-test.md")
	cmdContent := `---
description: Remove test
---
# Meta Remove Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "import", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Note: We cannot easily verify metadata file location in CLI tests
	// because the CLI uses XDG data home, not a custom repo path.
	// We'll verify that the show command works before removal.
	showOutput, err := runAimgr(t, "repo", "show", "command/meta-remove-test")
	if err != nil {
		t.Fatalf("Failed to show command before removal: %v", err)
	}
	if !strings.Contains(showOutput, "meta-remove-test") {
		t.Error("Command should exist before removal")
	}

	// Remove the command (use --force to skip confirmation)
	removeOutput, err := runAimgr(t, "repo", "remove", "command/meta-remove-test", "--force")
	if err != nil {
		t.Fatalf("Failed to remove command: %v\nOutput: %s", err, removeOutput)
	}

	// Verify the command is gone by trying to show it (should fail)
	_, err = runAimgr(t, "repo", "show", "command/meta-remove-test")
	if err == nil {
		t.Error("Command should not exist after removal")
	}
}
