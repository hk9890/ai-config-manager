package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

const (
	// MaxRecursiveDepth is the maximum depth for recursive agent search
	MaxRecursiveDepth = 5
)

// DiscoverAgents discovers agent resources (single .md files) in a repository
// It searches in priority locations first, then falls back to recursive search
// if no agents are found.
//
// Priority locations:
//   - basePath/subpath/agents/
//   - basePath/subpath/.claude/agents/
//   - basePath/subpath/.opencode/agents/
//
// If no agents found in priority locations, performs recursive search (max depth 5)
// looking for .md files with valid agent frontmatter (description required).
//
// Returns deduplicated list of agents by agent name.
func DiscoverAgents(basePath string, subpath string) ([]*resource.Resource, error) {
	resources, _, err := DiscoverAgentsWithErrors(basePath, subpath)
	return resources, err
}

// DiscoverAgentsWithErrors discovers agent resources and returns both successful discoveries and errors
// It searches in priority locations first, then falls back to recursive search
// if no agents are found.
func DiscoverAgentsWithErrors(basePath string, subpath string) ([]*resource.Resource, []DiscoveryError, error) {
	if basePath == "" {
		return nil, nil, fmt.Errorf("basePath cannot be empty")
	}

	var allErrors []DiscoveryError

	// Build full search path
	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Check if search path exists, but be lenient with subpaths
	searchPathInfo, searchPathErr := os.Stat(searchPath)

	// If searchPath doesn't exist but we have a subpath, try parent directories
	if searchPathErr != nil && subpath != "" {
		currentPath := searchPath
		for {
			parentPath := filepath.Dir(currentPath)
			if parentPath == currentPath || parentPath == basePath {
				searchPath = basePath
				break
			}
			if info, err := os.Stat(parentPath); err == nil && info.IsDir() {
				searchPath = parentPath
				break
			}
			currentPath = parentPath
		}
		// Re-check the new search path
		searchPathInfo, searchPathErr = os.Stat(searchPath)
	}

	// Check if search path exists (after trying to find a valid parent)
	if searchPathErr != nil {
		return nil, allErrors, fmt.Errorf("search path does not exist: %w", searchPathErr)
	}

	if !searchPathInfo.IsDir() {
		return nil, allErrors, fmt.Errorf("search path is not a directory: %s", searchPath)
	}

	// Priority locations to search
	priorityLocations := []string{
		filepath.Join(searchPath, "agents"),
		filepath.Join(searchPath, ".claude", "agents"),
		filepath.Join(searchPath, ".opencode", "agents"),
	}

	// Try priority locations first (with recursive search within them)
	agents := make([]*resource.Resource, 0)
	for _, location := range priorityLocations {
		found, errors := searchAgentsInDirectory(location, 0)
		agents = append(agents, found...)
		allErrors = append(allErrors, errors...)
	}

	// If no agents found in priority locations, do recursive search
	if len(agents) == 0 {
		found, errors, err := discoverAgentsRecursive(searchPath, 0)
		if err != nil {
			return nil, allErrors, fmt.Errorf("recursive search failed: %w", err)
		}
		agents = append(agents, found...)
		allErrors = append(allErrors, errors...)
	}

	// Deduplicate by agent name (keep first occurrence)
	agents = deduplicateAgents(agents)

	return agents, allErrors, nil
}

// searchAgentsInDirectory recursively searches for agent .md files within a directory
// This is used for priority locations to find agents at any depth within that location
// Returns both successfully loaded agents and any errors encountered
func searchAgentsInDirectory(dir string, depth int) ([]*resource.Resource, []DiscoveryError) {
	if depth > MaxRecursiveDepth {
		return nil, nil
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directory doesn't exist, not an error
		}
		return nil, []DiscoveryError{{
			Path:  dir,
			Error: fmt.Errorf("failed to stat directory: %w", err),
		}}
	}

	if !info.IsDir() {
		return nil, []DiscoveryError{{
			Path:  dir,
			Error: fmt.Errorf("path is not a directory: %s", dir),
		}}
	}

	agents := make([]*resource.Resource, 0)
	var errors []DiscoveryError

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []DiscoveryError{{
			Path:  dir,
			Error: fmt.Errorf("failed to read directory: %w", err),
		}}
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively search subdirectories
			subAgents, subErrors := searchAgentsInDirectory(entryPath, depth+1)
			agents = append(agents, subAgents...)
			errors = append(errors, subErrors...)
			continue
		}

		// If it's a .md file, try to load as agent
		if filepath.Ext(entry.Name()) == ".md" {
			agent, err := resource.LoadAgent(entryPath)
			if err != nil {
				// Collect error instead of silently skipping
				errors = append(errors, DiscoveryError{
					Path:  entryPath,
					Error: err,
				})
				continue
			}
			agents = append(agents, agent)
		}
	}

	return agents, errors
}

// discoverAgentsInDirectory finds agent .md files in a specific directory
// If recursive is true, it will search subdirectories up to MaxRecursiveDepth
// Returns both successfully loaded agents and any errors encountered
func discoverAgentsInDirectory(dirPath string, recursive bool) ([]*resource.Resource, []DiscoveryError) {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directory doesn't exist, not an error
		}
		return nil, []DiscoveryError{{
			Path:  dirPath,
			Error: fmt.Errorf("failed to stat directory: %w", err),
		}}
	}

	if !info.IsDir() {
		return nil, []DiscoveryError{{
			Path:  dirPath,
			Error: fmt.Errorf("path is not a directory: %s", dirPath),
		}}
	}

	agents := make([]*resource.Resource, 0)
	var errors []DiscoveryError

	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, []DiscoveryError{{
			Path:  dirPath,
			Error: fmt.Errorf("failed to read directory: %w", err),
		}}
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())

		// Skip hidden directories and files (except .claude and .opencode)
		if strings.HasPrefix(entry.Name(), ".") && !recursive {
			continue
		}

		// If it's a .md file, try to load as agent
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			// Only parse files that are in an agents/ subtree
			if !isInAgentsSubtree(entryPath) {
				continue
			}

			agent, err := resource.LoadAgent(entryPath)
			if err != nil {
				// Collect error instead of silently skipping
				errors = append(errors, DiscoveryError{
					Path:  entryPath,
					Error: err,
				})
				continue
			}
			agents = append(agents, agent)
		}
	}

	return agents, errors
}

// discoverAgentsRecursive performs recursive search for agent files
// up to MaxRecursiveDepth
// Returns both successfully loaded agents and any errors encountered
func discoverAgentsRecursive(dirPath string, currentDepth int) ([]*resource.Resource, []DiscoveryError, error) {
	// Check depth limit
	if currentDepth > MaxRecursiveDepth {
		return nil, nil, nil
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, nil, nil
	}

	agents := make([]*resource.Resource, 0)
	var allErrors []DiscoveryError

	// Find agents in current directory
	found, errors := discoverAgentsInDirectory(dirPath, true)
	agents = append(agents, found...)
	allErrors = append(allErrors, errors...)

	// Recursively search subdirectories
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return agents, allErrors, nil // Return what we have so far
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip common directories that typically don't contain agents
		if shouldSkipDirectory(entry.Name()) {
			continue
		}

		subPath := filepath.Join(dirPath, entry.Name())
		subAgents, subErrors, err := discoverAgentsRecursive(subPath, currentDepth+1)
		if err != nil {
			// Continue on error
			continue
		}
		agents = append(agents, subAgents...)
		allErrors = append(allErrors, subErrors...)
	}

	return agents, allErrors, nil
}

// shouldSkipDirectory returns true if the directory should be skipped during recursive search
func shouldSkipDirectory(name string) bool {
	skipDirs := []string{
		"commands", // Commands are handled by command discovery
		"skills",   // Skills are handled by skill discovery
		"documentation",
		"docs",
		"node_modules",
		".git",
		".svn",
		".hg",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
		"obj",
		"__pycache__",
		".pytest_cache",
		".venv",
		"venv",
		"test",
		"tests",
		"examples",
	}

	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}

	return false
}

// isInAgentsSubtree checks if a file path is within an agents/ directory subtree
// Returns true if the path contains /agents/ or /.claude/agents/ or /.opencode/agents/
func isInAgentsSubtree(path string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check for agents/ directories
	return strings.Contains(normalizedPath, "/agents/") ||
		strings.Contains(normalizedPath, "/.claude/agents/") ||
		strings.Contains(normalizedPath, "/.opencode/agents/") ||
		strings.HasSuffix(normalizedPath, "/agents") ||
		strings.HasSuffix(normalizedPath, "/.claude/agents") ||
		strings.HasSuffix(normalizedPath, "/.opencode/agents")
}

// deduplicateAgents removes duplicate agents, keeping the first occurrence
// Deduplication is by agent name
func deduplicateAgents(agents []*resource.Resource) []*resource.Resource {
	seen := make(map[string]bool)
	result := make([]*resource.Resource, 0, len(agents))

	for _, agent := range agents {
		if !seen[agent.Name] {
			seen[agent.Name] = true
			result = append(result, agent)
		}
	}

	return result
}
