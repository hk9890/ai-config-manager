# Architecture Rules

This document establishes strict architectural rules for the ai-config-manager codebase. These rules ensure consistency, performance, and maintainability across the project.

## Table of Contents
- [Rule 1: All Git Operations Must Use Workspace Cache](#rule-1-all-git-operations-must-use-workspace-cache)
- [Rule 2: XDG Base Directory Specification](#rule-2-xdg-base-directory-specification)
- [Rule 3: Build Tags for Test Categories](#rule-3-build-tags-for-test-categories)
- [Rule 4: Error Wrapping Requirements](#rule-4-error-wrapping-requirements)
- [Rule 5: Symlink Handling for Filesystem Operations](#rule-5-symlink-handling-for-filesystem-operations)
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

## Rule 5: Symlink Handling for Filesystem Operations

**Status**: ✅ Active (Since 2026-02-05)

### Statement

All filesystem traversal code **MUST** support both real files/directories (COPY mode) and symlinks (SYMLINK mode). When checking if a path is a directory, use `os.Stat()` which follows symlinks, not `entry.IsDir()` from `os.ReadDir()`.

### Context

Resources can be imported via two modes:
- **COPY mode**: GitHub imports - creates real files/directories in repository
- **SYMLINK mode**: Local imports - creates symlinks to source locations

All filesystem traversal code must support BOTH modes transparently.

### Problem

`os.ReadDir()` returns directory entries where `entry.IsDir()` reports `false` for symlinks to directories. This causes code to skip symlinked directories when it should process them as regular directories.

**Example of the issue:**
```bash
# Real directory
$ ls -ld /repo/skills/real-skill/
drwxr-xr-x  /repo/skills/real-skill/
# entry.IsDir() = true ✓

# Symlinked directory  
$ ls -ld /repo/skills/symlinked-skill/
lrwxrwxrwx  /repo/skills/symlinked-skill/ -> /source/symlinked-skill/
# entry.IsDir() = false ✗ (reports as symlink, not directory)
```

### Correct Usage

✅ **DO**: Use `os.Stat()` which follows symlinks

```go
entries, err := os.ReadDir(dir)
if err != nil {
    return fmt.Errorf("failed to read directory: %w", err)
}

for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    
    // Follow symlinks to check if target is a directory
    info, err := os.Stat(path)
    if err != nil {
        // Handle error (broken symlink, permissions, etc.)
        continue
    }
    
    if info.IsDir() {
        // Process directory (works for both real and symlinked dirs)
        processDirectory(path)
    }
}
```

✅ **DO**: Use `os.Lstat()` when you need symlink metadata

```go
// When you explicitly need to know if something IS a symlink
info, err := os.Lstat(path)
if err != nil {
    return err
}

if info.Mode()&os.ModeSymlink != 0 {
    // This is a symlink - get its target
    target, err := os.Readlink(path)
    // ...
}
```

### Prohibited Patterns

❌ **DON'T**: Use `entry.IsDir()` directly from `os.ReadDir()`

```go
// WRONG: Skips symlinked directories
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    if entry.IsDir() {  // ← Returns false for symlinks!
        processDirectory(entry.Name())
    }
}
```

❌ **DON'T**: Skip symlinks without justification

```go
// WRONG: Silently excludes symlinked resources
if info.Mode()&os.ModeSymlink != 0 {
    continue  // Why? This breaks SYMLINK mode!
}
```

### Guidelines

1. **Default Behavior**: Follow symlinks during directory traversal
2. **Document Exceptions**: If you intentionally skip symlinks, document why
3. **Test Both Modes**: Every discovery function must test both real and symlinked resources
4. **Error Handling**: Handle broken symlinks gracefully (permissions, missing targets)

### Common Use Cases

**Directory Traversal for Resource Discovery:**
```go
// Pattern: List all skill directories
func findSkills(repoDir string) ([]string, error) {
    skillsDir := filepath.Join(repoDir, "skills")
    entries, err := os.ReadDir(skillsDir)
    if err != nil {
        return nil, err
    }
    
    var skills []string
    for _, entry := range entries {
        path := filepath.Join(skillsDir, entry.Name())
        
        // ✓ Follow symlinks
        info, err := os.Stat(path)
        if err != nil {
            log.Printf("skipping %s: %v", path, err)
            continue
        }
        
        if info.IsDir() {
            skills = append(skills, path)
        }
    }
    return skills, nil
}
```

**Recursive Directory Walk:**
```go
// Pattern: Walk directory tree, following symlinks
func walkResources(root string, fn func(path string) error) error {
    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // filepath.Walk automatically follows symlinks via os.Lstat
        // But you still need os.Stat if you need the target's info
        if info.IsDir() {
            return fn(path)
        }
        return nil
    })
}
```

### Testing Requirements

Every resource discovery/listing function **MUST** test both:
- ✅ Real directories (COPY mode)
- ✅ Symlinked directories (SYMLINK mode)

**Test Pattern:**
```go
func TestDiscoverSkills_WithSymlinks(t *testing.T) {
    // Test 1: Real directory
    realDir := filepath.Join(t.TempDir(), "real-skill")
    os.MkdirAll(realDir, 0755)
    os.WriteFile(filepath.Join(realDir, "SKILL.md"), []byte("# Skill"), 0644)
    
    // Test 2: Symlinked directory
    sourceDir := filepath.Join(t.TempDir(), "source-skill")
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte("# Skill"), 0644)
    
    symlinkDir := filepath.Join(t.TempDir(), "symlink-skill")
    os.Symlink(sourceDir, symlinkDir)
    
    // Both should be discovered
    skills, err := discoverSkills(t.TempDir())
    require.NoError(t, err)
    assert.Len(t, skills, 2)
}
```

**Test Helper Available:**
```go
import "github.com/hk9890/ai-config-manager/test/testutil"

// Creates source directory and symlink, returns both paths + cleanup
sourceDir, symlinkPath, cleanup := testutil.CreateSymlinkedDir(t, "my-skill")
defer cleanup()
```

### Enforcement

#### Code Review Checklist
- [ ] If code uses `os.ReadDir()` + `entry.IsDir()`, verify symlinks are handled
- [ ] Tests include both real and symlinked directory cases
- [ ] Intentional symlink skipping is documented with rationale

#### Static Analysis
```bash
# Detect potential symlink issues in discovery code
rg "entry\.IsDir\(\)" pkg/discovery/ pkg/repo/
```

### Historical Context

This rule was added to prevent recurrence of symlink-related bugs:

- **ai-config-manager-nm6i** (2026-02-03): Skills listing skipped symlinked skills
  - Root cause: Used `entry.IsDir()` which returns false for symlinks
  - Impact: Local symlinked skills invisible to `aimgr list`
  - Fix: Changed to `os.Stat()` + `info.IsDir()`

- **ai-config-manager-pepy** (2026-02-03): copyDir and package discovery skip symlinks
  - Same pattern in multiple code locations
  - Discovered during comprehensive audit

### Why This Matters

1. **Feature Parity**: SYMLINK mode should behave identically to COPY mode
2. **User Expectations**: Users expect symlinked resources to work transparently
3. **Development Workflow**: Local development often uses symlinks for rapid iteration
4. **Data Integrity**: Skipping symlinks silently loses user data

### Related Rules

- **Rule 1**: Git workspace caching (affects how resources are initially stored)
- **Rule 4**: Error wrapping (handle broken symlinks with clear error messages)

---

## Version History

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 1.0 | 2026-01-27 | Initial version with Git workspace rule | AI Agent |
| 1.0 | 2026-01-27 | Added XDG, build tags, error wrapping rules | AI Agent |
| 1.1 | 2026-02-05 | Added Rule 5: Symlink handling for filesystem operations | AI Agent |

---

## Related Documentation

- [Workspace Caching](workspace-caching.md) - Detailed workspace cache documentation
- [Test Refactoring](test-refactoring.md) - Test organization and build tags
- [AGENTS.md](../AGENTS.md) - Development guidelines for AI agents
- [Resource Formats](resource-formats.md) - Resource specification formats
