package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
)

// SkillCandidate represents a skill discovered during search
type SkillCandidate struct {
	Path     string
	Resource *resource.Resource
}

// DiscoverSkills searches for skills following the priority-based algorithm
// It searches in standard locations first, then falls back to recursive search
func DiscoverSkills(basePath string, subpath string) ([]*resource.Resource, error) {
	// Validate base path exists
	if _, err := os.Stat(basePath); err != nil {
		return nil, fmt.Errorf("base path does not exist: %w", err)
	}

	// Build the search root
	searchRoot := basePath
	if subpath != "" {
		searchRoot = filepath.Join(basePath, subpath)
		// Validate search root exists
		if _, err := os.Stat(searchRoot); err != nil {
			return nil, fmt.Errorf("search path does not exist: %w", err)
		}
	}

	// Check if searchRoot itself is a skill (has SKILL.md)
	if isSkillDir(searchRoot) {
		skill, err := resource.LoadSkill(searchRoot)
		if err == nil && skill.Name != "" && skill.Description != "" {
			return []*resource.Resource{skill}, nil
		}
	}

	// Priority locations to search (relative to searchRoot)
	priorityLocations := []string{
		"skills",
		".claude/skills",
		".opencode/skills",
		".github/skills",
		".codex/skills",
		".cursor/skills",
		".goose/skills",
		".kilocode/skills",
		".kiro/skills",
		".roo/skills",
		".trae/skills",
		".agents/skills",
		".agent/skills",
	}

	// Search in priority locations
	candidates := make(map[string]*resource.Resource) // Map by name for deduplication
	for _, location := range priorityLocations {
		locationPath := filepath.Join(searchRoot, location)
		if skills := searchSkillsInDir(locationPath); len(skills) > 0 {
			for _, skill := range skills {
				// First found wins (deduplication by name)
				if _, exists := candidates[skill.Name]; !exists {
					candidates[skill.Name] = skill
				}
			}
		}
	}

	// If we found skills, return them
	if len(candidates) > 0 {
		return mapToSlice(candidates), nil
	}

	// Fall back to recursive search (max depth 5)
	recursiveSkills, err := recursiveSearchSkills(searchRoot, 0)
	if err != nil {
		// If recursive search fails, just return what we have
		return mapToSlice(candidates), nil
	}

	for _, skill := range recursiveSkills {
		if _, exists := candidates[skill.Name]; !exists {
			candidates[skill.Name] = skill
		}
	}

	return mapToSlice(candidates), nil
}

// isSkillDir checks if a directory contains a SKILL.md file
func isSkillDir(path string) bool {
	skillMdPath := filepath.Join(path, "SKILL.md")
	info, err := os.Stat(skillMdPath)
	return err == nil && !info.IsDir()
}

// searchSkillsInDir searches for skills in a specific directory
// Returns all valid skills found (directories with SKILL.md)
func searchSkillsInDir(dirPath string) []*resource.Resource {
	var skills []*resource.Resource

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return skills
	}

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return skills
	}

	// Check each subdirectory for SKILL.md
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(dirPath, entry.Name())
		if !isSkillDir(skillPath) {
			continue
		}

		// Try to load the skill
		skill, err := resource.LoadSkill(skillPath)
		if err != nil {
			continue // Skip invalid skills
		}

		// Skip if name or description is missing
		if skill.Name == "" || skill.Description == "" {
			continue
		}

		skills = append(skills, skill)
	}

	return skills
}

// recursiveSearchSkills performs a recursive directory search for skills
// Limited to maxDepth (5) to prevent excessive searching
func recursiveSearchSkills(rootPath string, currentDepth int) ([]*resource.Resource, error) {
	var skills []*resource.Resource
	const maxDepth = 5

	// Stop if we've reached max depth
	if currentDepth >= maxDepth {
		return skills, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Check each entry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(rootPath, entry.Name())

		// Skip hidden directories (except those in our priority list)
		if len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}

		// Check if this directory is a skill
		if isSkillDir(entryPath) {
			skill, err := resource.LoadSkill(entryPath)
			if err == nil && skill.Name != "" && skill.Description != "" {
				skills = append(skills, skill)
			}
			// Don't recurse into skill directories
			continue
		}

		// Recurse into subdirectory
		subSkills, err := recursiveSearchSkills(entryPath, currentDepth+1)
		if err == nil {
			skills = append(skills, subSkills...)
		}
	}

	return skills, nil
}

// mapToSlice converts a map of resources to a slice
func mapToSlice(m map[string]*resource.Resource) []*resource.Resource {
	result := make([]*resource.Resource, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
