package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Tool represents an AI coding tool that supports commands and/or skills
type Tool int

const (
	// Claude represents Claude Code (supports commands and skills)
	Claude Tool = iota
	// OpenCode represents OpenCode (supports commands and skills)
	OpenCode
	// Copilot represents GitHub Copilot (supports skills only)
	Copilot
	// VSCode is an alias for Copilot (GitHub Copilot in VSCode)
	VSCode = Copilot
)

// String returns the string representation of a Tool
func (t Tool) String() string {
	switch t {
	case Claude:
		return "claude"
	case OpenCode:
		return "opencode"
	case Copilot: // VSCode is an alias, same value
		return "copilot"
	default:
		return "unknown"
	}
}

// ParseTool converts a string to a Tool type
func ParseTool(s string) (Tool, error) {
	switch strings.ToLower(s) {
	case "claude":
		return Claude, nil
	case "opencode":
		return OpenCode, nil
	case "copilot", "vscode":
		return Copilot, nil
	default:
		return -1, fmt.Errorf("unknown tool: %s (must be: claude, opencode, copilot, or vscode)", s)
	}
}

// ToolInfo contains directory path information for a specific tool
type ToolInfo struct {
	// Name is the human-readable name of the tool
	Name string
	// CommandsDir is the project-level directory for commands (empty if not supported)
	CommandsDir string
	// SkillsDir is the project-level directory for skills
	SkillsDir string
	// AgentsDir is the project-level directory for agents (empty if not supported)
	AgentsDir string
	// SupportsCommands indicates whether this tool supports slash commands
	SupportsCommands bool
	// SupportsSkills indicates whether this tool supports agent skills
	SupportsSkills bool
	// SupportsAgents indicates whether this tool supports agents
	SupportsAgents bool
}

// GetToolInfo returns the directory path information for a given tool
func GetToolInfo(tool Tool) ToolInfo {
	switch tool {
	case Claude:
		return ToolInfo{
			Name:             "Claude Code",
			CommandsDir:      ".claude/commands",
			SkillsDir:        ".claude/skills",
			AgentsDir:        ".claude/agents",
			SupportsCommands: true,
			SupportsSkills:   true,
			SupportsAgents:   true,
		}
	case OpenCode:
		return ToolInfo{
			Name:             "OpenCode",
			CommandsDir:      ".opencode/commands",
			SkillsDir:        ".opencode/skills",
			AgentsDir:        ".opencode/agents",
			SupportsCommands: true,
			SupportsSkills:   true,
			SupportsAgents:   true,
		}
	case Copilot: // VSCode is an alias for Copilot
		return ToolInfo{
			Name:             "GitHub Copilot / VSCode",
			CommandsDir:      "", // Not supported
			SkillsDir:        ".github/skills",
			AgentsDir:        "", // Not supported
			SupportsCommands: false,
			SupportsSkills:   true,
			SupportsAgents:   false,
		}
	default:
		return ToolInfo{}
	}
}

// DetectExistingTools scans a project directory for existing tool configuration directories
// and returns a list of detected tools.
// It checks for the presence of tool-specific directories like .claude, .opencode, .github
func DetectExistingTools(projectPath string) ([]Tool, error) {
	var detected []Tool

	// Check for Claude (.claude directory)
	claudePath := filepath.Join(projectPath, ".claude")
	if exists, err := dirExists(claudePath); err != nil {
		return nil, fmt.Errorf("checking .claude directory: %w", err)
	} else if exists {
		detected = append(detected, Claude)
	}

	// Check for OpenCode (.opencode directory)
	opencodePath := filepath.Join(projectPath, ".opencode")
	if exists, err := dirExists(opencodePath); err != nil {
		return nil, fmt.Errorf("checking .opencode directory: %w", err)
	} else if exists {
		detected = append(detected, OpenCode)
	}

	// Check for GitHub Copilot (.github/skills directory)
	// Note: We check for .github/skills specifically, not just .github,
	// because .github is commonly used for GitHub Actions and other purposes
	copilotPath := filepath.Join(projectPath, ".github", "skills")
	if exists, err := dirExists(copilotPath); err != nil {
		return nil, fmt.Errorf("checking .github/skills directory: %w", err)
	} else if exists {
		detected = append(detected, Copilot)
	}

	return detected, nil
}

// dirExists checks if a directory exists at the given path
func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// AllTools returns a slice containing all supported tools
func AllTools() []Tool {
	return []Tool{Claude, OpenCode, Copilot}
}
