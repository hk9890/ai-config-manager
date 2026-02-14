# User Guide

This directory contains user-facing documentation for **aimgr** (ai-config-manager), a CLI tool for managing AI resources across multiple AI coding tools.

## Overview

The user guide provides comprehensive documentation for:
- **Resource Management**: Understanding and working with commands, skills, agents, and packages
- **Source Management**: Tracking sources in `ai.repo.yaml` for automatic synchronization
- **Configuration**: Customizing repository paths and installation targets
- **Pattern Matching**: Using patterns to filter and select resources
- **Output Formats**: Controlling CLI output for scripting and automation
- **GitHub Integration**: Working with GitHub repositories as resource sources
- **Performance Optimization**: Leveraging workspace caching for efficient Git operations
- **Resource Validation**: Testing resources before publishing

## Documentation Index

### [Getting Started](getting-started.md)
**Start here if you're new to aimgr!** This guide covers installation, first steps, common operations, and practical workflows.

**Key Topics:**
- Installation on Linux, macOS, and Windows
- Configuring your AI tool targets
- Adding sources with `repo add` (tracks in `ai.repo.yaml`)
- Installing resources into projects
- Common operations (install, uninstall, list, sync)
- Practical workflows and examples
- Troubleshooting tips

### [Configuration](configuration.md)
Complete guide to configuring aimgr, including repository path customization and installation targets.

**Key Topics:**
- Config file location (`~/.config/aimgr/aimgr.yaml`)
- Repository path configuration (config file vs environment variable)
- Precedence rules (ENV > config > XDG default)
- Path expansion (tilde, relative, absolute)
- Installation targets configuration
- Complete examples and troubleshooting

**Note:** Source management is configured in `ai.repo.yaml` (see [Sync Sources](sync-sources.md)).

### [Sync Sources](sync-sources.md)
Comprehensive guide to source management using `ai.repo.yaml`, explaining remote URLs vs local paths, symlink behavior, and development workflows.

**Key Topics:**
- `ai.repo.yaml` manifest format and structure
- Adding sources with `repo add` command
- Remote sources: Git repositories (copied, mode=copy)
- Local sources: Filesystem paths (symlinked, mode=symlink)
- Syncing resources with `repo sync`
- Development workflows and best practices
- Troubleshooting broken symlinks and permission issues

### [Pattern Matching](pattern-matching.md)
Learn how to use pattern matching to filter resources when listing, installing, or removing them. Covers syntax, wildcards, type-specific patterns, and practical examples.

**Key Topics:**
- Pattern syntax and wildcards (`*`, `**`)
- Type-specific patterns (`skill/pdf*`, `command/test`)
- Nested resource matching
- Common use cases and examples

### [Resource Formats](resource-formats.md)
Complete reference for all resource types supported by aimgr, including file formats, structure requirements, and examples.

**Key Topics:**
- Command format (Markdown files in `commands/` directories)
- Skill format (directories with `SKILL.md`)
- Agent format (Markdown with YAML frontmatter)
- Package format (`.package.json` files)
- Project manifests (`ai.package.yaml`)
- Marketplace format (`marketplace.json`)

### [Output Formats](output-formats.md)
Guide to CLI output formats for human-readable display and programmatic consumption.

**Key Topics:**
- Table format (default, human-readable)
- JSON format (structured, for scripting)
- YAML format (structured, human-readable)
- Commands supporting `--format` flag
- Scripting examples and best practices

### [GitHub Sources](github-sources.md)
Documentation for importing resources directly from GitHub repositories using the `gh:` prefix.

**Key Topics:**
- GitHub URL syntax (`gh:owner/repo`, `gh:owner/repo@branch`)
- Importing from public and private repositories
- Authentication and token management
- Path-specific imports (`gh:owner/repo/path/to/resources`)
- Filtering imports with patterns

### [Workspace Caching](workspace-caching.md)
Understanding workspace caching for Git repositories, which dramatically improves performance for repeated operations.

**Key Topics:**
- How workspace caching works
- Performance benefits (10-50x faster)
- Cache location and management
- Cache pruning and maintenance
- Troubleshooting cache issues

### [Validating Resources](validating-resources.md)
Guide for resource developers on validating skills, agents, commands, and packages before publishing them.

**Key Topics:**
- Quick validation workflow using `--dry-run`
- Validating skills, agents, commands, and packages
- Testing installation locally
- Common validation errors and fixes
- CI/CD integration for automated testing

### [Resource Developer Guide](developer-guide.md)
Developer guide for creating and testing AI resources (skills, agents, commands, packages) that work with aimgr.

**Key Topics:**
- Resource validation methods
- Local testing workflows
- Common validation errors and solutions
- Best practices for resource development
- Publishing resources to repositories

### [Git-Backed Repository Tracking](git-tracking.md)
Learn how to enable git-based tracking of your AI resource repository for full auditability, change history, and recoverability.

**Key Topics:**
- Initializing git tracking with `aimgr repo init`
- Automatic commits for all import/sync/remove operations
- Viewing change history and reverting operations
- Working with remote repositories for team collaboration
- Branch strategies and tagging milestones
- Integration with CI/CD workflows

## Getting Started

New to aimgr? **Read the [Getting Started Guide](getting-started.md)** for a complete tutorial.

Quick start:

1. **Install aimgr**: Download binary or build from source
   ```bash
   make install
   ```

2. **Initialize repository**: Create `ai.repo.yaml` manifest
   ```bash
   aimgr repo init
   ```

3. **Add sources**: Track sources in `ai.repo.yaml`
   ```bash
   aimgr repo add gh:hk9890/ai-tools
   aimgr repo add ~/my-local-resources
   ```

4. **Install resources**: Install resources to your projects
   ```bash
   cd ~/my-project
   aimgr install skill/pdf-processing
   ```

## Quick Reference

| Command | Description |
|---------|-------------|
| `aimgr repo init` | Initialize repository and create `ai.repo.yaml` |
| `aimgr repo add <source>` | Add source to `ai.repo.yaml` (imports resources) |
| `aimgr repo sync` | Sync all sources from `ai.repo.yaml` |
| `aimgr repo info` | View sources and repository status |
| `aimgr repo list` | List all resources in repository |
| `aimgr list [tool]` | List installed resources for a tool |
| `aimgr install <pattern>` | Install resources matching pattern |
| `aimgr uninstall <pattern>` | Uninstall resources matching pattern |
| `aimgr repo prune` | Clean up workspace cache |

## Related Documentation

For developer documentation and contributing guidelines, see:
- [AGENTS.md](../../AGENTS.md) - Guidelines for AI coding agents
- [docs/](../) - Technical documentation and architecture

## Support

For issues, questions, or contributions:
- GitHub Issues: Report bugs and request features
- GitHub Discussions: Ask questions and share ideas
- Pull Requests: Contribute improvements and fixes
