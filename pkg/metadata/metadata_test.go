package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestGetMetadataPath(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		resourceType resource.ResourceType
		repoPath     string
		wantPath     string
	}{
		{
			name:         "command metadata path",
			resourceName: "test-cmd",
			resourceType: resource.Command,
			repoPath:     "/home/user/.local/share/ai-config/repo",
			wantPath:     "/home/user/.local/share/ai-config/repo/.metadata/commands/test-cmd-metadata.json",
		},
		{
			name:         "skill metadata path",
			resourceName: "pdf-processor",
			resourceType: resource.Skill,
			repoPath:     "/home/user/.local/share/ai-config/repo",
			wantPath:     "/home/user/.local/share/ai-config/repo/.metadata/skills/pdf-processor-metadata.json",
		},
		{
			name:         "agent metadata path",
			resourceName: "code-reviewer",
			resourceType: resource.Agent,
			repoPath:     "/home/user/.local/share/ai-config/repo",
			wantPath:     "/home/user/.local/share/ai-config/repo/.metadata/agents/code-reviewer-metadata.json",
		},
		{
			name:         "command with different repo path",
			resourceName: "my-command",
			resourceType: resource.Command,
			repoPath:     "/tmp/test-repo",
			wantPath:     "/tmp/test-repo/.metadata/commands/my-command-metadata.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath := GetMetadataPath(tt.resourceName, tt.resourceType, tt.repoPath)
			if gotPath != tt.wantPath {
				t.Errorf("GetMetadataPath() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestSaveMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		metadata  *ResourceMetadata
		repoPath  string
		wantError bool
	}{
		{
			name: "save command metadata",
			metadata: &ResourceMetadata{
				Name:           "test-command",
				Type:           resource.Command,
				SourceType:     "github",
				SourceURL:      "gh:owner/repo/commands/test-command.md",
				FirstInstalled: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				LastUpdated:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			},
			repoPath:  tmpDir,
			wantError: false,
		},
		{
			name: "save skill metadata",
			metadata: &ResourceMetadata{
				Name:           "pdf-skill",
				Type:           resource.Skill,
				SourceType:     "local",
				SourceURL:      "file:///home/user/skills/pdf-skill",
				FirstInstalled: time.Date(2024, 1, 5, 10, 30, 0, 0, time.UTC),
				LastUpdated:    time.Date(2024, 1, 5, 10, 30, 0, 0, time.UTC),
			},
			repoPath:  tmpDir,
			wantError: false,
		},
		{
			name: "save agent metadata",
			metadata: &ResourceMetadata{
				Name:           "code-agent",
				Type:           resource.Agent,
				SourceType:     "file",
				SourceURL:      "file:///tmp/code-agent.md",
				FirstInstalled: time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
				LastUpdated:    time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
			},
			repoPath:  tmpDir,
			wantError: false,
		},
		{
			name:      "nil metadata",
			metadata:  nil,
			repoPath:  tmpDir,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Save(tt.metadata, tt.repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("Save() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Verify file exists
				metadataPath := GetMetadataPath(tt.metadata.Name, tt.metadata.Type, tt.repoPath)
				if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
					t.Errorf("Save() did not create file at %s", metadataPath)
				}

				// Verify file content
				data, err := os.ReadFile(metadataPath)
				if err != nil {
					t.Fatalf("Failed to read saved metadata file: %v", err)
				}

				var savedMetadata ResourceMetadata
				if err := json.Unmarshal(data, &savedMetadata); err != nil {
					t.Fatalf("Failed to unmarshal saved metadata: %v", err)
				}

				// Verify fields
				if savedMetadata.Name != tt.metadata.Name {
					t.Errorf("Name = %v, want %v", savedMetadata.Name, tt.metadata.Name)
				}
				if savedMetadata.Type != tt.metadata.Type {
					t.Errorf("Type = %v, want %v", savedMetadata.Type, tt.metadata.Type)
				}
				if savedMetadata.SourceType != tt.metadata.SourceType {
					t.Errorf("SourceType = %v, want %v", savedMetadata.SourceType, tt.metadata.SourceType)
				}
				if savedMetadata.SourceURL != tt.metadata.SourceURL {
					t.Errorf("SourceURL = %v, want %v", savedMetadata.SourceURL, tt.metadata.SourceURL)
				}

				// Verify timestamps (using Equal for time.Time comparison)
				if !savedMetadata.FirstInstalled.Equal(tt.metadata.FirstInstalled) {
					t.Errorf("FirstInstalled = %v, want %v", savedMetadata.FirstInstalled, tt.metadata.FirstInstalled)
				}
				if !savedMetadata.LastUpdated.Equal(tt.metadata.LastUpdated) {
					t.Errorf("LastUpdated = %v, want %v", savedMetadata.LastUpdated, tt.metadata.LastUpdated)
				}
			}
		})
	}
}

func TestLoadMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test metadata
	testMetadata := &ResourceMetadata{
		Name:           "test-load",
		Type:           resource.Command,
		SourceType:     "github",
		SourceURL:      "gh:owner/repo/test.md",
		FirstInstalled: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		LastUpdated:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// Save it first
	if err := Save(testMetadata, tmpDir); err != nil {
		t.Fatalf("Failed to save test metadata: %v", err)
	}

	tests := []struct {
		name         string
		resourceName string
		resourceType resource.ResourceType
		repoPath     string
		wantError    bool
		wantMetadata *ResourceMetadata
	}{
		{
			name:         "load existing metadata",
			resourceName: "test-load",
			resourceType: resource.Command,
			repoPath:     tmpDir,
			wantError:    false,
			wantMetadata: testMetadata,
		},
		{
			name:         "load non-existent metadata",
			resourceName: "does-not-exist",
			resourceType: resource.Command,
			repoPath:     tmpDir,
			wantError:    true,
			wantMetadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := Load(tt.resourceName, tt.resourceType, tt.repoPath)
			if (err != nil) != tt.wantError {
				t.Errorf("Load() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if metadata == nil {
					t.Fatal("Load() returned nil metadata")
				}

				// Verify fields
				if metadata.Name != tt.wantMetadata.Name {
					t.Errorf("Name = %v, want %v", metadata.Name, tt.wantMetadata.Name)
				}
				if metadata.Type != tt.wantMetadata.Type {
					t.Errorf("Type = %v, want %v", metadata.Type, tt.wantMetadata.Type)
				}
				if metadata.SourceType != tt.wantMetadata.SourceType {
					t.Errorf("SourceType = %v, want %v", metadata.SourceType, tt.wantMetadata.SourceType)
				}
				if metadata.SourceURL != tt.wantMetadata.SourceURL {
					t.Errorf("SourceURL = %v, want %v", metadata.SourceURL, tt.wantMetadata.SourceURL)
				}
				if !metadata.FirstInstalled.Equal(tt.wantMetadata.FirstInstalled) {
					t.Errorf("FirstInstalled = %v, want %v", metadata.FirstInstalled, tt.wantMetadata.FirstInstalled)
				}
				if !metadata.LastUpdated.Equal(tt.wantMetadata.LastUpdated) {
					t.Errorf("LastUpdated = %v, want %v", metadata.LastUpdated, tt.wantMetadata.LastUpdated)
				}
			}
		})
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Now().UTC()
	original := &ResourceMetadata{
		Name:           "roundtrip-test",
		Type:           resource.Skill,
		SourceType:     "local",
		SourceURL:      "file:///home/user/skills/roundtrip",
		FirstInstalled: now,
		LastUpdated:    now,
	}

	// Save
	if err := Save(original, tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := Load(original.Name, original.Type, tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare
	if loaded.Name != original.Name {
		t.Errorf("Name = %v, want %v", loaded.Name, original.Name)
	}
	if loaded.Type != original.Type {
		t.Errorf("Type = %v, want %v", loaded.Type, original.Type)
	}
	if loaded.SourceType != original.SourceType {
		t.Errorf("SourceType = %v, want %v", loaded.SourceType, original.SourceType)
	}
	if loaded.SourceURL != original.SourceURL {
		t.Errorf("SourceURL = %v, want %v", loaded.SourceURL, original.SourceURL)
	}

	// Time comparison - allow for microsecond differences due to JSON marshaling
	if !loaded.FirstInstalled.Equal(original.FirstInstalled) {
		t.Errorf("FirstInstalled = %v, want %v", loaded.FirstInstalled, original.FirstInstalled)
	}
	if !loaded.LastUpdated.Equal(original.LastUpdated) {
		t.Errorf("LastUpdated = %v, want %v", loaded.LastUpdated, original.LastUpdated)
	}
}

func TestJSONMarshaling(t *testing.T) {
	now := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)

	metadata := &ResourceMetadata{
		Name:           "json-test",
		Type:           resource.Agent,
		SourceType:     "github",
		SourceURL:      "gh:owner/repo/agents/json-test.md",
		FirstInstalled: now,
		LastUpdated:    now,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify JSON structure
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check all fields are present
	expectedFields := []string{"name", "type", "source_type", "source_url", "first_installed", "last_updated"}
	for _, field := range expectedFields {
		if _, exists := rawJSON[field]; !exists {
			t.Errorf("JSON missing field: %s", field)
		}
	}

	// Unmarshal back
	var unmarshaled ResourceMetadata
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() into struct error = %v", err)
	}

	// Verify all fields
	if unmarshaled.Name != metadata.Name {
		t.Errorf("Name = %v, want %v", unmarshaled.Name, metadata.Name)
	}
	if unmarshaled.Type != metadata.Type {
		t.Errorf("Type = %v, want %v", unmarshaled.Type, metadata.Type)
	}
	if unmarshaled.SourceType != metadata.SourceType {
		t.Errorf("SourceType = %v, want %v", unmarshaled.SourceType, metadata.SourceType)
	}
	if unmarshaled.SourceURL != metadata.SourceURL {
		t.Errorf("SourceURL = %v, want %v", unmarshaled.SourceURL, metadata.SourceURL)
	}
	if !unmarshaled.FirstInstalled.Equal(metadata.FirstInstalled) {
		t.Errorf("FirstInstalled = %v, want %v", unmarshaled.FirstInstalled, metadata.FirstInstalled)
	}
	if !unmarshaled.LastUpdated.Equal(metadata.LastUpdated) {
		t.Errorf("LastUpdated = %v, want %v", unmarshaled.LastUpdated, metadata.LastUpdated)
	}
}

func TestTimestampFormat(t *testing.T) {
	// Test that timestamps are stored in RFC3339 format
	tmpDir := t.TempDir()

	testTime := time.Date(2024, 6, 1, 15, 30, 45, 0, time.UTC)
	metadata := &ResourceMetadata{
		Name:           "time-format-test",
		Type:           resource.Command,
		SourceType:     "local",
		SourceURL:      "file:///test",
		FirstInstalled: testTime,
		LastUpdated:    testTime,
	}

	// Save
	if err := Save(metadata, tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read raw JSON
	metadataPath := GetMetadataPath(metadata.Name, metadata.Type, tmpDir)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	// Parse as map to check raw string format
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify timestamps are strings in RFC3339 format
	firstInstalled, ok := rawJSON["first_installed"].(string)
	if !ok {
		t.Fatal("first_installed is not a string")
	}

	lastUpdated, ok := rawJSON["last_updated"].(string)
	if !ok {
		t.Fatal("last_updated is not a string")
	}

	// Parse timestamps to verify RFC3339 format
	_, err = time.Parse(time.RFC3339, firstInstalled)
	if err != nil {
		t.Errorf("first_installed not in RFC3339 format: %v", err)
	}

	_, err = time.Parse(time.RFC3339, lastUpdated)
	if err != nil {
		t.Errorf("last_updated not in RFC3339 format: %v", err)
	}

	// Expected format: 2024-06-01T15:30:45Z
	expectedFormat := "2024-06-01T15:30:45Z"
	if firstInstalled != expectedFormat {
		t.Errorf("first_installed = %v, want %v", firstInstalled, expectedFormat)
	}
	if lastUpdated != expectedFormat {
		t.Errorf("last_updated = %v, want %v", lastUpdated, expectedFormat)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid JSON file
	invalidPath := filepath.Join(tmpDir, ".metadata", "commands", "invalid-metadata.json")
	if err := os.MkdirAll(filepath.Dir(invalidPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	invalidJSON := []byte(`{"name": "invalid", "type": "command", invalid json}`)
	if err := os.WriteFile(invalidPath, invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Try to load
	_, err := Load("invalid", resource.Command, tmpDir)
	if err == nil {
		t.Error("Load() expected error for invalid JSON, got nil")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Use a nested path that doesn't exist
	nestedRepo := filepath.Join(tmpDir, "nested", "repo", "path")

	metadata := &ResourceMetadata{
		Name:           "nested-test",
		Type:           resource.Skill,
		SourceType:     "local",
		SourceURL:      "file:///test",
		FirstInstalled: time.Now().UTC(),
		LastUpdated:    time.Now().UTC(),
	}

	// Save should create all necessary directories
	if err := Save(metadata, nestedRepo); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	metadataPath := GetMetadataPath(metadata.Name, metadata.Type, nestedRepo)
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Errorf("Save() did not create file at %s", metadataPath)
	}
}
