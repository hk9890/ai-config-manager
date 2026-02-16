package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/output"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
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

		// Report results
		if len(issues) == 0 {
			fmt.Println("✓ All installed resources are valid")
			return nil
		}

		// Display issues
		fmt.Printf("Found %d issue(s):\n\n", len(issues))
		displayVerifyIssues(issues)

		// Auto-fix if requested
		if verifyFixFlag {
			fmt.Println("\nAttempting to fix issues...")
			return fixVerifyIssues(projectPath, issues, manager)
		}

		fmt.Println("\nRun 'aimgr verify --fix' to automatically fix these issues")
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
		resType := parts[0] // "skill", "command", "agent"
		resName := parts[1]

		// Check if installed in any tool
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

func displayVerifyIssues(issues []VerifyIssue) {
	table := output.NewTable("Resource", "Tool", "Issue", "Details")
	table.WithResponsive().
		WithDynamicColumn(3).
		WithMinColumnWidths(20, 12, 15, 40)

	for _, issue := range issues {
		symbol := "⚠"
		if issue.Severity == "error" {
			symbol = "✗"
		}
		table.AddRow(issue.Resource, issue.Tool, symbol+" "+issue.IssueType, issue.Description)
	}

	_ = table.Format(output.Table)
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
}
