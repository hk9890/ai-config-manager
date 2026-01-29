# User Guide

This directory contains user-facing documentation for **aimgr** (ai-config-manager), a CLI tool for managing AI resources across multiple AI coding tools.

## Overview

The user guide provides comprehensive documentation for:
- **Resource Management**: Understanding and working with commands, skills, agents, and packages
- **Pattern Matching**: Using patterns to filter and select resources
- **Output Formats**: Controlling CLI output for scripting and automation
- **GitHub Integration**: Working with GitHub repositories as resource sources
- **Performance Optimization**: Leveraging workspace caching for efficient Git operations

## Documentation Index

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

## Getting Started

New to aimgr? Start with these steps:

1. **Install aimgr**: Build and install the CLI tool
   ```bash
   make install
   ```

2. **Initialize your first resource**: Import resources from a directory or GitHub
   ```bash
   aimgr repo import ~/.opencode
   aimgr repo import gh:owner/repo
   ```

3. **List available resources**: See what's in your repository
   ```bash
   aimgr repo list
   ```

4. **Install resources**: Install resources to your AI tools
   ```bash
   aimgr install skill/pdf-processing --tool=claude
   ```

## Quick Reference

| Command | Description |
|---------|-------------|
| `aimgr repo import <source>` | Import resources from directory or GitHub |
| `aimgr repo list` | List all resources in repository |
| `aimgr repo sync <source>` | Sync and update resources from source |
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
