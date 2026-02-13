# aimgr CLI Reference: Discovery Commands

Commands for discovering and listing available resources in the repository and installed in projects.

## Table of Contents

- [repo list](#repo-list) - List all resources in repository
- [list](#list) - List resources installed in current project
- [repo show](#repo-show) - Show detailed information about a resource

---

## repo list

List all resources available in the repository.

**Syntax:**
```bash
aimgr repo list [TYPE] [OPTIONS]
```

**Arguments:**
- `TYPE` - Optional resource type filter: `skill`, `command`, `agent`, or `package`

**Options:**
- `--format=FORMAT` - Output format: `text` (default), `json`, or `yaml`

**Examples:**

```bash
# List all resources in repository
aimgr repo list

# List all resources as JSON (recommended for programmatic use)
aimgr repo list --format=json

# List only skills
aimgr repo list skill

# List only commands with JSON output
aimgr repo list command --format=json

# List only agents with YAML output
aimgr repo list agent --format=yaml
```

**Output Format (JSON):**
```json
{
  "skills": [
    {
      "name": "pdf-processing",
      "description": "Process PDF documents",
      "path": "/home/user/.local/share/ai-config/repo/skills/pdf-processing",
      "source": "gh:owner/repo",
      "version": "1.2.0"
    }
  ],
  "commands": [
    {
      "name": "test",
      "description": "Run tests",
      "path": "/home/user/.local/share/ai-config/repo/commands/test.md"
    }
  ],
  "agents": [
    {
      "name": "code-reviewer",
      "description": "Review code changes",
      "path": "/home/user/.local/share/ai-config/repo/agents/code-reviewer.md"
    }
  ]
}
```

---

## list

List resources installed in the current project.

**Syntax:**
```bash
aimgr list [TYPE] [OPTIONS]
```

**Arguments:**
- `TYPE` - Optional resource type filter: `skill`, `command`, `agent`, or `package`

**Options:**
- `--format=FORMAT` - Output format: `text` (default), `json`, or `yaml`
- `--project-path=PATH` - Specify project directory (default: current directory)

**Examples:**

```bash
# List all installed resources in current project
aimgr list

# List installed resources as JSON
aimgr list --format=json

# List only installed skills
aimgr list skill

# List installed commands in specific project
aimgr list command --project-path /path/to/project
```

**Output Format (JSON):**
```json
{
  "claude": {
    "skills": ["pdf-processing", "react-testing"],
    "commands": ["test", "build"],
    "agents": ["code-reviewer"]
  },
  "opencode": {
    "skills": ["pdf-processing"],
    "commands": ["test"],
    "agents": []
  }
}
```

---

## repo show

Show detailed information about a resource in the repository.

**Syntax:**
```bash
aimgr repo show TYPE/NAME
```

**Arguments:**
- `TYPE/NAME` - Resource in format `type/name` (e.g., `skill/pdf-processing`, `command/test`, `agent/code-reviewer`)

**Examples:**

```bash
# Show skill details
aimgr repo show skill/pdf-processing

# Show command details
aimgr repo show command/test

# Show agent details
aimgr repo show agent/code-reviewer
```

**Output Example:**
```
Name: pdf-processing
Type: skill
Description: Process PDF documents with extraction and analysis
Path: /home/user/.local/share/ai-config/repo/skills/pdf-processing
Source: gh:owner/repo
Version: 1.2.0
Author: john-doe
License: MIT
Last Updated: 2026-01-25T14:30:00Z

Files:
  - SKILL.md (2.3 KB)
  - scripts/extract.py (1.5 KB)
  - README.md (890 B)
```
