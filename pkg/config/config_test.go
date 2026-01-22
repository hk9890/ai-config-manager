package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

func TestLoad_NoConfigFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Load config from directory without config file
	config, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Should return default config
	if len(config.Install.Targets) != 1 || config.Install.Targets[0] != DefaultToolValue {
		t.Errorf("Load().Install.Targets = %v, want [%v]", config.Install.Targets, DefaultToolValue)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tests := []struct {
		name          string
		configYAML    string
		wantTargets   []string
		wantToolTypes []tools.Tool
	}{
		{
			name:          "new format: single target",
			configYAML:    "install:\n  targets: [claude]\n",
			wantTargets:   []string{"claude"},
			wantToolTypes: []tools.Tool{tools.Claude},
		},
		{
			name:          "new format: multiple targets",
			configYAML:    "install:\n  targets: [claude, opencode]\n",
			wantTargets:   []string{"claude", "opencode"},
			wantToolTypes: []tools.Tool{tools.Claude, tools.OpenCode},
		},
		{
			name:          "old format: claude (migrated)",
			configYAML:    "default-tool: claude\n",
			wantTargets:   []string{"claude"},
			wantToolTypes: []tools.Tool{tools.Claude},
		},
		{
			name:          "old format: opencode (migrated)",
			configYAML:    "default-tool: opencode\n",
			wantTargets:   []string{"opencode"},
			wantToolTypes: []tools.Tool{tools.OpenCode},
		},
		{
			name:          "empty defaults to claude",
			configYAML:    "",
			wantTargets:   []string{"claude"},
			wantToolTypes: []tools.Tool{tools.Claude},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "config-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Write config file
			configPath := filepath.Join(tmpDir, DefaultConfigFileName)
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Load config
			config, err := Load(tmpDir)
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}

			// Check install targets
			if len(config.Install.Targets) != len(tt.wantTargets) {
				t.Errorf("Load().Install.Targets length = %v, want %v", len(config.Install.Targets), len(tt.wantTargets))
			}
			for i, target := range tt.wantTargets {
				if i >= len(config.Install.Targets) || config.Install.Targets[i] != target {
					t.Errorf("Load().Install.Targets[%d] = %v, want %v", i, config.Install.Targets, tt.wantTargets)
					break
				}
			}

			// Check GetDefaultTargets
			toolTypes, err := config.GetDefaultTargets()
			if err != nil {
				t.Fatalf("GetDefaultTargets() error = %v, want nil", err)
			}
			if len(toolTypes) != len(tt.wantToolTypes) {
				t.Errorf("GetDefaultTargets() length = %v, want %v", len(toolTypes), len(tt.wantToolTypes))
			}
			for i, wantType := range tt.wantToolTypes {
				if i >= len(toolTypes) || toolTypes[i] != wantType {
					t.Errorf("GetDefaultTargets()[%d] = %v, want %v", i, toolTypes[i], wantType)
				}
			}
		})
	}
}

func TestLoad_InvalidTool(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
	}{
		{
			name:       "invalid tool name in new format",
			configYAML: "install:\n  targets: [invalid]\n",
		},
		{
			name:       "unsupported tool in new format",
			configYAML: "install:\n  targets: [cursor]\n",
		},
		{
			name:       "invalid tool name in old format",
			configYAML: "default-tool: invalid\n",
		},
		{
			name:       "unsupported tool in old format",
			configYAML: "default-tool: cursor\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "config-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Write config file
			configPath := filepath.Join(tmpDir, DefaultConfigFileName)
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Load config - should fail validation
			_, err = Load(tmpDir)
			if err == nil {
				t.Errorf("Load() expected error for invalid tool, got nil")
			}
		})
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write malformed YAML
	configPath := filepath.Join(tmpDir, DefaultConfigFileName)
	malformedYAML := "default-tool: [invalid yaml\n"
	if err := os.WriteFile(configPath, []byte(malformedYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config - should fail parsing
	_, err = Load(tmpDir)
	if err == nil {
		t.Errorf("Load() expected error for malformed YAML, got nil")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid single target",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
			},
			wantError: false,
		},
		{
			name: "valid multiple targets",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude", "opencode", "copilot"},
				},
			},
			wantError: false,
		},
		{
			name: "empty is valid (uses default)",
			config: Config{
				Install: InstallConfig{
					Targets: []string{},
				},
			},
			wantError: false,
		},
		{
			name: "invalid tool in targets",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"invalid"},
				},
			},
			wantError: true,
		},
		{
			name: "mixed valid and invalid tools",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude", "invalid"},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError && err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestGetDefaultTargets(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantTools []tools.Tool
		wantError bool
	}{
		{
			name: "single target",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
			},
			wantTools: []tools.Tool{tools.Claude},
			wantError: false,
		},
		{
			name: "multiple targets",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude", "opencode"},
				},
			},
			wantTools: []tools.Tool{tools.Claude, tools.OpenCode},
			wantError: false,
		},
		{
			name: "all three tools",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude", "opencode", "copilot"},
				},
			},
			wantTools: []tools.Tool{tools.Claude, tools.OpenCode, tools.Copilot},
			wantError: false,
		},
		{
			name: "empty defaults to claude",
			config: Config{
				Install: InstallConfig{
					Targets: []string{},
				},
			},
			wantTools: []tools.Tool{tools.Claude},
			wantError: false,
		},
		{
			name: "invalid tool",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"invalid"},
				},
			},
			wantTools: nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolsList, err := tt.config.GetDefaultTargets()
			if tt.wantError {
				if err == nil {
					t.Errorf("GetDefaultTargets() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetDefaultTargets() unexpected error: %v", err)
			}
			if len(toolsList) != len(tt.wantTools) {
				t.Errorf("GetDefaultTargets() length = %v, want %v", len(toolsList), len(tt.wantTools))
				return
			}
			for i, wantTool := range tt.wantTools {
				if toolsList[i] != wantTool {
					t.Errorf("GetDefaultTargets()[%d] = %v, want %v", i, toolsList[i], wantTool)
				}
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Test that GetConfigPath returns a valid path
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v, want nil", err)
	}

	// Path should end with aimgr/aimgr.yaml
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() = %v, want absolute path", path)
	}

	// Path should contain .config/aimgr/aimgr.yaml
	if !filepath.IsAbs(path) || filepath.Base(path) != DefaultConfigFileName {
		t.Errorf("GetConfigPath() = %v, want path ending with %v", path, DefaultConfigFileName)
	}
}

func TestMigrateConfig(t *testing.T) {
	// Create temp directory for old config
	oldDir := t.TempDir()
	oldPath := filepath.Join(oldDir, OldConfigFileName)

	// Create temp directory for new config
	newDir := t.TempDir()
	newPath := filepath.Join(newDir, "aimgr", DefaultConfigFileName)

	// Write old config
	oldConfigData := []byte("default-tool: opencode\n")
	if err := os.WriteFile(oldPath, oldConfigData, 0644); err != nil {
		t.Fatalf("failed to write old config: %v", err)
	}

	// Migrate
	if err := migrateConfig(oldPath, newPath); err != nil {
		t.Fatalf("migrateConfig() error = %v, want nil", err)
	}

	// Check new config exists
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("new config file does not exist at %v", newPath)
	}

	// Check old config still exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		t.Errorf("old config file should still exist at %v", oldPath)
	}

	// Verify content
	newData, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("failed to read new config: %v", err)
	}
	if string(newData) != string(oldConfigData) {
		t.Errorf("new config content = %v, want %v", string(newData), string(oldConfigData))
	}
}

func TestLoadGlobal_NoConfig(t *testing.T) {
	// Create temp directory for isolated test environment
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override XDG_CONFIG_HOME to use temp directory
	// This isolates the test from the actual user's config
	oldXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfigHome)
		xdg.Reload() // Restore original paths
	}()

	// Also override HOME to prevent fallback to ~/.ai-repo.yaml
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Reload XDG paths to pick up new environment variables
	xdg.Reload()

	// LoadGlobal should return defaults when no config exists
	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v, want nil", err)
	}

	// Should return default config
	if len(cfg.Install.Targets) != 1 || cfg.Install.Targets[0] != DefaultToolValue {
		t.Errorf("LoadGlobal().Install.Targets = %v, want [%v]", cfg.Install.Targets, DefaultToolValue)
	}
}
