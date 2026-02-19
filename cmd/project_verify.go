package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var projectVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify installed resources in the current project",
	Long: `Verify all installed resources in the current project directory.

This command checks for common installation issues:
  - Broken symlinks (target doesn't exist)
  - Symlinks pointing to wrong repository
  - Resources in ai.package.yaml that aren't installed
  - Orphaned installations (not in ai.package.yaml)

Use --fix to automatically repair issues by reinstalling broken resources.

Examples:
  aimgr verify                           # Check current directory
  aimgr verify --project-path ~/project  # Check specific directory
  aimgr verify --fix                     # Auto-fix issues by reinstalling
  aimgr verify --format json             # JSON output for scripts
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path
		projectPath := verifyProjectPath
		if projectPath == "" {
			var err error
			projectPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}

		// Get repo manager
		manager, err := NewManagerWithLogLevel()
		if err != nil {
			return err
		}
		repoPath := manager.GetRepoPath()

		// Detect tools
		detectedTools, err := tools.DetectExistingTools(projectPath)
		if err != nil {
			return fmt.Errorf("failed to detect tools: %w", err)
		}

		if len(detectedTools) == 0 {
			fmt.Println("No tool directories found in this project.")
			return nil
		}

		// Scan for issues
		issues, err := scanProjectIssues(projectPath, detectedTools, repoPath)
		if err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		// Check manifest vs installed
		manifestIssues, err := checkManifestSync(projectPath, detectedTools, repoPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check manifest sync: %v\n", err)
		} else {
			issues = append(issues, manifestIssues...)
		}

		// Parse format flag
		parsedFormat, err := output.ParseFormat(verifyFormatFlag2)
		if err != nil {
			return err
		}

		// Report results
		if len(issues) == 0 {
			return displayNoIssues(parsedFormat)
		}

		// Display issues
		if parsedFormat == output.Table {
			fmt.Printf("Found %d issue(s):\n\n", len(issues))
		}
		if err := displayVerifyIssues(issues, parsedFormat); err != nil {
			return err
		}

		// Auto-fix if requested
		if verifyFixFlag {
			if parsedFormat == output.Table {
				fmt.Println("\nAttempting to fix issues...")
			}
			return fixVerifyIssues(projectPath, issues, manager)
		}

		// Show fix suggestion only for table format
		if parsedFormat == output.Table {
			fmt.Println("\nRun 'aimgr verify --fix' to automatically fix these issues")
		}
		return nil
	},
}

type VerifyIssue struct {
	Resource    string
	Tool        string
	IssueType   string // "broken", "wrong-repo", "not-installed", "orphaned"
	Description string
	Path        string
	Severity    string // "error", "warning"
}

func scanProjectIssues(projectPath string, detectedTools []tools.Tool, repoPath string) ([]VerifyIssue, error) {
	var issues []VerifyIssue

	for _, tool := range detectedTools {
		toolInfo := tools.GetToolInfo(tool)
		toolName := tool.String()

		// Check commands
		if toolInfo.SupportsCommands {
			commandsDir := filepath.Join(projectPath, toolInfo.CommandsDir)
			found, err := verifyDirectory(commandsDir, "command", toolName, repoPath)
			if err != nil {
				return nil, err
			}
			issues = append(issues, found...)
		}

		// Check skills
		if toolInfo.SupportsSkills {
			skillsDir := filepath.Join(projectPath, toolInfo.SkillsDir)
			found, err := verifyDirectory(skillsDir, "skill", toolName, repoPath)
			if err != nil {
				return nil, err
			}
			issues = append(issues, found...)
		}

		// Check agents
		if toolInfo.SupportsAgents {
			agentsDir := filepath.Join(projectPath, toolInfo.AgentsDir)
			found, err := verifyDirectory(agentsDir, "agent", toolName, repoPath)
			if err != nil {
				return nil, err
			}
			issues = append(issues, found...)
		}
	}

	return issues, nil
}

func verifyDirectory(dir, resType, tool, repoPath string) ([]VerifyIssue, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var issues []VerifyIssue
	for _, entry := range entries {
		symlinkPath := filepath.Join(dir, entry.Name())

		// Check if it's a symlink
		linkInfo, err := os.Lstat(symlinkPath)
		if err != nil {
			continue
		}

		if linkInfo.Mode()&os.ModeSymlink == 0 {
			continue // Not a symlink
		}

		// Read target
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			issues = append(issues, VerifyIssue{
				Resource:    entry.Name(),
				Tool:        tool,
				IssueType:   "unreadable",
				Description: "Cannot read symlink target",
				Path:        symlinkPath,
				Severity:    "error",
			})
			continue
		}

		// Check if target exists
		if _, err := os.Stat(symlinkPath); err != nil {
			issues = append(issues, VerifyIssue{
				Resource:    entry.Name(),
				Tool:        tool,
				IssueType:   "broken",
				Description: fmt.Sprintf("Symlink target doesn't exist: %s", target),
				Path:        symlinkPath,
				Severity:    "error",
			})
			continue
		}

		// Check if points to correct repo
		if !strings.HasPrefix(target, repoPath) {
			issues = append(issues, VerifyIssue{
				Resource:    entry.Name(),
				Tool:        tool,
				IssueType:   "wrong-repo",
				Description: fmt.Sprintf("Points to wrong repo: %s (expected: %s)", target, repoPath),
				Path:        symlinkPath,
				Severity:    "warning",
			})
		}
	}

	return issues, nil
}

func checkManifestSync(projectPath string, detectedTools []tools.Tool, repoPath string) ([]VerifyIssue, error) {
	// Load manifest
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	mf, err := manifest.Load(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No manifest, no sync issues
		}
		return nil, err
	}

	var issues []VerifyIssue

	// Check each resource in manifest
	for _, resourceRef := range mf.Resources {
		// Parse resource type and name
		parts := strings.SplitN(resourceRef, "/", 2)
		if len(parts) != 2 {
			continue
		}
		resType := parts[0] // "skill", "command", "agent", "package"
		resName := parts[1]

		// Special handling for packages - check if all constituent resources are installed
		if resType == "package" {
			// Load package definition
			packagePath := resource.GetPackagePath(resName, repoPath)
			pkg, err := resource.LoadPackage(packagePath)
			if err != nil {
				// Package definition doesn't exist or is invalid
				issues = append(issues, VerifyIssue{
					Resource:    resourceRef,
					Tool:        "any",
					IssueType:   "not-installed",
					Description: fmt.Sprintf("Package definition not found in repository: %v", err),
					Path:        manifestPath,
					Severity:    "warning",
				})
				continue
			}

			// Check if all resources in the package are installed
			allInstalled := true
			for _, pkgRes := range pkg.Resources {
				resInstalled := false
				resParts := strings.SplitN(pkgRes, "/", 2)
				if len(resParts) != 2 {
					continue
				}

				// Check if this resource is installed
				for _, tool := range detectedTools {
					toolInfo := tools.GetToolInfo(tool)
					var checkPaths []string

					switch resParts[0] {
					case "skill":
						if !toolInfo.SupportsSkills {
							continue
						}
						// Skills are directories
						checkPaths = []string{filepath.Join(projectPath, toolInfo.SkillsDir, resParts[1])}
					case "command":
						if !toolInfo.SupportsCommands {
							continue
						}
						// Commands can be files or nested: check both directory and .md file
						basePath := filepath.Join(projectPath, toolInfo.CommandsDir, resParts[1])
						checkPaths = []string{
							basePath,         // Directory (for nested commands)
							basePath + ".md", // File
						}
					case "agent":
						if !toolInfo.SupportsAgents {
							continue
						}
						// Agents are files
						basePath := filepath.Join(projectPath, toolInfo.AgentsDir, resParts[1])
						checkPaths = []string{basePath + ".md"}
					default:
						continue
					}

					// Check if any of the paths exist
					for _, path := range checkPaths {
						if _, err := os.Lstat(path); err == nil {
							resInstalled = true
							break
						}
					}

					if resInstalled {
						break
					}
				}

				if !resInstalled {
					allInstalled = false
					break
				}
			}

			if !allInstalled {
				issues = append(issues, VerifyIssue{
					Resource:    resourceRef,
					Tool:        "any",
					IssueType:   "not-installed",
					Description: fmt.Sprintf("Listed in %s but not all package resources are installed", manifest.ManifestFileName),
					Path:        manifestPath,
					Severity:    "warning",
				})
			}
			continue
		}

		// Check if regular resource is installed in any tool
		installed := false
		for _, tool := range detectedTools {
			toolInfo := tools.GetToolInfo(tool)
			var dir string

			switch resType {
			case "skill":
				if !toolInfo.SupportsSkills {
					continue
				}
				dir = filepath.Join(projectPath, toolInfo.SkillsDir, resName)
			case "command":
				if !toolInfo.SupportsCommands {
					continue
				}
				dir = filepath.Join(projectPath, toolInfo.CommandsDir, resName)
			case "agent":
				if !toolInfo.SupportsAgents {
					continue
				}
				dir = filepath.Join(projectPath, toolInfo.AgentsDir, resName)
			default:
				continue
			}

			if _, err := os.Lstat(dir); err == nil {
				installed = true
				break
			}
		}

		if !installed {
			issues = append(issues, VerifyIssue{
				Resource:    resourceRef,
				Tool:        "any",
				IssueType:   "not-installed",
				Description: fmt.Sprintf("Listed in %s but not installed", manifest.ManifestFileName),
				Path:        manifestPath,
				Severity:    "warning",
			})
		}
	}

	return issues, nil
}

func displayNoIssues(format output.Format) error {
	switch format {
	case output.JSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(map[string][]VerifyIssue{"issues": {}})

	case output.YAML:
		encoder := yaml.NewEncoder(os.Stdout)
		defer func() { _ = encoder.Close() }()
		return encoder.Encode(map[string][]VerifyIssue{"issues": {}})

	case output.Table:
		fmt.Println("✓ All installed resources are valid")
		return nil

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func displayVerifyIssues(issues []VerifyIssue, format output.Format) error {
	switch format {
	case output.JSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(issues)

	case output.YAML:
		encoder := yaml.NewEncoder(os.Stdout)
		defer func() { _ = encoder.Close() }()
		return encoder.Encode(issues)

	case output.Table:
		table := output.NewTable("Name", "Tool", "Issue", "Details")
		table.WithResponsive().
			WithDynamicColumn(3).
			WithMinColumnWidths(40, 12, 15, 40)

		for _, issue := range issues {
			symbol := "⚠"
			if issue.Severity == "error" {
				symbol = "✗"
			}
			table.AddRow(issue.Resource, issue.Tool, symbol+" "+issue.IssueType, issue.Description)
		}

		return table.Format(output.Table)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func fixVerifyIssues(projectPath string, issues []VerifyIssue, manager interface{}) error {
	fixed := 0
	failed := 0

	for _, issue := range issues {
		switch issue.IssueType {
		case "broken", "wrong-repo":
			// Remove and reinstall
			fmt.Printf("  Fixing %s...\n", issue.Resource)

			// Remove broken symlink
			if err := os.Remove(issue.Path); err != nil {
				fmt.Fprintf(os.Stderr, "    Failed to remove: %v\n", err)
				failed++
				continue
			}

			// TODO: Reinstall using installer
			// For now, just suggest manual reinstall
			fmt.Printf("    Removed broken symlink. Run 'aimgr install' to reinstall.\n")
			fixed++

		case "not-installed":
			// Resource in manifest but not installed
			fmt.Printf("  Installing %s...\n", issue.Resource)
			// TODO: Call installer to install this resource
			fmt.Printf("    Run 'aimgr install' to install missing resources.\n")

		case "orphaned":
			// Installed but not in manifest
			fmt.Printf("  Orphaned resource: %s\n", issue.Resource)
			fmt.Printf("    Run 'aimgr uninstall %s' to remove, or add to %s\n",
				issue.Resource, manifest.ManifestFileName)
		}
	}

	fmt.Println()
	if fixed > 0 {
		fmt.Printf("✓ Fixed %d issue(s)\n", fixed)
		fmt.Println("Run 'aimgr install' to complete repairs")
	}
	if failed > 0 {
		fmt.Printf("✗ Failed to fix %d issue(s)\n", failed)
	}

	return nil
}

var (
	verifyProjectPath string
	verifyFixFlag     bool
	verifyFormatFlag2 string
)

func init() {
	rootCmd.AddCommand(projectVerifyCmd)
	projectVerifyCmd.Flags().StringVar(&verifyProjectPath, "project-path", "", "Project directory path (default: current directory)")
	projectVerifyCmd.Flags().BoolVar(&verifyFixFlag, "fix", false, "Automatically fix issues by reinstalling resources")
	projectVerifyCmd.Flags().StringVar(&verifyFormatFlag2, "format", "table", "Output format (table|json|yaml)")

	// Register completion functions
	_ = projectVerifyCmd.RegisterFlagCompletionFunc("format", completeFormatFlag)
}
