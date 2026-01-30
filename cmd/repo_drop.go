package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hk9890/ai-config-manager/pkg/repo"
)

// repoDropCmd represents the drop command
var repoDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Delete the entire repository",
	Long: `Delete all resources from the repository. Requires --force flag.

WARNING: This is destructive and cannot be undone.
Use 'repo sync' after drop to rebuild from configured sources.

The drop command removes the entire repository directory and recreates
an empty structure with the standard subdirectories (commands, skills,
agents, packages).

When to use:
  - To completely reset the repository
  - To remove all cached resources at once
  - Before migrating to a new repository location

Examples:
  aimgr repo drop --force          # Delete and recreate repository`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("refusing to drop repo without --force flag")
		}

		mgr, err := repo.NewManager()
		if err != nil {
			return err
		}

		if err := mgr.Drop(); err != nil {
			return err
		}

		fmt.Println("✓ Repository dropped successfully")
		fmt.Println("  Run 'aimgr repo sync' to rebuild from sources")
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoDropCmd)
	repoDropCmd.Flags().Bool("force", false, "Confirm deletion")
	repoDropCmd.MarkFlagRequired("force")
}
