package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentResource represents an agent resource
type AgentResource struct {
	Resource
	Type         string   `yaml:"type,omitempty"`         // Agent type/role (OpenCode format)
	Instructions string   `yaml:"instructions,omitempty"` // Agent instructions (OpenCode format)
	Capabilities []string `yaml:"capabilities,omitempty"` // Optional capabilities list
	Content      string   `yaml:"-"`                      // The markdown content
}

// LoadAgent loads an agent resource from a markdown file
// Supports both OpenCode format (type, instructions) and Claude format
func LoadAgent(filePath string) (*Resource, error) {
	// Validate it's a .md file
	if filepath.Ext(filePath) != ".md" {
		return nil, WrapLoadError(filePath, Agent, fmt.Errorf("agent must be a .md file"))
	}

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, WrapLoadError(filePath, Agent, fmt.Errorf("file does not exist: %w", err))
	}

	// Extract name from filename (without .md extension)
	name := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Parse frontmatter
	frontmatter, _, err := ParseFrontmatter(filePath)
	if err != nil {
		return nil, NewValidationError(filePath, "agent", name, "frontmatter", err)
	}

	// Build resource
	resource := &Resource{
		Name:        name,
		Type:        Agent,
		Description: frontmatter.GetString("description"),
		Version:     frontmatter.GetString("version"),
		Author:      frontmatter.GetString("author"),
		License:     frontmatter.GetString("license"),
		Path:        filePath,
		Metadata:    frontmatter.GetMap("metadata"),
	}

	// Validate
	if err := resource.Validate(); err != nil {
		return nil, NewValidationError(filePath, "agent", name, "", err)
	}

	return resource, nil
}

// LoadAgentResource loads a full agent resource with all details
func LoadAgentResource(filePath string) (*AgentResource, error) {
	// Load base resource
	base, err := LoadAgent(filePath)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter again for agent-specific fields
	frontmatter, content, err := ParseFrontmatter(filePath)
	if err != nil {
		return nil, err
	}

	agent := &AgentResource{
		Resource: *base,
		Content:  content,
	}

	// Extract agent-specific fields
	agent.Type = frontmatter.GetString("type")
	agent.Instructions = frontmatter.GetString("instructions")

	// Extract capabilities
	if capVal, ok := frontmatter["capabilities"]; ok {
		if capSlice, ok := capVal.([]interface{}); ok {
			for _, item := range capSlice {
				if str, ok := item.(string); ok {
					agent.Capabilities = append(agent.Capabilities, str)
				}
			}
		}
	}

	return agent, nil
}

// ValidateAgent validates an agent resource structure
func ValidateAgent(filePath string) error {
	_, err := LoadAgent(filePath)
	return err
}

// NewAgentResource creates a new agent resource
func NewAgentResource(name, description string) *AgentResource {
	return &AgentResource{
		Resource: Resource{
			Name:        name,
			Type:        Agent,
			Description: description,
			Metadata:    make(map[string]string),
		},
	}
}

// WriteAgent writes an agent resource to a file
func WriteAgent(agent *AgentResource, filePath string) error {
	// Build frontmatter
	frontmatter := Frontmatter{
		"description": agent.Description,
	}

	if agent.Type != "" {
		frontmatter["type"] = agent.Type
	}
	if agent.Instructions != "" {
		frontmatter["instructions"] = agent.Instructions
	}
	if len(agent.Capabilities) > 0 {
		frontmatter["capabilities"] = agent.Capabilities
	}
	if agent.Version != "" {
		frontmatter["version"] = agent.Version
	}
	if agent.Author != "" {
		frontmatter["author"] = agent.Author
	}
	if agent.License != "" {
		frontmatter["license"] = agent.License
	}
	if len(agent.Metadata) > 0 {
		frontmatter["metadata"] = agent.Metadata
	}

	return WriteFrontmatter(filePath, frontmatter, agent.Content)
}
