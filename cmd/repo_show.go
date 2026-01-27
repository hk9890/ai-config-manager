package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// repoShowCmd represents the repo show command
var repoShowCmd = &cobra.Command{
	Use:   "show <pattern>",
	Short: "Display detailed resource information",
	Long: `Display detailed information about resources in the repository.

Examples:
  aimgr repo show skill/pdf-processing    # Show specific skill
  aimgr repo show command/test            # Show specific command
  aimgr repo show agent/code-reviewer     # Show specific agent
  aimgr repo show skill/*                 # Show all skills (summary)
  aimgr repo show *pdf*                   # Show all resources with "pdf"

Supports glob patterns: *, ?, [abc], {a,b}

When multiple resources match, shows a summary list.
When a single resource matches, shows detailed information.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeResourceArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]

		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Expand pattern to matching resources
		matches, err := ExpandPattern(manager, pattern)
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			return fmt.Errorf("no resources found matching '%s'", pattern)
		}

		if len(matches) == 1 {
			// Single match - show detailed view
			return showDetailedResource(manager, matches[0])
		}

		// Multiple matches - show summary
		return showResourceSummary(manager, matches)
	},
}

// showDetailedResource displays detailed information for a single resource
func showDetailedResource(manager *repo.Manager, resourceArg string) error {
	// Parse the resource argument
	resourceType, name, err := ParseResourceArg(resourceArg)
	if err != nil {
		return err
	}

	// Load the resource
	res, err := manager.Get(name, resourceType)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", resourceType, err)
	}

	// Load metadata
	meta, err := manager.GetMetadata(name, resourceType)
	metadataAvailable := err == nil

	// Display based on resource type
	switch resourceType {
	case resource.Skill:
		return showSkillDetails(manager, res, metadataAvailable, meta)
	case resource.Command:
		return showCommandDetails(manager, res, metadataAvailable, meta)
	case resource.Agent:
		return showAgentDetails(manager, res, metadataAvailable, meta)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// showSkillDetails displays detailed information for a skill
func showSkillDetails(manager *repo.Manager, res *resource.Resource, metadataAvailable bool, meta *metadata.ResourceMetadata) error {
	// Load full skill resource for additional details
	skillPath := manager.GetPath(res.Name, resource.Skill)
	skill, err := resource.LoadSkillResource(skillPath)
	if err != nil {
		return fmt.Errorf("failed to load skill details: %w", err)
	}

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
}

// showCommandDetails displays detailed information for a command
func showCommandDetails(manager *repo.Manager, res *resource.Resource, metadataAvailable bool, meta *metadata.ResourceMetadata) error {
	// Load full command resource for additional details
	commandPath := manager.GetPath(res.Name, resource.Command)
	command, err := resource.LoadCommandResource(commandPath)
	if err != nil {
		return fmt.Errorf("failed to load command details: %w", err)
	}

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
}

// showAgentDetails displays detailed information for an agent
func showAgentDetails(manager *repo.Manager, res *resource.Resource, metadataAvailable bool, meta *metadata.ResourceMetadata) error {
	// Load full agent resource for additional details
	agentPath := manager.GetPath(res.Name, resource.Agent)
	agent, err := resource.LoadAgentResource(agentPath)
	if err != nil {
		return fmt.Errorf("failed to load agent details: %w", err)
	}

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
}

// showResourceSummary displays a summary table for multiple resources
func showResourceSummary(manager *repo.Manager, matches []string) error {
	fmt.Printf("Found %d matching resources:\n\n", len(matches))

	// Display table header
	fmt.Printf("%-10s %-30s %s\n", "TYPE", "NAME", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 80))

	// Display each resource
	for _, match := range matches {
		resourceType, name, err := ParseResourceArg(match)
		if err != nil {
			// Skip invalid matches
			continue
		}

		res, err := manager.Get(name, resourceType)
		if err != nil {
			// Skip if can't load
			continue
		}

		// Truncate description if too long
		desc := res.Description
		if len(desc) > 45 {
			desc = desc[:42] + "..."
		}

		fmt.Printf("%-10s %-30s %s\n", resourceType, name, desc)
	}

	fmt.Println()
	fmt.Println("Use 'aimgr repo show <type>/<name>' to see detailed information")

	return nil
}

// formatTimestamp formats a timestamp in a human-readable format
func formatTimestamp(t time.Time) string {
	// Format: "Jan 2, 2006 at 3:04pm (MST)"
	return t.Format("Jan 2, 2006 at 3:04pm (MST)")
}

func init() {
	repoCmd.AddCommand(repoShowCmd)
}
