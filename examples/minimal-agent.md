---
description: Minimal agent example with only required fields
---

# Minimal Agent

This is a minimal agent example that demonstrates the simplest possible agent configuration.

## Overview

This agent has only the required `description` field in its frontmatter. This format is compatible with both Claude Code and OpenCode.

## When to Use Minimal Format

Use this minimal format when:
- You're just getting started with agents
- You want to define instructions entirely in the markdown body
- You're using Claude Code format (which doesn't require type/instructions in frontmatter)
- You prefer simplicity over structured metadata

## Agent Instructions

All agent behavior and instructions are defined here in the markdown body instead of in the frontmatter.

### What This Agent Does

This agent serves as a template for creating simple agents without the additional OpenCode-specific fields like `type`, `instructions`, or `capabilities`.

### Usage

Simply add your agent instructions and documentation in the markdown body. The agent will use this content to understand its role and behavior.

## Expanding This Agent

To convert this to an OpenCode format agent with more structure, add these fields to the frontmatter:

```yaml
---
description: Minimal agent example with only required fields
type: your-agent-type
instructions: Detailed instructions here
capabilities:
  - capability-1
  - capability-2
version: "1.0.0"
author: your-name
license: MIT
---
```

## See Also

- [sample-agent.md](sample-agent.md) - Full OpenCode format example
- [claude-style-agent.md](claude-style-agent.md) - Claude Code format example
