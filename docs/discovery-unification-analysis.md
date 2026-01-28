# Discovery Recursive Functions Analysis

## Current State: 4 Duplicate Functions

### 1. recursiveSearchCommands (commands.go:167-230)
```go
func recursiveSearchCommands(currentPath string, depth int, basePath string) 
    ([]*resource.Resource, []DiscoveryError)
```
**What it does:**
- Searches for `.md` files (commands)
- Skips: `agents/`, `skills/` directories
- Skips: hidden directories (except `.claude`, `.opencode`)
- **BUG**: Parses files OUTSIDE commands/ subtree

### 2. discoverAgentsRecursive (agents.go:245-299)
```go
func discoverAgentsRecursive(dirPath string, currentDepth int) 
    ([]*resource.Resource, []DiscoveryError, error)
```
**What it does:**
- Searches for `.md` files (agents)
- Skips: `commands/`, `skills/`, `node_modules/`, `.git/`, build dirs
- **BUG**: Parses files OUTSIDE agents/ subtree

### 3. recursiveSearchSkills (skills.go:202-256)
```go
func recursiveSearchSkills(rootPath string, currentDepth int) 
    ([]*resource.Resource, []DiscoveryError, error)
```
**What it does:**
- Searches for directories with `SKILL.md` file
- Skips: hidden directories
- Stops at skill directories (doesn't recurse into them)
- **BUG**: Parses SKILL.md files OUTSIDE skills/ subtree

### 4. recursiveSearchPackages (packages.go:75-117)
```go
func recursiveSearchPackages(basePath string, depth int) 
    ([]*resource.Package, error)
```
**What it does:**
- Searches for `.package.json` files
- Skips: `agents/`, `skills/`, `commands/` directories
- Skips: hidden directories
- **BUG**: Same issue - parses files outside packages/ subtree

## Common Pattern

All 4 functions follow the same algorithm:

```
1. Check max depth
2. Search current directory for target files
3. Read directory entries
4. For each subdirectory:
   a. Skip if hidden (with exceptions)
   b. Skip if it's another resource type
   c. Recursively search subdirectory
5. Collect and return results
```

## Key Differences

| Aspect | Commands | Agents | Skills | Packages |
|--------|----------|--------|--------|----------|
| **Target** | `*.md` files | `*.md` files | dirs with `SKILL.md` | `*.package.json` |
| **Loader** | LoadCommandWithBase | LoadAgent | LoadSkill | LoadPackage |
| **Return** | Resources + Errors | Resources + Errors + err | Resources + Errors + err | Packages + err |
| **Skip dirs** | agents, skills | commands, skills, node_modules, .git | (none specific) | agents, skills, commands |

## The Root Problem

**None of these functions check if they're IN the correct resource subtree.**

They all start from a root and search EVERYWHERE, only skipping other resource types.

## Proposed Unified Solution

### Design: Generic Recursive Walker

```go
// ResourceMatcher defines how to identify and load a resource
type ResourceMatcher struct {
    ResourceType  string   // "command", "agent", "skill", "package"
    FilePattern   string   // "*.md", "*.package.json", "SKILL.md"
    IsDirectory   bool     // true for skills, false for files
    SkipDirs      []string // directories to skip
    LoadFunc      func(string) (interface{}, error)
    ValidateFunc  func(string) bool // check if path is in valid subtree
}

// recursiveDiscover is the unified traversal function
func recursiveDiscover(
    rootPath string,
    currentDepth int,
    maxDepth int,
    matcher ResourceMatcher,
) ([]interface{}, []DiscoveryError, error) {
    if currentDepth > maxDepth {
        return nil, nil, nil
    }

    var results []interface{}
    var errors []DiscoveryError

    // Search current directory
    found, errs := searchDirectory(rootPath, matcher)
    results = append(results, found...)
    errors = append(errors, errs...)

    // Recursively search subdirectories
    entries, err := os.ReadDir(rootPath)
    if err != nil {
        return results, errors, err
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        // Skip hidden directories (with exceptions)
        if shouldSkipDir(entry.Name(), matcher.SkipDirs) {
            continue
        }

        subPath := filepath.Join(rootPath, entry.Name())
        subResults, subErrs, _ := recursiveDiscover(subPath, currentDepth+1, maxDepth, matcher)
        results = append(results, subResults...)
        errors = append(errors, subErrs...)
    }

    return results, errors, nil
}

// searchDirectory searches for resources in a single directory
func searchDirectory(dir string, matcher ResourceMatcher) ([]interface{}, []DiscoveryError) {
    var results []interface{}
    var errors []DiscoveryError

    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, []DiscoveryError{{Path: dir, Error: err}}
    }

    for _, entry := range entries {
        path := filepath.Join(dir, entry.Name())

        // Check if this is the target type
        if matcher.IsDirectory {
            if entry.IsDir() && matcher.ValidateFunc(path) {
                result, err := matcher.LoadFunc(path)
                if err != nil {
                    errors = append(errors, DiscoveryError{Path: path, Error: err})
                } else {
                    results = append(results, result)
                }
            }
        } else {
            if !entry.IsDir() && matchesPattern(entry.Name(), matcher.FilePattern) {
                // **KEY FIX**: Check if path is in valid resource subtree
                if !isInResourceSubtree(path, matcher.ResourceType) {
                    continue // Skip files outside resource directories
                }

                result, err := matcher.LoadFunc(path)
                if err != nil {
                    errors = append(errors, DiscoveryError{Path: path, Error: err})
                } else {
                    results = append(results, result)
                }
            }
        }
    }

    return results, errors
}

// isInResourceSubtree checks if path contains a resource directory in its tree
func isInResourceSubtree(path string, resourceType string) bool {
    parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
    
    for _, part := range parts {
        switch resourceType {
        case "command":
            if part == "commands" {
                return true
            }
        case "agent":
            if part == "agents" {
                return true
            }
        case "skill":
            if part == "skills" {
                return true
            }
        case "package":
            if part == "packages" {
                return true
            }
        }
    }
    return false
}
```

## Usage Examples

### Commands
```go
matcher := ResourceMatcher{
    ResourceType: "command",
    FilePattern:  "*.md",
    IsDirectory:  false,
    SkipDirs:     []string{"agents", "skills", "node_modules", ".git"},
    LoadFunc:     func(p string) (interface{}, error) { 
        return resource.LoadCommandWithBase(p, basePath) 
    },
    ValidateFunc: func(p string) bool { return isInResourceSubtree(p, "command") },
}
results, errs, err := recursiveDiscover(rootPath, 0, maxDepth, matcher)
```

### Agents
```go
matcher := ResourceMatcher{
    ResourceType: "agent",
    FilePattern:  "*.md",
    IsDirectory:  false,
    SkipDirs:     []string{"commands", "skills", "node_modules", ".git"},
    LoadFunc:     func(p string) (interface{}, error) { 
        return resource.LoadAgent(p) 
    },
    ValidateFunc: func(p string) bool { return isInResourceSubtree(p, "agent") },
}
```

## Migration Plan

### Phase 1: Add Unified Function
1. Create new file: `pkg/discovery/traverse.go`
2. Implement `recursiveDiscover()` with path filtering
3. Add comprehensive tests

### Phase 2: Migrate Commands
1. Update `recursiveSearchCommands()` to use unified function
2. Verify tests pass
3. Keep old function signature for compatibility

### Phase 3: Migrate Agents, Skills, Packages
1. Update each to use unified function
2. Verify all tests pass
3. Remove old recursive functions

### Phase 4: Cleanup
1. Remove deprecated functions
2. Update documentation
3. Final test run

## Benefits

1. **Single source of truth** for traversal logic
2. **Fixes the bug** with path-based filtering in one place
3. **Easier to maintain** - change once, affects all
4. **Easier to test** - test generic function once
5. **More consistent** behavior across resource types
6. **Less code** - remove ~300 lines of duplication

## Acceptance Criteria

- [ ] New unified `recursiveDiscover()` function created
- [ ] Path filtering (`isInResourceSubtree`) implemented
- [ ] All 4 resource types migrated to use it
- [ ] All existing tests pass
- [ ] No false positive discovery errors
- [ ] make test passes
