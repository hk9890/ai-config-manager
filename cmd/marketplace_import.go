package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/marketplace"
	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var (
	importDryRun    bool
	importForce     bool
	importFilter    string
	importSourceURL string
)

// marketplaceImportCmd represents the marketplace import command
var marketplaceImportCmd = &cobra.Command{
	Use:   "import <path-or-url>",
	Short: "Import Claude marketplace configuration",
	Long: `Import a Claude marketplace.json file and generate packages.

This command parses a Claude marketplace.json file, discovers resources in each
plugin's source directory, and creates aimgr packages. All discovered resources
are imported into the repository.

Supported Sources:
  - Local paths: /path/to/marketplace.json
  - GitHub URLs: gh:owner/repo/path/to/marketplace.json (requires gh CLI)

Flags:
  --dry-run       Preview what would be imported without making changes
  --force         Overwrite existing packages and resources
  --filter        Import only plugins matching the pattern (glob pattern)
  --source-url    Override source URL for metadata (default: inferred from path)

Examples:
  # Import from local marketplace
  aimgr marketplace import ~/.claude-plugin/marketplace.json

  # Import from GitHub (requires gh CLI)
  aimgr marketplace import gh:anthropics/claude-code/.claude-plugin/marketplace.json

  # Preview without importing
  aimgr marketplace import marketplace.json --dry-run

  # Import specific plugins only
  aimgr marketplace import marketplace.json --filter "code-*"

  # Force overwrite existing packages
  aimgr marketplace import marketplace.json --force`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceImport,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceImportCmd)

	marketplaceImportCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview without importing")
	marketplaceImportCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Overwrite existing packages and resources")
	marketplaceImportCmd.Flags().StringVar(&importFilter, "filter", "", "Import only matching plugins (glob pattern)")
	marketplaceImportCmd.Flags().StringVar(&importSourceURL, "source-url", "", "Override source URL for metadata")
}

func runMarketplaceImport(cmd *cobra.Command, args []string) error {
	pathOrURL := args[0]

	// Resolve the marketplace file path
	marketplacePath, cleanupFunc, err := resolveMarketplacePath(pathOrURL)
	if err != nil {
		return err
	}
	if cleanupFunc != nil {
		defer cleanupFunc()
	}

	// Parse marketplace.json
	fmt.Printf("Parsing marketplace: %s\n", pathOrURL)
	marketplaceConfig, err := marketplace.ParseMarketplace(marketplacePath)
	if err != nil {
		return fmt.Errorf("failed to parse marketplace: %w", err)
	}

	// Get base path (directory containing marketplace.json)
	basePath := filepath.Dir(marketplacePath)

	// Filter plugins if requested
	filteredPlugins := marketplaceConfig.Plugins
	if importFilter != "" {
		filteredPlugins, err = filterPlugins(marketplaceConfig.Plugins, importFilter)
		if err != nil {
			return fmt.Errorf("invalid filter pattern: %w", err)
		}
		if len(filteredPlugins) == 0 {
			fmt.Printf("No plugins matched filter: %s\n", importFilter)
			return nil
		}
	}

	// Display marketplace info
	fmt.Printf("\nImporting marketplace: %s (%d plugin(s))\n", marketplaceConfig.Name, len(filteredPlugins))
	if marketplaceConfig.Description != "" {
		fmt.Printf("Description: %s\n", marketplaceConfig.Description)
	}

	// Create temporary marketplace config with filtered plugins for generation
	filteredMarketplace := &marketplace.MarketplaceConfig{
		Name:        marketplaceConfig.Name,
		Version:     marketplaceConfig.Version,
		Description: marketplaceConfig.Description,
		Owner:       marketplaceConfig.Owner,
		Plugins:     filteredPlugins,
	}

	// Generate packages
	fmt.Println("\nGenerating packages:")
	packageInfos, err := marketplace.GeneratePackages(filteredMarketplace, basePath)
	if err != nil {
		return fmt.Errorf("failed to generate packages: %w", err)
	}

	if len(packageInfos) == 0 {
		fmt.Println("No packages generated (plugins may have no resources)")
		return nil
	}

	// Initialize repository manager
	manager, err := repo.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create repository manager: %w", err)
	}

	if !importDryRun {
		if err := manager.Init(); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
	}

	// Track statistics
	stats := &importStats{
		packageCount:     0,
		resourcesAdded:   make(map[resource.ResourceType]int),
		resourcesSkipped: make(map[resource.ResourceType]int),
		errors:           []string{},
	}

	// Determine source URL for metadata
	sourceURL := importSourceURL
	if sourceURL == "" {
		sourceURL = inferSourceURL(pathOrURL, marketplacePath)
	}

	// Import packages and resources
	for _, pkgInfo := range packageInfos {
		pkg := pkgInfo.Package
		pluginSourcePath := pkgInfo.SourcePath

		if err := importPackage(manager, pkg, pluginSourcePath, sourceURL, stats); err != nil {
			errorMsg := fmt.Sprintf("  ✗ %s: %v", pkg.Name, err)
			fmt.Println(errorMsg)
			stats.errors = append(stats.errors, errorMsg)
			if !importForce {
				return err
			}
			continue
		}

		// Print success
		resourceCount := len(pkg.Resources)
		fmt.Printf("  ✓ %s (%d resource(s))\n", pkg.Name, resourceCount)
		stats.packageCount++
	}

	// Print summary
	printImportSummary(stats)

	return nil
}

// resolveMarketplacePath resolves a path or URL to a local file path.
// Returns the local path and a cleanup function (for temporary files).
func resolveMarketplacePath(pathOrURL string) (string, func(), error) {
	// Check if it's a GitHub URL (gh:owner/repo/path)
	if strings.HasPrefix(pathOrURL, "gh:") {
		// TODO: Implement GitHub fetching using gh CLI
		// For now, return an error
		return "", nil, fmt.Errorf("GitHub URLs are not yet supported (coming soon)")
	}

	// Treat as local path
	absPath, err := filepath.Abs(pathOrURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("marketplace file not found: %s", absPath)
		}
		return "", nil, fmt.Errorf("failed to access marketplace file: %w", err)
	}

	return absPath, nil, nil
}

// filterPlugins filters plugins by name pattern using glob matching
func filterPlugins(plugins []marketplace.Plugin, pattern string) ([]marketplace.Plugin, error) {
	var filtered []marketplace.Plugin

	for _, plugin := range plugins {
		matched, err := filepath.Match(pattern, plugin.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %w", err)
		}
		if matched {
			filtered = append(filtered, plugin)
		}
	}

	return filtered, nil
}

// importPackage imports a single package and its resources.
// pluginSourcePath is the absolute path to the plugin's source directory where resources are located.
func importPackage(manager *repo.Manager, pkg *resource.Package, pluginSourcePath, sourceURL string, stats *importStats) error {
	// Check if package already exists
	packagePath := resource.GetPackagePath(pkg.Name, manager.GetRepoPath())
	_, err := os.Stat(packagePath)
	packageExists := err == nil

	if packageExists && !importForce {
		if !importDryRun {
			return fmt.Errorf("package already exists (use --force to overwrite)")
		}
		// In dry-run mode, just note it exists
		fmt.Printf("  ⊙ %s (already exists, would skip)\n", pkg.Name)
		return nil
	}

	// For each resource reference, discover and import the actual resource
	discoveredResources := []string{}
	for _, ref := range pkg.Resources {
		// Parse resource reference
		resType, resName, err := resource.ParseResourceReference(ref)
		if err != nil {
			return fmt.Errorf("invalid resource reference %q: %w", ref, err)
		}

		// Import the resource
		if err := importResource(manager, resType, resName, pluginSourcePath, sourceURL, stats); err != nil {
			if !importForce {
				return err
			}
			// In force mode, continue with other resources
			continue
		}

		discoveredResources = append(discoveredResources, ref)
	}

	// Update package with successfully imported resources
	pkg.Resources = discoveredResources

	if importDryRun {
		return nil
	}

	// Remove existing package if force mode
	if packageExists && importForce {
		if err := os.Remove(packagePath); err != nil {
			return fmt.Errorf("failed to remove existing package: %w", err)
		}
	}

	// Save package
	if err := resource.SavePackage(pkg, manager.GetRepoPath()); err != nil {
		return fmt.Errorf("failed to save package: %w", err)
	}

	// Save package metadata
	now := time.Now()
	pkgMetadata := &metadata.PackageMetadata{
		Name:           pkg.Name,
		SourceType:     "marketplace",
		SourceURL:      sourceURL,
		FirstAdded:     now,
		LastUpdated:    now,
		ResourceCount:  len(pkg.Resources),
		OriginalFormat: "claude-plugin",
	}

	if err := metadata.SavePackageMetadata(pkgMetadata, manager.GetRepoPath()); err != nil {
		return fmt.Errorf("failed to save package metadata: %w", err)
	}

	return nil
}

// importResource imports a single resource into the repository.
// pluginSourcePath is the absolute path to the plugin's source directory where the resource is located.
func importResource(manager *repo.Manager, resType resource.ResourceType, resName string, pluginSourcePath, sourceURL string, stats *importStats) error {
	// Check if resource already exists
	resourcePath := manager.GetPath(resName, resType)
	_, err := os.Stat(resourcePath)
	resourceExists := err == nil

	if resourceExists {
		if importForce {
			// Remove existing resource
			if !importDryRun {
				if err := manager.Remove(resName, resType); err != nil {
					return fmt.Errorf("failed to remove existing resource %s/%s: %w", resType, resName, err)
				}
			}
		} else {
			// Skip existing resource
			stats.resourcesSkipped[resType]++
			return nil
		}
	}

	// Find the resource source file in plugin source directory
	sourcePath, err := findResourceSource(pluginSourcePath, resType, resName)
	if err != nil {
		return fmt.Errorf("resource %s/%s not found in plugin source: %w", resType, resName, err)
	}

	if importDryRun {
		stats.resourcesAdded[resType]++
		return nil
	}

	// Import the resource
	switch resType {
	case resource.Command:
		err = manager.AddCommand(sourcePath, sourceURL, "marketplace")
	case resource.Skill:
		err = manager.AddSkill(sourcePath, sourceURL, "marketplace")
	case resource.Agent:
		err = manager.AddAgent(sourcePath, sourceURL, "marketplace")
	default:
		return fmt.Errorf("unsupported resource type: %s", resType)
	}

	if err != nil {
		return fmt.Errorf("failed to import %s/%s: %w", resType, resName, err)
	}

	stats.resourcesAdded[resType]++
	return nil
}

// findResourceSource finds the source file for a resource in the plugin source directory.
// pluginSourcePath is the absolute path to the plugin's source directory.
func findResourceSource(pluginSourcePath string, resType resource.ResourceType, resName string) (string, error) {
	var candidatePaths []string

	switch resType {
	case resource.Command:
		// Check standard locations
		candidatePaths = []string{
			filepath.Join(pluginSourcePath, "commands", resName+".md"),
			filepath.Join(pluginSourcePath, ".claude", "commands", resName+".md"),
			filepath.Join(pluginSourcePath, ".opencode", "commands", resName+".md"),
		}
	case resource.Skill:
		// Check standard locations
		candidatePaths = []string{
			filepath.Join(pluginSourcePath, "skills", resName, "SKILL.md"),
			filepath.Join(pluginSourcePath, ".claude", "skills", resName, "SKILL.md"),
			filepath.Join(pluginSourcePath, ".opencode", "skills", resName, "SKILL.md"),
		}
	case resource.Agent:
		// Check standard locations
		candidatePaths = []string{
			filepath.Join(pluginSourcePath, "agents", resName+".md"),
			filepath.Join(pluginSourcePath, ".claude", "agents", resName+".md"),
			filepath.Join(pluginSourcePath, ".opencode", "agents", resName+".md"),
		}
	}

	// Try each candidate path
	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("source file not found for %s/%s", resType, resName)
}

// inferSourceURL infers the source URL from the path or URL
func inferSourceURL(pathOrURL, resolvedPath string) string {
	// If it's a GitHub URL, return it as-is
	if strings.HasPrefix(pathOrURL, "gh:") {
		return pathOrURL
	}

	// For local paths, use file:// URL
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		absPath = resolvedPath
	}
	return "file://" + absPath
}

// importStats tracks import statistics
type importStats struct {
	packageCount     int
	resourcesAdded   map[resource.ResourceType]int
	resourcesSkipped map[resource.ResourceType]int
	errors           []string
}

// printImportSummary prints a summary of the import operation
func printImportSummary(stats *importStats) {
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Import Summary:")
	fmt.Printf("  Packages: %d created\n", stats.packageCount)

	totalAdded := 0
	for _, count := range stats.resourcesAdded {
		totalAdded += count
	}

	totalSkipped := 0
	for _, count := range stats.resourcesSkipped {
		totalSkipped += count
	}

	if totalAdded > 0 {
		fmt.Printf("  Resources: %d imported", totalAdded)
		details := []string{}
		if stats.resourcesAdded[resource.Command] > 0 {
			details = append(details, fmt.Sprintf("%d command(s)", stats.resourcesAdded[resource.Command]))
		}
		if stats.resourcesAdded[resource.Skill] > 0 {
			details = append(details, fmt.Sprintf("%d skill(s)", stats.resourcesAdded[resource.Skill]))
		}
		if stats.resourcesAdded[resource.Agent] > 0 {
			details = append(details, fmt.Sprintf("%d agent(s)", stats.resourcesAdded[resource.Agent]))
		}
		if len(details) > 0 {
			fmt.Printf(" (%s)", strings.Join(details, ", "))
		}
		fmt.Println()
	}

	if totalSkipped > 0 {
		fmt.Printf("  Skipped: %d (already exist)\n", totalSkipped)
	}

	if len(stats.errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(stats.errors))
	}

	if importDryRun {
		fmt.Println("\n✓ Dry run complete - no changes made")
	}
}
