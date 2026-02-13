# aimgr CLI Reference: Advanced Topics

Advanced topics including pattern syntax, shell completion, resource formats, and exit codes.

## Table of Contents

- [Pattern Syntax](#pattern-syntax) - Glob patterns for matching resources
- [Shell Completion](#shell-completion) - Tab completion setup
- [Resource Formats](#resource-formats) - Format specifications
- [Exit Codes](#exit-codes) - Command exit codes
- [Version](#version) - Version information
- [Help](#help) - Getting help

---
---

## Pattern Syntax

Many commands support glob patterns for matching multiple resources. Patterns enable bulk operations and flexible resource selection.

### Pattern Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `*` | Matches any sequence of characters | `test*` matches `test`, `test-unit`, `testing` |
| `?` | Matches any single character | `test?` matches `test1`, `testa`, not `test` |
| `[abc]` | Matches any character in the set | `test[123]` matches `test1`, `test2`, `test3` |
| `{a,b}` | Matches any alternative (a or b) | `{build,test}` matches `build` or `test` |

### Pattern Format

Patterns can be specified in two formats:

1. **Typed Pattern:** `type/pattern`
   - Matches only resources of the specified type
   - Example: `skill/pdf*` matches skills starting with "pdf"

2. **Untyped Pattern:** `pattern`
   - Matches across all resource types
   - Example: `*test*` matches any resource with "test" in the name

3. **Exact Name:** `name` (no wildcards)
   - Matches the exact resource name
   - Example: `test` matches only resource named "test"

### Pattern Examples

**Installation Patterns:**
```bash
# Install all skills
aimgr install "skill/*"

# Install all resources with "test" in name
aimgr install "*test*"

# Install PDF-related skills
aimgr install "skill/pdf*"

# Install build and test commands
aimgr install "command/{build,test}"

# Install multiple patterns
aimgr install "skill/pdf*" "command/test*" "agent/qa*"
```

**Filter Patterns (repo add):**
```bash
# Add only skills from source
aimgr repo add gh:owner/repo --filter "skill/*"

# Add only test-related resources
aimgr repo add ./resources/ --filter "*test*"

# Add specific skills
aimgr repo add gh:owner/repo --filter "skill/pdf*" --filter "skill/doc*"
```

**Uninstall Patterns:**
```bash
# Uninstall all legacy skills
aimgr uninstall "skill/legacy-*"

# Uninstall all test resources
aimgr uninstall "*test*"

# Uninstall specific commands
aimgr uninstall "command/{old,deprecated}*"
```

### Pattern Matching Rules

1. **Case Sensitivity:** Patterns are case-sensitive
   - `Test` does NOT match `test`

2. **Quotes Required:** Patterns with wildcards should be quoted
   - Correct: `aimgr install "skill/*"`
   - Incorrect: `aimgr install skill/*` (shell may expand wildcard)

3. **Type Prefix:** Type prefix is optional but recommended for clarity
   - `skill/test` is more explicit than `test`
   - Prevents ambiguity when multiple types have same name

4. **Empty Results:** If no resources match pattern, command fails
   - Use `--dry-run` with `repo add` to preview matches

### Advanced Pattern Examples

**Complex Combinations:**
```bash
# Install all PDF and document skills
aimgr install "skill/{pdf,doc}*"

# Install version-specific resources
aimgr install "skill/*-v2"

# Install by category
aimgr install "skill/web-*" "skill/api-*"

# Exclude specific version
aimgr install "skill/*" --filter "!skill/*-deprecated"
```

**Character Class Patterns:**
```bash
# Install numbered test commands
aimgr install "command/test[0-9]"

# Install lettered variants
aimgr install "skill/feature-[abc]"

# Install version ranges
aimgr install "skill/tool-v[1-3]*"
```

---

## Shell Completion

`aimgr` supports tab completion for Bash, Zsh, Fish, and PowerShell. Completion provides:
- Command and subcommand completion
- Resource name completion
- Option and flag completion
- Path completion

### Installation

**Bash:**
```bash
# System-wide (requires sudo)
aimgr completion bash | sudo tee /etc/bash_completion.d/aimgr

# User-only
aimgr completion bash > ~/.bash_completion.d/aimgr
echo 'source ~/.bash_completion.d/aimgr' >> ~/.bashrc
```

**Zsh:**
```bash
# Ensure completion directory is in fpath
aimgr completion zsh > "${fpath[1]}/_aimgr"

# Or user-only
mkdir -p ~/.zsh/completions
aimgr completion zsh > ~/.zsh/completions/_aimgr
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
```

**Fish:**
```bash
# User-only (recommended)
aimgr completion fish > ~/.config/fish/completions/aimgr.fish

# System-wide (requires sudo)
aimgr completion fish | sudo tee /usr/share/fish/vendor_completions.d/aimgr.fish
```

**PowerShell:**
```powershell
# Generate completion script
aimgr completion powershell > aimgr.ps1

# Add to profile
Add-Content $PROFILE '. /path/to/aimgr.ps1'
```

### Usage Examples

After installation, use **TAB** to trigger completion:

```bash
# Command completion
aimgr <TAB>
# Shows: install  uninstall  repo  list  config  completion

# Subcommand completion
aimgr repo <TAB>
# Shows: add  list  show  update  remove

# Resource type completion
aimgr install skill/<TAB>
# Shows: pdf-processing  react-testing  web-scraper  ...

# Resource type completion
aimgr install <TAB>
# Shows: skill/  command/  agent/

# Command completion
aimgr install command/<TAB>
# Shows: build  test  lint  format  ...

# Agent completion
aimgr install agent/<TAB>
# Shows: code-reviewer  qa-agent  documentation-writer  ...

# Option completion
aimgr install --<TAB>
# Shows: --target  --force  --project-path

# Target completion
aimgr install skill/test --target <TAB>
# Shows: claude  opencode  copilot
```

### Reload Completion

After installing or updating `aimgr`, reload shell completion:

**Bash:**
```bash
source ~/.bashrc
# Or
source /etc/bash_completion.d/aimgr
```

**Zsh:**
```bash
# Clear completion cache
rm -f ~/.zcompdump
# Restart shell or run
source ~/.zshrc
```

**Fish:**
```bash
# Fish automatically reloads completions
# Or manually reload
source ~/.config/fish/completions/aimgr.fish
```

**PowerShell:**
```powershell
. $PROFILE
```

---

## Resource Formats

Resources must follow specific formats for proper detection and installation.

### Commands

Commands are single `.md` files with YAML frontmatter.

**Format:**
```yaml
---
description: What this command does (required)
agent: build (optional)
model: anthropic/claude-3-5-sonnet-20241022 (optional)
allowed-tools: [bash, read, write] (optional)
---

# Command Body

Markdown content with instructions for the AI.

## Examples

Provide examples and usage guidance.
```

**Required Fields:**
- `description` - Brief description of command purpose

**Optional Fields:**
- `agent` - Agent to route command to (e.g., `build`, `test`, `review`)
- `model` - Specific AI model to use
- `allowed-tools` - List of tools agent can use

**File Location:**
- Repository: `~/.local/share/ai-config/repo/commands/NAME.md`
- Claude Code: `.claude/commands/NAME.md`
- OpenCode: `.opencode/commands/NAME.md`

**Naming:** File name (without `.md`) is the command name (must follow [naming requirements](#resource-naming-requirements))

---

### Skills

Skills are directories containing a `SKILL.md` file (plus optional supporting files).

**Format:**
```yaml
---
name: my-skill                # Must match directory name (required)
description: What this skill does (required)
license: MIT (optional)
metadata: (optional)
  author: your-name
  version: "1.0.0"
  keywords: [keyword1, keyword2]
---

# Skill Documentation

Markdown content with skill instructions, workflows, and examples.

## Usage

How to use this skill effectively.

## Examples

Practical examples and use cases.
```

**Required Fields:**
- `name` - Skill name (must match directory name)
- `description` - Brief description of skill purpose

**Optional Fields:**
- `license` - License type (e.g., `MIT`, `Apache-2.0`)
- `metadata` - Additional metadata:
  - `author` - Skill author
  - `version` - Semantic version
  - `keywords` - Search keywords

**Directory Structure:**
```
my-skill/
├── SKILL.md           # Required: Main skill file
├── README.md          # Optional: Additional documentation
├── scripts/           # Optional: Supporting scripts
│   ├── process.py
│   └── utils.sh
└── examples/          # Optional: Example files
    └── sample.json
```

**File Location:**
- Repository: `~/.local/share/ai-config/repo/skills/NAME/`
- Claude Code: `.claude/skills/NAME/`
- OpenCode: `.opencode/skills/NAME/`
- GitHub Copilot: `.github/skills/NAME/`

**Naming:** Directory name must match `name` field in frontmatter (and follow [naming requirements](#resource-naming-requirements))

---

### Agents

Agents are single `.md` files with YAML frontmatter. Supports both OpenCode and Claude Code formats.

**OpenCode Format:**
```yaml
---
description: What this agent does (required)
type: code-reviewer (optional)
instructions: Detailed behavior (optional)
capabilities: [static-analysis, security, performance] (optional)
version: "1.0.0" (optional)
author: your-name (optional)
license: MIT (optional)
---

# Agent Documentation

Markdown content with agent instructions and guidance.

## Role

What this agent does and when to use it.

## Capabilities

Detailed capabilities and limitations.
```

**Claude Code Format:**
```yaml
---
description: What this agent does (required)
model: anthropic/claude-3-5-sonnet-20241022 (optional)
allowed-tools: [bash, read, edit] (optional)
version: "1.0.0" (optional)
author: your-name (optional)
---

# Agent Instructions

Markdown content with agent behavior and guidelines.

## Usage

When and how to invoke this agent.
```

**Required Fields:**
- `description` - Brief description of agent purpose

**Optional Fields (OpenCode):**
- `type` - Agent type/category
- `instructions` - Detailed behavior instructions
- `capabilities` - List of agent capabilities

**Optional Fields (Claude Code):**
- `model` - Specific AI model to use
- `allowed-tools` - List of tools agent can use

**Common Optional Fields:**
- `version` - Semantic version
- `author` - Agent author
- `license` - License type

**File Location:**
- Repository: `~/.local/share/ai-config/repo/agents/NAME.md`
- Claude Code: `.claude/agents/NAME.md`
- OpenCode: `.opencode/agents/NAME.md`

**Naming:** File name (without `.md`) is the agent name (must follow [naming requirements](#resource-naming-requirements))

---

## Resource Naming Requirements

All resources (commands, skills, agents) must follow these naming rules:

**Rules:**
- **Length:** 1-64 characters
- **Characters:** Lowercase letters (a-z), numbers (0-9), hyphens (-)
- **Start/End:** Must start and end with alphanumeric character
- **Hyphens:** No consecutive hyphens (`--`)
- **Case:** Lowercase only (no uppercase)

**Valid Examples:**
- `test`
- `run-coverage`
- `pdf-processing`
- `skill-v2`
- `beads-task-agent`
- `code-reviewer-v3`

**Invalid Examples:**
- `Test` - Contains uppercase
- `test_coverage` - Contains underscore
- `-test` - Starts with hyphen
- `test-` - Ends with hyphen
- `test--cmd` - Consecutive hyphens
- `UPPERCASE` - All uppercase
- `my.skill` - Contains period

**Enforcement:**
- `aimgr repo add` validates names and rejects invalid resources
- Invalid names cause command failure with clear error message
- Existing resources with invalid names cannot be installed

---

## Exit Codes

`aimgr` uses standard exit codes to indicate success or failure.

| Code | Meaning | Description |
|------|---------|-------------|
| 0 | Success | Command completed successfully |
| 1 | General Error | Command failed due to user error or invalid input |
| 2 | Command Error | Command failed due to system or tool error |
| 3 | Resource Not Found | Specified resource does not exist |
| 4 | Already Exists | Resource already exists (without `--force`) |
| 5 | Invalid Pattern | Pattern syntax is invalid or matches nothing |
| 6 | Permission Denied | Insufficient permissions to perform operation |

**Examples:**

```bash
# Check exit code in Bash
aimgr install skill/test
echo $?  # 0 = success, non-zero = error

# Use in scripts
if aimgr install skill/pdf-processing; then
    echo "Installation successful"
else
    echo "Installation failed with code $?"
fi

# Chain commands (stop on failure)
aimgr repo add gh:owner/repo && aimgr install "skill/*"

# Continue on failure
aimgr install skill/test || echo "Installation failed, continuing..."
```

**Error Output:**

Errors are written to stderr with clear messages:

```bash
$ aimgr install skill/nonexistent
Error: Resource not found: skill/nonexistent
Run 'aimgr repo list skill' to see available skills.
$ echo $?
3
```

---

## Version

View `aimgr` version information:

```bash
# Show version
aimgr version

# Show version with detailed build info
aimgr version --full
```

**Output:**
```
aimgr version 1.12.0
Build: 2026-01-25
Commit: a1b2c3d
```

---

## Help

Get help on commands:

```bash
# General help
aimgr help
aimgr --help
aimgr -h

# Command-specific help
aimgr install --help
aimgr repo add --help
aimgr config --help
```

---

## Additional Resources

- **Repository:** https://github.com/hk9890/ai-config-manager
- **Issues:** https://github.com/hk9890/ai-config-manager/issues
- **Discussions:** https://github.com/hk9890/ai-config-manager/discussions
- **Specifications:**
  - Claude Code commands: https://code.claude.com/docs/en/slash-commands
  - Agent Skills: https://agentskills.io/specification
  - Claude Code agents: https://code.claude.com/docs/agents
  - OpenCode agents: https://opencode.ai/docs/agents
