package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		wantError    bool
		checkDesc    string
		checkContent string
	}{
		{
			name:         "valid command frontmatter",
			filePath:     "testdata/commands/test-command.md",
			wantError:    false,
			checkDesc:    "Run tests with coverage",
			checkContent: "# Test Command",
		},
		{
			name:         "valid skill frontmatter",
			filePath:     "testdata/skills/valid-skill/SKILL.md",
			wantError:    false,
			checkDesc:    "Extract text and tables from PDF files",
			checkContent: "# Valid Skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, content, err := ParseFrontmatter(tt.filePath)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseFrontmatter() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				desc := fm.GetString("description")
				if desc != tt.checkDesc {
					t.Errorf("ParseFrontmatter() description = %v, want %v", desc, tt.checkDesc)
				}

				if !strings.Contains(content, tt.checkContent) {
					t.Errorf("ParseFrontmatter() content does not contain %v", tt.checkContent)
				}
			}
		})
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "no-frontmatter.md")

	content := "# Just a regular markdown file\n\nNo frontmatter here."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, _, err := ParseFrontmatter(filePath)
	if err == nil {
		t.Error("ParseFrontmatter() expected error for file without frontmatter, got nil")
	}
}

func TestWriteAndReadFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")

	frontmatter := Frontmatter{
		"description": "Test description",
		"version":     "1.0.0",
		"author":      "Test Author",
	}
	content := "# Test Content\n\nThis is the body."

	err := WriteFrontmatter(filePath, frontmatter, content)
	if err != nil {
		t.Fatalf("WriteFrontmatter() error = %v", err)
	}

	// Read it back
	fm, readContent, err := ParseFrontmatter(filePath)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}

	// Verify frontmatter
	if fm.GetString("description") != "Test description" {
		t.Errorf("description = %v, want 'Test description'", fm.GetString("description"))
	}
	if fm.GetString("version") != "1.0.0" {
		t.Errorf("version = %v, want '1.0.0'", fm.GetString("version"))
	}
	if fm.GetString("author") != "Test Author" {
		t.Errorf("author = %v, want 'Test Author'", fm.GetString("author"))
	}

	// Verify content
	if !strings.Contains(readContent, "Test Content") {
		t.Errorf("content does not contain 'Test Content'")
	}
}

func TestFrontmatterGetMap(t *testing.T) {
	fm := Frontmatter{
		"metadata": map[string]interface{}{
			"author":  "test-author",
			"version": "1.0.0",
		},
	}

	metadata := fm.GetMap("metadata")
	if len(metadata) != 2 {
		t.Errorf("GetMap() returned %d items, want 2", len(metadata))
	}
	if metadata["author"] != "test-author" {
		t.Errorf("metadata[author] = %v, want 'test-author'", metadata["author"])
	}
	if metadata["version"] != "1.0.0" {
		t.Errorf("metadata[version] = %v, want '1.0.0'", metadata["version"])
	}
}
