package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// ResourceMetadata tracks source information and timestamps for a resource
type ResourceMetadata struct {
	Name           string                `json:"name"`
	Type           resource.ResourceType `json:"type"`
	SourceType     string                `json:"source_type"`
	SourceURL      string                `json:"source_url"`
	FirstInstalled time.Time             `json:"first_installed"`
	LastUpdated    time.Time             `json:"last_updated"`
}

// Save writes metadata to a JSON file in the repository
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

// Load reads metadata from a JSON file in the repository
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

// GetMetadataPath returns the path to the metadata file for a resource
// Pattern: <repoPath>/<type>s/<type>-<name>-metadata.json
// Examples:
//   - command: ~/.local/share/ai-config/repo/commands/command-mycmd-metadata.json
//   - skill: ~/.local/share/ai-config/repo/skills/skill-myskill-metadata.json
//   - agent: ~/.local/share/ai-config/repo/agents/agent-myagent-metadata.json
func GetMetadataPath(name string, resourceType resource.ResourceType, repoPath string) string {
	// Get the resource type directory (commands, skills, agents)
	typeDir := string(resourceType) + "s"

	// Build filename: <type>-<name>-metadata.json
	filename := fmt.Sprintf("%s-%s-metadata.json", resourceType, name)

	return filepath.Join(repoPath, typeDir, filename)
}
