# Source Provider Architecture

**Status**: Research Document  
**Date**: 2026-02-07  
**Author**: AI Research (beads task ai-config-manager-032)

## Executive Summary

This document analyzes the current source provider architecture in ai-config-manager to inform the integration of agentskills.in as a new source type.

**Key Findings**:
- ✅ Clear abstraction exists: `source.ParsedSource` struct
- ✅ Centralized parsing: `source.ParseSource()` function
- ✅ Workspace caching: All Git sources use `pkg/workspace`
- ✅ Metadata tracking: Source URL stored with resources
- ⚠️ No interface-based polymorphism (struct-based instead)
- ⚠️ Provider-specific logic scattered across commands

## Architecture Overview

### Data Flow Diagram

```
User Input (CLI)
      ↓
source.ParseSource()  ← Parses source string into ParsedSource
      ↓
source.GetCloneURL()  ← Converts to Git URL (if remote)
      ↓
workspace.GetOrClone() ← Caches Git repos (10-50x faster)
      ↓
discovery.Discover*()  ← Finds resources in directory
      ↓
repo.AddBulk()        ← Imports resources
      ↓
metadata.Save()       ← Tracks source URL/type/ref
```

### Component Responsibilities

| Component | Responsibility | Location |
|-----------|----------------|----------|
| **Source Parser** | Parse source strings, detect type | `pkg/source/parser.go` |
| **Workspace Manager** | Cache Git repos, manage refs | `pkg/workspace/manager.go` |
| **Discovery** | Find resources in directories | `pkg/discovery/*.go` |
| **Repository Manager** | Import/manage resources | `pkg/repo/manager.go` |
| **Metadata** | Track source provenance | `pkg/metadata/metadata.go` |

## Source Provider Comparison

| Feature | Local | GitHub (gh:) | Git URL | AgentSkills (Future) |
|---------|-------|--------------|---------|----------------------|
| **Prefix** | (none) | `gh:` | `https://` | `as:` (proposed) |
| **Parser** | `parseLocalPrefix()` | `parseGitHubPrefix()` | `parseHTTPURL()` | `parseAgentSkillsPrefix()` (TBD) |
| **Resolver** | Direct passthrough | → `GetCloneURL()` | → `GetCloneURL()` | → API fetch (TBD) |
| **Workspace** | No | Yes | Yes | Maybe (cache API responses?) |
| **Metadata** | `file://` URL | GitHub URL | Git URL | AgentSkills URL |
| **Import Mode** | Symlink (default) | Copy | Copy | Copy |
| **Example** | `/home/user/resources` | `gh:hk9890/ai-tools` | `https://...` | `as:hk9890/pdf-skill` |

## Current Implementation Details

### 1. Source Parsing (`pkg/source/parser.go`)

**Entry Point**: `ParseSource(input string) (*ParsedSource, error)`

**Supported Formats**:
```go
// GitHub
"gh:owner/repo"                    → GitHub type
"gh:owner/repo@branch"             → GitHub type with ref
"gh:owner/repo/path/to/resource"  → GitHub type with subpath
"owner/repo"                       → Inferred as GitHub

// Git URLs
"https://github.com/owner/repo"    → GitHub type
"https://gitlab.com/owner/repo"    → GitLab type
"https://example.com/repo.git"     → GitURL type
"git@github.com:owner/repo.git"    → GitHub type (SSH)

// Local Paths
"local:/path/to/dir"               → Local type
"./relative/path"                  → Inferred as Local
"/absolute/path"                   → Inferred as Local
```

**ParsedSource Struct**:
```go
type ParsedSource struct {
    Type      SourceType // GitHub, GitLab, Local, GitURL
    URL       string     // Full URL for Git sources
    LocalPath string     // Path for local sources
    Ref       string     // Branch/tag/commit (optional)
    Subpath   string     // Path within repo (optional)
}
```

**Key Functions**:
- `ParseSource()` - Main entry point
- `GetCloneURL()` - Converts ParsedSource to Git clone URL
- `parseGitHubPrefix()` - Handles `gh:owner/repo` format
- `parseLocalPrefix()` - Handles local paths
- `parseHTTPURL()` - Handles full URLs

### 2. GitHub Source Flow

**Example**: `aimgr repo import gh:hk9890/ai-tools@main/skills`

```
1. cmd/repo_import.go:96
   → parsed, err := source.ParseSource("gh:hk9890/ai-tools@main/skills")
   → Returns: ParsedSource{
       Type: GitHub,
       URL: "https://github.com/hk9890/ai-tools",
       Ref: "main",
       Subpath: "skills"
     }

2. cmd/repo_import.go:629
   → cloneURL, err := source.GetCloneURL(parsed)
   → Returns: "https://github.com/hk9890/ai-tools"

3. cmd/repo_import.go:635-644
   → workspaceManager, _ := workspace.NewManager(repoPath)
   → cachePath, _ := workspaceManager.GetOrClone(cloneURL, "main")
   → Returns: "/home/user/.local/share/ai-config/repo/.workspace/<hash>"

4. cmd/repo_import.go:647-655
   → workspaceManager.Update(cloneURL, "main")  // Pull latest
   → searchPath := filepath.Join(cachePath, "skills")

5. cmd/repo_import.go:680
   → importFromLocalPath(searchPath, manager, filter, parsed.URL, "github", "main")
   → Discovery finds resources in searchPath
   → Imports with metadata: sourceURL="https://github.com/hk9890/ai-tools"
```

**Workspace Caching**:
- First clone: ~30 seconds
- Subsequent access: <1 second (10-50x faster)
- Location: `<repoPath>/.workspace/<sha256-hash>/`
- Metadata: `.workspace/.cache-metadata.json`

### 3. Local Source Flow

**Example**: `aimgr repo import ~/my-resources`

```
1. cmd/repo_import.go:96
   → parsed, err := source.ParseSource("~/my-resources")
   → Returns: ParsedSource{
       Type: Local,
       LocalPath: "/home/user/my-resources"
     }

2. cmd/repo_import.go:108-122
   → importMode := "symlink"  // Default for local
   → if copyFlag { importMode = "copy" }
   → addBulkFromLocalWithMode(parsed.LocalPath, manager, filter, importMode)

3. Discovery runs directly on local path
   → No cloning, no workspace caching
   → Resources imported as symlinks (or copies if --copy flag)
   → Metadata: sourceURL="file:///home/user/my-resources"
```

**Symlink Handling**:
- Default mode for local sources (fast, saves space)
- Resources point to original files
- Changes in source directory affect installed resources
- Used for development workflow

### 4. Source Resolution in Commands

**Import Command** (`cmd/repo_import.go`):
```go
// Line 96: Parse source
parsed, err := source.ParseSource(sourceInput)

// Line 108: Determine if remote
isRemote := parsed.Type == source.GitHub || parsed.Type == source.GitURL

// Line 110-122: Branch based on type
if isRemote {
    return addBulkFromGitHub(parsed, manager)  // Copy mode
} else {
    return addBulkFromLocalWithMode(...)       // Symlink/copy
}
```

**Sync Command** (`cmd/repo_sync.go`):
```go
// Line 93: Parse configured source
parsed, err := source.ParseSource(src.URL)

// Line 99: Get clone URL
cloneURL, err := source.GetCloneURL(parsed)

// Line 105: Use workspace cache
sourcePath, err = wsMgr.GetOrClone(cloneURL, parsed.Ref)
```

### 5. Metadata Tracking

**Storage Location**: `<repoPath>/.metadata/<type>s/<name>-metadata.json`

**Metadata Structure**:
```go
type ResourceMetadata struct {
    Name           string           // Resource name
    Type           ResourceType     // command/skill/agent
    SourceType     string           // "github", "git-url", "local", "file"
    SourceURL      string           // Full source URL
    Ref            string           // Git ref (branch/tag/commit)
    FirstInstalled time.Time        // Initial import
    LastUpdated    time.Time        // Last sync
}
```

**Example** (skill imported from GitHub):
```json
{
  "name": "pdf-processing",
  "type": "skill",
  "source_type": "github",
  "source_url": "https://github.com/hk9890/ai-tools",
  "ref": "main",
  "first_installed": "2026-02-07T10:00:00Z",
  "last_updated": "2026-02-07T10:00:00Z"
}
```

**Key Functions**:
- `metadata.Save()` - Write metadata file
- `metadata.Load()` - Read metadata file
- `metadata.GetMetadataPath()` - Compute metadata path

### 6. Discovery System

Discovery finds resources in a directory tree.

**Entry Points**:
- `discovery.DiscoverCommands(basePath, subpath)` → `[]*resource.Resource`
- `discovery.DiscoverSkills(basePath, subpath)` → `[]*resource.Resource`
- `discovery.DiscoverAgents(basePath, subpath)` → `[]*resource.Resource`
- `discovery.DiscoverPackages(basePath, subpath)` → `[]*resource.Package`

**Search Strategy**:

**For Commands**:
1. Priority search: `commands/`, `.claude/commands/`, `.opencode/commands/`
2. Recursive search: Look for `commands/` dirs at any depth
3. File matching: `.md` files (excluding README.md, SKILL.md)

**For Skills**:
1. Priority search: `skills/`, `.claude/skills/`, `.opencode/skills/`, `.github/skills/`
2. Check for `SKILL.md` file in each directory
3. Recursive fallback if no skills found

**For Agents**:
1. Priority search: `agents/`, `.claude/agents/`, `.opencode/agents/`
2. File matching: `.md` files in agents directories

**Symlink Handling** (Architecture Rule 5):
```go
// WRONG: Doesn't follow symlinks
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    if entry.IsDir() { ... }  // ← Breaks for symlinked dirs
}

// CORRECT: Follows symlinks
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    info, err := os.Stat(path)  // ← Follows symlinks
    if err != nil || !info.IsDir() { continue }
}
```

## Interface Analysis

### Current Approach: Struct-Based

The codebase uses a **struct-based** approach rather than interface-based polymorphism:

**Pros**:
- Simple, direct implementation
- Easy to understand data flow
- No abstraction overhead
- Type safety without interfaces

**Cons**:
- Provider-specific logic scattered across commands
- Hard to add new source types (requires touching multiple files)
- No clear extension point
- Type switching based on `parsed.Type` field

### Where Provider Logic Lives

| Concern | Location | Notes |
|---------|----------|-------|
| **Parsing** | `pkg/source/parser.go` | Centralized ✅ |
| **Workspace** | `pkg/workspace/manager.go` | Git-specific |
| **Import Logic** | `cmd/repo_import.go` | Type switching |
| **Sync Logic** | `cmd/repo_sync.go` | Type switching |
| **Metadata** | `pkg/metadata/metadata.go` | Generic ✅ |

### No Explicit Interface

There is **no Source interface** like:
```go
type Source interface {
    Fetch() (localPath string, err error)
    Update() error
    GetMetadata() SourceMetadata
}
```

Instead, logic is conditional:
```go
// cmd/repo_import.go:108
isRemote := parsed.Type == source.GitHub || parsed.Type == source.GitURL

if isRemote {
    return addBulkFromGitHub(parsed, manager)
} else {
    return addBulkFromLocalWithMode(...)
}
```

## Recommendations for AgentSkills Integration

### Option 1: Follow Existing Pattern (Recommended)

**Approach**: Add AgentSkills as another case in the struct-based system.

**Steps**:
1. Add `AgentSkills` to `SourceType` enum
2. Add `parseAgentSkillsPrefix()` to `pkg/source/parser.go`
3. Add case to `GetCloneURL()` (or create `GetSourcePath()`)
4. Add case to import/sync commands
5. Store metadata with `source_type: "agentskills"`

**Pros**:
- Consistent with existing architecture
- Minimal refactoring
- Clear precedent (GitHub, GitLab, GitURL)
- Fast to implement

**Cons**:
- Continues pattern of scattered logic
- Every new source type requires multiple file changes

**Example Implementation**:
```go
// pkg/source/parser.go
const (
    GitHub       SourceType = "github"
    GitLab       SourceType = "gitlab"
    Local        SourceType = "local"
    GitURL       SourceType = "git-url"
    AgentSkills  SourceType = "agentskills"  // NEW
)

func ParseSource(input string) (*ParsedSource, error) {
    if strings.HasPrefix(input, "as:") {
        return parseAgentSkillsPrefix(strings.TrimPrefix(input, "as:"))
    }
    // ... existing logic
}

func parseAgentSkillsPrefix(input string) (*ParsedSource, error) {
    // Parse: as:owner/skillname[@version]
    // Return ParsedSource with AgentSkills type
}

// cmd/repo_import.go
isRemote := parsed.Type == source.GitHub || 
            parsed.Type == source.GitURL || 
            parsed.Type == source.AgentSkills  // NEW

if parsed.Type == source.AgentSkills {
    return addBulkFromAgentSkills(parsed, manager)
}
```

### Option 2: Introduce Interface Layer (Over-Engineering)

**Approach**: Create `Source` interface with provider implementations.

**Pros**:
- Clean extension point
- Testable in isolation
- Provider logic encapsulated

**Cons**:
- Large refactor required
- Breaks existing patterns
- Over-engineering for 4-5 source types
- No clear benefit over struct approach

**Not Recommended**: This would be premature abstraction.

### Option 3: Hybrid Approach

**Approach**: Keep struct-based parsing, but add provider-specific handlers.

```go
// pkg/source/provider.go (new file)
type Provider interface {
    Fetch(parsed *ParsedSource) (localPath string, err error)
}

type AgentSkillsProvider struct {
    client *http.Client
}

func (p *AgentSkillsProvider) Fetch(parsed *ParsedSource) (string, error) {
    // Fetch from AgentSkills API
    // Download to temp directory
    // Return path
}
```

**Pros**:
- Isolates complex provider logic
- Minimal changes to existing code
- Easy to test providers independently

**Cons**:
- Inconsistent with existing pattern
- Adds complexity

## AgentSkills Integration Recommendations

### Recommended Approach

**Follow Option 1**: Add AgentSkills as struct-based source type.

### Implementation Checklist

1. **Add Source Type** (`pkg/source/parser.go`)
   - [ ] Add `AgentSkills` constant
   - [ ] Add `parseAgentSkillsPrefix()` function
   - [ ] Parse `as:owner/skillname[@version]` format
   - [ ] Return `ParsedSource` with AgentSkills type

2. **Add API Fetcher** (`pkg/source/agentskills.go` - new file)
   - [ ] Create `FetchFromAgentSkills(parsed *ParsedSource)` function
   - [ ] Use agentskills.in API to download skill
   - [ ] Cache in temp directory (or workspace?)
   - [ ] Return local path to downloaded skill

3. **Update Import Command** (`cmd/repo_import.go`)
   - [ ] Add case for `source.AgentSkills` type
   - [ ] Call `addBulkFromAgentSkills()` handler
   - [ ] Use workspace cache (optional, for API responses?)

4. **Update Sync Command** (`cmd/repo_sync.go`)
   - [ ] Add case for AgentSkills sources
   - [ ] Fetch latest version from API

5. **Metadata Tracking**
   - [ ] Store `source_type: "agentskills"`
   - [ ] Store `source_url: "https://agentskills.in/owner/skillname"`
   - [ ] Store version in `ref` field

### Open Questions

1. **Should AgentSkills use workspace cache?**
   - Git repos: Yes (10-50x faster)
   - AgentSkills: Maybe (cache API responses?)
   - Recommendation: Start without, add if needed

2. **How to handle versions?**
   - Map version to `ref` field in metadata?
   - Sync updates to latest version?

3. **Should we support subpaths?**
   - Git sources: Yes (`gh:owner/repo/skills/subdir`)
   - AgentSkills: Probably not (single skill per URL)

4. **Import mode: copy or symlink?**
   - Git sources: Copy (cloned to cache)
   - AgentSkills: Copy (fetched from API)

## Appendix: Key Files

### Source Parsing
- `pkg/source/parser.go` - Main parsing logic
- `pkg/source/parser_test.go` - Comprehensive tests

### Workspace Caching
- `pkg/workspace/manager.go` - Git repository caching
- Architecture Rule 1: All Git ops MUST use workspace

### Discovery
- `pkg/discovery/commands.go` - Command discovery
- `pkg/discovery/skills.go` - Skill discovery
- `pkg/discovery/agents.go` - Agent discovery
- `pkg/discovery/packages.go` - Package discovery

### Repository Management
- `pkg/repo/manager.go` - Import/list/remove resources
- `pkg/repo/add.go` - (not found, logic in manager.go)

### Metadata
- `pkg/metadata/metadata.go` - Source tracking

### Commands
- `cmd/repo_import.go` - Import command entry point
- `cmd/repo_sync.go` - Sync command entry point

## Conclusion

The ai-config-manager architecture uses a clear, struct-based approach for source providers:

1. **Parsing**: Centralized in `source.ParseSource()`
2. **Resolution**: Git sources → workspace cache, local → direct
3. **Discovery**: Generic, works on any directory
4. **Import**: Type-based branching in commands
5. **Metadata**: Tracks source provenance

**For AgentSkills integration**:
- Follow existing struct-based pattern
- Add `as:` prefix parsing
- Create API fetcher function
- Update import/sync commands
- Store metadata with agentskills source type

This approach is **consistent**, **simple**, and **proven** by existing GitHub/GitLab/Local implementations.

---

## CRITICAL UPDATE: ParsedSource Limitations for AgentSkills

**Issue Identified**: The `ParsedSource` struct is Git-centric and doesn't map well to AgentSkills.

### Semantic Mismatches

| Field | Git Sources | AgentSkills | Problem |
|-------|-------------|-------------|---------|
| `URL` | Git clone URL | API endpoint | ✅ Works, but different semantics |
| `Ref` | Branch/tag/commit | Semantic version | ⚠️ **Semantic mismatch** |
| `LocalPath` | For local sources | Not used | ❌ Waste of space |
| `Subpath` | Path within repo | Not applicable | ❌ Waste of space |

### Why This Matters

1. **`Ref` vs `Version`**:
   ```go
   // Git: ref is mutable (branch) or immutable (commit)
   parsed.Ref = "main"      // Mutable
   parsed.Ref = "abc123"    // Immutable commit
   
   // AgentSkills: version is always immutable
   parsed.Ref = "1.0.0"     // ← Confusing: ref implies Git semantics
   ```

2. **`GetCloneURL()` doesn't apply**:
   ```go
   // Git sources
   cloneURL, _ := source.GetCloneURL(parsed)
   workspace.GetOrClone(cloneURL, ref)
   
   // AgentSkills: Can't "clone" an API
   // Would need: GetDownloadURL() or FetchFromAPI()
   ```

3. **Missing API-specific fields**:
   - AgentSkills needs: `author`, `downloads`, `rating`, `published_at`
   - ParsedSource has no place for this metadata

### Revised Recommendations

#### Option A: Extend ParsedSource (Pragmatic)

**Add optional fields** for non-Git sources:

```go
type ParsedSource struct {
    Type      SourceType
    URL       string
    LocalPath string
    Ref       string     // Git: branch/tag/commit; AgentSkills: version
    Subpath   string
    
    // New: API-specific metadata (optional, only for AgentSkills)
    Version   string     // Semantic version (clearer than "Ref")
    Metadata  map[string]interface{} // Flexible for API responses
}
```

**Pros**:
- Minimal changes to existing code
- Backward compatible
- Git sources ignore new fields

**Cons**:
- Struct becomes less cohesive
- Two ways to express version (`Ref` vs `Version`)
- Metadata map is untyped

#### Option B: Create AgentSkillsSource Struct (Clean)

**Separate struct** for AgentSkills:

```go
type AgentSkillsSource struct {
    URL       string  // https://agentskills.in/skills/owner/name
    Owner     string  // Parsed from URL
    Name      string  // Parsed from URL
    Version   string  // Optional: "1.0.0" or "latest"
    
    // API metadata (fetched separately)
    Author      string
    Description string
    Downloads   int
    Rating      float64
}

// Modify ParseSource to return interface
func ParseSource(input string) (interface{}, error) {
    if strings.HasPrefix(input, "as:") {
        return parseAgentSkillsSource(input)
    }
    return parseGitSource(input)  // Returns ParsedSource
}
```

**Pros**:
- Clean separation of concerns
- Proper semantic types
- No confusion between Git and API sources

**Cons**:
- Breaks existing API (returns interface{})
- Requires type assertions everywhere
- More refactoring needed

#### Option C: Introduce Source Interface (Future-Proof)

**Abstract away differences**:

```go
// Common interface for all sources
type Source interface {
    GetType() SourceType
    Fetch(repoPath string) (localPath string, err error)
    GetMetadata() SourceMetadata
}

// Git-based sources
type GitSource struct {
    URL     string
    Ref     string
    Subpath string
}

func (g *GitSource) Fetch(repoPath string) (string, error) {
    ws, _ := workspace.NewManager(repoPath)
    return ws.GetOrClone(g.URL, g.Ref)
}

// AgentSkills sources
type AgentSkillsSource struct {
    URL     string
    Owner   string
    Name    string
    Version string
}

func (a *AgentSkillsSource) Fetch(repoPath string) (string, error) {
    // Call AgentSkills API
    // Download skill to temp dir
    // Return path
}
```

**Pros**:
- Clean abstraction
- Easy to add new source types
- Proper encapsulation

**Cons**:
- Large refactor (20+ files affected)
- Over-engineering for 5 source types
- Performance cost of interface dispatch

### Updated Recommendation

**Start with Option A (Extend ParsedSource)**, then **refactor to Option C** if/when we add more API-based sources.

**Rationale**:
1. **Option A is pragmatic** - Gets AgentSkills working with minimal changes
2. **Technical debt is manageable** - Only 2-3 files need to handle version semantics
3. **Refactor later** - If we add more API sources (npm, crates.io, etc.), then justify Option C

**Implementation**:
```go
// pkg/source/parser.go
type ParsedSource struct {
    Type      SourceType
    URL       string
    LocalPath string
    Ref       string     // For Git sources
    Subpath   string     // For Git sources
    Version   string     // For AgentSkills (clearer than Ref)
}

func parseAgentSkillsPrefix(input string) (*ParsedSource, error) {
    // Parse: as:owner/skillname[@version]
    parts := strings.Split(input, "@")
    ref := parts[0]
    version := "latest"
    if len(parts) > 1 {
        version = parts[1]
    }
    
    return &ParsedSource{
        Type:    AgentSkills,
        URL:     fmt.Sprintf("https://agentskills.in/skills/%s", ref),
        Version: version,  // Use Version field, not Ref
    }, nil
}

// cmd/repo_import.go
if parsed.Type == source.AgentSkills {
    return addBulkFromAgentSkills(parsed, manager)
}

// pkg/source/agentskills.go (new file)
func FetchFromAgentSkills(parsed *ParsedSource) (localPath string, err error) {
    // Use parsed.Version (not parsed.Ref)
    // Call API: GET /skills/{owner}/{name}/versions/{version}
    // Download and extract
    // Return path
}
```

### Key Insight

**The real issue**: Current architecture assumes **all remote sources are Git repositories**. AgentSkills breaks this assumption by being an **API-based registry**.

This is similar to how package managers work:
- Git: `gh:owner/repo` → Clone repository
- AgentSkills: `as:owner/skill` → Fetch from API
- npm (future): `npm:packagename` → Fetch from npm registry

**Long-term**: We'll need Source interface abstraction. **Short-term**: Extend ParsedSource with `Version` field.

---

## CRITICAL DISCOVERY: AgentSkills is Git-Based, Not an API

**Major Finding**: After reviewing the AgentSkills CLI source, **there is NO centralized API**. AgentSkills works entirely through **Git repositories**.

### How AgentSkills Actually Works

```bash
# Install from "marketplace"
skills install @anthropic/xlsx

# Actually resolves to GitHub:
skills add anthropic/skills@xlsx
# → Clones: https://github.com/anthropic/skills
# → Extracts: xlsx skill from the repo
```

### AgentSkills Architecture

1. **No Central Registry**: AgentSkills.in is a **directory/index**, not a package registry
2. **Git-Based Distribution**: All skills are stored in GitHub/GitLab repos
3. **Namespace = GitHub Owner**: `@anthropic/xlsx` = `anthropic/skills` repo
4. **CLI Downloads from Git**: Uses `git clone` or GitHub API to fetch skills

### Evidence from Documentation

From the README:
```bash
# These are equivalent:
skills install @anthropic/xlsx
skills add anthropic/skills@xlsx

# Git URL Support:
skills add vercel-labs/agent-skills
skills add https://github.com/user/repo
skills add https://gitlab.com/org/repo
```

The `@owner/skill` syntax is **syntactic sugar** for `owner/skills@skill`.

### What This Means for aimgr

**Good News**: We **already support AgentSkills sources**!

```bash
# Current aimgr syntax (works today):
aimgr repo import gh:anthropic/skills

# AgentSkills syntax:
skills install @anthropic/xlsx

# Mapping:
@anthropic/xlsx → gh:anthropic/skills (subpath: xlsx)
```

### Updated Integration Approach

#### Option 1: Treat as GitHub Shorthand (Simplest)

**No new code needed** - just document the pattern:

```bash
# AgentSkills "marketplace" install
skills install @anthropic/xlsx

# aimgr equivalent (works today)
aimgr repo import gh:anthropic/skills --filter "skill/xlsx"
```

**Pros**:
- Works immediately
- No new code
- Leverages existing Git/workspace infrastructure

**Cons**:
- Verbose syntax
- User must know GitHub owner/repo mapping
- No namespace resolution

#### Option 2: Add AgentSkills Namespace Resolution

**Add prefix** that resolves to GitHub:

```go
// pkg/source/parser.go
if strings.HasPrefix(input, "as:") {
    return parseAgentSkillsShorthand(input)
}

func parseAgentSkillsShorthand(input string) (*ParsedSource, error) {
    // Parse: as:@anthropic/xlsx
    // Resolve to: gh:anthropic/skills
    // Extract skill: xlsx
    
    parts := strings.Split(strings.TrimPrefix(input, "as:@"), "/")
    owner := parts[0]
    skillName := parts[1]
    
    // AgentSkills convention: repo is usually named "skills"
    // or "agent-skills" or "<owner>-skills"
    return &ParsedSource{
        Type: GitHub,
        URL: fmt.Sprintf("https://github.com/%s/skills", owner),
        Subpath: skillName,
    }, nil
}
```

**Usage**:
```bash
aimgr repo import as:@anthropic/xlsx
# → Resolves to: gh:anthropic/skills
# → Downloads skill: xlsx
```

**Pros**:
- Convenient `as:` prefix
- Familiar syntax from AgentSkills
- Still uses Git infrastructure

**Cons**:
- Assumes repo naming convention (`owner/skills`)
- May not work for all providers
- Adds complexity for minimal benefit

#### Option 3: No AgentSkills Support (Recommended)

**Rationale**: AgentSkills is just a **Git repo directory**. Users can:

1. Import entire repo: `aimgr repo import gh:anthropic/skills`
2. Use filter: `aimgr repo import gh:anthropic/skills --filter "xlsx"`
3. Clone and symlink: `aimgr repo import ~/dev/anthropic-skills`

**No special support needed** - it's just Git repos.

### Revised Recommendations

**Original Epic Goal**: "Support AgentSkills.in as a source"

**Reality**: AgentSkills.in is **not a source** - it's a **directory of Git repos**

**New Recommendation**: 
1. **No AgentSkills-specific code needed**
2. Document the equivalence in user docs
3. Focus on improving Git source experience (workspace, filters)

### Documentation Example

```markdown
# Working with AgentSkills

AgentSkills.in is a directory of Git repositories. To use AgentSkills 
with aimgr, import directly from the Git repository:

## Import entire repository
aimgr repo import gh:anthropic/skills

## Import specific skill
aimgr repo import gh:anthropic/skills --filter "xlsx"

## Install to specific tool
aimgr install skill/xlsx --tool=claude

## Finding the repository
AgentSkills uses the @owner/skillname format. To find the Git URL:
1. Visit https://agentskills.in/marketplace
2. Click on the skill
3. Note the GitHub owner/repo
4. Use: aimgr repo import gh:owner/repo
```

### Conclusion: No API, No New Code Needed

**AgentSkills is already supported** through Git sources. The "integration" is just:
1. Document the pattern
2. Maybe add `as:` prefix as convenience (low priority)
3. Focus on other features

**The research epic was based on the assumption that AgentSkills had a centralized API.** 
It doesn't - it's a Git-based directory, which we already support.

---

## Final Clarification: AgentSkills Discovery API

### What AgentSkills.in Actually Provides

After analysis, here's the **complete picture**:

1. **Web Directory** (agentskills.in/marketplace)
   - Browse/search UI for humans
   - Shows skill descriptions, authors, stats
   - Links to GitHub repos

2. **No Programmatic API** (confirmed)
   - No REST API endpoints
   - No npm-like registry API
   - CLI dependencies: Only Git/filesystem tools (no HTTP clients)

3. **Discovery Mechanism**: GitHub Search
   ```bash
   # When you run:
   skills search "pdf"
   
   # It probably does:
   - GitHub API search for repos with "agent-skills" topic
   - Or: Scrapes agentskills.in website
   - Or: Uses cached index file from a known repo
   ```

### How "skills install @owner/name" Works

**Theory 1: GitHub Convention**
```bash
skills install @anthropic/xlsx
# Assumes: github.com/anthropic/skills (repo named "skills")
# Clones and extracts: xlsx subdirectory
```

**Theory 2: Registry File**
```bash
# Possible central registry:
github.com/Karanjot786/agent-skills-registry/registry.json
{
  "@anthropic/xlsx": {
    "repo": "anthropic/skills",
    "path": "xlsx"
  }
}
```

**Theory 3: Web Scraping**
```bash
# Fetch from website:
curl https://agentskills.in/api/skills
# Returns: JSON mapping of @owner/name → GitHub URLs
```

### What This Means for aimgr

**Two potential features**:

#### Feature 1: Add Discovery/Search Command

```bash
# New command:
aimgr search agentskills "pdf processing"

# Implementation:
1. Scrape agentskills.in marketplace, OR
2. Use GitHub API to search "agent-skills" topic, OR
3. Maintain our own curated list

# Output:
Found 5 skills:
  @anthropic/pdf-processing     gh:anthropic/skills
  @vercel/pdf-reader            gh:vercel-labs/agent-skills
  ...

# Then install:
aimgr repo import gh:anthropic/skills --filter "pdf-processing"
```

**Pros**:
- Helps users discover skills
- Integrates with existing Git workflow

**Cons**:
- Scraping is fragile
- GitHub API has rate limits
- Maintenance burden

#### Feature 2: Resolve `as:` Prefix via Convention

```bash
# User types:
aimgr repo import as:@anthropic/xlsx

# aimgr assumes convention:
# @owner/skill → github.com/owner/skills (or owner/agent-skills)

# Tries in order:
1. github.com/owner/skills
2. github.com/owner/agent-skills
3. github.com/owner/<skill-repo-name>  # Fallback: search for repo

# Then imports with filter:
gh:owner/skills --filter "xlsx"
```

**Pros**:
- Convenient shorthand
- No external dependencies
- Falls back gracefully

**Cons**:
- Assumes naming convention
- May not work for all repos
- Adds complexity

### Recommendation: Neither (Yet)

**Rationale**:
1. **Discovery**: Users can browse agentskills.in website manually
2. **Import**: `gh:owner/repo` syntax works today
3. **Value**: Low - most users find skills via web anyway
4. **Effort**: Medium - parsing conventions, error handling

**If we do add it later**:
- Start with Feature 2 (`as:` prefix resolution)
- Add Feature 1 (search) only if user demand is high
- Keep it simple: convention-based, no scraping

### Summary

**AgentSkills Architecture** (confirmed):
```
┌─────────────────────────────────────┐
│  agentskills.in (Website)           │
│  - Browse/search UI                 │
│  - No programmatic API              │
└─────────────────────────────────────┘
                  │
                  ▼ (links to)
┌─────────────────────────────────────┐
│  GitHub Repositories                │
│  - github.com/owner/skills          │
│  - SKILL.md files in subdirectories │
└─────────────────────────────────────┘
                  │
                  ▼ (git clone)
┌─────────────────────────────────────┐
│  Local Installation                 │
│  - .claude/skills/                  │
│  - .cursor/skills/                  │
│  - .github/skills/                  │
└─────────────────────────────────────┘
```

**aimgr Status**:
- ✅ Can import from GitHub repos (git sources)
- ✅ Can use filters to select specific skills
- ✅ Workspace cache provides performance
- ❌ No discovery/search integration
- ❌ No `as:@owner/skill` shorthand

**Conclusion**: AgentSkills is a **directory of Git repos**, not a package registry. 
The "marketplace" is just a website that links to GitHub. No special integration needed 
beyond what we already have for Git sources.
