# Contributor Guide

Welcome to the **aimgr** contributor guide! This section contains documentation for developers working on the ai-config-manager project.

## Getting Started

### Prerequisites

- **Go 1.25 or higher** (check with `go version`)
- **Make** (optional, but recommended)
- **Git**

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager

# Build the binary
make build

# Run all checks (format, vet, test, build)
make all
```

For detailed setup instructions, code style guidelines, and development workflow, see the main [CONTRIBUTING.md](../../CONTRIBUTING.md) file in the repository root.

## Contributor Documentation

### Core Guides

- **[CONTRIBUTING.md](../../CONTRIBUTING.md)** - Main contributor guide
  - Development environment setup
  - Project architecture overview
  - Development workflow
  - Code style guidelines
  - Testing guidelines
  - Submitting changes

- **[Release Process](release-process.md)** - How to create and publish releases
  - Version numbering
  - Creating releases
  - Testing releases locally
  - Release workflow
  - Troubleshooting

### Architecture Documentation

- **[Architecture Overview](../architecture/)** - Detailed architecture documentation
  - System design
  - Component interactions
  - Design patterns

### Additional Resources

- **[AGENTS.md](../../AGENTS.md)** - Quick reference for AI coding agents
  - Build & test commands
  - Common patterns
  - Resource formats
  - Code style quick reference

## Essential Commands

### Building

```bash
make build      # Build binary
make install    # Build and install to ~/bin
make all        # Run all checks and build
```

### Testing

```bash
make test              # Run all tests (unit + integration + vet)
make unit-test         # Run only unit tests
make integration-test  # Run only integration tests
```

### Linting & Formatting

```bash
make fmt        # Format all Go code
make vet        # Run go vet
make deps       # Download dependencies
make clean      # Clean build artifacts
```

## Code Style Quick Reference

### Import Organization

Three groups with blank lines:

```go
import (
    "fmt"           // 1. Standard library
    "os"

    "github.com/spf13/cobra"  // 2. External dependencies

    "github.com/hk9890/ai-config-manager/pkg/resource"  // 3. Internal
)
```

### Naming Conventions

- **Files**: `lowercase_with_underscores.go`
- **Packages**: Short, lowercase, single word (`resource`, `config`)
- **Types**: PascalCase (`ResourceType`, `CommandResource`)
- **Functions**: PascalCase (exported), camelCase (unexported)
- **Variables**: camelCase (`repoPath`, `skillsDir`)

### Error Handling

Always wrap errors with context:

```go
if err != nil {
    return fmt.Errorf("failed to load command: %w", err)
}
```

## Before Submitting Changes

- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter warnings (`make vet`)
- [ ] New code has tests
- [ ] Documentation updated (if user-facing change)
- [ ] Commit messages are clear and descriptive

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)
- **Main Documentation**: See [README.md](../../README.md) for user documentation

---

Thank you for contributing to aimgr! 🎉
