package test

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCLIRepoVerifyBasic tests basic verify command execution
func TestCLIRepoVerifyBasic(t *testing.T) {
	// Just test that verify runs without crashing
	output, _ := runAimgr(t, "repo", "verify")

	// Should have the header
	if !strings.Contains(output, "Repository Verification") {
		t.Errorf("Expected 'Repository Verification' header, got: %s", output)
	}

	// Should complete successfully (either "No issues found" or some status)
	if !strings.Contains(output, "Repository Verification") {
		t.Errorf("Expected verification output, got: %s", output)
	}
}

// TestCLIRepoVerifyResourceWithoutMetadata tests detection of resources without metadata
func TestCLIRepoVerifyResourceWithoutMetadata(t *testing.T) {
	repoDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create command directly in repo (bypassing Add to skip metadata creation)
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "no-metadata.md")
	cmdContent := `---
description: Command without metadata
---
# No Metadata Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Test: aimgr repo verify (should detect missing metadata)
	output, err := runAimgr(t, "repo", "verify")
	if err != nil {
		t.Fatalf("Verify failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Resources without metadata") {
		t.Errorf("Expected 'Resources without metadata', got: %s", output)
	}

	if !strings.Contains(output, "no-metadata") {
		t.Errorf("Expected 'no-metadata' to be listed, got: %s", output)
	}

	if !strings.Contains(output, "Warnings only") {
		t.Errorf("Expected 'Warnings only' status, got: %s", output)
	}
}

// TestCLIRepoVerifyOrphanedMetadata tests detection of orphaned metadata
func TestCLIRepoVerifyOrphanedMetadata(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "orphan-test.md")
	cmdContent := `---
description: Command that will be orphaned
---
# Orphan Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command (creates metadata)
	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the command file directly (leaving orphaned metadata)
	resourcePath := filepath.Join(repoDir, "commands", "orphan-test.md")
	if err := os.Remove(resourcePath); err != nil {
		t.Fatalf("Failed to remove command: %v", err)
	}

	// Test: aimgr repo verify (should detect orphaned metadata)
	output, err := runAimgr(t, "repo", "verify")
	if err == nil {
		t.Errorf("Expected non-zero exit code for errors")
	}

	if !strings.Contains(output, "Orphaned metadata") {
		t.Errorf("Expected 'Orphaned metadata', got: %s", output)
	}

	if !strings.Contains(output, "orphan-test") {
		t.Errorf("Expected 'orphan-test' to be listed, got: %s", output)
	}

	if !strings.Contains(output, "ERRORS found") {
		t.Errorf("Expected 'ERRORS found' status, got: %s", output)
	}
}

// TestCLIRepoVerifyMissingSourcePath tests detection of missing source paths
func TestCLIRepoVerifyMissingSourcePath(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "source-test.md")
	cmdContent := `---
description: Command with missing source
---
# Source Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the original source file (metadata still points to it)
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to remove source file: %v", err)
	}

	// Test: aimgr repo verify (should detect missing source path)
	output, err := runAimgr(t, "repo", "verify")
	if err != nil {
		t.Fatalf("Verify failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Metadata with missing source paths") {
		t.Errorf("Expected 'Metadata with missing source paths', got: %s", output)
	}

	if !strings.Contains(output, "source-test") {
		t.Errorf("Expected 'source-test' to be listed, got: %s", output)
	}

	if !strings.Contains(output, "Warnings only") {
		t.Errorf("Expected 'Warnings only' status (exit 0), got: %s", output)
	}
}

// TestCLIRepoVerifyTypeMismatch tests detection of type mismatches
func TestCLIRepoVerifyTypeMismatch(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a test command
	cmdPath := filepath.Join(testDir, "type-test.md")
	cmdContent := `---
description: Command for type mismatch test
---
# Type Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Add the command
	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Manually modify metadata to have wrong type
	meta, err := metadata.Load("type-test", resource.Command, repoDir)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Change type to agent (wrong type)
	meta.Type = resource.Agent
	if err := metadata.Save(meta, repoDir); err != nil {
		t.Fatalf("Failed to save modified metadata: %v", err)
	}

	// Test: aimgr repo verify (should detect type mismatch)
	output, err := runAimgr(t, "repo", "verify")
	if err == nil {
		t.Errorf("Expected non-zero exit code for errors")
	}

	if !strings.Contains(output, "Type mismatches") {
		t.Errorf("Expected 'Type mismatches', got: %s", output)
	}

	if !strings.Contains(output, "type-test") {
		t.Errorf("Expected 'type-test' to be listed, got: %s", output)
	}

	if !strings.Contains(output, "ERRORS found") {
		t.Errorf("Expected 'ERRORS found' status, got: %s", output)
	}
}

// TestCLIRepoVerifyFix tests automatic fixing of issues
func TestCLIRepoVerifyFix(t *testing.T) {
	repoDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create command directly in repo (no metadata)
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "fix-test.md")
	cmdContent := `---
description: Command to test fix
---
# Fix Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Also create orphaned metadata
	metaDir := filepath.Join(repoDir, ".metadata", "commands")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatalf("Failed to create metadata directory: %v", err)
	}

	orphanedMeta := &metadata.ResourceMetadata{
		Name:       "orphaned",
		Type:       resource.Command,
		SourceType: "local",
		SourceURL:  "file:///nonexistent/path.md",
	}
	if err := metadata.Save(orphanedMeta, repoDir); err != nil {
		t.Fatalf("Failed to create orphaned metadata: %v", err)
	}

	// Test: aimgr repo verify --fix
	output, err := runAimgr(t, "repo", "verify", "--fix")
	if err != nil {
		t.Fatalf("Verify --fix failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Created metadata for fix-test") {
		t.Errorf("Expected 'Created metadata for fix-test', got: %s", output)
	}

	if !strings.Contains(output, "Removed orphaned metadata for orphaned") {
		t.Errorf("Expected 'Removed orphaned metadata for orphaned', got: %s", output)
	}

	// Verify fix worked - run verify again (should be clean)
	output2, err := runAimgr(t, "repo", "verify")
	if err != nil {
		t.Fatalf("Verify after fix failed: %v\nOutput: %s", err, output2)
	}

	if !strings.Contains(output2, "No issues found") {
		t.Errorf("Expected 'No issues found' after fix, got: %s", output2)
	}

	// Verify metadata was created
	metaPath := filepath.Join(repoDir, ".metadata", "commands", "fix-test-metadata.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Errorf("Expected metadata file to be created at %s", metaPath)
	}

	// Verify orphaned metadata was removed
	orphanedMetaPath := filepath.Join(repoDir, ".metadata", "commands", "orphaned-metadata.json")
	if _, err := os.Stat(orphanedMetaPath); err == nil {
		t.Errorf("Expected orphaned metadata to be removed at %s", orphanedMetaPath)
	}
}

// TestCLIRepoVerifyJSON tests JSON output
func TestCLIRepoVerifyJSON(t *testing.T) {
	repoDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create command without metadata
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "json-test.md")
	cmdContent := `---
description: JSON test command
---
# JSON Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Test: aimgr repo verify --json
	output, err := runAimgr(t, "repo", "verify", "--json")
	if err != nil {
		t.Fatalf("Verify --json failed: %v\nOutput: %s", err, output)
	}

	// Parse JSON output
	var result struct {
		ResourcesWithoutMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"resources_without_metadata"`
		HasErrors   bool `json:"has_errors"`
		HasWarnings bool `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify JSON structure
	if len(result.ResourcesWithoutMetadata) != 1 {
		t.Errorf("Expected 1 resource without metadata, got %d", len(result.ResourcesWithoutMetadata))
	}

	if result.ResourcesWithoutMetadata[0].Name != "json-test" {
		t.Errorf("Expected name 'json-test', got %s", result.ResourcesWithoutMetadata[0].Name)
	}

	if result.HasErrors {
		t.Errorf("Expected HasErrors=false for warnings only")
	}

	if !result.HasWarnings {
		t.Errorf("Expected HasWarnings=true")
	}
}

// TestCLIRepoVerifyEmpty tests verify on empty repository
func TestCLIRepoVerifyEmpty(t *testing.T) {
	repoDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Test: aimgr repo verify on non-initialized repo
	output, err := runAimgr(t, "repo", "verify")
	if err != nil {
		t.Fatalf("Verify on empty repo failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(output, "Repository not initialized") {
		t.Errorf("Expected 'Repository not initialized', got: %s", output)
	}
}
