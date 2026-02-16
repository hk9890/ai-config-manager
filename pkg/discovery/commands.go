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
	resources, _, err := DiscoverCommandsWithErrors(basePath, subpath)
	return resources, err
}

// DiscoverCommandsWithErrors discovers command resources and returns both successful discoveries and errors
// It searches in priority locations first, then also does recursive search
// to find commands outside of priority directories
func DiscoverCommandsWithErrors(basePath string, subpath string) ([]*resource.Resource, []DiscoveryError, error) {
	var allErrors []DiscoveryError

	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Log discovery start
	if logger != nil {
		logger.Debug("starting command discovery",
			"base_path", basePath,
			"subpath", subpath,
			"search_path", searchPath)
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
		return nil, allErrors, fmt.Errorf("path does not exist: %w", searchPathErr)
	}

	if !searchPathInfo.IsDir() {
		return nil, allErrors, fmt.Errorf("path is not a directory: %s", searchPath)
	}

	var allCommands []*resource.Resource

	// If searchPath itself is a commands directory, search it directly
	if isCommandsDirectory(searchPath) {
		if logger != nil {
			logger.Debug("search path is a commands directory, searching directly",
				"path", searchPath)
		}
		commands, errors := searchCommandsInDirectory(searchPath, 0, searchPath)
		allCommands = append(allCommands, commands...)
		allErrors = append(allErrors, errors...)

		if logger != nil {
			logger.Debug("command discovery completed",
				"commands_found", len(allCommands),
				"errors", len(allErrors))
		}
		return deduplicateCommands(allCommands), allErrors, nil
	}

	// Search priority locations
	priorityCommands, priorityErrors := searchPriorityLocations(searchPath)
	allErrors = append(allErrors, priorityErrors...)
	allCommands = append(allCommands, priorityCommands...)

	if logger != nil {
		logger.Debug("priority locations search completed",
			"commands_found", len(priorityCommands),
			"errors", len(priorityErrors))
	}

	// Also do recursive search to find commands outside priority directories
	// (e.g., in nested/ or other top-level directories)
	recursiveCommands, recursiveErrors := recursiveSearchCommands(searchPath, 0, searchPath)
	allErrors = append(allErrors, recursiveErrors...)
	allCommands = append(allCommands, recursiveCommands...)

	if logger != nil {
		logger.Debug("recursive search completed",
			"commands_found", len(recursiveCommands),
			"errors", len(recursiveErrors))
	}

	// Return deduplicated results (may be empty)
	dedupedCommands := deduplicateCommands(allCommands)

	if logger != nil {
		logger.Debug("command discovery completed",
			"total_commands", len(dedupedCommands),
			"total_errors", len(allErrors))
	}

	return dedupedCommands, allErrors, nil
}

// searchPriorityLocations searches standard command directories RECURSIVELY
func searchPriorityLocations(basePath string) ([]*resource.Resource, []DiscoveryError) {
	priorityDirs := []string{
		filepath.Join(basePath, "commands"),
		filepath.Join(basePath, ".claude", "commands"),
		filepath.Join(basePath, ".opencode", "commands"),
	}

	var allCommands []*resource.Resource
	var allErrors []DiscoveryError

	for _, dir := range priorityDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Directory doesn't exist, skip
		}

		// Recursively search within priority directory (starting at depth 0)
		// Use the priority directory as the basePath for relative path calculation
		commands, errors := searchCommandsInDirectory(dir, 0, dir)
		allCommands = append(allCommands, commands...)
		allErrors = append(allErrors, errors...)
	}

	return allCommands, allErrors
}

// searchCommandsInDirectory recursively searches for command .md files within a directory
// This is used for priority locations to find commands at any depth within that location
// basePath is used to calculate relative paths for commands
func searchCommandsInDirectory(dir string, depth int, basePath string) ([]*resource.Resource, []DiscoveryError) {
	if depth > maxDepth {
		return nil, nil
	}

	if logger != nil {
		logger.Debug("scanning directory for commands",
			"directory", dir,
			"depth", depth)
	}

	var commands []*resource.Resource
	var errors []DiscoveryError

	entries, err := os.ReadDir(dir)
	if err != nil {
		if logger != nil {
			logger.Error("failed to read directory",
				"directory", dir,
				"error", err)
		}
		errors = append(errors, DiscoveryError{
			Path:  dir,
			Error: fmt.Errorf("failed to read directory: %w", err),
		})
		return nil, errors
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		// Follow symlinks with os.Stat
		entryInfo, err := os.Stat(entryPath)
		if err != nil {
			continue
		}

		if entryInfo.IsDir() {
			// Recursively search subdirectories
			subCommands, subErrors := searchCommandsInDirectory(entryPath, depth+1, basePath)
			commands = append(commands, subCommands...)
			errors = append(errors, subErrors...)
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
			// Collect error instead of silently skipping
			if logger != nil {
				logger.Debug("failed to load command",
					"path", entryPath,
					"error", err)
			}
			errors = append(errors, DiscoveryError{
				Path:  entryPath,
				Error: err,
			})
			continue
		}

		if logger != nil {
			logger.Debug("found command",
				"name", cmd.Name,
				"path", entryPath)
		}
		commands = append(commands, cmd)
	}

	return commands, errors
}

// recursiveSearchCommands performs a recursive search for command files
// basePath is used to calculate relative paths
// This function ONLY looks for commands/ directories and delegates to searchCommandsInDirectory
// It does NOT parse files directly - this prevents false positives from non-resource directories
func recursiveSearchCommands(currentPath string, depth int, basePath string) ([]*resource.Resource, []DiscoveryError) {
	if depth > maxDepth {
		return nil, nil
	}

	var allCommands []*resource.Resource
	var allErrors []DiscoveryError

	// Recursively search subdirectories (looking for commands/ directories)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		allErrors = append(allErrors, DiscoveryError{
			Path:  currentPath,
			Error: fmt.Errorf("failed to read directory: %w", err),
		})
		return allCommands, allErrors
	}

	for _, entry := range entries {
		entryPath := filepath.Join(currentPath, entry.Name())

		// Follow symlinks with os.Stat
		entryInfo, err := os.Stat(entryPath)
		if err != nil || !entryInfo.IsDir() {
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

		// Skip common non-resource directories
		if shouldSkipNonResourceDirectory(entry.Name()) {
			continue
		}

		// Handle 'commands' directories specially
		if entry.Name() == "commands" {
			// Skip root-level 'commands' (already handled by priority search)
			if depth == 0 {
				continue
			}
			// For nested 'commands' directories (depth > 0):
			// Search them recursively using the 'commands' directory as basePath
			// (similar to how priority search works)
			cmdDirCommands, cmdDirErrors := searchCommandsInDirectory(entryPath, 0, entryPath)
			allCommands = append(allCommands, cmdDirCommands...)
			allErrors = append(allErrors, cmdDirErrors...)
			// Don't recurse into this 'commands' directory again with recursiveSearchCommands
			continue
		}

		// For non-'commands' directories, continue recursive search
		subCommands, subErrors := recursiveSearchCommands(entryPath, depth+1, basePath)
		allCommands = append(allCommands, subCommands...)
		allErrors = append(allErrors, subErrors...)
	}

	return allCommands, allErrors
}

// searchDirectory searches a single directory for command files
// basePath is used to calculate relative paths
// NOTE: This function is only called from within commands/ directories,
// so no path filtering is needed here
func searchDirectory(dir string, basePath string) ([]*resource.Resource, []DiscoveryError) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []DiscoveryError{{
			Path:  dir,
			Error: fmt.Errorf("failed to read directory: %w", err),
		}}
	}

	var commands []*resource.Resource
	var errors []DiscoveryError

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
			// Collect error instead of silently skipping
			errors = append(errors, DiscoveryError{
				Path:  filePath,
				Error: err,
			})
			continue
		}

		commands = append(commands, cmd)
	}

	return commands, errors
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

// isCommandsDirectory checks if a directory path IS a commands directory
// (i.e., ends with "/commands" or "/.claude/commands" or "/.opencode/commands")
func isCommandsDirectory(path string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check if the path ends with a commands directory name
	return strings.HasSuffix(normalizedPath, "/commands") ||
		strings.HasSuffix(normalizedPath, "/.claude/commands") ||
		strings.HasSuffix(normalizedPath, "/.opencode/commands") ||
		normalizedPath == "commands" ||
		strings.HasSuffix(normalizedPath, ".claude/commands") ||
		strings.HasSuffix(normalizedPath, ".opencode/commands")
}

// isInCommandsSubtree checks if a file path is within a commands/ directory subtree
// Returns true if the path contains /commands/ or /.claude/commands/ or /.opencode/commands/
func isInCommandsSubtree(path string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(path)

	// Check for commands/ directories
	return strings.Contains(normalizedPath, "/commands/") ||
		strings.Contains(normalizedPath, "/.claude/commands/") ||
		strings.Contains(normalizedPath, "/.opencode/commands/") ||
		isCommandsDirectory(path)
}

// shouldSkipNonResourceDirectory returns true if the directory should be skipped during recursive search
// These are common directories that typically don't contain resources
func shouldSkipNonResourceDirectory(name string) bool {
	skipDirs := []string{
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
