# Validation Improvements for Resource Developers

## Current State Analysis

### Existing Validation Capabilities

aimgr already has robust validation infrastructure:

1. **Built-in validation during loading:**
   - `LoadCommand()`, `LoadSkill()`, `LoadAgent()`, `LoadPackage()`
   - Automatic validation of name format, description, required fields
   - Rich error messages with actionable suggestions via `ValidationError`

2. **Repository verification:**
   - `aimgr repo verify` - checks metadata consistency and package references
   - Supports patterns (e.g., `skill/*`, `command/test*`)
   - Auto-fix mode with `--fix` flag
   - Multiple output formats (table, JSON, YAML)

3. **Validation functions:**
   - `ValidateName()` - enforces agentskills.io naming rules
   - `ValidateDescription()` - checks description constraints
   - `resource.Validate()` - comprehensive resource validation

### Current Gaps for Developers

1. **No standalone validation command before adding to repo**
   - Developers must use `repo import` to validate
   - No explicit "validate-only" command

2. **Limited documentation for developers**
   - No dedicated guide for resource creators
   - Validation process not clearly documented
   - Testing workflow unclear

3. **No examples of validation workflows**
   - Missing CI/CD integration examples
   - No testing workflow documentation

## Proposed Improvements

### 1. Documentation (High Priority) ✅ DONE

**Status:** Created `docs/user-guide/developer-guide.md`

**Contents:**
- Validation methods (dry-run import, repo verify)
- Common validation errors with examples
- Local testing workflows
- Best practices for each resource type
- Publishing checklist
- CI/CD integration examples

**Impact:** Immediate value for developers without code changes

### 2. CLI Enhancement: Make Dry-Run More Discoverable (Low Priority)

**Problem:** `--dry-run` flag exists but is not well-documented as a validation tool

**Current:**
```bash
aimgr repo import ./my-skill --dry-run
```

**Proposed:** Add alias or dedicated command
```bash
# Option A: Add alias to import command
aimgr repo import ./my-skill --validate

# Option B: New subcommand (may be overkill)
aimgr repo validate ./my-skill
```

**Implementation:**
- Add `--validate` as an alias for `--dry-run` in `cmd/repo_import.go`
- Update help text to emphasize validation use case
- Add examples in `--help` output

**Estimated effort:** 1-2 hours
**Value:** Medium (improves discoverability)

### 3. Enhanced Validation Output (Medium Priority)

**Problem:** Validation errors could provide more context

**Current output:**
```
Error: skill 'my-skill' field 'name' in path/to/skill: name must match directory name
```

**Proposed enhancements:**

A. **Add file/line information for frontmatter errors:**
```
Error: skill 'my-skill' in path/to/skill/SKILL.md:3
  name must match directory name 'my-other-skill'
  → Suggestion: Rename directory to 'my-skill' or update frontmatter
```

B. **Summary statistics for bulk validation:**
```
Validation Summary:
  ✓ 12 resources valid
  ✗ 3 resources failed
  ⚠ 2 warnings

Errors:
  - skill/broken-skill: name mismatch
  - command/bad-cmd: missing description
  - package/bad-pkg: invalid resource reference

Use --format=json for detailed output
```

**Implementation:**
- Enhance `ValidationError` to include line numbers
- Add summary output to `repo import --dry-run`
- Track validation statistics

**Estimated effort:** 4-6 hours
**Value:** High (better UX for bulk validation)

### 4. Pre-commit Hook / Git Integration (Low Priority)

**Problem:** Developers might commit invalid resources

**Proposed:** Provide example git hooks

Create `examples/git-hooks/pre-commit`:
```bash
#!/bin/bash
# Validate resources before commit

echo "Validating AI resources..."

# Find changed resource files
CHANGED_SKILLS=$(git diff --cached --name-only --diff-filter=ACM | grep 'skills/.*/SKILL.md')
CHANGED_COMMANDS=$(git diff --cached --name-only --diff-filter=ACM | grep 'commands/.*\.md')
CHANGED_AGENTS=$(git diff --cached --name-only --diff-filter=ACM | grep 'agents/.*\.md')

# Validate each
for file in $CHANGED_SKILLS $CHANGED_COMMANDS $CHANGED_AGENTS; do
  dir=$(dirname "$file")
  if ! aimgr repo import "$dir" --dry-run --format=json > /dev/null 2>&1; then
    echo "❌ Validation failed for $file"
    aimgr repo import "$dir" --dry-run
    exit 1
  fi
done

echo "✅ All resources valid"
```

**Installation docs:**
```bash
# Copy to your project
cp examples/git-hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

**Estimated effort:** 2-3 hours
**Value:** Medium (prevents invalid commits)

### 5. Interactive Validation Mode (Future Enhancement)

**Problem:** Batch errors can be overwhelming

**Proposed:** Interactive fixer

```bash
$ aimgr repo validate ./my-resources --interactive

Checking skill/my-skill...
  ✗ Error: name 'my_skill' contains invalid characters

  Fix options:
    1. Rename to 'my-skill' (recommended)
    2. Skip this resource
    3. Abort validation

  Choice [1-3]: 1
  ✓ Renamed to 'my-skill'

Checking command/test...
  ✗ Error: missing description field

  Fix options:
    1. Add default description
    2. Edit frontmatter now
    3. Skip this resource

  Choice [1-3]: 2
  [Opens editor]
  ✓ Description added

Validation complete: 2 resources fixed, 0 skipped
```

**Estimated effort:** 8-12 hours
**Value:** High (great DX for beginners)

## Recommendations

### Immediate Actions (This PR)

1. ✅ **Add developer-guide.md** - Already created
2. **Update docs/user-guide/README.md** - Add link to developer guide
3. **Update main README.md** - Add small section linking to developer guide

### Short-term Improvements (Next PR)

1. **Make dry-run more discoverable**
   - Add `--validate` alias to `repo import`
   - Update help text
   - Add examples

2. **Enhanced validation output**
   - Add summary statistics
   - Improve error formatting

### Long-term Enhancements (Future)

1. **Interactive validation mode**
2. **Pre-commit hook examples**
3. **Validation API for programmatic use**

## Implementation Plan

### Phase 1: Documentation (This PR) ✅

- [x] Create `docs/user-guide/developer-guide.md`
- [ ] Update `docs/user-guide/README.md` to link to developer guide
- [ ] Add small section in main README.md
- [ ] Review and merge

### Phase 2: CLI Improvements (Next PR)

**File changes:**
- `cmd/repo_import.go` - Add `--validate` alias
- `pkg/repo/manager.go` - Add validation summary output
- `cmd/repo_import_test.go` - Add tests

**Testing:**
- Unit tests for new flag
- Integration test for validation workflow

### Phase 3: Future Enhancements

- Interactive mode
- Git hooks
- Enhanced error formatting

## Testing Strategy

For documentation changes (Phase 1):
- Manual review of developer-guide.md
- Test all command examples
- Verify links work

For CLI changes (Phase 2):
- Unit tests for new flags
- Integration tests for validation workflows
- Update existing tests if needed

## Success Metrics

1. **Documentation:**
   - Developer guide exists and is discoverable
   - All examples work as documented
   - Common errors are documented with solutions

2. **CLI improvements:**
   - Validation workflow is clear and documented
   - Error messages are actionable
   - Exit codes are correct for CI/CD use

3. **Developer experience:**
   - Time to validate resources decreases
   - Fewer invalid resources added to repositories
   - Positive community feedback

## Questions to Resolve

1. Should we add `--validate` as an alias or create a new command?
   - **Recommendation:** Alias (simpler, less code)

2. Should validation be part of `repo` subcommand or top-level?
   - **Current:** `aimgr repo import --dry-run`
   - **Recommendation:** Keep in `repo` subcommand for consistency

3. What level of verbosity is appropriate for validation output?
   - **Recommendation:** Concise by default, verbose with `--verbose` flag

## Conclusion

The existing validation infrastructure is solid. The main gap is **documentation and discoverability**. 

**Immediate value** can be delivered by:
1. Adding developer-guide.md (done)
2. Making validation workflow more prominent in docs
3. Adding examples and best practices

**Future improvements** can focus on:
1. Better CLI UX for validation
2. Interactive validation mode
3. CI/CD integration helpers

The proposed developer-guide.md addresses the most critical need: helping developers understand how to validate their resources before publishing.
