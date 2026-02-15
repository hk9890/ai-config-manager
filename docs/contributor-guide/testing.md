# Testing Guide

Comprehensive testing approach for ai-config-manager.

## Test Types

**Unit Tests** (fast, fixtures-based):
- Located in `pkg/*/` packages alongside source code
- Use fixtures in `pkg/*/testdata/ (e.g., pkg/discovery/testdata/)`
- No network calls, `//go:build unit`
- Run with: `make unit-test`
- Execution: <5 seconds

**Integration Tests** (slower, network-dependent):
- Located in `test/` directory
- Tagged with `//go:build integration`
- Use real GitHub repositories
- Run with: `make integration-test`
- Execution: ~30 seconds

**E2E Tests** (slowest, full CLI testing):
- Located in `test/e2e/`
- Build actual binary and test commands
- Tagged with `//go:build e2e`
- Run with: `make e2e-test`
- Execution: ~1-2 minutes

## Running Tests

```bash
# Run all tests (unit + integration + e2e + vet)
make test

# Run only unit tests (fast)
make unit-test

# Run only integration tests
make integration-test

# Run only E2E tests
make e2e-test

# Run specific test
go test -v ./pkg/resource/command_test.go
go test -v ./pkg/config -run TestLoad_ValidConfig

# Run with coverage
go test -v -cover ./pkg/...
```

## Test Isolation

**All tests use isolated temporary repositories** to prevent polluting the user's actual repository at `~/.local/share/ai-config/`.

### Test Setup Pattern

```go
func TestSomething(t *testing.T) {
    // Create isolated temporary directory (auto-cleanup)
    tmpDir := t.TempDir()
    
    // Create manager with custom path (NOT NewManager())
    manager := repo.NewManagerWithPath(tmpDir)
    
    // ... perform test operations ...
    
    // No cleanup needed - t.TempDir() auto-removes tmpDir
}
```

### Benefits

1. **Automatic Cleanup**: `t.TempDir()` removes directories when tests complete
2. **Parallel Safety**: Each test gets isolated directory
3. **No User Impact**: Tests never write to `~/.local/share/ai-config/`
4. **Realistic**: Tests use same code paths as production

## Writing New Tests

### Basic Structure

```go
func TestNewFeature(t *testing.T) {
    // 1. Create isolated temp directory
    tmpDir := t.TempDir()
    repoPath := filepath.Join(tmpDir, "repo")
    
    // 2. Create isolated manager
    manager := repo.NewManagerWithPath(repoPath)
    
    // 3. Create test fixtures
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
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {name: "valid", input: "test", wantErr: false},
        {name: "invalid", input: "", wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

Use build tag for network-dependent tests:

```go
//go:build integration

package test

func TestGitClone(t *testing.T) {
    // Can use real network operations
}
```

## Best Practices

### Do's
- ✅ Always use `t.TempDir()` for temporary directories
- ✅ Use `NewManagerWithPath()` instead of `NewManager()` in tests
- ✅ Use table-driven tests for multiple scenarios
- ✅ Test both success and error cases
- ✅ Use `t.Helper()` in helper functions
- ✅ Add tests for new functionality
- ✅ Follow existing patterns

### Don'ts
- ❌ Never use `NewManager()` in tests (uses user's repo)
- ❌ Don't hardcode paths to `~/.local/share/ai-config/`
- ❌ Don't skip error checking in tests
- ❌ Don't use `os.MkdirTemp()` directly (use `t.TempDir()`)
- ❌ Don't share global state between tests
- ❌ Don't add integration tests unless necessary

### File Permissions

Use consistent permissions:
```go
os.MkdirAll(dir, 0755)   // Directories
os.WriteFile(file, data, 0644)  // Files
```

## Troubleshooting

### Test Fails with "Permission Denied"

**Problem**: Test trying to write to system directories.
**Solution**: Ensure test uses `t.TempDir()` and doesn't write outside temp directories.

### Tests Fail When Run in Parallel

**Problem**: Tests sharing state or writing to same locations.
**Solution**: Ensure each test uses isolated temporary directory.

### Integration Tests Don't Run

**Problem**: Integration tests skipped by default.
**Solution**: Use `make integration-test` explicitly.

## Summary

Key principles:
- ✅ All tests use isolated temporary directories
- ✅ Tests never write to user's repository
- ✅ Tests automatically clean up
- ✅ Tests run safely in parallel
- ✅ Unit tests use fixtures, integration tests use real repos
- ✅ Prefer unit tests over integration tests
