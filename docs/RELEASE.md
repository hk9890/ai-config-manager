# Release Process

## Creating a New Release

1. **Update version** in code if needed
2. **Commit all changes**
3. **Create and push tag**:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```
4. **GitHub Actions automatically**:
   - Builds binaries for all platforms
   - Creates GitHub Release
   - Uploads binaries and checksums
   - Generates changelog

## Version Numbering

Follow semantic versioning (MAJOR.MINOR.PATCH):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

## Testing Releases Locally

```bash
# Test GoReleaser without publishing
goreleaser release --snapshot --clean

# Check generated artifacts
ls dist/
```

## Supported Platforms

The release process builds binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Version Injection

Version information is injected at build time via ldflags:
- `Version`: Git tag (e.g., v0.1.0)
- `GitCommit`: Short commit hash
- `BuildDate`: Build timestamp

These values are injected into `pkg/version/version.go`.

## Release Workflow

1. **Developer** creates and pushes a version tag
2. **GitHub Actions** detects the tag
3. **GoReleaser** runs:
   - Builds binaries for all platforms
   - Creates archives (tar.gz/zip)
   - Generates checksums
   - Creates GitHub Release with changelog
4. **Users** can download from Releases page

## Manual Release (If Needed)

If you need to create a release manually:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Create release (requires GITHUB_TOKEN)
export GITHUB_TOKEN="your-token"
goreleaser release --clean
```

## Troubleshooting

### Release Failed

1. Check GitHub Actions logs
2. Verify tag format matches `v*` pattern
3. Ensure all tests pass
4. Check GITHUB_TOKEN permissions

### Wrong Version in Binary

Verify ldflags paths match `pkg/version/version.go`:
```bash
go build -ldflags "-X github.com/hans-m-leitner/ai-config-manager/pkg/version.Version=test" .
./ai-repo --version
```
