# Test Fixtures for Discovery Tests

This directory contains committed test fixtures used by discovery unit tests. These fixtures replace slow GitHub integration tests with fast file-based tests.

## Purpose

The fixtures in this directory allow us to test resource discovery without relying on external GitHub repositories or Git operations. This makes tests:

- **Fast**: No network calls or git clones (50+ seconds saved per test run)
- **Reliable**: No dependency on external services
- **Deterministic**: Fixed test data ensures consistent results
- **Easy to debug**: All test data is visible in the repository

## Structure

Each subdirectory represents a complete test repository with various resource layouts:

- **skills-standard**: Standard skills directory structure (2 skills)
- **commands-nested**: Commands with nested directory structure (3 commands)
- **mixed-resources**: Mix of commands, skills, and agents
- **dotdir-resources**: Tool-specific directories (.claude, .opencode)
- **malformed-resources**: Edge cases for error handling
- **subpath-test**: Deeply nested resources
- **empty-repo**: Repository with no resources

## Adding New Fixtures

To add a new test fixture:

1. Create a new directory under `testdata/repos/`
2. Add a `README.md` explaining the fixture's purpose and origin
3. Create valid resource files following the formats below
4. Keep content minimal but realistic
5. Commit all files to the repository

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

## Attribution

These fixtures are based on structures from:
- Anthropic's skills repository (https://github.com/anthropics/skills)
- Anthropic's quickstarts repository (https://github.com/anthropics/quickstarts)
- OpenCode examples
- Real-world usage patterns

## Usage in Tests

Import fixtures in tests:

```go
import "path/filepath"

func TestDiscovery(t *testing.T) {
    fixturePath := filepath.Join("testdata", "repos", "skills-standard")
    resources, err := discovery.Discover(fixturePath, opts)
    // ... assertions
}
```

## Maintenance

- Keep fixtures small and focused on specific test scenarios
- Update fixtures when resource formats change
- Document any non-obvious test cases in fixture README files
- Avoid binary files or large content
