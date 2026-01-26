package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hk9890/ai-config-manager/pkg/manifest"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
	"github.com/hk9890/ai-config-manager/pkg/tools"
)

// Installer manages resource installations in a project
type Installer struct {
	projectPath string
	targetTools []tools.Tool // tools to install to
}

// NewInstaller creates a new installer for a project with default tools
func NewInstaller(projectPath string, defaultTools []tools.Tool) (*Installer, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	installer := &Installer{
		projectPath: absPath,
	}

	// Detect install targets
	targets, err := DetectInstallTargets(absPath, defaultTools)
	if err != nil {
		return nil, fmt.Errorf("detecting install targets: %w", err)
	}

	installer.targetTools = targets
	return installer, nil
}

// NewInstallerWithTargets creates a new installer with explicit target tools
// This bypasses tool detection and uses the specified targets directly
func NewInstallerWithTargets(projectPath string, targets []tools.Tool) (*Installer, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	installer := &Installer{
		projectPath: absPath,
		targetTools: targets,
	}

	return installer, nil
}

// DetectInstallTargets determines which tools to install to
// Precedence (highest to lowest):
// 1. Existing tool directories (if any exist, installs to ALL)
// 2. ai.package.yaml install.targets (if manifest exists and has targets)
// 3. Global config install.targets (defaultTools parameter)
func DetectInstallTargets(projectPath string, defaultTools []tools.Tool) ([]tools.Tool, error) {
	existingTools, err := tools.DetectExistingTools(projectPath)
	if err != nil {
		return nil, fmt.Errorf("detecting existing tools: %w", err)
	}

	// If tools are found, install to ALL of them
	if len(existingTools) > 0 {
		return existingTools, nil
	}

	// Check for ai.package.yaml manifest
	manifestPath := filepath.Join(projectPath, manifest.ManifestFileName)
	if manifest.Exists(manifestPath) {
		m, err := manifest.Load(manifestPath)
		if err != nil {
			// If manifest exists but can't be loaded, continue with defaults
			// (Don't fail the install operation due to manifest errors)
		} else if len(m.Install.Targets) > 0 {
			// Use manifest targets (overrides global config)
			manifestTargets := make([]tools.Tool, 0, len(m.Install.Targets))
			for _, targetStr := range m.Install.Targets {
				tool, err := tools.ParseTool(targetStr)
				if err != nil {
					return nil, fmt.Errorf("invalid install.targets in manifest '%s': %w", targetStr, err)
				}
				manifestTargets = append(manifestTargets, tool)
			}
			return manifestTargets, nil
		}
	}

	// Fall back to global config defaults
	return defaultTools, nil
}

// InstallCommand installs a command resource by creating symlinks to target tools
func (i *Installer) InstallCommand(name string, repoManager *repo.Manager) error {
	// Get command from repo
	res, err := repoManager.Get(name, resource.Command)
	if err != nil {
		return fmt.Errorf("command not found in repository: %w", err)
	}

	// Install to each target tool
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)

		// Skip tools that don't support commands
		if !toolInfo.SupportsCommands {
			continue
		}

		// Create commands directory if needed
		commandsDir := filepath.Join(i.projectPath, toolInfo.CommandsDir)
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return fmt.Errorf("failed to create commands directory for %s: %w", tool, err)
		}

		// Symlink path
		symlinkPath := filepath.Join(commandsDir, filepath.Base(res.Path))

		// Check if already installed
		if _, err := os.Lstat(symlinkPath); err == nil {
			// Already exists, skip this tool
			continue
		}

		// Create symlink
		if err := os.Symlink(res.Path, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", tool, err)
		}
	}

	return nil
}

// InstallSkill installs a skill resource by creating symlinks to target tools
func (i *Installer) InstallSkill(name string, repoManager *repo.Manager) error {
	// Get skill from repo
	res, err := repoManager.Get(name, resource.Skill)
	if err != nil {
		return fmt.Errorf("skill not found in repository: %w", err)
	}

	// Install to each target tool
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)

		// Skip tools that don't support skills (though all current tools do)
		if !toolInfo.SupportsSkills {
			continue
		}

		// Create skills directory if needed
		skillsDir := filepath.Join(i.projectPath, toolInfo.SkillsDir)
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory for %s: %w", tool, err)
		}

		// Symlink path
		symlinkPath := filepath.Join(skillsDir, name)

		// Check if already installed
		if _, err := os.Lstat(symlinkPath); err == nil {
			// Already exists, skip this tool
			continue
		}

		// Create symlink
		if err := os.Symlink(res.Path, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", tool, err)
		}
	}

	return nil
}

// InstallAgent installs an agent resource by creating symlinks to target tools
func (i *Installer) InstallAgent(name string, repoManager *repo.Manager) error {
	// Get agent from repo
	res, err := repoManager.Get(name, resource.Agent)
	if err != nil {
		return fmt.Errorf("agent not found in repository: %w", err)
	}

	// Install to each target tool
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)

		// Skip tools that don't support agents
		if !toolInfo.SupportsAgents {
			continue
		}

		// Create agents directory if needed
		agentsDir := filepath.Join(i.projectPath, toolInfo.AgentsDir)
		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			return fmt.Errorf("failed to create agents directory for %s: %w", tool, err)
		}

		// Symlink path
		symlinkPath := filepath.Join(agentsDir, filepath.Base(res.Path))

		// Check if already installed
		if _, err := os.Lstat(symlinkPath); err == nil {
			// Already exists, skip this tool
			continue
		}

		// Create symlink
		if err := os.Symlink(res.Path, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", tool, err)
		}
	}

	return nil
}

// Uninstall removes an installed resource by removing symlinks from all tool directories
func (i *Installer) Uninstall(name string, resourceType resource.ResourceType) error {
	removed := false
	var lastErr error

	// Try to uninstall from all target tools
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)
		var symlinkPath string

		switch resourceType {
		case resource.Command:
			if !toolInfo.SupportsCommands {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.CommandsDir, name+".md")
		case resource.Skill:
			if !toolInfo.SupportsSkills {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.SkillsDir, name)
		case resource.Agent:
			if !toolInfo.SupportsAgents {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.AgentsDir, name+".md")
		default:
			return fmt.Errorf("invalid resource type: %s", resourceType)
		}

		// Check if symlink exists
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			// Doesn't exist in this tool directory, continue
			continue
		}

		// Verify it's a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			lastErr = fmt.Errorf("'%s' in %s is not a symlink (manual installation?)", name, tool)
			continue
		}

		// Remove the symlink
		if err := os.Remove(symlinkPath); err != nil {
			lastErr = fmt.Errorf("failed to remove symlink from %s: %w", tool, err)
			continue
		}

		removed = true
	}

	if !removed {
		if lastErr != nil {
			return lastErr
		}
		return fmt.Errorf("resource '%s' is not installed", name)
	}

	return nil
}

// List lists all installed resources in the project (deduplicated across tools)
func (i *Installer) List() ([]resource.Resource, error) {
	// Use a map to deduplicate resources by name
	resourceMap := make(map[string]resource.Resource)

	// Scan all target tools
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)

		// List commands
		if toolInfo.SupportsCommands {
			commandsDir := filepath.Join(i.projectPath, toolInfo.CommandsDir)
			if _, err := os.Stat(commandsDir); err == nil {
				entries, err := os.ReadDir(commandsDir)
				if err != nil {
					return nil, fmt.Errorf("failed to read commands directory for %s: %w", tool, err)
				}

				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}

					symlinkPath := filepath.Join(commandsDir, entry.Name())
					info, err := os.Lstat(symlinkPath)
					if err != nil {
						continue
					}

					// Only list symlinks
					if info.Mode()&os.ModeSymlink == 0 {
						continue
					}

					// Read the symlink target
					target, err := os.Readlink(symlinkPath)
					if err != nil {
						continue
					}

					// Load the resource
					res, err := resource.LoadCommand(target)
					if err != nil {
						continue
					}

					// Deduplicate by name
					resourceMap[res.Name] = *res
				}
			}
		}

		// List skills
		if toolInfo.SupportsSkills {
			skillsDir := filepath.Join(i.projectPath, toolInfo.SkillsDir)
			if _, err := os.Stat(skillsDir); err == nil {
				entries, err := os.ReadDir(skillsDir)
				if err != nil {
					return nil, fmt.Errorf("failed to read skills directory for %s: %w", tool, err)
				}

				for _, entry := range entries {
					symlinkPath := filepath.Join(skillsDir, entry.Name())
					info, err := os.Lstat(symlinkPath)
					if err != nil {
						continue
					}

					// Only list symlinks
					if info.Mode()&os.ModeSymlink == 0 {
						continue
					}

					// Read the symlink target
					target, err := os.Readlink(symlinkPath)
					if err != nil {
						continue
					}

					// Verify target is a directory (skill)
					targetInfo, err := os.Stat(target)
					if err != nil || !targetInfo.IsDir() {
						continue
					}

					// Load the resource
					res, err := resource.LoadSkill(target)
					if err != nil {
						continue
					}

					// Deduplicate by name
					resourceMap[res.Name] = *res
				}
			}
		}

		// List agents
		if toolInfo.SupportsAgents {
			agentsDir := filepath.Join(i.projectPath, toolInfo.AgentsDir)
			if _, err := os.Stat(agentsDir); err == nil {
				entries, err := os.ReadDir(agentsDir)
				if err != nil {
					return nil, fmt.Errorf("failed to read agents directory for %s: %w", tool, err)
				}

				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}

					symlinkPath := filepath.Join(agentsDir, entry.Name())
					info, err := os.Lstat(symlinkPath)
					if err != nil {
						continue
					}

					// Only list symlinks
					if info.Mode()&os.ModeSymlink == 0 {
						continue
					}

					// Read the symlink target
					target, err := os.Readlink(symlinkPath)
					if err != nil {
						continue
					}

					// Load the resource
					res, err := resource.LoadAgent(target)
					if err != nil {
						continue
					}

					// Deduplicate by name
					resourceMap[res.Name] = *res
				}
			}
		}
	}

	// Convert map to slice
	resources := make([]resource.Resource, 0, len(resourceMap))
	for _, res := range resourceMap {
		resources = append(resources, res)
	}

	return resources, nil
}

// IsInstalled checks if a resource is installed in any tool directory
func (i *Installer) IsInstalled(name string, resourceType resource.ResourceType) bool {
	// Check if installed in any target tool
	for _, tool := range i.targetTools {
		toolInfo := tools.GetToolInfo(tool)
		var symlinkPath string

		switch resourceType {
		case resource.Command:
			if !toolInfo.SupportsCommands {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.CommandsDir, name+".md")
		case resource.Skill:
			if !toolInfo.SupportsSkills {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.SkillsDir, name)
		case resource.Agent:
			if !toolInfo.SupportsAgents {
				continue
			}
			symlinkPath = filepath.Join(i.projectPath, toolInfo.AgentsDir, name+".md")
		default:
			return false
		}

		// Check if symlink exists
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			continue
		}

		// Verify it's a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	}

	return false
}

// GetTargetTools returns the list of tools being targeted for installation
func (i *Installer) GetTargetTools() []tools.Tool {
	return i.targetTools
}
