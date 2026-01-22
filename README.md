# aimgr - AI Resources Manager

A command-line tool for discovering, installing, and managing AI resources (commands and skills) across multiple AI coding tools including Claude Code, OpenCode, and GitHub Copilot.

## Features

- 📦 **Repository Management**: Centralized repository for AI commands, skills, and agents
- 🔗 **Symlink-based Installation**: Install resources in projects without duplication
- 🤖 **Multi-Tool Support**: Works with Claude Code, OpenCode, and GitHub Copilot
- 🤖 **Agent Support**: Manage AI agents with specialized roles and capabilities
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
| **[GitHub Copilot](https://github.com/features/copilot)** | ❌ | ✅ | ❌ | `.github/skills/` |

**Notes:** 
- GitHub Copilot only supports Agent Skills, not slash commands or agents.
- Agents provide specialized roles with specific capabilities for different AI tools.

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

Resources can be added from multiple sources: local files, GitHub repositories, or bulk imports from tool directories.

#### From Local Files

```bash
# Add a command (single .md file)
aimgr repo add command ~/.claude/commands/my-command.md

# Add a skill (directory with SKILL.md)
aimgr repo add skill ~/my-skills/pdf-processing

# Add an agent (single .md file)
aimgr repo add agent ~/.claude/agents/code-reviewer.md
```

#### From GitHub Repositories

```bash
# Add a skill from GitHub
aimgr repo add skill gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo add skill gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific branch or tag
aimgr repo add skill gh:anthropics/skills@v1.0.0

# Shorthand - GitHub prefix is inferred for owner/repo format
aimgr repo add skill vercel-labs/agent-skills

# Add commands from GitHub
aimgr repo add command gh:myorg/commands

# Add agents from GitHub
aimgr repo add agent gh:myorg/agents/code-reviewer
```

#### From Tool Directories (Bulk Import)

```bash
# Import from OpenCode directory
aimgr repo add opencode ~/.opencode

# Import from Claude directory
aimgr repo add claude ~/.claude
```


## Bulk Import

Import multiple resources at once from Claude plugins or Claude configuration folders.

### Import from Claude Plugin

Import all commands and skills from a Claude plugin directory:

```bash
# Import from a Claude plugin
aimgr repo add plugin ~/.claude/plugins/marketplaces/claude-plugins-official/plugins/example-plugin

# Example output:
# Importing from plugin: example-plugin
#   Description: A comprehensive example plugin
# 
# Found: 1 commands, 1 skills
# 
# ✓ Added command 'example-command'
# ✓ Added skill 'example-skill'
# 
# Summary: 2 added, 0 skipped, 0 failed
```

**Flags:**
- `--force` / `-f`: Overwrite existing resources
- `--skip-existing`: Skip conflicts silently
- `--dry-run`: Preview without importing

### Import from Claude Folder

Import all commands, skills, and agents from a `.claude/` configuration folder:

```bash
# Import from .claude folder
aimgr repo add claude ~/.claude

# Import from project's .claude folder
aimgr repo add claude ~/my-project/.claude

# Example output:
# Importing from Claude folder: /home/user/.claude
# 
# Found: 5 commands, 3 skills, 2 agents
#
# ✓ Added command 'test'
# ✓ Added command 'review'
# ✓ Added command 'deploy'
# ✓ Added skill 'pdf-processor'
# ✓ Added skill 'code-analyzer'
# ✓ Added agent 'code-reviewer'
# ⊘ Skipped 'git-release' (already exists)
# 
# Summary: 6 added, 1 skipped, 0 failed
```

**What gets imported:**
- Commands from `.claude/commands/*.md`
- Skills from `.claude/skills/*/SKILL.md`
- Agents from `.claude/agents/*.md`
- Works with both `.claude` directory and parent directories containing `.claude/`

### Import from OpenCode Folder

Import all resources from an `.opencode/` configuration folder:

```bash
# Import from .opencode folder
aimgr repo add opencode ~/.opencode

# Import from project's .opencode folder
aimgr repo add opencode ~/my-project/.opencode

# Example output:
# Importing from OpenCode folder: /home/user/.opencode
# 
# Found: 3 commands, 2 skills, 1 agents
#
# ✓ Added command 'build'
# ✓ Added command 'test'
# ✓ Added skill 'data-processor'
# ✓ Added agent 'qa-reviewer'
# 
# Summary: 4 added, 0 skipped, 0 failed
```

**What gets imported:**
- Commands from `.opencode/commands/*.md`
- Skills from `.opencode/skills/*/SKILL.md`
- Agents from `.opencode/agents/*.md`

### Handling Conflicts

Control how conflicts are handled when importing:

```bash
# Force overwrite existing resources
aimgr repo add plugin ./my-plugin --force

# Skip existing resources (no error)
aimgr repo add claude ~/.claude --skip-existing

# Preview what would be imported (dry run)
aimgr repo add plugin ./plugin --dry-run

# Default behavior: fail on first conflict
aimgr repo add plugin ./plugin  # Error if any resource already exists
```

### Bulk Import Use Cases

**Scenario 1: Setting up your aimgr from existing Claude setup**
```bash
# Import all your existing commands and skills at once
aimgr repo add claude ~/.claude
```

**Scenario 2: Installing a complete plugin**
```bash
# Import all resources from a plugin in one command
aimgr repo add plugin ~/.claude/plugins/marketplaces/claude-plugins-official/plugins/pr-review-toolkit
```

**Scenario 3: Migrating resources between machines**
```bash
# On machine A: your resources are already in aimgr
# On machine B: import from a cloned .claude folder
aimgr repo add claude ~/cloned-config/.claude
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

# Resources are symlinked to tool-specific directories
# Example: .claude/commands/, .opencode/commands/, etc.
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

- **Claude Code** and **OpenCode** support both commands and skills
- **GitHub Copilot** only supports skills (no commands)
- When installing commands to a Copilot-only project, the command is not installed
- When installing to multiple tools including Copilot, commands are installed to Claude/OpenCode only

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
aimgr repo add skill gh:owner/repo

# With specific path (for multi-resource repos)
aimgr repo add skill gh:owner/repo/path/to/skill

# With branch or tag reference
aimgr repo add skill gh:owner/repo@branch-name
aimgr repo add skill gh:owner/repo@v1.0.0

# Combined: path and reference
aimgr repo add skill gh:owner/repo/skills/my-skill@main
```

**Examples:**
```bash
# Add a skill from Vercel's agent-skills repository
aimgr repo add skill gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo add skill gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific version tag
aimgr repo add skill gh:anthropics/skills@v2.1.0
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
aimgr repo add skill local:./my-skill
aimgr repo add skill local:/absolute/path/to/skill

# Direct path (local: is implied)
aimgr repo add skill ./my-skill
aimgr repo add skill ~/my-skills/pdf-processing
```

**Note:** Local sources work exactly as before - the `local:` prefix is optional for backward compatibility.

### Git URL Sources

Add from any Git repository using full URLs:

```bash
# HTTPS URLs
aimgr repo add skill https://github.com/owner/repo.git
aimgr repo add skill https://gitlab.com/owner/repo.git

# SSH URLs
aimgr repo add skill git@github.com:owner/repo.git

# With branch reference
aimgr repo add skill https://github.com/owner/repo.git@develop
```

### Shorthand Syntax

For convenience, `aimgr` infers the `gh:` prefix for GitHub-style `owner/repo` patterns:

```bash
# These are equivalent:
aimgr repo add skill vercel-labs/agent-skills
aimgr repo add skill gh:vercel-labs/agent-skills

# With path:
aimgr repo add skill vercel-labs/agent-skills/skills/frontend-design
aimgr repo add skill gh:vercel-labs/agent-skills/skills/frontend-design
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

### `aimgr repo add`

Add resources to the repository from various sources.

```bash
# Add from GitHub (with auto-discovery)
aimgr repo add skill gh:owner/repo
aimgr repo add skill gh:owner/repo/path/to/skill
aimgr repo add skill gh:owner/repo@v1.0.0

# Add from local path
aimgr repo add command <path-to-file.md>
aimgr repo add skill <path-to-directory>
aimgr repo add agent <path-to-file.md>

# Add using shorthand (infers gh: for owner/repo)
aimgr repo add skill vercel-labs/agent-skills

# Add from Git URL
aimgr repo add skill https://github.com/owner/repo.git
aimgr repo add skill git@github.com:owner/repo.git

# Add with explicit local prefix
aimgr repo add skill local:./my-skill
aimgr repo add skill local:/absolute/path/to/skill

# Add all resources from a Claude plugin
aimgr repo add plugin <path-to-plugin>

# Add all resources from a Claude folder
aimgr repo add claude <path-to-.claude-folder>

# Add all resources from an OpenCode folder
aimgr repo add opencode <path-to-.opencode-folder>

# Overwrite existing resource
aimgr repo add command my-command.md --force

# Bulk import with conflict handling
aimgr repo add plugin ./plugin --skip-existing
aimgr repo add claude ~/.claude --dry-run
aimgr repo add opencode ~/.opencode --skip-existing
```

**Source Formats:**
- `gh:owner/repo[/path][@ref]` - GitHub repository
- `local:path` or just `path` - Local filesystem
- `https://...` or `git@...` - Any Git repository
- `owner/repo` - Shorthand for GitHub (infers `gh:` prefix)

See [Source Formats](#source-formats) for detailed documentation.

### `aimgr repo list`

List resources in the repository.

```bash
# List all
aimgr repo list

# Filter by type
aimgr repo list command
aimgr repo list skill
aimgr repo list agent

# Output formats
aimgr repo list --format=table  # Default
aimgr repo list --format=json
aimgr repo list --format=yaml
```

### `aimgr install`

Install resources to a project using type/name format.

```bash
# Install single resource
aimgr install skill/pdf-processing
aimgr install command/test
aimgr install agent/code-reviewer

# Install multiple resources at once
aimgr install skill/foo skill/bar command/test agent/reviewer

# Custom project path
aimgr install command/test --project-path ~/my-project

# Force reinstall
aimgr install skill/utils --force
aimgr install agent/code-reviewer --force

# Install to specific tool(s) - overrides defaults and existing directories
aimgr install command/test --target claude
aimgr install skill/utils --target opencode
aimgr install agent/reviewer --target claude,opencode
```

**Resource Format:** `type/name` where type is `skill`, `command`, or `agent`

**Flags:**
- `--project-path`: Specify project directory (defaults to current directory)
- `--force`: Overwrite existing installations
- `--target`: Specify which tool(s) to install to (accepts comma-separated values). Overrides both default targets and existing directory detection.

### `aimgr repo remove`

Remove resources from the repository.

```bash
# Remove with confirmation
aimgr repo remove command <name>
aimgr repo remove skill <name>
aimgr repo remove agent <name>

# Skip confirmation
aimgr repo remove command test --force

# Alias
aimgr repo rm command old-test
aimgr repo rm agent old-reviewer
```

### `aimgr uninstall`

Uninstall resources from a project (removes symlinks).

```bash
# Uninstall single resource
aimgr uninstall skill/pdf-processing
aimgr uninstall command/test
aimgr uninstall agent/code-reviewer

# Uninstall multiple resources at once
aimgr uninstall skill/foo skill/bar command/test

# Uninstall from specific project
aimgr uninstall skill/foo --project-path ~/my-project

# Force uninstall
aimgr uninstall command/review --force
```

**Resource Format:** `type/name` where type is `skill`, `command`, or `agent`

**Safety:**
- Only removes symlinks pointing to the aimgr repository
- Warns about non-symlinks or symlinks pointing elsewhere
- Automatically detects and removes from all tool directories

**Flags:**
- `--project-path`: Specify project directory (defaults to current directory)
- `--force`: Force uninstall (placeholder for future confirmation prompts)

### `aimgr repo show`

Display detailed information about a resource, including metadata and source information.

```bash
# Show skill details
aimgr repo show skill pdf-processing

# Show command details
aimgr repo show command test

# Show agent details
aimgr repo show agent code-reviewer
```

**Output includes:**
- Resource name, type, and description
- Version, author, and license information
- Source details (GitHub URL, local path, etc.)
- Metadata tracking information
- Installation status

**Use cases:**
- Check resource metadata before installing
- Verify resource source for updates
- Review resource details and documentation
- Debug installation issues

### `aimgr repo update`

Update resources from their original sources (requires metadata tracking).

```bash
# Update all resources
aimgr repo update

# Update specific skill
aimgr repo update skill pdf-processing

# Update specific command
aimgr repo update command test

# Update specific agent
aimgr repo update agent code-reviewer

# Preview updates without applying (dry run)
aimgr repo update --dry-run

# Force update, overwriting local changes
aimgr repo update --force
```

**How it works:**
1. Reads source information from resource metadata (`.aimgr-meta.yaml`)
2. Fetches latest version from original source (GitHub, local path, etc.)
3. Updates the repository copy
4. Preserves existing symlinks to projects

**Supported sources:**
- **GitHub**: Re-clones and updates from repository
- **Local**: Copies latest version from local path
- **File**: Re-copies from original file location

**Metadata tracking:**
Resources added from GitHub or other sources automatically store metadata including:
- Source type (github, local, file)
- Source URL or path
- Git reference (branch/tag)
- Added and updated timestamps

**Flags:**
- `--dry-run`: Preview what would be updated without making changes
- `--force`: Force update even if local changes detected

## Resource Formats

### Source Formats

`aimgr` supports multiple source formats for adding resources, making it easy to share and discover resources across teams and the community.

### GitHub Sources

Add resources directly from GitHub repositories using the `gh:` prefix:

```bash
# Basic syntax
aimgr repo add skill gh:owner/repo

# With specific path (for multi-resource repos)
aimgr repo add skill gh:owner/repo/path/to/skill

# With branch or tag reference
aimgr repo add skill gh:owner/repo@branch-name
aimgr repo add skill gh:owner/repo@v1.0.0

# Combined: path and reference
aimgr repo add skill gh:owner/repo/skills/my-skill@main
```

**Examples:**
```bash
# Add a skill from Vercel's agent-skills repository
aimgr repo add skill gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo add skill gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific version tag
aimgr repo add skill gh:anthropics/skills@v2.1.0
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
aimgr repo add skill local:./my-skill
aimgr repo add skill local:/absolute/path/to/skill

# Direct path (local: is implied)
aimgr repo add skill ./my-skill
aimgr repo add skill ~/my-skills/pdf-processing
```

**Note:** Local sources work exactly as before - the `local:` prefix is optional for backward compatibility.

### Git URL Sources

Add from any Git repository using full URLs:

```bash
# HTTPS URLs
aimgr repo add skill https://github.com/owner/repo.git
aimgr repo add skill https://gitlab.com/owner/repo.git

# SSH URLs
aimgr repo add skill git@github.com:owner/repo.git

# With branch reference
aimgr repo add skill https://github.com/owner/repo.git@develop
```

### Shorthand Syntax

For convenience, `aimgr` infers the `gh:` prefix for GitHub-style `owner/repo` patterns:

```bash
# These are equivalent:
aimgr repo add skill vercel-labs/agent-skills
aimgr repo add skill gh:vercel-labs/agent-skills

# With path:
aimgr repo add skill vercel-labs/agent-skills/skills/frontend-design
aimgr repo add skill gh:vercel-labs/agent-skills/skills/frontend-design
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

aimgr automatically tracks metadata about resource sources, enabling features like `repo update` and `repo show`.

### What is Tracked

When you add a resource from GitHub or other sources, aimgr stores:

- **Source type**: `github`, `local`, or `file`
- **Source location**: URL, path, or file location
- **Git reference**: Branch or tag (for GitHub sources)
- **Subpath**: Path within repository (if applicable)
- **Timestamps**: When resource was added and last updated

### Metadata File Location

Metadata is stored alongside resources in `.aimgr-meta.yaml` files:

**Skills:**
```
~/.local/share/ai-config/repo/skills/my-skill/.aimgr-meta.yaml
```

**Commands:**
```
~/.local/share/ai-config/repo/commands/.aimgr-meta/my-command.yaml
```

**Agents:**
```
~/.local/share/ai-config/repo/agents/.aimgr-meta/my-agent.yaml
```

### Metadata Format

```yaml
name: pdf-processing
type: skill
source:
  type: github
  url: https://github.com/owner/repo
  ref: main
  path: skills/pdf-processing
added: "2026-01-22T10:30:00Z"
updated: "2026-01-22T10:30:00Z"
```

**Source Types:**

| Type | Description | Example |
|------|-------------|---------|
| `github` | GitHub repository | `gh:owner/repo` or GitHub URLs |
| `local` | Local directory/file | `./my-skill` or `/path/to/skill` |
| `file` | Direct file copy | `~/commands/test.md` |

### Using Metadata

**View metadata:**
```bash
# Show detailed resource info including metadata
aimgr repo show skill pdf-processing
aimgr repo show command test
aimgr repo show agent code-reviewer
```

**Update from source:**
```bash
# Update specific resource from its original source
aimgr repo update skill pdf-processing

# Update all resources that have source metadata
aimgr repo update

# Preview what would be updated
aimgr repo update --dry-run
```

### Manual Metadata

For resources added before metadata tracking or from sources without auto-tracking, you can manually create metadata files to enable updates.

**Example: Add metadata for a local skill**

Create `~/.local/share/ai-config/repo/skills/my-skill/.aimgr-meta.yaml`:

```yaml
name: my-skill
type: skill
source:
  type: local
  path: /home/user/projects/my-skill
added: "2026-01-22T10:00:00Z"
updated: "2026-01-22T10:00:00Z"
```

Then run:
```bash
aimgr repo update skill my-skill
```

### Metadata Best Practices

1. **Add from GitHub when possible** - Enables automatic updates
2. **Use specific refs** - Tag versions (e.g., `@v1.0.0`) for stability
3. **Keep local sources accessible** - Update requires access to original path
4. **Check metadata with `repo show`** - Verify source info before updating
5. **Use `--dry-run`** - Preview updates before applying

### Metadata Privacy

Metadata files are stored locally in your repository and are **not** shared with projects when you install resources. Symlinks point directly to resource files, not metadata.

## Repository Structure

Resources are stored in `~/.local/share/ai-config/repo/` (XDG data directory):

```
~/.local/share/ai-config/repo/
├── commands/
│   ├── test.md
│   └── review.md
├── skills/
│   ├── pdf-processing/
│   │   ├── SKILL.md
│   │   └── scripts/
│   └── git-release/
│       └── SKILL.md
└── agents/
    ├── code-reviewer.md
    └── qa-tester.md
```

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

### Config Format Changed

The configuration format has changed to support multiple default installation targets:

**Old format:**
```yaml
default-tool: claude
```

**New format:**
```yaml
install:
  targets: [claude]
```

**Multiple targets:**
```yaml
install:
  targets: [claude, opencode]
```

**Migration:** Automatic. When you run any `aimgr` command, the tool will automatically convert `default-tool` to `install.targets` format. Your config file will be updated to the new format.

### Command Changes

**Old commands:**
```bash
aimgr config set default-tool claude
aimgr config get default-tool
```

**New commands:**
```bash
aimgr config set install.targets claude
aimgr config set install.targets claude,opencode  # Multiple tools
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

## Creating Resources

### Create a Command

1. Create a `.md` file with valid name: `my-command.md`
2. Add YAML frontmatter with `description`
3. Write command body in markdown
4. Test: `aimgr repo add command ./my-command.md`

### Create a Skill

1. Create directory with valid name: `my-skill/`
2. Create `SKILL.md` with frontmatter (name must match directory)
3. Optionally add `scripts/`, `references/`, `assets/`
4. Test: `aimgr repo add skill ./my-skill`

### Create an Agent

1. Create a `.md` file with valid name: `my-agent.md`
2. Add YAML frontmatter with `description`
3. Optionally add `type`, `instructions`, `capabilities` (OpenCode format)
4. Write agent documentation in markdown body
5. Test: `aimgr repo add agent ./my-agent.md`

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
- Try specifying a direct path: `aimgr repo add skill gh:owner/repo/path/to/skill`
- Check that resources have valid frontmatter (SKILL.md with name and description)
- Use the repository's documentation to find resource locations

**Problem: "Multiple resources found, please specify path"**

Solution:
- Add the specific path to your command: `aimgr repo add skill gh:owner/repo/skills/specific-skill`
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
- Use `--force` flag to overwrite: `aimgr repo add skill gh:owner/repo --force`
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
aimgr repo add skill gh:owner/repo -v
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
