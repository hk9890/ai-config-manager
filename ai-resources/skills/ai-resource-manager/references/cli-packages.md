# aimgr CLI Reference: Package Management

Commands for managing packages (collections of related resources).

## Table of Contents

- [repo create-package](#repo-create-package) - Create a new package
- [install (packages)](#install-packages) - Install a package
- [uninstall (packages)](#uninstall-packages) - Uninstall a package
- [repo remove (packages)](#repo-remove-packages) - Remove package from repository
- [repo list (packages)](#repo-list-packages) - List all packages
- [repo show (packages)](#repo-show-packages) - Show package details
- [Package Format](#package-format-specification) - Package file format

---

## Package Management

### repo create-package

Create a package that groups multiple resources together for collective management.

**Syntax:**
```bash
aimgr repo create-package NAME --description=DESC --resources=RESOURCES [OPTIONS]
```

**Arguments:**
- `NAME` - Package name (must follow [resource naming requirements](#resource-naming-requirements))

**Options:**
- `--description=DESC` - Package description (required)
- `--resources=RESOURCES` - Comma-separated list of resources in `type/name` format (required)
  - Example: `skill/foo,command/bar,agent/baz`
- `--force` - Overwrite existing package with same name

**Examples:**

**Basic Package Creation:**
```bash
# Create web development toolkit
aimgr repo create-package web-tools \
  --description="Web development toolkit" \
  --resources="command/build,skill/typescript-helper,agent/code-reviewer"

# Create testing suite
aimgr repo create-package testing-suite \
  --description="Complete testing workflow" \
  --resources="command/test,command/coverage,skill/jest-helper,agent/test-reviewer"
```

**Domain-Specific Packages:**
```bash
# Python data science package
aimgr repo create-package data-science \
  --description="Python data science tools" \
  --resources="skill/pandas-helper,skill/numpy-helper,command/jupyter,agent/data-reviewer"

# DevOps package
aimgr repo create-package devops-tools \
  --description="DevOps and infrastructure tools" \
  --resources="skill/docker-helper,skill/k8s-helper,command/deploy,agent/infra-reviewer"
```

**Team Packages:**
```bash
# Company standard tools
aimgr repo create-package company-standard \
  --description="Company development standards" \
  --resources="skill/coding-standards,command/lint,command/format,agent/style-reviewer"
```

**Overwrite Existing:**
```bash
# Update package definition
aimgr repo create-package web-tools \
  --description="Updated web toolkit" \
  --resources="skill/react-helper,skill/vue-helper,command/build,command/dev" \
  --force
```

**Validation:**
- All resources must exist in repository before package creation
- Resource names must be in `type/name` format (e.g., `skill/foo`, `command/bar`)
- Package names must follow [resource naming requirements](#resource-naming-requirements)
- Duplicate resources in list are rejected

**Output:**
```
Creating package...
✓ Validated 4 resources
✓ Created package/web-tools

Package: web-tools
Description: Web development toolkit
Resources:
  - command/build
  - skill/typescript-helper
  - agent/code-reviewer

Successfully created package.
```

---

### install (packages)

Install all resources from a package with a single command.

**Syntax:**
```bash
aimgr install package/NAME [OPTIONS]
```

**Arguments:**
- `package/NAME` - Package to install (e.g., `package/web-tools`)

**Options:**
- `--target=TOOL` - Override target tool(s): `claude`, `opencode`, `copilot`
- `--force` - Force reinstall all package resources
- `--project-path=PATH` - Specify project directory

**Examples:**

**Basic Package Installation:**
```bash
# Install web development package
aimgr install package/web-tools

# Install multiple packages at once
aimgr install package/web-tools package/testing-suite

# Preview package contents before installing
aimgr repo show package/web-tools
```

**Target Override:**
```bash
# Install package to specific tool
aimgr install package/web-tools --target claude

# Install to multiple tools
aimgr install package/web-tools --target claude,opencode
```

**Force Reinstall:**
```bash
# Force reinstall all package resources
aimgr install package/web-tools --force
```

**Behavior:**
- Installs ALL resources listed in the package
- Skips resources that are already installed (unless `--force` used)
- Creates symlinks to repository resources
- Follows same installation logic as individual resource installs

**Output:**
```
Installing package: web-tools
Resources to install: 4

Installing resources to project...
✓ command/build -> .claude/commands/build.md
✓ command/build -> .opencode/commands/build.md
✓ skill/typescript-helper -> .claude/skills/typescript-helper
✓ skill/typescript-helper -> .opencode/skills/typescript-helper
✓ agent/code-reviewer -> .claude/agents/code-reviewer.md
✓ agent/code-reviewer -> .opencode/agents/code-reviewer.md

Successfully installed package 'web-tools' (4 resources to 2 tools).

⚠️  Restart Required: Close and reopen your AI tool to load changes.
```

---

### uninstall (packages)

Remove all resources from a package from the current project.

**Syntax:**
```bash
aimgr uninstall package/NAME [OPTIONS]
```

**Arguments:**
- `package/NAME` - Package to uninstall (e.g., `package/web-tools`)

**Options:**
- `--project-path=PATH` - Specify project directory

**Examples:**

**Basic Package Uninstallation:**
```bash
# Uninstall web development package
aimgr uninstall package/web-tools

# Uninstall from specific project
aimgr uninstall package/web-tools --project-path ~/projects/my-app
```

**Behavior:**
- Uninstalls ALL resources listed in the package
- Only removes symlinked resources (safety check)
- Removes from ALL detected tool directories
- Original resources in repository remain intact
- Package definition remains in repository

**Output:**
```
Uninstalling package: web-tools
Resources to uninstall: 4

Uninstalling resources from project...
✓ Removed .claude/commands/build.md
✓ Removed .opencode/commands/build.md
✓ Removed .claude/skills/typescript-helper
✓ Removed .opencode/skills/typescript-helper
✓ Removed .claude/agents/code-reviewer.md
✓ Removed .opencode/agents/code-reviewer.md

Successfully uninstalled package 'web-tools' (4 resources from 2 tools).
```

---

### repo remove (packages)

Remove a package from the repository. Optionally remove all package resources as well.

**Syntax:**
```bash
aimgr repo remove package/NAME [OPTIONS]
```

**Arguments:**
- `package/NAME` - Package to remove (e.g., `package/web-tools`)

**Options:**
- `--force` - Skip confirmation prompt
- `--with-resources` - Also remove all resources listed in the package

**Examples:**

**Remove Package Only:**
```bash
# Remove package definition (keeps resources)
aimgr repo remove package/web-tools

# Skip confirmation
aimgr repo remove package/web-tools --force
```

**Remove Package and Resources:**
```bash
# Remove package AND all its resources
aimgr repo remove package/web-tools --with-resources

# With force flag
aimgr repo remove package/web-tools --with-resources --force
```

**Confirmation Prompt (without --force):**
```
⚠️  WARNING: This will permanently delete the package from the repository.

Package: web-tools
Description: Web development toolkit
Resources (4):
  - command/build
  - skill/typescript-helper
  - agent/code-reviewer
  - command/dev

Are you sure you want to remove this package? [y/N]: 
```

**With --with-resources Flag:**
```
⚠️  WARNING: This will permanently delete the package AND all its resources.

Package: web-tools
Resources to delete (4):
  - command/build
  - skill/typescript-helper
  - agent/code-reviewer
  - command/dev

⚠️  This will break symlinks in ALL projects using these resources!

Are you sure you want to remove package and resources? [y/N]: 
```

**Output (Package Only):**
```
Removing package from repository...
✓ Deleted package definition

Successfully removed package/web-tools from repository.
Note: Resources remain in repository.
```

**Output (Package + Resources):**
```
Removing package and resources from repository...
✓ Deleted package definition
✓ Deleted command/build
✓ Deleted skill/typescript-helper
✓ Deleted agent/code-reviewer
✓ Deleted command/dev

Successfully removed package/web-tools and 4 resources from repository.
⚠️  Broken symlinks may exist in projects. Run 'aimgr list' to check.
```

---

### repo list (packages)

List all packages in the repository.

**Syntax:**
```bash
aimgr repo list package [OPTIONS]
aimgr repo list --type=package [OPTIONS]
```

**Options:**
- `--format=FORMAT` - Output format: `text` (default), `json`, or `yaml`

**Examples:**

**List Packages:**
```bash
# List all packages
aimgr repo list package

# List packages as JSON
aimgr repo list package --format=json

# List packages as YAML
aimgr repo list --type=package --format=yaml
```

**Output (Text):**
```
Packages in repository:

  web-tools
    Web development toolkit
    Resources: 4 (command/build, skill/typescript-helper, agent/code-reviewer, command/dev)

  testing-suite
    Complete testing workflow
    Resources: 3 (command/test, command/coverage, skill/jest-helper)

  data-science
    Python data science tools
    Resources: 4 (skill/pandas-helper, skill/numpy-helper, command/jupyter, agent/data-reviewer)

Total: 3 packages
```

**Output (JSON):**
```json
{
  "packages": [
    {
      "name": "web-tools",
      "description": "Web development toolkit",
      "resources": [
        "command/build",
        "skill/typescript-helper",
        "agent/code-reviewer",
        "command/dev"
      ],
      "resource_count": 4,
      "path": "/home/user/.local/share/ai-config/repo/packages/web-tools.json"
    },
    {
      "name": "testing-suite",
      "description": "Complete testing workflow",
      "resources": [
        "command/test",
        "command/coverage",
        "skill/jest-helper"
      ],
      "resource_count": 3,
      "path": "/home/user/.local/share/ai-config/repo/packages/testing-suite.json"
    }
  ]
}
```

---

### repo show (packages)

Show detailed information about a package.

**Syntax:**
```bash
aimgr repo show package/NAME
```

**Arguments:**
- `package/NAME` - Package to show (e.g., `package/web-tools`)

**Examples:**

```bash
# Show package details
aimgr repo show package/web-tools

# Show testing suite package
aimgr repo show package/testing-suite
```

**Output:**
```
Name: web-tools
Type: package
Description: Web development toolkit
Path: /home/user/.local/share/ai-config/repo/packages/web-tools.json

Resources (4):
  Commands:
    - build
  
  Skills:
    - typescript-helper
  
  Agents:
    - code-reviewer

Usage:
  Install package:    aimgr install package/web-tools
  Uninstall package:  aimgr uninstall package/web-tools
```

---

### Package Format Specification

Packages are stored as JSON files in `~/.local/share/ai-config/repo/packages/`.

**File Format:**
```json
{
  "name": "package-name",
  "description": "Package description",
  "resources": [
    "type/resource-name",
    "type/resource-name"
  ]
}
```

**Schema:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Package name (must match filename without .json) |
| `description` | string | Yes | Brief description of package purpose |
| `resources` | array | Yes | List of resources in `type/name` format |

**Validation Rules:**
- Package name must follow [resource naming requirements](#resource-naming-requirements)
- Filename must be `name.json` (e.g., `web-tools.json` for package `web-tools`)
- All resources must exist in repository
- Resources must use `type/name` format (e.g., `skill/foo`, `command/bar`)
- No duplicate resources allowed in list
- Minimum 1 resource required

**Example Package:**
```json
{
  "name": "web-dev-tools",
  "description": "Complete web development toolkit with build, test, and review tools",
  "resources": [
    "command/build",
    "command/dev",
    "command/test",
    "skill/typescript-helper",
    "skill/react-helper",
    "agent/code-reviewer"
  ]
}
```

**Storage Location:**
- Repository: `~/.local/share/ai-config/repo/packages/NAME.json`
- Packages are NOT symlinked to projects (only their resources are)

---
