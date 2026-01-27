package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIBulkAdd tests the 'aimgr repo add bulk' command
func TestCLIBulkAdd(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create a folder with mixed resources
	resourcesDir := filepath.Join(testDir, "resources")

	// Create commands directory
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmd1 := []byte("---\ndescription: Bulk test command 1\n---\n# Command 1")
	if err := os.WriteFile(filepath.Join(commandsDir, "bulk-cmd1.md"), cmd1, 0644); err != nil {
		t.Fatalf("Failed to write command 1: %v", err)
	}

	cmd2 := []byte("---\ndescription: Bulk test command 2\n---\n# Command 2")
	if err := os.WriteFile(filepath.Join(commandsDir, "bulk-cmd2.md"), cmd2, 0644); err != nil {
		t.Fatalf("Failed to write command 2: %v", err)
	}

	// Create skills directory
	skillsDir := filepath.Join(resourcesDir, "skills")
	skill1Dir := filepath.Join(skillsDir, "bulk-skill1")
	if err := os.MkdirAll(skill1Dir, 0755); err != nil {
		t.Fatalf("Failed to create skill1 dir: %v", err)
	}

	skill1 := []byte("---\nname: bulk-skill1\ndescription: Bulk test skill 1\n---\n# Skill 1")
	if err := os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), skill1, 0644); err != nil {
		t.Fatalf("Failed to write skill 1: %v", err)
	}

	// Create agents directory
	agentsDir := filepath.Join(resourcesDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	agent1 := []byte("---\ndescription: Bulk test agent 1\n---\n# Agent 1")
	if err := os.WriteFile(filepath.Join(agentsDir, "bulk-agent1.md"), agent1, 0644); err != nil {
		t.Fatalf("Failed to write agent 1: %v", err)
	}

	// Test: aimgr repo add (unified command)
	output, err := runAimgr(t, "repo", "import", resourcesDir)
	if err != nil {
		t.Fatalf("Failed to add resources: %v\nOutput: %s", err, output)
	}

	// Verify output mentions resources
	if !strings.Contains(output, "Found: 2 commands, 1 skills, 1 agents") {
		t.Errorf("Output should show discovered resources, got: %s", output)
	}

	if !strings.Contains(output, "Summary: 4 added") {
		t.Errorf("Output should show 4 added resources, got: %s", output)
	}

	// Verify resources are in repository
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	expectedResources := []string{"bulk-cmd1", "bulk-cmd2", "bulk-skill1", "bulk-agent1"}
	for _, resource := range expectedResources {
		if !strings.Contains(listOutput, resource) {
			t.Errorf("List should contain '%s', got: %s", resource, listOutput)
		}
	}
}

// TestCLIBulkAddForce tests bulk add with force flag
func TestCLIBulkAddForce(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create and add a command first
	cmdPath := filepath.Join(testDir, "conflict.md")
	cmd := []byte("---\ndescription: Original command\n---\n# Original")
	if err := os.WriteFile(cmdPath, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	_, err := runAimgr(t, "repo", "import", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command first time: %v", err)
	}

	// Create resources dir with same command
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmdPath2 := filepath.Join(commandsDir, "conflict.md")
	cmd2 := []byte("---\ndescription: Updated command\n---\n# Updated")
	if err := os.WriteFile(cmdPath2, cmd2, 0644); err != nil {
		t.Fatalf("Failed to write command 2: %v", err)
	}

	// Try unified add with force
	output, err := runAimgr(t, "repo", "import", "--force", resourcesDir)
	if err != nil {
		t.Fatalf("Force add failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Should overwrite with force flag, got: %s", output)
	}
}

// TestCLIBulkAddSkipExisting tests bulk add with skip-existing flag
func TestCLIBulkAddSkipExisting(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create and add a command
	cmdPath := filepath.Join(testDir, "skip.md")
	cmd := []byte("---\ndescription: Existing command\n---\n# Existing")
	if err := os.WriteFile(cmdPath, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	_, err := runAimgr(t, "repo", "import", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Create resources dir with same command
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmdPath2 := filepath.Join(commandsDir, "skip.md")
	if err := os.WriteFile(cmdPath2, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command 2: %v", err)
	}

	// Try unified add with skip-existing
	output, err := runAimgr(t, "repo", "import", "--skip-existing", resourcesDir)
	if err != nil {
		t.Fatalf("Skip-existing add failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "1 skipped") {
		t.Errorf("Should skip existing resource, got: %s", output)
	}
	if !strings.Contains(output, "0 added") {
		t.Errorf("Should not add any resources, got: %s", output)
	}
}

// TestCLIBulkAddNoFlagsFailsOnConflict tests that bulk add without flags fails on conflict
func TestCLIBulkAddNoFlagsFailsOnConflict(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create and add a command
	cmdPath := filepath.Join(testDir, "fail.md")
	cmd := []byte("---\ndescription: Existing command\n---\n# Existing")
	if err := os.WriteFile(cmdPath, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	_, err := runAimgr(t, "repo", "import", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Create resources dir with same command
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmdPath2 := filepath.Join(commandsDir, "fail.md")
	if err := os.WriteFile(cmdPath2, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command 2: %v", err)
	}

	// Try unified add without flags (should fail)
	_, err = runAimgr(t, "repo", "import", resourcesDir)
	if err == nil {
		t.Error("Expected error on conflict without flags, got nil")
	}
}

// TestCLIBulkAddDryRun tests bulk add with dry-run flag
func TestCLIBulkAddDryRun(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create a resources directory with one command
	resourcesDir := filepath.Join(testDir, "resources")
	commandsDir := filepath.Join(resourcesDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "dryrun.md")
	cmd := []byte("---\ndescription: Dry run test\n---\n# Dry Run")
	if err := os.WriteFile(cmdPath, cmd, 0644); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	// Test unified add with dry-run
	output, err := runAimgr(t, "repo", "import", "--dry-run", resourcesDir)
	if err != nil {
		t.Fatalf("Dry run add failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("Output should mention dry run mode, got: %s", output)
	}

	if !strings.Contains(output, "Summary: 1 added") {
		t.Errorf("Output should show 1 added in preview, got: %s", output)
	}

	// Verify nothing was actually added
	listOutput, err := runAimgr(t, "repo", "list")
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if strings.Contains(listOutput, "dryrun") {
		t.Errorf("Dry run should not add resources, but found 'dryrun' in: %s", listOutput)
	}
}

// TestCLIBulkAddEmptyFolder tests unified add on an empty folder
func TestCLIBulkAddEmptyFolder(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create an empty resources directory
	resourcesDir := filepath.Join(testDir, "empty")
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	// Test unified add on empty folder (should fail)
	_, err := runAimgr(t, "repo", "import", resourcesDir)
	if err == nil {
		t.Error("Expected error on empty folder, got nil")
	}
}

// TestCLIAddSingleFile tests unified add on a single file (should succeed for .md files)
func TestCLIAddSingleFile(t *testing.T) {
	testDir := t.TempDir()
	xdgData := filepath.Join(testDir, "xdg-data")

	t.Setenv("XDG_DATA_HOME", xdgData)

	// Create a command file
	cmdPath := filepath.Join(testDir, "test-cmd.md")
	cmd := []byte("---\ndescription: Test command\n---\n# Test Command")
	if err := os.WriteFile(cmdPath, cmd, 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	// Test unified add on single file (should succeed)
	output, err := runAimgr(t, "repo", "import", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add single file: %v\nOutput: %s", err, output)
	}

	// Verify success
	if !strings.Contains(output, "Added command") {
		t.Errorf("Output should show added command, got: %s", output)
	}
}
