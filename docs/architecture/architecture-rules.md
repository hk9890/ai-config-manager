# Architecture Rules

This document establishes strict architectural rules for the ai-config-manager codebase. These rules ensure consistency, performance, and maintainability across the project.

## Table of Contents
- [Rule 1: All Git Operations Must Use Workspace Cache](#rule-1-all-git-operations-must-use-workspace-cache)
- [Rule 2: XDG Base Directory Specification](#rule-2-xdg-base-directory-specification)
- [Rule 3: Build Tags for Test Categories](#rule-3-build-tags-for-test-categories)
- [Rule 4: Error Wrapping Requirements](#rule-4-error-wrapping-requirements)
- [Version History](#version-history)

---

## Rule 1: All Git Operations Must Use Workspace Cache

**Status**: ✅ Active (Since 2026-01-27)

### Statement

All external Git repository operations **MUST** use the `pkg/workspace` cache component. Direct temporary directory cloning is **PROHIBITED** for production code.

### Rationale

#### Performance
- **10-50x faster** for subsequent operations on the same repository
- Cached clones eliminate redundant network operations
- Automatic batching for resources from the same source

#### Consistency
- Single source of truth for repository state
- Shared cache across all commands (`repo import`, `repo sync`)
- Predictable behavior across the application

#### Maintainability
- Centralized Git operation logic
- Easier to add features (ref switching, update strategies)
- Simplified error handling and recovery

### Correct Usage

✅ **DO**: Use workspace.Manager for Git operations

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

#### Common Patterns

**Pattern 1: Adding resources from Git**
```go
// Correct: Use workspace cache
mgr, _ := workspace.NewManager(repoPath)
clonePath, err := mgr.GetOrClone(url, ref)
if err != nil {
    return err
}

// Extract resources from clonePath
resources := discovery.DiscoverInDirectory(clonePath)
```

**Pattern 2: Updating resources from Git**
```go
// Correct: Update cached repository
mgr, _ := workspace.NewManager(repoPath)
if err := mgr.Update(url, ref); err != nil {
    return err
}

// Get updated path (reuses cache)
clonePath, err := mgr.GetOrClone(url, ref)
if err != nil {
    return err
}
```

**Pattern 3: Pruning unused caches**
```go
// Correct: Let workspace manage cache lifecycle
mgr, _ := workspace.NewManager(repoPath)
removed, err := mgr.Prune(referencedURLs)
if err != nil {
    return err
}
fmt.Printf("Removed %d unused caches\n", len(removed))
```

### Prohibited Patterns

❌ **DON'T**: Create temporary directories for Git clones

```go
// WRONG: Direct temporary directory clone
tempDir, err := os.MkdirTemp("", "git-clone-*")
if err != nil {
    return err
}
defer os.RemoveAll(tempDir)

cmd := exec.Command("git", "clone", url, tempDir)
if err := cmd.Run(); err != nil {
    return err
}

// This bypasses caching, causes performance issues
```

❌ **DON'T**: Use pkg/source.CloneRepo in production code

```go
// WRONG: pkg/source.CloneRepo is deprecated for production use
import "github.com/hk9890/ai-config-manager/pkg/source"

tempDir, err := source.CloneRepo(url, ref)
if err != nil {
    return err
}
defer source.CleanupTempDir(tempDir)

// This function exists for backward compatibility only
// It should NOT be used in new code
```

### Exceptions

The following cases are **exempt** from this rule:

1. **Unit Tests**: Tests may use temporary directories for isolated testing
   ```go
   // OK in tests: Use t.TempDir() for isolated test environments
   func TestGitOperations(t *testing.T) {
       tempDir := t.TempDir()
       // Test with temporary clone
   }
   ```

2. **Integration Tests**: Integration tests in `test/` directory may use temp dirs
   ```go
   // OK in integration tests: Testing end-to-end behavior
   tempDir, _ := os.MkdirTemp("", "integration-test-*")
   defer os.RemoveAll(tempDir)
   ```

3. **Legacy Code**: `pkg/source/git.go` maintains temp-based functions for backward compatibility
   - These functions should NOT be called by new code
   - They exist only to avoid breaking existing callers during migration

### Enforcement

#### Code Review
- All PRs adding Git operations must use `pkg/workspace`
- Reviewers should flag any `os.MkdirTemp` + `git clone` patterns
- Exceptions require explicit justification in PR description

#### Static Analysis
Future enhancement: Add linting rule to detect prohibited patterns:
```bash
# Detect temporary directory + git clone patterns
grep -r "os.MkdirTemp.*git.*clone" pkg/ --include="*.go" --exclude="*_test.go"
```

#### Migration Tracking
Active migration efforts tracked in beads:
- Gate: `ai-config-manager-xo9y` - Verify no temporary Git cloning remains
- Epic: Migrate all Git operations to workspace cache

### References

- **Implementation**: `pkg/workspace/manager.go`
- **Documentation**: `docs/workspace-caching.md`
- **Design Comments**: See `pkg/workspace/manager.go` header comments

---

## Rule 2: XDG Base Directory Specification

**Status**: ✅ Active (Since Project Inception)

### Statement

All application data **MUST** follow the XDG Base Directory Specification for cross-platform compatibility.

### Implementation

```go
import "github.com/adrg/xdg"

// Data directory: ~/.local/share/ai-config/repo/
repoPath := filepath.Join(xdg.DataHome, "ai-config", "repo")

// Config directory: ~/.config/ai-config/
configPath := filepath.Join(xdg.ConfigHome, "ai-config")

// Cache directory: ~/.cache/ai-config/
cachePath := filepath.Join(xdg.CacheHome, "ai-config")
```

### Why This Matters

- **Linux/macOS**: Follows standard directory conventions
- **Windows**: Automatically maps to appropriate directories (AppData, etc.)
- **User Control**: Respects `XDG_DATA_HOME`, `XDG_CONFIG_HOME` environment variables
- **System Integration**: Integrates with backup tools, sync utilities

### Prohibited Patterns

❌ **DON'T**: Hardcode home directory paths
```go
// WRONG: Hardcoded path
homeDir, _ := os.UserHomeDir()
repoPath := filepath.Join(homeDir, ".ai-config")  // Not XDG compliant
```

---

## Rule 3: Build Tags for Test Categories

**Status**: ✅ Active (Since 2026-01-26)

### Statement

Tests **MUST** use build tags to categorize test types: `unit`, `integration`, or both.

### Implementation

**Unit Tests**:
```go
//go:build unit

package mypackage_test

func TestMyFunction(t *testing.T) {
    // Fast, isolated test
}
```

**Integration Tests**:
```go
//go:build integration

package test

func TestEndToEnd(t *testing.T) {
    // Slower, system-level test
}
```

### Running Tests

```bash
# Run only unit tests (fast)
go test -tags=unit ./...

# Run only integration tests
go test -tags=integration ./...

# Run all tests
go test -tags="unit integration" ./...
```

### Why This Matters

- **CI/CD Optimization**: Run fast unit tests on every commit, integration tests on merge
- **Developer Experience**: Quick feedback loop with unit tests
- **Resource Management**: Integration tests may require network, filesystem

---

## Rule 4: Error Wrapping Requirements

**Status**: ✅ Active (Since Project Inception)

### Statement

All errors **MUST** be wrapped with context using `fmt.Errorf` with `%w` verb for error chain preservation.

### Correct Usage

✅ **DO**: Wrap errors with context
```go
file, err := os.Open(path)
if err != nil {
    return fmt.Errorf("failed to open configuration file: %w", err)
}

data, err := parser.Parse(content)
if err != nil {
    return fmt.Errorf("failed to parse resource metadata: %w", err)
}
```

### Prohibited Patterns

❌ **DON'T**: Return raw errors without context
```go
// WRONG: No context
if err != nil {
    return err
}

// WRONG: String formatting loses error chain
if err != nil {
    return fmt.Errorf("error: %s", err.Error())
}

// WRONG: Using %v instead of %w
if err != nil {
    return fmt.Errorf("failed to load: %v", err)
}
```

### Why This Matters

- **Debugging**: Full error chain shows exactly where failure occurred
- **Error Handling**: Enables `errors.Is()` and `errors.As()` for typed error checks
- **User Experience**: Clear, actionable error messages

### Guidelines

1. **Be Specific**: Include the operation that failed
   - Good: `"failed to load command from ~/.claude/commands/build.md"`
   - Bad: `"load failed"`

2. **Include Context**: Add relevant identifiers
   - Good: `"failed to install skill/pdf-processing: skill directory already exists"`
   - Bad: `"install failed"`

3. **Preserve Chain**: Always use `%w` when wrapping errors
   ```go
   return fmt.Errorf("failed to %s: %w", operation, err)
   ```

---

## Version History

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 1.0 | 2026-01-27 | Initial version with Git workspace rule | AI Agent |
| 1.0 | 2026-01-27 | Added XDG, build tags, error wrapping rules | AI Agent |

---

## Related Documentation

- [Workspace Caching](workspace-caching.md) - Detailed workspace cache documentation
- [Test Refactoring](test-refactoring.md) - Test organization and build tags
- [AGENTS.md](../AGENTS.md) - Development guidelines for AI agents
- [Resource Formats](resource-formats.md) - Resource specification formats
