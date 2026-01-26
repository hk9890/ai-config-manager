package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// SkillCandidate represents a skill discovered during search
type SkillCandidate struct {
	Path     string
	Resource *resource.Resource
}

// DiscoveryError represents an error encountered during resource discovery
type DiscoveryError struct {
	Path  string // Path where the error occurred
	Error error  // The error that occurred
}

// SkillDiscoveryResult contains discovered skills and any errors encountered
type SkillDiscoveryResult struct {
	Skills []*resource.Resource
	Errors []DiscoveryError
}

// DiscoverSkills searches for skills following the priority-based algorithm
// It searches in standard locations first, then falls back to recursive search
func DiscoverSkills(basePath string, subpath string) ([]*resource.Resource, error) {
	skills, _, err := DiscoverSkillsWithErrors(basePath, subpath)
	return skills, err
}

// DiscoverSkillsWithErrors searches for skills and returns both successful discoveries and errors
// It searches in standard locations first, then falls back to recursive search
// Returns successfully discovered skills, any errors encountered, and fatal errors
func DiscoverSkillsWithErrors(basePath string, subpath string) ([]*resource.Resource, []DiscoveryError, error) {
	var allErrors []DiscoveryError

	// Build the initial search root
	searchRoot := basePath
	if subpath != "" {
		searchRoot = filepath.Join(basePath, subpath)
	}

	// Check if searchRoot exists and is accessible
	searchRootInfo, searchRootErr := os.Stat(searchRoot)

	// If searchRoot exists, check if it's a skill directory
	if searchRootErr == nil && searchRootInfo.IsDir() {
		if isSkillDir(searchRoot) {
			skill, err := resource.LoadSkill(searchRoot)
			if err != nil {
				// Collect error instead of silently failing
				allErrors = append(allErrors, DiscoveryError{
					Path:  searchRoot,
					Error: err,
				})
			} else if skill.Name != "" && skill.Description != "" {
				return []*resource.Resource{skill}, allErrors, nil
			}
		}
	}

	// If searchRoot doesn't exist, try searching from immediate parent directory
	// This handles cases where the path might be slightly incorrect (e.g., typo in skill name)
	// but we don't want to fall back too far (which would mask real errors)
	if searchRootErr != nil {
		parentPath := filepath.Dir(searchRoot)
		// Only fall back if the parent is a valid, different directory
		if parentPath != searchRoot && parentPath != "." && parentPath != "/" {
			if info, err := os.Stat(parentPath); err == nil && info.IsDir() {
				searchRoot = parentPath
			} else {
				// Parent doesn't exist either, this is a real error
				return nil, allErrors, fmt.Errorf("base path does not exist: %w", searchRootErr)
			}
		} else {
			// Can't fall back, this is a real error
			return nil, allErrors, fmt.Errorf("base path does not exist: %w", searchRootErr)
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
		skills, errs := searchSkillsInDir(locationPath)
		allErrors = append(allErrors, errs...)
		if len(skills) > 0 {
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
		return mapToSlice(candidates), allErrors, nil
	}

	// Fall back to recursive search (max depth 5)
	recursiveSkills, recursiveErrs, err := recursiveSearchSkills(searchRoot, 0)
	allErrors = append(allErrors, recursiveErrs...)
	if err != nil {
		// If recursive search fails, just return what we have
		return mapToSlice(candidates), allErrors, nil
	}

	for _, skill := range recursiveSkills {
		if _, exists := candidates[skill.Name]; !exists {
			candidates[skill.Name] = skill
		}
	}

	return mapToSlice(candidates), allErrors, nil
}

// isSkillDir checks if a directory contains a SKILL.md file
func isSkillDir(path string) bool {
	skillMdPath := filepath.Join(path, "SKILL.md")
	info, err := os.Stat(skillMdPath)
	return err == nil && !info.IsDir()
}

// searchSkillsInDir searches for skills in a specific directory
// Returns all valid skills found (directories with SKILL.md) and any errors encountered
func searchSkillsInDir(dirPath string) ([]*resource.Resource, []DiscoveryError) {
	var skills []*resource.Resource
	var errors []DiscoveryError

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return skills, errors
	}

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return skills, errors
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
			// Collect error instead of silently skipping
			errors = append(errors, DiscoveryError{
				Path:  skillPath,
				Error: err,
			})
			continue
		}

		// Skip if name or description is missing
		if skill.Name == "" || skill.Description == "" {
			continue
		}

		skills = append(skills, skill)
	}

	return skills, errors
}

// recursiveSearchSkills performs a recursive directory search for skills
// Limited to maxDepth (5) to prevent excessive searching
// Returns both discovered skills and any errors encountered
func recursiveSearchSkills(rootPath string, currentDepth int) ([]*resource.Resource, []DiscoveryError, error) {
	var skills []*resource.Resource
	var errors []DiscoveryError
	const maxDepth = 5

	// Stop if we've reached max depth
	if currentDepth >= maxDepth {
		return skills, errors, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, errors, fmt.Errorf("failed to read directory: %w", err)
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
			if err != nil {
				// Collect error instead of silently skipping
				errors = append(errors, DiscoveryError{
					Path:  entryPath,
					Error: err,
				})
			} else if skill.Name != "" && skill.Description != "" {
				skills = append(skills, skill)
			}
			// Don't recurse into skill directories
			continue
		}

		// Recurse into subdirectory
		subSkills, subErrs, err := recursiveSearchSkills(entryPath, currentDepth+1)
		errors = append(errors, subErrs...)
		if err == nil {
			skills = append(skills, subSkills...)
		}
	}

	return skills, errors, nil
}

// mapToSlice converts a map of resources to a slice
func mapToSlice(m map[string]*resource.Resource) []*resource.Resource {
	result := make([]*resource.Resource, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
