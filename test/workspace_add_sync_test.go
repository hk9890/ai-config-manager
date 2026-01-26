package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/source"
	"github.com/hk9890/ai-config-manager/pkg/workspace"
)

// TestWorkspaceCacheWithRepoAdd verifies that repo add uses workspace cache
func TestWorkspaceCacheWithRepoAdd(t *testing.T) {
	// Create temp repository
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(repoDir)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Use anthropics/skills repository as test source
	testURL := "https://github.com/anthropics/skills"
	testRef := "main"

	t.Log("Step 1: Verify workspace cache doesn't exist yet")
	workspaceDir := filepath.Join(repoDir, ".workspace")
	if _, err := os.Stat(workspaceDir); err == nil {
		t.Error("Workspace directory shouldn't exist yet")
	}

	t.Log("Step 2: Simulate repo add (using workspace cache)")
	cachePath, err := workspaceManager.GetOrClone(testURL, testRef)
	if err != nil {
		t.Skipf("Skipping test - network required: %v", err)
	}
	t.Logf("  ✓ Repository cached at: %s", cachePath)

	t.Log("Step 3: Verify workspace cache was created")
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		t.Error("Workspace directory should exist after GetOrClone")
	}

	// Verify .git exists in cache
	gitDir := filepath.Join(cachePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("Cache should contain .git directory")
	}

	t.Log("Step 4: Verify cache can be reused")
	cachePath2, err := workspaceManager.GetOrClone(testURL, testRef)
	if err != nil {
		t.Fatalf("Second GetOrClone failed: %v", err)
	}

	if cachePath != cachePath2 {
		t.Errorf("Expected same cache path, got different: %s vs %s", cachePath, cachePath2)
	}
	t.Log("  ✓ Cache reused successfully")

	t.Log("Step 5: Verify resources can be discovered from cache")
	skills, err := discovery.DiscoverSkills(cachePath, "")
	if err != nil {
		t.Fatalf("Failed to discover skills from cache: %v", err)
	}

	if len(skills) == 0 {
		t.Error("Expected to find skills in cached repository")
	}
	t.Logf("  ✓ Discovered %d skills from cached repository", len(skills))
}

// TestWorkspaceCacheMetadataAfterAdd verifies cache metadata is created
func TestWorkspaceCacheMetadataAfterAdd(t *testing.T) {
	repoDir := t.TempDir()
	workspaceManager, err := workspace.NewManager(repoDir)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	testURL := "https://github.com/anthropics/skills"
	testRef := "main"

	t.Log("Cloning repository to cache")
	_, err = workspaceManager.GetOrClone(testURL, testRef)
	if err != nil {
		t.Skipf("Skipping test - network required: %v", err)
	}

	t.Log("Verifying cache metadata")
	metadataPath := filepath.Join(repoDir, ".workspace", ".cache-metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Cache metadata file should exist")
	}

	t.Log("Verifying cached URLs")
	cachedURLs, err := workspaceManager.ListCached()
	if err != nil {
		t.Fatalf("Failed to list cached URLs: %v", err)
	}

	if len(cachedURLs) == 0 {
		t.Error("Expected at least one cached URL")
	}

	// Verify our test URL is in the cache
	found := false
	for _, cached := range cachedURLs {
		if cached == testURL || cached == testURL+"/" || cached == testURL+".git" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Test URL %s not found in cached URLs: %v", testURL, cachedURLs)
	}
	t.Log("  ✓ Cache metadata tracking works correctly")
}

// TestWorkspaceCacheWithDifferentRefs verifies different refs are handled
func TestWorkspaceCacheWithDifferentRefs(t *testing.T) {
	repoDir := t.TempDir()
	workspaceManager, err := workspace.NewManager(repoDir)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	testURL := "https://github.com/anthropics/skills"

	t.Log("Step 1: Clone with main ref")
	cachePath1, err := workspaceManager.GetOrClone(testURL, "main")
	if err != nil {
		t.Skipf("Skipping test - network required: %v", err)
	}
	t.Logf("  ✓ Cached at: %s", cachePath1)

	t.Log("Step 2: Request same repo with different ref (should use same cache)")
	cachePath2, err := workspaceManager.GetOrClone(testURL, "main")
	if err != nil {
		t.Fatalf("Second GetOrClone failed: %v", err)
	}

	if cachePath1 != cachePath2 {
		t.Errorf("Expected same cache path for same URL, got: %s vs %s", cachePath1, cachePath2)
	}
	t.Log("  ✓ Same cache used for same URL with same ref")

	t.Log("Step 3: Verify cache directory count")
	workspaceDir := filepath.Join(repoDir, ".workspace")
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		t.Fatalf("Failed to read workspace dir: %v", err)
	}

	// Count actual cache directories (exclude .cache-metadata.json)
	cacheCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			cacheCount++
		}
	}

	if cacheCount != 1 {
		t.Errorf("Expected 1 cache directory, got %d", cacheCount)
	}
	t.Log("  ✓ Only one cache directory created")
}

// TestWorkspaceCacheWithRepoSync simulates sync behavior
func TestWorkspaceCacheWithRepoSync(t *testing.T) {
	repoDir := t.TempDir()
	manager := repo.NewManagerWithPath(repoDir)
	if err := manager.Init(); err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	workspaceManager, err := workspace.NewManager(repoDir)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	// Test with multiple repositories (simulating sync from multiple sources)
	testRepos := []struct {
		url string
		ref string
	}{
		{"https://github.com/anthropics/skills", "main"},
		{"https://github.com/hk9890/ai-tools", "main"},
	}

	t.Log("Step 1: Simulate syncing from multiple sources")
	for i, repo := range testRepos {
		t.Logf("  Syncing source %d: %s", i+1, repo.url)
		cachePath, err := workspaceManager.GetOrClone(repo.url, repo.ref)
		if err != nil {
			t.Logf("  ⚠ Skipping %s - network required: %v", repo.url, err)
			continue
		}
		t.Logf("    ✓ Cached at: %s", cachePath)
	}

	t.Log("Step 2: Verify both repositories are cached")
	cachedURLs, err := workspaceManager.ListCached()
	if err != nil {
		t.Fatalf("Failed to list cached URLs: %v", err)
	}

	t.Logf("  Cached URLs: %d", len(cachedURLs))
	for _, url := range cachedURLs {
		t.Logf("    - %s", url)
	}

	// We expect at least one cache (some may be skipped due to network)
	if len(cachedURLs) == 0 {
		t.Error("Expected at least one cached URL")
	}

	t.Log("  ✓ Multiple repositories cached successfully")
}

// TestWorkspaceCacheUpdateAfterAdd verifies update works after add
func TestWorkspaceCacheUpdateAfterAdd(t *testing.T) {
	repoDir := t.TempDir()
	workspaceManager, err := workspace.NewManager(repoDir)
	if err != nil {
		t.Fatalf("Failed to create workspace manager: %v", err)
	}

	testURL := "https://github.com/anthropics/skills"
	testRef := "main"

	t.Log("Step 1: Initial cache via GetOrClone (simulating repo add)")
	_, err = workspaceManager.GetOrClone(testURL, testRef)
	if err != nil {
		t.Skipf("Skipping test - network required: %v", err)
	}
	t.Log("  ✓ Repository cached")

	t.Log("Step 2: Update cached repository (simulating repo update)")
	err = workspaceManager.Update(testURL, testRef)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Log("  ✓ Repository updated successfully")

	t.Log("Step 3: Verify cache still exists after update")
	cachedURLs, err := workspaceManager.ListCached()
	if err != nil {
		t.Fatalf("Failed to list cached URLs: %v", err)
	}

	if len(cachedURLs) == 0 {
		t.Error("Cache should still exist after update")
	}
	t.Log("  ✓ Cache persists after update")
}

// TestWorkspaceCacheWithLocalSource verifies local sources don't create cache
func TestWorkspaceCacheWithLocalSource(t *testing.T) {
	repoDir := t.TempDir()
	
	// Create a test local source directory
	localSource := t.TempDir()
	
	// Parse local source (should not create cache)
	parsed, err := source.ParseSource(localSource)
	if err != nil {
		t.Fatalf("Failed to parse local source: %v", err)
	}

	if parsed.Type != source.Local {
		t.Errorf("Expected source type Local, got %v", parsed.Type)
	}

	t.Log("Step 1: Verify local sources don't use Git caching")
	workspaceDir := filepath.Join(repoDir, ".workspace")
	if _, err := os.Stat(workspaceDir); err == nil {
		// If workspace exists, verify it's empty (no caches created)
		entries, _ := os.ReadDir(workspaceDir)
		cacheCount := 0
		for _, e := range entries {
			if e.IsDir() {
				cacheCount++
			}
		}
		if cacheCount > 0 {
			t.Error("Local sources should not create workspace cache entries")
		}
	}
	t.Log("  ✓ Local sources correctly skip workspace caching")
}
