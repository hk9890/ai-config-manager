# Documentation Index

This directory contains detailed documentation for various aspects of `aimgr`.

## User Documentation

### [Output Formats](output-formats.md)
Comprehensive guide to using the `--format` flag for structured output (JSON, YAML, table). Includes examples for scripting, automation, and CI/CD integration.

**Use this when:**
- Using `aimgr` in scripts or automation
- Parsing command output with `jq` or other tools
- Implementing CI/CD pipelines
- Generating audit logs

### [GitHub Sources](github-sources.md)
Guide to adding resources from GitHub repositories. Covers GitHub URLs, auto-discovery, and workspace caching.

**Use this when:**
- Adding resources from GitHub repositories
- Understanding source URL formats (`gh:owner/repo`, `owner/repo@tag`)
- Working with Git-based resource sources
- Troubleshooting GitHub imports

## Developer Documentation

### [Release Process](RELEASE.md)
Step-by-step guide for creating and publishing releases. Covers version management, GoReleaser configuration, and release workflow.

**Use this when:**
- Creating a new release
- Understanding the release workflow
- Troubleshooting release issues
- Contributing to release automation

## Additional Resources

- **Main README**: [../README.md](../README.md) - Getting started, installation, and command reference
- **Agent Guide**: [../AGENTS.md](../AGENTS.md) - Guidelines for AI coding agents working in this repository
- **Contributing**: [../CONTRIBUTING.md](../CONTRIBUTING.md) - How to contribute to `aimgr`

## Archive

Historical documentation and completed proposals are stored in the [archive/](archive/) directory.
