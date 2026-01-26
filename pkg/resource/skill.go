package resource

import (
	"fmt"
	"os"
	"path/filepath"
)

// SkillResource represents a skill resource
type SkillResource struct {
	Resource
	Compatibility []string `yaml:"compatibility,omitempty"`
	Content       string   `yaml:"-"` // The markdown content from SKILL.md
	HasScripts    bool     `yaml:"-"`
	HasReferences bool     `yaml:"-"`
	HasAssets     bool     `yaml:"-"`
}

// LoadSkill loads a skill resource from a directory
func LoadSkill(dirPath string) (*Resource, error) {
	// Validate it's a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, WrapLoadError(dirPath, Skill, fmt.Errorf("failed to stat directory: %w", err))
	}
	if !info.IsDir() {
		return nil, WrapLoadError(dirPath, Skill, fmt.Errorf("skill must be a directory"))
	}

	// Check for SKILL.md
	skillMdPath := filepath.Join(dirPath, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err != nil {
		return nil, WrapLoadError(dirPath, Skill, fmt.Errorf("directory must contain SKILL.md: %w", err))
	}

	// Parse SKILL.md frontmatter
	frontmatter, _, err := ParseFrontmatter(skillMdPath)
	if err != nil {
		return nil, NewValidationError(skillMdPath, "skill", filepath.Base(dirPath), "frontmatter", err)
	}

	// Extract name from directory
	dirName := filepath.Base(dirPath)

	// Get name from frontmatter if present, otherwise use directory name
	name := frontmatter.GetString("name")
	if name == "" {
		name = dirName
	}

	// Validate name matches directory name
	if name != dirName {
		err := fmt.Errorf("skill name '%s' must match directory name '%s'", name, dirName)
		return nil, NewValidationError(dirPath, "skill", name, "name", err)
	}

	// Build resource
	resource := &Resource{
		Name:        name,
		Type:        Skill,
		Description: frontmatter.GetString("description"),
		Version:     frontmatter.GetString("version"),
		Author:      frontmatter.GetString("author"),
		License:     frontmatter.GetString("license"),
		Path:        dirPath,
		Metadata:    frontmatter.GetMap("metadata"),
	}

	// Validate
	if err := resource.Validate(); err != nil {
		return nil, NewValidationError(dirPath, "skill", name, "", err)
	}

	return resource, nil
}

// LoadSkillResource loads a full skill resource with all details
func LoadSkillResource(dirPath string) (*SkillResource, error) {
	// Load base resource
	base, err := LoadSkill(dirPath)
	if err != nil {
		return nil, err
	}

	// Parse SKILL.md frontmatter again for skill-specific fields
	skillMdPath := filepath.Join(dirPath, "SKILL.md")
	frontmatter, content, err := ParseFrontmatter(skillMdPath)
	if err != nil {
		return nil, err
	}

	skill := &SkillResource{
		Resource: *base,
		Content:  content,
	}

	// Extract compatibility
	if compatVal, ok := frontmatter["compatibility"]; ok {
		if compatSlice, ok := compatVal.([]interface{}); ok {
			for _, item := range compatSlice {
				if str, ok := item.(string); ok {
					skill.Compatibility = append(skill.Compatibility, str)
				}
			}
		}
	}

	// Check for optional subdirectories
	skill.HasScripts = dirExists(filepath.Join(dirPath, "scripts"))
	skill.HasReferences = dirExists(filepath.Join(dirPath, "references"))
	skill.HasAssets = dirExists(filepath.Join(dirPath, "assets"))

	return skill, nil
}

// ValidateSkill validates a skill resource structure
func ValidateSkill(dirPath string) error {
	_, err := LoadSkill(dirPath)
	return err
}

// NewSkillResource creates a new skill resource
func NewSkillResource(name, description string) *SkillResource {
	return &SkillResource{
		Resource: Resource{
			Name:        name,
			Type:        Skill,
			Description: description,
			Metadata:    make(map[string]string),
		},
	}
}

// WriteSkill writes a skill resource to a directory
func WriteSkill(skill *SkillResource, dirPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build frontmatter
	frontmatter := Frontmatter{
		"name":        skill.Name,
		"description": skill.Description,
	}

	if skill.License != "" {
		frontmatter["license"] = skill.License
	}
	if len(skill.Compatibility) > 0 {
		frontmatter["compatibility"] = skill.Compatibility
	}
	if len(skill.Metadata) > 0 {
		frontmatter["metadata"] = skill.Metadata
	}
	if skill.Version != "" {
		frontmatter["version"] = skill.Version
	}
	if skill.Author != "" {
		frontmatter["author"] = skill.Author
	}

	// Write SKILL.md
	skillMdPath := filepath.Join(dirPath, "SKILL.md")
	return WriteFrontmatter(skillMdPath, frontmatter, skill.Content)
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
