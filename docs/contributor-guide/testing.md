# Testing Guide

This guide covers the testing approach, best practices, and procedures for the ai-config-manager project.

## Table of Contents

- [Overview](#overview)
- [Running Tests](#running-tests)
- [Test Isolation](#test-isolation)
- [Writing New Tests](#writing-new-tests)
- [Test Fixtures and Helpers](#test-fixtures-and-helpers)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Overview

The ai-config-manager project uses a comprehensive testing strategy that includes both unit tests and integration tests. All tests are designed to be isolated, fast, and safe to run in parallel.

### Test Types

**Unit Tests** (fast, fixtures-based):
- Located in `pkg/*/` packages alongside source code
- Use committed fixtures in `testdata/repos/`
- No network calls
- Run by default with `make test`
- Execution time: <5 seconds

**Integration Tests** (slower, network-dependent):
- Located in `test/` directory
- Tagged with `//go:build integration`
- Use real GitHub repositories (e.g., hk9890/ai-tools)
- Run explicitly with `make integration-test`
- Execution time: ~30 seconds

### Test Philosophy

- **Prefer unit tests with fixtures** for most functionality
- **Only add integration tests** for features requiring real Git operations
- **All tests must be isolated** to prevent interference with user data
- **Tests should be deterministic** and safe to run in parallel

---

## Running Tests

### Quick Reference

```bash
# Run all tests (unit + integration + vet)
make test

# Run only unit tests (fast)
make unit-test

# Run only integration tests (slow, requires network)
make integration-test

# Run specific test file
go test -v ./pkg/resource/command_test.go

# Run specific test by name
go test -v ./pkg/config -run TestLoad_ValidConfig

# Run with coverage report
go test -v -cover ./pkg/...

# Run tests in parallel
go test -v -parallel 4 ./pkg/...
```

### Makefile Targets

| Command | Description | Includes |
|---------|-------------|----------|
| `make test` | Run all tests | Unit + Integration + vet |
| `make unit-test` | Run only unit tests | Fast, fixture-based tests |
| `make integration-test` | Run integration tests | Network-dependent tests |
| `make fmt` | Format Go code | gofmt |
| `make vet` | Run Go vet | Static analysis |

### Test Output

Tests use Go's standard testing output:
- `PASS` - Test passed
- `FAIL` - Test failed
- `SKIP` - Test skipped (e.g., integration test without tag)

Example output:
```
=== RUN   TestAddCommand
=== RUN   TestAddCommand/success
=== RUN   TestAddCommand/duplicate
--- PASS: TestAddCommand (0.01s)
    --- PASS: TestAddCommand/success (0.00s)
    --- PASS: TestAddCommand/duplicate (0.00s)
```

---

## Test Isolation

All tests in ai-config-manager use **isolated temporary repositories** to prevent pollution of the user's actual aimgr repository at `~/.local/share/ai-config/`.

### Why Test Isolation?

Tests that create resources could potentially pollute the user's repository with:
1. Test resources interfering with actual user resources
2. Test metadata files accumulating in the user's repo
3. Orphaned files from failed tests
4. Conflicts when tests run in parallel

### How It Works

The repository manager supports custom repository paths for testing:

```go
// Production: Uses user's actual repo at ~/.local/share/ai-config/repo/
func NewManager() (*Manager, error)

// Testing: Uses custom temporary path
func NewManagerWithPath(repoPath string) *Manager
```

### Test Setup Pattern

All tests follow this standard pattern:

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

### Benefits of Test Isolation

1. **Automatic Cleanup**: `t.TempDir()` automatically removes temporary directories when tests complete
2. **Parallel Safety**: Each test gets its own isolated directory, enabling safe parallel execution
3. **No User Impact**: Tests never write to `~/.local/share/ai-config/`
4. **Realistic Testing**: Tests use the same code paths as production, just with different base paths

### Verification

You can verify test isolation by checking that no files are created in your aimgr repository:

```bash
# Before tests
ls -la ~/.local/share/ai-config/

# Run all tests
make test

# After tests - should be unchanged
ls -la ~/.local/share/ai-config/
```

### Test Coverage

**Unit Tests** (isolated):
- `pkg/repo/manager_test.go` - Repository operations
- `pkg/install/installer_test.go` - Installation operations
- `pkg/metadata/metadata_test.go` - Metadata operations
- `pkg/resource/*_test.go` - Resource loading/validation

**Integration Tests** (isolated):
- `test/integration_test.go` - Complete workflows
- `test/bulk_import_test.go` - Bulk import operations
- `test/cli_integration_test.go` - CLI command execution
- `test/bulk_add_*.go` - Bulk add operations

**Command Tests** (isolated):
- `cmd/install_test.go` - Pattern expansion and installation
- `cmd/uninstall_test.go` - Pattern-based uninstallation
- `cmd/list_installed_test.go` - Listing installed resources

---

## Writing New Tests

### Basic Structure

Follow this pattern for all new tests:

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

### Table-Driven Tests

Prefer table-driven tests for testing multiple scenarios:

```go
func TestCommandValidation(t *testing.T) {
    tests := []struct {
        name    string
        content string
        wantErr bool
    }{
        {
            name:    "valid command",
            content: "---\ndescription: Valid\n---\nContent",
            wantErr: false,
        },
        {
            name:    "missing description",
            content: "---\n---\nContent",
            wantErr: true,
        },
        {
            name:    "invalid yaml",
            content: "---\nkey: [unclosed\n---",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            testFile := filepath.Join(tmpDir, "test.md")
            os.WriteFile(testFile, []byte(tt.content), 0644)
            
            _, err := resource.LoadCommand(testFile)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadCommand() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

For tests requiring network operations, use the `integration` build tag:

```go
//go:build integration

package test

import (
    "testing"
)

func TestGitClone(t *testing.T) {
    // This test only runs when: make integration-test
    // It can use real network operations
}
```

### Error Testing

Always test both success and error cases:

```go
func TestAddCommand(t *testing.T) {
    tmpDir := t.TempDir()
    manager := repo.NewManagerWithPath(tmpDir)
    
    // Test success case
    err := manager.AddCommand(validPath, validSource, "file")
    if err != nil {
        t.Fatalf("expected success, got error: %v", err)
    }
    
    // Test duplicate error
    err = manager.AddCommand(validPath, validSource, "file")
    if err == nil {
        t.Errorf("expected duplicate error, got nil")
    }
    
    // Test invalid path error
    err = manager.AddCommand("nonexistent.md", "file://bad", "file")
    if err == nil {
        t.Errorf("expected error for nonexistent file, got nil")
    }
}
```

---

## Test Fixtures and Helpers

### Helper Functions

The test suite provides several helper functions for common test setup:

#### createTestRepo()

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

#### createTestProject()

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

### Using Test Helpers

```go
func TestInstallCommand(t *testing.T) {
    // Create test repository with resources
    repoPath, cleanup := createTestRepo(t)
    defer cleanup()
    
    // Create test project directory
    projectDir := createTestProject(t)
    
    // Test installation
    manager := repo.NewManagerWithPath(repoPath)
    installer := install.NewInstaller(manager, projectDir)
    
    err := installer.Install("command/test", tools.Claude)
    if err != nil {
        t.Fatalf("Install failed: %v", err)
    }
}
```

### Fixture Paths

Never hardcode paths to the user's actual repository:

```go
// ❌ BAD - Hardcoded user path
repoPath := "/home/user/.local/share/ai-config/repo"

// ✅ GOOD - Use temporary directory
tmpDir := t.TempDir()
repoPath := filepath.Join(tmpDir, "repo")
```

### Example Paths in Tests

Example paths are OK for validation tests, not file operations:

```go
// ✅ OK - Just checking path construction logic
{
    repoPath: "/home/user/.local/share/ai-config/repo",
    wantPath: "/home/user/.local/share/ai-config/repo/.metadata/commands/test.json",
}

// ❌ BAD - Don't actually write to example paths
os.WriteFile("/home/user/.local/share/ai-config/repo/test.md", data, 0644)
```

---

## Best Practices

### Do's

1. **Always use `t.TempDir()`** for temporary directories
2. **Use `NewManagerWithPath()`** instead of `NewManager()` in tests
3. **Use table-driven tests** for multiple scenarios
4. **Test both success and error cases**
5. **Use `t.Helper()`** in helper functions
6. **Use `defer file.Close()`** for cleanup
7. **Add tests for new functionality**
8. **Follow existing code patterns**
9. **Run `make fmt` before committing**
10. **Verify tests pass** with `make test`

### Don'ts

1. **Never hardcode paths** to `~/.local/share/ai-config/`
2. **Don't use `NewManager()`** in tests (uses user's repo)
3. **Don't skip error checking** in tests
4. **Don't use `os.MkdirTemp()`** directly (use `t.TempDir()`)
5. **Don't use `defer os.RemoveAll()`** with `t.TempDir()` (redundant)
6. **Don't write outside** temporary directories
7. **Don't assume test execution order**
8. **Don't share global state** between tests
9. **Don't use hardcoded timestamps** or paths
10. **Don't add integration tests** unless necessary

### File Permissions

Use consistent file permissions:
- Directories: `0755`
- Files: `0644`

```go
os.MkdirAll(dir, 0755)
os.WriteFile(file, data, 0644)
```

### Test Organization

Organize tests by functionality:

```
pkg/
├── resource/
│   ├── command.go
│   ├── command_test.go       # Unit tests for command.go
│   ├── skill.go
│   └── skill_test.go         # Unit tests for skill.go
└── repo/
    ├── manager.go
    └── manager_test.go       # Unit tests for manager.go

test/
├── integration_test.go       # Integration tests
└── testdata/
    └── repos/                # Test fixtures
```

---

## Troubleshooting

### Test Fails with "Permission Denied"

**Problem**: Test trying to write to system directories.

**Solution**: Ensure the test uses `t.TempDir()` and doesn't try to write outside temporary directories.

```go
// ❌ BAD
repoPath := "/usr/local/share/ai-config"

// ✅ GOOD
tmpDir := t.TempDir()
repoPath := filepath.Join(tmpDir, "repo")
```

### Test Leaves Files Behind

**Problem**: Temporary files not cleaned up after test.

**Solution**: Verify:
1. Test uses `t.TempDir()` (not manual `os.MkdirTemp()`)
2. Test doesn't use `defer os.RemoveAll()` (unnecessary with `t.TempDir()`)
3. Test doesn't write outside temporary directory

```go
// ❌ BAD - Manual cleanup needed
tmpDir, _ := os.MkdirTemp("", "test")
defer os.RemoveAll(tmpDir)

// ✅ GOOD - Automatic cleanup
tmpDir := t.TempDir()
```

### Tests Fail When Run in Parallel

**Problem**: Tests sharing state or writing to same locations.

**Solution**: Ensure each test uses its own isolated temporary directory and doesn't share global state.

```go
// ❌ BAD - Shared state
var sharedManager *repo.Manager

func TestA(t *testing.T) {
    sharedManager = repo.NewManager()
}

// ✅ GOOD - Isolated state
func TestA(t *testing.T) {
    tmpDir := t.TempDir()
    manager := repo.NewManagerWithPath(tmpDir)
}
```

### Integration Tests Don't Run

**Problem**: Integration tests skipped by default.

**Solution**: Use the integration-test target:

```bash
# ❌ Won't run integration tests
make test

# ✅ Runs integration tests
make integration-test
```

### Test Fixtures Not Found

**Problem**: Test can't find fixture files.

**Solution**: Use relative paths from the test file location:

```go
// ✅ GOOD - Relative to test file
fixtureData, err := os.ReadFile("testdata/sample-command.md")
```

---

## Summary

Key principles for testing in ai-config-manager:

- ✅ All tests use isolated temporary directories
- ✅ Tests never write to the user's repository
- ✅ Tests automatically clean up after themselves
- ✅ Tests can run safely in parallel
- ✅ Tests provide realistic testing environments
- ✅ Unit tests use fixtures, integration tests use real repos
- ✅ Always prefer unit tests over integration tests

Following these guidelines ensures that tests are:
- **Safe**: Never affect user data
- **Fast**: Most tests run in seconds
- **Reliable**: Deterministic and reproducible
- **Maintainable**: Easy to understand and modify

For more information, see:
- [Test Refactoring Documentation](../test-refactoring.md)
- [Release Process](release-process.md)
