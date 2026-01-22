# 🎉 aimgr v1.0.0 - Major Release

> **Major milestone:** Production-ready AI resource management tool with modern architecture

**aimgr** (AI Manager) v1.0.0 is a complete refactoring that delivers a feature-complete, production-ready CLI tool for managing AI resources across Claude Code, OpenCode, and GitHub Copilot.

---

## ⚡ Quick Overview

| Aspect | Before (`ai-repo`) | After (`aimgr`) |
|--------|-------------------|-----------------|
| **Binary name** | `ai-repo` | `aimgr` |
| **Command structure** | `ai-repo add skill foo` | `aimgr repo add skill foo` |
| **Install syntax** | `ai-repo install skill foo` | `aimgr install skill/foo` |
| **Config location** | `~/.aimgr.yaml` | `~/.config/aimgr/aimgr.yaml` |
| **Config format** | `default-tool: claude` | `install.targets: [claude]` |

**Key improvements:**
- 🏗️ Reorganized command structure with `repo` group
- 📊 Metadata tracking for resource sources
- 🔄 Update resources from original sources
- 📦 Enhanced install with type prefix syntax
- 🎯 Multi-tool support with multiple default targets

---

## 🚨 Breaking Changes

> **⚠️ IMPORTANT:** This release includes significant breaking changes.  
> **Read the [Migration Guide](MIGRATION.md) before upgrading.**

<table>
<tr><td>

### 🔧 Binary Name Change

```bash
# Before
ai-repo --version

# After  
aimgr --version
```

**Action required:** Replace `ai-repo` binary with `aimgr`

</td></tr>
<tr><td>

### 🗂️ Command Structure Reorganization

Repository management commands now use the `repo` subcommand:

```bash
# Before
ai-repo add skill pdf-processing
ai-repo list
ai-repo remove command old-test

# After
aimgr repo add skill pdf-processing
aimgr repo list
aimgr repo remove command old-test
```

**Action required:** Add `repo` prefix to repository commands

</td></tr>
<tr><td>

### 📦 Install/Uninstall Syntax

New type prefix syntax (`type/name`):

```bash
# Before
ai-repo install skill pdf-processing
ai-repo install command test

# After
aimgr install skill/pdf-processing
aimgr install command/test
```

**Action required:** Use `type/name` format for install/uninstall

</td></tr>
<tr><td>

### ⚙️ Configuration Changes

| Change | Details |
|--------|---------|
| **Location** | `~/.aimgr.yaml` → `~/.config/aimgr/aimgr.yaml` (XDG) |
| **Format** | `default-tool` → `install.targets` |
| **Migration** | Automatic on first run ✅ |

```yaml
# Before
default-tool: claude

# After
install:
  targets: [claude]
```

**Action required:** None - auto-migration handles this

</td></tr>
</table>

---

## ✨ Major Features

### 1️⃣ Reorganized Command Structure

Clear separation between repository management and project operations:

<table>
<tr>
<td width="50%">

**Repository Management**
```bash
aimgr repo add skill gh:vercel-labs/agent-skills
aimgr repo list
aimgr repo show skill pdf-processing
aimgr repo update
aimgr repo remove skill old-skill
```

</td>
<td width="50%">

**Project Operations**
```bash
aimgr install skill/pdf-processing
aimgr uninstall skill/old-skill
aimgr config set install.targets claude,opencode
```

</td>
</tr>
</table>

---

### 2️⃣ Metadata Tracking System

Automatic tracking of resource sources enables powerful update functionality.

**What's tracked:**
- 🔗 Source type (GitHub, local, file)
- 📍 Source URL or path
- 🏷️ Git references (branches/tags)
- 🕒 Timestamps (added, updated)

**Benefits:**
- ♻️ Update resources from original sources
- 📋 View detailed resource information
- 🔍 Track resource provenance

```bash
# View resource metadata
aimgr repo show skill pdf-processing

# Update from source
aimgr repo update skill pdf-processing
aimgr repo update  # Update all resources
```

---

### 3️⃣ Enhanced Install Command

**New capabilities:**

| Feature | Description | Example |
|---------|-------------|---------|
| Type prefix syntax | Use `type/name` format | `skill/pdf`, `command/test`, `agent/reviewer` |
| Multiple resources | Install several at once | `aimgr install skill/foo skill/bar command/test` |
| Target override | Override default targets | `aimgr install skill/utils --target claude,opencode` |

```bash
# Install multiple resources at once
aimgr install skill/foo skill/bar command/test agent/reviewer

# Override installation targets
aimgr install skill/utils --target claude,opencode
```

---

### 4️⃣ Resource Update Command

Keep resources in sync with their original sources (GitHub, local paths):

```bash
aimgr repo update                          # Update all resources
aimgr repo update skill pdf-processing     # Update specific resource
aimgr repo update --dry-run                # Preview updates
aimgr repo update --force                  # Force update, overwrite changes
```

---

### 5️⃣ Resource Details Command

Comprehensive information about any resource:

```bash
aimgr repo show skill pdf-processing
aimgr repo show command test
aimgr repo show agent code-reviewer
```

**Output includes:**
- 📝 Name, type, description
- 🏷️ Version, author, license
- 🔗 Source information (URL, path, ref)
- ✅ Installation status
- 🗂️ Full metadata

---

### 6️⃣ Improved Configuration Management

**Features:**
- ✅ Multiple default targets - install to multiple tools by default
- ✅ XDG compliance - config stored in `~/.config/aimgr/`
- ✅ Auto-migration - seamless upgrade from old config format

```bash
# Set multiple default installation targets
aimgr config set install.targets claude,opencode

# View current settings
aimgr config get install.targets
```

---

## 🆕 What's New

<table>
<tr><td>

### 🐙 GitHub Source Support <sub>v0.3.0</sub>

Add resources directly from GitHub with powerful auto-discovery:

- ✅ Auto-discovery in 13+ standard locations
- ✅ Support for branches, tags, and subpaths
- ✅ Interactive selection for multi-resource repos

```bash
aimgr repo add skill gh:vercel-labs/agent-skills
aimgr repo add skill gh:owner/repo/path/to/skill@v1.0.0
aimgr repo add skill owner/repo  # Shorthand
```

</td></tr>
<tr><td>

### 🤖 Agent Support <sub>v0.2.x</sub>

Full support for AI agents across OpenCode and Claude formats:

- ✅ OpenCode and Claude format compatibility
- ✅ Auto-discovery from GitHub and local sources
- ✅ Bulk import from `.claude/agents/` and `.opencode/agents/`

```bash
aimgr repo add agent gh:myorg/agents/code-reviewer
aimgr install agent/code-reviewer
```

</td></tr>
<tr><td>

### 🔧 OpenCode Tool Support <sub>v0.2.x</sub>

First-class support for OpenCode alongside Claude Code:

- ✅ Bulk import from `.opencode/` directories
- ✅ Agent format compatibility (type, instructions, capabilities)
- ✅ Multi-tool default targets

```bash
aimgr repo add opencode ~/.opencode
aimgr config set install.targets opencode
```

</td></tr>
<tr><td>

### 📦 Bulk Import <sub>v0.2.x</sub>

Import all resources from tool directories in one command:

```bash
aimgr repo add claude ~/.claude                 # Import from Claude folder
aimgr repo add opencode ~/.opencode             # Import from OpenCode folder
aimgr repo add plugin ~/.claude/plugins/.../my-plugin  # Import from plugin
```

**Conflict handling options:**

| Flag | Behavior |
|------|----------|
| `--force` | Overwrite existing resources |
| `--skip-existing` | Skip conflicts silently |
| `--dry-run` | Preview without importing |

</td></tr>
<tr><td>

### ⌨️ Shell Completion <sub>v0.2.x</sub>

Dynamic tab completion for resource names:

```bash
# Install completion (Bash example)
aimgr completion bash > /etc/bash_completion.d/aimgr

# Use completion
aimgr install skill/<TAB>     # Shows all available skills
aimgr install command/<TAB>   # Shows all available commands
```

**Supported shells:** Bash, Zsh, Fish, PowerShell

</td></tr>
</table>

---

## 🔧 Improvements

<table>
<tr>
<td width="50%">

### 🗂️ Command Organization
- Clear separation: `repo` for repository, root for projects
- Consistent command naming and structure
- Improved help text and error messages

### 📦 Installation Behavior
- **Fresh projects:** Uses configured default targets
- **Existing tool dirs:** Installs to detected tools
- **Multiple tools:** Installs to all detected directories
- **Target override:** `--target` flag for explicit control

</td>
<td width="50%">

### 🚨 Error Handling
- Better error messages with context
- Helpful suggestions for common issues
- Validation at command invocation time

### ✅ Testing
- **200+** unit tests
- Integration tests for major workflows
- End-to-end tests with real repositories
- **100%** pass rate across all test suites

</td>
</tr>
</table>

---

## 📚 Documentation

### 🆕 New Documentation
- **[MIGRATION.md](MIGRATION.md)** - Complete migration guide from ai-repo to aimgr
- Enhanced CONTRIBUTING.md with architecture details
- Comprehensive command reference

### 🔄 Updated Documentation
- ✅ All examples use new command structure
- ✅ Breaking changes clearly highlighted
- ✅ Troubleshooting section expanded
- ✅ Multi-tool scenarios documented

---

## 🐛 Bug Fixes

- ✅ Fixed module path to match GitHub repository
- ✅ Fixed install command to respect existing tool directories
- ✅ Fixed integration tests for configuration management
- ✅ Fixed remaining references to old command structure
- ✅ Fixed subpath handling for GitHub sources

---

## 📊 Development Statistics

<table>
<tr>
<td width="33%">

### 💼 Development Effort
- **Epic:** ai-config-manager-6hn
- **Tasks:** 15+ completed
- **Code:** 2,500+ lines
- **Docs:** 1,500+ lines
- **Tests:** 200+ unit, 9 integration

</td>
<td width="33%">

### 📅 Feature Timeline
- **v0.1.0** - Initial release
- **v0.1.1** - Bug fixes
- **v0.2.0** - Multi-tool support
- **v0.2.1** - Documentation
- **v0.3.0** - GitHub sources
- **v0.3.1** - Module path fix
- **v1.0.0** - Major refactor ⭐

</td>
<td width="33%">

### ✅ Completed Tasks
- ✅ Metadata tracking
- ✅ `repo` command group
- ✅ Type prefix syntax
- ✅ Resource updates
- ✅ Resource details
- ✅ Documentation

</td>
</tr>
</table>

---

## 🚀 Getting Started

### 📥 Installation

<table>
<tr><td>

**Linux (amd64)**
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.0.0/aimgr_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

</td></tr>
<tr><td>

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.0.0/aimgr_1.0.0_darwin_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

</td></tr>
<tr><td>

**From source**
```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
make install
```

</td></tr>
</table>

### ⚡ Quick Start

```bash
# 1. Configure default targets
aimgr config set install.targets claude

# 2. Add resources from GitHub
aimgr repo add skill gh:vercel-labs/agent-skills

# 3. List available resources
aimgr repo list

# 4. Install in your project
cd ~/my-project
aimgr install skill/pdf-processing command/test agent/code-reviewer

# 5. Update resources from their sources
aimgr repo update
```

---

## 🔄 Migration Guide

### 📋 For Existing Users

<table>
<tr><td>

**Step 1: Install new binary**
```bash
# Download aimgr v1.0.0 from releases
# Remove old ai-repo binary
sudo rm /usr/local/bin/ai-repo
```

</td></tr>
<tr><td>

**Step 2: Configuration auto-migrates** ✅
- Old config at `~/.aimgr.yaml` is automatically migrated
- New location: `~/.config/aimgr/aimgr.yaml`
- `default-tool` → `install.targets` conversion is automatic
- **No manual action required!**

</td></tr>
<tr><td>

**Step 3: Update scripts and aliases**
```bash
# Update shell scripts
sed -i 's/ai-repo /aimgr repo /g' ~/scripts/*.sh

# Update shell aliases
# Before: alias air='ai-repo'
# After:  alias air='aimgr repo'
```

</td></tr>
<tr><td>

**Step 4: Repository is fully compatible** ✅
- ✅ Existing repository at `~/.local/share/ai-config/repo/` works as-is
- ✅ All existing symlinks continue to function
- ✅ No need to reinstall resources

</td></tr>
</table>

### 🆕 For New Users

Simply install **aimgr v1.0.0** and follow the [Quick Start](#-quick-start) guide above!

---

## 📖 Resources

| Document | Description |
|----------|-------------|
| **[README.md](README.md)** | Full documentation and feature guide |
| **[MIGRATION.md](MIGRATION.md)** | Detailed migration instructions |
| **[CONTRIBUTING.md](CONTRIBUTING.md)** | Development guide and architecture |
| **[examples/](examples/)** | Example resources (commands, skills, agents) |

---

## 🙏 Acknowledgments

| Project | Contribution |
|---------|-------------|
| **Claude Code** | Command and agent format specifications |
| **agentskills.io** | Skill format specification |
| **OpenCode** | Multi-tool ecosystem support |
| **Vercel** | Inspiration from add-skill tool |
| **Contributors** | Feedback and testing |

---

## 💬 Support & Community

<table>
<tr>
<td width="33%">

**🐛 Issues**  
[Report bugs](https://github.com/hk9890/ai-config-manager/issues)

</td>
<td width="33%">

**💭 Discussions**  
[Ask questions](https://github.com/hk9890/ai-config-manager/discussions)

</td>
<td width="33%">

**📚 Documentation**  
[Read the docs](https://github.com/hk9890/ai-config-manager)

</td>
</tr>
</table>

---

## 🎯 What's Next

**Future enhancements planned:**

- 🪟 Windows support (junction instead of symlinks)
- 🔍 Resource search functionality
- 🦊 GitLab source support
- 🏷️ Resource versioning and compatibility checks
- 🛒 Marketplace integration

---

## 🎊 Thank You!

Thank you for using **aimgr**! We hope this major release provides a solid foundation for managing AI resources across your projects.

> **Full Changelog:** [v0.3.1...v1.0.0](https://github.com/hk9890/ai-config-manager/compare/v0.3.1...v1.0.0)

---

**Ready to get started?** See the [Quick Start](#-quick-start) guide above! 🚀
