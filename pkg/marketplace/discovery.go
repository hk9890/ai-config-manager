package marketplace

import (
	"os"
	"path/filepath"
)

// DiscoverMarketplace searches for marketplace.json files in common locations
// and returns the parsed marketplace configuration along with the file path.
//
// It searches in the following locations (relative to basePath/subpath):
// 1. .claude-plugin/marketplace.json
// 2. marketplace.json (root)
// 3. .opencode/marketplace.json
//
// Returns:
// - *MarketplaceConfig: The parsed marketplace configuration (nil if not found)
// - string: The absolute path to the marketplace.json file (empty if not found)
// - error: Any error during parsing (nil if not found or successful)
func DiscoverMarketplace(basePath string, subpath string) (*MarketplaceConfig, string, error) {
	searchPath := basePath
	if subpath != "" {
		searchPath = filepath.Join(basePath, subpath)
	}

	// Verify search path exists
	if _, err := os.Stat(searchPath); err != nil {
		// Path doesn't exist, return nil (not an error)
		return nil, "", nil
	}

	// Common locations to search for marketplace.json
	candidatePaths := []string{
		filepath.Join(searchPath, ".claude-plugin", "marketplace.json"),
		filepath.Join(searchPath, "marketplace.json"),
		filepath.Join(searchPath, ".opencode", "marketplace.json"),
	}

	// Try each candidate path
	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to parse it
			config, err := ParseMarketplace(path)
			if err != nil {
				return nil, path, err
			}
			return config, path, nil
		}
	}

	// No marketplace.json found (not an error)
	return nil, "", nil
}
