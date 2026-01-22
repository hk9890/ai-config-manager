package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/config"
	"github.com/hk9890/ai-config-manager/pkg/install"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	projectPathFlag   string
	installForceFlag  bool
	installTargetFlag string
)

// installResult tracks the result of installing a single resource
type installResult struct {
	resourceType resource.ResourceType
	name         string
	success      bool
	skipped      bool
	message      string
	toolsAdded   []tools.Tool
}

// parseTargetFlag parses the --target flag and returns a list of tools
// If the flag is empty, returns nil (use defaults)
// If the flag contains values, parses and validates them
func parseTargetFlag(targetFlag string) ([]tools.Tool, error) {
	if targetFlag == "" {
		return nil, nil
	}

	var targets []tools.Tool
	targetStrs := strings.Split(targetFlag, ",")
	for _, t := range targetStrs {
		tool, err := tools.ParseTool(strings.TrimSpace(t))
		if err != nil {
			return nil, fmt.Errorf("invalid target '%s': %w", t, err)
		}
		targets = append(targets, tool)
	}

	return targets, nil
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install <resource>...",
	Short: "Install resources to a project",
	Long: `Install one or more resources (commands, skills, or agents) to a project.

Resources are specified using the format 'type/name':
  - command/name (or commands/name)
  - skill/name (or skills/name)
  - agent/name (or agents/name)

Multi-tool behavior:
  - If tool directories exist (.claude, .opencode, .github/skills), installs to ALL of them
  - If no tool directories exist, creates and installs to your default tool
  - Default tool is configured in ~/.aimgr.yaml (use 'aimgr config set default-tool <tool>')

Supported tools:
  - claude:   Claude Code (.claude/commands, .claude/skills, .claude/agents)
  - opencode: OpenCode (.opencode/commands, .opencode/skills, .opencode/agents)
  - copilot:  GitHub Copilot (.github/skills only - no commands or agents support)

Examples:
  # Install a single skill
  aimgr install skill/pdf-processing

  # Install multiple resources at once
  aimgr install skill/foo command/bar agent/reviewer

  # Install to specific project
  aimgr install skill/test --project-path ~/project

  # Force reinstall
  aimgr install command/test --force

  # Install to specific target
  aimgr install skill/utils --target claude`,
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeInstallResources,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path (current directory or flag)
		projectPath := projectPathFlag
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		// Parse target flag (if provided)
		explicitTargets, err := parseTargetFlag(installTargetFlag)
		if err != nil {
			return err
		}

		// Create installer
		var installer *install.Installer
		if explicitTargets != nil {
			// Use explicit targets from --target flag (bypass detection)
			installer, err = install.NewInstallerWithTargets(projectPath, explicitTargets)
		} else {
			// Auto-detect existing tools or use config defaults
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			cfg, err := config.Load(home)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			defaultTargets, err := cfg.GetDefaultTargets()
			if err != nil {
				return fmt.Errorf("invalid default targets in config: %w", err)
			}
			// NewInstaller will auto-detect existing tool directories
			installer, err = install.NewInstaller(projectPath, defaultTargets)
		}
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		// Create repo manager
		manager, err := repo.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create repository manager: %w", err)
		}

		// Track results
		var results []installResult

		// Process each resource argument
		for _, arg := range args {
			result := processInstall(arg, installer, manager)
			results = append(results, result)
		}

		// Print results
		printInstallSummary(results)

		// Return error if any resource failed
		for _, result := range results {
			if !result.success && !result.skipped {
				return fmt.Errorf("some resources failed to install")
			}
		}

		return nil
	},
}

// processInstall processes installing a single resource
func processInstall(arg string, installer *install.Installer, manager *repo.Manager) installResult {
	// Parse resource argument
	resourceType, name, err := parseResourceArg(arg)
	if err != nil {
		return installResult{
			name:    arg,
			success: false,
			message: err.Error(),
		}
	}

	result := installResult{
		resourceType: resourceType,
		name:         name,
		toolsAdded:   []tools.Tool{},
	}

	// Verify resource exists in repo
	res, err := manager.Get(name, resourceType)
	if err != nil {
		result.success = false
		result.message = fmt.Sprintf("%s '%s' not found in repository. Use 'aimgr list' to see available resources.", resourceType, name)
		return result
	}

	// Check if already installed
	if !installForceFlag && installer.IsInstalled(name, resourceType) {
		result.skipped = true
		result.message = "already installed (use --force to reinstall)"
		return result
	}

	// Remove existing if force mode
	if installForceFlag && installer.IsInstalled(name, resourceType) {
		if err := installer.Uninstall(name, resourceType); err != nil {
			result.success = false
			result.message = fmt.Sprintf("failed to remove existing installation: %v", err)
			return result
		}
	}

	// Install based on resource type
	var installErr error
	switch resourceType {
	case resource.Command:
		installErr = installer.InstallCommand(name, manager)
	case resource.Skill:
		installErr = installer.InstallSkill(name, manager)
	case resource.Agent:
		installErr = installer.InstallAgent(name, manager)
	default:
		result.success = false
		result.message = fmt.Sprintf("unsupported resource type: %s", resourceType)
		return result
	}

	if installErr != nil {
		result.success = false
		result.message = fmt.Sprintf("failed to install: %v", installErr)
		return result
	}

	// Success
	result.success = true
	result.toolsAdded = installer.GetTargetTools()

	// Add description/version to message if available
	var metaParts []string
	if res.Version != "" {
		metaParts = append(metaParts, fmt.Sprintf("Version: %s", res.Version))
	}
	if res.Description != "" {
		metaParts = append(metaParts, fmt.Sprintf("Description: %s", res.Description))
	}
	if len(metaParts) > 0 {
		result.message = strings.Join(metaParts, ", ")
	}

	return result
}

// printInstallSummary prints a summary of install results
func printInstallSummary(results []installResult) {
	successCount := 0
	skipCount := 0
	failCount := 0

	for _, result := range results {
		if result.success {
			successCount++
			// Print success
			fmt.Printf("✓ Installed %s '%s'\n", result.resourceType, result.name)
			for _, tool := range result.toolsAdded {
				toolInfo := tools.GetToolInfo(tool)
				var installPath string
				switch result.resourceType {
				case resource.Command:
					if toolInfo.SupportsCommands {
						installPath = fmt.Sprintf("%s/%s.md", toolInfo.CommandsDir, result.name)
					}
				case resource.Skill:
					if toolInfo.SupportsSkills {
						installPath = fmt.Sprintf("%s/%s", toolInfo.SkillsDir, result.name)
					}
				case resource.Agent:
					if toolInfo.SupportsAgents {
						installPath = fmt.Sprintf("%s/%s.md", toolInfo.AgentsDir, result.name)
					}
				}
				if installPath != "" {
					fmt.Printf("  → %s\n", installPath)
				}
			}
			if result.message != "" {
				fmt.Printf("  %s\n", result.message)
			}
		} else if result.skipped {
			skipCount++
			// Print skipped
			fmt.Printf("⊘ Skipped %s '%s': %s\n", result.resourceType, result.name, result.message)
		} else {
			failCount++
			// Print failure
			if result.resourceType != "" {
				fmt.Printf("✗ Failed to install %s '%s': %s\n", result.resourceType, result.name, result.message)
			} else {
				fmt.Printf("✗ Failed: %s\n", result.message)
			}
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %d installed, %d skipped, %d failed\n", successCount, skipCount, failCount)
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Add flags to install command
	installCmd.Flags().StringVar(&projectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	installCmd.Flags().BoolVarP(&installForceFlag, "force", "f", false, "Overwrite existing installation")
	installCmd.Flags().StringVar(&installTargetFlag, "target", "", "Target tools (comma-separated: claude,opencode,copilot)")
}
