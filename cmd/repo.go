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

Use 'aimgr repo --help' to see all available subcommands.`,
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
