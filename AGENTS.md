# AGENTS.md

**This document provides essential guidelines for AI coding agents working in the ai-config-manager repository.**

---

## Project Overview

**aimgr** is a CLI tool for managing AI resources (commands, skills, and agents) across multiple AI coding tools (Claude Code, OpenCode, GitHub Copilot). It uses a centralized repository with symlink-based installation.

**Key Concepts:**
- **Language**: Go 1.25.6
- **Architecture**: CLI built with Cobra, resource management with symlinks  
- **Storage**: `~/.local/share/ai-config/repo/` (XDG data directory)
- **Supported Resources**: Commands, Skills, Agents, Packages

The tool discovers resources from various sources (local directories, Git repositories, GitHub), stores them in a central repository, and installs them via symlinks to tool-specific directories (`.claude/`, `.opencode/`, etc.).

---

## Repository Structure

```
ai-config-manager/
├── cmd/              # Cobra command definitions (CLI entry points)
├── pkg/              # Core business logic (see breakdown below)
├── test/             # Integration tests
├── docs/             # Documentation (user-guide/, contributor-guide/, architecture/)
├── examples/         # Example resources
└── main.go           # Entry point
```

### Key Packages (pkg/)

| Package | Purpose |
|---------|---------|
| `config/` | Configuration management |
| `discovery/` | Auto-discovery of resources from directories/repos |
| `install/` | Installation/symlink logic |
| `manifest/` | ai.package.yaml handling |
| `repo/` | Repository management (add, list, remove) |
| `resource/` | Resource types (command, skill, agent, package) |
| `source/` | Source parsing and Git operations |
| `tools/` | Tool-specific info (Claude, OpenCode, Copilot) |
| `workspace/` | Workspace caching for Git repos (10-50x faster) |

**For detailed architecture:** See `docs/architecture/architecture-rules.md`

---

## Quick Reference

### Build Commands
```bash
make build      # Build binary
make install    # Build and install to ~/bin
make test       # Run all tests (unit + integration + vet)
make fmt        # Format all Go code
make vet        # Run go vet
```

### Run Tests
```bash
make test                # All tests
make unit-test           # Fast unit tests only (<5 seconds)
make integration-test    # Slow integration tests (~30 seconds)

# Run specific tests
go test -v ./pkg/resource/command_test.go
go test -v ./pkg/config -run TestLoad_ValidConfig
```

### Documentation Locations

- **User Guide**: `docs/user-guide/` - Resource formats, pattern matching, output formats
- **Architecture**: `docs/architecture/architecture-rules.md` - Strict architectural rules
- **Contributor Guide**: `docs/contributor-guide/release-process.md`
- **Planning/Archive**: `docs/planning/`, `docs/archive/`

---

## Code Style Guidelines

### Import Organization
Three groups with blank lines:
```go
import (
	"fmt"           // 1. Standard library
	"os"

	"github.com/spf13/cobra"  // 2. External dependencies

	"github.com/hk9890/ai-config-manager/pkg/resource"  // 3. Internal
)
```

### Naming Conventions
- **Files**: `lowercase_with_underscores.go`
- **Packages**: Short, lowercase, single word (`resource`, `config`)
- **Types**: PascalCase (`ResourceType`, `CommandResource`)
- **Functions**: PascalCase (exported), camelCase (unexported)
- **Variables**: camelCase (`repoPath`, `skillsDir`)

### Resource Names (User-Facing)
Resources must follow agentskills.io naming:
- Lowercase alphanumeric + hyphens only
- Cannot start/end with hyphen
- No consecutive hyphens (`--`)
- 1-64 characters max
- Examples: `test-command`, `pdf-processing`, `code-reviewer`

### Error Handling
Always wrap errors with context:
```go
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}
```

### File Operations
- Use `filepath.Join()` for path construction
- Check existence with `os.Stat(path)`
- Use `defer file.Close()` for cleanup
- Permissions: `0755` for dirs, `0644` for files

---

## Common Patterns

### Loading Resources

#### Commands
Commands are loaded using `LoadCommand(filePath)` which automatically detects
the base path by finding the nearest `commands/` directory in the path.

```go
// LoadCommand auto-detects base path and preserves nested structure
res, err := resource.LoadCommand("path/to/commands/test.md")
// → name = "test"

res, err := resource.LoadCommand("path/to/commands/api/deploy.md")
// → name = "api/deploy"

// Commands MUST be in a directory named `commands/`
// (or `.claude/commands/`, `.opencode/commands/`, etc.)
```

#### Skills, Agents, and Packages
```go
// Skills: directory with SKILL.md
res, err := resource.LoadSkill("path/to/skill-dir")

// Agents: single .md file
res, err := resource.LoadAgent("path/to/agent.md")

// Packages: .package.json file
pkg, err := resource.LoadPackage("path/to/package.package.json")
```

### Repository Operations
```go
mgr, err := repo.NewManager()
mgr.AddCommand(sourcePath)
mgr.AddSkill(sourcePath)
mgr.AddAgent(sourcePath)
resources, err := mgr.List(nil)  // nil = all types

// Bulk operations
opts := repo.BulkImportOptions{
    Force:        false,
    SkipExisting: false,
    DryRun:       false,
}
result, err := mgr.AddBulk(paths, opts)
```

### Tool Detection
```go
tool, err := tools.ParseTool("claude")  // Returns tools.Claude
tool, err := tools.ParseTool("opencode") // Returns tools.OpenCode
info := tools.GetToolInfo(tool)         // Get dirs, etc.
```

### Pattern Matching
```go
import "github.com/hk9890/ai-config-manager/pkg/pattern"

// Parse pattern
resourceType, patternStr, isPattern := pattern.ParsePattern("skill/pdf*")

// Create matcher
matcher, err := pattern.NewMatcher("skill/pdf*")
if matcher.Match(res) {
    fmt.Println("Matched!")
}
```

See [docs/user-guide/pattern-matching.md](docs/user-guide/pattern-matching.md) for detailed examples.

### Workspace Caching (Critical)

**All Git operations MUST use `pkg/workspace` cache** (see Architecture Rule 1):

```go
import "github.com/hk9890/ai-config-manager/pkg/workspace"

// Get repository path (clone if needed, reuse if cached)
mgr, err := workspace.NewManager(repoPath)
if err != nil {
    return fmt.Errorf("failed to create workspace manager: %w", err)
}

// GetOrClone returns cached path or clones if needed
clonePath, err := mgr.GetOrClone(gitURL, ref)
if err != nil {
    return fmt.Errorf("failed to get repository: %w", err)
}

// Use clonePath to access repository contents
// No cleanup needed - cache is managed automatically
```

**Performance:** First clone is slow, subsequent operations are 10-50x faster.  
**Commands:** `repo import`, `repo sync` use this automatically.  
**See:** `docs/user-guide/workspace-caching.md`

---

## Testing

Tests are split into fast unit tests (default) and slow integration tests (opt-in).

### Unit Tests (fixtures)
- Use committed fixtures in `testdata/repos/`
- No network calls
- Run by default: `make test`
- Execution time: <5 seconds
- Build tag: `//go:build unit`

### Integration Tests (network)
- Use real GitHub repos (hk9890/ai-tools)
- Run with: `make integration-test`
- Execution time: ~30 seconds
- Build tag: `//go:build integration`

### Best Practices
- Use table-driven tests
- Use `t.TempDir()` for temporary directories
- Test both success and error cases
- Prefer unit tests with fixtures
- Only add integration tests for new Git features

**See:** `docs/planning/test-refactoring.md`

---

## Output Formats

Commands support multiple output formats via `--format` flag:

```bash
aimgr repo import ~/resources/               # Table (default, human-readable)
aimgr repo import ~/resources/ --format=json # JSON (for scripting)
aimgr repo import ~/resources/ --format=yaml # YAML (structured)
```

**Commands supporting --format:**
- `repo import`, `repo sync`, `repo list`
- `list`, `install`, `uninstall`

**See:** [docs/user-guide/output-formats.md](docs/user-guide/output-formats.md)

---

## When Making Changes

1. Run `make fmt` before committing
2. Run `make test` to verify all tests pass
3. Add tests for new functionality
4. Update documentation if adding user-facing features
5. Follow existing code patterns and conventions
6. Ensure Git operations use `pkg/workspace` (Architecture Rule 1)

---

## Detailed Documentation

For comprehensive information, see:

- **[docs/architecture/architecture-rules.md](docs/architecture/architecture-rules.md)** - Strict architectural rules (Git workspace, XDG, error wrapping, build tags)
- **[docs/user-guide/resource-formats.md](docs/user-guide/resource-formats.md)** - Complete format specifications for all resource types
- **[docs/user-guide/workspace-caching.md](docs/user-guide/workspace-caching.md)** - Git repository caching, performance optimization
- **[docs/user-guide/pattern-matching.md](docs/user-guide/pattern-matching.md)** - Pattern matching syntax, examples
- **[docs/user-guide/output-formats.md](docs/user-guide/output-formats.md)** - CLI output formats with scripting examples
- **[docs/user-guide/github-sources.md](docs/user-guide/github-sources.md)** - Adding resources from GitHub repositories
- **[docs/contributor-guide/release-process.md](docs/contributor-guide/release-process.md)** - Release workflow and GoReleaser configuration
