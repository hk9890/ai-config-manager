package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/repomanifest"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestRepoRemove_ByName(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	source := &repomanifest.Source{
		Name: "test-source",
		Path: "/home/user/resources",
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Add a test resource with metadata pointing to this source
	commandPath := filepath.Join(tempDir, "commands", "test-command.md")
	commandContent := `---
name: test-command
description: Test command
---
# Test Command`
	if err := os.WriteFile(commandPath, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Create metadata for the resource
	meta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources",
		SourceName:     "test-source",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta, tempDir, "test-source"); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Verify source exists
	manifest, _ = repomanifest.Load(tempDir)
	if !manifest.HasSource("test-source") {
		t.Fatal("Source should exist before removal")
	}

	// Verify resource exists
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Fatal("Resource should exist before removal")
	}

	// Remove the source by name
	if err := performRemove(mgr, "test-source", false, false); err != nil {
		t.Fatalf("Failed to remove source: %v", err)
	}

	// Verify source is removed from manifest
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource("test-source") {
		t.Error("Source should not exist after removal")
	}

	// Verify orphaned resource is removed
	if _, err := os.Stat(commandPath); !os.IsNotExist(err) {
		t.Error("Orphaned resource should be removed")
	}

	// Verify metadata is removed
	metadataPath := metadata.GetMetadataPath("test-command", resource.Command, tempDir)
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		t.Error("Resource metadata should be removed")
	}
}

func TestRepoRemove_ByPath(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	sourcePath := "/home/user/my-resources"
	source := &repomanifest.Source{
		Name: "my-resources",
		Path: sourcePath,
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Remove the source by path
	if err := performRemove(mgr, sourcePath, false, false); err != nil {
		t.Fatalf("Failed to remove source: %v", err)
	}

	// Verify source is removed
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource(sourcePath) {
		t.Error("Source should not exist after removal")
	}
}

func TestRepoRemove_ByURL(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	sourceURL := "https://github.com/owner/repo"
	source := &repomanifest.Source{
		Name: "owner-repo",
		URL:  sourceURL,
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Remove the source by URL
	if err := performRemove(mgr, sourceURL, false, false); err != nil {
		t.Fatalf("Failed to remove source: %v", err)
	}

	// Verify source is removed
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource(sourceURL) {
		t.Error("Source should not exist after removal")
	}
}

func TestRepoRemove_DryRun(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	source := &repomanifest.Source{
		Name: "test-source",
		Path: "/home/user/resources",
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Add a test resource
	commandPath := filepath.Join(tempDir, "commands", "test-command.md")
	commandContent := `---
name: test-command
description: Test command
---
# Test Command`
	if err := os.WriteFile(commandPath, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Create metadata for the resource
	meta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources",
		SourceName:     "test-source",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta, tempDir, "test-source"); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Run with dry-run flag
	if err := performRemove(mgr, "test-source", true, false); err != nil {
		t.Fatalf("Failed to run dry-run: %v", err)
	}

	// Verify source still exists (not actually removed)
	manifest, _ = repomanifest.Load(tempDir)
	if !manifest.HasSource("test-source") {
		t.Error("Source should still exist after dry-run")
	}

	// Verify resource still exists (not actually removed)
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Error("Resource should still exist after dry-run")
	}
}

func TestRepoRemove_KeepResources(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	source := &repomanifest.Source{
		Name: "test-source",
		Path: "/home/user/resources",
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Add a test resource
	commandPath := filepath.Join(tempDir, "commands", "test-command.md")
	commandContent := `---
name: test-command
description: Test command
---
# Test Command`
	if err := os.WriteFile(commandPath, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Create metadata for the resource
	meta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources",
		SourceName:     "test-source",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta, tempDir, "test-source"); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Run with keep-resources flag
	if err := performRemove(mgr, "test-source", false, true); err != nil {
		t.Fatalf("Failed to remove source with keep-resources: %v", err)
	}

	// Verify source is removed from manifest
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource("test-source") {
		t.Error("Source should not exist after removal")
	}

	// Verify resource still exists (kept with --keep-resources)
	if _, err := os.Stat(commandPath); os.IsNotExist(err) {
		t.Error("Resource should still exist with --keep-resources")
	}
}

func TestRepoRemove_NotFound(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Try to remove non-existent source
	err := performRemove(mgr, "nonexistent-source", false, false)
	if err == nil {
		t.Fatal("Expected error when removing non-existent source")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error message, got: %v", err)
	}
}

func TestRepoRemove_MultipleResources(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add a source to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	source := &repomanifest.Source{
		Name: "test-source",
		Path: "/home/user/resources",
	}

	if err := manifest.AddSource(source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Add multiple test resources
	commandPath := filepath.Join(tempDir, "commands", "test-command.md")
	commandContent := `---
name: test-command
description: Test command
---
# Test Command`
	if err := os.WriteFile(commandPath, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to create test command: %v", err)
	}

	// Create skill directory and SKILL.md
	skillPath := filepath.Join(tempDir, "skills", "test-skill")
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	skillFile := filepath.Join(skillPath, "SKILL.md")
	skillContent := `---
name: test-skill
description: Test skill
---
# Test Skill`
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Create metadata for command
	commandMeta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources",
		SourceName:     "test-source",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(commandMeta, tempDir, "test-source"); err != nil {
		t.Fatalf("Failed to save command metadata: %v", err)
	}

	// Create metadata for skill
	skillMeta := &metadata.ResourceMetadata{
		Name:           "test-skill",
		Type:           resource.Skill,
		SourceType:     "local",
		SourceURL:      "/home/user/resources",
		SourceName:     "test-source",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(skillMeta, tempDir, "test-source"); err != nil {
		t.Fatalf("Failed to save skill metadata: %v", err)
	}

	// Remove the source
	if err := performRemove(mgr, "test-source", false, false); err != nil {
		t.Fatalf("Failed to remove source: %v", err)
	}

	// Verify source is removed
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource("test-source") {
		t.Error("Source should not exist after removal")
	}

	// Verify all orphaned resources are removed
	if _, err := os.Stat(commandPath); !os.IsNotExist(err) {
		t.Error("Command should be removed")
	}
	if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
		t.Error("Skill should be removed")
	}
}

func TestRepoRemove_OnlyOrphansFromSource(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	mgr := repo.NewManagerWithPath(tempDir)

	// Initialize repo
	if err := mgr.Init(); err != nil {
		t.Fatalf("Failed to initialize repo: %v", err)
	}

	// Add two sources to the manifest
	manifest, err := repomanifest.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	source1 := &repomanifest.Source{
		Name: "source-1",
		Path: "/home/user/resources1",
	}
	source2 := &repomanifest.Source{
		Name: "source-2",
		Path: "/home/user/resources2",
	}

	if err := manifest.AddSource(source1); err != nil {
		t.Fatalf("Failed to add source 1: %v", err)
	}
	if err := manifest.AddSource(source2); err != nil {
		t.Fatalf("Failed to add source 2: %v", err)
	}

	if err := manifest.Save(tempDir); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Add resources from both sources
	command1Path := filepath.Join(tempDir, "commands", "command1.md")
	command1Content := `---
name: command1
description: Command from source 1
---
# Command 1`
	if err := os.WriteFile(command1Path, []byte(command1Content), 0644); err != nil {
		t.Fatalf("Failed to create command1: %v", err)
	}

	command2Path := filepath.Join(tempDir, "commands", "command2.md")
	command2Content := `---
name: command2
description: Command from source 2
---
# Command 2`
	if err := os.WriteFile(command2Path, []byte(command2Content), 0644); err != nil {
		t.Fatalf("Failed to create command2: %v", err)
	}

	// Create metadata linking to different sources
	meta1 := &metadata.ResourceMetadata{
		Name:           "command1",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources1",
		SourceName:     "source-1",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta1, tempDir, "source-1"); err != nil {
		t.Fatalf("Failed to save metadata for command1: %v", err)
	}

	meta2 := &metadata.ResourceMetadata{
		Name:           "command2",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "/home/user/resources2",
		SourceName:     "source-2",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta2, tempDir, "source-2"); err != nil {
		t.Fatalf("Failed to save metadata for command2: %v", err)
	}

	// Remove source-1
	if err := performRemove(mgr, "source-1", false, false); err != nil {
		t.Fatalf("Failed to remove source: %v", err)
	}

	// Verify source-1 is removed but source-2 still exists
	manifest, _ = repomanifest.Load(tempDir)
	if manifest.HasSource("source-1") {
		t.Error("source-1 should not exist after removal")
	}
	if !manifest.HasSource("source-2") {
		t.Error("source-2 should still exist")
	}

	// Verify only command1 is removed (from source-1), command2 should remain
	if _, err := os.Stat(command1Path); !os.IsNotExist(err) {
		t.Error("command1 should be removed")
	}
	if _, err := os.Stat(command2Path); os.IsNotExist(err) {
		t.Error("command2 should still exist (from source-2)")
	}
}
