package cmd

import (
	"fmt"

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
  - Creates ai.repo.yaml manifest for source tracking
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

		// Init() handles everything: directories, git init, .gitignore, initial commit
		if err := mgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}

		fmt.Printf("✓ Repository initialized at: %s\n", mgr.GetRepoPath())
		fmt.Println("\n✨ Repository ready for git-tracked operations")
		fmt.Println("   Created ai.repo.yaml manifest for source tracking")
		fmt.Println("   All aimgr operations (add, sync, remove) will be tracked in git")

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoInitCmd)
}
