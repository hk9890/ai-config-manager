package cmd

import (
	"fmt"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/spf13/cobra"
)

// parseResourceArg parses a resource argument in the format "type/name"
// Returns the resource type, name, and any error
func parseResourceArg(arg string) (resource.ResourceType, string, error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format: must be 'type/name' (e.g., skill/my-skill, command/my-command, agent/my-agent)")
	}

	typeStr := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if name == "" {
		return "", "", fmt.Errorf("resource name cannot be empty")
	}

	// Normalize type (support plural forms)
	var resourceType resource.ResourceType
	switch strings.ToLower(typeStr) {
	case "skill", "skills":
		resourceType = resource.Skill
	case "command", "commands":
		resourceType = resource.Command
	case "agent", "agents":
		resourceType = resource.Agent
	default:
		return "", "", fmt.Errorf("invalid resource type '%s': must be one of 'skill', 'command', or 'agent'", typeStr)
	}

	return resourceType, name, nil
}

// completeResourceArgs provides shell completion for resource arguments in "type/name" format
// Supports completing both the type prefix and resource names after the slash
func completeResourceArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	manager, err := repo.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// If toComplete doesn't contain a slash, suggest type prefixes
	if !strings.Contains(toComplete, "/") {
		prefixes := []string{"skill/", "command/", "agent/"}
		var matches []string
		for _, prefix := range prefixes {
			if strings.HasPrefix(prefix, toComplete) {
				matches = append(matches, prefix)
			}
		}
		return matches, cobra.ShellCompDirectiveNoSpace
	}

	// Parse the type prefix
	parts := strings.SplitN(toComplete, "/", 2)
	if len(parts) != 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	typeStr := parts[0]
	namePrefix := parts[1]

	// Determine resource type
	var resourceType resource.ResourceType
	switch strings.ToLower(typeStr) {
	case "skill", "skills":
		resourceType = resource.Skill
	case "command", "commands":
		resourceType = resource.Command
	case "agent", "agents":
		resourceType = resource.Agent
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get resources of that type
	resources, err := manager.List(&resourceType)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Build completions with type prefix
	var suggestions []string
	for _, res := range resources {
		if strings.HasPrefix(res.Name, namePrefix) {
			suggestions = append(suggestions, typeStr+"/"+res.Name)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

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

// completeInstallResources provides shell completion for install command with type prefix support
// Handles both old-style subcommands and new type/name format
func completeInstallResources(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	manager, err := repo.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Check how many args we already have
	argIndex := len(args)

	// If this is the first argument (or continuing after previous args)
	// If toComplete doesn't contain a slash, suggest type prefixes
	if !strings.Contains(toComplete, "/") {
		prefixes := []string{"skill/", "command/", "agent/"}
		var matches []string
		for _, prefix := range prefixes {
			if strings.HasPrefix(prefix, toComplete) {
				matches = append(matches, prefix)
			}
		}
		return matches, cobra.ShellCompDirectiveNoSpace
	}

	// Parse the type prefix
	parts := strings.SplitN(toComplete, "/", 2)
	if len(parts) != 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	typeStr := parts[0]
	namePrefix := parts[1]

	// Determine resource type
	var resourceType resource.ResourceType
	switch strings.ToLower(typeStr) {
	case "skill", "skills":
		resourceType = resource.Skill
	case "command", "commands":
		resourceType = resource.Command
	case "agent", "agents":
		resourceType = resource.Agent
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get resources of that type
	resources, err := manager.List(&resourceType)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Build completions with type prefix
	var suggestions []string
	for _, res := range resources {
		if strings.HasPrefix(res.Name, namePrefix) {
			suggestions = append(suggestions, typeStr+"/"+res.Name)
		}
	}

	// Allow multiple resources: after completing one, suggest prefixes again
	if argIndex > 0 {
		// Already have some args, so also suggest type prefixes for next resource
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

// completeUninstallResources provides shell completion for uninstall command with type prefix support
// Same as completeInstallResources but for uninstall (could be merged if logic is identical)
func completeUninstallResources(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For now, reuse the same logic as install
	return completeInstallResources(cmd, args, toComplete)
}

// completeResourceTypes provides completion for resource type arguments
func completeResourceTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	types := []string{"command", "skill", "agent"}
	return types, cobra.ShellCompDirectiveNoFileComp
}

// completeConfigKeys provides completion for config key names
func completeConfigKeys(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	keys := []string{"install.targets", "default-tool"}
	return keys, cobra.ShellCompDirectiveNoFileComp
}

// completeConfigSetArgs provides completion for config set command (key and value)
func completeConfigSetArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		// Complete config keys
		return []string{"install.targets"}, cobra.ShellCompDirectiveNoFileComp
	} else if len(args) == 1 && args[0] == "install.targets" {
		// Complete tool names for install.targets
		return []string{"claude", "opencode", "copilot"}, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
