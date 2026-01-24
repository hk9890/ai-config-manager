package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/config"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"gopkg.in/yaml.v3"
)

// Test helper: create a minimal test source directory with resources
func createTestSource(t *testing.T) string {
	t.Helper()

	sourceDir := t.TempDir()

	// Create directory structure
	if err := os.MkdirAll(filepath.Join(sourceDir, "commands"), 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sourceDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create skills dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(sourceDir, "agents"), 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create test commands
	testCommands := []struct {
		name    string
		content string
	}{
		{"sync-test-cmd", "Sync test command"},
		{"test-command", "Test command"},
		{"pdf-command", "PDF command"},
	}
	for _, cmd := range testCommands {
		content := fmt.Sprintf("---\ndescription: %s\n---\n# %s", cmd.content, cmd.name)
		path := filepath.Join(sourceDir, "commands", cmd.name+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create command %s: %v", cmd.name, err)
		}
	}

	// Create test skills
	testSkills := []struct {
		name    string
		content string
	}{
		{"sync-test-skill", "Sync test skill"},
		{"pdf-processing", "PDF processing skill"},
		{"image-processing", "Image processing skill"},
	}
	for _, skill := range testSkills {
		skillDir := filepath.Join(sourceDir, "skills", skill.name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatalf("failed to create skill dir %s: %v", skill.name, err)
		}
		content := fmt.Sprintf("---\ndescription: %s\n---\n# %s", skill.content, skill.name)
		path := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create skill %s: %v", skill.name, err)
		}
	}

	// Create test agents
	testAgents := []struct {
		name    string
		content string
	}{
		{"sync-test-agent", "Sync test agent"},
		{"code-reviewer", "Code reviewer agent"},
	}
	for _, agent := range testAgents {
		content := fmt.Sprintf("---\ndescription: %s\n---\n# %s", agent.content, agent.name)
		path := filepath.Join(sourceDir, "agents", agent.name+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create agent %s: %v", agent.name, err)
		}
	}

	return sourceDir
}

// Test helper: verify resources exist in repo
func verifyResourcesInRepo(t *testing.T, repoPath string, resourceType resource.ResourceType, names ...string) {
	t.Helper()

	for _, name := range names {
		var path string
		switch resourceType {
		case resource.Command:
			path = filepath.Join(repoPath, "commands", name+".md")
		case resource.Skill:
			path = filepath.Join(repoPath, "skills", name)
		case resource.Agent:
			path = filepath.Join(repoPath, "agents", name+".md")
		}

		if _, err := os.Stat(path); err != nil {
			t.Errorf("resource %s/%s not found in repo at %s: %v", resourceType, name, path, err)
		}
	}
}

// Test helper: verify resources do NOT exist in repo
func verifyResourcesNotInRepo(t *testing.T, repoPath string, resourceType resource.ResourceType, names ...string) {
	t.Helper()

	for _, name := range names {
		var path string
		switch resourceType {
		case resource.Command:
			path = filepath.Join(repoPath, "commands", name+".md")
		case resource.Skill:
			path = filepath.Join(repoPath, "skills", name)
		case resource.Agent:
			path = filepath.Join(repoPath, "agents", name+".md")
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("resource %s/%s should not exist in repo at %s", resourceType, name, path)
		}
	}
}

// setupTestConfig temporarily replaces the user's config for testing
// Returns a cleanup function that must be called with defer
func setupTestConfig(t *testing.T, cfg *config.Config) (repoPath string, cleanup func()) {
	t.Helper()

	// Get actual config path that LoadGlobal will use
	realConfigPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	// Backup existing config if it exists
	var backupContent []byte
	var hadConfig bool
	if data, err := os.ReadFile(realConfigPath); err == nil {
		backupContent = data
		hadConfig = true
	}

	// Write test config
	if err := os.MkdirAll(filepath.Dir(realConfigPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(realConfigPath, data, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Get the actual repo path that repo.NewManager() will use
	// Since we can't override xdg.DataHome after package init, we need to use the real one
	// and clean it up after the test
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get home dir: %v", err)
		}
	}

	// The real path will be ~/.local/share/ai-config/repo (or XDG_DATA_HOME/ai-config/repo)
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}
	repoPath = filepath.Join(dataHome, "ai-config", "repo")

	// Backup existing repo contents if they exist
	var repoBackup string
	if _, err := os.Stat(repoPath); err == nil {
		repoBackup = repoPath + ".test-backup"
		if err := os.Rename(repoPath, repoBackup); err != nil {
			t.Fatalf("failed to backup repo: %v", err)
		}
	}

	// Create fresh repo directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	cleanup = func() {
		// Clean up test repo
		os.RemoveAll(repoPath)

		// Restore backed up repo if it existed
		if repoBackup != "" {
			os.Rename(repoBackup, repoPath)
		}

		// Restore original config
		if hadConfig {
			os.WriteFile(realConfigPath, backupContent, 0644)
		} else {
			os.Remove(realConfigPath)
		}
	}

	return repoPath, cleanup
}

// TestRunSync_SingleSource tests syncing from a single local source
func TestRunSync_SingleSource(t *testing.T) {
	source1 := createTestSource(t)

	// Create test config with one source
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: source1},
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command directly
	err := runSync(syncCmd, []string{})

	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify resources were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}

// TestRunSync_MultipleSources tests syncing from multiple sources
func TestRunSync_MultipleSources(t *testing.T) {
	source1 := createTestSource(t)
	source2 := createTestSource(t) // Create a second source

	// Create test config with multiple sources
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: source1},
				{URL: source2},
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command directly
	err := runSync(syncCmd, []string{})

	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify resources from both sources were imported (should be deduplicated)
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}

// TestRunSync_WithFilter tests syncing with per-source filters
func TestRunSync_WithFilter(t *testing.T) {
	tests := []struct {
		name               string
		filter             string
		expectedCommands   []string
		expectedSkills     []string
		expectedAgents     []string
		unexpectedCommands []string
		unexpectedSkills   []string
		unexpectedAgents   []string
	}{
		{
			name:             "filter skills only",
			filter:           "skill/*",
			expectedCommands: []string{},
			expectedSkills:   []string{"sync-test-skill", "pdf-processing", "image-processing"},
			expectedAgents:   []string{},
			// These should NOT be imported
			unexpectedCommands: []string{"sync-test-cmd", "test-command"},
			unexpectedAgents:   []string{"sync-test-agent"},
		},
		{
			name:             "filter commands only",
			filter:           "command/*",
			expectedCommands: []string{"sync-test-cmd", "test-command", "pdf-command"},
			expectedSkills:   []string{},
			expectedAgents:   []string{},
			unexpectedSkills: []string{"sync-test-skill"},
			unexpectedAgents: []string{"sync-test-agent"},
		},
		{
			name:               "filter agents only",
			filter:             "agent/*",
			expectedCommands:   []string{},
			expectedSkills:     []string{},
			expectedAgents:     []string{"sync-test-agent", "code-reviewer"},
			unexpectedCommands: []string{"sync-test-cmd"},
			unexpectedSkills:   []string{"sync-test-skill"},
		},
		{
			name:               "filter by name pattern - pdf",
			filter:             "*pdf*",
			expectedCommands:   []string{"pdf-command"},
			expectedSkills:     []string{"pdf-processing"},
			expectedAgents:     []string{},
			unexpectedCommands: []string{"sync-test-cmd", "test-command"},
			unexpectedSkills:   []string{"sync-test-skill", "image-processing"},
			unexpectedAgents:   []string{"sync-test-agent", "code-reviewer"},
		},
		{
			name:               "filter by type and pattern",
			filter:             "skill/pdf*",
			expectedCommands:   []string{},
			expectedSkills:     []string{"pdf-processing"},
			expectedAgents:     []string{},
			unexpectedCommands: []string{"pdf-command"},
			unexpectedSkills:   []string{"sync-test-skill", "image-processing"},
			unexpectedAgents:   []string{"sync-test-agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source1 := createTestSource(t)

			// Create test config with filter
			cfg := &config.Config{
				Sync: config.SyncConfig{
					Sources: []config.SyncSource{
						{URL: source1, Filter: tt.filter},
					},
				},
			}

			repoPath, cleanup := setupTestConfig(t, cfg)
			defer cleanup()

			// Run sync command directly
			err := runSync(syncCmd, []string{})

			if err != nil {
				t.Fatalf("sync command failed: %v", err)
			}

			// Verify expected resources were imported
			if len(tt.expectedCommands) > 0 {
				verifyResourcesInRepo(t, repoPath, resource.Command, tt.expectedCommands...)
			}
			if len(tt.expectedSkills) > 0 {
				verifyResourcesInRepo(t, repoPath, resource.Skill, tt.expectedSkills...)
			}
			if len(tt.expectedAgents) > 0 {
				verifyResourcesInRepo(t, repoPath, resource.Agent, tt.expectedAgents...)
			}

			// Verify unexpected resources were NOT imported
			if len(tt.unexpectedCommands) > 0 {
				verifyResourcesNotInRepo(t, repoPath, resource.Command, tt.unexpectedCommands...)
			}
			if len(tt.unexpectedSkills) > 0 {
				verifyResourcesNotInRepo(t, repoPath, resource.Skill, tt.unexpectedSkills...)
			}
			if len(tt.unexpectedAgents) > 0 {
				verifyResourcesNotInRepo(t, repoPath, resource.Agent, tt.unexpectedAgents...)
			}
		})
	}
}

// TestRunSync_DryRun tests that --dry-run doesn't actually import
func TestRunSync_DryRun(t *testing.T) {
	source1 := createTestSource(t)

	// Create test config
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: source1},
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command with --dry-run
	syncDryRunFlag = true
	defer func() { syncDryRunFlag = false }()
	err := runSync(syncCmd, []string{})

	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify NO resources were actually imported (dry run)
	verifyResourcesNotInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesNotInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesNotInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}

// TestRunSync_SkipExisting tests that --skip-existing doesn't overwrite
func TestRunSync_SkipExisting(t *testing.T) {
	source1 := createTestSource(t)

	// Create test config
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: source1},
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Pre-populate repo with one resource (modified version)
	existingCmdDir := filepath.Join(repoPath, "commands")
	if err := os.MkdirAll(existingCmdDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	existingContent := "---\ndescription: EXISTING VERSION\n---\n# existing"
	existingPath := filepath.Join(existingCmdDir, "sync-test-cmd.md")
	if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing resource: %v", err)
	}

	// Read original content
	originalData, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read existing resource: %v", err)
	}
	originalContent := string(originalData)

	// Run sync command with --skip-existing
	syncSkipExistingFlag = true
	defer func() { syncSkipExistingFlag = false }()
	err = runSync(syncCmd, []string{})

	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify existing resource was NOT overwritten
	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read resource after sync: %v", err)
	}
	newContent := string(data)

	if newContent != originalContent {
		t.Errorf("--skip-existing failed: existing resource was overwritten")
		t.Logf("Original: %s", originalContent)
		t.Logf("After sync: %s", newContent)
	}

	// Verify other resources were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}

// TestRunSync_DefaultForce tests that by default, existing resources are overwritten
func TestRunSync_DefaultForce(t *testing.T) {
	source1 := createTestSource(t)

	// Create test config
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: source1},
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Pre-populate repo with one resource (modified version)
	existingCmdDir := filepath.Join(repoPath, "commands")
	if err := os.MkdirAll(existingCmdDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	existingContent := "---\ndescription: EXISTING VERSION\n---\n# existing"
	existingPath := filepath.Join(existingCmdDir, "sync-test-cmd.md")
	if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing resource: %v", err)
	}

	// Read original content
	originalData, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read existing resource: %v", err)
	}
	originalContent := string(originalData)

	// Run sync command (default force behavior)
	err = runSync(syncCmd, []string{})

	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify existing resource WAS overwritten (force is default)
	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read resource after sync: %v", err)
	}
	newContent := string(data)

	if newContent == originalContent {
		t.Errorf("default force failed: existing resource was not overwritten")
		t.Logf("Content unchanged: %s", newContent)
	}

	// Verify it contains the new content from source
	if !contains(newContent, "Sync test command") {
		t.Errorf("resource content doesn't match source after force update")
		t.Logf("Content: %s", newContent)
	}

	// Verify all resources were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}

// TestRunSync_NoSources tests error when no sources configured
func TestRunSync_NoSources(t *testing.T) {
	// Create test config with NO sources
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{},
		},
	}

	_, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command directly
	err := runSync(syncCmd, []string{})

	// Should return error
	if err == nil {
		t.Fatal("expected error when no sources configured, got nil")
	}

	// Check error message
	expectedMsg := "no sync sources configured"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("error message doesn't contain expected text\nGot: %s\nWant substring: %s", err.Error(), expectedMsg)
	}
}

// TestRunSync_InvalidSource tests error handling for invalid sources
func TestRunSync_InvalidSource(t *testing.T) {
	// Create test config with invalid source (non-existent path)
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: "/nonexistent/path/that/does/not/exist"},
			},
		},
	}

	_, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command directly
	err := runSync(syncCmd, []string{})

	// Should return error since all sources failed
	if err == nil {
		t.Fatal("expected error for invalid source, got nil")
	}

	// Check error message indicates failure
	if !contains(err.Error(), "all sources failed") {
		t.Errorf("error message doesn't indicate all sources failed\nGot: %s", err.Error())
	}
}

// TestRunSync_MixedValidInvalidSources tests partial success with mixed sources
func TestRunSync_MixedValidInvalidSources(t *testing.T) {
	validSource := createTestSource(t)

	// Create test config with one valid and one invalid source
	cfg := &config.Config{
		Sync: config.SyncConfig{
			Sources: []config.SyncSource{
				{URL: validSource},         // Valid
				{URL: "/nonexistent/path"}, // Invalid
			},
		},
	}

	repoPath, cleanup := setupTestConfig(t, cfg)
	defer cleanup()

	// Run sync command directly
	err := runSync(syncCmd, []string{})

	// Should succeed since at least one source worked
	if err != nil {
		t.Fatalf("sync command should succeed with partial success, got: %v", err)
	}

	// Verify resources from valid source were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")
}
