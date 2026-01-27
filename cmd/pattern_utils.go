package cmd

import (
	"fmt"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// ExpandPattern expands a pattern to matching resources from repository.
// If the argument is not a pattern, returns it as-is in a single-element slice.
// Returns a slice of resource arguments in "type/name" format.
func ExpandPattern(mgr *repo.Manager, resourceArg string) ([]string, error) {
	// Parse pattern
	resourceType, _, isPattern := pattern.ParsePattern(resourceArg)
	if !isPattern {
		// Not a pattern, return as-is
		return []string{resourceArg}, nil
	}

	// Create matcher
	matcher, err := pattern.NewMatcher(resourceArg)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern '%s': %w", resourceArg, err)
	}

	// List resources from repository
	var resources []resource.Resource
	if resourceType != "" {
		// Specific type filter
		resources, err = mgr.List(&resourceType)
	} else {
		// All types
		resources, err = mgr.List(nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Filter by pattern
	var matches []string
	for _, res := range resources {
		if matcher.Match(&res) {
			// Return in "type/name" format
			matches = append(matches, FormatResourceArg(&res))
		}
	}

	return matches, nil
}

// ParseResourceArg parses a resource argument in the format "type/name".
// Supports both singular and plural type names (e.g., "skill" and "skills").
// Returns the resource type, name, and any error.
func ParseResourceArg(arg string) (resource.ResourceType, string, error) {
	parts := strings.SplitN(arg, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format: must be 'type/name' (e.g., skill/my-skill, command/my-command, agent/my-agent)")
	}

	typeStr := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if name == "" {
		return "", "", fmt.Errorf("resource name cannot be empty")
	}

	// Parse and validate type
	resourceType, err := parseResourceType(typeStr)
	if err != nil {
		return "", "", err
	}

	return resourceType, name, nil
}

// FormatResourceArg formats a resource as "type/name".
func FormatResourceArg(res *resource.Resource) string {
	return fmt.Sprintf("%s/%s", res.Type, res.Name)
}

// parseResourceType converts a string to ResourceType.
// Supports both singular and plural forms (e.g., "skill" and "skills").
func parseResourceType(s string) (resource.ResourceType, error) {
	// Normalize: lowercase and handle plural forms
	s = strings.ToLower(s)
	switch s {
	case "skill", "skills":
		return resource.Skill, nil
	case "command", "commands":
		return resource.Command, nil
	case "agent", "agents":
		return resource.Agent, nil
	default:
		return "", fmt.Errorf("invalid resource type '%s': must be one of 'skill', 'command', or 'agent'", s)
	}
}
