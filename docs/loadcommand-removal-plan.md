# 🎯 REVISED PLAN: Fix LoadCommand Duplication (REMOVE, Don't Deprecate)

## Executive Summary

**Goal:** Fix LoadCommand dual-API by making LoadCommand auto-detect base path and **REMOVE** LoadCommandWithBase entirely.

**Scope:** Complete the LoadCommand epic (ai-config-manager-ss4d) with NO deprecation cycle - clean removal.

**Timeline:** 1-2 days with systematic testing.

---

## ✅ USER DECISIONS

Based on your feedback:

1. ❌ **NO Skills/Agents base path support** - Don't fix what isn't broken yet
2. ✅ **KEEP discovery as-is** - Priority + recursive works, leave it alone
3. ✅ **Enforce structure** - Files must be in proper directories (details below)
4. ✅ **REMOVE LoadCommandWithBase** - No deprecation, clean removal

---

## 🔥 WHAT EXACTLY WE'RE DOING

### Phase 1: Review All 17 LoadCommand Callsites (ai-config-manager-5qg4)

**Goal:** Ensure every place that calls LoadCommand will work with auto-detect.

**The Change:**
```go
// BEFORE (current behavior after auto-detect was added):
LoadCommand(path) → Auto-detects commands/ directory, returns nested path

// AFTER (same, but we verify it works everywhere):
LoadCommand(path) → Auto-detects commands/ directory, returns nested path
```

**17 Callsites to Review:**

| File | Line | Function | Will It Work? | Action Needed |
|------|------|----------|---------------|---------------|
| pkg/repo/manager.go | 754 | importResource() | ✅ YES - sourcePath from discovery | None |
| cmd/repo_import.go | 145 | Direct file import | ✅ YES - checks parent="commands" | None |
| cmd/repo_import.go | 165 | Auto-detect type | ⚠️ MAYBE - ad-hoc imports | Test & document error |
| cmd/repo_import.go | 738 | findCommandFile() | ✅ YES - fixes the bug! | Verify fix works |
| pkg/install/installer.go | 330 | Load installed file | ✅ YES - in .claude/commands/ | None |
| pkg/resource/command.go | 157 | ValidateCommand | ✅ YES - just validation | None |
| pkg/resource/resource.go | 33 | LoadResource | ✅ YES - delegates to LoadCommand | None |
| pkg/resource/resource.go | 98 | DetectType | ✅ YES - type detection | None |
| Tests | various | Test fixtures | ✅ YES - already in commands/ | Verify all pass |

**Specific Actions:**

1. **cmd/repo_import.go:165** - Test ad-hoc import behavior:
   ```bash
   # This will now FAIL (by design):
   aimgr repo import /tmp/random-file.md
   # Error: "command file must be in a 'commands/' directory"
   
   # This still WORKS:
   aimgr repo import ~/project/commands/test.md
   ```
   **Decision:** Document that ad-hoc imports must use proper structure.

2. **cmd/repo_import.go:738** - Verify findCommandFile fix:
   ```bash
   # This was BROKEN before, should work now:
   aimgr repo update command/opencode-coder/doctor
   ```

3. **All tests** - Run full suite and fix any that fail:
   ```bash
   make test
   make test-integration
   ```

---

### Phase 2: Remove LoadCommandWithBase Entirely

**Goal:** Delete the WithBase function and update all code that uses it.

**Current Usage of LoadCommandWithBase:**

| File | Line | Context | Replacement |
|------|------|---------|-------------|
| pkg/repo/manager.go | 146 | addCommand() | Use LoadCommand (auto-detects) |
| pkg/repo/manager.go | 376 | AddBulk commands | Use LoadCommand |
| pkg/repo/manager.go | 533 | Get() command | Use LoadCommand |
| pkg/discovery/commands.go | 149 | searchCommandsInDirectory | Use LoadCommand |
| pkg/discovery/commands.go | 264 | recursiveSearchCommands | Use LoadCommand |
| pkg/install/nested_install_test.go | 49 | Test nested install | Use LoadCommand |

**Total: 6 callsites to update**

**The Changes:**

1. **pkg/resource/command.go** - DELETE LoadCommandWithBase:
   ```go
   // DELETE THIS ENTIRE FUNCTION:
   func LoadCommandWithBase(filePath string, basePath string) (*Resource, error) {
       // 40+ lines of code - DELETE ALL
   }
   
   // DELETE THIS TOO:
   func LoadCommandResourceWithBase(filePath string, basePath string) (*CommandResource, error) {
       // DELETE ALL
   }
   ```

2. **pkg/repo/manager.go:146** - Update addCommand():
   ```go
   // BEFORE:
   res, err := resource.LoadCommandWithBase(sourcePath, basePath)
   
   // AFTER:
   res, err := resource.LoadCommand(sourcePath)
   ```

3. **pkg/repo/manager.go:376** - Update AddBulk():
   ```go
   // BEFORE:
   res, err := resource.LoadCommandWithBase(path, commandsPath)
   
   // AFTER:
   res, err := resource.LoadCommand(path)
   ```

4. **pkg/repo/manager.go:533** - Update Get():
   ```go
   // BEFORE:
   return resource.LoadCommandWithBase(path, commandsPath)
   
   // AFTER:
   return resource.LoadCommand(path)
   ```

5. **pkg/discovery/commands.go:149 & 264** - Update discovery:
   ```go
   // BEFORE (both locations):
   cmd, err := resource.LoadCommandWithBase(entryPath, basePath)
   
   // AFTER:
   cmd, err := resource.LoadCommand(entryPath)
   ```

6. **pkg/install/nested_install_test.go:49** - Update test:
   ```go
   // BEFORE:
   res, err := resource.LoadCommandWithBase(expectedRepoPath, filepath.Join(repoPath, "commands"))
   
   // AFTER:
   res, err := resource.LoadCommand(expectedRepoPath)
   ```

---

### Phase 3: Update Test Fixtures (ai-config-manager-j70m)

**Goal:** Ensure all test fixtures use proper commands/ directory structure.

**Action:** Review testdata/ directories and ensure structure like:
```
testdata/
  repos/
    test-repo/
      commands/           ✅ Proper structure
        test.md
        nested/
          command.md
      /tmp/test.md        ❌ Would fail (by design)
```

**Commands to run:**
```bash
# Find any test files that might use improper structure
grep -r "LoadCommand" test/ pkg/ --include="*_test.go" | grep -v "commands/"

# Fix any found issues
```

---

### Phase 4: Add Integration Tests (ai-config-manager-1dlz)

**Goal:** Comprehensive tests covering all workflows with nested commands.

**Tests to Add:**

1. **Nested command import:**
   ```go
   // Test: Import nested command from local directory
   // Verify: Name is "opencode-coder/doctor", not "doctor"
   ```

2. **Nested command update:**
   ```go
   // Test: Update nested command (was broken, now fixed)
   // Verify: findCommandFile matches correctly
   ```

3. **Nested command install:**
   ```go
   // Test: Install nested command
   // Verify: Symlink created correctly
   ```

4. **Auto-detect error handling:**
   ```go
   // Test: Load command NOT in commands/ directory
   // Verify: Clear error message returned
   ```

---

### Phase 5: Verification & Documentation

**Goal:** Ensure everything works and is properly documented.

**Testing Checklist:**
```bash
# Unit tests
make test

# Integration tests
make test-integration

# Manual verification
aimgr repo import ~/.opencode              # Nested commands work
aimgr repo update command/nested/test      # Update works
aimgr repo sync                            # Sync works
aimgr install command/nested/test          # Install works
aimgr repo import /tmp/test.md             # Fails with clear error (expected)
```

**Documentation Updates:**

1. **AGENTS.md** - Update "Common Patterns" section:
   ```markdown
   ### Loading Resources
   ```go
   // Commands: Auto-detects base path from commands/ directory
   res, err := resource.LoadCommand("path/to/commands/nested/file.md")
   // Returns: name = "nested/file"
   
   // Commands MUST be in commands/ directory:
   res, err := resource.LoadCommand("/tmp/random.md")
   // Error: "command file must be in a 'commands/' directory"
   ```

2. **CHANGELOG.md** - Add entry:
   ```markdown
   ## [v1.14.0] - 2026-01-XX
   
   ### Changed
   - **BREAKING:** LoadCommand now requires files to be in commands/ directory
   - LoadCommand auto-detects base path for nested structure
   
   ### Removed
   - **BREAKING:** LoadCommandWithBase removed (use LoadCommand instead)
   - **BREAKING:** LoadCommandResourceWithBase removed
   
   ### Fixed
   - Nested commands work correctly in `repo update`
   - Command identity is consistent regardless of load context
   ```

3. **docs/autodetect-base-path-analysis.md** - Add completion note:
   ```markdown
   ## Implementation Status
   
   ✅ COMPLETED - 2026-01-28
   - Auto-detect implemented
   - LoadCommandWithBase REMOVED (not deprecated)
   - All callsites updated
   - Tests pass
   ```

---

## 🚨 BREAKING CHANGES EXPLAINED

### What Will Break?

1. **Ad-hoc imports from random locations:**
   ```bash
   # BEFORE (maybe worked?):
   aimgr repo import /tmp/my-command.md
   
   # AFTER (fails):
   Error: command file must be in a 'commands/' directory
   
   # WORKAROUND:
   mkdir -p commands && mv /tmp/my-command.md commands/
   aimgr repo import commands/my-command.md
   ```

2. **External code using LoadCommandWithBase:**
   ```go
   // BEFORE:
   cmd, err := resource.LoadCommandWithBase(path, base)
   
   // AFTER (compile error - function doesn't exist):
   // Fix: Use LoadCommand instead (auto-detects base)
   cmd, err := resource.LoadCommand(path)
   ```

3. **Tests with improper structure:**
   ```go
   // BEFORE (might have worked):
   tmpFile := filepath.Join(t.TempDir(), "test.md")
   cmd, err := resource.LoadCommand(tmpFile)
   
   // AFTER (fails):
   // Fix: Use proper structure
   tmpDir := t.TempDir()
   commandsDir := filepath.Join(tmpDir, "commands")
   os.MkdirAll(commandsDir, 0755)
   tmpFile := filepath.Join(commandsDir, "test.md")
   cmd, err := resource.LoadCommand(tmpFile)
   ```

### Who Is Affected?

1. **Internal code:** Already updated in this plan
2. **Integration tests:** Already use proper structure (mostly)
3. **End users:** Only if they were doing ad-hoc imports
4. **External consumers:** Only if they import the package (unlikely)

### Migration Guide

If someone is affected:
1. Ensure commands are in `commands/` directories
2. Replace `LoadCommandWithBase` with `LoadCommand`
3. That's it - auto-detect handles the rest

---

## ✅ SUCCESS CRITERIA

After completion:

1. **No LoadCommandWithBase exists** - Function completely removed
2. **No LoadCommandResourceWithBase exists** - Also removed
3. **All tests pass** - Unit + integration
4. **Manual testing works** - Import, update, sync, install
5. **Clear errors** - Ad-hoc imports fail with helpful message
6. **Documentation updated** - AGENTS.md, CHANGELOG.md, analysis doc
7. **Epic closed** - ai-config-manager-ss4d marked complete

---

## 📊 IMPLEMENTATION SUMMARY

| Task | Files Changed | Lines Changed | Risk |
|------|--------------|---------------|------|
| Review 17 callsites | 0 (verification only) | 0 | Low |
| Remove WithBase functions | pkg/resource/command.go | -80 lines | Low |
| Update 6 callsites | pkg/repo/, pkg/discovery/ | ~12 lines | Low |
| Update test fixtures | testdata/ | ~10 files | Low |
| Add integration tests | test/ | +100 lines | None |
| Update docs | docs/, AGENTS.md | +50 lines | None |
| **TOTAL** | **~15 files** | **+50, -80** | **Low** |

**Net result:** Code is SIMPLER (fewer functions, single way to load commands).

---

## 🚀 EXECUTION PLAN

### Day 1 Morning
- ✅ Review 17 callsites (Phase 1)
- ✅ Test ad-hoc import behavior
- ✅ Verify findCommandFile fix
- ✅ Run existing test suite

### Day 1 Afternoon
- ✅ Remove LoadCommandWithBase functions (Phase 2)
- ✅ Update 6 callsites that use WithBase
- ✅ Run tests after each change
- ✅ Fix any broken tests

### Day 2 Morning
- ✅ Update test fixtures (Phase 3)
- ✅ Add integration tests (Phase 4)
- ✅ Run full test suite (unit + integration)
- ✅ Manual testing of all workflows

### Day 2 Afternoon
- ✅ Update documentation (Phase 5)
- ✅ Final verification
- ✅ Close epic and all tasks
- ✅ Create git tag v1.14.0

---

## 🤔 FINAL CONFIRMATION

Before I start, confirm:

1. ✅ Remove LoadCommandWithBase entirely (NO deprecation)
2. ✅ Breaking change is acceptable (enforce commands/ structure)
3. ✅ Ad-hoc imports from random locations will fail (by design)
4. ✅ Update all 6 callsites to use LoadCommand
5. ✅ Remove ~80 lines of duplicate code

**Ready to proceed?**

If yes, I'll start with Phase 1 (reviewing the 17 callsites) and work systematically through each phase with full testing at every step.
