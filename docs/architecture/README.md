# Architecture Documentation

This directory contains technical design documentation for the ai-config-manager project. It covers architectural decisions, design rules, and system design principles that guide the codebase.

## Overview

**aimgr** is a CLI tool for managing AI resources (commands, skills, agents, packages) across multiple AI coding tools (Claude Code, OpenCode, GitHub Copilot). The architecture is built around several core principles:

### Core Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│                    (cobra commands)                          │
└───────────────────┬─────────────────────────────────────────┘
                    │
┌───────────────────┴─────────────────────────────────────────┐
│                   Business Logic Layer                       │
├──────────────────────────────────────────────────────────────┤
│  • Resource Management (resource/)                           │
│  • Repository Operations (repo/)                             │
│  • Installation/Symlinks (install/)                          │
│  • Auto-Discovery (discovery/)                               │
│  • Pattern Matching (pattern/)                               │
│  • Workspace Caching (workspace/)                            │
└───────────────────┬─────────────────────────────────────────┘
                    │
┌───────────────────┴─────────────────────────────────────────┐
│                   Infrastructure Layer                       │
├──────────────────────────────────────────────────────────────┤
│  • Configuration (config/)                                   │
│  • Git Operations (source/)                                  │
│  • Metadata Tracking (metadata/)                             │
│  • Tool Detection (tools/)                                   │
│  • XDG Directory Support                                     │
└──────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Resource-Centric**: Everything is a resource (command, skill, agent, package)
2. **Symlink-Based Installation**: Resources stored centrally, symlinked to tool directories
3. **Workspace Caching**: Git repositories cached for 10-50x performance improvement
4. **XDG Compliance**: Cross-platform directory conventions
5. **Pattern Matching**: Flexible resource filtering and selection
6. **Tool Agnostic**: Works with Claude Code, OpenCode, GitHub Copilot

### Storage Layout

```
~/.local/share/ai-config/
└── repo/
    ├── commands/          # Command resources
    ├── skills/            # Skill resources
    ├── agents/            # Agent resources
    ├── packages/          # Package resources
    ├── .workspace/        # Git repository cache
    └── .metadata/         # Metadata tracking
```

## Architecture Documentation

### Design Rules

- **[architecture-rules.md](architecture-rules.md)** - Strict architectural rules and patterns
  - Rule 1: All Git Operations Must Use Workspace Cache
  - Rule 2: XDG Base Directory Specification
  - Rule 3: Build Tags for Test Categories
  - Rule 4: Error Wrapping Requirements

### Component Documentation

For detailed component documentation, see:

- **[../workspace-caching.md](../workspace-caching.md)** - Git repository caching system
- **[../pattern-matching.md](../pattern-matching.md)** - Resource filtering and selection
- **[../resource-formats.md](../resource-formats.md)** - Resource specifications
- **[../output-formats.md](../output-formats.md)** - CLI output formatting
- **[../test-refactoring.md](../test-refactoring.md)** - Test organization

## Purpose

This architecture section serves to:

1. **Document Design Decisions**: Capture the "why" behind technical choices
2. **Establish Standards**: Define rules and patterns for consistent implementation
3. **Guide Development**: Provide clear architectural direction for contributors
4. **Enable Maintenance**: Make architectural intent explicit for future work

### When to Add Documentation Here

Add documentation to `docs/architecture/` when:

- Establishing new architectural rules or patterns
- Documenting system-wide design decisions
- Defining component interactions and boundaries
- Capturing performance-critical design choices
- Recording migration strategies for breaking changes

### When to Use Other Documentation

- **User-facing features**: Add to main docs/ directory
- **API documentation**: Use godoc comments in code
- **Development setup**: Update AGENTS.md or README.md
- **Package-specific details**: Add to pkg/ subdirectory READMEs

## Related Documentation

- **[../../AGENTS.md](../../AGENTS.md)** - Development guidelines for AI agents
- **[../../README.md](../../README.md)** - Project overview and usage
- **[../](../)** - General documentation directory
