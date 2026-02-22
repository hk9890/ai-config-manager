# Sources Guide

This guide explains how to manage sources in your aimgr repository. Sources are locations (local paths or remote URLs) containing AI resources that you want to import and track.

---

## Overview

**aimgr** uses a **repository-local** approach to source management. Each repository tracks its own sources in an `ai.repo.yaml` manifest file, which is version-controlled and portable.

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Source** | A location (local path or remote URL) containing AI resources |
| **ai.repo.yaml** | Manifest file tracking all configured sources |
| **Import Mode** | How resources are stored - `symlink` for paths, `copy` for URLs |
| **Sync** | Re-import resources from configured sources to get latest changes |

### Benefits

- **Self-contained repositories**: All source information lives in the repository
- **Git-tracked manifests**: Share source configurations via version control
- **Automatic synchronization**: Re-import resources with a single command
- **Orphan cleanup**: Automatically remove resources when sources are removed

---

## The ai.repo.yaml File

The `ai.repo.yaml` file is automatically created and maintained in your repository root (`~/.local/share/ai-config/repo/ai.repo.yaml` by default). It tracks all sources you've added and their metadata.

### File Format

```yaml
version: 1
sources:
  # Local source (uses symlink mode)
  - name: my-local-commands
    path: /home/user/my-resources
    
  # Remote source (uses copy mode)
  - name: agentskills-catalog
    url: https://github.com/agentskills/catalog
    ref: main
    subpath: resources
```

### Field Reference

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `version` | integer | Manifest format version (currently 1) | Yes |
| `sources` | array | List of source configurations | Yes |
| `name` | string | Unique identifier for the source | Yes |
| `path` | string | Absolute path to local directory (for local sources) | One of path/url |
| `url` | string | Git repository URL (for remote sources) | One of path/url |
| `ref` | string | Git branch/tag/commit (for remote sources) | No |
| `subpath` | string | Subdirectory within repository (for remote sources) | No |

**Note:** Import mode is implicit based on source type. Path sources use `symlink` mode; URL sources use `copy` mode.

---

## Local Sources

Local sources point to directories on your filesystem. Resources are automatically **symlinked** for live editing.

### Adding Local Sources

```bash
# Add from local directory
aimgr repo add ~/my-skills
aimgr repo add /opt/team-resources
aimgr repo add ./local-resources

# With custom name
aimgr repo add ~/my-skills --name=personal-skills
```

### How Symlink Mode Works

When you add a local source:
1. Resources are discovered in the source directory
2. Symbolic links are created in your repository pointing to the original files
3. Changes to source files immediately reflect in the repository

### Benefits

- **Live editing**: Changes appear instantly without re-importing
- **No duplication**: Original files stay in their location
- **Perfect for development**: Quick iteration without sync commands

### Use Cases

- Local development and testing
- Personal resource collections
- Active development with rapid iteration

### Example ai.repo.yaml Entry

```yaml
sources:
  - name: my-local-skills
    path: /home/user/dev/my-skills
```

---

## Remote Sources

Remote sources point to Git repositories (GitHub, GitLab, etc.). Resources are automatically **copied** to your repository.

### Adding Remote Sources

```bash
# Basic GitHub URL
aimgr repo add https://github.com/owner/repo

# With custom name
aimgr repo add https://github.com/owner/repo --name=community-tools
```

### How Copy Mode Works

When you add a remote source:
1. The repository is cloned to a workspace cache
2. Resources are discovered and copied to your repository
3. Source is tracked in `ai.repo.yaml` for future syncing

### Benefits

- **Stable, versioned resources**: Pin to specific versions
- **Works offline**: No network needed after initial import
- **No external dependencies**: Resources are self-contained

### Use Cases

- Production environments
- Shared team resources
- Community packages
- Versioned resource collections

### Example ai.repo.yaml Entry

```yaml
sources:
  - name: community-catalog
    url: https://github.com/owner/repo
    ref: v1.2.0
```

---

## GitHub-Specific Syntax

GitHub sources support special syntax for specifying branches, tags, and subdirectories.

### URL Formats

```bash
# Standard GitHub URL
aimgr repo add https://github.com/owner/repo

# Git URL with .git extension
aimgr repo add https://github.com/owner/repo.git

# SSH URL (requires configured keys)
aimgr repo add git@github.com:owner/repo.git
```

### Specifying Refs (Branches/Tags)

Use `#` to specify a branch, tag, or commit:

```bash
# Specific branch
aimgr repo add https://github.com/owner/repo#develop

# Specific tag (recommended for stability)
aimgr repo add https://github.com/owner/repo#v1.2.0

# Specific commit
aimgr repo add https://github.com/owner/repo#abc1234
```

### Specifying Subpaths

Add a path after the repository name to target a specific directory:

```bash
# Resources in a subdirectory
aimgr repo add https://github.com/owner/repo/skills/frontend

# Subpath with ref
aimgr repo add https://github.com/owner/repo/commands#v1.0.0
```

### Complete Examples

```bash
# All resources from latest main branch
aimgr repo add https://github.com/owner/repo

# Specific version
aimgr repo add https://github.com/owner/repo#v2.0.0 --name=stable-tools

# Specific directory and version
aimgr repo add https://github.com/owner/mono-repo/skills/frontend#v1.0.0

# SSH with branch
aimgr repo add git@github.com:owner/repo.git#develop
```

### Auto-Discovery

When adding from GitHub, `aimgr` automatically discovers resources in standard locations:

**Skills** are searched in order:
1. Direct path (if subpath specified)
2. `skills/`, `.claude/skills/`, `.opencode/skills/`, `.github/skills/`, etc.
3. Recursive search (max depth 5)

**Commands** are searched in:
1. `commands/`, `.claude/commands/`, `.opencode/commands/`
2. Recursive search for `.md` files

**Agents** are searched in:
1. `agents/`, `.claude/agents/`, `.opencode/agents/`
2. Recursive search for `.md` files

### Authentication

For private repositories:

```bash
# HTTPS with credentials (configure Git credential helper)
git config --global credential.helper store
aimgr repo add https://github.com/private/repo

# SSH (requires configured SSH keys)
ssh -T git@github.com  # Test authentication
aimgr repo add git@github.com:private/repo.git
```

### Workspace Caching

GitHub operations use a workspace cache for performance:

- **First add**: Full clone (~30s for large repos)
- **Subsequent syncs**: Uses cached clone (<1s)
- **Cache location**: `~/.cache/aimgr/workspace/`
- **Safe to clear**: `rm -rf ~/.cache/aimgr/workspace/`

---

## Syncing and Updating Sources

The `repo sync` command re-imports resources from all configured sources.

### Basic Usage

```bash
# Sync all sources
aimgr repo sync

# Preview without changes
aimgr repo sync --dry-run

# Skip existing resources (don't overwrite)
aimgr repo sync --skip-existing
```

### What Sync Does

For each source in `ai.repo.yaml`:
- **Path sources**: Re-create symlinks to source files
- **URL sources**: Download latest version, copy to repository

### When to Sync

- After upstream changes to remote repositories
- To refresh all sources at once
- After manually editing `ai.repo.yaml`
- To verify source availability

### Options

| Flag | Description |
|------|-------------|
| `--skip-existing` | Don't overwrite existing resources |
| `--dry-run` | Preview without importing |
| `--format=<format>` | Output format: table, json, yaml |

### Handling Failures

If a source becomes unavailable:
1. Error is reported
2. Remaining sources continue syncing
3. Command exits with error status

The failed source remains in `ai.repo.yaml` for future retries. Remove it with `repo drop-source` if no longer needed.

---

## Removing Sources

Use `repo drop-source` to remove a source and optionally clean up its resources.

### Basic Usage

```bash
# Remove by name
aimgr repo drop-source my-source

# Remove by path (local sources)
aimgr repo drop-source ~/my-resources/

# Remove by URL (remote sources)
aimgr repo drop-source https://github.com/owner/repo
```

### Options

| Flag | Description |
|------|-------------|
| `--keep-resources` | Keep resources, only remove source entry |
| `--dry-run` | Preview what would be removed |

### Behavior

By default, `drop-source`:
1. Removes the source entry from `ai.repo.yaml`
2. Deletes resources that came from that source (orphan cleanup)

Use `--keep-resources` to preserve resources (they become "untracked").

### Examples

```bash
# Preview removal
aimgr repo drop-source my-source --dry-run

# Remove source and its resources
aimgr repo drop-source my-source

# Remove source but keep resources
aimgr repo drop-source my-source --keep-resources
```

---

## Commands Reference

### repo add

Import resources from a source and track it in `ai.repo.yaml`.

```bash
aimgr repo add <source> [flags]
```

**Options:**

| Flag | Description |
|------|-------------|
| `--name=<name>` | Custom name for source |
| `--filter=<pattern>` | Only import matching resources |
| `--force` | Overwrite existing resources |
| `--skip-existing` | Skip existing resources |
| `--dry-run` | Preview without importing |
| `--format=<format>` | Output format |

### repo sync

Re-import resources from all configured sources.

```bash
aimgr repo sync [flags]
```

### repo drop-source

Remove a source and optionally clean up orphaned resources.

```bash
aimgr repo drop-source <name|path|url> [flags]
```

### repo info

Display repository information including configured sources.

```bash
aimgr repo info [flags]
```

**Example output:**
```
Repository: /home/user/.local/share/ai-config/repo

Resources:
  Commands: 12
  Skills:   8
  Agents:   3
  Packages: 2
  Total:    25

Configured Sources (2):
  my-local-skills (symlink)
    Path: /home/user/dev/my-skills
    Last synced: 2026-02-14 15:45:00
    
  community-catalog (copy)
    URL: https://github.com/owner/repo@v1.2.0
    Last synced: 2026-02-14 15:45:00
```

### repo drop

Delete the entire repository.

```bash
aimgr repo drop [flags]
```

| Flag | Description |
|------|-------------|
| `--force` | Required confirmation |
| `--full-delete` | Delete everything including `ai.repo.yaml` |

**Recovery after soft drop:**
```bash
# ai.repo.yaml is preserved, rebuild from sources
aimgr repo sync
```

---

## Workflows

### Initial Setup

```bash
# Add your first source
aimgr repo add gh:agentskills/catalog

# Check what was imported
aimgr repo list

# View source configuration
aimgr repo info

# Install resources in your project
cd ~/my-project
aimgr install skill/pdf-processing
```

### Multi-Source Setup

```bash
# Add community resources
aimgr repo add gh:anthropics/skills --name=anthropic-skills

# Add company resources
aimgr repo add gh:mycompany/ai-resources --name=company-resources

# Add local development resources
aimgr repo add ~/dev/my-skills --name=dev-skills

# Verify all sources
aimgr repo info
```

### Promoting Local to Remote

When ready to share your local resources:

```bash
# Currently using local source
cd ~/dev/my-skills

# Push to GitHub
git init && git add . && git commit -m "Initial commit"
git remote add origin https://github.com/myuser/my-skills.git
git push -u origin main
git tag v1.0.0 && git push origin v1.0.0

# Remove local source, add remote
aimgr repo drop-source my-skills --keep-resources
aimgr repo add gh:myuser/my-skills@v1.0.0 --name=my-skills --force
```

### Converting Remote to Local

For active development on an upstream repository:

```bash
# Clone locally
git clone https://github.com/owner/repo ~/dev/upstream-repo

# Remove remote source
aimgr repo drop-source upstream-repo

# Add as local source (symlink mode)
aimgr repo add ~/dev/upstream-repo --name=upstream-repo

# Now changes reflect immediately via symlinks
```

---

## Troubleshooting

### Source Not Found After Moving Directory

**Problem:** Local source was moved to a new location.

```
Error syncing source 'my-skills': path does not exist: /old/path/my-skills
```

**Solutions:**

```bash
# Option 1: Edit ai.repo.yaml
vim ~/.local/share/ai-config/repo/ai.repo.yaml
# Change path: /old/path/my-skills to /new/path/my-skills

# Option 2: Remove and re-add
aimgr repo drop-source my-skills --keep-resources
aimgr repo add /new/path/my-skills --name=my-skills --force
```

### Remote Repository Unavailable

**Problem:** Remote repository was deleted, made private, or URL changed.

```
Error syncing source 'company-resources': failed to clone: repository not found
```

**Solutions:**

```bash
# If temporarily unavailable - other sources still sync
# Sync will continue and report the error

# If permanently unavailable
aimgr repo drop-source company-resources --keep-resources

# If URL changed - edit ai.repo.yaml
vim ~/.local/share/ai-config/repo/ai.repo.yaml
```

### Git Not Installed

**Problem:** Git executable not found.

```
git clone failed: exec: "git": executable file not found
```

**Solution:** Install Git:

```bash
# Ubuntu/Debian
sudo apt-get install git

# macOS
brew install git
```

### Authentication Failed

**Problem:** Cannot access private repository.

```
git clone failed: authentication required
```

**Solutions:**

```bash
# For HTTPS - configure credential helper
git config --global credential.helper store

# For SSH - ensure keys are configured
ssh -T git@github.com
```

### No Resources Found

**Problem:** Repository exists but no resources discovered.

```
Error: no resources found in repository
```

**Solutions:**

- Check repository structure matches expected locations
- Use a subpath: `aimgr repo add https://github.com/owner/repo/path/to/resources`
- Verify resources have valid frontmatter (SKILL.md with name and description)

### Duplicate Source Names

**Problem:** Source name collision.

```
Error: source with name 'my-source' already exists
```

**Solutions:**

```bash
# Use custom name
aimgr repo add ~/new-resources --name=my-source-v2

# Or remove old source first
aimgr repo drop-source my-source
aimgr repo add ~/new-resources
```

### Corrupted ai.repo.yaml

**Problem:** Invalid YAML syntax.

```
Error: failed to load manifest: invalid YAML
```

**Solutions:**

```bash
# Validate syntax
yamllint ~/.local/share/ai-config/repo/ai.repo.yaml

# Or restore from backup
cp ~/.local/share/ai-config/repo/ai.repo.yaml.backup \
   ~/.local/share/ai-config/repo/ai.repo.yaml

# Or recreate from scratch
aimgr repo drop --force
aimgr repo add gh:owner/repo --name=source1
```

### Symlinks Not Working on Windows

**Problem:** Windows requires special permissions for symlinks.

**Solutions:**

```bash
# Option 1: Enable Developer Mode
# Settings > Update & Security > For Developers > Developer Mode

# Option 2: Use remote source instead (copy mode)
# Push local source to Git, then add as URL source
aimgr repo drop-source my-source --keep-resources
aimgr repo add https://github.com/myuser/my-resources --name=my-source --force
```

---

## Best Practices

### 1. Use Descriptive Source Names

```bash
# Good - clear and descriptive
aimgr repo add gh:owner/repo --name=community-pdf-tools

# Less clear - auto-generated name
aimgr repo add gh:owner/repo  # → "repo"
```

### 2. Pin Remote Sources to Versions

```bash
# Stable - pinned to version
aimgr repo add gh:owner/repo#v1.2.0

# Risky - tracks latest (may break)
aimgr repo add gh:owner/repo
```

### 3. Use Filters for Selective Importing

```bash
# Only import skills
aimgr repo add gh:owner/all-resources --filter "skill/*"
```

### 4. Separate Development and Production

```yaml
# ai.repo.yaml
sources:
  # Production (versioned, copy mode)
  - name: prod-resources
    url: https://github.com/company/resources
    ref: v2.1.0
    
  # Development (local, symlink mode)
  - name: dev-resources
    path: /home/user/dev/resources
```

### 5. Preview Before Destructive Operations

```bash
aimgr repo drop-source old-source --dry-run
aimgr repo sync --dry-run
```

### 6. Back Up ai.repo.yaml

```bash
# Manual backup
cp ~/.local/share/ai-config/repo/ai.repo.yaml \
   ~/.local/share/ai-config/repo/ai.repo.yaml.backup

# Or version control it
cd ~/.local/share/ai-config/repo
git init && git add ai.repo.yaml && git commit -m "Backup sources"
```

### 7. Sync Regularly

```bash
# Keep resources up-to-date
aimgr repo sync

# Or automate with cron
0 9 * * * /usr/local/bin/aimgr repo sync
```

---

## Related Documentation

- [Getting Started](./getting-started.md) - First steps with aimgr
- [Configuration Guide](./configuration.md) - Global and project configuration
- [Pattern Matching](../reference/pattern-matching.md) - Filter pattern syntax
- [Supported Tools](../reference/supported-tools.md) - Tool support and resource format documentation
