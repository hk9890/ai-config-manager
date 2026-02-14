# GitHub Sources Guide

This guide provides detailed information about using GitHub sources with `aimgr`.

## Overview

`aimgr` can add resources directly from GitHub repositories, making it easy to discover, share, and use community-contributed resources. When you add a resource from GitHub, `aimgr`:

1. Clones the repository to a workspace cache
2. Auto-discovers resources in standard locations
3. Copies resources to your centralized repository (`~/.local/share/ai-config/repo/`)
4. Tracks the source in `ai.repo.yaml` for future syncing

GitHub sources are stored in **copy mode** (not symlink mode), meaning resources are copied to your repository rather than symlinked.

## Source Format Syntax

### Basic GitHub Source

```bash
aimgr repo add https://github.com/owner/repo
```

This clones the repository and auto-discovers all resources (skills, commands, agents).

### GitHub Source with Ref (Branch/Tag)

```bash
# With branch
aimgr repo add https://github.com/owner/repo#branch-name

# With tag
aimgr repo add https://github.com/owner/repo#v1.0.0
```

Uses a specific branch or tag reference.

### GitHub Source with Subpath

```bash
aimgr repo add https://github.com/owner/repo/path/to/resources
```

Looks for resources in a specific path within the repository.

### Combined Subpath and Reference

```bash
aimgr repo add https://github.com/owner/repo/skills/my-skill#v2.0.0
```

Uses both a specific subpath and a version reference.

### Custom Source Name

```bash
# Auto-generated name (e.g., "owner-repo")
aimgr repo add https://github.com/owner/repo

# Custom name
aimgr repo add https://github.com/owner/repo --name=my-gh-source
```

Custom names make it easier to identify and manage sources.

### Git URL Sources

Use any Git URL directly:

```bash
# HTTPS
aimgr repo add https://github.com/owner/repo.git
aimgr repo add https://gitlab.com/owner/repo.git

# SSH
aimgr repo add git@github.com:owner/repo.git

# With ref
aimgr repo add https://github.com/owner/repo.git#develop
```

## Source Tracking in ai.repo.yaml

When you add a GitHub source, it's automatically tracked in `ai.repo.yaml` at the root of your repository. This manifest file records:

- Source URL and ref
- Subpath (if specified)
- Source name

**Example ai.repo.yaml entry:**

```yaml
sources:
  - name: owner-repo
    url: https://github.com/owner/repo
    ref: main
  
  - name: my-skills
    url: https://github.com/myorg/skills
    ref: v2.0.0
    subpath: skills/frontend
```

**Note:** Import mode is **implicit** based on source type. URL sources (like GitHub) always use **copy mode** (resources are copied to your repository). Path sources use **symlink mode** (resources are symlinked for live editing). You don't need to configure this - it's automatic.

### Syncing GitHub Sources

Once tracked, you can sync sources to pull latest changes:

```bash
# Sync all sources
aimgr repo sync

# Sync specific source
aimgr repo sync my-skills

# Sync with specific ref
aimgr repo sync owner-repo --ref=v2.1.0
```

Syncing updates resources in your repository to match the current state of the GitHub source.

### Removing GitHub Sources

Remove sources and optionally clean up their resources:

```bash
# Remove by name (keeps resources)
aimgr repo remove my-gh-source

# Remove by URL (keeps resources)
aimgr repo remove https://github.com/owner/repo

# Remove and delete orphaned resources
aimgr repo remove my-gh-source --clean-orphans
```

When you remove a source, its entry is removed from `ai.repo.yaml`. By default, resources are kept in the repository. Use `--clean-orphans` to delete resources that are no longer tracked by any source.

## Auto-Discovery Algorithm

When adding resources from GitHub, `aimgr` searches for resources in standard locations following tool conventions.

### Skills Discovery

**Priority search order:**

1. **Direct path** (if subpath specified):
   - `SKILL.md` at the exact path: `repo/path/to/skill/SKILL.md`

2. **Standard directories** (searched in order):
   - `skills/`
   - `.claude/skills/`
   - `.opencode/skills/`
   - `.github/skills/`
   - `.codex/skills/`
   - `.cursor/skills/`
   - `.goose/skills/`
   - `.kilocode/skills/`
   - `.kiro/skills/`
   - `.roo/skills/`
   - `.trae/skills/`
   - `.agents/skills/`
   - `.agent/skills/`

3. **Recursive search** (if not found above):
   - Searches up to 5 levels deep
   - Looks for any directory containing `SKILL.md`

**Validation:**
- Each skill must have a valid `SKILL.md` file
- `SKILL.md` must have YAML frontmatter with `name` and `description`
- Directory name must match the `name` field in `SKILL.md`

**Deduplication:**
- If multiple skills with the same name are found, the first one (by priority order) is used

### Commands Discovery

**Search locations:**
1. `commands/`
2. `.claude/commands/`
3. `.opencode/commands/`
4. Recursive search for `.md` files (max depth 5)

**Filtering:**
- Only `.md` files are considered
- `SKILL.md` and `README.md` are excluded
- Files must have valid command frontmatter (at minimum, a `description` field)

### Agents Discovery

**Search locations:**
1. `agents/`
2. `.claude/agents/`
3. `.opencode/agents/`
4. Recursive search for `.md` files (max depth 5)

**Validation:**
- Files must have valid agent frontmatter (at minimum, a `description` field)
- Deduplication by agent name

## Resource Selection

### Auto-Discovery in Action

When you add a GitHub source, all discoverable resources are automatically added:

```bash
$ aimgr repo add https://github.com/myorg/my-resources
Cloning repository...
Discovering resources...
✓ Added 2 skills, 3 commands, 1 agent
✓ Source 'myorg-my-resources' tracked in ai.repo.yaml
```

### Specific Subpath

To add resources from a specific directory, use a subpath:

```bash
$ aimgr repo add https://github.com/myorg/mono-repo/skills/frontend
Cloning repository...
Discovering resources in 'skills/frontend'...
✓ Added skill 'frontend-design'
✓ Source 'myorg-mono-repo' tracked in ai.repo.yaml
```

### No Resources Found

If no resources are found, an error is displayed:

```bash
$ aimgr repo add https://github.com/myorg/empty-repo
Cloning repository...
Error: no resources found in repository: https://github.com/myorg/empty-repo
```

Check the repository's structure or documentation to find where resources are located. You may need to specify a subpath.

## Precedence Rules

When multiple search paths exist, resources are discovered in this precedence:

1. **Explicit subpath** (if provided in source)
2. **Standard tool directories** (in order: skills/, .claude/skills/, .opencode/skills/, etc.)
3. **Recursive search** (fallback, max depth 5)

The first valid resource found wins.

## Examples

### Example 1: Add All Resources from GitHub

```bash
# Add entire repository (auto-discovers all resources)
aimgr repo add https://github.com/myorg/my-resources
```

### Example 2: Add with Custom Name

```bash
# Custom name for easier management
aimgr repo add https://github.com/myorg/resources --name=my-tools
```

### Example 3: Add from Specific Version

```bash
# Add from a tagged release
aimgr repo add https://github.com/myorg/skills#v2.1.0

# Add from a branch
aimgr repo add https://github.com/myorg/experimental-skills#develop
```

### Example 4: Add Specific Subpath

```bash
# Add resources from specific directory
aimgr repo add https://github.com/myorg/mono-repo/skills/frontend

# Add with both subpath and ref
aimgr repo add https://github.com/myorg/repo/commands#v1.0.0
```

### Example 5: Use Git URLs

```bash
# HTTPS URL
aimgr repo add https://github.com/myorg/custom-skills.git

# SSH URL (if you have keys configured)
aimgr repo add git@github.com:myorg/private-skills.git

# GitLab or other Git hosting
aimgr repo add https://gitlab.com/mygroup/skills.git
```

### Example 6: Sync and Update Sources

```bash
# Add source
aimgr repo add https://github.com/myorg/skills --name=my-skills

# Later, sync to get latest changes
aimgr repo sync my-skills

# Update to different version
aimgr repo sync my-skills --ref=v2.0.0
```

### Example 7: Remove Sources

```bash
# Remove source (keeps resources)
aimgr repo remove my-skills

# Remove and clean up orphaned resources
aimgr repo remove my-skills --clean-orphans

# Remove by URL
aimgr repo remove https://github.com/myorg/skills
```

## Centralized Storage

All resources added from GitHub are stored in your centralized repository:

```
~/.local/share/ai-config/repo/
├── ai.repo.yaml          # Source tracking manifest
├── commands/
│   └── deploy.md
├── skills/
│   ├── frontend-design/
│   └── pdf-processing/
└── agents/
    └── code-reviewer.md
```

This means:
- Resources are downloaded once and stored centrally
- Sources are tracked in `ai.repo.yaml` for easy syncing
- GitHub sources use **copy mode** (resources are copied, not symlinked)
- You can sync sources to pull latest changes from GitHub
- No duplication across projects when you install resources to project directories

## Performance Considerations

### Workspace Caching

`aimgr` uses a workspace cache to dramatically improve performance for Git operations:

**First add** (cold cache):
```bash
$ aimgr repo add https://github.com/large-org/big-repo
Cloning repository...  # Takes ~30 seconds for large repos
Discovering resources...
✓ Added resources
```

**Subsequent syncs** (warm cache):
```bash
$ aimgr repo sync large-org-big-repo
Using cached repository...  # Takes <1 second
Pulling latest changes...
✓ Synced resources
```

**Benefits:**
- First clone is slower (full git clone)
- Subsequent operations are 10-50x faster (uses cached repository)
- Cache is stored in `~/.cache/aimgr/workspace/` (XDG cache directory)
- Safe to delete cache - will re-clone on next sync

### Shallow Clones

For initial clones, `aimgr` uses shallow clones (`git clone --depth 1`) for speed:
- Only downloads the latest commit
- Significantly faster for large repositories
- Sufficient for resource discovery

### Storage Efficiency

- Workspace cache stores cloned repositories for reuse
- Resources are copied to your repository (not symlinked for GitHub sources)
- Cache grows with number of unique GitHub sources you add
- Clear cache with: `rm -rf ~/.cache/aimgr/workspace/`

## Troubleshooting

### Problem: Git not installed

**Error:** `git clone failed: exec: "git": executable file not found`

**Solution:** Install Git:
```bash
# Ubuntu/Debian
sudo apt-get install git

# macOS
brew install git

# Verify installation
git --version
```

### Problem: Repository not found

**Error:** `git clone failed: repository not found`

**Solutions:**
- Verify the repository URL is correct
- Check that the repository is public (or you have access)
- For private repos, ensure SSH keys or credentials are configured

### Problem: Network timeout

**Error:** `git clone failed: connection timed out`

**Solutions:**
- Check your internet connection
- Try again (may be temporary network issue)
- For large repos, consider cloning manually first

### Problem: Invalid credentials

**Error:** `git clone failed: authentication required`

**Solutions:**
- For HTTPS URLs, configure Git credentials:
  ```bash
  git config --global credential.helper store
  ```
- For SSH URLs, ensure your SSH keys are configured:
  ```bash
  ssh -T git@github.com
  ```

### Problem: Resource not found after clone

**Error:** `no resources found in repository`

**Solutions:**
- Check the repository structure - resources must be in standard locations
- Use a specific subpath: `aimgr repo add https://github.com/owner/repo/path/to/resources`
- Verify the resource has valid frontmatter (SKILL.md with name and description)
- Clone manually and inspect: `git clone https://github.com/owner/repo && ls -R`

### Problem: Source already exists

**Error:** `source already exists: my-source`

**Solutions:**
- Use a different name: `aimgr repo add URL --name=my-source-2`
- Remove existing source first: `aimgr repo remove my-source`
- Update existing source: `aimgr repo sync my-source --ref=new-version`

## Best Practices

### 1. Use Custom Names for Clarity

Give sources meaningful names for easier management:

```bash
# Good - clear name
aimgr repo add https://github.com/myorg/skills --name=company-skills

# Auto-generated name works but less clear
aimgr repo add https://github.com/myorg/skills  # → "myorg-skills"
```

### 2. Pin to Version Tags for Stability

For production use, pin to specific versions:

```bash
# Good - pinned version
aimgr repo add https://github.com/myorg/skills#v1.2.0

# Risky - uses latest code
aimgr repo add https://github.com/myorg/skills
```

### 3. Use Subpaths for Large Repositories

For mono-repos, specify the exact subpath:

```bash
# Specific - only gets what you need
aimgr repo add https://github.com/myorg/mono-repo/skills/frontend

# Generic - may pull unnecessary resources
aimgr repo add https://github.com/myorg/mono-repo
```

### 4. Track Sources in Version Control

Commit your `ai.repo.yaml` to share sources with your team:

```bash
git add ai.repo.yaml
git commit -m "Add GitHub sources for project resources"
git push
```

Team members can then sync sources:

```bash
aimgr repo sync  # Syncs all sources in ai.repo.yaml
```

### 5. Regular Syncing for Updates

Keep resources up-to-date by syncing periodically:

```bash
# Sync all sources
aimgr repo sync

# Or sync specific source
aimgr repo sync company-skills
```

### 6. Check Repository Documentation

Before adding, check the repository's README for:
- Resource locations (if not in standard directories)
- Available resources and their purposes
- Version compatibility
- Recommended version tags

## Contributing GitHub Sources

If you're creating resources to share on GitHub:

### Repository Structure

Use standard directory structures for easy discovery:

```
my-skills-repo/
├── README.md
├── skills/
│   ├── skill-one/
│   │   └── SKILL.md
│   └── skill-two/
│       └── SKILL.md
├── commands/
│   ├── command-one.md
│   └── command-two.md
└── agents/
    ├── agent-one.md
    └── agent-two.md
```

Or use tool-specific directories:

```
my-skills-repo/
├── README.md
└── .claude/
    ├── skills/
    │   └── my-skill/
    │       └── SKILL.md
    ├── commands/
    │   └── my-command.md
    └── agents/
        └── my-agent.md
```

### Metadata Best Practices

**Skills (SKILL.md):**
```yaml
---
name: my-skill
description: Clear, concise description
version: "1.0.0"
license: MIT
metadata:
  author: Your Name
  tags: category, feature
---
```

**Commands:**
```yaml
---
description: What this command does
---
```

**Agents:**
```yaml
---
description: What this agent does
type: agent-role
---
```

### Version Tags

Use semantic versioning tags:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

Users can then reference specific versions:

```bash
aimgr repo add https://github.com/yourorg/your-repo#v1.0.0
```

### Documentation

Include in your README:
- List of resources and their purposes
- Installation instructions
- Resource locations (if not in standard directories)
- Version compatibility
- Examples

## Related Documentation

- [README.md](../README.md) - Main user documentation
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guide with GitHub source architecture
- [Source Formats](../README.md#source-formats) - Overview of all source formats
- [Auto-Discovery](../README.md#auto-discovery) - Summary of discovery algorithms
