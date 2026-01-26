# Release Notes: v1.4.0

**Release Date:** January 24, 2026

## Overview

Version 1.4.0 is a **critical maintenance release** that fixes a serious bug affecting all Git-sourced resources and introduces powerful new repository maintenance commands. This release ensures reliable resource updates and provides tools for repository health management.

## 🔴 Critical Bug Fix

### Git Source Metadata Storage (P0)

**Issue:** When adding resources from Git sources (GitHub, GitLab, Git URLs), metadata incorrectly stored temporary clone paths (`file:///tmp/ai-repo-clone-*/`) instead of the original Git source URL. This made `aimgr repo update` fail for all Git-sourced resources after the temporary directory was cleaned up.

**Impact:** All resources added from Git sources since v1.0.0 could not be updated.

**Root Cause:** The `AddBulk` import chain always converted source paths to `file://` URLs, losing the original Git source information.

**Fix:**
- Modified `BulkImportOptions` to include `SourceURL` and `SourceType` fields
- Updated `importResource()` to use provided source info or fall back to `file://`
- Fixed `addBulkFromGitHub()` to pass original Git URLs and source types
- Metadata now correctly stores: `https://github.com/owner/repo` or `gh:owner/repo`

**Upgrade Impact:** 
- ✅ **Future adds work correctly** - New resources from Git sources will have proper metadata
- ⚠️ **Existing broken metadata** - Use new `repo prune` command to clean up orphaned entries
- ℹ️ **Workaround for existing resources** - Remove and re-add from original Git source

## 🆕 New Commands

### 1. `aimgr repo prune`

Remove orphaned metadata entries where source paths no longer exist.

**Usage:**
```bash
aimgr repo prune              # Scan and remove with confirmation
aimgr repo prune --dry-run    # Preview without deleting
aimgr repo prune --force      # Skip confirmation prompt
```

**Features:**
- Detects metadata pointing to non-existent local/file sources
- Interactive confirmation with detailed list of orphaned entries
- Skips Git sources (transient by nature)
- Summary with counts of removed/failed entries

**When to use:**
- After encountering update errors for missing sources
- During repository cleanup and maintenance
- To remove test artifacts and stale metadata

### 2. `aimgr repo verify`

Check repository health and metadata consistency (similar to `git fsck`).

**Usage:**
```bash
aimgr repo verify             # Check for issues
aimgr repo verify --fix       # Auto-resolve fixable issues
aimgr repo verify --json      # Machine-readable output
```

**Detects:**
- ⚠️ **Resources without metadata** (warning) - Resources in repo with no metadata tracking
- ✗ **Orphaned metadata** (error) - Metadata files for non-existent resources
- ⚠️ **Missing source paths** (warning) - Metadata references to deleted/moved sources
- ✗ **Type mismatches** (error) - Resource type differs from metadata type

**Exit Codes:**
- `0` - No errors found (warnings are acceptable)
- `1` - Errors found (requires attention)

**When to use:**
- Regular repository health checks
- Before important operations
- CI/CD integration for repository validation
- Troubleshooting resource issues

### 3. Improved `aimgr repo update`

Enhanced UX for handling missing source paths gracefully.

**Previous behavior:**
```
✗ skill 'pdf-processor': source path no longer exists: /tmp/...
✗ skill 'doc-parser': source path no longer exists: /tmp/...
... (aborts on errors, partial updates)
```

**New behavior:**
```
⊘ Skipped 7 resources with missing sources
✓ Updated 16 resources successfully
✗ Failed 2 resources with errors

Hint: Run 'aimgr repo prune' to remove orphaned metadata
```

**Improvements:**
- Continues updating all resources even when some sources are missing
- Distinguishes between skipped (⊘) and failed (✗) resources
- Provides actionable hints to resolve issues
- Separates warnings from errors in summary

## 🧪 Testing & Quality

### Test Isolation

All tests now use isolated temporary repositories, preventing pollution of user's actual repository.

**Changes:**
- Added `AIMGR_REPO_PATH` environment variable support for test isolation
- All tests use `t.TempDir()` with `NewManagerWithPath()`
- Tests no longer write to `~/.local/share/ai-config/`
- Can safely run tests in parallel
- Added `TEST_ISOLATION.md` documentation

### Comprehensive Integration Tests

Added 25+ new integration tests covering all new functionality:

**`repo prune` tests (10 tests):**
- Basic dry-run and force modes
- Multiple orphaned entries across resource types
- Git source handling (should not be pruned)
- Empty repositories and mixed source types

**`repo verify` tests (15 tests):**
- Resources without metadata detection
- Orphaned metadata detection
- Missing source paths detection
- Type mismatch detection
- `--fix` flag functionality
- `--json` output format
- Exit code validation
- Multi-issue scenarios

**Test Results:**
- ✅ 200+ total tests passing
- ✅ 100% of new features covered
- ✅ Zero regressions in existing functionality

## 📊 Statistics

- **Commits:** 8 feature commits + integration
- **Files Changed:** 15+ files across core packages
- **New Tests:** 25+ integration tests
- **Bug Severity:** P0 (Critical - affects core functionality)
- **Lines Added:** ~2000 (including tests and docs)

## 🚀 Upgrade Guide

### For All Users

1. **Update aimgr:**
   ```bash
   # Download and install v1.4.0
   # See installation instructions in README
   ```

2. **Verify repository health:**
   ```bash
   aimgr repo verify
   ```

3. **Clean up orphaned metadata (if any):**
   ```bash
   aimgr repo prune --dry-run  # Preview
   aimgr repo prune            # Execute
   ```

### For Users with Git-Sourced Resources

If you previously added resources from GitHub/Git URLs and encounter update errors:

**Option 1: Clean and re-add (recommended)**
```bash
# List resources that need re-adding
aimgr repo list | grep <source>

# Remove the resource
aimgr repo remove <resource-type> <resource-name>

# Re-add from original Git source
aimgr repo add gh:owner/repo
```

**Option 2: Use prune to clean up**
```bash
# Remove orphaned metadata
aimgr repo prune

# Future adds will work correctly
```

### Migration Checklist

- [ ] Install aimgr v1.4.0
- [ ] Run `aimgr repo verify` to identify issues
- [ ] Run `aimgr repo prune` to clean orphaned metadata
- [ ] Test `aimgr repo update` - should complete without errors
- [ ] Re-add any critical resources that were broken

## 💡 Best Practices

### Regular Maintenance

```bash
# Weekly health check
aimgr repo verify

# Clean up orphaned metadata
aimgr repo prune --dry-run
aimgr repo prune

# Keep resources updated
aimgr repo update
```

### CI/CD Integration

```bash
# Validate repository in CI
aimgr repo verify --json
if [ $? -ne 0 ]; then
  echo "Repository validation failed"
  exit 1
fi
```

### Safe Experimentation

```bash
# Always preview destructive operations
aimgr repo prune --dry-run

# Use verify before making changes
aimgr repo verify --fix
```

## 🐛 Known Issues

None identified in this release.

## 🔜 What's Next (v1.5.0)

Planned features for the next release:
- Resource dependency tracking
- Enhanced source type support (GitLab, Bitbucket)
- Repository synchronization between machines
- Performance optimizations for large repositories
- Advanced filtering and search capabilities

## 📝 Breaking Changes

**None.** This release is fully backward compatible.

## 🙏 Contributors

Thank you to all contributors who made this release possible!

Special thanks to community members who reported the Git source metadata issue.

---

**Full Changelog:** https://github.com/hk9890/ai-config-manager/compare/v1.3.0...v1.4.0

**Download:** See [Releases](https://github.com/hk9890/ai-config-manager/releases/tag/v1.4.0)
