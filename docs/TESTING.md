# Testing Guide

Quick reference for testing ai-config-manager.

## Commands

```bash
make test             # All tests (vet -> unit -> integration)
make unit-test        # Fast unit tests (<5s)
make integration-test # Network-dependent tests (~30s)
make e2e-test         # Full CLI tests (~1-2min)
```

## Critical Rules

- **ALWAYS** use `t.TempDir()` for temporary directories
- **ALWAYS** use `repo.NewManagerWithPath(tmpDir)` in tests -- NEVER `NewManager()`
- **NEVER** write to `~/.local/share/ai-config/` in tests

## Full Guide

Read **[docs/contributor-guide/testing.md](contributor-guide/testing.md)** for test types, isolation patterns, table-driven tests, and troubleshooting.
