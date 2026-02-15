# AGENTS.md

Quick reference for AI coding agents working in the ai-config-manager repository.

## Project Overview

**aimgr** is a CLI tool (Go 1.25.6) for managing AI resources (commands, skills, agents, packages) across multiple AI coding tools. It uses a centralized repository (`~/.local/share/ai-config/repo/`) with symlink-based installation to tool directories (`.claude/`, `.opencode/`, `.github/skills/`, `.windsurf/skills/`).

**Architecture**: CLI (Cobra) → Business Logic (`pkg/`) → Storage (XDG directories)

## Quick Commands

```bash
# Build & Install
make build      # Build binary
make install    # Build and install to ~/bin

# Testing
make test                # All tests (vet → unit → integration)
make unit-test           # Fast unit tests only
make integration-test    # Slow integration tests

# Code Quality
make fmt        # Format all Go code
make vet        # Run go vet
```

## Use Case Guide

Delegate to the right documentation based on what you need to do:

### Writing Code / Implementing Features

→ See **[docs/contributor-guide/code-style.md](docs/contributor-guide/code-style.md)** for:
- Naming conventions (files, packages, types, resources)
- Import organization (3 groups: stdlib, external, internal)
- Error handling (always wrap with `%w`)
- File operations (paths, permissions, defer)
- Symlink handling (CRITICAL: use `os.Stat()` not `entry.IsDir()`)
- Best practices summary

→ See **[docs/contributor-guide/architecture.md](docs/contributor-guide/architecture.md)** for:
- System overview and components
- Package structure (17 packages) and responsibilities
- Architecture rules (5 critical rules)
- Data flows (import, install, sync)
- Key design patterns

### Writing Tests / Analyzing Test Failures

→ See **[docs/contributor-guide/testing.md](docs/contributor-guide/testing.md)** for:
- Test types (unit, integration, E2E)
- Test isolation with `t.TempDir()`
- Table-driven test patterns
- Running specific tests
- Common test failures and fixes

### Planning / Adding New Functionality

→ See **[docs/contributor-guide/architecture.md](docs/contributor-guide/architecture.md)** for:
- System overview and components
- Package responsibilities (`pkg/` structure)
- Architecture rules you MUST follow
- Data flows for different operations

### Releasing New Versions

→ Use the **github-releases skill** (if available in your environment)

If not available, see **[docs/contributor-guide/release-process.md](docs/contributor-guide/release-process.md)**

## Before Committing

1. `make fmt` - Format code
2. `make test` - All tests pass
3. Follow code style guide (see docs/contributor-guide/code-style.md)
4. Git operations use `pkg/workspace` (see architecture.md)
5. Tests use `t.TempDir()` and `NewManagerWithPath()` (see testing.md)
