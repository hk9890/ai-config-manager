package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestFindOrphanedMetadata(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo structure
	metadataDir := filepath.Join(tempDir, ".metadata")
	commandsDir := filepath.Join(metadataDir, "commands")
	skillsDir := filepath.Join(metadataDir, "skills")
	agentsDir := filepath.Join(metadataDir, "agents")

	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid source file
	validSourcePath := filepath.Join(tempDir, "valid-command.md")
	if err := os.WriteFile(validSourcePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create metadata for command with valid source (should NOT be orphaned)
	validMeta := &metadata.ResourceMetadata{
		Name:           "valid-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      validSourcePath,
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(validMeta, tempDir); err != nil {
		t.Fatal(err)
	}

	// Create metadata for command with non-existent source (should be orphaned)
	orphanedMeta := &metadata.ResourceMetadata{
		Name:           "orphaned-command",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      filepath.Join(tempDir, "non-existent.md"),
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(orphanedMeta, tempDir); err != nil {
		t.Fatal(err)
	}

	// Create metadata for git source (should NOT be checked)
	gitMeta := &metadata.ResourceMetadata{
		Name:           "git-command",
		Type:           resource.Command,
		SourceType:     "github",
		SourceURL:      "github.com/user/repo/commands/test.md",
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(gitMeta, tempDir); err != nil {
		t.Fatal(err)
	}

	// Create metadata with file:// URL scheme
	orphanedFileURL := &metadata.ResourceMetadata{
		Name:           "orphaned-file-url",
		Type:           resource.Skill,
		SourceType:     "file",
		SourceURL:      "file://" + filepath.Join(tempDir, "missing-skill"),
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(orphanedFileURL, tempDir); err != nil {
		t.Fatal(err)
	}

	// Create manager with temp repo path
	manager := repo.NewManagerWithPath(tempDir)

	// Find orphaned metadata
	orphaned, err := findOrphanedMetadata(manager)
	if err != nil {
		t.Fatalf("findOrphanedMetadata failed: %v", err)
	}

	// Verify results
	if len(orphaned) != 2 {
		t.Errorf("Expected 2 orphaned entries, got %d", len(orphaned))
		for _, o := range orphaned {
			t.Logf("Found orphaned: %s %s", o.Type, o.Name)
		}
	}

	// Check that specific entries are in the orphaned list
	foundOrphanedCommand := false
	foundOrphanedFileURL := false
	for _, o := range orphaned {
		if o.Name == "orphaned-command" && o.Type == resource.Command {
			foundOrphanedCommand = true
		}
		if o.Name == "orphaned-file-url" && o.Type == resource.Skill {
			foundOrphanedFileURL = true
		}
		// Ensure valid-command and git-command are NOT in orphaned list
		if o.Name == "valid-command" || o.Name == "git-command" {
			t.Errorf("Non-orphaned entry %s should not be in orphaned list", o.Name)
		}
	}

	if !foundOrphanedCommand {
		t.Error("orphaned-command should be in orphaned list")
	}
	if !foundOrphanedFileURL {
		t.Error("orphaned-file-url should be in orphaned list")
	}
}

func TestIsOrphaned(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	// Create a valid file
	validPath := filepath.Join(tempDir, "exists.md")
	if err := os.WriteFile(validPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		meta     *metadata.ResourceMetadata
		expected bool
	}{
		{
			name: "local source exists",
			meta: &metadata.ResourceMetadata{
				SourceType: "local",
				SourceURL:  validPath,
			},
			expected: false,
		},
		{
			name: "local source missing",
			meta: &metadata.ResourceMetadata{
				SourceType: "local",
				SourceURL:  filepath.Join(tempDir, "missing.md"),
			},
			expected: true,
		},
		{
			name: "file URL exists",
			meta: &metadata.ResourceMetadata{
				SourceType: "file",
				SourceURL:  "file://" + validPath,
			},
			expected: false,
		},
		{
			name: "file URL missing",
			meta: &metadata.ResourceMetadata{
				SourceType: "file",
				SourceURL:  "file://" + filepath.Join(tempDir, "missing.md"),
			},
			expected: true,
		},
		{
			name: "github source (not checked)",
			meta: &metadata.ResourceMetadata{
				SourceType: "github",
				SourceURL:  "github.com/user/repo",
			},
			expected: false,
		},
		{
			name: "git-url source (not checked)",
			meta: &metadata.ResourceMetadata{
				SourceType: "git-url",
				SourceURL:  "https://example.com/repo.git",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOrphaned(tt.meta)
			if result != tt.expected {
				t.Errorf("isOrphaned() = %v, expected %v for %s", result, tt.expected, tt.name)
			}
		})
	}
}

func TestRemoveOrphanedMetadata(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create test metadata files
	file1 := filepath.Join(tempDir, "orphan1.json")
	file2 := filepath.Join(tempDir, "orphan2.json")

	if err := os.WriteFile(file1, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	orphaned := []OrphanedMetadata{
		{
			Name:     "orphan1",
			Type:     resource.Command,
			FilePath: file1,
		},
		{
			Name:     "orphan2",
			Type:     resource.Skill,
			FilePath: file2,
		},
	}

	// Remove orphaned metadata
	removed, failed := removeOrphanedMetadata(orphaned)

	// Check counts
	if removed != 2 {
		t.Errorf("Expected 2 removed, got %d", removed)
	}
	if failed != 0 {
		t.Errorf("Expected 0 failed, got %d", failed)
	}

	// Verify files are deleted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("file1 should be deleted")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("file2 should be deleted")
	}
}

func TestRemoveOrphanedMetadata_PartialFailure(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create one valid file
	validFile := filepath.Join(tempDir, "valid.json")
	if err := os.WriteFile(validFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	orphaned := []OrphanedMetadata{
		{
			Name:     "valid",
			Type:     resource.Command,
			FilePath: validFile,
		},
		{
			Name:     "non-existent",
			Type:     resource.Skill,
			FilePath: filepath.Join(tempDir, "non-existent.json"),
		},
	}

	// Remove orphaned metadata
	removed, failed := removeOrphanedMetadata(orphaned)

	// Check counts - one should succeed, one should fail
	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}
	if failed != 1 {
		t.Errorf("Expected 1 failed, got %d", failed)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		singular string
		plural   string
		count    int
		expected string
	}{
		{"entry", "entries", 0, "entries"},
		{"entry", "entries", 1, "entry"},
		{"entry", "entries", 2, "entries"},
		{"entry", "entries", 100, "entries"},
		{"file", "files", 1, "file"},
		{"file", "files", 5, "files"},
	}

	for _, tt := range tests {
		result := pluralize(tt.singular, tt.plural, tt.count)
		if result != tt.expected {
			t.Errorf("pluralize(%q, %q, %d) = %q, expected %q",
				tt.singular, tt.plural, tt.count, result, tt.expected)
		}
	}
}
