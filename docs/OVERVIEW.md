# Overview

## What this repository builds

- Go module: `github.com/dynatrace-oss/ai-config-manager/v3`
- CLI entrypoint: `cmd/aimgr`
- Output binary: `./aimgr` from `make build`
- Supported install targets live under project directories such as `.claude/`, `.opencode/`, `.github/skills/`, `.github/agents/`, and `.windsurf/skills/`

`aimgr` manages reusable AI resources — commands, skills, agents, and packages — through a central repository plus per-project installs.

## Core runtime model

1. Source definitions are tracked in `ai.repo.yaml`.
2. Repository import and sync logic lives in `pkg/repo/`, `pkg/source/`, `pkg/workspace/`, and `pkg/sourcemetadata/`.
3. Project dependency state lives in `ai.package.yaml` plus optional `ai.package.local.yaml` overlays; `ai.package.yaml` may also declare remote bootstrap hints in `sources:` for install-time source bootstrap.
4. Install and uninstall flows create or repair project-local resource links through `pkg/install/`, `pkg/tools/`, `cmd/install.go`, and `cmd/uninstall.go`.
5. Validation and repair flows are exposed through commands such as `cmd/resource_validate.go`, `cmd/project_verify.go`, and `cmd/repair.go`.

## Repo map

| Path | Use it for |
| --- | --- |
| `cmd/` | Cobra command wiring, flags, and CLI-facing behavior |
| `pkg/config/` | Config loading, XDG paths, environment expansion, and install-target defaults |
| `pkg/discovery/` | Resource discovery across local and cloned repositories |
| `pkg/repo/` | Central repository mutations and lookups |
| `pkg/repomanifest/` | `ai.repo.yaml` load/save and source-definition helpers |
| `pkg/source/` | Source parsing, GitHub/local URL handling, and source-level helpers |
| `pkg/sourcemetadata/` | Persisted source sync state and source metadata tracking |
| `pkg/workspace/` | Cached Git workspaces and clone/update performance |
| `pkg/install/` | Project install, uninstall, and symlink/copy behavior |
| `pkg/manifest/` | Project manifest parsing for `ai.package.yaml` (`resources`, optional `sources`) and overlays |
| `pkg/metadata/` | Resource metadata helpers for repo-managed state |
| `pkg/pattern/` | Pattern parsing and compiled glob matching for resource selection |
| `pkg/tools/` | Tool-specific target directories and install conventions |
| `pkg/resource/` | Resource loaders, validation, and shared resource model |
| `pkg/repolock/` | Repo-level lock acquisition, timeout behavior, and shared/exclusive semantics |
| `pkg/fileutil/` | Atomic file writes and low-level filesystem helpers used by repo state updates |
| `pkg/errors/` | Typed error categories and shared error helpers |
| `pkg/logging/` | Structured logging setup and log writer helpers |
| `pkg/marketplace/` | Marketplace discovery, parsing, and generation helpers |
| `pkg/frontmatter/` | Frontmatter parsing helpers shared by resource loaders |
| `pkg/giturl/` | Git URL normalization helpers |
| `pkg/modifications/` | Change tracking helpers for install/repair flows |
| `pkg/output/` | Table/JSON/YAML output formatting |
| `pkg/version/` | Build-time version metadata exposed by the CLI |
| `test/` | Integration coverage against real repo and CLI workflows |
| `test/e2e/` | Full binary end-to-end coverage behind the `e2e` build tag |
| `docs/` | Canonical project docs and contributor guidance |
| `scripts/` | Install/bootstrap helper scripts validated in CI |
| `.github/workflows/` | CI and release automation |
| `examples/` | Sample resources and example layouts used in docs/tests |

## Start points by task

| Task | Start here |
| --- | --- |
| Add or change a CLI command | `cmd/`, then matching `cmd/*_test.go` files |
| Change repository sync/import behavior | `pkg/repo/`, `pkg/source/`, `pkg/workspace/`, `test/*sync*` |
| Change install targets or tool support | `pkg/install/`, `pkg/tools/`, `docs/reference/supported-tools.md` |
| Change resource parsing or validation | `pkg/resource/`, `cmd/resource_validate.go`, `test/resource_validate_test.go` |
| Investigate lock or mutation safety | `pkg/repolock/`, `pkg/fileutil/`, `docs/CODING.md`, `docs/TESTING.md` |
| Update contributor or agent workflows | `AGENTS.md`, `CONTRIBUTING.md`, and top-level docs in `docs/` |

## Related docs

- `docs/CODING.md` — implementation constraints, build commands, and mutation safety
- `docs/TESTING.md` — required checks, isolation rules, and change-to-test mapping
- `docs/RELEASING.md` — repo-local release facts that supplement the `github-releases` skill
- `docs/CHANGE-WORKFLOW.md` — commit, push, branch, PR, and merge expectations
