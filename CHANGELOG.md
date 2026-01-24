# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-01-24

### Fixed
- **[CRITICAL]** Fixed P0 bug where Git source metadata stored temporary clone paths instead of original URLs
  - Resources from GitHub/Git URLs now correctly store source URLs in metadata
  - Enables successful updates for Git-sourced resources
  - Affects all resources added from Git sources since v1.0.0

### Added
- New `aimgr repo prune` command to remove orphaned metadata entries
  - Interactive confirmation with detailed list of orphaned entries
  - Supports `--dry-run` and `--force` flags
  - Focuses on local/file sources (Git sources are transient)
- New `aimgr repo verify` command for repository health checks
  - Detects resources without metadata, orphaned metadata, missing sources, type mismatches
  - Supports `--fix` flag for automatic resolution
  - Supports `--json` flag for machine-readable output
  - Returns appropriate exit codes for CI/CD integration
- Added `AIMGR_REPO_PATH` environment variable for test isolation
- Added `TEST_ISOLATION.md` documentation for test patterns

### Changed
- Improved `aimgr repo update` UX for handling missing source paths
  - Continues updating all resources even when some sources are missing
  - Distinguishes between skipped (⊘) and failed (✗) resources  
  - Provides actionable hints to resolve issues
  - Separates warnings from errors in summary

### Testing
- Added 25+ comprehensive integration tests for new commands
- All tests now use isolated temporary repositories
- Tests no longer pollute user's actual repository
- Can safely run tests in parallel

## [1.3.0] - 2026-01-24

### Added
- Metadata reorganization into cleaner `.metadata/` directory hierarchy
- Pattern matching support for install and uninstall commands
  - Supports glob operators: `*`, `?`, `[abc]`, `{a,b}`
  - Type-specific patterns: `skill/pdf*`, `command/*test*`
  - Cross-type patterns: `*review*`
- Unified `repo add` command with auto-discovery
  - Replaces separate `add-command`, `add-skill`, `add-agent` commands
  - Supports `--filter` flag for pattern-based filtering
  - Auto-discovers from `.claude/`, `.opencode/` directories

### Fixed
- Fixed agent detection for minimal agents (only `description` field)
- Fixed bulk import conflict handling due to incorrect type detection

### Changed
- Backward compatibility for deprecated `default-tool` configuration key

## [1.2.0] - 2026-01-23

### Added
- New `aimgr repo add bulk <folder|url>` command for auto-discovering and importing all resource types from a directory
  - Automatically detects commands, skills, and agents in standard locations (.claude/, .opencode/, etc.)
  - Simplifies bulk imports - no need to specify resource types individually
  - Supports Claude folders, OpenCode folders, and plugin directories
- Enhanced shell completion support:
  - Autocomplete for `aimgr list` with resource type filtering
  - Autocomplete for `aimgr config` with configuration keys
  - Autocomplete for `aimgr repo update` with resource names

### Changed
- Removed redundant commands in favor of unified bulk add:
  - Removed `aimgr repo add opencode` (use `aimgr repo add bulk` instead)
  - Removed `aimgr repo add claude` (use `aimgr repo add bulk` instead)
  - Removed `aimgr repo add plugin` (use `aimgr repo add bulk` instead)
- Updated README with simplified bulk import workflow
- Improved discovery logic to avoid duplicate detection in commands/skills subdirectories

### Fixed
- Discovery now correctly skips nested directories to prevent duplicate resource detection

## [1.1.0] - 2026-01-23

### Added
- Shell completion support for Bash, Zsh, Fish, and PowerShell
- Dynamic resource name completion for install commands
- Completion helper functions for resource types and config keys
- Shell completion troubleshooting guide

### Changed
- Enhanced CLI with completion flags
- Improved user experience with tab completion

## [1.0.0] - 2026-01-22

### Added
- Initial stable release
- Repository management for AI resources
- Multi-tool support (Claude Code, OpenCode, GitHub Copilot)
- Agent resource support with OpenCode and Claude formats
- Command and skill management
- Symlink-based installation
- Format validation and type safety
- Cross-platform support (Linux, macOS)
- Configuration management with XDG base directory support

[1.4.0]: https://github.com/hk9890/ai-config-manager/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/hk9890/ai-config-manager/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/hk9890/ai-config-manager/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/hk9890/ai-config-manager/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/hk9890/ai-config-manager/releases/tag/v1.0.0
