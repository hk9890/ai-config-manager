package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PluginMetadata represents the metadata from a Claude plugin
type PluginMetadata struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Version     string       `json:"version,omitempty"`
	Author      PluginAuthor `json:"author,omitempty"`
	Homepage    string       `json:"homepage,omitempty"`
	Repository  string       `json:"repository,omitempty"`
	License     string       `json:"license,omitempty"`
}

// PluginAuthor represents plugin author information
type PluginAuthor struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// DetectPlugin checks if the given path contains a valid Claude plugin structure
// Returns true if the path contains .claude-plugin/plugin.json
func DetectPlugin(path string) (bool, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path does not exist: %w", err)
	}

	// Must be a directory
	if !info.IsDir() {
		return false, nil
	}

	// Check for .claude-plugin/plugin.json
	pluginJsonPath := filepath.Join(path, ".claude-plugin", "plugin.json")
	if _, err := os.Stat(pluginJsonPath); err != nil {
		return false, nil
	}

	return true, nil
}

// LoadPluginMetadata loads and parses the plugin.json metadata file
func LoadPluginMetadata(path string) (*PluginMetadata, error) {
	// Validate it's a plugin directory
	isPlugin, err := DetectPlugin(path)
	if err != nil {
		return nil, err
	}
	if !isPlugin {
		return nil, fmt.Errorf("not a valid plugin directory: missing .claude-plugin/plugin.json")
	}

	// Read plugin.json
	pluginJsonPath := filepath.Join(path, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(pluginJsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	// Parse JSON
	var metadata PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	// Validate required fields
	if metadata.Name == "" {
		return nil, fmt.Errorf("plugin.json missing required field: name")
	}

	return &metadata, nil
}

// ScanPluginResources scans a plugin directory for commands, skills, and agents
func ScanPluginResources(path string) (commandPaths []string, skillPaths []string, agentPaths []string, err error) {
	// Validate it's a plugin
	isPlugin, err := DetectPlugin(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if !isPlugin {
		return nil, nil, nil, fmt.Errorf("not a valid plugin directory")
	}

	// Scan commands/ directory
	commandsDir := filepath.Join(path, "commands")
	if info, err := os.Stat(commandsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(commandsDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to read commands directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".md" {
				commandPaths = append(commandPaths, filepath.Join(commandsDir, entry.Name()))
			}
		}
	}

	// Scan skills/ directory
	skillsDir := filepath.Join(path, "skills")
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to read skills directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			// Check if directory contains SKILL.md
			skillPath := filepath.Join(skillsDir, entry.Name())
			skillMdPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMdPath); err == nil {
				skillPaths = append(skillPaths, skillPath)
			}
		}
	}

	// Scan agents/ directory
	agentsDir := filepath.Join(path, "agents")
	if info, err := os.Stat(agentsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(agentsDir)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to read agents directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".md" {
				agentPaths = append(agentPaths, filepath.Join(agentsDir, entry.Name()))
			}
		}
	}

	return commandPaths, skillPaths, agentPaths, nil
}

// ClaudeFolderContents represents the contents found in a Claude folder
type ClaudeFolderContents struct {
	CommandPaths []string
	SkillPaths   []string
	AgentPaths   []string
}

// DetectClaudeFolder checks if the given path is a Claude configuration folder
// Returns true if the path is named .claude or contains commands/ or skills/ subdirectories
func DetectClaudeFolder(path string) (bool, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path does not exist: %w", err)
	}

	// Must be a directory
	if !info.IsDir() {
		return false, nil
	}

	// Check if directory is named .claude
	baseName := filepath.Base(path)
	if baseName == ".claude" {
		return true, nil
	}

	// Check for .claude subdirectory
	claudeSubDir := filepath.Join(path, ".claude")
	if info, err := os.Stat(claudeSubDir); err == nil && info.IsDir() {
		return true, nil
	}

	// Check if has commands/ or skills/ subdirectories (indicating it might be a Claude folder)
	commandsDir := filepath.Join(path, "commands")
	skillsDir := filepath.Join(path, "skills")

	hasCommands := false
	hasSkills := false

	if info, err := os.Stat(commandsDir); err == nil && info.IsDir() {
		hasCommands = true
	}
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		hasSkills = true
	}

	// If it has both commands and skills, it's likely a Claude folder
	return hasCommands || hasSkills, nil
}

// scanToolFolder is a shared helper that scans a tool's configuration folder for
// commands, skills, and agents. It handles resolving the actual tool directory
// (either the path itself or a subdirectory named toolDirName).
func scanToolFolder(path, toolDirName string, commandPaths, skillPaths, agentPaths *[]string) error {
	// Determine the actual tool directory
	toolPath := path
	baseName := filepath.Base(path)
	if baseName != toolDirName {
		// Check if there's a tool subdirectory
		toolSubDir := filepath.Join(path, toolDirName)
		if info, err := os.Stat(toolSubDir); err == nil && info.IsDir() {
			toolPath = toolSubDir
		}
	}

	// Scan commands/ directory
	commandsDir := filepath.Join(toolPath, "commands")
	if info, err := os.Stat(commandsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(commandsDir)
		if err != nil {
			return fmt.Errorf("failed to read commands directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".md" {
				*commandPaths = append(*commandPaths, filepath.Join(commandsDir, entry.Name()))
			}
		}
	}

	// Scan skills/ directory
	skillsDir := filepath.Join(toolPath, "skills")
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			return fmt.Errorf("failed to read skills directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			// Check if directory contains SKILL.md
			skillPath := filepath.Join(skillsDir, entry.Name())
			skillMdPath := filepath.Join(skillPath, "SKILL.md")
			if _, err := os.Stat(skillMdPath); err == nil {
				*skillPaths = append(*skillPaths, skillPath)
			}
		}
	}

	// Scan agents/ directory
	agentsDir := filepath.Join(toolPath, "agents")
	if info, err := os.Stat(agentsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(agentsDir)
		if err != nil {
			return fmt.Errorf("failed to read agents directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) == ".md" {
				*agentPaths = append(*agentPaths, filepath.Join(agentsDir, entry.Name()))
			}
		}
	}

	return nil
}

// ScanClaudeFolder scans a Claude folder for commands and skills
func ScanClaudeFolder(path string) (*ClaudeFolderContents, error) {
	// Validate it's a Claude folder
	isClaudeFolder, err := DetectClaudeFolder(path)
	if err != nil {
		return nil, err
	}
	if !isClaudeFolder {
		return nil, fmt.Errorf("not a valid Claude folder")
	}

	contents := &ClaudeFolderContents{
		CommandPaths: []string{},
		SkillPaths:   []string{},
		AgentPaths:   []string{},
	}

	if err := scanToolFolder(path, ".claude", &contents.CommandPaths, &contents.SkillPaths, &contents.AgentPaths); err != nil {
		return nil, err
	}

	return contents, nil
}

// OpenCodeFolderContents represents the contents found in an OpenCode folder
type OpenCodeFolderContents struct {
	CommandPaths []string
	SkillPaths   []string
	AgentPaths   []string
}

// DetectOpenCodeFolder checks if the given path is an OpenCode configuration folder
// Returns true if the path is named .opencode or contains commands/, skills/, or agents/ subdirectories
func DetectOpenCodeFolder(path string) (bool, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path does not exist: %w", err)
	}

	// Must be a directory
	if !info.IsDir() {
		return false, nil
	}

	// Check if directory is named .opencode
	baseName := filepath.Base(path)
	if baseName == ".opencode" {
		return true, nil
	}

	// Check for .opencode subdirectory
	opencodeSubDir := filepath.Join(path, ".opencode")
	if info, err := os.Stat(opencodeSubDir); err == nil && info.IsDir() {
		return true, nil
	}

	// Check if has commands/, skills/, or agents/ subdirectories (indicating it might be an OpenCode folder)
	commandsDir := filepath.Join(path, "commands")
	skillsDir := filepath.Join(path, "skills")
	agentsDir := filepath.Join(path, "agents")

	hasCommands := false
	hasSkills := false
	hasAgents := false

	if info, err := os.Stat(commandsDir); err == nil && info.IsDir() {
		hasCommands = true
	}
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		hasSkills = true
	}
	if info, err := os.Stat(agentsDir); err == nil && info.IsDir() {
		hasAgents = true
	}

	// If it has any of commands, skills, or agents, it's likely an OpenCode folder
	return hasCommands || hasSkills || hasAgents, nil
}

// ScanOpenCodeFolder scans an OpenCode folder for commands, skills, and agents
func ScanOpenCodeFolder(path string) (*OpenCodeFolderContents, error) {
	// Validate it's an OpenCode folder
	isOpenCodeFolder, err := DetectOpenCodeFolder(path)
	if err != nil {
		return nil, err
	}
	if !isOpenCodeFolder {
		return nil, fmt.Errorf("not a valid OpenCode folder")
	}

	contents := &OpenCodeFolderContents{
		CommandPaths: []string{},
		SkillPaths:   []string{},
		AgentPaths:   []string{},
	}

	if err := scanToolFolder(path, ".opencode", &contents.CommandPaths, &contents.SkillPaths, &contents.AgentPaths); err != nil {
		return nil, err
	}

	return contents, nil
}
