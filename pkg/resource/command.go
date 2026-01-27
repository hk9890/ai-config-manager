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
// For backward compatibility, this does not calculate RelativePath
func LoadCommand(filePath string) (*Resource, error) {
	return LoadCommandWithBase(filePath, "")
}

// LoadCommandWithBase loads a command resource and calculates RelativePath if basePath is provided
// basePath should be the directory to calculate relative paths from (e.g., "commands/")
// If basePath is empty, RelativePath will not be set (backward compatibility)
func LoadCommandWithBase(filePath string, basePath string) (*Resource, error) {
	// Validate it's a .md file
	if filepath.Ext(filePath) != ".md" {
		return nil, WrapLoadError(filePath, Command, fmt.Errorf("command must be a .md file"))
	}

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, WrapLoadError(filePath, Command, fmt.Errorf("file does not exist: %w", err))
	}

	// Extract name from filename (without .md extension)
	name := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Parse frontmatter
	frontmatter, _, err := ParseFrontmatter(filePath)
	if err != nil {
		return nil, NewValidationError(filePath, "command", name, "frontmatter", err)
	}

	// Calculate relative path if basePath is provided
	var relativePath string
	if basePath != "" {
		// Clean paths for consistent comparison
		cleanFilePath := filepath.Clean(filePath)
		cleanBasePath := filepath.Clean(basePath)

		// Get relative path from basePath to filePath
		relPath, err := filepath.Rel(cleanBasePath, cleanFilePath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			// Remove the .md extension from relative path
			relativePath = strings.TrimSuffix(relPath, ".md")
		}
	}

	// Use relativePath as name if available (for nested commands)
	// Otherwise fallback to basename (for backward compatibility)
	if relativePath != "" {
		name = relativePath
	}

	// Build resource
	resource := &Resource{
		Name:         name,
		Type:         Command,
		Description:  frontmatter.GetString("description"),
		Version:      frontmatter.GetString("version"),
		Author:       frontmatter.GetString("author"),
		License:      frontmatter.GetString("license"),
		Path:         filePath,
		RelativePath: relativePath,
		Metadata:     frontmatter.GetMap("metadata"),
	}

	// Validate
	if err := resource.Validate(); err != nil {
		return nil, NewValidationError(filePath, "command", name, "", err)
	}

	return resource, nil
}

// LoadCommandResource loads a command resource with full details including content
// For backward compatibility, this does not calculate RelativePath
func LoadCommandResource(filePath string) (*CommandResource, error) {
	return LoadCommandResourceWithBase(filePath, "")
}

// LoadCommandResourceWithBase loads a command resource with full details and calculates RelativePath
func LoadCommandResourceWithBase(filePath string, basePath string) (*CommandResource, error) {
	// Validate it's a .md file
	if filepath.Ext(filePath) != ".md" {
		return nil, fmt.Errorf("command must be a .md file")
	}

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	// Parse frontmatter and content
	frontmatter, content, err := ParseFrontmatter(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract name from filename (without .md extension)
	name := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Calculate relative path if basePath is provided
	var relativePath string
	if basePath != "" {
		cleanFilePath := filepath.Clean(filePath)
		cleanBasePath := filepath.Clean(basePath)

		relPath, err := filepath.Rel(cleanBasePath, cleanFilePath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			relativePath = strings.TrimSuffix(relPath, ".md")
		}
	}

	// Use relativePath as name if available (for nested commands)
	// Otherwise fallback to basename (for backward compatibility)
	if relativePath != "" {
		name = relativePath
	}

	// Build command resource
	cmd := &CommandResource{
		Resource: Resource{
			Name:         name,
			Type:         Command,
			Description:  frontmatter.GetString("description"),
			Version:      frontmatter.GetString("version"),
			Author:       frontmatter.GetString("author"),
			License:      frontmatter.GetString("license"),
			Path:         filePath,
			RelativePath: relativePath,
			Metadata:     frontmatter.GetMap("metadata"),
		},
		Agent:        frontmatter.GetString("agent"),
		Model:        frontmatter.GetString("model"),
		AllowedTools: frontmatter.GetStringSlice("allowed-tools"),
		Content:      content,
	}

	// Validate
	if err := cmd.Resource.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return cmd, nil
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
