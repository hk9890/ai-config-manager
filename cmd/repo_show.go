package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// repoShowCmd represents the repo show command
var repoShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display detailed resource information",
	Long: `Display detailed information about a resource in the repository.

Use the subcommands to show details for specific resource types:
  aimgr repo show skill <name>       # Show skill details
  aimgr repo show command <name>     # Show command details
  aimgr repo show agent <name>       # Show agent details`,
}

// repoShowSkillCmd shows details for a skill
var repoShowSkillCmd = &cobra.Command{
	Use:               "skill <name>",
	Short:             "Show detailed skill information",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Load the resource
		res, err := manager.Get(name, resource.Skill)
		if err != nil {
			return fmt.Errorf("failed to load skill: %w", err)
		}

		// Load full skill resource for additional details
		skillPath := manager.GetPath(name, resource.Skill)
		skill, err := resource.LoadSkillResource(skillPath)
		if err != nil {
			return fmt.Errorf("failed to load skill details: %w", err)
		}

		// Load metadata
		meta, err := manager.GetMetadata(name, resource.Skill)
		metadataAvailable := err == nil

		// Display information
		fmt.Printf("Skill: %s\n", res.Name)
		if res.Version != "" {
			fmt.Printf("Version: %s\n", res.Version)
		}
		fmt.Printf("Description: %s\n", res.Description)
		if res.License != "" {
			fmt.Printf("License: %s\n", res.License)
		}
		if res.Author != "" {
			fmt.Printf("Author: %s\n", res.Author)
		}

		// Display compatibility if available
		if len(skill.Compatibility) > 0 {
			fmt.Printf("Compatibility: %s\n", strings.Join(skill.Compatibility, ", "))
		}

		// Display skill structure
		features := []string{}
		if skill.HasScripts {
			features = append(features, "scripts")
		}
		if skill.HasReferences {
			features = append(features, "references")
		}
		if skill.HasAssets {
			features = append(features, "assets")
		}
		if len(features) > 0 {
			fmt.Printf("Features: %s\n", strings.Join(features, ", "))
		}

		fmt.Println()

		// Display metadata if available
		if metadataAvailable {
			fmt.Printf("Source: %s\n", meta.SourceURL)
			fmt.Printf("Source Type: %s\n", meta.SourceType)
			fmt.Printf("First Installed: %s\n", formatTimestamp(meta.FirstInstalled))
			fmt.Printf("Last Updated: %s\n", formatTimestamp(meta.LastUpdated))
			fmt.Println()
		} else {
			fmt.Println("Metadata: Not available")
			fmt.Println()
		}

		fmt.Printf("Location: %s\n", skillPath)

		return nil
	},
}

// repoShowCommandCmd shows details for a command
var repoShowCommandCmd = &cobra.Command{
	Use:               "command <name>",
	Short:             "Show detailed command information",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeCommandNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Load the resource
		res, err := manager.Get(name, resource.Command)
		if err != nil {
			return fmt.Errorf("failed to load command: %w", err)
		}

		// Load full command resource for additional details
		commandPath := manager.GetPath(name, resource.Command)
		command, err := resource.LoadCommandResource(commandPath)
		if err != nil {
			return fmt.Errorf("failed to load command details: %w", err)
		}

		// Load metadata
		meta, err := manager.GetMetadata(name, resource.Command)
		metadataAvailable := err == nil

		// Display information
		fmt.Printf("Command: %s\n", res.Name)
		if res.Version != "" {
			fmt.Printf("Version: %s\n", res.Version)
		}
		fmt.Printf("Description: %s\n", res.Description)
		if res.License != "" {
			fmt.Printf("License: %s\n", res.License)
		}
		if res.Author != "" {
			fmt.Printf("Author: %s\n", res.Author)
		}

		// Display command-specific fields
		if command.Agent != "" {
			fmt.Printf("Agent: %s\n", command.Agent)
		}
		if command.Model != "" {
			fmt.Printf("Model: %s\n", command.Model)
		}
		if len(command.AllowedTools) > 0 {
			fmt.Printf("Allowed Tools: %s\n", strings.Join(command.AllowedTools, ", "))
		}

		fmt.Println()

		// Display metadata if available
		if metadataAvailable {
			fmt.Printf("Source: %s\n", meta.SourceURL)
			fmt.Printf("Source Type: %s\n", meta.SourceType)
			fmt.Printf("First Installed: %s\n", formatTimestamp(meta.FirstInstalled))
			fmt.Printf("Last Updated: %s\n", formatTimestamp(meta.LastUpdated))
			fmt.Println()
		} else {
			fmt.Println("Metadata: Not available")
			fmt.Println()
		}

		fmt.Printf("Location: %s\n", commandPath)

		return nil
	},
}

// repoShowAgentCmd shows details for an agent
var repoShowAgentCmd = &cobra.Command{
	Use:               "agent <name>",
	Short:             "Show detailed agent information",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeAgentNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Load the resource
		res, err := manager.Get(name, resource.Agent)
		if err != nil {
			return fmt.Errorf("failed to load agent: %w", err)
		}

		// Load full agent resource for additional details
		agentPath := manager.GetPath(name, resource.Agent)
		agent, err := resource.LoadAgentResource(agentPath)
		if err != nil {
			return fmt.Errorf("failed to load agent details: %w", err)
		}

		// Load metadata
		meta, err := manager.GetMetadata(name, resource.Agent)
		metadataAvailable := err == nil

		// Display information
		fmt.Printf("Agent: %s\n", res.Name)
		if res.Version != "" {
			fmt.Printf("Version: %s\n", res.Version)
		}
		fmt.Printf("Description: %s\n", res.Description)
		if res.License != "" {
			fmt.Printf("License: %s\n", res.License)
		}
		if res.Author != "" {
			fmt.Printf("Author: %s\n", res.Author)
		}

		// Display agent-specific fields (OpenCode format)
		if agent.Type != "" {
			fmt.Printf("Type: %s\n", agent.Type)
		}
		if agent.Instructions != "" {
			fmt.Printf("Instructions: %s\n", agent.Instructions)
		}
		if len(agent.Capabilities) > 0 {
			fmt.Printf("Capabilities: %s\n", strings.Join(agent.Capabilities, ", "))
		}

		fmt.Println()

		// Display metadata if available
		if metadataAvailable {
			fmt.Printf("Source: %s\n", meta.SourceURL)
			fmt.Printf("Source Type: %s\n", meta.SourceType)
			fmt.Printf("First Installed: %s\n", formatTimestamp(meta.FirstInstalled))
			fmt.Printf("Last Updated: %s\n", formatTimestamp(meta.LastUpdated))
			fmt.Println()
		} else {
			fmt.Println("Metadata: Not available")
			fmt.Println()
		}

		fmt.Printf("Location: %s\n", agentPath)

		return nil
	},
}

// formatTimestamp formats a timestamp in a human-readable format
func formatTimestamp(t time.Time) string {
	// Format: "Jan 2, 2006 at 3:04pm (MST)"
	return t.Format("Jan 2, 2006 at 3:04pm (MST)")
}

func init() {
	repoCmd.AddCommand(repoShowCmd)
	repoShowCmd.AddCommand(repoShowSkillCmd)
	repoShowCmd.AddCommand(repoShowCommandCmd)
	repoShowCmd.AddCommand(repoShowAgentCmd)
}
