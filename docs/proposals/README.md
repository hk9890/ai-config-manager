# Package System Proposals

This directory contains proposals for adding package/grouping functionality to aimgr.

## Documents

### 1. [package-system-proposal.md](./package-system-proposal.md)
**Full-featured package system** (50+ pages)

Comprehensive proposal inspired by Claude Code plugins and npm/pip package managers.

**Key Features:**
- Semantic versioning
- Author/license metadata
- Package-level dependency management
- Dedicated package commands (`package list/show/update/remove`)
- Marketplace/registry support
- Package entity tracking

**Implementation**: 12+ weeks (5 phases)

---

### 2. [package-system-simplified.md](./package-system-simplified.md)
**Simplified package system** (addendum to above)

Lightweight approach based on user feedback: packages as grouping concept only.

**Key Principles:**
1. Packages are **grouping/distribution** mechanisms, not managed entities
2. Installing a package installs **individual resources** (no package tracking)
3. Works same way at project and repo level
4. No versioning, author info, or complex metadata
5. Resources can be removed individually after package install

**Key Features:**
- Simple `package.json` manifest (name, description, resource list)
- `aimgr install package/<name>` installs all resources
- No new package-specific commands (use existing resource commands)
- Optional: lightweight tracking via resource metadata (hybrid approach)

**Implementation**: 2-3 weeks (simplified) or 3-5 weeks (hybrid)

---

## Quick Comparison

| Aspect | Full Proposal | Simplified |
|--------|---------------|------------|
| **Versioning** | Full semver | None (use Git tags) |
| **Metadata** | Author, license, homepage, etc. | Name + description only |
| **Dependencies** | Package dependencies with constraints | System commands only |
| **Tracking** | Packages tracked separately | Resources only (package = install-time concept) |
| **Commands** | 10+ new commands | 0 new commands (extends existing) |
| **Updates** | Package-level updates | Resource-level (or bulk by source) |
| **Marketplace** | Full registry/search system | GitHub topics |
| **Complexity** | High | Low |
| **Timeline** | 12+ weeks | 2-5 weeks |

---

## User Feedback (2026-01-25)

Key points from user review:

1. **No versioning needed** - Packages should be simple grouping concept
2. **Install = install resources** - Not package as entity
3. **Individual removal** - User can delete resources after package install
4. **Same at all levels** - Project install and repo add work the same way

This feedback led to the simplified proposal.

---

## Recommendation

**Start with simplified approach** (package-system-simplified.md):
- Faster to implement
- Easier to understand
- Less code to maintain
- Can add features later if needed

**Hybrid option** (in simplified doc):
- Add lightweight tracking (package name in resource metadata)
- Enables convenience commands like `aimgr list --from-package=<name>`
- Still much simpler than full proposal
- Good balance of features and complexity

---

## Next Steps

1. **Review** both proposals
2. **Decide** on approach (simplified or hybrid)
3. **Finalize** `package.json` schema
4. **Implement** Phase 1 (package detection and install)
5. **Test** with example packages (beads-workflow, pdf-toolkit, etc.)

---

## Example Package Structure

**Simplified approach:**
```
my-package/
├── package.json              # Minimal manifest
├── commands/
│   └── *.md
├── skills/
│   └── */SKILL.md
├── agents/
│   └── *.md
└── README.md
```

**Minimal package.json:**
```json
{
  "name": "my-package",
  "description": "My package description",
  "resources": {
    "commands": ["commands/cmd1.md"],
    "skills": ["skills/skill1"],
    "agents": ["agents/agent1.md"]
  },
  "requires": ["git", "jq"]
}
```

**Installation:**
```bash
# Install package (installs all resources individually)
$ aimgr install gh:user/my-package

# Resources appear in list
$ aimgr list
COMMANDS
  cmd1
AGENTS
  agent1
...

# Remove individual resource
$ aimgr uninstall cmd1
```

---

## Questions?

See the individual proposal documents for detailed information, or open an issue for discussion.
