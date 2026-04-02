# Pull Request Guide

Short reference for branch workflow, pull request expectations, and review follow-up in the `aimgr` repository.

## Branch Workflow

- Start from an up-to-date `main` branch.
- Create a descriptive feature branch such as `feature/your-feature-name`.
- Keep each branch focused on one logical change.
- Push your branch to your fork or remote before opening a PR.

## Before Opening a PR

- Run `make fmt` to format Go code.
- Run `make vet` for static analysis.
- Run `make test` to cover the normal contributor test flow.
- Update documentation when the change affects users, contributors, or AI agents.

## PR Description

Open the PR from your feature branch to `main` and include:

1. What changed
2. Why the change was needed
3. How to test it

Link related issues with GitHub keywords such as `Fixes #42` when applicable.

## Review Follow-up

- Wait for CI to pass before asking for merge.
- Address review feedback promptly and respectfully.
- Maintainers merge after approval.

## Related Docs

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution workflow and commit message format
- [docs/CODING.md](CODING.md) - Build commands and code conventions
- [docs/TESTING.md](TESTING.md) - Test commands and isolation rules
