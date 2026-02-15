# Contributing to aimgr

Thank you for your interest in contributing to aimgr! This guide will help you get started quickly.

## Quick Start

### Prerequisites

- **Go 1.25.6+** (or use [mise](https://mise.jdx.dev/) for automatic version management)
- **Make** (build automation)
- **Git** (version control)

### Setup

```bash
# Clone the repository
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager

# Build the binary
make build

# Run tests
make test
```

### Verify Your Setup

```bash
# Check Go version
go version  # Should show 1.25.6 or higher

# Build and test
make build
make test   # Should pass all tests
```

## Development Workflow

### 1. Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/hk9890/ai-config-manager.git
   ```

### 2. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Make Changes

1. Write your code
2. Add tests for new functionality
3. Follow the [Code Style Guide](docs/contributor-guide/code-style.md)
4. Run tests frequently: `make test`

### 4. Commit Your Changes

Follow conventional commit format (see below)

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub.

## Submitting Changes

### Before Submitting

- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make vet`)
- [ ] New code has tests
- [ ] Documentation updated (if user-facing change)
- [ ] Commit messages are clear and descriptive

### Commit Message Format

Use conventional commits:

```
type(scope): short description

Longer description if needed, explaining what and why.

Fixes #issue-number
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, missing semicolons, etc.
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance, dependencies, etc.

**Examples:**

```
feat(repo): add bulk import support for plugins

Add ability to import multiple commands and skills from Claude plugins
in a single operation.

Fixes #42
```

```
fix(install): handle symlink creation on Windows

Use junction points instead of symlinks for Windows compatibility.
```

### Pull Request Checklist

1. **Create PR** from your feature branch to `main`
2. **Fill out description** with:
   - What changes were made
   - Why the changes were necessary
   - How to test the changes
3. **Link related issues** using keywords (e.g., "Fixes #42")
4. **Wait for CI to pass** - All checks must pass
5. **Address review feedback** promptly and respectfully
6. **Maintainer will merge** when approved

## Detailed Guides

For in-depth information, see the comprehensive guides in `docs/contributor-guide/`:

### Writing Code

→ **[Code Style Guide](docs/contributor-guide/code-style.md)**
- Naming conventions (files, packages, types)
- Import organization (3 groups: stdlib, external, internal)
- Error handling (always wrap with `%w`)
- Symlink handling (CRITICAL: use `os.Stat()` not `entry.IsDir()`)
- File operations and best practices

→ **[Architecture Guide](docs/contributor-guide/architecture.md)**
- System overview and components
- Package structure (17 packages)
- Architecture rules (5 critical rules)
- Data flows (import, install, sync)
- Key concepts and patterns

### Testing

→ **[Testing Guide](docs/contributor-guide/testing.md)**
- Test types (unit, integration, E2E)
- Test isolation with `t.TempDir()`
- Table-driven test patterns
- Running specific tests
- Troubleshooting test failures

### Setup & Tools

→ **[Development Environment](docs/contributor-guide/development-environment.md)**
- mise setup for Go version management
- IDE configuration (VS Code, GoLand)
- Build tools and workflow
- CI/CD consistency

### Releasing

→ **[Release Process](docs/contributor-guide/release-process.md)**
- Version management
- Release workflow
- See [docs/RELEASING.md](docs/RELEASING.md) for complete guide

### Documentation Overview

→ **[Contributor Guide README](docs/contributor-guide/README.md)**
- Overview of all contributor documentation
- Quick navigation to all guides

## Common Tasks

### Running Tests

```bash
make test              # All tests (vet → unit → integration → test/...)
make unit-test         # Fast unit tests only
make integration-test  # Integration tests (slower)
make e2e-test          # End-to-end tests (slowest)
```

### Code Quality

```bash
make fmt               # Format all Go code
make vet               # Run go vet (static analysis)
make build             # Build binary
```

### Building

```bash
make build             # Build to ./aimgr
make install           # Build and install to ~/bin
make clean             # Remove build artifacts
```

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)
- **Documentation**: 
  - User docs: [README.md](README.md)
  - AI agent guide: [AGENTS.md](AGENTS.md)
  - Contributor docs: [docs/contributor-guide/](docs/contributor-guide/)

---

Thank you for contributing to aimgr! 🎉
