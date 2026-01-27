# Malformed Resources Fixture

## Purpose

Edge cases and error conditions to test robust error handling in discovery.

## Structure

```
skills/
  missing-description/
    SKILL.md      # Missing required description field
  invalid-yaml/
    SKILL.md      # Malformed YAML frontmatter
  empty-file/
    SKILL.md      # Empty file
```

Three malformed skills to test error handling.

## Test Use Cases

- Missing required fields (description)
- Invalid YAML frontmatter syntax
- Empty resource files
- Graceful error handling
- Error message clarity
- Partial discovery success (continue after errors)

## Expected Behavior

Discovery should:
- Skip invalid resources with clear error messages
- Continue processing valid resources
- Not crash on malformed input
- Report all errors encountered
