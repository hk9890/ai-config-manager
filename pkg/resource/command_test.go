package resource

import (
	"os"
	"strings"
	"path/filepath"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := LoadCommand(tt.filePath)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadCommand() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if res.Name != tt.checkName {
					t.Errorf("LoadCommand() name = %v, want %v", res.Name, tt.checkName)
				}
				if res.Description != tt.checkDesc {
					t.Errorf("LoadCommand() description = %v, want %v", res.Description, tt.checkDesc)
				}
				if res.Type != Command {
					t.Errorf("LoadCommand() type = %v, want %v", res.Type, Command)
				}
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
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
			name:      "non-existent file",
			filePath:  "testdata/commands/nonexistent.md",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.filePath)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateCommand() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestWriteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-write.md")

	cmd := NewCommandResource("test-write", "A test command for writing")
	cmd.Agent = "test-agent"
	cmd.Model = "test-model"
	cmd.Content = "# Test Content\n\nThis is test content."

	err := WriteCommand(cmd, filePath)
	if err != nil {
		t.Fatalf("WriteCommand() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("File was not created: %v", err)
	}

	// Load it back and verify
	res, err := LoadCommand(filePath)
	if err != nil {
		t.Fatalf("LoadCommand() after write error = %v", err)
	}

	if res.Name != "test-write" {
		t.Errorf("Loaded command name = %v, want test-write", res.Name)
	}
	if res.Description != "A test command for writing" {
		t.Errorf("Loaded command description = %v, want 'A test command for writing'", res.Description)
	}
}

// TestLoadCommandWithBase tests the LoadCommandWithBase function
func TestLoadCommandWithBase(t *testing.T) {
	tests := []struct {
		name             string
		setupPath        string   // Path structure to create
		basePath         string   // Base path for relative calculation
		expectedRelPath  string   // Expected RelativePath
	}{
		{
			name:            "nested command with base path",
			setupPath:       "commands/api/v2/deploy.md",
			basePath:        "commands",
			expectedRelPath: "api/v2/deploy",
		},
		{
			name:            "nested command with absolute base path",
			setupPath:       "commands/db/deploy.md",
			basePath:        "commands",
			expectedRelPath: "db/deploy",
		},
		{
			name:            "flat command with base path",
			setupPath:       "commands/test.md",
			basePath:        "commands",
			expectedRelPath: "test",
		},
		{
			name:            "no base path provided",
			setupPath:       "commands/api/deploy.md",
			basePath:        "",
			expectedRelPath: "",
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

			// Check RelativePath
			if res.RelativePath != tt.expectedRelPath {
				t.Errorf("RelativePath mismatch. Got: %s, Want: %s", res.RelativePath, tt.expectedRelPath)
			}

			// Verify name is still just the filename
			expectedName := strings.TrimSuffix(filepath.Base(fullPath), ".md")
			if res.Name != expectedName {
				t.Errorf("Name mismatch. Got: %s, Want: %s", res.Name, expectedName)
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

	// RelativePath should be empty when file is not under basePath
	if res.RelativePath != "" {
		t.Errorf("RelativePath should be empty for file outside basePath. Got: %s", res.RelativePath)
	}
}
