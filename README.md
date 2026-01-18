# ai-repo - AI Resources Manager

A command-line tool for discovering, installing, and managing AI resources (commands and skills) across multiple AI coding tools including Claude Code, OpenCode, and GitHub Copilot.

## Features

- 📦 **Repository Management**: Centralized repository for AI commands and skills
- 🔗 **Symlink-based Installation**: Install resources in projects without duplication
- 🤖 **Multi-Tool Support**: Works with Claude Code, OpenCode, and GitHub Copilot
- ✅ **Format Validation**: Automatic validation of command and skill formats
- 🎯 **Type Safety**: Strong validation following agentskills.io and Claude Code specifications
- 🗂️ **Organized Storage**: Clean separation between commands and skills
- 🔧 **Smart Installation**: Automatically detects existing tool directories
- 💻 **Cross-platform**: Works on Linux and macOS (Windows support planned)

## Supported AI Tools

`ai-repo` supports three major AI coding tools:

| Tool | Commands | Skills | Directory |
|------|----------|--------|-----------|
| **[Claude Code](https://code.claude.com/)** | ✅ | ✅ | `.claude/` |
| **[OpenCode](https://opencode.ai/)** | ✅ | ✅ | `.opencode/` |
| **[GitHub Copilot](https://github.com/features/copilot)** | ❌ | ✅ | `.github/skills/` |

**Note:** GitHub Copilot only supports Agent Skills, not slash commands.

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

## Quick Start

### Check Version

```bash
# Display version information
ai-repo --version
# Output: ai-repo version 0.1.0 (commit: a1b2c3d, built: 2026-01-18T19:30:00Z)

# Short form
ai-repo -v
```

### 1. Configure Your Default Tool

```bash
# Set your preferred AI tool (claude, opencode, or copilot)
ai-repo config set default-tool claude

# Check current setting
ai-repo config get default-tool
```

### 2. Add Resources to Repository

```bash
# Add a command (single .md file)
ai-repo add command ~/.claude/commands/my-command.md

# Add a skill (directory with SKILL.md)
ai-repo add skill ~/my-skills/pdf-processing
```

### 3. List Available Resources

```bash
# List all resources
ai-repo list

# List only commands
ai-repo list command

# List only skills
ai-repo list skill

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

# Resources are symlinked to tool-specific directories
# Example: .claude/commands/, .opencode/commands/, etc.
```

### 5. Remove Resources

```bash
# Remove from repository (with confirmation)
ai-repo remove command my-command

# Force remove (skip confirmation)
ai-repo remove skill old-skill --force
```

## Multi-Tool Support

### Installation Behavior

`ai-repo` intelligently handles installation based on your project's existing tool directories:

#### Scenario 1: Fresh Project (No Tool Directories)
When installing to a project with no existing tool directories, `ai-repo` creates and uses your configured default tool directory:

```bash
# Set default tool
ai-repo config set default-tool claude

# Install in fresh project
cd ~/my-new-project
ai-repo install command test

# Result: Creates .claude/commands/test.md
```

#### Scenario 2: Existing Tool Directory
When a tool directory already exists (e.g., `.opencode/`), `ai-repo` installs to that directory, ignoring your default tool setting:

```bash
# Project already has .opencode directory
cd ~/existing-opencode-project
ai-repo install command test

# Result: Uses existing .opencode/commands/test.md
# (Even if your default is set to 'claude')
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

### Configuring Default Tool

Use the `config` command to set or view your default tool:

```bash
# Set default tool
ai-repo config set default-tool claude
ai-repo config set default-tool opencode
ai-repo config set default-tool copilot

# View current setting
ai-repo config get default-tool

# Configuration is stored in ~/.ai-repo.yaml
```

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
ai-repo config get default-tool

# Set a setting
ai-repo config set default-tool <tool>

# Valid tools: claude, opencode, copilot
```

### `ai-repo add`

Add resources to the repository.

```bash
# Add a command
ai-repo add command <path-to-file.md>

# Add a skill
ai-repo add skill <path-to-directory>

# Overwrite existing resource
ai-repo add command my-command.md --force
```

### `ai-repo list`

List resources in the repository.

```bash
# List all
ai-repo list

# Filter by type
ai-repo list command
ai-repo list skill

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

# Custom project path
ai-repo install command test --project-path ~/my-project

# Force reinstall
ai-repo install skill utils --force
```

### `ai-repo remove`

Remove resources from the repository.

```bash
# Remove with confirmation
ai-repo remove command <name>
ai-repo remove skill <name>

# Skip confirmation
ai-repo remove command test --force

# Alias
ai-repo rm command old-test
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

## Name Validation

Both commands and skills must follow these naming rules:

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
└── skills/
    ├── pdf-processing/
    │   ├── SKILL.md
    │   └── scripts/
    └── git-release/
        └── SKILL.md
```

## Project Installation

When you install resources in a project, symlinks are created in tool-specific directories:

### Claude Code Project
```
your-project/
└── .claude/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    └── skills/
        └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
```

### OpenCode Project
```
your-project/
└── .opencode/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    └── skills/
        └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
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
│   └── skills/
│       └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
└── .opencode/
    ├── commands/
    │   └── test.md -> ~/.local/share/ai-config/repo/commands/test.md
    └── skills/
        └── pdf-processing -> ~/.local/share/ai-config/repo/skills/pdf-processing/
```

The tool automatically detects existing tool directories and installs to all of them, ensuring resources are available in whichever tool you're using.

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

See [examples/README.md](examples/README.md) for detailed instructions.

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional)

### Build

```bash
make build
```

### Run Tests

```bash
# All tests
make test

# Unit tests only
make unit-test

# Integration tests only
make integration-test
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make vet

# All checks
make all
```

## Project Structure

```
.
├── cmd/               # CLI commands
│   ├── root.go
│   ├── add.go
│   ├── list.go
│   ├── install.go
│   └── remove.go
├── pkg/
│   ├── config/        # Configuration
│   ├── repo/          # Repository management
│   ├── resource/      # Resource types and validation
│   └── install/       # Installation logic
├── test/              # Integration tests
├── examples/          # Example resources
├── main.go
└── README.md
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

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
- [ ] Shell completion
