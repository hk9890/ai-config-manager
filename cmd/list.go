package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var formatFlag string
var typeFlag string

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

// ListResourceOutput represents a resource with installation and sync information
// for JSON and YAML output formats
type ListResourceOutput struct {
	resource.Resource
	Targets    []string `json:"targets" yaml:"targets"`         // Tools where resource is installed (e.g., ["claude", "opencode"])
	SyncStatus string   `json:"sync_status" yaml:"sync_status"` // Sync status (e.g., "in-sync", "not-in-manifest")
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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List resources in the repository",
	Long: `List all resources in the aimgr repository, optionally filtered by pattern.

The list command shows resources with their installation status and sync state:
  - NAME: Resource reference (e.g., skill/pdf-processing, command/test)
  - TARGETS: Tools where the resource is installed (e.g., claude, opencode, copilot)
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

Patterns support wildcards (* for multiple characters, ? for single character) 
and optional type prefixes.

Examples:
  aimgr repo list                    # List all resources with sync status
  aimgr repo list skill/*            # List all skills
  aimgr repo list command/test*      # List commands starting with "test"
  aimgr repo list *pdf*              # List all resources with "pdf" in name
  aimgr repo list --format=json      # Output as JSON with full details
  aimgr repo list --format=yaml      # Output as YAML

Output Format Examples:
  Table format (default):
  ┌──────────────────────┬───────────────────┬──────────┬────────────────────┐
  │         NAME         │      TARGETS      │   SYNC   │    DESCRIPTION     │
  ├──────────────────────┼───────────────────┼──────────┼────────────────────┤
  │ skill/skill-creator  │ claude, opencode  │    ✓     │ Guide for creating │
  │ skill/webapp-testing │ claude            │    *     │ Toolkit for inter  │
  └──────────────────────┴───────────────────┴──────────┴────────────────────┘

  JSON format includes targets array and sync_status field for scripting:
  {
    "resources": [
      {
        "type": "skill",
        "name": "skill-creator",
        "targets": ["claude", "opencode"],
        "sync_status": "in-sync"
      }
    ]
  }

See also:
  aimgr repo list         # List all resources in repository (not just installed)
  aimgr install <resource>  # Install a resource
  aimgr uninstall <resource>  # Uninstall a resource`,
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
	// Get current working directory for installation detection
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

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

	// Create table with new header including TARGETS and SYNC columns
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Name", "Targets", "Sync", "Description")

	// Add commands
	for _, cmd := range commands {
		targets := formatInstalledTargets(projectPath, cmd.Name, cmd.Type)
		syncSymbol := formatSyncStatus(projectPath, fmt.Sprintf("command/%s", cmd.Name), targets != "-")
		desc := truncateString(cmd.Description, 40)
		if err := table.Append(fmt.Sprintf("command/%s", cmd.Name), targets, syncSymbol, desc); err != nil {
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
		targets := formatInstalledTargets(projectPath, skill.Name, skill.Type)
		syncSymbol := formatSyncStatus(projectPath, fmt.Sprintf("skill/%s", skill.Name), targets != "-")
		desc := truncateString(skill.Description, 40)
		if err := table.Append(fmt.Sprintf("skill/%s", skill.Name), targets, syncSymbol, desc); err != nil {
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
		targets := formatInstalledTargets(projectPath, agent.Name, agent.Type)
		syncSymbol := formatSyncStatus(projectPath, fmt.Sprintf("agent/%s", agent.Name), targets != "-")
		desc := truncateString(agent.Description, 40)
		if err := table.Append(fmt.Sprintf("agent/%s", agent.Name), targets, syncSymbol, desc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add empty row between agents and packages if both exist
	if len(agents) > 0 && len(packages) > 0 {
		if err := table.Append("", "", "", ""); err != nil {
			return fmt.Errorf("failed to add separator: %w", err)
		}
	}

	// Add packages
	for _, pkg := range packages {
		desc := truncateString(pkg.Description, 40)
		countStr := fmt.Sprintf("%d resources", pkg.ResourceCount)
		fullDesc := fmt.Sprintf("%s %s", countStr, desc)
		if err := table.Append(fmt.Sprintf("package/%s", pkg.Name), "-", "-", fullDesc); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Render the table
	if err := table.Render(); err != nil {
		return err
	}

	// Print legend
	fmt.Println("\nLegend:")
	fmt.Println("  ✓ = In sync  * = Not in manifest  ⚠ = Not installed  - = No manifest")

	return nil
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
	// Get current working directory for installation detection
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Build ListResourceOutput for each resource with targets and sync status
	enhancedResources := make([]ListResourceOutput, 0, len(resources))
	for _, res := range resources {
		// Get installed targets as Tool slice
		targetTools := getInstalledTargets(projectPath, res.Name, res.Type)

		// Convert Tool slice to string slice
		targetStrings := make([]string, 0, len(targetTools))
		for _, tool := range targetTools {
			targetStrings = append(targetStrings, tool.String())
		}

		// Determine if resource is installed (has at least one target)
		isInstalled := len(targetTools) > 0

		// Build resource reference for manifest lookup
		resourceRef := fmt.Sprintf("%s/%s", res.Type, res.Name)

		// Get sync status
		syncStatus := getSyncStatus(projectPath, resourceRef, isInstalled)

		// Create enhanced resource output
		enhancedResources = append(enhancedResources, ListResourceOutput{
			Resource:   res,
			Targets:    targetStrings,
			SyncStatus: string(syncStatus),
		})
	}

	output := map[string]interface{}{
		"resources": enhancedResources,
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
	// Get current working directory for installation detection
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Build ListResourceOutput for each resource with targets and sync status
	enhancedResources := make([]ListResourceOutput, 0, len(resources))
	for _, res := range resources {
		// Get installed targets as Tool slice
		targetTools := getInstalledTargets(projectPath, res.Name, res.Type)

		// Convert Tool slice to string slice
		targetStrings := make([]string, 0, len(targetTools))
		for _, tool := range targetTools {
			targetStrings = append(targetStrings, tool.String())
		}

		// Determine if resource is installed (has at least one target)
		isInstalled := len(targetTools) > 0

		// Build resource reference for manifest lookup
		resourceRef := fmt.Sprintf("%s/%s", res.Type, res.Name)

		// Get sync status
		syncStatus := getSyncStatus(projectPath, resourceRef, isInstalled)

		// Create enhanced resource output
		enhancedResources = append(enhancedResources, ListResourceOutput{
			Resource:   res,
			Targets:    targetStrings,
			SyncStatus: string(syncStatus),
		})
	}

	output := map[string]interface{}{
		"resources": enhancedResources,
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

// formatInstalledTargets gets the list of installed tools for a resource and formats it as a comma-separated string
func formatInstalledTargets(projectPath string, resourceName string, resourceType resource.ResourceType) string {
	installedTools := getInstalledTargets(projectPath, resourceName, resourceType)

	if len(installedTools) == 0 {
		return "-"
	}

	// Convert tools to string names
	toolNames := make([]string, len(installedTools))
	for i, tool := range installedTools {
		toolNames[i] = tool.String()
	}

	return strings.Join(toolNames, ", ")
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

// getInstalledTargets detects which tools (claude, opencode, copilot) have the given resource installed.
// It checks tool-specific directories for the resource and returns a list of tools where it's found.
//
// Parameters:
//   - projectPath: Path to the project directory
//   - resourceName: Name of the resource (e.g., "test-command", "api/deploy")
//   - resourceType: Type of resource (Command, Skill, Agent)
//
// Returns:
//   - A slice of Tool types where the resource is installed (empty if not installed anywhere)
//
// The function:
//   - Iterates over all supported tools using tools.AllTools()
//   - Checks if each tool supports the given resource type
//   - Verifies if the resource exists in the tool-specific directory
//   - Handles symlinks correctly (checks both existence and validity)
//   - Gracefully handles missing directories (not an error, just not installed)
func getInstalledTargets(projectPath string, resourceName string, resourceType resource.ResourceType) []tools.Tool {
	var installedTools []tools.Tool

	// Iterate over all supported tools
	for _, tool := range tools.AllTools() {
		toolInfo := tools.GetToolInfo(tool)

		// Determine the directory and file path based on resource type
		var resourcePath string
		var supported bool

		switch resourceType {
		case resource.Command:
			if !toolInfo.SupportsCommands {
				continue
			}
			supported = true
			// Commands use .md extension
			resourcePath = filepath.Join(projectPath, toolInfo.CommandsDir, resourceName+".md")

		case resource.Skill:
			if !toolInfo.SupportsSkills {
				continue
			}
			supported = true
			// Skills are directories
			resourcePath = filepath.Join(projectPath, toolInfo.SkillsDir, resourceName)

		case resource.Agent:
			if !toolInfo.SupportsAgents {
				continue
			}
			supported = true
			// Agents use .md extension
			resourcePath = filepath.Join(projectPath, toolInfo.AgentsDir, resourceName+".md")

		default:
			// Unknown resource type, skip
			continue
		}

		// Skip if tool doesn't support this resource type
		if !supported {
			continue
		}

		// Check if the resource exists at the expected path
		// Use Lstat to check symlink existence without following it
		info, err := os.Lstat(resourcePath)
		if err != nil {
			// Resource doesn't exist in this tool's directory (not an error, just not installed)
			continue
		}

		// For symlinks, verify the target is valid
		if info.Mode()&os.ModeSymlink != 0 {
			// Check if symlink target exists
			_, err := os.Stat(resourcePath)
			if err != nil {
				// Symlink target is invalid (broken symlink), skip
				continue
			}
		}

		// Resource is installed for this tool
		installedTools = append(installedTools, tool)
	}

	return installedTools
}
func init() {
	repoCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&formatFlag, "format", "table", "Output format (table|json|yaml)")
	listCmd.Flags().StringVar(&typeFlag, "type", "", "Filter by type (command|skill|agent|package)")
	listCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
}
