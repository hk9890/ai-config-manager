# Supported Tools

aimgr supports installing AI resources to multiple AI coding tools. This reference documents which tools are supported and what resource types each tool accepts.

## Tool Support Matrix

| Tool | Commands | Skills | Agents | Directory |
|------|:--------:|:------:|:------:|-----------|
| Claude Code | Yes | Yes | Yes | `.claude/` |
| OpenCode | Yes | Yes | Yes | `.opencode/` |
| Windsurf | - | Yes | - | `.windsurf/skills/` |
| GitHub Copilot | - | Yes | - | `.github/skills/` |

**Key:**
- **Commands**: Slash commands (e.g., `/review`, `/deploy`)
- **Skills**: Agent skills that provide specialized knowledge or workflows
- **Agents**: Custom agent definitions with specific behaviors

## Tool Details

### Claude Code

Claude Code is Anthropic's AI coding assistant.

| Property | Value |
|----------|-------|
| Config Directory | `.claude/` |
| Commands Path | `.claude/commands/` |
| Skills Path | `.claude/skills/` |
| Agents Path | `.claude/agents/` |
| CLI Alias | `claude` |

**Documentation:**
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Custom Slash Commands](https://docs.anthropic.com/en/docs/claude-code/slash-commands#custom-slash-commands)

### OpenCode

OpenCode is an open-source AI coding assistant that runs in your terminal.

| Property | Value |
|----------|-------|
| Config Directory | `.opencode/` |
| Commands Path | `.opencode/commands/` |
| Skills Path | `.opencode/skills/` |
| Agents Path | `.opencode/agents/` |
| CLI Alias | `opencode` |

**Documentation:**
- [OpenCode GitHub Repository](https://github.com/sst/opencode)
- [OpenCode Skills Documentation](https://opencode.ai/docs/skills)

### Windsurf

Windsurf is Codeium's AI-powered IDE built on VSCode.

| Property | Value |
|----------|-------|
| Config Directory | `.windsurf/` |
| Skills Path | `.windsurf/skills/` |
| Commands | Not supported |
| Agents | Not supported |
| CLI Alias | `windsurf` |

**Documentation:**
- [Windsurf IDE](https://codeium.com/windsurf)

### GitHub Copilot (VSCode)

GitHub Copilot is GitHub's AI pair programmer, integrated into VSCode.

| Property | Value |
|----------|-------|
| Config Directory | `.github/` |
| Skills Path | `.github/skills/` |
| Commands | Not supported |
| Agents | Not supported |
| CLI Aliases | `copilot`, `vscode` |

**Documentation:**
- [GitHub Copilot Documentation](https://docs.github.com/en/copilot)

## Resource Formats

**Important:** Resource file formats (SKILL.md structure, frontmatter fields, markdown syntax) are defined by the tools themselves and the [AgentSkills.io](https://agentskills.io/home) skill format specification, not by aimgr.

aimgr's responsibility is:
- Managing and organizing resources in a central repository
- Installing resources to the correct tool directories via symlinks
- Syncing resources from remote sources
- Applying [field mappings](../user-guide/configuration.md#field-mappings) for tool-specific values

For resource format specifications, refer to:
- [AgentSkills.io](https://agentskills.io/home) - Community skill format specification
- Individual tool documentation (linked above)

## Configuration

### Setting Default Tools

Configure which tools to install to by default in `~/.config/aimgr/aimgr.yaml`:

```yaml
install:
  targets:
    - claude
    - opencode
```

### Per-Install Override

Override default targets for a single installation:

```bash
aimgr install skill/code-review --target windsurf
```

### Multi-Tool Installation

Install to multiple tools at once:

```bash
aimgr install skill/code-review --target claude --target opencode
```

## Auto-Detection

When installing to a project, aimgr automatically detects which tools are already configured by checking for tool-specific directories:

- `.claude/` - Claude Code detected
- `.opencode/` - OpenCode detected
- `.github/skills/` - GitHub Copilot detected
- `.windsurf/skills/` - Windsurf detected

If tool directories already exist, aimgr installs to those tools. If no tool directories exist, it uses your configured default targets.

## See Also

- [Configuration Guide](../user-guide/configuration.md) - Default target configuration
- [Field Mappings](../user-guide/configuration.md#field-mappings) - Tool-specific field transformations
