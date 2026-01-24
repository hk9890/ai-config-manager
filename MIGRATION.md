# Migration Guide: ai-repo → aimgr

This guide helps you migrate from the old `ai-repo` command to the new `aimgr` (AI Manager) command structure.

## Overview

**aimgr v0.4.0** introduces a major rebranding and command structure reorganization:
- Binary renamed from `ai-repo` to `aimgr`
- Commands reorganized under `repo` subcommand group
- New install/uninstall syntax with type prefixes (e.g., `skill/name`)
- New commands: `repo show`, `repo update`
- Metadata tracking for resource sources

## Quick Reference

| Old Command (ai-repo) | New Command (aimgr) |
|----------------------|---------------------|
| `ai-repo add skill foo` | `aimgr repo add skill foo` |
| `ai-repo add command bar` | `aimgr repo add command bar` |
| `ai-repo add agent reviewer` | `aimgr repo add agent reviewer` |
| `ai-repo list` | `aimgr repo list` |
| `ai-repo list skill` | `aimgr repo list skill` |
| `ai-repo remove skill old` | `aimgr repo remove skill old` |
| `ai-repo install skill foo` | `aimgr install skill/foo` |
| `ai-repo install command bar` | `aimgr install command/bar` |
| `ai-repo install agent reviewer` | `aimgr install agent/reviewer` |
| N/A | `aimgr repo show skill foo` |
| N/A | `aimgr repo update` |

## Breaking Changes

### 1. Binary Name Change

The executable has been renamed from `ai-repo` to `aimgr`.

**Before:**
```bash
ai-repo --version
```

**After:**
```bash
aimgr --version
```

**Migration:** Replace or update your binary. If you installed to a PATH location, remove the old binary:
```bash
# Remove old binary
sudo rm /usr/local/bin/ai-repo

# Install new binary
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

### 2. Repo Command Group

Repository management commands now require the `repo` subcommand prefix.

**Before:**
```bash
ai-repo add skill pdf-processing
ai-repo list
ai-repo remove command old-test
```

**After:**
```bash
aimgr repo add skill pdf-processing
aimgr repo list
aimgr repo remove command old-test
```

**Migration:** Add `repo` between `aimgr` and the operation name (`add`, `list`, `remove`).

### 3. Install/Uninstall Syntax Change

Installing and uninstalling now uses type prefixes in the format `type/name`.

**Before:**
```bash
ai-repo install skill pdf-processing
ai-repo install command test
ai-repo install agent code-reviewer
```

**After:**
```bash
aimgr install skill/pdf-processing
aimgr install command/test
aimgr install agent/code-reviewer
```

**Migration:** Change `install TYPE NAME` to `install TYPE/NAME`.

**Multiple resources:**
```bash
# Before (multiple commands)
ai-repo install skill foo
ai-repo install skill bar
ai-repo install command test

# After (single command)
aimgr install skill/foo skill/bar command/test
```

### 4. Shell Completion Command Changes

Shell completion generation follows the new command structure.

**Before:**
```bash
ai-repo completion bash > /etc/bash_completion.d/ai-repo
```

**After:**
```bash
aimgr completion bash > /etc/bash_completion.d/aimgr
```

**Migration:** Regenerate your shell completion scripts with the new binary name.

## New Features

### 1. `repo show` Command

Display detailed information about a resource, including metadata.

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
- Source information (GitHub URL, local path, etc.)
- Version, author, license
- Installation locations
- Full frontmatter metadata

### 2. `repo update` Command

Update resources from their original sources (GitHub, local paths).

```bash
# Update all resources
aimgr repo update

# Update specific skill
aimgr repo update skill pdf-processing

# Update specific command
aimgr repo update command test

# Preview updates without applying
aimgr repo update --dry-run

# Force update, overwriting local changes
aimgr repo update --force
```

**How it works:**
- Tracks source information in metadata files (stored in `.metadata/` directory)
- Fetches latest version from original source
- Updates repository copy
- Preserves symlinks to projects

### 3. Metadata Tracking

Resources now store metadata about their sources for updates in a centralized `.metadata/` directory.

**Metadata file location:**
```
~/.local/share/ai-config/repo/.metadata/skills/my-skill-metadata.json
~/.local/share/ai-config/repo/.metadata/commands/my-command-metadata.json
~/.local/share/ai-config/repo/.metadata/agents/my-agent-metadata.json
```

**Metadata format (JSON):**
```json
{
  "name": "my-skill",
  "type": "skill",
  "source_type": "github",
  "source_url": "https://github.com/owner/repo",
  "first_installed": "2026-01-22T10:30:00Z",
  "last_updated": "2026-01-22T10:30:00Z"
}
```

**Source types:**
- `github`: GitHub repositories
- `local`: Local filesystem paths
- `file`: Direct file sources

**Migration from old metadata format (Historical - Command Removed in v1.4.0):**

**Note:** The `aimgr repo migrate-metadata` command has been removed as all repositories have been successfully migrated. If you have an older repository from before v1.3.0, you will need to use aimgr v1.3.x to perform the migration.

If you have metadata files in the old location (`.aimgr-meta.yaml` files), use the migration command from v1.3.x:
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
1. Found all old metadata files in your repository
2. Converted YAML format to JSON
3. Moved files to the new `.metadata/` directory structure
4. Cleaned up old metadata files after successful migration

### 4. Batch Installation

Install multiple resources in a single command.

```bash
# Before (multiple commands)
ai-repo install skill foo
ai-repo install skill bar
ai-repo install command test
ai-repo install agent reviewer

# After (single command)
aimgr install skill/foo skill/bar command/test agent/reviewer
```

### 5. Uninstall Command

New dedicated `uninstall` command (replaces manual removal).

```bash
# Uninstall a single resource
aimgr uninstall skill/old-skill

# Uninstall multiple resources
aimgr uninstall skill/foo skill/bar command/old-test

# Uninstall from specific project
aimgr uninstall skill/foo --project-path ~/my-project
```

**Features:**
- Safely removes symlinks only
- Warns about non-symlinks
- Multi-tool support (removes from all detected tools)
- Summary of what was removed

## Migration Workflow

### Step 1: Install New Binary

Download and install the new `aimgr` binary:

```bash
# Download for your platform
curl -L https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr_VERSION_linux_amd64.tar.gz | tar xz

# Install
sudo mv aimgr /usr/local/bin/

# Verify
aimgr --version
```

### Step 2: Update Shell Completion (Optional)

If you use shell completion, regenerate it:

**Bash:**
```bash
# Remove old completion
sudo rm /etc/bash_completion.d/ai-repo

# Add new completion
aimgr completion bash | sudo tee /etc/bash_completion.d/aimgr

# Reload shell
source ~/.bashrc
```

**Zsh:**
```bash
# Remove old completion
rm "${fpath[1]}/_ai-repo"

# Add new completion
aimgr completion zsh > "${fpath[1]}/_aimgr"

# Restart shell
exec zsh
```

### Step 3: Update Scripts and Aliases

If you have scripts or shell aliases using `ai-repo`, update them:

**Scripts:**
```bash
# Find all shell scripts using ai-repo
grep -r "ai-repo" ~/scripts/

# Update them manually or with sed
sed -i 's/ai-repo /aimgr repo /g' ~/scripts/my-script.sh
sed -i 's/install skill \([a-z-]*\)/install skill\/\1/g' ~/scripts/my-script.sh
```

**Aliases:**
```bash
# Check your shell rc file
grep "ai-repo" ~/.bashrc ~/.zshrc

# Update aliases
# Before: alias air='ai-repo'
# After:  alias air='aimgr repo'
```

### Step 4: Update Project Documentation

If your project docs reference `ai-repo`, update them:

```bash
# Find markdown files with ai-repo references
grep -r "ai-repo" *.md docs/

# Update manually
```

### Step 5: Remove Old Binary (Optional)

After verifying the new binary works:

```bash
# Remove old binary
sudo rm /usr/local/bin/ai-repo

# Or wherever you installed it
which ai-repo  # Check if it still exists
```

## Compatibility Notes

### Repository Structure Unchanged

Your existing repository at `~/.local/share/ai-config/repo/` is **fully compatible** with the new version. No migration needed.

```
~/.local/share/ai-config/repo/
├── commands/
├── skills/
└── agents/
```

### Symlinks Still Work

Existing symlinks in your projects continue to work without changes:

```
your-project/.claude/skills/foo -> ~/.local/share/ai-config/repo/skills/foo/
```

### Config File Auto-Migration

Your configuration at `~/.config/aimgr/aimgr.yaml` (or old `~/.aimgr.yaml`) is automatically migrated on first run.

## Common Migration Issues

### Issue: Command Not Found After Installation

**Symptom:**
```bash
$ aimgr --version
bash: aimgr: command not found
```

**Solution:**
1. Verify binary location: `which aimgr`
2. Check PATH includes install location: `echo $PATH | grep /usr/local/bin`
3. Reinstall to correct location or update PATH

### Issue: Old `ai-repo` Still in PATH

**Symptom:**
```bash
$ which ai-repo
/usr/local/bin/ai-repo
```

**Solution:**
Remove the old binary:
```bash
sudo rm /usr/local/bin/ai-repo
```

### Issue: Shell Completion Not Working

**Symptom:**
Tab completion doesn't work or completes old commands.

**Solution:**
1. Remove old completion: `sudo rm /etc/bash_completion.d/ai-repo`
2. Install new completion: `aimgr completion bash | sudo tee /etc/bash_completion.d/aimgr`
3. Reload shell: `source ~/.bashrc` or `exec bash`

### Issue: Scripts Failing with New Syntax

**Symptom:**
```bash
Error: unknown command "add" for "aimgr"
```

**Solution:**
Update script to use `repo` subcommand:
```bash
# Change this:
aimgr add skill foo

# To this:
aimgr repo add skill foo
```

### Issue: Install Command Failing

**Symptom:**
```bash
$ aimgr install skill pdf-processing
Error: accepts 1 arg(s), received 2
```

**Solution:**
Use type prefix format:
```bash
# Change this:
aimgr install skill pdf-processing

# To this:
aimgr install skill/pdf-processing
```

## Gradual Migration Strategy

If you need to support both old and new commands during transition:

### Option 1: Wrapper Script

Create a wrapper that works with both:

```bash
#!/bin/bash
# ~/bin/ai-install
# Wrapper for installing resources

if command -v aimgr &> /dev/null; then
    # New version
    aimgr install "$@"
elif command -v ai-repo &> /dev/null; then
    # Old version
    ai-repo install "$@"
else
    echo "Error: Neither aimgr nor ai-repo found"
    exit 1
fi
```

### Option 2: Shell Function

Add to your `.bashrc` or `.zshrc`:

```bash
# Support both ai-repo and aimgr
air() {
    if command -v aimgr &> /dev/null; then
        aimgr repo "$@"
    elif command -v ai-repo &> /dev/null; then
        ai-repo "$@"
    else
        echo "Error: Neither aimgr nor ai-repo found"
        return 1
    fi
}
```

Usage:
```bash
air add skill foo       # Works with both versions
air list                # Works with both versions
```

## Getting Help

If you encounter migration issues:

1. **Documentation:** Read the [README.md](README.md) for full command reference
2. **Command Help:** Run `aimgr --help` or `aimgr repo --help`
3. **Issues:** Report problems at https://github.com/hk9890/ai-config-manager/issues
4. **Discussions:** Ask questions at https://github.com/hk9890/ai-config-manager/discussions

## Summary

The migration from `ai-repo` to `aimgr` is straightforward:

1. ✅ Replace binary: `ai-repo` → `aimgr`
2. ✅ Add `repo` for repository commands: `aimgr repo add/list/remove`
3. ✅ Use type prefixes for install: `aimgr install skill/name`
4. ✅ Update shell completion and scripts
5. ✅ Enjoy new features: `repo show`, `repo update`, metadata tracking

Your existing repository and symlinks continue to work without changes!
