package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/spf13/cobra"
)

var (
	migrateDryRunFlag bool
	migrateForceFlag  bool
)

// repoMigrateMetadataCmd represents the migrate-metadata command
var repoMigrateMetadataCmd = &cobra.Command{
	Use:   "migrate-metadata",
	Short: "Migrate metadata files to .metadata/ directory structure",
	Long: `Migrate existing metadata files to the new .metadata/ directory structure.

This command migrates metadata files from the old layout:
  /repo/<type>s/<type>-<name>-metadata.json

To the new layout:
  /repo/.metadata/<type>s/<name>-metadata.json

The migration:
- Moves metadata files to the new location
- Removes the type prefix from filenames
- Creates .metadata/ directory structure
- Preserves all metadata content
- Skips files already at the new location

Examples:
  aimgr repo migrate-metadata              # Run migration with confirmation prompt
  aimgr repo migrate-metadata --dry-run    # Preview changes without moving files
  aimgr repo migrate-metadata --force      # Skip confirmation prompt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get repository manager
		manager, err := repo.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}

		repoPath := manager.GetRepoPath()

		// Dry run mode - preview message
		if migrateDryRunFlag {
			fmt.Println("=== DRY RUN MODE ===")
			fmt.Println("The following changes would be made:")
			fmt.Println()
		}

		// Run migration
		result, err := runMigration(repoPath, migrateDryRunFlag)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		// Display results
		displayMigrationResults(result, migrateDryRunFlag)

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoMigrateMetadataCmd)
	repoMigrateMetadataCmd.Flags().BoolVar(&migrateDryRunFlag, "dry-run", false, "Preview changes without moving files")
	repoMigrateMetadataCmd.Flags().BoolVar(&migrateForceFlag, "force", false, "Skip confirmation prompt")
}

// runMigration executes the migration with optional confirmation
func runMigration(repoPath string, dryRun bool) (*metadata.MigrationResult, error) {
	// Check if confirmation is needed (not in dry-run and not forced)
	if !dryRun && !migrateForceFlag {
		confirmed, err := confirmMigration()
		if err != nil {
			return nil, err
		}
		if !confirmed {
			fmt.Println("Migration cancelled.")
			return &metadata.MigrationResult{}, nil
		}
	}

	// Run the migration
	if dryRun {
		fmt.Println("Scanning for metadata files to migrate...")
		fmt.Println()
	} else {
		fmt.Println("Starting metadata migration...")
		fmt.Println()
	}

	opts := metadata.MigrationOptions{
		DryRun: dryRun,
	}

	result, err := metadata.MigrateMetadataFilesWithOptions(repoPath, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// confirmMigration prompts the user for confirmation
func confirmMigration() (bool, error) {
	fmt.Println("This will migrate metadata files to the new .metadata/ directory structure.")
	fmt.Print("Continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// displayMigrationResults displays the results of the migration
func displayMigrationResults(result *metadata.MigrationResult, dryRun bool) {
	fmt.Println()
	fmt.Println("=== Migration Summary ===")

	if dryRun {
		fmt.Printf("Files found:        %d\n", result.TotalFiles)
		fmt.Printf("Would be moved:     %d\n", result.MovedFiles)
		fmt.Printf("Would be skipped:   %d\n", result.SkippedFiles)
	} else {
		fmt.Printf("Files found:        %d\n", result.TotalFiles)
		fmt.Printf("Files moved:        %d\n", result.MovedFiles)
		fmt.Printf("Files skipped:      %d\n", result.SkippedFiles)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("Errors encountered: %d\n", len(result.Errors))
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  ✗ %v\n", err)
		}
	} else {
		if dryRun {
			fmt.Println("\n✓ No errors found (dry run)")
		} else {
			fmt.Println("\n✓ Migration completed successfully")
		}
	}

	// Additional message for dry run
	if dryRun {
		fmt.Println("\nRun without --dry-run to perform the migration.")
	}
}
