# Configuration Guide

This guide covers all configuration options for `aimgr`, including repository path customization, installation targets, and sync sources.

## Config File Location

`aimgr` uses XDG Base Directory standards for configuration:

- **Default location**: `~/.config/aimgr/aimgr.yaml`
- **Legacy location**: `~/.ai-repo.yaml` (automatically migrated)
- **Custom location**: Use `--config` flag (not implemented yet)

The config file is automatically created when you first run commands that need configuration.

## Config File Format

The config file uses YAML format with three main sections:

```yaml
# Repository configuration
repo:
  path: ~/my-custom-repo

# Installation configuration
install:
  targets:
    - claude
    - opencode

# Sync configuration
sync:
  sources:
    - url: gh:anthropics/skills
    - url: gh:myorg/resources
      filter: "skill/*"
```

---

## Repository Path Configuration

By default, `aimgr` stores resources in `~/.local/share/ai-config/repo` (XDG data directory). You can customize this location using three methods.

### Precedence Rules

The repository path is determined using this precedence order (highest to lowest):

1. **`AIMGR_REPO_PATH` environment variable** (highest priority)
2. **`repo.path` in config file** (`~/.config/aimgr/aimgr.yaml`)
3. **XDG default** (`~/.local/share/ai-config/repo`)

Each level overrides all levels below it.

### Option 1: Config File (Recommended)

Set `repo.path` in your config file for a persistent custom location:

**Edit `~/.config/aimgr/aimgr.yaml`:**

```yaml
repo:
  path: ~/my-custom-repo

install:
  targets:
    - claude
```

**Benefits:**
- ✅ Persistent across all sessions
- ✅ Version controllable (can be checked into dotfiles)
- ✅ Supports path expansion (tilde, relative paths)
- ✅ Validated on load (errors shown early)

**Use this when:**
- You want a permanent custom repository location
- You're managing dotfiles in version control
- You want path validation and error checking

### Option 2: Environment Variable

Set `AIMGR_REPO_PATH` to override the config file:

**Temporary (current session only):**

```bash
export AIMGR_REPO_PATH=/path/to/custom/repo
aimgr list
```

**Permanent (add to shell profile):**

```bash
# For Bash
echo 'export AIMGR_REPO_PATH=~/my-repo' >> ~/.bashrc
source ~/.bashrc

# For Zsh
echo 'export AIMGR_REPO_PATH=~/my-repo' >> ~/.zshrc
source ~/.zshrc

# For Fish
echo 'set -x AIMGR_REPO_PATH ~/my-repo' >> ~/.config/fish/config.fish
```

**Benefits:**
- ✅ Overrides all other settings (highest priority)
- ✅ Easy to set temporarily for testing
- ✅ Works across different projects
- ✅ Useful for CI/CD environments

**Use this when:**
- You need to temporarily test with a different repository
- You want different repos for different shell sessions
- You're running in CI/CD with different environments
- You need to override config file setting without editing it

### Option 3: XDG Default

If no custom path is configured, `aimgr` uses the XDG data directory:

```
~/.local/share/ai-config/repo/
```

This follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

**Benefits:**
- ✅ Standard Linux/Unix location for application data
- ✅ Keeps home directory clean
- ✅ No configuration needed
- ✅ Works out of the box

**Use this when:**
- You're happy with the default location
- You want to follow XDG standards
- You don't need custom repository locations

### Path Expansion

All custom paths support expansion and normalization:

#### Tilde Expansion

```yaml
repo:
  path: ~/my-repo
# Expands to: /home/username/my-repo
```

```bash
export AIMGR_REPO_PATH=~/custom/repo
# Expands to: /home/username/custom/repo
```

#### Relative Paths

Relative paths are automatically converted to absolute paths:

```yaml
repo:
  path: ./repo
# Converts to: /home/username/project/repo (from current directory)
```

```yaml
repo:
  path: ../shared-repo
# Converts to: /home/username/shared-repo
```

**Note:** Relative paths in config file are resolved relative to the current working directory when `aimgr` is run.

#### Absolute Paths

Absolute paths are used as-is:

```yaml
repo:
  path: /opt/ai-resources
# Used directly: /opt/ai-resources
```

All paths are cleaned and normalized (e.g., `//double/slashes` becomes `/double/slashes`).

---

## Environment Variable Interpolation

aimgr supports Docker Compose-style environment variable interpolation in the config file. This allows you to reference environment variables with optional default values.

### Syntax

| Pattern | Description | Example |
|---------|-------------|---------|
| `${VAR}` | Simple substitution | `${HOME}` → `/home/user` |
| `${VAR:-default}` | Default if unset/empty | `${PORT:-5432}` → `5432` if PORT not set |

### Variable Names

Variable names must match the pattern: `[A-Za-z_][A-Za-z0-9_]*`
- Start with letter or underscore
- Contain only letters, numbers, and underscores
- Case-sensitive (standard shell convention)

### Examples

**Basic usage:**
```yaml
repo:
  path: ${AIMGR_REPO_PATH}  # Use env var or empty if not set
```

**With default values:**
```yaml
repo:
  path: ${AIMGR_REPO_PATH:-~/.local/share/ai-config/repo}

sync:
  sources:
    - url: ${SYNC_REPO:-https://github.com/hk9890/ai-tools}
      filter: ${RESOURCE_FILTER:-skill/*}
```

**Multiple variables:**
```yaml
sync:
  sources:
    - url: ${PROTOCOL:-https}://${HOST}/owner/repo
```

**Environment-specific configs:**
```yaml
# Development
repo:
  path: ${DEV_REPO_PATH:-~/dev/ai-resources}

# Production  
repo:
  path: ${PROD_REPO_PATH:-/var/lib/ai-config/repo}
```

### How It Works

1. Environment variables are expanded **before** YAML parsing
2. Variable expansion happens in both `Load()` and `LoadGlobal()`
3. Works in **any config field** (repo.path, sync.sources, etc.)
4. If variable is unset/empty and no default provided, expands to empty string
5. Existing validation applies **after** expansion

### Use Cases

**CI/CD environments:**
Different paths per environment without maintaining multiple config files:

```yaml
repo:
  path: ${CI_REPO_PATH:-~/.local/share/ai-config/repo}

sync:
  sources:
    - url: ${CI_RESOURCE_REPO:-https://github.com/myorg/resources}
```

**Team configurations:**
Shared config with user-specific overrides:

```yaml
repo:
  path: ${USER_REPO_PATH:-~/team-ai-resources}

sync:
  sources:
    - url: ${TEAM_RESOURCES:-https://github.com/team/resources}
      filter: ${USER_FILTER:-*}
```

**Testing:**
Override paths without modifying config file:

```bash
# Run tests with temporary repository
export AIMGR_REPO_PATH=/tmp/test-repo
aimgr repo import ~/test-resources/
aimgr install skill/test-skill

# Original config unchanged
unset AIMGR_REPO_PATH
aimgr list  # Uses default config path
```

**Secret management:**
Reference secrets from environment (for private repositories):

```yaml
sync:
  sources:
    - url: https://${GH_TOKEN}@github.com/private/repo
```

---

## Installation Targets

Configure which AI tools to install resources to by default:

```yaml
install:
  targets:
    - claude
    - opencode
    - copilot
```

**Valid tools:**
- `claude` - Claude Code (`.claude/` directories)
- `opencode` - OpenCode (`.opencode/` directories)
- `copilot` - GitHub Copilot (`.github/skills/` directories)

**Behavior:**
- Used when installing to fresh projects (no existing tool directories)
- Overridden by existing tool directories in the project
- Can be overridden per-command with `--target` flag

See the [README](../../README.md#multi-tool-support) for more details on multi-tool installation.

### Setting Targets via CLI

You can also manage installation targets via the CLI:

```bash
# Set single target
aimgr config set install.targets claude

# Set multiple targets
aimgr config set install.targets claude,opencode

# Get current setting
aimgr config get install.targets
```

---

## Sync Configuration

Configure sources to automatically sync resources from:

```yaml
sync:
  sources:
    - url: https://github.com/anthropics/skills
    - url: gh:myorg/ai-resources@v1.0.0
      filter: "skill/*"
    - url: ~/local/resources
      filter: "*test*"
```

### Source Fields

**`url`** (required): Source location
- GitHub: `gh:owner/repo` or `owner/repo`
- Git URLs: `https://github.com/owner/repo.git`
- Local paths: `~/path/to/resources` or `/absolute/path`
- Version tags: `gh:owner/repo@v1.0.0`

**`filter`** (optional): Glob pattern to filter resources
- `"skill/*"` - Only skills
- `"command/*"` - Only commands
- `"agent/*"` - Only agents
- `"*test*"` - Resources with "test" in name
- `"skill/pdf*"` - Skills starting with "pdf"

### Running Sync

Once configured, sync all sources with one command:

```bash
# Sync all configured sources (overwrites existing)
aimgr repo sync

# Sync without overwriting existing resources
aimgr repo sync --skip-existing

# Preview what would be synced
aimgr repo sync --dry-run
```

See the [README](../../README.md#aimgr-repo-sync) for more details on sync functionality.

---

## Complete Example

Here's a complete example config file with all options:

```yaml
# ~/.config/aimgr/aimgr.yaml

# Repository location (optional)
# If not specified, uses: ~/.local/share/ai-config/repo
repo:
  path: ~/ai-resources

# Default installation targets (required)
install:
  targets:
    - claude
    - opencode

# Sync sources (optional)
sync:
  sources:
    # Public repositories
    - url: gh:anthropics/skills

    # Organization resources with filter
    - url: gh:myorg/company-resources
      filter: "skill/*"

    # Specific version
    - url: gh:myorg/stable-tools@v2.1.0

    # Local directory with filter
    - url: ~/dev/custom-tools
      filter: "*test*"

    # Direct Git URL
    - url: https://github.com/community/awesome-tools.git
```

---

## Configuration Examples

### Example 1: Developer with Custom Repo

```yaml
# Developer using custom repo location
repo:
  path: ~/dev/ai-repo

install:
  targets:
    - claude
    - opencode
```

### Example 2: Team with Shared Network Location

```yaml
# Team sharing resources on network drive
repo:
  path: /mnt/shared/ai-resources

install:
  targets:
    - claude
```

### Example 3: CI/CD Environment

Use environment variable for dynamic paths:

```bash
# .gitlab-ci.yml or .github/workflows/main.yml
export AIMGR_REPO_PATH=/tmp/ci-ai-repo
aimgr repo import gh:myorg/resources
aimgr install skill/linter
```

### Example 4: Multi-Source Sync

```yaml
# Automatically sync from multiple sources
repo:
  path: ~/ai-resources

install:
  targets:
    - claude

sync:
  sources:
    # Base skills from Anthropic
    - url: gh:anthropics/skills

    # Company internal tools
    - url: gh:mycompany/internal-tools
      filter: "skill/*"

    # Personal custom tools
    - url: ~/personal/custom-tools
```

### Example 5: Testing Environment

```bash
# Test with temporary repository
export AIMGR_REPO_PATH=/tmp/test-repo
aimgr repo import ./test-resources/
aimgr install skill/test-skill

# Original repository unchanged
unset AIMGR_REPO_PATH
aimgr list  # Uses default or config file path
```

---

## Troubleshooting

### Issue: Config file not found

**Error:**
```
Error: no config found

Please create a config file at: ~/.config/aimgr/aimgr.yaml
```

**Solution:**

Create the config file with minimum required settings:

```bash
mkdir -p ~/.config/aimgr
cat > ~/.config/aimgr/aimgr.yaml << 'EOF'
install:
  targets:
    - claude
EOF
```

### Issue: Custom repo path not working

**Symptom:** Resources still going to `~/.local/share/ai-config/repo`

**Cause:** Environment variable `AIMGR_REPO_PATH` is overriding config file

**Solution:**

Check for environment variable:

```bash
echo $AIMGR_REPO_PATH

# If set and you want to use config file instead:
unset AIMGR_REPO_PATH
```

Or verify config file path:

```bash
cat ~/.config/aimgr/aimgr.yaml
```

### Issue: Tilde (~) not expanding

**Symptom:** Path like `~/repo` is not expanding to home directory

**Cause:** Path might be quoted incorrectly or not processed by config loader

**Solution:**

Ensure path is in config file (not just environment variable):

```yaml
# ✅ Correct - will expand ~
repo:
  path: ~/my-repo

# ❌ Incorrect - quotes prevent expansion in some contexts
repo:
  path: "~/my-repo"
```

### Issue: Relative path unexpected behavior

**Symptom:** Relative path resolves to unexpected location

**Cause:** Relative paths resolve from current working directory when `aimgr` runs

**Solution:**

Use absolute paths or tilde expansion instead:

```yaml
# ✅ Recommended - use tilde
repo:
  path: ~/ai-repo

# ✅ Also good - absolute path
repo:
  path: /home/username/ai-repo

# ⚠️  Avoid - resolves from current directory
repo:
  path: ./ai-repo
```

### Issue: Permission denied

**Error:**
```
Error: failed to create repo directory: permission denied
```

**Solution:**

Ensure you have write permissions to the configured path:

```bash
# Check permissions
ls -ld ~/my-repo

# Fix permissions if needed
chmod 755 ~/my-repo

# Or use a different path you own
```

### Issue: Precedence confusion

**Symptom:** Not sure which repo path is being used

**Solution:**

Check precedence in order:

```bash
# 1. Check environment variable (highest priority)
echo $AIMGR_REPO_PATH

# 2. Check config file
cat ~/.config/aimgr/aimgr.yaml | grep -A1 "^repo:"

# 3. Default XDG location (if nothing else set)
echo ~/.local/share/ai-config/repo

# Verify actual repo location
aimgr repo list  # Lists resources from active repo
```

---

## Migration Notes

### Migrating from XDG Default to Custom Path

If you want to move your repository from the default location:

```bash
# Copy existing resources to new location
cp -r ~/.local/share/ai-config/repo ~/my-custom-repo

# Update config file
cat >> ~/.config/aimgr/aimgr.yaml << 'EOF'
repo:
  path: ~/my-custom-repo
EOF

# Verify it works
aimgr repo list

# Optional: Remove old location
rm -rf ~/.local/share/ai-config/repo
```

### Migrating from Legacy Config

If you have `~/.ai-repo.yaml`, it's automatically migrated to `~/.config/aimgr/aimgr.yaml` on first use.

The old file is left intact for safety. You can delete it once you verify the migration worked:

```bash
# Verify migration
cat ~/.config/aimgr/aimgr.yaml

# Delete old config if migration successful
rm ~/.ai-repo.yaml
```

---

## See Also

- [README.md](../../README.md) - Quick start and overview
- [workspace-caching.md](./workspace-caching.md) - Git repository caching
- [output-formats.md](./output-formats.md) - CLI output formats
- [pattern-matching.md](./pattern-matching.md) - Pattern syntax for filtering
