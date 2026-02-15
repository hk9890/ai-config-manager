# Resource Developer Guide

This guide is for developers creating skills, agents, commands, and packages for use with aimgr. It covers validation, testing, and best practices for ensuring your resources work correctly.

## Table of Contents

- [Quick Start](#quick-start)
- [Validating Your Resources](#validating-your-resources)
- [Common Validation Errors](#common-validation-errors)
- [Testing Locally](#testing-locally)
- [Best Practices](#best-practices)
- [Publishing Resources](#publishing-resources)

---

## Quick Start

Before adding resources to aimgr, you should:

1. **Follow the correct format** (see [Resource Formats](resource-formats.md))
2. **Validate locally** using the methods below
3. **Test installation** in a test project
4. **Add to repository** once validated

---

## Validating Your Resources

aimgr provides several ways to validate resources before adding them to the repository.

### Method 1: Dry-Run Import (Recommended)

Test if aimgr can successfully parse your resources without actually adding them:

```bash
# Validate a single skill
aimgr repo add ./my-skill --dry-run

# Validate a command
aimgr repo add ./commands/my-command.md --dry-run

# Validate an agent
aimgr repo add ./agents/my-agent.md --dry-run

# Validate a package
aimgr repo add ./packages/my-package.package.json --dry-run

# Validate an entire directory
aimgr repo add ./my-resources --dry-run
```

**Output formats:**

```bash
# Human-readable table (default)
aimgr repo add ./my-skill --dry-run

# JSON for scripting/CI
aimgr repo add ./my-skill --dry-run --format=json

# YAML for structured review
aimgr repo add ./my-skill --dry-run --format=yaml
```

**Exit codes:**
- 0: All resources are valid
- 1: Validation errors found

### Method 2: Add and Verify

Add resources to the repository and then run verification:

```bash
# Add resources
aimgr repo add ./my-resources

# Verify integrity
aimgr repo verify

# Verify specific pattern
aimgr repo verify skill/my-*
```

### Method 3: Repository Verification

After adding resources, verify repository consistency:

```bash
# Check all resources
aimgr repo verify

# Check and auto-fix issues
aimgr repo verify --fix

# Check specific type
aimgr repo verify skill/*
aimgr repo verify command/*
aimgr repo verify agent/*
aimgr repo verify package/*

# JSON output for CI/CD
aimgr repo verify --format=json
```

**What repo verify checks:**
- Resources without metadata (warning)
- Orphaned metadata files (error)
- Non-existent source paths (warning)
- Type mismatches (error)
- Broken package references (error)

---

## Common Validation Errors

aimgr provides detailed error messages with actionable suggestions. Here are the most common issues:

### Name Validation Errors

**Invalid characters:**
```
Error: name must be lowercase alphanumeric + hyphens, cannot start/end with hyphen
Suggestion: Use lowercase alphanumeric characters and hyphens only
```

Valid names:
- my-skill
- test-command
- pdf-processing
- code-reviewer

Invalid names:
- My-Skill (uppercase)
- -my-skill (starts with hyphen)
- my-skill- (ends with hyphen)
- my--skill (consecutive hyphens)
- my_skill (underscores not allowed)

**Name too long:**
```
Error: name too long (78 chars, max 64)
Suggestion: Shorten the name to 64 characters or less
```

### Frontmatter Errors

**Missing frontmatter:**
```
Error: no frontmatter found
Suggestion: Add YAML frontmatter at the top of the file
```

**Missing closing delimiter:**
```
Error: no closing frontmatter delimiter
Suggestion: Add closing delimiter after the frontmatter section
```

**YAML syntax errors:**
```
Error: mapping values are not allowed in this context
Suggestion: Quote the description field if it contains colons or special characters
```

### Description Errors

**Missing description:**
```
Error: description is required
Suggestion: Add a description field to the frontmatter with a brief explanation
```

**Description too long:**
```
Error: skill description too long (1500 chars, max 1024)
Suggestion: Shorten the description to 1024 characters or less
```

### Skill-Specific Errors

**Missing SKILL.md:**
```
Error: directory must contain SKILL.md
Suggestion: Create a SKILL.md file in the skill directory with proper frontmatter
```

**Name mismatch:**
```
Error: skill name must match directory name
Suggestion: Rename the directory or file to match the name field in frontmatter
```

**Wrong structure:**
```
Error: skill must be a directory
Suggestion: Skills must be directories containing SKILL.md, not single files
```

### Command-Specific Errors

**Not in commands/ directory:**
```
Error: command file must be in a commands directory
Suggestion: Move the file to a commands directory
```

**Wrong file extension:**
```
Error: command must be a .md file
Suggestion: Rename the file with a .md extension
```

### Package-Specific Errors

**Missing resource reference:**
```
Error: package references non-existent resource
Suggestion: Ensure all resources exist in the repository before referencing them
```

**Invalid resource format:**
```
Error: invalid resource format (expected type/name)
Suggestion: Use format type/name (e.g., skill/pdf-processing, command/test)
```

---

## Testing Locally

### Testing Resource Installation

Create a test project to verify resources install correctly:

```bash
# 1. Add resource to repository
aimgr repo add ./my-skill

# 2. Create test project
mkdir -p /tmp/test-project
cd /tmp/test-project

# 3. Install and test
aimgr install skill/my-skill
ls .claude/skills/my-skill  # Verify files exist

# 4. Test with different tools
aimgr install skill/my-skill --tool=opencode
ls .opencode/skills/my-skill

aimgr install skill/my-skill --tool=copilot
ls .github/skills/my-skill

# 5. Clean up
cd -
rm -rf /tmp/test-project
```

### Testing Nested Commands

For commands with nested structure:

```bash
# Directory structure:
# commands/
#   api/
#     deploy.md
#     rollback.md

# Import with nested structure preserved
aimgr repo add ./commands

# Verify nested names
aimgr repo list command/api/*

# Install and test
cd /tmp/test-project
aimgr install command/api/deploy
ls .claude/commands/api/deploy.md
```

### Testing Packages

For packages that bundle multiple resources:

```bash
# 1. Create package JSON
cat > my-package.package.json << 'EOF'
{
  "name": "my-package",
  "description": "My resource bundle",
  "resources": [
    "skill/my-skill",
    "command/my-command"
  ]
}
EOF

# 2. Ensure referenced resources exist
aimgr repo add ./my-skill
aimgr repo add ./commands/my-command.md

# 3. Import package
aimgr repo add my-package.package.json

# 4. Verify package references
aimgr repo verify package/my-package

# 5. Test package installation
cd /tmp/test-project
aimgr install package/my-package
# Should install both skill and command
```

### Testing with ai.package.yaml

Test automatic manifest updates:

```bash
cd /tmp/test-project

# Initialize manifest
aimgr init

# Install resource (should auto-update manifest)
aimgr install skill/my-skill

# Verify manifest
cat ai.package.yaml

# Test reinstall from manifest
rm -rf .claude .opencode .github .windsurf
aimgr install  # Reads from ai.package.yaml
```

---

## Best Practices

### Skill Development

DO:
- Use clear, descriptive names (lowercase-with-hyphens)
- Provide detailed descriptions (1-1024 characters)
- Include examples in the SKILL.md content
- Use optional directories (scripts/, references/, assets/) when needed
- Test with multiple tools (Claude, OpenCode, Copilot, Windsurf)
- Follow agentskills.io specification

DONT:
- Use uppercase or special characters in names
- Create single-file skills (must be directories with SKILL.md)
- Make directory names different from the skill name in frontmatter
- Exceed 1024 characters in descriptions
- Forget to include required frontmatter fields

**Example structure:**
```
my-skill/
├── SKILL.md              # Required: metadata + docs
├── scripts/              # Optional: executable scripts
│   └── helper.sh
├── references/           # Optional: additional docs
│   └── api-docs.md
└── assets/              # Optional: images, diagrams
    └── diagram.png
```

### Command Development

DO:
- Place commands in a commands directory
- Use .md extension
- Include description in frontmatter
- Support nested commands for organization
- Add optional fields (agent, model, allowed-tools) when needed

DONT:
- Place command files outside commands directories
- Use non-.md extensions
- Forget the description field
- Mix commands with other resource types in the same directory

### Agent Development

DO:
- Use .md extension
- Include description in frontmatter
- Add type/instructions for OpenCode format
- Include capabilities array when applicable
- Provide clear role definitions

DONT:
- Confuse agent and command formats
- Forget required description field
- Use ambiguous agent types

### Package Development

DO:
- Group related resources logically
- Use clear package names and descriptions
- Verify all referenced resources exist before creating package
- Use type/name format for resource references
- Test package installation

DONT:
- Reference non-existent resources
- Use invalid resource reference formats
- Create circular dependencies
- Bundle unrelated resources

---

## Publishing Resources

### Pre-Publishing Checklist

Before sharing your resources publicly:

- Validation passes
- Repository verification clean
- Tested installation in a test project
- Documentation complete
- Naming conventions followed
- License specified
- Version tagged

### Publishing to GitHub

Resources can be shared via GitHub repositories. Users can import with:

```bash
aimgr repo add gh:username/my-ai-resources
```

### Validation in CI/CD

Automate validation in your CI pipeline using dry-run mode and JSON output.

---

## Getting Help

- Documentation: Resource Formats guide
- Examples: See examples directory in the repository
- Issues: GitHub Issues
- Discussions: GitHub Discussions

---

## Quick Reference

**Validate before adding:**
```bash
aimgr repo add ./my-resource --dry-run
```

**Add to repository:**
```bash
aimgr repo add ./my-resource
```

**Verify repository:**
```bash
aimgr repo verify
aimgr repo verify --fix  # Auto-fix issues
```

**Test installation:**
```bash
cd test-project
aimgr install skill/my-skill
```

**Check resource details:**
```bash
aimgr repo describe skill/my-skill
```

**List resources:**
```bash
aimgr repo list skill/*
```
