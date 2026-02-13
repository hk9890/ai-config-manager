---
name: ai-resource-manager
description: "Manage AI resources (skills, commands, agents) using aimgr CLI. Use when user asks to: (1) INSTALL - 'Install skill X', 'Add resource', 'Set up tools' (2) DISCOVER - 'What skills are available?', 'Recommend resources', 'Show me skills' (3) MANAGE - 'Update repository', 'Remove skill', 'Sync resources' (4) PACKAGES - 'Install package', 'Create package', 'Manage collections' (5) TROUBLESHOOT - 'Skills not working', 'aimgr issues', 'Fix installation'. NOT triggered automatically - only when user explicitly asks about skills/resources/aimgr."
---

# AI Resource Manager

Manage AI resources (skills, commands, agents) using the `aimgr` CLI. This skill helps you discover, install, and maintain resources across AI tools (Claude Code, OpenCode, GitHub Copilot) using symlink-based installation.

**Core Principle:** Resources are stored once in `~/.local/share/ai-config/repo/` and linked to projects, enabling zero-duplication sharing across tools.

---

## Quick Reference

```bash
# Discovery
aimgr repo list --format=json   # All available resources
aimgr list --format=json        # Installed in current project

# Installation
aimgr install skill/name        # Install specific resource
aimgr install "skill/*"         # Install using patterns
aimgr install package/name      # Install package (v1.12.0+)

# Packages (v1.12.0+)
aimgr repo create-package name --description="..." --resources="skill/a,command/b"
aimgr repo list package         # List packages
aimgr uninstall package/name    # Uninstall package

# Management
aimgr repo add ./path           # Add to repository
aimgr repo update               # Update from sources
aimgr uninstall skill/name      # Remove from project

# Configuration
aimgr config set install.targets claude

📚 **Need command syntax?** See [cli-installation.md](references/cli-installation.md) for `install` command details

---

## 1. Installation Assistant Workflow

Help users discover and install AI resources interactively.

### Step 1: Query Available Resources

```bash
# Get all resources in repository
aimgr repo list --format=json

# Get installed resources in current project
aimgr list --format=json
```

Parse the JSON output to present resources in a user-friendly format. Do NOT dump raw JSON.

### Step 2: Interactive Selection

Use the `question` tool to present options in a user-friendly format (NOT raw JSON).

### Step 3: Batch Installation

Install multiple resources efficiently:

```bash
# Install multiple resources at once
aimgr install skill/pdf-processing skill/react-testing command/test

# Install using patterns for bulk operations
aimgr install "skill/pdf*"       # All PDF-related skills
aimgr install "*test*"            # All test resources
```

### Step 4: Restart Reminder (CRITICAL)

**ALWAYS remind users to restart their AI tool after installation.** Skills load at startup - close and reopen Claude Code/OpenCode/VS Code to activate new resources.

### Step 5: Verification & Sync

```bash
# Verify installation
aimgr repo show skill/pdf-processing

# Enable sync (v1.12.0+)
aimgr config set repository.sync.enabled true
aimgr config set repository.sync.auto_update true
```

📚 **Need More Details?**
- **Installation commands**: See [cli-installation.md](references/cli-installation.md) for complete install/uninstall syntax
- **Discovery commands**: See [cli-discovery.md](references/cli-discovery.md) for listing and showing resources
- **Configuration**: See [cli-configuration.md](references/cli-configuration.md) for sync settings

---

## 2. Auto-Discovery Workflow

Analyze projects and recommend relevant skills based on detected technologies.

### Overview

Quick-scan (<2s) project indicator files (package.json, requirements.txt, Dockerfile) to recommend 3-8 relevant skills. Check only root-level files, no deep traversal.

### Step 1: Detect Project Technologies

Use `glob` to check for indicator files:

```bash
# Common indicators
glob "package.json"           # Node.js
glob "requirements.txt"       # Python  
glob "Dockerfile"             # Docker
glob ".github/workflows/*"    # GitHub Actions
glob "jest.config.*"          # Jest testing
```

**Pattern Matching:**
- **Node.js:** `package.json` → check dependencies for `react`, `vue`, `next`, `express`
- **Python:** `requirements.txt`, `pyproject.toml` → check for `django`, `flask`, `fastapi`
- **Docker:** `Dockerfile`, `docker-compose.yml`
- **CI/CD:** `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`

📚 **Complete pattern mappings:** See [references/project-patterns.md](references/project-patterns.md)

### Step 2: Query Available Skills

```bash
# Get skills from repository
aimgr repo list skill --format=json
```

Parse JSON and filter skills matching detected patterns.

### Step 3: Present Recommendations

Format as user-friendly summary prioritizing: (1) language/framework, (2) testing, (3) CI/CD, (4) infrastructure.

### Step 4: Interactive Installation

Use `question` tool for selection, then install via Workflow 1.


📚 **Need More Details?**
- **Project patterns**: See [project-patterns.md](references/project-patterns.md) for complete framework detection patterns
- **Discovery commands**: See [cli-discovery.md](references/cli-discovery.md) for `repo list` options
---

## 3. Repository Management Workflow

Manage the central skill repository by adding, updating, and syncing resources.

### Overview

Manage central repository at `~/.local/share/ai-config/repo/` - add from local/GitHub, update, and sync.

### Step 1: Add Resources

```bash
aimgr repo add ~/my-skills/pdf-processing    # local path
aimgr repo add ~/.opencode/                  # bulk import
aimgr repo add gh:owner/repo                 # GitHub
aimgr repo add gh:owner/repo@v1.0.0          # specific version
aimgr repo add owner/repo                    # shorthand
aimgr repo add gh:owner/repo --filter "skill/*"  # filter
aimgr repo add ./resources/ --dry-run            # preview
```

**Options:** `--force`, `--skip-existing`, `--dry-run`, `--filter=PATTERN`

### Step 2: Update Resources

```bash
aimgr repo update                          # all tracked sources
aimgr repo update skill/pdf-processing     # specific resource
aimgr repo update --dry-run                # preview
aimgr repo update --force                  # overwrite local changes
```

**Note:** Only GitHub/Git sources auto-update. Symlinks reflect updates automatically.

### Step 3: Configure Sync

```bash
# Enable repository sync (v1.12.0+)
aimgr config set repository.sync.enabled true
aimgr config set repository.sync.auto_update true
```

### Step 4: Remove & Verify

```bash
# Remove (permanently deletes, breaks symlinks)
aimgr repo remove skill/old-skill          # with confirmation
aimgr repo remove skill/old-skill --force  # skip confirmation

# Verify status
aimgr repo list --format=json              # all in repo
aimgr repo show skill/pdf-processing       # details
aimgr list --format=json                   # installed in project
```

### Example Workflows

```bash
# Team sharing
aimgr repo add gh:team/skills && aimgr install "skill/*"

# Custom skills
aimgr repo add ~/my-skills/custom && aimgr install skill/custom

# Maintenance
aimgr repo update --dry-run && aimgr repo update
```
📚 **Need More Details?**
- **Repository commands**: See [cli-repository.md](references/cli-repository.md) for complete `repo add`, `repo update`, `repo remove` syntax
- **Pattern syntax**: See [cli-advanced.md](references/cli-advanced.md#pattern-syntax) for filtering patterns

---

## 4. Package Management Workflow

Manage collections of related resources (skills, commands, agents) as a single unit using packages.

### Overview

Packages allow you to group multiple resources together and manage them collectively. This is ideal for:
- **Project Setup** - Bundle all tools needed for a new project type
- **Team Onboarding** - Share your team's standard resources in one install
- **Testing Workflows** - Group testing commands, skills, and agents
- **Domain Toolkits** - Organize resources by domain (web, data science, DevOps)

### Step 1: List Available Packages

```bash
# List all packages in repository
aimgr repo list --type=package --format=json

# List all package resources
aimgr repo list package
```

### Step 2: Install Packages

```bash
# Install a package (installs all its resources)
aimgr install package/web-dev-tools

# Install multiple packages
aimgr install package/web-dev-tools package/testing-suite

# View what resources are in a package before installing
aimgr repo show package/web-dev-tools
```

### Step 3: Create Custom Packages

```bash
# Create a package from existing repository resources
aimgr repo create-package my-toolkit \
  --description="My custom development toolkit" \
  --resources="skill/typescript-helper,command/build,command/test,agent/code-reviewer"

# Verify package creation
aimgr repo show package/my-toolkit
```

### Step 4: Uninstall Packages

```bash
# Uninstall all resources from a package
aimgr uninstall package/web-dev-tools

# Remove package from repository (keeps resources)
aimgr repo remove package/web-dev-tools

# Remove package AND all its resources from repository
aimgr repo remove package/web-dev-tools --with-resources
```

### Package Format

Packages are stored as JSON files with a flat structure:

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

### Use Cases

**Team Onboarding:**
```bash
# Team lead creates package
aimgr repo create-package team-standard \
  --description="Company standard development tools" \
  --resources="skill/coding-standards,command/test,command/lint,agent/reviewer"

# Team member installs in one command
aimgr install package/team-standard
```

**Project Templates:**
```bash
# Create packages for different project types
aimgr repo create-package react-project \
  --description="React project starter" \
  --resources="skill/react-helper,command/dev,command/build,agent/react-reviewer"

aimgr repo create-package python-api \
  --description="Python API project" \
📚 **Need More Details?**
- **Package commands**: See [cli-packages.md](references/cli-packages.md) for complete package management syntax
- **Package format**: See [cli-packages.md](references/cli-packages.md#package-format-specification) for JSON schema
```

📚 **Complete package commands:** [cli-packages.md](references/cli-packages.md)

---

## 5. Repository Resource Validation Workflow

Validate AI resources (skills, commands, agents) before adding them to the repository to ensure quality and correctness.

### Overview

When creating or modifying resources for the repository, validation catches common errors early:
- **Frontmatter errors** - Missing required fields, invalid YAML syntax
- **dt- specific issues** - Dynatrace skills have additional validation rules
- **Structural problems** - Missing files, incorrect directory structure
- **Quality issues** - Best practice violations, style inconsistencies

**Why Validation Matters:**
- Prevents broken resources from entering the repository
- Catches errors before users encounter them
- Ensures consistency across all resources
- Provides clear feedback on what needs fixing

### When to Validate

Run validation before:
- Adding new resources to repository: `aimgr repo add ./my-skill`
- Packaging skills for distribution: `python package_skill.py skill-name`
- Submitting pull requests with new/modified resources
- Updating existing resources: `aimgr repo update`

**Quick Validation Check:**
```bash
# Validate a skill before adding to repository
python src/skills/dt-dev-skill-creator/scripts/quick_validate.py ./my-skill

# Validate during packaging
python src/skills/dt-dev-skill-creator/scripts/package_skill.py ./my-skill
```

### Step 1: Create Your Resource

Follow the standard resource structure:

**Skills:**
```bash
# Initialize with skill-creator
python src/skills/dt-dev-skill-creator/scripts/init_skill.py my-skill

# Or create manually with SKILL.md containing:
# ---
# name: my-skill
# description: Brief description. Use when [trigger conditions].
# ---
```

**Commands/Agents:**
```markdown
---
description: Brief description of what this command/agent does
---

# Command/Agent content here
```

### Step 2: Run Validation

**Standard Resources:**
```bash
# Validate any skill
cd src/skills/dt-dev-skill-creator/scripts
python quick_validate.py /path/to/my-skill

# Expected output on success:
# ✅ Skill is valid!

# Expected output on failure:
# ❌ Validation failed: [specific error message]
```

**dt- Prefixed Resources:**

Resources starting with `dt-` trigger additional Dynatrace-specific validations. See [docs/DT_VALIDATION_DESIGN.md](../../../docs/DT_VALIDATION_DESIGN.md) for complete specification.

```bash
# Validate dt- skill (runs standard + dt- specific checks)
python quick_validate.py /path/to/dt-my-skill

# Expected output with warnings:
# ✅ Skill is valid!
# 
# dt- validation warnings:
#   Line 10: Consider adding '--plain' flag
#   Consider referencing 'dynatrace-control' skill for dtctl fundamentals
```

**Validation Scope:**
- ✅ Checks: Frontmatter syntax, required fields, basic structure
- ✅ dt- checks: dtctl command syntax, DQL query structure, naming conventions
- ❌ Does NOT check: Functional correctness, business logic, runtime behavior
- ❌ Does NOT verify: dtctl actually works, DQL returns correct results

**⚠️ Important:** Validation catches common errors but doesn't guarantee correctness. Always test resources functionally after validation passes.

### Step 3: Fix Issues

**Common Errors and Solutions:**

| Error | Cause | Fix |
|-------|-------|-----|
| `"Frontmatter not found"` | Missing `---` delimiters | Add YAML frontmatter at top of file |
| `"name field is required"` | Skills missing `name:` | Add `name: skill-name` to frontmatter |
| `"description field is required"` | Missing `description:` | Add `description:` to frontmatter |
| `"Invalid YAML syntax"` | YAML parsing error | Check indentation, quotes, colons |
| `"dt- skills must mention 'Dynatrace' or 'dtctl'"` | dt- skill lacks context | Add "Dynatrace" or "dtctl" to description |
| `"dtctl command requires resource type"` | Incomplete dtctl command | Use format: `dtctl verb resource [options]` |
| `"DQL query must start with 'fetch' or 'timeseries'"` | Invalid DQL syntax | Start queries with `fetch` or `timeseries` |

**Example Fix Workflow:**
```bash
# 1. Run validation
python quick_validate.py ./my-skill
# ❌ Validation failed: description field is required

# 2. Fix the issue
# Edit SKILL.md frontmatter to add:
# description: My skill description. Use when [trigger].

# 3. Re-validate
python quick_validate.py ./my-skill
# ✅ Skill is valid!

# 4. Add to repository
aimgr repo add ./my-skill
```

### Step 4: Add to Repository

After validation passes:

```bash
# Add validated resource to repository
aimgr repo add ./my-skill

# Verify it was added correctly
aimgr repo show skill/my-skill

# Install in current project to test
aimgr install skill/my-skill
```

**⚠️ Restart Required:** After installing, restart your AI tool (Claude Code/OpenCode/VS Code) to load the new resource.

### dt- Specific Rules

Resources with `dt-` prefix undergo additional validation. These rules ensure Dynatrace skills maintain quality standards:

**Naming Requirements:**
- ✅ Name must start with `dt-` followed by descriptive suffix (e.g., `dt-workflow-builder`)
- ❌ Name cannot be just `dt-` without description
- ✅ Description must mention "Dynatrace" or "dtctl" for context

**dtctl Command Validation:**
- ✅ Basic syntax: `dtctl <verb> <resource> [options]`
- ✅ Valid verbs: `get`, `describe`, `apply`, `delete`, `query`, `exec`, `history`, `restore`, `config`, `auth`
- ⚠️ Warnings: Missing `--plain` flag (recommended for AI consumption)
- ⚠️ Warnings: Destructive operations without context verification

**DQL Query Validation:**
- ✅ Must start with `fetch` or `timeseries`
- ✅ Cannot contain `dtctl` inside query string (common mistake)
- ⚠️ Warnings: Pipe syntax without spaces (use ` | ` not `|`)

**Quality Standards:**
- ⚠️ Should reference `dynatrace-control` skill for dtctl fundamentals
- ⚠️ Should verify context before destructive operations

**Example Valid dt- Skill:**

Frontmatter:
```yaml
---
name: dt-workflow-builder
description: Create and manage Dynatrace workflows using dtctl. Use when building automation workflows.
---
```

Content with valid commands:
```bash
# List workflows (valid)
dtctl get workflows --mine --plain

# Apply workflow (valid)
dtctl apply -f workflow.yaml

# Query logs (valid)
dtctl query "fetch logs | filter loglevel == 'ERROR' | limit 100" --plain
```

**dt- Validation Reference:** [docs/DT_VALIDATION_DESIGN.md](../../../docs/DT_VALIDATION_DESIGN.md)

### Validation Checklist

Before adding resources to repository:

- [ ] **Structure**: Resource has correct directory/file structure
- [ ] **Frontmatter**: Valid YAML with required fields (`name`, `description`)
- [ ] **Syntax**: No YAML parsing errors
- [ ] **Description**: Clear, concise, includes "Use when..." triggers
- [ ] **dt- Rules** (if applicable): Passes dt- specific validations
- [ ] **Manual Test**: Resource loads and functions correctly in AI tool
- [ ] **Documentation**: Usage examples are clear and accurate
- [ ] **Quality**: Follows best practices from skill-creator guidelines

**Run Validation:**
```bash
python src/skills/dt-dev-skill-creator/scripts/quick_validate.py ./my-resource
```

**Add to Repository:**
```bash
aimgr repo add ./my-resource
aimgr repo show [type]/my-resource  # Verify
```

📚 **Need More Details?**
- **dt- Validation Design**: [docs/DT_VALIDATION_DESIGN.md](../../../docs/DT_VALIDATION_DESIGN.md)
- **Skill Creator Guide**: [dt-dev-skill-creator/SKILL.md](../dt-dev-skill-creator/SKILL.md)
- **Contributing Guidelines**: [CONTRIBUTING.md](../../../CONTRIBUTING.md)

---

## 6. Troubleshooting Workflow

Diagnose and fix common aimgr and resource issues.

### Installation Check

```bash
# Verify aimgr
which aimgr

# If not found, detect platform
python3 scripts/detect_platform.py
```

Platform-specific installation:

**Linux (x86_64):**
```bash
curl -fsSL https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr-linux-amd64 -o ~/.local/bin/aimgr
chmod +x ~/.local/bin/aimgr
export PATH="$HOME/.local/bin:$PATH"
```

**Linux (ARM64):**
```bash
curl -fsSL https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr-linux-arm64 -o ~/.local/bin/aimgr
chmod +x ~/.local/bin/aimgr
export PATH="$HOME/.local/bin:$PATH"
```

**macOS (Intel):**
```bash
curl -fsSL https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr-darwin-amd64 -o /usr/local/bin/aimgr
chmod +x /usr/local/bin/aimgr
```

**macOS (Apple Silicon):**
```bash
curl -fsSL https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr-darwin-arm64 -o /usr/local/bin/aimgr
chmod +x /usr/local/bin/aimgr
```

**Windows (WSL x86_64):**
```bash
curl -fsSL https://github.com/hk9890/ai-config-manager/releases/latest/download/aimgr-windows-amd64.exe -o ~/.local/bin/aimgr
chmod +x ~/.local/bin/aimgr
export PATH="$HOME/.local/bin:$PATH"
```

```bash
# Verify installation
aimgr --version
aimgr config set install.targets claude
```

**⚠️ Restart terminal after first-time setup.**

Platform detection returns: `linux_amd64`, `linux_arm64`, `darwin_amd64`, `darwin_arm64`, `windows_amd64`, `windows_arm64`, or `unknown`.

### Common Issues

- **Skills not loading?** → Restart AI tool (Claude Code/OpenCode/VS Code)
- **Resource not found?** → Run `aimgr repo update`
- **Broken symlinks?** → Reinstall: `aimgr uninstall skill/name && aimgr install skill/name`
- **Wrong directory?** → Use `--target`: `aimgr install skill/name --target claude`
- **Permission denied?** → Fix: `chmod +x $(which aimgr)`
- **Config missing?** → Create: `aimgr config set install.targets claude`

### Diagnostics

```bash
aimgr --version                 # Check version
aimgr config                    # Check configuration
aimgr list --format=json        # List installed
aimgr repo list --format=json   # Check repository
ls -la .claude/skills/          # Verify symlinks
```

📚 **Need More Details?**
- **Complete troubleshooting**: See [troubleshooting.md](references/troubleshooting.md) for detailed solutions
- **Platform detection**: Use [scripts/detect_platform.py](scripts/detect_platform.py) for OS/architecture detection
- **Advanced formats**: See [cli-advanced.md](references/cli-advanced.md#resource-formats) for resource specifications

---

## 7. Safety Notes

### Uninstall Safety

**ALWAYS verify symlink (→ symbol) before removing:**

```bash
ls -la .claude/skills/
# pdf-processing -> /path/to/repo/skills/pdf-processing  ← SAFE (symlink)
# my-custom-skill/                                        ← DO NOT remove
```

**Rule:** Only uninstall resources with `→` symbol. Never remove manually-created resources.

### Restart Requirements

**Always remind users to restart AI tool after:**
- Installing/uninstalling resources
- Repository updates affecting installed resources  
- First-time aimgr configuration

**Template:** `⚠️ **Restart Required:** Close and reopen [Tool Name] to load changes.`

---

## Resource Formats

**Commands/Agents:** Single `.md` file with `description:` frontmatter
**Skills:** Directory with `SKILL.md` containing `name:` and `description:` frontmatter

📚 [cli-advanced.md](references/cli-advanced.md#resource-formats)

---

## Pattern Syntax

```bash
aimgr install "skill/*"              # All skills
aimgr install "*test*"               # All test resources
aimgr install "skill/pdf*"           # PDF-related skills
aimgr install "command/{build,test}" # Multiple patterns
```

📚 [cli-advanced.md](references/cli-advanced.md#pattern-syntax)

---

## Supported AI Tools

| Tool | Skills | Commands | Agents | Directory |
|------|--------|----------|--------|-----------|
| **Claude Code** | ✅ | ✅ | ✅ | `.claude/` |
| **OpenCode** | ✅ | ✅ | ✅ | `.opencode/` |
| **GitHub Copilot** | ✅ | ❌ | ❌ | `.github/skills/` |

---

## Additional Resources

- **CLI References:** See individual reference files for detailed documentation:
  - [cli-discovery.md](references/cli-discovery.md) - Discovery and listing commands
  - [cli-installation.md](references/cli-installation.md) - Installation and removal
  - [cli-repository.md](references/cli-repository.md) - Repository management
  - [cli-packages.md](references/cli-packages.md) - Package management
  - [cli-configuration.md](references/cli-configuration.md) - Configuration settings
  - [cli-advanced.md](references/cli-advanced.md) - Patterns, completion, formats
- **Project Patterns:** [references/project-patterns.md](references/project-patterns.md) - Tech stack → skill mappings
- **Troubleshooting:** [references/troubleshooting.md](references/troubleshooting.md) - Solutions to common issues
- **Platform Detection:** [scripts/detect_platform.py](scripts/detect_platform.py) - OS/architecture detection

**Repository:** https://github.com/hk9890/ai-config-manager
**Issues:** https://github.com/hk9890/ai-config-manager/issues

**Specifications:**
- Claude Code commands: https://code.claude.com/docs/en/slash-commands
- Agent Skills: https://agentskills.io/specification
- Claude Code agents: https://code.claude.com/docs/agents
- OpenCode agents: https://opencode.ai/docs/agents
