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

## Repository Structure

```
ai-config-manager/
├── cmd/              # Cobra command definitions (CLI entry points)
├── pkg/              # Core business logic packages
│   ├── config/       # Configuration management
│   ├── discovery/    # Auto-discovery of resources
│   ├── install/      # Installation/symlink logic
│   ├── repo/         # Repository management
│   ├── resource/     # Resource types (command, skill, agent, package)
│   ├── workspace/    # Git repository caching (10-50x faster)
│   └── ...
├── test/             # Integration and E2E tests
├── docs/             # Documentation
└── main.go           # Entry point
```

## Code Style Essentials

**Import Organization** - Three groups with blank lines:
```go
import (
    "fmt"           // 1. Standard library

    "github.com/spf13/cobra"  // 2. External dependencies

    "github.com/hk9890/ai-config-manager/pkg/resource"  // 3. Internal
)
```

**Naming**:
- Files: `lowercase_with_underscores.go`
- Types: `PascalCase`
- Functions: `PascalCase` (exported), `camelCase` (unexported)
- Resources: `lowercase-with-hyphens` (1-64 chars, no start/end hyphens)

**Error Handling** - Always wrap:
```go
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}
```

**Symlink Handling** - CRITICAL for COPY and SYMLINK modes:
```go
// ❌ WRONG: Skips symlinked directories
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    if entry.IsDir() { ... }  // Returns false for symlinks!
}

// ✅ CORRECT: Follows symlinks
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    info, err := os.Stat(path)  // os.Stat follows symlinks
    if err != nil || !info.IsDir() { continue }
    // Process directory...
}
```

## Critical Patterns

**Git Operations** - Always use workspace cache:
```go
import "github.com/hk9890/ai-config-manager/pkg/workspace"

mgr, _ := workspace.NewManager(repoPath)
clonePath, err := mgr.GetOrClone(gitURL, ref)  // Cached, 10-50x faster
// Use clonePath - no cleanup needed
```

**Loading Resources**:
```go
// Commands (auto-detects base path)
res, err := resource.LoadCommand("path/to/commands/test.md")

// Skills (directory with SKILL.md)
res, err := resource.LoadSkill("path/to/skill-dir")

// Agents (single .md file)
res, err := resource.LoadAgent("path/to/agent.md")
```

**Repository Operations**:
```go
mgr, err := repo.NewManager()
mgr.AddCommand(sourcePath)
resources, err := mgr.List(nil)  // nil = all types
```

## Testing

Use isolated temporary directories:
```go
func TestFeature(t *testing.T) {
    tmpDir := t.TempDir()  // Auto-cleanup
    manager := repo.NewManagerWithPath(tmpDir)  // NOT NewManager()
    // ... test operations ...
}
```

**Test Types**:
- Unit tests: Use fixtures in `testdata/`, no network, `//go:build unit`
- Integration tests: Real repos, network calls, `//go:build integration`

## Documentation

- **User Guide**: `docs/user-guide/` - Resource formats, patterns, output
- **Contributor Guide**: `docs/contributor-guide/` - Architecture, testing, development
- **README.md**: User-facing installation and usage
- **CONTRIBUTING.md**: Complete development guide

## Before Committing

1. `make fmt` - Format code
2. `make test` - All tests pass
3. Add tests for new functionality
4. Follow existing patterns
5. Git operations use `pkg/workspace`
