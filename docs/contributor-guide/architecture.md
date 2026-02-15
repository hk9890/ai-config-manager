# Architecture Guide

High-level architecture overview and design rules for ai-config-manager contributors.

## System Overview

**aimgr** manages AI resources across multiple AI coding environments with:

1. **Centralized Repository**: Single source at `~/.local/share/ai-config/repo/` (configurable)
2. **Multi-Tool Support**: Claude Code (`.claude/`), OpenCode (`.opencode/`), GitHub Copilot (`.github/`), Windsurf (`.windsurf/`)
3. **Symlink-Based Installation**: No duplication - symlinks to central repository
4. **Git Integration**: Imports from Git with intelligent workspace caching
5. **Pattern Matching**: Install/uninstall using glob patterns (`skill/pdf*`)

### Resource Types

- **Commands**: Single `.md` files (e.g., `build.md`, `api/deploy.md`)
- **Skills**: Directories with `SKILL.md` (e.g., `pdf-processing/`)
- **Agents**: Single `.md` files with YAML frontmatter (e.g., `code-reviewer.md`)
- **Packages**: Collections in `.package.json` (e.g., `web-tools.package.json`)

### Architecture

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

### Storage Layout

```
~/.local/share/ai-config/repo/
├── commands/          # Command resources
├── skills/            # Skill resources
├── agents/            # Agent resources
├── packages/          # Package resources
├── .workspace/        # Git repository cache
└── .metadata/         # Metadata tracking
```

## Package Structure

### cmd/ - Command Definitions

Cobra command tree matching CLI structure. Each command delegates to `pkg/` for logic.

### pkg/ - Business Logic


| Package | Purpose |
|---------|---------|
| `config/` | Configuration file parsing and validation |
| `discovery/` | Auto-discovery of resources in directories |
| `errors/` | Error categories and structured error handling |
| `install/` | Symlink creation and installation logic |
| `manifest/` | Project manifests (ai.package.yaml) management |
| `marketplace/` | Marketplace discovery and parsing (marketplace.json) |
| `metadata/` | Resource metadata tracking (.metadata/ directory) |
| `output/` | Output formatting (JSON, YAML, tables) |
| `pattern/` | Glob pattern matching for resource selection |
| `repo/` | Central repository management (add, list, remove) |
| `repomanifest/` | Repository manifests (ai.repo.yaml) management |
| `resource/` | Resource type definitions and loaders |
| `source/` | Git operations and source parsing |
| `sourcemetadata/` | Source state tracking (.metadata/sources.json) |
| `tools/` | Tool-specific information (Claude, OpenCode, Copilot, Windsurf) |
| `version/` | Version information (build-time injection) |
| `workspace/` | Git repository caching (10-50x performance) |


### test/ - Integration Tests

End-to-end tests exercising the full system.

## Architecture Rules

### Rule 1: Git Operations Use Workspace Cache

**All Git operations MUST use `pkg/workspace` cache.** Direct temporary cloning is prohibited.

**Why**: 10-50x faster for repeated operations, single source of truth, simplified error handling.

**Correct Usage**:
```go
import "github.com/hk9890/ai-config-manager/pkg/workspace"

mgr, _ := workspace.NewManager(repoPath)
clonePath, err := mgr.GetOrClone(gitURL, ref)  // Cached
// Use clonePath - no cleanup needed
```

**Prohibited**:
```go
// WRONG: Temporary directory clone
tempDir, _ := os.MkdirTemp("", "git-clone-*")
defer os.RemoveAll(tempDir)
cmd := exec.Command("git", "clone", url, tempDir)
```

**Exceptions**: Unit/integration tests may use temporary directories.

### Rule 2: XDG Base Directory Specification

**All application data MUST follow XDG Base Directory Specification.**

```go
import "github.com/adrg/xdg"

// Data: ~/.local/share/ai-config/repo/
repoPath := filepath.Join(xdg.DataHome, "ai-config", "repo")

// Config: ~/.config/ai-config/
configPath := filepath.Join(xdg.ConfigHome, "ai-config")
```

**Why**: Cross-platform, respects environment variables, integrates with backup tools.

### Rule 3: Build Tags for Test Categories

**Tests MUST use build tags** to categorize: `unit` or `integration`.

```go
//go:build unit
package mypackage_test
func TestFast(t *testing.T) { }

//go:build integration
package test
func TestSlow(t *testing.T) { }
```

**Running Tests**:
```bash
go test -tags=unit ./...           # Fast unit tests
go test -tags=integration ./...    # Slow integration tests
go test -tags="unit integration" ./...  # All tests
```

### Rule 4: Error Wrapping Requirements

**All errors MUST be wrapped** with context using `fmt.Errorf` with `%w`.

```go
// ✅ CORRECT
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}

// ❌ WRONG: No context
if err != nil {
    return err
}

// ❌ WRONG: Loses error chain
if err != nil {
    return fmt.Errorf("error: %s", err.Error())
}
```

**Why**: Full error chain for debugging, enables `errors.Is()` and `errors.As()`, clear user messages.

### Rule 5: Symlink Handling

**All filesystem traversal MUST support both real files (COPY mode) and symlinks (SYMLINK mode).**

**Problem**: `entry.IsDir()` from `os.ReadDir()` returns `false` for symlinks to directories.

**Correct Usage**:
```go
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    
    // Follow symlinks to check if target is a directory
    info, err := os.Stat(path)  // os.Stat follows symlinks
    if err != nil || !info.IsDir() { continue }
    
    processDirectory(path)  // Works for both real and symlinked dirs
}
```

**Prohibited**:
```go
// WRONG: Skips symlinked directories
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    if entry.IsDir() {  // ← Returns false for symlinks!
        processDirectory(entry.Name())
    }
}
```

**Testing Requirements**: Every discovery function MUST test both real and symlinked resources.

## Key Concepts

### Central Repository

Single source at `~/.local/share/ai-config/repo/` (configurable via `repo.path` or `AIMGR_REPO_PATH`).

### Resources

```go
type Resource struct {
    Type        ResourceType  // command, skill, agent, package
    Name        string        // e.g., "build", "api/deploy", "pdf-processing"
    Description string
    Path        string        // Absolute path to resource
    Source      string        // Origin (local path or git URL)
}
```

### Symlink-Based Installation

```
# Repository (source)
~/.local/share/ai-config/repo/commands/build.md

# Installation (symlink)
.claude/commands/build.md -> ~/.local/share/ai-config/repo/commands/build.md
```

**Benefits**: No duplication, automatic updates, simple uninstall.

### Workspace Caching

Git repositories cached in `.workspace/` for 10-50x performance improvement:

```
~/.local/share/ai-config/repo/.workspace/
└── github.com/owner/repo/main/  # Cached clone at 'main' ref
```

## Data Flows

### Flow 1: Import from Git Repository

```
1. Parse source → Extract Git URL and ref
2. Workspace cache → Get or clone repository
3. Discovery → Find all resources in clone
4. Repository → Add resources to central repo
5. Metadata → Track added resources
```

**Packages**: `source` → `workspace` → `discovery` → `repo` → `metadata`

### Flow 2: Install Resource to Tool

```
1. Pattern matching → Resolve pattern to resources
2. Repository → Find resource in central repo
3. Tool detection → Get tool directories (.claude/)
4. Install → Create symlink to target directory
```

**Packages**: `pattern` → `repo` → `tools` → `install`

### Flow 3: Sync Resources

```
1. Metadata → Load all resources with Git sources
2. Workspace → Update cached repositories (git pull)
3. Discovery → Find updated resources
4. Repository → Replace resources in central repo
```

**Packages**: `metadata` → `workspace` → `discovery` → `repo`

## Directory Layout

### User Data

```
~/.local/share/ai-config/
├── repo/
│   ├── commands/
│   ├── skills/
│   ├── agents/
│   ├── packages/
│   ├── .metadata/           # Resource tracking
│   └── .workspace/          # Git cache
└── config.yaml              # Optional
```

### Project Installation

```
project-root/
├── .claude/
│   ├── commands/            # Claude Code commands
│   ├── skills/              # Claude Code skills
│   └── agents/              # Claude Code agents
├── .opencode/               # OpenCode resources
├── .github/skills/          # GitHub Copilot skills
├── .windsurf/skills/        # Windsurf skills
└── ai.package.yaml          # Project manifest (optional)
```

## Related Documentation

- **[Testing Guide](testing.md)** - Comprehensive testing documentation
- **[Development Environment](development-environment.md)** - Setup and tools
- **[Release Process](release-process.md)** - Release workflow
- **[CONTRIBUTING.md](../../CONTRIBUTING.md)** - Complete development guide
