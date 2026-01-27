# Test Refactoring Guide

> **Status**: Inventory Phase (2026-01-27)  
> **Purpose**: Document all tests making external calls to enable systematic refactoring  
> **Goal**: Reduce test execution time from ~3 minutes to <30 seconds

---

## Table of Contents
- [Executive Summary](#executive-summary)
- [Quick Reference](#quick-reference)
- [Detailed Test Inventory](#detailed-test-inventory)
- [External Resources Used](#external-resources-used)
- [Refactoring Strategy](#refactoring-strategy)
- [Implementation Roadmap](#implementation-roadmap)

---

## Executive Summary

### Current State
- **Total test files**: 60
- **Tests with external calls**: 47+ tests across 6 files
- **Network-dependent execution time**: ~2.4-3.4 minutes
- **Primary bottleneck**: Cloning real GitHub repositories during tests

### Target State
- **Fast unit tests**: <5 seconds (no external calls)
- **Integration tests**: <30 seconds (local Git repos)
- **E2E smoke tests**: <2 minutes (real GitHub repos, optional in CI)

### Impact
- **10-50x speedup** for tests using local repos instead of network clones
- **Faster developer feedback loop** (seconds instead of minutes)
- **Reliable CI/CD** (no network flakiness)

---

## Quick Reference

### Test Files by Category

| Category | Files | Tests | Est. Time | Priority |
|----------|-------|-------|-----------|----------|
| **Critical Integration** | 2 | 9 | ~30s | Keep as-is |
| **Can Optimize** | 6 | 38+ | ~2-3min | **HIGH** |
| **Fast Unit Tests** | 52 | ~85+ | <10s | No change |

### Files to Refactor (Priority Order)

1. 🔥 **test/workspace_cache_test.go** (8 tests, ~56s)
2. 🔥 **test/github_sources_test.go** (11 tests, ~45-90s)
3. 🔥 **test/workspace_add_sync_test.go** (6 tests, ~30s)
4. 🔴 **test/repo_update_batching_test.go** (6 tests, ~15-25s)
5. 🟡 **test/cli_integration_test.go** (depends on CLI commands tested)

---

## Detailed Test Inventory

### Critical Integration Tests (Keep Real External Calls)

These tests validate actual Git operations and network connectivity - core features that must be tested against real services.

#### 1. `pkg/workspace/manager_integration_test.go`

**Build tag**: `//go:build integration`

| Test Name | External Calls | Purpose | Keep? |
|-----------|----------------|---------|-------|
| `TestGetOrClone_Integration` | `https://github.com/hk9890/ai-config-manager-test-repo` | Validate GetOrClone with real Git | ✅ Yes |
| `TestUpdate_Integration` | Same repo | Validate Update (git pull) | ✅ Yes |
| `TestListCached_WithCaches` | Same repo | Validate cache listing | ✅ Yes |
| `TestRemove` | Same repo | Validate cache removal | ✅ Yes |

**Action**: Keep as-is. These are true integration tests validating core Git functionality.

---

#### 2. `pkg/source/git_integration_test.go`

**Build tag**: `//go:build integration`

| Test Name | External Calls | Purpose | Keep? |
|-----------|----------------|---------|-------|
| `TestCloneRepo/clone public repo` | `https://github.com/github/gitignore` | Clone without ref | ✅ Yes |
| `TestCloneRepo/with branch ref` | Same repo | Clone with branch | ✅ Yes |
| `TestCloneRepo/invalid repo URL` | `https://github.com/nonexistent/...` | Error handling | ✅ Yes |
| `TestCloneRepo/invalid branch` | `github/gitignore` + bad branch | Error handling | ✅ Yes |

**Action**: Keep as-is. Core Git functionality must be tested with real repos.

---

### Can Be Optimized (High Priority Refactoring)

#### 3. `test/workspace_cache_test.go` 🔥

**Estimated time**: ~56 seconds (8 tests × ~7s each)  
**Primary bottleneck**: Each test clones `gh:anthropics/anthropic-quickstarts`

| Test Name | External Calls | Why Needed | Refactor Strategy |
|-----------|----------------|------------|-------------------|
| `TestWorkspaceCacheFirstUpdate` | Clone anthropics/anthropic-quickstarts | Test first-time cache creation | 🔧 Use local Git repo |
| `TestWorkspaceCacheSubsequentUpdate` | Clone + update same repo | Test cache reuse | 🔧 Use local Git repo |
| `TestWorkspaceCacheCorruption` | Clone repo, corrupt .git | Test corruption detection | 🔧 Use local Git repo |
| `TestWorkspacePrune` | Clone 2 repos (anthropics/quickstarts + hk9890/ai-config-manager) | Test pruning unreferenced | 🔧 Use 2 local Git repos |
| `TestWorkspacePruneDryRun` | Clone anthropics/quickstarts | Test dry-run mode | 🔧 Use local Git repo |
| `TestWorkspaceCacheConcurrent` | Clone repo, concurrent access | Test concurrent operations | 🔧 Use local Git repo |
| `TestWorkspaceCacheMetadata` | Clone anthropics/quickstarts | Test metadata tracking | 🔧 Use local Git repo |
| `TestWorkspaceCacheEmptyRef` | Clone anthropics/skills | Test default branch handling | 🔧 Use local Git repo |

**Refactoring approach**:
```go
// Before (slow):
githubSource := "gh:anthropics/anthropic-quickstarts"
parsed, _ := source.ParseSource(githubSource)
// ... clone from GitHub (~5-7s)

// After (fast):
localRepo := setupLocalTestRepo(t) // Creates local Git repo in testdata
// ... use local repo (~0.1s)
```

**Shared helper function needed**:
```go
// test/testutil/git_helpers.go
func CreateLocalGitRepo(t *testing.T, name string) string {
    // Create temp dir with initialized Git repo
    // Add some test files and commits
    // Return path to repo
}
```

---

#### 4. `test/github_sources_test.go` 🔥

**Estimated time**: ~45-90 seconds (11 tests)  
**Primary bottleneck**: Multiple clones of `anthropics/anthropic-quickstarts` with different refs/subpaths

| Test Name | External Calls | Purpose | Refactor Strategy |
|-----------|----------------|---------|-------------------|
| `TestGitHubSourceSkillDiscovery` | Clone anthropics/quickstarts, discover skills | Test discovery from cloned repo | 🔧 Use local repo with test skills |
| `TestGitHubSourceCommandDiscovery` | Clone anthropics/quickstarts, discover commands | Test command discovery | 🔧 Use local repo with test commands |
| `TestGitHubSourceAgentDiscovery` | Clone anthropics/quickstarts, discover agents | Test agent discovery | 🔧 Use local repo with test agents |
| `TestGitHubSourceWithSubpath` | Clone with subpath `/computer-use-demo` | Test subpath handling | 🔧 Use local repo with subdirs |
| `TestGitHubSourceWithBranch` | Clone with `@main` branch | Test branch reference | 🔧 Use local repo with branches |
| `TestGitHubSourceErrorHandling/invalid repo` | Try clone nonexistent repo | Test error handling | ✅ Keep (tests real Git errors) |
| `TestGitHubSourceErrorHandling/invalid branch` | Clone with bad branch | Test error handling | ✅ Keep (tests real Git errors) |
| `TestGitHubSourceErrorHandling/cleanup security` | Test cleanup safety | Test security checks | ✅ Keep (no network) |
| `TestLocalSourceStillWorks` | Local file operations only | Test local sources | ✅ Keep (no network) |
| `TestGitHubURLFormats` | Parsing only (has network skip) | Test URL parsing | ✅ Keep (no network) |
| `TestCleanupOnError` | Clone + cleanup | Test cleanup on success | 🔧 Use local repo |

**Special note**: `isOnline()` check:
```go
func isOnline() bool {
    conn, err := net.DialTimeout("tcp", "github.com:443", 2*time.Second)
    // ...
}
```
**Refactor**: Replace with `skipIfNoGit(t)` helper that only checks `git --version`

---

#### 5. `test/workspace_add_sync_test.go` 🔥

**Estimated time**: ~30 seconds (7 tests)  
**Primary bottleneck**: Clones `anthropics/skills` and `hk9890/ai-tools`

| Test Name | External Calls | Purpose | Refactor Strategy |
|-----------|----------------|---------|-------------------|
| `TestWorkspaceCacheWithRepoAdd` | Clone anthropics/skills | Test cache on repo add | 🔧 Use local Git repo |
| `TestWorkspaceCacheMetadataAfterAdd` | Clone anthropics/skills | Test metadata creation | 🔧 Use local Git repo |
| `TestWorkspaceCacheWithDifferentRefs` | Clone same repo twice | Test ref handling | 🔧 Use local Git repo |
| `TestWorkspaceCacheWithRepoSync` | Clone 2 repos (skills + ai-tools) | Test multi-repo sync | 🔧 Use 2 local Git repos |
| `TestWorkspaceCacheUpdateAfterAdd` | Clone + update skills | Test update after add | 🔧 Use local Git repo |
| `TestWorkspaceCacheWithLocalSource` | Local source only | Test local vs Git | ✅ Keep (no network) |
| `TestWorkspaceCacheSyncPullsLatest` | Clone + update skills | Test sync pulls changes | 🔧 Use local Git repo |

**Pattern to refactor**:
```go
// Before:
testURL := "https://github.com/anthropics/skills"
cachePath, err := workspaceManager.GetOrClone(testURL, testRef)

// After:
testURL := setupLocalGitRepoURL(t, "test-skills-repo")
cachePath, err := workspaceManager.GetOrClone(testURL, testRef)
```

---

#### 6. `test/repo_update_batching_test.go` 🔴

**Estimated time**: ~15-25 seconds (6 tests)  
**Primary bottleneck**: Uses `anthropics/quickstarts` and `openai/cookbook` for batching tests

| Test Name | External Calls | Purpose | Refactor Strategy |
|-----------|----------------|---------|-------------------|
| `TestUpdateBatching` | Metadata only (gh:anthropics/quickstarts) | Test batching setup | ✅ Keep (no actual clone) |
| `TestUpdateBatching_MixedSources` | Metadata only | Test mixed Git/local | ✅ Keep (no actual clone) |
| `TestUpdateBatching_MultipleResourceTypes` | Metadata only | Test batching across types | ✅ Keep (no actual clone) |
| `TestUpdateBatching_DryRun` | Metadata only | Test dry-run flag | ✅ Keep (no actual clone) |
| `TestUpdateBatching_VerifyGrouping` | Metadata only | Test grouping logic | ✅ Keep (no actual clone) |
| `TestCLIUpdateBatching_LocalSources` | Local sources + CLI | Test CLI update | ✅ Keep (no network) |

**Note**: These tests mostly set up metadata without actually cloning. May not need refactoring if they're already fast.

**Verification needed**: Run `go test -v ./test -run TestUpdateBatching` to measure actual time.

---

#### 7. `test/cli_integration_test.go`

**External calls**: `exec.Command(binPath, args...)` - depends on which CLI commands are tested

**Action**: Audit which commands use network, refactor those to use local repos or fixtures.

---

## External Resources Used

### GitHub Repositories (Frequency of Use)

| Repository | Used In | Frequency | Purpose | Replace With |
|------------|---------|-----------|---------|--------------|
| `gh:anthropics/anthropic-quickstarts` | workspace_cache, github_sources, repo_update_batching | ~30+ calls | General testing, caching, parsing | Local Git repo with test structure |
| `gh:anthropics/skills` | workspace_add_sync, workspace_cache | ~10+ calls | Skill discovery, sync testing | Local Git repo with skill dirs |
| `gh:hk9890/ai-config-manager` | workspace_cache, workspace_add_sync | ~5 calls | Multi-repo testing | Local Git repo #2 |
| `https://github.com/hk9890/ai-config-manager-test-repo` | pkg/workspace/manager_integration_test.go | 4 tests | Integration testing | ✅ Keep for integration tests |
| `https://github.com/github/gitignore` | pkg/source/git_integration_test.go | 2 calls | Clone testing | ✅ Keep for integration tests |
| `gh:openai/openai-cookbook` | repo_update_batching | 1 call | Grouping tests | Local Git repo #3 (if needed) |
| `gh:hk9890/ai-tools` | workspace_add_sync | 1 call | Multi-repo sync | Local Git repo #4 (if needed) |

### Network Checks

| Check | Implementation | Used In | Refactor To |
|-------|----------------|---------|-------------|
| `isOnline()` | `net.DialTimeout("tcp", "github.com:443", ...)` | workspace_cache, github_sources (16+ places) | `skipIfNoGit(t)` - only check git binary |
| `isGitAvailable()` | `exec.Command("git", "--version").Run()` | repo_update_batching, git_test, workspace tests | ✅ Keep |

---

## Refactoring Strategy

### Phase 1: Create Test Infrastructure (Week 1)

#### 1.1 Create Test Utilities Package

**File**: `test/testutil/git_helpers.go`

```go
package testutil

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// CreateLocalGitRepo creates a local Git repository for testing
// Returns the absolute path to the repo
func CreateLocalGitRepo(t *testing.T, name string) string {
    t.Helper()
    
    repoDir := filepath.Join(t.TempDir(), name)
    if err := os.MkdirAll(repoDir, 0755); err != nil {
        t.Fatalf("Failed to create repo dir: %v", err)
    }
    
    // Initialize Git repo
    runGit(t, repoDir, "init")
    runGit(t, repoDir, "config", "user.email", "test@example.com")
    runGit(t, repoDir, "config", "user.name", "Test User")
    
    // Create initial commit
    readmePath := filepath.Join(repoDir, "README.md")
    if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err != nil {
        t.Fatalf("Failed to create README: %v", err)
    }
    
    runGit(t, repoDir, "add", "README.md")
    runGit(t, repoDir, "commit", "-m", "Initial commit")
    
    return repoDir
}

// CreateLocalGitRepoWithSkills creates a Git repo with test skill structure
func CreateLocalGitRepoWithSkills(t *testing.T, name string, skillNames ...string) string {
    t.Helper()
    
    repoDir := CreateLocalGitRepo(t, name)
    
    skillsDir := filepath.Join(repoDir, "skills")
    if err := os.MkdirAll(skillsDir, 0755); err != nil {
        t.Fatalf("Failed to create skills dir: %v", err)
    }
    
    for _, skillName := range skillNames {
        skillDir := filepath.Join(skillsDir, skillName)
        if err := os.MkdirAll(skillDir, 0755); err != nil {
            t.Fatalf("Failed to create skill dir: %v", err)
        }
        
        skillContent := fmt.Sprintf(`---
name: %s
description: Test skill for %s
---
# %s

Test skill content.
`, skillName, skillName, skillName)
        
        skillMdPath := filepath.Join(skillDir, "SKILL.md")
        if err := os.WriteFile(skillMdPath, []byte(skillContent), 0644); err != nil {
            t.Fatalf("Failed to create SKILL.md: %v", err)
        }
    }
    
    runGit(t, repoDir, "add", ".")
    runGit(t, repoDir, "commit", "-m", "Add test skills")
    
    return repoDir
}

// CreateLocalGitRepoWithBranches creates a repo with multiple branches
func CreateLocalGitRepoWithBranches(t *testing.T, name string, branches ...string) string {
    t.Helper()
    
    repoDir := CreateLocalGitRepo(t, name)
    
    for _, branch := range branches {
        runGit(t, repoDir, "checkout", "-b", branch)
        
        // Create a file specific to this branch
        branchFile := filepath.Join(repoDir, fmt.Sprintf("%s.txt", branch))
        content := fmt.Sprintf("Content for branch %s\n", branch)
        if err := os.WriteFile(branchFile, []byte(content), 0644); err != nil {
            t.Fatalf("Failed to create branch file: %v", err)
        }
        
        runGit(t, repoDir, "add", ".")
        runGit(t, repoDir, "commit", "-m", fmt.Sprintf("Add content for %s", branch))
    }
    
    // Return to main/master
    runGit(t, repoDir, "checkout", "main")
    
    return repoDir
}

// CreateLocalGitRepoWithSubdirs creates a repo with subdirectory structure
func CreateLocalGitRepoWithSubdirs(t *testing.T, name string, subdirs ...string) string {
    t.Helper()
    
    repoDir := CreateLocalGitRepo(t, name)
    
    for _, subdir := range subdirs {
        subdirPath := filepath.Join(repoDir, subdir)
        if err := os.MkdirAll(subdirPath, 0755); err != nil {
            t.Fatalf("Failed to create subdir: %v", err)
        }
        
        // Add a file in the subdir
        filePath := filepath.Join(subdirPath, "README.md")
        content := fmt.Sprintf("# %s\n\nTest content for subdir.\n", subdir)
        if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
            t.Fatalf("Failed to create subdir file: %v", err)
        }
    }
    
    runGit(t, repoDir, "add", ".")
    runGit(t, repoDir, "commit", "-m", "Add subdirectories")
    
    return repoDir
}

// runGit is a helper to run git commands
func runGit(t *testing.T, dir string, args ...string) {
    t.Helper()
    
    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    if output, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("Git command failed: git %v\nOutput: %s\nError: %v", 
            args, string(output), err)
    }
}

// SkipIfNoGit skips the test if git is not available
func SkipIfNoGit(t *testing.T) {
    t.Helper()
    
    if err := exec.Command("git", "--version").Run(); err != nil {
        t.Skip("git not available, skipping test")
    }
}
```

#### 1.2 Create Static Test Fixtures

**File**: `testdata/repos/README.md`

```markdown
# Test Repositories

This directory contains static Git repositories used for testing.
These are committed to the repository to avoid network calls during tests.

## Structure

- `quickstarts-minimal/` - Minimal version of anthropics/quickstarts
- `skills-test/` - Test repository with skill structures
- `multi-resource/` - Repository with commands, skills, and agents
```

**Action**: Create minimal static Git repos in `testdata/repos/` and commit them.

---

### Phase 2: Refactor Tests (Week 2-3)

#### 2.1 Refactor `test/workspace_cache_test.go`

**Before**:
```go
func TestWorkspaceCacheFirstUpdate(t *testing.T) {
    if !isOnline() {
        t.Skip("Skipping test: network not available")
    }
    
    githubSource := "gh:anthropics/anthropic-quickstarts"
    parsed, err := source.ParseSource(githubSource)
    // ... ~7s clone from GitHub
}
```

**After**:
```go
func TestWorkspaceCacheFirstUpdate(t *testing.T) {
    testutil.SkipIfNoGit(t)
    
    // Create local test repo (~0.1s)
    localRepo := testutil.CreateLocalGitRepo(t, "test-quickstarts")
    githubSource := "file://" + localRepo
    parsed, err := source.ParseSource(githubSource)
    // ... fast local operations
}
```

**Estimated speedup**: 7s → 0.1s per test (70x faster)

#### 2.2 Refactor `test/github_sources_test.go`

Focus on discovery tests:

```go
func TestGitHubSourceSkillDiscovery(t *testing.T) {
    testutil.SkipIfNoGit(t)
    
    // Create local repo with test skills
    localRepo := testutil.CreateLocalGitRepoWithSkills(t, 
        "skills-repo", "test-skill-1", "test-skill-2")
    
    // Test discovery from local repo instead of cloning from GitHub
    skills, err := discovery.DiscoverSkills(localRepo, "")
    // ... assertions
}
```

**Keep these tests as-is** (they test real Git errors):
- `TestGitHubSourceErrorHandling/invalid repo`
- `TestGitHubSourceErrorHandling/invalid branch`

#### 2.3 Refactor `test/workspace_add_sync_test.go`

```go
func TestWorkspaceCacheWithRepoAdd(t *testing.T) {
    testutil.SkipIfNoGit(t)
    
    localRepo := testutil.CreateLocalGitRepoWithSkills(t, "skills-test", "skill1")
    testURL := "file://" + localRepo
    
    // Rest of test uses local repo instead of GitHub
    cachePath, err := workspaceManager.GetOrClone(testURL, "main")
    // ... fast operations
}
```

---

### Phase 3: Separate Test Types (Week 3)

#### 3.1 Add Build Tags

**Fast unit tests** (default):
```go
// No build tag - runs by default
package workspace

func TestNormalizeURL(t *testing.T) {
    // Fast unit test, no external calls
}
```

**Integration tests** (opt-in):
```go
//go:build integration

package workspace

func TestGetOrClone_Integration(t *testing.T) {
    // Real GitHub clone, network required
}
```

**E2E tests** (opt-in):
```go
//go:build e2e

package test

func TestFullWorkflowWithRealGitHub(t *testing.T) {
    // End-to-end with real GitHub repos
}
```

#### 3.2 Update Makefile

```makefile
# Run fast unit tests only (default)
test:
	go test -v ./...

# Run unit + integration tests
test-integration:
	go test -v -tags=integration ./...

# Run all tests including E2E
test-all:
	go test -v -tags=integration,e2e ./...

# Run with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

---

### Phase 4: Add Test Caching (Week 4)

#### 4.1 CI/CD Optimization

**GitHub Actions** (`.github/workflows/test.yml`):

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      # Run fast unit tests (no network)
      - name: Run unit tests
        run: make test
      
      # Upload coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      # Run integration tests (with local repos)
      - name: Run integration tests
        run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    # Only run E2E on main branch or release tags
    if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      # Run E2E tests with real GitHub repos (slow)
      - name: Run E2E tests
        run: make test-all
```

---

## Implementation Roadmap

### Week 1: Infrastructure
- [ ] Create `test/testutil/git_helpers.go` with helper functions
- [ ] Create `testdata/repos/` with minimal static Git repos
- [ ] Add `make test`, `make test-integration`, `make test-all` targets
- [ ] Document new test patterns in README

### Week 2: Refactor High-Impact Tests
- [ ] Refactor `test/workspace_cache_test.go` (8 tests)
- [ ] Refactor `test/github_sources_test.go` (discovery tests)
- [ ] Verify speedup: run `time make test` before and after

### Week 3: Complete Refactoring
- [ ] Refactor `test/workspace_add_sync_test.go` (7 tests)
- [ ] Add build tags to integration tests
- [ ] Audit `test/cli_integration_test.go` for optimization
- [ ] Update documentation

### Week 4: CI/CD & Validation
- [ ] Update GitHub Actions workflow
- [ ] Add test execution time benchmarks
- [ ] Validate all tests pass with new structure
- [ ] Measure and document performance improvements

### Success Metrics
- ✅ **Unit tests**: <5 seconds total
- ✅ **Integration tests**: <30 seconds total
- ✅ **E2E tests**: <2 minutes (optional in CI)
- ✅ **CI/CD feedback loop**: <1 minute for unit tests
- ✅ **Zero flaky tests** from network issues

---

## Appendix

### Test Execution Time Benchmarks

**Baseline** (before refactoring):
```bash
# Run all tests
$ time go test -v ./test ./pkg/...
# Expected: ~3-4 minutes

# Run just slow tests
$ time go test -v ./test -run "Cache|GitHub"
# Expected: ~2.5 minutes
```

**After refactoring** (target):
```bash
# Run unit tests (default)
$ time make test
# Target: <5 seconds

# Run unit + integration tests
$ time make test-integration
# Target: <30 seconds

# Run all tests including E2E
$ time make test-all
# Target: <2 minutes
```

### Helper Function Reference

| Function | Purpose | Use In |
|----------|---------|--------|
| `testutil.CreateLocalGitRepo(t, name)` | Basic Git repo | All cache tests |
| `testutil.CreateLocalGitRepoWithSkills(t, name, skills...)` | Repo with skills | Discovery tests |
| `testutil.CreateLocalGitRepoWithBranches(t, name, branches...)` | Multi-branch repo | Branch tests |
| `testutil.CreateLocalGitRepoWithSubdirs(t, name, subdirs...)` | Repo with subdirs | Subpath tests |
| `testutil.SkipIfNoGit(t)` | Skip if no Git | All Git tests |

### Files NOT Needing Refactoring

These test files are already fast (no external calls):

- `pkg/config/config_test.go`
- `pkg/discovery/commands_test.go`, `skills_test.go`, `agents_test.go`, `packages_test.go`
- `pkg/errors/types_test.go`
- `pkg/install/installer_test.go`
- `pkg/manifest/manifest_test.go`
- `pkg/marketplace/parser_test.go`, `generator_test.go`, `discovery_test.go`
- `pkg/metadata/metadata_test.go`
- `pkg/output/formatter_test.go`
- `pkg/pattern/matcher_test.go`
- `pkg/repo/manager_test.go`, `package_test.go`
- `pkg/resource/*.go` (all resource tests)
- `pkg/tools/types_test.go`
- `pkg/workspace/manager_test.go` (unit tests only)
- `cmd/*.go` (most cmd tests)

---

**Last Updated**: 2026-01-27  
**Related Issues**: ai-config-manager-ed2t, ai-config-manager-eaze  
**Next Review**: After Phase 1 completion
