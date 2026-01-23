package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/spf13/cobra"
)

// addBulkCmd represents the bulk add command that auto-discovers all resources
var addBulkCmd = &cobra.Command{
	Use:   "bulk <folder|url>",
	Short: "Auto-discover and add all resources from a folder or URL",
	Long: `Auto-discover and add all resources (commands, skills, agents) from a folder or URL.

This command scans the source location for all supported resource types and adds them
to the repository in bulk.

Source Formats:
  Local folders:
    ./path or ~/path           Local directory (relative or home)
    /absolute/path             Absolute local path
  
  GitHub repositories:
    gh:owner/repo              GitHub repository
    gh:owner/repo@branch       Specific branch or tag
    owner/repo                 GitHub shorthand (gh: inferred)
    https://github.com/...     Full HTTPS Git URL
    git@github.com:...         SSH Git URL

Examples:
  # From local folder (discovers all commands, skills, agents)
  aimgr repo add bulk ~/.opencode/
  aimgr repo add bulk ~/project/.claude/
  aimgr repo add bulk ./my-resources/
  
  # From GitHub
  aimgr repo add bulk gh:owner/repo
  aimgr repo add bulk owner/repo
  aimgr repo add bulk https://github.com/owner/repo
  
  # With options
  aimgr repo add bulk ~/resources/ --force
  aimgr repo add bulk gh:owner/repo --skip-existing
  aimgr repo add bulk ./test/ --dry-run`,
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
	addCmd.AddCommand(addBulkCmd)

	// Add flags
	addBulkCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resources")
	addBulkCmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	addBulkCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without importing")
}

// addBulkFromLocal handles bulk add from a local folder
func addBulkFromLocal(localPath string, manager *repo.Manager) error {
	// Validate path exists
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("path does not exist: %s", localPath)
	}

	// Validate it's a directory
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", localPath)
	}

	// Discover all resources
	commands, err := discovery.DiscoverCommands(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	skills, err := discovery.DiscoverSkills(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	agents, err := discovery.DiscoverAgents(localPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	// Check if any resources found
	totalResources := len(commands) + len(skills) + len(agents)
	if totalResources == 0 {
		return fmt.Errorf("no resources found in: %s\nExpected commands (*.md), skills (*/SKILL.md), or agents (*.md)", localPath)
	}

	// Print header
	absPath, _ := filepath.Abs(localPath)
	fmt.Printf("Importing from: %s\n", absPath)
	if dryRunFlag {
		fmt.Println("  Mode: DRY RUN (preview only)")
	}
	fmt.Println()
	fmt.Printf("Found: %d commands, %d skills, %d agents\n\n", len(commands), len(skills), len(agents))

	// Collect all resource paths
	var allPaths []string
	
	// Add commands
	for _, cmd := range commands {
		cmdPath, err := findCommandFile(localPath, cmd.Name)
		if err == nil {
			allPaths = append(allPaths, cmdPath)
		}
	}
	
	// Add skills
	for _, skill := range skills {
		skillPath, err := findSkillDir(localPath, skill.Name)
		if err == nil {
			allPaths = append(allPaths, skillPath)
		}
	}
	
	// Add agents
	for _, agent := range agents {
		agentPath, err := findAgentFile(localPath, agent.Name)
		if err == nil {
			allPaths = append(allPaths, agentPath)
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

	// Print results
	printImportResults(result)

	return nil
}

// addBulkFromGitHub handles bulk add from a GitHub repository
func addBulkFromGitHub(parsed *source.ParsedSource, manager *repo.Manager) error {
	// Clone repository to temp directory
	cloneURL, err := source.GetCloneURL(parsed)
	if err != nil {
		return fmt.Errorf("failed to get clone URL: %w", err)
	}

	tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	defer source.CleanupTempDir(tempDir)

	// Determine search path
	searchPath := tempDir
	if parsed.Subpath != "" {
		searchPath = filepath.Join(tempDir, parsed.Subpath)
	}

	// Discover all resources
	commands, err := discovery.DiscoverCommands(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	skills, err := discovery.DiscoverSkills(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	agents, err := discovery.DiscoverAgents(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	// Check if any resources found
	totalResources := len(commands) + len(skills) + len(agents)
	if totalResources == 0 {
		return fmt.Errorf("no resources found in repository: %s", parsed.URL)
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
	fmt.Println()
	fmt.Printf("Found: %d commands, %d skills, %d agents\n\n", len(commands), len(skills), len(agents))

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

	// Print results
	printImportResults(result)

	return nil
}
