package sourcemetadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SourceState tracks runtime state for a single source
type SourceState struct {
	SourceID   string    `json:"source_id,omitempty"`
	Added      time.Time `json:"added"`
	LastSynced time.Time `json:"last_synced,omitempty"`
}

// SourceMetadata tracks state for all sources in a repository
type SourceMetadata struct {
	Version int                     `json:"version"`
	Sources map[string]*SourceState `json:"sources"`
}

// Load reads source metadata from .metadata/sources.json
func Load(repoPath string) (*SourceMetadata, error) {
	metadataPath := filepath.Join(repoPath, ".metadata", "sources.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata if file doesn't exist
			return &SourceMetadata{
				Version: 1,
				Sources: make(map[string]*SourceState),
			}, nil
		}
		return nil, fmt.Errorf("failed to read source metadata: %w", err)
	}

	var metadata SourceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse source metadata: %w", err)
	}

	if metadata.Sources == nil {
		metadata.Sources = make(map[string]*SourceState)
	}

	return &metadata, nil
}

// Save writes source metadata to .metadata/sources.json
func (m *SourceMetadata) Save(repoPath string) error {
	metadataDir := filepath.Join(repoPath, ".metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadataPath := filepath.Join(metadataDir, "sources.json")

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal source metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write source metadata: %w", err)
	}

	return nil
}

// Get returns the state for a given source name
func (m *SourceMetadata) Get(sourceName string) *SourceState {
	return m.Sources[sourceName]
}

// SetAdded sets the Added timestamp for a source
func (m *SourceMetadata) SetAdded(sourceName string, t time.Time) {
	if m.Sources[sourceName] == nil {
		m.Sources[sourceName] = &SourceState{}
	}
	m.Sources[sourceName].Added = t
}

// SetLastSynced sets the LastSynced timestamp for a source
func (m *SourceMetadata) SetLastSynced(sourceName string, t time.Time) {
	if m.Sources[sourceName] == nil {
		m.Sources[sourceName] = &SourceState{}
	}
	m.Sources[sourceName].LastSynced = t
}

// Delete removes a source from the metadata
func (m *SourceMetadata) Delete(sourceName string) {
	delete(m.Sources, sourceName)
}
