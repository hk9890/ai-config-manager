# End-to-End (E2E) Tests

This directory contains end-to-end tests for the `aimgr` CLI tool. E2E tests verify complete workflows by running the actual CLI binary against test configurations and repositories.

## Purpose

E2E tests validate:
- Complete CLI workflows (import → install → uninstall → remove)
- Integration between multiple components (config, repo, discovery, installation)
- Real-world scenarios with actual Git repositories and file operations
- Cross-tool functionality (Claude, OpenCode, Copilot)
- Config file handling and environment variable overrides

Unlike unit or integration tests that may use mocks or fixtures, E2E tests run the actual CLI binary with real configurations.

## Directory Structure

```
test/e2e/
├── README.md                    # This file
├── configs/                     # Test configuration files
│   ├── e2e-test.yaml           # Base test config
│   └── *.local.yaml            # Git-ignored local configs
├── repos/                       # Git-ignored test repositories
│   └── test-repo-1/            # Created during test runs
└── *_e2e_test.go               # E2E test files
```

## Configuration Files

### Base Config: `configs/e2e-test.yaml`

The base E2E test config includes:
- **Test repo path**: `test/e2e/repos/test-repo-1` (relative path, git-ignored)
- **Realistic sync sources**: Same sources as production config for real-world testing
- **Multiple tool targets**: Tests Claude, OpenCode, and Copilot integration

You can override this config in tests using:
```go
config := filepath.Join(testDir, "configs", "e2e-test.yaml")
cmd := exec.Command("aimgr", "--config", config, "repo", "import", sourcePath)
```

### Local Configs: `*.local.yaml`

For custom test scenarios, create `*.local.yaml` files (git-ignored):
```yaml
# configs/custom-test.local.yaml
repo:
  path: test/e2e/repos/custom-repo
sync:
  sources:
    - url: "https://github.com/your-org/test-resources"
```

## How Configs Work

### Config Loading Priority

`aimgr` loads configuration in this order (first found wins):
1. `--config` flag: `aimgr --config path/to/config.yaml`
2. `AIMGR_CONFIG_PATH` environment variable
3. `~/.config/aimgr/aimgr.yaml` (default user config)

### Path Resolution

- **Relative paths** in configs are resolved from the current working directory
- **Absolute paths** are used as-is
- **Environment variables** in paths are expanded (e.g., `$HOME/repo`)

### E2E Test Config Usage

```go
// Option 1: --config flag (recommended)
cmd := exec.Command("aimgr", "--config", "test/e2e/configs/e2e-test.yaml", "repo", "list")

// Option 2: Environment variable
cmd := exec.Command("aimgr", "repo", "list")
cmd.Env = append(os.Environ(), "AIMGR_CONFIG_PATH=test/e2e/configs/e2e-test.yaml")

// Option 3: Override repo path via env var
cmd := exec.Command("aimgr", "repo", "list")
cmd.Env = append(os.Environ(), "AIMGR_REPO_PATH=test/e2e/repos/test-repo-2")
```

## Running E2E Tests

### Run All E2E Tests

```bash
# Run with go test
go test -v ./test/e2e/... -tags=e2e

# Or with make (if configured)
make e2e-test
```

### Run Specific E2E Test

```bash
go test -v ./test/e2e -run TestE2EWorkflow -tags=e2e
```

### Test Environment

E2E tests should:
1. Use `t.TempDir()` for temporary directories when needed
2. Reference `test/e2e/configs/e2e-test.yaml` for config
3. Let tests create `test/e2e/repos/` directories as needed (git-ignored)
4. Clean up after themselves (temp dirs auto-cleaned, test repos persist for debugging)

## Writing New E2E Tests

### Test Structure

```go
//go:build e2e

package e2e

import (
    "os/exec"
    "path/filepath"
    "testing"
)

func TestMyE2EScenario(t *testing.T) {
    // 1. Setup: Prepare test data
    testDir, err := filepath.Abs(".")
    if err != nil {
        t.Fatal(err)
    }
    
    configPath := filepath.Join(testDir, "configs", "e2e-test.yaml")
    sourcePath := filepath.Join(testDir, "..", "testdata", "repos", "example")
    
    // 2. Execute: Run CLI commands
    cmd := exec.Command("aimgr", "--config", configPath, "repo", "import", sourcePath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("import failed: %v\nOutput: %s", err, output)
    }
    
    // 3. Verify: Check results
    cmd = exec.Command("aimgr", "--config", configPath, "repo", "list", "--format=json")
    output, err = cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("list failed: %v", err)
    }
    
    // Parse and assert on output
    var resources []Resource
    if err := json.Unmarshal(output, &resources); err != nil {
        t.Fatal(err)
    }
    
    if len(resources) == 0 {
        t.Error("expected resources to be imported")
    }
    
    // 4. Cleanup (optional - temp dirs auto-cleaned)
    // Test repos in test/e2e/repos/ are git-ignored and can be inspected after test
}
```

### Best Practices

1. **Use build tags**: Add `//go:build e2e` to separate E2E from unit tests
2. **Use --format=json**: Parse JSON output for assertions instead of parsing tables
3. **Test complete workflows**: Import → List → Install → Uninstall → Remove
4. **Verify file system state**: Check that symlinks, directories, and files are created correctly
5. **Test error cases**: Invalid configs, missing sources, permission errors
6. **Use realistic data**: Reference actual testdata repos from `test/testdata/repos/`
7. **Isolate tests**: Each test should use its own config or repo path to avoid conflicts

### Example Workflows to Test

- **Basic import workflow**: Import resources from directory → verify in repo
- **Sync workflow**: Configure sync sources → run sync → verify resources
- **Install workflow**: Import → install to tool directories → verify symlinks
- **Uninstall workflow**: Install → uninstall → verify symlinks removed
- **Multi-tool workflow**: Install to Claude + OpenCode simultaneously
- **Config override workflow**: Test --config flag, env vars, default config precedence
- **Error handling**: Invalid paths, missing repos, permission errors

## Debugging E2E Tests

### Inspect Test Repositories

Test repositories are git-ignored but persist after test runs:
```bash
ls -la test/e2e/repos/test-repo-1/
```

### Run with Verbose Output

```bash
go test -v ./test/e2e -run TestE2EWorkflow -tags=e2e
```

### Manual CLI Testing

Run CLI commands manually with test config:
```bash
./aimgr --config test/e2e/configs/e2e-test.yaml repo list
./aimgr --config test/e2e/configs/e2e-test.yaml repo import examples/
```

### Clean Test Artifacts

```bash
# Remove test repos (recreated on next test run)
rm -rf test/e2e/repos/

# Keep configs - they're committed and versioned
```

## CI/CD Considerations

### Local Path Sources

The base config includes a commented-out local path source:
```yaml
# - url: "/home/hans/dev/dt/copilot/knowledge-base/..."
```

This is intentional - local paths won't work in CI. For CI-compatible tests:
- Use GitHub sources (e.g., `https://github.com/hk9890/ai-tools`)
- Or copy test fixtures to `test/testdata/repos/` (committed)
- Or create `*.local.yaml` configs for local-only scenarios

### Network Dependencies

E2E tests that use `repo sync` will:
- Clone real GitHub repositories
- Use workspace caching (10-50x faster on subsequent runs)
- Require network access (may be slow on first run)

Consider:
- Mocking or skipping sync tests in CI if network is unreliable
- Using fixture repos in `test/testdata/` instead of live GitHub repos
- Setting reasonable timeouts for network operations

## References

- **Architecture**: See `docs/architecture/architecture-rules.md`
- **Config format**: See `docs/user-guide/config.md`
- **Testing strategy**: See `docs/planning/test-refactoring.md`
- **Unit tests**: See `test/*_test.go` with `//go:build unit`
- **Integration tests**: See `test/*_test.go` with `//go:build integration`
