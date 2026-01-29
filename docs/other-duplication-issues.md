# 🎯 OTHER DUPLICATION ISSUES TO FIX (Beyond LoadCommand)

Based on analysis of the codebase, here are the OTHER duplication issues that need fixing:

---

## 1. 🔴 **CRITICAL: addCommandFile vs addAgentFile Duplication**

**Location:** `cmd/repo_import.go:176-241`

**Problem:** Two functions doing IDENTICAL logic, just with different resource types:

```go
// Lines 176-207: addCommandFile
func addCommandFile(filePath string, res *resource.Resource, manager *repo.Manager) error {
    // Check if already exists (if not force mode)
    if !forceFlag {
        existing, _ := manager.Get(res.Name, resource.Command)
        if existing != nil {
            return fmt.Errorf("command '%s' already exists...", res.Name)
        }
    } else {
        _ = manager.Remove(res.Name, resource.Command)
    }
    
    // Determine source info
    absPath, err := filepath.Abs(filePath)
    sourceURL := "file://" + absPath
    
    // Add the command
    if err := manager.AddCommand(filePath, sourceURL, "file"); err != nil {
        return fmt.Errorf("failed to add command: %w", err)
    }
    
    // Success message
    fmt.Printf("✓ Added command '%s' to repository\n", res.Name)
    return nil
}

// Lines 210-241: addAgentFile - EXACT SAME LOGIC!
func addAgentFile(filePath string, res *resource.Resource, manager *repo.Manager) error {
    // Check if already exists (if not force mode)
    if !forceFlag {
        existing, _ := manager.Get(res.Name, resource.Agent)
        if existing != nil {
            return fmt.Errorf("agent '%s' already exists...", res.Name)
        }
    } else {
        _ = manager.Remove(res.Name, resource.Agent)
    }
    
    // Determine source info
    absPath, err := filepath.Abs(filePath)
    sourceURL := "file://" + absPath
    
    // Add the agent
    if err := manager.AddAgent(filePath, sourceURL, "file"); err != nil {
        return fmt.Errorf("failed to add agent: %w", err)
    }
    
    // Success message
    fmt.Printf("✓ Added agent '%s' to repository\n", res.Name)
    return nil
}
```

**The ONLY differences:**
1. `resource.Command` vs `resource.Agent`
2. `manager.AddCommand` vs `manager.AddAgent`
3. `"command"` vs `"agent"` in messages

**Impact:** 
- 66 lines of duplicate code
- If you fix a bug in one, must remember to fix in the other
- No addSkillDir equivalent (skills handled differently)

---

### 🔧 **FIX: Consolidate into Single Function**

```go
// Replace BOTH functions with single generic one:
func addResourceFile(filePath string, res *resource.Resource, resType resource.ResourceType, manager *repo.Manager) error {
    // Check if already exists (if not force mode)
    if !forceFlag {
        existing, _ := manager.Get(res.Name, resType)
        if existing != nil {
            return fmt.Errorf("%s '%s' already exists in repository (use --force to overwrite)", 
                resType.String(), res.Name)
        }
    } else {
        // Remove existing if force mode
        _ = manager.Remove(res.Name, resType)
    }
    
    // Determine source info
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return fmt.Errorf("failed to get absolute path: %w", err)
    }
    sourceURL := "file://" + absPath
    
    // Add the resource based on type
    var addErr error
    switch resType {
    case resource.Command:
        addErr = manager.AddCommand(filePath, sourceURL, "file")
    case resource.Agent:
        addErr = manager.AddAgent(filePath, sourceURL, "file")
    case resource.Skill:
        addErr = manager.AddSkill(filePath, sourceURL, "file")
    default:
        return fmt.Errorf("unsupported resource type: %s", resType)
    }
    
    if addErr != nil {
        return fmt.Errorf("failed to add %s: %w", resType.String(), addErr)
    }
    
    // Success message
    fmt.Printf("✓ Added %s '%s' to repository\n", resType.String(), res.Name)
    if res.Description != "" {
        fmt.Printf("  Description: %s\n", res.Description)
    }
    
    return nil
}

// Update callsites (4 places):
// Line 140: return addResourceFile(filePath, agent, resource.Agent, manager)
// Line 149: return addResourceFile(filePath, cmd, resource.Command, manager)
// Line 160: return addResourceFile(filePath, agent, resource.Agent, manager)
// Line 168: return addResourceFile(filePath, cmd, resource.Command, manager)
```

**Result:**
- 66 duplicate lines → 1 generic function
- Easier to maintain
- Consistent behavior across all resource types
- Can easily add Skill support

---

## 2. 🟡 **MEDIUM: find*File/Dir Functions in repo_import.go**

**Location:** `cmd/repo_import.go:720-876`

**Problem:** Four similar functions with same pattern:

```go
// Line 720: findCommandFile(searchPath, name) - walks directory, matches name
// Line 758: findSkillDir(searchPath, name) - walks directory, matches name
// Line 798: findAgentFile(searchPath, name) - walks directory, matches name
// Line 836: findPackageFile(searchPath, name) - walks directory, matches name
// Line 879: findResourceInPath(sourcePath, resType, resName) - dispatches to above
```

**Total:** ~160 lines of similar filepath.Walk logic

**Question:** According to bug ai-config-manager-ksbt, these were supposed to be REMOVED because import now uses discovered .Path directly. Let me check if they're still used:

```bash
# Are these functions still called?
grep -n "findCommandFile\|findSkillDir\|findAgentFile\|findPackageFile" cmd/repo_import.go | grep -v "^[0-9]*:func"
```

**If they're still used:** Consolidate into single generic findResource function
**If they're NOT used:** DELETE them (leftover from old implementation)

**Action:** Verify usage and either consolidate or remove

---

## 3. ✅ **EVALUATED & DECIDED: LoadCommandResource vs LoadCommand Duplication**

**Status:** ✅ **LEAVE AS-IS** (Option C) - Decision documented 2026-01-29  
**Issue:** ai-config-manager-exxz  
**Documentation:** See `docs/loadxresource-evaluation.md` for complete analysis

**Location:** `pkg/resource/command.go`, `skill.go`, `agent.go`

**Problem:** Two sets of Load functions per resource type:

```go
// Pattern 1: Load minimal Resource
LoadCommand(path) → *Resource        // 38 uses
LoadSkill(path) → *Resource          // 26 uses
LoadAgent(path) → *Resource          // 38 uses

// Pattern 2: Load full *TypeResource with content
LoadCommandResource(path) → *CommandResource   // Only 4 uses (3 in repo_show.go, 1 in repo_import.go)
LoadSkillResource(path) → *SkillResource       // Only 2 uses (repo_show.go, tests)
LoadAgentResource(path) → *AgentResource       // Only 4 uses (repo_show.go, repo_import.go, tests)
```

**The Duplication:**
- Frontmatter parsed TWICE (once in LoadCommand, again in LoadCommandResource)
- Name calculation duplicated
- Relative path calculation duplicated
- Validation duplicated

---

### ✅ **DECISION: LEAVE AS-IS (Option C)**

**Rationale:**
1. **Small scale:** Only 4 production callsites affected (3 in repo_show.go, 1 in repo_import.go)
2. **Not broken:** No bugs, no performance issues (negligible ~4ms overhead on interactive commands)
3. **Legitimate pattern:** Clear separation between common case (LoadX, 38 uses) and special needs (LoadXResource, 4 uses)
4. **Alternatives add complexity:** Both Option A and B make code harder to understand without meaningful benefit
5. **All callsites have legitimate needs:**
   - `repo_show.go`: Display type-specific fields (Agent, Model, Capabilities, etc.)
   - `repo_import.go`: Distinguish agents from commands during type detection

**Code comments added** to explain this design choice in:
- `pkg/resource/command.go:LoadCommandResource`
- `pkg/resource/skill.go:LoadSkillResource`
- `pkg/resource/agent.go:LoadAgentResource`

**When to revisit:**
- v2.0 refactor: Consider as part of larger resource API redesign
- If usage patterns change (if LoadXResource becomes heavily used >20 callsites)
- If performance matters (if used in hot paths - not currently)
- If bugs occur (none so far)

**Action:** ✅ NO FIX NEEDED - Working as intended

---

## 4. 🟢 **LOW: Discovery Pattern Duplication**

**Location:** `pkg/discovery/commands.go`, `skills.go`, `agents.go`

**Problem:** Same pattern repeated 3 times:
1. Priority location search
2. Recursive fallback search
3. Deduplication

**Total Lines:**
- commands.go: 314 lines
- skills.go: 265 lines
- agents.go: 345 lines
- **Total: 924 lines** doing similar things

**Analysis:**
- Each resource type has slightly different search logic
- Commands: .md files in commands/
- Skills: directories with SKILL.md in skills/
- Agents: .md files in agents/
- Different enough that generic version would be complex

**Your Decision:** "keep this" - Don't touch discovery, it works

**Action:** SKIP - Not fixing per user request

---

## 📋 **PRIORITY ACTION PLAN**

### ✅ **High Priority: Fix These**

#### 1. Consolidate addCommandFile + addAgentFile (1-2 hours)
- **Impact:** Remove 66 lines of duplicate code
- **Risk:** Low (just refactoring, same behavior)
- **Files:** cmd/repo_import.go
- **Testing:** Run `aimgr repo import` for commands and agents

#### 2. Verify find*File Functions Still Needed (30 mins)
- **Check:** Are they still called or leftover from old code?
- **If unused:** Delete ~160 lines
- **If used:** Consolidate into single findResource()
- **Files:** cmd/repo_import.go

### ⏸️ **Low Priority: Consider Later**

#### 3. LoadCommandResource Duplication
- **Impact:** Low (only 9 uses, works correctly)
- **Effort:** Medium (would need API redesign)
- **Recommendation:** Leave for now, revisit in v2.0

#### 4. Discovery Pattern Duplication
- **Your decision:** Keep as-is
- **Action:** SKIP

---

## 🎯 **IMMEDIATE NEXT STEPS**

If you want me to proceed:

1. **Fix #1: Consolidate add*File functions**
   - Create single addResourceFile()
   - Update 4 callsites
   - Test commands and agents
   - **Result:** -66 lines, cleaner code

2. **Fix #2: Check find*File functions**
   - grep for usage
   - If unused: delete them
   - If used: consolidate or document why needed
   - **Result:** Either -160 lines or better understanding

**Ready to start? Or want more details on any of these?**
