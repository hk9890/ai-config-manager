package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
)

// TestImportNestedCommands_EndToEnd verifies the complete import workflow
// (discovery → DetectType → AddBulk → storage) imports all discovered resources.
// This test reproduces bug ai-config-manager-2tzg where nested commands were
// silently skipped during import.
func TestImportNestedCommands_EndToEnd(t *testing.T) {
	// 1. Setup temp repo (NOT ~/.local/share/ai-config/repo)
	tempRepo := t.TempDir()
	mgr := NewManagerWithPath(tempRepo)

	// 2. Use existing testdata fixture
	fixtureDir := filepath.Join("..", "..", "testdata", "repos", "commands-nested")

	// Verify fixture exists
	if _, err := os.Stat(fixtureDir); err != nil {
		t.Fatalf("Fixture directory not found: %s (error: %v)", fixtureDir, err)
	}

	// 3. Discover resources (what SHOULD be imported)
	commands, err := discovery.DiscoverCommands(fixtureDir, "")
	if err != nil {
		t.Fatalf("Failed to discover commands: %v", err)
	}

	t.Logf("Discovered %d commands:", len(commands))
	for _, cmd := range commands {
		t.Logf("  - %s (path: %s)", cmd.Name, cmd.Path)
	}

	// We expect exactly 3 commands: build, test, nested/deploy
	if len(commands) != 3 {
		t.Errorf("Expected to discover 3 commands, but found %d", len(commands))
	}

	// 4. Import using AddBulk (matches real workflow)
	var paths []string
	for _, cmd := range commands {
		paths = append(paths, cmd.Path)
	}

	result, err := mgr.AddBulk(paths, BulkImportOptions{})
	if err != nil {
		t.Fatalf("AddBulk failed: %v", err)
	}

	// 5. CRITICAL ASSERTION: discovered count == imported count
	// This is where the bug manifests: nested commands are silently skipped
	if result.CommandCount != len(commands) {
		t.Errorf("Imported %d commands but discovered %d", result.CommandCount, len(commands))
		t.Logf("Added: %v", result.Added)
		t.Logf("Failed: %v", result.Failed)
		t.Logf("Skipped: %v", result.Skipped)
	}

	// 6. Verify each command file exists in repo
	expectedCommands := []string{
		"build.md",
		"test.md",
		"nested/deploy.md",
	}

	for _, cmdFile := range expectedCommands {
		expectedPath := filepath.Join(tempRepo, "commands", cmdFile)
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("Command not imported: %s (expected at: %s)", cmdFile, expectedPath)
		} else {
			t.Logf("✓ Command imported: %s", cmdFile)
		}
	}

	// 7. Report on failures
	if len(result.Failed) > 0 {
		t.Errorf("Import had %d failures:", len(result.Failed))
		for _, f := range result.Failed {
			t.Errorf("  - %s: %s", f.Path, f.Message)
		}
	}
}
