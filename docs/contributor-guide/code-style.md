# Code Style Guide

Complete code style guidelines for ai-config-manager contributors.

## General Principles

- **Clarity over cleverness**: Write obvious code
- **Single Responsibility**: Functions do one thing well
- **Error handling**: Always handle errors, provide context
- **Documentation**: Export functions/types must have GoDoc comments
- **Testing**: All new code must have tests

## Naming Conventions

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

## Import Organization

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

## Error Handling

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}
```

**Rules**:
- Use `%w` to wrap errors (preserves error chain)
- Provide descriptive context
- Don't panic (except in main/init for fatal errors)
- Check errors immediately

**Examples**:

```go
// ✅ CORRECT: Wrapped with context
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

See [Architecture Rules - Rule 4](architecture.md#rule-4-error-wrapping-requirements) for complete details.

## Comments and Documentation

```go
// LoadCommand loads a command resource from a markdown file.
// It validates the file format and parses the YAML frontmatter.
// Returns an error if the file is not a valid command resource.
func LoadCommand(filePath string) (*Resource, error) {
    // Implementation
}
```

**Rules**:
- All exported items must have GoDoc comments
- Start with the item name
- Describe what, not how
- Keep comments up-to-date with code

## File Operations

### Cross-Platform Paths

```go
// ✅ GOOD: Use filepath.Join for cross-platform paths
path := filepath.Join(dir, "commands", "test.md")
```

### File Existence

```go
// ✅ GOOD: Check file existence
if _, err := os.Stat(path); err != nil {
    return fmt.Errorf("file does not exist: %w", err)
}
```

### Cleanup with Defer

```go
// ✅ GOOD: Use defer for cleanup
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()
```

### File Permissions

```go
// ✅ GOOD: Set appropriate permissions
os.MkdirAll(dir, 0755)        // Directories
os.WriteFile(path, data, 0644) // Files
```

## Symlink Handling

**CRITICAL:** Resources can be stored as real files (COPY mode) or symlinks (SYMLINK mode). All code must support both transparently.

### The Problem

`entry.IsDir()` from `os.ReadDir()` returns `false` for symlinks to directories!

### Wrong Approach

```go
// ❌ WRONG: Skips symlinked directories
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    if entry.IsDir() {  // Returns false for symlinks!
        processDirectory(entry.Name())
    }
}
```

### Correct Approach

```go
// ✅ CORRECT: Follows symlinks
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    info, err := os.Stat(path)  // os.Stat follows symlinks
    if err != nil {
        continue  // Handle broken symlinks gracefully
    }
    if info.IsDir() {
        processDirectory(path)  // Works for both real and symlinked dirs
    }
}
```

**Key Rule:** Use `os.Stat()` to follow symlinks, not `entry.IsDir()` from `os.ReadDir()`.

**Testing Requirement:** Every discovery function MUST test both real and symlinked resources.

See [Architecture Rules - Rule 5](architecture.md#rule-5-symlink-handling) for complete details.

## Best Practices Summary

### Do's

- ✅ Use `filepath.Join()` for paths
- ✅ Wrap errors with `fmt.Errorf(..., %w, err)`
- ✅ Use `os.Stat()` for symlink-aware checks
- ✅ Group imports in 3 sections
- ✅ Add GoDoc comments to exported items
- ✅ Use descriptive variable names
- ✅ Follow existing patterns in codebase

### Don'ts

- ❌ Don't use `entry.IsDir()` for discovery
- ❌ Don't return raw errors without context
- ❌ Don't use string concatenation for paths
- ❌ Don't panic (except main/init)
- ❌ Don't use `%s` or `%v` for error wrapping
- ❌ Don't skip error checking

## Related Documentation

- **[Architecture Guide](architecture.md)** - System overview and design rules
- **[Testing Guide](testing.md)** - Testing patterns and practices
- **[CONTRIBUTING.md](../../CONTRIBUTING.md)** - Quick start and workflow

## Examples from Codebase

For real-world examples, see:
- `pkg/resource/` - Resource loading patterns
- `pkg/discovery/` - Directory traversal with symlink support
- `pkg/repo/manager.go` - Error wrapping examples
