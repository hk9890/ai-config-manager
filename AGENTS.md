# aimgr (ai-config-manager)

CLI tool (Go 1.25.6) for managing AI resources (commands, skills, agents, packages) across AI coding tools. Centralized repository with symlink-based installation.

**Tech Stack**: Go 1.25.6, Cobra, Viper, XDG directories, GoReleaser

## Coding

Read `docs/CODING.md` for build commands, project structure, code conventions, and safety rules.
Read `CONTRIBUTING.md` for contribution workflow and commit conventions.

## Testing

Read `docs/TESTING.md` for test commands, isolation requirements, and patterns.

## Releases

Start release work by loading the **github-releases** skill.
Read `docs/RELEASING.md` for repo-specific release details.

## Monitoring

Load the **observability-triage** skill for log analysis and issue triage.
Read `docs/MONITORING.md` for available signals and local triage workflow.

## Pull Requests

Read `docs/PULL-REQUESTS.md` for branch workflow, PR expectations, and review follow-up.

## Issue Tracking

This repo uses **bd (beads)** for task tracking.
Read `.beads/README.md` for tracker basics.

## Landing the Plane (Session Completion)

**When ending a work session**, complete ALL steps:

1. **File issues for remaining work** -- Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) -- Tests, linters, builds
3. **Update issue status** -- Close finished work, update in-progress items
4. **PUSH TO REMOTE** -- This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** -- Clear stashes, prune remote branches
6. **Verify** -- All changes committed AND pushed
7. **Hand off** -- Provide context for next session

**CRITICAL**: Work is NOT complete until `git push` succeeds.
