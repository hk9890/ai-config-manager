package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

const maxPackageDepth = 5

// DiscoverPackages discovers package resources (*.package.json files) in a repository
// It searches in priority locations first, then falls back to recursive search
func DiscoverPackages(basePath string, subpath string) ([]*resource.Package, error) {
	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Verify search path exists
	searchPathInfo, err := os.Stat(searchPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	if !searchPathInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", searchPath)
	}

	// Try priority locations first
	packages, err := searchPriorityPackageLocations(searchPath)
	if err == nil && len(packages) > 0 {
		return deduplicatePackages(packages), nil
	}

	// Fall back to recursive search (max depth 5)
	packages, err = recursiveSearchPackages(searchPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search for packages: %w", err)
	}

	return deduplicatePackages(packages), nil
}

// searchPriorityPackageLocations searches standard package directories
func searchPriorityPackageLocations(basePath string) ([]*resource.Package, error) {
	priorityDirs := []string{
		filepath.Join(basePath, "packages"),
		filepath.Join(basePath, ".claude", "packages"),
		filepath.Join(basePath, ".opencode", "packages"),
	}

	var allPackages []*resource.Package

	for _, dir := range priorityDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Directory doesn't exist, skip
		}

		packages, err := searchPackagesDirectory(dir)
		if err != nil {
			// Continue searching other directories even if one fails
			continue
		}

		allPackages = append(allPackages, packages...)
	}

	return allPackages, nil
}

// recursiveSearchPackages performs a recursive search for package files
func recursiveSearchPackages(basePath string, depth int) ([]*resource.Package, error) {
	if depth > maxPackageDepth {
		return nil, nil
	}

	var allPackages []*resource.Package

	// Search current directory for *.package.json files
	packages, err := searchPackagesDirectory(basePath)
	if err == nil {
		allPackages = append(allPackages, packages...)
	}

	// Recursively search subdirectories
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(basePath, entry.Name())

		// Follow symlinks with os.Stat
		entryInfo, err := os.Stat(entryPath)
		if err != nil || !entryInfo.IsDir() {
			continue
		}

		// Skip hidden directories (except .claude and .opencode which are handled in priority)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip other resource type directories
		if entry.Name() == "agents" || entry.Name() == "skills" || entry.Name() == "commands" {
			continue
		}

		subPackages, err := recursiveSearchPackages(entryPath, depth+1)
		if err == nil {
			allPackages = append(allPackages, subPackages...)
		}
	}

	return allPackages, nil
}

// searchPackagesDirectory searches a directory for package files
func searchPackagesDirectory(dir string) ([]*resource.Package, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var packages []*resource.Package

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a .package.json file
		if !strings.HasSuffix(entry.Name(), ".package.json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())

		// Try to load as package
		pkg, err := resource.LoadPackage(filePath)
		if err != nil {
			// Invalid package file, skip it
			continue
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

// deduplicatePackages removes duplicate packages by name, keeping the first occurrence
func deduplicatePackages(packages []*resource.Package) []*resource.Package {
	seen := make(map[string]bool)
	var unique []*resource.Package

	for _, pkg := range packages {
		if !seen[pkg.Name] {
			seen[pkg.Name] = true
			unique = append(unique, pkg)
		}
	}

	return unique
}
