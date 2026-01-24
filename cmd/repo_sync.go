package cmd

import (
	"fmt"

	"github.com/hk9890/ai-config-manager/pkg/config"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/spf13/cobra"
)

var (
	syncSkipExistingFlag bool
	syncDryRunFlag       bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync resources from configured sources",
	Long: `Sync resources from configured sources.

This command reads the sync.sources list from your global configuration file
(~/.config/aimgr/aimgr.yaml) and imports all resources from each source.

By default, existing resources will be overwritten (force mode). Use --skip-existing
to skip resources that already exist in the repository.

Configuration format in aimgr.yaml:

  sync:
    sources:
      - url: https://github.com/owner/repo
        filter: "skill/*"            # optional: filter by pattern
      - url: gh:another/repo@v1.0.0
        filter: "*test*"             # optional: filter by pattern
      - url: ~/local/path
                                     # no filter: import all

Each source can have an optional filter pattern to limit which resources are
imported. Filters use glob patterns:
  - "skill/*"         - Only skills
  - "command/*"       - Only commands  
  - "agent/*"         - Only agents
  - "*test*"          - Any resource with "test" in the name
  - "skill/pdf*"      - Skills starting with "pdf"

Examples:
  # Sync all configured sources (overwrites existing)
  aimgr repo sync

  # Sync without overwriting existing resources
  aimgr repo sync --skip-existing

  # Preview what would be synced
  aimgr repo sync --dry-run`,
	RunE: runSync,
}

func init() {
	repoCmd.AddCommand(syncCmd)

	// Add flags
	syncCmd.Flags().BoolVar(&syncSkipExistingFlag, "skip-existing", false, "Skip conflicts silently")
	syncCmd.Flags().BoolVar(&syncDryRunFlag, "dry-run", false, "Preview without importing")
}

// runSync executes the sync command
func runSync(cmd *cobra.Command, args []string) error {
	// Load global config
	cfg, err := config.LoadGlobal()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if any sources configured
	if len(cfg.Sync.Sources) == 0 {
		return fmt.Errorf("no sync sources configured\n\nAdd sources to your config file (~/.config/aimgr/aimgr.yaml):\n\n  sync:\n    sources:\n      - url: https://github.com/owner/repo\n        filter: \"skill/*\"  # optional")
	}

	// Create manager
	manager, err := repo.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create repo manager: %w", err)
	}

	// Print header
	fmt.Printf("Syncing from %d configured source(s)...\n", len(cfg.Sync.Sources))
	if syncDryRunFlag {
		fmt.Println("Mode: DRY RUN (preview only)")
	}
	fmt.Println()

	// Track overall results
	totalAdded := 0
	totalSkipped := 0
	totalFailed := 0
	sourcesProcessed := 0
	sourcesFailed := 0

	// Set forceFlag and dryRunFlag for the duration of the sync operation
	originalForceFlag := forceFlag
	originalDryRunFlag := dryRunFlag
	originalSkipExistingFlag := skipExistingFlag

	// Default to force=true unless --skip-existing specified
	forceFlag = !syncSkipExistingFlag
	skipExistingFlag = syncSkipExistingFlag
	dryRunFlag = syncDryRunFlag

	// Restore original flags when done
	defer func() {
		forceFlag = originalForceFlag
		dryRunFlag = originalDryRunFlag
		skipExistingFlag = originalSkipExistingFlag
	}()

	// Process each source
	for i, src := range cfg.Sync.Sources {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(cfg.Sync.Sources), src.URL)
		if src.Filter != "" {
			fmt.Printf("  Filter: %s\n", src.Filter)
		}

		// Parse source
		parsed, err := source.ParseSource(src.URL)
		if err != nil {
			fmt.Printf("  ✗ Error: invalid source format: %v\n\n", err)
			sourcesFailed++
			continue
		}

		// Import from source based on type
		if parsed.Type == source.GitHub || parsed.Type == source.GitURL {
			// GitHub source
			err = addBulkFromGitHubWithFilter(parsed, manager, src.Filter)
		} else {
			// Local source
			err = addBulkFromLocalWithFilter(parsed.LocalPath, manager, src.Filter)
		}

		if err != nil {
			fmt.Printf("  ✗ Error: %v\n\n", err)
			sourcesFailed++
			continue
		}

		// Source processed successfully
		sourcesProcessed++
		fmt.Println()
	}

	// Print overall summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Sync Complete: %d/%d sources processed successfully\n", sourcesProcessed, len(cfg.Sync.Sources))
	if sourcesFailed > 0 {
		fmt.Printf("  %d source(s) failed\n", sourcesFailed)
	}
	fmt.Println()

	// Note about tracking individual results
	if totalAdded > 0 || totalSkipped > 0 || totalFailed > 0 {
		fmt.Printf("Total: %d added, %d skipped, %d failed\n", totalAdded, totalSkipped, totalFailed)
	}

	// Return error if all sources failed
	if sourcesFailed == len(cfg.Sync.Sources) {
		return fmt.Errorf("all sources failed to sync")
	}

	return nil
}
