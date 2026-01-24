package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/hk9890/ai-config-manager/pkg/pattern"
	"github.com/hk9890/ai-config-manager/pkg/tools"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFileName is the name of the config file (without leading dot for XDG)
	DefaultConfigFileName = "aimgr.yaml"
	// OldConfigFileName is the legacy config file name in home directory (ai-repo for migration)
	OldConfigFileName = ".ai-repo.yaml"
	// DefaultToolValue is the default tool when not specified (legacy)
	DefaultToolValue = "claude"
)

var (
	// DefaultTargets is the default installation targets when not specified
	DefaultTargets = []string{"claude"}
)

// Config represents the application configuration
type Config struct {
	// Install configuration for default installation targets
	Install InstallConfig `yaml:"install"`

	// Sync configuration for syncing resources from external sources
	Sync SyncConfig `yaml:"sync"`

	// DefaultTool is deprecated - use Install.Targets instead
	// Kept for backward compatibility during migration
	DefaultTool string `yaml:"default-tool,omitempty"`
}

// InstallConfig holds installation-related configuration
type InstallConfig struct {
	// Targets specifies which AI tools to install to by default
	// Valid values: claude, opencode, copilot
	Targets []string `yaml:"targets"`
}

// SyncSource represents a source for syncing resources
type SyncSource struct {
	// URL is the source URL (required, e.g., "https://github.com/owner/repo" or "gh:owner/repo")
	URL string `yaml:"url"`
	// Filter is an optional glob pattern to filter resources (e.g., "skill/*", "skill/pdf*")
	Filter string `yaml:"filter,omitempty"`
}

// SyncConfig holds sync-related configuration
type SyncConfig struct {
	// Sources is a list of sources to sync from
	Sources []SyncSource `yaml:"sources"`
}

// GetConfigPath returns the path to the config file in XDG config directory
// Returns ~/.config/aimgr/aimgr.yaml
func GetConfigPath() (string, error) {
	configDir := filepath.Join(xdg.ConfigHome, "aimgr")
	return filepath.Join(configDir, DefaultConfigFileName), nil
}

// getOldConfigPath returns the path to the legacy config file in home directory
// Returns ~/.ai-repo.yaml
func getOldConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, OldConfigFileName), nil
}

// Load loads the configuration from file in the specified directory
// If no config file exists, returns default configuration
// For backward compatibility, checks old location (~/.ai-repo.yaml) and migrates if found
func Load(projectPath string) (*Config, error) {
	configPath := filepath.Join(projectPath, DefaultConfigFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			Install: InstallConfig{
				Targets: DefaultTargets,
			},
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Migrate from old format if needed
	if err := config.migrate(); err != nil {
		return nil, fmt.Errorf("migrating config: %w", err)
	}

	// Apply defaults if not set
	if len(config.Install.Targets) == 0 {
		config.Install.Targets = DefaultTargets
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// LoadGlobal loads the global configuration from XDG config directory
// Checks ~/.config/aimgr/aimgr.yaml and falls back to ~/.ai-repo.yaml
// Automatically migrates from old location if found
func LoadGlobal() (*Config, error) {
	// Get new config path
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("getting config path: %w", err)
	}

	// Check if new config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Check for old config location
		oldPath, err := getOldConfigPath()
		if err != nil {
			return nil, fmt.Errorf("getting old config path: %w", err)
		}

		if _, err := os.Stat(oldPath); err == nil {
			// Old config exists - migrate it
			if err := migrateConfig(oldPath, configPath); err != nil {
				return nil, fmt.Errorf("migrating config: %w", err)
			}
			fmt.Fprintf(os.Stderr, "⚠️  Migrated config from %s to %s\n", oldPath, configPath)
			fmt.Fprintf(os.Stderr, "   Old config file left intact for safety\n")
		} else {
			// No config exists - return default
			return &Config{
				Install: InstallConfig{
					Targets: DefaultTargets,
				},
			}, nil
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Migrate from old format if needed
	if err := config.migrate(); err != nil {
		return nil, fmt.Errorf("migrating config: %w", err)
	}

	// Apply defaults if not set
	if len(config.Install.Targets) == 0 {
		config.Install.Targets = DefaultTargets
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// migrateConfig copies config from old location to new XDG location
func migrateConfig(oldPath, newPath string) error {
	// Read old config
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("reading old config: %w", err)
	}

	// Ensure new config directory exists
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Write to new location
	if err := os.WriteFile(newPath, data, 0644); err != nil {
		return fmt.Errorf("writing new config: %w", err)
	}

	return nil
}

// migrate handles migration from old config format (default-tool) to new format (install.targets)
func (c *Config) migrate() error {
	// If we have the old default-tool field set, migrate it
	if c.DefaultTool != "" && len(c.Install.Targets) == 0 {
		// Migrate single tool to targets array
		c.Install.Targets = []string{c.DefaultTool}
		// Clear the old field to avoid confusion
		c.DefaultTool = ""
	}
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate install targets
	for _, target := range c.Install.Targets {
		if _, err := tools.ParseTool(target); err != nil {
			return fmt.Errorf("install.targets: invalid tool '%s': %w", target, err)
		}
	}

	// Validate sync sources
	for i, source := range c.Sync.Sources {
		// Validate URL is non-empty
		if source.URL == "" {
			return fmt.Errorf("sync.sources[%d]: url cannot be empty", i)
		}

		// Validate filter pattern if present
		if source.Filter != "" {
			_, err := pattern.NewMatcher(source.Filter)
			if err != nil {
				return fmt.Errorf("sync.sources[%d]: invalid filter pattern '%s': %w", i, source.Filter, err)
			}
		}
	}

	// Validate legacy default-tool if present (for backward compatibility)
	if c.DefaultTool != "" {
		if _, err := tools.ParseTool(c.DefaultTool); err != nil {
			return fmt.Errorf("default-tool: %w", err)
		}
	}

	return nil
}

// GetDefaultTool returns the configured default tool (first target)
// Deprecated: Use GetDefaultTargets() instead
func (c *Config) GetDefaultTool() (tools.Tool, error) {
	targets, err := c.GetDefaultTargets()
	if err != nil {
		return -1, err
	}
	if len(targets) == 0 {
		return tools.ParseTool(DefaultToolValue)
	}
	return targets[0], nil
}

// GetDefaultTargets returns the configured default installation targets
func (c *Config) GetDefaultTargets() ([]tools.Tool, error) {
	targetStrs := c.Install.Targets
	if len(targetStrs) == 0 {
		targetStrs = DefaultTargets
	}

	targets := make([]tools.Tool, 0, len(targetStrs))
	for _, targetStr := range targetStrs {
		tool, err := tools.ParseTool(targetStr)
		if err != nil {
			return nil, fmt.Errorf("invalid target '%s': %w", targetStr, err)
		}
		targets = append(targets, tool)
	}

	return targets, nil
}
