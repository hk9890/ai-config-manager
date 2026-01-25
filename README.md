# aimgr - AI Resources Manager

A command-line tool for discovering, installing, and managing AI resources (commands and skills) across multiple AI coding tools including Claude Code, OpenCode, and GitHub Copilot.

## Features

- 📦 **Repository Management**: Centralized repository for AI commands, skills, and agents
- 🔗 **Symlink-based Installation**: Install resources in projects without duplication
- 🤖 **Multi-Tool Support**: Works with Claude Code, OpenCode, and GitHub Copilot
- 🤖 **Agent Support**: Manage AI agents with specialized roles and capabilities
- ⚡ **Workspace Caching**: Git repositories cached for 10-50x faster subsequent updates
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

Resources can be added from multiple sources: local files, GitHub repositories, or bulk discovery from folders.

#### Add Individual Resources

Add specific resources (type is auto-detected):

```bash
# Add a command (single .md file - auto-detected as command)
aimgr repo add ~/.claude/commands/my-command.md

# Add a skill (directory with SKILL.md - auto-detected as skill)
aimgr repo add ~/my-skills/pdf-processing

# Add an agent (single .md file - auto-detected as agent)
aimgr repo add ~/.claude/agents/code-reviewer.md
```

#### Auto-Discovery from Folders and Repositories

Auto-discover and add all resources (commands, skills, agents, packages) from a folder or GitHub repository:

```bash
# From local folders
aimgr repo add ~/.opencode/           # Discovers all resources in .opencode
aimgr repo add ~/project/.claude/     # Discovers all resources in .claude
aimgr repo add ./my-resources/        # Any folder with resources

# From GitHub
aimgr repo add gh:owner/repo          # Auto-discovers all resources in repo
aimgr repo add gh:owner/repo@v1.0.0   # Specific version
aimgr repo add owner/repo             # Shorthand (gh: inferred)

# With options
aimgr repo add ~/.opencode/ --force         # Overwrite existing
aimgr repo add ./resources/ --skip-existing # Skip conflicts
aimgr repo add ./test/ --dry-run            # Preview without importing

# Filter resources with patterns
aimgr repo add gh:owner/repo --filter "skill/*"     # Only add skills
aimgr repo add ./resources/ --filter "skill/pdf*"   # Skills starting with "pdf"
aimgr repo add ~/.claude/ --filter "*test*"         # Resources with "test" in name
aimgr repo add gh:owner/repo --filter "package/*"   # Only add packages

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
aimgr repo add gh:vercel-labs/agent-skills

# Add a specific skill from a multi-skill repo
aimgr repo add gh:vercel-labs/agent-skills/skills/frontend-design

# Add from a specific branch or tag
aimgr repo add gh:anthropics/skills@v1.0.0

# Add specific resource types using filters
aimgr repo add gh:myorg/repo --filter "command/*"
aimgr repo add gh:myorg/repo --filter "agent/*"
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

# Use patterns to install multiple resources
aimgr install "skill/*"              # Install all skills
aimgr install "*test*"               # Install all resources with "test" in name
aimgr install "skill/pdf*"           # Install skills starting with "pdf"
aimgr install "command/test*" "agent/qa*"  # Multiple patterns

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

**Packages** are searched in:
1. `packages/` directory
2. Recursive search for `*.package.json` files

**Interactive Selection:**
- If a single resource is found, it's added automatically
- If multiple resources are found, you'll be prompted to select one
- If a specific subpath is provided, exactly one resource should exist at that location

## Pattern Syntax

Many commands support glob patterns for matching multiple resources. Patterns work with `repo add --filter`, `install`, and `uninstall` commands.

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
aimgr repo add gh:owner/repo --filter "skill/*"

# Add skills matching a pattern
aimgr repo add gh:owner/repo --filter "skill/pdf*"

# Add all resources with "test" in the name
aimgr repo add ./resources/ --filter "*test*"

# Add commands and agents only (no skills)
aimgr repo add ~/.opencode/ --filter "command/*" --filter "agent/*"

# Add only packages from a repository
aimgr repo add gh:owner/repo --filter "package/*"
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

Add resources to the repository from various sources. Resource types are auto-detected from file structure and content.

```bash
# Add from GitHub (with auto-discovery)
aimgr repo add gh:owner/repo
aimgr repo add gh:owner/repo/path/to/resource
aimgr repo add gh:owner/repo@v1.0.0

# Add from local path (type auto-detected)
aimgr repo add <path-to-file.md>           # Auto-detects command or agent
aimgr repo add <path-to-directory>         # Auto-detects skill
aimgr repo add ~/.claude/commands/test.md  # Command
aimgr repo add ~/my-skills/pdf-processing  # Skill
aimgr repo add ~/.claude/agents/reviewer.md # Agent

# Add using shorthand (infers gh: for owner/repo)
aimgr repo add vercel-labs/agent-skills

# Add from Git URL
aimgr repo add https://github.com/owner/repo.git
aimgr repo add git@github.com:owner/repo.git

# Add with explicit local prefix
aimgr repo add local:./my-resource
aimgr repo add local:/absolute/path/to/resource

# Add all resources from a folder (auto-discovery)
aimgr repo add ~/.opencode/
aimgr repo add ~/project/.claude/
aimgr repo add ./my-resources/

# Filter resources during import
aimgr repo add gh:owner/repo --filter "skill/*"       # Only skills
aimgr repo add ~/.opencode/ --filter "skill/pdf*"     # Skills starting with "pdf"
aimgr repo add ./resources/ --filter "*test*"         # Resources with "test" in name

# Import options
aimgr repo add <source> --force                       # Overwrite existing
aimgr repo add <source> --skip-existing               # Skip conflicts
aimgr repo add <source> --dry-run                     # Preview without importing
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
aimgr repo add gh:anthropics/skills
aimgr repo add gh:myorg/commands --filter "command/*"
aimgr repo add ~/local/resources --filter "skill/*"
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

### `aimgr repo create-package`

Create a package from existing resources in the repository.

```bash
# Create package with description and resources
aimgr repo create-package web-dev-tools \
  --description="Web development toolkit" \
  --resources="command/build,skill/typescript-helper,agent/code-reviewer"

# Create package with multiple resources
aimgr repo create-package testing-suite \
  --description="Complete testing toolkit" \
  --resources="command/test,command/coverage,skill/test-generator"

# Overwrite existing package
aimgr repo create-package my-tools \
  --description="Updated tools" \
  --resources="command/lint,skill/formatter" \
  --force
```

**Flags:**
- `--description`: Package description (required)
- `--resources`: Comma-separated list of resources in `type/name` format (required)
- `--force, -f`: Overwrite existing package

**Resource Format:** `type/name` where type is `command`, `skill`, or `agent`

All specified resources must exist in the repository before creating the package.

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

**Workspace Caching:**
Git repositories are cached in `.workspace/` for efficient reuse:
- **First update**: Full git clone (creates cache)
- **Subsequent updates**: Git pull only (10-50x faster)
- Caches automatically shared across resources from the same source
- Dramatically improves performance for bulk updates

**How it works:**
1. Reads source information from resource metadata (stored in `.metadata/` directory)
2. Fetches latest version from original source (GitHub, local path, etc.)
3. Uses cached Git repository if available (much faster)
4. Updates the repository copy
5. Preserves existing symlinks to projects

**Supported sources:**
- **GitHub**: Uses workspace cache for fast updates
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

## Packages

Packages allow you to group related resources (commands, skills, agents) together and install them as a unit. This makes it easy to distribute collections of tools or share themed resource sets.

### What Are Packages?

A package is a JSON file that references existing resources in your repository. When you install a package, all its resources are installed together. This is useful for:

- **Themed Collections**: Group resources for specific workflows (e.g., "testing-tools", "documentation-helpers")
- **Project Templates**: Create reusable resource sets for different project types
- **Distribution**: Share curated collections of resources with your team
- **Dependency Management**: Install all needed resources for a task in one command

### Creating Packages

Create a package from existing resources in your repository:

```bash
# Create a package with description and resources
aimgr repo create-package web-dev-tools \
  --description="Web development toolkit" \
  --resources="command/build,skill/typescript-helper,agent/code-reviewer"

# Create a package with multiple resources
aimgr repo create-package testing-suite \
  --description="Complete testing toolkit" \
  --resources="command/test,command/coverage,skill/test-generator,agent/qa-tester"

# Overwrite existing package
aimgr repo create-package my-tools \
  --description="Updated tools collection" \
  --resources="command/lint,skill/formatter" \
  --force
```

**Resource Format**: Resources are specified as `type/name`:
- `command/name` - Command resource
- `skill/name` - Skill resource  
- `agent/name` - Agent resource

All resources must already exist in your repository before creating a package.

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
# Create a package for React projects
aimgr repo create-package react-starter \
  --description="React development essentials" \
  --resources="command/dev,command/build,skill/react-helper,agent/react-reviewer"

# Install in new React project
cd my-react-app
aimgr install package/react-starter
```

**2. Testing Workflow:**
```bash
# Testing tools package
aimgr repo create-package test-tools \
  --description="Complete testing suite" \
  --resources="command/test,command/coverage,skill/test-generator,agent/qa-tester"

# Install for testing workflow
aimgr install package/test-tools
```

**3. Documentation:**
```bash
# Documentation helpers
aimgr repo create-package docs-tools \
  --description="Documentation toolkit" \
  --resources="skill/markdown-formatter,agent/doc-writer,command/doc-gen"
```

## Marketplace Import

Import entire Claude plugin marketplaces to quickly bootstrap your repository with curated collections of resources. Each plugin in the marketplace becomes an aimgr package containing all discovered resources.

### What is Marketplace Import?

Claude plugin marketplaces use a `marketplace.json` file to define collections of plugins with commands, skills, and agents. The marketplace import feature:

- **Parses** Claude marketplace.json files
- **Discovers** resources in each plugin's source directory
- **Creates** aimgr packages for each plugin
- **Imports** all resources into your repository
- **Tracks** metadata for future updates

This makes it easy to import entire plugin ecosystems in one command.

### Basic Usage

```bash
# Import from local marketplace
aimgr marketplace import ~/.claude-plugin/marketplace.json

# Import from GitHub (requires gh CLI)
aimgr marketplace import gh:anthropics/claude-code/.claude-plugin/marketplace.json

# Preview without importing
aimgr marketplace import <path> --dry-run
```

### Command Options

**Flags:**
- `--dry-run`: Preview what would be imported without making changes
- `--force, -f`: Overwrite existing packages and resources
- `--filter <pattern>`: Import only plugins matching the glob pattern
- `--source-url <url>`: Override source URL for metadata tracking

**Examples:**

```bash
# Import specific plugins only
aimgr marketplace import marketplace.json --filter "code-*"

# Force overwrite existing packages
aimgr marketplace import marketplace.json --force

# Specify source URL for metadata
aimgr marketplace import ./marketplace.json \
  --source-url "gh:myorg/plugins"
```

### How It Works

When you import a marketplace:

1. **Parse**: Reads the marketplace.json file
2. **Filter**: Applies optional pattern filters to select specific plugins
3. **Discover**: For each plugin, searches for resources in standard locations:
   - Commands: `commands/*.md`, `.claude/commands/*.md`
   - Skills: `skills/*/SKILL.md`, `.claude/skills/*/SKILL.md`
   - Agents: `agents/*.md`, `.claude/agents/*.md`
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
# Distribute marketplace.json to team members
# They import with one command:
aimgr marketplace import gh:myorg/plugins/.claude-plugin/marketplace.json

# All plugins become packages they can install:
aimgr install package/web-dev-tools
aimgr install package/testing-suite
```

**2. Public Plugin Collections:**
```bash
# Import community plugin collections
aimgr marketplace import gh:community/awesome-claude-plugins/marketplace.json

# Browse imported packages
aimgr repo list package

# Install what you need
aimgr install package/pdf-tools package/git-helpers
```

**3. Project Bootstrapping:**
```bash
# Create project-specific marketplace
# Import during project setup
aimgr marketplace import ./project-plugins/marketplace.json

# All project resources available immediately
aimgr install package/backend-tools package/frontend-tools
```

### Metadata and Updates

Imported packages and resources are tracked with metadata:

```bash
# View package metadata
aimgr repo show package/web-dev-tools

# Update all marketplace-imported resources
aimgr repo update

# Update specific package resources
aimgr repo update skill/typescript-helper
```

Metadata includes:
- Source URL (marketplace location)
- Original format (claude-plugin)
- Import timestamp
- Resource count

### Package Management After Import

After importing, packages work like regular aimgr packages:

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
aimgr marketplace import marketplace.json --filter "web-*"

# Import testing plugins
aimgr marketplace import marketplace.json --filter "*-test"

# Import multiple patterns (run multiple commands)
aimgr marketplace import marketplace.json --filter "code-*"
aimgr marketplace import marketplace.json --filter "dev-*"
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

aimgr automatically tracks metadata about resource sources, enabling features like `repo update` and `repo show`.

### What is Tracked

When you add a resource from GitHub or other sources, aimgr stores:

- **Source type**: `github`, `local`, or `file`
- **Source location**: URL, path, or file location
- **Git reference**: Branch or tag (for GitHub sources)
- **Subpath**: Path within repository (if applicable)
- **Timestamps**: When resource was added and last updated

### Metadata File Location

Metadata is stored in a centralized `.metadata/` directory organized by resource type:

**Skills:**
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

**Note:** The `source_type` and `source_url` fields are used for update tracking.

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

Create `~/.local/share/ai-config/repo/.metadata/skills/my-skill-metadata.json`:

```json
{
  "name": "my-skill",
  "type": "skill",
  "source_type": "local",
  "source_url": "file:///home/user/projects/my-skill",
  "first_installed": "2026-01-22T10:00:00Z",
  "last_updated": "2026-01-22T10:00:00Z"
}
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

The `.metadata/` directory contains JSON files with source tracking information (GitHub URLs, local paths, timestamps) used for resource updates.

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
| `aimgr repo add bulk <source>` | `aimgr repo add <source>` | Auto-discovery is now the default behavior |
| `aimgr repo add command <file>` | `aimgr repo add <file>` | Type auto-detected from file content |
| `aimgr repo add skill <dir>` | `aimgr repo add <dir>` | Type auto-detected from SKILL.md |
| `aimgr repo add agent <file>` | `aimgr repo add <file>` | Type auto-detected from frontmatter |
| N/A | `aimgr repo add <source> --filter "type/*"` | New: Filter resources during import |

### Before and After Examples

**Adding all resources from a folder:**
```bash
# Old way
aimgr repo add bulk ~/.opencode/

# New way (exactly the same)
aimgr repo add ~/.opencode/
```

**Adding only skills from a repository:**
```bash
# Old way (not possible - needed multiple commands)
# 1. Clone repo manually
# 2. Add each skill individually

# New way
aimgr repo add gh:owner/repo --filter "skill/*"
```

**Adding a single resource:**
```bash
# Old way
aimgr repo add command ~/my-command.md
aimgr repo add skill ~/my-skill/
aimgr repo add agent ~/my-agent.md

# New way (type auto-detected)
aimgr repo add ~/my-command.md
aimgr repo add ~/my-skill/
aimgr repo add ~/my-agent.md
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

- ✅ Replace `repo add bulk` with `repo add` (both work)
- ✅ Replace `repo add command/skill/agent` with `repo add` (type auto-detected)
- ✅ Use `--filter` flag to selectively import resources
- ✅ Use patterns with `install` and `uninstall` for batch operations
- ✅ Update scripts and documentation to use new syntax

All old commands continue to work - no breaking changes!

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
