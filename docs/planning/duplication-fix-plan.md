# 🎯 IMMEDIATE ACTION PLAN: Fix Duplication in ai-config-manager

## Executive Summary

**Goal:** Stop the bleeding by consolidating duplicate code paths and fixing the LoadCommand dual-API problem.

**Scope:** Fix 5 major duplication issues causing the current chaos.

**Timeline:** 1-2 days if done systematically with proper testing.

---

## 🔍 DUPLICATION ISSUES IDENTIFIED

### 1. **LoadCommand Dual API** (CRITICAL - Currently Being Fixed)
**Status:** Epic ai-config-manager-ss4d in progress

**Problem:**
- `LoadCommand(path)` → returns basename only
- `LoadCommandWithBase(path, base)` → returns nested path
- Same file, different identity depending on which function you call

**Already Done:**
✅ Auto-detect implementation in LoadCommand (autoDetectCommandsBase)
✅ Analysis document written

**Still TODO:**
- [ ] Review all 17 callsites (ai-config-manager-5qg4)
- [ ] Update test fixtures (ai-config-manager-j70m)
- [ ] Add integration tests (ai-config-manager-1dlz)
- [ ] Deprecate LoadCommandWithBase
- [ ] Verify all systems working

**Impact:** 17 callsites, 2 confirmed bugs, root cause of recent chaos

---

### 2. **Skills and Agents Have NO Base Path Support** (HIGH PRIORITY)

**Problem:**
```go
// Commands: Has dual API
LoadCommand(path)          → basename only
LoadCommandWithBase(path, base) → nested path

// Skills: NO WithBase variant - WILL HAVE SAME BUG
LoadSkill(dirPath)         → basename only (filepath.Base)
// NO LoadSkillWithBase exists!

// Agents: NO WithBase variant - WILL HAVE SAME BUG  
LoadAgent(filePath)        → basename only (filepath.Base)
// NO LoadAgentWithBase exists!
```

**Why This Matters:**
- Currently NO nested skills or agents in repos (checked)
- But the SAME problem will happen when someone tries
- Skills are DIRECTORIES (like `skills/pdf-processing/`)
- Agents are FILES (like `agents/code-reviewer.md`)
- If we want `opencode-agents/code-reviewer`, LoadAgent will fail

**Observed Code:**
```go
// skill.go:43
dirName := filepath.Base(dirPath)  // ❌ Only takes last segment!

// agent.go:33
name := strings.TrimSuffix(filepath.Base(filePath), ".md")  // ❌ Only takes basename!
```

**What We Need:**
1. Skills: Auto-detect nearest `skills/` directory (same as commands)
2. Agents: Auto-detect nearest `agents/` directory (same as commands)
3. Use relative path from base as Name for nested resources

**Priority:** HIGH - Prevents future repeat of command disaster

---

### 3. **Import Logic Duplication** (ALREADY FIXED BUT INCOMPLETE)

**Status:** Bug ai-config-manager-ksbt marked CLOSED but needs verification

**Problem (before fix):**
Three different code paths for importing:
1. Tests: Use `AddBulk()` with `cmd.Path` ✅
2. Local import: Use `cmd.Path` directly ✅
3. GitHub import: Use `findCommandFile(name)` ❌

**Fix Applied:**
Commit 4fade69: "fix: consolidate import logic to use discovered .Path directly"
- GitHub import now uses `cmd.Path` like local import
- Removed need for `findCommandFile()` for commands, skills, agents

**What Still Needs Checking:**
1. ❓ Does `findPackageFile()` still exist? (packages don't have .Path)
2. ❓ Are `addCommandFile()`, `addAgentFile()` helper functions still needed?
3. ❓ Could we consolidate into single `importResource()` function?

**Files to Review:**
- `cmd/repo_import.go:175` - addCommandFile()
- `cmd/repo_import.go:210` - addAgentFile()
- `cmd/repo_import.go:738` - findCommandFile() usage (should be gone)
- `pkg/repo/manager.go:732` - importResource() function

**Action:** Verify the fix is complete and no code paths remain duplicated

---

### 4. **Discovery Logic Has Priority + Recursive Search** (MEDIUM PRIORITY)

**Problem:**
`pkg/discovery/commands.go` has TWO search strategies:
1. **Priority locations:** `commands/`, `.claude/commands/`, `.opencode/commands/`
2. **Recursive fallback:** Search everywhere else

**Code Flow:**
```go
// Line 66-78 in commands.go
priorityCommands, priorityErrors := searchPriorityLocations(searchPath)
allCommands = append(allCommands, priorityCommands...)

recursiveCommands, recursiveErrors := recursiveSearchCommands(searchPath, 0, searchPath)
allCommands = append(allCommands, recursiveCommands...)

return deduplicateCommands(allCommands), allErrors, nil
```

**Why This Is Duplication:**
- Both functions walk directory trees
- Both call `LoadCommandWithBase()`
- Need `deduplicateCommands()` because may find same command twice
- Complexity: 314 lines for commands.go + 305 lines tests

**Same Pattern Repeated:**
- `pkg/discovery/skills.go` (265 lines) - same pattern
- `pkg/discovery/agents.go` (345 lines) - same pattern
- `pkg/discovery/packages.go` (166 lines) - similar but simpler

**Questions:**
1. Can we simplify to single recursive search with priority ordering?
2. Do we need recursive fallback at all? (strict: only find in standard locations)
3. Why deduplication? Shouldn't paths be unique?

**Action:** 
- Document WHY we have both priority + recursive
- Consider simplification in future (not urgent - works now)

---

### 5. **Resource Loading Has THREE Entry Points** (LOW PRIORITY)

**Problem:**
Multiple ways to load the same resource type:

```go
// Pattern 1: Load base Resource (minimal)
LoadCommand(path) → *Resource
LoadSkill(path) → *Resource
LoadAgent(path) → *Resource
LoadPackage(path) → *Package

// Pattern 2: Load full *Resource (with details)
LoadCommandResource(path) → *CommandResource
LoadSkillResource(path) → *SkillResource
LoadAgentResource(path) → *AgentResource
// NO LoadPackageResource!

// Pattern 3: Load with base path (only commands)
LoadCommandWithBase(path, base) → *Resource
LoadCommandResourceWithBase(path, base) → *CommandResource
// NO WithBase for skills/agents!
```

**Why This Is Confusing:**
- 3 different function patterns doing similar things
- Not consistent across resource types
- Commands have 4 functions, packages have 1

**Impact:**
- Low urgency (doesn't cause bugs, just confusion)
- Could be cleaned up during refactoring
- Good candidate for interface-based approach

**Action:** Document the pattern, fix during larger refactoring

---

## 📋 RECOMMENDED EXECUTION ORDER

### **Phase 1: Complete LoadCommand Fix** (1 day)

**Tasks:**
1. ✅ ai-config-manager-5qg4: Review 17 LoadCommand callsites
2. ✅ ai-config-manager-j70m: Update test fixtures  
3. ✅ ai-config-manager-1dlz: Add integration tests
4. ✅ Mark LoadCommandWithBase as deprecated
5. ✅ Run full test suite
6. ✅ Close epic ai-config-manager-ss4d

**Acceptance Criteria:**
- All tests pass
- `repo import` works for nested commands
- No emergency reverts needed

---

### **Phase 2: Add Base Path Support for Skills & Agents** (0.5 day)

**Why Now:** Prevent future repeat of command disaster

**Tasks:**

#### 2.1 Skills
```go
// New function
func autoDetectSkillsBase(dirPath string) string {
    // Walk up looking for "skills/" directory
    // Same algorithm as autoDetectCommandsBase
}

// Update LoadSkill
func LoadSkill(dirPath string) (*Resource, error) {
    basePath := autoDetectSkillsBase(dirPath)
    if basePath == "" {
        return nil, fmt.Errorf("skill must be in a 'skills/' directory")
    }
    return LoadSkillWithBase(dirPath, basePath)
}

// Add new function
func LoadSkillWithBase(dirPath string, basePath string) (*Resource, error) {
    // Calculate relative path from basePath
    // Use as Name for nested skills
    // Otherwise same as current LoadSkill
}
```

#### 2.2 Agents
```go
// New function
func autoDetectAgentsBase(filePath string) string {
    // Walk up looking for "agents/" directory
}

// Update LoadAgent  
func LoadAgent(filePath string) (*Resource, error) {
    basePath := autoDetectAgentsBase(filePath)
    if basePath == "" {
        return nil, fmt.Errorf("agent must be in an 'agents/' directory")
    }
    return LoadAgentWithBase(filePath, basePath)
}

// Add new function
func LoadAgentWithBase(filePath string, basePath string) (*Resource, error) {
    // Calculate relative path from basePath
    // Use as Name for nested agents
}
```

**Tests:**
- Unit tests for auto-detect logic
- Test fixtures with nested skills/agents
- Integration tests for discovery

**Acceptance Criteria:**
- Skills support nested structure: `skills/opencode/typescript-helper/`
- Agents support nested structure: `agents/opencode/code-reviewer.md`
- All existing flat skills/agents still work
- No breaking changes (no nested resources exist yet)

---

### **Phase 3: Verify Import Consolidation** (0.5 day)

**Tasks:**
1. Review commit 4fade69 changes
2. Verify findCommandFile/findSkillDir/findAgentFile are gone
3. Check if addCommandFile/addAgentFile can be consolidated
4. Look for remaining duplication in cmd/repo_import.go
5. Add tests for consolidated paths

**Questions to Answer:**
- Is importResource() in manager.go now the single truth?
- Do we still need separate addCommandFile() vs addAgentFile()?
- Can we unify into: addResource(res *Resource, type ResourceType)?

**Acceptance Criteria:**
- Single code path for local + GitHub import
- No findXXX() functions for commands/skills/agents
- Tests verify both paths use same logic

---

### **Phase 4: Document Discovery Pattern** (0.25 day)

**Tasks:**
1. Add comments explaining priority + recursive search
2. Document why deduplication is needed
3. Create docs/discovery-architecture.md
4. Note for future: consider simplification

**Not Fixing Yet:**
- Discovery works, don't break it
- Simplification can wait for v2.0

**Acceptance Criteria:**
- Clear comments in discovery/*.go files
- Architecture doc explains the design
- TODOs added for future simplification

---

### **Phase 5: Deprecate LoadXWithBase Functions** (0.25 day)

**After Phase 1 & 2 complete:**

```go
// Deprecated: LoadCommand now auto-detects base path.
// This function will be removed in v2.0.0.
// Use LoadCommand(filePath) instead.
func LoadCommandWithBase(filePath string, basePath string) (*Resource, error) {
    // Keep implementation for backward compatibility
}

// Deprecated: LoadSkill now auto-detects base path.
// This function will be removed in v2.0.0.
func LoadSkillWithBase(dirPath string, basePath string) (*Resource, error) {
    // Implementation
}

// Deprecated: LoadAgent now auto-detects base path.
// This function will be removed in v2.0.0.
func LoadAgentWithBase(filePath string, basePath string) (*Resource, error) {
    // Implementation
}
```

**Update CHANGELOG.md:**
```markdown
## [v1.14.0] - 2026-01-XX

### Changed
- LoadCommand now auto-detects base path (LoadCommandWithBase deprecated)
- LoadSkill now supports nested structure (LoadSkillWithBase added, will be deprecated in v2.0)
- LoadAgent now supports nested structure (LoadAgentWithBase added, will be deprecated in v2.0)

### Deprecated
- LoadCommandWithBase: Use LoadCommand instead (removal in v2.0.0)

### Fixed
- Import logic consolidated (no more duplicate code paths)
```

---

## ✅ SUCCESS CRITERIA

### Testing Checklist
```bash
# All tests pass
make test
make test-integration

# Manual verification
aimgr repo import ~/.opencode     # Nested commands work
aimgr repo sync                   # Sync works
aimgr install skill/test          # Install works

# Verify no regressions
git bisect start
# Test at each commit - no breaking changes
```

### Documentation Checklist
- [ ] AGENTS.md updated with new patterns
- [ ] docs/discovery-architecture.md created
- [ ] CHANGELOG.md updated
- [ ] LoadXWithBase marked deprecated
- [ ] Comments added to discovery code

### Beads Checklist
- [ ] Epic ai-config-manager-ss4d closed
- [ ] All related tasks closed
- [ ] No open bugs related to nested commands
- [ ] Verification gate passed

---

## 🚨 RISK MITIGATION

### Before Starting
1. **Create feature branch:** `git checkout -b fix/consolidate-load-functions`
2. **Commit frequently:** Each phase = separate commits
3. **Run tests after each commit:** Catch regressions early

### During Implementation
1. **One resource type at a time:** Commands → Skills → Agents
2. **Test before moving on:** Full test suite must pass
3. **Manual testing:** Actually use the commands

### If Things Break
1. **Don't rush a revert:** Understand WHY it broke
2. **Fix forward if possible:** Add test, fix issue, verify
3. **Revert only if blocked:** Can't fix quickly, revert cleanly

---

## 📊 ESTIMATED EFFORT

| Phase | Estimated Time | Risk Level |
|-------|---------------|------------|
| 1. Complete LoadCommand | 1 day | Medium (in progress) |
| 2. Skills & Agents Base Path | 0.5 day | Low (no nested resources exist yet) |
| 3. Verify Import Consolidation | 0.5 day | Low (already fixed, just verify) |
| 4. Document Discovery | 0.25 day | None (documentation only) |
| 5. Deprecate WithBase | 0.25 day | None (backward compatible) |
| **TOTAL** | **2.5 days** | **Low-Medium** |

---

## 🎯 WHAT THIS FIXES

After completing this plan:

✅ **No more dual API confusion** - One way to load each resource type
✅ **No more import duplication** - Single code path for all imports  
✅ **No more nested command bugs** - Proper base path detection
✅ **No more future disasters** - Skills/agents won't repeat the problem
✅ **Clear path forward** - Deprecation cycle to v2.0.0

---

## 📝 NOTES

### What We're NOT Fixing (Yet)
- Discovery complexity (works, document it)
- Multiple Load* function variants (confusing but not buggy)
- Test infrastructure (already being refactored separately)
- Metadata system redesign (needs larger effort)

### What Comes After
Once this is stable:
1. Plan metadata system migration
2. Simplify discovery (single strategy)
3. Unify Load* function patterns
4. Extract common interfaces

---

## 🤔 OPEN QUESTIONS FOR USER

1. **Skills/Agents Base Path:** Should we add it NOW or wait until someone needs nested structure?
   - Pro: Prevents future disaster
   - Con: Changes code that currently works
   
2. **Discovery Simplification:** Should we keep priority + recursive, or go strict (only standard locations)?
   - Pro: Simpler, clearer
   - Con: May break existing repos with weird layouts
   
3. **Breaking Changes:** Are you OK with enforcing structure (must be in commands/skills/agents dirs)?
   - Pro: Clean architecture
   - Con: May break ad-hoc workflows

4. **Timeline:** Is 2-3 days acceptable, or need it faster?

---

## 🚀 READY TO START?

If approved, I'll:
1. Start with Phase 1 (complete LoadCommand epic)
2. Create beads issues for Phases 2-5
3. Execute systematically with full testing
4. Keep you updated on progress

Let me know if you want to adjust the plan!
