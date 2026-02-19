package resource

import (
	"fmt"
	"regexp"
	"strings"
)

// ResourceType represents the type of AI resource
type ResourceType string

const (
	// Command represents a single command resource (markdown file)
	Command ResourceType = "command"
	// Skill represents a skill resource (folder with SKILL.md)
	Skill ResourceType = "skill"
	// Agent represents an agent resource (single .md file)
	Agent ResourceType = "agent"
	// PackageType represents a package resource (JSON file with resource references)
	PackageType ResourceType = "package"
)

// ResourceHealth represents the health status of an installed resource
type ResourceHealth string

const (
	// HealthOK indicates the resource symlink is valid and target exists
	HealthOK ResourceHealth = "ok"
	// HealthBroken indicates the resource symlink target doesn't exist
	HealthBroken ResourceHealth = "broken"
)

// Resource represents an AI resource (command, skill, or agent)
type Resource struct {
	Name        string            `json:"name" yaml:"name"` // For nested commands, contains full path (e.g., "api/deploy")
	Type        ResourceType      `json:"type" yaml:"type"`
	Description string            `json:"description" yaml:"description"`
	Version     string            `json:"version,omitempty" yaml:"version,omitempty"`
	Author      string            `json:"author,omitempty" yaml:"author,omitempty"`
	License     string            `json:"license,omitempty" yaml:"license,omitempty"`
	Path        string            `json:"path" yaml:"path"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Health      ResourceHealth    `json:"health,omitempty" yaml:"health,omitempty"`
}

var (
	// nameRegex matches valid resource names: lowercase alphanumeric + hyphens
	// Must not start/end with hyphen, no consecutive hyphens
	nameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

	// namePartRegex matches valid resource name parts (for nested paths)
	// Each part separated by / must be valid
	namePartRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// ValidateName validates a resource name according to agentskills.io spec
// Rules:
// - 1-64 characters (per segment for nested paths)
// - Lowercase alphanumeric + hyphens only (+ slashes for nested paths)
// - Cannot start or end with hyphen
// - No consecutive hyphens
// - For nested paths (containing /), each segment must be valid
func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	// Check for consecutive hyphens
	if strings.Contains(name, "--") {
		return fmt.Errorf("name cannot contain consecutive hyphens")
	}

	// If name contains slashes, validate each part separately (nested commands)
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		for i, part := range parts {
			if len(part) == 0 {
				return fmt.Errorf("empty segment in path at position %d", i)
			}
			if len(part) > 64 {
				return fmt.Errorf("segment '%s' too long (%d chars, max 64)", part, len(part))
			}
			if !namePartRegex.MatchString(part) {
				return fmt.Errorf("segment '%s' invalid: must be lowercase alphanumeric + hyphens, cannot start/end with hyphen", part)
			}
		}
		return nil
	}

	// Single-part name (no slashes)
	if len(name) > 64 {
		return fmt.Errorf("name too long (%d chars, max 64)", len(name))
	}

	if !nameRegex.MatchString(name) {
		return fmt.Errorf("name must be lowercase alphanumeric + hyphens, cannot start/end with hyphen")
	}

	return nil
}

// ValidateDescription validates a resource description
// For skills: 1-1024 characters
// For commands: flexible (minimum 1 character)
func ValidateDescription(description string, resourceType ResourceType) error {
	if len(description) == 0 {
		return fmt.Errorf("description cannot be empty")
	}

	if resourceType == Skill {
		if len(description) > 1024 {
			return fmt.Errorf("skill description too long (%d chars, max 1024)", len(description))
		}
	}

	return nil
}

// Validate validates the resource
func (r *Resource) Validate() error {
	if err := ValidateName(r.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	if err := ValidateDescription(r.Description, r.Type); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	if r.Type != Command && r.Type != Skill && r.Type != Agent {
		return fmt.Errorf("invalid resource type: %s (must be 'command', 'skill', or 'agent')", r.Type)
	}

	return nil
}
