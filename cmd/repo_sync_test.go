package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/repomanifest"
	"github.com/hk9890/ai-config-manager/pkg/resource"
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

// setupTestManifest creates a test repository with ai.repo.yaml
// Returns the repo path for verification and a cleanup function
func setupTestManifest(t *testing.T, sources []*repomanifest.Source) (string, func()) {
	t.Helper()

	// Create temp directory for repo
	repoPath := t.TempDir()

	// Create manifest
	manifest := &repomanifest.Manifest{
		Version: 1,
		Sources: sources,
	}

	// Save manifest to repo
	if err := manifest.Save(repoPath); err != nil {
		t.Fatalf("failed to save test manifest: %v", err)
	}

	// Use AIMGR_REPO_PATH to override repo location
	// This is the highest priority override in repo.NewManager()
	originalRepoPath := os.Getenv("AIMGR_REPO_PATH")
	os.Setenv("AIMGR_REPO_PATH", repoPath)

	cleanup := func() {
		if originalRepoPath != "" {
			os.Setenv("AIMGR_REPO_PATH", originalRepoPath)
		} else {
			os.Unsetenv("AIMGR_REPO_PATH")
		}
	}

	return repoPath, cleanup
}

// verifySourceSynced checks that a source was successfully synced
// Verifies that last_synced timestamp was updated
func verifySourceSynced(t *testing.T, repoPath string, sourceName string) {
	t.Helper()

	// Load manifest
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	// Find source
	source, found := manifest.GetSource(sourceName)
	if !found {
		t.Fatalf("source %s not found in manifest", sourceName)
	}

	// Verify last_synced was updated
	if source.LastSynced.IsZero() {
		t.Errorf("source %s last_synced timestamp not updated", sourceName)
	}

	// Verify timestamp is recent (within last minute)
	if time.Since(source.LastSynced) > time.Minute {
		t.Errorf("source %s last_synced timestamp is too old: %v", sourceName, source.LastSynced)
	}
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

// TestRunSync_SingleSource tests syncing from a single local source
func TestRunSync_SingleSource(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest with one source
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
	err := runSync(syncCmd, []string{})
	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify resources were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")

	// Verify source was marked as synced
	verifySourceSynced(t, repoPath, "test-source-1")
}

// TestRunSync_MultipleSources tests syncing from multiple sources
func TestRunSync_MultipleSources(t *testing.T) {
	source1 := createTestSource(t)
	source2 := createTestSource(t)

	// Create manifest with multiple sources
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: time.Now(),
		},
		{
			Name:  "test-source-2",
			Path:  source2,
			Mode:  "copy",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
	err := runSync(syncCmd, []string{})
	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Verify resources from both sources were imported (should be deduplicated)
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")

	// Verify both sources were marked as synced
	verifySourceSynced(t, repoPath, "test-source-1")
	verifySourceSynced(t, repoPath, "test-source-2")
}

// TestRunSync_DryRun tests that --dry-run doesn't actually import
func TestRunSync_DryRun(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
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

	// Verify last_synced was NOT updated (dry run)
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	source, found := manifest.GetSource("test-source-1")
	if !found {
		t.Fatalf("source not found")
	}
	if !source.LastSynced.IsZero() {
		t.Errorf("dry run should not update last_synced timestamp")
	}
}

// TestRunSync_SkipExisting tests that --skip-existing doesn't overwrite
func TestRunSync_SkipExisting(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
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

	// Verify source was marked as synced
	verifySourceSynced(t, repoPath, "test-source-1")
}

// TestRunSync_DefaultForce tests that by default, existing resources are overwritten
func TestRunSync_DefaultForce(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
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

	// Verify source was marked as synced
	verifySourceSynced(t, repoPath, "test-source-1")
}

// TestRunSync_NoSources tests error when no sources configured
func TestRunSync_NoSources(t *testing.T) {
	// Create manifest with NO sources
	sources := []*repomanifest.Source{}
	_, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
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
	// Create manifest with invalid source (non-existent path)
	sources := []*repomanifest.Source{
		{
			Name:  "invalid-source",
			Path:  "/nonexistent/path/that/does/not/exist",
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	_, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
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

	// Create manifest with one valid and one invalid source
	sources := []*repomanifest.Source{
		{
			Name:  "valid-source",
			Path:  validSource,
			Mode:  "symlink",
			Added: time.Now(),
		},
		{
			Name:  "invalid-source",
			Path:  "/nonexistent/path",
			Mode:  "symlink",
			Added: time.Now(),
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
	err := runSync(syncCmd, []string{})

	// Should succeed since at least one source worked
	if err != nil {
		t.Fatalf("sync command should succeed with partial success, got: %v", err)
	}

	// Verify resources from valid source were imported
	verifyResourcesInRepo(t, repoPath, resource.Command, "sync-test-cmd", "test-command", "pdf-command")
	verifyResourcesInRepo(t, repoPath, resource.Skill, "sync-test-skill", "pdf-processing", "image-processing")
	verifyResourcesInRepo(t, repoPath, resource.Agent, "sync-test-agent", "code-reviewer")

	// Verify valid source was marked as synced
	verifySourceSynced(t, repoPath, "valid-source")

	// Verify invalid source was NOT marked as synced
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	invalidSource, found := manifest.GetSource("invalid-source")
	if !found {
		t.Fatalf("invalid-source not found in manifest")
	}
	if !invalidSource.LastSynced.IsZero() {
		t.Errorf("invalid source should not have last_synced timestamp updated")
	}
}

// TestRunSync_UpdatesLastSynced tests that last_synced timestamps are updated
func TestRunSync_UpdatesLastSynced(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest with old last_synced timestamp
	oldTimestamp := time.Now().Add(-24 * time.Hour)
	sources := []*repomanifest.Source{
		{
			Name:       "test-source-1",
			Path:       source1,
			Mode:       "symlink",
			Added:      time.Now().Add(-48 * time.Hour),
			LastSynced: oldTimestamp,
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
	err := runSync(syncCmd, []string{})
	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Load manifest and verify timestamp was updated
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	source, found := manifest.GetSource("test-source-1")
	if !found {
		t.Fatalf("source not found")
	}

	// Verify timestamp was updated (should be recent, not old)
	if source.LastSynced.Equal(oldTimestamp) {
		t.Errorf("last_synced timestamp was not updated")
	}

	if time.Since(source.LastSynced) > time.Minute {
		t.Errorf("last_synced timestamp is not recent: %v", source.LastSynced)
	}
}

// TestRunSync_PreservesManifest tests that manifest structure is preserved
func TestRunSync_PreservesManifest(t *testing.T) {
	source1 := createTestSource(t)

	// Create manifest with specific structure
	addedTime := time.Now().Add(-48 * time.Hour)
	sources := []*repomanifest.Source{
		{
			Name:  "test-source-1",
			Path:  source1,
			Mode:  "symlink",
			Added: addedTime,
		},
	}
	repoPath, cleanup := setupTestManifest(t, sources)
	defer cleanup()

	// Run sync command
	err := runSync(syncCmd, []string{})
	if err != nil {
		t.Fatalf("sync command failed: %v", err)
	}

	// Load manifest and verify structure is preserved
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	// Verify version
	if manifest.Version != 1 {
		t.Errorf("manifest version changed: expected 1, got %d", manifest.Version)
	}

	// Verify source count
	if len(manifest.Sources) != 1 {
		t.Errorf("source count changed: expected 1, got %d", len(manifest.Sources))
	}

	// Verify source fields are preserved
	source := manifest.Sources[0]
	if source.Name != "test-source-1" {
		t.Errorf("source name changed: expected test-source-1, got %s", source.Name)
	}
	if source.Path != source1 {
		t.Errorf("source path changed: expected %s, got %s", source1, source.Path)
	}
	if source.Mode != "symlink" {
		t.Errorf("source mode changed: expected symlink, got %s", source.Mode)
	}

	// Verify added timestamp is preserved (within 1 second tolerance for rounding)
	if source.Added.Sub(addedTime).Abs() > time.Second {
		t.Errorf("added timestamp changed: expected %v, got %v", addedTime, source.Added)
	}

	// Verify last_synced was added
	if source.LastSynced.IsZero() {
		t.Errorf("last_synced timestamp was not added")
	}
}

// TestRunSync_WithFilter is skipped since sync doesn't support filtering
// Filters are per-source at add time, not at sync time
func TestRunSync_WithFilter(t *testing.T) {
	t.Skip("Sync does not support filtering - filters are applied at 'repo add' time, not sync time")
}
