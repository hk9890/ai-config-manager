# ai-repo - AI Resources Manager

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

`ai-repo` supports three major AI coding tools:

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
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/ai-repo_VERSION_linux_amd64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/
```

**Linux (arm64)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/ai-repo_VERSION_linux_arm64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/
```

**macOS (Intel)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/ai-repo_VERSION_darwin_amd64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/
```

**macOS (Apple Silicon)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/ai-repo_VERSION_darwin_arm64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/
```

**Windows (PowerShell)**:
```powershell
Invoke-WebRequest -Uri "https://github.com/hk9890/ai-config-manager/releases/latest/download/ai-repo_VERSION_windows_amd64.zip" -OutFile "ai-repo.zip"
Expand-Archive -Path "ai-repo.zip" -DestinationPath "."
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

`ai-repo` supports shell completion for Bash, Zsh, Fish, and PowerShell, making it easier to discover and install resources.

### Features

- **Auto-complete resource names** when typing `ai-repo install skill <TAB>`
- **Auto-complete commands** like `ai-repo install command <TAB>`
- **Auto-complete agents** like `ai-repo install agent <TAB>`
- **Dynamically queries your repository** to show available resources

### Setup

#### Bash

**Option 1: Load for current session**
```bash
source <(ai-repo completion bash)
```

**Option 2: Load automatically for all sessions**

Linux:
```bash
ai-repo completion bash > /etc/bash_completion.d/ai-repo
```

macOS (with Homebrew):
```bash
ai-repo completion bash > $(brew --prefix)/etc/bash_completion.d/ai-repo
```

#### Zsh

**Enable completions (if not already enabled):**
```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

**Install completion:**
```bash
ai-repo completion zsh > "${fpath[1]}/_ai-repo"
```

Then start a new shell.

#### Fish

**Option 1: Load for current session**
```bash
ai-repo completion fish | source
```

**Option 2: Load automatically for all sessions**
```bash
ai-repo completion fish > ~/.config/fish/completions/ai-repo.fish
```

#### PowerShell

**Option 1: Load for current session**
```powershell
ai-repo completion powershell | Out-String | Invoke-Expression
```

**Option 2: Load automatically for all sessions**
```powershell
# Generate completion script
ai-repo completion powershell > ai-repo.ps1

# Add to your PowerShell profile
# Find profile location: $PROFILE
```

### Usage Examples

After setting up completion, you can use TAB to auto-complete resource names:

```bash
# List all available skills
ai-repo install skill <TAB>
# Shows: atlassian-cli  dynatrace-api  dynatrace-control  github-docs  skill-creator

# List all available commands
ai-repo install command <TAB>
# Shows: test  review  deploy  build

# List all available agents
ai-repo install agent <TAB>
# Shows: code-reviewer  qa-tester  beads-task-agent
```

The completion dynamically queries your repository, so newly added resources are immediately available for completion!



## Quick Start

### Check Version

```bash
# Display version information
ai-repo --version
# Output: ai-repo version 0.1.0 (commit: a1b2c3d, built: 2026-01-18T19:30:00Z)

# Short form
ai-repo -v
```

### 1. Configure Your Default Installation Targets

```bash
# Set your preferred AI tools (claude, opencode, or copilot)
ai-repo config set install.targets claude
ai-repo config set install.targets claude,opencode  # Multiple tools

# Check current setting
ai-repo config get install.targets
```

### 2. Add Resources to Repository

```bash
# Add a command (single .md file)
ai-repo add command ~/.claude/commands/my-command.md

# Add a skill (directory with SKILL.md)
ai-repo add skill ~/my-skills/pdf-processing

# Add an agent (single .md file)
ai-repo add agent ~/.claude/agents/code-reviewer.md

# Import from OpenCode directory
ai-repo add opencode ~/.opencode
```

## Bulk Import

Import multiple resources at once from Claude plugins or Claude configuration folders.

### Import from Claude Plugin

Import all commands and skills from a Claude plugin directory:

```bash
# Import from a Claude plugin
ai-repo add plugin ~/.claude/plugins/marketplaces/claude-plugins-official/plugins/example-plugin

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
ai-repo add claude ~/.claude

# Import from project's .claude folder
ai-repo add claude ~/my-project/.claude

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
ai-repo add opencode ~/.opencode

# Import from project's .opencode folder
ai-repo add opencode ~/my-project/.opencode

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
ai-repo add plugin ./my-plugin --force

# Skip existing resources (no error)
ai-repo add claude ~/.claude --skip-existing

# Preview what would be imported (dry run)
ai-repo add plugin ./plugin --dry-run

# Default behavior: fail on first conflict
ai-repo add plugin ./plugin  # Error if any resource already exists
```

### Bulk Import Use Cases

**Scenario 1: Setting up your ai-repo from existing Claude setup**
```bash
# Import all your existing commands and skills at once
ai-repo add claude ~/.claude
```

**Scenario 2: Installing a complete plugin**
```bash
# Import all resources from a plugin in one command
ai-repo add plugin ~/.claude/plugins/marketplaces/claude-plugins-official/plugins/pr-review-toolkit
```

**Scenario 3: Migrating resources between machines**
```bash
# On machine A: your resources are already in ai-repo
# On machine B: import from a cloned .claude folder
ai-repo add claude ~/cloned-config/.claude
```



### 3. List Available Resources

```bash
# List all resources
ai-repo list

# List only commands
ai-repo list command

# List only skills
ai-repo list skill

# List only agents
ai-repo list agent

# JSON output
ai-repo list --format=json
```

### 4. Install in a Project

```bash
cd your-project/

# Install a command
ai-repo install command my-command

# Install a skill
ai-repo install skill pdf-processing

# Install an agent
ai-repo install agent code-reviewer

# Resources are symlinked to tool-specific directories
# Example: .claude/commands/, .opencode/commands/, etc.
```

### 5. Remove Resources

```bash
# Remove from repository (with confirmation)
ai-repo remove command my-command

# Force remove (skip confirmation)
ai-repo remove skill old-skill --force

# Remove an agent
ai-repo remove agent old-agent
```

## Multi-Tool Support

### Installation Behavior

`ai-repo` intelligently handles installation based on your project's existing tool directories:

#### Scenario 1: Fresh Project (No Tool Directories)
When installing to a project with no existing tool directories, `ai-repo` creates and uses your configured default installation targets:

```bash
# Set default targets
ai-repo config set install.targets claude

# Install in fresh project
cd ~/my-new-project
ai-repo install command test

# Result: Creates .claude/commands/test.md

# Or install to multiple tools by default
ai-repo config set install.targets claude,opencode
ai-repo install command test
# Result: Creates both .claude/commands/test.md and .opencode/commands/test.md
```

#### Scenario 2: Existing Tool Directory
When a tool directory already exists (e.g., `.opencode/`), `ai-repo` installs to that directory, ignoring your default installation targets:

```bash
# Project already has .opencode directory
cd ~/existing-opencode-project
ai-repo install command test

# Result: Uses existing .opencode/commands/test.md
# (Even if your default targets are set to 'claude')
```

#### Scenario 3: Multiple Tool Directories
When multiple tool directories exist, `ai-repo` installs to **ALL** of them:

```bash
# Project has both .claude and .opencode directories
cd ~/multi-tool-project
ai-repo install skill pdf-processing

# Result: Installs to BOTH:
#   - .claude/skills/pdf-processing
#   - .opencode/skills/pdf-processing
```

This ensures resources are available regardless of which tool you're using.

### Configuring Default Installation Targets

Use the `config` command to set or view your default installation targets:

```bash
# Set default target (single tool)
ai-repo config set install.targets claude
ai-repo config set install.targets opencode
ai-repo config set install.targets copilot

# Set multiple default targets
ai-repo config set install.targets claude,opencode

# View current setting
ai-repo config get install.targets

# Configuration is stored in ~/.config/ai-repo/ai-repo.yaml
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

## Commands

### `ai-repo config`

View and manage configuration settings.

```bash
# Get a setting
ai-repo config get install.targets

# Set a setting (single tool)
ai-repo config set install.targets <tool>

# Set multiple targets
ai-repo config set install.targets <tool1>,<tool2>

# Valid tools: claude, opencode, copilot
# Examples:
ai-repo config set install.targets claude
ai-repo config set install.targets claude,opencode
```

### `ai-repo add`

Add resources to the repository.

```bash
# Add a single command
ai-repo add command <path-to-file.md>

# Add a single skill
ai-repo add skill <path-to-directory>

# Add a single agent
ai-repo add agent <path-to-file.md>

# Add all resources from a Claude plugin
ai-repo add plugin <path-to-plugin>

# Add all resources from a Claude folder
ai-repo add claude <path-to-.claude-folder>

# Add all resources from an OpenCode folder
ai-repo add opencode <path-to-.opencode-folder>

# Overwrite existing resource
ai-repo add command my-command.md --force

# Bulk import with conflict handling
ai-repo add plugin ./plugin --skip-existing
ai-repo add claude ~/.claude --dry-run
ai-repo add opencode ~/.opencode --skip-existing
```

### `ai-repo list`

List resources in the repository.

```bash
# List all
ai-repo list

# Filter by type
ai-repo list command
ai-repo list skill
ai-repo list agent

# Output formats
ai-repo list --format=table  # Default
ai-repo list --format=json
ai-repo list --format=yaml
```

### `ai-repo install`

Install resources to a project.

```bash
# Install command
ai-repo install command <name>

# Install skill
ai-repo install skill <name>

# Install agent
ai-repo install agent <name>

# Custom project path
ai-repo install command test --project-path ~/my-project

# Force reinstall
ai-repo install skill utils --force
ai-repo install agent code-reviewer --force

# Install to specific tool(s) - overrides defaults and existing directories
ai-repo install command test --target claude
ai-repo install skill utils --target opencode
ai-repo install agent reviewer --target claude,opencode
```

**Flags:**
- `--project-path`: Specify project directory (defaults to current directory)
- `--force`: Overwrite existing installations
- `--target`: Specify which tool(s) to install to (accepts comma-separated values). Overrides both default targets and existing directory detection.

### `ai-repo remove`

Remove resources from the repository.

```bash
# Remove with confirmation
ai-repo remove command <name>
ai-repo remove skill <name>
ai-repo remove agent <name>

# Skip confirmation
ai-repo remove command test --force

# Alias
ai-repo rm command old-test
ai-repo rm agent old-reviewer
```

## Resource Formats

### Commands

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

**Old location:** `~/.ai-repo.yaml`  
**New location:** `~/.config/ai-repo/ai-repo.yaml`

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

**Migration:** Automatic. When you run any `ai-repo` command, the tool will automatically convert `default-tool` to `install.targets` format. Your config file will be updated to the new format.

### Command Changes

**Old commands:**
```bash
ai-repo config set default-tool claude
ai-repo config get default-tool
```

**New commands:**
```bash
ai-repo config set install.targets claude
ai-repo config set install.targets claude,opencode  # Multiple tools
ai-repo config get install.targets
```

### New Feature: --target Flag

You can now override the installation target for individual operations:

```bash
# Install to specific tool(s), ignoring defaults
ai-repo install command test --target claude
ai-repo install skill utils --target claude,opencode
```

This is useful when you want to install a resource to a specific tool without changing your global configuration.

## Creating Resources

### Create a Command

1. Create a `.md` file with valid name: `my-command.md`
2. Add YAML frontmatter with `description`
3. Write command body in markdown
4. Test: `ai-repo add command ./my-command.md`

### Create a Skill

1. Create directory with valid name: `my-skill/`
2. Create `SKILL.md` with frontmatter (name must match directory)
3. Optionally add `scripts/`, `references/`, `assets/`
4. Test: `ai-repo add skill ./my-skill`

### Create an Agent

1. Create a `.md` file with valid name: `my-agent.md`
2. Add YAML frontmatter with `description`
3. Optionally add `type`, `instructions`, `capabilities` (OpenCode format)
4. Write agent documentation in markdown body
5. Test: `ai-repo add agent ./my-agent.md`

See [examples/README.md](examples/README.md) for detailed instructions.

## Development

For information on building, testing, and contributing to ai-repo, please see [CONTRIBUTING.md](CONTRIBUTING.md).

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

## Roadmap

- [ ] Windows support (junction instead of symlinks)
- [ ] Search functionality for resources
- [ ] Remote repository support
- [ ] Resource versioning
- [ ] Update/upgrade commands
- [x] Shell completion
