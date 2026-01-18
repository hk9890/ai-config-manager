package resource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkill(t *testing.T) {
	tests := []struct {
		name      string
		dirPath   string
		wantError bool
		checkName string
		checkDesc string
	}{
		{
			name:      "valid skill",
			dirPath:   "testdata/skills/valid-skill",
			wantError: false,
			checkName: "valid-skill",
			checkDesc: "Extract text and tables from PDF files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := LoadSkill(tt.dirPath)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadSkill() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if res.Name != tt.checkName {
					t.Errorf("LoadSkill() name = %v, want %v", res.Name, tt.checkName)
				}
				if res.Description != tt.checkDesc {
					t.Errorf("LoadSkill() description = %v, want %v", res.Description, tt.checkDesc)
				}
				if res.Type != Skill {
					t.Errorf("LoadSkill() type = %v, want %v", res.Type, Skill)
				}
			}
		})
	}
}

func TestLoadSkillResource(t *testing.T) {
	skill, err := LoadSkillResource("testdata/skills/valid-skill")
	if err != nil {
		t.Fatalf("LoadSkillResource() error = %v", err)
	}

	if skill.Name != "valid-skill" {
		t.Errorf("LoadSkillResource() name = %v, want valid-skill", skill.Name)
	}
	if skill.License != "Apache-2.0" {
		t.Errorf("LoadSkillResource() license = %v, want Apache-2.0", skill.License)
	}
	if skill.Content == "" {
		t.Error("LoadSkillResource() content is empty")
	}
}

func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name      string
		dirPath   string
		wantError bool
	}{
		{
			name:      "valid skill",
			dirPath:   "testdata/skills/valid-skill",
			wantError: false,
		},
		{
			name:      "non-existent directory",
			dirPath:   "testdata/skills/nonexistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkill(tt.dirPath)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSkill() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestWriteSkill(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")

	skill := NewSkillResource("test-skill", "A test skill for writing")
	skill.License = "MIT"
	skill.Content = "# Test Skill\n\nThis is test content."
	skill.Metadata = map[string]string{
		"author":  "test-author",
		"version": "1.0.0",
	}

	err := WriteSkill(skill, skillDir)
	if err != nil {
		t.Fatalf("WriteSkill() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(skillDir); err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}

	// Verify SKILL.md exists
	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err != nil {
		t.Fatalf("SKILL.md was not created: %v", err)
	}

	// Load it back and verify
	res, err := LoadSkill(skillDir)
	if err != nil {
		t.Fatalf("LoadSkill() after write error = %v", err)
	}

	if res.Name != "test-skill" {
		t.Errorf("Loaded skill name = %v, want test-skill", res.Name)
	}
	if res.Description != "A test skill for writing" {
		t.Errorf("Loaded skill description = %v, want 'A test skill for writing'", res.Description)
	}
	if res.License != "MIT" {
		t.Errorf("Loaded skill license = %v, want MIT", res.License)
	}
}

func TestSkillNameMustMatchDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "dir-name")

	// Create skill with mismatched name
	skill := NewSkillResource("different-name", "A test skill")
	skill.Content = "# Test Skill"

	err := WriteSkill(skill, skillDir)
	if err != nil {
		t.Fatalf("WriteSkill() error = %v", err)
	}

	// Try to load - should fail due to name mismatch
	_, err = LoadSkill(skillDir)
	if err == nil {
		t.Error("LoadSkill() expected error for mismatched name, got nil")
	}
}
