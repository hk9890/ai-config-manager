package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
	"github.com/spf13/cobra"
)

var (
	updateForceFlag  bool
	updateDryRunFlag bool
)

// UpdateResult tracks the results of update operations
type UpdateResult struct {
	Name    string
	Type    resource.ResourceType
	Success bool
	Skipped bool // True if resource was skipped (e.g., missing source)
	Message string
}

// UpdateContext tracks progress during update operations
type UpdateContext struct {
	Current int
	Total   int
}

// repoUpdateCmd represents the update command
var repoUpdateCmd = &cobra.Command{
	Use:   "update [pattern]...",
	Short: "Update resources from their original sources",
	Long: `Update resources from their original sources.

Updates can refresh resources from GitHub repositories, local paths, or file sources.
The source information is retrieved from the resource metadata.

Workspace Caching: Git repositories are cached in .workspace/ for faster subsequent
updates. The first update clones the full repository, while subsequent updates only
pull changes. This dramatically improves performance:
  - First update: Full git clone (slower)
  - Subsequent updates: Git pull only (10-50x faster)
  - Cached repos are reused across all resources from the same source

Performance: Resources from the same Git repository are batched together for efficient
updates. This means a single git clone operation is shared across multiple resources
from the same source, significantly improving update speed for bulk operations.

Patterns support wildcards (* and ?) and can filter by type using 'type/pattern' format.
If no patterns are provided, all resources are updated.

Examples:
  aimgr repo update                    # Update all resources
  aimgr repo update skill/my-skill     # Update specific skill
  aimgr repo update skill/*            # Update all skills
  aimgr repo update command/test*      # Update commands starting with 'test'
  aimgr repo update skill/* agent/*    # Update all skills and agents
  aimgr repo update --dry-run          # Preview what would be updated
  aimgr repo update --force            # Force update even with local changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var results []UpdateResult
		var ctx UpdateContext

		// Build list of resources to update
		var toUpdate []string
		if len(args) == 0 {
			// Update all resources - build list of all resources
			typeFilter := resource.Skill
			skills, _ := manager.List(&typeFilter)
			for _, res := range skills {
				toUpdate = append(toUpdate, fmt.Sprintf("skill/%s", res.Name))
			}

			typeFilter = resource.Command
			commands, _ := manager.List(&typeFilter)
			for _, res := range commands {
				toUpdate = append(toUpdate, fmt.Sprintf("command/%s", res.Name))
			}

			typeFilter = resource.Agent
			agents, _ := manager.List(&typeFilter)
			for _, res := range agents {
				toUpdate = append(toUpdate, fmt.Sprintf("agent/%s", res.Name))
			}
		} else {
			// Update by patterns
			for _, pattern := range args {
				matches, err := ExpandPattern(manager, pattern)
				if err != nil {
					return err
				}
				toUpdate = append(toUpdate, matches...)
			}

			// Remove duplicates
			toUpdate = uniqueStrings(toUpdate)
		}

		ctx.Total = len(toUpdate)

		if len(toUpdate) > 0 {
			fmt.Printf("Updating %d resources...\n\n", len(toUpdate))
		}

		// Group resources by source for batched updates
		gitSources, localSources, err := groupResourcesBySource(manager, toUpdate)
		if err != nil {
			return err
		}

		// Process Git sources with batching
		for sourceURL, resources := range gitSources {
			if len(resources) > 1 {
				fmt.Printf("Batch: Updating %d resources from %s\n\n", len(resources), sourceURL)
			}
			batchResults := updateBatchFromGitSource(manager, sourceURL, resources, &ctx)
			results = append(results, batchResults...)
		}

		// Process local/file sources individually (no batching)
		for _, res := range localSources {
			ctx.Current++

			// Show progress counter
			fmt.Printf("[%d/%d] %s '%s'\n", ctx.Current, ctx.Total, res.resourceType, res.name)

			result := UpdateResult{
				Name:    res.name,
				Type:    res.resourceType,
				Success: false,
				Skipped: false,
			}

			// Dry run mode - just report what would be done
			if updateDryRunFlag {
				result.Success = true
				result.Message = fmt.Sprintf("Would update from %s (%s)", res.metadata.SourceURL, res.metadata.SourceType)
				fmt.Printf("  ↓ %s\n\n", result.Message)
				results = append(results, result)
				continue
			}

			// Show operation type
			fmt.Printf("  ↓ Updating from local source...\n")

			// Update from local source
			skipped, updateErr := updateFromLocalSource(manager, res.name, res.resourceType, res.metadata)
			if skipped {
				result.Skipped = true
				result.Message = updateErr.Error()
				fmt.Printf("  ⊘ %s\n\n", result.Message)
				results = append(results, result)
				continue
			}

			if updateErr != nil {
				result.Message = updateErr.Error()
				fmt.Printf("  ✗ %s\n\n", result.Message)
				results = append(results, result)
				continue
			}

			// Update LastUpdated timestamp
			res.metadata.LastUpdated = time.Now()
			if err := metadata.Save(res.metadata, manager.GetRepoPath()); err != nil {
				result.Message = fmt.Sprintf("Updated but failed to save metadata: %v", err)
				fmt.Printf("  ✗ %s\n\n", result.Message)
				results = append(results, result)
				continue
			}

			result.Success = true
			result.Message = "Updated successfully"
			fmt.Printf("  ✓ %s\n\n", result.Message)
			results = append(results, result)
		}

		// Display summary
		displayUpdateSummary(results)

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoUpdateCmd)

	// Add flags to update command
	repoUpdateCmd.Flags().BoolVar(&updateForceFlag, "force", false, "Force update, overwriting local changes")
	repoUpdateCmd.Flags().BoolVar(&updateDryRunFlag, "dry-run", false, "Preview updates without making changes")
}

// updateResourceTypeWithProgress updates all resources of a specific type with progress tracking
func updateResourceTypeWithProgress(manager *repo.Manager, resourceType resource.ResourceType, name string, ctx *UpdateContext) ([]UpdateResult, error) {
	var results []UpdateResult

	if name != "" {
		// Update specific resource
		ctx.Current++
		result := updateSingleResourceWithProgress(manager, name, resourceType, ctx)
		results = append(results, result)
	} else {
		// Update all resources of this type
		typeFilter := resourceType
		resources, err := manager.List(&typeFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to list %ss: %w", resourceType, err)
		}

		for _, res := range resources {
			ctx.Current++
			result := updateSingleResourceWithProgress(manager, res.Name, resourceType, ctx)
			results = append(results, result)
		}
	}

	return results, nil
}

// updateResourceType updates all resources of a specific type or a specific resource by name
func updateResourceType(manager *repo.Manager, resourceType resource.ResourceType, name string) ([]UpdateResult, error) {
	ctx := &UpdateContext{Total: 1}
	return updateResourceTypeWithProgress(manager, resourceType, name, ctx)
}

// updateSingleResourceWithProgress updates a single resource with progress display
func updateSingleResourceWithProgress(manager *repo.Manager, name string, resourceType resource.ResourceType, ctx *UpdateContext) UpdateResult {
	// Show progress counter
	fmt.Printf("[%d/%d] %s '%s'\n", ctx.Current, ctx.Total, resourceType, name)

	result := UpdateResult{
		Name:    name,
		Type:    resourceType,
		Success: false,
		Skipped: false,
	}

	// Load metadata
	meta, err := manager.GetMetadata(name, resourceType)
	if err != nil {
		result.Message = fmt.Sprintf("Metadata not found: %v", err)
		fmt.Printf("  ✗ %s\n\n", result.Message)
		return result
	}

	// Dry run mode - just report what would be done
	if updateDryRunFlag {
		result.Success = true
		result.Message = fmt.Sprintf("Would update from %s (%s)", meta.SourceURL, meta.SourceType)
		fmt.Printf("  ↓ %s\n\n", result.Message)
		return result
	}

	// Show operation type
	switch meta.SourceType {
	case "github", "git-url", "gitlab":
		fmt.Printf("  ↓ Updating cached repo from %s...\n", meta.SourceURL)
	case "local", "file":
		fmt.Printf("  ↓ Updating from local source...\n")
	}

	// Update based on source type
	switch meta.SourceType {
	case "github", "git-url", "gitlab":
		err = updateFromGitSource(manager, name, resourceType, meta)
	case "local", "file":
		skipped, updateErr := updateFromLocalSource(manager, name, resourceType, meta)
		if skipped {
			result.Skipped = true
			result.Message = updateErr.Error()
			fmt.Printf("  ⊘ %s\n\n", result.Message)
			return result
		}
		err = updateErr
	default:
		result.Message = fmt.Sprintf("Unknown source type: %s", meta.SourceType)
		fmt.Printf("  ✗ %s\n\n", result.Message)
		return result
	}

	if err != nil {
		result.Message = err.Error()
		fmt.Printf("  ✗ %s\n\n", result.Message)
		return result
	}

	// Update LastUpdated timestamp
	meta.LastUpdated = time.Now()
	if err := metadata.Save(meta, manager.GetRepoPath()); err != nil {
		result.Message = fmt.Sprintf("Updated but failed to save metadata: %v", err)
		fmt.Printf("  ✗ %s\n\n", result.Message)
		return result
	}

	result.Success = true
	result.Message = "Updated successfully"
	fmt.Printf("  ✓ %s\n\n", result.Message)
	return result
}

// updateSingleResource updates a single resource from its original source
func updateSingleResource(manager *repo.Manager, name string, resourceType resource.ResourceType) UpdateResult {
	ctx := &UpdateContext{Current: 1, Total: 1}
	return updateSingleResourceWithProgress(manager, name, resourceType, ctx)
}

// updateFromGitSource updates a resource from a GitHub or Git source
func updateFromGitSource(manager *repo.Manager, name string, resourceType resource.ResourceType, meta *metadata.ResourceMetadata) error {
	// Parse the source URL
	parsed, err := source.ParseSource(meta.SourceURL)
	if err != nil {
		return fmt.Errorf("failed to parse source URL: %w", err)
	}

	// Get ref from metadata, default to "main" if empty
	ref := meta.Ref
	if ref == "" {
		ref = "main"
	}

	// Get clone URL
	cloneURL, err := source.GetCloneURL(parsed)
	if err != nil {
		return fmt.Errorf("failed to get clone URL: %w", err)
	}

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(manager.GetRepoPath())
	if err != nil {
		return fmt.Errorf("failed to create workspace manager: %w", err)
	}

	// Get or clone repository using workspace cache
	cachePath, err := workspaceManager.GetOrClone(cloneURL, ref)
	if err != nil {
		return fmt.Errorf("failed to get cached repository: %w", err)
	}

	// Update cached repository to latest
	if err := workspaceManager.Update(cloneURL, ref); err != nil {
		// If update fails, log warning but continue with existing cache
		// This handles cases like network failures where cache is still usable
		fmt.Fprintf(os.Stderr, "warning: failed to update cached repo (using existing cache): %v\n", err)
	}

	// Discover resources in the cached repo
	searchPath := cachePath
	if parsed.Subpath != "" {
		searchPath = filepath.Join(cachePath, parsed.Subpath)
	}

	// Find and update the specific resource
	switch resourceType {
	case resource.Command:
		return updateCommandFromClone(manager, name, searchPath, meta.SourceURL, meta.SourceType)
	case resource.Skill:
		return updateSkillFromClone(manager, name, searchPath, meta.SourceURL, meta.SourceType)
	case resource.Agent:
		return updateAgentFromClone(manager, name, searchPath, meta.SourceURL, meta.SourceType)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// updateBatchFromGitSource updates multiple resources from a single Git clone
func updateBatchFromGitSource(manager *repo.Manager, sourceURL string, resources []resourceInfo, ctx *UpdateContext) []UpdateResult {
	var results []UpdateResult

	// Parse the source URL
	parsed, err := source.ParseSource(sourceURL)
	if err != nil {
		// If clone fails, all resources in batch fail
		for _, res := range resources {
			results = append(results, UpdateResult{
				Name:    res.name,
				Type:    res.resourceType,
				Success: false,
				Skipped: false,
				Message: fmt.Sprintf("failed to parse source URL: %v", err),
			})
		}
		return results
	}

	// Get ref from first resource's metadata, default to "main" if empty
	// All resources in batch should have the same ref (grouped by source URL)
	ref := "main"
	if len(resources) > 0 && resources[0].metadata.Ref != "" {
		ref = resources[0].metadata.Ref
	}

	// Clone repository to temp directory
	cloneURL, err := source.GetCloneURL(parsed)
	if err != nil {
		// If clone fails, all resources in batch fail
		for _, res := range resources {
			results = append(results, UpdateResult{
				Name:    res.name,
				Type:    res.resourceType,
				Success: false,
				Skipped: false,
				Message: fmt.Sprintf("failed to get clone URL: %v", err),
			})
		}
		return results
	}

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(manager.GetRepoPath())
	if err != nil {
		// If workspace manager creation fails, all resources in batch fail
		for _, res := range resources {
			results = append(results, UpdateResult{
				Name:    res.name,
				Type:    res.resourceType,
				Success: false,
				Skipped: false,
				Message: fmt.Sprintf("failed to create workspace manager: %v", err),
			})
		}
		return results
	}

	// Get or clone repository using workspace cache
	cachePath, err := workspaceManager.GetOrClone(cloneURL, ref)
	if err != nil {
		// If cache retrieval fails, all resources in batch fail
		for _, res := range resources {
			results = append(results, UpdateResult{
				Name:    res.name,
				Type:    res.resourceType,
				Success: false,
				Skipped: false,
				Message: fmt.Sprintf("failed to get cached repository: %v", err),
			})
		}
		return results
	}

	// Update cached repository to latest
	updateErr := workspaceManager.Update(cloneURL, ref)
	if updateErr != nil {
		// If update fails, log warning but continue with existing cache
		fmt.Fprintf(os.Stderr, "warning: failed to update cached repo (using existing cache): %v\n", updateErr)
	}

	// Determine search path
	searchPath := cachePath
	if parsed.Subpath != "" {
		searchPath = filepath.Join(cachePath, parsed.Subpath)
	}

	// Show update message (once for the whole batch)
	if !updateDryRunFlag && len(resources) > 0 {
		if updateErr == nil {
			fmt.Printf("  ↓ Updating cached repo from %s... (refreshing %d resources)\n\n", sourceURL, len(resources))
		} else {
			fmt.Printf("  ↓ Using cached repo from %s... (refreshing %d resources)\n\n", sourceURL, len(resources))
		}
	}

	// Update each resource in the batch
	for _, res := range resources {
		ctx.Current++

		// Show progress counter
		fmt.Printf("[%d/%d] %s '%s'\n", ctx.Current, ctx.Total, res.resourceType, res.name)

		result := UpdateResult{
			Name:    res.name,
			Type:    res.resourceType,
			Success: false,
			Skipped: false,
		}

		// Dry run mode - just report what would be done
		if updateDryRunFlag {
			result.Success = true
			result.Message = fmt.Sprintf("Would update from %s (%s)", sourceURL, res.metadata.SourceType)
			fmt.Printf("  ↓ %s\n\n", result.Message)
			results = append(results, result)
			continue
		}

		// Find and update the specific resource
		var resourceUpdateErr error
		switch res.resourceType {
		case resource.Command:
			resourceUpdateErr = updateCommandFromClone(manager, res.name, searchPath, sourceURL, res.metadata.SourceType)
		case resource.Skill:
			resourceUpdateErr = updateSkillFromClone(manager, res.name, searchPath, sourceURL, res.metadata.SourceType)
		case resource.Agent:
			resourceUpdateErr = updateAgentFromClone(manager, res.name, searchPath, sourceURL, res.metadata.SourceType)
		default:
			resourceUpdateErr = fmt.Errorf("unsupported resource type: %s", res.resourceType)
		}

		if resourceUpdateErr != nil {
			result.Message = resourceUpdateErr.Error()
			fmt.Printf("  ✗ %s\n\n", result.Message)
			results = append(results, result)
			continue
		}

		// Update LastUpdated timestamp
		res.metadata.LastUpdated = time.Now()
		if err := metadata.Save(res.metadata, manager.GetRepoPath()); err != nil {
			result.Message = fmt.Sprintf("Updated but failed to save metadata: %v", err)
			fmt.Printf("  ✗ %s\n\n", result.Message)
			results = append(results, result)
			continue
		}

		result.Success = true
		result.Message = "Updated successfully"
		fmt.Printf("  ✓ %s\n\n", result.Message)
		results = append(results, result)
	}

	return results
}

// updateFromLocalSource updates a resource from a local source
// Returns (skipped bool, error) where skipped=true indicates the source path is missing
func updateFromLocalSource(manager *repo.Manager, name string, resourceType resource.ResourceType, meta *metadata.ResourceMetadata) (bool, error) {
	// Extract local path from source URL (format: file:///path/to/resource)
	localPath := meta.SourceURL
	if strings.HasPrefix(localPath, "file://") {
		localPath = filepath.Clean(localPath[7:]) // Remove "file://" prefix
	}

	// Check if source still exists
	if _, err := os.Stat(localPath); err != nil {
		// Source path missing - return as skipped with informative message
		return true, fmt.Errorf("source path no longer exists (consider running 'aimgr repo prune' to clean up orphaned metadata)")
	}

	// Remove existing resource (force mode is implicit for update)
	if err := manager.Remove(name, resourceType); err != nil {
		return false, fmt.Errorf("failed to remove existing resource: %w", err)
	}

	// Re-add from local source
	var addErr error
	switch resourceType {
	case resource.Command:
		addErr = manager.AddCommand(localPath, meta.SourceURL, meta.SourceType)
	case resource.Skill:
		addErr = manager.AddSkill(localPath, meta.SourceURL, meta.SourceType)
	case resource.Agent:
		addErr = manager.AddAgent(localPath, meta.SourceURL, meta.SourceType)
	default:
		addErr = fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	return false, addErr
}

// updateCommandFromClone updates a command from a cloned repository
func updateCommandFromClone(manager *repo.Manager, name, searchPath, sourceURL, sourceType string) error {
	// Find the command file
	commandPath, err := findCommandFile(searchPath, name)
	if err != nil {
		return fmt.Errorf("command not found in source: %w", err)
	}

	// Remove existing command
	if err := manager.Remove(name, resource.Command); err != nil {
		return fmt.Errorf("failed to remove existing command: %w", err)
	}

	// Add updated command
	return manager.AddCommand(commandPath, sourceURL, sourceType)
}

// updateSkillFromClone updates a skill from a cloned repository
func updateSkillFromClone(manager *repo.Manager, name, searchPath, sourceURL, sourceType string) error {
	// Find the skill directory
	skillPath, err := findSkillDir(searchPath, name)
	if err != nil {
		return fmt.Errorf("skill not found in source: %w", err)
	}

	// Remove existing skill
	if err := manager.Remove(name, resource.Skill); err != nil {
		return fmt.Errorf("failed to remove existing skill: %w", err)
	}

	// Add updated skill
	return manager.AddSkill(skillPath, sourceURL, sourceType)
}

// updateAgentFromClone updates an agent from a cloned repository
func updateAgentFromClone(manager *repo.Manager, name, searchPath, sourceURL, sourceType string) error {
	// Find the agent file
	agentPath, err := findAgentFile(searchPath, name)
	if err != nil {
		return fmt.Errorf("agent not found in source: %w", err)
	}

	// Remove existing agent
	if err := manager.Remove(name, resource.Agent); err != nil {
		return fmt.Errorf("failed to remove existing agent: %w", err)
	}

	// Add updated agent
	return manager.AddAgent(agentPath, sourceURL, sourceType)
}

// resourceInfo holds information about a resource for grouping purposes
type resourceInfo struct {
	name         string
	resourceType resource.ResourceType
	metadata     *metadata.ResourceMetadata
}

// groupResourcesBySource groups resources by their source URL for Git sources
// and separates local/file sources that should not be batched.
//
// Returns:
//   - map[string][]resourceInfo: Git sources grouped by URL
//   - []resourceInfo: Local/file sources (not grouped)
//   - error: Critical error if grouping fails
//
// Resources with missing metadata are skipped gracefully (not treated as errors).
func groupResourcesBySource(manager *repo.Manager, resources []string) (map[string][]resourceInfo, []resourceInfo, error) {
	gitSources := make(map[string][]resourceInfo)
	var localSources []resourceInfo

	for _, resArg := range resources {
		// Parse the resource argument (format: type/name)
		resType, name, err := ParseResourceArg(resArg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse resource argument %q: %w", resArg, err)
		}

		// Load metadata for the resource
		meta, err := manager.GetMetadata(name, resType)
		if err != nil {
			// Skip resources with missing metadata gracefully
			continue
		}

		info := resourceInfo{
			name:         name,
			resourceType: resType,
			metadata:     meta,
		}

		// Group by source type
		switch meta.SourceType {
		case "github", "git-url", "gitlab":
			// Group Git sources by their URL
			gitSources[meta.SourceURL] = append(gitSources[meta.SourceURL], info)
		case "local", "file":
			// Local/file sources are not batched
			localSources = append(localSources, info)
		default:
			// Unknown source types are treated as local (not batched)
			localSources = append(localSources, info)
		}
	}

	return gitSources, localSources, nil
}

// displayUpdateSummary displays a summary of update operations
func displayUpdateSummary(results []UpdateResult) {
	if len(results) == 0 {
		fmt.Println("No resources to update")
		return
	}

	successCount := 0
	failCount := 0
	skipCount := 0

	// Count results (don't display individual results - already shown inline)
	for _, result := range results {
		if result.Success {
			successCount++
		} else if result.Skipped {
			skipCount++
		} else {
			failCount++
		}
	}

	// Display summary
	if updateDryRunFlag {
		fmt.Printf("Summary (dry run): %d would be updated, %d would fail, %d would be skipped\n", successCount, failCount, skipCount)
	} else {
		fmt.Printf("Summary: %d updated, %d failed, %d skipped\n", successCount, failCount, skipCount)
	}

	// Display hint if there are skipped resources
	if skipCount > 0 {
		fmt.Println()
		fmt.Println("Hint: Run 'aimgr repo prune' to clean up orphaned metadata for missing source paths")
	}
}
