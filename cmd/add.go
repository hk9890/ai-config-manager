package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/spf13/cobra"
)

var forceFlag bool

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [command|skill|agent]",
	Short: "Add a resource to the repository",
	Long: `Add a command, skill, or agent resource to the aimgr repository.

Commands are single .md files with YAML frontmatter.
Skills are directories containing a SKILL.md file.
Agents are single .md files with YAML frontmatter.`,
}

// addCommandCmd represents the add command subcommand
var addCommandCmd = &cobra.Command{
	Use:   "command <source>",
	Short: "Add a command resource",
	Long: `Add a command resource to the repository.

A command is a single .md file with YAML frontmatter containing at minimum
a description field.

Source Formats:
  gh:owner/repo              GitHub repository (auto-discovers commands)
  gh:owner/repo/path         Specific command in repo
  local:path or path         Local .md file
  owner/repo                 GitHub shorthand

Examples:
  aimgr repo add command gh:myorg/commands
  aimgr repo add command gh:myorg/repo/commands/test-command.md
  aimgr repo add command ~/.claude/commands/my-command.md
  aimgr repo add command ./my-command.md --force
  aimgr repo add command myorg/commands`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceInput := args[0]

		// Parse source
		parsed, err := source.ParseSource(sourceInput)
		if err != nil {
			return fmt.Errorf("invalid source format: %w", err)
		}

		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Handle different source types
		if parsed.Type == source.GitHub || parsed.Type == source.GitURL {
			// GitHub/Git source - use auto-discovery workflow
			return addCommandFromGitHub(parsed, manager)
		}

		// Local source - existing behavior
		sourcePath := parsed.LocalPath

		// Validate path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return fmt.Errorf("path does not exist: %s", sourcePath)
		}

		// Validate it's a file
		info, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, expected a .md file: %s", sourcePath)
		}

		// Validate it's a .md file
		if filepath.Ext(sourcePath) != ".md" {
			return fmt.Errorf("file must have .md extension: %s", sourcePath)
		}

		// Try to load and validate the command
		res, err := resource.LoadCommand(sourcePath)
		if err != nil {
			return fmt.Errorf("invalid command resource: %w", err)
		}

		// Check if already exists (if not force mode)
		if !forceFlag {
			existing, _ := manager.Get(res.Name, resource.Command)
			if existing != nil {
				return fmt.Errorf("command '%s' already exists in repository (use --force to overwrite)", res.Name)
			}
		} else {
			// Remove existing if force mode
			_ = manager.Remove(res.Name, resource.Command)
		}

		// Add the command
		if err := manager.AddCommand(sourcePath); err != nil {
			return fmt.Errorf("failed to add command: %w", err)
		}

		// Success message
		fmt.Printf("✓ Added command '%s' to repository\n", res.Name)
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

// addSkillCmd represents the add skill subcommand
var addSkillCmd = &cobra.Command{
	Use:   "skill <source>",
	Short: "Add a skill resource",
	Long: `Add a skill resource to the repository.

A skill is a directory containing a SKILL.md file with YAML frontmatter.
The directory name must match the 'name' field in SKILL.md.

Source Formats:
  gh:owner/repo              GitHub repository (auto-discovers skills)
  gh:owner/repo/path         Specific skill in repo
  local:path or path         Local directory
  owner/repo                 GitHub shorthand

Examples:
  aimgr repo add skill gh:myorg/skills
  aimgr repo add skill gh:myorg/repo/skills/pdf-processing
  aimgr repo add skill ~/my-skills/pdf-processing
  aimgr repo add skill ./my-skill --force
  aimgr repo add skill myorg/skills`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceInput := args[0]

		// Parse source
		parsed, err := source.ParseSource(sourceInput)
		if err != nil {
			return fmt.Errorf("invalid source format: %w", err)
		}

		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Handle different source types
		if parsed.Type == source.GitHub || parsed.Type == source.GitURL {
			// GitHub/Git source - use auto-discovery workflow
			return addSkillFromGitHub(parsed, manager)
		}

		// Local source - existing behavior
		sourcePath := parsed.LocalPath

		// Validate path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return fmt.Errorf("path does not exist: %s", sourcePath)
		}

		// Validate it's a directory
		info, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("path is a file, expected a directory: %s", sourcePath)
		}

		// Check for SKILL.md
		skillMdPath := filepath.Join(sourcePath, "SKILL.md")
		if _, err := os.Stat(skillMdPath); err != nil {
			return fmt.Errorf("directory must contain SKILL.md: %s", sourcePath)
		}

		// Try to load and validate the skill
		res, err := resource.LoadSkill(sourcePath)
		if err != nil {
			return fmt.Errorf("invalid skill resource: %w", err)
		}

		// Validate folder name matches frontmatter name
		dirName := filepath.Base(sourcePath)
		if res.Name != dirName {
			return fmt.Errorf("folder name '%s' does not match skill name '%s' in SKILL.md", dirName, res.Name)
		}

		// Check if already exists (if not force mode)
		if !forceFlag {
			existing, _ := manager.Get(res.Name, resource.Skill)
			if existing != nil {
				return fmt.Errorf("skill '%s' already exists in repository (use --force to overwrite)", res.Name)
			}
		} else {
			// Remove existing if force mode
			_ = manager.Remove(res.Name, resource.Skill)
		}

		// Add the skill
		if err := manager.AddSkill(sourcePath); err != nil {
			return fmt.Errorf("failed to add skill: %w", err)
		}

		// Success message
		fmt.Printf("✓ Added skill '%s' to repository\n", res.Name)
		if res.Version != "" {
			fmt.Printf("  Version: %s\n", res.Version)
		}
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

// addAgentCmd represents the add agent subcommand
var addAgentCmd = &cobra.Command{
	Use:   "agent <source>",
	Short: "Add an agent resource",
	Long: `Add an agent resource to the repository.

An agent is a single .md file with YAML frontmatter containing at minimum
a description field.

Source Formats:
  gh:owner/repo              GitHub repository (auto-discovers agents)
  gh:owner/repo/path         Specific agent in repo
  local:path or path         Local .md file
  owner/repo                 GitHub shorthand

Examples:
  aimgr repo add agent gh:myorg/agents
  aimgr repo add agent gh:myorg/repo/agents/code-reviewer.md
  aimgr repo add agent ~/.opencode/agents/my-agent.md
  aimgr repo add agent ./my-agent.md --force
  aimgr repo add agent myorg/agents`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceInput := args[0]

		// Parse source
		parsed, err := source.ParseSource(sourceInput)
		if err != nil {
			return fmt.Errorf("invalid source format: %w", err)
		}

		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Handle different source types
		if parsed.Type == source.GitHub || parsed.Type == source.GitURL {
			// GitHub/Git source - use auto-discovery workflow
			return addAgentFromGitHub(parsed, manager)
		}

		// Local source - existing behavior
		sourcePath := parsed.LocalPath

		// Validate path exists
		if _, err := os.Stat(sourcePath); err != nil {
			return fmt.Errorf("path does not exist: %s", sourcePath)
		}

		// Validate it's a file
		info, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, expected a .md file: %s", sourcePath)
		}

		// Validate it's a .md file
		if filepath.Ext(sourcePath) != ".md" {
			return fmt.Errorf("file must have .md extension: %s", sourcePath)
		}

		// Try to load and validate the agent
		res, err := resource.LoadAgent(sourcePath)
		if err != nil {
			return fmt.Errorf("invalid agent resource: %w", err)
		}

		// Check if already exists (if not force mode)
		if !forceFlag {
			existing, _ := manager.Get(res.Name, resource.Agent)
			if existing != nil {
				return fmt.Errorf("agent '%s' already exists in repository (use --force to overwrite)", res.Name)
			}
		} else {
			// Remove existing if force mode
			_ = manager.Remove(res.Name, resource.Agent)
		}

		// Add the agent
		if err := manager.AddAgent(sourcePath); err != nil {
			return fmt.Errorf("failed to add agent: %w", err)
		}

		// Success message
		fmt.Printf("✓ Added agent '%s' to repository\n", res.Name)
		if res.Description != "" {
			fmt.Printf("  Description: %s\n", res.Description)
		}

		return nil
	},
}

func init() {
	repoCmd.AddCommand(addCmd)
	addCmd.AddCommand(addCommandCmd)
	addCmd.AddCommand(addSkillCmd)
	addCmd.AddCommand(addAgentCmd)

	// Add --force flag to all subcommands
	addCommandCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resource")
	addSkillCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resource")
	addAgentCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing resource")
}

// Helper functions for GitHub source integration

// addCommandFromGitHub handles adding a command from a GitHub source
func addCommandFromGitHub(parsed *source.ParsedSource, manager *repo.Manager) error {
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

	// Discover commands
	searchPath := tempDir
	if parsed.Subpath != "" {
		searchPath = filepath.Join(tempDir, parsed.Subpath)
	}

	commands, err := discovery.DiscoverCommands(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands found in repository: %s", parsed.URL)
	}

	// Handle resource selection
	selectedCommand, err := selectResource(commands, "command")
	if err != nil {
		return err
	}

	// Get the actual file path from the resource
	commandPath := filepath.Join(searchPath, selectedCommand.Name+".md")

	// Find the actual command file in the discovered locations
	commandPath, err = findCommandFile(searchPath, selectedCommand.Name)
	if err != nil {
		return fmt.Errorf("failed to find command file: %w", err)
	}

	// Check if already exists (if not force mode)
	if !forceFlag {
		existing, _ := manager.Get(selectedCommand.Name, resource.Command)
		if existing != nil {
			return fmt.Errorf("command '%s' already exists in repository (use --force to overwrite)", selectedCommand.Name)
		}
	} else {
		// Remove existing if force mode
		_ = manager.Remove(selectedCommand.Name, resource.Command)
	}

	// Add the command using manager
	if err := manager.AddCommand(commandPath); err != nil {
		return fmt.Errorf("failed to add command: %w", err)
	}

	// Success message
	fmt.Printf("✓ Added command '%s' to repository\n", selectedCommand.Name)
	if selectedCommand.Description != "" {
		fmt.Printf("  Description: %s\n", selectedCommand.Description)
	}

	return nil
}

// addSkillFromGitHub handles adding a skill from a GitHub source
func addSkillFromGitHub(parsed *source.ParsedSource, manager *repo.Manager) error {
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

	// Discover skills
	searchPath := tempDir
	if parsed.Subpath != "" {
		searchPath = filepath.Join(tempDir, parsed.Subpath)
	}

	skills, err := discovery.DiscoverSkills(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	if len(skills) == 0 {
		return fmt.Errorf("no skills found in repository: %s", parsed.URL)
	}

	// Handle resource selection
	selectedSkill, err := selectResource(skills, "skill")
	if err != nil {
		return err
	}

	// Find the skill directory
	skillPath, err := findSkillDir(searchPath, selectedSkill.Name)
	if err != nil {
		return fmt.Errorf("failed to find skill directory: %w", err)
	}

	// Check if already exists (if not force mode)
	if !forceFlag {
		existing, _ := manager.Get(selectedSkill.Name, resource.Skill)
		if existing != nil {
			return fmt.Errorf("skill '%s' already exists in repository (use --force to overwrite)", selectedSkill.Name)
		}
	} else {
		// Remove existing if force mode
		_ = manager.Remove(selectedSkill.Name, resource.Skill)
	}

	// Add the skill using manager
	if err := manager.AddSkill(skillPath); err != nil {
		return fmt.Errorf("failed to add skill: %w", err)
	}

	// Success message
	fmt.Printf("✓ Added skill '%s' to repository\n", selectedSkill.Name)
	if selectedSkill.Version != "" {
		fmt.Printf("  Version: %s\n", selectedSkill.Version)
	}
	if selectedSkill.Description != "" {
		fmt.Printf("  Description: %s\n", selectedSkill.Description)
	}

	return nil
}

// addAgentFromGitHub handles adding an agent from a GitHub source
func addAgentFromGitHub(parsed *source.ParsedSource, manager *repo.Manager) error {
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

	// Discover agents
	searchPath := tempDir
	if parsed.Subpath != "" {
		searchPath = filepath.Join(tempDir, parsed.Subpath)
	}

	agents, err := discovery.DiscoverAgents(searchPath, "")
	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	if len(agents) == 0 {
		return fmt.Errorf("no agents found in repository: %s", parsed.URL)
	}

	// Handle resource selection
	selectedAgent, err := selectResource(agents, "agent")
	if err != nil {
		return err
	}

	// Find the agent file
	agentPath, err := findAgentFile(searchPath, selectedAgent.Name)
	if err != nil {
		return fmt.Errorf("failed to find agent file: %w", err)
	}

	// Check if already exists (if not force mode)
	if !forceFlag {
		existing, _ := manager.Get(selectedAgent.Name, resource.Agent)
		if existing != nil {
			return fmt.Errorf("agent '%s' already exists in repository (use --force to overwrite)", selectedAgent.Name)
		}
	} else {
		// Remove existing if force mode
		_ = manager.Remove(selectedAgent.Name, resource.Agent)
	}

	// Add the agent using manager
	if err := manager.AddAgent(agentPath); err != nil {
		return fmt.Errorf("failed to add agent: %w", err)
	}

	// Success message
	fmt.Printf("✓ Added agent '%s' to repository\n", selectedAgent.Name)
	if selectedAgent.Description != "" {
		fmt.Printf("  Description: %s\n", selectedAgent.Description)
	}

	return nil
}

// selectResource handles resource selection when multiple resources are found
func selectResource(resources []*resource.Resource, resourceType string) (*resource.Resource, error) {
	if len(resources) == 1 {
		// Single resource found, use it automatically
		return resources[0], nil
	}

	// Multiple resources found, prompt user to select
	fmt.Printf("Multiple %ss found:\n", resourceType)
	for i, res := range resources {
		desc := res.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %d. %s - %s\n", i+1, res.Name, desc)
	}

	fmt.Print("\nSelect a resource (1-", len(resources), "): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read selection: %w", err)
	}

	input = strings.TrimSpace(input)
	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(resources) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return resources[selection-1], nil
}

// findCommandFile finds the actual file path for a command by name
func findCommandFile(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// Load and check if it's the command we're looking for
		cmd, err := resource.LoadCommand(path)
		if err != nil {
			return nil // Skip invalid commands
		}
		if cmd.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("command file not found for: %s", name)
	}
	return found, nil
}

// findSkillDir finds the directory path for a skill by name
func findSkillDir(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		// Check if this directory contains SKILL.md
		skillMdPath := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillMdPath); err != nil {
			return nil // No SKILL.md, continue
		}
		// Load and check if it's the skill we're looking for
		skill, err := resource.LoadSkill(path)
		if err != nil {
			return nil // Skip invalid skills
		}
		if skill.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("skill directory not found for: %s", name)
	}
	return found, nil
}

// findAgentFile finds the actual file path for an agent by name
func findAgentFile(searchPath, name string) (string, error) {
	var found string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// Load and check if it's the agent we're looking for
		agent, err := resource.LoadAgent(path)
		if err != nil {
			return nil // Skip invalid agents
		}
		if agent.Name == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("agent file not found for: %s", name)
	}
	return found, nil
}
