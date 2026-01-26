# AGENTS.md

This document provides essential guidelines for AI coding agents working in the ai-config-manager repository.

## Table of Contents
- [Project Overview](#project-overview)
- [Build & Test Commands](#build--test-commands)
- [Code Style Guidelines](#code-style-guidelines)
- [Common Patterns](#common-patterns)
- [Resource Formats Quick Reference](#resource-formats-quick-reference)
- [Version & Dependencies](#version--dependencies)
- [Detailed Documentation](#detailed-documentation)

---

## Project Overview

**aimgr** is a CLI tool for managing AI resources (commands, skills, and agents) across multiple AI coding tools (Claude Code, OpenCode, GitHub Copilot). It uses a centralized repository with symlink-based installation.

- **Language**: Go 1.25.6
- **Architecture**: CLI built with Cobra, resource management with symlinks
- **Storage**: `~/.local/share/ai-config/repo/` (XDG data directory)
- **Supported Resources**: Commands, Skills, Agents, Packages

### Package Structure
```
ai-config-manager/
├── cmd/              # Cobra command definitions
├── pkg/
│   ├── config/       # Configuration management
│   ├── discovery/    # Auto-discovery of resources
│   ├── install/      # Installation/symlink logic
│   ├── manifest/     # ai.package.yaml handling
│   ├── marketplace/  # Marketplace parsing and generation
│   ├── metadata/     # Metadata tracking and migration
│   ├── pattern/      # Pattern matching for resources
│   ├── repo/         # Repository management
│   ├── resource/     # Resource types (command, skill, agent, package)
│   ├── source/       # Source parsing and Git operations
│   ├── tools/        # Tool-specific info (Claude, OpenCode, Copilot)
│   ├── version/      # Version information
│   └── workspace/    # Workspace caching for Git repos
├── test/             # Integration tests
├── examples/         # Example resources
└── main.go           # Entry point
```

---

## Build & Test Commands

### Building
```bash
make build      # Build binary
make install    # Build and install to ~/bin
make all        # Run all checks and build
```

### Testing
```bash
make test       # Run all tests (unit + integration + vet)
make unit-test  # Run only unit tests
make integration-test  # Run only integration tests

# Run specific tests
go test -v ./pkg/resource/command_test.go
go test -v ./pkg/config -run TestLoad_ValidConfig
go test -v -cover ./pkg/...
```

### Linting & Formatting
```bash
make fmt        # Format all Go code
make vet        # Run go vet
make deps       # Download dependencies
make clean      # Clean build artifacts
```

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

### Testing
- Use table-driven tests
- Use `t.TempDir()` for temporary directories
- Test both success and error cases

### File Operations
- Use `filepath.Join()` for path construction
- Check existence with `os.Stat(path)`
- Use `defer file.Close()` for cleanup
- Permissions: `0755` for dirs, `0644` for files

---

## Common Patterns

### Loading Resources
```go
// Commands: single .md file
res, err := resource.LoadCommand("path/to/command.md")

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

See [docs/pattern-matching.md](docs/pattern-matching.md) for detailed examples.

### Output Formats

Commands support multiple output formats via `--format` flag:

**Table (default)**:
```bash
aimgr repo add ~/resources/
# Shows human-readable table with results
```

**JSON (for scripting)**:
```bash
aimgr repo add ~/resources/ --format=json
# Returns structured JSON output
```

**YAML (structured, human-readable)**:
```bash
aimgr repo add ~/resources/ --format=yaml
# Returns YAML output
```

**Commands supporting --format:**
- `repo add`, `repo sync`, `repo list`, `repo update`
- `list`, `install`, `uninstall`

See [docs/output-formats.md](docs/output-formats.md) for comprehensive documentation.

### Workspace Caching

Git repositories are cached in `.workspace/` for efficient reuse:
- **First operation**: Full clone (creates cache)
- **Subsequent operations**: Reuse cache (10-50x faster)
- **Commands**: `repo add`, `repo sync`, `repo update`
- **Cache management**: `aimgr repo prune`

See [docs/workspace-caching.md](docs/workspace-caching.md) for details.

---

## Resource Formats Quick Reference

### Package Format

**File**: `packages/<name>.package.json`

```json
{
  "name": "package-name",
  "description": "Package description",
  "resources": [
    "command/build",
    "skill/typescript-helper",
    "agent/code-reviewer"
  ]
}
```

**Code**:
```go
type Package struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Resources   []string `json:"resources"`
}
```

### Agent Format

**File**: `agents/<name>.md` with YAML frontmatter

```yaml
---
description: Agent description (required)
type: code-reviewer (optional, OpenCode)
capabilities: [static-analysis, security] (optional, OpenCode)
version: "1.0.0" (optional)
---

# Agent Name

Agent instructions and documentation here.
```

**Code**:
```go
type AgentResource struct {
    Resource
    Type         string   `yaml:"type,omitempty"`
    Instructions string   `yaml:"instructions,omitempty"`
    Capabilities []string `yaml:"capabilities,omitempty"`
    Content      string   `yaml:"-"`
}
```

### Marketplace Format

**File**: `marketplace.json`

```json
{
  "name": "marketplace-name",
  "plugins": [
    {
      "name": "plugin-name",
      "description": "Plugin description",
      "source": "path/to/plugin"
    }
  ]
}
```

**Import**:
```bash
aimgr marketplace import marketplace.json
```

### Project Manifest

**File**: `ai.package.yaml` (project root)

```yaml
resources:
  - skill/pdf-processing
  - command/test
  - agent/code-reviewer
  - package/web-tools

targets:  # Optional
  - claude
  - opencode
```

**Usage**:
```bash
aimgr install                    # Install all from manifest
aimgr install skill/test         # Install and add to manifest
aimgr install skill/test --no-save  # Install without adding
```

### Tool Directories

| Tool | Commands | Skills | Agents |
|------|----------|--------|--------|
| Claude Code | `.claude/commands/` | `.claude/skills/` | `.claude/agents/` |
| OpenCode | `.opencode/commands/` | `.opencode/skills/` | `.opencode/agents/` |
| GitHub Copilot | N/A | `.github/skills/` | N/A |

### Bulk Import

Import from directories with auto-discovery:
```bash
aimgr repo add ~/.opencode
aimgr repo add ~/.claude
aimgr repo add gh:owner/repo --filter "skill/*"
```

See [docs/resource-formats.md](docs/resource-formats.md) for complete specifications.

---

## Version & Dependencies

### Version Information
Version embedded at build time via ldflags in Makefile:
- `Version`, `GitCommit`, `BuildDate` in `pkg/version/version.go`

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/adrg/xdg` - XDG base directory support
- `github.com/olekukonko/tablewriter` - Table output formatting

### Testing Philosophy
- Unit tests in `pkg/*/` packages
- Integration tests in `test/`
- Use testdata directories for fixtures
- Mock filesystem operations (use temp dirs)

---

## Detailed Documentation

For comprehensive information, see:

- **[docs/resource-formats.md](docs/resource-formats.md)** - Complete format specifications for all resource types (packages, agents, marketplaces, manifests)
- **[docs/workspace-caching.md](docs/workspace-caching.md)** - Git repository caching, performance optimization, cache management
- **[docs/pattern-matching.md](docs/pattern-matching.md)** - Pattern matching syntax, examples, and use cases
- **[docs/output-formats.md](docs/output-formats.md)** - CLI output formats (JSON, YAML, table) with scripting examples

---

## When Making Changes

1. Run `make fmt` before committing
2. Run `make test` to verify all tests pass
3. Add tests for new functionality
4. Update documentation if adding user-facing features
5. Follow existing code patterns and conventions
