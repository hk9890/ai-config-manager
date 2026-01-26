package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalizeURL verifies URL normalization logic
func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic URL unchanged",
			input:    "https://github.com/anthropics/skills",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "https://GitHub.com/Anthropics/Skills",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "trailing slash removed",
			input:    "https://github.com/anthropics/skills/",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     ".git suffix removed",
			input:    "https://github.com/anthropics/skills.git",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "multiple transformations",
			input:    "  https://GitHub.com/Anthropics/Skills.git/  ",
			expected: "https://github.com/anthropics/skills",
		},
		{
			name:     "whitespace trimmed",
			input:    "  https://github.com/anthropics/skills  ",
			expected: "https://github.com/anthropics/skills",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeURL(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestComputeHash verifies hash computation consistency
func TestComputeHash(t *testing.T) {
	// Same URL should always produce same hash
	url1 := "https://github.com/anthropics/skills"
	hash1 := computeHash(url1)
	hash2 := computeHash(url1)

	if hash1 != hash2 {
		t.Errorf("computeHash produced inconsistent results: %s != %s", hash1, hash2)
	}

	// Different variations of same URL should produce same hash (after normalization)
	variations := []string{
		"https://github.com/anthropics/skills",
		"https://GitHub.com/Anthropics/Skills",
		"https://github.com/anthropics/skills/",
		"https://github.com/anthropics/skills.git",
		"  https://github.com/anthropics/skills  ",
	}

	expectedHash := computeHash(variations[0])
	for _, variation := range variations {
		hash := computeHash(variation)
		if hash != expectedHash {
			t.Errorf("computeHash(%q) = %s; want %s", variation, hash, expectedHash)
		}
	}

	// Different URLs should produce different hashes
	url2 := "https://github.com/different/repo"
	hashDifferent := computeHash(url2)
	if hashDifferent == hash1 {
		t.Errorf("computeHash produced same hash for different URLs")
	}

	// Hash should be 64 characters (SHA256 hex encoded)
	if len(hash1) != 64 {
		t.Errorf("computeHash produced hash of length %d; want 64", len(hash1))
	}
}

// TestNewManager verifies Manager creation
func TestNewManager(t *testing.T) {
	repoPath := "/tmp/test-repo"
	mgr, err := NewManager(repoPath)

	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	expectedWorkspace := filepath.Join(repoPath, ".workspace")
	if mgr.workspaceDir != expectedWorkspace {
		t.Errorf("Manager.workspaceDir = %q; want %q", mgr.workspaceDir, expectedWorkspace)
	}
}

// TestInit verifies workspace directory initialization
func TestInit(t *testing.T) {
	// Use temporary directory for testing
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "repo")

	mgr, err := NewManager(repoPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Workspace should not exist yet
	if _, err := os.Stat(mgr.workspaceDir); err == nil {
		t.Errorf("workspace directory should not exist before Init()")
	}

	// Initialize
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Workspace should now exist
	info, err := os.Stat(mgr.workspaceDir)
	if err != nil {
		t.Errorf("workspace directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("workspace path exists but is not a directory")
	}

	// Should be idempotent
	if err := mgr.Init(); err != nil {
		t.Errorf("Init should be idempotent but failed on second call: %v", err)
	}
}

// TestGetCachePath verifies cache path computation
func TestGetCachePath(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)

	url := "https://github.com/anthropics/skills"
	cachePath := mgr.getCachePath(url)

	// Should be under workspace directory
	expectedPrefix := filepath.Join(tmpDir, ".workspace")
	if !filepath.HasPrefix(cachePath, expectedPrefix) {
		t.Errorf("getCachePath returned path outside workspace: %s", cachePath)
	}

	// Should end with hash
	hash := computeHash(url)
	expectedPath := filepath.Join(expectedPrefix, hash)
	if cachePath != expectedPath {
		t.Errorf("getCachePath = %s; want %s", cachePath, expectedPath)
	}
}

// TestIsValidCache verifies cache validation logic
func TestIsValidCache(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)

	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name:     "non-existent directory",
			setup:    func(path string) {},
			expected: false,
		},
		{
			name: "directory without .git",
			setup: func(path string) {
				os.MkdirAll(path, 0755)
			},
			expected: false,
		},
		{
			name: "directory with .git file (not directory)",
			setup: func(path string) {
				os.MkdirAll(path, 0755)
				os.WriteFile(filepath.Join(path, ".git"), []byte("gitdir: ../repo/.git"), 0644)
			},
			expected: false,
		},
		{
			name: "valid cache with .git directory",
			setup: func(path string) {
				os.MkdirAll(path, 0755)
				os.MkdirAll(filepath.Join(path, ".git"), 0755)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := filepath.Join(tmpDir, tt.name)
			tt.setup(cachePath)

			result := mgr.isValidCache(cachePath)
			if result != tt.expected {
				t.Errorf("isValidCache = %v; want %v", result, tt.expected)
			}
		})
	}
}

// TestMetadataOperations verifies metadata loading and saving
func TestMetadataOperations(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)
	mgr.Init()

	// Load metadata (should create empty if not exists)
	metadata, err := mgr.loadMetadata()
	if err != nil {
		t.Fatalf("loadMetadata failed: %v", err)
	}

	if metadata.Version != "1.0" {
		t.Errorf("metadata.Version = %q; want %q", metadata.Version, "1.0")
	}

	if len(metadata.Caches) != 0 {
		t.Errorf("new metadata should have empty Caches map")
	}

	// Add an entry
	url := "https://github.com/anthropics/skills"
	hash := computeHash(url)
	metadata.Caches[hash] = CacheEntry{
		URL: normalizeURL(url),
		Ref: "main",
	}

	// Save metadata
	if err := mgr.saveMetadata(metadata); err != nil {
		t.Fatalf("saveMetadata failed: %v", err)
	}

	// Load again and verify
	loaded, err := mgr.loadMetadata()
	if err != nil {
		t.Fatalf("loadMetadata after save failed: %v", err)
	}

	if len(loaded.Caches) != 1 {
		t.Errorf("loaded metadata has %d caches; want 1", len(loaded.Caches))
	}

	entry, exists := loaded.Caches[hash]
	if !exists {
		t.Errorf("cache entry not found after reload")
	}

	if entry.URL != normalizeURL(url) {
		t.Errorf("entry.URL = %q; want %q", entry.URL, normalizeURL(url))
	}

	if entry.Ref != "main" {
		t.Errorf("entry.Ref = %q; want %q", entry.Ref, "main")
	}
}

// TestUpdateMetadataEntry verifies metadata entry updates
func TestUpdateMetadataEntry(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)
	mgr.Init()

	url := "https://github.com/anthropics/skills"
	ref := "main"

	// Add new entry
	if err := mgr.updateMetadataEntry(url, ref, "clone"); err != nil {
		t.Fatalf("updateMetadataEntry failed: %v", err)
	}

	// Verify entry was added
	metadata, _ := mgr.loadMetadata()
	hash := computeHash(url)
	entry, exists := metadata.Caches[hash]

	if !exists {
		t.Fatalf("cache entry not created")
	}

	if entry.URL != normalizeURL(url) {
		t.Errorf("entry.URL = %q; want %q", entry.URL, normalizeURL(url))
	}

	if entry.Ref != ref {
		t.Errorf("entry.Ref = %q; want %q", entry.Ref, ref)
	}

	if entry.LastAccessed.IsZero() {
		t.Errorf("entry.LastAccessed not set")
	}

	if entry.LastUpdated.IsZero() {
		t.Errorf("entry.LastUpdated not set for clone operation")
	}

	firstUpdated := entry.LastUpdated

	// Update with access only (no update)
	if err := mgr.updateMetadataEntry(url, ref, "access"); err != nil {
		t.Fatalf("updateMetadataEntry (access) failed: %v", err)
	}

	metadata, _ = mgr.loadMetadata()
	entry = metadata.Caches[hash]

	// LastUpdated should not change for access-only
	if entry.LastUpdated != firstUpdated {
		t.Errorf("LastUpdated changed on access-only operation")
	}
}

// TestGetOrClone_Integration tests GetOrClone with a real Git repository
// This test requires network access and git to be installed
func TestGetOrClone_Integration(t *testing.T) {
	// Skip in short mode or if git is not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if git is available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("git not available, skipping integration test")
	}

	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Use a small, stable public repository for testing
	testURL := "https://github.com/hk9890/ai-config-manager-test-repo"
	testRef := "main"

	// First call should clone
	cachePath1, err := mgr.GetOrClone(testURL, testRef)
	if err != nil {
		// If this fails, the test repo might not exist - skip test
		t.Skipf("GetOrClone failed (test repo may not exist): %v", err)
	}

	// Verify cache path exists
	if !mgr.isValidCache(cachePath1) {
		t.Errorf("cache path is not a valid git repo: %s", cachePath1)
	}

	// Second call should use cache
	cachePath2, err := mgr.GetOrClone(testURL, testRef)
	if err != nil {
		t.Fatalf("GetOrClone (cached) failed: %v", err)
	}

	// Should return same path
	if cachePath1 != cachePath2 {
		t.Errorf("GetOrClone returned different paths: %s != %s", cachePath1, cachePath2)
	}

	// Verify metadata was created
	metadata, err := mgr.loadMetadata()
	if err != nil {
		t.Fatalf("loadMetadata failed: %v", err)
	}

	hash := computeHash(testURL)
	entry, exists := metadata.Caches[hash]
	if !exists {
		t.Errorf("metadata entry not created")
	}

	if entry.URL != normalizeURL(testURL) {
		t.Errorf("metadata URL = %q; want %q", entry.URL, normalizeURL(testURL))
	}

	if entry.Ref != testRef {
		t.Errorf("metadata Ref = %q; want %q", entry.Ref, testRef)
	}
}

// TestGetOrClone_EmptyInputs tests error handling for empty inputs
func TestGetOrClone_EmptyInputs(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)

	tests := []struct {
		name        string
		url         string
		ref         string
		wantErr     bool
		checkErrMsg string // Optional: check that error is about validation, not clone failure
	}{
		{
			name:        "empty URL",
			url:         "",
			ref:         "main",
			wantErr:     true,
			checkErrMsg: "url cannot be empty",
		},
		{
			name:    "empty ref allowed (attempts clone)",
			url:     "https://github.com/test/repo",
			ref:     "",
			wantErr: true, // Will fail to clone non-existent repo, but not due to empty ref validation
		},
		{
			name:        "both empty",
			url:         "",
			ref:         "",
			wantErr:     true,
			checkErrMsg: "url cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.GetOrClone(tt.url, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrClone() error = %v; wantErr %v", err, tt.wantErr)
			}
			// If we want to check specific error message (for validation errors)
			if tt.checkErrMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.checkErrMsg) {
					t.Errorf("GetOrClone() error = %v; want error containing %q", err, tt.checkErrMsg)
				}
			}
		})
	}
}

// TestUpdate_EmptyInputs tests error handling for empty inputs
func TestUpdate_EmptyInputs(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)

	tests := []struct {
		name    string
		url     string
		ref     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			ref:     "main",
			wantErr: true,
		},
		{
			name:    "empty ref allowed (updates current branch)",
			url:     "https://github.com/test/repo",
			ref:     "",
			wantErr: true, // Still errors because cache doesn't exist
		},
		{
			name:    "both empty",
			url:     "",
			ref:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Update(tt.url, tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v; wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUpdate_CacheNotExists tests error when cache doesn't exist
func TestUpdate_CacheNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, _ := NewManager(tmpDir)
	mgr.Init()

	// Try to update a repo that was never cloned
	err := mgr.Update("https://github.com/test/repo", "main")
	if err == nil {
		t.Errorf("Update should fail when cache doesn't exist")
	}

	// Error should mention using GetOrClone first
	if !strings.Contains(err.Error(), "GetOrClone") {
		t.Errorf("Error should suggest using GetOrClone first, got: %v", err)
	}
}

// TestUpdate_Integration tests Update with a real Git repository
// This test requires network access and git to be installed
func TestUpdate_Integration(t *testing.T) {
	// Skip in short mode or if git is not available
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if git is available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("git not available, skipping integration test")
	}

	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Use a small, stable public repository for testing
	testURL := "https://github.com/hk9890/ai-config-manager-test-repo"
	testRef := "main"

	// First clone the repo
	cachePath, err := mgr.GetOrClone(testURL, testRef)
	if err != nil {
		t.Skipf("GetOrClone failed (test repo may not exist): %v", err)
	}

	// Now update it
	if err := mgr.Update(testURL, testRef); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify cache still exists and is valid
	if !mgr.isValidCache(cachePath) {
		t.Errorf("cache is not valid after update: %s", cachePath)
	}

	// Verify metadata was updated
	metadata, err := mgr.loadMetadata()
	if err != nil {
		t.Fatalf("loadMetadata failed: %v", err)
	}

	hash := computeHash(testURL)
	entry, exists := metadata.Caches[hash]
	if !exists {
		t.Errorf("metadata entry not found after update")
	}

	if entry.LastUpdated.IsZero() {
		t.Errorf("metadata LastUpdated not set")
	}
}

// TestListCached_Empty verifies ListCached with no caches
func TestListCached_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tempDir := t.TempDir()
	mgr, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Initialize workspace
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// List caches (should be empty)
	urls, err := mgr.ListCached()
	if err != nil {
		t.Fatalf("ListCached failed: %v", err)
	}

	if len(urls) != 0 {
		t.Errorf("expected 0 cached URLs, got %d", len(urls))
	}
}

// TestListCached_WithCaches verifies ListCached returns cached URLs
func TestListCached_WithCaches(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Check if git is available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("git not available")
	}

	tempDir := t.TempDir()
	mgr, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Use the test repository
	testURL := "https://github.com/hk9890/ai-config-manager-test-repo"

	// Clone to cache
	_, err = mgr.GetOrClone(testURL, "main")
	if err != nil {
		t.Skipf("GetOrClone failed (test repo may not exist): %v", err)
	}

	// List caches
	urls, err := mgr.ListCached()
	if err != nil {
		t.Fatalf("ListCached failed: %v", err)
	}

	if len(urls) != 1 {
		t.Fatalf("expected 1 cached URL, got %d", len(urls))
	}

	// Verify URL matches (normalized)
	expectedURL := normalizeURL(testURL)
	if urls[0] != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, urls[0])
	}
}

// TestRemove verifies removing a cached repository
func TestRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Check if git is available
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Skip("git not available")
	}

	tempDir := t.TempDir()
	mgr, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Use the test repository
	testURL := "https://github.com/hk9890/ai-config-manager-test-repo"

	// Clone to cache
	cachePath, err := mgr.GetOrClone(testURL, "main")
	if err != nil {
		t.Skipf("GetOrClone failed (test repo may not exist): %v", err)
	}

	// Verify cache exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatalf("cache directory should exist after GetOrClone")
	}

	// Remove cache
	if err := mgr.Remove(testURL); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify cache was removed
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Errorf("cache directory should be removed after Remove")
	}

	// Verify metadata was updated
	urls, err := mgr.ListCached()
	if err != nil {
		t.Fatalf("ListCached failed: %v", err)
	}

	if len(urls) != 0 {
		t.Errorf("expected 0 cached URLs after removal, got %d", len(urls))
	}
}

// TestRemove_NonExistent verifies removing a non-existent cache returns error
func TestRemove_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	mgr, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Try to remove non-existent cache
	err = mgr.Remove("https://github.com/nonexistent/repo")
	if err == nil {
		t.Error("Remove should return error for non-existent cache")
	}
}

// TestPrune verifies pruning unreferenced caches
func TestPrune(t *testing.T) {
	// This is a unit test that doesn't need real Git repos
	tempDir := t.TempDir()
	mgr, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Create fake cache directories manually
	workspaceDir := filepath.Join(tempDir, ".workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatal(err)
	}

	testURL1 := "https://github.com/test/repo1"
	testURL2 := "https://github.com/test/repo2"

	hash1 := computeHash(testURL1)
	hash2 := computeHash(testURL2)

	cache1 := filepath.Join(workspaceDir, hash1)
	cache2 := filepath.Join(workspaceDir, hash2)

	// Create fake .git directories
	if err := os.MkdirAll(filepath.Join(cache1, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cache2, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create metadata pointing to both repos
	metadata := &CacheMetadata{
		Version: "1.0",
		Caches: map[string]CacheEntry{
			hash1: {
				URL: normalizeURL(testURL1),
				Ref: "main",
			},
			hash2: {
				URL: normalizeURL(testURL2),
				Ref: "main",
			},
		},
	}
	if err := mgr.saveMetadata(metadata); err != nil {
		t.Fatal(err)
	}

	// Verify both caches exist
	urls, err := mgr.ListCached()
	if err != nil {
		t.Fatalf("ListCached failed: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 cached URLs, got %d", len(urls))
	}

	// Prune with only URL1 referenced
	removed, err := mgr.Prune([]string{testURL1})
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	// Verify URL2 was removed
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed URL, got %d", len(removed))
	}

	normalizedURL2 := normalizeURL(testURL2)
	if removed[0] != normalizedURL2 {
		t.Errorf("expected removed URL %s, got %s", normalizedURL2, removed[0])
	}

	// Verify only URL1 remains cached
	urls, err = mgr.ListCached()
	if err != nil {
		t.Fatalf("ListCached failed: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("expected 1 cached URL after prune, got %d", len(urls))
	}
	if urls[0] != normalizeURL(testURL1) {
		t.Errorf("expected remaining URL %s, got %s", normalizeURL(testURL1), urls[0])
	}
}
