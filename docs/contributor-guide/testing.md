# Testing Guide

Start with [docs/TESTING.md](../TESTING.md) for command selection, minimum checks, and repo-wide test rules. This file covers test-authoring patterns for contributors changing Go code.

## Where tests live

- `cmd/*_test.go` and `pkg/**/*_test.go` without `//go:build integration` hold fast unit coverage run by `make unit-test`.
- Integration coverage is split between `cmd/**/*_test.go` and `pkg/**/*_test.go` files tagged with `//go:build integration` and higher-level CLI/integration coverage under `test/`.
- End-to-end coverage lives under `test/e2e/` and uses the `e2e` build tag.
- Reusable fixtures usually live under package-local `testdata/` directories such as `pkg/discovery/testdata/`.

## Useful commands while authoring tests

```bash
# Fast package-level checks while editing
go test -run TestLoad_ValidConfig -v ./pkg/config
go test -run TestLoadCommand -v ./pkg/resource

# Repo-standard suites
make unit-test
make integration-test
make e2e-test
```

## Core test-authoring rules

- Use `t.TempDir()` for temp data and repo state.
- Use `repo.NewManagerWithPath(...)` instead of `NewManager()`.
- Keep tests isolated from `~/.local/share/ai-config/`.
- Prefer table-driven coverage for validation and parser logic.
- Put network-dependent or slower checks in the repo's integration/e2e layers instead of the fast unit loop.

## Basic isolation pattern

```go
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    manager := repo.NewManagerWithPath(tmpDir)

    // ... perform test operations against manager ...
}
```

## Writing New Tests

### Basic Structure

```go
func TestNewFeature(t *testing.T) {
    // 1. Create isolated temp directory
    tmpDir := t.TempDir()
    repoPath := filepath.Join(tmpDir, "repo")
    
    // 2. Create isolated manager
    manager := repo.NewManagerWithPath(repoPath)
    
    // 3. Create test fixtures in a real commands/ directory
    commandsDir := filepath.Join(tmpDir, "commands")
    os.MkdirAll(commandsDir, 0755)
    testCmd := filepath.Join(commandsDir, "test-cmd.md")
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

### Integration or E2E tests

Use the existing build tags when a test belongs in a slower suite:

```go
//go:build integration

package test

func TestGitClone(t *testing.T) {
    // Can use real network operations
}

//go:build e2e

package e2e

func TestCLIWorkflow(t *testing.T) {
    // Exercises the built binary end to end
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

## Related docs

- [docs/TESTING.md](../TESTING.md) - command selection, minimum checks, concurrency, and atomic-write expectations
- [docs/CODING.md](../CODING.md) - repo safety rules and locally built binary guidance
