# AI Tool Directory Structure Research

Research conducted on January 18, 2026, documenting where major AI coding tools store their commands and skills.

## Summary

This document provides directory paths and structure information for three AI coding tools:
- **Claude Code**: Comprehensive support for commands, skills, and plugins
- **OpenCode**: Open-source AI coding agent with commands and skills support
- **GitHub Copilot**: Support for Agent Skills (standardized format)

---

## 1. Claude Code

**Official Documentation**: https://code.claude.com/docs

Claude Code has extensive support for custom commands, skills, agents, and plugins with a well-defined directory structure.

### Commands (Slash Commands)

**Purpose**: Frequently used prompts that can be invoked with `/command-name`

**Storage Locations**:
- **Personal (User)**: `~/.claude/commands/`
  - Available across all projects
- **Project**: `.claude/commands/`
  - Stored in repository, shared with team

**File Format**: Markdown files with `.md` extension

**Example File** (`.claude/commands/test.md`):
```markdown
---
description: Run tests with coverage
agent: build
model: anthropic/claude-3-5-sonnet-20241022
allowed-tools:
  - bash
  - read
---

# Run Tests

Run the full test suite with coverage report and show any failures.
Focus on the failing tests and suggest fixes.
```

**Features**:
- Support for arguments (`$ARGUMENTS`, `$1`, `$2`, etc.)
- Bash command execution with `!` prefix
- File references with `@` prefix
- YAML frontmatter for configuration

---

### Skills (Agent Skills)

**Purpose**: Automatically discovered capabilities that Claude applies when relevant to the task

**Storage Locations**:
- **Personal (User)**: `~/.claude/skills/`
- **Project**: `.claude/skills/`

**File Structure**: Each skill is a directory containing `SKILL.md` (required) plus optional supporting files

**Example Structure**:
```
~/.claude/skills/
└── pdf-processing/
    ├── SKILL.md
    ├── README.md
    ├── scripts/
    └── references/
```

**SKILL.md Example**:
```markdown
---
name: pdf-processing
description: Process PDF files and extract data
license: MIT
metadata:
  author: your-name
  version: "1.0.0"
---

# PDF Processing Skill

Instructions for processing PDF files...
```

---

## 2. OpenCode

**Official Documentation**: https://opencode.ai/docs

OpenCode is an open-source AI coding agent with support for custom commands and Agent Skills.

### Commands

**Purpose**: Custom commands for repetitive tasks

**Storage Locations**:
- **Global**: `~/.config/opencode/commands/`
- **Project Config**: `.opencode/commands/`
- **Claude-compatible**: `.claude/commands/` (also supported)

**File Format**: Markdown files with `.md` extension

**Example File** (`.opencode/commands/test.md`):
```markdown
---
description: Run tests with coverage
agent: build
model: anthropic/claude-3-5-sonnet-20241022
---

Run the full test suite with coverage report and show any failures.
Focus on the failing tests and suggest fixes.
```

**Features**:
- Support for arguments (`$ARGUMENTS`, `$1`, `$2`, `$3`, etc.)
- Shell command output injection with `!`command`` syntax
- File references with `@filename`
- Optional agent, model, subtask configuration
- Can override built-in commands

**Naming Rules**:
- Length: 1-64 characters
- Lowercase alphanumeric with single hyphen separators
- Cannot start or end with `-`
- No consecutive hyphens

---

### Skills (Agent Skills)

**Purpose**: Reusable behavior loaded on-demand via the native `skill` tool

**Storage Locations**:
- **Global**: `~/.config/opencode/skills/<name>/SKILL.md`
- **Project Config**: `.opencode/skills/<name>/SKILL.md`  
- **Claude-compatible (Global)**: `~/.claude/skills/<name>/SKILL.md`
- **Claude-compatible (Project)**: `.claude/skills/<name>/SKILL.md`

**Discovery**: OpenCode walks up from current directory to git worktree, loading all matching `SKILL.md` files.

**File Structure**: One folder per skill name with `SKILL.md` inside

**Example Structure**:
```
.opencode/skills/
└── git-release/
    └── SKILL.md

~/.config/opencode/skills/
└── pdf-processing/
    └── SKILL.md
```

**SKILL.md Format**:
```markdown
---
name: git-release
description: Create consistent releases and changelogs
license: MIT
compatibility: opencode
metadata:
  audience: maintainers
  workflow: github
---

## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
Ask clarifying questions if the target versioning scheme is unclear.
```

**Required Frontmatter**:
- `name` (required) - Must match directory name
- `description` (required) - 1-1024 characters

**Optional Frontmatter**:
- `license`
- `compatibility`
- `metadata` (string-to-string map)

**Naming Rules** (same as commands):
- 1-64 characters
- Lowercase alphanumeric with single hyphen separators
- Regex: `^[a-z0-9]+(-[a-z0-9]+)*$`

**Permissions**: Can control skill access per agent using pattern-based permissions in `opencode.json`:
```json
{
  "permission": {
    "skill": {
      "*": "allow",
      "internal-*": "deny",
      "experimental-*": "ask"
    }
  }
}
```

---

## 3. GitHub Copilot

**Official Documentation**: https://docs.github.com/en/copilot

GitHub Copilot supports Agent Skills using the standardized Agent Skills format (an open standard shared with Claude Code).

### Commands (Prompts)

**Note**: GitHub Copilot does NOT have a slash command system like Claude Code or OpenCode. Custom prompts are not currently supported in the same way.

**Status**: Not supported (or called "prompts" - requires further clarification)

---

### Agent Skills

**Purpose**: Folders of instructions, scripts, and resources that Copilot can load when relevant

**Storage Locations**:
- **Project**: `.github/skills/` or `.claude/skills/`
- **Personal**: `~/.copilot/skills/` or `~/.claude/skills/`
  - Note: Personal skills only supported in Copilot coding agent and GitHub Copilot CLI (not VS Code stable yet)

**File Structure**: Each skill is a directory containing `SKILL.md`

**Example Structure**:
```
.github/skills/
└── github-actions-failure-debugging/
    └── SKILL.md

~/.copilot/skills/
└── webapp-testing/
    ├── SKILL.md
    └── test-templates/
        └── integration.test.js
```

**SKILL.md Format** (compatible with Claude Code and OpenCode):
```markdown
---
name: github-actions-failure-debugging
description: Guide for debugging failing GitHub Actions workflows. Use this when asked to debug failing GitHub Actions workflows.
license: MIT
---

To debug failing GitHub Actions workflows:

1. Use the `list_workflow_runs` tool to look up recent runs
2. Use the `summarize_job_log_failures` tool for AI summary
3. Use `get_job_logs` for full detailed logs if needed
4. Reproduce the failure in your environment
5. Fix the failing build
```

**Limitations**:
- No custom slash commands
- Personal skills not yet in VS Code stable
- Skills are the primary customization mechanism

---

## Directory Paths Quick Reference

### Claude Code

**User Level**:
- Commands: `~/.claude/commands/`
- Skills: `~/.claude/skills/`

**Project Level**:
- Commands: `.claude/commands/`
- Skills: `.claude/skills/`

---

### OpenCode

**User Level (XDG-compliant)**:
- Commands: `~/.config/opencode/commands/`
- Skills: `~/.config/opencode/skills/<name>/SKILL.md`

**User Level (Claude-compatible)**:
- Commands: `~/.claude/commands/`
- Skills: `~/.claude/skills/<name>/SKILL.md`

**Project Level**:
- Commands: `.opencode/commands/`
- Skills: `.opencode/skills/<name>/SKILL.md`

**Project Level (Claude-compatible)**:
- Commands: `.claude/commands/`
- Skills: `.claude/skills/<name>/SKILL.md`

---

### GitHub Copilot

**User Level**:
- Skills: `~/.copilot/skills/<name>/SKILL.md` or `~/.claude/skills/<name>/SKILL.md`

**Project Level**:
- Skills: `.github/skills/<name>/SKILL.md` or `.claude/skills/<name>/SKILL.md`

---

## Key Differences Summary

| Feature | Claude Code | OpenCode | GitHub Copilot |
|---------|-------------|----------|----------------|
| **Commands** | ✅ `~/.claude/commands/`, `.claude/commands/` | ✅ `~/.config/opencode/commands/`, `.opencode/commands/` | ❌ Not supported |
| **Skills** | ✅ `~/.claude/skills/`, `.claude/skills/` | ✅ `~/.config/opencode/skills/`, `.opencode/skills/` | ✅ `~/.copilot/skills/`, `.github/skills/` |
| **Command Format** | Markdown with YAML frontmatter | Markdown with YAML frontmatter | N/A |
| **Skill Format** | Agent Skills standard (SKILL.md) | Agent Skills standard (SKILL.md) | Agent Skills standard (SKILL.md) |
| **Cross-compatible** | Yes (OpenCode reads `.claude/`) | Yes (reads `.claude/` and `.opencode/`) | Limited (only skills) |

---

## Recommendations for ai-config-manager

Based on this research, the tool should:

1. **Support 3 Tools**:
   - Claude Code (`.claude/`)
   - OpenCode (`.opencode/` and `~/.config/opencode/`)
   - GitHub Copilot (`.github/` and `~/.copilot/`) - skills only

2. **Directory Priority**:
   - For **commands**: Support `.claude/commands/` and `.opencode/commands/`
   - For **skills**: Support all: `.claude/skills/`, `.opencode/skills/`, `.github/skills/`, `~/.copilot/skills/`

3. **Format Compatibility**:
   - Use Agent Skills standard for skills (SKILL.md) - compatible with all three tools
   - Use Claude/OpenCode markdown format for commands

4. **Installation Logic**:
   - Detect existing tool directories (.claude, .opencode, .github)
   - Install to appropriate tool-specific directories
   - Allow user to configure default tool

5. **User-level vs Project-level**:
   - Repository in `~/.local/share/ai-config/repo/` (or XDG data directory)
   - Install creates symlinks to tool-specific directories

---

## Sources

- Claude Code Documentation: https://code.claude.com/docs
- OpenCode Documentation: https://opencode.ai/docs
  - Commands: https://opencode.ai/docs/commands
  - Skills: https://opencode.ai/docs/skills
- GitHub Copilot Documentation: https://docs.github.com/en/copilot
- Agent Skills Open Standard: https://github.com/agentskills/agentskills
