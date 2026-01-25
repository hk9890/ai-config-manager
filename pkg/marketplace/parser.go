package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
)

// MarketplaceConfig represents a Claude marketplace.json configuration file
type MarketplaceConfig struct {
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description"`
	Owner       *Author  `json:"owner,omitempty"`
	Plugins     []Plugin `json:"plugins"`
}

// Plugin represents an individual plugin in the marketplace
type Plugin struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Source      string  `json:"source"`
	Category    string  `json:"category,omitempty"`
	Version     string  `json:"version,omitempty"`
	Author      *Author `json:"author,omitempty"`
}

// Author represents author information for a marketplace or plugin
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// ParseMarketplace parses a Claude marketplace.json file and validates required fields
func ParseMarketplace(filePath string) (*MarketplaceConfig, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read marketplace file: %w", err)
	}

	// Parse JSON
	var config MarketplaceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse marketplace JSON: %w", err)
	}

	// Validate required fields
	if err := validateMarketplaceConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateMarketplaceConfig validates that all required fields are present
func validateMarketplaceConfig(config *MarketplaceConfig) error {
	if config.Name == "" {
		return fmt.Errorf("marketplace name is required")
	}

	if config.Description == "" {
		return fmt.Errorf("marketplace description is required")
	}

	// Validate each plugin
	for i, plugin := range config.Plugins {
		if err := validatePlugin(&plugin, i); err != nil {
			return err
		}
	}

	return nil
}

// validatePlugin validates that a plugin has all required fields
func validatePlugin(plugin *Plugin, index int) error {
	if plugin.Name == "" {
		return fmt.Errorf("plugin at index %d: name is required", index)
	}

	if plugin.Description == "" {
		return fmt.Errorf("plugin at index %d: description is required", index)
	}

	if plugin.Source == "" {
		return fmt.Errorf("plugin at index %d: source is required", index)
	}

	return nil
}
