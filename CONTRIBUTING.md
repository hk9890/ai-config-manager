# Contributing to ai-repo

Thank you for your interest in contributing to ai-repo! This document provides guidelines and information for developers.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)

## Development Setup

### Prerequisites

- **Go 1.25 or higher** (check with `go version`)
- **Make** (optional, but recommended)
- **Git**

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager

# Build the binary
make build

# Or build and install to ~/bin
make install

# Run all checks (format, vet, test, build)
make all
```

### Quick Commands

```bash
# Build
make build              # Build binary to ./ai-repo

# Testing
make test              # Run all tests (unit + integration + vet)
make unit-test         # Run only unit tests
make integration-test  # Run only integration tests

# Code Quality
make fmt               # Format all Go code
make vet               # Run go vet
make deps              # Download dependencies

# Cleanup
make clean             # Remove build artifacts
```

## Project Architecture

### Overview

ai-repo is a CLI tool built with Go that manages AI resources (commands and skills) across multiple AI coding tools. It uses a centralized repository with symlink-based installation.

**Key Concepts:**
- **Resources**: Commands (single .md files) and Skills (directories with SKILL.md)
- **Repository**: Centralized storage at `~/.local/share/ai-config/repo/`
- **Installation**: Creates symlinks in tool-specific directories (.claude/, .opencode/, .github/skills/)
- **Multi-Tool**: Intelligently detects and installs to all present tool directories

### Project Structure

```
ai-config-manager/
├── cmd/                    # CLI command definitions (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── add.go             # Add command parent
│   ├── add_command.go     # Add single command
│   ├── add_skill.go       # Add single skill
│   ├── add_plugin.go      # Add from plugin (bulk)
│   ├── add_claude.go      # Add from .claude/ (bulk)
│   ├── config.go          # Configuration management
│   ├── install.go         # Install resources to projects
│   ├── list.go            # List resources
│   └── remove.go          # Remove resources
│
├── pkg/                    # Core packages
│   ├── config/            # Configuration management
│   │   └── config.go      # Load/save ~/.ai-repo.yaml
│   │
│   ├── repo/              # Repository management
│   │   └── manager.go     # Add/remove/list resources, bulk import
│   │
│   ├── resource/          # Resource types and validation
│   │   ├── resource.go    # Base resource type
│   │   ├── command.go     # Command resource logic
│   │   ├── skill.go       # Skill resource logic
│   │   ├── plugin.go      # Plugin and .claude/ detection
│   │   ├── frontmatter.go # YAML frontmatter parsing
│   │   └── validation.go  # Name and format validation
│   │
│   ├── install/           # Installation logic
│   │   └── installer.go   # Symlink creation, tool detection
│   │
│   ├── tools/             # Tool-specific information
│   │   └── tools.go       # Claude/OpenCode/Copilot definitions
│   │
│   └── version/           # Version information
│       └── version.go     # Embedded at build time
│
├── test/                   # Integration tests
│   ├── integration_test.go       # End-to-end workflow tests
│   └── bulk_import_test.go       # Bulk import integration tests
│
├── examples/               # Example resources
│   ├── sample-command.md
│   ├── sample-skill/
│   └── README.md
│
├── main.go                 # Entry point
├── Makefile                # Build automation
├── go.mod                  # Go module definition
├── README.md               # User documentation
├── AGENTS.md               # AI agent quick reference
└── CONTRIBUTING.md         # This file
```

### Architecture Flow

1. **User runs CLI command** → `cmd/` (Cobra)
2. **Command validates input** → `pkg/resource/` (validation)
3. **Manager handles operation** → `pkg/repo/` (add/remove/list)
4. **Resources stored** → `~/.local/share/ai-config/repo/`
5. **Installation creates symlinks** → `pkg/install/` → `.claude/`, `.opencode/`, etc.

### Key Design Patterns

**Resource Abstraction:**
- `Resource` struct represents both commands and skills
- Type-specific logic in `command.go` and `skill.go`
- Common operations in `resource.go`

**Repository Pattern:**
- `Manager` struct encapsulates all repo operations
- Centralized storage with consistent structure
- Copy-on-add to ensure immutability

**Tool Detection:**
- Smart detection of existing tool directories
- Multi-tool installation support
- Respects user's default-tool configuration

**Symlink-based Installation:**
- No duplication of resources
- Single source of truth in repository
- Easy updates (modify in repo, all projects updated)

## Development Workflow

### Adding a New Feature

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Write tests first** (TDD approach recommended)
   - Unit tests in `pkg/*/`
   - Integration tests in `test/`

3. **Implement the feature**
   - Follow code style guidelines
   - Add inline documentation
   - Handle errors properly

4. **Test thoroughly**
   ```bash
   make test          # All tests must pass
   make vet           # No linter warnings
   ```

5. **Update documentation**
   - User-facing: Update README.md
   - Developer-facing: Update CONTRIBUTING.md or AGENTS.md
   - Add examples if applicable

6. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   git push origin feature/your-feature-name
   ```

7. **Create Pull Request**
   - Describe what and why
   - Link related issues
   - Ensure CI passes

### Testing Guidelines

**Unit Tests:**
- Located in same package as code (`*_test.go`)
- Test individual functions/methods
- Use table-driven tests
- Mock external dependencies
- Use `t.TempDir()` for filesystem operations

**Integration Tests:**
- Located in `test/` directory
- Test end-to-end workflows
- Use real files and directories (with temp dirs)
- Verify actual behavior of CLI commands

**Test Coverage:**
- Aim for >80% coverage on new code
- Both success and error paths must be tested
- Edge cases and invalid inputs tested

**Example Test Structure:**
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        want      string
        wantError bool
    }{
        {name: "valid input", input: "test", want: "result", wantError: false},
        {name: "invalid input", input: "", want: "", wantError: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantError {
                t.Errorf("Feature() error = %v, wantError %v", err, tt.wantError)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Code Style Guidelines

### General Principles

- **Clarity over cleverness**: Write obvious code
- **Single Responsibility**: Functions do one thing well
- **Error handling**: Always handle errors, provide context
- **Documentation**: Export functions/types must have GoDoc comments
- **Testing**: All new code must have tests

### File Naming

- `lowercase_with_underscores.go` for regular files
- `*_test.go` for test files
- Descriptive names (e.g., `command_validation.go`, not `util.go`)

### Package Naming

- Short, lowercase, single word
- Examples: `resource`, `config`, `install`, `repo`
- No underscores or mixed caps

### Type and Function Naming

- **Exported**: PascalCase (e.g., `ResourceType`, `LoadCommand`)
- **Unexported**: camelCase (e.g., `resourcePath`, `loadConfig`)
- **Constants**: PascalCase for exported, camelCase for unexported

### Import Organization

Group imports in three sections with blank lines:

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

### Error Handling

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}
```

- Use `%w` to wrap errors (preserves error chain)
- Provide descriptive context
- Don't panic (except in main/init for fatal errors)
- Check errors immediately

### Comments and Documentation

```go
// LoadCommand loads a command resource from a markdown file.
// It validates the file format and parses the YAML frontmatter.
// Returns an error if the file is not a valid command resource.
func LoadCommand(filePath string) (*Resource, error) {
    // Implementation
}
```

- All exported items must have GoDoc comments
- Start with the item name
- Describe what, not how
- Keep comments up-to-date with code

### File Operations

```go
// Good: Use filepath.Join for cross-platform paths
path := filepath.Join(dir, "commands", "test.md")

// Good: Check file existence
if _, err := os.Stat(path); err != nil {
    return fmt.Errorf("file does not exist: %w", err)
}

// Good: Use defer for cleanup
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()

// Good: Set appropriate permissions
os.MkdirAll(dir, 0755)        // Directories
os.WriteFile(path, data, 0644) // Files
```

### Resource Name Validation

Resources must follow agentskills.io naming:
- Lowercase alphanumeric + hyphens only
- Cannot start/end with hyphen
- No consecutive hyphens
- 1-64 characters max

```go
// Valid
"test", "run-coverage", "pdf-processing", "skill-v2"

// Invalid
"Test", "test_coverage", "-test", "test--cmd"
```

## Testing

### Running Tests

```bash
# All tests
make test

# Unit tests only
make unit-test

# Integration tests only
make integration-test

# Single test file
go test -v ./pkg/resource/command_test.go

# Specific test function
go test -v ./pkg/config -run TestLoad_ValidConfig

# With coverage
go test -v -cover ./pkg/...

# Specific package
go test -v ./pkg/resource/
```

### Writing Tests

Use table-driven tests:

```go
func TestValidateName(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantError bool
    }{
        {name: "valid simple", input: "test", wantError: false},
        {name: "valid with hyphen", input: "test-command", wantError: false},
        {name: "invalid uppercase", input: "Test", wantError: true},
        {name: "invalid underscore", input: "test_cmd", wantError: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateName(tt.input)
            if (err != nil) != tt.wantError {
                t.Errorf("ValidateName(%q) error = %v, wantError %v", 
                    tt.input, err, tt.wantError)
            }
        })
    }
}
```

Use `t.TempDir()` for filesystem tests:

```go
func TestAddCommand(t *testing.T) {
    tmpDir := t.TempDir() // Auto-cleanup
    repoPath := filepath.Join(tmpDir, "repo")
    manager := repo.NewManagerWithPath(repoPath)
    
    // Test with temp directory
}
```

## Submitting Changes

### Before Submitting

- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make vet`)
- [ ] New code has tests
- [ ] Documentation updated (if user-facing change)
- [ ] Commit messages are clear and descriptive

### Commit Message Format

Use conventional commits:

```
type(scope): short description

Longer description if needed, explaining what and why.

Fixes #issue-number
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance, dependencies, etc.

**Examples:**
```
feat(repo): add bulk import support for plugins

Add ability to import multiple commands and skills from Claude plugins
in a single operation.

Fixes #42
```

```
fix(install): handle symlink creation on Windows

Use junction points instead of symlinks for Windows compatibility.
```

### Pull Request Process

1. Create PR from your feature branch
2. Fill out PR template (if available)
3. Link related issues
4. Wait for CI to pass
5. Address review feedback
6. Maintainer will merge when approved

## Key Dependencies

- **[Cobra](https://github.com/spf13/cobra)**: CLI framework
- **[yaml.v3](https://gopkg.in/yaml.v3)**: YAML parsing for frontmatter
- **[XDG](https://github.com/adrg/xdg)**: XDG Base Directory support
- **[tablewriter](https://github.com/olekukonko/tablewriter)**: Table output formatting

## Questions or Need Help?

- **Issues**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)
- **Documentation**: See [README.md](README.md) for user docs, [AGENTS.md](AGENTS.md) for AI agent quick reference

Thank you for contributing to ai-repo! 🎉
