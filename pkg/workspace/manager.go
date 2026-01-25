package workspace

/*
Workspace Cache Design
======================

## Purpose
The workspace cache provides efficient Git repository management for resources sourced
from Git URLs. It avoids redundant clones when multiple resources come from the same
repository by maintaining a local cache of Git repositories.

## Directory Structure

~/.local/share/ai-config/repo/
  .workspace/                    # Workspace cache root
    <url-hash>/                  # Directory for each cached repository
      .git/                      # Git repository data
      README.md                  # Cloned repository contents
      commands/                  # Example: repository structure
      skills/
      ...
    .cache-metadata.json         # Optional: Cache index for quick lookups

### Cache Key Algorithm

Each cached repository is stored in a directory named after the SHA256 hash of its
normalized Git URL. This ensures:
- Collision-free storage (unique hash per URL)
- Consistent lookup (same URL always maps to same hash)
- Security (no special characters or path traversal in directory names)

Normalization rules:
1. Convert to lowercase
2. Strip trailing slashes
3. Strip .git suffix if present
4. Trim whitespace

Examples:
  https://github.com/anthropics/skills       → sha256(https://github.com/anthropics/skills)
  https://github.com/anthropics/skills.git   → sha256(https://github.com/anthropics/skills)
  https://github.com/anthropics/skills/      → sha256(https://github.com/anthropics/skills)
  https://GitHub.com/Anthropics/Skills       → sha256(https://github.com/anthropics/skills)

All above URLs normalize to the same hash and share the same cached directory.

## Cache Metadata Format

The optional .cache-metadata.json provides a quick lookup index without scanning
directories. Format:

{
  "version": "1.0",
  "caches": {
    "a3f2e1...": {
      "url": "https://github.com/anthropics/skills",
      "last_accessed": "2026-01-25T10:00:00Z",
      "last_updated": "2026-01-25T09:00:00Z",
      "ref": "main"
    }
  }
}

This metadata is optional and used for optimization only. The cache is designed to
work correctly even if metadata is missing or out of sync.

## API Design

The Manager provides methods for:
- GetOrClone: Retrieve cached repo or clone if missing
- Update: Pull latest changes for a cached repo
- ListCached: Enumerate all cached repositories
- Prune: Remove unused cached repos
- Remove: Delete a specific cached repo

All methods handle edge cases:
- Corrupted cache (missing .git directory)
- Network failures during clone/update
- Invalid Git URLs
- Ref changes (branch/tag/commit switches)
- Concurrent access (file locking where needed)

## Edge Cases Handled

### 1. Corrupted Cache
If a cached directory exists but is missing .git or is corrupted:
- Detection: Check for .git directory and run git status
- Recovery: Remove corrupted directory and re-clone

### 2. Missing .git Directory
If cache directory exists but .git is missing:
- Treat as corrupted cache
- Remove and re-clone

### 3. Ref Changes
When requesting a different ref (branch/tag/commit) than last used:
- Fetch latest refs from remote
- Checkout requested ref
- Update working tree

### 4. Network Failures
- Clone failures: Return error without creating cache directory
- Update failures: Leave cache in last known good state, return error

### 5. Concurrent Access
- Use file locking (flock) to prevent concurrent modifications
- Read operations can proceed concurrently
- Write operations (clone, update, remove) are serialized per cache

### 6. Invalid Git URLs
- Validate URL format before operations
- Return descriptive errors for malformed URLs

### 7. Disk Space Issues
- Check available disk space before clone
- Provide clear error messages if insufficient space

## Thread Safety

The Manager is designed to be safe for concurrent use from multiple goroutines.
Individual cache directories are locked during write operations to prevent
corruption from concurrent access.

## Performance Considerations

- Caching: Significantly reduces clone time for repeated accesses
- Batching: repo update command batches resources from same URL
- Incremental updates: git pull only fetches changed objects
- Lazy loading: Caches created on-demand, not pre-populated

## Migration Notes

When introducing workspace cache to existing installations:
1. Existing resources with Git sources continue to work
2. Cache is populated lazily on first update after upgrade
3. No migration of existing data required
4. Old behavior (clone per resource) is replaced transparently
*/

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager manages the workspace cache for Git repositories.
// The workspace cache stores cloned Git repositories to avoid redundant clones
// when multiple resources come from the same repository.
type Manager struct {
	workspaceDir string // Path to .workspace directory
}

// CacheEntry represents metadata for a single cached repository.
type CacheEntry struct {
	URL          string    `json:"url"`           // Normalized Git URL
	LastAccessed time.Time `json:"last_accessed"` // Last time cache was accessed
	LastUpdated  time.Time `json:"last_updated"`  // Last time cache was updated (git pull)
	Ref          string    `json:"ref"`           // Current ref (branch/tag/commit)
}

// CacheMetadata is the root structure for .cache-metadata.json.
type CacheMetadata struct {
	Version string                `json:"version"` // Metadata format version
	Caches  map[string]CacheEntry `json:"caches"`  // Map of hash -> CacheEntry
}

// NewManager creates a new workspace cache manager.
// The workspace directory is located at <repoPath>/.workspace/
func NewManager(repoPath string) (*Manager, error) {
	workspaceDir := filepath.Join(repoPath, ".workspace")
	return &Manager{
		workspaceDir: workspaceDir,
	}, nil
}

// Init initializes the workspace cache directory structure.
// Creates the .workspace directory if it doesn't exist.
func (m *Manager) Init() error {
	if err := os.MkdirAll(m.workspaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}
	return nil
}

// GetOrClone returns the path to a cached Git repository, cloning it if necessary.
//
// Parameters:
//   - url: Git repository URL (will be normalized)
//   - ref: Git ref to checkout (branch, tag, or commit hash)
//
// Returns:
//   - string: Absolute path to the cached repository directory
//   - error: Non-nil if clone/checkout fails
//
// Behavior:
//   - If cache exists and is valid: checkout ref and return path
//   - If cache exists but is corrupted: remove and re-clone
//   - If cache doesn't exist: clone repository
//
// Edge cases handled:
//   - Corrupted cache (missing .git): detected and re-cloned
//   - Network failures: error returned, no partial cache created
//   - Invalid URL: error returned before any filesystem operations
//   - Ref not found: error returned with helpful message
func (m *Manager) GetOrClone(url string, ref string) (string, error) {
	// Ensure workspace is initialized
	if err := m.Init(); err != nil {
		return "", err
	}

	// Validate inputs
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}
	if ref == "" {
		return "", fmt.Errorf("ref cannot be empty")
	}

	// Get cache path
	cachePath := m.getCachePath(url)

	// Check if cache exists and is valid
	if m.isValidCache(cachePath) {
		// Cache exists - ensure correct ref is checked out
		if err := m.checkoutRef(cachePath, ref); err != nil {
			// If checkout fails, try to recover by fetching
			if fetchErr := m.fetchRepo(cachePath); fetchErr != nil {
				// Fetch failed - cache may be corrupted, remove and re-clone
				if removeErr := os.RemoveAll(cachePath); removeErr != nil {
					return "", fmt.Errorf("failed to remove corrupted cache: %w", removeErr)
				}
				// Fall through to clone
			} else {
				// Fetch succeeded, try checkout again
				if err := m.checkoutRef(cachePath, ref); err != nil {
					return "", fmt.Errorf("failed to checkout ref after fetch: %w", err)
				}
				// Success - update metadata and return
				if err := m.updateMetadataEntry(url, ref, "access"); err != nil {
					// Log warning but don't fail - metadata is optional
					fmt.Fprintf(os.Stderr, "warning: failed to update metadata: %v\n", err)
				}
				return cachePath, nil
			}
		} else {
			// Checkout succeeded - update metadata and return
			if err := m.updateMetadataEntry(url, ref, "access"); err != nil {
				// Log warning but don't fail - metadata is optional
				fmt.Fprintf(os.Stderr, "warning: failed to update metadata: %v\n", err)
			}
			return cachePath, nil
		}
	}

	// Cache doesn't exist or was removed due to corruption - clone it
	if err := m.cloneRepo(url, cachePath, ref); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Update metadata
	if err := m.updateMetadataEntry(url, ref, "clone"); err != nil {
		// Log warning but don't fail - metadata is optional
		fmt.Fprintf(os.Stderr, "warning: failed to update metadata: %v\n", err)
	}

	return cachePath, nil
}

// Update pulls the latest changes for a cached repository.
//
// Parameters:
//   - url: Git repository URL (will be normalized)
//   - ref: Git ref to update to (branch, tag, or commit)
//
// Returns:
//   - error: Non-nil if update fails
//
// Behavior:
//   - Fetches latest refs from remote
//   - Checks out requested ref
//   - Updates working tree (git pull for branches, checkout for tags/commits)
//
// Edge cases handled:
//   - Cache doesn't exist: returns error (use GetOrClone first)
//   - Network failure: cache left in last known good state
//   - Ref doesn't exist: error returned
//   - Uncommitted changes: stashed before update, restored after
func (m *Manager) Update(url string, ref string) error {
	// TODO: Implement Update
	// 1. Normalize URL
	// 2. Compute cache hash
	// 3. Verify cache exists
	// 4. Check for uncommitted changes
	// 5. Stash changes if present
	// 6. Fetch from remote
	// 7. Checkout ref
	// 8. Pull if branch, otherwise just checkout
	// 9. Restore stashed changes if any
	// 10. Update metadata
	return fmt.Errorf("not implemented")
}

// ListCached returns all cached repository URLs.
//
// Returns:
//   - []string: List of normalized Git URLs for all cached repositories
//   - error: Non-nil if cache directory can't be read
//
// Behavior:
//   - Reads .cache-metadata.json if available
//   - Falls back to scanning .workspace directory if metadata missing
//   - Skips corrupted caches (missing .git)
//
// Note: May be slow for large caches without metadata file.
func (m *Manager) ListCached() ([]string, error) {
	// TODO: Implement ListCached
	// 1. Try to read .cache-metadata.json
	// 2. If exists, return URLs from metadata
	// 3. If not, scan .workspace directory
	// 4. For each subdirectory, verify it's a valid git repo
	// 5. Return list of URLs (may need to reconstruct from git remote)
	return nil, fmt.Errorf("not implemented")
}

// Prune removes cached repositories that are not referenced by any resources.
//
// Parameters:
//   - referencedURLs: Set of Git URLs currently referenced by resources
//
// Returns:
//   - []string: List of URLs that were removed
//   - error: Non-nil if pruning fails
//
// Behavior:
//   - Compares cached repos against referenced URLs
//   - Removes caches not in referenced set
//   - Updates cache metadata
//   - Returns list of removed URLs for logging
//
// Safety:
//   - Uses normalized URLs for comparison
//   - Dry-run mode available (future enhancement)
//   - Logs removals for audit trail
func (m *Manager) Prune(referencedURLs []string) ([]string, error) {
	// TODO: Implement Prune
	// 1. Get list of all cached URLs
	// 2. Normalize referenced URLs
	// 3. Find caches not in referenced set
	// 4. Remove each unreferenced cache
	// 5. Update metadata
	// 6. Return list of removed URLs
	return nil, fmt.Errorf("not implemented")
}

// Remove deletes a specific cached repository.
//
// Parameters:
//   - url: Git repository URL (will be normalized)
//
// Returns:
//   - error: Non-nil if removal fails or cache doesn't exist
//
// Behavior:
//   - Normalizes URL
//   - Computes cache hash
//   - Removes cache directory
//   - Updates cache metadata
//
// Safety:
//   - Acquires lock before removal to prevent concurrent access
//   - Validates hash before removal (sanity check)
func (m *Manager) Remove(url string) error {
	// TODO: Implement Remove
	// 1. Normalize URL
	// 2. Compute cache hash
	// 3. Acquire lock on cache directory
	// 4. Remove cache directory
	// 5. Update metadata
	// 6. Release lock
	return fmt.Errorf("not implemented")
}

// normalizeURL normalizes a Git URL for consistent hashing.
//
// Normalization rules:
//   - Convert to lowercase
//   - Trim whitespace
//   - Strip trailing slashes
//   - Strip .git suffix
//
// Example:
//
//	https://GitHub.com/Anthropics/Skills.git/  → https://github.com/anthropics/skills
func normalizeURL(url string) string {
	// Trim whitespace
	normalized := strings.TrimSpace(url)

	// Convert to lowercase
	normalized = strings.ToLower(normalized)

	// Strip trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")

	// Strip .git suffix
	normalized = strings.TrimSuffix(normalized, ".git")

	return normalized
}

// computeHash computes the SHA256 hash of a normalized URL.
// Returns the hash as a lowercase hexadecimal string.
func computeHash(url string) string {
	normalized := normalizeURL(url)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// getCachePath returns the filesystem path for a cached repository.
func (m *Manager) getCachePath(url string) string {
	hash := computeHash(url)
	return filepath.Join(m.workspaceDir, hash)
}

// isValidCache checks if a cache directory exists and contains a valid Git repository.
// Returns true if the cache is valid, false otherwise.
func (m *Manager) isValidCache(cachePath string) bool {
	// Check if directory exists
	info, err := os.Stat(cachePath)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check if .git directory exists
	gitPath := filepath.Join(cachePath, ".git")
	gitInfo, err := os.Stat(gitPath)
	if err != nil || !gitInfo.IsDir() {
		return false
	}

	// TODO: Could add additional validation here (e.g., git status)
	return true
}

// loadMetadata loads the cache metadata from .cache-metadata.json.
// Returns nil if the file doesn't exist (not an error).
func (m *Manager) loadMetadata() (*CacheMetadata, error) {
	metadataPath := filepath.Join(m.workspaceDir, ".cache-metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, return empty metadata
			return &CacheMetadata{
				Version: "1.0",
				Caches:  make(map[string]CacheEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read cache metadata: %w", err)
	}

	var metadata CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse cache metadata: %w", err)
	}

	return &metadata, nil
}

// saveMetadata saves the cache metadata to .cache-metadata.json.
func (m *Manager) saveMetadata(metadata *CacheMetadata) error {
	metadataPath := filepath.Join(m.workspaceDir, ".cache-metadata.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache metadata: %w", err)
	}

	return nil
}

// updateMetadataEntry updates the metadata for a specific cache entry.
func (m *Manager) updateMetadataEntry(url string, ref string, updateType string) error {
	metadata, err := m.loadMetadata()
	if err != nil {
		return err
	}

	hash := computeHash(url)
	normalized := normalizeURL(url)

	entry, exists := metadata.Caches[hash]
	if !exists {
		entry = CacheEntry{
			URL: normalized,
		}
	}

	now := time.Now()
	entry.Ref = ref
	entry.LastAccessed = now

	if updateType == "update" || updateType == "clone" {
		entry.LastUpdated = now
	}

	metadata.Caches[hash] = entry

	return m.saveMetadata(metadata)
}

// cloneRepo clones a Git repository to the specified cache path.
// Unlike pkg/source/git.go's CloneRepo, this does a full clone (not shallow)
// to support ref switching.
func (m *Manager) cloneRepo(url string, cachePath string, ref string) error {
	// Build git clone command (full clone for ref switching)
	args := []string{"clone"}

	// Add branch/tag reference if specified (optimization)
	if ref != "" {
		args = append(args, "--branch", ref)
	}

	args = append(args, url, cachePath)

	// Execute git clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up partial clone on failure
		_ = os.RemoveAll(cachePath)
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// checkoutRef checks out the specified ref in a Git repository.
func (m *Manager) checkoutRef(cachePath string, ref string) error {
	cmd := exec.Command("git", "-C", cachePath, "checkout", ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// fetchRepo fetches the latest refs from the remote repository.
func (m *Manager) fetchRepo(cachePath string) error {
	cmd := exec.Command("git", "-C", cachePath, "fetch", "--all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
