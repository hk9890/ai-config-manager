package resource

import (
	"os"
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
