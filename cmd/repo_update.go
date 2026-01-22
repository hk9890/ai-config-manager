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
	Message string
}

// repoUpdateCmd represents the update command group
var repoUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update resources from their original sources",
	Long: `Update resources from their original sources.

Updates can refresh resources from GitHub repositories, local paths, or file sources.
The source information is retrieved from the resource metadata.

Examples:
  aimgr repo update                    # Update all resources
  aimgr repo update skill my-skill     # Update specific skill
  aimgr repo update command my-cmd     # Update specific command
  aimgr repo update agent my-agent     # Update specific agent
  aimgr repo update --dry-run          # Preview what would be updated
  aimgr repo update --force            # Force update even with local changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Update all resources (skills, commands, agents)
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var results []UpdateResult

		// Update skills
		skillResults, err := updateResourceType(manager, resource.Skill, "")
		if err != nil {
			return err
		}
		results = append(results, skillResults...)

		// Update commands
		commandResults, err := updateResourceType(manager, resource.Command, "")
		if err != nil {
			return err
		}
		results = append(results, commandResults...)

		// Update agents
		agentResults, err := updateResourceType(manager, resource.Agent, "")
		if err != nil {
			return err
		}
		results = append(results, agentResults...)

		// Display summary
		displayUpdateSummary(results)

		return nil
	},
}

// repoUpdateSkillCmd handles updating skills
var repoUpdateSkillCmd = &cobra.Command{
	Use:   "skill [name]",
	Short: "Update skill(s) from their original source",
	Long: `Update a specific skill or all skills from their original sources.

If a skill name is provided, only that skill is updated.
If no name is provided, all skills are updated.

Examples:
  aimgr repo update skill my-skill     # Update specific skill
  aimgr repo update skill              # Update all skills
  aimgr repo update skill my-skill --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		results, err := updateResourceType(manager, resource.Skill, name)
		if err != nil {
			return err
		}

		displayUpdateSummary(results)
		return nil
	},
}

// repoUpdateCommandCmd handles updating commands
var repoUpdateCommandCmd = &cobra.Command{
	Use:   "command [name]",
	Short: "Update command(s) from their original source",
	Long: `Update a specific command or all commands from their original sources.

If a command name is provided, only that command is updated.
If no name is provided, all commands are updated.

Examples:
  aimgr repo update command my-cmd     # Update specific command
  aimgr repo update command            # Update all commands
  aimgr repo update command my-cmd --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		results, err := updateResourceType(manager, resource.Command, name)
		if err != nil {
			return err
		}

		displayUpdateSummary(results)
		return nil
	},
}

// repoUpdateAgentCmd handles updating agents
var repoUpdateAgentCmd = &cobra.Command{
	Use:   "agent [name]",
	Short: "Update agent(s) from their original source",
	Long: `Update a specific agent or all agents from their original sources.

If an agent name is provided, only that agent is updated.
If no name is provided, all agents are updated.

Examples:
  aimgr repo update agent my-agent     # Update specific agent
  aimgr repo update agent              # Update all agents
  aimgr repo update agent my-agent --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		results, err := updateResourceType(manager, resource.Agent, name)
		if err != nil {
			return err
		}

		displayUpdateSummary(results)
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoUpdateCmd)
	repoUpdateCmd.AddCommand(repoUpdateSkillCmd)
	repoUpdateCmd.AddCommand(repoUpdateCommandCmd)
	repoUpdateCmd.AddCommand(repoUpdateAgentCmd)

	// Add flags to all update commands
	for _, cmd := range []*cobra.Command{repoUpdateCmd, repoUpdateSkillCmd, repoUpdateCommandCmd, repoUpdateAgentCmd} {
		cmd.Flags().BoolVar(&updateForceFlag, "force", false, "Force update, overwriting local changes")
		cmd.Flags().BoolVar(&updateDryRunFlag, "dry-run", false, "Preview updates without making changes")
	}
}

// updateResourceType updates all resources of a specific type or a specific resource by name
func updateResourceType(manager *repo.Manager, resourceType resource.ResourceType, name string) ([]UpdateResult, error) {
	var results []UpdateResult

	if name != "" {
		// Update specific resource
		result := updateSingleResource(manager, name, resourceType)
		results = append(results, result)
	} else {
		// Update all resources of this type
		typeFilter := resourceType
		resources, err := manager.List(&typeFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to list %ss: %w", resourceType, err)
		}

		for _, res := range resources {
			result := updateSingleResource(manager, res.Name, resourceType)
			results = append(results, result)
		}
	}

	return results, nil
}

// updateSingleResource updates a single resource from its original source
func updateSingleResource(manager *repo.Manager, name string, resourceType resource.ResourceType) UpdateResult {
	result := UpdateResult{
		Name:    name,
		Type:    resourceType,
		Success: false,
	}

	// Load metadata
	meta, err := manager.GetMetadata(name, resourceType)
	if err != nil {
		result.Message = fmt.Sprintf("Metadata not found: %v", err)
		return result
	}

	// Dry run mode - just report what would be done
	if updateDryRunFlag {
		result.Success = true
		result.Message = fmt.Sprintf("Would update from %s (%s)", meta.SourceURL, meta.SourceType)
		return result
	}

	// Update based on source type
	switch meta.SourceType {
	case "github", "git-url", "gitlab":
		err = updateFromGitSource(manager, name, resourceType, meta)
	case "local", "file":
		err = updateFromLocalSource(manager, name, resourceType, meta)
	default:
		result.Message = fmt.Sprintf("Unknown source type: %s", meta.SourceType)
		return result
	}

	if err != nil {
		result.Message = err.Error()
		return result
	}

	// Update LastUpdated timestamp
	meta.LastUpdated = time.Now()
	if err := metadata.Save(meta, manager.GetRepoPath()); err != nil {
		result.Message = fmt.Sprintf("Updated but failed to save metadata: %v", err)
		return result
	}

	result.Success = true
	result.Message = "Updated successfully"
	return result
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
func updateFromLocalSource(manager *repo.Manager, name string, resourceType resource.ResourceType, meta *metadata.ResourceMetadata) error {
	// Extract local path from source URL (format: file:///path/to/resource)
	localPath := meta.SourceURL
	if strings.HasPrefix(localPath, "file://") {
		localPath = filepath.Clean(localPath[7:]) // Remove "file://" prefix
	}

	// Check if source still exists
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("source path no longer exists: %s", localPath)
	}

	// Remove existing resource (force mode is implicit for update)
	if err := manager.Remove(name, resourceType); err != nil {
		return fmt.Errorf("failed to remove existing resource: %w", err)
	}

	// Re-add from local source
	switch resourceType {
	case resource.Command:
		return manager.AddCommand(localPath, meta.SourceURL, meta.SourceType)
	case resource.Skill:
		return manager.AddSkill(localPath, meta.SourceURL, meta.SourceType)
	case resource.Agent:
		return manager.AddAgent(localPath, meta.SourceURL, meta.SourceType)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
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

	// Display individual results
	for _, result := range results {
		if result.Success {
			if updateDryRunFlag {
				fmt.Printf("✓ [DRY RUN] %s '%s': %s\n", result.Type, result.Name, result.Message)
			} else {
				fmt.Printf("✓ %s '%s': %s\n", result.Type, result.Name, result.Message)
			}
			successCount++
		} else {
			fmt.Printf("✗ %s '%s': %s\n", result.Type, result.Name, result.Message)
			failCount++
		}
	}

	// Display summary
	fmt.Println()
	if updateDryRunFlag {
		fmt.Printf("Summary (dry run): %d would be updated, %d would fail\n", successCount, failCount)
	} else {
		fmt.Printf("Summary: %d updated, %d failed\n", successCount, failCount)
	}
}
