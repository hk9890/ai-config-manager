package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/install"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	uninstallProjectPathFlag string
	uninstallForceFlag       bool
)

// uninstallResult tracks the result of uninstalling a single resource
type uninstallResult struct {
	resourceType resource.ResourceType
	name         string
	success      bool
	skipped      bool
	message      string
	toolsRemoved []tools.Tool
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall <resource>...",
	Short: "Uninstall resources from a project",
	Long: `Uninstall one or more resources (commands, skills, or agents) from a project.

Resources are specified using the format 'type/name':
  - command/name (or commands/name)
  - skill/name (or skills/name)
  - agent/name (or agents/name)

Safety:
  - Only removes symlinks that point to the aimgr repository
  - Skips non-symlinks with a warning
  - Skips symlinks pointing to other locations with a warning

Multi-tool behavior:
  - Automatically detects all tool directories (.claude, .opencode, .github/skills)
  - Removes resource from ALL detected tool directories
  - Shows summary of what was removed from where

Examples:
  # Uninstall a single skill
  aimgr uninstall skill/pdf-processing

  # Uninstall multiple resources at once
  aimgr uninstall skill/foo skill/bar command/test agent/my-agent

  # Uninstall from a specific project
  aimgr uninstall skill/foo --project-path ~/my-project

  # Force uninstall (placeholder for future confirmation prompts)
  aimgr uninstall command/review --force`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeResourceArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path (current directory or flag)
		projectPath := uninstallProjectPathFlag
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		// Create repo manager to get repository path
		manager, err := repo.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create repository manager: %w", err)
		}
		repoPath := manager.GetRepoPath()

		// Create installer (auto-detect existing tools)
		installer, err := install.NewInstaller(projectPath, nil)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		// Track results
		var results []uninstallResult

		// Process each resource argument
		for _, arg := range args {
			result := processUninstall(arg, projectPath, repoPath, installer.GetTargetTools())
			results = append(results, result)
		}

		// Print results
		printUninstallSummary(results)

		// Return error if any resource failed
		for _, result := range results {
			if !result.success && !result.skipped {
				return fmt.Errorf("some resources failed to uninstall")
			}
		}

		return nil
	},
}

// processUninstall processes uninstalling a single resource
func processUninstall(arg string, projectPath string, repoPath string, targetTools []tools.Tool) uninstallResult {
	// Parse resource argument
	resourceType, name, err := parseResourceArg(arg)
	if err != nil {
		return uninstallResult{
			name:    arg,
			success: false,
			message: err.Error(),
		}
	}

	result := uninstallResult{
		resourceType: resourceType,
		name:         name,
		toolsRemoved: []tools.Tool{},
	}

	removed := false
	skipped := false
	var messages []string

	// Try to uninstall from all target tools
	for _, tool := range targetTools {
		toolInfo := tools.GetToolInfo(tool)
		var symlinkPath string

		// Determine symlink path based on resource type
		switch resourceType {
		case resource.Command:
			if !toolInfo.SupportsCommands {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.CommandsDir, name+".md")
		case resource.Skill:
			if !toolInfo.SupportsSkills {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.SkillsDir, name)
		case resource.Agent:
			if !toolInfo.SupportsAgents {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.AgentsDir, name+".md")
		default:
			result.success = false
			result.message = fmt.Sprintf("invalid resource type: %s", resourceType)
			return result
		}

		// Check if symlink exists
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			// Doesn't exist in this tool directory, continue
			continue
		}

		// Verify it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			messages = append(messages, fmt.Sprintf("%s: not a symlink (skipping)", tool))
			skipped = true
			continue
		}

		// Read symlink target
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			messages = append(messages, fmt.Sprintf("%s: failed to read symlink: %v", tool, err))
			continue
		}

		// Resolve to absolute path for comparison
		absTarget, err := filepath.Abs(target)
		if err != nil {
			// If we can't resolve, use the original target
			absTarget = target
		}

		// Check if target points to our repository
		if !strings.HasPrefix(absTarget, repoPath) {
			messages = append(messages, fmt.Sprintf("%s: not managed by aimgr (points to %s, skipping)", tool, target))
			skipped = true
			continue
		}

		// Remove the symlink
		if err := os.Remove(symlinkPath); err != nil {
			messages = append(messages, fmt.Sprintf("%s: failed to remove: %v", tool, err))
			continue
		}

		removed = true
		result.toolsRemoved = append(result.toolsRemoved, tool)
	}

	// Determine overall result
	if removed {
		result.success = true
		if len(messages) > 0 {
			result.message = strings.Join(messages, "; ")
		}
	} else if skipped {
		result.skipped = true
		result.message = strings.Join(messages, "; ")
	} else {
		result.success = false
		if len(messages) > 0 {
			result.message = strings.Join(messages, "; ")
		} else {
			result.message = fmt.Sprintf("%s/%s not installed in project", resourceType, name)
		}
	}

	return result
}

// printUninstallSummary prints a summary of uninstall results
func printUninstallSummary(results []uninstallResult) {
	successCount := 0
	skipCount := 0
	failCount := 0

	for _, result := range results {
		if result.success {
			successCount++
			// Print success
			fmt.Printf("✓ Uninstalled %s '%s'\n", result.resourceType, result.name)
			for _, tool := range result.toolsRemoved {
				toolInfo := tools.GetToolInfo(tool)
				var dirName string
				switch result.resourceType {
				case resource.Command:
					dirName = toolInfo.CommandsDir
				case resource.Skill:
					dirName = toolInfo.SkillsDir
				case resource.Agent:
					dirName = toolInfo.AgentsDir
				}
				fmt.Printf("  → Removed from %s (%s)\n", tool, dirName)
			}
			if result.message != "" {
				fmt.Printf("  Note: %s\n", result.message)
			}
		} else if result.skipped {
			skipCount++
			// Print skipped
			fmt.Printf("⊘ Skipped %s '%s': %s\n", result.resourceType, result.name, result.message)
		} else {
			failCount++
			// Print failure
			if result.resourceType != "" {
				fmt.Printf("✗ Failed to uninstall %s '%s': %s\n", result.resourceType, result.name, result.message)
			} else {
				fmt.Printf("✗ Failed: %s\n", result.message)
			}
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %d uninstalled, %d skipped, %d failed\n", successCount, skipCount, failCount)
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringVar(&uninstallProjectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	uninstallCmd.Flags().BoolVarP(&uninstallForceFlag, "force", "f", false, "Force uninstall (placeholder for future use)")
}
