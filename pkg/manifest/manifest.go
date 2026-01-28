package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// ManifestFileName is the default name for project manifest files
	ManifestFileName = "ai.package.yaml"
)

// Manifest represents a project's AI resource dependencies
// Similar to npm's package.json, it declares which resources should be installed
// InstallConfig holds installation-related configuration
type InstallConfig struct {
	// Targets specifies which AI tools to install to
	// Valid values: claude, opencode, copilot
	Targets []string `yaml:"targets"`
}

type Manifest struct {
	// Resources is an array of resource references in "type/name" format
	// Examples: "skill/pdf-processing", "command/test", "agent/code-reviewer"
	Resources []string `yaml:"resources"`

	// Install configuration for installation targets
	Install InstallConfig `yaml:"install,omitempty"`

	// Deprecated: Use Install.Targets instead
	// Kept for backward compatibility when reading old manifests
	Targets []string `yaml:"targets,omitempty"`
}

// Load loads a manifest from a YAML file
// Returns error if file doesn't exist, is invalid YAML, or missing required fields
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Migrate from old format if needed
	if len(m.Targets) > 0 && len(m.Install.Targets) == 0 {
		m.Install.Targets = m.Targets
		m.Targets = nil // Clear old field
	}

	// Validate the manifest
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &m, nil
}

// LoadOrCreate loads a manifest from a file, or creates a new empty manifest if it doesn't exist
func LoadOrCreate(path string) (*Manifest, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Manifest{
			Resources: []string{},
		}, nil
	}

	return Load(path)
}

// Save writes the manifest to a YAML file with pretty formatting
func (m *Manifest) Save(path string) error {
	if m == nil {
		return fmt.Errorf("cannot save nil manifest")
	}

	// Validate before saving
	if err := m.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML with pretty printing
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Validate checks if the manifest is valid
// Returns error if any validation rules are violated
func (m *Manifest) Validate() error {
	if m == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Resources array is optional (can be empty)
	// But if present, validate format
	for _, res := range m.Resources {
		if err := validateResourceReference(res); err != nil {
			return fmt.Errorf("invalid resource '%s': %w", res, err)
		}
	}

	// Install targets are optional
	// If present, validate they're known tools (basic check)
	for _, target := range m.Install.Targets {
		if !isValidTarget(target) {
			return fmt.Errorf("invalid install.targets '%s': must be 'claude', 'opencode', or 'copilot'", target)
		}
	}

	// Also validate old Targets field if present (backward compatibility)
	for _, target := range m.Targets {
		if !isValidTarget(target) {
			return fmt.Errorf("invalid targets '%s': must be 'claude', 'opencode', or 'copilot'", target)
		}
	}

	return nil
}

// Add adds a resource to the manifest
// Resource should be in "type/name" format (e.g., "skill/pdf-processing")
// If the resource already exists, it's not added again
func (m *Manifest) Add(resource string) error {
	if m == nil {
		return fmt.Errorf("cannot add to nil manifest")
	}

	// Validate resource format
	if err := validateResourceReference(resource); err != nil {
		return fmt.Errorf("invalid resource: %w", err)
	}

	// Check if already exists
	if m.Has(resource) {
		return nil // Already exists, nothing to do
	}

	// Add to resources
	m.Resources = append(m.Resources, resource)
	return nil
}

// Remove removes a resource from the manifest
// Returns nil even if the resource doesn't exist
func (m *Manifest) Remove(resource string) error {
	if m == nil {
		return fmt.Errorf("cannot remove from nil manifest")
	}

	// Find and remove the resource
	newResources := make([]string, 0, len(m.Resources))
	for _, r := range m.Resources {
		if r != resource {
			newResources = append(newResources, r)
		}
	}

	m.Resources = newResources
	return nil
}

// Has checks if a resource exists in the manifest
func (m *Manifest) Has(resource string) bool {
	if m == nil {
		return false
	}

	for _, r := range m.Resources {
		if r == resource {
			return true
		}
	}
	return false
}

// Exists checks if a manifest file exists at the given path
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// validateResourceReference validates that a resource reference is in the correct format
// Expected format: "type/name" where type is command|skill|agent|package
func validateResourceReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("resource reference cannot be empty")
	}

	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format (expected type/name): %q", ref)
	}

	resourceType, name := parts[0], parts[1]

	// Validate resource type
	validTypes := []string{"command", "skill", "agent", "package"}
	isValidType := false
	for _, t := range validTypes {
		if resourceType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid resource type %q (expected: command, skill, agent, or package)", resourceType)
	}

	// Validate name is not empty
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	return nil
}

// isValidTarget checks if a target is a known AI tool
func isValidTarget(target string) bool {
	validTargets := []string{"claude", "opencode", "copilot"}
	for _, t := range validTargets {
		if target == t {
			return true
		}
	}
	return false
}
