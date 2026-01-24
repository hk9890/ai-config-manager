# Release Notes: v1.3.0

**Release Date:** January 24, 2026

## Overview

Version 1.3.0 introduces significant improvements to metadata management, adds powerful pattern matching capabilities for resource operations, and includes important bug fixes for agent detection and bulk import operations.

## Major Features

### 🗂️ Metadata Reorganization

The repository metadata structure has been reorganized into a cleaner `.metadata/` directory hierarchy for better organization and maintainability.

**New structure:**
```
~/.local/share/ai-config/repo/.metadata/
├── by-name/
│   ├── commands/
│   ├── skills/
│   └── agents/
└── by-source/
    └── <source-hash>/
```

**Migration Command:**
```bash
aimgr repo migrate-metadata [--dry-run]
```

The migration utility automatically:
- Detects old metadata format (`.metadata.json` files)
- Converts to new directory structure
- Preserves all metadata information
- Supports dry-run mode for safe testing
- Handles edge cases and validates integrity

### 🎯 Pattern Matching Support

Install and uninstall commands now support powerful glob pattern matching for batch operations.

**Pattern syntax:**
```bash
# Type-specific patterns
aimgr install skill/pdf*           # All skills starting with "pdf"
aimgr uninstall command/*test*     # All commands containing "test"

# Cross-type patterns
aimgr install *review*             # All resources containing "review"
aimgr uninstall agent/code-*       # All agents starting with "code-"
```

**Supported operators:**
- `*` - Match any characters
- `?` - Match single character
- `[abc]` - Character class
- `{a,b}` - Alternatives

### 🔄 Unified `repo add` Command

The `repo add` command has been refactored to provide a cleaner, more intuitive interface with built-in filtering.

**Before:**
```bash
aimgr repo add-command path/to/command.md
aimgr repo add-skill path/to/skill/
aimgr repo add-agent path/to/agent.md
```

**After:**
```bash
aimgr repo add path/to/resources
aimgr repo add ~/.opencode --filter "skill/*"
aimgr repo add ~/project/.claude --filter "agent/*"
```

The unified command:
- Auto-discovers resource types from directory structure
- Supports bulk imports from `.claude/`, `.opencode/`, etc.
- Provides pattern-based filtering with `--filter` flag
- Maintains backward compatibility with individual resource paths

## Bug Fixes

### Agent Detection Fix

Fixed critical bug where minimal agents (containing only required `description` field) were not being properly detected during repository scanning.

**Issue:** Agents without optional OpenCode-specific fields (`type`, `instructions`, `capabilities`) were being skipped.

**Fix:** Updated agent detection logic to correctly identify agents with minimal frontmatter.

### Bulk Import Conflict Handling

Fixed resource type detection bug in `AddBulk` that caused incorrect conflict handling.

**Issue:** During bulk imports, resource type was incorrectly determined from the target repository path instead of the source resource type, leading to false positive conflicts.

**Fix:** Corrected type detection to use actual resource type, ensuring proper conflict resolution.

## Configuration Updates

### Backward Compatibility

Added backward compatibility for the deprecated `default-tool` configuration key.

**Change:** The `config get` command now recognizes both `default-tool` (deprecated) and `default_tool` (current) keys, ensuring smooth transitions for existing users.

## Documentation

- Updated all documentation to reflect new `.metadata/` directory structure
- Added comprehensive pattern matching examples and syntax guide
- Documented unified `repo add` command with filter support
- Updated AGENTS.md with testing patterns and conventions

## Testing

- Added comprehensive unit tests for metadata migration utility
- Updated integration tests for new metadata structure  
- Added extensive test coverage for pattern matching functionality
- Verified migration on real-world repositories

## Breaking Changes

None. All changes are backward compatible:
- Old metadata format is automatically detected and can be migrated
- Previous `repo add-*` commands continue to work alongside unified `repo add`
- Configuration changes maintain backward compatibility

## Upgrade Notes

### Migrating Metadata

If you have an existing repository, migrate to the new metadata structure:

```bash
# Test migration first
aimgr repo migrate-metadata --dry-run

# Perform migration
aimgr repo migrate-metadata
```

The migration is safe and preserves all existing metadata. The old `.metadata.json` files are removed after successful migration.

### Using New Features

**Pattern matching is automatically available:**
```bash
aimgr install skill/*             # Install all skills
aimgr uninstall command/*-test    # Remove all test commands
```

**Unified repo add:**
```bash
aimgr repo add ~/.opencode                    # Import all resources
aimgr repo add ~/project/.claude --filter "skill/*"  # Import only skills
```

## Statistics

- **Commits:** 35+ commits since v1.2.0
- **Files Changed:** 50+ files across core packages
- **New Tests:** 100+ new test cases added
- **Test Coverage:** Maintained high coverage across all packages

## Contributors

Thank you to all contributors who made this release possible!

## What's Next

Looking ahead to v1.4.0:
- Enhanced resource validation
- Improved error messages and user feedback
- Performance optimizations for large repositories
- Additional pattern matching features

---

**Full Changelog:** https://github.com/hk9890/ai-config-manager/compare/v1.2.0...v1.3.0
