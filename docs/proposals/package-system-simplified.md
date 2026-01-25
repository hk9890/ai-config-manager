# Simplified Package System (Addendum)

**Date**: 2026-01-25  
**Status**: Revised Proposal  
**Based on**: package-system-proposal.md + user feedback

---

## Key Simplifications

This document revises the main proposal based on the following principles:

### 1. **Packages are Grouping Concepts Only**

Packages are **NOT** versioned, complex entities. They are simply:
- A **directory structure** that groups related resources
- A **lightweight manifest** that lists what's included
- A **convenience mechanism** for bulk installation

**What packages DO NOT have:**
- ❌ Version numbers
- ❌ Author information
- ❌ License fields
- ❌ Complex dependency trees
- ❌ Marketplace metadata
- ❌ Update tracking

**What packages DO have:**
- ✅ Name
- ✅ Description
- ✅ List of resources (commands, skills, agents)
- ✅ Optional: system command requirements (e.g., "git", "jq")

### 2. **Installation Installs Individual Resources**

When you install a package:
```bash
$ aimgr install package/beads-workflow
```

**What happens:**
1. Package is detected/downloaded
2. Individual resources are extracted
3. Each resource is installed **separately** (commands, skills, agents)
4. Resources are tracked independently in metadata
5. **No package-level tracking** (just resource-level)

**Result:**
- You see individual commands/skills/agents in `aimgr list`
- You can remove individual resources: `aimgr uninstall beads-init`
- No "package" appears in listings (it's just a delivery mechanism)

### 3. **Both Project and Repo Level Work the Same**

**Project Level:**
```bash
# Install package to project
$ aimgr install package/beads-workflow

# Individual resources appear in .claude/, .opencode/, etc.
# Can remove individual resources
$ aimgr uninstall beads-init  # Just removes one command
```

**Repo Level:**
```bash
# Add package to repo
$ aimgr repo add gh:user/beads-workflow

# Individual resources stored in repo
# ~/.local/share/ai-config/repo/commands/beads-init.md
# ~/.local/share/ai-config/repo/agents/beads-planner.md

# Can remove individual resources
$ aimgr repo remove beads-init  # Just removes one command
```

---

## Simplified Package Structure

### Minimal Directory Structure

```
my-package/
├── package.json              # Minimal manifest (NEW LOCATION)
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

**Key Changes:**
- `package.json` at **root level** (not `.aimgr-package/` subdirectory)
- Simpler to create and manage
- Follows common package conventions (npm, pip, etc.)

### Minimal package.json Schema

```json
{
  "name": "beads-workflow",
  "description": "Complete beads workflow tools",
  "resources": {
    "commands": [
      "commands/beads-init.md",
      "commands/beads-create.md"
    ],
    "skills": [
      "skills/beads-planning"
    ],
    "agents": [
      "agents/beads-planner.md",
      "agents/beads-task-agent.md"
    ]
  },
  "requires": ["git", "jq"]
}
```

**Fields:**
- `name` (required): Package name
- `description` (required): Brief description
- `resources` (required): Lists of paths to resources
- `requires` (optional): System commands needed

**That's it!** No version, author, license, dependencies, tools config, etc.

---

## Simplified CLI Commands

### Core Commands

#### `aimgr install package/<name>`
Install all resources from a package.

```bash
# From GitHub
$ aimgr install package/beads-workflow
# or
$ aimgr install gh:user/beads-workflow

# From local directory
$ aimgr install package/~/dev/my-package

# What it does:
# 1. Downloads/locates package
# 2. Reads package.json
# 3. Installs each resource individually
# 4. No package tracking (just resources)
```

#### `aimgr repo add <package-source>`
Add package resources to repository.

```bash
$ aimgr repo add gh:user/beads-workflow

# What it does:
# 1. Downloads package
# 2. Reads package.json
# 3. Adds each resource to repo individually
# 4. Metadata tracks source URL (like current behavior)
```

### No New Package Commands

**We do NOT need:**
- ❌ `aimgr package list` (resources are not grouped after install)
- ❌ `aimgr package show` (no package metadata to show)
- ❌ `aimgr package update` (resources updated individually)
- ❌ `aimgr package remove` (remove resources individually)

**Instead, existing commands work:**
```bash
# List all installed resources (including those from packages)
$ aimgr list

# Remove individual resources
$ aimgr uninstall beads-init
$ aimgr repo remove beads-planner
```

---

## Implementation Approach

### Phase 1: Package Detection & Parsing

**Goal**: Recognize package directories and parse manifest

**Code Changes:**
```go
// pkg/resource/package.go

type Package struct {
    Name        string
    Description string
    Resources   PackageResources
    Requires    []string
}

type PackageResources struct {
    Commands []string
    Skills   []string
    Agents   []string
}

func DetectPackage(path string) (bool, error)
func LoadPackage(path string) (*Package, error)
func InstallPackage(pkg *Package, sourcePath string) error
```

**Detection Logic:**
1. Check for `package.json` at root
2. Validate structure (name, description, resources)
3. Return Package struct

### Phase 2: Install Integration

**Goal**: `aimgr install package/<name>` works

**Code Changes:**
```go
// cmd/install.go

func runInstall(cmd *cobra.Command, args []string) error {
    pattern := args[0]
    
    // Check if pattern is "package/..." or "gh:user/repo"
    if strings.HasPrefix(pattern, "package/") || isGitSource(pattern) {
        // Try to detect as package
        isPackage, err := resource.DetectPackage(sourcePath)
        if err == nil && isPackage {
            return installPackage(sourcePath)
        }
    }
    
    // Fall back to existing resource install
    return installResource(pattern)
}

func installPackage(sourcePath string) error {
    // 1. Load package.json
    pkg, err := resource.LoadPackage(sourcePath)
    if err != nil {
        return err
    }
    
    // 2. Check system requirements
    for _, cmd := range pkg.Requires {
        if !commandExists(cmd) {
            fmt.Printf("Warning: Required command '%s' not found\n", cmd)
        }
    }
    
    // 3. Install each resource individually
    for _, cmdPath := range pkg.Resources.Commands {
        fullPath := filepath.Join(sourcePath, cmdPath)
        installCommand(fullPath)
    }
    for _, skillPath := range pkg.Resources.Skills {
        fullPath := filepath.Join(sourcePath, skillPath)
        installSkill(fullPath)
    }
    for _, agentPath := range pkg.Resources.Agents {
        fullPath := filepath.Join(sourcePath, agentPath)
        installAgent(fullPath)
    }
    
    return nil
}
```

### Phase 3: Repo Integration

**Goal**: `aimgr repo add` handles packages

**Code Changes:**
```go
// pkg/repo/manager.go

func (m *Manager) Add(sourcePath string) error {
    // Detect if source is a package
    isPackage, _ := resource.DetectPackage(sourcePath)
    
    if isPackage {
        return m.AddPackage(sourcePath)
    }
    
    // Fall back to single resource
    return m.AddResource(sourcePath)
}

func (m *Manager) AddPackage(sourcePath string) error {
    pkg, err := resource.LoadPackage(sourcePath)
    if err != nil {
        return err
    }
    
    // Add each resource to repo individually
    for _, cmdPath := range pkg.Resources.Commands {
        fullPath := filepath.Join(sourcePath, cmdPath)
        m.AddCommand(fullPath)
    }
    // ... (skills, agents)
    
    return nil
}
```

---

## Example Workflows

### Creating a Package

**Manual Creation:**
```bash
# Create directory structure
mkdir my-package
cd my-package
mkdir commands skills agents

# Create resources
echo "# My Command" > commands/my-cmd.md
# ... add frontmatter, etc.

# Create package.json
cat > package.json << 'PKGJSON'
{
  "name": "my-package",
  "description": "My custom package",
  "resources": {
    "commands": ["commands/my-cmd.md"],
    "skills": [],
    "agents": []
  },
  "requires": []
}
PKGJSON

# Push to GitHub
git init
git add .
git commit -m "Initial commit"
git remote add origin gh:user/my-package
git push -u origin main
```

**Optional: Helper Command (future)**
```bash
# Interactive package creator
$ aimgr create-package my-package
Package name: my-package
Description: My custom package
Add existing resources? [y/N]: y
  Select commands: [✓] my-cmd
  Select skills: [ ] 
  Select agents: [ ]
Created package at: ./my-package/
```

### Installing a Package

```bash
# From GitHub
$ aimgr install gh:user/beads-workflow
Detected package: beads-workflow
Installing resources...
  ✓ command/beads-init
  ✓ command/beads-create
  ✓ skill/beads-planning
  ✓ agent/beads-planner
  ✓ agent/beads-task-agent
Installed 5 resources to Claude Code, OpenCode

# Check system requirements
Warning: Required command 'jq' not found
Install with: apt install jq

# Resources now available
$ claude /beads-init
```

### Managing Installed Resources

```bash
# List all resources (including from packages)
$ aimgr list

COMMANDS
beads-init
beads-create
...

AGENTS
beads-planner
beads-task-agent
...

# Remove individual resource
$ aimgr uninstall beads-init
Removed beads-init from Claude Code, OpenCode

# Install individual resource back
$ aimgr install command/beads-init
```

### Adding Package to Repo

```bash
# Add package to central repo
$ aimgr repo add gh:user/beads-workflow
Detected package: beads-workflow
Adding resources to repository...
  ✓ beads-init → repo/commands/
  ✓ beads-create → repo/commands/
  ✓ beads-planning → repo/skills/
  ✓ beads-planner → repo/agents/
  ✓ beads-task-agent → repo/agents/

# Metadata tracks source for each resource
$ cat ~/.local/share/ai-config/repo/.metadata/commands/beads-init-metadata.json
{
  "name": "beads-init",
  "source_url": "gh:user/beads-workflow",
  "source_type": "github",
  ...
}

# Update individual resources
$ aimgr repo update beads-init
# or update all from same source
$ aimgr repo update --source=gh:user/beads-workflow
```

---

## Comparison: Original vs. Simplified

| Feature | Original Proposal | Simplified |
|---------|-------------------|------------|
| **Versioning** | Semantic versioning | None |
| **Author/License** | Tracked in metadata | None |
| **Dependencies** | Package dependencies | System commands only |
| **Package Tracking** | Separate package list | None (resources only) |
| **Manifest Location** | `.aimgr-package/package.json` | `package.json` (root) |
| **Package Commands** | `package list/show/update/remove` | None (use resource commands) |
| **Installation** | Installs package entity | Installs individual resources |
| **Removal** | Remove package or resources | Remove resources only |
| **Metadata Storage** | Package + resource metadata | Resource metadata only |
| **Update Strategy** | Package-level updates | Resource-level updates |
| **Marketplace** | Complex registry system | Simple GitHub + topics |
| **Tool Config** | Per-package tool compat | Per-resource (existing) |

---

## Benefits of Simplified Approach

### 1. **Simpler Mental Model**
- Users think: "I'm installing resources from a package"
- Not: "I'm installing a package that contains resources"

### 2. **Less Code, Less Complexity**
- No package entity tracking
- No package-level commands
- No version resolution
- Reuse existing resource management code

### 3. **Flexibility**
- Install whole package or individual resources
- Remove resources without "orphaned package" concerns
- Mix-and-match from different packages

### 4. **Backward Compatible**
- All existing workflows unchanged
- Packages are opt-in convenience feature
- No new concepts for users to learn

### 5. **Easier to Implement**
- Phase 1: Detection & parsing (~1 week)
- Phase 2: Install integration (~1 week)
- Phase 3: Repo integration (~3 days)
- **Total: ~2-3 weeks** (vs. 3-4 months for full proposal)

---

## What We Lose vs. Original Proposal

### Features We're NOT Implementing

1. **Versioning**
   - Can't track package versions
   - Can't require specific versions
   - **Workaround**: Use Git tags in source URL (`gh:user/pkg@v1.0.0`)

2. **Package-Level Updates**
   - Can't "update all resources from beads-workflow at once"
   - **Workaround**: `aimgr repo update --source=gh:user/beads-workflow`

3. **Package Metadata**
   - No author, license, homepage tracking
   - **Workaround**: This info can be in README.md

4. **Dependency Management**
   - No automatic package dependencies
   - **Workaround**: Manual installation, or list in README

5. **Package Listings**
   - Can't see "which packages are installed"
   - **Workaround**: Resources show source in metadata

6. **Marketplace Features**
   - No package search/browse
   - **Workaround**: GitHub search, topics, awesome lists

### Are These Losses Acceptable?

**Arguments FOR simplified approach:**
- Packages are primarily for **distribution**, not **management**
- Version control is better handled at **Git level** (tags, branches)
- Most users care about **resources**, not packaging
- Simpler = less bugs, easier maintenance
- Can always add features later if needed

**Arguments AGAINST (from original proposal):**
- Professional users expect versioning
- Package updates are more convenient than resource updates
- Marketplace needs metadata for discovery
- Other package managers (npm, pip) have these features

---

## Hybrid Approach (Compromise)

If we want **some** metadata without full complexity:

### Option: Lightweight Metadata Tracking

Track which resources came from which package, but don't manage packages as entities.

**Resource Metadata:**
```json
{
  "name": "beads-init",
  "source_url": "gh:user/beads-workflow",
  "source_type": "github",
  "package_name": "beads-workflow",  // NEW
  "package_source": "gh:user/beads-workflow"  // NEW
}
```

**Benefits:**
- Can list resources by package: `aimgr list --from-package=beads-workflow`
- Can update all resources from package: `aimgr update --package=beads-workflow`
- Can show "installed packages": `aimgr packages` (derived from resource metadata)

**Implementation:**
- Add `package_name` and `package_source` fields to resource metadata
- Populate during package installation
- Query across resource metadata for package-related commands

**CLI Examples:**
```bash
# List resources from a package
$ aimgr list --from-package=beads-workflow
COMMANDS (from beads-workflow)
  beads-init
  beads-create

AGENTS (from beads-workflow)
  beads-planner
  beads-task-agent

# Update all resources from a package
$ aimgr update --package=beads-workflow
Checking for updates: gh:user/beads-workflow
Found updates for:
  - beads-init (changed)
  - beads-planner (changed)
Update these resources? [Y/n]: y

# Show packages (derived from resource metadata)
$ aimgr packages
PACKAGES (installed resources)
beads-workflow    5 resources    gh:user/beads-workflow
pdf-toolkit       3 resources    gh:org/pdf-toolkit
```

**Cost:**
- Slightly more complex than "pure simplified"
- Still much simpler than original proposal
- Good middle ground between convenience and simplicity

---

## Recommendation

### Proposed Implementation Strategy

**Phase 1: Core (Simplified)**
- Package detection and parsing
- Install package → installs resources individually
- No package tracking, no metadata

**Phase 2: Lightweight Tracking (Hybrid)**
- Add `package_name`/`package_source` to resource metadata
- Implement `aimgr list --from-package=<name>`
- Implement `aimgr update --package=<name>`
- Implement `aimgr packages` (derived list)

**Phase 3: Optional Enhancements**
- Discovery helpers (search GitHub topics)
- Package creation wizard
- Validation tools

**NOT Implementing (for now):**
- Versioning
- Author/license tracking
- Complex dependency resolution
- Marketplace/registry
- Package-level commands (except derived ones)

---

## Updated Timeline

**Phase 1: Core Simplified (2 weeks)**
- Week 1: Detection, parsing, install logic
- Week 2: Testing, documentation

**Phase 2: Lightweight Tracking (1 week)**
- Week 3: Metadata extensions, queries, CLI updates

**Phase 3: Enhancements (1-2 weeks, optional)**
- Week 4-5: Discovery, creation tools, polish

**Total: 3-5 weeks** (vs. 12+ weeks for original proposal)

---

## Action Items

### Immediate
1. ✅ Review this simplified proposal
2. ⬜ Decide: Pure simplified or hybrid approach?
3. ⬜ Finalize `package.json` schema
4. ⬜ Create example package for testing

### Short-term
1. ⬜ Implement package detection (`pkg/resource/package.go`)
2. ⬜ Integrate into `aimgr install`
3. ⬜ Integrate into `aimgr repo add`
4. ⬜ Write tests (unit + integration)
5. ⬜ Update documentation

### Medium-term
1. ⬜ Add lightweight tracking (if hybrid approach)
2. ⬜ Create package creation helper
3. ⬜ Create example packages (beads-workflow, etc.)
4. ⬜ Write tutorial/guide

---

## Questions for Review

1. **Approach**: Pure simplified or hybrid (with lightweight tracking)?
2. **Manifest Location**: `package.json` at root (simplified) or `.aimgr-package/package.json` (namespaced)?
3. **Required Fields**: Just `name`/`description`/`resources`, or also `requires`?
4. **System Requirements**: Show warnings for missing commands, or error?
5. **Resource Paths**: Relative to package root, or allow absolute?
6. **Naming**: "package" or something else ("bundle", "collection", "group")?

---

**End of Simplified Proposal**

---

## Addendum: Reference Model (2026-01-25)

### User Feedback: References Over Containers

**Key insight**: Packages as "reference collections" rather than resource containers.

### The Reference Model

**Package structure:**
```
dynatrace-rca-analysis/
└── package.json
```

**package.json (references external resources):**
```json
{
  "name": "dynatrace-rca-analysis",
  "description": "Dynatrace root cause analysis tools",
  "references": {
    "commands": [
      "gh:user/dynatrace-commands/analyze-rca",
      "gh:user/dynatrace-commands/trace-analyzer"
    ],
    "skills": [
      "gh:shared/dynatrace-skills/dynatrace-dql-core",
      "gh:user/dynatrace-skills/log-analysis"
    ],
    "agents": [
      "gh:user/dynatrace-agents/rca-agent"
    ]
  },
  "requires": ["jq", "curl"]
}
```

### Example: Shared Resources

**Scenario**: Two packages both need `dynatrace-dql-core` skill

```
Package: dynatrace-rca-analysis
├─ references: gh:shared/dynatrace-skills/dynatrace-dql-core ─┐
├─ references: gh:user/dynatrace-commands/analyze-rca         │
└─ references: gh:user/dynatrace-agents/rca-agent             │
                                                               │
Package: dynatrace-config                                     │
├─ references: gh:shared/dynatrace-skills/dynatrace-dql-core ─┤─→ Shared!
├─ references: gh:user/dynatrace-commands/configure-dt        │
└─ references: gh:user/dynatrace-agents/config-agent          │
```

**Installation:**
```bash
# Install package 1
$ aimgr install package/dynatrace-rca-analysis
Resolving references...
  ✓ dynatrace-dql-core (from gh:shared/dynatrace-skills)
  ✓ analyze-rca (from gh:user/dynatrace-commands)
  ✓ rca-agent (from gh:user/dynatrace-agents)
Installed 3 resources

# Install package 2
$ aimgr install package/dynatrace-config
Resolving references...
  ○ dynatrace-dql-core (already installed, skipping)
  ✓ configure-dt (from gh:user/dynatrace-commands)
  ✓ config-agent (from gh:user/dynatrace-agents)
Installed 2 resources (1 skipped)
```

### Removal: Defer the Problem

**User insight**: "Remove is not a main use case, usually update"

**Simplified removal approach:**
1. **Don't track reference counts** (simpler implementation)
2. **Removal removes resources directly** (user's responsibility to check)
3. **Warn on remove** if unsure

**Options for removal:**

**Option 1: Simple removal (warn user)**
```bash
$ aimgr uninstall dynatrace-dql-core
Warning: This resource may be used by multiple packages.
Remove anyway? [y/N]: n
```

**Option 2: No package-level removal (resources only)**
```bash
# Remove individual resources (existing command)
$ aimgr uninstall dynatrace-dql-core

# No "uninstall package" command
# User manages resources individually
```

**Option 3: Defer to future (recommended)**
```bash
# For now: only implement install
$ aimgr install package/dynatrace-rca-analysis
✓ Installed

# Remove not implemented yet
$ aimgr uninstall package/dynatrace-rca-analysis
Error: Package removal not yet implemented.
Use 'aimgr uninstall <resource>' to remove individual resources.

# Or: treat as regular uninstall of all resources
$ aimgr uninstall package/dynatrace-rca-analysis
This will remove 5 resources:
  - dynatrace-dql-core (skill)
  - analyze-rca (command)
  - rca-agent (agent)
  ...
Continue? [y/N]:
```

### Advantages of Reference Model

1. **No duplication**: Skills/commands/agents stored once
2. **Shared resources**: Multiple packages reference same skill
3. **Composability**: Build packages from existing resources
4. **Update propagation**: Update resource → all packages benefit
5. **Resource independence**: Resources exist standalone or in packages

### Implementation: References Only

**Simplified package.json schema:**
```json
{
  "name": "package-name",
  "description": "Package description",
  "references": {
    "commands": ["source1", "source2"],
    "skills": ["source3"],
    "agents": ["source4"]
  },
  "requires": ["system-command"]
}
```

**Reference formats:**
- GitHub: `gh:user/repo/path/to/resource`
- Git URL: `https://github.com/user/repo/path/to/resource`
- Local: `~/path/to/resource`
- Repo resource: `repo:resource-name` (from aimgr repo)

**Installation logic:**
```go
func installPackage(pkg *Package) error {
    for _, ref := range pkg.References.Commands {
        // Download/locate resource
        resourcePath := resolveReference(ref)
        
        // Check if already installed
        if isInstalled(resourcePath) {
            fmt.Printf("○ %s (already installed, skipping)\n", ref)
            continue
        }
        
        // Install resource
        installCommand(resourcePath)
        fmt.Printf("✓ %s\n", ref)
    }
    // ... same for skills, agents
    return nil
}
```

### Hybrid: References + Contained Resources

Support both patterns in same package:

```json
{
  "name": "my-package",
  "description": "Hybrid package",
  
  "resources": {
    "commands": ["commands/my-cmd.md"]
  },
  
  "references": {
    "skills": ["gh:shared/skills/common-skill"]
  }
}
```

**Use case:**
- Package contains specialized resources (commands, agents)
- Package references common shared resources (skills)

### Main Workflow: Update, Not Remove

**User pattern:**
```bash
# Initial setup: install packages
$ aimgr install package/dynatrace-rca-analysis
$ aimgr install package/dynatrace-config
$ aimgr install package/web-scraping

# Regular workflow: update everything
$ aimgr repo update --all

# Or update specific sources
$ aimgr repo update --source=gh:shared/dynatrace-skills

# Or update individual resources
$ aimgr repo update dynatrace-dql-core
```

**Remove is rare:**
- Users typically don't remove packages
- If needed, remove resources individually
- Reference counting not critical for MVP

### Recommendation: Reference Model Without Removal

**Phase 1 Implementation:**
1. ✅ Support `references` in package.json
2. ✅ Resolve references during install
3. ✅ Skip if resource already installed
4. ✅ Track which resources came from which package (metadata)
5. ❌ No package-level removal (defer)

**Future Phase (if needed):**
1. Add reference counting
2. Implement smart package removal
3. Handle dangling references

**Timeline:**
- Phase 1: 2-3 weeks (same as before)
- No additional complexity for removal
- Clean implementation without edge cases

### Updated package.json Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "aimgr Package (Reference Model)",
  "type": "object",
  "required": ["name", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$"
    },
    "description": {
      "type": "string",
      "minLength": 1,
      "maxLength": 500
    },
    "references": {
      "type": "object",
      "properties": {
        "commands": {
          "type": "array",
          "items": {"type": "string"}
        },
        "skills": {
          "type": "array",
          "items": {"type": "string"}
        },
        "agents": {
          "type": "array",
          "items": {"type": "string"}
        }
      }
    },
    "resources": {
      "type": "object",
      "properties": {
        "commands": {
          "type": "array",
          "items": {"type": "string"}
        },
        "skills": {
          "type": "array",
          "items": {"type": "string"}
        },
        "agents": {
          "type": "array",
          "items": {"type": "string"}
        }
      }
    },
    "requires": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}
```

### Example Packages

**Package 1: RCA Analysis (references only)**
```json
{
  "name": "dynatrace-rca-analysis",
  "description": "Root cause analysis for Dynatrace",
  "references": {
    "skills": [
      "gh:shared/dynatrace-skills/dynatrace-dql-core",
      "gh:user/dynatrace-skills/log-analysis"
    ],
    "commands": [
      "gh:user/dynatrace-commands/analyze-rca"
    ],
    "agents": [
      "gh:user/dynatrace-agents/rca-agent"
    ]
  },
  "requires": ["jq"]
}
```

**Package 2: Config Tools (hybrid)**
```json
{
  "name": "dynatrace-config",
  "description": "Dynatrace configuration tools",
  "resources": {
    "commands": [
      "commands/configure-dt.md"
    ]
  },
  "references": {
    "skills": [
      "gh:shared/dynatrace-skills/dynatrace-dql-core"
    ]
  },
  "requires": ["curl"]
}
```

**Package 3: Self-contained (resources only)**
```json
{
  "name": "pdf-toolkit",
  "description": "PDF processing tools",
  "resources": {
    "commands": [
      "commands/pdf-extract.md",
      "commands/pdf-merge.md"
    ],
    "skills": [
      "skills/pdf-processing"
    ]
  },
  "requires": ["pdftotext"]
}
```

### Summary

**Key decisions:**
1. ✅ Support reference model (no duplication)
2. ✅ Install resolves references and installs resources
3. ✅ Skip already-installed resources
4. ✅ Support hybrid (references + contained)
5. ❌ Defer package removal (not main use case)
6. ✅ Focus on install and update workflows

**This matches user workflow:**
- Install packages (collections of references)
- Update resources regularly
- Rarely remove (defer complexity)
- Manage resources individually when needed

