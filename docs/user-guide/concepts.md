# aimgr Concepts

This page explains the main concepts of **aimgr** from a user perspective.

## What aimgr does

`aimgr` helps you manage AI resources such as **skills**, **agents**, and **commands** across projects and AI tools.

It separates two concerns:

1. **Where resources come from**
2. **Which resources a project wants**

Those are represented by two different manifest files.

---

## 1. The repository: `ai.repo.yaml`

The aimgr repository is your local catalog of available resources.

- You add **sources** to it
- You **sync** those sources to refresh available resources
- You can **remove** sources you no longer want

Typical source commands:

```bash
aimgr repo add gh:your-org/ai-tools
aimgr repo add local:~/my-resources
aimgr repo sync
aimgr repo remove my-source
```

### What is a source?

A source is a place where resources live:

- a GitHub repository
- another Git repository URL
- a local folder on your machine

### What `repo sync` does

`aimgr repo sync` refreshes your repository from all configured sources.

- For **remote sources**, it fetches the latest content
- For **local sources**, aimgr keeps them connected for live editing workflows

Think of `ai.repo.yaml` as:

> “Where can aimgr find resources?”

---

## 2. The project manifest: `ai.package.yaml`

`ai.package.yaml` belongs to a specific project.

It defines which resources that project wants installed.

Typical project commands:

```bash
aimgr install
aimgr install skill/code-review
aimgr uninstall skill/code-review
aimgr list
aimgr verify
aimgr repair
```

If an `ai.package.yaml` exists in the current project, running:

```bash
aimgr install
```

installs everything declared in that manifest.

Think of `ai.package.yaml` as:

> “What does this project want to use?”

---

## How they work together

The two manifests have different jobs:

| File | Purpose |
|------|---------|
| `ai.repo.yaml` | Tracks resource sources for your local aimgr repository |
| `ai.package.yaml` | Tracks which resources a project depends on |

In short:

1. Add and sync sources into your repository
2. Install selected resources into a project
3. Track project dependencies in `ai.package.yaml`

---

## Mental model

You can think about aimgr like this:

- **`ai.repo.yaml`** = your supply side
- **`ai.package.yaml`** = your project requirements
- **`aimgr install`** = bring required resources into the project
- **`aimgr repo sync`** = refresh the supply side

---

## Common workflow

```bash
# 1. Configure where resources come from
aimgr repo add gh:your-org/ai-tools
aimgr repo sync

# 2. Move into a project
cd ~/my-project

# 3. Install what the project needs
aimgr install
```

If the project has no `ai.package.yaml` yet, you can still install resources directly,
and aimgr can track them in the project manifest.

---

## When to use which file

Use **`ai.repo.yaml`** when you want to:

- add a new source
- sync updates from upstream
- remove a source
- manage your available resource catalog

Use **`ai.package.yaml`** when you want to:

- define project dependencies
- reproduce the same setup in another clone
- install all project resources with `aimgr install`
- verify whether a project is in sync

---

## See also

- [Getting Started](getting-started.md)
- [Sources](sources.md)
- [Configuration](configuration.md)
- [Repairing Resources](repair.md)
