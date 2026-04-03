# Coding Guide

Essential coding reference for ai-config-manager contributors.

## CRITICAL: Repository Safety for Testing

**NEVER run `aimgr repo` commands against the global repository during testing or bug reproduction!**

The default repository location is `~/.local/share/ai-config/repo/` which contains your real aimgr configuration. Testing against this will corrupt your development environment.

**Safe methods:**

| Method | Usage |
|--------|-------|
| Environment variable (recommended) | `export AIMGR_REPO_PATH=/tmp/test-repo-$(date +%s)` |
| Config file | `aimgr --config /tmp/test-config.yaml repo init` |
| Go tests | `repo.NewManagerWithPath(t.TempDir())` |

**Bottom line**: Every test operation MUST explicitly specify a temporary repository location. No exceptions.

## CRITICAL: Use Locally Built Binary

**ALWAYS use `./aimgr` (the locally built binary) when testing changes, NOT `aimgr` from PATH!**

Version managers (mise, asdf, etc.) may install older versions that are found first in PATH.

```bash
# CORRECT: Use local binary
./aimgr --version
./aimgr repo init

# WRONG: May use mise/asdf version
aimgr --version
```

## Concurrency and Locking Model (Repo/Workspace Mutations)

All repo mutations are coordinated with OS-backed advisory file locks under:

`<repo>/.workspace/locks/`

Lock files:

- Repo-wide lock: `<repo>/.workspace/locks/repo.lock` (single path for both read and write)
- Workspace metadata lock: `<repo>/.workspace/locks/workspace-metadata.lock`
- Per-cache lock: `<repo>/.workspace/locks/caches/<cache-hash>.lock`

Repo lock modes:

- **Read/shared lock**: allows concurrent readers on POSIX (`flock` `LOCK_SH`)
- **Write/exclusive lock**: blocks all other readers/writers (`flock` `LOCK_EX`)
- Read and write modes use the **same `repo.lock` file**.

Lock ordering is strict and must never be reversed:

1. repo lock
2. cache lock
3. workspace metadata lock

Rules:

- Top-level mutating CLI commands are outermost repo-lock holders.
- Workspace cache operations may take cache lock and metadata lock, but must not try to take repo lock from inside those sections.
- Workspace metadata lock is only for short metadata read-modify-write sections (not long clone/fetch/pull work).
- Cache and workspace metadata locks remain **exclusive-only**.
- Locks are intentionally **non-reentrant per process and path**: acquiring the same lock path twice in one process fails with `ErrNonReentrantLock`.
- Same-process lock transition attempts on the same path (read→write or write→read) also fail through the same non-reentrant contract (no implicit upgrade/downgrade protocol).

Scope / limitations:

- These locks are implemented with `flock` on Unix/POSIX builds.
- On Windows, locks use `LockFileEx`/`UnlockFileEx` from the Windows API.
- **Windows fallback behavior**: repo read-lock APIs currently map to exclusive `LockFileEx` locking, so concurrent shared readers are not currently enabled on Windows.
- `flock` does not provide fairness guarantees; under heavy contention, lock starvation is possible. `aimgr` uses timeout-bounded acquisition rather than starvation prevention.
- Locks serialize *aimgr* mutation paths; they do not prevent arbitrary external tools from modifying files directly.

## Repo Lock Command Matrix

Use this matrix when adding or changing commands that touch shared repo state.

| Command / Path | Lock mode | Why |
|---|---|---|
| `aimgr repo init` | write | Creates/updates repo structure and metadata. |
| `aimgr repo add` | write | Adds resources/metadata to shared repo state. |
| `aimgr repo sync` | write | Mutates repo content and source metadata. |
| `aimgr repo remove` | write | Removes source entries and associated metadata. |
| `aimgr repo drop` | write | Deletes repository data. |
| `aimgr repo apply-manifest` | write | Applies source mutations from manifest. |
| `aimgr repo override-source` | write | Mutates source overrides/metadata. |
| `aimgr repo prune` | write | Deletes stale resources/metadata. |
| `aimgr repo repair` | write | Rewrites repo metadata/content for consistency. |
| `aimgr repo verify --fix` | write | Applies corrective repo mutations. |
| `aimgr repo verify` (read-only) | read | Scans shared repo files/metadata without mutation. |
| `aimgr repo info` | read | Reads repo manifest/resources/metadata snapshot. |
| `aimgr repo list` | read | Reads resource/package/metadata listings. |
| `aimgr repo describe` / `repo show` | read | Reads resource/package definitions + metadata. |
| `aimgr repo show-manifest` | read | Reads repo manifest snapshot. |
| `aimgr install` | read | Expands repo patterns/packages and validates resources during install planning. |
| `installFromManifest` (internal) | read | Reads repo package/resource definitions while expanding manifest refs. |
| package install expansion (`installPackage`) | read (via `install`) | Reads package definitions and member resources from shared repo state. |
| `aimgr repair` (project) | read | Expands package refs and validates repo-backed declared resources before reconcile. |
| `aimgr verify` (project) | read | Resolves package refs and repo-backed membership during manifest checks/reconcile wrapper. |
| `aimgr list` (installed resources) | read | Expands manifest package refs by reading repo package definitions. |
| `aimgr uninstall` | read | Reads repo package definitions for `package/*` uninstall expansion. |
| repo-backed shell completions | try-read (non-blocking) | Completions must stay responsive; return no dynamic suggestions while writers hold lock. |

Intentional no-lock helpers:

- `cmd.ExpandPattern` does not acquire repo locks itself. Callers must hold the appropriate repo lock before invoking it when a stable repo snapshot is required.

## Atomic Write Model (Repo-Managed State Files)

Repo-managed state files are written with **atomic replacement**, not in-place
truncating writes.

State files covered by this model include:

- `ai.repo.yaml`
- `.metadata/sources.json`
- Resource metadata files under `.metadata/...` (for example
  `.metadata/skills/<name>-metadata.json`)
- `.workspace/.cache-metadata.json`

Write sequence (via `pkg/fileutil.AtomicWrite`):

1. Create a temporary file in the **same directory** as the target file.
2. Write full file contents to the temp file.
3. `fsync` the temp file.
4. Rename the temp file over the destination path.
5. `fsync` the parent directory where supported.

Important limits:

- Parent directories must already exist before writing.
- Atomic replacement protects against partial-file writes during process crashes,
  but it does not by itself resolve concurrent read-modify-write races; locks
  still provide serialization for mutation paths.
- Parent-directory `fsync` is best-effort by platform; on Windows this layer
  currently does not perform directory `fsync`.

## Quick Commands

```bash
# Build
make build      # Build binary to ./aimgr
make install    # Build and install to ~/bin

# Test
make test             # All tests (vet -> unit [cmd+pkg] -> integration)
make unit-test        # Fast unit tests only
make integration-test # Integration tests

# Code Quality
make fmt        # Format all Go code
make vet        # Run go vet
```

## Project Structure

```
cmd/    CLI command implementations (Cobra)
pkg/    Business logic (20 packages)
test/   Integration and E2E tests
docs/   Documentation
```

**Architecture**: CLI (Cobra) -> Business Logic (`pkg/`) -> Storage (XDG directories)

## Detailed Guides

- **[Code Style](contributor-guide/code-style.md)** -- Naming, imports, error handling, symlink handling, best practices
- **[Architecture](contributor-guide/architecture.md)** -- System overview, package responsibilities, 5 critical rules, data flows
- **[Development Environment](contributor-guide/development-environment.md)** -- IDE setup, mise, build tools

## Before Committing

1. `make fmt` -- Format code
2. `make test` -- All tests pass
3. Follow [code style guide](contributor-guide/code-style.md)
4. Git operations use `pkg/workspace` (see [architecture](contributor-guide/architecture.md))
5. Tests use `t.TempDir()` and `NewManagerWithPath()` (see [testing](contributor-guide/testing.md))
