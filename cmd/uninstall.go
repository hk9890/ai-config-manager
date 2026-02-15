package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/install"
	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
)

var (
	uninstallProjectPathFlag string
	uninstallForceFlag       bool
)

// uninstallResult tracks the result of uninstalling a single resource
type uninstallResult struct {
	resourceType resource.ResourceType
	name         string
	success      bool
	skipped      bool
	message      string
	toolsRemoved []tools.Tool
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [resource]...",
	Short: "Uninstall resources from a project",
	Long: `Uninstall one or more resources (commands, skills, or agents) from a project.

If no resources are specified, uninstalls all resources that are currently installed
(all symlinks pointing to the aimgr repository).

Resources are specified using the format 'type/name':
  - command/name (or commands/name)
  - skill/name (or skills/name)
  - agent/name (or agents/name)


Pattern matching:
  - Use * to match any sequence of characters
  - Use ? to match any single character
  - Use [abc] to match any character in the set
  - Use {a,b} to match alternatives
  - Patterns are expanded by scanning installed resources in the project

Safety:
  - Only removes symlinks that point to the aimgr repository
  - Skips non-symlinks with a warning
  - Skips symlinks pointing to other locations with a warning

Multi-tool behavior:
  - Automatically detects all tool directories (.claude, .opencode, .github/skills)
  - Removes resource from ALL detected tool directories
  - Shows summary of what was removed from where

Examples:
  # Uninstall all installed resources
  aimgr uninstall

  # Uninstall a single skill
  aimgr uninstall skill/pdf-processing

  # Uninstall multiple resources at once
  aimgr uninstall skill/foo skill/bar command/test agent/my-agent

  # Uninstall all skills
  aimgr uninstall "skill/*"

  # Uninstall all test resources (any type)
  aimgr uninstall "*test*"

  # Uninstall skills starting with "pdf"
  aimgr uninstall "skill/pdf*"

  # Uninstall multiple patterns
  aimgr uninstall "skill/pdf*" "command/test*"

  # Uninstall from a specific project
  aimgr uninstall skill/foo --project-path ~/my-project

  # Force uninstall (placeholder for future confirmation prompts)
  aimgr uninstall command/review --force`,
	Args:              cobra.ArbitraryArgs, // Allow 0 or more args
	ValidArgsFunction: completeResourceArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path (current directory or flag)
		projectPath := uninstallProjectPathFlag
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		// Create repo manager to get repository path
		manager, err := repo.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create repository manager: %w", err)
		}
		repoPath := manager.GetRepoPath()

		// Create installer (auto-detect existing tools)
		installer, err := install.NewInstaller(projectPath, nil)
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}

		// Handle zero-arg uninstall (uninstall all)
		if len(args) == 0 {
			return uninstallAll(projectPath, repoPath, installer.GetTargetTools())
		}

		// Check if uninstalling a package
		if len(args) == 1 && strings.HasPrefix(args[0], "package/") {
			packageName := strings.TrimPrefix(args[0], "package/")
			// Uninstall package
			return uninstallPackage(packageName, installer, manager)
		}

		// Expand patterns in arguments
		var expandedArgs []string
		for _, arg := range args {
			// Check if this is a pattern or exact name
			_, _, isPattern := pattern.ParsePattern(arg)

			if isPattern {
				// Expand the pattern
				expanded, err := expandUninstallPattern(projectPath, arg, installer.GetTargetTools())
				if err != nil {
					return fmt.Errorf("failed to expand pattern '%s': %w", arg, err)
				}
				if len(expanded) == 0 {
					fmt.Printf("Warning: pattern '%s' matches no installed resources\n", arg)
				}
				expandedArgs = append(expandedArgs, expanded...)
			} else {
				// Not a pattern, add as-is (will be validated by processUninstall)
				expandedArgs = append(expandedArgs, arg)
			}
		}

		// Deduplicate the expanded list
		expandedArgs = deduplicateStrings(expandedArgs)

		// Track results
		var results []uninstallResult

		// Process each resource argument
		for _, arg := range expandedArgs {
			result := processUninstall(arg, projectPath, repoPath, installer.GetTargetTools(), manager)
			results = append(results, result)
		}
		// Print results
		printUninstallSummary(results)

		// Return error if any resource failed
		for _, result := range results {
			if !result.success && !result.skipped {
				return fmt.Errorf("some resources failed to uninstall")
			}
		}

		return nil
	},
}

// uninstallAll uninstalls all resources currently installed in the project
func uninstallAll(projectPath string, repoPath string, targetTools []tools.Tool) error {
	fmt.Println("Uninstalling all resources from project...")
	fmt.Println()

	var results []uninstallResult

	// Scan each tool directory
	for _, tool := range targetTools {
		toolInfo := tools.GetToolInfo(tool)

		// Scan commands
		if toolInfo.SupportsCommands {
			commandsResults := uninstallAllFromDir(projectPath, repoPath, toolInfo.CommandsDir, resource.Command, tool)
			results = append(results, commandsResults...)
		}

		// Scan skills
		if toolInfo.SupportsSkills {
			skillsResults := uninstallAllFromDir(projectPath, repoPath, toolInfo.SkillsDir, resource.Skill, tool)
			results = append(results, skillsResults...)
		}

		// Scan agents
		if toolInfo.SupportsAgents {
			agentsResults := uninstallAllFromDir(projectPath, repoPath, toolInfo.AgentsDir, resource.Agent, tool)
			results = append(results, agentsResults...)
		}
	}

	// Print results
	printUninstallSummary(results)

	// Return error if any resource failed
	for _, result := range results {
		if !result.success && !result.skipped {
			return fmt.Errorf("some resources failed to uninstall")
		}
	}

	return nil
}

// uninstallAllFromDir scans a tool directory and uninstalls all symlinks pointing to the repo
func uninstallAllFromDir(projectPath, repoPath, toolDir string, resourceType resource.ResourceType, tool tools.Tool) []uninstallResult {
	var results []uninstallResult

	// Build full path
	fullPath := filepath.Join(projectPath, toolDir)

	// Read directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		// Directory doesn't exist or can't be read, skip
		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		entryPath := filepath.Join(fullPath, name)

		// Get entry info
		info, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}

		// Check if it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			// Not a symlink, skip
			continue
		}

		// Read symlink target
		target, err := os.Readlink(entryPath)
		if err != nil {
			continue
		}

		// Resolve to absolute path for comparison
		absTarget, err := filepath.Abs(target)
		if err != nil {
			absTarget = target
		}

		// Check if target points to our repository
		if !strings.HasPrefix(absTarget, repoPath) {
			// Not managed by aimgr, skip
			continue
		}

		// Extract resource name
		var resourceName string
		if resourceType == resource.Command || resourceType == resource.Agent {
			// Remove .md extension
			if strings.HasSuffix(name, ".md") {
				resourceName = strings.TrimSuffix(name, ".md")
			} else {
				continue
			}
		} else {
			// Skill - name is the directory name
			resourceName = name
		}

		// Remove the symlink
		if err := os.Remove(entryPath); err != nil {
			results = append(results, uninstallResult{
				resourceType: resourceType,
				name:         resourceName,
				success:      false,
				message:      fmt.Sprintf("failed to remove: %v", err),
			})
			continue
		}

		// Record success
		results = append(results, uninstallResult{
			resourceType: resourceType,
			name:         resourceName,
			success:      true,
			toolsRemoved: []tools.Tool{tool},
		})
	}

	return results
}

// processUninstall processes uninstalling a single resource
func processUninstall(arg string, projectPath string, repoPath string, targetTools []tools.Tool, manager *repo.Manager) uninstallResult {
	// Parse resource argument
	resourceType, name, err := parseResourceArg(arg)
	if err != nil {
		return uninstallResult{
			name:    arg,
			success: false,
			message: err.Error(),
		}
	}

	result := uninstallResult{
		resourceType: resourceType,
		name:         name,
		toolsRemoved: []tools.Tool{},
	}

	removed := false
	skipped := false
	var messages []string

	// Try to uninstall from all target tools
	for _, tool := range targetTools {
		toolInfo := tools.GetToolInfo(tool)
		var symlinkPath string

		// Determine symlink path based on resource type
		switch resourceType {
		case resource.Command:
			if !toolInfo.SupportsCommands {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.CommandsDir, name+".md")
		case resource.Skill:
			if !toolInfo.SupportsSkills {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.SkillsDir, name)
		case resource.Agent:
			if !toolInfo.SupportsAgents {
				continue
			}
			symlinkPath = filepath.Join(projectPath, toolInfo.AgentsDir, name+".md")
		default:
			result.success = false
			result.message = fmt.Sprintf("invalid resource type: %s", resourceType)
			return result
		}

		// Check if symlink exists
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			// Doesn't exist in this tool directory, continue
			continue
		}

		// Verify it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			messages = append(messages, fmt.Sprintf("%s: not a symlink (skipping)", tool))
			skipped = true
			continue
		}

		// Read symlink target
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			messages = append(messages, fmt.Sprintf("%s: failed to read symlink: %v", tool, err))
			continue
		}

		// Resolve to absolute path for comparison
		absTarget, err := filepath.Abs(target)
		if err != nil {
			// If we can't resolve, use the original target
			absTarget = target
		}

		// Check if target points to our repository
		if !strings.HasPrefix(absTarget, repoPath) {
			messages = append(messages, fmt.Sprintf("%s: not managed by aimgr (points to %s, skipping)", tool, target))
			skipped = true
			continue
		}

		// Remove the symlink
		if err := os.Remove(symlinkPath); err != nil {
			messages = append(messages, fmt.Sprintf("%s: failed to remove: %v", tool, err))
			continue
		}

		// Log successful uninstallation
		if logger := manager.GetLogger(); logger != nil {
			logger.Info("resource uninstalled",
				"operation", "uninstall",
				"resource_type", resourceType,
				"resource_name", name,
				"tool", tool.String(),
				"dest_path", symlinkPath,
			)
		}

		removed = true
		result.toolsRemoved = append(result.toolsRemoved, tool)
	}

	// Determine overall result
	if removed {
		result.success = true
		if len(messages) > 0 {
			result.message = strings.Join(messages, "; ")
		}
	} else if skipped {
		result.skipped = true
		result.message = strings.Join(messages, "; ")
	} else {
		result.success = false
		if len(messages) > 0 {
			result.message = strings.Join(messages, "; ")
		} else {
			result.message = fmt.Sprintf("%s/%s not installed in project", resourceType, name)
		}
	}

	return result
}

// printUninstallSummary prints a summary of uninstall results
func printUninstallSummary(results []uninstallResult) {
	successCount := 0
	skipCount := 0
	failCount := 0

	for _, result := range results {
		if result.success {
			successCount++
			// Print success
			fmt.Printf("✓ Uninstalled %s '%s'\n", result.resourceType, result.name)
			for _, tool := range result.toolsRemoved {
				toolInfo := tools.GetToolInfo(tool)
				var dirName string
				switch result.resourceType {
				case resource.Command:
					dirName = toolInfo.CommandsDir
				case resource.Skill:
					dirName = toolInfo.SkillsDir
				case resource.Agent:
					dirName = toolInfo.AgentsDir
				}
				fmt.Printf("  → Removed from %s (%s)\n", tool, dirName)
			}
			if result.message != "" {
				fmt.Printf("  Note: %s\n", result.message)
			}
		} else if result.skipped {
			skipCount++
			// Print skipped
			fmt.Printf("⊘ Skipped %s '%s': %s\n", result.resourceType, result.name, result.message)
		} else {
			failCount++
			// Print failure
			if result.resourceType != "" {
				fmt.Printf("✗ Failed to uninstall %s '%s': %s\n", result.resourceType, result.name, result.message)
			} else {
				fmt.Printf("✗ Failed: %s\n", result.message)
			}
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Summary: %d uninstalled, %d skipped, %d failed\n", successCount, skipCount, failCount)
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringVar(&uninstallProjectPathFlag, "project-path", "", "Project directory path (default: current directory)")
	uninstallCmd.Flags().BoolVarP(&uninstallForceFlag, "force", "f", false, "Force uninstall (placeholder for future use)")
}

// expandUninstallPattern finds installed resources matching a pattern
func expandUninstallPattern(projectPath, resourceArg string, detectedTools []tools.Tool) ([]string, error) {
	// Parse pattern
	resourceType, _, isPattern := pattern.ParsePattern(resourceArg)

	// If not a pattern, return as-is
	if !isPattern {
		return []string{resourceArg}, nil
	}

	// Create matcher
	matcher, err := pattern.NewMatcher(resourceArg)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	// Scan installed resources from tool directories
	var matches []string

	for _, tool := range detectedTools {
		toolInfo := tools.GetToolInfo(tool)

		// Scan each resource type directory
		if resourceType == "" || resourceType == resource.Command {
			if toolInfo.SupportsCommands {
				foundMatches := scanToolDir(projectPath, toolInfo.CommandsDir, resource.Command, matcher)
				matches = append(matches, foundMatches...)
			}
		}
		if resourceType == "" || resourceType == resource.Skill {
			if toolInfo.SupportsSkills {
				foundMatches := scanToolDir(projectPath, toolInfo.SkillsDir, resource.Skill, matcher)
				matches = append(matches, foundMatches...)
			}
		}
		if resourceType == "" || resourceType == resource.Agent {
			if toolInfo.SupportsAgents {
				foundMatches := scanToolDir(projectPath, toolInfo.AgentsDir, resource.Agent, matcher)
				matches = append(matches, foundMatches...)
			}
		}
	}

	// Deduplicate
	return deduplicateStrings(matches), nil
}

// scanToolDir scans a tool directory for resources matching a pattern
func scanToolDir(projectPath, toolDir string, resourceType resource.ResourceType, matcher *pattern.Matcher) []string {
	// Build full path
	fullPath := filepath.Join(projectPath, toolDir)

	// Read directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		// Directory doesn't exist or can't be read
		return nil
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()

		// For commands and agents, remove .md extension
		if resourceType == resource.Command || resourceType == resource.Agent {
			if strings.HasSuffix(name, ".md") {
				name = strings.TrimSuffix(name, ".md")
			} else {
				// Skip non-.md files
				continue
			}
		}

		// For skills, name is the directory name (no extension to remove)

		// Test if name matches the pattern
		if matcher.MatchName(name) {
			// Format as type/name
			matches = append(matches, fmt.Sprintf("%s/%s", resourceType, name))
		}
	}

	return matches
}

// deduplicateStrings removes duplicate strings from a slice
func deduplicateStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range input {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// uninstallPackage uninstalls all resources from a package
func uninstallPackage(packageName string, installer *install.Installer, manager *repo.Manager) error {
	repoPath := manager.GetRepoPath()
	pkgPath := resource.GetPackagePath(packageName, repoPath)

	// Load package
	pkg, err := resource.LoadPackage(pkgPath)
	if err != nil {
		return fmt.Errorf("package '%s' not found in repository: %w", packageName, err)
	}

	fmt.Printf("Uninstalling package: %s\n", pkg.Name)
	if pkg.Description != "" {
		fmt.Printf("Description: %s\n", pkg.Description)
	}
	fmt.Println()

	uninstalled := 0
	skipped := 0
	errors := []string{}

	// Uninstall each resource
	for _, ref := range pkg.Resources {
		// Parse type/name format
		resType, resName, err := resource.ParseResourceReference(ref)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", ref, err))
			continue
		}

		// Check if resource is installed
		if !installer.IsInstalled(resName, resType) {
			fmt.Printf("  ○ %s - not installed, skipping\n", ref)
			skipped++
			continue
		}

		// Uninstall the resource
		err = installer.Uninstall(resName, resType, manager)
		if err != nil {
			fmt.Printf("  ✗ %s - failed to uninstall: %v\n", ref, err)
			errors = append(errors, fmt.Sprintf("%s: %v", ref, err))
		} else {
			fmt.Printf("  ✓ %s\n", ref)
			uninstalled++
		}
	}

	// Print summary
	fmt.Println()
	if skipped > 0 {
		fmt.Printf("Skipped %d resource(s) not installed\n", skipped)
	}
	if len(errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range errors {
			fmt.Printf("  ✗ %s\n", e)
		}
	}

	totalResources := len(pkg.Resources)
	fmt.Printf("Uninstalled %d of %d resources from package '%s'\n", uninstalled, totalResources, pkg.Name)

	if len(errors) > 0 {
		return fmt.Errorf("package uninstallation completed with errors")
	}

	return nil
}
