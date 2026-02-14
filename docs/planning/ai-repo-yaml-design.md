# ai.repo.yaml Design Specification

**Status**: Draft
**Created**: 2026-02-14
**Beads Issue**: ai-config-manager-d9g

## Problem Statement

Currently, sync sources are defined globally in `~/.config/aimgr/aimgr.yaml` under `sync.sources`. This creates several problems:

1. **No self-describing repositories**: A Git repository containing AI resources has no way to declare what it contains, how resources should be filtered, or what metadata applies.
2. **Source knowledge lives outside the source**: The user must manually configure URLs, filters, and modes in their global config. If someone shares a repository URL, the recipient must know what filter patterns to use.
3. **No per-source configuration**: All sources share the same global behavior. There's no way for a repository author to specify defaults (e.g., "this repo should always be imported as copy mode" or "only these resources are public").
4. **Discovery is implicit**: `repo import` discovers resources by scanning directory structure. There's no manifest to declare intent vs. accident (e.g., test fixtures that look like real resources).

## Goals

- Allow repository authors to declare what AI resources their repository contains
- Enable `aimgr repo import <url>` to work without additional flags (the repo self-describes)
- Maintain backward compatibility with repositories that don't have `ai.repo.yaml`
- Keep the design simple and consistent with existing `ai.package.yaml` patterns

## Non-Goals

- Replacing `sync.sources` in global config (that remains for user-side orchestration)
- Package registry or marketplace features
- Authentication or access control
- Dependency resolution between repositories

## Design

### File Location and Name

```
my-ai-resources/
  ai.repo.yaml          # Repository manifest (root level)
  commands/
    build.md
    deploy.md
  skills/
    pdf-processing/
      SKILL.md
  agents/
    code-reviewer.md
```

The file MUST be named `ai.repo.yaml` and placed at the repository root (or the root of the imported path for monorepo subpaths).

### YAML Schema

```yaml
# ai.repo.yaml - Repository manifest for AI resources

# Human-readable name for this repository (optional)
name: "My AI Resources"

# Short description (optional)
description: "Collection of skills and commands for development workflows"

# Resources to expose from this repository
# If omitted, falls back to auto-discovery (current behavior)
resources:
  # Explicit list of resources in type/name format
  - command/build
  - command/deploy
  - skill/pdf-processing
  - agent/code-reviewer

# Alternatively, use include/exclude patterns for declarative filtering
# (mutually exclusive with explicit resources list)
filter:
  include:
    - "skill/*"
    - "command/build"
  exclude:
    - "command/test-*"     # Exclude test fixtures
    - "skill/experimental-*"

# Default import mode (optional, can be overridden by consumer)
# "copy" = copy files into repo (default for remote)
# "symlink" = create symlinks (default for local)
# "auto" = let aimgr decide based on source type (default)
mode: auto
```

### Struct Definition

```go
// pkg/repomanifest/manifest.go

package repomanifest

const FileName = "ai.repo.yaml"

// Manifest represents a repository's AI resource declaration
type Manifest struct {
    // Name is a human-readable name for this repository (optional)
    Name string `yaml:"name,omitempty"`

    // Description is a short description (optional)
    Description string `yaml:"description,omitempty"`

    // Resources is an explicit list of resources in "type/name" format
    // Mutually exclusive with Filter
    Resources []string `yaml:"resources,omitempty"`

    // Filter declares include/exclude patterns for resource discovery
    // Mutually exclusive with Resources
    Filter *FilterConfig `yaml:"filter,omitempty"`

    // Mode is the default import mode: "copy", "symlink", or "auto" (default)
    Mode string `yaml:"mode,omitempty"`
}

// FilterConfig declares which resources to include/exclude via glob patterns
type FilterConfig struct {
    // Include patterns (glob). If set, only matching resources are exposed.
    Include []string `yaml:"include,omitempty"`

    // Exclude patterns (glob). Matched resources are excluded, even if they match include.
    Exclude []string `yaml:"exclude,omitempty"`
}
```

### Package Location

New package: `pkg/repomanifest/`

Rationale: The existing `pkg/manifest/` handles `ai.package.yaml` (project-level, consumer-side). The repo manifest is fundamentally different -- it's source-side, declaring what a repository offers. A separate package keeps these concerns cleanly separated.

Files:
- `pkg/repomanifest/manifest.go` -- Manifest struct, Load, Save, Validate
- `pkg/repomanifest/manifest_test.go` -- Unit tests
- `pkg/repomanifest/doc.go` -- Package documentation

### API

```go
// Load loads a repo manifest from a YAML file
func Load(path string) (*Manifest, error)

// LoadFromDir loads ai.repo.yaml from a directory (convenience)
func LoadFromDir(dir string) (*Manifest, error)

// Exists checks if ai.repo.yaml exists in a directory
func Exists(dir string) bool

// Save writes the manifest to a YAML file
func (m *Manifest) Save(path string) error

// Validate checks if the manifest is valid
func (m *Manifest) Validate() error

// MatchResource checks if a resource (in "type/name" format) is exposed by this manifest
func (m *Manifest) MatchResource(resource string) bool

// ListExposedResources returns all resources this manifest exposes,
// given a list of discovered resources
func (m *Manifest) ListExposedResources(discovered []string) []string
```

### Validation Rules

1. `name` -- optional, max 128 characters
2. `description` -- optional, max 512 characters
3. `resources` and `filter` are mutually exclusive (error if both set)
4. If `resources` is set, each entry must be valid `type/name` format (reuse `validateResourceReference` logic from `pkg/manifest`)
5. If `filter.include` patterns are set, they must be valid glob patterns (reuse `pkg/pattern`)
6. If `filter.exclude` patterns are set, they must be valid glob patterns
7. `mode` must be one of: `""` (empty/default), `"auto"`, `"copy"`, `"symlink"`
8. If neither `resources` nor `filter` is set, all discovered resources are exposed (backward compatible)

### Integration Points

#### 1. `repo import` Command

Current flow:
```
repo import <path-or-url>
  -> discover resources in directory
  -> AddBulk() all discovered resources
```

New flow with ai.repo.yaml:
```
repo import <path-or-url>
  -> check for ai.repo.yaml in source directory
  -> if found:
      -> load and validate manifest
      -> discover resources in directory
      -> filter discovered resources through manifest
      -> AddBulk() only exposed resources
      -> store manifest metadata (name, description) in source metadata
  -> if not found:
      -> current behavior (discover all, no filtering)
```

Changes needed in: `cmd/repo_import.go` -- after source path is resolved, before `addBulkFromLocalWithMode()`.

#### 2. `repo sync` Command

Current flow:
```
repo sync
  -> load global config sync.sources
  -> for each source:
      -> resolve path (clone if remote)
      -> apply source.filter from global config
      -> addBulkFromLocalWithMode()
```

New flow:
```
repo sync
  -> load global config sync.sources
  -> for each source:
      -> resolve path (clone if remote)
      -> check for ai.repo.yaml in source
      -> if found:
          -> apply repo manifest filtering FIRST
          -> then apply source.filter from global config ON TOP
      -> addBulkFromLocalWithMode()
```

This means the global `filter` is an additional constraint, not a replacement. The repo author controls what's exposed; the consumer can further narrow it.

#### 3. `repo info` / `repo describe` Commands

If the source has an `ai.repo.yaml`, display its metadata:

```
$ aimgr repo info
Repository: ~/.local/share/ai-config/repo/
Resources: 12 (5 commands, 4 skills, 3 agents)

Sources with manifests:
  gh:company/tools@v2.0.0
    Name: Company AI Tools
    Description: Standard tools for engineering team
    Mode: copy
    Exposed: 8/12 resources (filtered by manifest)
```

#### 4. Metadata Storage

When importing from a source with `ai.repo.yaml`, store additional metadata:

```json
// .metadata/skills/pdf-processing-metadata.json
{
  "name": "pdf-processing",
  "type": "skill",
  "source_type": "github",
  "source_url": "https://github.com/company/tools",
  "ref": "v2.0.0",
  "repo_manifest": true,
  "repo_name": "Company AI Tools",
  "first_installed": "2026-02-14T10:00:00Z",
  "last_updated": "2026-02-14T10:00:00Z"
}
```

New fields: `repo_manifest` (bool), `repo_name` (string, from manifest).

### Resource Matching Logic

The `MatchResource` method implements the following priority:

1. **Explicit resources list**: If `resources` is set, resource must be in the list.
2. **Include/Exclude patterns**: If `filter` is set:
   - If `include` is non-empty, resource must match at least one include pattern
   - If `exclude` is non-empty, resource must NOT match any exclude pattern
   - Exclude takes priority over include
3. **No filter**: If neither `resources` nor `filter` is set, all resources match (backward compatible).

```go
func (m *Manifest) MatchResource(resource string) bool {
    // Explicit list mode
    if len(m.Resources) > 0 {
        for _, r := range m.Resources {
            if r == resource {
                return true
            }
        }
        return false
    }

    // Filter mode
    if m.Filter != nil {
        // Check exclude first (takes priority)
        for _, pattern := range m.Filter.Exclude {
            if matchGlob(pattern, resource) {
                return false
            }
        }
        // Check include
        if len(m.Filter.Include) > 0 {
            for _, pattern := range m.Filter.Include {
                if matchGlob(pattern, resource) {
                    return true
                }
            }
            return false // Didn't match any include pattern
        }
    }

    // No resources list, no filter -> expose everything
    return true
}
```

### Backward Compatibility

- Repositories without `ai.repo.yaml` continue to work exactly as today (auto-discover all)
- The global `sync.sources[].filter` continues to work and is applied ON TOP of repo manifest filtering
- No changes to `ai.package.yaml` behavior (consumer-side manifest is separate)
- Existing metadata format is extended (new optional fields), not replaced

### Example Scenarios

#### Scenario 1: Community Repository

```yaml
# ai.repo.yaml in gh:community/ai-skills
name: "Community AI Skills"
description: "Curated collection of AI skills for common tasks"
filter:
  include:
    - "skill/*"
  exclude:
    - "skill/experimental-*"
    - "skill/deprecated-*"
```

User imports with: `aimgr repo import gh:community/ai-skills`
Result: Only stable skills are imported. Test fixtures, experimental, and deprecated skills are excluded.

#### Scenario 2: Monorepo with Mixed Content

```yaml
# ai.repo.yaml in company monorepo under /ai-tools/
name: "Company AI Tools"
description: "Engineering team AI resources"
resources:
  - skill/code-review
  - skill/pr-summary
  - command/deploy
  - agent/oncall-helper
```

User imports with: `aimgr repo import gh:company/monorepo/ai-tools`
Result: Only the 4 declared resources are imported, even though the directory may contain other files.

#### Scenario 3: Global Config + Repo Manifest

```yaml
# ~/.config/aimgr/aimgr.yaml (consumer)
sync:
  sources:
    - url: "gh:community/ai-skills@v2.0.0"
      filter: "skill/pdf*"    # Consumer wants only PDF skills
```

```yaml
# ai.repo.yaml in gh:community/ai-skills (source)
filter:
  include:
    - "skill/*"
  exclude:
    - "skill/experimental-*"
```

Result: Repo manifest allows all skills except experimental. Consumer further narrows to only `skill/pdf*`. Final result: only stable PDF skills are imported.

## Implementation Plan

### Phase 1: Core Package (pkg/repomanifest)
- Manifest struct, Load, Save, Validate
- MatchResource and ListExposedResources
- Unit tests with fixtures

### Phase 2: Import Integration
- Modify `cmd/repo_import.go` to check for ai.repo.yaml
- Apply manifest filtering before AddBulk
- Store manifest metadata

### Phase 3: Sync Integration
- Modify `cmd/repo_sync.go` to layer repo manifest + global filter
- Update sync summary output to show manifest info

### Phase 4: Info/Describe Integration
- Show ai.repo.yaml metadata in `repo info` and `repo describe`
- Surface manifest filtering in output

### Phase 5: Documentation
- User guide for ai.repo.yaml format
- Update sync-sources.md with manifest interaction
- Update resource-formats.md

## Open Questions

1. **Should `aimgr repo init` generate an ai.repo.yaml?** Currently it creates directory structure + git init. Could optionally scaffold a manifest too.

2. **Should there be an `aimgr repo manifest` subcommand?** For generating/editing ai.repo.yaml interactively (similar to `npm init`).

3. **Version field?** Should ai.repo.yaml declare a version for the manifest format itself (for future schema evolution)?

4. **Dependencies between repos?** Should ai.repo.yaml be able to declare dependencies on other repos? (Probably not in v1 -- keep it simple.)

5. **Should `mode` in ai.repo.yaml override or just suggest?** Current design: it's a default that the consumer can override. Is that the right behavior?
