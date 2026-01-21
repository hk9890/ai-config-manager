package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
)

const maxDepth = 5

// DiscoverCommands discovers command resources (.md files) in a repository
// It searches in priority locations first, then falls back to recursive search
func DiscoverCommands(basePath string, subpath string) ([]*resource.Resource, error) {
	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Verify base path exists
	if _, err := os.Stat(searchPath); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Try priority locations first
	commands, err := searchPriorityLocations(searchPath)
	if err == nil && len(commands) > 0 {
		return deduplicateCommands(commands), nil
	}

	// Fall back to recursive search (max depth 5)
	commands, err = recursiveSearchCommands(searchPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search for commands: %w", err)
	}

	return deduplicateCommands(commands), nil
}

// searchPriorityLocations searches standard command directories
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

		commands, err := searchDirectory(dir)
		if err != nil {
			// Continue searching other directories even if one fails
			continue
		}

		allCommands = append(allCommands, commands...)
	}

	return allCommands, nil
}

// recursiveSearchCommands performs a recursive search for command files
func recursiveSearchCommands(basePath string, depth int) ([]*resource.Resource, error) {
	if depth > maxDepth {
		return nil, nil
	}

	var allCommands []*resource.Resource

	// Search current directory
	commands, err := searchDirectory(basePath)
	if err == nil {
		allCommands = append(allCommands, commands...)
	}

	// Recursively search subdirectories
	entries, err := os.ReadDir(basePath)
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

		subdirPath := filepath.Join(basePath, entry.Name())
		subCommands, err := recursiveSearchCommands(subdirPath, depth+1)
		if err == nil {
			allCommands = append(allCommands, subCommands...)
		}
	}

	return allCommands, nil
}

// searchDirectory searches a single directory for command files
func searchDirectory(dir string) ([]*resource.Resource, error) {
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

		// Try to load as command
		cmd, err := resource.LoadCommand(filePath)
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
