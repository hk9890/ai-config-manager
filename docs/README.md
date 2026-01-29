# Documentation Index

Welcome to the **aimgr** (ai-config-manager) documentation! This page serves as the central index for all documentation in the project.

## Documentation Structure

The documentation is organized into the following sections:

### 📘 [User Guide](user-guide/)

User-facing documentation covering all aspects of using **aimgr** effectively.

**What's inside:**
- [Getting Started Guide](user-guide/README.md) - Quick start and overview
- [Pattern Matching](user-guide/pattern-matching.md) - Filter resources with patterns (`skill/*`, `command/test`)
- [Resource Formats](user-guide/resource-formats.md) - Complete specifications for commands, skills, agents, and packages
- [Output Formats](user-guide/output-formats.md) - Control CLI output (JSON, YAML, table) for scripting and automation
- [GitHub Sources](user-guide/github-sources.md) - Import resources from GitHub repositories (`gh:owner/repo`)
- [Workspace Caching](user-guide/workspace-caching.md) - Understand Git repository caching for 10-50x performance improvements

**Use this section for:** Learning how to use aimgr, understanding resource formats, scripting with CLI output, and optimizing performance.

---

### 🛠️ [Contributor Guide](contributor-guide/)

Documentation for developers contributing to the **aimgr** project.

**What's inside:**
- [Contributor Overview](contributor-guide/README.md) - Getting started with development
- [Release Process](contributor-guide/release-process.md) - Step-by-step guide for creating and publishing releases

**See also:**
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Main contributor guide (setup, workflow, code style)
- [Architecture Documentation](#-architecture) - Technical design and system architecture

**Use this section for:** Setting up a development environment, understanding the contribution workflow, and managing releases.

---

### 🏗️ [Architecture](architecture/)

Technical design documentation covering architectural decisions, design rules, and system design principles.

**What's inside:**
- [Architecture Overview](architecture/README.md) - System design, core principles, and component layout
- [Architecture Rules](architecture/architecture-rules.md) - Strict architectural rules and patterns
  - Git operations with workspace cache
  - XDG Base Directory compliance
  - Build tags for test categories
  - Error wrapping requirements

**Use this section for:** Understanding design decisions, following architectural patterns, and maintaining consistency across the codebase.

---

### 📦 [Planning Archive](planning/)

Historical planning and analysis documents from the development of **aimgr**.

**What's inside:**
- Technical investigations and evaluations
- Implementation strategies and refactoring plans
- Historical decision-making context

**Note:** These documents are for reference and historical context. They are **not required for general use** of aimgr. See the [User Guide](#-user-guide) for current documentation.

**Use this section for:** Understanding past architectural decisions, finding context for refactoring needs, and learning from historical development processes.

---

### 📜 [Release Notes Archive](archive/release-notes/)

Archived release notes from previous versions of **aimgr**.

**What's inside:**
- Historical release announcements (v0.3.0 through v1.4.0)
- Feature additions and bug fixes from past releases

**For current releases, see:**
- [CHANGELOG.md](../CHANGELOG.md) - Active changelog in the root directory
- [GitHub Releases](https://github.com/hk9890/ai-config-manager/releases) - Official releases with binaries

**Use this section for:** Historical reference and understanding the evolution of the project.

---

## For AI Agents

If you're an AI coding agent working on this repository, see **[AGENTS.md](../AGENTS.md)** for:
- Quick reference for build & test commands
- Code style guidelines
- Common patterns and resource loading
- Testing best practices
- Essential development workflows

This guide is specifically designed to help AI agents work effectively in the ai-config-manager codebase.

---

## Quick Links

### Project Information
- **[README.md](../README.md)** - Project overview, installation, and command reference
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - How to contribute to aimgr
- **[CHANGELOG.md](../CHANGELOG.md)** - Current changelog

### External Resources
- **[GitHub Repository](https://github.com/hk9890/ai-config-manager)** - Source code and issue tracker
- **[GitHub Releases](https://github.com/hk9890/ai-config-manager/releases)** - Download releases and view release notes

---

## Getting Help

- **Questions**: [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)
- **Bug Reports**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Feature Requests**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) and the [Contributor Guide](contributor-guide/) to get started.
