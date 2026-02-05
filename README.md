# aimgr - AI Resources Manager

A command-line tool for discovering, installing, and managing AI resources (commands and skills) across multiple AI coding tools including Claude Code, OpenCode, and GitHub Copilot.

## Features

- 📦 **Repository Management**: Centralized repository for AI commands, skills, and agents
- 🔗 **Symlink-based Installation**: Install resources in projects without duplication
- 🤖 **Multi-Tool Support**: Works with Claude Code, OpenCode, and GitHub Copilot
- 🤖 **Agent Support**: Manage AI agents with specialized roles and capabilities
- ⚡ **Workspace Caching**: Git repositories cached for 10-50x faster subsequent operations
- 🧹 **Cache Management**: `repo prune` command to clean up unused Git caches
- ✅ **Format Validation**: Automatic validation of command, skill, and agent formats
- 🎯 **Type Safety**: Strong validation following agentskills.io and Claude Code specifications
- 🗂️ **Organized Storage**: Clean separation between commands, skills, and agents
- 🔧 **Smart Installation**: Automatically detects existing tool directories
- 💻 **Cross-platform**: Works on Linux and macOS (Windows support planned)
- 🔨 **Shell Completion**: Tab completion for Bash, Zsh, Fish, and PowerShell

## Supported AI Tools

`aimgr` supports three major AI coding tools:

| Tool | Commands | Skills | Agents | Directory |
|------|----------|--------|--------|-----------|
| **[Claude Code](https://code.claude.com/)** | ✅ | ✅ | ✅ | `.claude/` |
| **[OpenCode](https://opencode.ai/)** | ✅ | ✅ | ✅ | `.opencode/` |
| **[VSCode / GitHub Copilot](https://github.com/features/copilot)** | ❌ | ✅ | ❌ | `.github/skills/` |

**Notes:** 
- VSCode / GitHub Copilot only supports [Agent Skills](https://www.agentskills.io/) (via the open standard at agentskills.io), not slash commands or agents.
- Skills for Copilot use the same `SKILL.md` format as other tools.
- Use `--tool=copilot` or `--tool=vscode` when installing resources (both names work).
- Agents provide specialized roles with specific capabilities for Claude Code and OpenCode.

## Installation

### Download Binary

Download the latest release for your platform from the [Releases page](https://github.com/hk9890/ai-config-manager/releases).

**Linux (amd64)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**Linux (arm64)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**macOS (Intel)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_darwin_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**macOS (Apple Silicon)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_darwin_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**Windows (PowerShell)**:
```powershell
Invoke-WebRequest -Uri "https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_windows_amd64.zip" -OutFile "aimgr.zip"
Expand-Archive -Path "aimgr.zip" -DestinationPath "."
```

*Note: Replace `VERSION` with the actual version number (e.g., `v0.1.0`).*

### From Source

```bash
# Clone the repository
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager

# Build and install
make install

# Or just build
make build
```

### Using Go

```bash
go install github.com/hk9890/ai-config-manager@latest
```

## Shell Completion

`aimgr` supports shell completion for Bash, Zsh, Fish, and PowerShell, making it easier to discover and install resources.

### Features

- **Auto-complete resource names** when typing `aimgr install skill <TAB>`
- **Auto-complete commands** like `aimgr install command <TAB>`
- **Auto-complete agents** like `aimgr install agent <TAB>`
- **Dynamically queries your repository** to show available resources

### Setup

#### Bash

**Option 1: Load for current session**
```bash
source <(aimgr completion bash)
```

**Option 2: Load automatically for all sessions**

Linux:
```bash
aimgr completion bash > /etc/bash_completion.d/aimgr
```

macOS (with Homebrew):
```bash
aimgr completion bash > $(brew --prefix)/etc/bash_completion.d/aimgr
```

#### Zsh

**Enable completions (if not already enabled):**
```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

**Install completion:**
```bash
aimgr completion zsh > "${fpath[1]}/_aimgr"
```

Then start a new shell.

#### Fish

**Option 1: Load for current session**
```bash
aimgr completion fish | source
```

**Option 2: Load automatically for all sessions**
```bash
aimgr completion fish > ~/.config/fish/completions/aimgr.fish
```

#### PowerShell

**Option 1: Load for current session**
```powershell
aimgr completion powershell | Out-String | Invoke-Expression
```

**Option 2: Load automatically for all sessions**
```powershell
# Generate completion script
aimgr completion powershell > aimgr.ps1

# Add to your PowerShell profile
# Find profile location: $PROFILE
```

### Usage Examples

After setting up completion, you can use TAB to auto-complete resource names:

```bash
# List all available skills
aimgr install skill <TAB>
# Shows: atlassian-cli  dynatrace-api  dynatrace-control  github-docs  skill-creator

# List all available commands
aimgr install command <TAB>
# Shows: test  review  deploy  build

# List all available agents
aimgr install agent <TAB>
# Shows: code-reviewer  qa-tester  beads-task-agent
```

The completion dynamically queries your repository, so newly added resources are immediately available for completion!

### Troubleshooting

**Issue: Completion doesn't work when using `./aimgr`**

Shell completion only works for commands in your `$PATH`, not relative paths like `./aimgr`.

**Solution:**
```bash
# Make sure aimgr is in your PATH
make install  # Installs to ~/bin

# Verify ~/bin is in PATH
echo $PATH | grep -q "$HOME/bin" && echo "✓ ~/bin is in PATH" || echo "✗ ~/bin NOT in PATH"

# If not in PATH, add to ~/.bashrc (or ~/.zshrc)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Now use without ./
aimgr install skill/<TAB>  # ✓ Works
```

**Issue: Completion still not working after setup**

**Solution:**
```bash
# Restart your shell
exec bash  # or: exec zsh

# Verify completion is loaded
complete -p aimgr  # Bash: should show completion function
which _aimgr       # Zsh: should show completion function
```

**Issue: Want to use `./aimgr` during development**

**Solution:**
```bash
# Option 1: Use installed version (recommended)
make install
aimgr install skill/<TAB>

# Option 2: Create an alias
alias a='./aimgr'
a install skill/<TAB>
```

## Configuration

`aimgr` uses a configuration file at `~/.config/aimgr/aimgr.yaml` for global settings.

### Repository Path

By default, resources are stored in `~/.local/share/ai-config/repo` (XDG data directory). You can customize this location using one of three methods:

**Precedence order (highest to lowest):**
1. **`AIMGR_REPO_PATH` environment variable** - Highest priority
2. **`repo.path` in config file** - `~/.config/aimgr/aimgr.yaml`
3. **XDG default** - `~/.local/share/ai-config/repo`

#### Using Config File

Create or edit `~/.config/aimgr/aimgr.yaml`:

```yaml
repo:
  path: ~/my-custom-repo  # Supports ~ expansion and relative paths

install:
  targets:
    - claude
    - opencode
```

The `repo.path` supports:
- **Tilde expansion**: `~/custom-repo` expands to your home directory
- **Relative paths**: `./repo` converts to absolute path automatically
- **Absolute paths**: `/absolute/path/to/repo`
- **Environment variables**: `${AIMGR_REPO_PATH:-~/.local/share/ai-config/repo}`

#### Using Environment Variable

Set `AIMGR_REPO_PATH` to override all other settings:

```bash
# Temporary override for current session
export AIMGR_REPO_PATH=/path/to/custom/repo
aimgr list

# Permanent override (add to ~/.bashrc or ~/.zshrc)
echo 'export AIMGR_REPO_PATH=~/my-repo' >> ~/.bashrc
```

**Note:** The environment variable takes precedence over the config file setting.

#### Environment Variable Interpolation

Config files support Docker Compose-style environment variable interpolation with optional defaults:

```yaml
repo:
  path: ${AIMGR_REPO_PATH:-~/.local/share/ai-config/repo}

sync:
  sources:
    - url: ${SYNC_REPO:-https://github.com/hk9890/ai-tools}
      filter: ${RESOURCE_FILTER:-skill/*}
```

**Syntax:**
- `${VAR}` - Use environment variable
- `${VAR:-default}` - Use default if VAR is not set or empty

This is useful for:
- CI/CD environments with dynamic paths
- Team configurations with user-specific overrides
- Testing with temporary repositories
- Managing secrets without hardcoding credentials

For detailed configuration options, see [docs/user-guide/configuration.md](docs/user-guide/configuration.md).

## Quick Start

### Check Version

```bash
# Display version information
aimgr --version
# Output: aimgr version 0.1.0 (commit: a1b2c3d, built: 2026-01-18T19:30:00Z)

# Short form
aimgr -v
```

### 1. Configure Your Default Installation Targets

```bash
# Set your preferred AI tools (claude, opencode, or copilot)
aimgr config set install.targets claude
aimgr config set install.targets claude,opencode  # Multiple tools

# Check current setting
aimgr config get install.targets
```

### 2. Add Resources to Repository

Resources can be added from multiple sources: local files, GitHub repositories, or bulk discovery from folders.

#### Add Individual Resources

Add specific resources (type is auto-detected):

```bash
# Add a command (single .md file - auto-detected as command)
aimgr repo import ~/.claude/commands/my-command.md

# Add a skill (directory with SKILL.md - auto-detected as skill)
aimgr repo import ~/my-skills/pdf-processing

# Add an agent (single .md file - auto-detected as agent)
aimgr repo import ~/.claude/agents/code-reviewer.md
```

#### Auto-Discovery from Folders and Repositories

Auto-discover and add all resources (commands, skills, agents, packages) from a folder or GitHub repository:

```bash
# From local folders
aimgr repo import ~/.opencode/           # Discovers all resources in .opencode
aimgr repo import ~/project/.claude/     # Discovers all resources in .claude
aimgr repo import ./my-resources/        # Any folder with resources

# From GitHub
aimgr repo import gh:owner/repo          # Auto-discovers all resources in repo
aimgr repo import gh:owner/repo@v1.0.0   # Specific version
aimgr repo import owner/repo             # Shorthand (gh: inferred)

# With options
aimgr repo import ~/.opencode/ --force         # Overwrite existing
aimgr repo import ./resources/ --skip-existing # Skip conflicts
aimgr repo import ./test/ --dry-run            # Preview without importing

# Filter resources with patterns
aimgr repo import gh:owner/repo --filter "skill/*"     # Only add skills
aimgr repo import ./resources/ --filter "skill/pdf*"   # Skills starting with "pdf"
aimgr repo import ~/.claude/ --filter "*test*"         # Resources with "test" in name
aimgr repo import gh:owner/repo --filter "package/*"   # Only add packages

# Example output:
# Importing from: /home/user/.opencode
# 
# Found: 5 commands, 3 skills, 2 agents, 1 package
# 
# ✓ Added command 'test-command'
# ✓ Added command 'debug-helper'
# ...
# ✓ Added skill 'pdf-processing'
# ...
# ✓ Added agent 'code-reviewer'
# ✓ Added package 'web-dev-tools'
# 
# Summary: 11 added, 0 skipped, 0 failed
```

**How Auto-Discovery Works:**

- Searches recursively in the folder for commands (*.md), skills (*/SKILL.md), agents (*.md), and packages (*.package.json)
- Automatically detects resource types and validates them
- Handles Claude (`.claude/`), OpenCode (`.opencode/`), and GitHub Copilot (`.github/`) structures
- Discovers packages from `packages/*.package.json` files
- Skips common directories like `node_modules`, `.git`, etc.
- Supports filtering with glob patterns via `--filter` flag

#### From GitHub (Individual Resources)

Add specific resources from GitHub repositories:

```bash
# Add resources from GitHub (type auto-detected)
aimgr repo import gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo import gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific branch or tag
aimgr repo import gh:anthropics/skills@v1.0.0

# Add specific resource types using filters
aimgr repo import gh:myorg/repo --filter "command/*"
aimgr repo import gh:myorg/repo --filter "agent/*"
```

### 3. List Available Resources

```bash
# List all resources
aimgr repo list

# List only commands
aimgr repo list command

# List only skills
aimgr repo list skill

# List only agents
aimgr repo list agent

# JSON output
aimgr repo list --format=json
```

### 4. Install in a Project

```bash
cd your-project/

# Install a command
aimgr install command/my-command

# Install a skill
aimgr install skill/pdf-processing

# Install an agent
aimgr install agent/code-reviewer

# Install multiple resources at once
aimgr install skill/pdf-processing command/my-command agent/code-reviewer

# Install for specific tools
aimgr install skill/pdf-processing --tool=copilot      # VSCode/Copilot only
aimgr install skill/pdf-processing --tool=vscode       # Same as copilot
aimgr install skill/pdf-processing --tool=claude,opencode,copilot  # Multiple tools

# Use patterns to install multiple resources
aimgr install "skill/*"              # Install all skills
aimgr install "*test*"               # Install all resources with "test" in name
aimgr install "skill/pdf*"           # Install skills starting with "pdf"
aimgr install "command/test*" "agent/qa*"  # Multiple patterns

# Resources are symlinked to tool-specific directories
# Example: .claude/commands/, .opencode/commands/, .github/skills/
```

### 5. Remove Resources

```bash
# Remove from repository (with confirmation)
aimgr repo remove command my-command

# Force remove (skip confirmation)
aimgr repo remove skill old-skill --force

# Remove an agent
aimgr repo remove agent old-agent

# Uninstall from project (removes symlinks)
aimgr uninstall skill/old-skill
aimgr uninstall command/my-command agent/old-agent
```

## Multi-Tool Support

### Installation Behavior

`aimgr` intelligently handles installation based on your project's existing tool directories:

#### Scenario 1: Fresh Project (No Tool Directories)
When installing to a project with no existing tool directories, `aimgr` creates and uses your configured default installation targets:

```bash
# Set default targets
aimgr config set install.targets claude

# Install in fresh project
cd ~/my-new-project
aimgr install command/test

# Result: Creates .claude/commands/test.md

# Or install to multiple tools by default
aimgr config set install.targets claude,opencode
aimgr install command/test
# Result: Creates both .claude/commands/test.md and .opencode/commands/test.md
```

#### Scenario 2: Existing Tool Directory
When a tool directory already exists (e.g., `.opencode/`), `aimgr` installs to that directory, ignoring your default installation targets:

```bash
# Project already has .opencode directory
cd ~/existing-opencode-project
aimgr install command/test

# Result: Uses existing .opencode/commands/test.md
# (Even if your default targets are set to 'claude')
```

#### Scenario 3: Multiple Tool Directories
When multiple tool directories exist, `aimgr` installs to **ALL** of them:

```bash
# Project has both .claude and .opencode directories
cd ~/multi-tool-project
aimgr install skill/pdf-processing

# Result: Installs to BOTH:
#   - .claude/skills/pdf-processing
#   - .opencode/skills/pdf-processing
```

This ensures resources are available regardless of which tool you're using.

### Configuring Default Installation Targets

Use the `config` command to set or view your default installation targets:

```bash
# Set default target (single tool)
aimgr config set install.targets claude
aimgr config set install.targets opencode
aimgr config set install.targets copilot

# Set multiple default targets
aimgr config set install.targets claude,opencode

# View current setting
aimgr config get install.targets

# Configuration is stored in ~/.config/aimgr/aimgr.yaml
```

The `install.targets` setting controls which tool directories are created when installing to a fresh project (one without existing tool directories). You can specify multiple tools to install resources to all of them by default.

### Tool-Specific Behavior

- **Claude Code** and **OpenCode** support commands, skills, and agents
- **VSCode / GitHub Copilot** only supports skills (no commands or agents)
- Skills for Copilot follow the [Agent Skills standard](https://www.agentskills.io/) at agentskills.io
- When installing commands to a Copilot-only project, the command is not installed
- When installing to multiple tools including Copilot, commands and agents are installed to Claude/OpenCode only
- Use either `--tool=copilot` or `--tool=vscode` (both names work)

### Migration from .ai Directory

If you were using an earlier version with `.ai/` directories:

1. **Rename your directory** to match your preferred tool:
   ```bash
   mv .ai .claude    # For Claude Code
   mv .ai .opencode  # For OpenCode
   ```

2. **Or keep both** by creating symlinks:
   ```bash
   ln -s .ai .claude
   # Now both .ai and .claude point to same resources
   ```

3. **Fresh start**: Delete `.ai/` and reinstall resources - they'll use your configured default tool

## Source Formats

`aimgr` supports multiple source formats for adding resources, making it easy to share and discover resources across teams and the community.

### GitHub Sources

Add resources directly from GitHub repositories using the `gh:` prefix:

```bash
# Basic syntax
aimgr repo import skill gh:owner/repo

# With specific path (for multi-resource repos)
aimgr repo import skill gh:owner/repo/path/to/skill

# With branch or tag reference
aimgr repo import skill gh:owner/repo@branch-name
aimgr repo import skill gh:owner/repo@v1.0.0

# Combined: path and reference
aimgr repo import skill gh:owner/repo/skills/my-skill@main
```

**Examples:**
```bash
# Add a skill from Vercel's agent-skills repository
aimgr repo import skill gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo import skill gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific version tag
aimgr repo import skill gh:anthropics/skills@v2.1.0
```

**How it works:**
1. `aimgr` clones the repository to a temporary directory
2. Auto-discovers resources in standard locations (see [Auto-Discovery](#auto-discovery))
3. Copies found resources to your centralized repository (`~/.local/share/ai-config/repo/`)
4. Cleans up the temporary directory

### Local Sources

Add resources from your local filesystem using the `local:` prefix or a direct path:

```bash
# Explicit local prefix
aimgr repo import skill local:./my-skill
aimgr repo import skill local:/absolute/path/to/skill

# Direct path (local: is implied)
aimgr repo import skill ./my-skill
aimgr repo import skill ~/my-skills/pdf-processing
```

**Note:** Local sources work exactly as before - the `local:` prefix is optional for backward compatibility.

### Git URL Sources

Add from any Git repository using full URLs:

```bash
# HTTPS URLs
aimgr repo import skill https://github.com/owner/repo.git
aimgr repo import skill https://gitlab.com/owner/repo.git

# SSH URLs
aimgr repo import skill git@github.com:owner/repo.git

# With branch reference
aimgr repo import skill https://github.com/owner/repo.git@develop
```

### Shorthand Syntax

For convenience, `aimgr` infers the `gh:` prefix for GitHub-style `owner/repo` patterns:

```bash
# These are equivalent:
aimgr repo import skill vercel-labs/agent-skills
aimgr repo import skill gh:vercel-labs/agent-skills

# With path:
aimgr repo import skill vercel-labs/agent-skills/skills/frontend-design
aimgr repo import skill gh:vercel-labs/agent-skills/skills/frontend-design
```

### Auto-Discovery

When adding resources from GitHub or Git URLs, `aimgr` automatically searches for resources in standard locations:

**Skills** are searched in this priority order:
1. `SKILL.md` in the specified path (if subpath provided)
2. `skills/` directory
3. `.claude/skills/`
4. `.opencode/skills/`
5. `.github/skills/`
6. `.codex/skills/`, `.cursor/skills/`, `.goose/skills/`, etc.
7. Recursive search (max depth 5) if not found above

**Commands** are searched in:
1. `commands/` directory
2. `.claude/commands/`
3. `.opencode/commands/`
4. Recursive search for `.md` files (excluding `SKILL.md`, `README.md`)

**Agents** are searched in:
1. `agents/` directory
2. `.claude/agents/`
3. `.opencode/agents/`
4. Recursive search for `.md` files with agent frontmatter

**Packages** are searched in:
1. `packages/` directory
2. Recursive search for `*.package.json` files

**Interactive Selection:**
- If a single resource is found, it's added automatically
- If multiple resources are found, you'll be prompted to select one
- If a specific subpath is provided, exactly one resource should exist at that location

## Pattern Syntax

Many commands support glob patterns for matching multiple resources. Patterns work with `repo import --filter`, `install`, and `uninstall` commands.

### Pattern Operators

- `*` - Matches any sequence of characters
- `?` - Matches any single character
- `[abc]` - Matches any character in the set
- `{a,b}` - Matches any alternative (a or b)

### Pattern Format

Patterns can optionally specify a resource type prefix:

- `type/pattern` - Match specific type (e.g., `skill/pdf*`)
- `pattern` - Match across all types (e.g., `*test*`)
- `exact-name` - Exact match (no wildcards)

### Pattern Examples

**Adding resources with filters:**
```bash
# Add all skills from a repository
aimgr repo import gh:owner/repo --filter "skill/*"

# Add skills matching a pattern
aimgr repo import gh:owner/repo --filter "skill/pdf*"

# Add all resources with "test" in the name
aimgr repo import ./resources/ --filter "*test*"

# Add commands and agents only (no skills)
aimgr repo import ~/.opencode/ --filter "command/*" --filter "agent/*"

# Add only packages from a repository
aimgr repo import gh:owner/repo --filter "package/*"
```

**Installing resources with patterns:**
```bash
# Install all skills
aimgr install "skill/*"

# Install all test resources (any type)
aimgr install "*test*"

# Install PDF-related skills
aimgr install "skill/pdf*"

# Install multiple patterns
aimgr install "skill/pdf*" "command/test*" "agent/qa*"

# Install specific versions/variants
aimgr install "skill/*-v2"           # All v2 skills
aimgr install "command/{build,test}" # build and test commands
```

**Uninstalling with patterns:**
```bash
# Uninstall all skills
aimgr uninstall "skill/*"

# Uninstall test resources
aimgr uninstall "*test*"

# Uninstall old versions
aimgr uninstall "skill/*-old"

# Uninstall multiple patterns
aimgr uninstall "skill/legacy-*" "command/deprecated-*"
```

### Pattern Behavior

- **Exact match**: If no wildcards are present, matches exact resource name
- **Type filtering**: `type/pattern` only matches resources of that type
- **Cross-type matching**: `pattern` (without type prefix) matches across all types
- **Empty matches**: If a pattern matches zero resources, a warning is shown
- **Case-sensitive**: All pattern matching is case-sensitive

## Output Formats

`aimgr` supports multiple output formats for commands that perform bulk operations, making it easy to use both interactively and in automation scripts.

### Available Formats

| Format | Use Case | Flag |
|--------|----------|------|
| **table** | Human-readable, interactive use | `--format=table` (default) |
| **json** | Scripting, automation, CI/CD | `--format=json` |
| **yaml** | Configuration files, audit logs | `--format=yaml` |

### Table Format (Default)

Provides clean, formatted output with status indicators:

```bash
$ aimgr repo import ~/my-resources/

┌─────────┬─────────────────────┬─────────┬──────────────────────┐
│ TYPE    │ NAME                │ STATUS  │ MESSAGE              │
├─────────┼─────────────────────┼─────────┼──────────────────────┤
│ skill   │ pdf-processing      │ SUCCESS │ Added to repository  │
│ skill   │ typescript-helper   │ SUCCESS │ Added to repository  │
│ command │ test                │ SUCCESS │ Added to repository  │
│ command │ build               │ SKIPPED │ Already exists       │
│ agent   │ code-reviewer       │ SUCCESS │ Added to repository  │
└─────────┴─────────────────────┴─────────┴──────────────────────┘

Summary: 4 added, 0 failed, 1 skipped (5 total)
```

### JSON Format

Perfect for scripting and automation:

```bash
$ aimgr repo import ~/my-resources/ --format=json
{
  "added": [
    {
      "name": "pdf-processing",
      "type": "skill",
      "path": "/home/user/.local/share/ai-config/repo/skills/pdf-processing"
    },
    {
      "name": "test",
      "type": "command",
      "path": "/home/user/.local/share/ai-config/repo/commands/test.md"
    }
  ],
  "skipped": [],
  "failed": [],
  "command_count": 1,
  "skill_count": 1,
  "agent_count": 0,
  "package_count": 0
}
```

**Use with jq for powerful filtering:**

```bash
# Extract only added resource names
aimgr repo import ~/resources/ --format=json | jq '.added[].name'

# Check for failures in scripts
if [ $(aimgr repo import ~/resources/ --format=json | jq '.failed | length') -gt 0 ]; then
  echo "Import failed!"
  exit 1
fi

# Get error messages
aimgr repo import ~/resources/ --format=json | jq '.failed[] | {name, error: .message}'
```

### YAML Format

Human-readable structured output:

```bash
$ aimgr repo import ~/my-resources/ --format=yaml
added:
  - name: pdf-processing
    type: skill
    path: /home/user/.local/share/ai-config/repo/skills/pdf-processing
  - name: test
    type: command
    path: /home/user/.local/share/ai-config/repo/commands/test.md
skipped: []
failed: []
command_count: 1
skill_count: 1
agent_count: 0
package_count: 0
```

**Save results for auditing:**

```bash
# Save import log
aimgr repo import gh:myorg/resources --format=yaml > import-log.yaml

# Review later
cat import-log.yaml
```

### Error Reporting

All formats include detailed error reporting:

**Table format** shows errors inline with helpful hints:
```bash
Summary: 2 added, 1 failed, 0 skipped (3 total)

⚠ Use --format=json to see detailed error messages
```

**JSON format** provides complete error details:
```json
{
  "failed": [
    {
      "name": "broken-skill",
      "type": "skill",
      "path": "/path/to/skills/broken-skill",
      "message": "missing required field: description"
    }
  ]
}
```

### Commands Supporting Output Formats

The `--format` flag is available on these commands:

```bash
# Repository operations
aimgr repo import <source> --format=json
aimgr repo sync --format=json
aimgr repo list --format=json

# Project operations
aimgr list --format=json
```

For comprehensive documentation with scripting examples, see [docs/output-formats.md](docs/output-formats.md).

## Commands

### `aimgr config`

View and manage configuration settings.

```bash
# Get a setting
aimgr config get install.targets

# Set a setting (single tool)
aimgr config set install.targets <tool>

# Set multiple targets
aimgr config set install.targets <tool1>,<tool2>

# Valid tools: claude, opencode, copilot
# Examples:
aimgr config set install.targets claude
aimgr config set install.targets claude,opencode
```

### `aimgr repo import`

Add resources to the repository from various sources. Resource types are auto-detected from file structure and content.

```bash
# Add from GitHub (with auto-discovery)
aimgr repo import gh:owner/repo
aimgr repo import gh:owner/repo/path/to/resource
aimgr repo import gh:owner/repo@v1.0.0

# Add from local path (type auto-detected)
aimgr repo import <path-to-file.md>           # Auto-detects command or agent
aimgr repo import <path-to-directory>         # Auto-detects skill
aimgr repo import ~/.claude/commands/test.md  # Command
aimgr repo import ~/my-skills/pdf-processing  # Skill
aimgr repo import ~/.claude/agents/reviewer.md # Agent

# Add using shorthand (infers gh: for owner/repo)
aimgr repo import vercel-labs/agent-skills

# Add from Git URL
aimgr repo import https://github.com/owner/repo.git
aimgr repo import git@github.com:owner/repo.git

# Add with explicit local prefix
aimgr repo import local:./my-resource
aimgr repo import local:/absolute/path/to/resource

# Add all resources from a folder (auto-discovery)
aimgr repo import ~/.opencode/
aimgr repo import ~/project/.claude/
aimgr repo import ./my-resources/

# Filter resources during import
aimgr repo import gh:owner/repo --filter "skill/*"       # Only skills
aimgr repo import ~/.opencode/ --filter "skill/pdf*"     # Skills starting with "pdf"
aimgr repo import ./resources/ --filter "*test*"         # Resources with "test" in name

# Import options
aimgr repo import <source> --force                       # Overwrite existing
aimgr repo import <source> --skip-existing               # Skip conflicts
aimgr repo import <source> --dry-run                     # Preview without importing
```

**Source Formats:**
- `gh:owner/repo[/path][@ref]` - GitHub repository
- `local:path` or just `path` - Local filesystem
- `https://...` or `git@...` - Any Git repository
- `owner/repo` - Shorthand for GitHub (infers `gh:` prefix)

**Pattern Filtering:**
- `--filter "type/*"` - Match specific resource type (skill/*, command/*, agent/*, package/*)
- `--filter "pattern"` - Match resources by name pattern
- `--filter "*test*"` - Match any resource with "test" in name

See [Pattern Syntax](#pattern-syntax) and [Source Formats](#source-formats) for detailed documentation.

### `aimgr repo sync`

Automatically sync resources from configured sources in your global configuration file.

This command reads the `sync.sources` list from `~/.config/aimgr/aimgr.yaml` and imports all resources from each source. It's perfect for:
- Keeping your repository up-to-date with organization resources
- Automatically importing from multiple repositories
- Maintaining consistent resource sets across teams
- Replacing manual `aimgr-init` setup scripts

**Configuration File Location:** `~/.config/aimgr/aimgr.yaml`

**Configuration Format:**

```yaml
sync:
  sources:
    - url: https://github.com/anthropics/skills
    - url: gh:myorg/ai-resources@v1.0.0
      filter: "skill/*"
    - url: ~/local/resources
      filter: "*test*"
```

Each source supports:
- **url** (required): Source location (GitHub, Git URL, or local path)
  - `https://github.com/owner/repo`
  - `gh:owner/repo` or `owner/repo`
  - `gh:owner/repo@v1.0.0` (with version tag)
  - `~/local/path` or `/absolute/path`
- **filter** (optional): Glob pattern to filter resources
  - `"skill/*"` - Only skills
  - `"command/*"` - Only commands
  - `"agent/*"` - Only agents
  - `"*test*"` - Any resource with "test" in name
  - `"skill/pdf*"` - Skills starting with "pdf"

**Usage:**

```bash
# Sync all configured sources (overwrites existing)
aimgr repo sync

# Sync without overwriting existing resources
aimgr repo sync --skip-existing

# Preview what would be synced without importing
aimgr repo sync --dry-run
```

**Flags:**
- `--skip-existing`: Skip resources that already exist (default: overwrite)
- `--dry-run`: Preview without importing

**Behavior:**
- By default, existing resources are **overwritten** (force mode)
- Use `--skip-existing` to preserve existing resources and skip conflicts
- Each source is processed independently - failures in one don't affect others
- Supports all source formats: GitHub, Git URLs, and local paths
- Per-source filtering allows precise control over imported resources

**Configuration Examples:**

**Basic sync without filters:**
```yaml
sync:
  sources:
    - url: https://github.com/anthropics/skills
    - url: gh:myorg/company-resources
```

**Sync with per-source filters:**
```yaml
sync:
  sources:
    # Import all skills from anthropics
    - url: gh:anthropics/skills
      filter: "skill/*"
    
    # Import only PDF-related skills from myorg
    - url: gh:myorg/ai-tools
      filter: "skill/pdf*"
    
    # Import all test resources from local directory
    - url: ~/dev/test-resources
      filter: "*test*"
    
    # Import everything from versioned release
    - url: gh:myorg/resources@v2.1.0
```

**Mixed sources and filters:**
```yaml
sync:
  sources:
    # GitHub repos
    - url: https://github.com/vercel-labs/agent-skills
      filter: "skill/*"
    
    - url: gh:company/ai-commands@stable
      filter: "command/*"
    
    # Local paths
    - url: ~/.claude/
      filter: "agent/*"
    
    - url: ~/projects/custom-skills
      filter: "skill/*-v2"
```

**Migration from `aimgr-init` Script:**

If you previously used an `aimgr-init` setup script to populate your repository, you can replace it with the sync configuration:

**Old approach (aimgr-init script):**
```bash
#!/bin/bash
# aimgr-init - Manual setup script
aimgr repo import gh:anthropics/skills
aimgr repo import gh:myorg/commands --filter "command/*"
aimgr repo import ~/local/resources --filter "skill/*"
```

**New approach (config file):**
```yaml
# ~/.config/aimgr/aimgr.yaml
sync:
  sources:
    - url: gh:anthropics/skills
    - url: gh:myorg/commands
      filter: "command/*"
    - url: ~/local/resources
      filter: "skill/*"
```

Then simply run:
```bash
aimgr repo sync
```

**Benefits of sync over manual scripts:**
- ✅ Declarative configuration (version controllable)
- ✅ One command to update everything
- ✅ Per-source filtering
- ✅ Automatic retry and error handling
- ✅ Dry-run mode for preview
- ✅ No need to maintain shell scripts

**Example Workflow:**

```bash
# 1. Create or edit your config file
vim ~/.config/aimgr/aimgr.yaml

# 2. Add sync sources
sync:
  sources:
    - url: gh:myorg/resources
      filter: "skill/*"

# 3. Preview what will be imported
aimgr repo sync --dry-run

# 4. Run the sync
aimgr repo sync

# 5. Update existing resources later
aimgr repo sync --skip-existing  # Keep local changes
```


### `aimgr repo list`

List resources in the repository.

```bash
# List all (includes packages)
aimgr repo list

# Filter by type
aimgr repo list command
aimgr repo list skill
aimgr repo list agent
aimgr repo list package

# Output formats
aimgr repo list --format=table  # Default
aimgr repo list --format=json
aimgr repo list --format=yaml
```

### `aimgr install`

Install resources or packages to a project using type/name format.

```bash
# Install single resource
aimgr install skill/pdf-processing
aimgr install command/test
aimgr install agent/code-reviewer

# Install package (installs all resources in package)
aimgr install package/web-dev-tools
aimgr install package/testing-suite

# Install multiple resources at once
aimgr install skill/foo skill/bar command/test agent/reviewer

# Custom project path
aimgr install command/test --project-path ~/my-project
aimgr install package/my-tools --project-path ~/my-project

# Force reinstall
aimgr install skill/utils --force
aimgr install package/my-tools --force

# Install to specific tool(s) - overrides defaults and existing directories
aimgr install command/test --target claude
aimgr install package/web-dev-tools --target opencode
aimgr install agent/reviewer --target claude,opencode
```

**Resource Format:** `type/name` where type is `skill`, `command`, `agent`, or `package`

**Flags:**
- `--project-path`: Specify project directory (defaults to current directory)
- `--force`: Overwrite existing installations
- `--target`: Specify which tool(s) to install to (accepts comma-separated values). Overrides both default targets and existing directory detection.


### `aimgr list`

List resources installed in the current project directory.

```bash
# List all installed resources
aimgr list

# Filter by type
aimgr list command
aimgr list skill
aimgr list agent

# Output formats
aimgr list --format=table  # Default
aimgr list --format=json
aimgr list --format=yaml

# List in specific directory
aimgr list --path ~/my-project
aimgr list skill --path /path/to/project
```

**Key Differences from `aimgr repo list`:**
- `aimgr repo list` - Shows resources in the centralized repository
- `aimgr list` - Shows resources installed in the current/specified project

**Flags:**
- `--format`: Output format (table, json, yaml)
- `--path`: Project directory path (defaults to current directory)

**Output includes:**
- Resource type (command, skill, agent)
- Resource name
- **Target tools** (claude, opencode, copilot) - shows where each resource is installed
- Description

Only resources installed via `aimgr install` (symlinks) are shown. Manually copied files are excluded.

### `aimgr repo remove`

Remove resources or packages from the repository.

```bash
# Remove resource with confirmation
aimgr repo remove command <name>
aimgr repo remove skill <name>
aimgr repo remove agent <name>

# Remove package (keeps resources by default)
aimgr repo remove package/my-package

# Remove package and all its resources
aimgr repo remove package/my-package --with-resources

# Skip confirmation
aimgr repo remove command test --force
aimgr repo remove package/old-tools --force

# Alias
aimgr repo rm command old-test
aimgr repo rm package/old-package
```

**Package Removal:**
- By default, only the package file is removed (resources are kept)
- Use `--with-resources` to also remove all referenced resources
- Confirmation prompt is shown before removing resources

### `aimgr uninstall`

Uninstall resources or packages from a project (removes symlinks).

```bash
# Uninstall single resource
aimgr uninstall skill/pdf-processing
aimgr uninstall command/test
aimgr uninstall agent/code-reviewer

# Uninstall package (removes all resources in package)
aimgr uninstall package/web-dev-tools
aimgr uninstall package/testing-suite

# Uninstall multiple resources at once
aimgr uninstall skill/foo skill/bar command/test

# Uninstall from specific project
aimgr uninstall skill/foo --project-path ~/my-project
aimgr uninstall package/my-tools --project-path ~/my-project

# Force uninstall
aimgr uninstall command/review --force
aimgr uninstall package/my-tools --force
```

**Resource Format:** `type/name` where type is `skill`, `command`, `agent`, or `package`

**Safety:**
- Only removes symlinks pointing to the aimgr repository
- Warns about non-symlinks or symlinks pointing elsewhere
- Automatically detects and removes from all tool directories

**Flags:**
- `--project-path`: Specify project directory (defaults to current directory)
- `--force`: Force uninstall (placeholder for future confirmation prompts)

### `aimgr repo describe`

Display detailed information about a resource, including metadata and source information.

```bash
# Show skill details
aimgr repo describe skill pdf-processing

# Show command details
aimgr repo describe command test

# Show agent details
aimgr repo describe agent code-reviewer

# Output in different formats
aimgr repo describe skill pdf-processing --format=json
aimgr repo describe command test --format=yaml
aimgr repo describe agent code-reviewer --format=table  # Default
```

**Output includes:**
- Resource name, type, and description
- Version, author, and license information
- Source details (GitHub URL, local path, etc.)
- Metadata tracking information
- Installation status

**Output formats:**
- `--format=table` - Human-readable table format (default)
- `--format=json` - JSON output for scripting and automation
- `--format=yaml` - YAML output for configuration files

**Use cases:**
- Check resource metadata before installing
- Verify resource source information
- Review resource details and documentation
- Debug installation issues
- Export resource information for automation

**Note:** The `repo show` command is deprecated and will be removed in a future version. Use `repo describe` instead.

### `aimgr repo prune`

Remove unreferenced Git repository caches from `.workspace/` to free disk space.

```bash
# Remove unreferenced caches (with confirmation)
aimgr repo prune

# Preview what would be removed
aimgr repo prune --dry-run

# Force remove without confirmation
aimgr repo prune --force
```

**What gets pruned:**
- Git repository caches in `.workspace/` not used by any current resources
- Cached repositories from removed or outdated resources
- Orphaned caches from failed operations

**What is NOT pruned:**
- Caches referenced by currently installed resources
- Local file sources (not cached in `.workspace/`)
- Resource files themselves (only Git caches are removed)

**When to run prune:**
- After removing many resources from the repository
- When `.workspace/` directory grows too large
- As periodic maintenance to reclaim disk space
- After changing Git source URLs for resources

**Output:**
Shows detailed list of unreferenced caches with sizes before removal.

**Flags:**
- `--dry-run`: Preview what would be removed without removing
- `--force`: Skip confirmation prompt

## Resource Formats

### Source Formats

`aimgr` supports multiple source formats for adding resources, making it easy to share and discover resources across teams and the community.

### GitHub Sources

Add resources directly from GitHub repositories using the `gh:` prefix:

```bash
# Basic syntax
aimgr repo import skill gh:owner/repo

# With specific path (for multi-resource repos)
aimgr repo import skill gh:owner/repo/path/to/skill

# With branch or tag reference
aimgr repo import skill gh:owner/repo@branch-name
aimgr repo import skill gh:owner/repo@v1.0.0

# Combined: path and reference
aimgr repo import skill gh:owner/repo/skills/my-skill@main
```

**Examples:**
```bash
# Add a skill from Vercel's agent-skills repository
aimgr repo import skill gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo import skill gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific version tag
aimgr repo import skill gh:anthropics/skills@v2.1.0
```

**How it works:**
1. `aimgr` clones the repository to a temporary directory
2. Auto-discovers resources in standard locations (see [Auto-Discovery](#auto-discovery))
3. Copies found resources to your centralized repository (`~/.local/share/ai-config/repo/`)
4. Cleans up the temporary directory

### Local Sources

Add resources from your local filesystem using the `local:` prefix or a direct path:

```bash
# Explicit local prefix
aimgr repo import skill local:./my-skill
aimgr repo import skill local:/absolute/path/to/skill

# Direct path (local: is implied)
aimgr repo import skill ./my-skill
aimgr repo import skill ~/my-skills/pdf-processing
```

**Note:** Local sources work exactly as before - the `local:` prefix is optional for backward compatibility.

### Git URL Sources

Add from any Git repository using full URLs:

```bash
# HTTPS URLs
aimgr repo import skill https://github.com/owner/repo.git
aimgr repo import skill https://gitlab.com/owner/repo.git

# SSH URLs
aimgr repo import skill git@github.com:owner/repo.git

# With branch reference
aimgr repo import skill https://github.com/owner/repo.git@develop
```

### Shorthand Syntax

For convenience, `aimgr` infers the `gh:` prefix for GitHub-style `owner/repo` patterns:

```bash
# These are equivalent:
aimgr repo import skill vercel-labs/agent-skills
aimgr repo import skill gh:vercel-labs/agent-skills

# With path:
aimgr repo import skill vercel-labs/agent-skills/skills/frontend-design
aimgr repo import skill gh:vercel-labs/agent-skills/skills/frontend-design
```

### Auto-Discovery

When adding resources from GitHub or Git URLs, `aimgr` automatically searches for resources in standard locations:

**Skills** are searched in this priority order:
1. `SKILL.md` in the specified path (if subpath provided)
2. `skills/` directory
3. `.claude/skills/`
4. `.opencode/skills/`
5. `.github/skills/`
6. `.codex/skills/`, `.cursor/skills/`, `.goose/skills/`, etc.
7. Recursive search (max depth 5) if not found above

**Commands** are searched in:
1. `commands/` directory
2. `.claude/commands/`
3. `.opencode/commands/`
4. Recursive search for `.md` files (excluding `SKILL.md`, `README.md`)

**Agents** are searched in:
1. `agents/` directory
2. `.claude/agents/`
3. `.opencode/agents/`
4. Recursive search for `.md` files with agent frontmatter

**Interactive Selection:**
- If a single resource is found, it's added automatically
- If multiple resources are found, you'll be prompted to select one
- If a specific subpath is provided, exactly one resource should exist at that location

## Packages

Packages allow you to group related resources (commands, skills, agents) together and install them as a unit. This makes it easy to distribute collections of tools or share themed resource sets.

### What Are Packages?

A package is a JSON file that references existing resources in your repository. When you install a package, all its resources are installed together. This is useful for:

- **Themed Collections**: Group resources for specific workflows (e.g., "testing-tools", "documentation-helpers")
- **Project Templates**: Create reusable resource sets for different project types
- **Distribution**: Share curated collections of resources with your team
- **Dependency Management**: Install all needed resources for a task in one command

### Creating Packages

Packages are defined as JSON files with the format `<name>.package.json` in a `packages/` directory. Create a package file with this structure:

```json
{
  "name": "web-dev-tools",
  "description": "Web development toolkit",
  "resources": [
    "command/build",
    "skill/typescript-helper",
    "agent/code-reviewer"
  ]
}
```

Then import it:

```bash
aimgr repo import ~/my-packages/
```

**Resource Format**: Resources are specified as `type/name`:
- `command/name` - Command resource
- `skill/name` - Skill resource  
- `agent/name` - Agent resource

All resources must already exist in your repository before the package can be installed.

### Installing Packages

Install all resources in a package at once:

```bash
# Install package to current project
aimgr install package/web-dev-tools

# Install to specific project
aimgr install package/testing-suite --project-path ~/my-project

# Install to specific tool(s)
aimgr install package/my-tools --target claude
aimgr install package/my-tools --target claude,opencode
```

When you install a package:
1. Each resource in the package is installed as a symlink
2. Resources are installed to the appropriate tool directories
3. If a resource is already installed, it's skipped (unless `--force` is used)

### Uninstalling Packages

Remove all resources from a package:

```bash
# Uninstall package from current project
aimgr uninstall package/web-dev-tools

# Uninstall from specific project
aimgr uninstall package/testing-suite --project-path ~/my-project

# Force uninstall
aimgr uninstall package/my-tools --force
```

Uninstalling a package removes all its resource symlinks from the project but keeps the resources in your repository.

### Removing Packages

Remove a package from your repository:

```bash
# Remove package (keeps resources)
aimgr repo remove package/web-dev-tools

# Remove package and all its resources
aimgr repo remove package/web-dev-tools --with-resources

# Force remove without confirmation
aimgr repo remove package/old-tools --force
```

**Important**: 
- By default, removing a package only deletes the package file, not the resources
- Use `--with-resources` to also remove all referenced resources
- You'll be prompted for confirmation before removing resources

### Listing Packages

View available packages in your repository:

```bash
# List all resources (includes packages section)
aimgr repo list

# List only packages
aimgr repo list package

# JSON output
aimgr repo list package --format=json
```

### Package File Format

Packages are stored as JSON files in `~/.local/share/ai-config/repo/packages/`:

```json
{
  "name": "web-dev-tools",
  "description": "Web development toolkit",
  "resources": [
    "command/build",
    "skill/typescript-helper",
    "agent/code-reviewer"
  ]
}
```

**Fields:**
- `name` (string, required): Package name (must match agentskills.io naming rules)
- `description` (string, required): Human-readable description
- `resources` (array, required): List of resource references in `type/name` format

See [examples/packages/](examples/packages/) for complete examples.

### Package Use Cases

**1. Project Setup:**
```bash
# Create a package file: react-starter.package.json
# {
#   "name": "react-starter",
#   "description": "React development essentials",
#   "resources": ["command/dev", "command/build", "skill/react-helper", "agent/react-reviewer"]
# }

# Import and install in new React project
aimgr repo import ~/packages/
cd my-react-app
aimgr install package/react-starter
```

**2. Testing Workflow:**
```bash
# Create test-tools.package.json with testing resources
# Then import and install
aimgr repo import ~/packages/
aimgr install package/test-tools
```

**3. Documentation:**
```bash
# Create docs-tools.package.json with documentation resources
# Then import and install
aimgr repo import ~/packages/
aimgr install package/docs-tools
```

## Marketplace Import

Import Claude plugin marketplaces to quickly bootstrap your repository with curated collections of resources. Marketplace files are automatically discovered and imported when you use `aimgr repo import`.

### What is Marketplace Import?

Claude plugin marketplaces use a `marketplace.json` file to define collections of plugins with commands, skills, and agents. When `aimgr repo import` detects a marketplace file, it automatically:

- **Parses** the marketplace.json file
- **Discovers** resources in each plugin's source directory  
- **Creates** aimgr packages for each plugin
- **Imports** all resources into your repository
- **Tracks** metadata for future updates

This makes it easy to import entire plugin ecosystems in one command.

### Basic Usage

The marketplace import feature is integrated into `aimgr repo import`. When you add a directory containing a `marketplace.json` file, it's automatically detected and processed:

```bash
# Import from local directory with marketplace
aimgr repo import ~/my-marketplace/

# Import from GitHub repository with marketplace
aimgr repo import gh:myorg/plugins

# Preview without importing
aimgr repo import ~/my-marketplace/ --dry-run
```

### Auto-Discovery

`aimgr repo import` automatically searches for marketplace files in standard locations:

**Marketplace Discovery:**
1. `marketplace.json` in the root directory
2. `.claude-plugin/marketplace.json` (Claude plugin convention)
3. Recursive search (max depth 3)

When found, the marketplace is automatically imported along with any other resources in the directory.

### Command Options

Use the same flags as `aimgr repo import`:

**Flags:**
- `--dry-run`: Preview what would be imported without making changes
- `--force, -f`: Overwrite existing packages and resources
- `--filter <pattern>`: Import only plugins matching the glob pattern
- `--skip-existing`: Skip existing resources

**Examples:**

```bash
# Import marketplace with all its plugins
aimgr repo import ~/claude-plugins/

# Import specific plugins only
aimgr repo import ~/claude-plugins/ --filter "web-*"

# Force overwrite existing packages
aimgr repo import gh:myorg/plugins --force

# Preview what would be imported
aimgr repo import ~/marketplace/ --dry-run
```

### How It Works

When `aimgr repo import` finds a marketplace:

1. **Parse**: Reads the marketplace.json file
2. **Filter**: Applies optional pattern filters to select specific plugins
3. **Discover**: For each plugin, searches for resources in standard locations:
   - Commands: `commands/*.md`, `.claude/commands/*.md`, `.opencode/commands/*.md`
   - Skills: `skills/*/SKILL.md`, `.claude/skills/*/SKILL.md`, `.opencode/skills/*/SKILL.md`
   - Agents: `agents/*.md`, `.claude/agents/*.md`, `.opencode/agents/*.md`
4. **Import**: Copies all discovered resources into your repository
5. **Package**: Creates an aimgr package for each plugin
6. **Track**: Saves metadata for future updates

### Marketplace Format

A Claude marketplace.json file has this structure:

```json
{
  "name": "my-marketplace",
  "version": "1.0.0",
  "description": "Collection of useful plugins",
  "owner": {
    "name": "Organization Name",
    "email": "contact@example.com"
  },
  "plugins": [
    {
      "name": "web-dev-tools",
      "description": "Web development toolkit",
      "source": "./plugins/web-dev-tools",
      "category": "development",
      "version": "1.0.0",
      "author": {
        "name": "Developer Name",
        "email": "dev@example.com"
      }
    },
    {
      "name": "testing-suite",
      "description": "Complete testing toolkit",
      "source": "./plugins/testing-suite",
      "category": "testing"
    }
  ]
}
```

**Required fields:**
- `name`: Marketplace name
- `description`: Marketplace description
- `plugins`: Array of plugin definitions

**Plugin fields:**
- `name` (required): Plugin name (becomes package name)
- `description` (required): Plugin description
- `source` (required): Relative path to plugin resources
- `category` (optional): Plugin category
- `version` (optional): Plugin version
- `author` (optional): Plugin author info

See [examples/marketplace/](examples/marketplace/) for complete examples.

### Use Cases

**1. Organization Plugin Distribution:**
```bash
# Create marketplace for your organization's plugins
# Distribute repository to team members via GitHub
# They import with one command:
aimgr repo import gh:myorg/plugins

# All plugins become packages they can install:
aimgr install package/web-dev-tools
aimgr install package/testing-suite
```

**2. Public Plugin Collections:**
```bash
# Import community plugin collections
aimgr repo import gh:community/awesome-claude-plugins

# Browse imported packages
aimgr repo list package

# Install what you need
aimgr install package/pdf-tools package/git-helpers
```

**3. Project Bootstrapping:**
```bash
# Clone project with marketplace
git clone https://github.com/myorg/project
cd project

# Import project-specific plugins
aimgr repo import ./.claude-plugin/

# All project resources available immediately
aimgr install package/backend-tools package/frontend-tools
```

### Metadata and Updates

Imported packages and resources are tracked with metadata:

```bash
# View package metadata
aimgr repo describe package/web-dev-tools
```

Metadata includes:
- Source URL (marketplace location)
- Original format (marketplace)
- Import timestamp
- Resource count

### Package Management After Import

After importing a marketplace, packages work like regular aimgr packages:

```bash
# Install marketplace package
aimgr install package/web-dev-tools

# List packages
aimgr repo list package

# Remove package (keeps resources)
aimgr repo remove package/web-dev-tools

# Remove package and resources
aimgr repo remove package/web-dev-tools --with-resources

# Uninstall from project
aimgr uninstall package/web-dev-tools
```

### Pattern Filtering

Use glob patterns to import specific plugins:

```bash
# Import only web-related plugins
aimgr repo import ~/marketplace/ --filter "web-*"

# Import testing plugins
aimgr repo import gh:myorg/plugins --filter "*-test"

# Import multiple patterns (multiple commands)
aimgr repo import ~/marketplace/ --filter "code-*"
aimgr repo import ~/marketplace/ --filter "dev-*"
```

### Integration with repo import

Marketplace import is seamlessly integrated into `aimgr repo import`:

- **Automatic detection**: No special command needed
- **Works with all sources**: Local paths and GitHub repositories
- **Same flags**: `--dry-run`, `--force`, `--filter`, `--skip-existing`
- **Unified workflow**: Import resources and marketplaces together
- **Preserves structure**: Plugin organization becomes package structure

**Example output:**
```bash
$ aimgr repo import ~/my-plugins/ --dry-run

Importing from: /home/user/my-plugins
  Mode: DRY RUN (preview only)

Found: 4 commands, 2 skills, 2 agents, 0 packages
Found marketplace: my-plugins/marketplace.json (3 plugins)

Generating packages from marketplace:
  ✓ web-dev-tools (4 resources)
  ✓ testing-suite (4 resources)
  ✓ docs-helpers (2 resources)

✓ Added command 'build'
✓ Added command 'test'
✓ Added skill 'typescript-helper'
✓ Added agent 'code-reviewer'
...

Summary: 16 added, 0 skipped, 0 failed
```
## Commands

Commands are single `.md` files with YAML frontmatter, following the Claude Code slash command format.

**Minimum format:**
```yaml
---
description: What this command does
---

# Command body

Your instructions here.
```

**Full format:**
```yaml
---
description: Run tests with coverage
agent: build
model: anthropic/claude-3-5-sonnet-20241022
allowed-tools:
  - bash
  - read
---

# Run Tests

Command instructions...
```

See [examples/sample-command.md](examples/sample-command.md) for a complete example.

**Specification:** https://code.claude.com/docs/en/slash-commands

### Skills

Skills are directories containing a `SKILL.md` file plus optional subdirectories, following the agentskills.io specification.

**Minimum structure:**
```
my-skill/
└── SKILL.md
```

**Full structure:**
```
my-skill/
├── SKILL.md           # Required: metadata + documentation
├── README.md          # Optional
├── scripts/           # Optional: executable scripts
├── references/        # Optional: reference docs
└── assets/            # Optional: images, etc.
```

**SKILL.md format:**
```yaml
---
name: my-skill         # Must match directory name
description: What this skill does
license: MIT
metadata:
  author: your-name
  version: "1.0.0"
---

# Skill documentation

Your skill details here.
```

See [examples/sample-skill/](examples/sample-skill/) for a complete example.

**Specification:** https://agentskills.io/specification

### Agents

Agents are single `.md` files with YAML frontmatter that define AI agents with specialized roles and capabilities. They support both OpenCode and Claude Code formats.

**Minimum format:**
```yaml
---
description: What this agent does
---

# Agent documentation

Your agent details here.
```

**OpenCode format (with type and instructions):**
```yaml
---
description: Code review agent that checks for best practices
type: code-reviewer
instructions: Review code for quality, security, and performance
capabilities:
  - static-analysis
  - security-scan
  - performance-review
version: "1.0.0"
author: your-name
license: MIT
---

# Code Reviewer Agent

Detailed documentation about the agent's role and behavior.
```

**Claude format (instructions in body):**
```yaml
---
description: Test automation agent
version: "2.0.0"
author: team-name
license: Apache-2.0
metadata:
  category: testing
  tags: qa,automation
---

# Test Automation Agent

This agent specializes in creating and maintaining test suites.

## Guidelines

- Write comprehensive test cases
- Follow testing best practices
- Ensure good coverage
```

See [examples/sample-agent.md](examples/sample-agent.md) for a complete example.

**Specifications:**
- OpenCode: https://opencode.ai/docs/agents
- Claude Code: https://code.claude.com/docs/agents

## Name Validation

Commands, skills, and agents must follow these naming rules:

- **Length:** 1-64 characters
- **Characters:** Lowercase letters (a-z), numbers (0-9), hyphens (-)
- **Start/End:** Must start and end with alphanumeric
- **Hyphens:** No consecutive hyphens

**Valid:** `test`, `run-coverage`, `pdf-processing`, `skill-v2`  
**Invalid:** `Test`, `test_coverage`, `-test`, `test--cmd`

## Metadata Tracking

aimgr automatically tracks metadata about resource sources, enabling features like `repo describe`.

```bash
aimgr repo describe skill pdf-processing
aimgr repo describe command test
aimgr repo describe agent code-reviewer
```
~/.local/share/ai-config/repo/.metadata/skills/my-skill-metadata.json
```

**Commands:**
```
~/.local/share/ai-config/repo/.metadata/commands/my-command-metadata.json
```

**Agents:**
```
~/.local/share/ai-config/repo/.metadata/agents/my-agent-metadata.json
```

This centralized structure keeps metadata separate from resource files and makes it easier to manage and back up.

### Metadata Format

Metadata files are stored in JSON format for better performance and tooling support:

```json
{
  "name": "pdf-processing",
  "type": "skill",
  "source_type": "github",
  "source_url": "https://github.com/owner/repo",
  "first_installed": "2026-01-22T10:30:00Z",
  "last_updated": "2026-01-22T10:30:00Z"
}
```

**Source Types:**

| Type | Description | Example |
|------|-------------|---------|
| `github` | GitHub repository | `gh:owner/repo` or GitHub URLs |
| `local` | Local directory | `./my-skill` or `/path/to/skill` |
| `file` | Direct file copy | `~/commands/test.md` |

### Using Metadata

**View metadata:**
```bash
# Show detailed resource info including metadata
aimgr repo describe skill pdf-processing
aimgr repo describe command test
aimgr repo describe agent code-reviewer
```

### Metadata Best Practices

1. **Add from GitHub when possible** - Enables better source tracking
2. **Use specific refs** - Tag versions (e.g., `@v1.0.0`) for stability
3. **Keep local sources accessible** - Maintain access to original paths
4. **Check metadata with `repo describe`** - Verify source info

### Metadata Privacy

Metadata files are stored locally in your repository and are **not** shared with projects when you install resources. Symlinks point directly to resource files, not metadata.

## Repository Structure

Resources are stored in `~/.local/share/ai-config/repo/` (XDG data directory):

```
~/.local/share/ai-config/repo/
├── .metadata/                   # Centralized metadata storage (JSON)
│   ├── commands/
│   │   ├── test-metadata.json
│   │   └── review-metadata.json
│   ├── skills/
│   │   ├── pdf-processing-metadata.json
│   │   └── git-release-metadata.json
│   └── agents/
│       ├── code-reviewer-metadata.json
│       └── qa-tester-metadata.json
├── commands/                    # Command resources
│   ├── test.md
│   └── review.md
├── skills/                      # Skill resources
│   ├── pdf-processing/
│   │   ├── SKILL.md
│   │   └── scripts/
│   └── git-release/
│       └── SKILL.md
└── agents/                      # Agent resources
    ├── code-reviewer.md
    └── qa-tester.md
```

The `.metadata/` directory contains JSON files with source tracking information (GitHub URLs, local paths, timestamps).

## Project Installation

When you install resources in a project, symlinks are created in tool-specific directories:

### Claude Code Project
```
your-project/
└── .claude/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    ├── skills/
    │   └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
    └── agents/
        └── code-reviewer.md -> ~/.local/share/ai-config/repo/agents/code-reviewer.md
```

### OpenCode Project
```
your-project/
└── .opencode/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    ├── skills/
    │   └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
    └── agents/
        └── code-reviewer.md -> ~/.local/share/ai-config/repo/agents/code-reviewer.md
```

### GitHub Copilot Project (Skills Only)
```
your-project/
└── .github/
    └── skills/
        └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
```

### Multi-Tool Project
```
your-project/
├── .claude/
│   ├── commands/
│   │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
│   ├── skills/
│   │   └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
│   └── agents/
│       └── code-reviewer.md -> ~/.local/share/ai-config/repo/agents/code-reviewer.md
└── .opencode/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    ├── skills/
    │   └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
    └── agents/
        └── code-reviewer.md -> ~/.local/share/ai-config/repo/agents/code-reviewer.md
```

The tool automatically detects existing tool directories and installs to all of them, ensuring resources are available in whichever tool you're using.

## Migrating from v0.2.0

If you're upgrading from v0.2.0 or earlier, there are two key changes:

### Config Location Changed

**Old location:** `~/.aimgr.yaml`  
**New location:** `~/.config/aimgr/aimgr.yaml`

**Migration:** Automatic on first run. The tool will detect your old config file and copy it to the new location. Your old file is left intact for safety.

### Config Format

The configuration format supports multiple default installation targets:

**Format:**
```yaml
install:
  targets: [claude]
```

**Multiple targets:**
```yaml
install:
  targets: [claude, opencode]
```

**Commands:**
```bash
# Set single target
aimgr config set install.targets claude

# Set multiple targets
aimgr config set install.targets claude,opencode

# Get current targets
aimgr config get install.targets
```

### New Feature: --target Flag

You can now override the installation target for individual operations:

```bash
# Install to specific tool(s), ignoring defaults
aimgr install command test --target claude
aimgr install skill utils --target claude,opencode
```

This is useful when you want to install a resource to a specific tool without changing your global configuration.

### Metadata Structure Migration

If you're upgrading from a version with the old metadata structure (`.aimgr-meta.yaml` files), the metadata files need to be migrated to the new centralized `.metadata/` directory structure.

**Old structure:**
```
~/.local/share/ai-config/repo/skills/my-skill/.aimgr-meta.yaml
~/.local/share/ai-config/repo/commands/.aimgr-meta/my-command.yaml
```

**New structure:**
```
~/.local/share/ai-config/repo/.metadata/skills/my-skill-metadata.json
~/.local/share/ai-config/repo/.metadata/commands/my-command-metadata.json
```

**Migration command (Removed in v1.4.0):**

**Note:** The `aimgr repo migrate-metadata` command has been removed as all repositories have been successfully migrated. If you have an older repository, you will need to use aimgr v1.3.x to perform the migration.

```bash
# Historical command (v1.3.x only):
# Preview migration without making changes
# aimgr repo migrate-metadata --dry-run

# Run migration (prompts for confirmation)
# aimgr repo migrate-metadata

# Force migration without confirmation
# aimgr repo migrate-metadata --force
```

The migration command (v1.3.x) performed:
1. Found all old metadata files (`.aimgr-meta.yaml`)
2. Converted YAML format to JSON
3. Moved files to the new `.metadata/` directory structure
4. Cleaned up old metadata files after successful migration

**Benefits of new structure:**
- Centralized location for all metadata
- Better organization and easier backup
- Faster lookup and processing
- JSON format for better tooling support

## Migration Guide: Command Simplification

In recent versions, aimgr unified its command structure to be simpler and more consistent. The old type-specific subcommands have been replaced with pattern-based filtering.

### Command Changes Summary

| Old Command | New Command | Notes |
|-------------|-------------|-------|
| `aimgr repo import bulk <source>` | `aimgr repo import <source>` | Auto-discovery is now the default behavior |
| `aimgr repo import command <file>` | `aimgr repo import <file>` | Type auto-detected from file content |
| `aimgr repo import skill <dir>` | `aimgr repo import <dir>` | Type auto-detected from SKILL.md |
| `aimgr repo import agent <file>` | `aimgr repo import <file>` | Type auto-detected from frontmatter |
| N/A | `aimgr repo import <source> --filter "type/*"` | New: Filter resources during import |

### Before and After Examples

**Adding all resources from a folder:**
```bash
# Old way
aimgr repo import bulk ~/.opencode/

# New way (exactly the same)
aimgr repo import ~/.opencode/
```

**Adding only skills from a repository:**
```bash
# Old way (not possible - needed multiple commands)
# 1. Clone repo manually
# 2. Add each skill individually

# New way
aimgr repo import gh:owner/repo --filter "skill/*"
```

**Adding a single resource:**
```bash
# Old way
aimgr repo import command ~/my-command.md
aimgr repo import skill ~/my-skill/
aimgr repo import agent ~/my-agent.md

# New way (type auto-detected)
aimgr repo import ~/my-command.md
aimgr repo import ~/my-skill/
aimgr repo import ~/my-agent.md
```

**Installing resources with patterns:**
```bash
# Old way (not possible - needed multiple commands)
aimgr install skill/pdf-processing
aimgr install skill/pdf-converter
aimgr install skill/pdf-merger

# New way
aimgr install "skill/pdf*"
```

**Uninstalling multiple resources:**
```bash
# Old way (one at a time)
aimgr uninstall skill/test-skill
aimgr uninstall command/test-command
aimgr uninstall agent/test-agent

# New way
aimgr uninstall "*test*"
```

### What's Improved

✅ **Simpler**: One `add` command instead of multiple type-specific commands  
✅ **More powerful**: Pattern matching for batch operations  
✅ **Auto-detection**: Resource types detected automatically  
✅ **Filtering**: Add only specific resources from large repos  
✅ **Backward compatible**: Old workflows still work with new commands

### Migration Checklist

- ✅ Replace `repo import bulk` with `repo import` (both work)
- ✅ Replace `repo import command/skill/agent` with `repo import` (type auto-detected)
- ✅ Use `--filter` flag to selectively import resources
- ✅ Use patterns with `install` and `uninstall` for batch operations
- ✅ Update scripts and documentation to use new syntax

All old commands continue to work - no breaking changes!

## Creating Resources

### Create a Command

1. Create a `.md` file with valid name: `my-command.md`
2. Add YAML frontmatter with `description`
3. Write command body in markdown
4. Test: `aimgr repo import command ./my-command.md`

### Create a Skill

1. Create directory with valid name: `my-skill/`
2. Create `SKILL.md` with frontmatter (name must match directory)
3. Optionally add `scripts/`, `references/`, `assets/`
4. Test: `aimgr repo import skill ./my-skill`

### Create an Agent

1. Create a `.md` file with valid name: `my-agent.md`
2. Add YAML frontmatter with `description`
3. Optionally add `type`, `instructions`, `capabilities` (OpenCode format)
4. Write agent documentation in markdown body
5. Test: `aimgr repo import agent ./my-agent.md`

See [examples/README.md](examples/README.md) for detailed instructions.

## Development

For information on building, testing, and contributing to aimgr, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup and workflow
- Code style guidelines
- Testing requirements
- Architecture overview

For quick questions or discussions, visit our [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Claude Code for command format specification
- agentskills.io for skill format specification
- Cobra for CLI framework
- XDG Base Directory specification

## Support

- Issues: https://github.com/hk9890/ai-config-manager/issues
- Discussions: https://github.com/hk9890/ai-config-manager/discussions

## Troubleshooting

### GitHub Source Issues

**Problem: "Git clone failed" error**

Solution:
- Ensure `git` is installed: `git --version`
- Check internet connectivity
- Verify repository URL is correct and accessible
- For private repositories, ensure you have SSH keys or credentials configured

**Problem: "No resources found in repository"**

Solution:
- Verify the repository contains resources in standard locations
- Try specifying a direct path: `aimgr repo import skill gh:owner/repo/path/to/skill`
- Check that resources have valid frontmatter (SKILL.md with name and description)
- Use the repository's documentation to find resource locations

**Problem: "Multiple resources found, please specify path"**

Solution:
- Add the specific path to your command: `aimgr repo import skill gh:owner/repo/skills/specific-skill`
- List available resources by cloning the repo manually: `git clone https://github.com/owner/repo && ls -R`

**Problem: Network timeout or slow clones**

Solution:
- Check your internet connection
- For large repositories, consider cloning manually first, then using `local:` source
- Repository is cloned with `--depth 1` (shallow) for speed, but may still be large

### Local Source Issues

**Problem: "Path does not exist"**

Solution:
- Verify the path is correct: `ls <path>`
- Use absolute paths to avoid confusion: `/home/user/skills/my-skill`
- Check for typos in the path

**Problem: "Directory must contain SKILL.md"**

Solution:
- Ensure your skill directory has a `SKILL.md` file (case-sensitive)
- Verify SKILL.md has valid YAML frontmatter with at least `name` and `description`

**Problem: "Folder name does not match skill name in SKILL.md"**

Solution:
- Rename the directory to match the `name` field in SKILL.md frontmatter
- Or update the `name` field in SKILL.md to match the directory name

### Installation Issues

**Problem: Symlink creation fails**

Solution:
- Ensure you have write permissions in the project directory
- On Windows, run as administrator (symlinks require elevated permissions)
- Check that the repository path is accessible: `ls ~/.local/share/ai-config/repo/`

**Problem: "Resource already exists"**

Solution:
- Use `--force` flag to overwrite: `aimgr repo import skill gh:owner/repo --force`
- Or remove the existing resource first: `aimgr repo remove skill <name>`
- Check what's installed: `aimgr list`

### Configuration Issues

**Problem: "Config file not found" or invalid config**

Solution:
- Config location: `~/.config/aimgr/aimgr.yaml`
- Reset config: Delete the file and run `aimgr config set install.targets claude`
- Check config syntax: `cat ~/.config/aimgr/aimgr.yaml`

**Problem: Resources installing to wrong tool directory**

Solution:
- Check your default targets: `aimgr config get install.targets`
- Set desired targets: `aimgr config set install.targets claude,opencode`
- Use `--target` flag to override: `aimgr install skill name --target claude`

### General Debugging

**Enable verbose output:**
```bash
# Most commands support verbose flags (check with --help)
aimgr repo import skill gh:owner/repo -v
```

**Check repository contents:**
```bash
# List all resources in your repository
aimgr repo list

# Check repository directory directly
ls -la ~/.local/share/ai-config/repo/skills/
ls -la ~/.local/share/ai-config/repo/commands/
ls -la ~/.local/share/ai-config/repo/agents/
```

**Verify Git installation:**
```bash
git --version
# Should output: git version 2.x.x or higher
```

**Check permissions:**
```bash
# Repository directory
ls -ld ~/.local/share/ai-config/repo/

# Project directories
ls -ld .claude/ .opencode/ .github/skills/
```

## Roadmap

- [x] GitHub source support with auto-discovery
- [x] Shell completion
- [ ] Windows support (junction instead of symlinks)
- [ ] Search functionality for resources
- [ ] GitLab source support
- [ ] Resource versioning
- [ ] Update/upgrade commands
