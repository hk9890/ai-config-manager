# Package System - Final Design

**Date**: 2026-01-25  
**Status**: Ready for Implementation  
**Version**: 2.0 (Updated with finalized decisions)

---

## Executive Summary

This document defines the final package system design for **aimgr**. Packages are a new resource type that group existing resources (commands, skills, agents) for convenient bulk installation and removal.

**Core Principles:**
1. Packages are a **new resource type** alongside commands, skills, agents
2. Packages contain **references** to resources using `type/name` format
3. Installation/removal works at **package level** (bulk operations)
4. `aimgr list` shows **resources only**, not packages
5. Packages are **stored as JSON files** in the repo

**Key Decisions:**
- ✅ Use `type/name` format in resources array (e.g., `"skill/pdf-processing"`)
- ✅ Auto-discover packages from `packages/` folder in repos
- ✅ Support Claude plugin/marketplace format conversion
- ✅ Update only package.json, not referenced resources
- ✅ Track package metadata (source, timestamps)
- ✅ No validation on add (lazy validation at install time)

---

## Package as Resource Type

### Resource Type Hierarchy

```
Resource Types:
├── command     (individual .md file)
├── skill       (directory with SKILL.md)
├── agent       (individual .md file)
└── package     (NEW: .package.json file)
```

### Package Storage Location

**Repository structure:**
```
~/.local/share/ai-config/repo/
├── commands/
│   ├── cmd1.md
│   └── cmd2.md
├── skills/
│   ├── skill1/
│   │   └── SKILL.md
│   └── skill2/
│       └── SKILL.md
├── agents/
│   ├── agent1.md
│   └── agent2.md
├── packages/                           # NEW
│   ├── dynatrace-rca.package.json      # NEW
│   ├── pdf-toolkit.package.json        # NEW
│   └── web-scraping.package.json       # NEW
└── .metadata/
    └── packages/                       # NEW
        ├── dynatrace-rca-metadata.json
        └── pdf-toolkit-metadata.json
```

**Key points:**
- Packages stored in `repo/packages/`
- Metadata stored in `repo/.metadata/packages/`
- Each package is a single JSON file: `<package-name>.package.json`
- Naming: `package-name` follows agentskills.io rules (lowercase, hyphens, alphanumeric)

---

## Package File Format

### Updated Schema (v2.0)

**Location**: `~/.local/share/ai-config/repo/packages/<package-name>.package.json`

**Schema:**
```json
{
  "name": "package-name",
  "description": "Brief description of the package",
  "resources": [
    "command/command-name-1",
    "command/command-name-2",
    "skill/skill-name-1",
    "skill/skill-name-2",
    "agent/agent-name-1"
  ]
}
```

**Changes from v1.0:**
- **Resources is now a flat array** (not nested object)
- **Uses `type/name` format** for each resource (e.g., `"skill/pdf-processing"`)
- **Self-documenting** - type is explicit in each entry

**Fields:**
- `name` (required): Package name (must match filename without `.package.json`)
- `description` (required): Human-readable description
- `resources` (required): Array of resource references in `type/name` format

**Constraints:**
- Package CANNOT contain other packages (no nesting)
- Resource format: `type/name` where type is `command`, `skill`, or `agent`
- Resource names reference existing resources in repo (validated at install time)
- Empty resources array is valid

### Example: dynatrace-rca.package.json

```json
{
  "name": "dynatrace-rca",
  "description": "Dynatrace root cause analysis tools and workflows",
  "resources": [
    "command/dt-analyze-rca",
    "command/dt-trace-analyzer",
    "skill/dynatrace-dql-core",
    "skill/dynatrace-log-analysis",
    "agent/dynatrace-rca-agent"
  ]
}
```

### Example: pdf-toolkit.package.json

```json
{
  "name": "pdf-toolkit",
  "description": "PDF processing tools for document analysis",
  "resources": [
    "command/pdf-extract",
    "command/pdf-merge",
    "command/pdf-split",
    "skill/pdf-processing"
  ]
}
```

**Benefits of `type/name` format:**
- ✅ Easy to copy/paste from `aimgr list` output
- ✅ Self-documenting (type is explicit)
- ✅ Consistent with CLI syntax (`aimgr install skill/name`)
- ✅ Simpler structure (flat array)
- ✅ Easier to parse and validate

---

## Project-Level Commands

### `aimgr install package/<package-name>`

**Install all resources from a package to project tools.**

**Syntax:**
```bash
aimgr install package/<package-name> [flags]
```

**Examples:**
```bash
# Install all resources from dynatrace-rca package
$ aimgr install package/dynatrace-rca

# Install to specific tools only
$ aimgr install package/dynatrace-rca --tool=claude

# Preview without installing
$ aimgr install package/dynatrace-rca --dry-run

# Force overwrite existing resources
$ aimgr install package/dynatrace-rca --force
```

**Behavior:**
1. Read `~/.local/share/ai-config/repo/packages/dynatrace-rca.package.json`
2. For each resource in `resources` array:
   - Parse `type/name` format
   - Check if resource exists in repo
   - If exists, install to project (symlink to `.claude/`, `.opencode/`, etc.)
   - If not exists, warn and continue with remaining resources
3. Skip already-installed resources (unless `--force`)

**Output:**
```
Installing package: dynatrace-rca
  ✓ command/dt-analyze-rca
  ✓ command/dt-trace-analyzer
  ✓ skill/dynatrace-dql-core
  ✗ skill/dynatrace-log-analysis - not found in repo
  ✓ agent/dynatrace-rca-agent

Warning: 1 resource not found (skill/dynatrace-log-analysis)
Installed 4 of 5 resources from package dynatrace-rca
```

**Flags:**
- `--tool=<tools>` - Install to specific tools only (comma-separated: claude,opencode)
- `--force` - Overwrite existing resources
- `--dry-run` - Show what would be installed without installing

**Error cases:**
- Package not found: `Error: Package 'dynatrace-rca' not found in repository`
- Resource not found: Warn and continue (install available resources)
- Invalid format: `Error: Invalid resource format 'bad-format' (expected type/name)`

---

### `aimgr uninstall package/<package-name>`

**Remove all resources from a package from project tools.**

**Syntax:**
```bash
aimgr uninstall package/<package-name> [flags]
```

**Examples:**
```bash
# Uninstall all resources from dynatrace-rca package
$ aimgr uninstall package/dynatrace-rca

# Uninstall from specific tools only
$ aimgr uninstall package/dynatrace-rca --tool=claude

# Preview without uninstalling
$ aimgr uninstall package/dynatrace-rca --dry-run
```

**Behavior:**
1. Read `~/.local/share/ai-config/repo/packages/dynatrace-rca.package.json`
2. For each resource in package:
   - Remove from project tools (delete symlinks from `.claude/`, `.opencode/`, etc.)
   - **Does NOT care** if resource was installed by another package
   - If resource not installed, skip silently
3. **Does NOT** remove resources from repo (only from project)

**Output:**
```
Uninstalling package: dynatrace-rca
  ✓ command/dt-analyze-rca
  ✓ command/dt-trace-analyzer
  ✓ skill/dynatrace-dql-core
  ○ skill/dynatrace-log-analysis - not installed
  ✓ agent/dynatrace-rca-agent
Uninstalled 4 resources from package dynatrace-rca
```

**Important:**
- **Aggressive removal**: Removes resources even if installed by other means
- **No reference counting**: Doesn't track which package installed what
- User is responsible for managing shared resources

**Flags:**
- `--tool=<tools>` - Uninstall from specific tools only
- `--dry-run` - Show what would be uninstalled

---

### `aimgr list`

**List installed resources (commands, skills, agents) - does NOT show packages.**

**Behavior:**
- Shows commands, skills, agents (unchanged from current behavior)
- **Does NOT** show packages
- No indication of which package a resource came from

**Example:**
```bash
$ aimgr list

COMMANDS
  dt-analyze-rca
  dt-trace-analyzer
  pdf-extract
  ...

SKILLS
  dynatrace-dql-core
  pdf-processing
  ...

AGENTS
  dynatrace-rca-agent
  ...
```

**Rationale:**
- Packages are grouping/installation mechanism, not installed entities
- After installation, resources are standalone
- Users manage resources, not packages

---

## Repository-Level Commands

### `aimgr repo list`

**List all resources in repository, including packages.**

**Syntax:**
```bash
aimgr repo list [--type=<type>]
```

**Examples:**
```bash
# List all resources (including packages)
$ aimgr repo list

COMMANDS (5)
  dt-analyze-rca
  dt-trace-analyzer
  pdf-extract
  ...

SKILLS (3)
  dynatrace-dql-core
  dynatrace-log-analysis
  pdf-processing

AGENTS (2)
  dynatrace-rca-agent
  pdf-analyzer-agent

PACKAGES (2)                          # NEW
  dynatrace-rca (5 resources)
  pdf-toolkit (4 resources)

# List only packages
$ aimgr repo list --type=package

PACKAGES
  dynatrace-rca          5 resources    Dynatrace root cause analysis tools
  pdf-toolkit            4 resources    PDF processing tools
  web-scraping           3 resources    Web scraping utilities
```

**Output format:**
```
PACKAGES
  <name>    <count> resources    <description>
```

---

### `aimgr repo add <source>` - FINALIZED

**Add resources or packages to repository with auto-discovery.**

**Decision: Auto-discovery + Adapter Pattern**

**Behavior:**

#### 1. Package Auto-Discovery (NEW)

When source contains `packages/` folder, auto-import all `*.package.json` files:

```bash
# Source directory structure:
my-repo/
├── packages/
│   ├── toolkit-a.package.json
│   └── toolkit-b.package.json
├── commands/
│   └── cmd1.md
└── skills/
    └── skill1/SKILL.md

# Auto-discovery behavior:
$ aimgr repo add gh:user/my-repo

Discovering resources...
  Found 1 command
  Found 1 skill
  Found 2 packages

Adding commands...
  ✓ cmd1

Adding skills...
  ✓ skill1

Adding packages...
  ✓ toolkit-a (3 resources)
  ✓ toolkit-b (2 resources)

Summary: Added 1 command, 1 skill, 2 packages
```

**Package files are:**
1. Copied to `repo/packages/<name>.package.json`
2. Metadata created in `.metadata/packages/<name>-metadata.json`

#### 2. Claude Plugin/Marketplace Adapter (NEW)

Detect and convert Claude plugin formats to aimgr packages:

**Supported formats:**
- Claude Desktop Plugin (`claude-plugin.json`)
- Claude Marketplace Manifest (`manifest.json`)
- VS Code Extension (`package.json` with Claude extensions)

**Example conversion:**

```bash
# Source contains Claude plugin
$ aimgr repo add gh:anthropic/claude-plugin-example

Detected Claude plugin format
Converting to aimgr package...
  ✓ Converted commands: 3
  ✓ Converted skills: 2
  ✓ Created package: claude-plugin-example

Added package 'claude-plugin-example' (5 resources)
```

**Conversion mapping:**
```
Claude Plugin          →  aimgr Package
─────────────────────────────────────
.claude/commands/*.md  →  command/name
.claude/skills/*/      →  skill/name
manifest commands      →  command/name
manifest tools         →  skill/name
```

#### 3. Individual Resource (Existing Behavior)

```bash
# Still works as before
$ aimgr repo add gh:user/commands/my-command
$ aimgr repo add ./local/skill/
```

**Filter support:**
```bash
# Only import packages
$ aimgr repo add gh:user/my-repo --filter "package/*"

# Only import specific resources
$ aimgr repo add gh:user/my-repo --filter "skill/*"
```

---

### `aimgr repo update` - FINALIZED

**Update packages and resources from source.**

**Decision: Package definition only (no cascade)**

**For packages:**
```bash
$ aimgr repo update package/dynatrace-rca
```

**Behavior:**
1. Check metadata for source URL
2. Re-fetch package.json file from source
3. Update `packages/<name>.package.json`
4. **Does NOT** update referenced resources
5. Update metadata timestamps

**Output:**
```
Updating package: dynatrace-rca
  Source: gh:user/dynatrace-rca-package
  ✓ Package definition updated
  
Changes:
  + Added: skill/dynatrace-metrics-analysis
  - Removed: command/dt-old-tool
  
Package now contains 6 resources (was 5)

Note: Resources themselves were not updated.
Run 'aimgr repo update <resource>' to update individual resources.
```

**For resources (existing behavior):**
```bash
$ aimgr repo update skill/dynatrace-dql-core
$ aimgr repo update command/dt-analyze-rca
```

---

### `aimgr repo remove package/<package-name>`

**Remove package from repository.**

**Syntax:**
```bash
aimgr repo remove package/<package-name> [--with-resources]
```

**Examples:**
```bash
# Remove package definition only (keep resources)
$ aimgr repo remove package/dynatrace-rca

# Remove package and all its resources
$ aimgr repo remove package/dynatrace-rca --with-resources
```

**Behavior (without `--with-resources`):**
1. Delete `packages/<package-name>.package.json`
2. Delete metadata file
3. Keep all resources in repo (commands/, skills/, agents/)

**Behavior (with `--with-resources`):**
1. Read package file
2. Delete all resources referenced by package
3. Delete package file and metadata

**Warning:**
```
$ aimgr repo remove package/dynatrace-rca --with-resources
Warning: This will remove 5 resources:
  - command/dt-analyze-rca
  - command/dt-trace-analyzer
  - skill/dynatrace-dql-core
  - skill/dynatrace-log-analysis
  - agent/dynatrace-rca-agent

These resources may be used by other packages or installed in projects.
Continue? [y/N]:
```

---

### `aimgr repo create-package <package-name>`

**Create a new package from existing resources in repo.**

**Syntax:**
```bash
aimgr repo create-package <package-name> [flags]
```

**Examples:**
```bash
# Interactive creation
$ aimgr repo create-package dynatrace-rca
Package name: dynatrace-rca
Description: Dynatrace root cause analysis tools

Select resources (space to toggle, enter to confirm):
  [x] command/dt-analyze-rca
  [x] command/dt-trace-analyzer
  [x] skill/dynatrace-dql-core
  [x] skill/dynatrace-log-analysis
  [x] agent/dynatrace-rca-agent
  [ ] command/unrelated-cmd

Created package: dynatrace-rca (5 resources)
Saved to: ~/.local/share/ai-config/repo/packages/dynatrace-rca.package.json

# Non-interactive with flags
$ aimgr repo create-package dynatrace-rca \
  --description="Dynatrace RCA tools" \
  --resources="command/dt-analyze-rca,command/dt-trace-analyzer,skill/dynatrace-dql-core,skill/dynatrace-log-analysis,agent/dynatrace-rca-agent"
```

**Flags:**
- `--description=<desc>` - Package description (required for non-interactive)
- `--resources=<list>` - Comma-separated resource list in `type/name` format
- `--force` - Overwrite existing package

**Output file:**
```json
{
  "name": "dynatrace-rca",
  "description": "Dynatrace root cause analysis tools",
  "resources": [
    "command/dt-analyze-rca",
    "command/dt-trace-analyzer",
    "skill/dynatrace-dql-core",
    "skill/dynatrace-log-analysis",
    "agent/dynatrace-rca-agent"
  ]
}
```

**Metadata file:**
```json
{
  "name": "dynatrace-rca",
  "source_type": "manual",
  "source_url": null,
  "first_added": "2026-01-25T10:00:00Z",
  "last_updated": "2026-01-25T10:00:00Z",
  "resource_count": 5
}
```

---

## Package Metadata - FINALIZED

**Decision: Track package metadata (Option B)**

**Metadata location:** `.metadata/packages/<package-name>-metadata.json`

**Schema:**
```json
{
  "name": "dynatrace-rca",
  "source_type": "github|local|manual|claude-plugin",
  "source_url": "gh:user/dynatrace-rca-package",
  "source_ref": "main",
  "first_added": "2026-01-25T10:00:00Z",
  "last_updated": "2026-01-25T12:30:00Z",
  "resource_count": 5,
  "original_format": "aimgr|claude-plugin|claude-marketplace"
}
```

**Fields:**
- `name`: Package name
- `source_type`: Where package came from
  - `github`: GitHub repository
  - `local`: Local directory
  - `manual`: Created with `repo create-package`
  - `claude-plugin`: Converted from Claude plugin
- `source_url`: Original source URL (null for manual)
- `source_ref`: Git ref/branch (for version control)
- `first_added`: When package was added to repo
- `last_updated`: Last update timestamp
- `resource_count`: Number of resources in package
- `original_format`: Original format if converted

**Benefits:**
- ✅ Enables `repo update` for packages
- ✅ Tracks package provenance
- ✅ Consistent with existing resource metadata
- ✅ Supports conversion tracking (Claude plugins)
- ✅ Low overhead (one JSON file per package)

---

## Validation Strategy - FINALIZED

**Decision: No validation on add, lazy validation at install time**

### Package Creation (`repo create-package`)
- **No validation** - accept any resource reference
- User is responsible for correctness
- Fast, simple implementation

### Package Addition (`repo add`)
- **No validation** - copy package file as-is
- Accept any format
- Errors surface at install time

### Package Installation (`install package/<name>`)
- **Lazy validation** - check resources exist during install
- Warn for missing resources, continue with available ones
- Clear error messages with suggestions

**Example:**
```bash
$ aimgr install package/my-package
Installing package: my-package
  ✓ command/cmd1
  ✗ skill/missing-skill - not found in repo
  ✓ agent/agent1

Warning: 1 resource not found
Installed 2 of 3 resources

Tip: Check available skills with 'aimgr repo list --type=skill'
```

**Rationale:**
- **Simple implementation** - no complex validation logic
- **User-friendly** - doesn't block on minor issues
- **Clear feedback** - errors happen where user can fix them
- **Flexible** - allows partial installs

**Future enhancement:** Optional validation command
```bash
$ aimgr repo validate package/my-package
Validating package: my-package
  ✓ command/cmd1 exists
  ✗ skill/missing-skill not found
  ✓ agent/agent1 exists

Validation failed: 1 resource not found
```

---

## Command Summary

### New Commands

| Command | Description | Level |
|---------|-------------|-------|
| `aimgr install package/<name>` | Install all resources from package | Project |
| `aimgr uninstall package/<name>` | Uninstall all resources from package | Project |
| `aimgr repo list --type=package` | List packages in repo | Repo |
| `aimgr repo remove package/<name>` | Remove package from repo | Repo |
| `aimgr repo create-package <name>` | Create new package | Repo |
| `aimgr repo update package/<name>` | Update package definition | Repo |

### Modified Commands

| Command | Change | Description |
|---------|--------|-------------|
| `aimgr repo list` | Show packages section | Lists resources + packages |
| `aimgr repo add` | Auto-discover packages | Import from `packages/` folder |
| `aimgr repo add` | Claude plugin adapter | Convert Claude plugins to packages |

### Unchanged Commands

| Command | Description |
|---------|-------------|
| `aimgr list` | List installed resources (NOT packages) |
| `aimgr install <type>/<name>` | Install individual resource |
| `aimgr uninstall <name>` | Uninstall individual resource |
| `aimgr repo update <type>/<name>` | Update individual resource |

---

## Implementation Details

### Package Type Definition

```go
// pkg/resource/package.go

type Package struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Resources   []string `json:"resources"`  // Changed: now flat array
}

type PackageMetadata struct {
    Name           string    `json:"name"`
    SourceType     string    `json:"source_type"`
    SourceURL      string    `json:"source_url,omitempty"`
    SourceRef      string    `json:"source_ref,omitempty"`
    FirstAdded     time.Time `json:"first_added"`
    LastUpdated    time.Time `json:"last_updated"`
    ResourceCount  int       `json:"resource_count"`
    OriginalFormat string    `json:"original_format,omitempty"`
}

// LoadPackage loads a package from .package.json file
func LoadPackage(filePath string) (*Package, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    var pkg Package
    if err := json.Unmarshal(data, &pkg); err != nil {
        return nil, err
    }
    
    // Validate required fields
    if pkg.Name == "" {
        return nil, fmt.Errorf("package name is required")
    }
    if pkg.Description == "" {
        return nil, fmt.Errorf("package description is required")
    }
    
    return &pkg, nil
}

// ParseResourceReference parses type/name format
func ParseResourceReference(ref string) (resourceType ResourceType, name string, err error) {
    parts := strings.SplitN(ref, "/", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("invalid resource format: %q (expected type/name)", ref)
    }
    
    typeStr, name := parts[0], parts[1]
    
    switch typeStr {
    case "command":
        resourceType = Command
    case "skill":
        resourceType = Skill
    case "agent":
        resourceType = Agent
    default:
        return "", "", fmt.Errorf("invalid resource type: %q (expected command/skill/agent)", typeStr)
    }
    
    return resourceType, name, nil
}

// SavePackage saves a package to .package.json file
func SavePackage(pkg *Package, repoPath string) error {
    filePath := filepath.Join(repoPath, "packages", pkg.Name+".package.json")
    
    data, err := json.MarshalIndent(pkg, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filePath, data, 0644)
}
```

### Install Package Logic

```go
// cmd/install.go

func installPackage(packageName string, opts InstallOptions) error {
    // Load package from repo
    repoPath := config.GetRepoPath()
    pkgPath := filepath.Join(repoPath, "packages", packageName+".package.json")
    
    pkg, err := resource.LoadPackage(pkgPath)
    if err != nil {
        return fmt.Errorf("package not found: %w", err)
    }
    
    fmt.Printf("Installing package: %s\n", pkg.Name)
    
    installed := 0
    missing := 0
    errors := []string{}
    
    // Install each resource
    for _, ref := range pkg.Resources {
        // Parse type/name format
        resType, resName, err := resource.ParseResourceReference(ref)
        if err != nil {
            errors = append(errors, fmt.Sprintf("%s: %v", ref, err))
            continue
        }
        
        // Install resource
        err = installResource(resType, resName, opts)
        if err != nil {
            if os.IsNotExist(err) {
                fmt.Printf("  ✗ %s - not found in repo\n", ref)
                missing++
            } else {
                errors = append(errors, fmt.Sprintf("%s: %v", ref, err))
            }
        } else {
            fmt.Printf("  ✓ %s\n", ref)
            installed++
        }
    }
    
    // Summary
    if missing > 0 {
        fmt.Printf("\nWarning: %d resource(s) not found\n", missing)
    }
    fmt.Printf("Installed %d of %d resources from package %s\n", 
               installed, len(pkg.Resources), pkg.Name)
    
    if len(errors) > 0 {
        fmt.Println("\nErrors:")
        for _, e := range errors {
            fmt.Printf("  ✗ %s\n", e)
        }
        return fmt.Errorf("package installation completed with errors")
    }
    
    return nil
}
```

### Claude Plugin Adapter

```go
// pkg/adapter/claude_plugin.go

type ClaudePlugin struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Commands    []string `json:"commands,omitempty"`
    Skills      []string `json:"skills,omitempty"`
    Tools       []string `json:"tools,omitempty"`
}

// ConvertClaudePluginToPackage converts Claude plugin to aimgr package
func ConvertClaudePluginToPackage(pluginPath string) (*resource.Package, error) {
    // Load Claude plugin JSON
    data, err := os.ReadFile(pluginPath)
    if err != nil {
        return nil, err
    }
    
    var plugin ClaudePlugin
    if err := json.Unmarshal(data, &plugin); err != nil {
        return nil, err
    }
    
    // Convert to aimgr package
    pkg := &resource.Package{
        Name:        plugin.Name,
        Description: plugin.Description,
        Resources:   []string{},
    }
    
    // Convert commands
    for _, cmd := range plugin.Commands {
        pkg.Resources = append(pkg.Resources, fmt.Sprintf("command/%s", cmd))
    }
    
    // Convert skills
    for _, skill := range plugin.Skills {
        pkg.Resources = append(pkg.Resources, fmt.Sprintf("skill/%s", skill))
    }
    
    // Convert tools to skills
    for _, tool := range plugin.Tools {
        pkg.Resources = append(pkg.Resources, fmt.Sprintf("skill/%s", tool))
    }
    
    return pkg, nil
}

// DetectClaudePlugin checks if directory contains Claude plugin
func DetectClaudePlugin(dirPath string) bool {
    pluginFiles := []string{
        "claude-plugin.json",
        "manifest.json",
    }
    
    for _, file := range pluginFiles {
        path := filepath.Join(dirPath, file)
        if _, err := os.Stat(path); err == nil {
            return true
        }
    }
    
    return false
}
```

---

## Shared Resources Behavior

### Scenario: Two Packages Share a Skill

**Setup:**
```
Package: dynatrace-rca
├─ command/dt-analyze-rca
├─ skill/dynatrace-dql-core    ← Shared
└─ agent/rca-agent

Package: dynatrace-config
├─ command/dt-configure
├─ skill/dynatrace-dql-core    ← Shared
└─ agent/config-agent
```

**Install both packages:**
```bash
$ aimgr install package/dynatrace-rca
Installing package: dynatrace-rca
  ✓ command/dt-analyze-rca
  ✓ skill/dynatrace-dql-core
  ✓ agent/rca-agent
Installed 3 of 3 resources

$ aimgr install package/dynatrace-config
Installing package: dynatrace-config
  ✓ command/dt-configure
  ○ skill/dynatrace-dql-core - already installed, skipping
  ✓ agent/config-agent
Installed 2 of 3 resources (1 skipped)
```

**Uninstall first package:**
```bash
$ aimgr uninstall package/dynatrace-rca
Uninstalling package: dynatrace-rca
  ✓ command/dt-analyze-rca
  ✓ skill/dynatrace-dql-core          ← REMOVES SHARED SKILL!
  ✓ agent/rca-agent
Uninstalled 3 resources
```

**Result:** `skill/dynatrace-dql-core` is removed even though `dynatrace-config` still references it!

**Implications:**
- **No reference counting**: Uninstall doesn't check if resource used elsewhere
- **User responsibility**: User must manage shared resources
- **Simple implementation**: No complex tracking needed

---

## Implementation Phases

### Phase 1: Core Package Support (MVP)

**Goal:** Basic package install/uninstall with manual creation

**Features:**
- `Package` type and JSON parsing (`type/name` format)
- `aimgr install package/<name>`
- `aimgr uninstall package/<name>`
- `aimgr repo list` shows packages
- `aimgr repo create-package` (manual creation)
- Package storage in `repo/packages/`
- Metadata tracking in `.metadata/packages/`

**Implementation tasks:**
1. Define `Package` and `PackageMetadata` types
2. Implement `ParseResourceReference()` function
3. Add package loading/saving functions
4. Create `install package/<name>` command
5. Create `uninstall package/<name>` command
6. Create `repo create-package` command
7. Update `repo list` to show packages
8. Implement metadata tracking

**Testing:**
- Unit tests for package parsing
- Unit tests for `type/name` parsing
- Install/uninstall with shared resources
- Error cases (missing resources, invalid format)
- Metadata creation and updates

**Estimated time:** 1-2 weeks

### Phase 2: Auto-Discovery & Adapters

**Goal:** Import packages from repos, convert Claude plugins

**Features:**
- `aimgr repo add` auto-discovers packages from `packages/` folder
- Claude plugin adapter (convert to aimgr packages)
- `aimgr repo remove package/<name>`
- `aimgr repo update package/<name>`

**Implementation tasks:**
1. Add package detection in `repo add`
2. Implement Claude plugin adapter
3. Create `repo remove package/<name>` command
4. Create `repo update package/<name>` command
5. Support `--with-resources` flag for remove

**Testing:**
- Add package from GitHub repo with `packages/` folder
- Convert Claude plugin to package
- Update package definition from source
- Remove package with/without resources

**Estimated time:** 1-2 weeks

### Phase 3: Polish & Documentation

**Goal:** User-facing polish and comprehensive docs

**Features:**
- Better error messages
- Shell completion for package names
- Example packages
- User documentation

**Implementation tasks:**
1. Improve error messages with suggestions
2. Add shell completion
3. Create example packages for testing
4. Write user documentation
5. Add package workflows to docs

**Testing:**
- End-to-end workflows
- Edge cases
- Documentation review

**Estimated time:** 1 week

---

## Example Workflows

### Workflow 1: Create and Install Package

```bash
# Add resources to repo
$ aimgr repo add gh:user/commands/dt-analyze-rca
$ aimgr repo add gh:user/skills/dynatrace-dql-core
$ aimgr repo add gh:user/agents/rca-agent

# Create package from resources
$ aimgr repo create-package dynatrace-rca \
  --description="Dynatrace RCA tools" \
  --resources="command/dt-analyze-rca,skill/dynatrace-dql-core,agent/rca-agent"

# Install package to project
$ aimgr install package/dynatrace-rca

# Verify installation
$ aimgr list
COMMANDS
  dt-analyze-rca
...
```

### Workflow 2: Import Package from Repo

```bash
# Repo structure:
# my-package-repo/
# ├── packages/
# │   └── toolkit.package.json
# ├── commands/
# │   └── cmd1.md
# └── skills/
#     └── skill1/SKILL.md

$ aimgr repo add gh:user/my-package-repo

Discovering resources...
  Found 1 command
  Found 1 skill
  Found 1 package

Adding commands...
  ✓ cmd1

Adding skills...
  ✓ skill1

Adding packages...
  ✓ toolkit (2 resources)

# Install the package
$ aimgr install package/toolkit
```

### Workflow 3: Convert Claude Plugin

```bash
# Source contains Claude plugin
$ aimgr repo add gh:anthropic/claude-pdf-plugin

Detected Claude plugin format
Converting to aimgr package...
  ✓ Converted commands: 2
  ✓ Converted skills: 1
  ✓ Created package: claude-pdf-plugin

Added package 'claude-pdf-plugin' (3 resources)

# Install converted package
$ aimgr install package/claude-pdf-plugin
```

### Workflow 4: Update Package

```bash
# Package definition changed upstream
$ aimgr repo update package/dynatrace-rca

Updating package: dynatrace-rca
  Source: gh:user/dynatrace-rca-package
  ✓ Package definition updated
  
Changes:
  + Added: skill/dynatrace-metrics-analysis
  - Removed: command/dt-old-tool
  
Package now contains 6 resources (was 5)

# Reinstall to get new resources
$ aimgr install package/dynatrace-rca
  ✓ command/dt-analyze-rca (already installed)
  ✓ skill/dynatrace-metrics-analysis (new!)
  ...
```

---

## Summary

**Core Design:**
- Packages = new resource type (alongside commands, skills, agents)
- Stored as JSON files: `repo/packages/<name>.package.json`
- Resources use `type/name` format (e.g., `"skill/pdf-processing"`)
- Metadata tracked for updates and provenance
- Cannot nest packages within packages

**Key Commands:**
- `aimgr install package/<name>` - Install all resources from package
- `aimgr uninstall package/<name>` - Remove all resources (aggressive, no ref counting)
- `aimgr repo create-package <name>` - Create package from existing resources
- `aimgr repo add <source>` - Auto-discover and import packages
- `aimgr repo update package/<name>` - Update package definition only
- `aimgr repo list` - Show packages (plus resources)

**Key Decisions:**
1. ✅ **Format**: Use flat array with `type/name` format
2. ✅ **Auto-discovery**: Import from `packages/` folder
3. ✅ **Adapters**: Convert Claude plugins to packages
4. ✅ **Updates**: Package definition only (no cascade)
5. ✅ **Metadata**: Track source and timestamps
6. ✅ **Validation**: Lazy validation at install time

**Philosophy:**
- Simple, explicit, no magic
- User manages shared resources
- Packages are convenience for bulk operations
- No hidden dependencies or complex tracking

---

**End of Design Document**
