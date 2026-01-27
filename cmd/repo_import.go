package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/marketplace"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var forceFlag bool

var (
	skipExistingFlag bool
	dryRunFlag       bool
	filterFlag       string
	importFormatFlag    string
)

// repoImportCmd represents the import command
var repoImportCmd = &cobra.Command{
	Use:   "import <source>",
	Short: "Import resources to the repository",
	Long: `Import resources to the aimgr repository.

This command auto-discovers and imports all resources (commands, skills, agents, packages)
from the specified source location. It also automatically detects and imports marketplace.json
files if present.

Source Formats:
  Local folders:
    ./path or ~/path           Local directory (relative or home)
    /absolute/path             Absolute local path
  
  GitHub repositories:
    gh:owner/repo              GitHub repository (shorthand)
    owner/repo                 GitHub shorthand (gh: inferred)
    https://github.com/...     Full HTTPS Git URL
    git@github.com:...         SSH Git URL
    gh:owner/repo@branch       Specific branch or tag

Commands are single .md files with YAML frontmatter.
Skills are directories containing a SKILL.md file.
Agents are single .md files with YAML frontmatter.
Packages are .package.json files in the packages/ directory.
Marketplace files (marketplace.json) are automatically discovered and converted to packages.

Examples:
  # Add all resources from local folder
  aimgr repo add ~/.opencode/
  aimgr repo add ~/project/.claude/
  aimgr repo add ./my-resources/

  # Add all resources from GitHub
  aimgr repo add https://github.com/owner/repo
  aimgr repo add git@github.com:owner/repo.git
  aimgr repo add gh:owner/repo
  aimgr repo add owner/repo
  
  # Add specific version
  aimgr repo add gh:owner/repo@v1.0.0
  
  # With options
  aimgr repo add ~/resources/ --force
  aimgr repo add gh:owner/repo --skip-existing
  aimgr repo add ./test/ --dry-run
  
  # Filter resources (pattern matching)
  aimgr repo add gh:owner/repo --filter "skill/*"
  aimgr repo add ./resources/ --filter "*test*"`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceInput := args[0]

		// Parse source
		parsed, err := source.ParseSource(sourceInput)
		if err != nil {
			return fmt.Errorf("invalid source format: %w", err)
		}

		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Handle different source types
		if parsed.Type == source.GitHub || parsed.Type == source.GitURL {
			return addBulkFromGitHub(parsed, manager)
		}

		// Local source
		return addBulkFromLocal(parsed.LocalPath, manager)
	},
}

func init() {
	repoCmd.AddCommand(repoImportCmd)

	// Add flags
	repoImportCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resources")
	repoImportCmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	repoImportCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without importing")
	repoImportCmd.Flags().StringVar(&filterFlag, "filter", "", "Filter resources by pattern (e.g., 'skill/*', '*test*')")
	repoImportCmd.Flags().StringVar(&importFormatFlag, "format", "table", "Output format: table, json, yaml")
}

// Helper functions for bulk add integration

// addSingleResource adds a single file (command or agent) to the repository
func addSingleResource(filePath string, manager *repo.Manager) error {
	// Validate it's a .md file
	if filepath.Ext(filePath) != ".md" {
		return fmt.Errorf("file must have .md extension: %s", filePath)
	}

	// Check parent directory to determine type
	parentDir := filepath.Base(filepath.Dir(filePath))

	// If in agents/ directory, treat as agent
	if parentDir == "agents" {
		agent, err := resource.LoadAgent(filePath)
		if err != nil {
			return fmt.Errorf("invalid agent resource: %w", err)
		}
		return addAgentFile(filePath, agent, manager)
	}

	// If in commands/ directory, treat as command
	if parentDir == "commands" {
		cmd, err := resource.LoadCommand(filePath)
		if err != nil {
			return fmt.Errorf("invalid command resource: %w", err)
		}
		return addCommandFile(filePath, cmd, manager)
	}

	// Otherwise, try to determine by content
	// Try as agent first (agents have more specific fields)
	agent, agentErr := resource.LoadAgent(filePath)
	if agentErr == nil {
		// Check if it has agent-specific fields (type, instructions, capabilities)
		agentRes, err := resource.LoadAgentResource(filePath)
		if err == nil && (agentRes.Type != "" || agentRes.Instructions != "" || len(agentRes.Capabilities) > 0) {
			// Has agent-specific fields, treat as agent
			return addAgentFile(filePath, agent, manager)
		}
	}

	// Try as command
	cmd, cmdErr := resource.LoadCommand(filePath)
	if cmdErr == nil {
		// It's a command
		return addCommandFile(filePath, cmd, manager)
	}

	// Neither worked, return both errors
	return fmt.Errorf("invalid resource file (tried as agent: %v; tried as command: %v)", agentErr, cmdErr)
}

// addCommandFile adds a single command file to the repository
func addCommandFile(filePath string, res *resource.Resource, manager *repo.Manager) error {
	// Check if already exists (if not force mode)
	if !forceFlag {
		existing, _ := manager.Get(res.Name, resource.Command)
		if existing != nil {
			return fmt.Errorf("command '%s' already exists in repository (use --force to overwrite)", res.Name)
		}
	} else {
		// Remove existing if force mode
		_ = manager.Remove(res.Name, resource.Command)
	}

	// Determine source info
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	sourceURL := "file://" + absPath

	// Add the command
	if err := manager.AddCommand(filePath, sourceURL, "file"); err != nil {
		return fmt.Errorf("failed to add command: %w", err)
	}

	// Success message
	fmt.Printf("✓ Added command '%s' to repository\n", res.Name)
	if res.Description != "" {
		fmt.Printf("  Description: %s\n", res.Description)
	}

	return nil
}

// addAgentFile adds a single agent file to the repository
func addAgentFile(filePath string, res *resource.Resource, manager *repo.Manager) error {
	// Check if already exists (if not force mode)
	if !forceFlag {
		existing, _ := manager.Get(res.Name, resource.Agent)
		if existing != nil {
			return fmt.Errorf("agent '%s' already exists in repository (use --force to overwrite)", res.Name)
		}
	} else {
		// Remove existing if force mode
		_ = manager.Remove(res.Name, resource.Agent)
	}

	// Determine source info
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	sourceURL := "file://" + absPath

	// Add the agent
	if err := manager.AddAgent(filePath, sourceURL, "file"); err != nil {
		return fmt.Errorf("failed to add agent: %w", err)
	}

	// Success message
	fmt.Printf("✓ Added agent '%s' to repository\n", res.Name)
	if res.Description != "" {
		fmt.Printf("  Description: %s\n", res.Description)
	}

	return nil
}

// applyFilter filters discovered resources based on a pattern.
// Returns filtered slices and a boolean indicating if filtering was applied.
func applyFilter(filterPattern string, commands, skills, agents []*resource.Resource, packages []*resource.Package) ([]*resource.Resource, []*resource.Resource, []*resource.Resource, []*resource.Package, error) {
	if filterPattern == "" {
		// No filter, return all resources
		return commands, skills, agents, packages, nil
	}

	// Parse the pattern
	matcher, err := pattern.NewMatcher(filterPattern)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// If not a pattern (exact name), try to match by name across all types
	if !matcher.IsPattern() {
		// Exact match - find the resource by name
		var filteredCommands []*resource.Resource
		var filteredSkills []*resource.Resource
		var filteredAgents []*resource.Resource
		var filteredPackages []*resource.Package

		// Get the resource type filter (if specified)
		resourceType := matcher.GetResourceType()

		// Check commands (if no type filter or type is command)
		if resourceType == "" || resourceType == resource.Command {
			for _, cmd := range commands {
				if cmd.Name == filterPattern || matcher.Match(cmd) {
					filteredCommands = append(filteredCommands, cmd)
				}
			}
		}

		// Check skills (if no type filter or type is skill)
		if resourceType == "" || resourceType == resource.Skill {
			for _, skill := range skills {
				if skill.Name == filterPattern || matcher.Match(skill) {
					filteredSkills = append(filteredSkills, skill)
				}
			}
		}

		// Check agents (if no type filter or type is agent)
		if resourceType == "" || resourceType == resource.Agent {
			for _, agent := range agents {
				if agent.Name == filterPattern || matcher.Match(agent) {
					filteredAgents = append(filteredAgents, agent)
				}
			}
		}

		// Check packages (if no type filter or type is package)
		if resourceType == "" || resourceType == resource.PackageType {
			for _, pkg := range packages {
				// Create a temporary Resource for matching
				tempRes := &resource.Resource{Type: resource.PackageType, Name: pkg.Name}
				if pkg.Name == filterPattern || matcher.Match(tempRes) {
					filteredPackages = append(filteredPackages, pkg)
				}
			}
		}

		return filteredCommands, filteredSkills, filteredAgents, filteredPackages, nil
	}

	// Pattern matching - filter each resource type
	var filteredCommands []*resource.Resource
	var filteredSkills []*resource.Resource
	var filteredAgents []*resource.Resource
	var filteredPackages []*resource.Package

	// Get the resource type filter (if specified)
	resourceType := matcher.GetResourceType()

	// Filter commands (if no type filter or type is command)
	if resourceType == "" || resourceType == resource.Command {
		for _, cmd := range commands {
			if matcher.Match(cmd) {
				filteredCommands = append(filteredCommands, cmd)
			}
		}
	}

	// Filter skills (if no type filter or type is skill)
	if resourceType == "" || resourceType == resource.Skill {
		for _, skill := range skills {
			if matcher.Match(skill) {
				filteredSkills = append(filteredSkills, skill)
			}
		}
	}

	// Filter agents (if no type filter or type is agent)
	if resourceType == "" || resourceType == resource.Agent {
		for _, agent := range agents {
			if matcher.Match(agent) {
				filteredAgents = append(filteredAgents, agent)
			}
		}
	}

	// Filter packages (if no type filter or type is package)
	if resourceType == "" || resourceType == resource.PackageType {
		for _, pkg := range packages {
			// Create a temporary Resource for matching
			tempRes := &resource.Resource{Type: resource.PackageType, Name: pkg.Name}
			if matcher.Match(tempRes) {
				filteredPackages = append(filteredPackages, pkg)
			}
		}
	}

	return filteredCommands, filteredSkills, filteredAgents, filteredPackages, nil
}

// addBulkFromLocal handles bulk add from a local folder or single file
func addBulkFromLocal(localPath string, manager *repo.Manager) error {
	return addBulkFromLocalWithFilter(localPath, manager, filterFlag)
}

// addBulkFromLocalWithFilter handles bulk add from a local folder or single file with a custom filter
func addBulkFromLocalWithFilter(localPath string, manager *repo.Manager, filter string) error {
	// Validate path exists
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("path does not exist: %s", localPath)
	}

	// Check if it's a file or directory
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	// If it's a single file, handle it specially
	if !info.IsDir() {
		return addSingleResource(localPath, manager)
	}

	// Discover all resources (with error collection)
	commands, err := discovery.DiscoverCommands(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	skills, skillErrors, err := discovery.DiscoverSkillsWithErrors(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	agents, err := discovery.DiscoverAgents(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	packages, err := discovery.DiscoverPackages(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover packages: %w", err)
	}

	// Collect all discovery errors
	var discoveryErrors []discovery.DiscoveryError
	discoveryErrors = append(discoveryErrors, skillErrors...)

	// Discover marketplace.json
	marketplaceConfig, marketplacePath, err := marketplace.DiscoverMarketplace(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to parse marketplace: %w", err)
	}

	// Check if any resources found
	totalResources := len(commands) + len(skills) + len(agents) + len(packages)
	if totalResources == 0 && marketplaceConfig == nil {
		return fmt.Errorf("no resources found in: %s\nExpected commands (*.md), skills (*/SKILL.md), agents (*.md), packages (*.package.json), or marketplace.json", localPath)
	}

	// Print header
	absPath, _ := filepath.Abs(localPath)
	fmt.Printf("Importing from: %s\n", absPath)
	if dryRunFlag {
		fmt.Println("  Mode: DRY RUN (preview only)")
	}
	if filter != "" {
		fmt.Printf("  Filter: %s\n", filter)
	}
	fmt.Println()

	// Store original counts for reporting
	origCommandCount := len(commands)
	origSkillCount := len(skills)
	origAgentCount := len(agents)
	origPackageCount := len(packages)

	// Apply filter if specified
	if filter != "" {
		var err error
		commands, skills, agents, packages, err = applyFilter(filter, commands, skills, agents, packages)
		if err != nil {
			return err
		}

		// Check if filter matched any resources
		filteredTotal := len(commands) + len(skills) + len(agents) + len(packages)
		if filteredTotal == 0 {
			fmt.Printf("⚠ Warning: Filter '%s' matched 0 resources (found %d total)\n\n", filter, totalResources)
			return nil
		}

		// Show filtered counts
		fmt.Printf("Found: %d commands, %d skills, %d agents, %d packages", origCommandCount, origSkillCount, origAgentCount, origPackageCount)
		if filteredTotal < totalResources {
			fmt.Printf(" (filtered to %d matching '%s')\n", filteredTotal, filter)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("Found: %d commands, %d skills, %d agents, %d packages\n", len(commands), len(skills), len(agents), len(packages))
	}

	// Display marketplace info if found
	var marketplacePackages []*marketplace.PackageInfo
	if marketplaceConfig != nil {
		relPath := strings.TrimPrefix(marketplacePath, absPath)
		if relPath == "" {
			relPath = "marketplace.json"
		} else {
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		}
		fmt.Printf("Found marketplace: %s (%d plugins)\n", relPath, len(marketplaceConfig.Plugins))

		// Generate packages from marketplace
		fmt.Println("\nGenerating packages from marketplace:")
		basePath := filepath.Dir(marketplacePath)
		marketplacePackages, err = marketplace.GeneratePackages(marketplaceConfig, basePath)
		if err != nil {
			return fmt.Errorf("failed to generate packages from marketplace: %w", err)
		}

		for _, pkgInfo := range marketplacePackages {
			fmt.Printf("  ✓ %s (%d resources)\n", pkgInfo.Package.Name, len(pkgInfo.Package.Resources))
		}
	}

	fmt.Println()

	// Collect all resource paths
	var allPaths []string

	// Add commands - use discovered paths directly
	for _, cmd := range commands {
		allPaths = append(allPaths, cmd.Path)
	}
	// Add skills - use discovered paths directly
	for _, skill := range skills {
		allPaths = append(allPaths, skill.Path)
	}
	// Add agents - use discovered paths directly
	for _, agent := range agents {
		allPaths = append(allPaths, agent.Path)
	}
	// Add packages - use discovered paths directly
	for _, pkg := range packages {
		allPaths = append(allPaths, pkg.Path)
	}

	// Add resources from marketplace-generated packages
	for _, pkgInfo := range marketplacePackages {
		// Import resources for this package
		for _, resRef := range pkgInfo.Package.Resources {
			resType, resName, err := resource.ParseResourceReference(resRef)
			if err != nil {
				continue // Skip invalid references
			}

			// Find the resource file in the plugin source directory
			resPath, err := findResourceInPath(pkgInfo.SourcePath, resType, resName)
			if err == nil {
				allPaths = append(allPaths, resPath)
			}
		}
	}

	// Import using bulk add
	opts := repo.BulkImportOptions{
		Force:        forceFlag,
		SkipExisting: skipExistingFlag,
		DryRun:       dryRunFlag,
	}

	result, err := manager.AddBulk(allPaths, opts)
	if err != nil && !skipExistingFlag {
		// Print partial results before error
		printImportResults(result)
		return err
	}

	// Save marketplace-generated packages if not in dry-run mode
	if !dryRunFlag {
		for _, pkgInfo := range marketplacePackages {
			// Save package to repository
			if err := resource.SavePackage(pkgInfo.Package, manager.GetRepoPath()); err != nil {
				fmt.Printf("⚠ Warning: Failed to save package %s: %v\n", pkgInfo.Package.Name, err)
				continue
			}
		}
	}

	// Print discovery errors if any
	if len(discoveryErrors) > 0 {
		printDiscoveryErrors(discoveryErrors)
		fmt.Println()
	}

	// Print results
	printImportResults(result)

	// Exit with error if there were failures
	if len(result.Failed) > 0 {
		return fmt.Errorf("failed to import %d resource(s)", len(result.Failed))
	}

	return nil
}

// addBulkFromGitHub handles bulk add from a GitHub repository
func addBulkFromGitHub(parsed *source.ParsedSource, manager *repo.Manager) error {
	return addBulkFromGitHubWithFilter(parsed, manager, filterFlag)
}

// addBulkFromGitHubWithFilter handles bulk add from a GitHub repository with a custom filter
func addBulkFromGitHubWithFilter(parsed *source.ParsedSource, manager *repo.Manager, filter string) error {
	// Clone repository to temp directory
	cloneURL, err := source.GetCloneURL(parsed)
	if err != nil {
		return fmt.Errorf("failed to get clone URL: %w", err)
	}

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(manager.GetRepoPath())
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	// Get or clone repository using workspace cache
	cachePath, err := workspaceManager.GetOrClone(cloneURL, parsed.Ref)
	if err != nil {
		return fmt.Errorf("failed to get cached repo: %w", err)
	}

	// Update cached repository to latest changes
	// This ensures repo sync always pulls the latest code from Git remotes
	if err := workspaceManager.Update(cloneURL, parsed.Ref); err != nil {
		// If update fails, log warning but continue with existing cache
		// This handles cases like network failures where cache is still usable
		fmt.Fprintf(os.Stderr, "warning: failed to update cached repo (using existing cache): %v\n", err)
	}

	// Determine search path
	searchPath := cachePath
	if parsed.Subpath != "" {
		searchPath = filepath.Join(cachePath, parsed.Subpath)
	}

	// Discover all resources (with error collection)
	commands, err := discovery.DiscoverCommands(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	skills, skillErrors, err := discovery.DiscoverSkillsWithErrors(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	agents, err := discovery.DiscoverAgents(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	packages, err := discovery.DiscoverPackages(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover packages: %w", err)
	}

	// Collect all discovery errors
	var discoveryErrors []discovery.DiscoveryError
	discoveryErrors = append(discoveryErrors, skillErrors...)

	// Discover marketplace.json
	marketplaceConfig, marketplacePath, err := marketplace.DiscoverMarketplace(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to parse marketplace: %w", err)
	}

	// Check if any resources found
	totalResources := len(commands) + len(skills) + len(agents) + len(packages)
	if totalResources == 0 && marketplaceConfig == nil {
		return fmt.Errorf("no resources found in repository: %s\nExpected commands (*.md), skills (*/SKILL.md), agents (*.md), packages (*.package.json), or marketplace.json", parsed.URL)
	}

	// Print header
	fmt.Printf("Importing from: %s\n", parsed.URL)
	if parsed.Ref != "" {
		fmt.Printf("  Branch/Tag: %s\n", parsed.Ref)
	}
	if parsed.Subpath != "" {
		fmt.Printf("  Subpath: %s\n", parsed.Subpath)
	}
	if dryRunFlag {
		fmt.Println("  Mode: DRY RUN (preview only)")
	}
	if filter != "" {
		fmt.Printf("  Filter: %s\n", filter)
	}
	fmt.Println()

	// Store original counts for reporting
	origCommandCount := len(commands)
	origSkillCount := len(skills)
	origAgentCount := len(agents)
	origPackageCount := len(packages)

	// Apply filter if specified
	if filter != "" {
		var err error
		commands, skills, agents, packages, err = applyFilter(filter, commands, skills, agents, packages)
		if err != nil {
			return err
		}

		// Check if filter matched any resources
		filteredTotal := len(commands) + len(skills) + len(agents) + len(packages)
		if filteredTotal == 0 {
			fmt.Printf("⚠ Warning: Filter '%s' matched 0 resources (found %d total)\n\n", filter, totalResources)
			return nil
		}

		// Show filtered counts
		fmt.Printf("Found: %d commands, %d skills, %d agents, %d packages", origCommandCount, origSkillCount, origAgentCount, origPackageCount)
		if filteredTotal < totalResources {
			fmt.Printf(" (filtered to %d matching '%s')\n", filteredTotal, filter)
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("Found: %d commands, %d skills, %d agents, %d packages\n", len(commands), len(skills), len(agents), len(packages))
	}

	// Display marketplace info if found
	var marketplacePackages []*marketplace.PackageInfo
	if marketplaceConfig != nil {
		relPath := strings.TrimPrefix(marketplacePath, searchPath)
		if relPath == "" {
			relPath = "marketplace.json"
		} else {
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		}
		fmt.Printf("Found marketplace: %s (%d plugins)\n", relPath, len(marketplaceConfig.Plugins))

		// Generate packages from marketplace
		fmt.Println("\nGenerating packages from marketplace:")
		basePath := filepath.Dir(marketplacePath)
		marketplacePackages, err = marketplace.GeneratePackages(marketplaceConfig, basePath)
		if err != nil {
			return fmt.Errorf("failed to generate packages from marketplace: %w", err)
		}

		for _, pkgInfo := range marketplacePackages {
			fmt.Printf("  ✓ %s (%d resources)\n", pkgInfo.Package.Name, len(pkgInfo.Package.Resources))
		}
	}

	fmt.Println()

	// Collect all resource paths
	var allPaths []string

	// Add commands
	for _, cmd := range commands {
		cmdPath, err := findCommandFile(searchPath, cmd.Name)
		if err == nil {
			allPaths = append(allPaths, cmdPath)
		}
	}

	// Add skills
	for _, skill := range skills {
		skillPath, err := findSkillDir(searchPath, skill.Name)
		if err == nil {
			allPaths = append(allPaths, skillPath)
		}
	}

	// Add agents
	for _, agent := range agents {
		agentPath, err := findAgentFile(searchPath, agent.Name)
		if err == nil {
			allPaths = append(allPaths, agentPath)
		}
	}

	// Add packages
	for _, pkg := range packages {
		pkgPath, err := findPackageFile(searchPath, pkg.Name)
		if err == nil {
			allPaths = append(allPaths, pkgPath)
		}
	}

	// Add resources from marketplace-generated packages
	for _, pkgInfo := range marketplacePackages {
		// Import resources for this package
		for _, resRef := range pkgInfo.Package.Resources {
			resType, resName, err := resource.ParseResourceReference(resRef)
			if err != nil {
				continue // Skip invalid references
			}

			// Find the resource file in the plugin source directory
			resPath, err := findResourceInPath(pkgInfo.SourcePath, resType, resName)
			if err == nil {
				allPaths = append(allPaths, resPath)
			}
		}
	}

	// Determine source type
	sourceType := "github"
	if parsed.Type == source.GitURL {
		sourceType = "git-url"
	}

	// Import using bulk add with original source URL and ref
	opts := repo.BulkImportOptions{
		Force:        forceFlag,
		SkipExisting: skipExistingFlag,
		DryRun:       dryRunFlag,
		SourceURL:    parsed.URL,
		SourceType:   sourceType,
		Ref:          parsed.Ref,
	}

	result, err := manager.AddBulk(allPaths, opts)
	if err != nil && !skipExistingFlag {
		// Print partial results before error
		printImportResults(result)
		return err
	}

	// Save marketplace-generated packages if not in dry-run mode
	if !dryRunFlag {
		for _, pkgInfo := range marketplacePackages {
			// Save package to repository
			if err := resource.SavePackage(pkgInfo.Package, manager.GetRepoPath()); err != nil {
				fmt.Printf("⚠ Warning: Failed to save package %s: %v\n", pkgInfo.Package.Name, err)
				continue
			}
		}
	}
	// Print discovery errors if any
	if len(discoveryErrors) > 0 {
		printDiscoveryErrors(discoveryErrors)
		fmt.Println()
	}

	// Print results
	printImportResults(result)

	// Exit with error if there were failures
	if len(result.Failed) > 0 {
		return fmt.Errorf("failed to import %d resource(s)", len(result.Failed))
	}

	return nil
}

// selectResource handles resource selection when multiple resources are found
func selectResource(resources []*resource.Resource, resourceType string) (*resource.Resource, error) {
	if len(resources) == 1 {
		// Single resource found, use it automatically
		return resources[0], nil
	}

	// Multiple resources found, prompt user to select
	// Print header
	fmt.Printf("Multiple %ss found in repository:\n\n", resourceType)

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("#", "Name", "Description")

	// Add rows
	for i, res := range resources {
		desc := res.Description
		if desc == "" {
			desc = "(no description)"
		}
		// Truncate description to 60 chars (same as list command)
		desc = truncateString(desc, 60)

		if err := table.Append(fmt.Sprintf("%d", i+1), res.Name, desc); err != nil {
			return nil, fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Render table
	if err := table.Render(); err != nil {
		return nil, fmt.Errorf("failed to render table: %w", err)
	}

	// Prompt for selection
	fmt.Printf("\nSelect a resource (1-%d, or 'q' to cancel): ", len(resources))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read selection: %w", err)
	}

	input = strings.TrimSpace(input)

	// Allow 'q' to cancel
	if input == "q" || input == "Q" {
		return nil, fmt.Errorf("selection cancelled by user")
	}

	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(resources) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return resources[selection-1], nil
}

// findCommandFile finds the actual file path for a command by name
func findCommandFile(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// Load and check if it's the command we're looking for
		cmd, err := resource.LoadCommand(path)
		if err != nil {
			return nil // Skip invalid commands
		}
		if cmd.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("command file not found for: %s", name)
	}
	return found, nil
}

// findSkillDir finds the directory path for a skill by name
func findSkillDir(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		// Check if this directory contains SKILL.md
		skillMdPath := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillMdPath); err != nil {
			return nil // No SKILL.md, continue
		}
		// Load and check if it's the skill we're looking for
		skill, err := resource.LoadSkill(path)
		if err != nil {
			return nil // Skip invalid skills
		}
		if skill.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("skill directory not found for: %s", name)
	}
	return found, nil
}

// findAgentFile finds the actual file path for an agent by name
func findAgentFile(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// Load and check if it's the agent we're looking for
		agent, err := resource.LoadAgent(path)
		if err != nil {
			return nil // Skip invalid agents
		}
		if agent.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("agent file not found for: %s", name)
	}
	return found, nil
}

// findPackageFile finds the actual file path for a package by name
func findPackageFile(searchPath, name string) (string, error) {
	// Use recursive search to find the package (packages can be nested in subdirectories)
	var foundPath string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a package file with matching name
		if strings.HasSuffix(path, ".package.json") {
			pkg, err := resource.LoadPackage(path)
			if err == nil && pkg.Name == name {
				foundPath = path
				// Use a sentinel error to stop the walk early
				return fmt.Errorf("found")
			}
		}

		return nil
	})

	// Check if we stopped because we found the package
	if err != nil && err.Error() == "found" {
		return foundPath, nil
	}

	// Check for other walk errors
	if err != nil && err.Error() != "found" {
		return "", fmt.Errorf("error searching for package: %w", err)
	}

	if foundPath == "" {
		return "", fmt.Errorf("package file not found for: %s", name)
	}

	return foundPath, nil
}
func findResourceInPath(sourcePath string, resType resource.ResourceType, resName string) (string, error) {
	var candidatePaths []string

	switch resType {
	case resource.Command:
		// Check standard locations
		candidatePaths = []string{
			filepath.Join(sourcePath, "commands", resName+".md"),
			filepath.Join(sourcePath, ".claude", "commands", resName+".md"),
			filepath.Join(sourcePath, ".opencode", "commands", resName+".md"),
		}
	case resource.Skill:
		// Check standard locations (skills use directory path, not SKILL.md path)
		candidatePaths = []string{
			filepath.Join(sourcePath, "skills", resName),
			filepath.Join(sourcePath, ".claude", "skills", resName),
			filepath.Join(sourcePath, ".opencode", "skills", resName),
		}
	case resource.Agent:
		// Check standard locations
		candidatePaths = []string{
			filepath.Join(sourcePath, "agents", resName+".md"),
			filepath.Join(sourcePath, ".claude", "agents", resName+".md"),
			filepath.Join(sourcePath, ".opencode", "agents", resName+".md"),
		}
	}

	// Try each candidate path
	for _, path := range candidatePaths {
		if resType == resource.Skill {
			// For skills, check if directory exists with SKILL.md
			skillMdPath := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillMdPath); err == nil {
				return path, nil
			}
		} else {
			// For commands and agents, check if file exists
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("source file not found for %s/%s", resType, resName)
}

// determineSourceInfo determines the source type and URL from a parsed source and local path
func determineSourceInfo(parsed *source.ParsedSource, localPath string) (string, string, error) {
	switch parsed.Type {
	case source.GitHub:
		// GitHub source: type="github", url="gh:owner/repo/path"
		return "github", formatGitHubShortURL(parsed), nil

	case source.GitURL:
		// Git URL: type="github", url=original URL
		return "github", parsed.URL, nil

	case source.Local:
		// Local source: need to determine if file or directory
		info, err := os.Stat(localPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to stat path: %w", err)
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Format as file:// URL
		fileURL := "file://" + absPath

		if info.IsDir() {
			return "local", fileURL, nil
		}
		return "file", fileURL, nil

	default:
		return "", "", fmt.Errorf("unsupported source type: %s", parsed.Type)
	}
}

// formatGitHubSourceInfo formats GitHub source info for metadata
func formatGitHubSourceInfo(parsed *source.ParsedSource) (string, string) {
	return "github", formatGitHubShortURL(parsed)
}

// formatGitHubShortURL formats a GitHub URL in the gh:owner/repo/path format
func formatGitHubShortURL(parsed *source.ParsedSource) string {
	// Extract owner/repo from the GitHub URL
	// URL format: https://github.com/owner/repo
	url := strings.TrimPrefix(parsed.URL, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")

	// Remove /tree/branch if present
	if strings.Contains(url, "/tree/") {
		parts := strings.SplitN(url, "/tree/", 2)
		url = parts[0]
	}

	// Build the gh: format URL
	result := "gh:" + url

	// Add subpath if present
	if parsed.Subpath != "" {
		result = result + "/" + parsed.Subpath
	}

	// Add ref if present
	if parsed.Ref != "" {
		result = result + "@" + parsed.Ref
	}

	return result
}

// printImportResults prints a formatted summary of import results
func printImportResults(result *repo.BulkImportResult) {
	// Parse format flag
	format, err := output.ParseFormat(importFormatFlag)
	if err != nil {
		// Fall back to default human-readable output if format is invalid
		fmt.Fprintf(os.Stderr, "Warning: %v, using table format\n", err)
		format = output.Table
	}

	// Use structured output formatter for all formats
	bulkResult := output.FromBulkImportResult(result)
	if err := output.FormatBulkResult(bulkResult, format); err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
	}
}

// printDiscoveryErrors prints discovery errors with helpful suggestions
func printDiscoveryErrors(errors []discovery.DiscoveryError) {
	if len(errors) == 0 {
		return
	}

	// Deduplicate errors by path (first error for each path wins)
	seen := make(map[string]bool)
	uniqueErrors := make([]discovery.DiscoveryError, 0, len(errors))
	for _, err := range errors {
		if !seen[err.Path] {
			seen[err.Path] = true
			uniqueErrors = append(uniqueErrors, err)
		}
	}

	fmt.Printf("⚠ Discovery Issues (%d):\n", len(uniqueErrors))
	fmt.Println()

	for _, err := range uniqueErrors {
		// Extract just the directory/file name for display
		shortPath := filepath.Base(err.Path)
		if filepath.Base(filepath.Dir(err.Path)) != "." {
			// Include parent directory for context
			shortPath = filepath.Join(filepath.Base(filepath.Dir(err.Path)), filepath.Base(err.Path))
		}

		fmt.Printf("  ✗ %s\n", shortPath)
		fmt.Printf("    Error: %v\n", err.Error)
		fmt.Println()
	}

	fmt.Println("  Tip: These resources were skipped due to validation errors.")
	fmt.Println("       Fix the issues above and re-run the import.")
}
