# aimgr CLI Reference: Repository Management

Commands for managing the central resource repository.

## Table of Contents

- [repo add](#repo-add) - Add resources to repository
- [repo update](#repo-update) - Update resources from sources
- [repo remove](#repo-remove) - Remove resources from repository

---
---

## Repository Management

### repo add

Add resources to the repository from local paths, GitHub, or Git URLs.

**Syntax:**
```bash
aimgr repo add SOURCE [OPTIONS]
```

**Arguments:**
- `SOURCE` - Resource source (supports multiple formats):
  - Local path: `/path/to/resource` or `./resource`
  - GitHub: `gh:owner/repo[/path][@version]` or `owner/repo` (shorthand)
  - Git URL: `https://github.com/owner/repo.git` or `git@github.com:owner/repo.git`

**Options:**
- `--force` - Overwrite existing resources
- `--skip-existing` - Skip resources that already exist (no error)
- `--dry-run` - Preview what would be imported without making changes
- `--filter=PATTERN` - Filter resources by pattern (see [Pattern Syntax](#pattern-syntax))

**Examples:**

**Add from Local Path (Auto-Detected Type):**
```bash
# Add single resource (type auto-detected from structure)
aimgr repo add ~/my-skills/pdf-processing        # Auto-detects as skill
aimgr repo add ~/.claude/commands/test.md        # Auto-detects as command
aimgr repo add ~/.opencode/agents/reviewer.md    # Auto-detects as agent
```

**Bulk Import from Folders:**
```bash
# Add all resources from a folder (auto-discovers all types)
aimgr repo add ~/.opencode/
aimgr repo add ~/project/.claude/
aimgr repo add ./my-resources/

# With options
aimgr repo add ~/.opencode/ --force         # Overwrite existing
aimgr repo add ./resources/ --skip-existing # Skip conflicts
aimgr repo add ./test/ --dry-run            # Preview without importing
```

**Add from GitHub:**
```bash
# Add all resources from a GitHub repo (auto-discovers)
aimgr repo add gh:owner/repo

# Add from specific path in repo
aimgr repo add gh:owner/repo/skills/pdf-processing

# Add from specific version/tag
aimgr repo add gh:owner/repo@v1.0.0
aimgr repo add gh:owner/repo@main

# Shorthand (infers gh: prefix)
aimgr repo add owner/repo
aimgr repo add owner/repo@v2.0.0

# Add from Git URL
aimgr repo add https://github.com/owner/repo.git
aimgr repo add git@github.com:owner/repo.git
```

**Filter Resources with Patterns:**
```bash
# Add only skills from GitHub repo
aimgr repo add gh:owner/repo --filter "skill/*"

# Add only skills starting with "pdf"
aimgr repo add ./resources/ --filter "skill/pdf*"

# Add only test-related resources
aimgr repo add ~/.claude/ --filter "*test*"

# Add multiple patterns
aimgr repo add gh:owner/repo --filter "skill/*" --filter "command/build*"
```

**Preview and Safety:**
```bash
# Preview what would be added (dry run)
aimgr repo add gh:owner/repo --dry-run

# Preview with filters
aimgr repo add ./resources/ --filter "skill/*" --dry-run

# Force overwrite existing resources
aimgr repo add ~/skills/ --force

# Skip existing resources without error
aimgr repo add gh:owner/repo --skip-existing
```

**Auto-Detection Rules:**

The `repo add` command automatically detects resource types:

1. **Skill:** Directory containing `SKILL.md`
2. **Command:** `.md` file with command frontmatter (description field)
3. **Agent:** `.md` file with agent frontmatter (description + type/instructions/capabilities)

**Add Output (v1.12.0+):**
```
Adding resources to repository...
✓ skill/pdf-processing (from gh:owner/repo)
✓ command/test (from ./resources/test.md)
✓ agent/reviewer (from ~/.claude/agents/reviewer.md)

Successfully added 3 resources to repository.
```

---

### repo update

Update resources in the repository from their original sources.

**Syntax:**
```bash
aimgr repo update [TYPE/NAME] [OPTIONS]
```

**Arguments:**
- `TYPE/NAME` - Optional resource in format `type/name` (if omitted, updates all resources)

**Options:**
- `--dry-run` - Preview updates without applying changes
- `--force` - Force update, overwriting local modifications

**Examples:**

**Update All Resources:**
```bash
# Update all resources from their sources
aimgr repo update

# Preview updates without applying
aimgr repo update --dry-run

# Force update all, overwriting local changes
aimgr repo update --force
```

**Update Specific Resource:**
```bash
# Update specific skill
aimgr repo update skill/pdf-processing

# Update specific command
aimgr repo update command/test

# Update specific agent with force
aimgr repo update agent/code-reviewer --force
```

**Update Output (v1.12.0+ with progress):**
```
Updating resources from sources...
[1/5] ✓ skill/pdf-processing (v1.2.0 -> v1.3.0)
[2/5] - skill/react-testing (already up-to-date)
[3/5] ✓ command/test (updated from gh:owner/repo)
[4/5] ✗ command/build (source unavailable)
[5/5] ✓ agent/reviewer (v2.0.0 -> v2.1.0)

Successfully updated 3 of 5 resources.
```

**Dry Run Output:**
```
Preview of updates (no changes will be applied):
[1/5] skill/pdf-processing: v1.2.0 -> v1.3.0 (gh:owner/repo)
[2/5] skill/react-testing: already up-to-date
[3/5] command/test: local changes detected (use --force to overwrite)
[4/5] command/build: source unavailable
[5/5] agent/reviewer: v2.0.0 -> v2.1.0 (gh:owner/repo)

Would update 2 of 5 resources.
```

**Behavior:**

- Only updates resources that have a tracked source (GitHub, Git URL)
- Resources added from local paths without source tracking cannot be auto-updated
- Local modifications are preserved unless `--force` is used
- Symlinks in projects automatically reflect updated resources (no reinstall needed)

---

### repo remove

Remove resources from the repository. **Warning:** This permanently deletes the resource and breaks existing symlinks in projects.

**Syntax:**
```bash
aimgr repo remove TYPE/NAME [OPTIONS]
```

**Aliases:** `rm`

**Arguments:**
- `TYPE/NAME` - Resource in format `type/name` (e.g., `skill/old-skill`, `command/test`, `agent/old-agent`)

**Options:**
- `--force` - Skip confirmation prompt

**Examples:**

**Basic Removal:**
```bash
# Remove with confirmation prompt
aimgr repo remove skill/old-skill
aimgr repo remove command/test
aimgr repo remove agent/old-agent

# Skip confirmation
aimgr repo remove skill/old-skill --force

# Using alias
aimgr repo rm command/old-test
```

**Confirmation Prompt:**
```
⚠️  WARNING: This will permanently delete the resource from the repository.
Installed symlinks in projects will break.

Resource: skill/old-skill
Path: /home/user/.local/share/ai-config/repo/skills/old-skill

Are you sure you want to remove this resource? [y/N]: 
```

**Remove Output:**
```
Removing resource from repository...
✓ Deleted /home/user/.local/share/ai-config/repo/skills/old-skill
✓ Deleted metadata

Successfully removed skill/old-skill from repository.
```

**⚠️ Impact:**
- Resource is permanently deleted from `~/.local/share/ai-config/repo/`
- Metadata in `.metadata/` is removed
- Symlinks in projects will break (point to non-existent path)
- Projects must run `aimgr uninstall` to clean up broken symlinks

---
