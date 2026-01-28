package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

const maxDepth = 5

// DiscoverCommands discovers command resources (.md files) in a repository
// It searches in priority locations first, then also does recursive search
// to find commands outside of priority directories
func DiscoverCommands(basePath string, subpath string) ([]*resource.Resource, error) {
	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Verify search path exists, but be lenient with subpaths
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

	// Verify base path exists (after trying to find a valid parent)
	if searchPathErr != nil {
		return nil, fmt.Errorf("path does not exist: %w", searchPathErr)
	}

	if !searchPathInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", searchPath)
	}

	var allCommands []*resource.Resource

	// Search priority locations
	priorityCommands, err := searchPriorityLocations(searchPath)
	if err == nil && len(priorityCommands) > 0 {
		allCommands = append(allCommands, priorityCommands...)
	}

	// Also do recursive search to find commands outside priority directories
	// (e.g., in nested/ or other top-level directories)
	recursiveCommands, err := recursiveSearchCommands(searchPath, 0, searchPath)
	if err == nil && len(recursiveCommands) > 0 {
		allCommands = append(allCommands, recursiveCommands...)
	}

	// Return deduplicated results (may be empty)
	return deduplicateCommands(allCommands), nil
}

// searchPriorityLocations searches standard command directories RECURSIVELY
func searchPriorityLocations(basePath string) ([]*resource.Resource, error) {
	priorityDirs := []string{
		filepath.Join(basePath, "commands"),
		filepath.Join(basePath, ".claude", "commands"),
		filepath.Join(basePath, ".opencode", "commands"),
	}

	var allCommands []*resource.Resource

	for _, dir := range priorityDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Directory doesn't exist, skip
		}

		// Recursively search within priority directory (starting at depth 0)
		// Use the priority directory as the basePath for relative path calculation
		commands, err := searchCommandsInDirectory(dir, 0, dir)
		if err != nil {
			// Continue searching other directories even if one fails
			continue
		}

		allCommands = append(allCommands, commands...)
	}

	return allCommands, nil
}

// searchCommandsInDirectory recursively searches for command .md files within a directory
// This is used for priority locations to find commands at any depth within that location
// basePath is used to calculate relative paths for commands
func searchCommandsInDirectory(dir string, depth int, basePath string) ([]*resource.Resource, error) {
	if depth > maxDepth {
		return nil, nil
	}

	var commands []*resource.Resource

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively search subdirectories
			subCommands, err := searchCommandsInDirectory(entryPath, depth+1, basePath)
			if err == nil {
				commands = append(commands, subCommands...)
			}
			continue
		}

		// Check if it's a .md file
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Exclude special files
		if isExcludedFile(entry.Name()) {
			continue
		}

		// Try to load as command with relative path calculation
		cmd, err := resource.LoadCommandWithBase(entryPath, basePath)
		if err != nil {
			// Invalid command file, skip it
			continue
		}

		commands = append(commands, cmd)
	}

	return commands, nil
}

// recursiveSearchCommands performs a recursive search for command files
// basePath is used to calculate relative paths
func recursiveSearchCommands(currentPath string, depth int, basePath string) ([]*resource.Resource, error) {
	if depth > maxDepth {
		return nil, nil
	}

	var allCommands []*resource.Resource

	// Search current directory
	commands, err := searchDirectory(currentPath, basePath)
	if err == nil {
		allCommands = append(allCommands, commands...)
	}

	// Recursively search subdirectories
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories (except .claude and .opencode which are handled in priority)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip agents and skills directories (they're handled by their own discovery)
		if entry.Name() == "agents" || entry.Name() == "skills" {
			continue
		}

		subdirPath := filepath.Join(currentPath, entry.Name())

		// Handle 'commands' directories specially
		if entry.Name() == "commands" {
			// Skip root-level 'commands' (already handled by priority search)
			if depth == 0 {
				continue
			}
			// For nested 'commands' directories (depth > 0):
			// Search them recursively using the 'commands' directory as basePath
			// (similar to how priority search works)
			cmdDirCommands, err := searchCommandsInDirectory(subdirPath, 0, subdirPath)
			if err == nil {
				allCommands = append(allCommands, cmdDirCommands...)
			}
			// Don't recurse into this 'commands' directory again with recursiveSearchCommands
			continue
		}

		// For non-'commands' directories, continue recursive search
		subCommands, err := recursiveSearchCommands(subdirPath, depth+1, basePath)
		if err == nil {
			allCommands = append(allCommands, subCommands...)
		}
	}

	return allCommands, nil
}

// searchDirectory searches a single directory for command files
// basePath is used to calculate relative paths
func searchDirectory(dir string, basePath string) ([]*resource.Resource, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var commands []*resource.Resource

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a .md file
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Exclude special files
		if isExcludedFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())

		// Try to load as command with relative path calculation
		cmd, err := resource.LoadCommandWithBase(filePath, basePath)
		if err != nil {
			// Invalid command file, skip it
			continue
		}

		commands = append(commands, cmd)
	}

	return commands, nil
}

// isExcludedFile checks if a filename should be excluded from command discovery
func isExcludedFile(filename string) bool {
	excluded := []string{
		"SKILL.md",
		"README.md",
		"readme.md",
		"Readme.md",
		"REFERENCE.md",
		"reference.md",
		"Reference.md",
	}

	for _, excl := range excluded {
		if filename == excl {
			return true
		}
	}

	return false
}

// deduplicateCommands removes duplicate commands by name, keeping the first occurrence
func deduplicateCommands(commands []*resource.Resource) []*resource.Resource {
	seen := make(map[string]bool)
	var unique []*resource.Resource

	for _, cmd := range commands {
		if !seen[cmd.Name] {
			seen[cmd.Name] = true
			unique = append(unique, cmd)
		}
	}

	return unique
}
