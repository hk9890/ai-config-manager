package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hans-m-leitner/ai-config-manager/pkg/tools"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigFileName is the name of the config file
	DefaultConfigFileName = ".ai-repo.yaml"
	// DefaultTool is the default tool when not specified
	DefaultToolValue = "claude"
)

// Config represents the application configuration
type Config struct {
	// DefaultTool specifies which AI tool to use by default
	// Valid values: claude, opencode, copilot
	DefaultTool string `yaml:"default-tool"`
}

// Load loads the configuration from file in the specified directory
// If no config file exists, returns default configuration
func Load(projectPath string) (*Config, error) {
	configPath := filepath.Join(projectPath, DefaultConfigFileName)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			DefaultTool: DefaultToolValue,
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

	// Apply defaults if not set
	if config.DefaultTool == "" {
		config.DefaultTool = DefaultToolValue
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate default-tool
	if c.DefaultTool != "" {
		if _, err := tools.ParseTool(c.DefaultTool); err != nil {
			return fmt.Errorf("default-tool: %w", err)
		}
	}
	return nil
}

// GetDefaultTool returns the configured default tool
func (c *Config) GetDefaultTool() (tools.Tool, error) {
	toolName := c.DefaultTool
	if toolName == "" {
		toolName = DefaultToolValue
	}
	return tools.ParseTool(toolName)
}
