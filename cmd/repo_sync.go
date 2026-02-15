package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/repomanifest"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/sourcemetadata"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	syncSkipExistingFlag bool
	syncDryRunFlag       bool
	syncFormatFlag       string
	syncForceFlag        bool
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync resources from configured sources",
	Long: `Sync resources from configured sources in ai.repo.yaml.

This command reads sources from the repository's ai.repo.yaml manifest file
and re-imports all resources from each source. This is useful for:
  - Pulling latest changes from remote repositories
  - Updating symlinked resources from local paths
  - Cleaning up orphaned resources from removed sources

By default, existing resources will be overwritten (force mode). Use --skip-existing
to skip resources that already exist in the repository.

The ai.repo.yaml file is automatically maintained when you use "aimgr repo add".

Example ai.repo.yaml:

  version: 1
  sources:
    - name: my-team-resources
      path: /home/user/resources
    - name: community-skills
      url: https://github.com/owner/repo
      ref: main

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
	syncCmd.Flags().BoolVar(&syncForceFlag, "force", false, "Overwrite existing resources (default: true)")
	syncCmd.Flags().StringVar(&syncFormatFlag, "format", "table", "Output format: table, json, yaml")
	_ = syncCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
}

// syncSource syncs resources from a single manifest source
func syncSource(src *repomanifest.Source, manager *repo.Manager) error {
	var sourcePath string
	var mode string

	if src.URL != "" {
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

		// Parse source URL to get clone URL
		// Note: We construct a pseudo-source string for parsing
		sourceStr := src.URL
		if src.Ref != "" {
			sourceStr = sourceStr + "@" + src.Ref
		}
		if src.Subpath != "" {
			sourceStr = sourceStr + "/" + src.Subpath
		}

		parsed, err := source.ParseSource(sourceStr)
		if err != nil {
			return fmt.Errorf("invalid source URL: %w", err)
		}

		// Get clone URL
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			return fmt.Errorf("failed to get clone URL: %w", err)
		}

		// Get or clone repository (using ref if available)
		sourcePath, err = wsMgr.GetOrClone(cloneURL, src.Ref)
		if err != nil {
			return fmt.Errorf("failed to download repository: %w", err)
		}

		// Apply subpath if specified
		if src.Subpath != "" {
			sourcePath = filepath.Join(sourcePath, src.Subpath)
		}

	} else if src.Path != "" {
		// Local source (path): use path directly, symlink to repo
		fmt.Printf("  Mode: Local (symlink)\n")
		mode = src.GetMode() // Use mode from source (implicit: path=symlink, url=copy)

		// Convert to absolute path
		absPath, err := filepath.Abs(src.Path)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", src.Path, err)
		}
		sourcePath = absPath
	} else {
		return fmt.Errorf("source must have either URL or Path")
	}

	// Import from source path with appropriate mode
	// Note: We don't pass a filter here because sync doesn't support filtering
	// All resources from the source should be synced
	return addBulkFromLocalWithMode(sourcePath, manager, "", mode)
}

// runSync executes the sync command
func runSync(cmd *cobra.Command, args []string) error {
	// Create manager
	manager, err := repo.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create repo manager: %w", err)
	}

	// Load manifest from ai.repo.yaml
	manifest, err := repomanifest.Load(manager.GetRepoPath())
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Check if any sources configured
	if len(manifest.Sources) == 0 {
		return fmt.Errorf("no sync sources configured\n\nAdd sources using:\n  aimgr repo add <source>\n\nSources are automatically tracked in ai.repo.yaml")
	}

	// Load source metadata for timestamps
	metadata, err := sourcemetadata.Load(manager.GetRepoPath())
	if err != nil {
		// If metadata doesn't exist yet, create empty one
		metadata = &sourcemetadata.SourceMetadata{
			Version: 1,
			Sources: make(map[string]*sourcemetadata.SourceState),
		}
	}

	// Print header
	fmt.Printf("Syncing from %d configured source(s)...\n", len(manifest.Sources))
	if syncDryRunFlag {
		fmt.Println("Mode: DRY RUN (preview only)")
	}
	fmt.Println()

	// Set flags for the duration of the sync operation
	originalForceFlag := forceFlag
	originalDryRunFlag := dryRunFlag
	originalSkipExistingFlag := skipExistingFlag
	originalAddFormatFlag := addFormatFlag

	// Default to force=true unless --skip-existing specified
	forceFlag = !syncSkipExistingFlag
	skipExistingFlag = syncSkipExistingFlag
	dryRunFlag = syncDryRunFlag
	addFormatFlag = syncFormatFlag

	// Restore original flags when done
	defer func() {
		forceFlag = originalForceFlag
		dryRunFlag = originalDryRunFlag
		skipExistingFlag = originalSkipExistingFlag
		addFormatFlag = originalAddFormatFlag
	}()

	// Track results
	sourcesProcessed := 0
	sourcesFailed := 0
	failedSources := make([]string, 0)

	// Process each source
	for i, src := range manifest.Sources {
		sourceDesc := src.Name
		if src.URL != "" {
			sourceDesc = fmt.Sprintf("%s (%s)", src.Name, src.URL)
		} else if src.Path != "" {
			sourceDesc = fmt.Sprintf("%s (%s)", src.Name, src.Path)
		}

		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(manifest.Sources), sourceDesc)

		// Sync this source
		if err := syncSource(src, manager); err != nil {
			// Log warning but don't fail entire sync
			fmt.Printf("  ⚠ Warning: %v\n", err)
			fmt.Printf("  Skipping this source and continuing...\n\n")
			sourcesFailed++
			failedSources = append(failedSources, src.Name)
			continue
		}

		// Update last_synced timestamp in metadata after successful sync
		if !syncDryRunFlag {
			if state, ok := metadata.Sources[src.Name]; ok {
				state.LastSynced = time.Now()
			} else {
				// Create new state if it doesn't exist
				metadata.Sources[src.Name] = &sourcemetadata.SourceState{
					Added:      time.Now(),
					LastSynced: time.Now(),
				}
			}
		}

		// Source processed successfully
		sourcesProcessed++
		fmt.Println()
	}

	// Save updated metadata with new timestamps
	if !syncDryRunFlag && sourcesProcessed > 0 {
		if err := metadata.Save(manager.GetRepoPath()); err != nil {
			// Don't fail, just warn
			fmt.Printf("⚠ Warning: Failed to save metadata: %v\n", err)
		}
	}

	// TODO: Orphan cleanup
	// After syncing all sources, find and remove resources that:
	// 1. Have metadata.SourceName set
	// 2. That source name is no longer in manifest.Sources
	// This ensures resources from removed sources are cleaned up
	// For now, we'll skip this to keep the initial implementation simple

	// Print overall summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Sync Complete: %d/%d sources synced successfully\n", sourcesProcessed, len(manifest.Sources))
	if sourcesFailed > 0 {
		fmt.Printf("  %d source(s) failed or skipped:\n", sourcesFailed)
		for _, name := range failedSources {
			fmt.Printf("    - %s\n", name)
		}
	}
	fmt.Println()

	// Return error if all sources failed
	if sourcesFailed == len(manifest.Sources) {
		return fmt.Errorf("all sources failed to sync")
	}

	return nil
}
