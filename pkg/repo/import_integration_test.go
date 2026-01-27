package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/discovery"
	"github.com/hk9890/ai-config-manager/pkg/metadata"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

// TestImportNestedCommands_EndToEnd verifies the complete import workflow:
// 1. Start with git repo on disk layout
// 2. Import using discovery + AddBulk
// 3. Verify all files are in repo
// 4. Verify all metadata entries exist and are correct
//
// This test reproduces bug ai-config-manager-2tzg where nested commands were
// silently skipped during import.
func TestImportNestedCommands_EndToEnd(t *testing.T) {
	// 1. Setup temp repo (NOT ~/.local/share/ai-config/repo)
	tempRepo := t.TempDir()
	mgr := NewManagerWithPath(tempRepo)

	// 2. Use existing testdata fixture (simulates git repo on disk)
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
	expectedCommands := []struct {
		name string
		file string
	}{
		{"build", "build.md"},
		{"test", "test.md"},
		{"nested/deploy", "nested/deploy.md"},
	}

	for _, cmd := range expectedCommands {
		expectedPath := filepath.Join(tempRepo, "commands", cmd.file)
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("Command file not imported: %s (expected at: %s)", cmd.file, expectedPath)
		} else {
			t.Logf("✓ Command file exists: %s", cmd.file)
		}
	}

	// 7. Verify metadata exists and is correct for each command
	for _, cmd := range expectedCommands {
		meta, err := metadata.Load(cmd.name, resource.Command, tempRepo)
		if err != nil {
			t.Errorf("Metadata not found for command '%s': %v", cmd.name, err)
			continue
		}

		// Check metadata fields
		if meta.Name != cmd.name {
			t.Errorf("Metadata name mismatch: got '%s', want '%s'", meta.Name, cmd.name)
		}
		if meta.Type != resource.Command {
			t.Errorf("Metadata type mismatch for '%s': got '%s', want 'command'", cmd.name, meta.Type)
		}
		if meta.SourceURL == "" {
			t.Errorf("Metadata missing SourceURL for '%s'", cmd.name)
		}
		if meta.SourceType == "" {
			t.Errorf("Metadata missing SourceType for '%s'", cmd.name)
		}

		t.Logf("✓ Metadata verified: %s (type=%s, source=%s)", cmd.name, meta.Type, meta.SourceType)
	}

	// 8. Report on import failures (should be none)
	if len(result.Failed) > 0 {
		t.Errorf("Import had %d failures:", len(result.Failed))
		for _, f := range result.Failed {
			t.Errorf("  - %s: %s", f.Path, f.Message)
		}
	}

	t.Logf("✓ Import test complete: %d commands discovered, %d imported, all files and metadata verified", 
		len(commands), result.CommandCount)
}
