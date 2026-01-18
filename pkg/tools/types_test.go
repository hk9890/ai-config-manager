package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTool_String(t *testing.T) {
	tests := []struct {
		tool     Tool
		expected string
	}{
		{Claude, "claude"},
		{OpenCode, "opencode"},
		{Copilot, "copilot"},
		{Tool(-1), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tool.String(); got != tt.expected {
				t.Errorf("Tool.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTool(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Tool
		wantError bool
	}{
		{
			name:      "claude lowercase",
			input:     "claude",
			want:      Claude,
			wantError: false,
		},
		{
			name:      "claude uppercase",
			input:     "CLAUDE",
			want:      Claude,
			wantError: false,
		},
		{
			name:      "opencode lowercase",
			input:     "opencode",
			want:      OpenCode,
			wantError: false,
		},
		{
			name:      "opencode mixed case",
			input:     "OpenCode",
			want:      OpenCode,
			wantError: false,
		},
		{
			name:      "copilot lowercase",
			input:     "copilot",
			want:      Copilot,
			wantError: false,
		},
		{
			name:      "copilot uppercase",
			input:     "COPILOT",
			want:      Copilot,
			wantError: false,
		},
		{
			name:      "invalid tool",
			input:     "invalid",
			want:      -1,
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			want:      -1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTool(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseTool(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseTool(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetToolInfo(t *testing.T) {
	tests := []struct {
		name             string
		tool             Tool
		wantName         string
		wantCommandsDir  string
		wantSkillsDir    string
		wantSupportsCmd  bool
		wantSupportsSkil bool
	}{
		{
			name:             "Claude",
			tool:             Claude,
			wantName:         "Claude Code",
			wantCommandsDir:  ".claude/commands",
			wantSkillsDir:    ".claude/skills",
			wantSupportsCmd:  true,
			wantSupportsSkil: true,
		},
		{
			name:             "OpenCode",
			tool:             OpenCode,
			wantName:         "OpenCode",
			wantCommandsDir:  ".opencode/commands",
			wantSkillsDir:    ".opencode/skills",
			wantSupportsCmd:  true,
			wantSupportsSkil: true,
		},
		{
			name:             "Copilot",
			tool:             Copilot,
			wantName:         "GitHub Copilot",
			wantCommandsDir:  "",
			wantSkillsDir:    ".github/skills",
			wantSupportsCmd:  false,
			wantSupportsSkil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetToolInfo(tt.tool)
			if info.Name != tt.wantName {
				t.Errorf("GetToolInfo(%v).Name = %v, want %v", tt.tool, info.Name, tt.wantName)
			}
			if info.CommandsDir != tt.wantCommandsDir {
				t.Errorf("GetToolInfo(%v).CommandsDir = %v, want %v", tt.tool, info.CommandsDir, tt.wantCommandsDir)
			}
			if info.SkillsDir != tt.wantSkillsDir {
				t.Errorf("GetToolInfo(%v).SkillsDir = %v, want %v", tt.tool, info.SkillsDir, tt.wantSkillsDir)
			}
			if info.SupportsCommands != tt.wantSupportsCmd {
				t.Errorf("GetToolInfo(%v).SupportsCommands = %v, want %v", tt.tool, info.SupportsCommands, tt.wantSupportsCmd)
			}
			if info.SupportsSkills != tt.wantSupportsSkil {
				t.Errorf("GetToolInfo(%v).SupportsSkills = %v, want %v", tt.tool, info.SupportsSkills, tt.wantSupportsSkil)
			}
		})
	}
}

func TestDetectExistingTools(t *testing.T) {
	tests := []struct {
		name          string
		setupDirs     []string
		expectedTools []Tool
	}{
		{
			name:          "no tool directories",
			setupDirs:     []string{},
			expectedTools: []Tool{},
		},
		{
			name:          "only Claude",
			setupDirs:     []string{".claude"},
			expectedTools: []Tool{Claude},
		},
		{
			name:          "only OpenCode",
			setupDirs:     []string{".opencode"},
			expectedTools: []Tool{OpenCode},
		},
		{
			name:          "only Copilot",
			setupDirs:     []string{".github/skills"},
			expectedTools: []Tool{Copilot},
		},
		{
			name:          "Claude and OpenCode",
			setupDirs:     []string{".claude", ".opencode"},
			expectedTools: []Tool{Claude, OpenCode},
		},
		{
			name:          "all tools",
			setupDirs:     []string{".claude", ".opencode", ".github/skills"},
			expectedTools: []Tool{Claude, OpenCode, Copilot},
		},
		{
			name:          ".github exists but not skills",
			setupDirs:     []string{".github/workflows"},
			expectedTools: []Tool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "tools-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Setup directories
			for _, dir := range tt.setupDirs {
				dirPath := filepath.Join(tmpDir, dir)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					t.Fatalf("failed to create directory %s: %v", dirPath, err)
				}
			}

			// Detect tools
			detected, err := DetectExistingTools(tmpDir)
			if err != nil {
				t.Fatalf("DetectExistingTools() error = %v", err)
			}

			// Check results
			if len(detected) != len(tt.expectedTools) {
				t.Errorf("DetectExistingTools() found %d tools, want %d", len(detected), len(tt.expectedTools))
				t.Errorf("Got: %v, Want: %v", detected, tt.expectedTools)
				return
			}

			// Check each expected tool is present
			for _, expected := range tt.expectedTools {
				found := false
				for _, detected := range detected {
					if detected == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tool %v not found in detected tools %v", expected, detected)
				}
			}
		})
	}
}

func TestDetectExistingTools_NonexistentPath(t *testing.T) {
	// Test with a path that doesn't exist
	_, err := DetectExistingTools("/nonexistent/path/that/should/not/exist")
	// Should return error or empty list, depending on implementation
	// Current implementation returns error for permission issues but not for nonexistent paths
	if err != nil {
		// This is acceptable
		t.Logf("Got error for nonexistent path: %v", err)
	}
}

func TestAllTools(t *testing.T) {
	tools := AllTools()
	if len(tools) != 3 {
		t.Errorf("AllTools() returned %d tools, want 3", len(tools))
	}

	// Check that all expected tools are present
	expectedTools := []Tool{Claude, OpenCode, Copilot}
	for _, expected := range expectedTools {
		found := false
		for _, tool := range tools {
			if tool == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllTools() missing %v", expected)
		}
	}
}
