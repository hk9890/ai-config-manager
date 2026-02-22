# Release Process

This document provides step-by-step instructions for releasing ai-config-manager.

## Build

```bash
make build
```

Output: `aimgr` binary in project root.

## Tests

```bash
make test
```

This runs: `go vet`, unit tests, and integration tests. All must pass.

## Version Files

**None**. Version is managed via Git tags only.

The version in `pkg/version/version.go` is a placeholder. The actual version is injected at build time via ldflags by GoReleaser:

```
-X github.com/hk9890/ai-config-manager/pkg/version.Version={{.Version}}
-X github.com/hk9890/ai-config-manager/pkg/version.GitCommit={{.ShortCommit}}
-X github.com/hk9890/ai-config-manager/pkg/version.BuildDate={{.Date}}
```

**Do not manually update version files.**

## Versioning

Follow Semantic Versioning:

- **Major**: Breaking changes to CLI or resource formats
- **Minor**: New features (new commands, flags, resource types)
- **Patch**: Bug fixes only

Analyze commits to determine bump:

```bash
git log v2.2.0..HEAD --pretty=format:"%s" | grep -E "^(feat|fix|BREAKING):"
```

Rules:
- `feat:` → minor
- `fix:` only → patch  
- `BREAKING CHANGE:` → major

## Pre-Release Checklist

- [ ] Tests pass: `make test`
- [ ] Build succeeds: `make build`
- [ ] No uncommitted changes: `git status`
- [ ] **CHANGELOG.md updated** with new version entry
- [ ] CI green: `gh run list --limit 1`

## Release Steps

This project uses **tag-triggered GitHub Actions**. Do NOT use `gh workflow run`.

### 1. Update CHANGELOG.md

Add entry for new version:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Fixed
- Bug fixes

### Changed
- Breaking changes
```

### 2. Commit and Push

```bash
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for vX.Y.Z"
git push origin main
```

### 3. Create and Push Tag

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

This triggers the GitHub Actions release workflow.

### 4. Monitor Workflow

```bash
gh run watch
```

The workflow builds binaries, creates archives, generates checksums, and publishes the GitHub release.

### 5. Enhance Release Notes

GitHub auto-generates basic notes. Enhance them:

```bash
cat > /tmp/release-notes.md << 'NOTES'
## Highlights

- **Feature** - Description
- **Tests** - Improvements
- **Config** - Changes

## What's Changed

### Added
- Item (commit)

### Fixed
- Item (commit)

### Changed
- Item (commit)

## Full Changelog
https://github.com/hk9890/ai-config-manager/compare/vOLD...vNEW
NOTES

gh release edit vX.Y.Z --notes-file /tmp/release-notes.md
```

### 6. Verify

```bash
gh release view vX.Y.Z
```

## Post-Release

1. Verify release: https://github.com/hk9890/ai-config-manager/releases
2. Test binary download works
3. Monitor issues

## Rollback

```bash
gh release delete vX.Y.Z -y
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z
```

Then fix and re-release with incremented version.

## Related

- [CHANGELOG.md](../../CHANGELOG.md) - Version history
- [GitHub Releases](https://github.com/hk9890/ai-config-manager/releases) - Official releases
