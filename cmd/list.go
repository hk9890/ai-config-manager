package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var formatFlag string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List resources in the repository",
	Long: `List all resources in the aimgr repository, optionally filtered by pattern.

Patterns support wildcards (* for multiple characters, ? for single character) and optional type prefixes.

Examples:
  aimgr repo list                    # List all resources
  aimgr repo list skill/*            # List all skills
  aimgr repo list command/test*      # List commands starting with "test"
  aimgr repo list *pdf*              # List all resources with "pdf" in name
  aimgr repo list agent/code-*       # List agents starting with "code-"
  aimgr repo list --format=json      # Output as JSON
  aimgr repo list --format=yaml      # Output as YAML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		var resources []resource.Resource

		if len(args) == 0 {
			// List all resources (no filter)
			resources, err = manager.List(nil)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}
		} else {
			// Parse pattern
			matcher, err := pattern.NewMatcher(args[0])
			if err != nil {
				return fmt.Errorf("invalid pattern '%s': %w", args[0], err)
			}

			// Get resource type filter if pattern specifies it
			resourceType, _, _ := pattern.ParsePattern(args[0])
			var typeFilter *resource.ResourceType
			if resourceType != "" {
				typeFilter = &resourceType
			}

			// List resources with optional type filter
			resources, err = manager.List(typeFilter)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			// Apply pattern matching
			var filtered []resource.Resource
			for _, res := range resources {
				if matcher.Match(&res) {
					filtered = append(filtered, res)
				}
			}
			resources = filtered
		}

		// Handle empty results
		if len(resources) == 0 {
			if len(args) == 0 {
				fmt.Println("No resources found in repository.")
			} else {
				fmt.Printf("No resources matching pattern '%s' found in repository.\n", args[0])
			}
			fmt.Println("\nAdd resources with: aimgr repo add command <file>, aimgr repo add skill <folder>, or aimgr repo add agent <file>")
			return nil
		}

		// Format output based on --format flag
		switch formatFlag {
		case "json":
			return outputJSON(resources)
		case "yaml":
			return outputYAML(resources)
		case "table":
			return outputTable(resources)
		default:
			return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'yaml')", formatFlag)
		}
	},
}

func outputTable(resources []resource.Resource) error {
	// Group by type
	commands := []resource.Resource{}
	skills := []resource.Resource{}
	agents := []resource.Resource{}

	for _, res := range resources {
		if res.Type == resource.Command {
			commands = append(commands, res)
		} else if res.Type == resource.Skill {
			skills = append(skills, res)
		} else if res.Type == resource.Agent {
			agents = append(agents, res)
		}
	}

	// Create table with new API
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Type", "Name", "Description")

	// Add commands
	for _, cmd := range commands {
		desc := truncateString(cmd.Description, 60)
		if err := table.Append("command", cmd.Name, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if commands exist and skills or agents exist
	if len(commands) > 0 && (len(skills) > 0 || len(agents) > 0) {
		if err := table.Append("", "", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add skills
	for _, skill := range skills {
		desc := truncateString(skill.Description, 60)
		if err := table.Append("skill", skill.Name, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if skills exist and agents exist
	if len(skills) > 0 && len(agents) > 0 {
		if err := table.Append("", "", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add agents
	for _, agent := range agents {
		desc := truncateString(agent.Description, 60)
		if err := table.Append("agent", agent.Name, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return table.Render()
}

func outputJSON(resources []resource.Resource) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resources)
}

func outputYAML(resources []resource.Resource) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(resources)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Truncate and add ellipsis
	return strings.TrimSpace(s[:maxLen-3]) + "..."
}

func init() {
	repoCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&formatFlag, "format", "table", "Output format (table|json|yaml)")
}
