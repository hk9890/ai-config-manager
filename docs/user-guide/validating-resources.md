# Validating Resources for aimgr

**For skill, agent, command, and package developers**

This guide shows you how to validate that your AI resources are compatible with aimgr before publishing them. Whether you're building a single skill or a complete package, you can test everything locally to ensure it works correctly.

## Table of Contents

- [Quick Validation Workflow](#quick-validation-workflow)
- [Before You Start](#before-you-start)
- [Validating Skills](#validating-skills)
- [Validating Agents](#validating-agents)
- [Validating Commands](#validating-commands)
- [Validating Packages](#validating-packages)
- [Common Issues and Fixes](#common-issues-and-fixes)
- [CI/CD Integration](#cicd-integration)

---

## Quick Validation Workflow

The fastest way to check if your resource works with aimgr:

```bash
# 1. Validate the format (doesn't add to repo)
aimgr repo import ./my-resource --dry-run

# 2. If validation passes, test installation
aimgr repo import ./my-resource      # Add to repo
cd /tmp/test-project                  # Go to test directory
aimgr install skill/my-resource       # Try installing it

# 3. Verify files are in place
ls .claude/skills/my-resource         # Check installation
```

If all three steps succeed, your resource is compatible with aimgr!

---

## Before You Start

### Understanding aimgr's Validation

aimgr validates resources at import time by checking:

1. **File structure** - Correct directory layout and filenames
2. **Naming rules** - Names follow agentskills.io specification
3. **Frontmatter format** - Valid YAML syntax and required fields
4. **Resource references** - For packages, all referenced resources exist

### Required Tools

You only need aimgr installed. No other tools required for validation.

```bash
# Check aimgr is installed
aimgr --version

# If not installed, build from source
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
make install
```

---

## Validating Skills

Skills are directories containing a `SKILL.md` file with YAML frontmatter.

### Expected Structure

```
my-skill/
├── SKILL.md              # Required: metadata + documentation
├── scripts/              # Optional: helper scripts
├── references/           # Optional: additional docs
└── assets/              # Optional: images, diagrams
```

### Step 1: Check Your SKILL.md Format

Your `SKILL.md` should look like this:

```markdown
---
name: my-skill
description: Brief description of what this skill does (1-1024 chars)
license: MIT
version: "1.0.0"
author: Your Name
---

# My Skill

Detailed documentation goes here.

## Usage

Examples and instructions...
```

**Required fields:**
- `description` - Brief explanation (1-1024 characters)

**Important:**
- Directory name MUST match the `name` field in frontmatter
- Use lowercase, alphanumeric, and hyphens only (e.g., `pdf-processing`, not `PDF_Processing`)

### Step 2: Validate with Dry-Run

```bash
# Validate without adding to repository
aimgr repo import ./my-skill --dry-run

# Expected success output:
# ✓ skill/my-skill - Successfully validated

# JSON output for scripting
aimgr repo import ./my-skill --dry-run --format=json
```

**Exit codes:**
- `0` = Validation passed
- `1` = Validation failed

### Step 3: Test Installation

```bash
# Add to repository
aimgr repo import ./my-skill

# Create test project
mkdir -p /tmp/test-aimgr
cd /tmp/test-aimgr

# Install the skill
aimgr install skill/my-skill

# Verify installation
ls .claude/skills/my-skill/SKILL.md     # Should exist
cat .claude/skills/my-skill/SKILL.md    # Should match your original
```

### Step 4: Test with Multiple Tools

```bash
# Test with OpenCode
aimgr install skill/my-skill --tool=opencode
ls .opencode/skills/my-skill/SKILL.md

# Test with GitHub Copilot
aimgr install skill/my-skill --tool=copilot
ls .github/skills/my-skill/SKILL.md

# Test with Windsurf
aimgr install skill/my-skill --tool=windsurf
ls .windsurf/skills/my-skill/SKILL.md
```

### Common Skill Validation Errors

**Error: Directory must contain SKILL.md**
```
Error: skill 'my-skill' in /path/to/my-skill: directory must contain SKILL.md
```
**Fix:** Create `SKILL.md` file in the skill directory.

**Error: Name mismatch**
```
Error: skill name 'my-skill' must match directory name 'myskill'
```
**Fix:** Rename directory to match the name in frontmatter, or update frontmatter to match directory name.

**Error: Invalid name format**
```
Error: skill 'My_Skill' field 'name': name must be lowercase alphanumeric + hyphens
```
**Fix:** Use only lowercase letters, numbers, and hyphens. Examples: `my-skill`, `pdf-processor`, `test-helper`

**Error: Missing description**
```
Error: skill 'my-skill' field 'description': description cannot be empty
```
**Fix:** Add `description` field to frontmatter.

---

## Validating Agents

Agents are single `.md` files with YAML frontmatter defining AI agents.

### Expected Structure

Single file: `my-agent.md`

### Step 1: Check Your Agent Format

```markdown
---
description: Brief description of what this agent does
type: code-reviewer
instructions: Detailed instructions for the agent's behavior
capabilities:
  - static-analysis
  - security-scan
version: "1.0.0"
author: Your Name
license: MIT
---

# My Agent

Agent documentation and guidelines go here.
```

**Required fields:**
- `description` - Brief explanation

**Optional fields (OpenCode format):**
- `type` - Agent role/category
- `instructions` - Detailed behavior instructions
- `capabilities` - Array of capability strings

### Step 2: Validate

```bash
# Validate the agent file
aimgr repo import ./my-agent.md --dry-run

# Expected success output:
# ✓ agent/my-agent - Successfully validated
```

### Step 3: Test Installation

```bash
# Add to repository
aimgr repo import ./my-agent.md

# Test installation
cd /tmp/test-aimgr
aimgr install agent/my-agent

# Verify
ls .claude/agents/my-agent.md
```

### Common Agent Validation Errors

**Error: Must be a .md file**
```
Error: agent must be a .md file
```
**Fix:** Rename file with `.md` extension.

**Error: Invalid name**
```
Error: agent 'my_agent' field 'name': name must be lowercase alphanumeric + hyphens
```
**Fix:** Rename file to use hyphens instead of underscores (e.g., `my-agent.md`).

---

## Validating Commands

Commands are `.md` files in a `commands/` directory.

### Expected Structure

```
commands/
├── my-command.md
└── api/                    # Nested commands supported
    ├── deploy.md
    └── rollback.md
```

### Step 1: Check Your Command Format

```markdown
---
description: Brief description of what this command does
agent: optional-agent-name
model: gpt-4
---

# My Command

Command documentation and usage instructions.
```

**Required fields:**
- `description` - Brief explanation

**Important:**
- File MUST be in a directory named `commands/` (or `.claude/commands/`, `.opencode/commands/`)
- Nested structure is supported (e.g., `commands/api/deploy.md` becomes `command/api/deploy`)

### Step 2: Validate

```bash
# Validate command file
aimgr repo import ./commands/my-command.md --dry-run

# Validate entire commands directory
aimgr repo import ./commands --dry-run

# Expected success output:
# ✓ command/my-command - Successfully validated
# ✓ command/api/deploy - Successfully validated
```

### Step 3: Test Installation

```bash
# Add to repository
aimgr repo import ./commands

# Test installation
cd /tmp/test-aimgr
aimgr install command/my-command

# Verify
ls .claude/commands/my-command.md

# Test nested command
aimgr install command/api/deploy
ls .claude/commands/api/deploy.md
```

### Common Command Validation Errors

**Error: Must be in commands/ directory**
```
Error: command file must be in a 'commands/' directory
```
**Fix:** Move file into a `commands/` directory:
```bash
mkdir -p commands
mv my-command.md commands/
```

**Error: Must be .md file**
```
Error: command must be a .md file
```
**Fix:** Rename with `.md` extension.

---

## Validating Packages

Packages are JSON files that bundle multiple resources together.

### Expected Structure

Single file: `my-package.package.json`

### Step 1: Create Package JSON

```json
{
  "name": "my-package",
  "description": "My collection of AI resources",
  "resources": [
    "skill/my-skill",
    "command/my-command",
    "agent/my-agent"
  ]
}
```

**Required fields:**
- `name` - Package name (must follow naming rules)
- `description` - Package description
- `resources` - Array of resource references in `type/name` format

**Resource reference format:**
- `skill/name` - Reference to a skill
- `command/name` - Reference to a command (including nested: `command/api/deploy`)
- `agent/name` - Reference to an agent

### Step 2: Ensure Referenced Resources Exist

**Important:** All resources referenced in the package MUST exist in the repository before you can add the package.

```bash
# Add individual resources first
aimgr repo import ./my-skill
aimgr repo import ./commands/my-command.md
aimgr repo import ./my-agent.md

# Verify they're in the repository
aimgr repo list skill/my-skill
aimgr repo list command/my-command
aimgr repo list agent/my-agent
```

### Step 3: Validate Package

```bash
# Validate package file
aimgr repo import ./my-package.package.json --dry-run

# Expected success output:
# ✓ package/my-package - Successfully validated
```

### Step 4: Verify Package References

After adding the package, verify all references are valid:

```bash
# Add package to repository
aimgr repo import ./my-package.package.json

# Verify package integrity
aimgr repo verify package/my-package

# Expected output:
# ✓ package/my-package - All references valid
```

### Step 5: Test Package Installation

```bash
# Install the entire package
cd /tmp/test-aimgr
aimgr install package/my-package

# This should install ALL resources in the package
ls .claude/skills/my-skill/
ls .claude/commands/my-command.md
ls .claude/agents/my-agent.md
```

### Common Package Validation Errors

**Error: Invalid resource reference format**
```
Error: invalid resource format: "my-skill" (expected type/name)
```
**Fix:** Use `type/name` format:
```json
{
  "resources": [
    "skill/my-skill",      // ✓ Correct
    "my-skill"              // ✗ Wrong
  ]
}
```

**Error: Referenced resource doesn't exist**
```
Error: package references non-existent resource 'skill/missing-skill'
```
**Fix:** Add the resource to the repository first:
```bash
aimgr repo import ./missing-skill
```

**Error: Invalid package name**
```
Error: invalid package name: "My_Package"
```
**Fix:** Use lowercase, alphanumeric, and hyphens only: `my-package`

---

## Common Issues and Fixes

### Issue: YAML Parsing Errors

**Error:**
```
Error: yaml: mapping values are not allowed in this context
```

**Cause:** Special characters in description (especially colons)

**Fix:** Quote the description field:
```yaml
---
description: "My tool: a helper script"  # ✓ Quoted
---
```

### Issue: Name Validation Fails

**Error:**
```
Error: name cannot contain consecutive hyphens
```

**Fix:** Remove double hyphens:
```yaml
---
name: my-skill      # ✓ Correct
name: my--skill     # ✗ Wrong
---
```

### Issue: Description Too Long

**Error:**
```
Error: skill description too long (1500 chars, max 1024)
```

**Fix:** Shorten description to 1024 characters or less. Put detailed info in the markdown body instead.

### Issue: Import Succeeds but Verify Fails

**Symptom:** `repo import` works but `repo verify` reports errors

**Cause:** Package references resources that don't exist

**Fix:**
```bash
# Check what's missing
aimgr repo verify package/my-package

# Add missing resources
aimgr repo import ./missing-resource

# Verify again
aimgr repo verify package/my-package
```

---

## CI/CD Integration

Automate validation in your CI/CD pipeline.

### GitHub Actions Example

Create `.github/workflows/validate-resources.yml`:

```yaml
name: Validate AI Resources

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Install aimgr
        run: |
          # Install Go
          sudo snap install go --classic
          
          # Install aimgr
          go install github.com/hk9890/ai-config-manager@latest
          
          # Add to PATH
          echo "$HOME/go/bin" >> $GITHUB_PATH
      
      - name: Validate all resources
        run: |
          # Validate with dry-run (doesn't modify anything)
          if aimgr repo import . --dry-run --format=json > validation.json; then
            echo "✅ All resources valid"
            cat validation.json | jq '.'
            exit 0
          else
            echo "❌ Validation failed"
            cat validation.json | jq '.'
            exit 1
          fi
```

### GitLab CI Example

Create `.gitlab-ci.yml`:

```yaml
validate-resources:
  image: golang:1.22
  stage: test
  script:
    - go install github.com/hk9890/ai-config-manager@latest
    - export PATH="$HOME/go/bin:$PATH"
    - aimgr repo import . --dry-run --format=json
  artifacts:
    reports:
      junit: validation.json
```

### Pre-commit Hook

Create `.git/hooks/pre-commit` to validate before committing:

```bash
#!/bin/bash

echo "Validating AI resources..."

# Find all SKILL.md files
SKILLS=$(find . -name "SKILL.md" -not -path "./.git/*" | xargs dirname)

# Find all agent and command files
AGENTS=$(find . -path "*/agents/*.md" -not -path "./.git/*")
COMMANDS=$(find . -path "*/commands/*.md" -not -path "./.git/*")

# Validate each
for skill in $SKILLS; do
  if ! aimgr repo import "$skill" --dry-run > /dev/null 2>&1; then
    echo "❌ Validation failed: $skill"
    aimgr repo import "$skill" --dry-run
    exit 1
  fi
done

for agent in $AGENTS; do
  if ! aimgr repo import "$agent" --dry-run > /dev/null 2>&1; then
    echo "❌ Validation failed: $agent"
    aimgr repo import "$agent" --dry-run
    exit 1
  fi
done

for command in $COMMANDS; do
  if ! aimgr repo import "$command" --dry-run > /dev/null 2>&1; then
    echo "❌ Validation failed: $command"
    aimgr repo import "$command" --dry-run
    exit 1
  fi
done

echo "✅ All resources validated successfully"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Complete Example: Validating a New Skill

Let's walk through validating a complete skill from scratch.

### 1. Create the skill structure

```bash
mkdir pdf-processing
cd pdf-processing

cat > SKILL.md << 'EOF'
---
name: pdf-processing
description: Skill for processing and extracting information from PDF files
license: MIT
version: "1.0.0"
author: John Doe
---

# PDF Processing Skill

This skill helps AI assistants process PDF files.

## Capabilities

- Extract text from PDFs
- Parse PDF metadata
- Convert PDF to other formats

## Usage

Ask the AI assistant to help with PDF-related tasks.
EOF

# Optional: Add helper scripts
mkdir scripts
cat > scripts/extract-text.sh << 'EOF'
#!/bin/bash
# Extract text from PDF using pdftotext
pdftotext "$1" -
EOF
chmod +x scripts/extract-text.sh
```

### 2. Validate the skill

```bash
cd ..  # Go to parent directory
aimgr repo import ./pdf-processing --dry-run

# Expected output:
# ✓ skill/pdf-processing - Successfully validated
```

### 3. Add to repository

```bash
aimgr repo import ./pdf-processing

# Verify it's in the repo
aimgr repo list skill/pdf-processing
```

### 4. Test installation

```bash
# Create test project
mkdir -p /tmp/test-pdf
cd /tmp/test-pdf

# Install the skill
aimgr install skill/pdf-processing

# Check installation
ls -la .claude/skills/pdf-processing/
cat .claude/skills/pdf-processing/SKILL.md

# Test with other tools
aimgr install skill/pdf-processing --tool=opencode
ls -la .opencode/skills/pdf-processing/
```

### 5. Create a package with the skill

```bash
cd /tmp

cat > pdf-toolkit.package.json << 'EOF'
{
  "name": "pdf-toolkit",
  "description": "Complete PDF processing toolkit",
  "resources": [
    "skill/pdf-processing"
  ]
}
EOF

# Validate package
aimgr repo import ./pdf-toolkit.package.json --dry-run

# Add package
aimgr repo import ./pdf-toolkit.package.json

# Verify package
aimgr repo verify package/pdf-toolkit
```

### 6. Test package installation

```bash
cd /tmp/test-pdf

# Install entire package
aimgr install package/pdf-toolkit

# Should install the skill
ls .claude/skills/pdf-processing/
```

✅ Success! Your skill is now validated, in the repository, and installable via package.

---

## Quick Reference

### Validation Commands

```bash
# Validate without adding
aimgr repo import <path> --dry-run

# Validate and add
aimgr repo import <path>

# Verify repository integrity
aimgr repo verify

# Verify specific resource
aimgr repo verify <pattern>

# JSON output for scripting
aimgr repo import <path> --dry-run --format=json
```

### Resource Formats

- **Skill:** Directory with `SKILL.md`
- **Agent:** Single `.md` file
- **Command:** `.md` file in `commands/` directory
- **Package:** `.package.json` file

### Naming Rules

- Lowercase alphanumeric + hyphens only
- Cannot start or end with hyphen
- No consecutive hyphens
- Max 64 characters per segment
- Examples: `my-skill`, `pdf-processor`, `api-helper`

### Exit Codes

- `0` - Validation passed
- `1` - Validation failed

---

## Need Help?

- **Full documentation:** [Resource Formats](resource-formats.md)
- **Developer guide:** [Developer Guide](developer-guide.md)
- **Report issues:** [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Ask questions:** [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)
