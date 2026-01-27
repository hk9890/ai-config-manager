# Dotdir Resources Fixture

## Purpose

Tool-specific directory structures (.claude, .opencode) to test discovery in tool config directories.

## Structure

```
.claude/
  commands/
    claude-cmd.md
  skills/
    claude-skill/
      SKILL.md
.opencode/
  agents/
    opencode-agent.md
```

Resources organized in tool-specific dotdirectories.

## Test Use Cases

- Discovery in .claude directories
- Discovery in .opencode directories
- Tool-specific resource organization
- Dotdirectory handling (not hidden from discovery)
- Priority handling when both generic and tool-specific dirs exist
