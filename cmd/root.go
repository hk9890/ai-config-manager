package cmd

import (
	"fmt"
	"os"

	"github.com/hk9890/ai-config-manager/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	versionFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ai-repo",
	Short: "Manage AI resources (commands and skills) for LLM tools",
	Long: `ai-repo is a CLI tool for discovering, installing, and managing
AI resources like commands and skills across different LLM tools
(Claude Code, OpenCode, etc.).

It helps you organize and share reusable AI configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println(version.GetVersion())
			os.Exit(0)
		}
		// If no version flag and no subcommand, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/ai-repo/ai-repo.yaml)")

	// Version flag
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in XDG config directory first (~/.config/ai-repo/)
		configDir := home + "/.config/ai-repo"
		viper.AddConfigPath(configDir)

		// Also search in home directory for backward compatibility
		viper.AddConfigPath(home)

		viper.SetConfigType("yaml")
		viper.SetConfigName("ai-repo") // Will find both ai-repo.yaml and .ai-repo.yaml
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig() // read in config file if available
}
