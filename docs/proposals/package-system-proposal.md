# Package System Proposal for AI Config Manager

**Date**: 2026-01-25  
**Status**: Proposal  
**Authors**: Research based on Claude Code plugins and ai-config-manager codebase analysis

---

## Executive Summary

This proposal outlines a package system for **aimgr** that enables grouping multiple AI resources (commands, skills, agents) into logical units called **packages**. This is inspired by Claude Code's plugin system but adapted to aimgr's multi-tool architecture (Claude Code, OpenCode, GitHub Copilot).

**Key Goals:**
- Group related resources together (e.g., "beads-workflow" package with commands + agents + skills)
- Simplify installation (`aimgr install package/beads-workflow` instead of 39 individual commands)
- Support dependency management between packages and resources
- Enable marketplace/discovery mechanisms
- Maintain backward compatibility with existing resources

---

## 1. Research Summary: Claude Code Plugins

### 1.1 Claude Code Plugin Architecture

**Plugin Structure:**
```
my-plugin/
├── .claude-plugin/
│   └── plugin.json          # Plugin metadata
├── commands/
│   ├── command1.md
│   └── command2.md
├── skills/
│   └── skill-name/
│       └── SKILL.md
├── agents/
│   ├── agent1.md
│   └── agent2.md
├── .mcp.json               # Optional: MCP server config
└── README.md
```

**Plugin Metadata (`.claude-plugin/plugin.json`):**
```json
{
  "name": "feature-dev",
  "version": "1.0.0",
  "description": "Feature development workflow",
  "author": {
    "name": "Anthropic",
    "email": "support@anthropic.com"
  },
  "homepage": "https://github.com/...",
  "repository": "https://github.com/...",
  "license": "MIT"
}
```

**Installation Tracking:**
- Central registry: `~/.claude/plugins/installed_plugins.json`
- Version-based cache: `~/.claude/plugins/cache/<marketplace>/<plugin-name>/<version>/`
- Supports marketplace scopes (e.g., `claude-code-plugins`, `claude-plugins-official`)

**Key Features:**
1. **Versioning**: Semantic versioning with Git commit tracking
2. **Scopes**: User-level vs. project-level installations
3. **Marketplaces**: Multiple plugin sources (official, community)
4. **Resource Bundling**: Commands, skills, agents in one package
5. **MCP Integration**: Optional MCP server configuration

### 1.2 Existing aimgr Plugin Support

The codebase **already has** plugin detection/scanning code (`pkg/resource/plugin.go`):
- `DetectPlugin()` - Checks for `.claude-plugin/plugin.json`
- `LoadPluginMetadata()` - Parses plugin metadata
- `ScanPluginResources()` - Discovers commands/skills/agents within plugin
- Similar functions for `.claude/` and `.opencode/` folder structures

**Current Limitations:**
- Plugin detection exists, but **not integrated into CLI workflow**
- No package-level metadata tracking
- No dependency management
- No versioning beyond source URL tracking
- Resources stored flat by type (not grouped by package)

---

## 2. Proposed Package System Architecture

### 2.1 Core Concepts

**Package vs. Resource:**
- **Resource**: Atomic unit (command, skill, agent)
- **Package**: Collection of related resources with metadata

**Package Naming Convention:**
- Format: `package/<package-name>` or just `<package-name>` in package contexts
- Examples: `package/beads-workflow`, `package/pdf-processing`, `package/web-scraping`
- Follows agentskills.io naming: lowercase, alphanumeric, hyphens only

**Package Types:**
1. **Standard Package**: Directory with `.aimgr-package/package.json` + resources
2. **Claude Plugin**: Directory with `.claude-plugin/plugin.json` (compatibility)
3. **OpenCode Plugin**: Directory with `.opencode-plugin/plugin.json` (future)

### 2.2 Package Directory Structure

**Option A: aimgr-native format**
```
my-package/
├── .aimgr-package/
│   └── package.json         # Package metadata + manifest
├── commands/
│   ├── cmd1.md
│   └── cmd2.md
├── skills/
│   └── skill-name/
│       └── SKILL.md
├── agents/
│   ├── agent1.md
│   └── agent2.md
└── README.md
```

**Option B: Claude-compatible format**
```
my-package/
├── .claude-plugin/
│   └── plugin.json          # Claude plugin metadata
├── .aimgr-manifest.json     # aimgr-specific extensions
├── commands/
│   └── ...
├── skills/
│   └── ...
└── agents/
    └── ...
```

**Recommendation**: **Option A** for new packages, with automatic detection of Option B for Claude plugin compatibility.

### 2.3 Package Metadata Schema

**File: `.aimgr-package/package.json`**

```json
{
  "name": "beads-workflow",
  "version": "1.0.0",
  "description": "Complete beads issue tracking workflow",
  "author": {
    "name": "Your Name",
    "email": "you@example.com",
    "url": "https://github.com/yourusername"
  },
  "homepage": "https://github.com/yourusername/beads-workflow",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourusername/beads-workflow.git"
  },
  "license": "MIT",
  "keywords": ["workflow", "issue-tracking", "beads"],
  
  "resources": {
    "commands": [
      {
        "name": "beads-init",
        "path": "commands/beads-init.md",
        "description": "Initialize beads repository"
      },
      {
        "name": "beads-create",
        "path": "commands/beads-create.md"
      }
    ],
    "skills": [
      {
        "name": "beads-planning",
        "path": "skills/beads-planning",
        "description": "Strategic planning with beads"
      }
    ],
    "agents": [
      {
        "name": "beads-planner",
        "path": "agents/beads-planner.md"
      },
      {
        "name": "beads-task-agent",
        "path": "agents/beads-task-agent.md"
      }
    ]
  },
  
  "dependencies": {
    "packages": [
      "package/git-helpers@^1.0.0"
    ],
    "commands": [
      "jq",
      "git"
    ]
  },
  
  "tools": {
    "claude": {
      "compatible": true,
      "version": ">=1.0.0"
    },
    "opencode": {
      "compatible": true
    },
    "copilot": {
      "compatible": false,
      "reason": "Uses agents not supported by Copilot"
    }
  },
  
  "config": {
    "mcp": {
      "servers": [
        {
          "name": "custom-server",
          "type": "stdio",
          "command": "node",
          "args": ["server.js"]
        }
      ]
    }
  }
}
```

**Required Fields:**
- `name` - Package name (lowercase, alphanumeric, hyphens)
- `version` - Semantic version (e.g., "1.0.0")
- `description` - Brief description
- `resources` - Manifest of included resources

**Optional Fields:**
- `author` - Author information
- `homepage`, `repository` - URLs
- `license` - License identifier
- `keywords` - Searchable tags
- `dependencies` - Package and system dependencies
- `tools` - Tool compatibility information
- `config` - Additional configuration (MCP servers, etc.)

### 2.4 Repository Storage Structure

**Current Structure:**
```
~/.local/share/ai-config/repo/
├── .metadata/
│   ├── commands/<name>-metadata.json
│   ├── skills/<name>-metadata.json
│   └── agents/<name>-metadata.json
├── commands/<name>.md
├── skills/<name>/SKILL.md
└── agents/<name>.md
```

**Proposed Structure (with packages):**
```
~/.local/share/ai-config/repo/
├── .metadata/
│   ├── commands/<name>-metadata.json
│   ├── skills/<name>-metadata.json
│   ├── agents/<name>-metadata.json
│   └── packages/<name>-metadata.json        # NEW
├── commands/<name>.md
├── skills/<name>/SKILL.md
├── agents/<name>.md
└── packages/                                 # NEW
    └── <package-name>/
        ├── .aimgr-package/package.json
        ├── commands/
        ├── skills/
        └── agents/
```

**Package Metadata File (`.metadata/packages/<name>-metadata.json`):**
```json
{
  "name": "beads-workflow",
  "version": "1.0.0",
  "source_type": "github",
  "source_url": "gh:yourusername/beads-workflow",
  "first_installed": "2026-01-25T10:00:00Z",
  "last_updated": "2026-01-25T10:00:00Z",
  "git_commit_sha": "abc123...",
  "installed_resources": {
    "commands": ["beads-init", "beads-create"],
    "skills": ["beads-planning"],
    "agents": ["beads-planner", "beads-task-agent"]
  }
}
```

---

## 3. CLI Integration

### 3.1 New Commands

#### `aimgr package list`
List all installed packages.

```bash
$ aimgr package list

PACKAGES
Name              Version  Resources         Source
beads-workflow    1.0.0    2 cmd, 1 sk, 2 ag gh:user/beads-workflow
pdf-processing    0.5.0    1 cmd, 1 sk       gh:org/pdf-tools
web-scraping      2.1.0    3 cmd, 2 ag       /local/path
```

#### `aimgr package show <package-name>`
Show detailed package information.

```bash
$ aimgr package show beads-workflow

Package: beads-workflow
Version: 1.0.0
Description: Complete beads issue tracking workflow
Author: Your Name
Source: gh:yourusername/beads-workflow
Installed: 2026-01-25

Resources:
  Commands:
    - beads-init (Initialize beads repository)
    - beads-create (Create beads issues)
  Skills:
    - beads-planning (Strategic planning)
  Agents:
    - beads-planner
    - beads-task-agent

Dependencies:
  - package/git-helpers@^1.0.0

Tool Compatibility:
  ✓ Claude Code
  ✓ OpenCode
  ✗ GitHub Copilot (agents not supported)
```

#### `aimgr package add <source>`
Install a package from a source.

```bash
# From GitHub
$ aimgr package add gh:yourusername/beads-workflow

# From GitHub with version
$ aimgr package add gh:yourusername/beads-workflow@v1.0.0

# From local directory
$ aimgr package add ~/dev/my-package

# With options
$ aimgr package add gh:user/pkg --force --tools=claude,opencode
```

**Options:**
- `--force` - Overwrite existing resources
- `--tools=<tools>` - Install only for specific tools (comma-separated)
- `--skip-deps` - Don't install dependencies
- `--dry-run` - Preview without installing

**Installation Process:**
1. Clone/download package source
2. Parse package.json metadata
3. Validate package structure
4. Check dependencies (install if missing)
5. Copy package to `repo/packages/<name>/`
6. Create package metadata in `.metadata/packages/`
7. Install resources to tools (based on `--tools` or default: all compatible)
8. Create symlinks in tool directories

#### `aimgr package remove <package-name>`
Uninstall a package.

```bash
$ aimgr package remove beads-workflow

# With options
$ aimgr package remove beads-workflow --purge-resources
```

**Options:**
- `--purge-resources` - Also remove all installed resources
- `--keep-data` - Keep package data but remove symlinks

**Removal Process:**
1. Check if other packages depend on this one (warn user)
2. Remove symlinks from tool directories
3. Remove package from `repo/packages/`
4. Remove package metadata
5. Optionally remove installed resources if `--purge-resources`

#### `aimgr package update [package-name]`
Update package(s) to latest version.

```bash
# Update specific package
$ aimgr package update beads-workflow

# Update all packages
$ aimgr package update --all

# Check for updates without installing
$ aimgr package update --check
```

#### `aimgr package create <name>`
Create a new package from existing resources.

```bash
$ aimgr package create my-package

# Interactive prompts:
# - Package name
# - Description
# - Author info
# - Select resources to include
# - Set version
# - Add dependencies
# - Configure tools

# Creates directory structure:
my-package/
├── .aimgr-package/package.json
├── commands/
├── skills/
├── agents/
└── README.md
```

### 3.2 Modified Commands

#### `aimgr install <pattern>`
Enhanced to support package installation.

```bash
# Current: Install individual resources
$ aimgr install command/test-runner

# NEW: Install packages with prefix
$ aimgr install package/beads-workflow

# NEW: Install all resources from a package
$ aimgr install beads-workflow/*

# Pattern matching
$ aimgr install package/*      # Install all packages
$ aimgr install package/web-*  # Install web-* packages
```

#### `aimgr uninstall <pattern>`
Enhanced to support package uninstallation.

```bash
# Uninstall package (removes symlinks, keeps resources)
$ aimgr uninstall package/beads-workflow

# Uninstall and purge resources
$ aimgr uninstall package/beads-workflow --purge
```

#### `aimgr list`
Enhanced to show packages.

```bash
# Current behavior (resources only)
$ aimgr list

# NEW: Show packages too
$ aimgr list --packages

# NEW: Filter by package
$ aimgr list --package=beads-workflow
```

#### `aimgr repo add <source>`
Auto-detect packages during bulk import.

```bash
# Current: Import individual resources
$ aimgr repo add ~/my-resources

# NEW: If directory is a package, import as package
$ aimgr repo add ~/my-package
# Detected package: my-package (3 commands, 1 skill, 2 agents)
# Import as package? [Y/n]
```

---

## 4. Dependency Management

### 4.1 Dependency Types

**1. Package Dependencies**
Packages can depend on other packages.

```json
{
  "dependencies": {
    "packages": [
      "package/git-helpers@^1.0.0",
      "package/json-utils@>=0.5.0"
    ]
  }
}
```

**Version Constraints:**
- `^1.0.0` - Compatible with 1.x.x (>= 1.0.0, < 2.0.0)
- `~1.2.0` - Approximately 1.2.x (>= 1.2.0, < 1.3.0)
- `>=1.0.0` - Greater than or equal to
- `1.0.0` - Exact version
- `*` - Any version (not recommended)

**2. System Dependencies**
Required system commands/tools.

```json
{
  "dependencies": {
    "commands": [
      "git",
      "jq",
      "python3"
    ]
  }
}
```

**3. Resource Dependencies**
Individual resources can depend on other resources (within same package or external).

```json
{
  "resources": {
    "commands": [
      {
        "name": "advanced-cmd",
        "path": "commands/advanced-cmd.md",
        "dependencies": {
          "commands": ["basic-cmd"],
          "skills": ["helper-skill"]
        }
      }
    ]
  }
}
```

### 4.2 Dependency Resolution

**Installation Order:**
1. Check if package already installed
2. Parse package.json
3. Resolve and install package dependencies (recursive)
4. Check system dependencies (warn if missing)
5. Install package resources
6. Verify resource dependencies

**Conflict Resolution:**
- **Same package, different versions**: Prompt user to choose
- **Circular dependencies**: Detect and error
- **Missing dependencies**: Offer to install or error if not found

### 4.3 Dependency Commands

```bash
# Show dependency tree
$ aimgr package deps beads-workflow
beads-workflow@1.0.0
├── package/git-helpers@^1.0.0
│   └── git (system)
└── jq (system)

# Check dependency health
$ aimgr package check beads-workflow
✓ package/git-helpers@1.2.0 installed
✓ git found at /usr/bin/git
✗ jq not found (install: apt install jq)

# Install missing dependencies
$ aimgr package install-deps beads-workflow
```

---

## 5. Marketplace & Discovery

### 5.1 Marketplace Architecture

**Decentralized Approach:**
- No central registry initially
- Use GitHub as primary marketplace
- Support for marketplace URLs in config

**Package Discovery Sources:**
1. **Official Registry** (future): `https://registry.aimgr.io/packages/`
2. **GitHub Topics**: Packages tagged with `aimgr-package`, `claude-plugin`
3. **Awesome Lists**: Curated GitHub repos (e.g., `awesome-aimgr-packages`)
4. **Custom Registries**: User-configured sources

### 5.2 Registry Configuration

**File: `~/.config/aimgr/aimgr.yaml`**

```yaml
registries:
  - name: official
    url: https://registry.aimgr.io/
    priority: 1
    enabled: true
  
  - name: community
    url: https://github.com/topics/aimgr-package
    type: github-topics
    priority: 2
    enabled: true
  
  - name: custom
    url: https://mycompany.com/aimgr-packages/
    priority: 3
    enabled: false
```

### 5.3 Discovery Commands

```bash
# Search for packages
$ aimgr package search beads
Found 3 packages:
  - beads-workflow (1.0.0) - Complete beads workflow
  - beads-cli (0.5.0) - CLI tools for beads
  - beads-helpers (1.2.0) - Utility functions for beads

# Search with filters
$ aimgr package search web --tool=claude --tag=scraping

# Show package details before installing
$ aimgr package info gh:user/beads-workflow
Name: beads-workflow
Version: 1.0.0
...

# Browse available packages
$ aimgr package browse
```

### 5.4 Publishing Packages

**Manual Publishing (GitHub):**
1. Create package repository on GitHub
2. Add topics: `aimgr-package`, `claude-plugin`, etc.
3. Add to `awesome-aimgr-packages` list
4. Users install via `aimgr package add gh:user/repo`

**Future: Official Registry:**
```bash
# Publish to official registry
$ aimgr package publish
# - Validates package.json
# - Checks for conflicts
# - Uploads to registry
# - Tags GitHub release
```

---

## 6. Tool-Specific Considerations

### 6.1 Multi-Tool Installation

Packages can specify tool compatibility:

```json
{
  "tools": {
    "claude": {
      "compatible": true,
      "version": ">=1.0.0"
    },
    "opencode": {
      "compatible": true
    },
    "copilot": {
      "compatible": false,
      "reason": "Uses agents not supported by Copilot"
    }
  }
}
```

**Installation Behavior:**
```bash
# Install for all compatible tools
$ aimgr package add gh:user/pkg

# Install for specific tools only
$ aimgr package add gh:user/pkg --tools=claude

# Skip tool compatibility check
$ aimgr package add gh:user/pkg --force-tools
```

### 6.2 Tool-Specific Resources

Some resources may be tool-specific:

```json
{
  "resources": {
    "commands": [
      {
        "name": "claude-only-cmd",
        "path": "commands/claude-only-cmd.md",
        "tools": ["claude"]
      },
      {
        "name": "universal-cmd",
        "path": "commands/universal-cmd.md",
        "tools": ["claude", "opencode", "copilot"]
      }
    ]
  }
}
```

**Installation Process:**
- Filter resources by tool compatibility
- Only install compatible resources for each tool
- Warn if some resources skipped

---

## 7. Backward Compatibility

### 7.1 Existing Resource Support

**All existing resources continue to work:**
- Individual commands, skills, agents can be installed independently
- No changes to existing `aimgr install`, `aimgr uninstall` behavior for resources
- Repository structure extends, not replaces

### 7.2 Claude Plugin Compatibility

**Automatic Detection:**
- If directory has `.claude-plugin/plugin.json`, treat as package
- Parse Claude plugin metadata into aimgr package format
- Support Claude plugin installation via `aimgr package add`

**Conversion:**
```bash
# Install Claude plugin
$ aimgr package add ~/.claude/plugins/cache/some-plugin

# Converts .claude-plugin/plugin.json → aimgr package metadata
# Installs resources to aimgr repository
```

### 7.3 Migration Path

**For Users:**
1. No action required - existing resources continue to work
2. Gradually adopt packages for new installations
3. Optional: Convert existing resources to packages

**For Resource Authors:**
1. Add `.aimgr-package/package.json` to existing resource repos
2. Group related resources into package structure
3. Publish as package for easier installation

---

## 8. Implementation Phases

### Phase 1: Core Package System (v2.0.0)
**Goal**: Basic package installation and management

**Features:**
- `pkg/resource/package.go` - Package type and loading
- `pkg/metadata/package_metadata.go` - Package metadata tracking
- Package storage in `repo/packages/`
- Commands: `package add`, `package remove`, `package list`, `package show`
- Auto-detection of `.aimgr-package/` and `.claude-plugin/`
- Resource manifest parsing
- Basic version tracking

**Testing:**
- Unit tests for package parsing
- Integration tests for installation/removal
- Test with example packages

**Timeline**: 2-3 weeks

### Phase 2: Dependency Management (v2.1.0)
**Goal**: Handle package and resource dependencies

**Features:**
- Dependency resolution algorithm
- Dependency installation
- Conflict detection
- Commands: `package deps`, `package check`, `package install-deps`
- Version constraint parsing (^, ~, >=, etc.)
- Circular dependency detection

**Testing:**
- Dependency resolution edge cases
- Conflict scenarios
- Version constraint logic

**Timeline**: 2-3 weeks

### Phase 3: Tool Integration (v2.2.0)
**Goal**: Multi-tool support and tool-specific resources

**Features:**
- Tool compatibility checking
- Selective installation by tool
- Tool-specific resource filtering
- `--tools` flag for installation
- Cross-tool package support

**Testing:**
- Installation across different tools
- Tool compatibility validation
- Resource filtering by tool

**Timeline**: 1-2 weeks

### Phase 4: Package Creation & Management (v2.3.0)
**Goal**: Authoring and updating packages

**Features:**
- Commands: `package create`, `package update`
- Interactive package creation wizard
- Package validation
- Update checking and notification
- Batch update support

**Testing:**
- Package creation workflows
- Update scenarios
- Validation edge cases

**Timeline**: 2 weeks

### Phase 5: Discovery & Marketplace (v3.0.0)
**Goal**: Package discovery and distribution

**Features:**
- Commands: `package search`, `package browse`, `package info`
- Registry configuration
- GitHub topic-based discovery
- Package publishing workflow (future)
- Official registry API (future)

**Testing:**
- Search functionality
- Registry integration
- Discovery mechanisms

**Timeline**: 3-4 weeks

---

## 9. Example Use Cases

### Use Case 1: Installing a Package

**User Story**: Install the beads-workflow package to get all beads commands, skills, and agents.

```bash
# Search for beads packages
$ aimgr package search beads
Found: beads-workflow (1.0.0) - Complete beads issue tracking workflow

# Install package
$ aimgr package add gh:yourusername/beads-workflow
Resolving dependencies...
  - package/git-helpers@^1.0.0 → installing git-helpers@1.2.0
Installing beads-workflow@1.0.0...
  ✓ 2 commands installed
  ✓ 1 skill installed
  ✓ 2 agents installed
Installed to: Claude Code, OpenCode

# Verify installation
$ aimgr package show beads-workflow
Package: beads-workflow
Version: 1.0.0
Resources: 2 commands, 1 skill, 2 agents
Status: Installed ✓

# Use the resources
$ claude /beads-init
```

### Use Case 2: Creating a Package

**User Story**: Create a package from existing PDF processing resources.

```bash
# Create package structure
$ aimgr package create pdf-toolkit
Package name: pdf-toolkit
Description: PDF processing tools for AI workflows
Author: Your Name
Email: you@example.com
Version [1.0.0]: 
License [MIT]: 

Select resources to include:
  [✓] command/pdf-extract
  [✓] command/pdf-merge
  [✓] skill/pdf-processing
  [ ] agent/pdf-analyzer

Created package at: ./pdf-toolkit/

# Edit package.json to add metadata
$ cd pdf-toolkit
$ vim .aimgr-package/package.json

# Add package to repository
$ aimgr repo add .
Detected package: pdf-toolkit (2 commands, 1 skill)
Import as package? [Y/n] y
Package added to repository

# Publish to GitHub
$ git init
$ git add .
$ git commit -m "Initial commit"
$ git remote add origin gh:yourusername/pdf-toolkit
$ git push -u origin main
$ git tag v1.0.0
$ git push --tags

# Others can now install
$ aimgr package add gh:yourusername/pdf-toolkit
```

### Use Case 3: Managing Dependencies

**User Story**: Create a web scraping package that depends on common utilities.

**Package: web-scraping**
```json
{
  "name": "web-scraping",
  "version": "1.0.0",
  "dependencies": {
    "packages": [
      "package/http-helpers@^1.0.0",
      "package/json-utils@^0.5.0"
    ],
    "commands": [
      "curl",
      "jq"
    ]
  }
}
```

**Installation:**
```bash
$ aimgr package add gh:user/web-scraping
Resolving dependencies...
  - package/http-helpers@^1.0.0 → not found
  - package/json-utils@^0.5.0 → installing json-utils@0.6.2
Checking system dependencies...
  ✓ curl found at /usr/bin/curl
  ✗ jq not found
  
Install missing packages? [Y/n] y
Installing http-helpers@1.1.0...
  ✓ http-helpers installed

Warning: System dependency 'jq' not found
Install with: apt install jq (Ubuntu/Debian) or brew install jq (macOS)

Continue anyway? [y/N] y
Installing web-scraping@1.0.0...
  ✓ 3 commands installed
  ✓ 1 agent installed
```

### Use Case 4: Tool-Specific Packages

**User Story**: Install a package only for Claude Code.

**Package: advanced-agents (uses Claude-specific features)**
```json
{
  "name": "advanced-agents",
  "tools": {
    "claude": {"compatible": true},
    "opencode": {"compatible": false, "reason": "Uses Claude-specific agent features"},
    "copilot": {"compatible": false, "reason": "No agent support"}
  }
}
```

**Installation:**
```bash
# Auto-install for compatible tools only
$ aimgr package add gh:user/advanced-agents
Warning: Package not compatible with: OpenCode, GitHub Copilot
Install for Claude Code only? [Y/n] y
Installing advanced-agents@1.0.0...
  ✓ Installed to Claude Code
  ✗ Skipped OpenCode (incompatible)
  ✗ Skipped GitHub Copilot (incompatible)

# Or force install for specific tool
$ aimgr package add gh:user/advanced-agents --tools=claude
```

---

## 10. Open Questions & Decisions

### Q1: Package Metadata Format
**Question**: Use `.aimgr-package/package.json` or `.claude-plugin/plugin.json` as primary format?

**Options:**
- **A**: `.aimgr-package/package.json` (aimgr-native, more flexible)
- **B**: `.claude-plugin/plugin.json` (Claude-compatible, limited features)
- **C**: Support both, prefer `.aimgr-package/`

**Recommendation**: **Option C** - Support both formats for maximum compatibility.

### Q2: Repository Storage
**Question**: Store packages as subdirectories or flattened?

**Options:**
- **A**: Subdirectories (`repo/packages/<name>/`) - keeps package structure intact
- **B**: Flattened (`repo/commands/`, `repo/skills/`) - simpler, no duplication

**Recommendation**: **Option A** - Preserves package boundaries, easier for updates/removal.

### Q3: Resource Duplication
**Question**: If a resource exists in multiple packages, how to handle?

**Options:**
- **A**: Allow duplication (each package has its own copy)
- **B**: Deduplicate (shared resources, track multiple owners)
- **C**: Error on conflict (force user to choose)

**Recommendation**: **Option B** with tracking - More efficient, track which packages provide each resource.

### Q4: Version Constraints
**Question**: Support full semver range or simplified constraints?

**Options:**
- **A**: Full semver (^, ~, >=, <, *, ||, etc.)
- **B**: Simplified (^, ~, >=, exact only)

**Recommendation**: **Option B** - Simpler to implement, covers 95% of use cases.

### Q5: Dependency Resolution
**Question**: Install all dependencies automatically or prompt user?

**Options:**
- **A**: Auto-install (silent, fast)
- **B**: Prompt for each (verbose, user control)
- **C**: Prompt once for all (middle ground)

**Recommendation**: **Option C** - Show all deps, ask once, install batch.

### Q6: Package Namespacing
**Question**: How to prevent package name conflicts?

**Options:**
- **A**: No namespacing (first-come-first-served)
- **B**: GitHub-style namespacing (`user/package-name`)
- **C**: Scope-based namespacing (`@scope/package-name`)

**Recommendation**: **Option B** for remote packages, simple names for local - Familiar to developers, clear ownership.

### Q7: Update Strategy
**Question**: Auto-update packages or manual?

**Options:**
- **A**: Auto-update (check on install, prompt to update)
- **B**: Manual only (`package update` command)
- **C**: Configurable (global setting in config)

**Recommendation**: **Option C** - Default to manual, allow auto-update in config.

---

## 11. Comparison: Claude Plugins vs. aimgr Packages

| Feature | Claude Plugins | aimgr Packages | Notes |
|---------|----------------|----------------|-------|
| **Metadata File** | `.claude-plugin/plugin.json` | `.aimgr-package/package.json` | aimgr can read both |
| **Resource Bundling** | Commands, Skills, Agents | Commands, Skills, Agents | Same |
| **Versioning** | Semantic versioning | Semantic versioning | Same |
| **Dependencies** | No built-in support | Package + system deps | aimgr extension |
| **Multi-Tool Support** | Claude Code only | Claude, OpenCode, Copilot | aimgr multi-tool |
| **Installation Scope** | User, Project | User (repo-based) | Different approach |
| **Marketplace** | Official marketplace | Decentralized (GitHub) | Different distribution |
| **Version Cache** | `~/.claude/plugins/cache/` | `~/.local/share/ai-config/repo/packages/` | Different storage |
| **MCP Support** | `.mcp.json` file | In package.json `config.mcp` | aimgr integrates metadata |
| **Update Mechanism** | Auto-update from marketplace | Manual or configured | Different philosophy |

---

## 12. Migration Plan for Existing Users

### For Current aimgr Users

**No Breaking Changes:**
- All existing commands continue to work
- Individual resource installation unchanged
- Repo structure extends, not replaces

**Adoption Path:**
1. **v2.0.0 release**: Package system available
2. **Gradual adoption**: Users choose when to use packages
3. **Optional conversion**: Convert existing resources to packages
4. **Documentation**: Examples of packaging existing resources

### For Claude Plugin Users

**Compatibility:**
- `aimgr package add` can install Claude plugins directly
- Automatic format detection and conversion
- No manual conversion needed

**Example:**
```bash
# Install from Claude plugin format
$ aimgr package add ~/.claude/plugins/cache/some-plugin

# aimgr auto-detects .claude-plugin/plugin.json
# Converts to aimgr package metadata
# Installs resources to repository
```

---

## 13. Success Metrics

**Adoption Metrics:**
- Number of packages created (target: 50+ in first 6 months)
- Package installations vs. individual resources (target: 30% via packages)
- Average resources per package (target: 4-5)

**Quality Metrics:**
- Package installation success rate (target: >95%)
- Dependency resolution failures (target: <5%)
- Update success rate (target: >90%)

**Community Metrics:**
- GitHub stars on example packages (target: 100+ per popular package)
- Community-contributed packages (target: 20+ in first year)
- Package author diversity (target: 10+ authors)

---

## 14. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **Low adoption** | High | Medium | Create compelling example packages, good docs |
| **Dependency hell** | High | Low | Simple version constraints, clear conflict errors |
| **Breaking changes** | High | Low | Maintain backward compatibility, versioning |
| **Security concerns** | High | Medium | Package verification, signing (future), community review |
| **Namespace conflicts** | Medium | Medium | GitHub-style namespacing, conflict detection |
| **Storage bloat** | Low | Medium | Deduplication, optional resource purging |

---

## 15. Next Steps

### Immediate (Week 1)
1. Review and discuss this proposal
2. Decide on open questions (Q1-Q7)
3. Create design doc from proposal
4. Define API contracts for package functions

### Short-term (Weeks 2-4)
1. Implement Phase 1 (Core Package System)
2. Create example packages for testing
3. Write integration tests
4. Update documentation

### Medium-term (Months 2-3)
1. Implement Phase 2 (Dependencies)
2. Implement Phase 3 (Tool Integration)
3. Beta release for early adopters
4. Gather feedback

### Long-term (Months 4-6)
1. Implement Phase 4 (Creation & Management)
2. Implement Phase 5 (Discovery & Marketplace)
3. Official registry planning
4. v3.0.0 release

---

## 16. References

### Research Sources
- Claude Code Plugins: `~/.claude/plugins/`
- Claude Quickstarts: `https://github.com/anthropics/claude-quickstarts`
- aimgr codebase: `pkg/resource/plugin.go`, `pkg/repo/manager.go`
- agentskills.io: Resource naming conventions

### Related Documentation
- `AGENTS.md` - aimgr development guide
- `pkg/resource/` - Resource type implementations
- `pkg/metadata/` - Metadata tracking system
- `test/bulk_import_test.go` - Bulk import patterns

### Similar Systems
- npm (Node.js packages)
- pip (Python packages)
- Claude Code plugins
- VS Code extensions
- Homebrew formulae

---

## Appendix A: Example Package Repository

**Repository: `gh:yourusername/beads-workflow`**

```
beads-workflow/
├── .aimgr-package/
│   └── package.json
├── commands/
│   ├── beads-init.md
│   ├── beads-create.md
│   ├── beads-update.md
│   └── beads-close.md
├── skills/
│   └── beads-planning/
│       └── SKILL.md
├── agents/
│   ├── beads-planner.md
│   ├── beads-task-agent.md
│   ├── beads-review-agent.md
│   └── beads-verify-agent.md
├── examples/
│   ├── basic-workflow.md
│   └── advanced-workflow.md
├── .gitignore
├── README.md
├── LICENSE
└── CHANGELOG.md
```

**`.aimgr-package/package.json`:**
```json
{
  "name": "beads-workflow",
  "version": "1.0.0",
  "description": "Complete beads issue tracking workflow with planning, execution, review, and verification agents",
  "author": {
    "name": "Your Name",
    "email": "you@example.com",
    "url": "https://github.com/yourusername"
  },
  "homepage": "https://github.com/yourusername/beads-workflow",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourusername/beads-workflow.git"
  },
  "license": "MIT",
  "keywords": ["workflow", "issue-tracking", "beads", "project-management", "agents"],
  
  "resources": {
    "commands": [
      {
        "name": "beads-init",
        "path": "commands/beads-init.md",
        "description": "Initialize beads repository in current project"
      },
      {
        "name": "beads-create",
        "path": "commands/beads-create.md",
        "description": "Create beads issues (epic, task, bug, gate)"
      },
      {
        "name": "beads-update",
        "path": "commands/beads-update.md",
        "description": "Update beads issue status and metadata"
      },
      {
        "name": "beads-close",
        "path": "commands/beads-close.md",
        "description": "Close beads issues and tasks"
      }
    ],
    "skills": [
      {
        "name": "beads-planning",
        "path": "skills/beads-planning",
        "description": "Strategic planning and task breakdown with beads"
      }
    ],
    "agents": [
      {
        "name": "beads-planner",
        "path": "agents/beads-planner.md",
        "description": "Planning agent for creating epics and tasks"
      },
      {
        "name": "beads-task-agent",
        "path": "agents/beads-task-agent.md",
        "description": "Execution agent for implementing tasks"
      },
      {
        "name": "beads-review-agent",
        "path": "agents/beads-review-agent.md",
        "description": "Review agent for validating plans"
      },
      {
        "name": "beads-verify-agent",
        "path": "agents/beads-verify-agent.md",
        "description": "Verification agent for gate checks"
      }
    ]
  },
  
  "dependencies": {
    "packages": [],
    "commands": [
      "git",
      "jq"
    ]
  },
  
  "tools": {
    "claude": {
      "compatible": true,
      "version": ">=1.0.0"
    },
    "opencode": {
      "compatible": true

    },
    "copilot": {
      "compatible": false,
      "reason": "Uses agents and skills not supported by GitHub Copilot"
    }
  }
}
```

---

## Appendix B: Package Metadata Schema (JSON Schema)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "aimgr Package Metadata",
  "type": "object",
  "required": ["name", "version", "description", "resources"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$",
      "minLength": 1,
      "maxLength": 64,
      "description": "Package name (lowercase, alphanumeric, hyphens)"
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+(-.+)?$",
      "description": "Semantic version (e.g., '1.0.0', '2.1.0-beta')"
    },
    "description": {
      "type": "string",
      "minLength": 1,
      "maxLength": 1024,
      "description": "Brief package description"
    },
    "author": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"},
        "url": {"type": "string", "format": "uri"}
      }
    },
    "homepage": {
      "type": "string",
      "format": "uri"
    },
    "repository": {
      "type": "object",
      "properties": {
        "type": {"type": "string", "enum": ["git"]},
        "url": {"type": "string", "format": "uri"}
      }
    },
    "license": {
      "type": "string"
    },
    "keywords": {
      "type": "array",
      "items": {"type": "string"}
    },
    "resources": {
      "type": "object",
      "properties": {
        "commands": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "path"],
            "properties": {
              "name": {"type": "string"},
              "path": {"type": "string"},
              "description": {"type": "string"},
              "dependencies": {"$ref": "#/definitions/resourceDependencies"},
              "tools": {"type": "array", "items": {"type": "string"}}
            }
          }
        },
        "skills": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "path"],
            "properties": {
              "name": {"type": "string"},
              "path": {"type": "string"},
              "description": {"type": "string"},
              "dependencies": {"$ref": "#/definitions/resourceDependencies"},
              "tools": {"type": "array", "items": {"type": "string"}}
            }
          }
        },
        "agents": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "path"],
            "properties": {
              "name": {"type": "string"},
              "path": {"type": "string"},
              "description": {"type": "string"},
              "dependencies": {"$ref": "#/definitions/resourceDependencies"},
              "tools": {"type": "array", "items": {"type": "string"}}
            }
          }
        }
      }
    },
    "dependencies": {
      "type": "object",
      "properties": {
        "packages": {
          "type": "array",
          "items": {"type": "string", "pattern": "^package/[a-z0-9-]+@[~^>=<]?\\d+\\.\\d+\\.\\d+$"}
        },
        "commands": {
          "type": "array",
          "items": {"type": "string"}
        }
      }
    },
    "tools": {
      "type": "object",
      "properties": {
        "claude": {"$ref": "#/definitions/toolCompatibility"},
        "opencode": {"$ref": "#/definitions/toolCompatibility"},
        "copilot": {"$ref": "#/definitions/toolCompatibility"}
      }
    },
    "config": {
      "type": "object",
      "properties": {
        "mcp": {
          "type": "object",
          "properties": {
            "servers": {
              "type": "array",
              "items": {
                "type": "object",
                "required": ["name", "type"],
                "properties": {
                  "name": {"type": "string"},
                  "type": {"type": "string", "enum": ["stdio", "http"]},
                  "command": {"type": "string"},
                  "args": {"type": "array", "items": {"type": "string"}},
                  "url": {"type": "string", "format": "uri"},
                  "headers": {"type": "object"}
                }
              }
            }
          }
        }
      }
    }
  },
  "definitions": {
    "resourceDependencies": {
      "type": "object",
      "properties": {
        "commands": {"type": "array", "items": {"type": "string"}},
        "skills": {"type": "array", "items": {"type": "string"}},
        "agents": {"type": "array", "items": {"type": "string"}}
      }
    },
    "toolCompatibility": {
      "type": "object",
      "required": ["compatible"],
      "properties": {
        "compatible": {"type": "boolean"},
        "version": {"type": "string"},
        "reason": {"type": "string"}
      }
    }
  }
}
```

---

**End of Proposal**
