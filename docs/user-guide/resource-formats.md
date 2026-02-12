# Resource Format Specifications

This document provides detailed format specifications for all aimgr resource types.

## Table of Contents
- [Package Format](#package-format)
- [Agent Format](#agent-format)
- [Marketplace Format](#marketplace-format)
- [Project Manifests](#project-manifests)

---

## Package Format

Packages are collections of resources that can be installed together as a unit. They are stored as JSON files in the `packages/` directory.

### File Structure

Package files are stored as:
- **Repository**: `~/.local/share/ai-config/repo/packages/<package-name>.package.json`
- **Metadata**: `~/.local/share/ai-config/repo/.metadata/packages/<package-name>-metadata.json`

Packages are **not** installed directly into projects. Instead, when you install a package, all its referenced resources are installed as symlinks.

### Package JSON Format

```json
{
  "name": "package-name",
  "description": "Human-readable description of the package",
  "resources": [
    "command/command-name",
    "skill/skill-name",
    "agent/agent-name"
  ]
}
```

**Required Fields:**

- **name** (string): Package name following agentskills.io naming rules
  - Lowercase alphanumeric + hyphens only
  - Cannot start/end with hyphen
  - No consecutive hyphens (`--`)
  - 1-64 characters max
  - Examples: `web-tools`, `testing-suite`, `docs-helpers`

- **description** (string): Human-readable description (1-1024 characters)

- **resources** ([]string): Array of resource references in `type/name` format
  - Format: `type/name` where type is `command`, `skill`, or `agent`
  - Examples: `command/test`, `skill/pdf-processing`, `agent/code-reviewer`
  - All referenced resources must exist in the repository

### Format Examples

**Minimal Package:**
```json
{
  "name": "minimal-tools",
  "description": "Minimal tool collection",
  "resources": [
    "command/test"
  ]
}
```

**Web Development Package:**
```json
{
  "name": "web-dev-tools",
  "description": "Complete web development toolkit",
  "resources": [
    "command/build",
    "command/dev",
    "skill/typescript-helper",
    "skill/react-helper",
    "agent/code-reviewer"
  ]
}
```

**Testing Suite Package:**
```json
{
  "name": "testing-suite",
  "description": "Complete testing toolkit with commands, skills, and QA agent",
  "resources": [
    "command/test",
    "command/coverage",
    "skill/test-generator",
    "skill/mock-helper",
    "agent/qa-tester"
  ]
}
```

### Code Structure

The package resource is represented by the `Package` struct in `pkg/resource/package.go`:

```go
type Package struct {
    Name        string   `json:"name"`        // Package name
    Description string   `json:"description"` // Human-readable description
    Resources   []string `json:"resources"`   // Resource references (type/name)
}
```

**Loading packages:**
```go
// Load package from file
pkg, err := resource.LoadPackage("path/to/package.package.json")

// Get package path in repo
packagePath := resource.GetPackagePath("package-name", repoPath)
```

**Creating packages:**
```go
pkg := &resource.Package{
    Name:        "web-tools",
    Description: "Web development tools",
    Resources:   []string{"command/build", "skill/typescript-helper"},
}

err := resource.SavePackage(pkg, repoPath)
```

**Parsing resource references:**
```go
// Parse type/name format
resType, resName, err := resource.ParseResourceReference("command/test")
// Returns: resource.Command, "test", nil
```

### CLI Commands


**Install a package:**
```bash
# Installs all resources in the package
aimgr install package/web-tools
```

**Uninstall a package:**
```bash
# Removes all resource symlinks
aimgr uninstall package/web-tools
```

**Remove a package:**
```bash
# Remove package only (keeps resources)
aimgr repo remove package/web-tools

# Remove package and all its resources
aimgr repo remove package/web-tools --with-resources
```

**List packages:**
```bash
# List all packages in repository
aimgr repo list package
```

---

## Agent Format

Agents are single `.md` files with YAML frontmatter that define AI agents with specialized roles and capabilities. The codebase supports both OpenCode and Claude Code formats.

### File Structure

Agent files are stored as:
- **Repository**: `~/.local/share/ai-config/repo/agents/<agent-name>.md`
- **Installed**: `.claude/agents/<agent-name>.md` or `.opencode/agents/<agent-name>.md`

### YAML Frontmatter Fields

**Required:**
- `description` (string): Brief description of what the agent does (1-1024 chars for validation)

**Optional (OpenCode format):**
- `type` (string): Agent type/role (e.g., "code-reviewer", "tester", "documentation")
- `instructions` (string): Detailed instructions for the agent's behavior
- `capabilities` ([]string): List of agent capabilities (e.g., "static-analysis", "security-scan")

**Optional (Common):**
- `version` (string): Semantic version (e.g., "1.0.0")
- `author` (string): Author name or organization
- `license` (string): License identifier (e.g., "MIT", "Apache-2.0")
- `metadata` (map): Additional metadata fields

### Format Examples

**Minimal Agent (Claude format):**
```yaml
---
description: Minimal agent for testing
---

# Minimal Agent

This agent has only the required description field.
```

**OpenCode Format Agent:**
```yaml
---
description: Code review agent that checks for best practices
type: code-reviewer
instructions: Review code for quality, security, and performance
capabilities:
  - static-analysis
  - security-scan
  - performance-review
version: "1.0.0"
author: team-name
license: MIT
---

# Code Reviewer Agent

This agent analyzes code for quality and security issues.
```

**Claude Format Agent:**
```yaml
---
description: A Claude format agent for testing
version: "2.0.0"
author: claude-team
license: Apache-2.0
metadata:
  category: development
  tags: testing,qa
---

# Claude Agent

Instructions and guidelines go in the markdown body for Claude format.
```

### Code Structure

The agent resource is represented by the `AgentResource` struct in `pkg/resource/agent.go`:

```go
type AgentResource struct {
    Resource                                     // Embedded base resource
    Type         string   `yaml:"type,omitempty"`         // Agent role (OpenCode)
    Instructions string   `yaml:"instructions,omitempty"` // Agent instructions (OpenCode)
    Capabilities []string `yaml:"capabilities,omitempty"` // Capability list
    Content      string   `yaml:"-"`                      // Markdown content
}
```

**Loading agents:**
```go
// Load base resource (minimal info)
res, err := resource.LoadAgent("path/to/agent.md")

// Load full agent with all details
agent, err := resource.LoadAgentResource("path/to/agent.md")
```

**Creating agents:**
```go
agent := resource.NewAgentResource("code-reviewer", "Reviews code quality")
agent.Type = "code-reviewer"
agent.Instructions = "Check for best practices"
agent.Capabilities = []string{"static-analysis", "security"}
agent.Content = "# Code Reviewer\n\nDetailed docs here."

err := resource.WriteAgent(agent, "code-reviewer.md")
```

### Tool Support

OpenCode and VSCode / GitHub Copilot are supported as first-class AI tools alongside Claude Code.

**Tool Directories:**

| Tool | Commands | Skills | Agents |
|------|----------|--------|--------|
| Claude Code | `.claude/commands/` | `.claude/skills/` | `.claude/agents/` |
| OpenCode | `.opencode/commands/` | `.opencode/skills/` | `.opencode/agents/` |
| VSCode / GitHub Copilot | N/A | `.github/skills/` | N/A |
| Windsurf | N/A | `.windsurf/skills/` | N/A |

**VSCode / GitHub Copilot Support:**
- Only supports skills (no commands or agents)
- Skills use the same `SKILL.md` format as Claude Code and OpenCode
- Follows the [Agent Skills standard](https://www.agentskills.io/) at agentskills.io
- Compatible with both the Copilot CLI and coding agent features
- Use `--tool=copilot` or `--tool=vscode` (both names work)

**Windsurf Support:**
- Only supports skills (no commands or agents)
- Skills use the same `SKILL.md` format as other tools
- Follows the [Agent Skills standard](https://www.agentskills.io/) at agentskills.io
- Use `--tool=windsurf`

**OpenCode-Specific Features:**

OpenCode agents support additional frontmatter fields:
- `type`: Agent role/category
- `instructions`: Detailed behavior instructions
- `capabilities`: Array of capability strings

These fields are optional and allow more structured agent definitions. Claude format agents typically put instructions in the markdown body instead.

### Bulk Import

Import all resources from directories using auto-discovery:

```bash
# Import from .opencode folder
aimgr repo import ~/.opencode
aimgr repo import ~/project/.opencode

# Import from .claude folder
aimgr repo import ~/.claude
aimgr repo import ~/project/.claude

# Filter specific resource types
aimgr repo import ~/.opencode --filter "skill/*"
aimgr repo import ~/project/.claude --filter "agent/*"
aimgr repo import gh:owner/repo --filter "package/*"
```

This imports:
- Commands from `.opencode/commands/*.md` or `.claude/commands/*.md`
- Skills from `.opencode/skills/*/SKILL.md` or `.claude/skills/*/SKILL.md`
- Agents from `.opencode/agents/*.md` or `.claude/agents/*.md`
- Packages from `packages/*.package.json`

---

## Marketplace Format

Claude plugin marketplaces are defined using a `marketplace.json` file that declares a collection of plugins. aimgr can import these marketplaces and automatically generate packages.

### File Structure

Marketplace files are typically located at:
- `.claude-plugin/marketplace.json` (Claude plugin convention)
- Any custom location (specified during import)

### Marketplace JSON Format

aimgr supports two marketplace JSON formats:

**Traditional Format** (description at top level):
```json
{
  "name": "marketplace-name",
  "version": "1.0.0",
  "description": "Human-readable description of the marketplace",
  "owner": {
    "name": "Organization Name",
    "email": "contact@example.com"
  },
  "plugins": [...]
}
```

**Anthropics Format** (description in metadata):
```json
{
  "name": "marketplace-name",
  "owner": {
    "name": "Organization Name",
    "email": "contact@example.com"
  },
  "metadata": {
    "description": "Human-readable description",
    "version": "1.0.0"
  },
  "plugins": [...]
}
```

### Required Fields

**Marketplace level:**
- **name** (string): Marketplace name
- **plugins** ([]Plugin): Array of plugin definitions

**Plugin level:**
- **name** (string): Plugin name (becomes package name)
- **description** (string): Plugin description
- **source** (string): Relative path to plugin resources

### Optional Fields

**Marketplace level:**
- **description** (string): Marketplace description (can be at top level or in metadata)
- **version** (string): Marketplace version (can be at top level or in metadata)
- **owner** (Author): Marketplace owner information
- **metadata** (Metadata): Metadata container (Anthropics format)

**Plugin level:**
- **category** (string): Plugin category (e.g., "development", "testing", "documentation")
- **version** (string): Plugin version (semver)
- **author** (Author): Plugin author information

**Author object:**
- **name** (string): Name
- **email** (string): Email address

**Metadata object:**
- **description** (string): Description (Anthropics format)
- **version** (string): Version (Anthropics format)

### Code Structure

The marketplace format is represented by structs in `pkg/marketplace/parser.go`:

```go
type MarketplaceConfig struct {
    Name        string    `json:"name"`
    Version     string    `json:"version,omitempty"`
    Description string    `json:"description,omitempty"`
    Owner       *Author   `json:"owner,omitempty"`
    Metadata    *Metadata `json:"metadata,omitempty"`
    Plugins     []Plugin  `json:"plugins"`
}

// GetDescription returns description from either top-level or metadata
func (c *MarketplaceConfig) GetDescription() string

// GetVersion returns version from either top-level or metadata
func (c *MarketplaceConfig) GetVersion() string

type Metadata struct {
    Description string `json:"description,omitempty"`
    Version     string `json:"version,omitempty"`
}

type Plugin struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Source      string  `json:"source"`
    Category    string  `json:"category,omitempty"`
    Version     string  `json:"version,omitempty"`
    Author      *Author `json:"author,omitempty"`
}

type Author struct {
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`
}
```

**Parsing:**
```go
import "github.com/hk9890/ai-config-manager/pkg/marketplace"

// Parse marketplace file
config, err := marketplace.ParseMarketplace("path/to/marketplace.json")
if err != nil {
    return err
}

// Access marketplace data
fmt.Println("Marketplace:", config.Name)
for _, plugin := range config.Plugins {
    fmt.Printf("Plugin: %s (%s)\n", plugin.Name, plugin.Description)
}
```

**Generating packages:**
```go
import "github.com/hk9890/ai-config-manager/pkg/marketplace"

// Generate aimgr packages from marketplace
basePath := "/path/to/marketplace/directory"
packages, err := marketplace.GeneratePackages(config, basePath)
if err != nil {
    return err
}

// Each plugin becomes a package
for _, pkg := range packages {
    fmt.Printf("Package: %s (%d resources)\n", pkg.Name, len(pkg.Resources))
}
```

### Resource Discovery

When importing a marketplace plugin, aimgr searches for resources in standard locations within the plugin's source directory:

**Commands:**
1. `commands/*.md`
2. `.claude/commands/*.md`
3. `.opencode/commands/*.md`

**Skills:**
1. `skills/*/SKILL.md`
2. `.claude/skills/*/SKILL.md`
3. `.opencode/skills/*/SKILL.md`

**Agents:**
1. `agents/*.md`
2. `.claude/agents/*.md`
3. `.opencode/agents/*.md`

All discovered resources are imported into the repository and referenced in the generated package.

### Package Generation

For each plugin in the marketplace:

1. **Resource Discovery**: Scan plugin source directory for resources
2. **Resource Import**: Copy resources to repository
3. **Package Creation**: Create package JSON file
4. **Metadata Tracking**: Save package metadata

**Generated package structure:**
```json
{
  "name": "plugin-name",
  "description": "Plugin description from marketplace",
  "resources": [
    "command/build",
    "skill/typescript-helper",
    "agent/code-reviewer"
  ]
}
```

**Package metadata structure:**
```json
{
  "name": "plugin-name",
  "source_type": "marketplace",
  "source_url": "file:///path/to/marketplace.json",
  "first_added": "2026-01-25T12:00:00Z",
  "last_updated": "2026-01-25T12:00:00Z",
  "resource_count": 3,
  "original_format": "claude-plugin"
}
```

### CLI Commands

**Import marketplace:**
```bash
# Import from local path
aimgr marketplace import ~/.claude-plugin/marketplace.json

# Import from GitHub (requires gh CLI)
aimgr marketplace import gh:owner/repo/.claude-plugin/marketplace.json

# With options
aimgr marketplace import marketplace.json --dry-run     # Preview
aimgr marketplace import marketplace.json --force       # Overwrite
aimgr marketplace import marketplace.json --filter "web-*"  # Filter plugins
```

**Import workflow:**
1. Parse `marketplace.json`
2. Filter plugins (if `--filter` specified)
3. For each plugin:
   - Discover resources in plugin source directory
   - Import resources to repository
   - Create package with resource references
   - Save package metadata
4. Display summary

### Testing

Test files for marketplace functionality:
- `pkg/marketplace/parser_test.go` - Marketplace parsing tests
- `pkg/marketplace/generator_test.go` - Package generation tests

**Example test:**
```go
func TestParseMarketplace(t *testing.T) {
    config, err := marketplace.ParseMarketplace("testdata/marketplace.json")
    if err != nil {
        t.Fatalf("Failed to parse: %v", err)
    }
    
    if config.Name != "expected-name" {
        t.Errorf("Expected name %q, got %q", "expected-name", config.Name)
    }
}
```

### Format Examples

See [examples/marketplace/](../examples/marketplace/) for complete examples:
- `marketplace.json` - Full marketplace with multiple plugins
- `minimal-marketplace.json` - Minimal required fields
- `plugin-structure/` - Example plugin directory structure

---

## Project Manifests

Similar to npm's `package.json`, the `ai.package.yaml` file allows you to declare project dependencies for AI resources. This makes it easy to share projects with consistent AI tooling and manage resources declaratively.

### File Structure

The manifest file is placed in your project root:
- **Location**: `./ai.package.yaml` (current directory)
- **Format**: YAML
- **Optional**: Projects can work without this file

### YAML Format

```yaml
# ai.package.yaml
resources:
  - skill/pdf-processing
  - skill/typescript-helper
  - command/test
  - command/build
  - agent/code-reviewer
  - package/web-tools

# Optional: override default install targets
targets:
  - claude
  - opencode
```

**Fields:**

- **resources** ([]string, required): Array of resource references in `type/name` format
  - Format: `type/name` where type is `command`, `skill`, `agent`, or `package`
  - Examples: `skill/pdf-processing`, `command/test`, `agent/code-reviewer`, `package/web-tools`
  - Uses the same format as CLI commands throughout aimgr

- **targets** ([]string, optional): Override default install targets
  - Valid values: `claude`, `opencode`, `copilot`, `windsurf`
  - If not specified, uses defaults from `~/.config/aimgr/aimgr.yaml`
  - Allows per-project tool preferences

### Workflows

**Create a new manifest:**
```bash
cd my-project
aimgr init
# Creates ai.package.yaml with empty resources array
```

**Or let it be created automatically:**
```bash
cd my-project
aimgr install skill/pdf-processing
# Installs skill and creates ai.package.yaml automatically
```

**Install all dependencies:**
```bash
cd existing-project
aimgr install
# Reads ai.package.yaml and installs all listed resources
```

**Install without saving:**
```bash
aimgr install skill/temporary-test --no-save
# Installs skill but doesn't add to ai.package.yaml
```

**Install specific resource:**
```bash
aimgr install command/build
# Installs command and adds it to ai.package.yaml automatically
```

### CLI Commands

**Install from manifest:**
```bash
# Installs all resources from ai.package.yaml
aimgr install
```

**Install with auto-save (default):**
```bash
# Installs and adds to ai.package.yaml
aimgr install skill/test
```

**Install without saving:**
```bash
# Installs but skips ai.package.yaml update
aimgr install skill/test --no-save
```

### Format Examples

**Minimal manifest:**
```yaml
resources:
  - skill/pdf-processing
```

**Basic project:**
```yaml
resources:
  - skill/web-development
  - skill/testing-helper
  - command/test
  - command/build
  - agent/code-reviewer
```

**With custom targets:**
```yaml
resources:
  - skill/typescript-helper
  - command/deploy
  - package/web-tools

targets:
  - claude
  - opencode
```

**Full example:**
```yaml
# Complete development environment
resources:
  # Skills for development
  - skill/typescript-helper
  - skill/react-helper
  - skill/testing-helper
  
  # Commands for automation
  - command/test
  - command/build
  - command/deploy
  
  # Agents for quality
  - agent/code-reviewer
  - agent/qa-tester
  
  # Package bundles
  - package/web-tools

# Install to both Claude and OpenCode
targets:
  - claude
  - opencode
```

### Code Structure

The manifest format is handled by the `pkg/manifest/` package:

```go
type Manifest struct {
    Resources []string `yaml:"resources"`           // Resource references (type/name)
    Targets   []string `yaml:"targets,omitempty"`   // Optional install targets
}
```

**Loading manifests:**
```go
import "github.com/hk9890/ai-config-manager/pkg/manifest"

// Load from file
m, err := manifest.Load("ai.package.yaml")

// Check if file exists
exists := manifest.Exists("ai.package.yaml")
```

**Creating/updating manifests:**
```go
m := &manifest.Manifest{
    Resources: []string{"skill/test", "command/build"},
    Targets:   []string{"claude"},
}

// Save to file
err := m.Save("ai.package.yaml")

// Add a resource
err = m.Add("skill/new-skill")

// Remove a resource
err = m.Remove("skill/old-skill")
```

### Best Practices

1. **Version control**: Commit `ai.package.yaml` to your repository
2. **Team consistency**: Everyone on the team gets the same AI resources
3. **Simple resources list**: Use plain `type/name` format for clarity
4. **Override targets when needed**: Set project-specific tool preferences
5. **Use packages**: Group related resources into packages for reuse
6. **Keep it minimal**: Only include resources actually used in the project

### Comparison with npm

| npm | aimgr |
|-----|-------|
| `package.json` | `ai.package.yaml` |
| `npm install` | `aimgr install` |
| `npm install <package>` | `aimgr install skill/name` |
| `npm install --no-save` | `aimgr install --no-save` |
| `dependencies` array | `resources` array |

### Notes

- **Version 1**: Simple format with just resources and targets
- **Extensible**: Future versions may add versioning, constraints, etc.
- **Backward compatible**: Optional file, projects work without it
- **No validation at parse time**: Resource existence checked at install time
