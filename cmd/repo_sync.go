package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/repomanifest"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/sourcemetadata"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/spf13/cobra"
)

// resourceInfo holds the name and type of a resource for pre-sync inventory tracking.
type resourceInfo struct {
	Name string
	Type resource.ResourceType
}

// collectResourcesBySource returns a map of source identifier -> []resourceInfo
// for all resources in the repo that have a source assigned.
// This is used before sync to build a pre-sync inventory, enabling orphan detection
// by comparing the "before" set with the "after" set.
func collectResourcesBySource(manager *repo.Manager, repoPath string) (map[string][]resourceInfo, error) {
	// List all resources in the repo (commands, skills, agents, packages)
	resources, err := manager.List(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	bySource := make(map[string][]resourceInfo)
	for _, res := range resources {
		var key string

		if res.Type == resource.PackageType {
			// Load package-specific metadata
			pkgMeta, err := metadata.LoadPackageMetadata(res.Name, repoPath)
			if err != nil {
				continue // Skip packages without metadata
			}
			key = pkgMeta.SourceID
			if key == "" {
				key = pkgMeta.SourceName
			}
		} else {
			// Load resource metadata (commands, skills, agents)
			meta, err := metadata.Load(res.Name, res.Type, repoPath)
			if err != nil {
				continue // Skip resources without metadata
			}
			key = meta.SourceID
			if key == "" {
				key = meta.SourceName
			}
		}

		if key == "" {
			continue // Skip resources without source info
		}

		bySource[key] = append(bySource[key], resourceInfo{Name: res.Name, Type: res.Type})
	}

	return bySource, nil
}

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

// scanSourceResources scans a source directory and returns the set of
// resource names it contains, keyed by type.
// This uses the same discovery functions as importFromLocalPathWithMode
// to ensure consistent resource detection.
func scanSourceResources(sourcePath string) (map[resource.ResourceType]map[string]bool, error) {
	result := make(map[resource.ResourceType]map[string]bool)

	// Discover commands
	commands, _, err := discovery.DiscoverCommandsWithErrors(sourcePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to discover commands: %w", err)
	}
	if len(commands) > 0 {
		cmdSet := make(map[string]bool, len(commands))
		for _, cmd := range commands {
			cmdSet[cmd.Name] = true
		}
		result[resource.Command] = cmdSet
	}

	// Discover skills
	skills, _, err := discovery.DiscoverSkillsWithErrors(sourcePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to discover skills: %w", err)
	}
	if len(skills) > 0 {
		skillSet := make(map[string]bool, len(skills))
		for _, skill := range skills {
			skillSet[skill.Name] = true
		}
		result[resource.Skill] = skillSet
	}

	// Discover agents
	agents, _, err := discovery.DiscoverAgentsWithErrors(sourcePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}
	if len(agents) > 0 {
		agentSet := make(map[string]bool, len(agents))
		for _, agent := range agents {
			agentSet[agent.Name] = true
		}
		result[resource.Agent] = agentSet
	}

	// Discover packages
	packages, err := discovery.DiscoverPackages(sourcePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to discover packages: %w", err)
	}
	if len(packages) > 0 {
		pkgSet := make(map[string]bool, len(packages))
		for _, pkg := range packages {
			pkgSet[pkg.Name] = true
		}
		result[resource.PackageType] = pkgSet
	}

	return result, nil
}

// syncSource syncs resources from a single manifest source.
// Returns the resolved source path (for use in post-sync scanning) and any error.
func syncSource(src *repomanifest.Source, manager *repo.Manager) (string, error) {
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
			return "", fmt.Errorf("failed to create workspace manager: %w", err)
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
			return "", fmt.Errorf("invalid source URL: %w", err)
		}

		// Get clone URL
		cloneURL, err := source.GetCloneURL(parsed)
		if err != nil {
			return "", fmt.Errorf("failed to get clone URL: %w", err)
		}

		// Get or clone repository (using ref if available)
		sourcePath, err = wsMgr.GetOrClone(cloneURL, src.Ref)
		if err != nil {
			return "", fmt.Errorf("failed to download repository: %w", err)
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
			return "", fmt.Errorf("invalid path %s: %w", src.Path, err)
		}
		sourcePath = absPath
	} else {
		return "", fmt.Errorf("source must have either URL or Path")
	}

	// Import from source path with appropriate mode
	// Note: We don't pass a filter here because sync doesn't support filtering
	// All resources from the source should be synced
	err := addBulkFromLocalWithMode(sourcePath, manager, "", src.ID, mode, src.Name)
	if err != nil {
		return "", err
	}
	return sourcePath, nil
}

// runSync executes the sync command
func runSync(cmd *cobra.Command, args []string) error {
	// Create manager
	manager, err := NewManagerWithLogLevel()
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

	// Collect pre-sync inventory for orphan detection (bmz1.1)
	// This captures which resources belong to each source before sync,
	// so we can compare with the post-sync state to find removals.
	repoPath := manager.GetRepoPath()
	preSyncResources, err := collectResourcesBySource(manager, repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not collect pre-sync inventory: %v\n", err)
		preSyncResources = make(map[string][]resourceInfo)
	}

	// Track results
	sourcesProcessed := 0
	sourcesFailed := 0
	failedSources := make([]string, 0)

	// Track removed resources detected per source (bmz1.2)
	// These are resources that existed in the repo pre-sync but are no longer
	// in the source directory. Collected here for the next task (bmz1.3) to act on.
	removedResources := make(map[string][]resourceInfo)

	// Process each source
	for i, src := range manifest.Sources {
		sourceDesc := src.Name
		if src.URL != "" {
			sourceDesc = fmt.Sprintf("%s (%s)", src.Name, src.URL)
		} else if src.Path != "" {
			sourceDesc = fmt.Sprintf("%s (%s)", src.Name, src.Path)
		}

		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(manifest.Sources), sourceDesc)

		// Sync this source (returns resolved source path for scanning)
		sourcePath, err := syncSource(src, manager)
		if err != nil {
			// Log warning but don't fail entire sync
			fmt.Printf("  ⚠ Warning: %v\n", err)
			fmt.Printf("  Skipping this source and continuing...\n\n")
			sourcesFailed++
			failedSources = append(failedSources, src.Name)
			continue
		}

		// Detect removed resources by comparing pre-sync inventory with current source (bmz1.2)
		// Only for successfully synced sources — failed sources are excluded from removal detection.
		sourceKey := src.ID
		if sourceKey == "" {
			sourceKey = src.Name
		}

		preSyncSet := preSyncResources[sourceKey]
		if len(preSyncSet) > 0 {
			sourceResources, scanErr := scanSourceResources(sourcePath)
			if scanErr != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not scan source for removal detection: %v\n", scanErr)
			} else {
				for _, res := range preSyncSet {
					// Check if this resource still exists in the source
					if typeSet, ok := sourceResources[res.Type]; ok {
						if typeSet[res.Name] {
							continue // Still in source, keep it
						}
					}
					// Resource was in repo but not in source -> removed
					removedResources[sourceKey] = append(removedResources[sourceKey], res)
				}
			}
		}

		// Update last_synced timestamp in metadata after successful sync
		if !syncDryRunFlag {
			if state, ok := metadata.Sources[src.Name]; ok {
				state.LastSynced = time.Now()
				if src.ID != "" {
					state.SourceID = src.ID
				}
			} else {
				// Create new state if it doesn't exist
				metadata.Sources[src.Name] = &sourcemetadata.SourceState{
					SourceID:   src.ID,
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
		} else {
			// Commit metadata changes to git
			if err := manager.CommitChanges("aimgr: update sync timestamps"); err != nil {
				// Don't fail if commit fails (e.g., not a git repo)
				fmt.Printf("⚠ Warning: Failed to commit metadata: %v\n", err)
			}
		}
	}

	// Report detected removals (actual removal is handled by bmz1.3)
	totalRemoved := 0
	for _, resources := range removedResources {
		totalRemoved += len(resources)
	}
	if totalRemoved > 0 {
		fmt.Printf("Detected %d resource(s) removed from source(s):\n", totalRemoved)
		for sourceKey, resources := range removedResources {
			var names []string
			for _, res := range resources {
				names = append(names, fmt.Sprintf("%s/%s", res.Type, res.Name))
			}
			fmt.Printf("  %s: %s\n", sourceKey, strings.Join(names, ", "))
		}
		fmt.Println()
	}

	// Print overall summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Sync Complete: %d/%d sources synced successfully\n", sourcesProcessed, len(manifest.Sources))
	if sourcesFailed > 0 {
		fmt.Printf("  %d source(s) failed or skipped:\n", sourcesFailed)
		for _, name := range failedSources {
			fmt.Printf("    - %s\n", name)
		}
	}
	if totalRemoved > 0 {
		fmt.Printf("  %d resource(s) detected as removed from source(s)\n", totalRemoved)
	}
	fmt.Println()

	// Return error if all sources failed
	if sourcesFailed == len(manifest.Sources) {
		return fmt.Errorf("all sources failed to sync")
	}

	return nil
}
