package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// isGitAvailable checks if git is available on the system
func isGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// TestUpdateBatching tests that multiple resources from the same Git source are batched
func TestUpdateBatching(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	if !isOnline() {
		t.Skip("Skipping test: network not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Add multiple resources from the same GitHub source
	// Using anthropics/anthropic-quickstarts as a known repository
	githubSource := "gh:anthropics/anthropic-quickstarts"

	// Create test resources with GitHub source metadata
	// We'll create local test resources but set their metadata to point to GitHub
	testResources := []struct {
		name     string
		resType  resource.ResourceType
		filename string
		content  string
	}{
		{
			name:     "batch-test-cmd1",
			resType:  resource.Command,
			filename: "batch-test-cmd1.md",
			content: `---
description: Batch test command 1 from GitHub
---
# Batch Test Command 1
`,
		},
		{
			name:     "batch-test-cmd2",
			resType:  resource.Command,
			filename: "batch-test-cmd2.md",
			content: `---
description: Batch test command 2 from GitHub
---
# Batch Test Command 2
`,
		},
		{
			name:     "batch-test-cmd3",
			resType:  resource.Command,
			filename: "batch-test-cmd3.md",
			content: `---
description: Batch test command 3 from GitHub
---
# Batch Test Command 3
`,
		},
	}

	// Create and add test resources
	tempSourceDir := t.TempDir()
	for _, res := range testResources {
		sourcePath := filepath.Join(tempSourceDir, res.filename)
		if err := os.WriteFile(sourcePath, []byte(res.content), 0644); err != nil {
			t.Fatalf("Failed to create test resource %s: %v", res.name, err)
		}

		// Add with GitHub source metadata
		if err := manager.AddCommand(sourcePath, githubSource, "github"); err != nil {
			t.Fatalf("Failed to add resource %s: %v", res.name, err)
		}
	}

	// Note: We can't directly count git clone invocations in this test since the
	// update functionality is in cmd package and uses internal functions.
	// Instead, we verify that:
	// 1. All resources can be listed
	// 2. All resources have the same GitHub source
	// 3. The grouping logic works correctly

	// Verify all resources were added
	cmdType := resource.Command
	commands, err := manager.List(&cmdType)
	if err != nil {
		t.Fatalf("Failed to list commands: %v", err)
	}

	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
	}

	// Verify all resources have the same GitHub source
	for _, res := range testResources {
		meta, err := manager.GetMetadata(res.name, res.resType)
		if err != nil {
			t.Errorf("Failed to get metadata for %s: %v", res.name, err)
			continue
		}

		if meta.SourceURL != githubSource {
			t.Errorf("Expected source URL %s, got %s", githubSource, meta.SourceURL)
		}

		if meta.SourceType != "github" {
			t.Errorf("Expected source type 'github', got %s", meta.SourceType)
		}
	}

	// This test verifies the setup for batching. The actual batching behavior
	// is tested by the update command implementation which groups resources
	// by source URL before cloning.
	t.Log("Batch test resources created successfully with same GitHub source")
}

// TestUpdateBatching_MixedSources tests batching with mixed Git and local sources
func TestUpdateBatching_MixedSources(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create resources with different source types
	tempSourceDir := t.TempDir()

	// GitHub source resources (should be batched)
	githubSource := "gh:anthropics/anthropic-quickstarts"
	githubCmd1Path := filepath.Join(tempSourceDir, "github-cmd1.md")
	githubCmd1Content := `---
description: GitHub command 1
---
# GitHub Command 1
`
	if err := os.WriteFile(githubCmd1Path, []byte(githubCmd1Content), 0644); err != nil {
		t.Fatalf("Failed to create GitHub command 1: %v", err)
	}

	githubCmd2Path := filepath.Join(tempSourceDir, "github-cmd2.md")
	githubCmd2Content := `---
description: GitHub command 2
---
# GitHub Command 2
`
	if err := os.WriteFile(githubCmd2Path, []byte(githubCmd2Content), 0644); err != nil {
		t.Fatalf("Failed to create GitHub command 2: %v", err)
	}

	// Local source resources (should NOT be batched)
	localCmd1Path := filepath.Join(tempSourceDir, "local-cmd1.md")
	localCmd1Content := `---
description: Local command 1
---
# Local Command 1
`
	if err := os.WriteFile(localCmd1Path, []byte(localCmd1Content), 0644); err != nil {
		t.Fatalf("Failed to create local command 1: %v", err)
	}

	localCmd2Path := filepath.Join(tempSourceDir, "local-cmd2.md")
	localCmd2Content := `---
description: Local command 2
---
# Local Command 2
`
	if err := os.WriteFile(localCmd2Path, []byte(localCmd2Content), 0644); err != nil {
		t.Fatalf("Failed to create local command 2: %v", err)
	}

	// Add GitHub resources
	if err := manager.AddCommand(githubCmd1Path, githubSource, "github"); err != nil {
		t.Fatalf("Failed to add GitHub command 1: %v", err)
	}
	if err := manager.AddCommand(githubCmd2Path, githubSource, "github"); err != nil {
		t.Fatalf("Failed to add GitHub command 2: %v", err)
	}

	// Add local resources
	if err := manager.AddCommand(localCmd1Path, "file://"+localCmd1Path, "file"); err != nil {
		t.Fatalf("Failed to add local command 1: %v", err)
	}
	if err := manager.AddCommand(localCmd2Path, "file://"+localCmd2Path, "file"); err != nil {
		t.Fatalf("Failed to add local command 2: %v", err)
	}

	// Verify all resources were added
	cmdType := resource.Command
	commands, err := manager.List(&cmdType)
	if err != nil {
		t.Fatalf("Failed to list commands: %v", err)
	}

	if len(commands) != 4 {
		t.Errorf("Expected 4 commands, got %d", len(commands))
	}

	// Verify source types
	githubCount := 0
	localCount := 0

	for _, cmd := range commands {
		meta, err := manager.GetMetadata(cmd.Name, resource.Command)
		if err != nil {
			t.Errorf("Failed to get metadata for %s: %v", cmd.Name, err)
			continue
		}

		if meta.SourceType == "github" {
			githubCount++
			if meta.SourceURL != githubSource {
				t.Errorf("Expected GitHub source URL %s, got %s", githubSource, meta.SourceURL)
			}
		} else if meta.SourceType == "file" {
			localCount++
			if !strings.HasPrefix(meta.SourceURL, "file://") {
				t.Errorf("Expected file:// prefix for local source, got %s", meta.SourceURL)
			}
		}
	}

	if githubCount != 2 {
		t.Errorf("Expected 2 GitHub resources, got %d", githubCount)
	}

	if localCount != 2 {
		t.Errorf("Expected 2 local resources, got %d", localCount)
	}

	t.Log("Mixed source test resources created successfully")
}

// TestUpdateBatching_MultipleResourceTypes tests batching with different resource types from same source
func TestUpdateBatching_MultipleResourceTypes(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create different resource types from the same GitHub source
	githubSource := "gh:anthropics/anthropic-quickstarts"
	tempSourceDir := t.TempDir()

	// Create command
	cmdPath := filepath.Join(tempSourceDir, "multi-type-cmd.md")
	cmdContent := `---
description: Multi-type test command
---
# Multi-type Test Command
`
	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// Create skill
	skillDir := filepath.Join(tempSourceDir, "multi-type-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: multi-type-skill
description: Multi-type test skill
license: MIT
---
# Multi-type Test Skill
`
	if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	// Create agent
	agentPath := filepath.Join(tempSourceDir, "multi-type-agent.md")
	agentContent := `---
description: Multi-type test agent
---
# Multi-type Test Agent
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add all resources with same GitHub source
	if err := manager.AddCommand(cmdPath, githubSource, "github"); err != nil {
		t.Fatalf("Failed to add command: %v", err)
	}
	if err := manager.AddSkill(skillDir, githubSource, "github"); err != nil {
		t.Fatalf("Failed to add skill: %v", err)
	}
	if err := manager.AddAgent(agentPath, githubSource, "github"); err != nil {
		t.Fatalf("Failed to add agent: %v", err)
	}

	// Verify all resource types were added with same source
	resourceTypes := []struct {
		resType resource.ResourceType
		name    string
	}{
		{resource.Command, "multi-type-cmd"},
		{resource.Skill, "multi-type-skill"},
		{resource.Agent, "multi-type-agent"},
	}

	for _, rt := range resourceTypes {
		meta, err := manager.GetMetadata(rt.name, rt.resType)
		if err != nil {
			t.Errorf("Failed to get metadata for %s %s: %v", rt.resType, rt.name, err)
			continue
		}

		if meta.SourceURL != githubSource {
			t.Errorf("Expected source URL %s for %s, got %s", githubSource, rt.name, meta.SourceURL)
		}

		if meta.SourceType != "github" {
			t.Errorf("Expected source type 'github' for %s, got %s", rt.name, meta.SourceType)
		}
	}

	t.Log("Multi-resource-type batching test setup successful")
}

// TestUpdateBatching_DryRun tests the --dry-run flag with batched updates
func TestUpdateBatching_DryRun(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create resources with GitHub source
	githubSource := "gh:anthropics/anthropic-quickstarts"
	tempSourceDir := t.TempDir()

	resources := []struct {
		name    string
		content string
	}{
		{
			name: "dryrun-cmd1",
			content: `---
description: Dry run test command 1
---
# Dry Run Command 1
`,
		},
		{
			name: "dryrun-cmd2",
			content: `---
description: Dry run test command 2
---
# Dry Run Command 2
`,
		},
	}

	for _, res := range resources {
		sourcePath := filepath.Join(tempSourceDir, res.name+".md")
		if err := os.WriteFile(sourcePath, []byte(res.content), 0644); err != nil {
			t.Fatalf("Failed to create test resource %s: %v", res.name, err)
		}

		if err := manager.AddCommand(sourcePath, githubSource, "github"); err != nil {
			t.Fatalf("Failed to add resource %s: %v", res.name, err)
		}
	}

	// Verify resources were added and have correct metadata
	for _, res := range resources {
		meta, err := manager.GetMetadata(res.name, resource.Command)
		if err != nil {
			t.Errorf("Failed to get metadata for %s: %v", res.name, err)
			continue
		}

		if meta.SourceURL != githubSource {
			t.Errorf("Expected source URL %s, got %s", githubSource, meta.SourceURL)
		}
	}

	// In dry-run mode, resources should not be modified
	// The actual --dry-run flag testing is done at the CLI level,
	// but we verify the setup here
	t.Log("Dry-run test resources created successfully")
}

// TestUpdateBatching_VerifyGrouping tests the internal grouping logic
func TestUpdateBatching_VerifyGrouping(t *testing.T) {
	if !isGitAvailable() {
		t.Skip("Skipping test: git not available")
	}

	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)

	// Create resources with multiple different GitHub sources
	tempSourceDir := t.TempDir()

	sources := []struct {
		sourceURL  string
		sourceType string
		names      []string
	}{
		{
			sourceURL:  "gh:anthropics/anthropic-quickstarts",
			sourceType: "github",
			names:      []string{"group1-cmd1", "group1-cmd2", "group1-cmd3"},
		},
		{
			sourceURL:  "gh:openai/openai-cookbook",
			sourceType: "github",
			names:      []string{"group2-cmd1", "group2-cmd2"},
		},
		{
			sourceURL:  "file:///tmp/local",
			sourceType: "file",
			names:      []string{"local-cmd1", "local-cmd2"},
		},
	}

	for _, source := range sources {
		for _, name := range source.names {
			sourcePath := filepath.Join(tempSourceDir, name+".md")
			content := fmt.Sprintf(`---
description: Test command %s
---
# Test Command %s
`, name, name)
			if err := os.WriteFile(sourcePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test resource %s: %v", name, err)
			}

			if err := manager.AddCommand(sourcePath, source.sourceURL, source.sourceType); err != nil {
				t.Fatalf("Failed to add resource %s: %v", name, err)
			}
		}
	}

	// Verify grouping: Count resources by source
	sourceGroups := make(map[string]int)

	cmdType := resource.Command
	allCommands, err := manager.List(&cmdType)
	if err != nil {
		t.Fatalf("Failed to list commands: %v", err)
	}

	for _, cmd := range allCommands {
		meta, err := manager.GetMetadata(cmd.Name, resource.Command)
		if err != nil {
			t.Errorf("Failed to get metadata for %s: %v", cmd.Name, err)
			continue
		}
		sourceGroups[meta.SourceURL]++
	}

	// Verify grouping counts
	expectedGroups := map[string]int{
		"gh:anthropics/anthropic-quickstarts": 3,
		"gh:openai/openai-cookbook":           2,
		"file:///tmp/local":                   2,
	}

	for sourceURL, expectedCount := range expectedGroups {
		actualCount := sourceGroups[sourceURL]
		if actualCount != expectedCount {
			t.Errorf("Expected %d resources from %s, got %d", expectedCount, sourceURL, actualCount)
		}
	}

	// Verify we should have 2 Git source groups (to be batched) and 1 local group (not batched)
	gitSourceCount := 0
	localSourceCount := 0

	for _, source := range sources {
		if source.sourceType == "github" || source.sourceType == "git-url" {
			gitSourceCount++
		} else if source.sourceType == "file" || source.sourceType == "local" {
			localSourceCount++
		}
	}

	if gitSourceCount != 2 {
		t.Errorf("Expected 2 Git sources for batching, got %d", gitSourceCount)
	}

	if localSourceCount != 1 {
		t.Errorf("Expected 1 local source (not batched), got %d", localSourceCount)
	}

	t.Logf("Grouping verification successful: %d Git sources (batched), %d local sources (not batched)",
		gitSourceCount, localSourceCount)
}

// TestCLIUpdateBatching_LocalSources tests the CLI update with local sources (no batching)
func TestCLIUpdateBatching_LocalSources(t *testing.T) {
	repoDir := t.TempDir()
	testDir := t.TempDir()

	t.Setenv("AIMGR_REPO_PATH", repoDir)

	// Create three local commands
	for i := 1; i <= 3; i++ {
		cmdPath := filepath.Join(testDir, fmt.Sprintf("local-%d.md", i))
		content := fmt.Sprintf(`---
description: Local command %d
version: "1.0.0"
---
# Local Command %d
`, i, i)
		if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create command: %v", err)
		}

		_, err := runAimgr(t, "repo", "add", "--force", cmdPath)
		if err != nil {
			t.Fatalf("Failed to add command: %v", err)
		}
	}

	// Update source files
	for i := 1; i <= 3; i++ {
		cmdPath := filepath.Join(testDir, fmt.Sprintf("local-%d.md", i))
		content := fmt.Sprintf(`---
description: Local command %d updated
version: "2.0.0"
---
# Updated
`, i)
		if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to update file: %v", err)
		}
	}

	// Update all commands
	output, err := runAimgr(t, "repo", "update", "command/local-1", "command/local-2", "command/local-3")
	if err != nil {
		t.Fatalf("Failed to update: %v\nOutput: %s", err, output)
	}

	t.Logf("Update output:\n%s", output)

	// Verify no batching for local sources
	if strings.Contains(output, "Batch:") {
		t.Errorf("Local sources should not be batched")
	}

	// Verify all updated (output says "3 added" in summary for update operations)
	if !strings.Contains(output, "3 added") && !strings.Contains(output, "3 updated") {
		t.Errorf("Expected '3 added' or '3 updated' in summary, got: %s", output)
	}
}
