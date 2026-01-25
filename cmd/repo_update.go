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

		if len(args) == 0 {
			// Count total resources for progress tracking
			totalCount := 0
			typeFilter := resource.Skill
			skills, _ := manager.List(&typeFilter)
			totalCount += len(skills)

			typeFilter = resource.Command
			commands, _ := manager.List(&typeFilter)
			totalCount += len(commands)

			typeFilter = resource.Agent
			agents, _ := manager.List(&typeFilter)
			totalCount += len(agents)

			ctx.Total = totalCount

			if totalCount > 0 {
				fmt.Printf("Updating %d resources...\n\n", totalCount)
			}

			// Update all resources
			skillResults, err := updateResourceTypeWithProgress(manager, resource.Skill, "", &ctx)
			if err != nil {
				return err
			}
			results = append(results, skillResults...)

			commandResults, err := updateResourceTypeWithProgress(manager, resource.Command, "", &ctx)
			if err != nil {
				return err
			}
			results = append(results, commandResults...)

			agentResults, err := updateResourceTypeWithProgress(manager, resource.Agent, "", &ctx)
			if err != nil {
				return err
			}
			results = append(results, agentResults...)
		} else {
			// Update by patterns
			var toUpdate []string
			for _, pattern := range args {
				matches, err := ExpandPattern(manager, pattern)
				if err != nil {
					return err
				}
				toUpdate = append(toUpdate, matches...)
			}

			// Remove duplicates
			toUpdate = uniqueStrings(toUpdate)

			ctx.Total = len(toUpdate)

			if len(toUpdate) > 0 {
				fmt.Printf("Updating %d resources...\n\n", len(toUpdate))
			}

			// Update each resource
			for _, tu := range toUpdate {
				resType, name, err := ParseResourceArg(tu)
				if err != nil {
					return err
				}
				ctx.Current++
				result := updateSingleResourceWithProgress(manager, name, resType, &ctx)
				results = append(results, result)
			}
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
		fmt.Printf("  ↓ Cloning from %s...\n", meta.SourceURL)
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

	// Clone repository to temp directory
	cloneURL, err := source.GetCloneURL(parsed)
	if err != nil {
		return fmt.Errorf("failed to get clone URL: %w", err)
	}

	tempDir, err := source.CloneRepo(cloneURL, parsed.Ref)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	defer source.CleanupTempDir(tempDir)

	// Discover resources in the cloned repo
	searchPath := tempDir
	if parsed.Subpath != "" {
		searchPath = filepath.Join(tempDir, parsed.Subpath)
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
