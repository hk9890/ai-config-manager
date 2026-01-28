package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCommand(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantError bool
		checkName string
		checkDesc string
	}{
		{
			name:      "valid command",
			filePath:  "testdata/commands/test-command.md",
			wantError: false,
			checkName: "test-command",
			checkDesc: "Run tests with coverage",
		},
		{
			name:      "command without frontmatter",
			filePath:  "testdata/commands/no-frontmatter.md",
			wantError: true,
		},
		{
			name:      "nonexistent file",
			filePath:  "testdata/commands/nonexistent.md",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := LoadCommand(tt.filePath)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if res.Name != tt.checkName {
				t.Errorf("Name mismatch. Got: %s, Want: %s", res.Name, tt.checkName)
			}

			if res.Description != tt.checkDesc {
				t.Errorf("Description mismatch. Got: %s, Want: %s", res.Description, tt.checkDesc)
			}
		})
	}
}

func TestLoadCommandWithBase_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	cmdPath := filepath.Join(tmpDir, "commands", "test.md")

	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	content := `---
description: Test command
---
# Test
`
	if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	basePath := filepath.Join(tmpDir, "commands")
	res, err := LoadCommandWithBase(cmdPath, basePath)
	if err != nil {
		t.Fatalf("LoadCommandWithBase failed: %v", err)
	}

	if res.Name != "test" {
		t.Errorf("Name mismatch. Got: %s, Want: test", res.Name)
	}
}

func TestLoadCommandResource(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		wantError bool
	}{
		{
			name:      "valid command",
			filePath:  "testdata/commands/test-command.md",
			wantError: false,
		},
		{
			name:      "nonexistent file",
			filePath:  "testdata/commands/nonexistent.md",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := LoadCommandResource(tt.filePath)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if res.Content == "" {
				t.Errorf("Content should not be empty")
			}
		})
	}
}

// TestLoadCommandWithBase tests the LoadCommandWithBase function
func TestLoadCommandWithBase(t *testing.T) {
	tests := []struct {
		name         string
		setupPath    string // Path structure to create
		basePath     string // Base path for relative calculation
		expectedName string // Expected Name field (nested path for nested commands)
	}{
		{
			name:         "nested command with base path",
			setupPath:    "commands/api/v2/deploy.md",
			basePath:     "commands",
			expectedName: "api/v2/deploy",
		},
		{
			name:         "nested command with absolute base path",
			setupPath:    "commands/db/deploy.md",
			basePath:     "commands",
			expectedName: "db/deploy",
		},
		{
			name:         "flat command with base path",
			setupPath:    "commands/test.md",
			basePath:     "commands",
			expectedName: "test",
		},
		{
			name:         "no base path provided",
			setupPath:    "commands/api/deploy.md",
			basePath:     "",
			expectedName: "deploy", // Without basePath, Name is just the filename
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory structure
			tmpDir := t.TempDir()
			fullPath := filepath.Join(tmpDir, tt.setupPath)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}

			// Write test command
			content := `---
description: Test command
---
# Test
`
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}

			// Calculate absolute base path if provided
			var absBasePath string
			if tt.basePath != "" {
				absBasePath = filepath.Join(tmpDir, tt.basePath)
			}

			// Load command with base
			res, err := LoadCommandWithBase(fullPath, absBasePath)
			if err != nil {
				t.Fatalf("LoadCommandWithBase failed: %v", err)
			}

			// Check Name field (now contains nested path for nested commands)
			if res.Name != tt.expectedName {
				t.Errorf("Name mismatch. Got: %s, Want: %s", res.Name, tt.expectedName)
			}
		})
	}
}

// TestLoadCommandWithBase_InvalidBasePath tests error handling
func TestLoadCommandWithBase_InvalidBasePath(t *testing.T) {
	tmpDir := t.TempDir()
	cmdPath := filepath.Join(tmpDir, "test.md")

	content := `---
description: Test
---
# Test
`
	if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Load with base path that doesn't contain the file
	res, err := LoadCommandWithBase(cmdPath, "/some/other/path")
	if err != nil {
		t.Fatalf("Should not error with invalid base: %v", err)
	}

	// Name should be just the filename when file is not under basePath
	expectedName := "test"
	if res.Name != expectedName {
		t.Errorf("Name should be just filename for file outside basePath. Got: %s, Want: %s", res.Name, expectedName)
	}
}

// TestAutoDetectCommandsBase tests the autoDetectCommandsBase helper function
func TestAutoDetectCommandsBase(t *testing.T) {
	tests := []struct {
		name         string
		setupPath    string // File path to create
		expectedBase string // Expected base path (relative to tmpDir)
	}{
		{
			name:         "flat command in commands/",
			setupPath:    "commands/test.md",
			expectedBase: "commands",
		},
		{
			name:         "nested command one level deep",
			setupPath:    "commands/api/deploy.md",
			expectedBase: "commands",
		},
		{
			name:         "deeply nested command",
			setupPath:    "commands/dt/cluster/overview.md",
			expectedBase: "commands",
		},
		{
			name:         "command in tool directory",
			setupPath:    ".claude/commands/test.md",
			expectedBase: ".claude/commands",
		},
		{
			name:         "command in nested tool directory",
			setupPath:    ".opencode/commands/build/deploy.md",
			expectedBase: ".opencode/commands",
		},
		{
			name:         "not in commands directory",
			setupPath:    "tmp/test.md",
			expectedBase: "", // Empty string - not found
		},
		{
			name:         "wrong directory name",
			setupPath:    "agents/test.md",
			expectedBase: "", // Empty string - not found
		},
		{
			name:         "commands in middle of path",
			setupPath:    "foo/commands/bar/test.md",
			expectedBase: "foo/commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory structure
			tmpDir := t.TempDir()
			fullPath := filepath.Join(tmpDir, tt.setupPath)

			// Create directory structure (file doesn't need to exist)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}

			// Call autoDetectCommandsBase
			result := autoDetectCommandsBase(fullPath)

			// Calculate expected absolute path
			var expectedAbs string
			if tt.expectedBase != "" {
				expectedAbs = filepath.Join(tmpDir, tt.expectedBase)
			}

			// Verify result
			if result != expectedAbs {
				t.Errorf("autoDetectCommandsBase() = %q, want %q", result, expectedAbs)
			}
		})
	}
}

// TestAutoDetectCommandsBase_EdgeCases tests edge cases
func TestAutoDetectCommandsBase_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "root path",
			input:    "/test.md",
			expected: "",
		},
		{
			name:     "relative path without commands",
			input:    "test.md",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := autoDetectCommandsBase(tt.input)
			if result != tt.expected {
				t.Errorf("autoDetectCommandsBase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLoadCommand_NotInCommandsDir tests that LoadCommand returns error when file is not in commands/ directory
func TestLoadCommand_NotInCommandsDir(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "file in root directory",
			filePath: "test.md",
		},
		{
			name:     "file in wrong directory",
			filePath: "agents/test.md",
		},
		{
			name:     "file in tmp directory",
			filePath: "tmp/test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory structure
			tmpDir := t.TempDir()
			fullPath := filepath.Join(tmpDir, tt.filePath)

			// Create directory structure and write test file
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}

			content := `---
description: Test command
---
# Test
`
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}

			// Attempt to load - should fail
			_, err := LoadCommand(fullPath)
			if err == nil {
				t.Error("Expected error when loading command not in commands/ directory, got nil")
			}

			// Check error message
			expectedErrMsg := "command file must be in a 'commands/' directory"
			if !strings.Contains(err.Error(), expectedErrMsg) {
				t.Errorf("Error message should mention commands/ directory requirement.\nGot: %v\nWant substring: %s", err, expectedErrMsg)
			}
		})
	}
}

// TestLoadCommand_NestedStructure tests that LoadCommand preserves nested structure
func TestLoadCommand_NestedStructure(t *testing.T) {
	tests := []struct {
		name         string
		setupPath    string
		expectedName string
	}{
		{
			name:         "flat command",
			setupPath:    "commands/test.md",
			expectedName: "test",
		},
		{
			name:         "nested one level",
			setupPath:    "commands/api/deploy.md",
			expectedName: "api/deploy",
		},
		{
			name:         "nested two levels",
			setupPath:    "commands/db/migrations/apply.md",
			expectedName: "db/migrations/apply",
		},
		{
			name:         "tool directory",
			setupPath:    ".claude/commands/test.md",
			expectedName: "test",
		},
		{
			name:         "tool directory nested",
			setupPath:    ".opencode/commands/build/deploy.md",
			expectedName: "build/deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory structure
			tmpDir := t.TempDir()
			fullPath := filepath.Join(tmpDir, tt.setupPath)

			// Create directory structure
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}

			// Write test command
			content := `---
description: Test command
---
# Test Command
`
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}

			// Load command
			res, err := LoadCommand(fullPath)
			if err != nil {
				t.Fatalf("LoadCommand failed: %v", err)
			}

			// Verify name is preserved correctly
			if res.Name != tt.expectedName {
				t.Errorf("Name mismatch.\nGot:  %s\nWant: %s", res.Name, tt.expectedName)
			}
		})
	}
}
