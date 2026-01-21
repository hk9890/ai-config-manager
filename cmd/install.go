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

// Completion functions for shell tab completion

// completeCommandNames provides completion for command names from the repository
func completeCommandNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	manager, err := repo.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	commandType := resource.Command
	resources, err := manager.List(&commandType)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, res := range resources {
		names = append(names, res.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeSkillNames provides completion for skill names from the repository
func completeSkillNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	manager, err := repo.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	skillType := resource.Skill
	resources, err := manager.List(&skillType)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, res := range resources {
		names = append(names, res.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeAgentNames provides completion for agent names from the repository
func completeAgentNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	manager, err := repo.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	agentType := resource.Agent
	resources, err := manager.List(&agentType)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, res := range resources {
		names = append(names, res.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [command|skill|agent]",
	Short: "Install a resource to a project",
	Long: `Install a command, skill, or agent resource from the repository to a project.

Multi-tool behavior:
  - If tool directories exist (.claude, .opencode, .github/skills), installs to ALL of them
  - If no tool directories exist, creates and installs to your default tool
  - Default tool is configured in ~/.ai-repo.yaml (use 'ai-repo config set default-tool <tool>')

Supported tools:
  - claude:   Claude Code (.claude/commands, .claude/skills, .claude/agents)
  - opencode: OpenCode (.opencode/commands, .opencode/skills, .opencode/agents)
  - copilot:  GitHub Copilot (.github/skills only - no commands or agents support)`,
}

// installCommandCmd represents the install command subcommand
var installCommandCmd = &cobra.Command{
	Use:   "command <name>",
	Short: "Install a command resource",
	Long: `Install a command resource from the repository to the current project.

Creates symlinks in tool-specific directories. If multiple AI tools are detected
in your project (e.g., both .claude and .opencode directories exist), the command
will be installed to all of them.

Note: GitHub Copilot does not support commands, only skills.

Examples:
  # Install to current project (uses detected tools or default)
  ai-repo install command my-command
  
  # Install to specific project
  ai-repo install command test --project-path ~/my-project
  
  # Install to specific target tool
  ai-repo install command my-command --target claude
  
  # Install to multiple specific tools
  ai-repo install command my-command --target claude,opencode
  
  # Force reinstall
  ai-repo install command review --force
  
  # Set your default tool
  ai-repo config set default-tool opencode`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeCommandNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

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
			return err
		}

		// Verify command exists in repo
		res, err := manager.Get(name, resource.Command)
		if err != nil {
			return fmt.Errorf("command '%s' not found in repository. Use 'ai-repo list' to see available resources.", name)
		}

		// Check if already installed
		if !installForceFlag && installer.IsInstalled(name, resource.Command) {
			return fmt.Errorf("command '%s' is already installed (use --force to reinstall)", name)
		}

		// Remove existing if force mode
		if installForceFlag && installer.IsInstalled(name, resource.Command) {
			if err := installer.Uninstall(name, resource.Command); err != nil {
				return fmt.Errorf("failed to remove existing installation: %w", err)
			}
		}

		// Install command
		if err := installer.InstallCommand(name, manager); err != nil {
			return fmt.Errorf("failed to install command: %w", err)
		}

		// Success message
		targetTools := installer.GetTargetTools()
		fmt.Printf("✓ Installed command '%s'\n", name)
		for _, tool := range targetTools {
			toolInfo := tools.GetToolInfo(tool)
			if toolInfo.SupportsCommands {
				installPath := fmt.Sprintf("%s/%s.md", toolInfo.CommandsDir, name)
				fmt.Printf("  → %s\n", installPath)
			}
		}
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

// installSkillCmd represents the install skill subcommand
var installSkillCmd = &cobra.Command{
	Use:   "skill <name>",
	Short: "Install a skill resource",
	Long: `Install a skill resource from the repository to the current project.

Creates symlinks in tool-specific directories. If multiple AI tools are detected
in your project (e.g., both .claude and .opencode directories exist), the skill
will be installed to all of them.

All three supported tools (Claude Code, OpenCode, and GitHub Copilot) support skills.

Examples:
  # Install to current project (uses detected tools or default)
  ai-repo install skill pdf-processing
  
  # Install to specific project
  ai-repo install skill my-skill --project-path ~/my-project
  
  # Install to specific target tool
  ai-repo install skill my-skill --target copilot
  
  # Install to multiple specific tools
  ai-repo install skill my-skill --target claude,opencode,copilot
  
  # Force reinstall
  ai-repo install skill utils --force
  
  # Set your default tool
  ai-repo config set default-tool copilot`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

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
			return err
		}

		// Verify skill exists in repo
		res, err := manager.Get(name, resource.Skill)
		if err != nil {
			return fmt.Errorf("skill '%s' not found in repository. Use 'ai-repo list' to see available resources.", name)
		}

		// Check if already installed
		if !installForceFlag && installer.IsInstalled(name, resource.Skill) {
			return fmt.Errorf("skill '%s' is already installed (use --force to reinstall)", name)
		}

		// Remove existing if force mode
		if installForceFlag && installer.IsInstalled(name, resource.Skill) {
			if err := installer.Uninstall(name, resource.Skill); err != nil {
				return fmt.Errorf("failed to remove existing installation: %w", err)
			}
		}

		// Install skill
		if err := installer.InstallSkill(name, manager); err != nil {
			return fmt.Errorf("failed to install skill: %w", err)
		}

		// Success message
		targetTools := installer.GetTargetTools()
		fmt.Printf("✓ Installed skill '%s'\n", name)
		for _, tool := range targetTools {
			toolInfo := tools.GetToolInfo(tool)
			if toolInfo.SupportsSkills {
				installPath := fmt.Sprintf("%s/%s", toolInfo.SkillsDir, name)
				fmt.Printf("  → %s\n", installPath)
			}
		}
		if res.Version != "" {
			fmt.Printf("  Version: %s\n", res.Version)
		}
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

// installAgentCmd represents the install agent subcommand
var installAgentCmd = &cobra.Command{
	Use:   "agent <name>",
	Short: "Install an agent resource",
	Long: `Install an agent resource from the repository to the current project.

Creates symlinks in tool-specific directories. If multiple AI tools are detected
in your project (e.g., both .claude and .opencode directories exist), the agent
will be installed to all of them.

Note: GitHub Copilot does not support agents. Claude Code and OpenCode support agents.

Examples:
  # Install to current project (uses detected tools or default)
  ai-repo install agent beads-task-agent
  
  # Install to specific project
  ai-repo install agent my-agent --project-path ~/my-project
  
  # Install to specific target tool
  ai-repo install agent my-agent --target opencode
  
  # Install to multiple specific tools
  ai-repo install agent my-agent --target claude,opencode
  
  # Force reinstall
  ai-repo install agent beads-verify-agent --force
  
  # Set your default tool
  ai-repo config set default-tool claude`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeAgentNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

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
			return err
		}

		// Verify agent exists in repo
		res, err := manager.Get(name, resource.Agent)
		if err != nil {
			return fmt.Errorf("agent '%s' not found in repository. Use 'ai-repo list' to see available resources.", name)
		}

		// Check if already installed
		if !installForceFlag && installer.IsInstalled(name, resource.Agent) {
			return fmt.Errorf("agent '%s' is already installed (use --force to reinstall)", name)
		}

		// Remove existing if force mode
		if installForceFlag && installer.IsInstalled(name, resource.Agent) {
			if err := installer.Uninstall(name, resource.Agent); err != nil {
				return fmt.Errorf("failed to remove existing installation: %w", err)
			}
		}

		// Install agent
		if err := installer.InstallAgent(name, manager); err != nil {
			return fmt.Errorf("failed to install agent: %w", err)
		}

		// Success message
		targetTools := installer.GetTargetTools()
		fmt.Printf("✓ Installed agent '%s'\n", name)
		for _, tool := range targetTools {
			toolInfo := tools.GetToolInfo(tool)
			if toolInfo.SupportsAgents {
				installPath := fmt.Sprintf("%s/%s.md", toolInfo.AgentsDir, name)
				fmt.Printf("  → %s\n", installPath)
			}
		}
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installCommandCmd)
	installCmd.AddCommand(installSkillCmd)
	installCmd.AddCommand(installAgentCmd)

	// Add flags to all subcommands
	installCommandCmd.Flags().StringVar(&projectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	installCommandCmd.Flags().BoolVarP(&installForceFlag, "force", "f", false, "Overwrite existing installation")
	installCommandCmd.Flags().StringVar(&installTargetFlag, "target", "", "Target tools (comma-separated: claude,opencode,copilot)")

	installSkillCmd.Flags().StringVar(&projectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	installSkillCmd.Flags().BoolVarP(&installForceFlag, "force", "f", false, "Overwrite existing installation")
	installSkillCmd.Flags().StringVar(&installTargetFlag, "target", "", "Target tools (comma-separated: claude,opencode,copilot)")

	installAgentCmd.Flags().StringVar(&projectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	installAgentCmd.Flags().BoolVarP(&installForceFlag, "force", "f", false, "Overwrite existing installation")
	installAgentCmd.Flags().StringVar(&installTargetFlag, "target", "", "Target tools (comma-separated: claude,opencode,copilot)")
}
