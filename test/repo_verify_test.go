package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// TestCLIRepoVerifyBasic tests basic verify command execution
func TestCLIRepoVerifyBasic(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Run verify on empty repo
	output, _ := runAimgr(t, "repo", "verify")

	// Should have the header
	if !strings.Contains(output, "Repository Verification") {
		t.Errorf("Expected 'Repository Verification' header, got: %s", output)
	}

	// Should have some status output
	validStatuses := []string{
		"No issues found",
		"ERRORS found",
		"Warnings only",
	}

	hasStatus := false
	for _, status := range validStatuses {
		if strings.Contains(output, status) {
			hasStatus = true
			break
		}
	}

	if !hasStatus {
		t.Errorf("Expected a status message in output, got: %s", output)
	}
}

// TestCLIRepoVerifyJSON tests JSON output format
func TestCLIRepoVerifyJSON(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	output, _ := runAimgr(t, "repo", "verify", "--json")

	// Parse JSON output
	var result struct {
		ResourcesWithoutMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"resources_without_metadata,omitempty"`
		OrphanedMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"orphaned_metadata,omitempty"`
		MissingSourcePaths []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Path       string `json:"path"`
			SourcePath string `json:"source_path"`
		} `json:"missing_source_paths,omitempty"`
		TypeMismatches []struct {
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
			MetadataType string `json:"metadata_type"`
		} `json:"type_mismatches,omitempty"`
		HasErrors   bool `json:"has_errors"`
		HasWarnings bool `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Just verify we got valid JSON with the right structure
	t.Logf("Verify JSON output parsed successfully")
	t.Logf("  Resources without metadata: %d", len(result.ResourcesWithoutMetadata))
	t.Logf("  Orphaned metadata: %d", len(result.OrphanedMetadata))
	t.Logf("  Missing source paths: %d", len(result.MissingSourcePaths))
	t.Logf("  Type mismatches: %d", len(result.TypeMismatches))
	t.Logf("  Has errors: %v", result.HasErrors)
	t.Logf("  Has warnings: %v", result.HasWarnings)
}

// TestCLIRepoVerifyHelp tests the help output
func TestCLIRepoVerifyHelp(t *testing.T) {
	output, err := runAimgr(t, "repo", "verify", "--help")
	if err != nil {
		t.Fatalf("Failed to get help: %v", err)
	}

	// Check for key help sections
	expectedContent := []string{
		"Check for consistency issues",
		"--fix",
		"--json",
		"Resources without metadata",
		"Orphaned metadata",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing '%s'\nOutput: %s", expected, output)
		}
	}
}

// TestCLIRepoVerifyResourcesWithoutMetadata tests detection of resources without metadata
func TestCLIRepoVerifyResourcesWithoutMetadata(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Manually create a resource without metadata
	manager := repo.NewManagerWithPath(repoDir)
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "no-meta-cmd.md")
	cmdContent := `---
description: Command without metadata
---
# No Metadata Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	// Test: verify should detect resource without metadata
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "Resources without metadata") {
		t.Errorf("Expected 'Resources without metadata' warning, got: %s", output)
	}

	if !strings.Contains(output, "no-meta-cmd") {
		t.Errorf("Expected 'no-meta-cmd' in output, got: %s", output)
	}

	// Verify it's a warning, not an error (exit code 0)
	if strings.Contains(output, "ERRORS found") {
		t.Errorf("Resources without metadata should be a warning, not an error")
	}

	// Verify metadata doesn't exist yet
	_, err := manager.GetMetadata("no-meta-cmd", resource.Command)
	if err == nil {
		t.Errorf("Metadata should not exist before --fix")
	}
}

// TestCLIRepoVerifyOrphanedMetadata tests detection of orphaned metadata
func TestCLIRepoVerifyOrphanedMetadata(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a resource
	cmdPath := filepath.Join(testDir, "orphan-test.md")
	cmdContent := `---
description: Will become orphaned
---
# Orphan Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the resource file (but leave metadata)
	manager := repo.NewManagerWithPath(repoDir)
	resourcePath := manager.GetPath("orphan-test", resource.Command)
	if err := os.Remove(resourcePath); err != nil {
		t.Fatalf("Failed to delete resource: %v", err)
	}

	// Test: verify should detect orphaned metadata
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "Orphaned metadata") {
		t.Errorf("Expected 'Orphaned metadata' error, got: %s", output)
	}

	if !strings.Contains(output, "orphan-test") {
		t.Errorf("Expected 'orphan-test' in output, got: %s", output)
	}

	// Should be an error (exit code 1)
	if !strings.Contains(output, "ERRORS found") {
		t.Errorf("Orphaned metadata should be an error")
	}
}

// TestCLIRepoVerifyMissingSourcePaths tests detection of missing source paths
func TestCLIRepoVerifyMissingSourcePaths(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a resource with local source
	cmdPath := filepath.Join(testDir, "source-test.md")
	cmdContent := `---
description: Test missing source path
---
# Source Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the source file (metadata and repo resource remain)
	if err := os.Remove(cmdPath); err != nil {
		t.Fatalf("Failed to delete source file: %v", err)
	}

	// Test: verify should detect missing source path
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "missing source paths") {
		t.Errorf("Expected 'missing source paths' warning, got: %s", output)
	}

	if !strings.Contains(output, "source-test") {
		t.Errorf("Expected 'source-test' in output, got: %s", output)
	}

	// Should be a warning, not an error
	if strings.Contains(output, "ERRORS found") {
		t.Errorf("Missing source paths should be a warning, not an error")
	}
}

// TestCLIRepoVerifyTypeMismatches tests detection of type mismatches
func TestCLIRepoVerifyTypeMismatches(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Manually create mismatched resource and metadata
	// Create a command file
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "type-mismatch.md")
	cmdContent := `---
description: Type mismatch test
---
# Mismatch Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Create metadata with wrong type (agent instead of command)
	meta := &metadata.ResourceMetadata{
		Name:       "type-mismatch",
		Type:       resource.Agent, // Wrong type!
		SourceType: "local",
		SourceURL:  "file://" + cmdPath,
	}
	if err := metadata.Save(meta, repoDir); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Test: verify should detect type mismatch
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "Type mismatches") {
		t.Errorf("Expected 'Type mismatches' error, got: %s", output)
	}

	if !strings.Contains(output, "type-mismatch") {
		t.Errorf("Expected 'type-mismatch' in output, got: %s", output)
	}

	if !strings.Contains(output, "command") || !strings.Contains(output, "agent") {
		t.Errorf("Expected both 'command' and 'agent' in output, got: %s", output)
	}

	// Should be an error (exit code 1)
	if !strings.Contains(output, "ERRORS found") {
		t.Errorf("Type mismatches should be an error")
	}
}

// TestCLIRepoVerifyFixFlag tests the --fix flag functionality
func TestCLIRepoVerifyFixFlag(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	manager := repo.NewManagerWithPath(repoDir)

	// Create resource without metadata
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "fix-test.md")
	cmdContent := `---
description: Test --fix flag
---
# Fix Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Verify metadata doesn't exist
	_, err := manager.GetMetadata("fix-test", resource.Command)
	if err == nil {
		t.Fatalf("Metadata should not exist before --fix")
	}

	// Test: verify --fix should create missing metadata
	output, _ := runAimgr(t, "repo", "verify", "--fix")

	if !strings.Contains(output, "fix-test") {
		t.Errorf("Expected 'fix-test' in output, got: %s", output)
	}

	if !strings.Contains(output, "Created metadata") {
		t.Errorf("Expected 'Created metadata' in output, got: %s", output)
	}

	// Verify metadata now exists
	meta, err := manager.GetMetadata("fix-test", resource.Command)
	if err != nil {
		t.Errorf("Metadata should exist after --fix, got error: %v", err)
	}

	if meta.Name != "fix-test" {
		t.Errorf("Metadata name = %s, want fix-test", meta.Name)
	}

	if meta.Type != resource.Command {
		t.Errorf("Metadata type = %s, want command", meta.Type)
	}
}

// TestCLIRepoVerifyFixOrphanedMetadata tests --fix removes orphaned metadata
func TestCLIRepoVerifyFixOrphanedMetadata(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a resource
	cmdPath := filepath.Join(testDir, "fix-orphan.md")
	cmdContent := `---
description: Test --fix with orphaned metadata
---
# Fix Orphan
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Delete the resource file
	manager := repo.NewManagerWithPath(repoDir)
	resourcePath := manager.GetPath("fix-orphan", resource.Command)
	if err := os.Remove(resourcePath); err != nil {
		t.Fatalf("Failed to delete resource: %v", err)
	}

	// Verify metadata exists before --fix
	_, err = manager.GetMetadata("fix-orphan", resource.Command)
	if err != nil {
		t.Fatalf("Metadata should exist before --fix, got error: %v", err)
	}

	// Test: verify --fix should remove orphaned metadata
	output, _ := runAimgr(t, "repo", "verify", "--fix")

	if !strings.Contains(output, "fix-orphan") {
		t.Errorf("Expected 'fix-orphan' in output, got: %s", output)
	}

	if !strings.Contains(output, "Removed orphaned metadata") {
		t.Errorf("Expected 'Removed orphaned metadata' in output, got: %s", output)
	}

	// Verify metadata is now removed
	_, err = manager.GetMetadata("fix-orphan", resource.Command)
	if err == nil {
		t.Errorf("Metadata should be removed after --fix")
	}
}

// TestCLIRepoVerifyJSONWithIssues tests JSON output with actual issues
func TestCLIRepoVerifyJSONWithIssues(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	manager := repo.NewManagerWithPath(repoDir)

	// Create resource without metadata
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "json-test.md")
	cmdContent := `---
description: Test JSON output
---
# JSON Test
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Create orphaned metadata
	orphanPath := filepath.Join(testDir, "orphan-json.md")
	if err := os.WriteFile(orphanPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create orphan command: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", orphanPath)
	if err != nil {
		t.Fatalf("Failed to add orphan command: %v", err)
	}

	// Delete the orphan resource
	orphanRepoPath := manager.GetPath("orphan-json", resource.Command)
	if err := os.Remove(orphanRepoPath); err != nil {
		t.Fatalf("Failed to delete orphan resource: %v", err)
	}

	// Test: verify --json
	output, _ := runAimgr(t, "repo", "verify", "--json")

	var result struct {
		ResourcesWithoutMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"resources_without_metadata"`
		OrphanedMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"orphaned_metadata"`
		HasErrors   bool `json:"has_errors"`
		HasWarnings bool `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Check that issues were detected
	foundNoMeta := false
	for _, issue := range result.ResourcesWithoutMetadata {
		if issue.Name == "json-test" {
			foundNoMeta = true
			break
		}
	}
	if !foundNoMeta {
		t.Errorf("Expected json-test in resources_without_metadata")
	}

	foundOrphan := false
	for _, issue := range result.OrphanedMetadata {
		if issue.Name == "orphan-json" {
			foundOrphan = true
			break
		}
	}
	if !foundOrphan {
		t.Errorf("Expected orphan-json in orphaned_metadata")
	}

	// Should have errors (orphaned metadata)
	if !result.HasErrors {
		t.Errorf("Expected has_errors=true (orphaned metadata)")
	}

	// Should have warnings (resource without metadata)
	if !result.HasWarnings {
		t.Errorf("Expected has_warnings=true (resource without metadata)")
	}
}

// TestCLIRepoVerifyHealthyRepo tests verification of a healthy repository
func TestCLIRepoVerifyHealthyRepo(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create and add a valid resource
	cmdPath := filepath.Join(testDir, "healthy-cmd.md")
	cmdContent := `---
description: A healthy command
---
# Healthy Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
	if err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}

	// Test: verify should find no issues
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "No issues found") {
		t.Errorf("Expected 'No issues found', got: %s", output)
	}

	if !strings.Contains(output, "healthy") {
		t.Errorf("Expected 'healthy' in output, got: %s", output)
	}

	// Verify JSON output also shows healthy
	jsonOutput, _ := runAimgr(t, "repo", "verify", "--json")

	var result struct {
		HasErrors   bool `json:"has_errors"`
		HasWarnings bool `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.HasErrors {
		t.Errorf("Healthy repo should have has_errors=false")
	}

	if result.HasWarnings {
		t.Errorf("Healthy repo should have has_warnings=false")
	}
}

// TestCLIRepoVerifySkillIssues tests verification with skill resources
func TestCLIRepoVerifySkillIssues(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create skill without metadata
	manager := repo.NewManagerWithPath(repoDir)
	skillsDir := filepath.Join(repoDir, "skills", "verify-skill")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills directory: %v", err)
	}

	skillPath := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: verify-skill
description: Skill for verify testing
---
# Verify Skill
`
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Test: verify should detect skill without metadata
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "Resources without metadata") {
		t.Errorf("Expected 'Resources without metadata', got: %s", output)
	}

	if !strings.Contains(output, "verify-skill") {
		t.Errorf("Expected 'verify-skill' in output, got: %s", output)
	}

	// Test --fix creates metadata for skill
	output, _ = runAimgr(t, "repo", "verify", "--fix")

	if !strings.Contains(output, "Created metadata") {
		t.Errorf("Expected 'Created metadata', got: %s", output)
	}

	// Verify metadata was created
	_, err := manager.GetMetadata("verify-skill", resource.Skill)
	if err != nil {
		t.Errorf("Metadata should be created for skill, got error: %v", err)
	}
}

// TestCLIRepoVerifyAgentIssues tests verification with agent resources
func TestCLIRepoVerifyAgentIssues(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create agent without metadata
	manager := repo.NewManagerWithPath(repoDir)
	agentsDir := filepath.Join(repoDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "verify-agent.md")
	agentContent := `---
description: Agent for verify testing
---
# Verify Agent
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test: verify should detect agent without metadata
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "Resources without metadata") {
		t.Errorf("Expected 'Resources without metadata', got: %s", output)
	}

	if !strings.Contains(output, "verify-agent") {
		t.Errorf("Expected 'verify-agent' in output, got: %s", output)
	}

	// Test --fix creates metadata for agent
	output, _ = runAimgr(t, "repo", "verify", "--fix")

	if !strings.Contains(output, "Created metadata") {
		t.Errorf("Expected 'Created metadata', got: %s", output)
	}

	// Verify metadata was created
	_, err := manager.GetMetadata("verify-agent", resource.Agent)
	if err != nil {
		t.Errorf("Metadata should be created for agent, got error: %v", err)
	}
}

// TestCLIRepoVerifyEmptyRepo tests verify on empty repository
func TestCLIRepoVerifyEmptyRepo(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create empty repo structure
	for _, dir := range []string{"commands", "skills", "agents"} {
		if err := os.MkdirAll(filepath.Join(repoDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create %s directory: %v", dir, err)
		}
	}

	// Test: verify on empty repo
	output, _ := runAimgr(t, "repo", "verify")

	if !strings.Contains(output, "No issues found") {
		t.Errorf("Expected 'No issues found' for empty repo, got: %s", output)
	}
}

// TestCLIRepoVerifyMultipleIssues tests verify with multiple types of issues
func TestCLIRepoVerifyMultipleIssues(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	manager := repo.NewManagerWithPath(repoDir)

	// Issue 1: Resource without metadata (command)
	commandsDir := filepath.Join(repoDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	noMetaPath := filepath.Join(commandsDir, "no-meta.md")
	if err := os.WriteFile(noMetaPath, []byte(`---
description: No metadata
---
# No Meta
`), 0644); err != nil {
		t.Fatalf("Failed to create no-meta command: %v", err)
	}

	// Issue 2: Orphaned metadata (skill)
	orphanPath := filepath.Join(testDir, "orphan-skill")
	if err := os.MkdirAll(orphanPath, 0755); err != nil {
		t.Fatalf("Failed to create orphan skill dir: %v", err)
	}
	skillPath := filepath.Join(orphanPath, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(`---
name: orphan-skill
description: Orphaned skill
---
# Orphan
`), 0644); err != nil {
		t.Fatalf("Failed to create orphan skill: %v", err)
	}

	_, err := runAimgr(t, "repo", "add", "--force", orphanPath)
	if err != nil {
		t.Fatalf("Failed to add orphan skill: %v", err)
	}

	// Delete the skill resource to make it orphaned
	skillRepoPath := manager.GetPath("orphan-skill", resource.Skill)
	if err := os.RemoveAll(skillRepoPath); err != nil {
		t.Fatalf("Failed to delete skill resource: %v", err)
	}

	// Test: verify should find both issues
	output, _ := runAimgr(t, "repo", "verify")

	// Should detect resource without metadata
	if !strings.Contains(output, "Resources without metadata") {
		t.Errorf("Expected 'Resources without metadata', got: %s", output)
	}
	if !strings.Contains(output, "no-meta") {
		t.Errorf("Expected 'no-meta' in output, got: %s", output)
	}

	// Should detect orphaned metadata
	if !strings.Contains(output, "Orphaned metadata") {
		t.Errorf("Expected 'Orphaned metadata', got: %s", output)
	}
	if !strings.Contains(output, "orphan-skill") {
		t.Errorf("Expected 'orphan-skill' in output, got: %s", output)
	}

	// Should have errors (orphaned metadata)
	if !strings.Contains(output, "ERRORS found") {
		t.Errorf("Expected 'ERRORS found', got: %s", output)
	}

	// Should suggest --fix
	if !strings.Contains(output, "--fix") {
		t.Errorf("Expected suggestion to use --fix, got: %s", output)
	}
}
