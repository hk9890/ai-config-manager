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

func TestValidate_SyncConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string // Expected error message substring
	}{
		{
			name: "valid: single source without filter",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: single source with filter",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "https://github.com/org/repo", Filter: "skill/*"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: multiple sources",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo"},
						{URL: "https://github.com/org/repo"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: multiple sources with filters",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/*"},
						{URL: "https://github.com/org/repo", Filter: "command/test*"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: mix of filtered and unfiltered sources",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/*"},
						{URL: "https://github.com/org/repo"},
						{URL: "gh:another/repo", Filter: "agent/*"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: complex filter patterns",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/pdf*"},
						{URL: "https://github.com/org/repo", Filter: "command/*test*"},
						{URL: "gh:another/repo", Filter: "*agent*"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid: empty URL",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: ""},
					},
				},
			},
			wantError: true,
			errorMsg:  "url cannot be empty",
		},
		{
			name: "invalid: empty URL in second source",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo"},
						{URL: ""},
					},
				},
			},
			wantError: true,
			errorMsg:  "sync.sources[1]: url cannot be empty",
		},
		{
			name: "invalid: malformed filter pattern (unmatched bracket)",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/[abc"},
					},
				},
			},
			wantError: true,
			errorMsg:  "invalid filter pattern",
		},
		{
			name: "invalid: malformed filter pattern (incomplete range)",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/[a-"},
					},
				},
			},
			wantError: true,
			errorMsg:  "invalid filter pattern",
		},
		{
			name: "invalid: bad pattern in second source",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "gh:owner/repo", Filter: "skill/*"},
						{URL: "https://github.com/org/repo", Filter: "[invalid"},
					},
				},
			},
			wantError: true,
			errorMsg:  "sync.sources[1]: invalid filter pattern",
		},
		{
			name: "valid: empty sync sources (no sync configured)",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sync: SyncConfig{
					Sources: []SyncSource{},
				},
			},
			wantError: false,
		},
		{
			name: "valid: no sync section (zero value)",
			config: Config{
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				// Check that error message contains expected substring
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLoad_ValidSyncConfig(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		wantSources  int
		validateFunc func(*testing.T, *Config)
	}{
		{
			name: "single sync source without filter",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: gh:owner/repo
`,
			wantSources: 1,
			validateFunc: func(t *testing.T, cfg *Config) {
				if cfg.Sync.Sources[0].URL != "gh:owner/repo" {
					t.Errorf("source[0].URL = %v, want 'gh:owner/repo'", cfg.Sync.Sources[0].URL)
				}
				if cfg.Sync.Sources[0].Filter != "" {
					t.Errorf("source[0].Filter = %v, want empty", cfg.Sync.Sources[0].Filter)
				}
			},
		},
		{
			name: "single sync source with filter",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: https://github.com/org/repo
      filter: skill/*
`,
			wantSources: 1,
			validateFunc: func(t *testing.T, cfg *Config) {
				if cfg.Sync.Sources[0].URL != "https://github.com/org/repo" {
					t.Errorf("source[0].URL = %v, want 'https://github.com/org/repo'", cfg.Sync.Sources[0].URL)
				}
				if cfg.Sync.Sources[0].Filter != "skill/*" {
					t.Errorf("source[0].Filter = %v, want 'skill/*'", cfg.Sync.Sources[0].Filter)
				}
			},
		},
		{
			name: "multiple sync sources",
			configYAML: `install:
  targets: [claude, opencode]
sync:
  sources:
    - url: gh:owner/repo1
    - url: gh:owner/repo2
      filter: command/*
    - url: https://github.com/org/repo3
      filter: skill/pdf*
`,
			wantSources: 3,
			validateFunc: func(t *testing.T, cfg *Config) {
				if cfg.Sync.Sources[0].URL != "gh:owner/repo1" {
					t.Errorf("source[0].URL = %v, want 'gh:owner/repo1'", cfg.Sync.Sources[0].URL)
				}
				if cfg.Sync.Sources[0].Filter != "" {
					t.Errorf("source[0].Filter = %v, want empty", cfg.Sync.Sources[0].Filter)
				}
				if cfg.Sync.Sources[1].URL != "gh:owner/repo2" {
					t.Errorf("source[1].URL = %v, want 'gh:owner/repo2'", cfg.Sync.Sources[1].URL)
				}
				if cfg.Sync.Sources[1].Filter != "command/*" {
					t.Errorf("source[1].Filter = %v, want 'command/*'", cfg.Sync.Sources[1].Filter)
				}
				if cfg.Sync.Sources[2].URL != "https://github.com/org/repo3" {
					t.Errorf("source[2].URL = %v, want 'https://github.com/org/repo3'", cfg.Sync.Sources[2].URL)
				}
				if cfg.Sync.Sources[2].Filter != "skill/pdf*" {
					t.Errorf("source[2].Filter = %v, want 'skill/pdf*'", cfg.Sync.Sources[2].Filter)
				}
			},
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

			// Check number of sources
			if len(config.Sync.Sources) != tt.wantSources {
				t.Errorf("len(Sync.Sources) = %v, want %v", len(config.Sync.Sources), tt.wantSources)
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, config)
			}
		})
	}
}

func TestLoad_InvalidSyncConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		errorMsg   string
	}{
		{
			name: "empty source URL",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: ""
`,
			errorMsg: "url cannot be empty",
		},
		{
			name: "invalid filter pattern",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: gh:owner/repo
      filter: "[invalid"
`,
			errorMsg: "invalid filter pattern",
		},
		{
			name: "malformed filter with incomplete range",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: gh:owner/repo
      filter: "[a-"
`,
			errorMsg: "invalid filter pattern",
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
				t.Errorf("Load() expected error, got nil")
				return
			}

			// Check that error message contains expected substring
			if !contains(err.Error(), tt.errorMsg) {
				t.Errorf("Load() error = %v, want error containing %q", err, tt.errorMsg)
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
