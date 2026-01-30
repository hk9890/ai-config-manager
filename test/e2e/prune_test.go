//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

// TestE2E_PruneOnCleanRepo verifies that prune operation works correctly on a clean repository.
// A clean repo is one where all cached Git repositories are referenced by installed resources.
func TestE2E_PruneOnCleanRepo(t *testing.T) {
	// Setup: Load config and create empty test repo
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)
	t.Logf("Using repo path: %s", repoPath)

	// Clean the repo before starting
	cleanTestRepo(t, repoPath)

	// Register cleanup for end of test
	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Step 1: Sync repo to populate it with resources
	// This will clone Git repos into .workspace/ and create metadata
	t.Log("Step 1: Syncing repository to populate with resources...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "sync")

	if err != nil {
		t.Fatalf("Sync failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Parse sync stats to verify resources were synced
	stats := parseSyncStats(t, stdout)
	totalSynced := stats.Added + stats.Updated
	t.Logf("Sync completed: %d resources synced (Added=%d, Updated=%d, Failed=%d)",
		totalSynced, stats.Added, stats.Updated, stats.Failed)

	if totalSynced == 0 {
		t.Fatal("No resources were synced - cannot test prune on clean repo")
	}

	if stats.Failed > 0 {
		t.Errorf("Sync had failures: Failed=%d", stats.Failed)
	}

	// Verify repo is populated
	if !isRepoPopulated(t, repoPath) {
		t.Fatal("Repository should be populated after sync")
	}

	// Step 2: Run prune with --dry-run on clean repo
	t.Log("Step 2: Running prune --dry-run on clean repository...")
	stdout, stderr, err = runAimgrWithEnv(t, configPath, env, "repo", "prune", "--dry-run")

	// Log output for debugging
	t.Logf("Prune dry-run stdout:\n%s", stdout)
	if stderr != "" {
		t.Logf("Prune dry-run stderr:\n%s", stderr)
	}

	// Verify prune succeeded
	if err != nil {
		t.Errorf("Prune dry-run failed: %v", err)
	}

	// Verify output indicates 0 resources would be pruned
	if !strings.Contains(stdout, "No unreferenced workspace caches found") {
		t.Errorf("Expected output to indicate no unreferenced caches, got:\n%s", stdout)
	}

	// Verify no resources are listed for removal
	if strings.Contains(stdout, "Would remove") {
		t.Errorf("Expected no resources to be listed for removal, but found 'Would remove' in output:\n%s", stdout)
	}

	t.Log("✓ Prune dry-run correctly identified clean repository with 0 unreferenced caches")
}

// TestE2E_PruneActuallyRemovesNothing verifies that prune (without --dry-run) removes nothing from a clean repo.
func TestE2E_PruneActuallyRemovesNothing(t *testing.T) {
	// Setup: Load config and create empty test repo
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)
	t.Logf("Using repo path: %s", repoPath)

	// Clean the repo before starting
	cleanTestRepo(t, repoPath)

	// Register cleanup for end of test
	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Step 1: Sync repo to populate it with resources
	t.Log("Step 1: Syncing repository to populate with resources...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "sync")

	if err != nil {
		t.Fatalf("Sync failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Parse sync stats
	stats := parseSyncStats(t, stdout)
	totalSynced := stats.Added + stats.Updated
	t.Logf("Sync completed: %d resources synced", totalSynced)

	if totalSynced == 0 {
		t.Fatal("No resources were synced - cannot test prune on clean repo")
	}

	// Step 2: Get initial resource list (JSON format for precise comparison)
	t.Log("Step 2: Getting initial resource list...")
	stdout, _, err = runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
	if err != nil {
		t.Fatalf("Initial repo list failed: %v", err)
	}

	initialResources := parseRepoListJSON(t, stdout)
	initialCount := len(initialResources)
	t.Logf("Initial resource count: %d", initialCount)

	if initialCount == 0 {
		t.Fatal("No resources found in repo after sync")
	}

	// Step 3: Run prune (without --dry-run) with --force to skip confirmation
	t.Log("Step 3: Running prune --force on clean repository...")
	stdout, stderr, err = runAimgrWithEnv(t, configPath, env, "repo", "prune", "--force")

	// Log output for debugging
	t.Logf("Prune stdout:\n%s", stdout)
	if stderr != "" {
		t.Logf("Prune stderr:\n%s", stderr)
	}

	// Verify prune succeeded
	if err != nil {
		t.Errorf("Prune failed: %v", err)
	}

	// Verify output indicates no resources were removed
	if !strings.Contains(stdout, "No unreferenced workspace caches found") {
		t.Errorf("Expected output to indicate no unreferenced caches, got:\n%s", stdout)
	}

	// Verify no removal messages
	if strings.Contains(stdout, "Removed") && !strings.Contains(stdout, "Removed 0") {
		t.Errorf("Expected no resources to be removed, but found removal message in output:\n%s", stdout)
	}

	// Step 4: Get final resource list and compare
	t.Log("Step 4: Verifying resource list unchanged after prune...")
	stdout, _, err = runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
	if err != nil {
		t.Fatalf("Final repo list failed: %v", err)
	}

	finalResources := parseRepoListJSON(t, stdout)
	finalCount := len(finalResources)
	t.Logf("Final resource count: %d", finalCount)

	// Assert: Resource count unchanged
	if finalCount != initialCount {
		t.Errorf("Resource count changed after prune: before=%d, after=%d", initialCount, finalCount)
	}

	// Verify each resource still exists
	for _, initialRes := range initialResources {
		name, ok := initialRes["Name"].(string)
		if !ok {
			continue
		}
		if findResourceInList(finalResources, name) == nil {
			t.Errorf("Resource %s was removed by prune (should not happen on clean repo)", name)
		}
	}

	t.Log("✓ Prune correctly removed nothing from clean repository")
	t.Logf("✓ Resource list unchanged: %d resources before and after", initialCount)
}

// TestE2E_PruneOutputFormats verifies that prune command produces correct output in various scenarios.
func TestE2E_PruneOutputFormats(t *testing.T) {
	// Setup: Load config and create empty test repo
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)
	t.Logf("Using repo path: %s", repoPath)

	// Clean the repo before starting
	cleanTestRepo(t, repoPath)

	// Register cleanup for end of test
	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Step 1: Sync repo to populate it
	t.Log("Step 1: Syncing repository...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	stdout, _, err := runAimgrWithEnv(t, configPath, env, "repo", "sync")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	stats := parseSyncStats(t, stdout)
	if (stats.Added + stats.Updated) == 0 {
		t.Fatal("No resources were synced")
	}

	// Step 2: Test dry-run output format
	t.Log("Step 2: Testing dry-run output format...")
	stdout, _, err = runAimgrWithEnv(t, configPath, env, "repo", "prune", "--dry-run")
	if err != nil {
		t.Errorf("Prune dry-run failed: %v", err)
	}

	// Check for expected message
	if !strings.Contains(stdout, "No unreferenced workspace caches found") {
		t.Errorf("Expected 'No unreferenced workspace caches found' in dry-run output, got:\n%s", stdout)
	}

	// Verify no file paths are listed (since there's nothing to prune)
	if strings.Contains(stdout, ".workspace") {
		t.Errorf("Expected no file paths in output for clean repo, but found .workspace reference:\n%s", stdout)
	}

	// Step 3: Test actual prune output format
	t.Log("Step 3: Testing actual prune output format...")
	stdout, _, err = runAimgrWithEnv(t, configPath, env, "repo", "prune", "--force")
	if err != nil {
		t.Errorf("Prune failed: %v", err)
	}

	// Check for expected message
	if !strings.Contains(stdout, "No unreferenced workspace caches found") {
		t.Errorf("Expected 'No unreferenced workspace caches found' in prune output, got:\n%s", stdout)
	}

	// Verify no removal summary (since nothing was removed)
	if strings.Contains(stdout, "✓ Removed") {
		// If "Removed" appears, it should say "Removed 0"
		if !strings.Contains(stdout, "Removed 0") {
			t.Errorf("Expected no removals or 'Removed 0', but got:\n%s", stdout)
		}
	}

	t.Log("✓ Output format verification passed")
}

// findResourceInList finds a resource in the list by name field.
// Returns the resource map or nil if not found.
func findResourceInList(items []map[string]interface{}, name string) map[string]interface{} {
	for _, item := range items {
		// Note: repo list JSON uses "Name" (capitalized) field
		if itemName, ok := item["Name"].(string); ok && itemName == name {
			return item
		}
	}
	return nil
}
