# aimgr list Command

The `aimgr list` command shows resources that are installed in the current project directory (or specified path).

## Overview

This command is different from `aimgr repo list`:
- **`aimgr repo list`**: Shows resources in the centralized repository (`~/.local/share/ai-config/repo/`)
- **`aimgr list`**: Shows resources installed in the current/specified project directory

Only resources installed via `aimgr install` (symlinks) are shown. Manually copied files are excluded.

## Usage

```bash
aimgr list [command|skill|agent] [flags]
```

## Arguments

- **[type]** (optional): Filter by resource type
  - `command` - Show only installed commands
  - `skill` - Show only installed skills
  - `agent` - Show only installed agents

## Flags

- `--format string` - Output format: `table`, `json`, or `yaml` (default: `table`)
- `--path string` - Project directory path (default: current directory)

## Examples

### Basic Usage

List all installed resources in current directory:
```bash
aimgr list
```

Output:
```
┌───────┬───────────────┬──────────────────┬────────────────────────────────────┐
│ TYPE  │     NAME      │     TARGETS      │            DESCRIPTION             │
├───────┼───────────────┼──────────────────┼────────────────────────────────────┤
│ skill │ skill-creator │ claude, opencode │ Guide for creating effective ski...│
└───────┴───────────────┴──────────────────┴────────────────────────────────────┘
```

### Filter by Type

List only installed skills:
```bash
aimgr list skill
```

List only installed commands:
```bash
aimgr list command
```

List only installed agents:
```bash
aimgr list agent
```

### Different Output Formats

JSON format:
```bash
aimgr list --format=json
```

Output:
```json
[
  {
    "type": "skill",
    "name": "skill-creator",
    "description": "Guide for creating effective skills...",
    "targets": [
      "claude",
      "opencode"
    ]
  }
]
```

YAML format:
```bash
aimgr list --format=yaml
```

Output:
```yaml
- type: skill
  name: skill-creator
  description: Guide for creating effective skills...
  targets:
    - claude
    - opencode
```

### Specify Path

List installed resources in a different directory:
```bash
aimgr list --path ~/my-project
aimgr list --path /path/to/project
```

## Target Tools

The `targets` column shows which tools each resource is installed to:
- **`claude`** - Installed in `.claude/` directory
- **`opencode`** - Installed in `.opencode/` directory
- **`copilot`** - Installed in `.github/skills/` directory

A resource can be installed to multiple tools simultaneously.

## Empty Results

If no tool directories exist:
```
No tool directories found in this project.

Expected directories: .claude, .opencode, or .github/skills
Install resources with: aimgr install <resource>
```

If no resources are installed:
```
No resources installed in this project.

Install resources with: aimgr install <resource>
```

## Symlink Detection

The command only shows resources installed via `aimgr install`, which creates symlinks to the centralized repository.

**Detected (shown):**
- Resources symlinked from repository via `aimgr install`

**Not detected (hidden):**
- Manually copied files
- Locally created resources
- Non-symlink files

This ensures you see only resources managed by aimgr.

## Integration with Install/Uninstall

The `list` command integrates naturally with install and uninstall workflows:

```bash
# Install a resource
aimgr install skill/my-skill

# Check what's installed
aimgr list

# Uninstall a resource
aimgr uninstall skill/my-skill

# Verify removal
aimgr list
```

## Comparison with `aimgr repo list`

| Command | Scope | Use Case |
|---------|-------|----------|
| `aimgr repo list` | Shows resources in the centralized repository | Browse available resources to install |
| `aimgr list` | Shows resources installed in current project | See what's installed in your project |

Typical workflow:
```bash
# 1. Browse available resources
aimgr repo list

# 2. Install desired resources
aimgr install skill/my-skill

# 3. Verify installation
aimgr list

# 4. Check in specific project
cd ~/my-project
aimgr list
```

## Exit Codes

- **0**: Success
- **1**: Error (invalid format, failed to detect tools, etc.)

## Related Commands

- [`aimgr install`](install-command.md) - Install resources to a project
- [`aimgr uninstall`](uninstall-command.md) - Uninstall resources from a project
- [`aimgr repo list`](repo-list-command.md) - List resources in the repository
