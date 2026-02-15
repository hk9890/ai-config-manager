package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/install"
	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	listInstalledFormatFlag string
	listInstalledPathFlag   string
)

// SyncStatus represents the synchronization state between installed resources and ai.package.yaml manifest
type SyncStatus string

const (
	// SyncStatusInSync means the resource is both in the manifest and installed
	SyncStatusInSync SyncStatus = "in-sync"
	// SyncStatusNotInManifest means the resource is installed but not in the manifest
	SyncStatusNotInManifest SyncStatus = "not-in-manifest"
	// SyncStatusNotInstalled means the resource is in the manifest but not installed
	SyncStatusNotInstalled SyncStatus = "not-installed"
	// SyncStatusNoManifest means no ai.package.yaml file exists in the project
	SyncStatusNoManifest SyncStatus = "no-manifest"
)

// listInstalledCmd represents the list command for installed resources
var listInstalledCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List installed resources in the current directory",
	Long: `List all resources installed in the current directory (or specified path).

This command shows resources that were installed using 'aimgr install',
displaying which tools (claude, opencode, copilot) each resource is installed to
and their synchronization status with ai.package.yaml.

Output columns:
  - NAME: Resource reference (e.g., skill/pdf-processing, command/test)
  - TARGETS: Tools where the resource is installed (e.g., claude, opencode)
  - SYNC: Synchronization status with ai.package.yaml manifest
  - DESCRIPTION: Brief description of the resource

Sync Status Symbols:
  ✓ = In sync (resource is both in manifest and installed)
  * = Not in manifest (installed but not declared in ai.package.yaml)
  ⚠ = Not installed (declared in manifest but not installed yet)
  - = No manifest (no ai.package.yaml file exists in current directory)

The sync status helps you keep your installations aligned with your project's
ai.package.yaml manifest file. Resources marked with * should be added to the
manifest if you want to track them. Resources marked with ⚠ need to be installed
using 'aimgr install <resource>'.

Only resources installed via aimgr (symlinks) are shown - manually copied files are excluded.

The optional pattern argument supports glob wildcards (* ? [ ]) and type prefixes:
  - No pattern: list all installed resources
  - Type patterns: skill/*, command/*, agent/*
  - Name patterns: *test*, *debug*, pdf*
  - Combined: skill/pdf*, command/test-*

Examples:
  aimgr list                         # List all installed resources with sync status
  aimgr list skill/*                 # List all installed skills
  aimgr list command/*               # List all installed commands
  aimgr list agent/*                 # List all installed agents
  aimgr list *test*                  # List all resources with "test" in name
  aimgr list skill/pdf*              # List installed skills starting with "pdf"
  aimgr list command/test-*          # List installed commands starting with "test-"
  aimgr list --format=json           # Output as JSON with sync_status field
  aimgr list --format=yaml           # Output as YAML
  aimgr list --path ~/project        # List in specific directory`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeInstalledResources,
	RunE: func(cmd *cobra.Command, args []string) error {

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

		// Filter by pattern if specified
		if len(args) > 0 {
			patternArg := args[0]
			matcher, err := pattern.NewMatcher(patternArg)
			if err != nil {
				return fmt.Errorf("invalid pattern '%s': %w", patternArg, err)
			}

			filtered := []resource.Resource{}
			for _, res := range resources {
				if matcher.Match(&res) {
					filtered = append(filtered, res)
				}
			}
			resources = filtered

			// Handle no matches
			if len(resources) == 0 {
				fmt.Printf("No installed resources match pattern '%s'.\n", patternArg)
				fmt.Println("\nInstall resources with: aimgr install <resource>")
				return nil
			}
		}

		// Handle empty results (no pattern specified)
		if len(resources) == 0 {
			fmt.Println("No resources installed in this project.")
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
			return outputInstalledTable(resourceInfos, projectPath)
		default:
			return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'yaml')", listInstalledFormatFlag)
		}
	},
}

// ResourceInfo extends Resource with installation target information and sync status
type ResourceInfo struct {
	Type        resource.ResourceType `json:"type" yaml:"type"`
	Name        string                `json:"name" yaml:"name"`
	Description string                `json:"description" yaml:"description"`
	Version     string                `json:"version,omitempty" yaml:"version,omitempty"`
	Targets     []string              `json:"targets" yaml:"targets"`
	SyncStatus  string                `json:"sync_status" yaml:"sync_status"`
}

// getSyncStatus determines the sync status of a resource relative to ai.package.yaml
// It compares the resource's installation state with its presence in the manifest file
func getSyncStatus(projectPath string, resourceRef string, isInstalled bool) SyncStatus {
	// Load the manifest file from the project directory
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	m, err := manifest.Load(manifestPath)
	if err != nil {
		// If manifest file doesn't exist, return no-manifest status
		if os.IsNotExist(err) {
			return SyncStatusNoManifest
		}
		// For other errors (e.g., invalid YAML), also treat as no manifest
		// This is graceful - we can't determine sync status without a valid manifest
		return SyncStatusNoManifest
	}

	// Check if the resource reference exists in the manifest
	inManifest := m.Has(resourceRef)

	// Determine sync status based on both installation and manifest presence
	if inManifest && isInstalled {
		return SyncStatusInSync
	} else if inManifest && !isInstalled {
		return SyncStatusNotInstalled
	} else if !inManifest && isInstalled {
		return SyncStatusNotInManifest
	}

	// This case shouldn't normally occur (not in manifest and not installed)
	// But we handle it by treating it as not-in-manifest
	return SyncStatusNotInManifest
}

// formatSyncStatus gets the sync status and returns the appropriate symbol
func formatSyncStatus(projectPath string, resourceRef string, isInstalled bool) string {
	status := getSyncStatus(projectPath, resourceRef, isInstalled)

	switch status {
	case SyncStatusInSync:
		return "✓"
	case SyncStatusNotInManifest:
		return "*"
	case SyncStatusNotInstalled:
		return "⚠"
	case SyncStatusNoManifest:
		return "-"
	default:
		return "-"
	}
}

// buildResourceInfo creates ResourceInfo entries with target tool information and sync status
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

		// Determine if resource is installed (has at least one target)
		isInstalled := len(info.Targets) > 0

		// Build resource reference for manifest lookup
		resourceRef := fmt.Sprintf("%s/%s", res.Type, res.Name)

		// Get sync status
		syncStatus := getSyncStatus(projectPath, resourceRef, isInstalled)
		info.SyncStatus = string(syncStatus)

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

func outputInstalledTable(infos []ResourceInfo, projectPath string) error {
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

	// Create table with NAME, TARGETS, SYNC, DESCRIPTION using shared infrastructure
	table := output.NewTable("Name", "Targets", "Sync", "Description")
	table.WithResponsive().
		WithDynamicColumn(3).              // Description stretches
		WithMinColumnWidths(40, 12, 4, 30) // Name min=40, Targets min=12, Sync min=4, Description min=30

	// Add commands
	for _, cmd := range commands {
		targets := strings.Join(cmd.Targets, ", ")
		resourceRef := fmt.Sprintf("command/%s", cmd.Name)
		syncSymbol := formatSyncStatus(projectPath, resourceRef, len(cmd.Targets) > 0)
		table.AddRow(resourceRef, targets, syncSymbol, cmd.Description)
	}

	// Add empty row between types if commands exist and skills or agents exist
	if len(commands) > 0 && (len(skills) > 0 || len(agents) > 0) {
		table.AddSeparator()
	}

	// Add skills
	for _, skill := range skills {
		targets := strings.Join(skill.Targets, ", ")
		resourceRef := fmt.Sprintf("skill/%s", skill.Name)
		syncSymbol := formatSyncStatus(projectPath, resourceRef, len(skill.Targets) > 0)
		table.AddRow(resourceRef, targets, syncSymbol, skill.Description)
	}

	// Add empty row between types if skills exist and agents exist
	if len(skills) > 0 && len(agents) > 0 {
		table.AddSeparator()
	}

	// Add agents
	for _, agent := range agents {
		targets := strings.Join(agent.Targets, ", ")
		resourceRef := fmt.Sprintf("agent/%s", agent.Name)
		syncSymbol := formatSyncStatus(projectPath, resourceRef, len(agent.Targets) > 0)
		table.AddRow(resourceRef, targets, syncSymbol, agent.Description)
	}

	// Render the table
	if err := table.Format(output.Table); err != nil {
		return err
	}

	// Print legend
	fmt.Println("\nLegend:")
	fmt.Println("  ✓ = In sync  * = Not in manifest  ⚠ = Not installed  - = No manifest")

	return nil
}

func outputInstalledJSON(infos []ResourceInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

func outputInstalledYAML(infos []ResourceInfo) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(infos)
}

func init() {
	rootCmd.AddCommand(listInstalledCmd)
	listInstalledCmd.Flags().StringVar(&listInstalledFormatFlag, "format", "table", "Output format (table|json|yaml)")
	listInstalledCmd.Flags().StringVar(&listInstalledPathFlag, "path", "", "Project directory path (default: current directory)")
	_ = listInstalledCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
}
