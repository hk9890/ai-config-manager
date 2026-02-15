# Development Environment Setup

This document describes how to set up your local development environment for ai-config-manager.

## Prerequisites

### Required Tools

1. **Go 1.25.6** (managed via mise - see below)
2. **Git** (version control)
3. **Make** (build automation)
4. **mise** (version manager) - **Recommended**

### Installing mise (Recommended)

mise ensures all developers use the same Go version (1.25.6) as CI/CD.

**macOS/Linux:**
```bash
curl https://mise.jdx.dev/install.sh | sh
```

**Alternative methods:**
- Homebrew: `brew install mise`
- See: https://mise.jdx.dev/getting-started.html

**Windows:**
```powershell
# PowerShell
irm https://mise.jdx.dev/install.ps1 | iex
```

### Activate mise

Add to your shell configuration:

**Bash** (`~/.bashrc`):
```bash
eval "$(mise activate bash)"
```

**Zsh** (`~/.zshrc`):
```zsh
eval "$(mise activate zsh)"
```

**Fish** (`~/.config/fish/config.fish`):
```fish
mise activate fish | source
```

Restart your shell or run: `source ~/.bashrc` (or equivalent)

## Project Setup

### 1. Clone the Repository

```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
```

### 2. Install Go via mise

When you enter the project directory, mise will automatically detect `.mise.toml` and prompt you to install Go 1.25.6:

```bash
cd ai-config-manager
# mise will show: "mise Go 1.25.6 is not installed. Install? [y/N]"
# Press 'y' to install
```

Or manually install:
```bash
mise install
```

### 3. Verify Setup

```bash
# Check Go version (should show 1.25.6)
go version

# Check CGO is disabled (should show CGO_ENABLED="0")
mise env | grep CGO_ENABLED

# Build the project
make build

# Run tests
make test
```

## Development Workflow

### Common Commands

```bash
# Build binary
make build

# Run all tests (vet → unit → integration)
make test

# Run only unit tests (fast)
make unit-test

# Run only integration tests
make integration-test

# Run E2E tests
make e2e-test

# Format code
make fmt

# Run linter (go vet)
make vet

# Install binary to ~/bin
make install

# Clean build artifacts
make clean

# See all available commands
make help
```

### Test Order

Tests run in this order (matching CI):
1. **vet** - Static analysis (catches syntax errors)
2. **unit-test** - Fast tests, no external dependencies
3. **integration-test** - Slower tests, uses git/network

### Building for Different Platforms

```bash
# Linux (default)
make build

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 make build

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 make build
```

## Environment Variables

The following environment variables are automatically set by mise (via `.env`):

- `CGO_ENABLED=0` - Disable CGO for static binary compilation

You can override these temporarily:
```bash
CGO_ENABLED=1 go build ...
```

## CI/CD Consistency

Our setup ensures **perfect consistency** between local and CI:

| Aspect | Local (mise) | CI (GitHub Actions) |
|--------|--------------|---------------------|
| Go Version | 1.25.6 | 1.25.6 |
| CGO | Disabled (0) | Disabled (0) |
| Test Order | vet → unit → int | vet → unit → int |
| Linter | go vet | go vet |
| Build Flags | Same ldflags | Same ldflags |

This means:
- ✅ If tests pass locally, they'll pass in CI
- ✅ Binaries built locally match CI builds
- ✅ No "works on my machine" problems

## Troubleshooting

### mise Not Found

If `mise` command is not found after installation:
1. Restart your terminal
2. Verify mise is in PATH: `which mise`
3. Re-run the activation command for your shell

### Wrong Go Version

If `go version` shows the wrong version:
```bash
# Reload mise
mise install

# Check mise status
mise doctor

# Verify Go is managed by mise
mise which go
```

### CGO_ENABLED Not Set

If CGO_ENABLED is not set:
```bash
# Check mise env
mise env | grep CGO

# Reload shell
cd .. && cd ai-config-manager
```

### Tests Fail Locally But Pass in CI

1. Ensure you're using Go 1.25.6: `go version`
2. Clean and rebuild: `make clean && make build`
3. Run tests in the same order as CI: `make test`
4. Check for uncommitted changes: `git status`

### Make Commands Don't Work

Ensure you have `make` installed:
```bash
# Linux
sudo apt-get install build-essential

# macOS
xcode-select --install

# Windows
# Install via chocolatey
choco install make
```

## Without mise (Manual Setup)

If you prefer not to use mise, install Go 1.25.6 manually:

1. Download Go 1.25.6: https://go.dev/dl/
2. Set CGO_ENABLED manually:
   ```bash
   export CGO_ENABLED=0
   ```
3. Verify: `go version` (should show 1.25.6)

**Note:** Manual setup is more error-prone. We strongly recommend using mise.

## IDE Setup

### VS Code

Recommended extensions:
- **Go** (golang.go)
- **mise** (jdxcode.mise)

Settings (`.vscode/settings.json`):
```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package"
}
```

### GoLand / IntelliJ IDEA

1. Settings → Go → GOROOT
2. Select mise-managed Go (usually `~/.local/share/mise/installs/go/1.25.6`)
3. Enable "Go Modules"

## Next Steps

- Read [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines
- Check [release-process.md](release-process.md) for release workflow
- See [AGENTS.md](../../AGENTS.md) for AI agent guidelines

## Resources

- mise documentation: https://mise.jdx.dev/
- Go documentation: https://go.dev/doc/
- Project repository: https://github.com/hk9890/ai-config-manager
