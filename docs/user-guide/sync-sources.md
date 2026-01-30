# Sync Sources Guide

This guide explains how to configure sync sources for automatically importing resources from remote repositories or local directories.

## Overview

The `aimgr repo sync` command allows you to automatically import resources from multiple sources defined in your configuration file. Sources can be either:

- **Remote URLs** (`url` field): Git repositories that are cloned and copied to your repository
- **Local paths** (`path` field): Local directories that are symlinked for live editing

## Source Types

### Remote URL Sources

Remote sources use the `url` field and are fetched from Git repositories. Resources are **copied** to your aimgr repository.

```yaml
sync:
  sources:
    - url: "https://github.com/anthropics/skills"
    - url: "gh:myorg/company-resources@v1.0.0"
      filter: "skill/*"
```

**Behavior:**
- Repository is cloned to workspace cache (see [workspace-caching.md](./workspace-caching.md))
- Resources are **copied** from workspace to your aimgr repository
- Changes to the source require running `aimgr repo sync` again to update
- Supports version pinning with `@branch` or `@tag` syntax

**Use cases:**
- Importing community resources from GitHub
- Pulling company-wide resources from organization repositories
- Using versioned, stable resources in production
- Sharing resources across teams

**Supported URL formats:**
```yaml
# GitHub shorthand
- url: "gh:owner/repo"
- url: "gh:owner/repo@v1.0.0"
- url: "gh:owner/repo/path/to/resources@main"

# Owner/repo shorthand (gh: implied)
- url: "owner/repo"

# Full Git URLs
- url: "https://github.com/owner/repo.git"
- url: "git@github.com:owner/repo.git"

# GitLab and other Git hosts
- url: "https://gitlab.com/group/project.git"
```

### Local Path Sources

Local sources use the `path` field and point to directories on your filesystem. Resources are **symlinked** to your aimgr repository.

```yaml
sync:
  sources:
    - path: "/home/user/my-skills"
    - path: "~/dev/custom-agents"
      filter: "agent/*"
```

**Behavior:**
- Resources are **symlinked** from the source directory to your aimgr repository
- Changes to source files are immediately reflected (live editing)
- No need to re-run `aimgr repo sync` after editing
- Original files remain in their location

**Use cases:**
- Active development of new resources
- Local customization and experimentation
- Personal resource collections
- Quick iteration without re-importing

**Supported path formats:**
```yaml
# Absolute paths
- path: "/home/user/resources"
- path: "/opt/shared/ai-tools"

# Home directory expansion
- path: "~/my-skills"
- path: "~/dev/projects/ai-resources"

# Relative paths (resolved from current working directory)
- path: "./local-resources"
- path: "../shared-resources"
```

## Configuration

### Basic Configuration

Add sources to your global config file (`~/.config/aimgr/aimgr.yaml`):

```yaml
sync:
  sources:
    # Remote source (copied)
    - url: "gh:anthropics/skills"
    
    # Local source (symlinked)
    - path: "~/my-custom-skills"
```

### Filtering Resources

Use the optional `filter` field to limit which resources are imported from a source:

```yaml
sync:
  sources:
    # Only import skills
    - url: "gh:myorg/all-resources"
      filter: "skill/*"
    
    # Only import commands with "test" in name
    - path: "~/dev-resources"
      filter: "command/*test*"
    
    # Only import specific skill
    - url: "gh:community/tools"
      filter: "skill/pdf-processor"
```

**Filter patterns:**
- `"skill/*"` - All skills
- `"command/*"` - All commands
- `"agent/*"` - All agents
- `"*test*"` - Resources with "test" in name
- `"skill/pdf*"` - Skills starting with "pdf"

See [pattern-matching.md](./pattern-matching.md) for complete filter syntax.

### Complete Example

```yaml
# ~/.config/aimgr/aimgr.yaml

repo:
  path: ~/ai-resources

install:
  targets:
    - claude
    - opencode

sync:
  sources:
    # Public community skills (remote, copied)
    - url: "gh:anthropics/skills"
    
    # Company resources, specific version (remote, copied)
    - url: "gh:mycompany/internal-tools@v2.1.0"
      filter: "skill/*"
    
    # Local development (symlinked for live editing)
    - path: "~/dev/my-skills"
      filter: "skill/*"
    
    # Shared team commands (symlinked)
    - path: "/mnt/shared/team-commands"
      filter: "command/*"
```

## Running Sync

Once sources are configured, sync them with:

```bash
# Sync all sources (overwrites existing)
aimgr repo sync

# Skip resources that already exist
aimgr repo sync --skip-existing

# Preview what would be synced
aimgr repo sync --dry-run

# JSON output for scripting
aimgr repo sync --format=json
```

## Development Workflow

### Live Editing with Local Sources

Local path sources enable immediate feedback during development:

```yaml
sync:
  sources:
    - path: "~/dev/my-skills"
```

```bash
# 1. Sync once to create symlinks
aimgr repo sync

# 2. Edit your skill directly
vim ~/dev/my-skills/pdf-processor/SKILL.md

# 3. Changes are immediately available (no re-sync needed!)
aimgr install skill/pdf-processor --tool=claude

# 4. Test in your AI tool
# (changes to SKILL.md are live)
```

### Promoting Local Resources to Remote

When ready to share local resources:

```bash
# 1. Develop locally with symlinks
vim ~/my-skills/new-skill/SKILL.md

# 2. Test thoroughly
aimgr install skill/new-skill --tool=claude

# 3. Push to Git repository
cd ~/my-skills
git add .
git commit -m "Add new-skill"
git push origin main

# 4. Update config to use remote URL
# Change from:
#   - path: "~/my-skills"
# To:
#   - url: "gh:myuser/my-skills@v1.0.0"

# 5. Re-sync to switch from symlink to copy
aimgr repo sync --skip-existing
```

### Mixed Development Workflow

Use both local and remote sources together:

```yaml
sync:
  sources:
    # Stable remote resources (copied)
    - url: "gh:anthropics/skills@v2.0.0"
    - url: "gh:mycompany/approved-tools@stable"
    
    # Development resources (symlinked)
    - path: "~/dev/experimental-skills"
      filter: "skill/experimental-*"
```

This gives you:
- Stable, versioned resources from remote sources
- Live editing for resources under development
- Clear separation between production and development

## Symlink Behavior

### How Symlinks Work

When using local path sources, aimgr creates symlinks in your repository:

```
~/.local/share/ai-config/repo/
├── skills/
│   ├── stable-skill/          # Copied from remote URL
│   │   └── SKILL.md
│   └── dev-skill/             # Symlink to local path
│       └── SKILL.md -> ~/dev/my-skills/dev-skill/SKILL.md
```

### Benefits

- **Immediate changes**: Edit source files and changes are instantly visible
- **Single source of truth**: No duplication, no sync lag
- **Version control**: Keep resources in their own Git repos
- **Easy cleanup**: Remove symlink without affecting source

### Limitations

- **Source must exist**: If source directory is moved/deleted, symlink breaks
- **Platform-specific**: Symlinks work best on Linux/macOS, limited on Windows
- **Relative paths**: Symlinks use absolute paths, not portable across machines

## Troubleshooting

### Broken Symlinks

**Symptom:** Resources appear in `aimgr repo list` but fail to install or show errors.

**Cause:** Source directory was moved, renamed, or deleted.

**Solution:**

```bash
# Check for broken symlinks
aimgr repo verify

# Option 1: Update config with new path
vim ~/.config/aimgr/aimgr.yaml
# Update path: field to new location
aimgr repo sync

# Option 2: Remove broken symlinks
aimgr repo remove skill/broken-skill
aimgr repo sync

# Option 3: Convert to remote URL source
# Change from:
#   - path: "/old/path"
# To:
#   - url: "gh:user/repo"
aimgr repo sync
```

### Source Not Found

**Symptom:** `aimgr repo sync` reports "no resources found" for a source.

**Cause:** Resources not in expected locations.

**Solution:**

```bash
# Check directory structure
ls -la ~/my-skills/

# Resources must be in standard locations:
# - commands/
# - skills/
# - agents/
# Or tool-specific locations:
# - .claude/commands/, .claude/skills/, etc.

# See resource-formats.md for details
```

### Permission Denied

**Symptom:** Cannot create symlinks or copy resources.

**Cause:** Insufficient permissions on source or destination.

**Solution:**

```bash
# Check source permissions
ls -ld ~/my-skills
chmod 755 ~/my-skills

# Check repository permissions
ls -ld ~/.local/share/ai-config/repo
chmod 755 ~/.local/share/ai-config/repo

# Check if running as correct user
whoami
```

### Symlink vs Copy Confusion

**Symptom:** Edited a local file but changes not showing up.

**Cause:** Source is using `url` field (copy) instead of `path` field (symlink).

**Solution:**

```bash
# Check your config
cat ~/.config/aimgr/aimgr.yaml

# Wrong (copies files):
sync:
  sources:
    - url: "~/my-skills"  # Local path in URL field = copy

# Correct (creates symlinks):
sync:
  sources:
    - path: "~/my-skills"  # Local path in path field = symlink

# Re-sync after fixing config
aimgr repo remove skill/my-skill --force
aimgr repo sync
```

### Workspace Cache Issues

**Symptom:** Remote sources show stale content after upstream changes.

**Cause:** Workspace cache contains old clone.

**Solution:**

```bash
# Clear workspace cache and re-sync
aimgr repo prune
aimgr repo sync

# Or force fresh clone by bumping version tag
# Change from:
#   - url: "gh:owner/repo@v1.0.0"
# To:
#   - url: "gh:owner/repo@v1.1.0"
```

## Migration Guide

### Migrating from Old Config Format

**Old format (ambiguous):**

```yaml
sync:
  sources:
    - url: "/home/user/skills"           # Local path in url field
    - url: "https://github.com/org/repo" # Remote in url field
```

**New format (explicit):**

```yaml
sync:
  sources:
    - path: "/home/user/skills"          # Local = path field (symlink)
    - url: "https://github.com/org/repo" # Remote = url field (copy)
```

**Migration steps:**

1. **Backup your config:**
   ```bash
   cp ~/.config/aimgr/aimgr.yaml ~/.config/aimgr/aimgr.yaml.backup
   ```

2. **Update config format:**
   ```bash
   vim ~/.config/aimgr/aimgr.yaml
   # Change local sources from `url:` to `path:`
   ```

3. **Remove old resources:**
   ```bash
   # Remove resources that should be symlinked
   aimgr repo remove skill/my-local-skill --force
   ```

4. **Re-sync with new config:**
   ```bash
   aimgr repo sync
   ```

5. **Verify:**
   ```bash
   aimgr repo list
   # Check that local resources show correct type
   ```

### Converting Remote to Local

To switch a remote source to local development:

```bash
# 1. Clone the remote repository locally
git clone https://github.com/owner/repo ~/local/repo

# 2. Update config
# Change from:
#   - url: "gh:owner/repo"
# To:
#   - path: "~/local/repo"

# 3. Remove old (copied) resources
aimgr repo remove skill/the-skill --force

# 4. Sync to create symlinks
aimgr repo sync

# 5. Verify symlink was created
ls -l ~/.local/share/ai-config/repo/skills/the-skill
```

### Converting Local to Remote

To publish local resources:

```bash
# 1. Create Git repository
cd ~/my-skills
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/user/my-skills.git
git push -u origin main

# 2. Tag a version
git tag v1.0.0
git push origin v1.0.0

# 3. Update config
# Change from:
#   - path: "~/my-skills"
# To:
#   - url: "gh:user/my-skills@v1.0.0"

# 4. Sync to switch from symlink to copy
aimgr repo sync
```

## Best Practices

### 1. Use Remote URLs for Stable Resources

Pin production resources to specific versions:

```yaml
sync:
  sources:
    - url: "gh:company/approved-tools@v2.1.0"  # Pinned version
    - url: "gh:community/skills@stable"         # Stable branch
```

### 2. Use Local Paths for Development

Enable fast iteration during development:

```yaml
sync:
  sources:
    - path: "~/dev/experimental-skills"
      filter: "skill/experimental-*"
```

### 3. Keep Sources Organized

Group related sources and use clear comments:

```yaml
sync:
  sources:
    # Production resources (stable, versioned)
    - url: "gh:company/prod-tools@v3.0.0"
      filter: "skill/*"
    
    # Development resources (live editing)
    - path: "~/dev/new-features"
      filter: "skill/feature-*"
    
    # Shared team resources (network mount)
    - path: "/mnt/shared/team-tools"
```

### 4. Use Filters to Avoid Conflicts

Prevent resource name collisions with specific filters:

```yaml
sync:
  sources:
    # Production skills
    - url: "gh:company/tools@v2.0.0"
      filter: "skill/*"
    
    # Dev skills with prefix
    - path: "~/dev/skills"
      filter: "skill/dev-*"
```

### 5. Document Source Purposes

Add comments explaining each source:

```yaml
sync:
  sources:
    # Official Anthropic skills - community maintained
    - url: "gh:anthropics/skills@v2.0.0"
    
    # Company internal tools - requires VPN access
    - url: "gh:mycompany/internal@stable"
      filter: "skill/company-*"
    
    # My personal development - not for production
    - path: "~/dev/personal-skills"
      filter: "skill/personal-*"
```

## Advanced Usage

### Conditional Sources with Environment Variables

Use environment variable interpolation for dynamic configs:

```yaml
sync:
  sources:
    # Production in CI/CD
    - url: "${PROD_REPO:-gh:company/tools@v1.0.0}"
    
    # Development on local machine
    - path: "${DEV_PATH:-~/dev/skills}"
```

```bash
# Production environment
export PROD_REPO="gh:company/tools@v2.0.0"
aimgr repo sync

# Development environment
export DEV_PATH="/home/dev/my-skills"
aimgr repo sync
```

### Network-Mounted Sources

Use symlinks for network-mounted team resources:

```yaml
sync:
  sources:
    # NFS or SMB mount
    - path: "/mnt/team-share/ai-resources"
      filter: "skill/team-*"
```

Benefits:
- Team-wide resource sharing
- Centralized updates
- No Git repository needed

### Monorepo Support

Import from monorepos with path filters:

```yaml
sync:
  sources:
    # Frontend skills only
    - url: "gh:company/monorepo/ai-tools/frontend@main"
      filter: "skill/*"
    
    # Backend skills only
    - url: "gh:company/monorepo/ai-tools/backend@main"
      filter: "skill/*"
```

## Related Documentation

- [configuration.md](./configuration.md) - Complete configuration guide
- [workspace-caching.md](./workspace-caching.md) - Git repository caching details
- [pattern-matching.md](./pattern-matching.md) - Filter pattern syntax
- [github-sources.md](./github-sources.md) - GitHub source formats
- [resource-formats.md](./resource-formats.md) - Resource structure requirements
