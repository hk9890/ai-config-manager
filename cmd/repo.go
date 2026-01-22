package cmd

import (
	"github.com/spf13/cobra"
)

// repoCmd represents the repo command group
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage resources in the aimgr repository",
	Long: `Manage resources in the aimgr repository.

The repo command group provides subcommands for adding, removing, and listing
resources (commands, skills, and agents) in the centralized aimgr repository.

Available subcommands:
  add     - Add resources to the repository
  remove  - Remove resources from the repository
  list    - List resources in the repository`,
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
