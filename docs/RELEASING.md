# Release Process

Quick reference for releasing ai-config-manager.

## Overview

Releases are **tag-triggered** via GitHub Actions + GoReleaser. Version is injected at build time -- do NOT manually update version files.

## Steps

1. Ensure tests pass: `make test`
2. Update `CHANGELOG.md`
3. Create and push a Git tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z" && git push origin vX.Y.Z`
4. Monitor: `gh run watch`

## Full Guide

Read **[docs/contributor-guide/release-process.md](contributor-guide/release-process.md)** for complete steps, versioning rules, and rollback procedures.

Load the **github-releases** skill for an automated release workflow.
