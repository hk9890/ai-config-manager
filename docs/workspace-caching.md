# Workspace Caching

This document explains how aimgr's workspace caching system optimizes Git repository operations.

## Overview

Git repositories are cached in the `.workspace/` directory for efficient reuse across all Git operations. This significantly improves performance when working with remote repositories.

## Performance Benefits

- **First operation** (`repo add`, `repo sync`, `repo update`): Full git clone (creates cache)
- **Subsequent operations**: Reuse cached repository (10-50x faster)
- **Automatic cache management** with SHA256 hash-based storage
- **Shared across all resources** from the same source repository

## Commands Using Workspace Cache

### repo add
Adds resources using cached clone:
```bash
aimgr repo add gh:owner/repo
aimgr repo add https://github.com/owner/repo
```

### repo sync
Syncs from configured sources using cached clones:
```bash
aimgr repo sync
aimgr repo sync --format=json
```

### repo update
Updates resources using cached clones (git pull):
```bash
aimgr repo update
aimgr repo update skill/name
```

## Batching Performance

Repository commands automatically batch resources from the same Git repository, cloning each unique source only once.

**Example**: 39 resources from one repository = 1 cached clone reused 39 times

This optimization significantly improves performance for bulk operations.

## Workspace Directory Structure

```
~/.local/share/ai-config/repo/
├── .workspace/                   # Git repository cache
│   ├── <hash-1>/                 # Cached repository 1 (by URL hash)
│   │   ├── .git/
│   │   ├── commands/
│   │   └── skills/
│   ├── <hash-2>/                 # Cached repository 2
│   │   └── ...
│   └── .cache-metadata.json      # Cache metadata (URLs, timestamps, refs)
├── .metadata/                    # Resource metadata
├── commands/                     # Command resources
├── skills/                       # Skill resources
└── agents/                       # Agent resources
```

### Hash-Based Storage

Each cached repository is stored in a directory named by the SHA256 hash of its URL. This ensures:
- **Unique storage** for each repository
- **Collision-free** caching
- **Efficient lookup** by URL

### Cache Metadata

The `.cache-metadata.json` file tracks:
- Repository URLs
- Clone timestamps
- Git refs (branches, tags, commits)
- Last access time

## Cache Management

### View Cache Status

Check workspace cache usage:
```bash
ls -lh ~/.local/share/ai-config/repo/.workspace/
```

### Prune Unreferenced Caches

Remove caches that are no longer referenced by any resources:

```bash
# Preview what would be removed
aimgr repo prune --dry-run

# Remove unreferenced caches
aimgr repo prune

# Force remove without confirmation
aimgr repo prune --force
```

**When to prune**:
- After removing many resources
- When `.workspace/` grows too large
- To free up disk space

### Manual Cache Cleanup

If needed, you can manually remove the workspace cache:
```bash
rm -rf ~/.local/share/ai-config/repo/.workspace/
```

The cache will be recreated on the next Git operation.

## How It Works

### 1. First Clone

When you add a resource from a Git repository for the first time:

```bash
aimgr repo add gh:owner/repo
```

1. Calculate SHA256 hash of the repository URL
2. Clone the repository to `.workspace/<hash>/`
3. Extract resources from the cached repository
4. Save metadata about the cache
5. Copy resources to the main repository

### 2. Subsequent Operations

When you add more resources from the same repository:

```bash
aimgr repo add gh:owner/repo --filter "skill/*"
```

1. Calculate SHA256 hash of the repository URL
2. Check if `.workspace/<hash>/` exists
3. Reuse existing cached repository (no clone needed)
4. Extract resources from the cache
5. Copy resources to the main repository

**Result**: 10-50x faster than re-cloning

### 3. Updates

When you update resources from a Git repository:

```bash
aimgr repo update
```

1. Identify Git sources from resource metadata
2. For each unique source, find cached repository
3. Run `git pull` in the cached repository
4. Extract updated resources
5. Update resources in the main repository

**Result**: Fast updates without full re-clone

## Cache Lifecycle

### Creation
- Cache created on first `repo add` from a Git source
- Full clone operation

### Reuse
- Cache reused for all subsequent operations on the same source
- No network operations for resource extraction

### Update
- Cache updated with `git pull` during `repo update`
- Incremental network operations only

### Pruning
- Cache removed if no resources reference it
- Triggered by `repo prune` command

## Implementation Details

### Cache Key Generation

```go
import "crypto/sha256"

// Generate cache key from repository URL
func getCacheKey(repoURL string) string {
    hash := sha256.Sum256([]byte(repoURL))
    return hex.EncodeToString(hash[:])
}
```

### Cache Lookup

```go
// Check if cache exists
func cacheExists(repoURL string) bool {
    key := getCacheKey(repoURL)
    cachePath := filepath.Join(workspaceDir, key)
    _, err := os.Stat(cachePath)
    return err == nil
}
```

### Cache Operations

The `pkg/workspace/` package provides:
- `Clone()` - Clone repository to cache
- `Pull()` - Update cached repository
- `GetPath()` - Get path to cached repository
- `Prune()` - Remove unreferenced caches

## Best Practices

1. **Let the cache build naturally**: Don't manually populate `.workspace/`
2. **Prune periodically**: Run `repo prune` after removing many resources
3. **Don't edit cached repos**: Cached repositories are read-only for aimgr
4. **Use patterns for bulk ops**: `--filter` flags work with cached repos
5. **Monitor disk usage**: Large repositories consume space even when cached

## Troubleshooting

### Cache Corruption

If a cached repository becomes corrupted:

```bash
# Remove the entire workspace cache
rm -rf ~/.local/share/ai-config/repo/.workspace/

# Or remove a specific cache
rm -rf ~/.local/share/ai-config/repo/.workspace/<hash>/
```

The cache will be recreated on the next operation.

### Stale Cache

If a cached repository has outdated content:

```bash
# Update all cached repositories
aimgr repo update

# Or force a fresh clone by removing the cache
rm -rf ~/.local/share/ai-config/repo/.workspace/
```

### Disk Space Issues

If `.workspace/` is consuming too much space:

```bash
# Check workspace size
du -sh ~/.local/share/ai-config/repo/.workspace/

# Prune unreferenced caches
aimgr repo prune

# Or remove entire workspace (will be recreated)
rm -rf ~/.local/share/ai-config/repo/.workspace/
```

## Related Documentation

- [Resource Formats](resource-formats.md) - Resource format specifications
- [Output Formats](output-formats.md) - CLI output formats
- [Pattern Matching](pattern-matching.md) - Pattern matching for resources
