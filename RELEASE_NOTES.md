# Release v0.3.0 - GitHub Source Support

## 🎉 Major Features

### GitHub Repository Support
Add resources directly from GitHub repositories with automatic discovery:

```bash
# Add a skill from GitHub
ai-repo add skill gh:anthropics/skills

# Add specific skill with subpath
ai-repo add skill gh:anthropics/skills/skills/pdf-processing

# Add with branch/tag
ai-repo add skill gh:anthropics/skills@v1.0.0

# Shorthand (infers gh: prefix)
ai-repo add skill anthropics/skills
```

### Auto-Discovery System
- Searches 13 standard locations for skills
- Automatic command and agent discovery
- Recursive fallback with depth limiting
- Interactive selection for multiple resources

### Source Format Support
- `gh:owner/repo` - GitHub repositories
- `gh:owner/repo/path` - Specific paths in repos
- `gh:owner/repo@branch` - Specific branches/tags
- `local:path` - Local directories (existing)
- `owner/repo` - GitHub shorthand
- HTTP/HTTPS Git URLs
- Git SSH URLs

## 📦 What's New

- **Source Parser** - Parses multiple source formats
- **Git Operations** - Shallow cloning with cleanup
- **Auto-Discovery** - Smart resource discovery
- **Interactive Selection** - Choose from multiple resources
- **Comprehensive Docs** - Full guide and examples

## 🧪 Testing

- 200+ unit tests
- 9 integration tests  
- 22 end-to-end test cases
- Tested with real repositories (anthropics/skills, softaworks/agent-toolkit)
- 100% pass rate

## 🐛 Bug Fixes

- Fixed subpath handling for GitHub sources
- Graceful fallback when exact paths don't exist

## 📚 Documentation

- Updated README with source formats and examples
- New comprehensive guide: docs/github-sources.md (350+ lines)
- Updated CONTRIBUTING.md with architecture details
- Updated command help text

## 🔄 Backward Compatibility

Fully backward compatible - all existing local source functionality maintained.

## 📊 Statistics

- 10 tasks completed
- 1 epic delivered
- 1,500+ lines of new code
- 1,145 lines of documentation
- 88 total issues closed

## 🙏 Acknowledgments

Inspired by Vercel's add-skill tool and the agentskills.io ecosystem.

## 📥 Installation

```bash
# Using go install
go install github.com/hans-m-leitner/ai-config-manager@v0.3.0

# Or download pre-built binaries from the release page
```

## 🔗 Links

- [GitHub Repository](https://github.com/hans-m-leitner/ai-config-manager)
- [Full Documentation](https://github.com/hans-m-leitner/ai-config-manager/blob/main/README.md)
- [GitHub Sources Guide](https://github.com/hans-m-leitner/ai-config-manager/blob/main/docs/github-sources.md)
- [Contributing Guide](https://github.com/hans-m-leitner/ai-config-manager/blob/main/CONTRIBUTING.md)
