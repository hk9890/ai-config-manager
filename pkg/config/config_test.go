package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hans-m-leitner/ai-config-manager/pkg/tools"
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
	if config.DefaultTool != DefaultToolValue {
		t.Errorf("Load().DefaultTool = %v, want %v", config.DefaultTool, DefaultToolValue)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		wantTool     string
		wantToolType tools.Tool
	}{
		{
			name:         "claude",
			configYAML:   "default-tool: claude\n",
			wantTool:     "claude",
			wantToolType: tools.Claude,
		},
		{
			name:         "opencode",
			configYAML:   "default-tool: opencode\n",
			wantTool:     "opencode",
			wantToolType: tools.OpenCode,
		},
		{
			name:         "copilot",
			configYAML:   "default-tool: copilot\n",
			wantTool:     "copilot",
			wantToolType: tools.Copilot,
		},
		{
			name:         "empty defaults to claude",
			configYAML:   "",
			wantTool:     DefaultToolValue,
			wantToolType: tools.Claude,
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

			// Check default tool
			if config.DefaultTool != tt.wantTool {
				t.Errorf("Load().DefaultTool = %v, want %v", config.DefaultTool, tt.wantTool)
			}

			// Check GetDefaultTool
			tool, err := config.GetDefaultTool()
			if err != nil {
				t.Fatalf("GetDefaultTool() error = %v, want nil", err)
			}
			if tool != tt.wantToolType {
				t.Errorf("GetDefaultTool() = %v, want %v", tool, tt.wantToolType)
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
			name:       "invalid tool name",
			configYAML: "default-tool: invalid\n",
		},
		{
			name:       "unsupported tool",
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
			name: "valid claude",
			config: Config{
				DefaultTool: "claude",
			},
			wantError: false,
		},
		{
			name: "valid opencode",
			config: Config{
				DefaultTool: "opencode",
			},
			wantError: false,
		},
		{
			name: "valid copilot",
			config: Config{
				DefaultTool: "copilot",
			},
			wantError: false,
		},
		{
			name: "empty is valid (uses default)",
			config: Config{
				DefaultTool: "",
			},
			wantError: false,
		},
		{
			name: "invalid tool",
			config: Config{
				DefaultTool: "invalid",
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

func TestGetDefaultTool(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantTool  tools.Tool
		wantError bool
	}{
		{
			name: "claude",
			config: Config{
				DefaultTool: "claude",
			},
			wantTool:  tools.Claude,
			wantError: false,
		},
		{
			name: "opencode",
			config: Config{
				DefaultTool: "opencode",
			},
			wantTool:  tools.OpenCode,
			wantError: false,
		},
		{
			name: "copilot",
			config: Config{
				DefaultTool: "copilot",
			},
			wantTool:  tools.Copilot,
			wantError: false,
		},
		{
			name: "empty defaults to claude",
			config: Config{
				DefaultTool: "",
			},
			wantTool:  tools.Claude,
			wantError: false,
		},
		{
			name: "invalid tool",
			config: Config{
				DefaultTool: "invalid",
			},
			wantTool:  -1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := tt.config.GetDefaultTool()
			if tt.wantError {
				if err == nil {
					t.Errorf("GetDefaultTool() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetDefaultTool() unexpected error: %v", err)
			}
			if tool != tt.wantTool {
				t.Errorf("GetDefaultTool() = %v, want %v", tool, tt.wantTool)
			}
		})
	}
}
