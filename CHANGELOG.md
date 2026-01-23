# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.2.0]: https://github.com/hk9890/ai-config-manager/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/hk9890/ai-config-manager/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/hk9890/ai-config-manager/releases/tag/v1.0.0
