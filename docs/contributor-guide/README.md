# Contributor Guide

Documentation for developers working on ai-config-manager.

## Quick Start

```bash
# Clone and setup
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager

# Build and test
make build
make test
```

For detailed setup, see [Development Environment](development-environment.md).

## Core Documentation

### Getting Started

- **[Development Environment](development-environment.md)** - Setup guide (Go, mise, tools)
- **[Code Style Guide](code-style.md)** - Naming, imports, error handling, symlinks
- **[Testing Guide](testing.md)** - Testing approach, best practices, troubleshooting
- **[Release Process](release-process.md)** - Version numbering, creating releases

### Architecture

- **[Architecture Guide](architecture.md)** - System overview, design rules, data flows
  - Architecture rules (Git workspace, XDG, error handling, symlinks)
  - Package structure (17 packages)
  - Key concepts and patterns

### Additional Resources

- **[CONTRIBUTING.md](../../CONTRIBUTING.md)** - Quick start and workflow guide
- **[AGENTS.md](../../AGENTS.md)** - Quick reference for AI agents

## Essential Commands

### Build & Test

```bash
make build      # Build binary
make install    # Install to ~/bin
make test       # All tests (vet → unit → integration → e2e)
```

### Code Quality

```bash
make fmt        # Format Go code
make vet        # Run static analysis
```

## Code Style Quick Reference

### Import Organization

```go
import (
    "fmt"           // 1. Standard library

    "github.com/spf13/cobra"  // 2. External dependencies

    "github.com/hk9890/ai-config-manager/pkg/resource"  // 3. Internal
)
```

### Naming

- Files: `lowercase_with_underscores.go`
- Types: `PascalCase`
- Functions: `PascalCase` (exported), `camelCase` (unexported)
- Resources: `lowercase-with-hyphens` (1-64 chars)

### Error Handling

```go
if err != nil {
    return fmt.Errorf("failed to load: %w", err)
}
```

### Symlink Handling (CRITICAL)

```go
// ✅ CORRECT: Use os.Stat() to follow symlinks
entries, _ := os.ReadDir(dir)
for _, entry := range entries {
    path := filepath.Join(dir, entry.Name())
    info, err := os.Stat(path)  // os.Stat follows symlinks
    if err != nil || !info.IsDir() { continue }
    processDirectory(path)
}
```

See [Code Style Guide](code-style.md) for complete details.

## Before Submitting

- [ ] Tests pass: `make test`
- [ ] Code formatted: `make fmt`
- [ ] No warnings: `make vet`
- [ ] Tests added for new functionality
- [ ] Documentation updated (if user-facing)

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/hk9890/ai-config-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/hk9890/ai-config-manager/discussions)

---

Thank you for contributing to aimgr! 🎉
