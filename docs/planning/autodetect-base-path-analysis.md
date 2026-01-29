# Analysis: Auto-Detect Base Path for LoadCommand

## Current Situation

### The Problem
Commands can be loaded in two ways that return DIFFERENT names for the SAME file:

```go
// Way 1: Without context - returns basename only
cmd := LoadCommand("repo/commands/opencode-coder/doctor.md")
// cmd.Name = "doctor" ❌

// Way 2: With context - returns nested path
cmd := LoadCommandWithBase("repo/commands/opencode-coder/doctor.md", "repo/commands")
// cmd.Name = "opencode-coder/doctor" ✅
```

**This is fundamentally broken** - a file's identity should not depend on which load function you call.

### Current Usage of LoadCommand (without Base)

Found **17 callsites** that use `LoadCommand` directly:

| File | Line | Context | Issue |
|------|------|---------|-------|
| **pkg/repo/manager.go** | 755 | `importResource()` | ❌ **BUG**: Nested commands get wrong name |
| **cmd/repo_import.go** | 145 | Direct file import | ⚠️ May be wrong |
| **cmd/repo_import.go** | 165 | Auto-detect type | ⚠️ May be wrong |
| pkg/install/installer.go | 330 | Loading installed file | ✅ OK (installed files are flat) |
| pkg/resource/command.go | 157 | ValidateCommand | ✅ OK (just validation) |
| pkg/resource/resource.go | 33 | LoadResource | ⚠️ Delegates, may be wrong |
| pkg/resource/resource.go | 98 | DetectType | ✅ OK (just type checking) |
| Tests | various | Test fixtures | ✅ OK (tests don't care about nested) |

**Confirmed bugs:**
1. ❌ `pkg/repo/manager.go:755` - `importResource()` - used by bulk operations

---

## Proposed Solution: Auto-Detect Base Path

### Assertive Approach

**Make `LoadCommand` auto-detect the base path and THROW HARD if file is not in a valid commands/ structure:**

```go
func LoadCommand(filePath string) (*Resource, error) {
    basePath := autoDetectCommandsBase(filePath)
    
    if basePath == "" {
        return nil, fmt.Errorf("command file must be in a 'commands/' directory: %s", filePath)
    }
    
    return LoadCommandWithBase(filePath, basePath)
}

func autoDetectCommandsBase(filePath string) string {
    // Clean and normalize path
    cleanPath := filepath.Clean(filePath)
    
    // Walk up the path looking for a "commands" directory
    dir := filepath.Dir(cleanPath)
    for {
        // Check if current directory is named "commands"
        if filepath.Base(dir) == "commands" {
            return dir
        }
        
        // Check if parent is "commands" (we're one level nested)
        parent := filepath.Dir(dir)
        if filepath.Base(parent) == "commands" {
            return parent
        }
        
        // Stop at filesystem root or after reasonable depth
        if dir == "." || dir == "/" || strings.Count(dir, string(filepath.Separator)) < 2 {
            break
        }
        
        dir = parent
    }
    
    return "" // Not found
}
```

### Examples

```go
// ✅ Valid - in commands/ directory
LoadCommand("/repo/commands/test.md")
// basePath = "/repo/commands"
// name = "test"

// ✅ Valid - in nested commands/
LoadCommand("/repo/commands/opencode-coder/doctor.md")  
// basePath = "/repo/commands"
// name = "opencode-coder/doctor"

// ✅ Valid - in tool installation
LoadCommand("/project/.claude/commands/test.md")
// basePath = "/project/.claude/commands"
// name = "test"

// ❌ HARD ERROR - not in commands/
LoadCommand("/tmp/random-file.md")
// Error: "command file must be in a 'commands/' directory: /tmp/random-file.md"

// ❌ HARD ERROR - in wrong directory
LoadCommand("/repo/agents/something.md")
// Error: "command file must be in a 'commands/' directory: /repo/agents/something.md"
```

---

## Impact Analysis

### Files That Need Checking

All 17 LoadCommand callsites need review:

#### 1. pkg/repo/manager.go:755 - importResource()

**Current:**
```go
case resource.Command:
    res, err = resource.LoadCommand(sourcePath)
```

**After auto-detect:** ✅ **WILL WORK**
- importResource is called from bulk import with discovered resources
- Discovery already finds commands in `commands/` directories
- sourcePath will be like `repo/commands/name.md` or `repo/commands/nested/name.md`
- Auto-detect will find base path correctly

**Test case:** Import commands from various directories

#### 2. cmd/repo_import.go:145 - Direct file import

**Current:**
```go
if parentDir == "commands" {
    cmd, err := resource.LoadCommand(filePath)
```

**After auto-detect:** ✅ **WILL WORK**
- Already checking that parent is "commands"
- filePath will be in commands/ directory
- Auto-detect will succeed

**Edge case:** What if user does `aimgr repo import /tmp/test.md`?
- **Current behavior:** Imports with name "test"
- **After auto-detect:** ❌ Error "must be in commands/ directory"
- **Question:** Is this acceptable?

#### 3. cmd/repo_import.go:165 - Auto-detect type

**Current:**
```go
cmd, cmdErr := resource.LoadCommand(filePath)
if cmdErr == nil {
    return addCommandFile(filePath, cmd, manager)
}
```

**After auto-detect:** ⚠️ **MIGHT BREAK**
- Used when user imports arbitrary file: `aimgr repo import /path/to/some-file.md`
- If file is not in commands/ directory, auto-detect will error
- But `addCommandFile` might copy it to commands/ anyway?

**Need to check:** What is the expected behavior for ad-hoc file imports?

#### 4. pkg/install/installer.go:330 - Loading installed file

**Current:**
```go
res, err := resource.LoadCommand(target)
```

**After auto-detect:** ✅ **WILL WORK**
- target is in `.claude/commands/` or `.opencode/commands/`
- Auto-detect will find these directories
- Will return correct name (flat, since installations are flat)

#### 5. pkg/resource/resource.go:33 - LoadResource()

**Current:**
```go
case Command:
    return LoadCommand(path)
```

**After auto-detect:** ⚠️ **DEPENDS ON CALLER**
- LoadResource is generic - used by various callers
- Need to trace all LoadResource callers
- If any caller passes non-commands/ paths, will break

#### 6. pkg/resource/resource.go:98 - DetectType()

**Current:**
```go
if _, cmdErr := LoadCommand(path); cmdErr == nil {
    return Command, nil
}
```

**After auto-detect:** ⚠️ **WILL BE STRICTER**
- Currently: any .md file could be detected as command
- After: only .md files in commands/ directories
- **This is actually GOOD** - more precise detection

---

## Test Plan

### Unit Tests

```go
func TestLoadCommand_AutoDetect(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        wantName string
        wantErr  bool
    }{
        {
            name:     "flat command",
            path:     "testdata/repo/commands/test.md",
            wantName: "test",
            wantErr:  false,
        },
        {
            name:     "nested command",
            path:     "testdata/repo/commands/api/deploy.md",
            wantName: "api/deploy",
            wantErr:  false,
        },
        {
            name:     "deeply nested command",
            path:     "testdata/repo/commands/dt/cluster/overview.md",
            wantName: "dt/cluster/overview",
            wantErr:  false,
        },
        {
            name:    "not in commands directory",
            path:    "/tmp/test.md",
            wantErr: true,
        },
        {
            name:    "in wrong directory",
            path:    "testdata/repo/agents/test.md",
            wantErr: true,
        },
        {
            name:     "tool installation directory",
            path:     "project/.claude/commands/test.md",
            wantName: "test",
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            res, err := LoadCommand(tt.path)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("LoadCommand() expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("LoadCommand() unexpected error: %v", err)
                return
            }
            
            if res.Name != tt.wantName {
                t.Errorf("LoadCommand() name = %q, want %q", res.Name, tt.wantName)
            }
        })
    }
}
```

### Integration Tests

1. **repo import from directory**
   ```bash
   aimgr repo import ~/.opencode/
   # Should discover and import nested commands correctly
   ```

2. **repo import single file NOT in commands/**
   ```bash
   aimgr repo import /tmp/test.md
   # Should FAIL with clear error
   ```

---

## Breaking Changes

### Potential Issues

1. **Ad-hoc file imports fail**
   ```bash
   # Used to work (maybe?)
   aimgr repo import /tmp/some-command.md
   
   # Will now fail
   Error: command file must be in a 'commands/' directory
   ```
   
   **Mitigation:** Document that commands must be in proper structure

2. **Tests that use temporary files**
   - Any test that creates commands in `/tmp/` without proper structure
   - Need to update test fixtures

3. **Scripts that generate commands dynamically**
   - Must place generated files in `commands/` directory
   - Shouldn't affect most users

---

## Skills and Agents

### Current State
- **Skills:** Only use `LoadSkill()` - no WithBase variant exists
- **Agents:** Only use `LoadAgent()` - no WithBase variant exists  
- **Nested structure:** Currently NO nested skills or agents in repos

### Should We Add Auto-Detect for Them?

**Skills:**
```go
func LoadSkill(dirPath string) (*Resource, error) {
    // Currently: name = filepath.Base(dirPath)
    
    // Could auto-detect: find nearest "skills/" directory
    // But: no nested skills exist yet, so not urgent
}
```

**Agents:**
```go
func LoadAgent(filePath string) (*Resource, error) {
    // Currently: name = basename of file
    
    // Could auto-detect: find nearest "agents/" directory
    // But: no nested agents exist yet, so not urgent
}
```

**Recommendation:** Start with commands only (they have the problem NOW), add for skills/agents when needed.

---

## Recommendation

### ✅ GO FOR IT - But with caution:

1. **Implement auto-detect for commands**
   - Assert that file is in `commands/` directory
   - Throw clear error if not
   - This fixes the current bugs

2. **Update all affected code**
   - Review all 17 callsites
   - Fix any that assume wrong behavior
   - Update tests

3. **Deprecate LoadCommandWithBase**
   - Add deprecation comment
   - Remove it in next major version
   - Single way to load = less confusion

4. **Document the requirement**
   - Commands must be in `commands/` directories
   - Clear error messages guide users
   - Update AGENTS.md and README

5. **Comprehensive testing**
   - Unit tests for auto-detect logic
   - Integration tests for all workflows
   - Test error cases thoroughly

### Implementation Checklist

- [ ] Implement `autoDetectCommandsBase()` helper
- [ ] Update `LoadCommand()` to use auto-detect
- [ ] Test all 17 callsites one by one
- [ ] Fix any broken callsites
- [ ] Update all tests to use proper structure
- [ ] Add new tests for auto-detect logic
- [ ] Run full test suite
- [ ] Test manually: import, update, sync, install
- [ ] Update documentation
- [ ] Mark LoadCommandWithBase as deprecated
- [ ] Create CHANGELOG entry

### Risk Level: MEDIUM

**Why medium:**
- Core function used everywhere
- Potential breaking changes
- But: fixes real bugs and makes API simpler

**Mitigation:**
- Thorough testing before merge
- Clear error messages guide users
- Backwards compatible for properly structured repos

---

## Alternative: Just Fix findCommandFile

**Quick fix:** Only update `findCommandFile` to use `LoadCommandWithBase`

**Pros:**
- Minimal change
- Fixes immediate bug
- Low risk

**Cons:**
- Doesn't fix `importResource` bug
- Dual API remains confusing
- Technical debt stays

**Verdict:** Not recommended. Fix it properly now, not later.
