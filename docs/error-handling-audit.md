# Error Handling Audit

## Executive Summary

Audit of bulk operations in aimgr to identify error handling patterns. Found one critical issue where operations stop on first error instead of collecting all errors.

**Date**: 2026-01-26
**Status**: Analysis Complete

## Error Categories

### 1. Fatal Errors (Should Stop Immediately)
- Internal bugs: null pointer dereference, index out of bounds
- System failures: out of memory, disk full
- Programming errors: should never happen in production

### 2. Validation Errors (Should Collect and Continue)
- Invalid YAML/JSON syntax
- Missing required fields in frontmatter
- Invalid field values (e.g., invalid name format)
- Schema validation failures
- Resource name conflicts ("already exists")

### 3. Resource Errors (Should Collect and Continue)
- File not found
- Permission denied
- Network failures (Git clone timeout)
- Source path doesn't exist
- Directory not accessible

## Bulk Operations Analysis

### ✅ GOOD: pkg/marketplace/generator.go - GeneratePackages()

**Location**: `pkg/marketplace/generator.go:28-99`

**Current behavior**: ✅ Continues on errors, collects all plugins

**Error handling**:
```go
for i, plugin := range marketplace.Plugins {
    // ...
    if info, err := os.Stat(sourcePath); err != nil {
        // Skip plugins with missing source directories
        continue  // ← Continues processing
    }
    // ...
    resources, err := discoverResources(sourcePath)
    if err != nil {
        return nil, fmt.Errorf("plugin %d (%s): failed to discover resources: %w", i, plugin.Name, err)
    }
}
```

**Error types**:
- Missing source directory (resource) - ✅ continues (line 56)
- Invalid plugin name (validation) - ❌ stops (line 64, but acceptable since name is critical)
- Discovery failure (mixed) - ❌ stops (line 75, should continue for validation errors)

**Recommendation**: 
- Current behavior is acceptable for marketplace import
- Consider continuing on discovery failures and reporting at end

---

### ❌ CRITICAL: pkg/repo/manager.go - AddBulk()

**Location**: `pkg/repo/manager.go:626-655`

**Current behavior**: ❌ Stops on first error (unless --skip-existing or --force)

**Problem code**:
```go
for _, sourcePath := range sources {
    if err := m.importResource(sourcePath, opts, result); err != nil {
        // If not skipping existing, fail on first error
        if !opts.SkipExisting && !opts.Force {
            return result, err  // ← STOPS on first error
        }
    }
}
```

**Impact**:
- `aimgr repo add --dry-run` stops after first "already exists" conflict
- User cannot see all potential conflicts before actual import
- Makes dry-run less useful for planning bulk operations

**Error types in importResource()**:
1. Line 666-671: Invalid resource type (validation) - should continue
2. Line 687-692: Failed to load resource (validation) - should continue
3. Line 718-723: Already exists (validation) - should continue
4. Line 704-710: Failed to remove existing (resource/fatal) - should continue on resource error
5. Line 757-762: Import failed (mixed) - depends on error type

**Recommendation**: 
- Distinguish between error types
- Only stop on fatal errors (internal bugs, system failures)
- Continue on validation/resource errors
- Collect all errors in result.Failed
- Return nil error but non-empty result.Failed indicates issues

---

### ✅ GOOD: pkg/discovery/skills.go - DiscoverSkillsWithErrors()

**Location**: `pkg/discovery/skills.go:88-138`

**Current behavior**: ✅ Collects errors, continues processing

**Error handling**:
```go
skills, errs := searchSkillsInDir(locationPath)
allErrors = append(allErrors, errs...)  // ← Collects errors
if len(skills) > 0 {
    // Continues processing
}
```

**Error types**:
- Invalid SKILL.md (validation) - ✅ continues, collects error (line 180-186)
- Directory not accessible (resource) - ✅ continues (line 156)
- Missing SKILL.md (resource) - ✅ continues (line 174)

**Recommendation**: ✅ Already correct, no changes needed

---

### ✅ GOOD: cmd/repo_update.go - Update operations

**Location**: `cmd/repo_update.go:220-232`

**Current behavior**: ✅ Continues on errors, collects results

**Error handling**:
```go
for _, res := range resources {
    ctx.Current++
    result := updateSingleResourceWithProgress(manager, res.Name, resourceType, ctx)
    results = append(results, result)  // ← Continues regardless of result
}
```

**Error types**:
- Metadata not found (resource) - ✅ continues (line 258-260)
- Unknown source type (validation) - ✅ continues (line 293-295)
- Update failed (resource/network) - ✅ continues (line 298-300)
- Source path no longer exists (resource) - ✅ continues (line 288)

**Recommendation**: ✅ Already correct, no changes needed

---

### ✅ GOOD: pkg/discovery/commands.go - DiscoverCommands()

**Location**: `pkg/discovery/commands.go:64-115`

**Current behavior**: ✅ Collects errors silently, continues processing

**Error handling**:
```go
// Check each candidate file
for _, filePath := range candidateFiles {
    cmd, err := resource.LoadCommand(filePath)
    if err != nil {
        continue  // Skip invalid commands, no error collection
    }
    // ...
}
```

**Error types**:
- Invalid command file (validation) - ✅ continues (line 176)
- Directory not accessible (resource) - ✅ continues (line 96)

**Recommendation**: 
- Consider adding error collection like skills discovery
- Would help users understand why commands were skipped

---

### ✅ GOOD: pkg/discovery/agents.go - DiscoverAgents()

**Location**: `pkg/discovery/agents.go:64-115`

**Current behavior**: ✅ Continues on errors, skips invalid agents

**Error handling**:
```go
for _, filePath := range candidateFiles {
    agent, err := resource.LoadAgent(filePath)
    if err != nil {
        continue  // Skip invalid agents
    }
    // ...
}
```

**Error types**:
- Invalid agent file (validation) - ✅ continues
- Directory not accessible (resource) - ✅ continues

**Recommendation**: 
- Consider adding error collection for consistency with skills

---

## Summary

### Critical Issues
1. **pkg/repo/manager.go AddBulk()** - Stops on first error (MUST FIX)
   - Impacts: repo add, repo sync, marketplace import
   - Severity: High
   - Users blocked from seeing all conflicts during dry-run

### Minor Improvements
1. **pkg/discovery/commands.go** - No error collection (nice to have)
2. **pkg/discovery/agents.go** - No error collection (nice to have)
3. **pkg/marketplace/generator.go** - Could continue on discovery failures (optional)

### Already Correct
1. **pkg/discovery/skills.go** - ✅ Collects errors, continues
2. **cmd/repo_update.go** - ✅ Collects results, continues
3. **Marketplace plugin iteration** - ✅ Continues on missing sources

## Recommendations

### Priority 1: Fix AddBulk Error Handling

1. Create error type hierarchy:
   - `IsFatal(error)` - Internal bugs, system failures
   - `IsValidation(error)` - Invalid configs, missing fields
   - `IsResource(error)` - File not found, permission denied

2. Update `AddBulk` to only stop on fatal errors:
   ```go
   for _, sourcePath := range sources {
       if err := m.importResource(sourcePath, opts, result); err != nil {
           if IsFatal(err) {
               return result, err  // Stop only on fatal errors
           }
           // Validation/resource errors already in result.Failed
           continue
       }
   }
   ```

3. Update error creation in `importResource` to use typed errors:
   - "already exists" → `ValidationError`
   - "failed to load resource" → `ValidationError`
   - "failed to remove" → `ResourceError` or `FatalError` (depends on cause)
   - "invalid resource type" → `ValidationError`

### Priority 2: Add Error Collection to Discovery

Update commands.go and agents.go to collect errors like skills.go:
- Add `DiscoverCommandsWithErrors()` function
- Add `DiscoverAgentsWithErrors()` function
- Return `(resources, errors, error)` tuple

## Implementation Plan

1. **Phase 1: Error Type System** (ai-config-manager-glis)
   - Create `pkg/errors/types.go`
   - Define `FatalError`, `ValidationError`, `ResourceError`
   - Add helper functions `IsFatal()`, `IsValidation()`, `IsResource()`
   - Add tests

2. **Phase 2: Update Error Creation** (ai-config-manager-glc7)
   - Update all error creation in `importResource()` to use typed errors
   - Update `AddCommand`, `AddSkill`, `AddAgent` to use typed errors
   - Update tests

3. **Phase 3: Update AddBulk Logic** (ai-config-manager-glc7)
   - Change loop to only stop on fatal errors
   - Ensure validation/resource errors are collected
   - Update tests

4. **Phase 4: Test Everything** (ai-config-manager-glc7)
   - Test dry-run shows all conflicts
   - Test continues on validation errors
   - Test stops on fatal errors
   - Integration tests

## Test Cases

### AddBulk with Multiple Conflicts
```bash
# Given: 3 resources, 2 already exist, 1 new
aimgr repo add /path/to/resources --dry-run

# Expected: Shows all 3 resources
#   - 2 in Failed (already exists)
#   - 1 in Added (would be added)
#
# Current: Stops after first conflict, only shows 1
```

### AddBulk with Mixed Errors
```bash
# Given: 5 resources
#   - 1 invalid YAML
#   - 1 already exists
#   - 1 missing required field
#   - 1 permission denied
#   - 1 valid

# Expected: Shows all 5
#   - 4 in Failed (with specific errors)
#   - 1 in Added
#
# Current: Stops after first error
```

### AddBulk with Fatal Error
```bash
# Given: Disk full during import
# Expected: Stops immediately, shows fatal error
# After fix: Should still stop on fatal errors
```
