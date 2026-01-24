package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestMigrateMetadataFiles_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create old-style metadata files for all resource types
	commandMeta := &ResourceMetadata{
		Name:           "test-cmd",
		Type:           resource.Command,
		SourceType:     "github",
		SourceURL:      "gh:owner/repo/commands/test-cmd.md",
		FirstInstalled: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	skillMeta := &ResourceMetadata{
		Name:           "pdf-processor",
		Type:           resource.Skill,
		SourceType:     "local",
		SourceURL:      "file:///home/user/skills/pdf-processor",
		FirstInstalled: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
	}

	agentMeta := &ResourceMetadata{
		Name:           "code-reviewer",
		Type:           resource.Agent,
		SourceType:     "github",
		SourceURL:      "gh:owner/repo/agents/code-reviewer.md",
		FirstInstalled: time.Date(2024, 3, 1, 14, 30, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 3, 2, 9, 15, 0, 0, time.UTC),
	}

	// Create old-style metadata files
	createOldMetadataFile(t, tmpDir, commandMeta)
	createOldMetadataFile(t, tmpDir, skillMeta)
	createOldMetadataFile(t, tmpDir, agentMeta)

	// Run migration
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Verify result counts
	if result.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", result.TotalFiles)
	}
	if result.MovedFiles != 3 {
		t.Errorf("MovedFiles = %d, want 3", result.MovedFiles)
	}
	if result.SkippedFiles != 0 {
		t.Errorf("SkippedFiles = %d, want 0", result.SkippedFiles)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}

	// Verify new files exist
	newCommandPath := GetMetadataPath("test-cmd", resource.Command, tmpDir)
	newSkillPath := GetMetadataPath("pdf-processor", resource.Skill, tmpDir)
	newAgentPath := GetMetadataPath("code-reviewer", resource.Agent, tmpDir)

	if _, err := os.Stat(newCommandPath); os.IsNotExist(err) {
		t.Errorf("New command metadata file not found at %s", newCommandPath)
	}
	if _, err := os.Stat(newSkillPath); os.IsNotExist(err) {
		t.Errorf("New skill metadata file not found at %s", newSkillPath)
	}
	if _, err := os.Stat(newAgentPath); os.IsNotExist(err) {
		t.Errorf("New agent metadata file not found at %s", newAgentPath)
	}

	// Verify old files are removed
	oldCommandPath := filepath.Join(tmpDir, "commands", "command-test-cmd-metadata.json")
	oldSkillPath := filepath.Join(tmpDir, "skills", "skill-pdf-processor-metadata.json")
	oldAgentPath := filepath.Join(tmpDir, "agents", "agent-code-reviewer-metadata.json")

	if _, err := os.Stat(oldCommandPath); !os.IsNotExist(err) {
		t.Errorf("Old command metadata file still exists at %s", oldCommandPath)
	}
	if _, err := os.Stat(oldSkillPath); !os.IsNotExist(err) {
		t.Errorf("Old skill metadata file still exists at %s", oldSkillPath)
	}
	if _, err := os.Stat(oldAgentPath); !os.IsNotExist(err) {
		t.Errorf("Old agent metadata file still exists at %s", oldAgentPath)
	}

	// Verify file contents are preserved
	verifyMetadataContent(t, newCommandPath, commandMeta)
	verifyMetadataContent(t, newSkillPath, skillMeta)
	verifyMetadataContent(t, newAgentPath, agentMeta)
}

func TestMigrateMetadataFiles_EmptyRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty resource directories
	os.MkdirAll(filepath.Join(tmpDir, "commands"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "skills"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "agents"), 0755)

	// Run migration
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Verify result counts - no files to migrate
	if result.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", result.TotalFiles)
	}
	if result.MovedFiles != 0 {
		t.Errorf("MovedFiles = %d, want 0", result.MovedFiles)
	}
	if result.SkippedFiles != 0 {
		t.Errorf("SkippedFiles = %d, want 0", result.SkippedFiles)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}
}

func TestMigrateMetadataFiles_MixedResources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create metadata only for commands and agents, not skills
	commandMeta := &ResourceMetadata{
		Name:           "my-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "file:///test/my-command.md",
		FirstInstalled: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	agentMeta := &ResourceMetadata{
		Name:           "test-agent",
		Type:           resource.Agent,
		SourceType:     "github",
		SourceURL:      "gh:owner/repo/agents/test-agent.md",
		FirstInstalled: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
	}

	createOldMetadataFile(t, tmpDir, commandMeta)
	createOldMetadataFile(t, tmpDir, agentMeta)

	// Create empty skills directory (no metadata files)
	os.MkdirAll(filepath.Join(tmpDir, "skills"), 0755)

	// Run migration
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Verify result counts
	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}
	if result.MovedFiles != 2 {
		t.Errorf("MovedFiles = %d, want 2", result.MovedFiles)
	}
	if result.SkippedFiles != 0 {
		t.Errorf("SkippedFiles = %d, want 0", result.SkippedFiles)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}

	// Verify new files exist
	newCommandPath := GetMetadataPath("my-command", resource.Command, tmpDir)
	newAgentPath := GetMetadataPath("test-agent", resource.Agent, tmpDir)

	if _, err := os.Stat(newCommandPath); os.IsNotExist(err) {
		t.Errorf("New command metadata file not found at %s", newCommandPath)
	}
	if _, err := os.Stat(newAgentPath); os.IsNotExist(err) {
		t.Errorf("New agent metadata file not found at %s", newAgentPath)
	}
}

func TestMigrateMetadataFiles_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory
	commandsDir := filepath.Join(tmpDir, "commands")
	os.MkdirAll(commandsDir, 0755)

	// Create valid metadata file
	validMeta := &ResourceMetadata{
		Name:           "valid-cmd",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "file:///test",
		FirstInstalled: time.Now().UTC(),
		LastUpdated:    time.Now().UTC(),
	}
	createOldMetadataFile(t, tmpDir, validMeta)

	// Create invalid JSON file (migration doesn't validate JSON, just moves files)
	// This file will be migrated just like any other metadata file
	invalidPath := filepath.Join(commandsDir, "command-invalid-metadata.json")
	invalidJSON := []byte(`{"name": "invalid", "type": "command", invalid json}`)
	if err := os.WriteFile(invalidPath, invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Run migration - moves files without validating JSON content
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Should count and move both files (migration doesn't parse JSON)
	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}
	if result.MovedFiles != 2 {
		t.Errorf("MovedFiles = %d, want 2 (both files migrated)", result.MovedFiles)
	}

	// Valid file should be migrated
	newValidPath := GetMetadataPath("valid-cmd", resource.Command, tmpDir)
	if _, err := os.Stat(newValidPath); os.IsNotExist(err) {
		t.Errorf("Valid metadata file should be migrated to %s", newValidPath)
	}

	// Invalid file should also be migrated (migration doesn't validate JSON)
	newInvalidPath := GetMetadataPath("invalid", resource.Command, tmpDir)
	if _, err := os.Stat(newInvalidPath); os.IsNotExist(err) {
		t.Errorf("Invalid metadata file should be migrated to %s", newInvalidPath)
	}

	// Old files should be removed
	if _, err := os.Stat(invalidPath); !os.IsNotExist(err) {
		t.Errorf("Old invalid metadata file should be removed from %s", invalidPath)
	}
}

func TestMigrateMetadataFiles_AlreadyMigrated(t *testing.T) {
	tmpDir := t.TempDir()

	commandMeta := &ResourceMetadata{
		Name:           "already-migrated",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "file:///test",
		FirstInstalled: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Create old-style metadata file
	createOldMetadataFile(t, tmpDir, commandMeta)

	// Also create new-style metadata file (simulating already migrated)
	newPath := GetMetadataPath("already-migrated", resource.Command, tmpDir)
	os.MkdirAll(filepath.Dir(newPath), 0755)
	data, _ := json.MarshalIndent(commandMeta, "", "  ")
	os.WriteFile(newPath, data, 0644)

	// Run migration
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Should skip the file since it already exists at new location
	if result.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want 1", result.TotalFiles)
	}
	if result.MovedFiles != 0 {
		t.Errorf("MovedFiles = %d, want 0", result.MovedFiles)
	}
	if result.SkippedFiles != 1 {
		t.Errorf("SkippedFiles = %d, want 1", result.SkippedFiles)
	}

	// Old file should still exist (not removed when skipped)
	oldPath := filepath.Join(tmpDir, "commands", "command-already-migrated-metadata.json")
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		t.Errorf("Old metadata file should still exist when skipped at %s", oldPath)
	}
}

func TestMigrateMetadataFiles_EmptyRepoPath(t *testing.T) {
	result, err := MigrateMetadataFiles("")
	if err == nil {
		t.Error("MigrateMetadataFiles(\"\") expected error, got nil")
	}
	if result != nil {
		t.Errorf("MigrateMetadataFiles(\"\") result = %v, want nil", result)
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Error message = %v, want 'cannot be empty'", err.Error())
	}
}

func TestMigrateMetadataFiles_NonExistentRepo(t *testing.T) {
	nonExistentPath := "/this/path/does/not/exist"
	result, err := MigrateMetadataFiles(nonExistentPath)
	if err == nil {
		t.Error("MigrateMetadataFiles() expected error for non-existent path, got nil")
	}
	if result != nil {
		t.Errorf("MigrateMetadataFiles() result = %v, want nil", result)
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Error message = %v, want 'does not exist'", err.Error())
	}
}

func TestMigrateMetadataFiles_PermissionError(t *testing.T) {
	// Skip on systems where we can't properly test permissions
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()

	// Create metadata file
	commandMeta := &ResourceMetadata{
		Name:           "perm-test",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "file:///test",
		FirstInstalled: time.Now().UTC(),
		LastUpdated:    time.Now().UTC(),
	}
	createOldMetadataFile(t, tmpDir, commandMeta)

	// Make the metadata directory read-only
	metadataDir := filepath.Join(tmpDir, ".metadata")
	os.MkdirAll(metadataDir, 0755)
	if err := os.Chmod(metadataDir, 0444); err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}
	defer os.Chmod(metadataDir, 0755) // Restore permissions for cleanup

	// Run migration - should fail to create new directory
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Should have errors due to permission issues
	if result.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want 1", result.TotalFiles)
	}
	if result.MovedFiles != 0 {
		t.Errorf("MovedFiles = %d, want 0 (permission denied)", result.MovedFiles)
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors due to permission issues, got none")
	}
}

func TestMigrateMetadataFiles_MultipleFilesPerType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple metadata files for each type
	commands := []string{"cmd1", "cmd2", "cmd3"}
	skills := []string{"skill1", "skill2"}
	agents := []string{"agent1"}

	totalFiles := len(commands) + len(skills) + len(agents)

	for _, name := range commands {
		meta := &ResourceMetadata{
			Name:           name,
			Type:           resource.Command,
			SourceType:     "local",
			SourceURL:      "file:///test/" + name,
			FirstInstalled: time.Now().UTC(),
			LastUpdated:    time.Now().UTC(),
		}
		createOldMetadataFile(t, tmpDir, meta)
	}

	for _, name := range skills {
		meta := &ResourceMetadata{
			Name:           name,
			Type:           resource.Skill,
			SourceType:     "local",
			SourceURL:      "file:///test/" + name,
			FirstInstalled: time.Now().UTC(),
			LastUpdated:    time.Now().UTC(),
		}
		createOldMetadataFile(t, tmpDir, meta)
	}

	for _, name := range agents {
		meta := &ResourceMetadata{
			Name:           name,
			Type:           resource.Agent,
			SourceType:     "local",
			SourceURL:      "file:///test/" + name,
			FirstInstalled: time.Now().UTC(),
			LastUpdated:    time.Now().UTC(),
		}
		createOldMetadataFile(t, tmpDir, meta)
	}

	// Run migration
	result, err := MigrateMetadataFiles(tmpDir)
	if err != nil {
		t.Fatalf("MigrateMetadataFiles() error = %v", err)
	}

	// Verify all files were migrated
	if result.TotalFiles != totalFiles {
		t.Errorf("TotalFiles = %d, want %d", result.TotalFiles, totalFiles)
	}
	if result.MovedFiles != totalFiles {
		t.Errorf("MovedFiles = %d, want %d", result.MovedFiles, totalFiles)
	}
	if result.SkippedFiles != 0 {
		t.Errorf("SkippedFiles = %d, want 0", result.SkippedFiles)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", result.Errors)
	}

	// Verify all new files exist
	for _, name := range commands {
		newPath := GetMetadataPath(name, resource.Command, tmpDir)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Errorf("New command metadata file not found: %s", newPath)
		}
	}
	for _, name := range skills {
		newPath := GetMetadataPath(name, resource.Skill, tmpDir)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Errorf("New skill metadata file not found: %s", newPath)
		}
	}
	for _, name := range agents {
		newPath := GetMetadataPath(name, resource.Agent, tmpDir)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			t.Errorf("New agent metadata file not found: %s", newPath)
		}
	}
}

func TestParseResourceName(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		resourceType resource.ResourceType
		wantName     string
		wantError    bool
	}{
		{
			name:         "command with simple name",
			filename:     "command-test-metadata.json",
			resourceType: resource.Command,
			wantName:     "test",
			wantError:    false,
		},
		{
			name:         "skill with hyphenated name",
			filename:     "skill-pdf-processor-metadata.json",
			resourceType: resource.Skill,
			wantName:     "pdf-processor",
			wantError:    false,
		},
		{
			name:         "agent with multi-hyphen name",
			filename:     "agent-code-review-bot-metadata.json",
			resourceType: resource.Agent,
			wantName:     "code-review-bot",
			wantError:    false,
		},
		{
			name:         "missing type prefix",
			filename:     "test-metadata.json",
			resourceType: resource.Command,
			wantName:     "",
			wantError:    true,
		},
		{
			name:         "missing metadata suffix",
			filename:     "command-test.json",
			resourceType: resource.Command,
			wantName:     "",
			wantError:    true,
		},
		{
			name:         "empty name",
			filename:     "command--metadata.json",
			resourceType: resource.Command,
			wantName:     "",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, err := parseResourceName(tt.filename, tt.resourceType)
			if (err != nil) != tt.wantError {
				t.Errorf("parseResourceName() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("parseResourceName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestIsOldMetadataFile(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		resourceType resource.ResourceType
		want         bool
	}{
		{
			name:         "valid command metadata",
			filename:     "command-test-metadata.json",
			resourceType: resource.Command,
			want:         true,
		},
		{
			name:         "valid skill metadata",
			filename:     "skill-pdf-processor-metadata.json",
			resourceType: resource.Skill,
			want:         true,
		},
		{
			name:         "valid agent metadata",
			filename:     "agent-code-reviewer-metadata.json",
			resourceType: resource.Agent,
			want:         true,
		},
		{
			name:         "resource file (not metadata)",
			filename:     "command-test.md",
			resourceType: resource.Command,
			want:         false,
		},
		{
			name:         "missing type prefix",
			filename:     "test-metadata.json",
			resourceType: resource.Command,
			want:         false,
		},
		{
			name:         "wrong type prefix",
			filename:     "skill-test-metadata.json",
			resourceType: resource.Command,
			want:         false,
		},
		{
			name:         "missing metadata suffix",
			filename:     "command-test.json",
			resourceType: resource.Command,
			want:         false,
		},
		{
			name:         "new style metadata (no type prefix)",
			filename:     "test-metadata.json",
			resourceType: resource.Command,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOldMetadataFile(tt.filename, tt.resourceType)
			if got != tt.want {
				t.Errorf("isOldMetadataFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNewMetadataPath(t *testing.T) {
	tests := []struct {
		name         string
		repoPath     string
		resourceName string
		resourceType resource.ResourceType
		wantPath     string
	}{
		{
			name:         "command metadata path",
			repoPath:     "/home/user/repo",
			resourceName: "test-cmd",
			resourceType: resource.Command,
			wantPath:     "/home/user/repo/.metadata/commands/test-cmd-metadata.json",
		},
		{
			name:         "skill metadata path",
			repoPath:     "/home/user/repo",
			resourceName: "pdf-processor",
			resourceType: resource.Skill,
			wantPath:     "/home/user/repo/.metadata/skills/pdf-processor-metadata.json",
		},
		{
			name:         "agent metadata path",
			repoPath:     "/home/user/repo",
			resourceName: "code-reviewer",
			resourceType: resource.Agent,
			wantPath:     "/home/user/repo/.metadata/agents/code-reviewer-metadata.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath := getNewMetadataPath(tt.repoPath, tt.resourceName, tt.resourceType)
			if gotPath != tt.wantPath {
				t.Errorf("getNewMetadataPath() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

// Helper functions

// createOldMetadataFile creates a metadata file in the old location with old naming pattern
func createOldMetadataFile(t *testing.T, repoPath string, meta *ResourceMetadata) {
	t.Helper()

	// Old pattern: /repo/<type>s/<type>-<name>-metadata.json
	dir := filepath.Join(repoPath, string(meta.Type)+"s")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	filename := fmt.Sprintf("%s-%s-metadata.json", meta.Type, meta.Name)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}
}

// verifyMetadataContent verifies that a metadata file contains the expected data
func verifyMetadataContent(t *testing.T, path string, expected *ResourceMetadata) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read metadata file %s: %v", path, err)
	}

	var actual ResourceMetadata
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if actual.Name != expected.Name {
		t.Errorf("Name = %v, want %v", actual.Name, expected.Name)
	}
	if actual.Type != expected.Type {
		t.Errorf("Type = %v, want %v", actual.Type, expected.Type)
	}
	if actual.SourceType != expected.SourceType {
		t.Errorf("SourceType = %v, want %v", actual.SourceType, expected.SourceType)
	}
	if actual.SourceURL != expected.SourceURL {
		t.Errorf("SourceURL = %v, want %v", actual.SourceURL, expected.SourceURL)
	}
	if !actual.FirstInstalled.Equal(expected.FirstInstalled) {
		t.Errorf("FirstInstalled = %v, want %v", actual.FirstInstalled, expected.FirstInstalled)
	}
	if !actual.LastUpdated.Equal(expected.LastUpdated) {
		t.Errorf("LastUpdated = %v, want %v", actual.LastUpdated, expected.LastUpdated)
	}
}
