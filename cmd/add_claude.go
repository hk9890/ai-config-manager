package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/repo"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// addClaudeCmd represents the add claude subcommand
var addClaudeCmd = &cobra.Command{
	Use:   "claude <path>",
	Short: "Add all resources from a Claude folder (.claude/)",
	Long: `Add all resources (commands, skills, and agents) from a Claude configuration folder.

A Claude folder contains .claude/commands/, .claude/skills/, and/or .claude/agents/ subdirectories.
You can specify either the .claude/ directory itself or its parent directory.

Example:
  ai-repo add claude ~/.claude
  ai-repo add claude ~/my-project/.claude
  ai-repo add claude ~/my-project --force
  ai-repo add claude ./.claude --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		claudePath := args[0]

		// Validate path exists
		if _, err := os.Stat(claudePath); err != nil {
			return fmt.Errorf("path does not exist: %s", claudePath)
		}

		// Validate it's a Claude folder
		isClaudeFolder, err := resource.DetectClaudeFolder(claudePath)
		if err != nil {
			return err
		}
		if !isClaudeFolder {
			return fmt.Errorf("path is not a valid Claude folder: %s\nMust be a .claude/ directory or contain commands/, skills/, or agents/ subdirectories", claudePath)
		}

		// Scan for resources
		contents, err := resource.ScanClaudeFolder(claudePath)
		if err != nil {
			return fmt.Errorf("failed to scan Claude folder: %w", err)
		}

		// Check if folder has any resources
		if len(contents.CommandPaths) == 0 && len(contents.SkillPaths) == 0 && len(contents.AgentPaths) == 0 {
			fmt.Printf("⚠ Claude folder contains no commands, skills, or agents to import\n")
			return nil
		}

		// Print header
		absPath, _ := filepath.Abs(claudePath)
		fmt.Printf("Importing from Claude folder: %s\n", absPath)
		if dryRunFlag {
			fmt.Println("  Mode: DRY RUN (preview only)")
		}
		fmt.Println()
		fmt.Printf("Found: %d commands, %d skills, %d agents\n\n", len(contents.CommandPaths), len(contents.SkillPaths), len(contents.AgentPaths))

		// Combine all resource paths
		allPaths := append(contents.CommandPaths, contents.SkillPaths...)
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
	addCmd.AddCommand(addClaudeCmd)

	// Add flags (reuse from add_plugin.go)
	addClaudeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resources")
	addClaudeCmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	addClaudeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without importing")
}
