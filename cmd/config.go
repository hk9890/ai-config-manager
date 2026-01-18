package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/config"
	"github.com/hans-m-leitner/ai-config-manager/pkg/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage ai-repo configuration",
	Long: `View and manage ai-repo configuration settings.

Configuration is stored in ~/.ai-repo.yaml and controls default behavior
for ai-repo commands.`,
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value.

Available keys:
  default-tool - The default AI tool to use (claude, opencode, or copilot)

Example:
  ai-repo config get default-tool`,
	Args: cobra.ExactArgs(1),
	RunE: configGet,
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  default-tool - The default AI tool to use

Valid tools:
  claude   - Claude Code (supports commands and skills)
  opencode - OpenCode (supports commands and skills)
  copilot  - GitHub Copilot (supports skills only)

Example:
  ai-repo config set default-tool opencode`,
	Args: cobra.ExactArgs(2),
	RunE: configSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

func configGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	if key != "default-tool" {
		return fmt.Errorf("unknown config key: %s (available: default-tool)", key)
	}

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Load config
	cfg, err := config.Load(home)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Print value
	fmt.Printf("default-tool: %s\n", cfg.DefaultTool)
	return nil
}

func configSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	if key != "default-tool" {
		return fmt.Errorf("unknown config key: %s (available: default-tool)", key)
	}

	// Validate tool name
	if _, err := tools.ParseTool(value); err != nil {
		return fmt.Errorf("invalid tool: %w\nValid tools: claude, opencode, copilot", err)
	}

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Load existing config (or create new)
	cfg, err := config.Load(home)
	if err != nil {
		// If config doesn't exist or has issues, create a new one
		cfg = &config.Config{}
	}

	// Update value
	cfg.DefaultTool = value

	// Save config
	configPath := filepath.Join(home, config.DefaultConfigFileName)
	if err := saveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("✓ Set default-tool to '%s'\n", value)
	return nil
}

func saveConfig(path string, cfg *config.Config) error {
	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
