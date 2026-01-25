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
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	pruneForceFlag  bool
	pruneDryRunFlag bool
)

// CachedRepo represents a cached repository to be pruned
type CachedRepo struct {
	URL  string
	Path string
	Size int64 // Size in bytes
}

// repoPruneCmd represents the prune command
var repoPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove unreferenced workspace caches",
	Long: `Remove Git repositories from .workspace/ that are not referenced by any installed resources.

What gets pruned:
  - Git repository clones in .workspace/ that are not used by any current resources
  - Cached repositories from removed or outdated resources
  - Orphaned caches from failed operations

What is NOT pruned:
  - Caches referenced by currently installed resources
  - Local file sources (not cached in .workspace/)
  - Resource files themselves (only Git caches are removed)

When to run prune:
  - After removing many resources from the repository
  - When .workspace/ directory grows too large
  - As periodic maintenance to reclaim disk space
  - After changing Git source URLs for resources

How it works:
  1. Scans all resource metadata to build list of referenced Git URLs
  2. Compares against cached repositories in .workspace/
  3. Identifies unreferenced caches (not used by any resource)
  4. Removes unreferenced caches and frees disk space
  5. Updates cache metadata (.cache-metadata.json)

Examples:
  aimgr repo prune                # Scan and remove unreferenced caches (with confirmation)
  aimgr repo prune --dry-run      # Preview what would be removed without removing
  aimgr repo prune --force        # Remove without confirmation prompt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Create workspace manager
		workspaceManager, err := workspace.NewManager(manager.GetRepoPath())
		if err != nil {
			return err
		}

		// Find unreferenced cached repos
		unreferenced, err := findUnreferencedCaches(manager, workspaceManager)
		if err != nil {
			return err
		}

		// If no unreferenced caches found
		if len(unreferenced) == 0 {
			fmt.Println("No unreferenced workspace caches found.")
			return nil
		}

		// Display unreferenced caches
		totalSize := displayUnreferencedCaches(unreferenced)

		// Dry run mode - just display, don't remove
		if pruneDryRunFlag {
			fmt.Printf("\n[DRY RUN] Would remove %d cached %s, freeing %s\n",
				len(unreferenced),
				pluralize("repository", "repositories", len(unreferenced)),
				formatSize(totalSize))
			return nil
		}

		// Confirmation prompt (unless --force)
		if !pruneForceFlag {
			fmt.Printf("\nRemove %d cached %s? This will free %s. [y/N] ",
				len(unreferenced),
				pluralize("repository", "repositories", len(unreferenced)),
				formatSize(totalSize))

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

		// Remove unreferenced caches
		removed, failed, freedSize := removeUnreferencedCaches(workspaceManager, unreferenced)

		// Display summary
		fmt.Printf("\n✓ Removed %d cached %s, freed %s",
			removed,
			pluralize("repository", "repositories", removed),
			formatSize(freedSize))
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
	repoPruneCmd.Flags().BoolVar(&pruneDryRunFlag, "dry-run", false, "Preview what would be removed without removing")
}

// findUnreferencedCaches finds all cached repos that are not referenced by any resource metadata
func findUnreferencedCaches(manager *repo.Manager, workspaceManager *workspace.Manager) ([]CachedRepo, error) {
	// Get all cached URLs
	cachedURLs, err := workspaceManager.ListCached()
	if err != nil {
		return nil, fmt.Errorf("failed to list cached repositories: %w", err)
	}

	// If no caches, nothing to prune
	if len(cachedURLs) == 0 {
		return []CachedRepo{}, nil
	}

	// Collect all referenced Git URLs from metadata
	referencedURLs, err := collectReferencedGitURLs(manager)
	if err != nil {
		return nil, fmt.Errorf("failed to collect referenced URLs: %w", err)
	}

	// Build set of referenced URLs (normalized) for fast lookup
	referencedSet := make(map[string]bool)
	for _, url := range referencedURLs {
		referencedSet[normalizeURL(url)] = true
	}

	// Find unreferenced caches
	var unreferenced []CachedRepo
	workspaceDir := filepath.Join(manager.GetRepoPath(), ".workspace")

	for _, cachedURL := range cachedURLs {
		normalized := normalizeURL(cachedURL)
		if !referencedSet[normalized] {
			// This cache is not referenced - calculate its path and size
			hash := workspace.ComputeHash(normalized)
			cachePath := filepath.Join(workspaceDir, hash)

			size, err := getDirSize(cachePath)
			if err != nil {
				// If we can't calculate size, use 0
				size = 0
			}

			unreferenced = append(unreferenced, CachedRepo{
				URL:  normalized,
				Path: cachePath,
				Size: size,
			})
		}
	}

	return unreferenced, nil
}

// collectReferencedGitURLs collects all Git URLs from resource metadata
func collectReferencedGitURLs(manager *repo.Manager) ([]string, error) {
	repoPath := manager.GetRepoPath()
	metadataDir := filepath.Join(repoPath, ".metadata")

	var gitURLs []string

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

			// Extract resource name from filename
			name := strings.TrimSuffix(entry.Name(), "-metadata.json")

			// Load metadata
			meta, err := metadata.Load(name, resType, repoPath)
			if err != nil {
				// Skip if we can't load metadata
				continue
			}

			// Only collect Git source URLs
			if isGitSource(meta.SourceType) {
				gitURLs = append(gitURLs, meta.SourceURL)
			}
		}
	}

	return gitURLs, nil
}

// isGitSource checks if a source type is a Git source
func isGitSource(sourceType string) bool {
	return sourceType == "github" || sourceType == "git-url" || sourceType == "gitlab"
}

// normalizeURL normalizes a Git URL for consistent comparison
func normalizeURL(url string) string {
	normalized := strings.TrimSpace(url)
	normalized = strings.ToLower(normalized)
	// Strip .git and / in a loop to handle cases like "/.git" or ".git/"
	for {
		oldNormalized := normalized
		normalized = strings.TrimSuffix(normalized, "/")
		normalized = strings.TrimSuffix(normalized, ".git")
		if normalized == oldNormalized {
			break
		}
	}
	return normalized
}

// getDirSize calculates the total size of a directory in bytes
func getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// displayUnreferencedCaches displays the list of unreferenced caches and returns total size
func displayUnreferencedCaches(unreferenced []CachedRepo) int64 {
	fmt.Printf("Found %d unreferenced cached %s:\n\n",
		len(unreferenced),
		pluralize("repository", "repositories", len(unreferenced)))

	var totalSize int64
	for _, cache := range unreferenced {
		totalSize += cache.Size
		fmt.Printf("  • %s (%s)\n", cache.URL, formatSize(cache.Size))
	}

	return totalSize
}

// removeUnreferencedCaches removes unreferenced caches and returns counts
func removeUnreferencedCaches(workspaceManager *workspace.Manager, unreferenced []CachedRepo) (removed int, failed int, freedSize int64) {
	for _, cache := range unreferenced {
		if err := workspaceManager.Remove(cache.URL); err != nil {
			fmt.Printf("✗ Failed to remove cache for %s: %v\n", cache.URL, err)
			failed++
		} else {
			removed++
			freedSize += cache.Size
		}
	}
	return removed, failed, freedSize
}

// formatSize formats a byte size as a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// pluralize returns singular or plural form based on count
func pluralize(singular, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}
