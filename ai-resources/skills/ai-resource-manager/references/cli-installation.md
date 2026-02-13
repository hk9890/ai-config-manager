# aimgr CLI Reference: Installation & Removal Commands

Commands for installing and removing resources in projects.

## Table of Contents

- [install](#install) - Install resources to project
- [uninstall](#uninstall) - Remove resources from project
- [Pattern Syntax](#pattern-syntax) - Pattern matching for bulk operations

---
## Installation Commands

### install

Install resources from the repository to the current project.

**Syntax:**
```bash
aimgr install PATTERN... [OPTIONS]
```

**Arguments:**
- `PATTERN` - Resource pattern(s) to install (supports wildcards, see [Pattern Syntax](#pattern-syntax))
  - Format: `type/pattern` (e.g., `skill/pdf-processing`)
  - Or: `pattern` (searches all types)

**Options:**
- `--target=TOOL` - Override target tool(s): `claude`, `opencode`, `copilot` (comma-separated for multiple)
- `--force` - Force reinstall, overwriting existing resources
- `--project-path=PATH` - Specify project directory (default: current directory)

**Examples:**

**Basic Installation:**
```bash
# Install using type/name format (NEW v1.12.0+ preferred format)
aimgr install skill/react-testing
aimgr install command/prettier-config
aimgr install agent/code-reviewer

# Install multiple resources at once
aimgr install skill/pdf-processing command/test agent/reviewer
```

**Pattern-Based Installation:**
```bash
# Install all skills
aimgr install "skill/*"

# Install all resources with "test" in name
aimgr install "*test*"

# Install skills starting with "pdf"
aimgr install "skill/pdf*"

# Install multiple patterns
aimgr install "command/test*" "agent/qa*"

# Install build and test commands
aimgr install "command/{build,test}"
```

**Target Override:**
```bash
# Install to specific tool only
aimgr install skill/utils --target claude

# Install to multiple tools
aimgr install command/test --target claude,opencode

# Install to all supported tools
aimgr install skill/react-testing --target claude,opencode,copilot
```

**Force Installation:**
```bash
# Force reinstall, overwriting existing
aimgr install skill/utils --force

# Force reinstall with specific target
aimgr install skill/utils --force --target claude
```

**Custom Project Path:**
```bash
# Install to specific project
aimgr install skill/foo --project-path /path/to/project

# Install patterns to another project
aimgr install "skill/*" --project-path ~/projects/my-app
```

**Behavior:**

1. **Fresh Project (no tool directories):**
   - Uses configured `install.targets` from config file
   - Creates tool directories automatically
   
2. **Existing Tool Directory:**
   - Auto-detects existing `.claude/`, `.opencode/`, or `.github/skills/`
   - Installs to all detected directories
   - If both `.claude/` and `.opencode/` exist, installs to both
   
3. **Target Override:**
   - `--target` flag overrides all auto-detection
   - Creates specified directories if they don't exist

**Installation Output (v1.12.0+):**
```
Installing resources to project...
✓ skill/pdf-processing -> .claude/skills/pdf-processing
✓ skill/pdf-processing -> .opencode/skills/pdf-processing
✓ command/test -> .claude/commands/test.md
✓ command/test -> .opencode/commands/test.md

Successfully installed 2 resources to 2 tools.
```

---

## Removal Commands

### uninstall

Remove resources from the current project. **Only removes symlinked resources** - manually created resources are never removed.

**Syntax:**
```bash
aimgr uninstall PATTERN... [OPTIONS]
```

**Arguments:**
- `PATTERN` - Resource pattern(s) to uninstall (supports wildcards, see [Pattern Syntax](#pattern-syntax))
  - Format: `type/pattern` (e.g., `skill/react-testing`)
  - Or: `pattern` (searches all types)

**Options:**
- `--project-path=PATH` - Specify project directory (default: current directory)

**Examples:**

**Basic Removal:**
```bash
# Uninstall using type/name format
aimgr uninstall skill/react-testing
aimgr uninstall command/prettier-config
aimgr uninstall agent/code-reviewer

# Uninstall multiple resources
aimgr uninstall skill/foo skill/bar command/test
```

**Pattern-Based Removal:**
```bash
# Uninstall all skills
aimgr uninstall "skill/*"

# Uninstall all test-related resources
aimgr uninstall "*test*"

# Uninstall legacy skills
aimgr uninstall "skill/legacy-*"

# Uninstall build and test commands
aimgr uninstall "command/{build,test}"
```

**Custom Project Path:**
```bash
# Uninstall from specific project
aimgr uninstall skill/foo --project-path /path/to/project
```

**What Gets Removed:**
- ✅ Symlinks from project directories (`.claude/`, `.opencode/`, `.github/`)
- ✅ Removes from ALL tool directories if present
- ❌ Original resource in `~/.local/share/ai-config/repo/` remains intact
- ❌ Manually created resources (NOT symlinks) are never removed

**⚠️ CRITICAL SAFETY CHECK:**

**ALWAYS verify a resource is a symlink before removing:**

```bash
# Check if resource is a symlink (look for ->)
ls -la .opencode/skills/
ls -la .claude/commands/
ls -la .opencode/agents/

# Example output:
# skill-creator -> /home/user/.local/share/ai-config/repo/skills/skill-creator
#   ↑ Safe to remove (symlink indicated by ->)
#
# my-custom-skill/
#   ↑ DO NOT remove (manually created, no -> symbol)
```

**Rule:** If you see `->` in the output, it's a symlink and safe to remove. If not, it's manually created and should NOT be removed via uninstall.

**Uninstall Output:**
```
Uninstalling resources from project...
✓ Removed .claude/skills/pdf-processing
✓ Removed .opencode/skills/pdf-processing
✓ Removed .claude/commands/test.md

Successfully uninstalled 2 resources from 2 tools.
```

