package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hk9890/ai-config-manager/pkg/repo"
)

// repoInitCmd represents the init command
var repoInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the aimgr repository with git tracking",
	Long: `Initialize the aimgr repository directory structure and git repository.

This command:
  - Creates the repository directory structure (commands/, skills/, agents/, packages/)
  - Initializes a git repository for tracking changes
  - Creates .gitignore to exclude .workspace/ cache directory

The repository location is determined by (in order of precedence):
  1. AIMGR_REPO_PATH environment variable
  2. repo.path in config file (~/.config/aimgr/aimgr.yaml)
  3. Default XDG location (~/.local/share/ai-config/repo/)

This command is idempotent - safe to run multiple times.

Examples:
  aimgr repo init                    # Initialize with default location
  AIMGR_REPO_PATH=/custom aimgr repo init  # Initialize at custom location`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := repo.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create manager: %w", err)
		}

		repoPath := mgr.GetRepoPath()

		// Check if already initialized
		gitDir := filepath.Join(repoPath, ".git")
		alreadyGit := false
		if _, err := os.Stat(gitDir); err == nil {
			alreadyGit = true
		}

		// Initialize directory structure
		if err := mgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize repository structure: %w", err)
		}

		fmt.Printf("✓ Repository structure initialized at: %s\n", repoPath)

		// Initialize git repository if not already done
		if alreadyGit {
			fmt.Println("✓ Git repository already initialized")
		} else {
			gitCmd := exec.Command("git", "init")
			gitCmd.Dir = repoPath
			if output, err := gitCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to initialize git repository: %w\nOutput: %s", err, output)
			}
			fmt.Println("✓ Git repository initialized")
		}

		// Create .gitignore
		gitignorePath := filepath.Join(repoPath, ".gitignore")
		gitignoreContent := `# aimgr workspace cache (Git clones for remote sources)
.workspace/

# macOS
.DS_Store

# Editor files
*.swp
*.swo
*~
.vscode/
.idea/
`

		// Check if .gitignore already exists
		if _, err := os.Stat(gitignorePath); err == nil {
			// .gitignore exists - check if it contains .workspace/
			content, err := os.ReadFile(gitignorePath)
			if err != nil {
				return fmt.Errorf("failed to read .gitignore: %w", err)
			}

			// If .workspace/ is already in .gitignore, we're done
			if containsWorkspaceIgnore(string(content)) {
				fmt.Println("✓ .gitignore already configured")
			} else {
				// Append to existing .gitignore
				f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open .gitignore for append: %w", err)
				}
				defer f.Close()

				if _, err := f.WriteString("\n" + gitignoreContent); err != nil {
					return fmt.Errorf("failed to append to .gitignore: %w", err)
				}
				fmt.Println("✓ .gitignore updated with .workspace/ entry")
			}
		} else {
			// Create new .gitignore
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
				return fmt.Errorf("failed to create .gitignore: %w", err)
			}
			fmt.Println("✓ .gitignore created")
		}

		// Initial commit
		if !alreadyGit {
			// Add .gitignore
			addCmd := exec.Command("git", "add", ".gitignore")
			addCmd.Dir = repoPath
			if output, err := addCmd.CombinedOutput(); err != nil {
				// Don't fail on add error - might not have anything to add
				fmt.Printf("Warning: git add failed: %s\n", output)
			}

			// Create initial commit
			commitCmd := exec.Command("git", "commit", "-m", "aimgr: initialize repository")
			commitCmd.Dir = repoPath
			if output, err := commitCmd.CombinedOutput(); err != nil {
				// Don't fail on commit error - might not have anything to commit
				fmt.Printf("Warning: git commit failed: %s\n", output)
			} else {
				fmt.Println("✓ Initial commit created")
			}
		}

		fmt.Println("\n✨ Repository ready for git-tracked operations")
		fmt.Println("   All aimgr operations (import, sync, remove) will be tracked in git")

		return nil
	},
}

// containsWorkspaceIgnore checks if .gitignore contains .workspace/ entry
func containsWorkspaceIgnore(content string) bool {
	lines := []string{".workspace/", ".workspace", "/.workspace/", "/.workspace"}
	for _, line := range lines {
		if strings.Contains(content, line) {
			return true
		}
	}
	return false
}

func init() {
	repoCmd.AddCommand(repoInitCmd)
}
