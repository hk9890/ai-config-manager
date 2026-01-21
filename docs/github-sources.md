# GitHub Sources Guide

This guide provides detailed information about using GitHub sources with `ai-repo`.

## Overview

`ai-repo` can add resources directly from GitHub repositories, making it easy to discover, share, and use community-contributed resources. When you add a resource from GitHub, `ai-repo`:

1. Clones the repository to a temporary directory
2. Auto-discovers resources in standard locations
3. Copies resources to your centralized repository (`~/.local/share/ai-config/repo/`)
4. Cleans up the temporary directory

## Source Format Syntax

### Basic GitHub Source

```bash
ai-repo add skill gh:owner/repo
```

This clones the repository and auto-discovers all skills.

### GitHub Source with Path

```bash
ai-repo add skill gh:owner/repo/path/to/skill
```

Looks for a skill in a specific path within the repository.

### GitHub Source with Branch/Tag

```bash
ai-repo add skill gh:owner/repo@branch-name
ai-repo add skill gh:owner/repo@v1.0.0
```

Clones from a specific branch or tag.

### Combined Path and Reference

```bash
ai-repo add skill gh:owner/repo/skills/my-skill@v2.0.0
```

Uses both a specific path and a version reference.

### Shorthand Syntax

The `gh:` prefix can be omitted for `owner/repo` format:

```bash
# These are equivalent
ai-repo add skill vercel-labs/agent-skills
ai-repo add skill gh:vercel-labs/agent-skills
```

### Git URL Sources

Use any Git URL directly:

```bash
# HTTPS
ai-repo add skill https://github.com/owner/repo.git
ai-repo add skill https://gitlab.com/owner/repo.git

# SSH
ai-repo add skill git@github.com:owner/repo.git

# With branch
ai-repo add skill https://github.com/owner/repo.git@develop
```

## Auto-Discovery Algorithm

When adding resources from GitHub, `ai-repo` searches for resources in standard locations following tool conventions.

### Skills Discovery

**Priority search order:**

1. **Direct path** (if subpath specified):
   - `SKILL.md` at the exact path: `repo/path/to/skill/SKILL.md`

2. **Standard directories** (searched in order):
   - `skills/`
   - `.claude/skills/`
   - `.opencode/skills/`
   - `.github/skills/`
   - `.codex/skills/`
   - `.cursor/skills/`
   - `.goose/skills/`
   - `.kilocode/skills/`
   - `.kiro/skills/`
   - `.roo/skills/`
   - `.trae/skills/`
   - `.agents/skills/`
   - `.agent/skills/`

3. **Recursive search** (if not found above):
   - Searches up to 5 levels deep
   - Looks for any directory containing `SKILL.md`

**Validation:**
- Each skill must have a valid `SKILL.md` file
- `SKILL.md` must have YAML frontmatter with `name` and `description`
- Directory name must match the `name` field in `SKILL.md`

**Deduplication:**
- If multiple skills with the same name are found, the first one (by priority order) is used

### Commands Discovery

**Search locations:**
1. `commands/`
2. `.claude/commands/`
3. `.opencode/commands/`
4. Recursive search for `.md` files (max depth 5)

**Filtering:**
- Only `.md` files are considered
- `SKILL.md` and `README.md` are excluded
- Files must have valid command frontmatter (at minimum, a `description` field)

### Agents Discovery

**Search locations:**
1. `agents/`
2. `.claude/agents/`
3. `.opencode/agents/`
4. Recursive search for `.md` files (max depth 5)

**Validation:**
- Files must have valid agent frontmatter (at minimum, a `description` field)
- Deduplication by agent name

## Resource Selection

### Single Resource Found

If exactly one resource is found, it's automatically added:

```bash
$ ai-repo add skill gh:myorg/my-skill
Cloning repository...
Found skill: my-skill
✓ Added skill 'my-skill' to repository
  Version: 1.0.0
  Description: My custom skill
```

### Multiple Resources Found

If multiple resources are found and no specific path was provided, you'll be prompted to select:

```bash
$ ai-repo add skill gh:myorg/multi-skill-repo
Cloning repository...
Found 3 skills:
  1. skill-one - Frontend development skill
  2. skill-two - Backend API skill
  3. skill-three - Database management skill

Select a skill to add (1-3): 2
✓ Added skill 'skill-two' to repository
```

To avoid the prompt, specify the exact path:

```bash
ai-repo add skill gh:myorg/multi-skill-repo/skills/skill-two
```

### No Resources Found

If no resources are found, an error is displayed:

```bash
$ ai-repo add skill gh:myorg/empty-repo
Cloning repository...
Error: no skills found in repository: https://github.com/myorg/empty-repo
```

Check the repository's structure or documentation to find where resources are located.

## Precedence Rules

When multiple search paths exist, resources are discovered in this precedence:

1. **Explicit subpath** (if provided in source)
2. **Standard tool directories** (in order: skills/, .claude/skills/, .opencode/skills/, etc.)
3. **Recursive search** (fallback, max depth 5)

The first valid resource found wins.

## Examples

### Example 1: Add a Single Skill from GitHub

```bash
# Add the entire vercel-labs/agent-skills repository
ai-repo add skill gh:vercel-labs/agent-skills

# Or use shorthand
ai-repo add skill vercel-labs/agent-skills
```

### Example 2: Add a Specific Skill from Multi-Skill Repo

```bash
# Specify the exact skill path
ai-repo add skill gh:vercel-labs/agent-skills/skills/frontend-design
```

### Example 3: Add from a Specific Version

```bash
# Add from a tagged release
ai-repo add skill gh:anthropics/skills@v2.1.0

# Add from a branch
ai-repo add skill gh:myorg/experimental-skills@develop
```

### Example 4: Add Commands from GitHub

```bash
# Auto-discover commands in a repository
ai-repo add command gh:myorg/commands

# Add a specific command
ai-repo add command gh:myorg/repo/commands/deploy.md
```

### Example 5: Add Agents from GitHub

```bash
# Auto-discover agents
ai-repo add agent gh:myorg/agents

# Add specific agent
ai-repo add agent gh:myorg/agents/code-reviewer.md
```

### Example 6: Use Git URLs

```bash
# HTTPS URL
ai-repo add skill https://github.com/myorg/custom-skills.git

# SSH URL (if you have keys configured)
ai-repo add skill git@github.com:myorg/private-skills.git

# GitLab or other Git hosting
ai-repo add skill https://gitlab.com/mygroup/skills.git
```

## Centralized Storage

All resources added from GitHub are stored in your centralized repository:

```
~/.local/share/ai-config/repo/
├── commands/
│   └── deploy.md
├── skills/
│   ├── frontend-design/
│   └── pdf-processing/
└── agents/
    └── code-reviewer.md
```

This means:
- Resources are downloaded once, regardless of how many projects use them
- Updates to a resource in the repo affect all projects (via symlinks)
- No duplication across projects

## Performance Considerations

### Shallow Clones

`ai-repo` uses shallow clones (`git clone --depth 1`) for speed:
- Only downloads the latest commit
- Significantly faster for large repositories
- Sufficient for resource discovery

### Temporary Directories

Repositories are cloned to temporary directories and cleaned up automatically:
- No disk space waste from cloned repos
- Cleanup happens even if errors occur
- Temp directory location respects system defaults

### Caching

Currently, `ai-repo` does not cache cloned repositories. Each `add` operation clones fresh. Future versions may add caching.

## Troubleshooting

### Problem: Git not installed

**Error:** `git clone failed: exec: "git": executable file not found`

**Solution:** Install Git:
```bash
# Ubuntu/Debian
sudo apt-get install git

# macOS
brew install git

# Verify installation
git --version
```

### Problem: Repository not found

**Error:** `git clone failed: repository not found`

**Solutions:**
- Verify the repository URL is correct
- Check that the repository is public (or you have access)
- For private repos, ensure SSH keys or credentials are configured

### Problem: Network timeout

**Error:** `git clone failed: connection timed out`

**Solutions:**
- Check your internet connection
- Try again (may be temporary network issue)
- For large repos, consider cloning manually first

### Problem: Invalid credentials

**Error:** `git clone failed: authentication required`

**Solutions:**
- For HTTPS URLs, configure Git credentials:
  ```bash
  git config --global credential.helper store
  ```
- For SSH URLs, ensure your SSH keys are configured:
  ```bash
  ssh -T git@github.com
  ```

### Problem: Resource not found after clone

**Error:** `no skills found in repository`

**Solutions:**
- Check the repository structure - resources must be in standard locations
- Use a specific path: `gh:owner/repo/path/to/skill`
- Verify the resource has valid frontmatter (SKILL.md with name and description)
- Clone manually and inspect: `git clone https://github.com/owner/repo && ls -R`

### Problem: Multiple resources found

**Error:** `multiple skills found, please specify path`

**Solutions:**
- Add the specific path: `ai-repo add skill gh:owner/repo/skills/specific-skill`
- Or select interactively from the prompt

## Best Practices

### 1. Use Version Tags

For production use, pin to specific versions:

```bash
# Good - pinned version
ai-repo add skill gh:myorg/skills@v1.2.0

# Risky - uses latest code
ai-repo add skill gh:myorg/skills
```

### 2. Specify Paths for Multi-Resource Repos

Avoid ambiguity by specifying the exact resource:

```bash
# Specific - no prompt needed
ai-repo add skill gh:myorg/mono-repo/skills/my-skill

# Generic - may prompt for selection
ai-repo add skill gh:myorg/mono-repo
```

### 3. Use Shorthand for Public Repos

The shorthand syntax is cleaner for GitHub repos:

```bash
# Preferred
ai-repo add skill vercel-labs/agent-skills

# Also works, but more verbose
ai-repo add skill gh:vercel-labs/agent-skills
```

### 4. Check Repository Documentation

Before adding, check the repository's README for:
- Resource locations (if not in standard directories)
- Available resources and their purposes
- Version compatibility

### 5. Keep Local Copies for Critical Resources

For production-critical resources, consider:
1. Add from GitHub to your repo
2. Export to local directory: `cp -r ~/.local/share/ai-config/repo/skills/my-skill ~/backups/`
3. Re-add from local if needed: `ai-repo add skill ~/backups/my-skill --force`

## Contributing GitHub Sources

If you're creating resources to share on GitHub:

### Repository Structure

Use standard directory structures for easy discovery:

```
my-skills-repo/
├── README.md
├── skills/
│   ├── skill-one/
│   │   └── SKILL.md
│   └── skill-two/
│       └── SKILL.md
├── commands/
│   ├── command-one.md
│   └── command-two.md
└── agents/
    ├── agent-one.md
    └── agent-two.md
```

Or use tool-specific directories:

```
my-skills-repo/
├── README.md
└── .claude/
    ├── skills/
    │   └── my-skill/
    │       └── SKILL.md
    ├── commands/
    │   └── my-command.md
    └── agents/
        └── my-agent.md
```

### Metadata Best Practices

**Skills (SKILL.md):**
```yaml
---
name: my-skill
description: Clear, concise description
version: "1.0.0"
license: MIT
metadata:
  author: Your Name
  tags: category, feature
---
```

**Commands:**
```yaml
---
description: What this command does
---
```

**Agents:**
```yaml
---
description: What this agent does
type: agent-role
---
```

### Version Tags

Use semantic versioning tags:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

Users can then reference specific versions:

```bash
ai-repo add skill gh:yourorg/your-repo@v1.0.0
```

### Documentation

Include in your README:
- List of resources and their purposes
- Installation instructions
- Resource locations (if not in standard directories)
- Version compatibility
- Examples

## Related Documentation

- [README.md](../README.md) - Main user documentation
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guide with GitHub source architecture
- [Source Formats](../README.md#source-formats) - Overview of all source formats
- [Auto-Discovery](../README.md#auto-discovery) - Summary of discovery algorithms
