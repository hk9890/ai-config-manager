# Source Management Guide

This guide explains how to manage sources in your aimgr repository using the `ai.repo.yaml` manifest file and source management commands.

---

## Overview

**aimgr** uses a **repository-local** approach to source management. Each repository tracks its own sources in an `ai.repo.yaml` manifest file, which is version-controlled and portable. This design enables:

- **Self-contained repositories**: All source information lives in the repository directory
- **Git-tracked manifests**: Share source configurations via version control
- **Automatic synchronization**: Re-import resources from all sources with a single command
- **Orphan cleanup**: Automatically remove resources when sources are removed

### Key Concepts

- **`ai.repo.yaml`**: Manifest file in your repository root that tracks all configured sources
- **Source**: A location (local path or remote URL) containing AI resources
- **Mode**: How resources are stored - `symlink` (live editing) or `copy` (stable snapshot)
- **Sync**: Re-import all resources from configured sources to get latest changes

---

## The ai.repo.yaml File

The `ai.repo.yaml` file is automatically created and maintained in your repository root (`~/.local/share/ai-config/repo/ai.repo.yaml` by default). It tracks all sources you've added and their metadata.

### File Format

```yaml
version: 1
sources:
  - name: my-local-commands
    path: /home/user/my-resources
    mode: symlink
    added: 2026-02-14T10:30:00Z
    last_synced: 2026-02-14T15:45:00Z
  - name: agentskills-catalog
    url: https://github.com/agentskills/catalog
    ref: main
    subpath: resources
    mode: copy
    added: 2026-02-14T11:00:00Z
    last_synced: 2026-02-14T15:45:00Z
```

### Field Reference

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `version` | integer | Manifest format version (currently 1) | Yes |
| `sources` | array | List of source configurations | Yes |
| `name` | string | Unique identifier for the source (auto-generated if not provided) | Yes |
| `path` | string | Absolute path to local directory (for local sources) | One of path/url |
| `url` | string | Git repository URL (for remote sources) | One of path/url |
| `ref` | string | Git branch/tag/commit (for remote sources) | No |
| `subpath` | string | Subdirectory within repository (for remote sources) | No |
| `mode` | string | Import mode: `symlink` or `copy` | Yes |
| `added` | timestamp | When source was first added | Yes |
| `last_synced` | timestamp | Last successful sync time | No |

### Source Types

#### Local Sources (`path`)

Local sources point to directories on your filesystem. Resources are typically **symlinked** for live editing.

```yaml
sources:
  - name: my-local-skills
    path: /home/user/dev/my-skills
    mode: symlink
    added: 2026-02-14T10:30:00Z
```

**Benefits:**
- Changes to source files immediately reflect in repository
- No need to re-sync after editing
- Perfect for active development
- Original files stay in their location

**Use cases:**
- Local development and testing
- Personal resource collections
- Quick iteration without re-importing

#### Remote Sources (`url`)

Remote sources point to Git repositories. Resources are **copied** to your repository.

```yaml
sources:
  - name: community-catalog
    url: https://github.com/owner/repo
    ref: v1.2.0
    mode: copy
    added: 2026-02-14T11:00:00Z
```

**Benefits:**
- Stable, versioned resources
- Works offline after initial import
- No dependency on remote availability
- Can pin to specific versions

**Use cases:**
- Production environments
- Shared team resources
- Versioned resource collections
- Community packages

### Import Modes

| Mode | Behavior | Best For |
|------|----------|----------|
| `symlink` | Creates symbolic links to source files | Local development, live editing |
| `copy` | Copies files to repository | Remote sources, stable deployments |

**Note:** Remote sources (URL) always use `copy` mode. Local sources (path) use `symlink` by default but can be forced to `copy` mode with the `--copy` flag.

---

## Commands

### repo add

Import resources from a source and track it in `ai.repo.yaml`.

```bash
aimgr repo add <source> [flags]
```

**Behavior:**
1. Discovers and imports all resources from the source
2. Adds source entry to `ai.repo.yaml`
3. Sets appropriate mode (`symlink` for local, `copy` for remote)
4. Records timestamps for tracking

#### Adding Local Sources

```bash
# Add from local directory (symlinked by default)
aimgr repo add ~/my-skills
aimgr repo add /opt/team-resources
aimgr repo add ./local-resources

# Force copy mode for local source
aimgr repo add ~/my-skills --copy
```

#### Adding Remote Sources

```bash
# Add from GitHub (various formats)
aimgr repo add gh:owner/repo
aimgr repo add owner/repo
aimgr repo add https://github.com/owner/repo
aimgr repo add git@github.com:owner/repo.git

# Add specific version
aimgr repo add gh:owner/repo@v1.0.0
aimgr repo add gh:owner/repo@main

# Add with subdirectory
aimgr repo add gh:owner/repo/resources@v1.0.0
```

#### Options

| Flag | Description |
|------|-------------|
| `--name=<name>` | Custom name for source (auto-generated if omitted) |
| `--copy` | Force copy mode for local paths (default: symlink) |
| `--filter=<pattern>` | Only import matching resources (e.g., `skill/*`) |
| `--force` | Overwrite existing resources |
| `--skip-existing` | Skip resources that already exist |
| `--dry-run` | Preview without importing |
| `--format=<format>` | Output format: table, json, yaml |

#### Examples

```bash
# Add with custom name
aimgr repo add ~/my-skills --name=personal-skills

# Add with filter
aimgr repo add gh:owner/repo --filter "skill/*"

# Preview before adding
aimgr repo add gh:owner/repo --dry-run

# Add and overwrite existing
aimgr repo add ~/resources --force
```

---

### repo drop-source

Remove a source from `ai.repo.yaml` and optionally clean up orphaned resources.

```bash
aimgr repo drop-source <name|path|url> [flags]
```

**Behavior:**
1. Finds source by name, path, or URL
2. Removes source entry from `ai.repo.yaml`
3. By default, removes orphaned resources (those that came from this source)

#### Matching Priority

Sources are matched in this order:
1. **Name** - The source's `name` field
2. **Path** - The source's `path` field (for local sources)
3. **URL** - The source's `url` field (for remote sources)

#### Options

| Flag | Description |
|------|-------------|
| `--keep-resources` | Keep resources, only remove source entry |
| `--dry-run` | Preview what would be removed |

#### Examples

```bash
# Remove source by name
aimgr repo drop-source my-source

# Remove source by path
aimgr repo drop-source ~/my-resources/

# Remove source by URL
aimgr repo drop-source https://github.com/owner/repo

# Preview removal
aimgr repo drop-source my-source --dry-run

# Remove source but keep resources
aimgr repo drop-source my-source --keep-resources
```

**Note:** By default, resources that came from the removed source are deleted. Use `--keep-resources` to preserve them (they become "untracked" resources).

---

### repo sync

Re-import resources from all configured sources in `ai.repo.yaml`.

```bash
aimgr repo sync [flags]
```

**Behavior:**
1. Reads all sources from `ai.repo.yaml`
2. For each source:
   - **Local sources**: Re-symlink from path
   - **Remote sources**: Download latest version, copy to repository
3. Updates `last_synced` timestamp for each source
4. By default, overwrites existing resources (force mode)

**When to use:**
- After upstream changes to remote repositories
- To refresh all sources at once
- After manually editing `ai.repo.yaml`
- To verify source availability

#### Options

| Flag | Description |
|------|-------------|
| `--skip-existing` | Skip resources that already exist (don't overwrite) |
| `--dry-run` | Preview without importing |
| `--format=<format>` | Output format: table, json, yaml |

#### Examples

```bash
# Sync all sources (overwrites existing)
aimgr repo sync

# Sync without overwriting
aimgr repo sync --skip-existing

# Preview sync
aimgr repo sync --dry-run

# Sync with JSON output
aimgr repo sync --format=json
```

#### Handling Unavailable Sources

If a source becomes unavailable (deleted directory, network error, etc.), `repo sync` will:
1. Report the error
2. Continue syncing remaining sources
3. Exit with error status

The failed source remains in `ai.repo.yaml` for future retries. Remove it with `repo drop-source` if no longer needed.

---

### repo info

Display repository information including configured sources.

```bash
aimgr repo info [flags]
```

**Output:**
- Repository location
- Total resource counts
- Configured sources with health status
- Last sync times

#### Options

| Flag | Description |
|------|-------------|
| `--format=<format>` | Output format: table, json, yaml |

#### Example Output

```bash
$ aimgr repo info

Repository: /home/user/.local/share/ai-config/repo

Resources:
  Commands: 12
  Skills:   8
  Agents:   3
  Packages: 2
  Total:    25

Configured Sources (2):
  ✓ my-local-skills (symlink)
    Path: /home/user/dev/my-skills
    Last synced: 2026-02-14 15:45:00
    
  ✓ community-catalog (copy)
    URL: https://github.com/owner/repo@v1.2.0
    Last synced: 2026-02-14 15:45:00
```

---

### repo drop

Delete the entire repository, including `ai.repo.yaml`.

```bash
aimgr repo drop [flags]
```

**Behavior:**
1. **Soft drop** (default): Removes all resources but keeps `ai.repo.yaml` and directory structure
2. **Full delete** (`--full-delete`): Completely removes repository directory

#### Options

| Flag | Description |
|------|-------------|
| `--force` | Required confirmation flag |
| `--full-delete` | Delete everything including `ai.repo.yaml` |

#### Examples

```bash
# Soft drop (keeps ai.repo.yaml)
aimgr repo drop --force

# Full delete (removes everything)
aimgr repo drop --force --full-delete

# After soft drop, rebuild from sources
aimgr repo sync
```

**Recovery after soft drop:**
If you performed a soft drop, `ai.repo.yaml` is preserved. Recover by running:
```bash
aimgr repo sync
```

---

## Workflows

### Initial Setup

Starting from scratch with a new repository:

```bash
# 1. Add your first source
aimgr repo add gh:agentskills/catalog

# 2. Check what was imported
aimgr repo list

# 3. View source configuration
aimgr repo info

# 4. Install resources in your project
cd ~/my-project
aimgr install skill/pdf-processing
```

---

### Adding Multiple Sources

Building up a repository with multiple sources:

```bash
# Add community resources
aimgr repo add gh:anthropics/skills --name=anthropic-skills

# Add company resources
aimgr repo add gh:mycompany/ai-resources --name=company-resources --filter "skill/*"

# Add local development resources
aimgr repo add ~/dev/my-skills --name=dev-skills

# Verify all sources
aimgr repo info

# List all resources
aimgr repo list
```

---

### Removing a Source

Cleaning up when you no longer need a source:

```bash
# Preview what would be removed
aimgr repo drop-source my-old-source --dry-run

# Remove source and its resources
aimgr repo drop-source my-old-source

# Or keep resources (make them untracked)
aimgr repo drop-source my-old-source --keep-resources

# Verify removal
aimgr repo info
```

---

### Syncing After Changes

Updating your repository when upstream sources change:

```bash
# Check current status
aimgr repo info

# Sync all sources
aimgr repo sync

# Or preview first
aimgr repo sync --dry-run

# Sync without overwriting local changes
aimgr repo sync --skip-existing
```

**Best practice:** Run `aimgr repo sync` regularly to keep resources up-to-date, especially for remote sources.

---

### Migrating Sources

Moving from local development to production:

#### Scenario: Promote local source to remote

```bash
# 1. Currently using local source
aimgr repo info
# Shows: my-skills (path: ~/dev/my-skills, mode: symlink)

# 2. Push to GitHub
cd ~/dev/my-skills
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/myuser/my-skills.git
git push -u origin main

# Tag a version
git tag v1.0.0
git push origin v1.0.0

# 3. Remove local source
aimgr repo drop-source my-skills --keep-resources

# 4. Add remote source
aimgr repo add gh:myuser/my-skills@v1.0.0 --name=my-skills --force

# 5. Verify change
aimgr repo info
# Shows: my-skills (url: gh:myuser/my-skills@v1.0.0, mode: copy)
```

#### Scenario: Convert remote to local for development

```bash
# 1. Currently using remote source
aimgr repo info
# Shows: upstream-repo (url: gh:owner/repo, mode: copy)

# 2. Clone locally
git clone https://github.com/owner/repo ~/dev/upstream-repo

# 3. Remove remote source
aimgr repo drop-source upstream-repo

# 4. Add as local source
aimgr repo add ~/dev/upstream-repo --name=upstream-repo

# 5. Now you can edit and changes reflect immediately
vim ~/dev/upstream-repo/skills/my-skill/SKILL.md
# Changes are live via symlinks
```

---

## Examples

### Example 1: Team Development Workflow

```bash
# Team lead sets up shared resources
aimgr repo add gh:mycompany/team-resources@stable --name=team-resources

# Developer adds local dev resources
aimgr repo add ~/dev/experimental --name=dev-experiments

# View configured sources
aimgr repo info
```

**Output:**
```
Repository: /home/user/.local/share/ai-config/repo

Resources:
  Commands: 5
  Skills:   12
  Total:    17

Configured Sources (2):
  ✓ team-resources (copy)
    URL: https://github.com/mycompany/team-resources@stable
    Last synced: 2026-02-14 10:30:00
    
  ✓ dev-experiments (symlink)
    Path: /home/user/dev/experimental
    Last synced: 2026-02-14 10:31:00
```

```bash
# Update to latest team resources
aimgr repo sync

# List all resources with installation status
aimgr repo list
```

---

### Example 2: Multi-Source Setup

```bash
# Add multiple sources at once
aimgr repo add gh:anthropics/skills --name=anthropic --filter "skill/*"
aimgr repo add gh:myorg/company-tools --name=company --filter "skill/*"
aimgr repo add ~/my-local-skills --name=local

# Preview sync
aimgr repo sync --dry-run
```

**Output:**
```
Syncing from 3 configured source(s)...
Mode: DRY RUN (preview only)

[1/3] Syncing source: anthropic
  Mode: Remote (download + copy)
  Would import: 8 skills

[2/3] Syncing source: company
  Mode: Remote (download + copy)
  Would import: 5 skills

[3/3] Syncing source: local
  Mode: Local (symlink)
  Would import: 3 skills

Summary (DRY RUN):
  Would import 16 resources total
```

```bash
# Actually sync
aimgr repo sync
```

---

### Example 3: Removing and Cleaning Up

```bash
# Add a temporary source
aimgr repo add ~/temp-resources --name=temp

# Later, remove it with orphan cleanup
aimgr repo drop-source temp --dry-run
```

**Output:**
```
Would remove source: temp
  Type: local
  Location: /home/user/temp-resources

Would remove 5 orphaned resource(s):
  - skill/test-skill
  - command/test-cmd
  - skill/another-test
  - agent/test-agent
  - command/temp-command
```

```bash
# Confirm removal
aimgr repo drop-source temp
```

**Output:**
```
✓ Removed source: temp
✓ Removed 5 orphaned resource(s)
```

---

### Example 4: Disaster Recovery

```bash
# Accidentally deleted all resources
aimgr repo list
# Shows: No resources found

# But ai.repo.yaml is intact
aimgr repo info
# Shows: 2 configured sources

# Recover everything
aimgr repo sync
```

**Output:**
```
Syncing from 2 configured source(s)...

[1/2] Syncing source: my-skills
  Mode: Local (symlink)
  ✓ Imported 8 resources

[2/2] Syncing source: community-catalog
  Mode: Remote (download + copy)
  ✓ Imported 12 resources

Summary:
  ✓ Synced 2 source(s)
  ✓ Imported 20 resources total
```

---

## Troubleshooting

### Source Not Found After Moving Directory

**Symptom:**
```bash
$ aimgr repo sync
Error syncing source 'my-skills': path does not exist: /old/path/my-skills
```

**Cause:** Local source was moved to a new location.

**Solution:**

```bash
# Option 1: Update ai.repo.yaml manually
vim ~/.local/share/ai-config/repo/ai.repo.yaml
# Change path: /old/path/my-skills to /new/path/my-skills

# Option 2: Remove and re-add
aimgr repo drop-source my-skills --keep-resources
aimgr repo add /new/path/my-skills --name=my-skills --force
```

---

### Remote Repository Unavailable

**Symptom:**
```bash
$ aimgr repo sync
Error syncing source 'company-resources': failed to clone: repository not found
```

**Cause:** Remote repository was deleted, made private, or URL changed.

**Solution:**

```bash
# If temporarily unavailable, skip for now
# Sync will continue with other sources

# If permanently unavailable, remove source
aimgr repo drop-source company-resources --keep-resources

# If URL changed, update ai.repo.yaml
vim ~/.local/share/ai-config/repo/ai.repo.yaml
# Change url: field to new URL
```

---

### Corrupted ai.repo.yaml

**Symptom:**
```bash
$ aimgr repo sync
Error: failed to load manifest: invalid YAML
```

**Cause:** Manual editing introduced syntax errors.

**Solution:**

```bash
# Validate YAML syntax
yamllint ~/.local/share/ai-config/repo/ai.repo.yaml

# Or restore from backup
cp ~/.local/share/ai-config/repo/ai.repo.yaml.backup \
   ~/.local/share/ai-config/repo/ai.repo.yaml

# Or recreate from scratch
aimgr repo drop --force  # Soft drop (keeps directory)
aimgr repo add gh:owner/repo --name=source1
aimgr repo add ~/path/to/local --name=source2
```

---

### Duplicate Source Names

**Symptom:**
```bash
$ aimgr repo add ~/new-resources
Error: source with name 'new-resources' already exists
```

**Cause:** Source name collision (auto-generated names based on directory/repo name).

**Solution:**

```bash
# Use custom name
aimgr repo add ~/new-resources --name=new-resources-v2

# Or remove old source first
aimgr repo drop-source new-resources
aimgr repo add ~/new-resources
```

---

### Recovering from Soft Drop

**Symptom:** Accidentally ran `aimgr repo drop --force` and lost all resources.

**Solution:**

```bash
# Check if ai.repo.yaml still exists
cat ~/.local/share/ai-config/repo/ai.repo.yaml

# If yes, rebuild from sources
aimgr repo sync

# If no, ai.repo.yaml was deleted (full delete)
# You'll need to re-add sources manually
aimgr repo add gh:owner/repo --name=source1
aimgr repo add ~/local/path --name=source2
```

**Prevention:** Use `--full-delete` flag only when you truly want to delete everything.

---

### Symlink Mode Not Working on Windows

**Symptom:** Symlinked resources not reflecting changes.

**Cause:** Windows symlink support requires administrator privileges or Developer Mode.

**Solution:**

```bash
# Option 1: Enable Developer Mode in Windows Settings
# Settings > Update & Security > For Developers > Developer Mode

# Option 2: Use copy mode instead
aimgr repo drop-source my-source --keep-resources
aimgr repo add ~/my-resources --name=my-source --copy

# Re-sync after each change
aimgr repo sync
```

---

## Best Practices

### 1. Use Descriptive Source Names

Good names make maintenance easier:

```bash
# Bad (auto-generated, unclear)
aimgr repo add gh:owner/repo
# Source name: "repo"

# Good (explicit, descriptive)
aimgr repo add gh:owner/repo --name=community-pdf-tools
```

---

### 2. Pin Remote Sources to Versions

Avoid breaking changes by pinning to stable versions:

```bash
# Bad (tracks latest, may break)
aimgr repo add gh:owner/repo

# Good (pinned to stable version)
aimgr repo add gh:owner/repo@v1.2.0

# Also good (tracks stable branch)
aimgr repo add gh:owner/repo@stable
```

---

### 3. Use Filters for Selective Importing

Keep your repository focused by filtering:

```bash
# Import only skills from a mixed repository
aimgr repo add gh:owner/all-resources --name=skills-only --filter "skill/*"

# Import only test-related resources
aimgr repo add ~/dev/experiments --name=tests --filter "*test*"
```

---

### 4. Separate Development and Production Sources

Maintain clear boundaries:

```yaml
# ai.repo.yaml example
sources:
  # Production (stable, versioned)
  - name: prod-resources
    url: https://github.com/company/resources
    ref: v2.1.0
    mode: copy
    
  # Development (local, live editing)
  - name: dev-resources
    path: /home/user/dev/resources
    mode: symlink
```

---

### 5. Run Sync Regularly

Keep resources fresh:

```bash
# Add to daily routine or CI/CD
aimgr repo sync

# Or set up a cron job
0 9 * * * /usr/local/bin/aimgr repo sync
```

---

### 6. Preview Before Destructive Operations

Always use `--dry-run` first:

```bash
# Before removing source
aimgr repo drop-source old-source --dry-run

# Before syncing
aimgr repo sync --dry-run

# Before full delete
# (There is no dry-run for drop, so be very careful!)
```

---

### 7. Back Up ai.repo.yaml

Protect your source configuration:

```bash
# Manual backup
cp ~/.local/share/ai-config/repo/ai.repo.yaml \
   ~/.local/share/ai-config/repo/ai.repo.yaml.backup

# Or commit to git
cd ~/.local/share/ai-config/repo
git init
git add ai.repo.yaml
git commit -m "Backup source configuration"
```

---

### 8. Document Custom Source Names

Add comments to your `ai.repo.yaml`:

```yaml
# ai.repo.yaml
version: 1
sources:
  # Company-approved resources for production use
  - name: company-prod
    url: https://github.com/company/resources
    ref: v2.0.0
    mode: copy
    added: 2026-02-14T10:00:00Z
    
  # Personal development experiments (not for production)
  - name: personal-dev
    path: /home/user/dev/experiments
    mode: symlink
    added: 2026-02-14T11:00:00Z
```

---

## Migrating from Global Config Sync

If you previously used the global config (`~/.config/aimgr/aimgr.yaml`) with `sync.sources`, here's how to migrate:

### Old Workflow (Deprecated)

```yaml
# ~/.config/aimgr/aimgr.yaml
sync:
  sources:
    - url: https://github.com/owner/repo
      filter: "skill/*"
    - url: ~/my-resources
```

```bash
aimgr repo sync  # Read from global config
```

### New Workflow (Current)

```bash
# Add sources (tracked in ai.repo.yaml)
aimgr repo add https://github.com/owner/repo --filter "skill/*"
aimgr repo add ~/my-resources

# Sync reads from ai.repo.yaml
aimgr repo sync
```

### Migration Steps

1. **List your old sources:**
   ```bash
   cat ~/.config/aimgr/aimgr.yaml | grep -A 10 "sync:"
   ```

2. **Add them using `repo add`:**
   ```bash
   aimgr repo add <source1>
   aimgr repo add <source2>
   ```

3. **Remove old config section:**
   ```bash
   vim ~/.config/aimgr/aimgr.yaml
   # Delete the sync: section
   ```

4. **Verify:**
   ```bash
   aimgr repo info
   aimgr repo list
   ```

**Benefits of new workflow:**
- Repository-specific source configuration
- Version control for source tracking
- No global configuration conflicts
- Portable repositories

---

## Related Documentation

- **[Getting Started](./getting-started.md)** - First steps with aimgr
- **[Configuration Guide](./configuration.md)** - Global and project configuration
- **[Pattern Matching](./pattern-matching.md)** - Filter pattern syntax
- **[GitHub Sources](./github-sources.md)** - GitHub-specific import details
- **[Workspace Caching](./workspace-caching.md)** - Git repository caching
- **[Resource Formats](./resource-formats.md)** - Resource structure requirements
- **[Developer Guide](./developer-guide.md)** - Creating and validating resources

---

## Summary

**Source management in aimgr:**

1. **Add sources** with `repo add` → tracked in `ai.repo.yaml`
2. **Remove sources** with `repo drop-source` → orphan cleanup
3. **Sync sources** with `repo sync` → get latest changes
4. **View sources** with `repo info` → health and status
5. **Self-contained** → each repository manages its own sources

The `ai.repo.yaml` manifest is your single source of truth for resource synchronization.
