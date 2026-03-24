# Monitoring Configuration

This project does not currently have a production monitoring backend wired into the repository. For now, monitoring is log-first and focuses on local tool/runtime health for `aimgr`, beads (`bd`), and OpenCode/opencode-coder.

## Available Data

### Local tool logs

- `.beads/daemon.log`
  - Beads daemon lifecycle, import/export activity, sync failures, and warnings.
- `.beads/dolt-server.log`
  - Dolt SQL server startup, connections, and database-level errors used by beads.
- `.beads/daemon-error`
  - Last fatal daemon/bootstrap issue. Treat this as high-signal when present.
- `logs/operations.log`
  - Aimgr structured repo-operation log when aimgr logging is active.
- `debug.log`
  - Historical local debug output from related tooling in this workspace.
- `~/.local/share/opencode/log/*.log`
  - OpenCode runtime logs, plugin startup, config warnings, and session errors.

### Useful commands

- `./aimgr --version`
- `bd version`
- `bd doctor`
- `git status --short --branch`

Use the local `./aimgr` binary for validation, not a globally installed `aimgr` from PATH.

## What To Look For

### Critical

- Beads repository/database mismatch warnings
- `bd doctor` hangs or never returns
- Repeated database errors in `.beads/dolt-server.log`
- Machine-readable command output being polluted by warnings/noise

### Important

- Structured logging corruption such as `!BADKEY`
- Repeated import/export validation failures in beads logs
- Plugin load failures or missing command/skill files in OpenCode logs
- Unexpected stderr output on otherwise successful commands

### Usually Safe To Ignore

- Historical warnings from unrelated repositories in shared OpenCode logs
- One-off config deprecation warnings unless they are blocking behavior
- Old failures that cannot be reproduced in the current checkout

## Context And Meaning

- `.beads/daemon-error` is high priority because it can indicate tracker-state corruption risk.
- `!BADKEY` in slog-style logs usually means formatted logging is being used incorrectly and the resulting logs are less trustworthy for diagnosis.
- `fatal: couldn't find remote ref main` in old daemon logs may be historical; confirm it still reproduces before treating it as an active bug.
- Empty `logs/operations.log` is not automatically a bug; it may simply mean no aimgr operations were run with file logging enabled.

## Current Triage Approach

When asked to analyze monitoring data for this repo:

1. Check `.beads/daemon-error`
2. Review `.beads/daemon.log` and `.beads/dolt-server.log`
3. Review current OpenCode logs in `~/.local/share/opencode/log/`
4. Run lightweight health commands (`bd version`, `bd doctor`, `./aimgr --version`)
5. Group findings by root cause and create/update beads issues for clear tooling bugs
