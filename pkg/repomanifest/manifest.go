package repomanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// ManifestFileName is the default name for repository manifest files
	ManifestFileName = "ai.repo.yaml"
)

// Manifest represents the ai.repo.yaml file that tracks synced sources
type Manifest struct {
	Version int       `yaml:"version"`
	Sources []*Source `yaml:"sources,omitempty"`
}

// Source represents a single synced source in the repository
type Source struct {
	Name       string    `yaml:"name"`
	Path       string    `yaml:"path,omitempty"`
	URL        string    `yaml:"url,omitempty"`
	Ref        string    `yaml:"ref,omitempty"`
	Subpath    string    `yaml:"subpath,omitempty"`
	Mode       string    `yaml:"mode"`
	Added      time.Time `yaml:"added"`
	LastSynced time.Time `yaml:"last_synced,omitempty"`
}

// Load loads a manifest from the repository's ai.repo.yaml file
// If the file doesn't exist, returns an empty manifest (not an error)
func Load(repoPath string) (*Manifest, error) {
	path := filepath.Join(repoPath, ManifestFileName)

	// If file doesn't exist, return empty manifest
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Manifest{
			Version: 1,
			Sources: []*Source{},
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Validate the manifest
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &m, nil
}

// Save writes the manifest to the repository's ai.repo.yaml file
func (m *Manifest) Save(repoPath string) error {
	if m == nil {
		return fmt.Errorf("cannot save nil manifest")
	}

	// Validate before saving
	if err := m.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML with pretty printing
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	path := filepath.Join(repoPath, ManifestFileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error {
	if m == nil {
		return fmt.Errorf("manifest is nil")
	}

	if m.Version != 1 {
		return fmt.Errorf("invalid version: %d (expected 1)", m.Version)
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, source := range m.Sources {
		if err := validateSource(source); err != nil {
			return fmt.Errorf("invalid source '%s': %w", source.Name, err)
		}

		if names[source.Name] {
			return fmt.Errorf("duplicate source name: %s", source.Name)
		}
		names[source.Name] = true
	}

	return nil
}

// AddSource adds a new source to the manifest
// Auto-generates name if not provided
// Returns error if source with same name already exists
func (m *Manifest) AddSource(source *Source) error {
	if m == nil {
		return fmt.Errorf("cannot add to nil manifest")
	}

	if source == nil {
		return fmt.Errorf("cannot add nil source")
	}

	// Set added timestamp if not set (must be done before validation)
	if source.Added.IsZero() {
		source.Added = time.Now()
	}

	// Auto-generate name if not provided
	if source.Name == "" {
		source.Name = generateSourceName(source)
	}

	// Validate source
	if err := validateSource(source); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// Check for duplicate name
	if m.HasSource(source.Name) {
		return fmt.Errorf("source with name '%s' already exists", source.Name)
	}

	m.Sources = append(m.Sources, source)
	return nil
}

// RemoveSource removes a source by name, path, or URL
// Returns the removed source and nil error if found
// Returns nil and error if not found
func (m *Manifest) RemoveSource(nameOrPath string) (*Source, error) {
	if m == nil {
		return nil, fmt.Errorf("cannot remove from nil manifest")
	}

	if nameOrPath == "" {
		return nil, fmt.Errorf("nameOrPath cannot be empty")
	}

	// Find source by name first, then by path/URL
	for i, source := range m.Sources {
		if source.Name == nameOrPath || source.Path == nameOrPath || source.URL == nameOrPath {
			// Remove from slice
			m.Sources = append(m.Sources[:i], m.Sources[i+1:]...)
			return source, nil
		}
	}

	return nil, fmt.Errorf("source not found: %s", nameOrPath)
}

// GetSource finds a source by name, path, or URL
// Returns the source and true if found, nil and false otherwise
func (m *Manifest) GetSource(nameOrPath string) (*Source, bool) {
	if m == nil || nameOrPath == "" {
		return nil, false
	}

	// Match by name first, then by path/URL
	for _, source := range m.Sources {
		if source.Name == nameOrPath || source.Path == nameOrPath || source.URL == nameOrPath {
			return source, true
		}
	}

	return nil, false
}

// HasSource checks if a source exists by name, path, or URL
func (m *Manifest) HasSource(nameOrPath string) bool {
	_, found := m.GetSource(nameOrPath)
	return found
}

// validateSource validates a single source
func validateSource(source *Source) error {
	if source == nil {
		return fmt.Errorf("source is nil")
	}

	if source.Name == "" {
		return fmt.Errorf("source name cannot be empty")
	}

	// Validate name format (agentskills.io naming rules)
	if !isValidSourceName(source.Name) {
		return fmt.Errorf("invalid source name '%s': must be lowercase alphanumeric with hyphens, 1-64 chars", source.Name)
	}

	// Must have either path or URL
	if source.Path == "" && source.URL == "" {
		return fmt.Errorf("source must have either path or url")
	}

	// Cannot have both path and URL
	if source.Path != "" && source.URL != "" {
		return fmt.Errorf("source cannot have both path and url")
	}

	// Mode must be symlink or copy
	if source.Mode != "symlink" && source.Mode != "copy" {
		return fmt.Errorf("invalid mode '%s': must be 'symlink' or 'copy'", source.Mode)
	}

	// Added timestamp is required
	if source.Added.IsZero() {
		return fmt.Errorf("added timestamp is required")
	}

	return nil
}

// isValidSourceName validates source name follows agentskills.io naming rules
func isValidSourceName(name string) bool {
	if len(name) < 1 || len(name) > 64 {
		return false
	}

	// Must be lowercase alphanumeric + hyphens only
	// Cannot start/end with hyphen
	// No consecutive hyphens
	pattern := `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`
	matched, _ := regexp.MatchString(pattern, name)
	if !matched {
		return false
	}

	// Check for consecutive hyphens
	if strings.Contains(name, "--") {
		return false
	}

	return true
}

// generateSourceName generates a filesystem-safe name from path or URL
func generateSourceName(source *Source) string {
	var base string

	if source.Path != "" {
		// Use last component of path
		base = filepath.Base(source.Path)
	} else if source.URL != "" {
		// Extract repo name from URL
		// Examples:
		// https://github.com/user/repo -> repo
		// https://github.com/user/repo.git -> repo
		// git@github.com:user/repo.git -> repo
		url := source.URL
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			base = parts[len(parts)-1]
		}
		// Handle git@ format
		if strings.Contains(base, ":") {
			parts := strings.Split(base, ":")
			if len(parts) > 1 {
				base = parts[len(parts)-1]
			}
		}
	}

	if base == "" {
		base = "source"
	}

	// Convert to lowercase
	base = strings.ToLower(base)

	// Replace invalid characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name := reg.ReplaceAllString(base, "-")

	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// Replace consecutive hyphens with single hyphen
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Ensure not empty
	if name == "" {
		name = "source"
	}

	// Truncate to 64 chars
	if len(name) > 64 {
		name = name[:64]
		name = strings.TrimRight(name, "-")
	}

	return name
}
