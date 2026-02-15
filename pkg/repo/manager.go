package repo

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/hk9890/ai-config-manager/pkg/config"
	pkgerrors "github.com/hk9890/ai-config-manager/pkg/errors"
	"github.com/hk9890/ai-config-manager/pkg/logging"
	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/repomanifest"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// Manager manages the AI resources repository
type Manager struct {
	repoPath string
	logger   *slog.Logger
}

// NewManager creates a new repository manager
// Repository path determined by 3-level precedence:
// 1. AIMGR_REPO_PATH environment variable (highest priority)
// 2. repo.path from config file (~/.config/aimgr/aimgr.yaml)
// 3. XDG default (~/.local/share/ai-config/repo/)
func NewManager() (*Manager, error) {
	var repoPath string

	// Priority 1: Check for environment variable override first
	repoPath = os.Getenv("AIMGR_REPO_PATH")
	if repoPath != "" {
		m := &Manager{
			repoPath: repoPath,
		}
		m.initLogger()
		return m, nil
	}

	// Priority 2: Check config file for repo.path
	cfg, err := config.LoadGlobal()
	if err == nil && cfg.Repo.Path != "" {
		// Config loaded successfully and has repo.path set
		// Path is already validated and expanded by config.Validate()
		m := &Manager{
			repoPath: cfg.Repo.Path,
		}
		m.initLogger()
		return m, nil
	}

	// Priority 3: Fall back to XDG default
	// Ignore config errors - user may not have config file
	repoPath = filepath.Join(xdg.DataHome, "ai-config", "repo")
	m := &Manager{
		repoPath: repoPath,
	}
	m.initLogger()
	return m, nil
}

// NewManagerWithPath creates a manager with a custom repository path (for testing)
func NewManagerWithPath(repoPath string) *Manager {
	m := &Manager{
		repoPath: repoPath,
	}
	m.initLogger()
	return m
}

// initLogger initializes the logger for this Manager.
// If logger creation fails, logs a warning to stderr but continues.
// Manager can operate without logging (graceful degradation).
func (m *Manager) initLogger() {
	logger, err := logging.NewRepoLogger(m.repoPath)
	if err != nil {
		// Log warning to stderr but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
		m.logger = nil
		return
	}
	m.logger = logger
}

// GetLogger returns the logger for this Manager, or nil if logger creation failed.
func (m *Manager) GetLogger() *slog.Logger {
	return m.logger
}

// Init initializes the repository directory structure and git repository.
// This is idempotent - safe to call multiple times.
func (m *Manager) Init() error {
	// Create main repo directory
	if err := os.MkdirAll(m.repoPath, 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create repo directory",
				"path", m.repoPath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Create commands subdirectory
	commandsPath := filepath.Join(m.repoPath, "commands")
	if err := os.MkdirAll(commandsPath, 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create commands directory",
				"path", commandsPath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	// Create skills subdirectory
	skillsPath := filepath.Join(m.repoPath, "skills")
	if err := os.MkdirAll(skillsPath, 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create skills directory",
				"path", skillsPath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Create agents subdirectory
	agentsPath := filepath.Join(m.repoPath, "agents")
	if err := os.MkdirAll(agentsPath, 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create agents directory",
				"path", agentsPath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Create packages subdirectory
	packagesPath := filepath.Join(m.repoPath, "packages")
	if err := os.MkdirAll(packagesPath, 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create packages directory",
				"path", packagesPath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Initialize git repository if not already initialized
	gitDir := filepath.Join(m.repoPath, ".git")
	alreadyGit := false
	if _, err := os.Stat(gitDir); err == nil {
		alreadyGit = true
	}

	if !alreadyGit {
		gitCmd := exec.Command("git", "init")
		gitCmd.Dir = m.repoPath
		if output, err := gitCmd.CombinedOutput(); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to initialize git repository",
					"path", m.repoPath,
					"error", err.Error(),
					"output", string(output),
				)
			}
			return fmt.Errorf("failed to initialize git repository: %w\nOutput: %s", err, output)
		}
	}

	// Log repo initialization
	if m.logger != nil {
		m.logger.Info("repo init",
			"path", m.repoPath,
		)
	}

	// Create ai.repo.yaml if it doesn't exist
	// NOTE: This handles the upgrade path for existing users. When upgrading from a version
	// without ai.repo.yaml, Init() will automatically create an empty manifest on first run.
	// This is idempotent and safe - the file is only created if missing.
	// TODO(release): Document in migration guide that ai.repo.yaml is auto-created on upgrade.
	manifestPath := filepath.Join(m.repoPath, repomanifest.ManifestFileName)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Create empty manifest
		manifest := &repomanifest.Manifest{
			Version: 1,
			Sources: []*repomanifest.Source{},
		}
		if err := manifest.Save(m.repoPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to create ai.repo.yaml",
					"path", manifestPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to create ai.repo.yaml: %w", err)
		}
		// Log that manifest was created (helpful for upgrades from pre-manifest versions)
		fmt.Fprintf(os.Stderr, "Created ai.repo.yaml\n")
	} else if err != nil {
		return fmt.Errorf("failed to check ai.repo.yaml: %w", err)
	}
	// If file exists, do nothing (idempotent)

	// Create/update .gitignore (idempotent)
	gitignorePath := filepath.Join(m.repoPath, ".gitignore")
	gitignoreContent := `# aimgr workspace cache (Git clones for remote sources)
.workspace/

# Log files
logs/
*.log

# macOS
.DS_Store

# Editor files
*.swp
*.swo
*~
.vscode/
.idea/
`

	if _, err := os.Stat(gitignorePath); err == nil {
		// .gitignore exists - check if it contains .workspace/
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("failed to read .gitignore",
					"path", gitignorePath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to read .gitignore: %w", err)
		}

		// If .workspace/ is not in .gitignore, append it
		if !strings.Contains(string(content), ".workspace/") &&
			!strings.Contains(string(content), ".workspace") {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				if m.logger != nil {
					m.logger.Error("failed to open .gitignore for append",
						"path", gitignorePath,
						"error", err.Error(),
					)
				}
				return fmt.Errorf("failed to open .gitignore for append: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString("\n" + gitignoreContent); err != nil {
				if m.logger != nil {
					m.logger.Error("failed to append to .gitignore",
						"path", gitignorePath,
						"error", err.Error(),
					)
				}
				return fmt.Errorf("failed to append to .gitignore: %w", err)
			}
		}
	} else {
		// Create new .gitignore
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to create .gitignore",
					"path", gitignorePath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	// Initial commit if git was just initialized
	if !alreadyGit {
		// Add .gitignore and ai.repo.yaml
		addCmd := exec.Command("git", "add", ".gitignore", repomanifest.ManifestFileName)
		addCmd.Dir = m.repoPath
		if _, err := addCmd.CombinedOutput(); err != nil {
			// Don't fail on add error - might not have anything to add
			// Continue to try commit anyway
		}

		// Create initial commit
		commitCmd := exec.Command("git", "commit", "-m", "aimgr: initialize repository")
		commitCmd.Dir = m.repoPath
		// Don't fail on commit error - might not have anything to commit
		commitCmd.CombinedOutput()
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
	return m.addCommandWithOptions(sourcePath, sourceURL, sourceType, ref, ImportOptions{ImportMode: "copy"})
}

// addCommandWithOptions is an internal method that adds a command with import options
func (m *Manager) addCommandWithOptions(sourcePath, sourceURL, sourceType, ref string, opts ImportOptions) error {
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
	// by looking for common markers (.git, .claude, etc.)
	if basePath == "" {
		dir := filepath.Dir(cleanPath)
		// Walk up to find a suitable base, but stop before system directories
		for dir != "." && dir != "/" && dir != "/tmp" {
			// Don't go too far up - stop at directories with less than 3 segments
			// This prevents walking into /tmp, /var, etc.
			segments := strings.Split(filepath.Clean(dir), string(filepath.Separator))
			if len(segments) < 3 {
				break
			}

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
		}
	}

	res, err := resource.LoadCommandWithBase(sourcePath, basePath)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to load command",
				"source", sourcePath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to load command: %w", err)
	}

	// Get destination path
	destPath := m.GetPathForResource(res)

	// Create parent directories if needed (for nested structure)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to create parent directories",
				"path", filepath.Dir(destPath),
				"resource", res.Name,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Log before copying/symlinking
	if m.logger != nil {
		m.logger.Info("repo add",
			"resource", res.Name,
			"type", "command",
			"source", sourcePath,
		)
	}

	// Copy or symlink based on mode
	if opts.ImportMode == "symlink" {
		// Ensure source is absolute path
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("failed to get absolute path",
					"source", sourcePath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Create symlink
		if err := os.Symlink(absSource, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to create symlink",
					"resource", res.Name,
					"type", "command",
					"source", absSource,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		// Copy the file (default)
		if err := copyFile(sourcePath, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to copy command",
					"resource", res.Name,
					"source", sourcePath,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to copy command: %w", err)
		}
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
	// Use explicit sourceName from opts if provided, otherwise derive
	sourceName := opts.SourceName
	if sourceName == "" {
		sourceName = metadata.DeriveSourceName(sourceURL)
	}
	if err := metadata.Save(meta, m.repoPath, sourceName); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to save metadata",
				"resource", res.Name,
				"type", "command",
				"path", metadata.GetMetadataPath(res.Name, resource.Command, m.repoPath),
				"error", err.Error(),
			)
		}
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
	return m.addSkillWithOptions(sourcePath, sourceURL, sourceType, ref, ImportOptions{ImportMode: "copy"})
}

// addSkillWithOptions is an internal method that adds a skill with import options
func (m *Manager) addSkillWithOptions(sourcePath, sourceURL, sourceType, ref string, opts ImportOptions) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the skill
	res, err := resource.LoadSkill(sourcePath)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to load skill",
				"source", sourcePath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to load skill: %w", err)
	}

	// Get destination path
	destPath := m.GetPath(res.Name, resource.Skill)

	// Log before copying/symlinking
	if m.logger != nil {
		m.logger.Info("repo add",
			"resource", res.Name,
			"type", "skill",
			"source", sourcePath,
		)
	}

	// Copy or symlink based on mode
	if opts.ImportMode == "symlink" {
		// Ensure source is absolute path
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("failed to get absolute path",
					"source", sourcePath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Create symlink to the entire directory
		if err := os.Symlink(absSource, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to create symlink",
					"resource", res.Name,
					"type", "skill",
					"source", absSource,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		// Copy the directory (default)
		if err := copyDir(sourcePath, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to copy skill",
					"resource", res.Name,
					"source", sourcePath,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to copy skill: %w", err)
		}
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
	// Use explicit sourceName from opts if provided, otherwise derive
	sourceName := opts.SourceName
	if sourceName == "" {
		sourceName = metadata.DeriveSourceName(sourceURL)
	}
	if err := metadata.Save(meta, m.repoPath, sourceName); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to save metadata",
				"resource", res.Name,
				"type", "skill",
				"path", metadata.GetMetadataPath(res.Name, resource.Skill, m.repoPath),
				"error", err.Error(),
			)
		}
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
	return m.addAgentWithOptions(sourcePath, sourceURL, sourceType, ref, ImportOptions{ImportMode: "copy"})
}

// addAgentWithOptions is an internal method that adds an agent with import options
func (m *Manager) addAgentWithOptions(sourcePath, sourceURL, sourceType, ref string, opts ImportOptions) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the agent
	res, err := resource.LoadAgent(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load agent: %w", err)
	}

	// Get destination path
	destPath := m.GetPath(res.Name, resource.Agent)

	// Log before copying/symlinking
	if m.logger != nil {
		m.logger.Info("repo add",
			"resource", res.Name,
			"type", "agent",
			"source", sourcePath,
		)
	}

	// Copy or symlink based on mode
	if opts.ImportMode == "symlink" {
		// Ensure source is absolute path
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Create symlink
		if err := os.Symlink(absSource, destPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		// Copy the file (default)
		if err := copyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy agent: %w", err)
		}
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
	// Use explicit sourceName from opts if provided, otherwise derive
	sourceName := opts.SourceName
	if sourceName == "" {
		sourceName = metadata.DeriveSourceName(sourceURL)
	}
	if err := metadata.Save(meta, m.repoPath, sourceName); err != nil {
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
	return m.addPackageWithOptions(sourcePath, sourceURL, sourceType, ref, ImportOptions{ImportMode: "copy"})
}

// addPackageWithOptions is an internal method that adds a package with import options
func (m *Manager) addPackageWithOptions(sourcePath, sourceURL, sourceType, ref string, opts ImportOptions) error {
	// Ensure repo is initialized
	if err := m.Init(); err != nil {
		return err
	}

	// Validate and load the package
	pkg, err := resource.LoadPackage(sourcePath)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to load package",
				"source", sourcePath,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to load package: %w", err)
	}

	// Get destination path
	destPath := resource.GetPackagePath(pkg.Name, m.repoPath)

	// Log before copying/symlinking
	if m.logger != nil {
		m.logger.Info("repo add",
			"resource", pkg.Name,
			"type", "package",
			"source", sourcePath,
		)
	}

	// Copy or symlink based on mode
	if opts.ImportMode == "symlink" {
		// Ensure source is absolute path
		absSource, err := filepath.Abs(sourcePath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("failed to get absolute path",
					"source", sourcePath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Create symlink
		if err := os.Symlink(absSource, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to create symlink",
					"package", pkg.Name,
					"source", absSource,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		// Copy the file (default)
		if err := copyFile(sourcePath, destPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to copy package",
					"package", pkg.Name,
					"source", sourcePath,
					"dest", destPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to copy package: %w", err)
		}
	}

	// Create and save metadata
	now := time.Now()
	// Use explicit sourceName from opts if provided, otherwise derive
	sourceName := opts.SourceName
	if sourceName == "" {
		sourceName = metadata.DeriveSourceName(sourceURL)
	}
	pkgMeta := &metadata.PackageMetadata{
		Name:          pkg.Name,
		SourceType:    sourceType,
		SourceURL:     sourceURL,
		SourceName:    sourceName,
		SourceRef:     ref,
		FirstAdded:    now,
		LastUpdated:   now,
		ResourceCount: len(pkg.Resources),
	}
	if err := metadata.SavePackageMetadata(pkgMeta, m.repoPath); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to save package metadata",
				"package", pkg.Name,
				"path", metadata.GetPackageMetadataPath(pkg.Name, m.repoPath),
				"error", err.Error(),
			)
		}
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
				skillPath := filepath.Join(skillsPath, entry.Name())

				// Follow symlinks to check if target is a directory
				info, err := os.Stat(skillPath)
				if err != nil || !info.IsDir() {
					continue
				}

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
	Name          string `json:"name" yaml:"name"`
	Description   string `json:"description" yaml:"description"`
	ResourceCount int    `json:"resource_count" yaml:"resource_count"`
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
		// Use LoadCommandWithBase to preserve nested names
		commandsPath := filepath.Join(m.repoPath, "commands")
		return resource.LoadCommandWithBase(path, commandsPath)
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

	// Log before removing
	if m.logger != nil {
		m.logger.Info("repo remove",
			"resource", name,
			"type", string(resourceType),
		)
	}

	// Remove the resource
	if err := os.RemoveAll(path); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to remove resource",
				"resource", name,
				"type", string(resourceType),
				"path", path,
				"error", err.Error(),
			)
		}
		return fmt.Errorf("failed to remove resource: %w", err)
	}

	// Remove metadata file
	metadataPath := metadata.GetMetadataPath(name, resourceType, m.repoPath)
	if _, err := os.Stat(metadataPath); err == nil {
		if err := os.Remove(metadataPath); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to remove metadata",
					"resource", name,
					"type", string(resourceType),
					"path", metadataPath,
					"error", err.Error(),
				)
			}
			return fmt.Errorf("failed to remove metadata: %w", err)
		}
	}

	// Commit the removal
	commitMsg := fmt.Sprintf("aimgr: remove %s: %s", resourceType, name)
	if err := m.CommitChanges(commitMsg); err != nil {
		// Log warning but don't fail the operation
		fmt.Fprintf(os.Stderr, "Warning: failed to commit changes: %v\n", err)
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
	case resource.PackageType:
		return filepath.Join(m.repoPath, "packages", name+".package.json")
	default:
		return ""
	}
}

// GetPathForResource returns the full path for a resource
// For commands with nested names (e.g., "api/deploy"), creates nested directory structure
func (m *Manager) GetPathForResource(res *resource.Resource) string {
	switch res.Type {
	case resource.Command:
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

// isGitRepo checks if the repository is a git repository
func (m *Manager) isGitRepo() bool {
	gitDir := filepath.Join(m.repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// CommitChanges commits all changes in the repository with a message
// Returns nil if successful, or an error if the commit fails
// If not a git repo, returns nil (non-fatal - operations work without git)
func (m *Manager) CommitChanges(message string) error {
	if !m.isGitRepo() {
		// Not a git repo - this is not an error, just skip
		return nil
	}

	// Stage all changes (respecting .gitignore)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = m.repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to stage changes",
				"path", m.repoPath,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to stage changes: %w\nOutput: %s", err, output)
	}

	// Check if there are changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = m.repoPath
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		if m.logger != nil {
			m.logger.Error("failed to check git status",
				"path", m.repoPath,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to check git status: %w\nOutput: %s", err, output)
	}

	// If no changes, nothing to commit
	if len(output) == 0 {
		return nil
	}

	// Create commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = m.repoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to commit changes",
				"path", m.repoPath,
				"message", message,
				"error", err.Error(),
				"output", string(output),
			)
		}
		return fmt.Errorf("failed to commit changes: %w\nOutput: %s", err, output)
	}

	return nil
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

		// Follow symlinks with os.Stat
		info, err := os.Stat(srcPath)
		if err != nil {
			// Skip entries we can't stat
			continue
		}

		if info.IsDir() {
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
	SourceName   string // Explicit source name from manifest (overrides derived name)
	ImportMode   string // "copy" or "symlink"
	Force        bool   // Overwrite existing resources
	SkipExisting bool   // Skip conflicts silently
	DryRun       bool   // Preview only, don't actually import
	SourceURL    string // Original source URL (for Git sources)
	SourceType   string // Source type (github, git-url, file, local)
	Ref          string // Git ref (branch/tag/commit), defaults to "main" if empty
}

// ImportOptions contains options for single resource import operations
type ImportOptions struct {
	SourceName string // Explicit source name from manifest (overrides derived name)
	ImportMode string // "copy" or "symlink"
	Force      bool   // Overwrite existing resources
}

// ImportError represents an error during resource import
type ImportError struct {
	Path    string
	Message string
}

// BulkImportResult contains the results of a bulk import operation
type BulkImportResult struct {
	Added        []string      // Successfully added resources (new)
	Updated      []string      // Successfully updated resources (already existed, re-added with Force)
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
		Updated:      []string{},
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

	// Commit changes if not a dry run and if any resources were added/updated
	if !opts.DryRun && (len(result.Added) > 0 || len(result.Updated) > 0) {
		// Build commit message
		totalChanges := len(result.Added) + len(result.Updated)
		commitMsg := fmt.Sprintf("aimgr: import %d resource(s)", totalChanges)

		// Add detail about resource types
		details := []string{}
		if result.CommandCount > 0 {
			details = append(details, fmt.Sprintf("%d command(s)", result.CommandCount))
		}
		if result.SkillCount > 0 {
			details = append(details, fmt.Sprintf("%d skill(s)", result.SkillCount))
		}
		if result.AgentCount > 0 {
			details = append(details, fmt.Sprintf("%d agent(s)", result.AgentCount))
		}
		if result.PackageCount > 0 {
			details = append(details, fmt.Sprintf("%d package(s)", result.PackageCount))
		}
		if len(details) > 0 {
			commitMsg += " (" + strings.Join(details, ", ") + ")"
		}

		if err := m.CommitChanges(commitMsg); err != nil {
			// Log warning but don't fail the operation
			// Git tracking is optional
			fmt.Fprintf(os.Stderr, "Warning: failed to commit changes: %v\n", err)
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

		// Create import options from bulk options
		importOpts := ImportOptions{
			SourceName: opts.SourceName,
			ImportMode: opts.ImportMode,
			Force:      opts.Force,
		}
		// Default to "copy" if not specified
		if importOpts.ImportMode == "" {
			importOpts.ImportMode = "copy"
		}

		switch resourceType {
		case resource.Command:
			err = m.addCommandWithOptions(sourcePath, sourceURL, sourceType, ref, importOpts)
		case resource.Skill:
			err = m.addSkillWithOptions(sourcePath, sourceURL, sourceType, ref, importOpts)
		case resource.Agent:
			err = m.addAgentWithOptions(sourcePath, sourceURL, sourceType, ref, importOpts)
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

	// Track whether this was an update (existed before) or a new addition
	if exists && opts.Force {
		result.Updated = append(result.Updated, sourcePath)
	} else {
		result.Added = append(result.Added, sourcePath)
	}
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

		// Create import options from bulk options
		importOpts := ImportOptions{
			SourceName: opts.SourceName,
			ImportMode: opts.ImportMode,
			Force:      opts.Force,
		}
		// Default to "copy" if not specified
		if importOpts.ImportMode == "" {
			importOpts.ImportMode = "copy"
		}

		err = m.addPackageWithOptions(sourcePath, sourceURL, sourceType, ref, importOpts)
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

	// Track whether this was an update (existed before) or a new addition
	if exists && opts.Force {
		result.Updated = append(result.Updated, sourcePath)
	} else {
		result.Added = append(result.Added, sourcePath)
	}
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

// Drop removes the entire repository directory and recreates the empty structure.
// WARNING: This is a destructive operation that cannot be undone.
// NOTE: After drop, Init() is called which recreates an empty ai.repo.yaml manifest
// with no sources. This is the expected behavior for soft drop - the repository
// is ready to accept new sources via 'repo add' or 'repo sync'.
func (m *Manager) Drop() error {
	// Remove entire repo directory
	if err := os.RemoveAll(m.repoPath); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	// Recreate empty structure (including empty ai.repo.yaml)
	return m.Init()
}
