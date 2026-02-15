package sourcemetadata

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NewRepo(t *testing.T) {
	tempDir := t.TempDir()

	metadata, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if metadata.Version != 1 {
		t.Errorf("Expected version 1, got %d", metadata.Version)
	}

	if metadata.Sources == nil {
		t.Error("Sources map should not be nil")
	}

	if len(metadata.Sources) != 0 {
		t.Errorf("Expected empty sources, got %d", len(metadata.Sources))
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	// Create metadata
	metadata := &SourceMetadata{
		Version: 1,
		Sources: map[string]*SourceState{
			"test-source": {
				Added:      time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC),
				LastSynced: time.Date(2026, 2, 14, 13, 0, 0, 0, time.UTC),
			},
		},
	}

	// Save
	if err := metadata.Save(tempDir); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	metadataPath := filepath.Join(tempDir, ".metadata", "sources.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Load
	loaded, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify
	if loaded.Version != 1 {
		t.Errorf("Expected version 1, got %d", loaded.Version)
	}

	state := loaded.Get("test-source")
	if state == nil {
		t.Fatal("Expected test-source to exist")
	}

	if !state.Added.Equal(metadata.Sources["test-source"].Added) {
		t.Errorf("Added timestamp mismatch")
	}

	if !state.LastSynced.Equal(metadata.Sources["test-source"].LastSynced) {
		t.Errorf("LastSynced timestamp mismatch")
	}
}

func TestSetAdded(t *testing.T) {
	metadata := &SourceMetadata{
		Version: 1,
		Sources: make(map[string]*SourceState),
	}

	now := time.Now()
	metadata.SetAdded("test-source", now)

	state := metadata.Get("test-source")
	if state == nil {
		t.Fatal("Expected test-source to exist")
	}

	if !state.Added.Equal(now) {
		t.Error("Added timestamp not set correctly")
	}
}

func TestSetLastSynced(t *testing.T) {
	metadata := &SourceMetadata{
		Version: 1,
		Sources: make(map[string]*SourceState),
	}

	now := time.Now()
	metadata.SetLastSynced("test-source", now)

	state := metadata.Get("test-source")
	if state == nil {
		t.Fatal("Expected test-source to exist")
	}

	if !state.LastSynced.Equal(now) {
		t.Error("LastSynced timestamp not set correctly")
	}
}

func TestDelete(t *testing.T) {
	metadata := &SourceMetadata{
		Version: 1,
		Sources: map[string]*SourceState{
			"test-source": {
				Added: time.Now(),
			},
		},
	}

	metadata.Delete("test-source")

	if metadata.Get("test-source") != nil {
		t.Error("Source should have been deleted")
	}
}
