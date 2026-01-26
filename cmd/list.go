package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var formatFlag string
var typeFlag string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List resources in the repository",
	Long: `List all resources in the aimgr repository, optionally filtered by pattern.

Patterns support wildcards (* for multiple characters, ? for single character) and optional type prefixes.

Examples:
  aimgr repo list                    # List all resources and packages
  aimgr repo list skill/*            # List all skills
  aimgr repo list command/test*      # List commands starting with "test"
  aimgr repo list *pdf*              # List all resources with "pdf" in name
  aimgr repo list agent/code-*       # List agents starting with "code-"
  aimgr repo list --type=package     # List only packages
  aimgr repo list --format=json      # Output as JSON
  aimgr repo list --format=yaml      # Output as YAML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create manager
		manager, err := repo.NewManager()
		if err != nil {
			return err
		}

		// Check if user wants only packages
		if typeFlag == "package" {
			packages, err := manager.ListPackages()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}

			if len(packages) == 0 {
				fmt.Println("No packages found in repository.")
				return nil
			}

			// Format output based on --format flag
			switch formatFlag {
			case "json":
				return outputPackagesJSON(packages)
			case "yaml":
				return outputPackagesYAML(packages)
			case "table":
				return outputPackagesTable(packages)
			default:
				return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'yaml')", formatFlag)
			}
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

		// Get packages (if not using pattern and not filtered by type)
		var packages []repo.PackageInfo
		if len(args) == 0 && typeFlag == "" {
			packages, err = manager.ListPackages()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}
		}

		// Handle empty results
		if len(resources) == 0 && len(packages) == 0 {
			if len(args) == 0 {
				fmt.Println("No resources or packages found in repository.")
			} else {
				fmt.Printf("No resources matching pattern '%s' found in repository.\n", args[0])
			}
			fmt.Println("\nAdd resources with: aimgr repo add command <file>, aimgr repo add skill <folder>, or aimgr repo add agent <file>")
			return nil
		}

		// Format output based on --format flag
		switch formatFlag {
		case "json":
			return outputWithPackagesJSON(resources, packages)
		case "yaml":
			return outputWithPackagesYAML(resources, packages)
		case "table":
			return outputWithPackagesTable(resources, packages)
		default:
			return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'yaml')", formatFlag)
		}
	},
}

func outputWithPackagesTable(resources []resource.Resource, packages []repo.PackageInfo) error {
	// Group resources by type
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

	// Sort packages alphabetically by name
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	// Create table with new API
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Name", "Description")

	// Add commands
	for _, cmd := range commands {
		desc := truncateString(cmd.Description, 60)
		if err := table.Append(fmt.Sprintf("command/%s", cmd.Name), desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if commands exist and skills or agents exist
	if len(commands) > 0 && (len(skills) > 0 || len(agents) > 0) {
		if err := table.Append("", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add skills
	for _, skill := range skills {
		desc := truncateString(skill.Description, 60)
		if err := table.Append(fmt.Sprintf("skill/%s", skill.Name), desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if skills exist and agents exist
	if len(skills) > 0 && len(agents) > 0 {
		if err := table.Append("", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add agents
	for _, agent := range agents {
		desc := truncateString(agent.Description, 60)
		if err := table.Append(fmt.Sprintf("agent/%s", agent.Name), desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between agents and packages if both exist
	if len(agents) > 0 && len(packages) > 0 {
		if err := table.Append("", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add packages
	for _, pkg := range packages {
		desc := truncateString(pkg.Description, 50)
		countStr := fmt.Sprintf("%d resources", pkg.ResourceCount)
		fullDesc := fmt.Sprintf("%s    %s", countStr, desc)
		if err := table.Append(fmt.Sprintf("package/%s", pkg.Name), fullDesc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return table.Render()
}

func outputPackagesTable(packages []repo.PackageInfo) error {
	// Sort packages alphabetically by name
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Name", "Resources", "Description")

	for _, pkg := range packages {
		desc := truncateString(pkg.Description, 60)
		if err := table.Append(pkg.Name, fmt.Sprintf("%d", pkg.ResourceCount), desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return table.Render()
}

func outputWithPackagesJSON(resources []resource.Resource, packages []repo.PackageInfo) error {
	output := map[string]interface{}{
		"resources": resources,
		"packages":  packages,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputPackagesJSON(packages []repo.PackageInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(packages)
}

func outputWithPackagesYAML(resources []resource.Resource, packages []repo.PackageInfo) error {
	output := map[string]interface{}{
		"resources": resources,
		"packages":  packages,
	}
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(output)
}

func outputPackagesYAML(packages []repo.PackageInfo) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(packages)
}

func outputTable(resources []resource.Resource) error {
	return outputWithPackagesTable(resources, nil)
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
	listCmd.Flags().StringVar(&typeFlag, "type", "", "Filter by type (command|skill|agent|package)")
}
