# Subpath Test Fixture

## Purpose

Deeply nested resources to test subpath discovery and recursive scanning.

## Structure

```
level1/
  level2/
    skills/
      deep-skill/
        SKILL.md
```

A single skill nested several directories deep.

## Test Use Cases

- Deep recursive discovery
- Subpath filtering (e.g., discovering only `level1/level2`)
- Path handling in nested structures
- Performance with deep directory trees
- Correct path resolution for nested resources

## Usage

This fixture is useful for testing:
- `aimgr repo add path/to/repo --subpath level1/level2`
- Deep directory scanning limits
- Path resolution correctness
