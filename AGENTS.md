# AGENTS.md

This document provides guidelines for AI coding agents working in the ai-config-manager repository.

## Project Overview

**ai-repo** is a CLI tool for managing AI resources (commands and skills) across multiple AI coding tools (Claude Code, OpenCode, GitHub Copilot). It uses a centralized repository with symlink-based installation.

- **Language**: Go 1.25.6
- **Architecture**: CLI built with Cobra, resource management with symlinks
- **Storage**: `~/.local/share/ai-config/repo/` (XDG data directory)

## Build & Test Commands

### Building
```bash
# Build binary
make build

# Build and install to ~/bin
make install

# Run all checks and build
make all
```

### Testing
```bash
# Run all tests (unit + integration + vet)
make test

# Run only unit tests
make unit-test

# Run only integration tests
make integration-test

# Run a single test file
go test -v ./pkg/resource/command_test.go

# Run a specific test function
go test -v ./pkg/config -run TestLoad_ValidConfig

# Run tests with coverage
go test -v -cover ./pkg/...

# Run tests in a specific package
go test -v ./pkg/resource/
```

### Linting & Formatting
```bash
# Format all Go code
make fmt
# Or: go fmt ./...

# Run go vet
make vet
# Or: go vet ./...

# Download dependencies
make deps
```

### Cleaning
```bash
# Clean build artifacts
make clean
```

## Code Style Guidelines

### Package Structure
```
ai-config-manager/
├── cmd/              # Cobra command definitions
├── pkg/
│   ├── config/       # Configuration management
│   ├── install/      # Installation/symlink logic
│   ├── repo/         # Repository management
│   ├── resource/     # Resource types (command, skill)
│   ├── tools/        # Tool-specific info (Claude, OpenCode, Copilot)
│   └── version/      # Version information
├── test/             # Integration tests
└── main.go           # Entry point
```

### Import Organization
Organize imports in three groups with blank lines:
1. Standard library
2. External dependencies
3. Internal packages

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/hans-m-leitner/ai-config-manager/pkg/resource"
)
```

### Naming Conventions
- **Files**: `lowercase_with_underscores.go` (e.g., `manager_test.go`)
- **Packages**: Short, lowercase, single word (e.g., `resource`, `config`)
- **Types**: PascalCase (e.g., `ResourceType`, `CommandResource`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `repoPath`, `skillsDir`)

### Resource Names (User-Facing)
Resources (commands/skills) must follow agentskills.io naming:
- Lowercase alphanumeric + hyphens only
- Cannot start/end with hyphen
- No consecutive hyphens (`--`)
- 1-64 characters max
- Examples: `test-command`, `pdf-processing`, `skill-v2`

### Types
- Use explicit types, avoid `interface{}` when possible
- Define custom types for domain concepts:
  ```go
  type ResourceType string
  
  const (
      Command ResourceType = "command"
      Skill   ResourceType = "skill"
  )
  ```

### Error Handling
- Always wrap errors with context using `fmt.Errorf` with `%w`:
  ```go
  if err != nil {
      return fmt.Errorf("failed to load command: %w", err)
  }
  ```
- Use descriptive error messages that include context
- Return errors, don't panic (except in main/init for fatal setup)
- Check errors immediately, don't defer error handling

### Functions
- Keep functions focused (single responsibility)
- Document exported functions with GoDoc comments:
  ```go
  // LoadCommand loads a command resource from a markdown file
  func LoadCommand(filePath string) (*Resource, error) {
  ```
- Return early for error cases (guard clauses)
- Use named return values sparingly (only for complex functions)

### Testing
- Test files: `*_test.go` in same package
- Table-driven tests preferred:
  ```go
  tests := []struct {
      name      string
      input     string
      wantError bool
  }{
      {name: "valid input", input: "test", wantError: false},
  }
  
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
          // test logic
      })
  }
  ```
- Use `t.TempDir()` for temporary test directories (auto-cleanup)
- Test both success and error cases

### File Operations
- Always use `filepath.Join()` for path construction (cross-platform)
- Check file existence before operations: `os.Stat(path)`
- Use defer for cleanup: `defer file.Close()`
- Set appropriate permissions: `0755` for dirs, `0644` for files

### YAML Frontmatter
- Commands and skills use YAML frontmatter in markdown files
- Format: delimited by `---` at start and end
- Required fields vary by resource type (see `pkg/resource/`)

### Comments
- Document all exported types, functions, constants
- Use `//` for single-line comments
- Keep comments concise and up-to-date with code

## Common Patterns

### Loading Resources
```go
// Commands: single .md file
res, err := resource.LoadCommand("path/to/command.md")

// Skills: directory with SKILL.md
res, err := resource.LoadSkill("path/to/skill-dir")
```

### Repository Operations
```go
mgr, err := repo.NewManager()
mgr.AddCommand(sourcePath)
mgr.AddSkill(sourcePath)
resources, err := mgr.List(nil)  // nil = all types
```

### Tool Detection
```go
tool, err := tools.ParseTool("claude")  // Returns tools.Claude
info := tools.GetToolInfo(tool)         // Get dirs, etc.
```

## Version Information
Version is embedded at build time via ldflags in Makefile:
- `Version`, `GitCommit`, `BuildDate` in `pkg/version/version.go`

## Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/adrg/xdg` - XDG base directory support
- `github.com/olekukonko/tablewriter` - Table output formatting

## Testing Philosophy
- Unit tests in `pkg/*/` packages
- Integration tests in `test/`
- Use testdata directories for test fixtures
- Mock filesystem operations where appropriate (use temp dirs)

## When Making Changes
1. Run `make fmt` before committing
2. Run `make test` to verify all tests pass
3. Add tests for new functionality
4. Update documentation if adding user-facing features
5. Follow existing code patterns and conventions
