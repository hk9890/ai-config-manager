# Discover & Recommend Resources

Scan project context, match against the aimgr repository, and recommend relevant resources.

**Sections:** [Workflow](#workflow) · [Signal Table](#signal-table) · [Filtering & Ranking](#filtering--ranking) · [Notes](#notes-for-agents)

---

## Workflow

### 1. Read Project Context

Check for `.coder/project.yaml` first (written by opencode-coder plugin at startup):

```bash
cat .coder/project.yaml 2>/dev/null
```

If present, parse:
- `git.platform` — github, gitlab, bitbucket, or null
- `beads.initialized` — whether beads is set up
- `aimgr.installed` — whether aimgr is available

**If absent:** Fall back to file-based scanning (step 2). Do NOT run inline detection commands.

### 2. Scan for Tech Signals

Look for files that indicate the project's tech stack:

```bash
ls package.json go.mod pyproject.toml Cargo.toml Dockerfile 2>/dev/null
ls -d .github/workflows .beads docs 2>/dev/null
```

See [Signal Table](#signal-table) for the full mapping.

### 3. Query Available Resources

```bash
aimgr repo list --format=json
aimgr list                      # Already installed — exclude from recommendations
```

### 4. Filter and Rank

Apply [exclusion rules](#exclusion-rules), then [rank](#ranking) by relevance.

### 5. Present Recommendations

Use the `question()` tool — **mandatory user interaction, do NOT auto-install**:

```text
Based on your project, these resources look useful:

| Resource | Why |
|----------|-----|
| skill/go-testing | Go project (go.mod) |
| skill/github-releases | GitHub repo detected |
| skill/pptx | .pptx files found |

Want me to install any of these?
```

After selection: `aimgr install <chosen-resources>`

⚠️ Remind user to restart their AI tool after install.

---

## Signal Table

| Signal | Indicates |
|--------|-----------|
| `package.json` | Node.js / JavaScript / TypeScript |
| `tsconfig.json` | TypeScript |
| `go.mod` | Go |
| `pyproject.toml`, `requirements.txt` | Python |
| `Cargo.toml` | Rust |
| `pom.xml`, `build.gradle` | Java / Kotlin |
| `Gemfile` | Ruby |
| `Dockerfile`, `docker-compose.yml` | Containers |
| `.github/workflows/` | GitHub Actions CI/CD |
| `.gitlab-ci.yml` | GitLab CI/CD |
| `*.pptx` | Presentation needs |
| `*.pdf` | PDF processing needs |
| `.beads/` | Beads task tracking |
| `docs/`, `CONTRIBUTING.md` | Documentation workflows |
| `terraform/`, `*.tf` | Infrastructure as code |

---

## Filtering & Ranking

### Exclusion Rules

Non-negotiable — agent MUST apply these:

| Rule | Action |
|------|--------|
| Skill mentions "bitbucket" and `git.platform` ≠ bitbucket | EXCLUDE |
| Skill mentions "github" and `git.platform` ≠ github | EXCLUDE |
| Skill mentions "gitlab" and `git.platform` ≠ gitlab | EXCLUDE |
| Skill name contains `-dev` (internal development skills) | EXCLUDE |
| Already installed (`aimgr list`) | EXCLUDE (mention as "already installed") |

### Ranking

- Platform-matching skills rank highest
- Skills complementing already-installed ones rank higher
- Generic skills (fix-documentation, observability-triage) are always candidates
- Check for packages that bundle related resources: `aimgr repo describe package/*`

---

## Notes for Agents

- **Don't hardcode mappings.** Always query `aimgr repo list` — repository contents change.
- **Be conservative.** Only recommend with a clear signal match.
- **Respect user choice.** Present options, never auto-install.
- **Match on descriptions.** Resource descriptions from `repo list --format=json` are the primary signal.
