package repo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
)

// Manager manages the AI resources repository
type Manager struct {
	repoPath string
}

// NewManager creates a new repository manager
// Repository is stored at ~/.ai-config/repo/ (XDG data directory)
func NewManager() (*Manager, error) {
	repoPath := filepath.Join(xdg.DataHome, "ai-config", "repo")
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

	return nil
}

// AddCommand adds a command resource to the repository
func (m *Manager) AddCommand(sourcePath string) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the command
	res, err := resource.LoadCommand(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load command: %w", err)
	}

	// Check for conflicts
	destPath := m.GetPath(res.Name, resource.Command)
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("command '%s' already exists in repository", res.Name)
	}

	// Copy the file
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy command: %w", err)
	}

	return nil
}

// AddSkill adds a skill resource to the repository
func (m *Manager) AddSkill(sourcePath string) error {
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

	return nil
}

// AddAgent adds an agent resource to the repository
func (m *Manager) AddAgent(sourcePath string) error {
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

	return nil
}

// List lists all resources, optionally filtered by type
func (m *Manager) List(resourceType *resource.ResourceType) ([]resource.Resource, error) {
	var resources []resource.Resource

	// List commands if no filter or filter is Command
	if resourceType == nil || *resourceType == resource.Command {
		commandsPath := filepath.Join(m.repoPath, "commands")
		if _, err := os.Stat(commandsPath); err == nil {
			entries, err := os.ReadDir(commandsPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read commands directory: %w", err)
			}

			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}

				cmdPath := filepath.Join(commandsPath, entry.Name())
				res, err := resource.LoadCommand(cmdPath)
				if err != nil {
					// Skip invalid commands
					continue
				}
				resources = append(resources, *res)
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

	return resources, nil
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

// Remove removes a resource from the repository
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

	return nil
}

// GetPath returns the full path to a resource in the repository
func (m *Manager) GetPath(name string, resourceType resource.ResourceType) string {
	switch resourceType {
	case resource.Command:
		return filepath.Join(m.repoPath, "commands", name+".md")
	case resource.Skill:
		return filepath.Join(m.repoPath, "skills", name)
	case resource.Agent:
		return filepath.Join(m.repoPath, "agents", name+".md")
	default:
		return ""
	}
}

// GetRepoPath returns the repository root path
func (m *Manager) GetRepoPath() string {
	return m.repoPath
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
	Force        bool // Overwrite existing resources
	SkipExisting bool // Skip conflicts silently
	DryRun       bool // Preview only, don't actually import
}

// ImportError represents an error during resource import
type ImportError struct {
	Path    string
	Message string
}

// BulkImportResult contains the results of a bulk import operation
type BulkImportResult struct {
	Added   []string      // Successfully added resources
	Skipped []string      // Skipped due to conflicts
	Failed  []ImportError // Failed imports with reasons
}

// AddBulk imports multiple resources at once
func (m *Manager) AddBulk(sources []string, opts BulkImportOptions) (*BulkImportResult, error) {
	result := &BulkImportResult{
		Added:   []string{},
		Skipped: []string{},
		Failed:  []ImportError{},
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
			// If not skipping existing, fail on first error
			if !opts.SkipExisting && !opts.Force {
				return result, err
			}
		}
	}

	return result, nil
}

// importResource imports a single resource (helper for AddBulk)
func (m *Manager) importResource(sourcePath string, opts BulkImportOptions, result *BulkImportResult) error {
	// Detect resource type
	resourceType, err := resource.DetectType(sourcePath)
	if err != nil {
		result.Failed = append(result.Failed, ImportError{
			Path:    sourcePath,
			Message: err.Error(),
		})
		return err
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
		result.Failed = append(result.Failed, ImportError{
			Path:    sourcePath,
			Message: fmt.Sprintf("failed to load resource: %v", err),
		})
		return err
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
					result.Failed = append(result.Failed, ImportError{
						Path:    sourcePath,
						Message: fmt.Sprintf("failed to remove existing resource: %v", err),
					})
					return err
				}
			}
		} else if opts.SkipExisting {
			// Skip mode: skip this resource
			result.Skipped = append(result.Skipped, sourcePath)
			return nil
		} else {
			// Default mode: fail on conflict
			err := fmt.Errorf("resource '%s' already exists in repository", res.Name)
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: err.Error(),
			})
			return err
		}
	}

	// Import the resource (unless dry run)
	if !opts.DryRun {
		switch resourceType {
		case resource.Command:
			err = m.AddCommand(sourcePath)
		case resource.Skill:
			err = m.AddSkill(sourcePath)
		case resource.Agent:
			err = m.AddAgent(sourcePath)
		}

		if err != nil {
			result.Failed = append(result.Failed, ImportError{
				Path:    sourcePath,
				Message: err.Error(),
			})
			return err
		}
	}

	result.Added = append(result.Added, sourcePath)
	return nil
}
