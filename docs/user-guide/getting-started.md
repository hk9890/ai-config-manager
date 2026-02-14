# Getting Started with aimgr

This guide will help you get started with **aimgr**, a command-line tool for managing AI resources (commands, skills, and agents) across multiple AI coding tools.

---

## What is aimgr?

`aimgr` helps you:

- **Store** AI resources (commands, skills, agents) in a centralized repository
- **Track** resource sources in `ai.repo.yaml` (local directories or Git repositories)
- **Install** them in your projects using symlinks (no duplication)
- **Share** resources across multiple AI tools (Claude Code, OpenCode, GitHub Copilot, Windsurf)
- **Sync** resources from tracked sources to keep them up-to-date
- **Organize** your AI workflow with packages and resource collections

**Key concept:** Sources (tracked in `ai.repo.yaml`) provide resources, which you install into projects.

---

## Quick Start

Here's the typical workflow in 4 steps:

```bash
# 1. Initialize your repository (creates ai.repo.yaml)
aimgr repo init

# 2. Add a source (local directory or GitHub repo)
aimgr repo add gh:hk9890/ai-tools

# 3. View what's available
aimgr repo info        # See sources
aimgr repo list        # See resources

# 4. Install resources into your project
cd ~/my-project
aimgr install skill/pdf-processing
```

**Key commands:**
- `repo add` - Add a new source (updates ai.repo.yaml)
- `repo sync` - Update resources from all sources
- `repo info` - View sources and repository status
- `install` - Install resources into current project

---

## Installation

Choose your platform and follow the instructions:

### Linux (amd64)
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

### Linux (arm64)
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_darwin_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

### macOS (Apple Silicon)
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_darwin_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

*Note: Replace `VERSION` with the actual version number (e.g., `v0.1.0`). Check the [Releases page](https://github.com/hk9890/ai-config-manager/releases) for the latest version.*

### Verify Installation

After installation, verify `aimgr` is working:

```bash
aimgr --version
```

You should see version information like:
```
aimgr version 0.1.0 (commit: a1b2c3d, built: 2026-01-18T19:30:00Z)
```

---

## First Steps

### 1. Configure Your AI Tool

First, tell `aimgr` which AI tool(s) you're using. This determines where resources are installed:

```bash
# For Claude Code
aimgr config set install.targets claude

# For OpenCode
aimgr config set install.targets opencode

# For VSCode / GitHub Copilot (both names work)
aimgr config set install.targets copilot
aimgr config set install.targets vscode

# For multiple tools (installs to all)
aimgr config set install.targets claude,opencode,copilot
```

**Notes:**
- If you're not sure which tool you're using, start with `claude` (most common)
- VSCode / GitHub Copilot only supports skills (no commands or agents)
- Use either `copilot` or `vscode` as the tool name (both work)

### 2. Initialize Your Repository

First, initialize your aimgr repository. This creates an `ai.repo.yaml` file to track your resource sources:

```bash
aimgr repo init
```

This creates the repository directory and the `ai.repo.yaml` manifest file that tracks all your sources.

### 3. Add Your First Resource Source

Now let's add a source to your repository. Sources are tracked in `ai.repo.yaml` and can be:
- Local directories (symlinked for live editing)
- GitHub repositories (copied and versioned)
- Existing tool directories (`.claude/`, `.opencode/`, etc.)

#### Example: Add from a Local Directory

If you have resources in a local directory:
```bash
# Add from your existing tool directory
aimgr repo add ~/.claude/

# Add from your own resource folder
aimgr repo add ~/my-resources/
```

**Note:** Local sources are symlinked by default for live editing. Use `--copy` to copy instead.

#### Example: Add from GitHub

You can also add resources directly from GitHub repositories:

```bash
# Add all resources from a GitHub repo
aimgr repo add gh:hk9890/ai-tools

# Add specific resources using filters
aimgr repo add gh:hk9890/ai-tools --filter "skill/*"

# Add a specific version
aimgr repo add gh:hk9890/ai-tools@v1.0.0
```

**Note:** GitHub sources are automatically copied to your repository.

#### What Happens When You Add

When you run `aimgr repo add`:
1. Resources are auto-discovered from the source
2. The source is recorded in `ai.repo.yaml` 
3. Resources are added to your repository (symlinked for local, copied for remote)
4. You can now install them in your projects

**Tip:** Run `aimgr repo info` to see all your sources and their status.

### 4. View Your Resources

Check what resources and sources are now in your repository:

```bash
# View repository information and sources
aimgr repo info

# List all resources
aimgr repo list

# List only commands
aimgr repo list command

# List only skills
aimgr repo list skill

# List only agents
aimgr repo list agent
```

The `repo info` command shows your tracked sources, while `repo list` shows individual resources.

---

## Common Operations

### Managing Sources

aimgr tracks resource sources in `ai.repo.yaml`. Here are the key commands:

```bash
# Add a new source (local or remote)
aimgr repo add ~/my-resources/          # Local (symlinked)
aimgr repo add gh:owner/repo            # GitHub (copied)
aimgr repo add gh:owner/repo@v1.0.0    # Specific version

# Sync resources from all tracked sources
aimgr repo sync

# View sources and their status
aimgr repo info

# Remove a source from tracking
aimgr repo drop-source source-name
```

**Three key concepts:**
- **Add** - Track a new source in ai.repo.yaml and import its resources
- **Sync** - Update resources from all tracked sources
- **Remove** - Stop tracking a source (use `repo drop-source`)

See [sync-sources.md](sync-sources.md) for detailed source management information.

### Install Resources in a Project

Navigate to your project directory and install resources:

```bash
cd ~/my-project

# Install a skill
aimgr install skill/pdf-processing

# Install a command
aimgr install command/my-command

# Install an agent
aimgr install agent/code-reviewer

# Install multiple resources at once
aimgr install skill/pdf-processing command/my-command agent/code-reviewer
```

**What happens:**
- Resources are symlinked from your repository to the project
- They're placed in tool-specific directories (`.claude/`, `.opencode/`, etc.)
- Your AI tool can now use them immediately

### List Installed Resources in Current Project

Check what's installed in your current project:

```bash
aimgr list
```

This shows resources installed in the current directory:

```bash
┌──────────────────────┬───────────────────┬────────────────────┐
│         NAME         │      TARGETS      │    DESCRIPTION     │
├──────────────────────┼───────────────────┼────────────────────┤
│ skill/skill-creator  │ claude, opencode  │ Guide for creating │
│ skill/webapp-testing │ claude            │ Toolkit for inter  │
└──────────────────────┴───────────────────┴────────────────────┘
```

**TARGETS** shows which AI tools (claude, opencode, copilot) have this resource installed.

### Check Repository Resources with Sync Status

To see all resources in your repository with synchronization status:

```bash
aimgr repo list
```

This shows resources with their sync status relative to your `ai.package.yaml` manifest:

```bash
┌──────────────────────┬───────────────────┬──────────┬────────────────────┐
│         NAME         │      TARGETS      │   SYNC   │    DESCRIPTION     │
├──────────────────────┼───────────────────┼──────────┼────────────────────┤
│ skill/skill-creator  │ claude, opencode  │    ✓     │ Guide for creating │
│ skill/webapp-testing │ claude            │    *     │ Toolkit for inter  │
│ command/test         │ -                 │    ⚠     │ Run tests          │
└──────────────────────┴───────────────────┴──────────┴────────────────────┘

Legend:
  ✓ = In sync  * = Not in manifest  ⚠ = Not installed  - = No manifest
```

**Understanding sync status:**
- **✓ (In sync)**: Resource is in manifest and installed
- **\* (Not in manifest)**: Resource is installed but not in ai.package.yaml
- **⚠ (Not installed)**: Resource is in manifest but needs installation
- **\- (No manifest)**: No ai.package.yaml file exists

**Tip:** If you see resources marked with `*`, consider adding them to your `ai.package.yaml` file to track them as project dependencies.

### Uninstall Resources from a Project

Remove resources from a project (doesn't delete from repository):

```bash
# Uninstall a resource
aimgr uninstall skill/pdf-processing

# Uninstall multiple resources
aimgr uninstall skill/foo command/bar agent/reviewer
```

**What happens:**
- Symlinks are removed from the project
- Resources remain in your repository for future use

### Remove Resources from Repository

To completely remove a resource from your repository:

```bash
# Remove with confirmation prompt
aimgr repo remove skill old-skill

# Force remove (skip confirmation)
aimgr repo remove command test-command --force
```

**Warning:** This removes the resource from your repository. Projects using it will have broken symlinks.

### Working with Packages

Packages let you install multiple related resources at once:

```bash
# Install a package (installs all its resources)
aimgr install package/web-dev-tools

# List available packages
aimgr repo list package

# Uninstall a package (removes all its resources)
aimgr uninstall package/web-dev-tools
```

---

## Common Workflows

### Workflow 1: Adding and Using a Skill from GitHub

```bash
# 1. Add GitHub source (adds all resources)
aimgr repo add gh:hk9890/ai-tools --filter "skill/typescript-helper"

# 2. Navigate to your project
cd ~/my-typescript-project

# 3. Install the skill
aimgr install skill/typescript-helper

# 4. Use the skill in your AI tool
# (Now available in Claude Code, OpenCode, etc.)
```

### Workflow 2: Sharing Resources with Your Team

```bash
# 1. Create a GitHub repository with your resources
# 2. Team members add your repo as a source:
aimgr repo add gh:myorg/ai-resources

# 3. They install what they need:
aimgr install skill/company-coding-standards
aimgr install agent/code-reviewer
```

### Workflow 3: Using Packages for Project Setup

```bash
# 1. Add a package collection source
aimgr repo add gh:myorg/project-templates

# 2. Start a new project
cd ~/new-react-app

# 3. Install the full package
aimgr install package/react-starter

# (All React-related commands, skills, and agents are now installed)
```

---

## Tips and Best Practices

### Tip 1: Use Shell Completion

Enable tab completion to speed up your workflow:

```bash
# Bash
aimgr completion bash > /etc/bash_completion.d/aimgr

# Zsh
aimgr completion zsh > "${fpath[1]}/_aimgr"

# Fish
aimgr completion fish > ~/.config/fish/completions/aimgr.fish
```

Now you can tab-complete resource names:
```bash
aimgr install skill/<TAB>
# Shows: atlassian-cli  dynatrace-api  github-docs  pdf-processing  skill-creator
```

### Tip 2: Use Patterns for Bulk Operations

Install multiple resources matching a pattern:

```bash
# Install all skills
aimgr install "skill/*"

# Install all test-related resources
aimgr install "*test*"

# Install PDF-related skills
aimgr install "skill/pdf*"
```

### Tip 3: Use --dry-run to Preview

Preview operations before executing:

```bash
# Preview what would be added
aimgr repo add ~/resources/ --dry-run

# Preview what would be synced
aimgr repo sync --dry-run
```

### Tip 4: Keep Your Repository Synced

Use `repo sync` to update from your tracked sources in `ai.repo.yaml`:

```bash
# Sync all sources (updates from remote, reflects local changes)
aimgr repo sync

# View your sources and their status
aimgr repo info
```

**Note:** Sources are tracked in `ai.repo.yaml` (created by `repo init` or `repo add`). See [sync-sources.md](sync-sources.md) for detailed source management.

### Tip 5: Manage Sources with Commands

Use `repo` commands to manage your sources:

```bash
# View all sources and their status
aimgr repo info

# Add a new source
aimgr repo add gh:myorg/resources

# Sync all sources
aimgr repo sync

# Remove a source (from ai.repo.yaml)
aimgr repo drop-source myorg-resources
```

Your sources are stored in `ai.repo.yaml` at the root of your repository. You can also edit this file directly if needed. See [sync-sources.md](sync-sources.md) for more details on source management.

---

## Troubleshooting

### Issue: Command not found after installation

**Solution:** Make sure `aimgr` is in your PATH:
```bash
# Check if ~/bin or /usr/local/bin is in PATH
echo $PATH

# Add to PATH if needed (add to ~/.bashrc or ~/.zshrc)
export PATH="/usr/local/bin:$PATH"
```

### Issue: Resources not showing in my AI tool

**Solution:** 
1. Check that resources are installed: `aimgr list`
   - Look at the **TARGETS** column to see which tools have the resource
   - If your tool isn't listed, the resource isn't installed for that tool
2. Verify your tool is using the right directory (`.claude/`, `.opencode/`, etc.)
3. Check installation targets in your config: `aimgr config get install.targets`
4. Restart your AI tool to pick up new resources

### Issue: Resource marked with * (not in manifest)

**What it means:** The resource is installed but not declared in your `ai.package.yaml` file.

**Solution:**
- If this is a project dependency you want to track, add it to `ai.package.yaml`:
  ```yaml
  resources:
    - skill/my-skill
    - command/my-command
  ```
- If it's just a temporary or local-only resource, you can ignore the `*` symbol

### Issue: Resource marked with ⚠ (not installed) in repo list

**What it means:** The resource is declared in `ai.package.yaml` but not installed yet.

**Solution:**
```bash
# Install the resource
aimgr install skill/my-skill

# Or install all resources from manifest
aimgr install $(aimgr repo list --format=json | jq -r '.resources[] | select(.sync_status == "not-installed") | .name')
```

### Issue: "Resource already exists" error

**Solution:** Use `--force` to overwrite:
```bash
aimgr repo add ~/resource/ --force
aimgr install skill/foo --force
```

### Issue: Source not showing in repo info

**What it means:** The source was not successfully added to `ai.repo.yaml`.

**Solution:**
1. Check if `ai.repo.yaml` exists in your repository:
   ```bash
   aimgr repo info
   ```
2. If missing, initialize the repository:
   ```bash
   aimgr repo init
   ```
3. Re-add your sources:
   ```bash
   aimgr repo add ~/my-resources/
   ```

### Issue: Symlinks are broken after removing resource

**Solution:** Uninstall from projects before removing from repository:
```bash
# 1. First uninstall from all projects
cd ~/project1 && aimgr uninstall skill/foo
cd ~/project2 && aimgr uninstall skill/foo

# 2. Then remove from repository
aimgr repo remove skill foo
```

---

## Next Steps

Now that you're familiar with the basics, explore these guides:

- **[Pattern Matching](pattern-matching.md)** - Learn advanced pattern syntax for filtering and bulk operations
- **[Output Formats](output-formats.md)** - Use JSON/YAML output for scripting and automation
- **[Sync Sources](sync-sources.md)** - Detailed guide to managing sources in ai.repo.yaml
- **[GitHub Sources](github-sources.md)** - Add resources from GitHub repositories
- **[Resource Formats](resource-formats.md)** - Understand resource file structures and create your own
- **[Workspace Caching](workspace-caching.md)** - Learn about Git repository caching for faster operations

### Creating Your Own Resources

Ready to create your own commands, skills, or agents? Check out:

- **Resource format specifications** in [Resource Formats](resource-formats.md)
- **Example resources** in the `examples/` directory of the repository
- **agentskills.io** for resource naming rules and best practices

### Advanced Features

Explore advanced `aimgr` features:

- **Source tracking** - Use ai.repo.yaml to manage resource sources
- **Packages** - Group related resources for easier distribution
- **Workspace caching** - Speed up Git operations with repository caching
- **Multi-tool support** - Install to multiple AI tools simultaneously
- **Local development** - Use symlinked sources for live editing

---

## Getting Help

- **Documentation:** [GitHub Repository](https://github.com/hk9890/ai-config-manager)
- **Issues:** [Report bugs or request features](https://github.com/hk9890/ai-config-manager/issues)
- **Command help:** Run `aimgr --help` or `aimgr <command> --help`

Happy managing your AI resources! 🚀
