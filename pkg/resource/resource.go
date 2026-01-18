package resource

import (
	"fmt"
	"os"
	"path/filepath"
)

// Load loads a resource from the filesystem
// It detects whether the path is a command (file) or skill (directory)
func Load(path string) (*Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		// Load as skill
		return LoadSkill(path)
	}

	// Load as command
	return LoadCommand(path)
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
		return Command, nil
	}

	return "", fmt.Errorf("not a valid resource (must be .md file or directory with SKILL.md)")
}
