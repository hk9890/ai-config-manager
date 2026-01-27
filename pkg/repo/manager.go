package repo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	pkgerrors "github.com/hk9890/ai-config-manager/pkg/errors"
	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// Manager manages the AI resources repository
type Manager struct {
	repoPath string
}

// NewManager creates a new repository manager
// Repository is stored at ~/.ai-config/repo/ (XDG data directory)
// Can be overridden with AIMGR_REPO_PATH environment variable
func NewManager() (*Manager, error) {
	// Check for environment variable override first
	repoPath := os.Getenv("AIMGR_REPO_PATH")
	if repoPath == "" {
		// Default to XDG data directory
		repoPath = filepath.Join(xdg.DataHome, "ai-config", "repo")
	}
	return &Manager{
		repoPath: repoPath,
	}, nil
}

// NewManagerWithPath creates a manager with a custom repository path (for testing)
func NewManagerWithPath(repoPath string) *Manager {
	return &Manager{
		repoPath: repoPath,
	}
}

// Init initializes the repository directory structure
func (m *Manager) Init() error {
	// Create main repo directory
	if err := os.MkdirAll(m.repoPath, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Create commands subdirectory
	commandsPath := filepath.Join(m.repoPath, "commands")
	if err := os.MkdirAll(commandsPath, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	// Create skills subdirectory
	skillsPath := filepath.Join(m.repoPath, "skills")
	if err := os.MkdirAll(skillsPath, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Create agents subdirectory
	agentsPath := filepath.Join(m.repoPath, "agents")
	if err := os.MkdirAll(agentsPath, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Create packages subdirectory
	packagesPath := filepath.Join(m.repoPath, "packages")
	if err := os.MkdirAll(packagesPath, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	return nil
}

// AddCommand adds a command resource to the repository.
// Metadata is automatically saved to .metadata/commands/<name>-metadata.json
func (m *Manager) AddCommand(sourcePath, sourceURL, sourceType string) error {
	return m.AddCommandWithRef(sourcePath, sourceURL, sourceType, "")
}

// AddCommandWithRef adds a command resource to the repository with a specified Git ref.
// Metadata is automatically saved to .metadata/commands/<name>-metadata.json
func (m *Manager) AddCommandWithRef(sourcePath, sourceURL, sourceType, ref string) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the command with base path for relative path calculation
	// Strategy: Look for "commands" directory ancestor, or use source's parent dir
	basePath := ""
	cleanPath := filepath.Clean(sourcePath)

	// First try to find a "commands" directory in the path
	if strings.Contains(cleanPath, "commands") {
		parts := strings.Split(cleanPath, string(filepath.Separator))
		for i, part := range parts {
			if part == "commands" {
				// Reconstruct path up to and including "commands"
				// If the original path is absolute, preserve that
				if filepath.IsAbs(cleanPath) {
					basePath = string(filepath.Separator) + filepath.Join(parts[:i+1]...)
				} else {
					basePath = filepath.Join(parts[:i+1]...)
				}
				// Clean to normalize the path
				basePath = filepath.Clean(basePath)
				break
			}
		}
	}

	// If no "commands" directory found, try to find source repo root
	// by looking for common markers (.git, .claude, etc.) or use file's grandparent
	if basePath == "" {
		dir := filepath.Dir(cleanPath)
		// Walk up to find a suitable base
		for dir != "." && dir != "/" {
			// Check for markers of a repo root
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				basePath = dir
				break
			}
			if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
				basePath = dir
				break
			}
			if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
				basePath = dir
				break
			}
			dir = filepath.Dir(dir)
			// Don't go too far up
			if len(strings.Split(dir, string(filepath.Separator))) < 2 {
				break
			}
		}
	}

	res, err := resource.LoadCommandWithBase(sourcePath, basePath)
	if err != nil {
		return fmt.Errorf("failed to load command: %w", err)
	}

	// Check for conflicts using the resource-aware path
	destPath := m.GetPathForResource(res)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("command '%s' already exists in repository", res.Name)
	}

	// Create parent directories if needed (for nested structure)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Copy the file
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy command: %w", err)
	}

	// Create and save metadata
	now := time.Now()
	meta := &metadata.ResourceMetadata{
		Name:           res.Name,
		Type:           resource.Command,
		SourceType:     sourceType,
		SourceURL:      sourceURL,
		Ref:            ref,
		FirstInstalled: now,
		LastUpdated:    now,
	}
	if err := metadata.Save(meta, m.repoPath); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// AddSkill adds a skill resource to the repository.
// Metadata is automatically saved to .metadata/skills/<name>-metadata.json
func (m *Manager) AddSkill(sourcePath, sourceURL, sourceType string) error {
	return m.AddSkillWithRef(sourcePath, sourceURL, sourceType, "")
}

// AddSkillWithRef adds a skill resource to the repository with a specified Git ref.
// Metadata is automatically saved to .metadata/skills/<name>-metadata.json
func (m *Manager) AddSkillWithRef(sourcePath, sourceURL, sourceType, ref string) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the skill
	res, err := resource.LoadSkill(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load skill: %w", err)
	}

	// Check for conflicts
	destPath := m.GetPath(res.Name, resource.Skill)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("skill '%s' already exists in repository", res.Name)
	}

	// Copy the directory
	if err := copyDir(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy skill: %w", err)
	}

	// Create and save metadata
	now := time.Now()
	meta := &metadata.ResourceMetadata{
		Name:           res.Name,
		Type:           resource.Skill,
		SourceType:     sourceType,
		SourceURL:      sourceURL,
		Ref:            ref,
		FirstInstalled: now,
		LastUpdated:    now,
	}
	if err := metadata.Save(meta, m.repoPath); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// AddAgent adds an agent resource to the repository.
// Metadata is automatically saved to .metadata/agents/<name>-metadata.json
func (m *Manager) AddAgent(sourcePath, sourceURL, sourceType string) error {
	return m.AddAgentWithRef(sourcePath, sourceURL, sourceType, "")
}

// AddAgentWithRef adds an agent resource to the repository with a specified Git ref.
// Metadata is automatically saved to .metadata/agents/<name>-metadata.json
func (m *Manager) AddAgentWithRef(sourcePath, sourceURL, sourceType, ref string) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the agent
	res, err := resource.LoadAgent(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load agent: %w", err)
	}

	// Check for conflicts
	destPath := m.GetPath(res.Name, resource.Agent)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("agent '%s' already exists in repository", res.Name)
	}

	// Copy the file
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy agent: %w", err)
	}

	// Create and save metadata
	now := time.Now()
	meta := &metadata.ResourceMetadata{
		Name:           res.Name,
		Type:           resource.Agent,
		SourceType:     sourceType,
		SourceURL:      sourceURL,
		Ref:            ref,
		FirstInstalled: now,
		LastUpdated:    now,
	}
	if err := metadata.Save(meta, m.repoPath); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// AddPackage adds a package resource to the repository.
// Metadata is automatically saved to .metadata/packages/<name>-metadata.json
func (m *Manager) AddPackage(sourcePath, sourceURL, sourceType string) error {
	return m.AddPackageWithRef(sourcePath, sourceURL, sourceType, "")
}

// AddPackageWithRef adds a package resource to the repository with a specified Git ref.
// Metadata is automatically saved to .metadata/packages/<name>-metadata.json
func (m *Manager) AddPackageWithRef(sourcePath, sourceURL, sourceType, ref string) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the package
	pkg, err := resource.LoadPackage(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	// Check for conflicts
	destPath := resource.GetPackagePath(pkg.Name, m.repoPath)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("package '%s' already exists in repository", pkg.Name)
	}

	// Copy the file
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}

	// Create and save metadata
	now := time.Now()
	pkgMeta := &metadata.PackageMetadata{
		Name:          pkg.Name,
		SourceType:    sourceType,
		SourceURL:     sourceURL,
		SourceRef:     ref,
		FirstAdded:    now,
		LastUpdated:   now,
		ResourceCount: len(pkg.Resources),
	}
	if err := metadata.SavePackageMetadata(pkgMeta, m.repoPath); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// validatePackageResources checks if all referenced resources exist in the repository.
// Returns a slice of missing resource references.
func (m *Manager) validatePackageResources(pkg *resource.Package) []string {
	var missing []string

	for _, ref := range pkg.Resources {
		// Parse the resource reference
		resType, resName, err := resource.ParseResourceReference(ref)
		if err != nil {
			// Invalid reference format
			missing = append(missing, ref)
			continue
		}

		// Check if resource exists
		resPath := m.GetPath(resName, resType)
		if _, err := os.Stat(resPath); os.IsNotExist(err) {
			missing = append(missing, ref)
		}
	}

	return missing
}

// List lists all resources, optionally filtered by type
func (m *Manager) List(resourceType *resource.ResourceType) ([]resource.Resource, error) {
	var resources []resource.Resource

	// List commands if no filter or filter is Command
	if resourceType == nil || *resourceType == resource.Command {
		commandsPath := filepath.Join(m.repoPath, "commands")
		if _, err := os.Stat(commandsPath); err == nil {
			// Use filepath.Walk to find commands recursively (supports nested structure)
			err := filepath.Walk(commandsPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip directories and non-.md files
				if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
					return nil
				}

				// Load command with base path to calculate RelativePath
				res, err := resource.LoadCommandWithBase(path, commandsPath)
				if err != nil {
					// Skip invalid commands
					return nil
				}
				resources = append(resources, *res)
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk commands directory: %w", err)
			}
		}
	}

	// List skills if no filter or filter is Skill
	if resourceType == nil || *resourceType == resource.Skill {
		skillsPath := filepath.Join(m.repoPath, "skills")
		if _, err := os.Stat(skillsPath); err == nil {
			entries, err := os.ReadDir(skillsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read skills directory: %w", err)
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				skillPath := filepath.Join(skillsPath, entry.Name())
				res, err := resource.LoadSkill(skillPath)
				if err != nil {
					// Skip invalid skills
					continue
				}
				resources = append(resources, *res)
			}
		}
	}

	// List agents if no filter or filter is Agent
	if resourceType == nil || *resourceType == resource.Agent {
		agentsPath := filepath.Join(m.repoPath, "agents")
		if _, err := os.Stat(agentsPath); err == nil {
			entries, err := os.ReadDir(agentsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read agents directory: %w", err)
			}

			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}

				agentPath := filepath.Join(agentsPath, entry.Name())
				res, err := resource.LoadAgent(agentPath)
				if err != nil {
					// Skip invalid agents
					continue
				}
				resources = append(resources, *res)
			}
		}
	}

	// List packages if no filter or filter is PackageType
	if resourceType == nil || *resourceType == resource.PackageType {
		packagesPath := filepath.Join(m.repoPath, "packages")
		if _, err := os.Stat(packagesPath); err == nil {
			entries, err := os.ReadDir(packagesPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read packages directory: %w", err)
			}

			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".package.json") {
					continue
				}

				packagePath := filepath.Join(packagesPath, entry.Name())
				pkg, err := resource.LoadPackage(packagePath)
				if err != nil {
					// Skip invalid packages
					continue
				}
				// Convert Package to Resource format
				res := resource.Resource{
					Name:        pkg.Name,
					Type:        resource.PackageType,
					Description: pkg.Description,
					Path:        packagePath,
				}
				resources = append(resources, res)
			}
		}
	}

	return resources, nil
}

// PackageInfo represents package information for listing
type PackageInfo struct {
	Name          string
	Description   string
	ResourceCount int
}

// ListPackages lists all packages in the repository
func (m *Manager) ListPackages() ([]PackageInfo, error) {
	var packages []PackageInfo

	packagesPath := filepath.Join(m.repoPath, "packages")
	if _, err := os.Stat(packagesPath); err != nil {
		// Packages directory doesn't exist, return empty list
		return packages, nil
	}

	entries, err := os.ReadDir(packagesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".package.json") {
			continue
		}

		pkgPath := filepath.Join(packagesPath, entry.Name())
		pkg, err := resource.LoadPackage(pkgPath)
		if err != nil {
			// Skip invalid packages
			continue
		}

		packages = append(packages, PackageInfo{
			Name:          pkg.Name,
			Description:   pkg.Description,
			ResourceCount: len(pkg.Resources),
		})
	}

	return packages, nil
}

// Get retrieves a specific resource by name and type
func (m *Manager) Get(name string, resourceType resource.ResourceType) (*resource.Resource, error) {
	path := m.GetPath(name, resourceType)

	// Check if resource exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("resource '%s' not found", name)
	}

	// Load the resource
	switch resourceType {
	case resource.Command:
		return resource.LoadCommand(path)
	case resource.Skill:
		return resource.LoadSkill(path)
	case resource.Agent:
		return resource.LoadAgent(path)
	default:
		return nil, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

// Remove removes a resource from the repository.
// Also removes associated metadata from .metadata/<type>s/<name>-metadata.json
func (m *Manager) Remove(name string, resourceType resource.ResourceType) error {
	path := m.GetPath(name, resourceType)

	// Check if resource exists
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("resource '%s' not found", name)
	}

	// Remove the resource
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove resource: %w", err)
	}

	// Remove metadata file
	metadataPath := metadata.GetMetadataPath(name, resourceType, m.repoPath)
	if _, err := os.Stat(metadataPath); err == nil {
		if err := os.Remove(metadataPath); err != nil {
			return fmt.Errorf("failed to remove metadata: %w", err)
		}
	}

	return nil
}

// GetPath returns the full path to a resource in the repository
// Handles nested paths for commands (e.g., "api/deploy" -> "commands/api/deploy.md")
func (m *Manager) GetPath(name string, resourceType resource.ResourceType) string {
	switch resourceType {
	case resource.Command:
		return filepath.Join(m.repoPath, "commands", name) + ".md"
	case resource.Skill:
		return filepath.Join(m.repoPath, "skills", name)
	case resource.Agent:
		return filepath.Join(m.repoPath, "agents", name+".md")
	default:
		return ""
	}
}

// GetPathForResource returns the full path for a resource, using RelativePath if available
func (m *Manager) GetPathForResource(res *resource.Resource) string {
	switch res.Type {
	case resource.Command:
		if res.RelativePath != "" {
			return filepath.Join(m.repoPath, "commands", res.RelativePath+".md")
		}
		return filepath.Join(m.repoPath, "commands", res.Name+".md")
	case resource.Skill:
		return filepath.Join(m.repoPath, "skills", res.Name)
	case resource.Agent:
		return filepath.Join(m.repoPath, "agents", res.Name+".md")
	default:
		return ""
	}
}

// GetRepoPath returns the repository root path
func (m *Manager) GetRepoPath() string {
	return m.repoPath
}

// GetMetadata retrieves metadata for a specific resource.
// Loads metadata from .metadata/<type>s/<name>-metadata.json
func (m *Manager) GetMetadata(name string, resourceType resource.ResourceType) (*metadata.ResourceMetadata, error) {
	return metadata.Load(name, resourceType, m.repoPath)
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// BulkImportOptions contains options for bulk import operations
type BulkImportOptions struct {
	Force        bool   // Overwrite existing resources
	SkipExisting bool   // Skip conflicts silently
	DryRun       bool   // Preview only, don't actually import
	SourceURL    string // Original source URL (for Git sources)
	SourceType   string // Source type (github, git-url, file, local)
	Ref          string // Git ref (branch/tag/commit), defaults to "main" if empty
}

// ImportError represents an error during resource import
type ImportError struct {
	Path    string
	Message string
}

// BulkImportResult contains the results of a bulk import operation
type BulkImportResult struct {
	Added        []string      // Successfully added resources
	Skipped      []string      // Skipped due to conflicts
	Failed       []ImportError // Failed imports with reasons
	CommandCount int           // Number of commands imported
	SkillCount   int           // Number of skills imported
	AgentCount   int           // Number of agents imported
	PackageCount int           // Number of packages imported
}

// AddBulk imports multiple resources at once
func (m *Manager) AddBulk(sources []string, opts BulkImportOptions) (*BulkImportResult, error) {
	result := &BulkImportResult{
		Added:        []string{},
		Skipped:      []string{},
		Failed:       []ImportError{},
		CommandCount: 0,
		SkillCount:   0,
		AgentCount:   0,
		PackageCount: 0,
	}

	// Ensure repo is initialized (even for dry run, to check paths)
	if !opts.DryRun {
		if err := m.Init(); err != nil {
			return nil, err
		}
	}

	// Process each source
	for _, sourcePath := range sources {
		if err := m.importResource(sourcePath, opts, result); err != nil {
			// Only stop on fatal errors (internal bugs, system failures)
			// Continue on validation/resource errors (they're already collected in result)
			if pkgerrors.IsFatal(err) {
				return result, err
			}
		}
	}

	return result, nil
}

// importResource imports a single resource (helper for AddBulk)
func (m *Manager) importResource(sourcePath string, opts BulkImportOptions, result *BulkImportResult) error {
	// Check if it's a package file
	if strings.HasSuffix(sourcePath, ".package.json") {
		return m.importPackage(sourcePath, opts, result)
	}

	// Detect resource type
	resourceType, err := resource.DetectType(sourcePath)
	if err != nil {
		// Validation error: invalid resource format
		typedErr := pkgerrors.Validation(err, "failed to detect resource type")
		result.Failed = append(result.Failed, ImportError{
			Path:    sourcePath,
			Message: typedErr.Error(),
		})
		return typedErr
	}

	// Load the resource to get its name
	var res *resource.Resource
	switch resourceType {
	case resource.Command:
		res, err = resource.LoadCommand(sourcePath)
	case resource.Skill:
		res, err = resource.LoadSkill(sourcePath)
	case resource.Agent:
		res, err = resource.LoadAgent(sourcePath)
	default:
		err = fmt.Errorf("unknown resource type: %s", resourceType)
	}

	if err != nil {
		// Validation error: invalid YAML/frontmatter
		typedErr := pkgerrors.Validation(err, "failed to load resource")
		result.Failed = append(result.Failed, ImportError{
			Path:    sourcePath,
			Message: typedErr.Error(),
		})
		return typedErr
	}

	// Check if resource already exists
	destPath := m.GetPath(res.Name, resourceType)
	_, statErr := os.Stat(destPath)
	exists := statErr == nil

	if exists {
		if opts.Force {
			// Force mode: remove existing and continue
			if !opts.DryRun {
				if err := m.Remove(res.Name, resourceType); err != nil {
					// Could be resource error (permission denied) or fatal (system failure)
					// Default to resource error for safety
					typedErr := pkgerrors.Resource(err, "failed to remove existing resource")
					result.Failed = append(result.Failed, ImportError{
						Path:    sourcePath,
						Message: typedErr.Error(),
					})
					return typedErr
				}
			}
		} else if opts.SkipExisting {
			// Skip mode: skip this resource
			result.Skipped = append(result.Skipped, sourcePath)
			return nil
		} else {
			// Default mode: fail on conflict
			// Validation error: resource name conflict
			typedErr := pkgerrors.Validation(fmt.Errorf("resource '%s' already exists in repository", res.Name), "")
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: typedErr.Error(),
			})
			return typedErr
		}
	}

	// Import the resource (unless dry run)
	if !opts.DryRun {
		// Determine source URL and type
		var sourceURL, sourceType, ref string

		// Use provided source info if available, otherwise fall back to file://
		if opts.SourceURL != "" && opts.SourceType != "" {
			sourceURL = opts.SourceURL
			sourceType = opts.SourceType
			ref = opts.Ref
		} else {
			// Fall back to file:// for local sources
			absPath, err := filepath.Abs(sourcePath)
			if err != nil {
				absPath = sourcePath
			}
			sourceURL = "file://" + absPath
			sourceType = "file"
			ref = ""
		}

		switch resourceType {
		case resource.Command:
			err = m.AddCommandWithRef(sourcePath, sourceURL, sourceType, ref)
		case resource.Skill:
			err = m.AddSkillWithRef(sourcePath, sourceURL, sourceType, ref)
		case resource.Agent:
			err = m.AddAgentWithRef(sourcePath, sourceURL, sourceType, ref)
		}

		if err != nil {
			// Could be various errors - wrap as validation for now
			typedErr := pkgerrors.Validation(err, "failed to import resource")
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: typedErr.Error(),
			})
			return typedErr
		}
	}

	// Increment resource type counter (count in both dry-run and real mode)
	switch resourceType {
	case resource.Command:
		result.CommandCount++
	case resource.Skill:
		result.SkillCount++
	case resource.Agent:
		result.AgentCount++
	}

	result.Added = append(result.Added, sourcePath)
	return nil
}

// importPackage imports a single package (helper for AddBulk)
func (m *Manager) importPackage(sourcePath string, opts BulkImportOptions, result *BulkImportResult) error {
	// Load the package to get its name
	pkg, err := resource.LoadPackage(sourcePath)
	if err != nil {
		// Validation error: invalid package format
		typedErr := pkgerrors.Validation(err, "failed to load package")
		result.Failed = append(result.Failed, ImportError{
			Path:    sourcePath,
			Message: typedErr.Error(),
		})
		return typedErr
	}

	// Check if package already exists
	destPath := resource.GetPackagePath(pkg.Name, m.repoPath)
	_, statErr := os.Stat(destPath)
	exists := statErr == nil

	if exists {
		if opts.Force {
			// Force mode: remove existing and continue
			if !opts.DryRun {
				// Remove package file
				if err := os.Remove(destPath); err != nil {
					// Resource error: failed to delete file
					typedErr := pkgerrors.Resource(err, "failed to remove existing package")
					result.Failed = append(result.Failed, ImportError{
						Path:    sourcePath,
						Message: typedErr.Error(),
					})
					return typedErr
				}
				// Remove metadata file
				metadataPath := metadata.GetPackageMetadataPath(pkg.Name, m.repoPath)
				if _, err := os.Stat(metadataPath); err == nil {
					if err := os.Remove(metadataPath); err != nil {
						// Resource error: failed to delete metadata
						typedErr := pkgerrors.Resource(err, "failed to remove metadata")
						result.Failed = append(result.Failed, ImportError{
							Path:    sourcePath,
							Message: typedErr.Error(),
						})
						return typedErr
					}
				}
			}
		} else if opts.SkipExisting {
			// Skip mode: skip this package
			result.Skipped = append(result.Skipped, sourcePath)
			return nil
		} else {
			// Default mode: fail on conflict
			// Validation error: package name conflict
			typedErr := pkgerrors.Validation(fmt.Errorf("package '%s' already exists in repository", pkg.Name), "")
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: typedErr.Error(),
			})
			return typedErr
		}
	}

	// Import the package (unless dry run)
	if !opts.DryRun {
		// Determine source URL and type
		var sourceURL, sourceType, ref string

		// Use provided source info if available, otherwise fall back to file://
		if opts.SourceURL != "" && opts.SourceType != "" {
			sourceURL = opts.SourceURL
			sourceType = opts.SourceType
			ref = opts.Ref
		} else {
			// Fall back to file:// for local sources
			absPath, err := filepath.Abs(sourcePath)
			if err != nil {
				absPath = sourcePath
			}
			sourceURL = "file://" + absPath
			sourceType = "file"
			ref = ""
		}

		err = m.AddPackageWithRef(sourcePath, sourceURL, sourceType, ref)
		if err != nil {
			// Could be various errors - wrap as validation for now
			typedErr := pkgerrors.Validation(err, "failed to import package")
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: typedErr.Error(),
			})
			return typedErr
		}
	}

	// Increment package counter (count in both dry-run and real mode)
	result.PackageCount++

	result.Added = append(result.Added, sourcePath)
	return nil
}

// ValidatePackageResources checks if all referenced resources exist in the repository.
// Returns a slice of missing resource references. (Public version for verify command)
func (m *Manager) ValidatePackageResources(pkg *resource.Package) []string {
	return m.validatePackageResources(pkg)
}

// GetPackage loads a package by name from the repository
func (m *Manager) GetPackage(name string) (*resource.Package, error) {
	pkgPath := resource.GetPackagePath(name, m.repoPath)
	return resource.LoadPackage(pkgPath)
}
