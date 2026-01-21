# Release v0.3.1 - Module Path Fix

## 🐛 Critical Bug Fix

This patch release fixes the Go module path to match the actual GitHub repository location.

### What Changed

**Module Path Correction**
- ❌ Old: `github.com/hans-m-leitner/ai-config-manager`
- ✅ New: `github.com/hk9890/ai-config-manager`

### Files Updated
- `go.mod` - Module declaration
- 33 `.go` files - Import statements
- `Makefile` - Build ldflags
- `.github/workflows/build.yml` - GitHub Actions workflow
- `.goreleaser.yaml` - Release configuration

### Why This Matters

Users can now correctly install via:

```bash
go install github.com/hk9890/ai-config-manager@latest
```

CI/CD builds will now:
- Use correct module paths
- Embed proper version information
- Work with GoReleaser properly

### Impact

- ✅ No functional changes to the tool
- ✅ No breaking changes to features
- ✅ Purely infrastructure/path corrections
- ✅ All v0.3.0 features remain unchanged

### Upgrading from v0.3.0

If you installed v0.3.0:

```bash
# Download new binary from releases
# Or rebuild from source
make install
```

No configuration or data migration needed.

## 📦 Installation

### Via Go Install
```bash
go install github.com/hk9890/ai-config-manager@v0.3.1
```

### Via Binary Download
Download for your platform from the [releases page](https://github.com/hk9890/ai-config-manager/releases/tag/v0.3.1).

### Via Source
```bash
git clone https://github.com/hk9890/ai-config-manager.git
cd ai-config-manager
git checkout v0.3.1
make install
```

## 🔗 Related Issues

- Fixes #ai-config-manager-091: Fix module path in go.mod
- Fixes #ai-config-manager-w4k: Update CI/CD config references

## 📊 Full Changelog

**v0.3.0...v0.3.1**
- dd67924 Fix remaining hans-m-leitner references in CI/CD configs
- 5dde803 Fix module path to match actual GitHub repository
