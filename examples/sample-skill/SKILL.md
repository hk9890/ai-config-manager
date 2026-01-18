---
name: sample-skill
description: Example skill demonstrating the format
license: MIT
metadata:
  author: example
  version: "1.0.0"
---

# Sample Skill

This is an example skill that demonstrates the proper format for agentskills.io compatible skills.

## Structure

A skill is a directory containing:
- **SKILL.md** (required): Metadata and documentation
- **scripts/** (optional): Executable scripts the skill can use
- **references/** (optional): Additional documentation and references
- **assets/** (optional): Images, diagrams, or other resources

## Frontmatter

The SKILL.md file must start with YAML frontmatter containing:
- `name`: Must match the folder name (required)
- `description`: 1-1024 character description (required)
- `license`: Optional license identifier (e.g., MIT, Apache-2.0)
- `metadata`: Optional key-value pairs for version, author, etc.

## Usage

This skill can be added to your repository and installed in projects:

```bash
# Add to repository
ai-repo add skill examples/sample-skill

# Install in a project
cd my-project/
ai-repo install skill sample-skill
```

## Creating Your Own Skills

1. Create a directory with a valid name (lowercase, alphanumeric, hyphens only)
2. Create SKILL.md with frontmatter matching the directory name
3. Add scripts, references, or assets as needed
4. Test locally before adding to repository
5. Add to repository: `ai-repo add skill ./your-skill`

## References

- Skills specification: https://agentskills.io/specification
- Claude Code documentation: https://code.claude.com/docs
