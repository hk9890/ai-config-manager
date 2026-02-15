# Development Environment Setup

Local development setup for ai-config-manager.

## Prerequisites

### Required Tools

1. **Go 1.25.6** (managed via mise - see below)
2. **Git** (version control)
3. **Make** (build automation)
4. **mise** (version manager) - **Recommended**

### Installing mise (Recommended)

mise ensures all developers use Go 1.25.6 (same as CI/CD).

**macOS/Linux:**
```bash
curl https://mise.jdx.dev/install.sh | sh
```

**Alternative**: `brew install mise` (macOS)

**Activate mise** - Add to shell config (`~/.bashrc`, `~/.zshrc`, etc.):
```bash
eval "$(mise activate bash)"  # or zsh, fish
```

Restart shell or run: `source ~/.bashrc`

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
```

### 2. Install Go via mise

mise auto-detects `.mise.toml` and prompts to install Go 1.25.6:

```bash
cd ai-config-manager  # Prompts to install Go
# Or manually: mise install
```

### 3. Verify Setup

```bash
# Check Go version (should show 1.25.6)
go version

# Check CGO is disabled
mise env | grep CGO_ENABLED  # Should show CGO_ENABLED="0"

# Build and test
make build
make test
```

## Development Workflow

### Common Commands

```bash
# Build & Install
make build      # Build binary to ./aimgr
make install    # Build and install to ~/bin

# Testing
make test              # All tests (vet → unit → integration → e2e)
make unit-test         # Fast unit tests only
make integration-test  # Integration tests
make e2e-test          # E2E tests

# Code Quality
make fmt        # Format all Go code
make vet        # Run go vet (static analysis)

# Cleanup
make clean      # Remove build artifacts

# Help
make help       # Show all targets
```

### Test Execution Order

Tests run in this order (matching CI):
1. **vet** - Static analysis (catches syntax errors)
2. **unit-test** - Fast tests, no external dependencies
3. **integration-test** - Slower tests, uses git/network
4. **e2e-test** - Full CLI tests with real binary

### Building for Different Platforms

```bash
# Linux (default)
make build

# macOS
GOOS=darwin GOARCH=amd64 make build

# Windows
GOOS=windows GOARCH=amd64 make build
```

## CI/CD Consistency

Setup ensures perfect consistency between local and CI:

| Aspect | Local (mise) | CI (GitHub Actions) |
|--------|--------------|---------------------|
| Go Version | 1.25.6 | 1.25.6 |
| CGO | Disabled (0) | Disabled (0) |
| Test Order | vet → unit → int → e2e | vet → unit → int → e2e |
| Linter | go vet | go vet |

**Result**: If tests pass locally, they'll pass in CI.

## Troubleshooting

### mise Not Found

After installation:
1. Restart terminal
2. Verify mise in PATH: `which mise`
3. Re-run activation for your shell

### Wrong Go Version

```bash
mise install         # Reload mise
mise doctor          # Check mise status
mise which go        # Verify Go managed by mise
```

### Tests Fail Locally But Pass in CI

1. Ensure Go 1.25.6: `go version`
2. Clean and rebuild: `make clean && make build`
3. Run tests in CI order: `make test`
4. Check for uncommitted changes: `git status`

## Without mise (Manual Setup)

If not using mise, install Go 1.25.6 manually:

1. Download from: https://go.dev/dl/
2. Set CGO manually: `export CGO_ENABLED=0`
3. Verify: `go version`

**Note**: Manual setup is error-prone. We recommend mise.

## IDE Setup

### VS Code

**Recommended extensions**:
- **Go** (golang.go)
- **mise** (jdxcode.mise)

**Settings** (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package"
}
```

### GoLand / IntelliJ IDEA

1. Settings → Go → GOROOT
2. Select mise-managed Go (`~/.local/share/mise/installs/go/1.25.6`)
3. Enable "Go Modules"

## Next Steps

- Read [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines
- Check [release-process.md](release-process.md) for release workflow
- See [AGENTS.md](../../AGENTS.md) for AI agent guidelines

## Resources

- mise: https://mise.jdx.dev/
- Go: https://go.dev/doc/
- Repository: https://github.com/hk9890/ai-config-manager
