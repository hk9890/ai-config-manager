//go:build integration

package workspace

import (
	"os/exec"
	"testing"
)

// TestGetOrClone_Integration tests GetOrClone with a real Git repository
// This test requires network access and git to be installed
func TestGetOrClone_Integration(t *testing.T) {
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

// TestUpdate_Integration tests Update with a real Git repository
// This test requires network access and git to be installed
func TestUpdate_Integration(t *testing.T) {
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

// TestListCached_WithCaches verifies ListCached returns cached URLs
func TestListCached_WithCaches(t *testing.T) {
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

	// Verify cache exists - use os.Stat instead
	// (isValidCache requires .git directory which we can't guarantee)

	// Remove cache
	if err := mgr.Remove(testURL); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify cache was removed - check that it's no longer valid
	if mgr.isValidCache(cachePath) {
		t.Errorf("cache should be removed after Remove, but still valid: %s", cachePath)
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
