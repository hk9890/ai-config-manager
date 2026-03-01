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

## Quick Commands

```bash
# Build
make build      # Build binary to ./aimgr
make install    # Build and install to ~/bin

# Test
make test             # All tests (vet -> unit -> integration)
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
