# Contributing to aimgr

Thank you for your interest in contributing to aimgr! This document provides guidelines and information for developers.

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
make build              # Build binary to ./aimgr

# Testing
make test              # Run all tests (unit + integration + vet)
make unit-test         # Run only unit tests
make integration-test  # Run only integration tests
make e2e-test          # Run only E2E tests

# Code Quality
make fmt               # Format all Go code
make vet               # Run go vet
make deps              # Download dependencies

# Cleanup
make clean             # Remove build artifacts
```

## Project Architecture

### Overview

aimgr is a CLI tool built with Go that manages AI resources (commands and skills) across multiple AI coding tools. It uses a centralized repository with symlink-based installation.

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
│   ├── add.go             # Add command (auto-discovery from sources)
│   ├── config.go          # Configuration management
│   ├── init.go            # Initialize ai.package.yaml
│   ├── install.go         # Install resources/packages
│   ├── list.go            # List repository resources
│   ├── list_installed.go  # List installed resources in project
│   ├── remove.go          # Remove resources
│   ├── uninstall.go       # Uninstall from project
│   ├── repo.go            # Repo subcommand group
│   ├── repo_create_package.go  # Create packages
│   ├── repo_show.go       # Show resource details
│   ├── repo_sync.go       # Sync from configured sources
│   ├── repo_update.go     # Update resources from sources
│   ├── repo_prune.go      # Prune unused caches
│   ├── repo_verify.go     # Verify repository health
│   ├── completion.go      # Shell completion
│   └── *_test.go          # Test files
│
├── pkg/                    # Core packages
│   ├── config/            # Configuration management
│   │   └── config.go      # Load/save ~/.config/aimgr/aimgr.yaml
│   │
│   ├── discovery/         # Auto-discovery of resources
│   │   ├── commands.go    # Command auto-discovery
│   │   ├── skills.go      # Skill auto-discovery
│   │   ├── agents.go      # Agent auto-discovery
│   │   ├── packages.go    # Package auto-discovery
│   │   └── testdata/      # Test fixtures
│   │
│   ├── install/           # Installation logic
│   │   └── installer.go   # Symlink creation, tool detection
│   │
│   ├── manifest/          # ai.package.yaml handling
│   │   └── manifest.go    # Load/save project manifests
│   │
│   ├── marketplace/       # Marketplace parsing and generation
│   │   ├── parser.go      # Parse marketplace.json
│   │   ├── generator.go   # Generate packages from plugins
│   │   └── discovery.go   # Auto-discover marketplace files
│   │
│   ├── metadata/          # Metadata tracking
│   │   └── metadata.go    # Resource metadata management
│   │
│   ├── pattern/           # Pattern matching
│   │   └── matcher.go     # Glob pattern matching for resources
│   │
│   ├── repo/              # Repository management
│   │   └── manager.go     # Add/remove/list resources, bulk import
│   │
│   ├── resource/          # Resource types and validation
│   │   ├── resource.go    # Base resource type
│   │   ├── command.go     # Command resource logic
│   │   ├── skill.go       # Skill resource logic
│   │   ├── agent.go       # Agent resource logic
│   │   ├── package.go     # Package resource logic
│   │   └── types.go       # Resource types and validation
│   │
│   ├── source/            # Source parsing and Git operations
│   │   ├── parser.go      # Parse source formats (gh:, local:, git:)
│   │   └── git.go         # Git clone and temp directory management
│   │
│   ├── tools/             # Tool-specific information
│   │   └── types.go       # Claude/OpenCode/Copilot definitions
│   │
│   ├── version/           # Version information
│   │   └── version.go     # Embedded at build time
│   │
│   └── workspace/         # Workspace caching
│       └── cache.go       # Git repository caching for repo operations
│
├── test/                   # Integration tests
│   ├── integration_test.go       # End-to-end workflow tests
│   ├── ai_package_test.go        # ai.package.yaml tests
│   ├── marketplace_test.go       # Marketplace import tests
│   ├── package_import_test.go    # Package auto-import tests
│   └── github_sources_test.go    # GitHub source tests
│
├── examples/               # Example resources
│   ├── sample-command.md
│   ├── sample-agent.md
│   ├── sample-skill/
│   ├── marketplace/        # Marketplace examples
│   ├── packages/           # Package examples
│   └── ai-package/         # ai.package.yaml examples
│
├── main.go                 # Entry point
├── Makefile                # Build automation
├── go.mod                  # Go module definition
├── README.md               # User documentation
├── AGENTS.md               # AI agent quick reference
└── CONTRIBUTING.md         # This file
```
### Architecture Flow

#### Adding Resources (Local)

1. **User runs CLI command** → `cmd/add.go` (Cobra)
2. **Command validates input** → `pkg/resource/` (validation)
3. **Manager handles operation** → `pkg/repo/manager.go` (add/remove/list)
4. **Resources stored** → `~/.local/share/ai-config/repo/`
5. **Installation creates symlinks** → `pkg/install/` → `.claude/`, `.opencode/`, etc.

#### Adding Resources (GitHub)

1. **User runs CLI command** → `cmd/add.go` with GitHub source
2. **Parse source format** → `pkg/source/parser.go` (gh:owner/repo → ParsedSource)
3. **Clone repository** → `pkg/source/git.go` (git clone to temp dir)
4. **Auto-discover resources** → `pkg/discovery/` (search standard locations)
5. **User selection** (if multiple resources found)
6. **Copy to repository** → `pkg/repo/manager.go` (centralized storage)
7. **Cleanup temp directory** → `pkg/source/git.go`
8. **Installation works as before** → symlinks to centralized repo

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
- Respects user's install.targets configuration

**Symlink-based Installation:**
- No duplication of resources
- Single source of truth in repository
- Easy updates (modify in repo, all projects updated)

**Source Parsing and Discovery:**
- Flexible source formats (gh:, local:, http:, git:)
- Auto-discovery follows tool conventions
- Priority-based search paths
- Handles multi-resource repositories

### GitHub Source Architecture

The GitHub source feature allows users to add resources directly from GitHub repositories with automatic discovery.

#### Source Parser (`pkg/source/parser.go`)

**Purpose:** Parse and normalize different source formats into a unified structure.

**Types:**
```go
type SourceType string

const (
    GitHub  SourceType = "github"
    GitLab  SourceType = "gitlab"  // Future
    Local   SourceType = "local"
    GitURL  SourceType = "git-url"
)

type ParsedSource struct {
    Type      SourceType
    URL       string      // Full Git URL
    LocalPath string      // For local sources
    Ref       string      // Branch/tag (optional)
    Subpath   string      // Path within repo (optional)
}
```

**Parsing Rules:**
- `gh:owner/repo` → GitHub with constructed URL
- `gh:owner/repo/path` → GitHub with subpath
- `gh:owner/repo@branch` → GitHub with ref
- `gh:owner/repo/path@branch` → Both subpath and ref
- `local:path` or bare paths → Local
- `https://` or `git@` URLs → GitURL
- `owner/repo` (no prefix) → Infer GitHub

#### Git Operations (`pkg/source/git.go`)

**Purpose:** Clone repositories and manage temporary directories.

**Functions:**
```go
// CloneRepo clones a Git repository to a temporary directory
func CloneRepo(url string, ref string) (tempDir string, err error)

// CleanupTempDir removes a temporary directory (with safety checks)
func CleanupTempDir(dir string) error
```

**Implementation Details:**
- Uses `git clone --depth 1` for shallow clones (faster)
- Creates temp directory using `os.MkdirTemp()`
- If ref specified, uses `--branch <ref>` flag
- Cleanup validates path is in temp directory (security)
- Always cleanup, even on errors (defer pattern)

#### Auto-Discovery (`pkg/discovery/`)

**Purpose:** Find resources in repositories following tool conventions.

**Skills Discovery** (`pkg/discovery/skills.go`):

Priority search order:
1. Direct path: `basePath/subpath/SKILL.md`
2. Standard directories:
   - `skills/`
   - `.claude/skills/`
   - `.opencode/skills/`
   - `.github/skills/`
   - `.codex/skills/`, `.cursor/skills/`, `.goose/skills/`
   - `.kilocode/skills/`, `.kiro/skills/`, `.roo/skills/`
   - `.trae/skills/`, `.agents/skills/`, `.agent/skills/`
3. Recursive search (max depth 5) if not found

**Commands Discovery** (`pkg/discovery/commands.go`):

Search locations:
1. `commands/`
2. `.claude/commands/`
3. `.opencode/commands/`
4. Recursive search for `.md` files (excluding `SKILL.md`, `README.md`)

**Agents Discovery** (`pkg/discovery/agents.go`):

Search locations:
1. `agents/`
2. `.claude/agents/`
3. `.opencode/agents/`
4. Recursive search for `.md` files with agent frontmatter

**Common Patterns:**
- Return `[]*resource.Resource`, not error if nothing found
- Deduplicate by resource name (first found wins)
- Validate frontmatter before including
- Max recursion depth of 5 levels

#### Integration in Add Command (`cmd/add.go`)

**Workflow:**
```go
// 1. Parse source
parsed := source.ParseSource(input)

// 2. Handle based on type
switch parsed.Type {
case source.Local:
    // Existing behavior - add directly
    manager.AddSkill(parsed.LocalPath)

case source.GitHub, source.GitURL:
    // Clone repository
    tempDir, err := source.CloneRepo(parsed.URL, parsed.Ref)
    defer source.CleanupTempDir(tempDir)
    
    // Discover resources
    searchPath := tempDir
    if parsed.Subpath != "" {
        searchPath = filepath.Join(tempDir, parsed.Subpath)
    }
    
    resources, err := discovery.DiscoverSkills(searchPath, "")
    
    // Handle selection
    if len(resources) == 0 {
        return fmt.Errorf("no skills found in %s", parsed.URL)
    }
    if len(resources) == 1 {
        // Add automatically
        manager.AddSkill(resources[0].Path)
    } else {
        // Interactive selection or error
        selected := promptUserSelection(resources)
        manager.AddSkill(selected.Path)
    }
}
```

### Testing GitHub Sources

When adding tests for GitHub sources:

1. **Use test fixtures** - Create mock repository structures in `testdata/`
2. **Mock git operations** - Use environment variables or test helpers to avoid real clones
3. **Test all source formats** - gh:, local:, https:, git@, shorthand
4. **Test discovery edge cases** - no resources, multiple resources, nested directories
5. **Test error handling** - invalid repos, network failures, cleanup on error

**Example test structure:**
```go
func TestDiscoverSkills(t *testing.T) {
    testRepo := setupTestRepo(t, "multi-skill-repo")
    defer cleanupTestRepo(testRepo)
    
    skills, err := discovery.DiscoverSkills(testRepo, "")
    if err != nil {
        t.Fatalf("DiscoverSkills failed: %v", err)
    }
    
    if len(skills) != 3 {
        t.Errorf("Expected 3 skills, got %d", len(skills))
    }
}
```

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

    "github.com/hk9890/ai-config-manager/pkg/resource"
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

# Unit tests only (fast, no external dependencies)
make unit-test

# Integration tests only (slower, requires git/network)
make integration-test

# E2E tests only (slowest, tests full CLI with real workflows)
make e2e-test

# Single test file
go test -v ./pkg/resource/command_test.go

# Specific test function
go test -v ./pkg/config -run TestLoad_ValidConfig

# With coverage
go test -v -cover ./pkg/...

# Specific package
go test -v ./pkg/resource/
```

### Test Types

**Unit Tests** (`make unit-test`):
- Located in package directories (`pkg/*/`)
- Test individual functions/methods in isolation
- Use fixtures from `testdata/` directories
- No network calls or git operations
- Execution time: <5 seconds
- Build tag: `//go:build unit` (optional)

**Integration Tests** (`make integration-test`):
- Located in `test/` directory
- Test package integration and workflows
- May use real git repositories (with workspace caching)
- Execution time: ~30 seconds
- Build tag: `//go:build integration`

**E2E Tests** (`make e2e-test`):
- Located in `test/e2e/` directory
- Test complete CLI workflows end-to-end
- Build actual binary and execute commands
- Use real filesystem operations and git repositories
- Test scenarios: sync, prune, read operations, helpers
- Execution time: ~1-2 minutes
- Build tag: `//go:build e2e`
- **Requirements**: git must be installed

### Running E2E Tests Locally

E2E tests build the full `aimgr` binary and test real command execution:

```bash
# Run all E2E tests
make e2e-test

# Run specific E2E test
go test -v -tags=e2e ./test/e2e/ -run TestSync

# Run with verbose output
go test -v -tags=e2e ./test/e2e/

# Run E2E tests with timeout
go test -v -tags=e2e -timeout 10m ./test/e2e/
```

**What E2E tests verify:**
- Full binary builds successfully
- Commands execute with correct exit codes
- Repository operations work end-to-end (import, sync, prune)
- Configuration and workspace caching function correctly
- Error messages are user-friendly
- Help output is correct

**Note**: E2E tests use git-ignored test repositories and temporary directories for isolation.

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

Thank you for contributing to aimgr! 🎉
