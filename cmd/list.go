package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var formatFlag string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [command|skill|agent]",
	Short: "List resources in the repository",
	Long: `List all resources in the ai-repo repository, optionally filtered by type.

Examples:
  ai-repo list                    # List all resources
  ai-repo list command            # List only commands
  ai-repo list skill              # List only skills
  ai-repo list agent              # List only agents
  ai-repo list --format=json      # Output as JSON
  ai-repo list --format=yaml      # Output as YAML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse optional type filter
		var resourceType *resource.ResourceType
		if len(args) == 1 {
			typeStr := args[0]
			switch typeStr {
			case "command":
				t := resource.Command
				resourceType = &t
			case "skill":
				t := resource.Skill
				resourceType = &t
			case "agent":
				t := resource.Agent
				resourceType = &t
			default:
				return fmt.Errorf("invalid resource type: %s (must be 'command', 'skill', or 'agent')", typeStr)
			}
		}

		// Create manager and list resources
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		resources, err := manager.List(resourceType)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		// Handle empty repository
		if len(resources) == 0 {
			if resourceType != nil {
				fmt.Printf("No %s resources found in repository.\n", *resourceType)
			} else {
				fmt.Println("No resources found in repository.")
			}
			fmt.Println("\nAdd resources with: ai-repo add command <file>, ai-repo add skill <folder>, or ai-repo add agent <file>")
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
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&formatFlag, "format", "table", "Output format (table|json|yaml)")
}
