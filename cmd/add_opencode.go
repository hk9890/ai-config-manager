package cmd

import (
	"fmt"
	"os"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// addOpenCodeCmd represents the add opencode subcommand
var addOpenCodeCmd = &cobra.Command{
	Use:   "opencode <path>",
	Short: "Add all resources from an OpenCode folder",
	Long: `Add all resources (commands, skills, and agents) from an OpenCode configuration folder.

An OpenCode folder is a directory named .opencode or containing .opencode/,
with commands/, skills/, and/or agents/ subdirectories.

Example:
  ai-repo add opencode ~/.opencode
  ai-repo add opencode ./my-project/.opencode --force
  ai-repo add opencode ./opencode-config --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opencodePath := args[0]

		// Validate path exists
		if _, err := os.Stat(opencodePath); err != nil {
			return fmt.Errorf("path does not exist: %s", opencodePath)
		}

		// Validate it's an OpenCode folder
		isOpenCodeFolder, err := resource.DetectOpenCodeFolder(opencodePath)
		if err != nil {
			return err
		}
		if !isOpenCodeFolder {
			return fmt.Errorf("path is not a valid OpenCode folder: %s\nMust be named .opencode or contain commands/, skills/, or agents/ subdirectories", opencodePath)
		}

		// Scan for resources
		contents, err := resource.ScanOpenCodeFolder(opencodePath)
		if err != nil {
			return fmt.Errorf("failed to scan OpenCode folder: %w", err)
		}

		// Check if folder has any resources
		totalResources := len(contents.CommandPaths) + len(contents.SkillPaths) + len(contents.AgentPaths)
		if totalResources == 0 {
			fmt.Printf("⚠ OpenCode folder contains no commands, skills, or agents to import\n")
			return nil
		}

		// Print header
		fmt.Printf("Importing from OpenCode folder: %s\n", opencodePath)
		if dryRunFlag {
			fmt.Println("  Mode: DRY RUN (preview only)")
		}
		fmt.Println()
		fmt.Printf("Found: %d commands, %d skills, %d agents\n\n",
			len(contents.CommandPaths), len(contents.SkillPaths), len(contents.AgentPaths))

		// Combine all resource paths
		allPaths := append([]string{}, contents.CommandPaths...)
		allPaths = append(allPaths, contents.SkillPaths...)
		allPaths = append(allPaths, contents.AgentPaths...)

		// Create manager and import
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

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
	},
}

func init() {
	addCmd.AddCommand(addOpenCodeCmd)

	// Add flags
	addOpenCodeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resources")
	addOpenCodeCmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	addOpenCodeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without importing")
}
