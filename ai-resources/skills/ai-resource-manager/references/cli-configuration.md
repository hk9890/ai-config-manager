# aimgr CLI Reference: Configuration

Commands for managing aimgr configuration settings.

## Table of Contents

- [config](#config) - View all configuration
- [config set](#config-set) - Set configuration value
- [config get](#config-get) - Get configuration value

---

## Configuration

### config

View or manage configuration settings.

**Configuration File:** `~/.config/aimgr/aimgr.yaml`

**Syntax:**
```bash
aimgr config [SUBCOMMAND]
```

**Subcommands:**
- (none) - Display all configuration
- `set KEY VALUE` - Set configuration value
- `get KEY` - Get configuration value

**Examples:**

**View Configuration:**
```bash
# Display all configuration
aimgr config
```

**Output:**
```yaml
install:
  targets:
    - claude
    - opencode
repository:
  path: /home/user/.local/share/ai-config/repo
  sync:
    enabled: true
    auto_update: false
```

---

### config set

Set a configuration value.

**Syntax:**
```bash
aimgr config set KEY VALUE
```

**Available Settings:**

| Key | Value | Description |
|-----|-------|-------------|
| `install.targets` | `claude`, `opencode`, `copilot` | Default installation targets (comma-separated) |
| `repository.sync.enabled` | `true`, `false` | Enable repository sync with remote (v1.12.0+) |
| `repository.sync.auto_update` | `true`, `false` | Auto-update on sync (v1.12.0+) |

**Examples:**

**Set Default Installation Targets:**
```bash
# Single tool
aimgr config set install.targets claude

# Multiple tools (comma-separated)
aimgr config set install.targets claude,opencode

# All tools
aimgr config set install.targets claude,opencode,copilot
```

**Set Repository Sync (v1.12.0+):**
```bash
# Enable repository sync
aimgr config set repository.sync.enabled true

# Enable auto-update on sync
aimgr config set repository.sync.auto_update true

# Disable sync
aimgr config set repository.sync.enabled false
```

**Output:**
```
✓ Configuration updated: install.targets = claude,opencode
```

---

### config get

Get a configuration value.

**Syntax:**
```bash
aimgr config get KEY
```

**Examples:**

```bash
# Get default installation targets
aimgr config get install.targets

# Get repository path
aimgr config get repository.path

# Get sync settings
aimgr config get repository.sync.enabled
aimgr config get repository.sync.auto_update
```

**Output:**
```
claude,opencode
```

