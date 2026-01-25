package cmd

import (
	"github.com/spf13/cobra"
)

// marketplaceCmd represents the marketplace command group
var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Import Claude marketplace configurations",
	Long: `Import Claude marketplace configurations and generate packages.

The marketplace command group provides subcommands for importing Claude marketplace
configurations (marketplace.json files) and automatically generating aimgr packages
from them.

Available subcommands:
  import  - Import marketplace configurations and generate packages`,
}

func init() {
	rootCmd.AddCommand(marketplaceCmd)
}
