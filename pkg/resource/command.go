package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CommandResource represents a command resource
type CommandResource struct {
	Resource
	Agent        string   `yaml:"agent,omitempty"`
	Model        string   `yaml:"model,omitempty"`
	AllowedTools []string `yaml:"allowed-tools,omitempty"`
	Content      string   `yaml:"-"` // The markdown content
}

// LoadCommand loads a command resource from a markdown file
func LoadCommand(filePath string) (*Resource, error) {
	// Validate it's a .md file
	if filepath.Ext(filePath) != ".md" {
		return nil, fmt.Errorf("command must be a .md file")
	}

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	// Parse frontmatter
	frontmatter, _, err := ParseFrontmatter(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract name from filename (without .md extension)
	name := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Build resource
	resource := &Resource{
		Name:        name,
		Type:        Command,
		Description: frontmatter.GetString("description"),
		Version:     frontmatter.GetString("version"),
		Author:      frontmatter.GetString("author"),
		License:     frontmatter.GetString("license"),
		Path:        filePath,
		Metadata:    frontmatter.GetMap("metadata"),
	}

	// Validate
	if err := resource.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return resource, nil
}

// ValidateCommand validates a command resource structure
func ValidateCommand(filePath string) error {
	_, err := LoadCommand(filePath)
	return err
}

// NewCommandResource creates a new command resource
func NewCommandResource(name, description string) *CommandResource {
	return &CommandResource{
		Resource: Resource{
			Name:        name,
			Type:        Command,
			Description: description,
			Metadata:    make(map[string]string),
		},
	}
}

// WriteCommand writes a command resource to a file
func WriteCommand(cmd *CommandResource, filePath string) error {
	// Build frontmatter
	frontmatter := Frontmatter{
		"description": cmd.Description,
	}

	if cmd.Agent != "" {
		frontmatter["agent"] = cmd.Agent
	}
	if cmd.Model != "" {
		frontmatter["model"] = cmd.Model
	}
	if len(cmd.AllowedTools) > 0 {
		frontmatter["allowed-tools"] = cmd.AllowedTools
	}
	if cmd.Version != "" {
		frontmatter["version"] = cmd.Version
	}
	if cmd.Author != "" {
		frontmatter["author"] = cmd.Author
	}
	if cmd.License != "" {
		frontmatter["license"] = cmd.License
	}
	if len(cmd.Metadata) > 0 {
		frontmatter["metadata"] = cmd.Metadata
	}

	return WriteFrontmatter(filePath, frontmatter, cmd.Content)
}
