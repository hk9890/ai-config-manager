package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Load loads a resource from the filesystem
// It detects whether the path is a command, agent, or skill
func Load(path string) (*Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		// Load as skill
		return LoadSkill(path)
	}

	// For .md files, detect type and load appropriately
	resourceType, err := DetectType(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect type: %w", err)
	}

	switch resourceType {
	case Agent:
		return LoadAgent(path)
	case Command:
		return LoadCommand(path)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// DetectType detects the resource type from a filesystem path
func DetectType(path string) (ResourceType, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		// Check if SKILL.md exists
		skillPath := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			return Skill, nil
		}
		return "", fmt.Errorf("directory does not contain SKILL.md")
	}

	// Check if it's a .md file
	if filepath.Ext(path) == ".md" {
		// Use path-based detection first (more reliable for bulk imports)
		// Check if file is in agents/ or commands/ directory
		cleanPath := filepath.ToSlash(filepath.Clean(path))
		
		// Check if path contains /agents/ anywhere (handles nested agents)
		if strings.Contains(cleanPath, "/agents/") || strings.HasPrefix(cleanPath, "agents/") {
			return Agent, nil
		}
		
		// Check if path contains /commands/ anywhere (handles nested commands)
		if strings.Contains(cleanPath, "/commands/") || strings.HasPrefix(cleanPath, "commands/") {
			return Command, nil
		}

		// Parse frontmatter to distinguish between agent and command
		frontmatter, _, err := ParseFrontmatter(path)
		if err != nil {
			// If we can't parse frontmatter, fall back to Command
			return Command, nil
		}

		// Check for agent-specific fields (type, instructions, capabilities)
		// If any exist, it's an agent
		_, hasType := frontmatter["type"]
		_, hasInstructions := frontmatter["instructions"]
		_, hasCapabilities := frontmatter["capabilities"]
		if hasType || hasInstructions || hasCapabilities {
			return Agent, nil
		}

		// Check for command-specific fields (agent, model, allowed-tools)
		// If any exist, it's a command
		_, hasAgent := frontmatter["agent"]
		_, hasModel := frontmatter["model"]
		_, hasAllowedTools := frontmatter["allowed-tools"]
		if hasAgent || hasModel || hasAllowedTools {
			return Command, nil
		}

		// If we can't determine from path or fields, try loading as both
		// Prefer command if both succeed (backward compatibility and more common)
		if _, cmdErr := LoadCommand(path); cmdErr == nil {
			return Command, nil
		}
		if _, agentErr := LoadAgent(path); agentErr == nil {
			return Agent, nil
		}

		// Default to command for backward compatibility
		return Command, nil
	}

	return "", fmt.Errorf("not a valid resource (must be .md file or directory with SKILL.md)")
}
