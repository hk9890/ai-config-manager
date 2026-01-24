package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

var (
	pruneForceFlag  bool
	pruneDryRunFlag bool
)

// OrphanedMetadata represents metadata with a non-existent source path
type OrphanedMetadata struct {
	Name       string
	Type       resource.ResourceType
	SourceURL  string
	SourceType string
	FilePath   string
}

// repoPruneCmd represents the prune command
var repoPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove orphaned metadata entries",
	Long: `Remove metadata entries for resources whose source paths no longer exist.

This command scans all metadata files and identifies entries where the source
path (for local/file sources) no longer exists on disk. Orphaned metadata is
typically created when source files are moved or deleted outside of aimgr.

Examples:
  aimgr repo prune                # Scan and remove orphaned metadata (with confirmation)
  aimgr repo prune --dry-run      # Preview orphaned entries without removing
  aimgr repo prune --force        # Remove orphaned metadata without confirmation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Find orphaned metadata
		orphaned, err := findOrphanedMetadata(manager)
		if err != nil {
			return err
		}

		// If no orphaned metadata found
		if len(orphaned) == 0 {
			fmt.Println("No orphaned metadata found.")
			return nil
		}

		// Display orphaned entries
		displayOrphanedMetadata(orphaned)

		// Dry run mode - just display, don't remove
		if pruneDryRunFlag {
			fmt.Printf("\n[DRY RUN] Would remove %d orphaned metadata %s\n",
				len(orphaned), pluralize("entry", "entries", len(orphaned)))
			return nil
		}

		// Confirmation prompt (unless --force)
		if !pruneForceFlag {
			fmt.Printf("\nRemove %d orphaned metadata %s? [y/N] ",
				len(orphaned), pluralize("entry", "entries", len(orphaned)))

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove orphaned metadata
		removed, failed := removeOrphanedMetadata(orphaned)

		// Display summary
		fmt.Printf("\n✓ Removed %d orphaned metadata %s",
			removed, pluralize("entry", "entries", removed))
		if failed > 0 {
			fmt.Printf(" (%d failed)", failed)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoPruneCmd)
	repoPruneCmd.Flags().BoolVar(&pruneForceFlag, "force", false, "Skip confirmation prompt")
	repoPruneCmd.Flags().BoolVar(&pruneDryRunFlag, "dry-run", false, "Preview orphaned metadata without removing")
}

// findOrphanedMetadata scans all metadata and returns entries with non-existent sources
func findOrphanedMetadata(manager *repo.Manager) ([]OrphanedMetadata, error) {
	var orphaned []OrphanedMetadata

	repoPath := manager.GetRepoPath()
	metadataDir := filepath.Join(repoPath, ".metadata")

	// Check all resource types
	resourceTypes := []resource.ResourceType{resource.Command, resource.Skill, resource.Agent}

	for _, resType := range resourceTypes {
		typeDir := filepath.Join(metadataDir, string(resType)+"s")

		// Check if type directory exists
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}

		// Read metadata files in directory
		entries, err := os.ReadDir(typeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata directory %s: %w", typeDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), "-metadata.json") {
				continue
			}

			// Extract resource name from filename (remove "-metadata.json")
			name := strings.TrimSuffix(entry.Name(), "-metadata.json")

			// Load metadata
			meta, err := metadata.Load(name, resType, repoPath)
			if err != nil {
				// Skip if we can't load metadata
				continue
			}

			// Check if source path exists
			if isOrphaned(meta) {
				orphaned = append(orphaned, OrphanedMetadata{
					Name:       meta.Name,
					Type:       meta.Type,
					SourceURL:  meta.SourceURL,
					SourceType: meta.SourceType,
					FilePath:   filepath.Join(typeDir, entry.Name()),
				})
			}
		}
	}

	return orphaned, nil
}

// isOrphaned checks if a metadata entry's source path no longer exists
func isOrphaned(meta *metadata.ResourceMetadata) bool {
	// Only check local and file sources (git sources are transient)
	if meta.SourceType != "local" && meta.SourceType != "file" {
		return false
	}

	// Extract local path from source URL
	localPath := meta.SourceURL
	if strings.HasPrefix(localPath, "file://") {
		localPath = filepath.Clean(localPath[7:]) // Remove "file://" prefix
	}

	// Check if path exists
	_, err := os.Stat(localPath)
	return os.IsNotExist(err)
}

// displayOrphanedMetadata displays a list of orphaned metadata entries
func displayOrphanedMetadata(orphaned []OrphanedMetadata) {
	fmt.Printf("Found %d orphaned metadata %s:\n\n",
		len(orphaned), pluralize("entry", "entries", len(orphaned)))

	for _, item := range orphaned {
		fmt.Printf("  • %s '%s'\n", item.Type, item.Name)
		fmt.Printf("    Source: %s (%s)\n", item.SourceURL, item.SourceType)
	}
}

// removeOrphanedMetadata removes orphaned metadata files and returns counts
func removeOrphanedMetadata(orphaned []OrphanedMetadata) (removed int, failed int) {
	for _, item := range orphaned {
		if err := os.Remove(item.FilePath); err != nil {
			fmt.Printf("✗ Failed to remove metadata for %s '%s': %v\n", item.Type, item.Name, err)
			failed++
		} else {
			removed++
		}
	}
	return removed, failed
}

// pluralize returns singular or plural form based on count
func pluralize(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}
