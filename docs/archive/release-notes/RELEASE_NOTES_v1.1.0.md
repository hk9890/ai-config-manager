# Release v1.1.0

## 🎉 New Features

### `aimgr list` - List Installed Resources

A new command to view resources installed in your current project directory!

```bash
# List all installed resources
aimgr list

# Filter by type
aimgr list skill
aimgr list command
aimgr list agent

# Output in different formats
aimgr list --format=json
aimgr list --format=yaml

# List in specific directory
aimgr list --path ~/my-project
```

**Key Features:**
- Shows which tools (claude, opencode, copilot) each resource is installed to
- Only displays resources installed via `aimgr install` (symlinks)
- Ignores manually copied files
- Supports filtering by resource type
- Multiple output formats (table, JSON, YAML)

**Example Output:**
```
┌───────┬───────────────┬──────────────────┬────────────────────────────┐
│ TYPE  │     NAME      │     TARGETS      │        DESCRIPTION         │
├───────┼───────────────┼──────────────────┼────────────────────────────┤
│ skill │ skill-creator │ claude, opencode │ Guide for creating skills  │
│ skill │ pdf-reader    │ claude           │ PDF processing skill       │
└───────┴───────────────┴──────────────────┴────────────────────────────┘
```

**Difference from `aimgr repo list`:**
- `aimgr repo list` - Shows resources in centralized repository
- `aimgr list` - Shows resources installed in your project

### `aimgr repo info` - Repository Overview

Get a quick overview of your repository statistics:

```bash
aimgr repo info
```

Shows:
- Total counts by resource type (commands, skills, agents)
- Repository location
- Quick summary of available resources

## 🐛 Bug Fixes

- Fixed tab completion for `repo add` and `repo remove` subcommands
- Improved multi-resource installation UI with better table formatting

## 📚 Documentation

### New Documentation
- **`removed in v1.9.0`** - Comprehensive guide for the new `list` command
- **`docs/shell-completion-troubleshooting.md`** - Troubleshooting guide for shell completion issues
- **`dev-completion.sh`** - Development helper script for enabling completion with `./aimgr`

### Updated Documentation
- README.md - Added `aimgr list` command reference
- Added source format examples to `repo add` help text
- Improved release notes readability

## 🔧 Improvements

- **Better UI**: Multi-resource selection now uses table format for improved readability
- **Development Workflow**: Added `dev-completion.sh` script for easier development
  - Usage: `source dev-completion.sh` to enable completion for `./aimgr`
  - Works with both bash and zsh

## 📦 Installation

### Upgrade Existing Installation

```bash
# From source
cd ai-config-manager
git pull
make install

# Using Go
go install github.com/hk9890/ai-config-manager@v1.1.0
```

### New Installation

Download the latest release for your platform from the [Releases page](https://github.com/hk9890/ai-config-manager/releases/tag/v1.1.0).

**Linux (amd64)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.1.0/aimgr_v1.1.0_linux_amd64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

**macOS (Apple Silicon)**:
```bash
curl -L https://github.com/hk9890/ai-config-manager/releases/download/v1.1.0/aimgr_v1.1.0_darwin_arm64.tar.gz | tar xz
sudo mv aimgr /usr/local/bin/
```

See the [README](https://github.com/hk9890/ai-config-manager#installation) for all installation options.

## 🔄 Migration Notes

No breaking changes in this release. All existing commands and workflows continue to work as before.

## 📝 Commits Since v1.0.0

- Add shell completion troubleshooting guide and dev helper (c377800)
- Add 'aimgr list' command to show installed resources (81531a0)
- Add 'aimgr repo info' command for repository overview (568db97)
- Improve multi-resource selection UI with table format (0c1eb5d)
- Fix: add tab completion to repo add/remove subcommands (2f4cdfe)
- Docs: add source format examples to repo add help text (449b1e7)
- Docs: polish v1.0.0 release notes for better readability (863a1db)

## 🙏 Acknowledgments

Thank you to everyone who provided feedback and helped improve aimgr!

## 🐛 Found a Bug?

Please [open an issue](https://github.com/hk9890/ai-config-manager/issues/new) on GitHub.

## 📖 Full Changelog

See the [full changelog](https://github.com/hk9890/ai-config-manager/compare/v1.0.0...v1.1.0) on GitHub.
