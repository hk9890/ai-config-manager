# GitHub Actions & Release Process Proposal

## Overview

This proposal outlines setting up automated CI/CD and release process for ai-repo, based on the proven approach used by [dtctl](https://github.com/dynatrace-oss/dtctl).

---

## 🎯 Goals

1. **Automated Testing** - Run tests on every push and PR
2. **Multi-Platform Builds** - Build for Linux, macOS, Windows (amd64 & arm64)
3. **Automated Releases** - Create releases with binaries, checksums, and changelog
4. **Version Management** - Inject version info at build time
5. **Professional Setup** - Industry-standard CI/CD workflow

---

## 📦 Components

### 1. GoReleaser

**Purpose**: Automate the release process

**Features**:
- Cross-platform compilation
- Archive creation (tar.gz, zip)
- Checksum generation
- GitHub Release creation
- Automatic changelog

**Config File**: `.goreleaser.yaml`

```yaml
version: 2
project_name: ai-repo

builds:
  - goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X pkg/version.Version={{.Version}}
      - -X pkg/version.GitCommit={{.ShortCommit}}
      - -X pkg/version.BuildDate={{.Date}}
```

---

### 2. GitHub Actions Workflows

#### **Build Workflow** (`.github/workflows/build.yml`)

**Triggers**: Every push to `main`, every PR

**Jobs**:
1. **Test** - Run `go test ./...` and `go vet ./...`
2. **Build** - Build for all platforms (6 combinations)

**Matrix Strategy**:
- linux/amd64
- linux/arm64
- darwin/amd64 (Intel Mac)
- darwin/arm64 (Apple Silicon)
- windows/amd64
- windows/arm64 (optional)

**Artifacts**: Uploaded for 30 days for testing

---

#### **Release Workflow** (`.github/workflows/release.yml`)

**Triggers**: Git tags matching `v*` (e.g., `v0.1.0`)

**Process**:
1. Checkout code
2. Setup Go
3. Run GoReleaser
4. Create GitHub Release
5. Upload binaries & checksums
6. Generate changelog

**Automation**: Fully automated, no manual steps

---

#### **Lint Workflow** (`.github/workflows/lint.yml`) - Optional

**Triggers**: Every push to `main`, every PR

**Purpose**: Code quality checks with `golangci-lint`

---

## 🔄 Release Process

### Developer Workflow

```bash
# 1. Make changes and commit
git add .
git commit -m "Add new feature"

# 2. Create and push tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# 3. GitHub Actions automatically:
#    - Builds all platform binaries
#    - Creates GitHub Release
#    - Uploads binaries with checksums
#    - Generates changelog
```

### What Users Get

After release, users can:

```bash
# Download specific platform
curl -L https://github.com/USER/ai-config-manager/releases/latest/download/ai-repo_0.1.0_linux_amd64.tar.gz | tar xz

# Or visit releases page
https://github.com/USER/ai-config-manager/releases
```

---

## 📊 Comparison with dtctl

| Feature | dtctl | ai-repo (Proposed) |
|---------|-------|-------------------|
| **Build Tool** | GoReleaser | ✅ GoReleaser |
| **CI Platform** | GitHub Actions | ✅ GitHub Actions |
| **Platforms** | Linux, macOS, Windows | ✅ Same (6 combinations) |
| **Version Injection** | ldflags | ✅ ldflags |
| **Checksums** | Yes | ✅ Yes |
| **Changelog** | Auto-generated | ✅ Auto-generated |
| **Test Workflow** | Included | ✅ Included |
| **Lint Workflow** | Optional | ✅ Optional |

---

## 📁 Files to Create

```
ai-config-manager/
├── .goreleaser.yaml              # GoReleaser config
├── .github/
│   └── workflows/
│       ├── build.yml             # Build on push/PR
│       ├── release.yml           # Release on tags
│       └── lint.yml              # Code quality (optional)
└── docs/
    └── RELEASE.md                # Release process docs
```

---

## 🚀 Benefits

### For Developers

- ✅ **Automated builds** on every commit
- ✅ **Quick feedback** from CI tests
- ✅ **Easy releases** - just push a tag
- ✅ **No manual compilation** for different platforms

### For Users

- ✅ **Pre-built binaries** for all major platforms
- ✅ **Easy installation** with curl/wget
- ✅ **Checksums** for verification
- ✅ **Proper versioning** - know what you're running
- ✅ **Release notes** - understand what changed

---

## 📝 Installation Instructions (for README)

### Quick Install (Linux/macOS)

```bash
# Linux (amd64)
curl -L https://github.com/USER/ai-config-manager/releases/latest/download/ai-repo_VERSION_linux_amd64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/USER/ai-config-manager/releases/latest/download/ai-repo_VERSION_darwin_arm64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/USER/ai-config-manager/releases/latest/download/ai-repo_VERSION_darwin_amd64.tar.gz | tar xz
sudo mv ai-repo /usr/local/bin/
```

### Windows

```powershell
# Download from releases page
https://github.com/USER/ai-config-manager/releases
```

---

## ⚙️ Configuration Requirements

### GitHub Repository Settings

**Permissions**: Enable "Read and write permissions" for GitHub Actions
- Settings → Actions → General → Workflow permissions

**Branch Protection** (Optional):
- Require PR reviews
- Require status checks (build, test) before merge

### No Secrets Required

GitHub automatically provides:
- `GITHUB_TOKEN` for creating releases
- Sufficient permissions when properly configured

---

## 🧪 Testing the Setup

### Before First Release

1. **Test GoReleaser locally**:
   ```bash
   # Install goreleaser
   go install github.com/goreleaser/goreleaser@latest
   
   # Test release (doesn't publish)
   goreleaser release --snapshot --clean
   
   # Check output
   ls dist/
   ```

2. **Test build workflow**:
   ```bash
   # Push to branch
   git push origin feature-branch
   
   # Check GitHub Actions tab
   # Verify all builds succeed
   ```

3. **Test release workflow**:
   ```bash
   # Create test tag
   git tag v0.0.1-test
   git push origin v0.0.1-test
   
   # Check GitHub Releases
   # Verify release created with binaries
   ```

---

## 🎯 Rollout Plan

### Phase 1: Setup (Day 1)
- [ ] Create `.goreleaser.yaml`
- [ ] Create `.github/workflows/build.yml`
- [ ] Create `.github/workflows/release.yml`
- [ ] Update `.gitignore`
- [ ] Update `README.md` with installation instructions

### Phase 2: Testing (Day 2)
- [ ] Test GoReleaser locally
- [ ] Push changes, verify build workflow
- [ ] Create test tag, verify release workflow
- [ ] Download and test binaries

### Phase 3: First Release (Day 3)
- [ ] Create `v0.1.0` tag
- [ ] Verify release created
- [ ] Test installation on different platforms
- [ ] Announce release

### Phase 4: Documentation (Day 4)
- [ ] Create `docs/RELEASE.md`
- [ ] Document release process
- [ ] Add badges to README (build status, latest release)
- [ ] Update CONTRIBUTING.md

---

## 💰 Cost

**GitHub Actions**: 
- Free for public repositories
- 2,000 minutes/month for private repos (free tier)
- Our builds ~5 min each = ~400 releases/month possible

**GoReleaser**:
- Free and open source
- No cost

**Total**: $0 for public repos

---

## 🔒 Security Considerations

1. **Token Security**: GitHub automatically provides `GITHUB_TOKEN` - no manual secrets needed
2. **Checksum Verification**: Users can verify downloads with SHA256 checksums
3. **Signed Commits**: Can add GPG signing later if needed
4. **Dependency Scanning**: Can add Dependabot for security updates

---

## 📚 References

- **dtctl**: https://github.com/dynatrace-oss/dtctl
- **GoReleaser**: https://goreleaser.com/
- **GitHub Actions**: https://docs.github.com/en/actions
- **Semantic Versioning**: https://semver.org/

---

## ❓ FAQ

**Q: What if I want to release a beta version?**
A: Use tags like `v0.2.0-beta.1` - GoReleaser handles pre-releases

**Q: Can I release from a branch other than main?**
A: Yes, but not recommended. Use tags from `main` for stability

**Q: How do I rollback a release?**
A: Delete the tag and GitHub release, then create a new release with fixed version

**Q: What about Homebrew/apt/other package managers?**
A: Can be added later with GoReleaser's `brews` and `nfpms` sections

---

## ✅ Decision Checklist

Before implementing, confirm:

- [ ] Repository is public (or have GitHub Actions minutes)
- [ ] Ready to follow semantic versioning
- [ ] Comfortable with automated releases (less control, more speed)
- [ ] Repository has sufficient test coverage
- [ ] Version code (`pkg/version/version.go`) is ready for ldflags injection

---

**Next Step**: Create task ai-config-manager-skq to implement this setup
