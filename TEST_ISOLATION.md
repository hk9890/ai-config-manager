# Test Isolation in ai-config-manager

This document explains how tests are isolated from the user's actual aimgr repository to prevent test metadata pollution.

## Problem

Tests that create resources can potentially pollute the user's actual aimgr repository at `~/.local/share/ai-config/` with test metadata, leftover files, and other artifacts. This is problematic because:

1. Test resources could interfere with actual user resources
2. Test metadata files could accumulate in the user's repo
3. Failed tests could leave orphaned files
4. Tests running in parallel could conflict with each other

## Solution

All tests use **isolated temporary repositories** that are automatically cleaned up after each test run.

## Implementation

### Core Mechanism

The repository manager supports custom repository paths for testing:

```go
// Production: Uses user's actual repo at ~/.local/share/ai-config/repo/
func NewManager() (*Manager, error)

// Testing: Uses custom temporary path
func NewManagerWithPath(repoPath string) *Manager
```

### Test Setup Pattern

All tests follow this pattern:

```go
func TestSomething(t *testing.T) {
    // Create isolated temporary directory (auto-cleanup)
    tmpDir := t.TempDir()
    
    // Create manager with custom path
    manager := repo.NewManagerWithPath(tmpDir)
    
    // ... perform test operations ...
    
    // No cleanup needed - t.TempDir() automatically removes tmpDir
}
```

### Key Benefits

1. **Automatic Cleanup**: `t.TempDir()` automatically removes temporary directories when tests complete
2. **Parallel Safety**: Each test gets its own isolated directory, enabling safe parallel execution
3. **No User Impact**: Tests never write to `~/.local/share/ai-config/`
4. **Realistic Testing**: Tests use the same code paths as production, just with different base paths

## Test Coverage

### Unit Tests

All unit tests in `pkg/` use isolated temporary repositories:

- `pkg/repo/manager_test.go` - Repository operations
- `pkg/install/installer_test.go` - Installation operations  
- `pkg/metadata/metadata_test.go` - Metadata operations
- `pkg/resource/*_test.go` - Resource loading/validation

### Integration Tests

All integration tests in `test/` use isolated temporary repositories:

- `test/integration_test.go` - Complete workflows
- `test/bulk_import_test.go` - Bulk import operations
- `test/cli_integration_test.go` - CLI command execution
- `test/bulk_add_*.go` - Bulk add operations

### Command Tests

Command tests in `cmd/` use isolated temporary repositories:

- `cmd/install_test.go` - Pattern expansion and installation
- `cmd/uninstall_test.go` - Pattern-based uninstallation
- `cmd/list_installed_test.go` - Listing installed resources

## Test Helper Functions

Several helper functions ensure consistent test isolation:

### createTestRepo()

Creates a temporary repository with test resources:

```go
func createTestRepo(t *testing.T) (repoPath string, cleanup func()) {
    t.Helper()
    
    // Create temp directory for repo
    tempDir := t.TempDir()
    repoPath = tempDir
    
    // Create directory structure
    os.MkdirAll(filepath.Join(repoPath, "commands"), 0755)
    os.MkdirAll(filepath.Join(repoPath, "skills"), 0755)
    os.MkdirAll(filepath.Join(repoPath, "agents"), 0755)
    
    // Create test resources...
    
    cleanup = func() {
        // Cleanup is automatic with t.TempDir()
    }
    
    return repoPath, cleanup
}
```

### createTestProject()

Creates a temporary project directory with tool directories:

```go
func createTestProject(t *testing.T) string {
    t.Helper()
    
    projectDir := t.TempDir()
    
    // Create .claude directories
    os.MkdirAll(filepath.Join(projectDir, ".claude", "commands"), 0755)
    os.MkdirAll(filepath.Join(projectDir, ".claude", "skills"), 0755)
    os.MkdirAll(filepath.Join(projectDir, ".claude", "agents"), 0755)
    
    return projectDir
}
```

## Verification

### No Hardcoded User Paths

Tests use example paths only for validation, not for actual file operations:

```go
// Example from metadata_test.go - just checking path construction
{
    repoPath: "/home/user/.local/share/ai-config/repo",
    wantPath: "/home/user/.local/share/ai-config/repo/.metadata/commands/test-cmd-metadata.json",
}
```

Actual file operations always use `t.TempDir()`:

```go
// Actual test that writes files
tmpDir := t.TempDir()  // Isolated temporary directory
manager := repo.NewManagerWithPath(tmpDir)
manager.AddCommand(...)  // Writes to tmpDir, not user's repo
```

### Running Tests Safely

You can verify test isolation by running tests and checking that no files are created in `~/.local/share/ai-config/`:

```bash
# Before tests
ls -la ~/.local/share/ai-config/

# Run all tests
make test

# After tests - should be unchanged
ls -la ~/.local/share/ai-config/
```

## Parallel Test Execution

Tests can run safely in parallel because each test has its own isolated directory:

```bash
# Run tests in parallel
go test -v -parallel 4 ./pkg/...
go test -v -parallel 4 ./test/...
```

## Best Practices for New Tests

When writing new tests:

1. **Always use `t.TempDir()`** for temporary directories
2. **Use `NewManagerWithPath()`** instead of `NewManager()` in tests
3. **Never hardcode paths** to `~/.local/share/ai-config/`
4. **Use helper functions** like `createTestRepo()` for consistency
5. **Verify cleanup** by checking that no files remain after tests

## Example: Adding a New Test

```go
func TestNewFeature(t *testing.T) {
    // 1. Create isolated temp directory
    tmpDir := t.TempDir()
    repoPath := filepath.Join(tmpDir, "repo")
    
    // 2. Create isolated manager
    manager := repo.NewManagerWithPath(repoPath)
    
    // 3. Create test fixtures in temp directory
    testCmd := filepath.Join(tmpDir, "test-cmd.md")
    os.WriteFile(testCmd, []byte("---\ndescription: Test\n---\n"), 0644)
    
    // 4. Perform test operations
    err := manager.AddCommand(testCmd, "file://"+testCmd, "file")
    if err != nil {
        t.Fatalf("AddCommand failed: %v", err)
    }
    
    // 5. Verify results
    res, err := manager.Get("test-cmd", resource.Command)
    if err != nil {
        t.Errorf("Get failed: %v", err)
    }
    
    // 6. No cleanup needed - automatic!
}
```

## Troubleshooting

### Test Fails with "Permission Denied"

Check that the test is using `t.TempDir()` and not trying to write to system directories.

### Test Leaves Files Behind

Verify that:
1. The test uses `t.TempDir()` (not manual `os.MkdirTemp()`)
2. The test doesn't use `defer os.RemoveAll()` (unnecessary with `t.TempDir()`)
3. The test doesn't write outside the temporary directory

### Tests Fail When Run in Parallel

Ensure each test uses its own isolated temporary directory and doesn't share global state.

## Summary

All tests in ai-config-manager use isolated temporary repositories that:

- ✅ Never write to the user's actual repository
- ✅ Automatically clean up after themselves
- ✅ Can run safely in parallel
- ✅ Provide realistic testing environments
- ✅ Prevent metadata pollution

This ensures that running tests is always safe and never affects the user's actual aimgr installation.
