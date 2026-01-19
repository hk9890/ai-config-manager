package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/repo"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var (
	skipExistingFlag bool
	dryRunFlag       bool
)

// addPluginCmd represents the add plugin subcommand
var addPluginCmd = &cobra.Command{
	Use:   "plugin <path>",
	Short: "Add all resources from a Claude plugin",
	Long: `Add all resources (commands and skills) from a Claude plugin directory.

A Claude plugin is a directory containing .claude-plugin/plugin.json with
commands/ and/or skills/ subdirectories.

Example:
  ai-repo add plugin ~/.claude/plugins/marketplaces/claude-plugins-official/plugins/example-plugin
  ai-repo add plugin ./my-plugin --force
  ai-repo add plugin ./plugin --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginPath := args[0]

		// Validate path exists
		if _, err := os.Stat(pluginPath); err != nil {
			return fmt.Errorf("path does not exist: %s", pluginPath)
		}

		// Validate it's a plugin
		isPlugin, err := resource.DetectPlugin(pluginPath)
		if err != nil {
			return err
		}
		if !isPlugin {
			return fmt.Errorf("path is not a valid Claude plugin: %s\nMust contain .claude-plugin/plugin.json", pluginPath)
		}

		// Load plugin metadata
		metadata, err := resource.LoadPluginMetadata(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin metadata: %w", err)
		}

		// Scan for resources
		commandPaths, skillPaths, err := resource.ScanPluginResources(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to scan plugin resources: %w", err)
		}

		// Check if plugin has any resources
		if len(commandPaths) == 0 && len(skillPaths) == 0 {
			fmt.Printf("⚠ Plugin '%s' contains no commands or skills to import\n", metadata.Name)
			return nil
		}

		// Print header
		fmt.Printf("Importing from plugin: %s\n", metadata.Name)
		if metadata.Description != "" {
			fmt.Printf("  Description: %s\n", metadata.Description)
		}
		if dryRunFlag {
			fmt.Println("  Mode: DRY RUN (preview only)")
		}
		fmt.Println()
		fmt.Printf("Found: %d commands, %d skills\n\n", len(commandPaths), len(skillPaths))

		// Combine all resource paths
		allPaths := append(commandPaths, skillPaths...)

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
	addCmd.AddCommand(addPluginCmd)

	// Add flags
	addPluginCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resources")
	addPluginCmd.Flags().BoolVar(&skipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	addPluginCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview without importing")
}

// printImportResults prints a formatted summary of import results
func printImportResults(result *repo.BulkImportResult) {
	// Print added resources
	for _, path := range result.Added {
		resourceType := "resource"
		name := filepath.Base(path)
		
		// Determine type
		if filepath.Ext(path) == ".md" {
			resourceType = "command"
			name = name[:len(name)-3] // Remove .md
		} else {
			resourceType = "skill"
		}
		
		fmt.Printf("✓ Added %s '%s'\n", resourceType, name)
	}

	// Print skipped resources
	for _, path := range result.Skipped {
		name := filepath.Base(path)
		if filepath.Ext(path) == ".md" {
			name = name[:len(name)-3]
		}
		fmt.Printf("⊘ Skipped '%s' (already exists)\n", name)
	}

	// Print failed resources
	for _, fail := range result.Failed {
		name := filepath.Base(fail.Path)
		if filepath.Ext(fail.Path) == ".md" {
			name = name[:len(name)-3]
		}
		fmt.Printf("✗ Failed '%s': %s\n", name, fail.Message)
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %d added, %d skipped, %d failed\n",
		len(result.Added), len(result.Skipped), len(result.Failed))
}
