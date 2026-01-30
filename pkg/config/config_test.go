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
	_, err = Load(tmpDir)

	// Should return error when no config file exists
	if err == nil {
		t.Fatalf("Load() expected error when no config file exists, got nil")
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

func TestLoad_MissingTargets(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
	}{
		{
			name:       "old format: claude (no auto-migration)",
			configYAML: "default-tool: claude\n",
		},
		{
			name:       "old format: opencode (no auto-migration)",
			configYAML: "default-tool: opencode\n",
		},
		{
			name:       "empty config requires targets",
			configYAML: "",
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

			// Load config - should fail because install.targets is required
			_, err = Load(tmpDir)
			if err == nil {
				t.Errorf("Load() expected error for missing install.targets, got nil")
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
			name: "empty targets returns error",
			config: Config{
				Install: InstallConfig{
					Targets: []string{},
				},
			},
			wantTools: nil,
			wantError: true,
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

	// LoadGlobal should return error when no config exists
	_, err = LoadGlobal()

	// Should return error when no config file exists
	if err == nil {
		t.Fatalf("LoadGlobal() expected error when no config exists, got nil")
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
			errorMsg:  "must specify either 'url' or 'path'",
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
			errorMsg:  "sync.sources[1]: must specify either 'url' or 'path'",
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
			errorMsg: "must specify either 'url' or 'path'",
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

func TestValidate_RepoPath(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		wantError bool
		checkPath func(t *testing.T, path string)
	}{
		{
			name:      "empty path is valid",
			repoPath:  "",
			wantError: false,
		},
		{
			name:      "absolute path unchanged",
			repoPath:  "/home/user/custom-repo",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if path != "/home/user/custom-repo" {
					t.Errorf("path = %q, want /home/user/custom-repo", path)
				}
			},
		},
		{
			name:      "tilde expansion",
			repoPath:  "~/my-repo",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if contains(path, "~") {
					t.Errorf("path still contains ~: %q", path)
				}
				if !filepath.IsAbs(path) {
					t.Errorf("path not absolute after expansion: %q", path)
				}
			},
		},
		{
			name:      "relative path converted to absolute",
			repoPath:  "relative/path",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if !filepath.IsAbs(path) {
					t.Errorf("path not absolute: %q", path)
				}
			},
		},
		{
			name:      "path with dot cleaning",
			repoPath:  "/home/user/./repo",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if contains(path, "/.") {
					t.Errorf("path still contains /.: %q", path)
				}
				if path != "/home/user/repo" {
					t.Errorf("path = %q, want /home/user/repo", path)
				}
			},
		},
		{
			name:      "path with double-dot cleaning",
			repoPath:  "/home/user/foo/../repo",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if contains(path, "..") {
					t.Errorf("path still contains ..: %q", path)
				}
				if path != "/home/user/repo" {
					t.Errorf("path = %q, want /home/user/repo", path)
				}
			},
		},
		{
			name:      "tilde with subdirectory",
			repoPath:  "~/projects/ai-repo",
			wantError: false,
			checkPath: func(t *testing.T, path string) {
				if contains(path, "~") {
					t.Errorf("path still contains ~: %q", path)
				}
				if !filepath.IsAbs(path) {
					t.Errorf("path not absolute after expansion: %q", path)
				}
				if !contains(path, "projects/ai-repo") && !contains(path, "projects\\ai-repo") {
					t.Errorf("path = %q, want to contain projects/ai-repo", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Repo:    RepoConfig{Path: tt.repoPath},
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && tt.checkPath != nil {
				tt.checkPath(t, cfg.Repo.Path)
			}
		})
	}
}

func TestLoad_WithRepoPath(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		checkPath func(t *testing.T, path string)
	}{
		{
			name:     "absolute path",
			repoPath: "/custom/repo/path",
			checkPath: func(t *testing.T, path string) {
				if path != "/custom/repo/path" {
					t.Errorf("Repo.Path = %q, want /custom/repo/path", path)
				}
			},
		},
		{
			name:     "tilde expansion",
			repoPath: "~/custom-repo",
			checkPath: func(t *testing.T, path string) {
				if contains(path, "~") {
					t.Errorf("Repo.Path still contains ~: %q", path)
				}
				if !filepath.IsAbs(path) {
					t.Errorf("Repo.Path not absolute: %q", path)
				}
			},
		},
		{
			name:     "relative path",
			repoPath: "my-repo",
			checkPath: func(t *testing.T, path string) {
				if !filepath.IsAbs(path) {
					t.Errorf("Repo.Path not absolute: %q", path)
				}
			},
		},
		{
			name:     "path with dots",
			repoPath: "/tmp/../home/user/repo",
			checkPath: func(t *testing.T, path string) {
				if contains(path, "..") {
					t.Errorf("Repo.Path still contains ..: %q", path)
				}
				if path != "/home/user/repo" {
					t.Errorf("Repo.Path = %q, want /home/user/repo", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			configYAML := "install:\n  targets: [claude]\nrepo:\n  path: " + tt.repoPath + "\n"

			configPath := filepath.Join(tmpDir, DefaultConfigFileName)
			if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			cfg, err := Load(tmpDir)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if tt.checkPath != nil {
				tt.checkPath(t, cfg.Repo.Path)
			}
		})
	}
}

func TestLoad_WithEmptyRepoPath(t *testing.T) {
	tmpDir := t.TempDir()

	configYAML := `install:
  targets: [claude]
repo:
  path: ""
`

	configPath := filepath.Join(tmpDir, DefaultConfigFileName)
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Empty path should remain empty
	if cfg.Repo.Path != "" {
		t.Errorf("Repo.Path = %q, want empty string", cfg.Repo.Path)
	}
}

func TestLoad_WithoutRepoSection(t *testing.T) {
	tmpDir := t.TempDir()

	configYAML := `install:
  targets: [claude]
`

	configPath := filepath.Join(tmpDir, DefaultConfigFileName)
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Repo path should be empty when section is omitted
	if cfg.Repo.Path != "" {
		t.Errorf("Repo.Path = %q, want empty string", cfg.Repo.Path)
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "path: ${HOME}/repo",
			envVars:  map[string]string{"HOME": "/home/user"},
			expected: "path: /home/user/repo",
		},
		{
			name:     "variable with default - var set",
			input:    "path: ${REPO_PATH:-/default/path}",
			envVars:  map[string]string{"REPO_PATH": "/custom/path"},
			expected: "path: /custom/path",
		},
		{
			name:     "variable with default - var unset",
			input:    "path: ${REPO_PATH:-/default/path}",
			envVars:  map[string]string{},
			expected: "path: /default/path",
		},
		{
			name:     "variable with default - var empty",
			input:    "path: ${REPO_PATH:-/default/path}",
			envVars:  map[string]string{"REPO_PATH": ""},
			expected: "path: /default/path",
		},
		{
			name:     "simple variable - unset becomes empty",
			input:    "path: ${MISSING_VAR}",
			envVars:  map[string]string{},
			expected: "path: ",
		},
		{
			name:     "multiple variables",
			input:    "url: ${PROTOCOL}://${HOST}:${PORT}",
			envVars:  map[string]string{"PROTOCOL": "https", "HOST": "example.com", "PORT": "8080"},
			expected: "url: https://example.com:8080",
		},
		{
			name:     "mixed with and without defaults",
			input:    "${VAR1}:${VAR2:-default2}:${VAR3}",
			envVars:  map[string]string{"VAR1": "value1", "VAR3": "value3"},
			expected: "value1:default2:value3",
		},
		{
			name:     "no variables",
			input:    "plain text without variables",
			envVars:  map[string]string{},
			expected: "plain text without variables",
		},
		{
			name:     "variable name with underscore",
			input:    "${MY_VAR_NAME}",
			envVars:  map[string]string{"MY_VAR_NAME": "test"},
			expected: "test",
		},
		{
			name:     "variable name with numbers",
			input:    "${VAR_123}",
			envVars:  map[string]string{"VAR_123": "test"},
			expected: "test",
		},
		{
			name:     "default with special characters",
			input:    "${VAR:-/path/to/default}",
			envVars:  map[string]string{},
			expected: "/path/to/default",
		},
		{
			name:     "default with spaces",
			input:    "${VAR:-default value}",
			envVars:  map[string]string{},
			expected: "default value",
		},
		{
			name:     "tilde in default",
			input:    "${REPO_PATH:-~/.local/share/repo}",
			envVars:  map[string]string{},
			expected: "~/.local/share/repo",
		},
		{
			name:     "empty default",
			input:    "${VAR:-}",
			envVars:  map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Clear any vars not in the map
			// (t.Setenv handles cleanup automatically)

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLoad_WithEnvVarExpansion(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		envVars      map[string]string
		wantRepoPath string
		wantError    bool
	}{
		{
			name: "expand env var in repo path",
			configYAML: `install:
  targets: [claude]
repo:
  path: ${CUSTOM_REPO}`,
			envVars:      map[string]string{"CUSTOM_REPO": "/tmp/test-repo"},
			wantRepoPath: "/tmp/test-repo",
			wantError:    false,
		},
		{
			name: "expand env var with default - var set",
			configYAML: `install:
  targets: [claude]
repo:
  path: ${CUSTOM_REPO:-/default/repo}`,
			envVars:      map[string]string{"CUSTOM_REPO": "/actual/repo"},
			wantRepoPath: "/actual/repo",
			wantError:    false,
		},
		{
			name: "expand env var with default - var unset",
			configYAML: `install:
  targets: [claude]
repo:
  path: ${CUSTOM_REPO:-/default/repo}`,
			envVars:      map[string]string{},
			wantRepoPath: "/default/repo",
			wantError:    false,
		},
		{
			name: "expand multiple env vars in sync source",
			configYAML: `install:
  targets: [claude]
sync:
  sources:
    - url: ${SYNC_PROTO:-https}://${SYNC_HOST}/repo
      filter: ${SYNC_FILTER:-skill/*}`,
			envVars: map[string]string{
				"SYNC_HOST": "github.com",
			},
			wantError: false,
		},
		{
			name: "no env vars - static config",
			configYAML: `install:
  targets: [claude]
repo:
  path: /static/path`,
			envVars:      map[string]string{},
			wantRepoPath: "/static/path",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Write config file
			configPath := filepath.Join(tmpDir, DefaultConfigFileName)
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			// Load config
			cfg, err := Load(tmpDir)
			if (err != nil) != tt.wantError {
				t.Errorf("Load() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			// Verify repo path if specified
			if tt.wantRepoPath != "" {
				if cfg.Repo.Path != tt.wantRepoPath {
					t.Errorf("Repo.Path = %q, want %q", cfg.Repo.Path, tt.wantRepoPath)
				}
			}

			// Verify sync sources if present
			if len(cfg.Sync.Sources) > 0 {
				source := cfg.Sync.Sources[0]
				if contains(tt.configYAML, "SYNC_HOST") {
					if !contains(source.URL, "github.com") {
						t.Errorf("URL = %q, expected to contain 'github.com'", source.URL)
					}
					if !contains(source.URL, "https://") {
						t.Errorf("URL = %q, expected to contain 'https://'", source.URL)
					}
				}
				if source.Filter != "" && source.Filter != "skill/*" {
					t.Errorf("Filter = %q, want 'skill/*'", source.Filter)
				}
			}
		})
	}
}

func TestLoadGlobal_WithEnvVarExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	xdgConfigDir := filepath.Join(tmpDir, "config")

	t.Setenv("XDG_CONFIG_HOME", xdgConfigDir)
	t.Setenv("TEST_REPO_PATH", "/test/repo/path")

	// Create config directory
	configDir := filepath.Join(xdgConfigDir, "aimgr")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configYAML := `install:
  targets: [claude]
repo:
  path: ${TEST_REPO_PATH:-/default/path}`

	configPath := filepath.Join(configDir, DefaultConfigFileName)
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Reload XDG paths
	xdg.Reload()

	// Load global config
	cfg, err := LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}

	// Verify expansion worked
	if cfg.Repo.Path != "/test/repo/path" {
		t.Errorf("Repo.Path = %q, want /test/repo/path", cfg.Repo.Path)
	}
}

func TestSyncSourceValidate(t *testing.T) {
	// Create a temp directory for path validation tests
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid-dir")
	if err := os.MkdirAll(validPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name      string
		source    SyncSource
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid: url only",
			source: SyncSource{
				URL: "https://github.com/owner/repo",
			},
			wantError: false,
		},
		{
			name: "valid: path only",
			source: SyncSource{
				Path: validPath,
			},
			wantError: false,
		},
		{
			name: "valid: http:// url",
			source: SyncSource{
				URL: "http://github.com/owner/repo",
			},
			wantError: false,
		},
		{
			name: "valid: git@ url",
			source: SyncSource{
				URL: "git@github.com:owner/repo.git",
			},
			wantError: false,
		},
		{
			name: "valid: git:// url",
			source: SyncSource{
				URL: "git://github.com/owner/repo.git",
			},
			wantError: false,
		},
		{
			name: "valid: gh: shorthand",
			source: SyncSource{
				URL: "gh:owner/repo",
			},
			wantError: false,
		},
		{
			name: "invalid: both url and path",
			source: SyncSource{
				URL:  "https://github.com/owner/repo",
				Path: validPath,
			},
			wantError: true,
			errorMsg:  "cannot specify both 'url' and 'path'",
		},
		{
			name:      "invalid: neither url nor path",
			source:    SyncSource{},
			wantError: true,
			errorMsg:  "must specify either 'url' or 'path'",
		},
		{
			name: "invalid: url format (no protocol)",
			source: SyncSource{
				URL: "github.com/owner/repo",
			},
			wantError: true,
			errorMsg:  "url must start with",
		},
		{
			name: "invalid: url format (ftp://)",
			source: SyncSource{
				URL: "ftp://github.com/owner/repo",
			},
			wantError: true,
			errorMsg:  "url must start with",
		},
		{
			name: "valid: path that doesn't exist (format is valid)",
			source: SyncSource{
				Path: "/nonexistent/path/that/does/not/exist",
			},
			wantError: false,
		},
		{
			name: "valid: relative path that exists",
			source: SyncSource{
				Path: ".", // Current directory always exists in tests
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
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

func TestValidate_SyncSourceUrlAndPath(t *testing.T) {
	// Create a temp directory for path validation tests
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid-dir")
	if err := os.MkdirAll(validPath, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid: url source",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "https://github.com/owner/repo"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: path source",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{Path: validPath},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid: multiple sources with mixed url and path",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "https://github.com/owner/repo"},
						{Path: validPath},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid: both url and path in same source",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{
							URL:  "https://github.com/owner/repo",
							Path: validPath,
						},
					},
				},
			},
			wantError: true,
			errorMsg:  "cannot specify both 'url' and 'path'",
		},
		{
			name: "invalid: neither url nor path",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{Filter: "skill/*"}, // Only filter, no url or path
					},
				},
			},
			wantError: true,
			errorMsg:  "must specify either 'url' or 'path'",
		},
		{
			name: "invalid: non-url string in url field",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "not-a-valid-url"},
					},
				},
			},
			wantError: true,
			errorMsg:  "url must start with",
		},
		{
			name: "valid: non-existent path (format is valid)",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{Path: "/this/path/does/not/exist/at/all"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid: error in second source",
			config: Config{
				Install: InstallConfig{Targets: []string{"claude"}},
				Sync: SyncConfig{
					Sources: []SyncSource{
						{URL: "https://github.com/owner/repo1"},
						{URL: "invalid-url"},
					},
				},
			},
			wantError: true,
			errorMsg:  "sync.sources[1]:",
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
