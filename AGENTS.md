# AGENTS.md

This document provides guidelines for AI coding agents working in the ai-config-manager repository.

## Project Overview

**aimgr** is a CLI tool for managing AI resources (commands, skills, and agents) across multiple AI coding tools (Claude Code, OpenCode, GitHub Copilot). It uses a centralized repository with symlink-based installation.

- **Language**: Go 1.25.6
- **Architecture**: CLI built with Cobra, resource management with symlinks
- **Storage**: `~/.local/share/ai-config/repo/` (XDG data directory)
- **Supported Resources**: Commands, Skills, Agents

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
│   ├── metadata/     # Metadata tracking and migration
│   ├── repo/         # Repository management
│   ├── resource/     # Resource types (command, skill, agent)
│   ├── tools/        # Tool-specific info (Claude, OpenCode, Copilot)
│   └── version/      # Version information
├── test/             # Integration tests
├── examples/         # Example resources (commands, skills, agents)
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

	"github.com/hk9890/ai-config-manager/pkg/resource"
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
Resources (commands/skills/agents) must follow agentskills.io naming:
- Lowercase alphanumeric + hyphens only
- Cannot start/end with hyphen
- No consecutive hyphens (`--`)
- 1-64 characters max
- Examples: `test-command`, `pdf-processing`, `skill-v2`, `code-reviewer`

### Types
- Use explicit types, avoid `interface{}` when possible
- Define custom types for domain concepts:
  ```go
  type ResourceType string
  
  const (
      Command ResourceType = "command"
      Skill   ResourceType = "skill"
      Agent   ResourceType = "agent"
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

// Agents: single .md file
res, err := resource.LoadAgent("path/to/agent.md")
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

**Workspace Caching**: Git repositories are cached in `.workspace/` directory for
efficient reuse across updates:
- First update: Full git clone (creates cache)
- Subsequent updates: Git pull only (10-50x faster)
- Automatic cache management with SHA256 hash-based storage
- Shared across all resources from the same source repository

**Update Performance**: The `repo update` command automatically batches resources from
the same Git repository, cloning each unique source only once. This optimization
significantly improves update speed for bulk operations (e.g., 39 resources from one
repository = 1 clone instead of 39).

**Workspace Directory Structure**:
```
~/.local/share/ai-config/repo/
├── .workspace/                   # Git repository cache
│   ├── <hash-1>/                 # Cached repository 1 (by URL hash)
│   │   ├── .git/
│   │   ├── commands/
│   │   └── skills/
│   ├── <hash-2>/                 # Cached repository 2
│   │   └── ...
│   └── .cache-metadata.json      # Cache metadata (URLs, timestamps, refs)
├── .metadata/                    # Resource metadata
├── commands/                     # Command resources
├── skills/                       # Skill resources
└── agents/                       # Agent resources
```

**Cache Management**:
```bash
# Remove unreferenced caches to free disk space
aimgr repo prune

# Preview what would be removed
aimgr repo prune --dry-run

# Force remove without confirmation
aimgr repo prune --force
```

Run `repo prune` after removing many resources or when `.workspace/` grows too large.

### Tool Detection
```go
tool, err := tools.ParseTool("claude")  // Returns tools.Claude
tool, err := tools.ParseTool("opencode") // Returns tools.OpenCode
info := tools.GetToolInfo(tool)         // Get dirs, etc.
```

### Pattern Matching

The `pkg/pattern` package provides glob pattern matching for resources:

```go
import "github.com/hk9890/ai-config-manager/pkg/pattern"

// Parse pattern to extract type and check if it's a pattern
resourceType, patternStr, isPattern := pattern.ParsePattern("skill/pdf*")
// Returns: resource.Skill, "pdf*", true

// Create a matcher
matcher, err := pattern.NewMatcher("skill/pdf*")
if err != nil {
    return err
}

// Match against resources
res := &resource.Resource{Type: resource.Skill, Name: "pdf-processing"}
if matcher.Match(res) {
    fmt.Println("Matched!")
}

// Check if pattern (vs exact name)
if matcher.IsPattern() {
    fmt.Println("This is a glob pattern")
}

// Get resource type filter (if specified)
resType := matcher.GetResourceType()  // Returns resource.Skill or ""

// Match by name only (useful when you already know the type)
if matcher.MatchName("pdf-processing") {
    fmt.Println("Name matches!")
}
```

**Pattern Features:**
- Supports standard glob operators: `*`, `?`, `[abc]`, `{a,b}`
- Optional type prefix: `type/pattern` or just `pattern`
- Type filtering: `skill/*` only matches skills
- Cross-type matching: `*test*` matches across all types
- Exact matching: No wildcards = exact name match

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

## Resource Format Details

### Package Format

Packages are collections of resources that can be installed together as a unit. They are stored as JSON files in the `packages/` directory.

#### File Structure

Package files are stored as:
- **Repository**: `~/.local/share/ai-config/repo/packages/<package-name>.package.json`
- **Metadata**: `~/.local/share/ai-config/repo/.metadata/packages/<package-name>-metadata.json`

Packages are **not** installed directly into projects. Instead, when you install a package, all its referenced resources are installed as symlinks.

#### Package JSON Format

```json
{
  "name": "package-name",
  "description": "Human-readable description of the package",
  "resources": [
    "command/command-name",
    "skill/skill-name",
    "agent/agent-name"
  ]
}
```

**Fields:**

- **name** (string, required): Package name following agentskills.io naming rules
  - Lowercase alphanumeric + hyphens only
  - Cannot start/end with hyphen
  - No consecutive hyphens (`--`)
  - 1-64 characters max
  - Examples: `web-tools`, `testing-suite`, `docs-helpers`

- **description** (string, required): Human-readable description (1-1024 characters)

- **resources** ([]string, required): Array of resource references in `type/name` format
  - Format: `type/name` where type is `command`, `skill`, or `agent`
  - Examples: `command/test`, `skill/pdf-processing`, `agent/code-reviewer`
  - All referenced resources must exist in the repository

#### Format Examples

**Minimal Package:**
```json
{
  "name": "minimal-tools",
  "description": "Minimal tool collection",
  "resources": [
    "command/test"
  ]
}
```

**Web Development Package:**
```json
{
  "name": "web-dev-tools",
  "description": "Complete web development toolkit",
  "resources": [
    "command/build",
    "command/dev",
    "skill/typescript-helper",
    "skill/react-helper",
    "agent/code-reviewer"
  ]
}
```

**Testing Suite Package:**
```json
{
  "name": "testing-suite",
  "description": "Complete testing toolkit with commands, skills, and QA agent",
  "resources": [
    "command/test",
    "command/coverage",
    "skill/test-generator",
    "skill/mock-helper",
    "agent/qa-tester"
  ]
}
```

#### Code Structure

The package resource is represented by the `Package` struct in `pkg/resource/package.go`:

```go
type Package struct {
    Name        string   `json:"name"`        // Package name
    Description string   `json:"description"` // Human-readable description
    Resources   []string `json:"resources"`   // Resource references (type/name)
}
```

**Loading packages:**
```go
// Load package from file
pkg, err := resource.LoadPackage("path/to/package.package.json")

// Get package path in repo
packagePath := resource.GetPackagePath("package-name", repoPath)
```

**Creating packages:**
```go
pkg := &resource.Package{
    Name:        "web-tools",
    Description: "Web development tools",
    Resources:   []string{"command/build", "skill/typescript-helper"},
}

err := resource.SavePackage(pkg, repoPath)
```

**Parsing resource references:**
```go
// Parse type/name format
resType, resName, err := resource.ParseResourceReference("command/test")
// Returns: resource.Command, "test", nil
```

#### CLI Commands

**Create a package:**
```bash
aimgr repo create-package web-tools \
  --description="Web development tools" \
  --resources="command/build,skill/typescript-helper"
```

**Install a package:**
```bash
# Installs all resources in the package
aimgr install package/web-tools
```

**Uninstall a package:**
```bash
# Removes all resource symlinks
aimgr uninstall package/web-tools
```

**Remove a package:**
```bash
# Remove package only (keeps resources)
aimgr repo remove package/web-tools

# Remove package and all its resources
aimgr repo remove package/web-tools --with-resources
```

**List packages:**
```bash
# List all packages in repository
aimgr repo list package
```

### Agent Resource Format

Agents are single `.md` files with YAML frontmatter that define AI agents with specialized roles and capabilities. The codebase supports both OpenCode and Claude Code formats.

#### File Structure

Agent files are stored as:
- **Repository**: `~/.local/share/ai-config/repo/agents/<agent-name>.md`
- **Installed**: `.claude/agents/<agent-name>.md` or `.opencode/agents/<agent-name>.md`

#### YAML Frontmatter Fields

**Required:**
- `description` (string): Brief description of what the agent does (1-1024 chars for validation)

**Optional (OpenCode format):**
- `type` (string): Agent type/role (e.g., "code-reviewer", "tester", "documentation")
- `instructions` (string): Detailed instructions for the agent's behavior
- `capabilities` ([]string): List of agent capabilities (e.g., "static-analysis", "security-scan")

**Optional (Common):**
- `version` (string): Semantic version (e.g., "1.0.0")
- `author` (string): Author name or organization
- `license` (string): License identifier (e.g., "MIT", "Apache-2.0")
- `metadata` (map): Additional metadata fields

#### Format Examples

**Minimal Agent (Claude format):**
```yaml
---
description: Minimal agent for testing
---

# Minimal Agent

This agent has only the required description field.
```

**OpenCode Format Agent:**
```yaml
---
description: Code review agent that checks for best practices
type: code-reviewer
instructions: Review code for quality, security, and performance
capabilities:
  - static-analysis
  - security-scan
  - performance-review
version: "1.0.0"
author: team-name
license: MIT
---

# Code Reviewer Agent

This agent analyzes code for quality and security issues.
```

**Claude Format Agent:**
```yaml
---
description: A Claude format agent for testing
version: "2.0.0"
author: claude-team
license: Apache-2.0
metadata:
  category: development
  tags: testing,qa
---

# Claude Agent

Instructions and guidelines go in the markdown body for Claude format.
```

#### Code Structure

The agent resource is represented by the `AgentResource` struct in `pkg/resource/agent.go`:

```go
type AgentResource struct {
    Resource                                     // Embedded base resource
    Type         string   `yaml:"type,omitempty"`         // Agent role (OpenCode)
    Instructions string   `yaml:"instructions,omitempty"` // Agent instructions (OpenCode)
    Capabilities []string `yaml:"capabilities,omitempty"` // Capability list
    Content      string   `yaml:"-"`                      // Markdown content
}
```

**Loading agents:**
```go
// Load base resource (minimal info)
res, err := resource.LoadAgent("path/to/agent.md")

// Load full agent with all details
agent, err := resource.LoadAgentResource("path/to/agent.md")
```

**Creating agents:**
```go
agent := resource.NewAgentResource("code-reviewer", "Reviews code quality")
agent.Type = "code-reviewer"
agent.Instructions = "Check for best practices"
agent.Capabilities = []string{"static-analysis", "security"}
agent.Content = "# Code Reviewer\n\nDetailed docs here."

err := resource.WriteAgent(agent, "code-reviewer.md")
```

### OpenCode Support

OpenCode is supported as a first-class AI tool alongside Claude Code and GitHub Copilot.

#### Tool Directories

| Tool | Commands | Skills | Agents |
|------|----------|--------|--------|
| Claude Code | `.claude/commands/` | `.claude/skills/` | `.claude/agents/` |
| OpenCode | `.opencode/commands/` | `.opencode/skills/` | `.opencode/agents/` |
| GitHub Copilot | N/A | `.github/skills/` | N/A |

#### OpenCode-Specific Features

OpenCode agents support additional frontmatter fields:
- `type`: Agent role/category
- `instructions`: Detailed behavior instructions
- `capabilities`: Array of capability strings

These fields are optional and allow more structured agent definitions. Claude format agents typically put instructions in the markdown body instead.

#### Bulk Import

Import all resources from directories using auto-discovery:

```bash
# Import from .opencode folder
aimgr repo add ~/.opencode
aimgr repo add ~/project/.opencode

# Import from .claude folder
aimgr repo add ~/.claude
aimgr repo add ~/project/.claude

# Filter specific resource types
aimgr repo add ~/.opencode --filter "skill/*"
aimgr repo add ~/project/.claude --filter "agent/*"
```

This imports:
- Commands from `.opencode/commands/*.md` or `.claude/commands/*.md`
- Skills from `.opencode/skills/*/SKILL.md` or `.claude/skills/*/SKILL.md`
- Agents from `.opencode/agents/*.md` or `.claude/agents/*.md`

## When Making Changes
1. Run `make fmt` before committing
2. Run `make test` to verify all tests pass
3. Add tests for new functionality
4. Update documentation if adding user-facing features
5. Follow existing code patterns and conventions
