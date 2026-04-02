# Documentation Index

Welcome to the **aimgr** (ai-config-manager) documentation. This page is the central index for project docs, contributor guides, and repo-specific workflow references.

## Repo Workflow Docs

These top-level docs capture repository-specific guidance used by contributors and AI agents:

- [CODING.md](CODING.md) - Build commands, project structure, conventions, and safety rules
- [TESTING.md](TESTING.md) - Test commands, isolation requirements, and concurrency expectations
- [RELEASING.md](RELEASING.md) - Repo-specific release companion for the `github-releases` skill
- [MONITORING.md](MONITORING.md) - Local logs, health signals, and monitoring triage workflow
- [PULL-REQUESTS.md](PULL-REQUESTS.md) - Branch workflow, PR expectations, and review follow-up

## User Guide

User-facing documentation for day-to-day aimgr usage:

- [Concepts](user-guide/concepts.md) - Mental model for repositories, manifests, and resources
- [Getting Started](user-guide/getting-started.md) - Installation, first steps, and common operations
- [Configuration](user-guide/configuration.md) - Repository path, installation targets, and field mappings
- [Repairing Resources](user-guide/repair.md) - Reconcile installs and clean broken project state safely
- [Sources](user-guide/sources.md) - Manage remote and local resource sources
- [Team and Multi-Project Workflows](user-guide/team-workflows.md) - Shared-manifest and multi-project patterns

## Reference

Technical reference material for exact syntax and supported features:

- [Supported Tools](reference/supported-tools.md) - AI tool compatibility and resource support
- [Pattern Matching](reference/pattern-matching.md) - Glob patterns for filtering and matching resources
- [Output Formats](reference/output-formats.md) - Table, JSON, and YAML output for scripting
- [Resource Validation](reference/resource-validation.md) - Validate resources by path or canonical ID
- [Troubleshooting](reference/troubleshooting.md) - Common issues and recovery guidance

## Contributor Guide

Deeper contributor documentation for development workflows:

- [Overview](contributor-guide/README.md) - Contributor guide entrypoint
- [Development Environment](contributor-guide/development-environment.md) - Tooling and local setup
- [Code Style](contributor-guide/code-style.md) - Naming, imports, error handling, and best practices
- [Architecture](contributor-guide/architecture.md) - System design, package structure, and design rules
- [Testing](contributor-guide/testing.md) - Test types, patterns, and troubleshooting
- [Release Process](contributor-guide/release-process.md) - Full release process details

See also [CONTRIBUTING.md](../CONTRIBUTING.md) for the contribution workflow and [PULL-REQUESTS.md](PULL-REQUESTS.md) for PR-specific expectations.

## Internals

Implementation details for contributors and advanced users:

- [Repository Layout](internals/repository-layout.md) - Internal folder structure of an aimgr repository
- [Workspace Caching](internals/workspace-caching.md) - How the workspace cache optimizes Git operations
- [Git Tracking](internals/git-tracking.md) - How aimgr tracks repository changes with Git

## For AI Agents

If you are working in this repository as an AI coding agent, start with [AGENTS.md](../AGENTS.md) and then follow the routed topic docs above.

## Quick Links

### Project Information
- [README.md](../README.md) - Project overview, installation, and command reference
- [CONTRIBUTING.md](../CONTRIBUTING.md) - How to contribute to aimgr
- [CHANGELOG.md](../CHANGELOG.md) - Version history

### External Resources
- [GitHub Repository](https://github.com/dynatrace-oss/ai-config-manager) - Source code and issue tracker
- [GitHub Releases](https://github.com/dynatrace-oss/ai-config-manager/releases) - Download releases and binaries
- [AgentSkills.io](https://agentskills.io/home) - Community skill format specification

## Getting Help

- Questions: [GitHub Discussions](https://github.com/dynatrace-oss/ai-config-manager/discussions)
- Bug Reports: [GitHub Issues](https://github.com/dynatrace-oss/ai-config-manager/issues)
- Feature Requests: [GitHub Issues](https://github.com/dynatrace-oss/ai-config-manager/issues)
