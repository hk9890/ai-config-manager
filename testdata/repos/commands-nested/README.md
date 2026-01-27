# Commands Nested Fixture

## Purpose

Commands with nested directory structure to test recursive discovery.

## Origin

Based on https://github.com/anthropics/quickstarts structure.

## Structure

```
commands/
  build.md
  test.md
  nested/
    deploy.md
```

Three commands: two at top level, one in a nested subdirectory.

## Test Use Cases

- Commands directory discovery
- Nested command discovery (commands in subdirectories)
- Multiple commands in one repository
- Command name extraction from filename
