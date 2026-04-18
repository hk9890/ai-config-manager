# Testing Guide

## Command map

| Goal | Command | What it covers |
| --- | --- | --- |
| Baseline contributor checks | `make test` | `make vet` + `make unit-test` + `make integration-test` |
| Fast local regression | `make unit-test` | `go test -v -short ./cmd/...` and `./pkg/...` (excludes `//go:build integration` files) |
| Integration coverage | `make integration-test` | `go test -v -tags=integration ./cmd/...` + `./pkg/...` plus `go test -v ./test/...` |
| Full CLI end-to-end coverage | `make e2e-test` | `go test -tags=e2e ./test/e2e/` |
| Main CI test-job parity | `go test -race ./...` | Mirrors the non-E2E `Run tests` step in `.github/workflows/build.yml` |

## Minimum checks by change type

- Docs-only edits under `docs/`, `README.md`, or `AGENTS.md`: verify affected links, paths, and commands; no Go test run is required unless the documented behavior changed in the same change.
- Changes in `cmd/` or `pkg/` with no user-facing workflow impact: run `make test`.
- Changes to install flows, project layout handling, script entrypoints, or supported-tool behavior: run `make test` and `make e2e-test`.
- Changes to repo locking, workspace caching, Git sync/import, or filesystem mutation safety: run `make integration-test` at minimum, and prefer `make test` for session close.

## Critical Rules

- **ALWAYS** use `t.TempDir()` for temporary directories
- **ALWAYS** use `repo.NewManagerWithPath(tmpDir)` in tests -- NEVER `NewManager()`
- **NEVER** write to `~/.local/share/ai-config/` in tests
- **ALWAYS** run `make integration-test` before finishing larger work such as a bigger feature, an epic, or broad cross-cutting changes if you were relying on focused tests during development.
- **ALWAYS** add `make e2e-test` when a change affects CLI entrypoints, installation flows, or user-visible end-to-end behavior.

## Test locations

- `cmd/*_test.go` and `pkg/**/*_test.go` without `//go:build integration` — fast local coverage run by `make unit-test`
- `cmd/**/*_test.go` and `pkg/**/*_test.go` with `//go:build integration`, plus files under `test/` — integration coverage run by `make integration-test`
- `test/e2e/*` with `//go:build e2e` — full binary coverage run by `make e2e-test`

## Concurrency Test Expectations

- Use deterministic coordination for concurrent-process tests (file-based ready/release markers).
- Do **not** rely on sleep-only timing to trigger contention.
- For CLI/process-level contention, run the locally built binary (`./aimgr`) or the E2E test-built binary, and always set `AIMGR_REPO_PATH` to a temp repo.
- Validate both:
  - safety outcomes (manifest/source metadata remains valid, no corrupted repo state)
  - user-facing behavior (second process waits or fails with clear lock-acquisition errors under contention)
- Keep contention tests isolated to temp repos and temp source fixtures.

For repo lock primitives, include focused coverage for:

- shared/shared behavior (concurrent readers succeed on POSIX)
- shared/exclusive contention both directions (reader blocks writer, writer blocks reader)
- in-process same-path non-reentrancy (second acquisition fails with `ErrNonReentrantLock`)
- same-path transition attempts (read→write and write→read fail with `ErrNonReentrantLock`)
- explicit Windows behavior verification for the read/write API (including exclusive-only fallback when applicable)

## Persistence / Atomic-Write Expectations

Repo-managed state persistence uses atomic replacement (not plain in-place
`os.WriteFile`) for:

- `ai.repo.yaml`
- `.metadata/sources.json`
- resource metadata files under `.metadata/...`
- `.workspace/.cache-metadata.json`

When adding/changing tests around these files, keep expectations aligned with
the implemented write sequence:

1. temp file in same directory
2. write + file `fsync`
3. rename replacement
4. parent-directory `fsync` where supported

Locking and atomic writes are complementary: tests that exercise concurrent
mutations should still validate lock-mediated serialization, while crash-safety
behavior focuses on atomic replacement semantics.

## Full Guide

Read **[docs/contributor-guide/testing.md](contributor-guide/testing.md)** for test-authoring patterns, example structures, and troubleshooting.
