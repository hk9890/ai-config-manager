package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/install"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listInstalledFormatFlag string
	listInstalledPathFlag   string
)

// listInstalledCmd represents the list command for installed resources
var listInstalledCmd = &cobra.Command{
	Use:   "list [command|skill|agent]",
	Short: "List installed resources in the current directory",
	Long: `List all resources installed in the current directory (or specified path).

This command shows resources that were installed using 'aimgr install',
displaying which tools (claude, opencode, copilot) each resource is installed to.

Only resources installed via aimgr (symlinks) are shown - manually copied files are excluded.

Examples:
  aimgr list                         # List all installed resources
  aimgr list command                 # List only installed commands
  aimgr list skill                   # List only installed skills
  aimgr list agent                   # List only installed agents
  aimgr list --format=json           # Output as JSON
  aimgr list --format=yaml           # Output as YAML
  aimgr list --path ~/project        # List in specific directory`,
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

		// Get project path (current directory or flag)
		projectPath := listInstalledPathFlag
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		// Detect which tools exist in the project
		detectedTools, err := tools.DetectExistingTools(projectPath)
		if err != nil {
			return fmt.Errorf("failed to detect tools: %w", err)
		}

		if len(detectedTools) == 0 {
			fmt.Println("No tool directories found in this project.")
			fmt.Println("\nExpected directories: .claude, .opencode, or .github/skills")
			fmt.Println("Install resources with: aimgr install <resource>")
			return nil
		}

		// Create installer to list resources
		installer, err := install.NewInstallerWithTargets(projectPath, detectedTools)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		// List all installed resources
		resources, err := installer.List()
		if err != nil {
			return fmt.Errorf("failed to list installed resources: %w", err)
		}

		// Filter by type if specified
		if resourceType != nil {
			filtered := []resource.Resource{}
			for _, res := range resources {
				if res.Type == *resourceType {
					filtered = append(filtered, res)
				}
			}
			resources = filtered
		}

		// Handle empty results
		if len(resources) == 0 {
			if resourceType != nil {
				fmt.Printf("No %s resources installed in this project.\n", *resourceType)
			} else {
				fmt.Println("No resources installed in this project.")
			}
			fmt.Println("\nInstall resources with: aimgr install <resource>")
			return nil
		}

		// Get tool installation info for each resource
		resourceInfos := buildResourceInfo(resources, projectPath, detectedTools)

		// Format output based on --format flag
		switch listInstalledFormatFlag {
		case "json":
			return outputInstalledJSON(resourceInfos)
		case "yaml":
			return outputInstalledYAML(resourceInfos)
		case "table":
			return outputInstalledTable(resourceInfos)
		default:
			return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'yaml')", listInstalledFormatFlag)
		}
	},
}

// ResourceInfo extends Resource with installation target information
type ResourceInfo struct {
	Type        resource.ResourceType `json:"type" yaml:"type"`
	Name        string                `json:"name" yaml:"name"`
	Description string                `json:"description" yaml:"description"`
	Version     string                `json:"version,omitempty" yaml:"version,omitempty"`
	Targets     []string              `json:"targets" yaml:"targets"`
}

// buildResourceInfo creates ResourceInfo entries with target tool information
func buildResourceInfo(resources []resource.Resource, projectPath string, detectedTools []tools.Tool) []ResourceInfo {
	infos := make([]ResourceInfo, 0, len(resources))

	for _, res := range resources {
		info := ResourceInfo{
			Type:        res.Type,
			Name:        res.Name,
			Description: res.Description,
			Version:     res.Version,
			Targets:     []string{},
		}

		// Check which tools have this resource installed
		for _, tool := range detectedTools {
			if isInstalledInTool(projectPath, res.Name, res.Type, tool) {
				info.Targets = append(info.Targets, tool.String())
			}
		}

		infos = append(infos, info)
	}

	return infos
}

// isInstalledInTool checks if a resource is installed in a specific tool directory
func isInstalledInTool(projectPath, name string, resType resource.ResourceType, tool tools.Tool) bool {
	toolInfo := tools.GetToolInfo(tool)
	var checkPath string

	switch resType {
	case resource.Command:
		if !toolInfo.SupportsCommands {
			return false
		}
		checkPath = fmt.Sprintf("%s/%s/%s.md", projectPath, toolInfo.CommandsDir, name)
	case resource.Skill:
		if !toolInfo.SupportsSkills {
			return false
		}
		checkPath = fmt.Sprintf("%s/%s/%s", projectPath, toolInfo.SkillsDir, name)
	case resource.Agent:
		if !toolInfo.SupportsAgents {
			return false
		}
		checkPath = fmt.Sprintf("%s/%s/%s.md", projectPath, toolInfo.AgentsDir, name)
	default:
		return false
	}

	// Check if symlink exists
	info, err := os.Lstat(checkPath)
	if err != nil {
		return false
	}

	// Verify it's a symlink (only count aimgr-managed installations)
	return info.Mode()&os.ModeSymlink != 0
}

func outputInstalledTable(infos []ResourceInfo) error {
	// Group by type
	commands := []ResourceInfo{}
	skills := []ResourceInfo{}
	agents := []ResourceInfo{}

	for _, info := range infos {
		if info.Type == resource.Command {
			commands = append(commands, info)
		} else if info.Type == resource.Skill {
			skills = append(skills, info)
		} else if info.Type == resource.Agent {
			agents = append(agents, info)
		}
	}

	// Create table with new API
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Type", "Name", "Targets", "Description")

	// Add commands
	for _, cmd := range commands {
		desc := truncateString(cmd.Description, 50)
		targets := strings.Join(cmd.Targets, ", ")
		if err := table.Append("command", cmd.Name, targets, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if commands exist and skills or agents exist
	if len(commands) > 0 && (len(skills) > 0 || len(agents) > 0) {
		if err := table.Append("", "", "", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add skills
	for _, skill := range skills {
		desc := truncateString(skill.Description, 50)
		targets := strings.Join(skill.Targets, ", ")
		if err := table.Append("skill", skill.Name, targets, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between types if skills exist and agents exist
	if len(skills) > 0 && len(agents) > 0 {
		if err := table.Append("", "", "", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add agents
	for _, agent := range agents {
		desc := truncateString(agent.Description, 50)
		targets := strings.Join(agent.Targets, ", ")
		if err := table.Append("agent", agent.Name, targets, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return table.Render()
}

func outputInstalledJSON(infos []ResourceInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

func outputInstalledYAML(infos []ResourceInfo) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(infos)
}

func init() {
	rootCmd.AddCommand(listInstalledCmd)
	listInstalledCmd.Flags().StringVar(&listInstalledFormatFlag, "format", "table", "Output format (table|json|yaml)")
	listInstalledCmd.Flags().StringVar(&listInstalledPathFlag, "path", "", "Project directory path (default: current directory)")
}
