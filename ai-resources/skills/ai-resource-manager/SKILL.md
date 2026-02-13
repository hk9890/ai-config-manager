---
name: ai-resource-manager
description: "Manage AI resources (skills, commands, agents) using aimgr CLI. Use when user asks to: (1) Install/uninstall resources, (2) Manage repository (import/sync/remove), (3) Troubleshoot aimgr issues."
---

# AI Resource Manager

Manage AI resources using `aimgr` CLI. Resources are stored once in `~/.local/share/ai-config/repo/` and symlinked to projects.

---

## Quick Reference

```bash
# Install/Uninstall
aimgr list                      # Show installed resources
aimgr repo list                 # Show available resources
aimgr install skill/name        # Install resource
aimgr uninstall skill/name      # Remove resource

# Repository Management
aimgr repo import ./path        # Import from local directory
aimgr repo import gh:user/repo  # Import from GitHub
aimgr repo sync                 # Sync from configured sources
aimgr repo remove skill/name    # Remove from repository
```

📚 **Command syntax:** Run `aimgr [command] --help` for detailed usage and examples

---

## Use Case 1: Install/Uninstall Resources

Install skills, commands, or agents to your current project.

### Workflow

**1. List available resources:**
```bash
aimgr repo list --format=json
```

Parse JSON and present to user in friendly format (don't dump raw JSON).

**2. Install resources:**
```bash
# Single resource
aimgr install skill/pdf-processing

# Multiple resources
aimgr install skill/pdf-processing command/test agent/reviewer

# Multiple tools (install to Claude, OpenCode, and Copilot)
aimgr install skill/pdf-processing --target=claude,opencode,copilot

# Pattern matching
aimgr install "skill/pdf*"
```

**3. Verify installation:**
```bash
aimgr list
```

**4. ⚠️ CRITICAL: Restart Reminder**

**ALWAYS remind users to restart their AI tool after installation.**

Skills load at startup. Users must close and reopen Claude Code/OpenCode/VS Code/Windsurf to activate new resources.

**Template:**
```
⚠️ **Restart Required:** Close and reopen [Tool Name] to load the new resources.
```

### Uninstall

```bash
aimgr uninstall skill/name
```

📚 **Detailed syntax:** `aimgr install --help` and `aimgr uninstall --help`

---

## Use Case 2: Manage Repository

Import, sync, or remove resources from the global repository.

### Key Operations

```bash
# Import resources
aimgr repo import ~/my-skills/          # Local directory
aimgr repo import gh:user/repo          # GitHub
aimgr repo import gh:user/repo@v1.0.0   # Specific version

# Sync from configured sources (reads sync.sources from ~/.config/aimgr/aimgr.yaml)
aimgr repo sync                         # All configured sources
aimgr repo sync --skip-existing         # Don't overwrite existing

# Remove resources
aimgr repo remove skill/name            # With confirmation
aimgr repo remove skill/name --force    # Skip confirmation
```

**⚠️ Safety Note:** `aimgr repo remove` permanently deletes resources and breaks symlinks.

**Note on sync vs import:**
- `repo import` - One-time import from a specific source
- `repo sync` - Recurring sync from all configured sources in `~/.config/aimgr/aimgr.yaml`

📚 **Detailed syntax:** `aimgr repo --help` for all repository commands

---

## Use Case 3: Troubleshooting

Common issues and quick fixes:

| Issue | Fix |
|-------|-----|
| Skills not loading | Restart AI tool |
| aimgr not found | Install aimgr (see below) |
| Resource not found | `aimgr repo sync` |
| Broken symlinks | `aimgr uninstall skill/name && aimgr install skill/name` |
| Permission denied | `chmod +x $(which aimgr)` |

### Install aimgr

**Using Go (Recommended):**
```bash
go install github.com/hk9890/ai-config-manager@latest
```

**From Source:**
```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
make install  # Installs to ~/bin
```

📚 **Complete troubleshooting guide:** [references/troubleshooting.md](references/troubleshooting.md)

---

## Additional Resources

**Built-in Help:**
- `aimgr --help` - Overview of all commands
- `aimgr install --help` - Install command with examples
- `aimgr repo --help` - Repository management commands
- `aimgr config --help` - Configuration options
- `aimgr [command] --help` - Help for any command

**Documentation:**
- [troubleshooting.md](references/troubleshooting.md) - Troubleshooting guide

**Supported Tools:**

| Tool | Skills | Commands | Agents |
|------|--------|----------|--------|
| Claude Code | ✅ | ✅ | ✅ |
| OpenCode | ✅ | ✅ | ✅ |
| GitHub Copilot | ✅ | ❌ | ❌ |
| Windsurf | ✅ | ❌ | ❌ |

**Links:**
- Repository: https://github.com/hk9890/ai-config-manager
- Issues: https://github.com/hk9890/ai-config-manager/issues
