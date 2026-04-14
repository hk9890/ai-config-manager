# Team and Multi-Project Workflows

This guide shows practical ways to use **aimgr** in larger development setups where:

- a team wants one shared, published baseline for a project
- a team wants reproducible onboarding for new clones
- some projects ship their own project-specific AI resources
- individual developers still want private additions
- one user works on many projects with a single local aimgr repository

## Recommended Mental Model

Use **one local aimgr repository per user**, then layer project setup on top of it:

- **`ai.repo.yaml`** = where resources come from
- **`ai.package.yaml`** = what a project depends on
- **project resource folders** = optional project-owned custom resources

In a larger setup, think in layers:

1. **Org / platform sources** — shared across many projects
2. **Team sources** — shared by one team or area
3. **Project sources** — only needed when a project ships its own resources
4. **Project package manifest** — the actual resources that project wants installed

`ai.package.yaml` can now also declare remote bootstrap sources in a `sources:` block.
That lets `aimgr install` bootstrap missing remote sources as part of project install.

## Workflow 1: Central Shared Team Manifest (Recommended Default)

Start by agreeing on the shared sources for a project, then publish that decision as a central `ai.repo.yaml`.

What the team owns:

1. agree on shared source repos/paths for the project
2. define them in one shared `ai.repo.yaml`
3. publish that manifest at a stable URL (for example in a team config repo)

Each project should still commit `ai.package.yaml` so every clone gets the same declared resource dependencies.

Example:

```yaml
resources:
  - package/company-base-dev
  - skill/project-review

install:
  targets:
    - claude
    - opencode
```

Developer bootstrap (new clone):

```bash
aimgr repo apply-manifest https://example.com/team/project-a/ai.repo.yaml
aimgr repo sync
aimgr install
```

CI bootstrap uses the exact same sequence:

```bash
aimgr repo apply-manifest https://example.com/team/project-a/ai.repo.yaml
aimgr repo sync
aimgr install
```

### Re-applying an updated shared manifest (additive behavior)

`repo apply-manifest` merges sources into local `ai.repo.yaml` and is **additive**.
If a source is removed from the shared manifest upstream, re-applying the updated
manifest will not auto-delete that source from existing local repos.

Team update guidance:

1. Publish updated shared `ai.repo.yaml`.
2. Re-apply it locally/in CI with `repo apply-manifest`.
3. Explicitly remove stale sources with `repo remove` when needed.

Example:

```bash
aimgr repo apply-manifest https://example.com/team/project-a/ai.repo.yaml
aimgr repo remove source-a
aimgr repo sync
aimgr install
```

### URL examples (including pinned/tagged GitHub raw URLs)

```bash
# Unpinned GitHub raw file URL
aimgr repo apply-manifest https://raw.githubusercontent.com/your-org/team-configs/main/manifests/project-a.ai.repo.yaml

# Pinned to a tag for stable/reproducible bootstrap
aimgr repo apply-manifest https://raw.githubusercontent.com/your-org/team-configs/v1.2.0/manifests/project-a.ai.repo.yaml
```

Important constraint for GitHub-hosted manifests:

- supported: direct file-content endpoints (for example `raw.githubusercontent.com/.../ai.repo.yaml`)
- not supported: GitHub web URLs such as `/blob/<ref>/.../ai.repo.yaml` or `/tree/<ref>/...`

This works best when shared manifests publish reusable `package/*` resources so projects can keep `ai.package.yaml` small.

## Workflow 2: Project Ships Its Own Resources (Secondary Pattern)

If a project has custom skills, commands, or agents, keep them in the repository and publish a project-specific repo manifest in addition to the central team baseline.

Example layout:

```text
my-project/
├── ai.package.yaml
├── aimgr/
│   └── ai.repo.yaml
└── ai-resources/
    ├── skills/
    ├── commands/
    └── packages/
```

Example `aimgr/ai.repo.yaml`:

```yaml
version: 1
sources:
  - name: company-platform
    url: https://github.com/acme/ai-resources
    ref: v1.4.0
    include:
      - package/company-*
      - skill/shared-*

  - name: project-local
    path: ../ai-resources
```

Clone workflow (central baseline + project-local resources):

```bash
aimgr repo apply-manifest https://example.com/team/project-a/ai.repo.yaml
aimgr repo apply-manifest ./aimgr/ai.repo.yaml
aimgr repo sync
aimgr install
```

Notes:

- relative `path:` entries work when applying a **local manifest file**
- for **remote** shared manifests, prefer `url:` sources
- project-local resources are a good fit for project-specific prompts, review skills, or helper packages

## Workflow 3: Personal Extras on Top of a Shared Project

For private additions, prefer the project-local overlay file `ai.package.local.yaml`.

Recommended pattern:

1. keep personal/private sources in a user-owned `ai.repo.yaml` (outside the project) when needed
2. apply and sync sources into your local repository
3. add private project-local resources to `ai.package.local.yaml`

Example:

```bash
aimgr repo apply-manifest ~/.config/aimgr/personal.ai.repo.yaml
aimgr repo sync

# ai.package.local.yaml (project-local, typically gitignored)
# resources:
#   - package/my-personal-tools

aimgr install
```

Use `--no-save` for temporary one-off installs only.

### How overlay coexists with shared team baselines

In team setups, committed `ai.package.yaml` remains the shared baseline, while
`ai.package.local.yaml` adds per-developer local-only entries.

Overlay behavior is explicit and opt-in:

- aimgr reads `ai.package.local.yaml` only when the file exists
- aimgr does not auto-create the overlay file
- aimgr does not auto-edit `.gitignore`

Effective project state is merged:

- effective resources = union of base + local resources
  - preserve committed `ai.package.yaml` order
  - append local-only additions
  - de-duplicate exact duplicates
- effective `install.targets` = union of base + local targets
- explicit CLI `--target` overrides manifest targets

Commands that use this merged view include `install`, `verify`, `repair`, and `list`.

Uninstall persistence removes resources from every manifest file that declares them
(committed manifest and/or local overlay).

If local overlay resources cannot be resolved from available sources in your local repo,
the failure is reported explicitly (not silently ignored).

## Workflow 4: One User, Many Projects

Using one local aimgr repository across many projects is the intended model.

Example:

```bash
# shared baseline
aimgr repo apply-manifest https://example.com/platform/ai.repo.yaml

# optional team layer
aimgr repo apply-manifest https://example.com/data/ai.repo.yaml

# project-specific layer from checked-out repo
cd ~/work/project-a
aimgr repo apply-manifest ./aimgr/ai.repo.yaml
aimgr repo sync
aimgr install
```

Later, another project can add more sources into the same local repo:

```bash
cd ~/work/project-b
aimgr repo apply-manifest ./aimgr/ai.repo.yaml
aimgr repo sync
aimgr install
```

This scales well as long as teams treat the local repo as a **shared catalog** and use stable source names.

## Workflow 5: Project-manifest bootstrap with `sources:`

If you want onboarding from a single file, declare remote sources directly in `ai.package.yaml`.

Example (`project-a/ai.package.yaml`):

```yaml
sources:
  - name: platform-default
    url: https://github.com/acme/platform-ai
    ref: v1.4.0
    subpath: resources

resources:
  - package/platform-dev
```

First install in a fresh environment:

```bash
# no prior repo init needed
cd ~/work/project-a
aimgr install
```

Behavior:

- If no local aimgr repo exists, install bootstraps it automatically.
- Install then bootstraps manifest-declared remote sources into local `ai.repo.yaml`.
- Resources from `resources:` are installed into the project.

Repeated install with unchanged manifest is idempotent:

```bash
aimgr install   # source bootstrap is a no-op when already satisfied
```

### Shared-catalog identity rules for manifest-declared sources

For `ai.package.yaml` `sources:` entries, canonical source reuse is based on normalized `url + subpath`.

Implications:

- `name` is a first-install hint/display alias only
- `ref` is a bootstrap hint only
- `ref` is not an isolation boundary in the shared local catalog

If an existing canonical source already matches `url + subpath`, install reuses it even if another project uses a different `name` or requests a different `ref`.

For project-manifest bootstrap entries (`ai.package.yaml` `sources:`), only `url`, `ref`, `subpath`, and optional `name` are supported. Use `include` filters only in repo-managed source workflows (`ai.repo.yaml`, `repo add --filter`, `repo apply-manifest`).

### Two projects sharing one upstream with different aliases/refs

`project-a/ai.package.yaml`:

```yaml
sources:
  - name: team-catalog
    url: https://github.com/acme/platform-ai
    ref: v1.4.0
    subpath: resources
resources:
  - package/platform-dev
```

`project-b/ai.package.yaml`:

```yaml
sources:
  - name: platform-tools
    url: https://github.com/acme/platform-ai
    ref: v2.0.0
    subpath: resources
resources:
  - package/platform-ops
```

With the same user-local repo:

1. `project-a` install bootstraps the canonical source
2. `project-b` install reuses that same canonical source (`url + subpath`)
3. install warns that requested `ref` differs and keeps existing source unchanged

To intentionally change the shared source ref, update the source via repo workflow (`aimgr repo add <url> --ref <new-ref> ...` or `aimgr repo sync` after explicit repo-source update), then re-run project install.

### Warning and failure behavior to expect

When reusing an existing canonical source for manifest bootstrap:

- **Ref mismatch warning**: existing source is reused under its current ref; requested manifest ref is advisory only.
- **Too-narrow filter failure**: if existing source include filters cannot satisfy required manifest resources, install fails with guidance to broaden filters or adjust source configuration before retrying.

This is intentional shared-catalog behavior: source definitions are reused across projects instead of silently forking per-project clones.

## Source Naming Rules for Multi-Project Setups

Source naming discipline matters.

For shared baseline manifests consumed through `repo apply-manifest`, source naming remains the primary merge key.

When `repo apply-manifest` merges manifests:

- **new source name** → source is added
- **same source name + identical definition** → no-op
- **same source name + same canonical source (`path`/`url`/`subpath`) but updated `ref`** → ref is updated (supported)
- **same source name + different canonical source definition** → explicit failure

Important consequence for `repo apply-manifest`:

- if two manifests reference the same upstream repo with different names, they are treated as separate source entries during apply/merge
- if two manifests use the same source name for different definitions, apply fails with a conflict (no silent overwrite)

### Canonical naming guidance

Use stable, canonical names based on ownership/scope, not per-project improvisation.

Good examples:

- `org-platform`
- `team-data`
- `project-foo-local`

Avoid having different projects invent different names for the same shared upstream repo.

### Collision troubleshooting for teams

Collision failures are explicit and should be treated as coordination problems in shared manifests.

If apply/sync fails, check:

1. **Source-definition conflict**
   - symptom: same `sources[].name` but different canonical source location/identity (for example different `url`, `path`, or `subpath`)
   - fix: align on one canonical source definition and reuse it across all shared manifests

2. **Canonical resource-name collision**
   - symptom: two different sources provide the same canonical resource ID (`type/name`)
   - fix: rename one resource at the source, or adjust source `include` patterns so only one canonical owner publishes that name

3. **Inconsistent naming conventions across teams**
   - symptom: repeated conflicts whenever manifests are combined
   - fix: maintain a short naming convention doc and treat shared `ai.repo.yaml` as the source-of-truth contract

## Recommended Team Pattern

For larger setups, prefer this structure:

### Org level

- one or a few shared resource repositories
- reusable `package/*` resources for common bundles
- published shared `ai.repo.yaml` manifests used as project baselines

### Project level

- commit `ai.package.yaml`
- reference a central shared manifest URL for team bootstrap
- add a project `aimgr/ai.repo.yaml` only if the project ships custom resources
- keep project-specific custom resources in a committed folder such as `ai-resources/`

### User level

- optional personal repo manifest outside the project
- `ai.package.local.yaml` for durable private project-local additions
- optional `--no-save` installs for temporary extras

## Practical Onboarding Recipe

For a project with shared team resources and optional project-local resources:

```bash
# 1. Extend local repo with shared and project-specific source manifests
aimgr repo apply-manifest https://example.com/team/project-a/ai.repo.yaml
aimgr repo apply-manifest ./aimgr/ai.repo.yaml

# 2. Import or refresh resources into the local repo
aimgr repo sync

# 3. Install project dependencies
aimgr install
```

Alternative (single-file onboarding): if the project commits remote `sources:` in `ai.package.yaml`, use `aimgr install` directly and let install bootstrap missing remote sources automatically.

## Choosing the right bootstrap workflow

- Use **`ai.package.yaml` `sources:` + `aimgr install`** when a project wants self-contained install/bootstrap behavior in one manifest.
- Use **`repo apply-manifest`** when a team publishes one shared `ai.repo.yaml` baseline (often reused by many projects and CI setups).
- Use **`repo add`** for ad-hoc/manual source administration outside committed project manifests.
