# Validating AI Resources for Developers

**For developers creating skills, agents, commands, and packages**

This guide shows how to validate that your AI resources work with aimgr before publishing them.

---

## Quick Validation Workflow

The fastest way to check if your resource is compatible with aimgr:

```bash
# Step 1: Validate format (doesn't modify repository)
aimgr repo import ./my-skill --dry-run

# Step 2: If validation passes, add to repository
aimgr repo import ./my-skill

# Step 3: Test installation in a project
cd /tmp/test-project
aimgr install skill/my-skill
ls .claude/skills/my-skill/  # Verify files exist
```

If all three steps succeed, your resource is compatible with aimgr!

---

## Validating Skills

Skills are directories containing a `SKILL.md` file.

### Expected Structure

```
my-skill/
├── SKILL.md              # Required: metadata + documentation
├── scripts/              # Optional: helper scripts
├── references/           # Optional: additional docs
└── assets/              # Optional: images, files
```

### Validation Steps

**1. Check SKILL.md format:**

```yaml
---
name: my-skill
description: Brief description (1-1024 chars, required)
license: MIT
version: "1.0.0"
---

# My Skill

Documentation goes here...
```

**Required fields:**
- `description` - Brief explanation (1-1024 characters)

**Important:**
- Directory name MUST match the `name` field
- Use lowercase, alphanumeric, and hyphens only (e.g., `pdf-processing`)

**2. Validate with dry-run:**

```bash
aimgr repo import ./my-skill --dry-run

# Success output:
# ✓ skill/my-skill - Successfully validated

# For scripting/CI:
aimgr repo import ./my-skill --dry-run --format=json
```

**Exit codes:**
- `0` = Valid
- `1` = Validation failed

**3. Test installation:**

```bash
# Add to repository
aimgr repo import ./my-skill

# Test in project
mkdir /tmp/test && cd /tmp/test
aimgr install skill/my-skill

# Verify
ls .claude/skills/my-skill/SKILL.md
```

### Common Skill Errors

**Name mismatch:**
```
Error: skill name 'my-skill' must match directory name 'myskill'
```
**Fix:** Rename directory or update frontmatter to match.

**Invalid name:**
```
Error: name must be lowercase alphanumeric + hyphens
```
**Fix:** Use only lowercase letters, numbers, hyphens (e.g., `my-skill`).

**Missing SKILL.md:**
```
Error: directory must contain SKILL.md
```
**Fix:** Create `SKILL.md` in the skill directory.

---

## Validating Agents

Agents are single `.md` files with YAML frontmatter.

### Expected Structure

Single file: `my-agent.md`

### Validation Steps

**1. Check format:**

```yaml
---
description: Brief description (required)
type: code-reviewer
instructions: Detailed instructions
---

# My Agent

Documentation...
```

**2. Validate:**

```bash
aimgr repo import ./my-agent.md --dry-run
```

**3. Test:**

```bash
aimgr repo import ./my-agent.md
cd /tmp/test
aimgr install agent/my-agent
ls .claude/agents/my-agent.md
```

---

## Validating Commands

Commands are `.md` files in a `commands/` directory.

### Expected Structure

```
commands/
├── my-command.md
└── api/                    # Nested commands supported
    └── deploy.md
```

### Validation Steps

**1. Check format:**

```yaml
---
description: Brief description (required)
---

# My Command

Documentation...
```

**Important:**
- File MUST be in a `commands/` directory
- Nested structure supported (e.g., `commands/api/deploy.md`)

**2. Validate:**

```bash
# Validate single command
aimgr repo import ./commands/my-command.md --dry-run

# Validate entire commands directory
aimgr repo import ./commands --dry-run
```

**3. Test:**

```bash
aimgr repo import ./commands
cd /tmp/test
aimgr install command/my-command
ls .claude/commands/my-command.md
```

---

## Validating Packages

Packages are JSON files that bundle multiple resources.

### Expected Structure

Single file: `my-package.package.json`

### Validation Steps

**1. Create package JSON:**

```json
{
  "name": "my-package",
  "description": "Collection of resources",
  "resources": [
    "skill/my-skill",
    "command/my-command",
    "agent/my-agent"
  ]
}
```

**Required fields:**
- `name` - Package name (lowercase, alphanumeric, hyphens)
- `description` - Package description
- `resources` - Array in `type/name` format

**2. Ensure referenced resources exist:**

**CRITICAL:** All resources must exist in repository BEFORE adding the package.

```bash
# Add individual resources first
aimgr repo import ./my-skill
aimgr repo import ./commands/my-command.md
aimgr repo import ./my-agent.md

# Verify they're in repository
aimgr repo list skill/my-skill
aimgr repo list command/my-command
aimgr repo list agent/my-agent
```

**3. Validate package:**

```bash
aimgr repo import ./my-package.package.json --dry-run
```

**4. Verify package references:**

```bash
# Add package
aimgr repo import ./my-package.package.json

# Verify all references valid
aimgr repo verify package/my-package
```

**5. Test package installation:**

```bash
cd /tmp/test
aimgr install package/my-package

# Should install ALL resources
ls .claude/skills/my-skill/
ls .claude/commands/my-command.md
ls .claude/agents/my-agent.md
```

### Common Package Errors

**Invalid resource format:**
```
Error: invalid resource format: "my-skill" (expected type/name)
```
**Fix:** Use `type/name` format: `"skill/my-skill"`

**Resource doesn't exist:**
```
Error: package references non-existent resource 'skill/missing'
```
**Fix:** Add the resource to repository first:
```bash
aimgr repo import ./missing-skill
```

---

## Common Validation Errors

### YAML Syntax Errors

**Error:**
```
Error: yaml: mapping values are not allowed in this context
```

**Cause:** Special characters in description (especially colons)

**Fix:** Quote the description:
```yaml
---
description: "My tool: a helper"  # ✓ Quoted
---
```

### Name Validation

**Consecutive hyphens:**
```
Error: name cannot contain consecutive hyphens
```
**Fix:** Use single hyphens: `my-skill` (not `my--skill`)

**Starts/ends with hyphen:**
```
Error: name cannot start/end with hyphen
```
**Fix:** Remove leading/trailing hyphens

### Description Errors

**Missing description:**
```
Error: description is required
```
**Fix:** Add `description` field to frontmatter

**Too long:**
```
Error: skill description too long (1500 chars, max 1024)
```
**Fix:** Shorten to 1024 characters or less

---

## CI/CD Integration

Automate validation in your pipeline.

### GitHub Actions

```yaml
name: Validate Resources

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install aimgr
        run: |
          go install github.com/hk9890/ai-config-manager@latest
          echo "$HOME/go/bin" >> $GITHUB_PATH
      
      - name: Validate resources
        run: |
          if aimgr repo import . --dry-run --format=json; then
            echo "✅ All resources valid"
            exit 0
          else
            echo "❌ Validation failed"
            exit 1
          fi
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Validating resources..."

# Find all skills
SKILLS=$(find . -name "SKILL.md" -not -path "./.git/*" | xargs dirname)

for skill in $SKILLS; do
  if ! aimgr repo import "$skill" --dry-run >/dev/null 2>&1; then
    echo "❌ Validation failed: $skill"
    aimgr repo import "$skill" --dry-run
    exit 1
  fi
done

echo "✅ All resources valid"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Complete Example

### Creating and Validating a Skill

**1. Create skill structure:**

```bash
mkdir pdf-processing
cd pdf-processing

cat > SKILL.md << 'EOF'
---
name: pdf-processing
description: Process and extract information from PDF files
license: MIT
version: "1.0.0"
---

# PDF Processing Skill

Helps AI assistants work with PDF files.

## Capabilities

- Extract text from PDFs
- Parse PDF metadata
EOF

# Optional: Add scripts
mkdir scripts
echo '#!/bin/bash' > scripts/extract.sh
```

**2. Validate:**

```bash
cd ..
aimgr repo import ./pdf-processing --dry-run
# ✓ skill/pdf-processing - Successfully validated
```

**3. Add to repository:**

```bash
aimgr repo import ./pdf-processing
```

**4. Test installation:**

```bash
mkdir /tmp/test && cd /tmp/test
aimgr install skill/pdf-processing
ls -la .claude/skills/pdf-processing/
```

**5. Create package:**

```bash
cd /tmp
cat > pdf-toolkit.package.json << 'EOF'
{
  "name": "pdf-toolkit",
  "description": "PDF processing tools",
  "resources": [
    "skill/pdf-processing"
  ]
}
EOF
```

**6. Validate and add package:**

```bash
aimgr repo import ./pdf-toolkit.package.json --dry-run
aimgr repo import ./pdf-toolkit.package.json
aimgr repo verify package/pdf-toolkit
```

**7. Test package:**

```bash
cd /tmp/test
aimgr install package/pdf-toolkit
ls .claude/skills/pdf-processing/
```

✅ Success! Skill is validated and ready to publish.

---

## Naming Rules

All resources must follow these naming conventions:

- Lowercase alphanumeric + hyphens only
- Cannot start or end with hyphen
- No consecutive hyphens (`--`)
- Max 64 characters per segment
- For nested commands: each segment validated separately

**Valid names:**
- `my-skill`
- `pdf-processor`
- `api-helper`

**Invalid names:**
- `My-Skill` (uppercase)
- `my_skill` (underscore)
- `-my-skill` (starts with hyphen)
- `my--skill` (consecutive hyphens)

---

## Quick Reference

```bash
# Validate without adding
aimgr repo import <path> --dry-run

# Validate and add
aimgr repo import <path>

# Verify repository
aimgr repo verify

# Verify specific resource
aimgr repo verify <pattern>

# JSON output for CI/CD
aimgr repo import <path> --dry-run --format=json
```

---

## Need Help?

- **Full documentation:** https://github.com/hk9890/ai-config-manager/tree/main/docs/user-guide
- **Resource formats:** https://github.com/hk9890/ai-config-manager/blob/main/docs/user-guide/resource-formats.md
- **Report issues:** https://github.com/hk9890/ai-config-manager/issues
