# Release v1.0.0 - aimgr Major Refactor

This is a **major milestone release** marking the completion of a comprehensive refactoring effort. Version 1.0.0 represents a production-ready, feature-complete AI resource management tool with a modernized architecture and command structure.

## 🚨 Breaking Changes

This release includes significant breaking changes. Please read the [Migration Guide](MIGRATION.md) for detailed upgrade instructions.

### Binary Name Change
- **Old:** `ai-repo`
- **New:** `aimgr` (AI Manager)

```bash
# Before
ai-repo --version

# After
aimgr --version
```

### Command Structure Reorganization
Repository management commands now require the `repo` subcommand prefix:

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

### Install/Uninstall Syntax
Installing and uninstalling now uses type prefix syntax (`type/name`):

```bash
# Before
ai-repo install skill pdf-processing
ai-repo install command test

# After
aimgr install skill/pdf-processing
aimgr install command/test
```

### Configuration Changes
- **Config location:** `~/.config/aimgr/aimgr.yaml` (XDG Base Directory)
- **Config format:** `install.targets` replaces `default-tool`
- **Auto-migration:** Existing configs are automatically migrated on first run

```yaml
# Before
default-tool: claude

# After
install:
  targets: [claude]
```

## 🎉 Major Features

### 1. Reorganized Command Structure
New `repo` command group provides clear separation between repository management and project operations:

```bash
# Repository management
aimgr repo add skill gh:vercel-labs/agent-skills
aimgr repo list
aimgr repo show skill pdf-processing
aimgr repo update
aimgr repo remove skill old-skill

# Project operations
aimgr install skill/pdf-processing
aimgr uninstall skill/old-skill
aimgr config set install.targets claude,opencode
```

### 2. Metadata Tracking System
Automatic tracking of resource sources enables update functionality:

**What's tracked:**
- Source type (GitHub, local, file)
- Source URL or path
- Git references (branches/tags)
- Timestamps (added, updated)

**Benefits:**
- Update resources from their original sources
- View detailed resource information
- Track resource provenance

```bash
# View resource metadata
aimgr repo show skill pdf-processing

# Update from source
aimgr repo update skill pdf-processing
aimgr repo update  # Update all resources
```

### 3. Enhanced Install Command
- **Type prefix syntax:** `skill/name`, `command/name`, `agent/name`
- **Multiple resources:** Install multiple resources in one command
- **Target override:** `--target` flag overrides default targets

```bash
# Install multiple resources at once
aimgr install skill/foo skill/bar command/test agent/reviewer

# Override installation targets
aimgr install skill/utils --target claude,opencode
```

### 4. Resource Update Command
Sync resources from their original sources (GitHub, local paths):

```bash
# Update all resources
aimgr repo update

# Update specific resource
aimgr repo update skill pdf-processing

# Preview updates (dry run)
aimgr repo update --dry-run

# Force update, overwriting local changes
aimgr repo update --force
```

### 5. Resource Details Command
View comprehensive information about resources:

```bash
aimgr repo show skill pdf-processing
aimgr repo show command test
aimgr repo show agent code-reviewer
```

**Output includes:**
- Name, type, description
- Version, author, license
- Source information (URL, path, ref)
- Installation status
- Full metadata

### 6. Improved Configuration Management
- **Multiple default targets:** Install to multiple tools by default
- **XDG compliance:** Config stored in `~/.config/aimgr/`
- **Auto-migration:** Seamless upgrade from old config format

```bash
# Set multiple default installation targets
aimgr config set install.targets claude,opencode

# View current settings
aimgr config get install.targets
```

## 📦 What's New

### GitHub Source Support (v0.3.0)
- Add resources directly from GitHub repositories
- Auto-discovery in 13+ standard locations
- Support for branches, tags, and subpaths
- Interactive selection for multi-resource repos

```bash
aimgr repo add skill gh:vercel-labs/agent-skills
aimgr repo add skill gh:owner/repo/path/to/skill@v1.0.0
aimgr repo add skill owner/repo  # Shorthand
```

### Agent Support (v0.2.x)
- Full support for AI agents (OpenCode and Claude formats)
- Agent auto-discovery from GitHub and local sources
- Bulk import from `.claude/agents/` and `.opencode/agents/`

```bash
aimgr repo add agent gh:myorg/agents/code-reviewer
aimgr install agent/code-reviewer
```

### OpenCode Tool Support (v0.2.x)
- First-class support for OpenCode alongside Claude Code
- Bulk import from `.opencode/` directories
- Agent format compatibility (type, instructions, capabilities)

```bash
aimgr repo add opencode ~/.opencode
aimgr config set install.targets opencode
```

### Bulk Import (v0.2.x)
Import all resources from tool directories in one command:

```bash
# Import from Claude folder
aimgr repo add claude ~/.claude

# Import from OpenCode folder
aimgr repo add opencode ~/.opencode

# Import from Claude plugin
aimgr repo add plugin ~/.claude/plugins/.../my-plugin
```

**Conflict handling:**
- `--force`: Overwrite existing resources
- `--skip-existing`: Skip conflicts silently
- `--dry-run`: Preview without importing

### Shell Completion (v0.2.x)
Dynamic tab completion for resource names:

```bash
# Install completion (Bash example)
aimgr completion bash > /etc/bash_completion.d/aimgr

# Use completion
aimgr install skill/<TAB>  # Shows all available skills
aimgr install command/<TAB>  # Shows all available commands
```

Supports Bash, Zsh, Fish, and PowerShell.

## 🔧 Improvements

### Command Organization
- Clear separation: `repo` for repository, root for projects
- Consistent command naming and structure
- Improved help text and error messages

### Installation Behavior
- **Fresh projects:** Uses configured default targets
- **Existing tool dirs:** Installs to detected tools
- **Multiple tools:** Installs to all detected directories
- **Target override:** `--target` flag for explicit control

### Error Handling
- Better error messages with context
- Helpful suggestions for common issues
- Validation at command invocation time

### Testing
- Comprehensive test coverage (200+ unit tests)
- Integration tests for major workflows
- End-to-end tests with real repositories
- 100% pass rate across all test suites

## 📚 Documentation

### New Documentation
- **[MIGRATION.md](MIGRATION.md)** - Complete migration guide from ai-repo to aimgr
- Updated README with all new features
- Enhanced CONTRIBUTING.md with architecture details
- Comprehensive command reference

### Updated Documentation
- All examples use new command structure
- Breaking changes clearly highlighted
- Troubleshooting section expanded
- Multi-tool scenarios documented

## 🐛 Bug Fixes

- Fixed module path to match GitHub repository
- Fixed install command to respect existing tool directories
- Fixed integration tests for configuration management
- Fixed remaining references to old command structure
- Fixed subpath handling for GitHub sources

## 📊 Statistics

### Development Effort
- **Epic:** ai-config-manager-6hn (aimgr refactor)
- **Tasks completed:** 15+ tasks across multiple epics
- **Code changes:** 2,500+ lines of new/refactored code
- **Documentation:** 1,500+ lines of documentation
- **Tests:** 200+ unit tests, 9 integration tests

### Feature Timeline
- **v0.1.0:** Initial release with basic repository management
- **v0.1.1:** Bug fixes and configuration improvements
- **v0.2.0:** Multi-tool support, bulk import, agents
- **v0.2.1:** Documentation fixes
- **v0.3.0:** GitHub source support with auto-discovery
- **v0.3.1:** Module path fix
- **v1.0.0:** Major refactor with new command structure

### Closed Issues
All epic tasks completed:
- ✅ Metadata tracking system
- ✅ New `repo` command group
- ✅ Type prefix syntax for install/uninstall
- ✅ Resource update functionality
- ✅ Resource details command
- ✅ Comprehensive documentation

## 🚀 Getting Started

### Installation

**Linux (amd64):**
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.0.0/aimgr_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.0.0/aimgr_1.0.0_darwin_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**From source:**
```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
make install
```

### Quick Start

```bash
# Configure default targets
aimgr config set install.targets claude

# Add resources from GitHub
aimgr repo add skill gh:vercel-labs/agent-skills

# List available resources
aimgr repo list

# Install in your project
cd ~/my-project
aimgr install skill/pdf-processing command/test agent/code-reviewer

# Update resources from their sources
aimgr repo update
```

## 🔄 Migration Path

### For Existing Users

1. **Install new binary**
   ```bash
   # Download aimgr v1.0.0
   # Remove old ai-repo binary
   sudo rm /usr/local/bin/ai-repo
   ```

2. **Configuration auto-migrates**
   - Old config at `~/.aimgr.yaml` is automatically migrated
   - New location: `~/.config/aimgr/aimgr.yaml`
   - `default-tool` → `install.targets` conversion is automatic

3. **Update scripts and aliases**
   ```bash
   # Update shell scripts
   sed -i 's/ai-repo /aimgr repo /g' ~/scripts/*.sh
   
   # Update shell aliases
   # Before: alias air='ai-repo'
   # After:  alias air='aimgr repo'
   ```

4. **Repository is fully compatible**
   - Your existing repository at `~/.local/share/ai-config/repo/` works as-is
   - All existing symlinks continue to function
   - No need to reinstall resources

### For New Users

Simply install aimgr v1.0.0 and follow the Quick Start guide above!

## 📖 Further Reading

- **[README.md](README.md)** - Full documentation and feature guide
- **[MIGRATION.md](MIGRATION.md)** - Detailed migration instructions
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guide and architecture
- **[examples/](examples/)** - Example resources (commands, skills, agents)

## 🙏 Acknowledgments

- **Claude Code** - Command and agent format specifications
- **agentskills.io** - Skill format specification
- **OpenCode** - Multi-tool ecosystem support
- **Vercel** - Inspiration from add-skill tool
- **Contributors** - All who provided feedback and testing

## 💬 Support

- **Issues:** https://github.com/hk9890/ai-config-manager/issues
- **Discussions:** https://github.com/hk9890/ai-config-manager/discussions
- **Documentation:** https://github.com/hk9890/ai-config-manager

## 🎯 What's Next

Future enhancements planned:
- Windows support (junction instead of symlinks)
- Resource search functionality
- GitLab source support
- Resource versioning and compatibility checks
- Marketplace integration

Thank you for using aimgr! We hope this major release provides a solid foundation for managing AI resources across your projects.

---

**Full Changelog:** v0.3.1...v1.0.0
