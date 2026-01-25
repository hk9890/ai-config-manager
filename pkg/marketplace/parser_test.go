package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseMarketplace tests the ParseMarketplace function with various inputs
func TestParseMarketplace(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string // Returns filepath to marketplace.json
		wantError    bool
		errorMessage string
		validate     func(t *testing.T, config *MarketplaceConfig)
	}{
		{
			name: "valid marketplace with all fields",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Version:     "1.0.0",
					Description: "A test marketplace",
					Owner: &Author{
						Name:  "Test Owner",
						Email: "owner@example.com",
					},
					Plugins: []Plugin{
						{
							Name:        "test-plugin",
							Description: "A test plugin",
							Source:      "./plugins/test",
							Category:    "testing",
							Version:     "1.0.0",
							Author: &Author{
								Name:  "Plugin Author",
								Email: "author@example.com",
							},
						},
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError: false,
			validate: func(t *testing.T, config *MarketplaceConfig) {
				if config.Name != "test-marketplace" {
					t.Errorf("Name = %v, want test-marketplace", config.Name)
				}
				if config.Version != "1.0.0" {
					t.Errorf("Version = %v, want 1.0.0", config.Version)
				}
				if config.Description != "A test marketplace" {
					t.Errorf("Description = %v, want 'A test marketplace'", config.Description)
				}
				if config.Owner == nil {
					t.Fatal("Owner is nil")
				}
				if config.Owner.Name != "Test Owner" {
					t.Errorf("Owner.Name = %v, want 'Test Owner'", config.Owner.Name)
				}
				if config.Owner.Email != "owner@example.com" {
					t.Errorf("Owner.Email = %v, want owner@example.com", config.Owner.Email)
				}
				if len(config.Plugins) != 1 {
					t.Fatalf("Plugins length = %v, want 1", len(config.Plugins))
				}
				plugin := config.Plugins[0]
				if plugin.Name != "test-plugin" {
					t.Errorf("Plugin.Name = %v, want test-plugin", plugin.Name)
				}
				if plugin.Description != "A test plugin" {
					t.Errorf("Plugin.Description = %v, want 'A test plugin'", plugin.Description)
				}
				if plugin.Source != "./plugins/test" {
					t.Errorf("Plugin.Source = %v, want ./plugins/test", plugin.Source)
				}
				if plugin.Category != "testing" {
					t.Errorf("Plugin.Category = %v, want testing", plugin.Category)
				}
				if plugin.Version != "1.0.0" {
					t.Errorf("Plugin.Version = %v, want 1.0.0", plugin.Version)
				}
				if plugin.Author == nil {
					t.Fatal("Plugin.Author is nil")
				}
				if plugin.Author.Name != "Plugin Author" {
					t.Errorf("Plugin.Author.Name = %v, want 'Plugin Author'", plugin.Author.Name)
				}
				if plugin.Author.Email != "author@example.com" {
					t.Errorf("Plugin.Author.Email = %v, want author@example.com", plugin.Author.Email)
				}
			},
		},
		{
			name: "valid marketplace with required fields only",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "minimal-marketplace",
					Description: "Minimal test marketplace",
					Plugins: []Plugin{
						{
							Name:        "minimal-plugin",
							Description: "Minimal plugin",
							Source:      "./plugins/minimal",
						},
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError: false,
			validate: func(t *testing.T, config *MarketplaceConfig) {
				if config.Name != "minimal-marketplace" {
					t.Errorf("Name = %v, want minimal-marketplace", config.Name)
				}
				if config.Version != "" {
					t.Errorf("Version = %v, want empty string", config.Version)
				}
				if config.Owner != nil {
					t.Errorf("Owner = %v, want nil", config.Owner)
				}
				if len(config.Plugins) != 1 {
					t.Fatalf("Plugins length = %v, want 1", len(config.Plugins))
				}
				plugin := config.Plugins[0]
				if plugin.Category != "" {
					t.Errorf("Plugin.Category = %v, want empty string", plugin.Category)
				}
				if plugin.Version != "" {
					t.Errorf("Plugin.Version = %v, want empty string", plugin.Version)
				}
				if plugin.Author != nil {
					t.Errorf("Plugin.Author = %v, want nil", plugin.Author)
				}
			},
		},
		{
			name: "valid marketplace with multiple plugins",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "multi-plugin-marketplace",
					Description: "Marketplace with multiple plugins",
					Plugins: []Plugin{
						{
							Name:        "plugin1",
							Description: "First plugin",
							Source:      "./plugins/plugin1",
						},
						{
							Name:        "plugin2",
							Description: "Second plugin",
							Source:      "./plugins/plugin2",
						},
						{
							Name:        "plugin3",
							Description: "Third plugin",
							Source:      "./plugins/plugin3",
						},
					},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError: false,
			validate: func(t *testing.T, config *MarketplaceConfig) {
				if len(config.Plugins) != 3 {
					t.Errorf("Plugins length = %v, want 3", len(config.Plugins))
				}
			},
		},
		{
			name: "valid marketplace with empty plugins array",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "empty-plugins",
					Description: "Marketplace with no plugins",
					Plugins:     []Plugin{},
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError: false,
			validate: func(t *testing.T, config *MarketplaceConfig) {
				if len(config.Plugins) != 0 {
					t.Errorf("Plugins length = %v, want 0", len(config.Plugins))
				}
			},
		},
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/marketplace.json"
			},
			wantError:    true,
			errorMessage: "failed to read marketplace file",
		},
		{
			name: "invalid JSON",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				content := `{"name": "test", "description": "test", invalid json}`
				if err := os.WriteFile(marketplaceFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "failed to parse marketplace JSON",
		},
		{
			name: "missing name field",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				content := `{"description": "test marketplace", "plugins": []}`
				if err := os.WriteFile(marketplaceFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "marketplace name is required",
		},
		{
			name: "empty name field",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "",
					Description: "test marketplace",
					Plugins:     []Plugin{},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "marketplace name is required",
		},
		{
			name: "missing description field",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				content := `{"name": "test-marketplace", "plugins": []}`
				if err := os.WriteFile(marketplaceFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "marketplace description is required",
		},
		{
			name: "empty description field",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Description: "",
					Plugins:     []Plugin{},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "marketplace description is required",
		},
		{
			name: "plugin missing name",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Description: "Test",
					Plugins: []Plugin{
						{
							Name:        "",
							Description: "Plugin without name",
							Source:      "./plugins/test",
						},
					},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "plugin at index 0: name is required",
		},
		{
			name: "plugin missing description",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Description: "Test",
					Plugins: []Plugin{
						{
							Name:        "test-plugin",
							Description: "",
							Source:      "./plugins/test",
						},
					},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "plugin at index 0: description is required",
		},
		{
			name: "plugin missing source",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Description: "Test",
					Plugins: []Plugin{
						{
							Name:        "test-plugin",
							Description: "Plugin without source",
							Source:      "",
						},
					},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "plugin at index 0: source is required",
		},
		{
			name: "multiple plugins with validation error in second plugin",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				config := MarketplaceConfig{
					Name:        "test-marketplace",
					Description: "Test",
					Plugins: []Plugin{
						{
							Name:        "plugin1",
							Description: "First plugin",
							Source:      "./plugins/plugin1",
						},
						{
							Name:        "plugin2",
							Description: "", // Invalid
							Source:      "./plugins/plugin2",
						},
					},
				}
				data, _ := json.Marshal(config)
				if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "plugin at index 1: description is required",
		},
		{
			name: "empty file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				if err := os.WriteFile(marketplaceFile, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError:    true,
			errorMessage: "failed to parse marketplace JSON",
		},
		{
			name: "null plugins field",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				marketplaceFile := filepath.Join(tmpDir, "marketplace.json")
				content := `{"name": "test-marketplace", "description": "test", "plugins": null}`
				if err := os.WriteFile(marketplaceFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return marketplaceFile
			},
			wantError: false,
			validate: func(t *testing.T, config *MarketplaceConfig) {
				if config.Plugins != nil {
					t.Errorf("Plugins = %v, want nil", config.Plugins)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup(t)
			config, err := ParseMarketplace(filePath)

			// Check error expectation
			if (err != nil) != tt.wantError {
				t.Errorf("ParseMarketplace() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Check error message if error expected
			if err != nil && tt.errorMessage != "" {
				if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("ParseMarketplace() error message = %v, want to contain %v", err.Error(), tt.errorMessage)
				}
				return
			}

			// Run validation if provided and no error
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestValidateMarketplaceConfig tests the validateMarketplaceConfig function
func TestValidateMarketplaceConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       *MarketplaceConfig
		wantError    bool
		errorMessage string
	}{
		{
			name: "valid config",
			config: &MarketplaceConfig{
				Name:        "test",
				Description: "test description",
				Plugins: []Plugin{
					{
						Name:        "plugin1",
						Description: "plugin description",
						Source:      "./source",
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			config: &MarketplaceConfig{
				Name:        "",
				Description: "test description",
				Plugins:     []Plugin{},
			},
			wantError:    true,
			errorMessage: "marketplace name is required",
		},
		{
			name: "missing description",
			config: &MarketplaceConfig{
				Name:        "test",
				Description: "",
				Plugins:     []Plugin{},
			},
			wantError:    true,
			errorMessage: "marketplace description is required",
		},
		{
			name: "invalid plugin",
			config: &MarketplaceConfig{
				Name:        "test",
				Description: "test description",
				Plugins: []Plugin{
					{
						Name:        "",
						Description: "plugin description",
						Source:      "./source",
					},
				},
			},
			wantError:    true,
			errorMessage: "plugin at index 0: name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMarketplaceConfig(tt.config)

			if (err != nil) != tt.wantError {
				t.Errorf("validateMarketplaceConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil && tt.errorMessage != "" {
				if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("validateMarketplaceConfig() error message = %v, want to contain %v", err.Error(), tt.errorMessage)
				}
			}
		})
	}
}

// TestValidatePlugin tests the validatePlugin function
func TestValidatePlugin(t *testing.T) {
	tests := []struct {
		name         string
		plugin       *Plugin
		index        int
		wantError    bool
		errorMessage string
	}{
		{
			name: "valid plugin",
			plugin: &Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				Source:      "./plugins/test",
			},
			index:     0,
			wantError: false,
		},
		{
			name: "valid plugin with all fields",
			plugin: &Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				Source:      "./plugins/test",
				Category:    "testing",
				Version:     "1.0.0",
				Author: &Author{
					Name:  "Test Author",
					Email: "test@example.com",
				},
			},
			index:     0,
			wantError: false,
		},
		{
			name: "missing name",
			plugin: &Plugin{
				Name:        "",
				Description: "A test plugin",
				Source:      "./plugins/test",
			},
			index:        5,
			wantError:    true,
			errorMessage: "plugin at index 5: name is required",
		},
		{
			name: "missing description",
			plugin: &Plugin{
				Name:        "test-plugin",
				Description: "",
				Source:      "./plugins/test",
			},
			index:        3,
			wantError:    true,
			errorMessage: "plugin at index 3: description is required",
		},
		{
			name: "missing source",
			plugin: &Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				Source:      "",
			},
			index:        0,
			wantError:    true,
			errorMessage: "plugin at index 0: source is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePlugin(tt.plugin, tt.index)

			if (err != nil) != tt.wantError {
				t.Errorf("validatePlugin() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err != nil && tt.errorMessage != "" {
				if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("validatePlugin() error message = %v, want to contain %v", err.Error(), tt.errorMessage)
				}
			}
		})
	}
}

// TestMarketplaceRoundTrip tests parsing and re-marshaling a marketplace config
func TestMarketplaceRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	marketplaceFile := filepath.Join(tmpDir, "marketplace.json")

	original := MarketplaceConfig{
		Name:        "roundtrip-test",
		Version:     "2.0.0",
		Description: "Testing round trip parse/marshal",
		Owner: &Author{
			Name:  "Test Owner",
			Email: "owner@test.com",
		},
		Plugins: []Plugin{
			{
				Name:        "plugin1",
				Description: "First plugin",
				Source:      "./plugins/plugin1",
				Category:    "tools",
				Version:     "1.0.0",
				Author: &Author{
					Name:  "Plugin Author",
					Email: "author@test.com",
				},
			},
			{
				Name:        "plugin2",
				Description: "Second plugin",
				Source:      "./plugins/plugin2",
			},
		},
	}

	// Write original
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal original: %v", err)
	}
	if err := os.WriteFile(marketplaceFile, data, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Parse it
	parsed, err := ParseMarketplace(marketplaceFile)
	if err != nil {
		t.Fatalf("ParseMarketplace() error = %v", err)
	}

	// Compare
	if parsed.Name != original.Name {
		t.Errorf("Round trip: name = %v, want %v", parsed.Name, original.Name)
	}
	if parsed.Version != original.Version {
		t.Errorf("Round trip: version = %v, want %v", parsed.Version, original.Version)
	}
	if parsed.Description != original.Description {
		t.Errorf("Round trip: description = %v, want %v", parsed.Description, original.Description)
	}
	if parsed.Owner == nil || original.Owner == nil {
		t.Fatalf("Round trip: owner is nil")
	}
	if parsed.Owner.Name != original.Owner.Name {
		t.Errorf("Round trip: owner.Name = %v, want %v", parsed.Owner.Name, original.Owner.Name)
	}
	if parsed.Owner.Email != original.Owner.Email {
		t.Errorf("Round trip: owner.Email = %v, want %v", parsed.Owner.Email, original.Owner.Email)
	}
	if len(parsed.Plugins) != len(original.Plugins) {
		t.Fatalf("Round trip: plugins length = %v, want %v", len(parsed.Plugins), len(original.Plugins))
	}
	for i := range parsed.Plugins {
		if parsed.Plugins[i].Name != original.Plugins[i].Name {
			t.Errorf("Round trip: plugins[%d].Name = %v, want %v", i, parsed.Plugins[i].Name, original.Plugins[i].Name)
		}
		if parsed.Plugins[i].Description != original.Plugins[i].Description {
			t.Errorf("Round trip: plugins[%d].Description = %v, want %v", i, parsed.Plugins[i].Description, original.Plugins[i].Description)
		}
		if parsed.Plugins[i].Source != original.Plugins[i].Source {
			t.Errorf("Round trip: plugins[%d].Source = %v, want %v", i, parsed.Plugins[i].Source, original.Plugins[i].Source)
		}
	}
}
