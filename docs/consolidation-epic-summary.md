# ✅ EPIC CREATED: Consolidate Duplicate Code in repo_import.go

## Epic Structure

**Epic ID:** ai-config-manager-6ptj
**Priority:** P1
**Status:** Open

---

## 📋 Tasks Created

### Task 1: Consolidate add*File Functions
**ID:** ai-config-manager-455n
**Priority:** P1
**Status:** Ready to work
**Effort:** 1-2 hours

**What it does:**
- Merges `addCommandFile()` and `addAgentFile()` into single `addResourceFile()`
- Removes 66 lines of duplicate code
- Updates 4 callsites
- Low risk, pure refactoring

---

### Task 2: Investigate find*File Functions
**ID:** ai-config-manager-m5zt
**Priority:** P1
**Status:** Ready to work
**Effort:** 2-3 hours

**What it does:**
- Checks if `findCommandFile`, `findSkillDir`, `findAgentFile`, `findPackageFile` are still called
- If dead code: DELETE (~160 lines removed)
- If still used: Consolidate or document why needed
- Medium risk, requires investigation

---

### Task 3: Evaluate LoadXResource Duplication
**ID:** ai-config-manager-exxz
**Priority:** P3 (Low)
**Status:** Ready to work
**Effort:** 1 hour

**What it does:**
- Evaluates whether LoadXResource duplication is worth fixing
- Only 9 uses total, works correctly
- Documents decision with rationale
- Low risk, just evaluation and documentation

---

### Task 4: Epic Acceptance
**ID:** ai-config-manager-ihpu
**Priority:** P1
**Status:** Blocked (waiting on tasks 1-3)

**What it does:**
- Verifies all tasks complete
- Runs full test suite
- Manual testing of import workflows
- Documents line count reduction
- Gates epic closure

---

## 🎯 Dependency Structure

```
Epic: ai-config-manager-6ptj
  ↓ depends on
Task: ai-config-manager-ihpu (Acceptance)
  ↓ depends on (all 3)
  ├─ Task: ai-config-manager-455n (add*File consolidation)
  ├─ Task: ai-config-manager-m5zt (find*File investigation)
  └─ Task: ai-config-manager-exxz (LoadXResource evaluation)
```

---

## 🚀 Ready to Work Now

```bash
bd ready
```

**Output:**
1. ai-config-manager-455n - Consolidate add*File ✅ Ready
2. ai-config-manager-m5zt - Investigate find*File ✅ Ready
3. ai-config-manager-exxz - Evaluate LoadXResource ✅ Ready

**All 3 tasks can be worked in parallel!**

---

## 📊 Expected Impact

| Item | Current | After | Savings |
|------|---------|-------|---------|
| add*File duplication | 66 lines | 0 lines | -66 lines |
| find*File functions | ~160 lines | 0-160 lines | Up to -160 lines |
| LoadXResource | Stays | Stays | 0 (decision: leave as-is) |
| **Total** | **226 lines** | **0-160 lines** | **-66 to -226 lines** |

---

## 🎬 Recommended Execution Order

### Option A: Sequential (Safer)
1. Start with Task 1 (add*File) - lowest risk, clear win
2. Then Task 2 (find*File) - investigate before touching
3. Finally Task 3 (LoadXResource) - just evaluation
4. Run acceptance verification

### Option B: Parallel (Faster)
- Task 1 & 2 can be done in parallel (different code areas)
- Task 3 can be done anytime (just evaluation)
- All merge before acceptance

---

## 📝 Testing Strategy

After EACH task:
```bash
# Run unit tests
make test

# Run integration tests  
make test-integration

# Test specific workflows
aimgr repo import ~/.opencode/commands/test.md
aimgr repo import ~/.opencode/agents/test.md
aimgr repo import --force commands/existing.md
```

After ALL tasks (acceptance):
```bash
# Full regression testing
make test
make test-integration

# Manual workflow testing
aimgr repo import ~/.opencode/
aimgr repo update command/test
```

---

## ✅ Success Criteria

Epic can be closed when:
- [ ] All 3 tasks completed
- [ ] Acceptance task verified
- [ ] All tests pass
- [ ] No regressions
- [ ] Code reduction documented (target: -66 to -226 lines)
- [ ] CHANGELOG.md updated

---

## 🤔 Next Steps

**To start work:**
```bash
# Pick a task
bd show ai-config-manager-455n  # Read details
bd update ai-config-manager-455n --status=in_progress  # Claim it
# Do the work
bd close ai-config-manager-455n  # Mark complete
```

**To check progress:**
```bash
bd stats
bd blocked
bd ready
```

**To close epic:**
```bash
# After all tasks done and acceptance passed
bd close ai-config-manager-6ptj --reason="All consolidation complete, -XX lines removed"
```
