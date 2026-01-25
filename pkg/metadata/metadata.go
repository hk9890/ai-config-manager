package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// ResourceMetadata tracks source information and timestamps for a resource.
// Metadata is stored in JSON format in the centralized .metadata/ directory.
//
// Example locations:
//   - Commands: ~/.local/share/ai-config/repo/.metadata/commands/mycmd-metadata.json
//   - Skills:   ~/.local/share/ai-config/repo/.metadata/skills/myskill-metadata.json
//   - Agents:   ~/.local/share/ai-config/repo/.metadata/agents/myagent-metadata.json
type ResourceMetadata struct {
	Name           string                `json:"name"`            // Resource name
	Type           resource.ResourceType `json:"type"`            // Resource type (command, skill, agent)
	SourceType     string                `json:"source_type"`     // Source type: "github", "local", "file"
	SourceURL      string                `json:"source_url"`      // Source URL or file path
	Ref            string                `json:"ref,omitempty"`   // Git ref (branch/tag/commit), defaults to "main" if empty
	FirstInstalled time.Time             `json:"first_installed"` // When resource was first added
	LastUpdated    time.Time             `json:"last_updated"`    // When resource was last updated
}

// Save writes metadata to a JSON file in the .metadata/ directory.
// Metadata is stored in a centralized location separate from resource files.
func Save(metadata *ResourceMetadata, repoPath string) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	metadataPath := GetMetadataPath(metadata.Name, metadata.Type, repoPath)

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(metadataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to file
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// Load reads metadata from a JSON file in the .metadata/ directory.
// Returns an error if the metadata file doesn't exist or can't be parsed.
func Load(name string, resourceType resource.ResourceType, repoPath string) (*ResourceMetadata, error) {
	metadataPath := GetMetadataPath(name, resourceType, repoPath)

	// Read file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata file not found: %s", metadataPath)
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	// Unmarshal JSON
	var metadata ResourceMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// GetMetadataPath returns the path to the metadata file for a resource.
//
// Metadata is stored in a centralized .metadata/ directory with the pattern:
//
//	<repoPath>/.metadata/<type>s/<name>-metadata.json
//
// This structure keeps metadata separate from resource files for better organization
// and easier backup/management.
//
// Examples:
//   - command: ~/.local/share/ai-config/repo/.metadata/commands/mycmd-metadata.json
//   - skill: ~/.local/share/ai-config/repo/.metadata/skills/myskill-metadata.json
//   - agent: ~/.local/share/ai-config/repo/.metadata/agents/myagent-metadata.json
func GetMetadataPath(name string, resourceType resource.ResourceType, repoPath string) string {
	// Get the resource type directory (commands, skills, agents)
	typeDir := string(resourceType) + "s"

	// Build filename: <name>-metadata.json
	filename := fmt.Sprintf("%s-metadata.json", name)

	return filepath.Join(repoPath, ".metadata", typeDir, filename)
}
