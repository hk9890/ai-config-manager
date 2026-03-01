# aimgr (ai-config-manager)

CLI tool (Go 1.25.6) for managing AI resources (commands, skills, agents, packages) across AI coding tools. Centralized repository with symlink-based installation.

**Tech Stack**: Go 1.25.6, Cobra, Viper, XDG directories, GoReleaser

## Coding

Read `docs/CODING.md` for build commands, project structure, code conventions, and **critical safety rules**.

Read `CONTRIBUTING.md` for contribution workflow and commit conventions.

## Testing

Read `docs/TESTING.md` for test commands, isolation requirements, and patterns.

## Releases

Load the **github-releases** skill for release workflow. Read `docs/RELEASING.md` for process details.

## Monitoring

Load the **observability-triage** skill for log analysis and issue triage.

## Landing the Plane (Session Completion)

**When ending a work session**, complete ALL steps:

1. **File issues for remaining work** -- Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) -- Tests, linters, builds
3. **Update issue status** -- Close finished work, update in-progress items
4. **PUSH TO REMOTE** -- This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Verify** -- All changes committed AND pushed

**CRITICAL**: Work is NOT complete until `git push` succeeds.
