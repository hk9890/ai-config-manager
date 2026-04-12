package manifest

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/giturl"
	"gopkg.in/yaml.v3"
)

// Package-level logger for manifest validation logging
var logger *slog.Logger

// SetLogger sets the logger for the manifest package.
// This should be called by the application during initialization.
func SetLogger(l *slog.Logger) {
	logger = l
}

const (
	// ManifestFileName is the default name for project manifest files
	ManifestFileName = "ai.package.yaml"
	// LocalManifestFileName is the optional local overlay manifest.
	LocalManifestFileName = "ai.package.local.yaml"
)

// ProjectManifests contains project manifest files and their merged effective view.
type ProjectManifests struct {
	BasePath  string
	LocalPath string

	Base      *Manifest
	Local     *Manifest
	Effective *Manifest
}

var windowsDrivePathPattern = regexp.MustCompile(`^[a-zA-Z]:[/\\]`)

// HasAny reports whether at least one project manifest file is present.
func (p *ProjectManifests) HasAny() bool {
	if p == nil {
		return false
	}
	return p.Base != nil || p.Local != nil
}

// LoadProjectManifests loads ai.package.yaml and optional ai.package.local.yaml,
// then builds a merged effective manifest view.
func LoadProjectManifests(projectPath string) (*ProjectManifests, error) {
	basePath := filepath.Join(projectPath, ManifestFileName)
	localPath := filepath.Join(projectPath, LocalManifestFileName)

	result := &ProjectManifests{
		BasePath:  basePath,
		LocalPath: localPath,
	}

	if Exists(basePath) {
		base, err := Load(basePath)
		if err != nil {
			return nil, err
		}
		result.Base = base
	}

	if Exists(localPath) {
		local, err := Load(localPath)
		if err != nil {
			return nil, err
		}
		result.Local = local
	}

	if result.HasAny() {
		result.Effective = Merge(result.Base, result.Local)
	}

	return result, nil
}

// Merge builds an effective manifest using additive overlay semantics:
//   - resources: base order preserved, local-only entries appended, exact duplicates removed
//   - install.targets: base order preserved, local-only entries appended, exact duplicates removed
//   - sources: base order preserved, local-only entries appended, canonical duplicates
//     (normalized url+subpath) removed while retaining first-declared name/ref
func Merge(base, local *Manifest) *Manifest {
	merged := &Manifest{
		Resources: []string{},
		Install: InstallConfig{
			Targets: []string{},
		},
		Sources: []ManifestSource{},
	}

	if base != nil {
		merged.Resources = append(merged.Resources, base.Resources...)
		merged.Install.Targets = append(merged.Install.Targets, base.Install.Targets...)
		merged.Sources = append(merged.Sources, base.Sources...)
	}

	if local != nil {
		merged.Resources = appendUniqueStrings(merged.Resources, local.Resources...)
		merged.Install.Targets = appendUniqueStrings(merged.Install.Targets, local.Install.Targets...)
		merged.Sources = appendUniqueSources(merged.Sources, local.Sources...)
	}

	return merged
}

func appendUniqueStrings(existing []string, candidates ...string) []string {
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		seen[item] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, ok := seen[candidate]; ok {
			continue
		}
		existing = append(existing, candidate)
		seen[candidate] = struct{}{}
	}

	return existing
}

func appendUniqueSources(existing []ManifestSource, candidates ...ManifestSource) []ManifestSource {
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		seen[canonicalSourceKey(item)] = struct{}{}
	}

	for _, candidate := range candidates {
		key := canonicalSourceKey(candidate)
		if _, ok := seen[key]; ok {
			continue
		}

		existing = append(existing, candidate)
		seen[key] = struct{}{}
	}

	return existing
}

func canonicalSourceKey(source ManifestSource) string {
	return giturl.NormalizeURL(strings.TrimSpace(source.URL)) + "|" + normalizeSourceSubpath(source.Subpath)
}

func normalizeSourceSubpath(subpathValue string) string {
	trimmed := strings.TrimSpace(subpathValue)
	if trimmed == "" {
		return ""
	}

	cleaned := path.Clean("/" + strings.Trim(trimmed, "/"))
	if cleaned == "/" || cleaned == "/." {
		return ""
	}

	return strings.TrimPrefix(cleaned, "/")
}

func isLocalPathLikeSourceURL(url string) bool {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return false
	}

	lowerTrimmed := strings.ToLower(trimmed)
	if strings.HasPrefix(trimmed, "./") ||
		strings.HasPrefix(trimmed, "../") ||
		strings.HasPrefix(trimmed, "/") ||
		strings.HasPrefix(trimmed, "~/") ||
		strings.HasPrefix(trimmed, "\\") ||
		strings.HasPrefix(lowerTrimmed, "file://") ||
		windowsDrivePathPattern.MatchString(trimmed) {
		return true
	}

	return false
}

// Manifest represents a project's AI resource dependencies
// Similar to npm's package.json, it declares which resources should be installed
// InstallConfig holds installation-related configuration
type InstallConfig struct {
	// Targets specifies which AI tools to install to
	// Valid values: claude, opencode, copilot
	Targets []string `yaml:"targets"`
}

// ManifestSource declares a remote source dependency for a project manifest.
// Only remote URL declarations are supported in ai.package.yaml.
type ManifestSource struct {
	URL     string `yaml:"url"`
	Ref     string `yaml:"ref,omitempty"`
	Subpath string `yaml:"subpath,omitempty"`
	Name    string `yaml:"name,omitempty"`
}

// UnmarshalYAML enforces the ai.package.yaml source schema.
// Local/path-style declarations are intentionally rejected.
func (s *ManifestSource) UnmarshalYAML(value *yaml.Node) error {
	type manifestSourceAlias ManifestSource

	var keys map[string]interface{}
	if err := value.Decode(&keys); err != nil {
		return err
	}

	for key := range keys {
		switch key {
		case "url", "ref", "subpath", "name":
		default:
			return fmt.Errorf("invalid sources entry field %q: only url, ref, subpath, and name are allowed", key)
		}
	}

	var alias manifestSourceAlias
	if err := value.Decode(&alias); err != nil {
		return err
	}

	*s = ManifestSource(alias)
	return nil
}

type Manifest struct {
	// Resources is an array of resource references in "type/name" format
	// Examples: "skill/pdf-processing", "command/test", "agent/code-reviewer"
	Resources []string `yaml:"resources"`

	// Install configuration for installation targets
	Install InstallConfig `yaml:"install,omitempty"`

	// Sources declares optional remote catalogs used by this project.
	// Canonical identity for merge/dedup is normalized url+subpath.
	Sources []ManifestSource `yaml:"sources,omitempty"`

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

	for i := range m.Sources {
		source := &m.Sources[i]
		source.URL = strings.TrimSpace(source.URL)
		source.Ref = strings.TrimSpace(source.Ref)
		source.Subpath = normalizeSourceSubpath(source.Subpath)
		source.Name = strings.TrimSpace(source.Name)

		if source.URL == "" {
			return fmt.Errorf("invalid sources[%d]: url is required", i)
		}

		if isLocalPathLikeSourceURL(source.URL) {
			return fmt.Errorf("invalid sources[%d]: local/path sources are not supported in ai.package.yaml", i)
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
// Names can contain slashes for nested resources (e.g., "command/api/deploy")
func validateResourceReference(ref string) error {
	if logger != nil {
		logger.Debug("validating resource reference",
			"reference", ref,
			"parser", "manifest.validateResourceReference")
	}

	if ref == "" {
		if logger != nil {
			logger.Debug("validation failed: empty reference")
		}
		return fmt.Errorf("resource reference cannot be empty")
	}

	parts := strings.Split(ref, "/")
	if logger != nil {
		logger.Debug("split resource reference",
			"parts", parts,
			"count", len(parts))
	}

	if len(parts) < 2 {
		if logger != nil {
			logger.Debug("validation failed: insufficient parts",
				"count", len(parts),
				"reason", "expected at least 2 parts (type/name)")
		}
		return fmt.Errorf("invalid format (expected type/name): %q", ref)
	}

	resourceType := parts[0]
	name := strings.Join(parts[1:], "/")

	if logger != nil {
		logger.Debug("parsed resource components",
			"type", resourceType,
			"name", name)
	}

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
		if logger != nil {
			logger.Debug("validation failed: invalid resource type",
				"type", resourceType,
				"valid_types", validTypes)
		}
		return fmt.Errorf("invalid resource type %q (expected: command, skill, agent, or package)", resourceType)
	}

	// Validate name is not empty
	if name == "" {
		if logger != nil {
			logger.Debug("validation failed: empty name")
		}
		return fmt.Errorf("resource name cannot be empty")
	}

	if logger != nil {
		logger.Debug("validation passed",
			"type", resourceType,
			"name", name)
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
