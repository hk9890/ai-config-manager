# Architecture Overview

This document provides a high-level architectural overview of the ai-config-manager project for contributors. For detailed architectural rules and implementation guidelines, see [docs/architecture/](../architecture/).

## Table of Contents
- [System Overview](#system-overview)
- [Package Structure](#package-structure)
- [Key Concepts](#key-concepts)
- [Data Flow](#data-flow)
- [Directory Layout](#directory-layout)
- [Related Documentation](#related-documentation)

---

## System Overview

**aimgr** (AI Config Manager) is a command-line tool that manages AI resources across multiple AI coding environments. It solves the problem of resource fragmentation by providing:

### What It Does

1. **Centralized Repository**: Maintains a single source of truth for all AI resources in `~/.local/share/ai-config/repo/` (configurable via `repo.path` or `AIMGR_REPO_PATH`)
2. **Multi-Tool Support**: Installs resources to Claude Code (`.claude/`), OpenCode (`.opencode/`), and GitHub Copilot (`.github/`)
3. **Symlink-Based Installation**: Creates symlinks from tool directories to the central repository (no duplication)
4. **Git Integration**: Imports resources from Git repositories with intelligent caching
5. **Pattern Matching**: Install/uninstall multiple resources using glob patterns (`skill/pdf*`)
6. **Package Management**: Group resources into packages for easy distribution

### Resource Types

The system manages four types of resources:

- **Commands**: Single-file markdown commands (e.g., `build.md`, `api/deploy.md`)
- **Skills**: Multi-file skills with documentation and resources (e.g., `pdf-processing/`)
- **Agents**: Single-file agent definitions with YAML frontmatter (e.g., `code-reviewer.md`)
- **Packages**: Collections of resources grouped together (e.g., `web-tools.package.json`)

### Architecture Style

- **Language**: Go 1.25.6
- **CLI Framework**: Cobra for command structure
- **Storage**: XDG Base Directory Specification compliant
- **Installation**: Symlink-based (not file copying)
- **Caching**: Workspace-based Git repository caching

---

## Package Structure

The codebase is organized into three main directories:

```
ai-config-manager/
├── cmd/              # CLI command definitions
├── pkg/              # Core business logic packages
├── test/             # Integration tests
├── examples/         # Example resources
└── main.go          # Application entry point
```

### cmd/ - Command Definitions

Cobra command tree mirroring the CLI structure:

```
cmd/
├── root.go           # Root command and global flags
├── install.go        # aimgr install
├── uninstall.go      # aimgr uninstall
├── list.go          # aimgr list
├── repo/            # aimgr repo subcommands
│   ├── repo.go      # Repo command group
│   ├── import.go    # aimgr repo import
│   ├── sync.go      # aimgr repo sync
│   ├── list.go      # aimgr repo list
│   └── prune.go     # aimgr repo prune
└── marketplace/     # aimgr marketplace subcommands
    └── import.go    # aimgr marketplace import
```

**Pattern**: Each command file contains Cobra command setup and delegates to pkg/ packages for logic.

### pkg/ - Business Logic

Core packages organized by responsibility:

| Package | Purpose |
|---------|---------|
| `config/` | Configuration file parsing and validation |
| `discovery/` | Auto-discovery of resources in directories |
| `install/` | Symlink creation and installation logic |
| `manifest/` | `ai.package.yaml` project manifest handling |
| `marketplace/` | Marketplace JSON parsing and generation |
| `metadata/` | Metadata tracking and schema migration |
| `pattern/` | Resource pattern matching (glob-style) |
| `repo/` | Central repository management (add, list, remove) |
| `resource/` | Resource type definitions and loaders |
| `source/` | Git operations and source parsing |
| `tools/` | Tool-specific information (Claude, OpenCode, Copilot) |
| `version/` | Version information embedded at build time |
| `workspace/` | Git repository caching and workspace management |

**Pattern**: Each package has a clear, single responsibility. Packages do not depend on `cmd/`.

### test/ - Integration Tests

End-to-end integration tests that exercise the full system:

```
test/
├── integration_test.go      # Main integration test suite
└── testdata/               # Test fixtures and resources
```

**Pattern**: Unit tests live alongside code in `pkg/*/`, integration tests live in `test/`.

---

## Key Concepts

### 1. Central Repository

**Default Location**: `~/.local/share/ai-config/repo/` (configurable)

The repository is the single source of truth for all resources:

```
~/.local/share/ai-config/repo/
├── commands/
│   ├── build.md
│   └── api/
│       └── deploy.md
├── skills/
│   └── pdf-processing/
│       ├── SKILL.md
│       └── resources/
├── agents/
│   └── code-reviewer.md
├── packages/
│   └── web-tools.package.json
└── .metadata/
    ├── commands.json
    ├── skills.json
    └── agents.json
```

**Why**: Single location eliminates duplication, simplifies updates, enables version control.

### 2. Resources

Resources are structured artifacts with metadata:

```go
type Resource struct {
    Type        ResourceType  // command, skill, agent, package
    Name        string       // e.g., "build", "api/deploy", "pdf-processing"
    Description string       // Human-readable description
    Path        string       // Absolute path to resource
    Source      string       // Origin (local path or git URL)
    Tool        tools.Tool   // Target tool (Claude, OpenCode, Copilot)
}
```

**Key Properties**:
- Resources have **types** (command, skill, agent, package)
- Resources have **names** (lowercase alphanumeric + hyphens, max 64 chars)
- Commands support **nested names** (`api/deploy`)
- Resources track their **source** for updates

### 3. Symlink-Based Installation

Installation creates symlinks, not copies:

```
# Repository (source)
~/.local/share/ai-config/repo/commands/build.md

# Installation (symlink)
.claude/commands/build.md -> ~/.local/share/ai-config/repo/commands/build.md
```

**Benefits**:
- No duplication (saves disk space)
- Updates propagate automatically (update repo, all installs updated)
- Uninstall is simple (remove symlink, keep source)

### 4. Metadata Tracking

The `.metadata/` directory tracks installed resources:

```json
{
  "version": 2,
  "resources": [
    {
      "name": "build",
      "description": "Build the project",
      "path": "/absolute/path/to/repo/commands/build.md",
      "source": "gh:owner/repo",
      "added": "2026-01-29T12:00:00Z"
    }
  ]
}
```

**Purpose**: Enables listing, updating, and removing resources without scanning filesystem.

### 5. Workspace Caching

Git repositories are cached in `.workspace/` for performance:

```
~/.local/share/ai-config/repo/.workspace/
└── github.com/
    └── hk9890/
        └── ai-tools/
            └── main/          # Cached clone at 'main' ref
                ├── .git/
                └── resources/
```

**Performance**: 10-50x faster for repeated operations (subsequent imports, syncs).

---

## Data Flow

### Flow 1: Import from Git Repository

```
User runs: aimgr repo import gh:owner/repo

1. Parse source → Extract Git URL and ref
2. Workspace cache → Get or clone repository
3. Discovery → Find all resources in clone
4. Repository → Add resources to central repo
5. Metadata → Track added resources
6. Output → Display results to user
```

**Packages Involved**: `source` → `workspace` → `discovery` → `repo` → `metadata`

### Flow 2: Install Resource to Tool

```
User runs: aimgr install skill/pdf-processing --tool=claude

1. Pattern matching → Resolve pattern to resources
2. Repository → Find resource in central repo
3. Tool detection → Get tool directories (.claude/)
4. Install → Create symlink to target directory
5. Manifest → Update ai.package.yaml (if exists)
6. Output → Confirm installation
```

**Packages Involved**: `pattern` → `repo` → `tools` → `install` → `manifest`

### Flow 3: Sync Resources (Update from Git)

```
User runs: aimgr repo sync

1. Metadata → Load all resources with Git sources
2. Workspace → Update cached repositories (git pull)
3. Discovery → Find updated resources
4. Repository → Replace resources in central repo
5. Metadata → Update tracking info
6. Output → Show updated resources
```

**Packages Involved**: `metadata` → `workspace` → `discovery` → `repo` → `metadata`

---

## Directory Layout

### User Data Directory

The application uses XDG Base Directory Specification:

**Linux/macOS**:
```
~/.local/share/ai-config/     # Data directory
├── repo/                     # Central repository
│   ├── commands/
│   ├── skills/
│   ├── agents/
│   ├── packages/
│   ├── .metadata/           # Resource tracking
│   └── .workspace/          # Git cache
└── config.yaml              # Optional user config
```

**Windows** (automatically mapped):
```
%LOCALAPPDATA%\ai-config\    # Data directory
```

### Project Installation Targets

Tool-specific directories (where symlinks are created):

```
project-root/
├── .claude/
│   ├── commands/            # Claude Code commands
│   ├── skills/              # Claude Code skills
│   └── agents/              # Claude Code agents
├── .opencode/
│   ├── commands/            # OpenCode commands
│   ├── skills/              # OpenCode skills
│   └── agents/              # OpenCode agents
├── .github/
│   └── skills/              # GitHub Copilot skills
└── ai.package.yaml          # Project manifest (optional)
```

### Tool Detection

The system automatically detects which tool to use based on:

1. **Explicit flag**: `--tool=claude`
2. **Project detection**: Presence of `.claude/`, `.opencode/`, `.github/`
3. **Manifest**: Tool specified in `ai.package.yaml`

---

## Related Documentation

### Architecture Details

For in-depth architectural information:

- **[Architecture Rules](../architecture/architecture-rules.md)** - Strict architectural rules and patterns (Git operations, XDG directories, error handling)

### Implementation Guides

For detailed implementation specifications:

- **[Resource Formats](../resource-formats.md)** - Complete format specifications for all resource types
- **[Workspace Caching](../workspace-caching.md)** - Git repository caching implementation and performance
- **[Pattern Matching](../pattern-matching.md)** - Pattern syntax and matching logic
- **[Output Formats](../output-formats.md)** - CLI output formats (JSON, YAML, table)

### Development

For development workflows:

- **[AGENTS.md](../../AGENTS.md)** - AI agent development guidelines
- **[Release Process](release-process.md)** - Release and versioning workflow

---

## Next Steps

- **Read**: [Architecture Rules](../architecture/architecture-rules.md) for strict coding patterns
- **Build**: Run `make build` to compile the project
- **Test**: Run `make test` to run the test suite
- **Contribute**: Follow patterns in existing code, add tests for new features
