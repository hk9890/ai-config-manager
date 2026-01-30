package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/hk9890/ai-config-manager/pkg/config"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	syncSkipExistingFlag bool
	syncDryRunFlag       bool
	syncFormatFlag       string
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
	syncCmd.Flags().StringVar(&syncFormatFlag, "format", "table", "Output format: table, json, yaml")
	syncCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
}

// syncSource syncs resources from a single sync source
func syncSource(src config.SyncSource, manager *repo.Manager) error {
	var sourcePath string
	var mode string

	if src.IsRemote() {
		// Remote source (url): download to workspace, copy to repo
		fmt.Printf("  Mode: Remote (download + copy)\n")
		mode = "copy"

		// Get repository path
		repoPath := manager.GetRepoPath()

		// Create workspace manager
		wsMgr, err := workspace.NewManager(repoPath)
		if err != nil {
			return fmt.Errorf("failed to create workspace manager: %w", err)
		}

		// Parse source URL
		parsed, err := source.ParseSource(src.URL)
		if err != nil {
			return fmt.Errorf("invalid source URL: %w", err)
		}

		// Get clone URL
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			return fmt.Errorf("failed to get clone URL: %w", err)
		}

		// Get or clone repository (using ref if available)
		sourcePath, err = wsMgr.GetOrClone(cloneURL, parsed.Ref)
		if err != nil {
			return fmt.Errorf("failed to download repository: %w", err)
		}

	} else {
		// Local source (path): use path directly, symlink to repo
		fmt.Printf("  Mode: Local (symlink)\n")
		mode = "symlink"

		// Convert to absolute path
		absPath, err := filepath.Abs(src.Path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", src.Path, err)
		}
		sourcePath = absPath
	}

	// Import from source path with appropriate mode
	return addBulkFromLocalWithMode(sourcePath, manager, src.Filter, mode)
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

	// Set flags for the duration of the sync operation
	originalForceFlag := forceFlag
	originalDryRunFlag := dryRunFlag
	originalSkipExistingFlag := skipExistingFlag
	originalImportFormatFlag := importFormatFlag

	// Default to force=true unless --skip-existing specified
	forceFlag = !syncSkipExistingFlag
	skipExistingFlag = syncSkipExistingFlag
	dryRunFlag = syncDryRunFlag
	importFormatFlag = syncFormatFlag

	// Restore original flags when done
	defer func() {
		forceFlag = originalForceFlag
		dryRunFlag = originalDryRunFlag
		skipExistingFlag = originalSkipExistingFlag
		importFormatFlag = originalImportFormatFlag
	}()

	// Process each source
	for i, src := range cfg.Sync.Sources {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(cfg.Sync.Sources), src.GetSourcePath())
		if src.Filter != "" {
			fmt.Printf("  Filter: %s\n", src.Filter)
		}

		// Sync this source
		if err := syncSource(src, manager); err != nil {
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
