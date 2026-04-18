//go:build integration

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/metadata"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repo"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/resource"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/workspace"
)

func TestFindUnreferencedCaches_NoWorkspace(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create repo manager
	manager := repo.NewManagerWithPath(tempDir)

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Find unreferenced caches (should return empty list, not error)
	unreferenced, err := findUnreferencedCaches(manager, workspaceManager)
	if err != nil {
		t.Fatalf("findUnreferencedCaches failed: %v", err)
	}

	if len(unreferenced) != 0 {
		t.Errorf("Expected 0 unreferenced caches, got %d", len(unreferenced))
	}
}

func TestFindUnreferencedCaches_NoCaches(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create workspace directory (but empty)
	workspaceDir := filepath.Join(tempDir, ".workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create repo manager
	manager := repo.NewManagerWithPath(tempDir)

	// Create workspace manager
	workspaceManager, err := workspace.NewManager(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Find unreferenced caches
	unreferenced, err := findUnreferencedCaches(manager, workspaceManager)
	if err != nil {
		t.Fatalf("findUnreferencedCaches failed: %v", err)
	}

	if len(unreferenced) != 0 {
		t.Errorf("Expected 0 unreferenced caches, got %d", len(unreferenced))
	}
}

func TestFindUnreferencedCaches_AllReferenced(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create metadata directory
	metadataDir := filepath.Join(tempDir, ".metadata", "commands")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tempDir, ".workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// URL to cache
	testURL := "https://github.com/test/repo"

	// Create a fake cached repo
	hash := workspace.ComputeHash(testURL)
	cachePath := filepath.Join(workspaceDir, hash)
	if err := os.MkdirAll(filepath.Join(cachePath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create workspace metadata pointing to this URL
	meta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "github",
		SourceURL:      testURL,
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta, tempDir, "test-source"); err != nil {
		t.Fatal(err)
	}

	// Create workspace manager and add cache metadata
	workspaceManager, err := workspace.NewManager(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Manually create cache metadata so ListCached works
	metadataPath := filepath.Join(workspaceDir, ".cache-metadata.json")
	metadataJSON := `{"version":"1.0","caches":{"` + hash + `":{"url":"` + normalizeURL(testURL) + `","last_accessed":"2026-01-01T00:00:00Z","last_updated":"2026-01-01T00:00:00Z","ref":"main"}}}`
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create repo manager
	manager := repo.NewManagerWithPath(tempDir)

	// Find unreferenced caches (should be empty - cache is referenced)
	unreferenced, err := findUnreferencedCaches(manager, workspaceManager)
	if err != nil {
		t.Fatalf("findUnreferencedCaches failed: %v", err)
	}

	if len(unreferenced) != 0 {
		t.Errorf("Expected 0 unreferenced caches (cache is referenced), got %d", len(unreferenced))
	}
}

func TestFindUnreferencedCaches_WithUnreferenced(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create metadata directory
	metadataDir := filepath.Join(tempDir, ".metadata", "commands")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create workspace directory
	workspaceDir := filepath.Join(tempDir, ".workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create two cached repos
	referencedURL := "https://github.com/test/referenced"
	unreferencedURL := "https://github.com/test/unreferenced"

	referencedHash := workspace.ComputeHash(referencedURL)
	unreferencedHash := workspace.ComputeHash(unreferencedURL)

	// Create fake cached repos
	referencedPath := filepath.Join(workspaceDir, referencedHash)
	unreferencedPath := filepath.Join(workspaceDir, unreferencedHash)

	if err := os.MkdirAll(filepath.Join(referencedPath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(unreferencedPath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	// Add some content to unreferenced cache for size testing
	if err := os.WriteFile(filepath.Join(unreferencedPath, "README.md"), []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create metadata only for the referenced repo
	meta := &metadata.ResourceMetadata{
		Name:           "test-command",
		Type:           resource.Command,
		SourceType:     "github",
		SourceURL:      referencedURL,
		FirstInstalled: time.Now(),
		LastUpdated:    time.Now(),
	}
	if err := metadata.Save(meta, tempDir, "test-source"); err != nil {
		t.Fatal(err)
	}

	// Create workspace manager and add cache metadata for both repos
	workspaceManager, err := workspace.NewManager(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create cache metadata
	metadataPath := filepath.Join(workspaceDir, ".cache-metadata.json")
	metadataJSON := `{"version":"1.0","caches":{
		"` + referencedHash + `":{"url":"` + normalizeURL(referencedURL) + `","last_accessed":"2026-01-01T00:00:00Z","last_updated":"2026-01-01T00:00:00Z","ref":"main"},
		"` + unreferencedHash + `":{"url":"` + normalizeURL(unreferencedURL) + `","last_accessed":"2026-01-01T00:00:00Z","last_updated":"2026-01-01T00:00:00Z","ref":"main"}
	}}`
	if err := os.WriteFile(metadataPath, []byte(metadataJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create repo manager
	manager := repo.NewManagerWithPath(tempDir)

	// Find unreferenced caches
	unreferenced, err := findUnreferencedCaches(manager, workspaceManager)
	if err != nil {
		t.Fatalf("findUnreferencedCaches failed: %v", err)
	}

	// Should find exactly one unreferenced cache
	if len(unreferenced) != 1 {
		t.Fatalf("Expected 1 unreferenced cache, got %d", len(unreferenced))
	}

	// Verify it's the correct one
	if unreferenced[0].URL != normalizeURL(unreferencedURL) {
		t.Errorf("Expected unreferenced URL %s, got %s", normalizeURL(unreferencedURL), unreferenced[0].URL)
	}

	// Verify size is calculated
	if unreferenced[0].Size == 0 {
		t.Error("Expected non-zero size for unreferenced cache")
	}

	// Verify path is correct
	expectedPath := filepath.Join(workspaceDir, unreferencedHash)
	if unreferenced[0].Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, unreferenced[0].Path)
	}
}

func TestRepoPrune_MissingRepoDoesNotCreateLockState(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)
	if err := os.RemoveAll(repoDir); err != nil {
		t.Fatalf("failed to remove repo dir: %v", err)
	}

	output := captureOutput(t, func() {
		if err := repoPruneCmd.RunE(repoPruneCmd, nil); err != nil {
			t.Fatalf("repo prune failed: %v", err)
		}
	})

	if !strings.Contains(output.Stdout, "No unreferenced workspace caches found.") {
		t.Fatalf("expected empty prune output, got:\n%s", output.Stdout)
	}
	if _, statErr := os.Stat(filepath.Join(repoDir, ".workspace")); !os.IsNotExist(statErr) {
		t.Fatalf("expected missing repo path to remain untouched, stat err: %v", statErr)
	}
}

func TestCollectReferencedGitURLs(t *testing.T) {
	// Create temp directory for test repo
	tempDir := t.TempDir()

	// Create metadata directories
	commandsDir := filepath.Join(tempDir, ".metadata", "commands")
	skillsDir := filepath.Join(tempDir, ".metadata", "skills")
	agentsDir := filepath.Join(tempDir, ".metadata", "agents")

	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create metadata with various source types
	metas := []metadata.ResourceMetadata{
		{
			Name:           "git-command",
			Type:           resource.Command,
			SourceType:     "github",
			SourceURL:      "https://github.com/test/repo1",
			FirstInstalled: time.Now(),
			LastUpdated:    time.Now(),
		},
		{
			Name:           "local-command",
			Type:           resource.Command,
			SourceType:     "local",
			SourceURL:      "/path/to/local",
			FirstInstalled: time.Now(),
			LastUpdated:    time.Now(),
		},
		{
			Name:           "git-skill",
			Type:           resource.Skill,
			SourceType:     "git-url",
			SourceURL:      "https://example.com/repo.git",
			FirstInstalled: time.Now(),
			LastUpdated:    time.Now(),
		},
		{
			Name:           "git-agent",
			Type:           resource.Agent,
			SourceType:     "gitlab",
			SourceURL:      "https://gitlab.com/test/repo",
			FirstInstalled: time.Now(),
			LastUpdated:    time.Now(),
		},
	}

	for _, meta := range metas {
		m := meta // Create copy for pointer
		if err := metadata.Save(&m, tempDir, "test-source"); err != nil {
			t.Fatal(err)
		}
	}

	// Create repo manager
	manager := repo.NewManagerWithPath(tempDir)

	// Collect referenced Git URLs
	urls, err := collectReferencedGitURLs(manager)
	if err != nil {
		t.Fatalf("collectReferencedGitURLs failed: %v", err)
	}

	// Should find 3 Git URLs (not the local one)
	if len(urls) != 3 {
		t.Fatalf("Expected 3 Git URLs, got %d: %v", len(urls), urls)
	}

	// Verify URLs are correct
	expectedURLs := map[string]bool{
		"https://github.com/test/repo1": false,
		"https://example.com/repo.git":  false,
		"https://gitlab.com/test/repo":  false,
	}

	for _, url := range urls {
		if _, exists := expectedURLs[url]; exists {
			expectedURLs[url] = true
		} else {
			t.Errorf("Unexpected URL found: %s", url)
		}
	}

	for url, found := range expectedURLs {
		if !found {
			t.Errorf("Expected URL not found: %s", url)
		}
	}
}
