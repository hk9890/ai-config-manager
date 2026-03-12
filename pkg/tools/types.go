package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Tool represents an AI coding target as modeled by aimgr.
//
// Note: these values encode aimgr's current direct-install contract, not every
// upstream customization feature a tool may expose. For example, GitHub
// Copilot/VS Code now supports workspace custom agents and prompt files, but
// aimgr intentionally does not model prompt-file installs for the
// copilot/vscode target.
type Tool int

const (
	// Claude represents Claude Code (supports commands and skills)
	Claude Tool = iota
	// OpenCode represents OpenCode (supports commands and skills)
	OpenCode
	// Copilot represents GitHub Copilot / VS Code.
	//
	// Validated upstream conventions:
	//   - skills: .github/skills/<name>/SKILL.md
	//   - custom agents: .github/agents/*.agent.md
	//   - VS Code prompt files (slash commands): .github/prompts/*.prompt.md
	//
	// aimgr supports skills and agents for this target. Prompt-file installs
	// (.github/prompts/*.prompt.md) remain intentionally unsupported.
	Copilot
	// Windsurf represents Windsurf IDE (supports skills only)
	Windsurf
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
	case Windsurf:
		return "windsurf"
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
	case "windsurf":
		return Windsurf, nil
	default:
		return -1, fmt.Errorf("unknown tool: %s (must be: claude, opencode, copilot, vscode, or windsurf)", s)
	}
}

// ToolInfo contains the directories and direct-install capabilities that aimgr
// currently models for a specific tool target.
type ToolInfo struct {
	// Name is the human-readable name of the tool
	Name string
	// CommandsDir is the project-level directory that aimgr installs command
	// resources into (empty if the current aimgr target does not support direct
	// command installation).
	CommandsDir string
	// SkillsDir is the project-level directory that aimgr installs skills into.
	SkillsDir string
	// AgentsDir is the project-level directory that aimgr installs agent
	// resources into (empty if the current aimgr target does not support direct
	// agent installation).
	AgentsDir string
	// SupportsCommands indicates whether aimgr currently supports direct
	// installation of its command resource type for this tool target.
	SupportsCommands bool
	// SupportsSkills indicates whether aimgr currently supports direct
	// installation of Agent Skills for this tool target.
	SupportsSkills bool
	// SupportsAgents indicates whether aimgr currently supports direct
	// installation of its agent resource type for this tool target.
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
			CommandsDir:      "", // VS Code uses .github/prompts/*.prompt.md; not yet modeled by aimgr
			SkillsDir:        ".github/skills",
			AgentsDir:        ".github/agents",
			SupportsCommands: false,
			SupportsSkills:   true,
			SupportsAgents:   true,
		}
	case Windsurf:
		return ToolInfo{
			Name:             "Windsurf",
			CommandsDir:      "",
			SkillsDir:        ".windsurf/skills",
			AgentsDir:        "",
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

	// Check for GitHub Copilot (.github/skills OR .github/agents).
	copilotSkillsPath := filepath.Join(projectPath, ".github", "skills")
	skillsExist, err := dirExists(copilotSkillsPath)
	if err != nil {
		return nil, fmt.Errorf("checking .github/skills directory: %w", err)
	}

	copilotAgentsPath := filepath.Join(projectPath, ".github", "agents")
	agentsExist, err := dirExists(copilotAgentsPath)
	if err != nil {
		return nil, fmt.Errorf("checking .github/agents directory: %w", err)
	}

	if skillsExist || agentsExist {
		detected = append(detected, Copilot)
	}

	// Check for Windsurf (.windsurf/skills directory)
	windsurfPath := filepath.Join(projectPath, ".windsurf", "skills")
	if exists, err := dirExists(windsurfPath); err != nil {
		return nil, fmt.Errorf("checking .windsurf/skills directory: %w", err)
	} else if exists {
		detected = append(detected, Windsurf)
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
	return []Tool{Claude, OpenCode, Copilot, Windsurf}
}

// AgentArtifactName maps a logical agent name to its installed filename for a tool.
func AgentArtifactName(tool Tool, logicalName string) string {
	if tool == Copilot {
		return logicalName + ".agent.md"
	}
	return logicalName + ".md"
}

// AgentLogicalName maps an installed agent filename back to its logical name for a tool.
// Returns false if the filename does not match the tool's agent artifact convention.
func AgentLogicalName(tool Tool, artifactName string) (string, bool) {
	if tool == Copilot {
		if !strings.HasSuffix(artifactName, ".agent.md") {
			return "", false
		}
		return strings.TrimSuffix(artifactName, ".agent.md"), true
	}

	if !strings.HasSuffix(artifactName, ".md") {
		return "", false
	}
	return strings.TrimSuffix(artifactName, ".md"), true
}

// AgentArtifactNameForToolName maps a logical agent name to its installed filename
// using a tool name string (e.g. "copilot", "claude"). Unknown tool names fall
// back to the default <name>.md behavior.
func AgentArtifactNameForToolName(toolName, logicalName string) string {
	tool, err := ParseTool(toolName)
	if err != nil {
		return logicalName + ".md"
	}
	return AgentArtifactName(tool, logicalName)
}

// AgentLogicalNameForToolName maps an installed agent filename back to logical name
// using a tool name string. Returns false if parsing fails or filename is invalid.
func AgentLogicalNameForToolName(toolName, artifactName string) (string, bool) {
	tool, err := ParseTool(toolName)
	if err != nil {
		return "", false
	}
	return AgentLogicalName(tool, artifactName)
}
