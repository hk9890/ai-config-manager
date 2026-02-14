# aimgr v2.0.0 - Repository Source Management Redesign

## 🎉 Major Release: ai.repo.yaml Source Tracking

This is a **major breaking release** that fundamentally redesigns how aimgr manages resource sources. The new system uses a self-contained `ai.repo.yaml` manifest in each repository, replacing the global config-based approach.

---

## 🚀 What's New

### New Self-Contained Source Management

**`ai.repo.yaml` Manifest:**
- Each repository now has its own source manifest
- Auto-created by `repo init`
- Tracks all sources with metadata (timestamps, mode, health)
- Git-tracked by default for version control

**Example `ai.repo.yaml`:**
```yaml
version: 1
sources:
  - name: my-local-resources
    path: /home/user/resources
    mode: symlink
    added: 2026-02-14T10:30:00Z
    last_synced: 2026-02-14T15:45:00Z
  - name: github-catalog
    url: https://github.com/owner/repo
    ref: main
    mode: copy
    added: 2026-02-14T11:00:00Z
```

### New Commands

| Command | Description |
|---------|-------------|
| `repo add` | Add source and track in ai.repo.yaml (replaces `repo import`) |
| `repo drop-source` | Remove source from manifest with orphan cleanup |
| `repo sync` | Sync all sources from ai.repo.yaml (no longer uses global config) |
| `repo info` | View sources with health status and sync times |
| `repo drop` | Soft drop (default) or full delete with `--full-delete` |

### Enhanced Features

**Source Tracking:**
- Auto-generated source names (customizable with `--name`)
- Timestamp tracking (added, last_synced)
- Health status indicators (reachable/unreachable)
- Mode tracking (symlink/copy)

**Orphan Cleanup:**
- `repo drop-source` automatically removes orphaned resources
- `repo sync` removes resources from deleted sources
- Prevents stale resources from accumulating

**Better UX:**
- Clear, semantic command names
- Symmetrical add/remove operations
- Self-documenting repositories
- No global config pollution

---

## ⚠️ Breaking Changes

### Command Changes

**REMOVED:**
- ❌ `repo import` - Use `repo add` instead
- ❌ Global config `sync.sources` - Now tracked per-repo in `ai.repo.yaml`

**RENAMED:**
- `repo import <source>` → `repo add <source>`
- All functionality preserved, just renamed

**CHANGED:**
- `repo sync` now reads from `ai.repo.yaml` instead of global config
- `repo drop` now does soft drop by default (use `--full-delete` for old behavior)

### Configuration Changes

**Removed from `~/.config/aimgr/aimgr.yaml`:**
```yaml
# ❌ NO LONGER SUPPORTED
sync:
  sources:
    - url: https://github.com/owner/repo
```

**New per-repository approach:**
```bash
# ✓ Now tracked in repository
aimgr repo add https://github.com/owner/repo
# Creates/updates ai.repo.yaml in the repository
```

---

## 📋 Migration Guide

### Upgrading from v1.x

**Step 1: Identify your sources**

Check your global config for sync sources:
```bash
cat ~/.config/aimgr/aimgr.yaml | grep -A 10 "sync:"
```

**Step 2: Re-add sources to repository**

For each source in your old config:
```bash
# Old way (no longer works)
# Sources were in config file

# New way
aimgr repo add <source-path-or-url> --name=<descriptive-name>
```

**Step 3: Verify sources are tracked**

```bash
aimgr repo info
# Shows all sources from ai.repo.yaml
```

**Step 4: Clean up global config**

The `sync:` section in your config is now ignored. You can safely remove it:
```bash
# Edit ~/.config/aimgr/aimgr.yaml
# Remove the entire "sync:" section
```

### Automatic Migration

When you run any repo command after upgrading:
- `ai.repo.yaml` is automatically created if it doesn't exist
- You'll see: `Created ai.repo.yaml` message
- This is normal and expected

---

## 🔧 What Changed Under the Hood

### New Package: `pkg/repomanifest`
- Complete manifest management (Load/Save/Validate)
- Source manipulation (Add/Remove/Get)
- Auto-name generation
- Timestamp handling

### Metadata Enhancement
- Resources now track their source name
- Enables orphan detection and cleanup
- Backward compatible with old metadata

### Removed Code
- 700+ lines of sync-related config code removed
- Cleaner, more maintainable codebase
- Simpler mental model

---

## 📚 Documentation Updates

All documentation has been updated:
- [Getting Started](docs/user-guide/getting-started.md) - Updated workflow
- [Source Management](docs/user-guide/sync-sources.md) - Complete rewrite (1,067 lines)
- [GitHub Sources](docs/user-guide/github-sources.md) - New `repo add` examples
- [Configuration](docs/user-guide/configuration.md) - Removed sync.sources

---

## ✅ Testing

- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Manually tested all commands
- ✅ Production ready

---

## 📊 Statistics

- **19 tasks** completed
- **31 files** changed
- **+3,939 / -2,186** lines
- **100%** test coverage maintained

---

## 🙏 Thanks

This major redesign was driven by user feedback requesting:
- Simpler source management
- Self-contained repositories
- Better tracking and visibility
- Cleaner command semantics

Thank you to everyone who provided feedback!

---

## 🔗 Links

- [Full Changelog](https://github.com/hk9890/ai-config-manager/compare/v1.23.1...v2.0.0)
- [Documentation](docs/user-guide/README.md)
- [Epic ai-config-manager-25z](https://github.com/hk9890/ai-config-manager/issues?q=is%3Aissue+label%3Aai-config-manager-25z)
