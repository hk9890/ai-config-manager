# Test Fixtures

This directory contains committed test fixtures used by unit tests to avoid slow network operations.

## Purpose

Tests in `test/discovery_*_test.go` use these fixtures instead of cloning real GitHub repositories. This provides:
- **Fast execution**: <1 second vs 30+ seconds for Git clone
- **Reliability**: No network dependency, no flaky tests
- **Consistency**: Fixtures don't change unexpectedly
- **Clarity**: Fixtures are minimal and focused

## Structure

Each directory is a self-contained fixture representing a repository structure:

| Fixture | Purpose | Based On |
|---------|---------|----------|
| `skills-standard/` | Standard skills layout | anthropics/skills |
| `commands-nested/` | Nested command structure | anthropics/quickstarts |
| `mixed-resources/` | Multiple resource types | Combined structure |
| `dotdir-resources/` | Hidden tool directories | .claude, .opencode patterns |
| `malformed-resources/` | Error handling | Edge cases |
| `subpath-test/` | Deep directory nesting | Subpath discovery testing |
| `empty-repo/` | No resources | Empty repository case |

## Adding New Fixtures

1. Create directory: `mkdir testdata/repos/new-fixture`
2. Add README.md with origin: `# Based on: <source>`
3. Create minimal resource structure
4. Keep files small and focused
5. Commit to repository
6. Document in this README

## Guidelines

- **Minimal**: Only include files needed for tests
- **Valid**: Resources should have proper frontmatter
- **Documented**: README.md explains purpose and origin
- **Small**: Keep total fixture size <1MB
- **No binaries**: Text files only
- **No .git**: These are file trees, not Git repos

### Resource Formats

**Command** (`*.md`):
```yaml
---
description: Command description (required)
agent: agent-name (optional)
model: model-name (optional)
---

# Command content here
```

**Skill** (`SKILL.md` in directory):
```yaml
---
name: skill-name (required, must match directory)
description: Skill description (required)
license: License (optional)
metadata:
  author: Author name
  version: "1.0"
---

# Skill content here
```

**Agent** (`*.md`):
```yaml
---
description: Agent description (required)
type: agent-type (optional, OpenCode)
instructions: Instructions (optional, OpenCode)
capabilities: [list] (optional, OpenCode)
version: "1.0.0" (optional)
author: Author name (optional)
license: License (optional)
---

# Agent content here
```

## Usage in Tests

```go
import "github.com/hk9890/ai-config-manager/test/testutil"

func TestExample(t *testing.T) {
    fixturePath := testutil.GetFixturePath("skills-standard")
    skills, err := discovery.DiscoverSkills(fixturePath, "")
    // ... assertions
}
```

## Attribution

Fixtures are inspired by real open-source projects but are minimal, modified versions created specifically for testing. See individual README files for attribution.
