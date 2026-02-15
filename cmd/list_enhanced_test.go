package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"gopkg.in/yaml.v3"
)

// TestListEnhanced_SingleTarget tests a resource installed to a single target (claude only)
func TestListEnhanced_SingleTarget(t *testing.T) {
	// Setup: Create temp repo with a test resource
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	// Create project directory with claude installation
	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create a command in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	commandPath := filepath.Join(sourceTempDir, "test-cmd.md")
	writeTestCommand(t, commandPath, "test-cmd", "Test command description")

	// Initialize repo manager
	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	// Add command to repo (this will copy it into repoPath/commands/)
	if err := mgr.AddCommand(commandPath, "file://"+commandPath, "file"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	// Install to claude only (create symlink)
	installResource(t, projectPath, repoPath, "claude", "commands", "test-cmd.md")

	// Change to project directory (list command uses os.Getwd())
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON output format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		// Parse JSON output
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, output)
		}

		// Verify resources array exists
		resources, ok := result["resources"].([]interface{})
		if !ok {
			t.Fatalf("expected resources to be an array, got: %T", result["resources"])
		}

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		// Verify first resource
		res := resources[0].(map[string]interface{})

		// Check targets field is an array with claude
		targets, ok := res["targets"].([]interface{})
		if !ok {
			t.Fatalf("expected targets to be array, got: %T", res["targets"])
		}

		if len(targets) != 1 {
			t.Fatalf("expected 1 target, got %d", len(targets))
		}

		if targets[0] != "claude" {
			t.Errorf("expected target 'claude', got %v", targets[0])
		}
	})

	// Test YAML output format
	t.Run("yaml_format", func(t *testing.T) {
		output := captureListOutput(t, "yaml", "")

		// Parse YAML output
		var result map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse YAML output: %v\nOutput: %s", err, output)
		}

		// Verify resources array
		resources, ok := result["resources"].([]interface{})
		if !ok {
			t.Fatalf("expected resources to be an array, got: %T", result["resources"])
		}

		res := resources[0].(map[string]interface{})

		// Check targets field is an array
		targets, ok := res["targets"].([]interface{})
		if !ok {
			t.Fatalf("expected targets to be array in YAML, got: %T", res["targets"])
		}

		if len(targets) != 1 {
			t.Fatalf("expected 1 target, got %d", len(targets))
		}

		if targets[0] != "claude" {
			t.Errorf("expected target 'claude', got %v", targets[0])
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Verify table contains "claude" in the targets column
		if !strings.Contains(output, "claude") {
			t.Errorf("expected table to contain 'claude', got:\n%s", output)
		}

		// Verify it's not showing multiple targets (no comma)
		lines := strings.Split(output, "\n")
		foundTargetLine := false
		for _, line := range lines {
			if strings.Contains(line, "test-cmd") {
				foundTargetLine = true
				// Note: There might be commas in description, so check specifically around "claude"
				if strings.Contains(line, "claude,") || strings.Contains(line, ", claude") {
					t.Errorf("expected single target without comma, got line: %s", line)
				}
			}
		}

		if !foundTargetLine {
			t.Errorf("did not find test-cmd in table output")
		}
	})
}

// TestListEnhanced_MultipleTargets tests a resource installed to multiple targets
func TestListEnhanced_MultipleTargets(t *testing.T) {
	// Setup
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude", "opencode"})

	// Create skill in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	skillPath := filepath.Join(sourceTempDir, "test-skill")
	writeTestSkill(t, skillPath, "test-skill", "Test skill description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	if err := mgr.AddSkill(skillPath, "file://"+skillPath, "file"); err != nil {
		t.Fatalf("failed to add skill: %v", err)
	}

	// Install to both claude and opencode
	installResource(t, projectPath, repoPath, "claude", "skills", "test-skill")
	installResource(t, projectPath, repoPath, "opencode", "skills", "test-skill")

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		targets, ok := res["targets"].([]interface{})
		if !ok {
			t.Fatalf("expected targets to be array, got: %T", res["targets"])
		}

		if len(targets) != 2 {
			t.Fatalf("expected 2 targets, got %d", len(targets))
		}

		// Verify both targets are present (order may vary)
		targetStrs := make([]string, len(targets))
		for i, t := range targets {
			targetStrs[i] = t.(string)
		}

		hasClaude := false
		hasOpenCode := false
		for _, target := range targetStrs {
			if target == "claude" {
				hasClaude = true
			}
			if target == "opencode" {
				hasOpenCode = true
			}
		}

		if !hasClaude || !hasOpenCode {
			t.Errorf("expected both 'claude' and 'opencode', got: %v", targetStrs)
		}
	})

	// Test table format - should show comma-separated list
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Should contain both tool names
		if !strings.Contains(output, "claude") {
			t.Errorf("expected table to contain 'claude'")
		}
		if !strings.Contains(output, "opencode") {
			t.Errorf("expected table to contain 'opencode'")
		}

		// Should have comma separating them (in the same line)
		lines := strings.Split(output, "\n")
		foundBoth := false
		for _, line := range lines {
			if strings.Contains(line, "test-skill") {
				if strings.Contains(line, "claude") && strings.Contains(line, "opencode") {
					foundBoth = true
					// Should have comma between them
					if !strings.Contains(line, ",") {
						t.Errorf("expected comma between targets in line: %s", line)
					}
				}
			}
		}

		if !foundBoth {
			t.Errorf("expected to find both claude and opencode in same line")
		}
	})
}

// TestListEnhanced_NotInManifest tests resource installed but not in ai.package.yaml
func TestListEnhanced_NotInManifest(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create agent in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	agentPath := filepath.Join(sourceTempDir, "test-agent.md")
	writeTestAgent(t, agentPath, "test-agent", "Test agent description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	if err := mgr.AddAgent(agentPath, "file://"+agentPath, "file"); err != nil {
		t.Fatalf("failed to add agent: %v", err)
	}

	// Install resource
	installResource(t, projectPath, repoPath, "claude", "agents", "test-agent.md")

	// Create manifest WITHOUT this resource
	m := &manifest.Manifest{
		Resources: []string{"skill/other-skill"}, // Different resource
	}
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		syncStatus, ok := res["sync_status"].(string)
		if !ok {
			t.Fatalf("expected sync_status to be string, got: %T", res["sync_status"])
		}

		if syncStatus != "not-in-manifest" {
			t.Errorf("expected sync_status 'not-in-manifest', got: %s", syncStatus)
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Check for the "*" symbol (not-in-manifest indicator)
		lines := strings.Split(output, "\n")
		foundSymbol := false
		for _, line := range lines {
			if strings.Contains(line, "test-agent") {
				// The sync column should have "*"
				if strings.Contains(line, "*") {
					foundSymbol = true
				}
			}
		}

		if !foundSymbol {
			t.Errorf("expected to find '*' symbol for not-in-manifest status")
		}

		// Verify legend exists
		if !strings.Contains(output, "* = Not in manifest") {
			t.Errorf("expected legend to explain '*' symbol")
		}
	})
}

// TestListEnhanced_NotInstalled tests resource in manifest but not installed
func TestListEnhanced_NotInstalled(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create command in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	commandPath := filepath.Join(sourceTempDir, "test-cmd.md")
	writeTestCommand(t, commandPath, "test-cmd", "Test command description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	// Add command to repo (this will copy it into repoPath/commands/)
	if err := mgr.AddCommand(commandPath, "file://"+commandPath, "file"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	if err := mgr.AddCommand(commandPath, "file://"+commandPath, "file"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	// Create manifest WITH this resource, but DON'T install it
	m := &manifest.Manifest{
		Resources: []string{"command/test-cmd"},
	}
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		// Check targets is empty array
		targets, ok := res["targets"].([]interface{})
		if !ok {
			t.Fatalf("expected targets to be array, got: %T", res["targets"])
		}

		if len(targets) != 0 {
			t.Errorf("expected empty targets array, got: %v", targets)
		}

		// Check sync status
		syncStatus := res["sync_status"].(string)
		if syncStatus != "not-installed" {
			t.Errorf("expected sync_status 'not-installed', got: %s", syncStatus)
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Check for the "⚠" symbol (not-installed indicator)
		lines := strings.Split(output, "\n")
		foundWarning := false
		for _, line := range lines {
			if strings.Contains(line, "test-cmd") {
				if strings.Contains(line, "⚠") {
					foundWarning = true
				}
			}
		}

		if !foundWarning {
			t.Errorf("expected to find '⚠' symbol for not-installed status")
		}

		// Verify legend
		if !strings.Contains(output, "⚠ = Not installed") {
			t.Errorf("expected legend to explain '⚠' symbol")
		}
	})
}

// TestListEnhanced_InSync tests resource in sync (installed and in manifest)
func TestListEnhanced_InSync(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create skill in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	skillPath := filepath.Join(sourceTempDir, "test-skill")
	writeTestSkill(t, skillPath, "test-skill", "Test skill description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	if err := mgr.AddSkill(skillPath, "file://"+skillPath, "file"); err != nil {
		t.Fatalf("failed to add skill: %v", err)
	}

	// Install resource
	installResource(t, projectPath, repoPath, "claude", "skills", "test-skill")

	// Create manifest WITH this resource
	m := &manifest.Manifest{
		Resources: []string{"skill/test-skill"},
	}
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	if err := m.Save(manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		syncStatus := res["sync_status"].(string)
		if syncStatus != "in-sync" {
			t.Errorf("expected sync_status 'in-sync', got: %s", syncStatus)
		}

		// Also verify targets is present
		targets := res["targets"].([]interface{})
		if len(targets) != 1 || targets[0] != "claude" {
			t.Errorf("expected targets ['claude'], got: %v", targets)
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Check for the "✓" symbol (in-sync indicator)
		lines := strings.Split(output, "\n")
		foundCheck := false
		for _, line := range lines {
			if strings.Contains(line, "test-skill") {
				if strings.Contains(line, "✓") {
					foundCheck = true
				}
			}
		}

		if !foundCheck {
			t.Errorf("expected to find '✓' symbol for in-sync status")
		}

		// Verify legend
		if !strings.Contains(output, "✓ = In sync") {
			t.Errorf("expected legend to explain '✓' symbol")
		}
	})
}

// TestListEnhanced_NoManifest tests behavior when no ai.package.yaml exists
func TestListEnhanced_NoManifest(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create command in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	commandPath := filepath.Join(sourceTempDir, "test-cmd.md")
	writeTestCommand(t, commandPath, "test-cmd", "Test command description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	// Add command to repo (this will copy it into repoPath/commands/)
	if err := mgr.AddCommand(commandPath, "file://"+commandPath, "file"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	if err := mgr.AddCommand(commandPath, "file://"+commandPath, "file"); err != nil {
		t.Fatalf("failed to add command: %v", err)
	}

	// Install resource
	installResource(t, projectPath, repoPath, "claude", "commands", "test-cmd.md")

	// Do NOT create manifest file

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		syncStatus := res["sync_status"].(string)
		if syncStatus != "no-manifest" {
			t.Errorf("expected sync_status 'no-manifest', got: %s", syncStatus)
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Check for the "-" symbol (no-manifest indicator) in sync column
		// Table format: | name | targets | sync | desc |
		// After Fields: [│, name, │, targets, │, sync, │, desc, │]
		lines := strings.Split(output, "\n")
		foundDash := false
		for _, line := range lines {
			if strings.Contains(line, "test-cmd") {
				fields := strings.Fields(line)
				// Sync column is at index 5 (after │, name, │, targets, │)
				if len(fields) >= 6 && fields[5] == "-" {
					foundDash = true
				}
			}
		}

		if !foundDash {
			t.Errorf("expected to find '-' symbol for no-manifest status")
		}

		// Verify legend
		if !strings.Contains(output, "- = No manifest") {
			t.Errorf("expected legend to explain '-' symbol")
		}
	})
}

// TestListEnhanced_CopilotOnly tests Copilot-only installation (skills only)
func TestListEnhanced_CopilotOnly(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	// Setup only copilot directory
	if err := os.MkdirAll(filepath.Join(projectPath, ".github", "skills"), 0755); err != nil {
		t.Fatalf("failed to create copilot skills dir: %v", err)
	}

	// Create skill in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	skillPath := filepath.Join(sourceTempDir, "test-skill")
	writeTestSkill(t, skillPath, "test-skill", "Test skill description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	if err := mgr.AddSkill(skillPath, "file://"+skillPath, "file"); err != nil {
		t.Fatalf("failed to add skill: %v", err)
	}

	// Install to copilot only
	installResource(t, projectPath, repoPath, "copilot", "skills", "test-skill")

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		resources := result["resources"].([]interface{})
		res := resources[0].(map[string]interface{})

		targets := res["targets"].([]interface{})
		if len(targets) != 1 {
			t.Fatalf("expected 1 target, got %d", len(targets))
		}

		if targets[0] != "copilot" {
			t.Errorf("expected target 'copilot', got: %v", targets[0])
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		if !strings.Contains(output, "copilot") {
			t.Errorf("expected table to contain 'copilot'")
		}

		// Should NOT contain claude or opencode
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "test-skill") {
				if strings.Contains(line, "claude") || strings.Contains(line, "opencode") {
					t.Errorf("expected only copilot target, but found: %s", line)
				}
			}
		}
	})
}

// TestListEnhanced_NoInstallations tests resource with no installations (empty targets)
func TestListEnhanced_NoInstallations(t *testing.T) {
	repoPath := t.TempDir()
	setupEnhancedTestRepo(t, repoPath)

	projectPath := t.TempDir()
	setupToolDirs(t, projectPath, []string{"claude"})

	// Create agent in a separate temp directory (not in repo)
	sourceTempDir := t.TempDir()
	agentPath := filepath.Join(sourceTempDir, "test-agent.md")
	writeTestAgent(t, agentPath, "test-agent", "Test agent description")

	t.Setenv("AIMGR_REPO_PATH", repoPath)

	mgr, err := repo.NewManager()
	if err != nil {
		t.Fatalf("failed to create repo manager: %v", err)
	}

	if err := mgr.AddAgent(agentPath, "file://"+agentPath, "file"); err != nil {
		t.Fatalf("failed to add agent: %v", err)
	}

	// Do NOT install the resource anywhere

	// Change to project directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to change to project dir: %v", err)
	}

	// Test JSON format
	t.Run("json_format", func(t *testing.T) {
		output := captureListOutput(t, "json", "")

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
		}

		resources := result["resources"].([]interface{})
		if len(resources) == 0 {
			t.Fatalf("expected at least 1 resource in repo, got none")
		}
		res := resources[0].(map[string]interface{})

		// Verify targets is empty array (not null, not string)
		targets, ok := res["targets"].([]interface{})
		if !ok {
			t.Fatalf("expected targets to be array, got: %T", res["targets"])
		}

		if len(targets) != 0 {
			t.Errorf("expected empty targets array [], got: %v", targets)
		}
	})

	// Test table format
	t.Run("table_format", func(t *testing.T) {
		output := captureListOutput(t, "table", "")

		// Check that targets column shows "-" for no installations
		// The table has format: | name | targets | sync | desc |
		// After Fields split, we get: [│, name, │, targets, │, sync, │, desc, │]
		lines := strings.Split(output, "\n")
		foundDash := false
		for _, line := range lines {
			if strings.Contains(line, "test-agent") {
				// Should have "-" in targets column (index 3 after splitting by whitespace)
				fields := strings.Fields(line)
				if len(fields) >= 4 && fields[3] == "-" {
					foundDash = true
				}
			}
		}

		if !foundDash {
			t.Errorf("expected to find '-' in targets column for uninstalled resource")
		}
	})
}

// Helper functions

// setupEnhancedTestRepo creates basic repo directory structure
func setupEnhancedTestRepo(t *testing.T, repoPath string) {
	t.Helper()

	dirs := []string{"commands", "skills", "agents", "packages"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(repoPath, dir), 0755); err != nil {
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
	}
}

// setupToolDirs creates tool-specific directories in project
func setupToolDirs(t *testing.T, projectPath string, tools []string) {
	t.Helper()

	for _, tool := range tools {
		var dirs []string
		switch tool {
		case "claude":
			dirs = []string{".claude/commands", ".claude/skills", ".claude/agents"}
		case "opencode":
			dirs = []string{".opencode/commands", ".opencode/skills", ".opencode/agents"}
		case "copilot":
			dirs = []string{".github/skills"}
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
				t.Fatalf("failed to create %s dir: %v", dir, err)
			}
		}
	}
}

// writeTestCommand creates a test command file
func writeTestCommand(t *testing.T, path string, name string, description string) {
	t.Helper()

	content := fmt.Sprintf(`---
description: %s
---
# %s

%s

`+"```bash"+`
echo "Test command"
`+"```"+`
`, description, name, description)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}
}

// writeTestSkill creates a test skill directory with SKILL.md
func writeTestSkill(t *testing.T, skillDir string, name string, description string) {
	t.Helper()

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := fmt.Sprintf(`---
description: %s
---
# %s

%s

## Instructions
Test skill instructions.
`, description, name, description)

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}
}

// writeTestAgent creates a test agent file
func writeTestAgent(t *testing.T, path string, name string, description string) {
	t.Helper()

	content := fmt.Sprintf(`---
description: %s
---
# %s

%s

## Instructions
Test agent instructions.
`, description, name, description)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}
}

// installResource creates a symlink from project tool dir to repo resource
func installResource(t *testing.T, projectPath string, repoPath string, tool string, resourceType string, resourceFile string) {
	t.Helper()

	var targetDir string
	switch tool {
	case "claude":
		switch resourceType {
		case "commands":
			targetDir = ".claude/commands"
		case "skills":
			targetDir = ".claude/skills"
		case "agents":
			targetDir = ".claude/agents"
		}
	case "opencode":
		switch resourceType {
		case "commands":
			targetDir = ".opencode/commands"
		case "skills":
			targetDir = ".opencode/skills"
		case "agents":
			targetDir = ".opencode/agents"
		}
	case "copilot":
		if resourceType == "skills" {
			targetDir = ".github/skills"
		}
	}

	if targetDir == "" {
		t.Fatalf("unsupported tool/resource type: %s/%s", tool, resourceType)
	}

	sourcePath := filepath.Join(repoPath, resourceType, resourceFile)
	targetPath := filepath.Join(projectPath, targetDir, resourceFile)

	if err := os.Symlink(sourcePath, targetPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
}

// captureListOutput runs the list command and captures output
func captureListOutput(t *testing.T, format string, pattern string) string {
	t.Helper()

	// Save and redirect stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Set format flag
	originalFormat := formatFlag
	formatFlag = format
	defer func() { formatFlag = originalFormat }()

	// Build args
	var args []string
	if pattern != "" {
		args = append(args, pattern)
	}

	// Capture output in goroutine
	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		buf.ReadFrom(r)
		outChan <- buf.String()
	}()

	// Execute command RunE directly
	err = listCmd.RunE(listCmd, args)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	output := <-outChan

	if err != nil {
		t.Fatalf("command execution failed: %v\nOutput: %s", err, output)
	}

	return output
}
