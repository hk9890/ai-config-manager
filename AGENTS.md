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

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
