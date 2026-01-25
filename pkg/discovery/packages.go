package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// DiscoverPackages discovers package resources (*.package.json files) in a repository
// It searches in the packages/ subdirectory of the basePath
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

	// Look for packages/ subdirectory
	packagesDir := filepath.Join(searchPath, "packages")

	// Check if packages directory exists
	if _, err := os.Stat(packagesDir); err != nil {
		// packages/ directory doesn't exist, return empty slice (not an error)
		return []*resource.Package{}, nil
	}

	// Search for *.package.json files in packages/ directory
	packages, err := searchPackagesDirectory(packagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to search packages directory: %w", err)
	}

	// Deduplicate packages by name
	return deduplicatePackages(packages), nil
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
