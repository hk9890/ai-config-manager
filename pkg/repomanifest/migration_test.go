package repomanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateSourceIDs_LoadGeneratesIDs(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a manifest without source IDs (simulating pre-migration repo)
	content := `version: 1
sources:
  - name: local-source
    path: /home/user/my-resources
  - name: git-source
    url: https://github.com/agentskills/catalog
    ref: main
    subpath: resources
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// Load should auto-generate IDs
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify all sources now have IDs
	for _, s := range m.Sources {
		if s.ID == "" {
			t.Errorf("source %q should have an ID after migration", s.Name)
		}
	}

	// Verify IDs match what GenerateSourceID would produce
	for _, s := range m.Sources {
		expected := GenerateSourceID(s)
		if s.ID != expected {
			t.Errorf("source %q: ID %q doesn't match GenerateSourceID result %q", s.Name, s.ID, expected)
		}
	}
}

func TestMigrateSourceIDs_DoesNotPersistToDisk(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a manifest without source IDs
	content := `version: 1
sources:
  - name: local-source
    path: /home/user/my-resources
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// Load triggers in-memory migration only
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	generatedID := m.Sources[0].ID
	if generatedID == "" {
		t.Fatal("expected source to have an ID after migration")
	}

	// Read the file back directly to verify ID was NOT persisted
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read manifest file: %v", err)
	}

	if strings.Contains(string(data), generatedID) {
		t.Errorf("manifest file should not contain generated runtime ID %q:\n%s", generatedID, string(data))
	}
	if strings.Contains(string(data), "id:") {
		t.Errorf("manifest file should not contain id field after load:\n%s", string(data))
	}
}

func TestMigrateSourceIDs_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a manifest without source IDs
	content := `version: 1
sources:
  - name: local-source
    path: /home/user/my-resources
  - name: git-source
    url: https://github.com/user/repo
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// First load: triggers migration
	m1, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("first Load() error = %v", err)
	}

	ids1 := make(map[string]string)
	for _, s := range m1.Sources {
		ids1[s.Name] = s.ID
	}

	// Second load: should produce same IDs (no re-migration needed)
	m2, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("second Load() error = %v", err)
	}

	for _, s := range m2.Sources {
		if s.ID != ids1[s.Name] {
			t.Errorf("source %q: ID changed on second load: %q → %q", s.Name, ids1[s.Name], s.ID)
		}
	}
}

func TestMigrateSourceIDs_SkipsExistingIDs(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a manifest where one source already has an ID
	content := `version: 1
sources:
  - name: has-id
    id: src-existing123
    url: https://github.com/user/repo
  - name: no-id
    path: /home/user/resources
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Source with existing ID should keep it unchanged
	hasID, _ := m.GetSource("has-id")
	if hasID.ID != "src-existing123" {
		t.Errorf("existing ID should be preserved, got %q", hasID.ID)
	}

	// Source without ID should get one generated
	noID, _ := m.GetSource("no-id")
	if noID.ID == "" {
		t.Error("source without ID should have gotten one generated")
	}
	expected := GenerateSourceID(noID)
	if noID.ID != expected {
		t.Errorf("generated ID %q doesn't match GenerateSourceID result %q", noID.ID, expected)
	}
}

func TestMigrateSourceIDs_ReadOnlyManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a manifest without source IDs
	content := `version: 1
sources:
  - name: local-source
    url: https://github.com/user/repo
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// Make file/dir read-only. Load no longer persists IDs, but it should still
	// succeed and populate runtime IDs in memory.
	if err := os.Chmod(path, 0444); err != nil {
		t.Fatalf("failed to chmod manifest file: %v", err)
	}
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatalf("failed to chmod directory: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions for cleanup
		os.Chmod(tmpDir, 0755) //nolint:errcheck
		os.Chmod(path, 0644)   //nolint:errcheck
	})

	// Load should work and generate in-memory IDs
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() should not fail for read-only manifest, got: %v", err)
	}

	// Sources should still have IDs in memory even if not persisted
	for _, s := range m.Sources {
		if s.ID == "" {
			t.Errorf("source %q should have in-memory ID even when file is read-only", s.Name)
		}
	}
}

func TestMigrateSourceIDs_EmptySources(t *testing.T) {
	tmpDir := t.TempDir()

	content := `version: 1
sources: []
`
	path := filepath.Join(tmpDir, ManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// Load should work fine with no sources
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(m.Sources) != 0 {
		t.Errorf("expected 0 sources, got %d", len(m.Sources))
	}
}

func TestMigrateSourceIDs_MethodDirectly(t *testing.T) {
	t.Run("returns true when IDs were generated", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{
				{Name: "test", URL: "https://github.com/user/repo"},
			},
		}

		if !m.migrateSourceIDs() {
			t.Error("migrateSourceIDs() should return true when IDs are generated")
		}

		if m.Sources[0].ID == "" {
			t.Error("source should have an ID after migration")
		}
	})

	t.Run("returns false when all sources already have IDs", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{
				{Name: "test", ID: "src-abc123def456", URL: "https://github.com/user/repo"},
			},
		}

		if m.migrateSourceIDs() {
			t.Error("migrateSourceIDs() should return false when all sources have IDs")
		}

		// ID should be unchanged
		if m.Sources[0].ID != "src-abc123def456" {
			t.Errorf("existing ID should not be modified, got %q", m.Sources[0].ID)
		}
	})

	t.Run("returns false for empty sources", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		if m.migrateSourceIDs() {
			t.Error("migrateSourceIDs() should return false for empty sources")
		}
	})

	t.Run("generated IDs match GenerateSourceID", func(t *testing.T) {
		sources := []*Source{
			{Name: "local", Path: "/home/user/resources"},
			{Name: "remote", URL: "https://github.com/user/repo"},
		}

		// Compute expected IDs before migration
		expected := make(map[string]string)
		for _, s := range sources {
			expected[s.Name] = GenerateSourceID(s)
		}

		m := &Manifest{Version: 1, Sources: sources}
		m.migrateSourceIDs()

		for _, s := range m.Sources {
			if s.ID != expected[s.Name] {
				t.Errorf("source %q: migration ID %q != GenerateSourceID %q", s.Name, s.ID, expected[s.Name])
			}
		}
	})
}

func TestLoad_DoesNotMigrateLegacyFormatOnRead(t *testing.T) {
	repoPath := t.TempDir()
	manifestPath := filepath.Join(repoPath, ManifestFileName)

	legacy := `version: 1
sources:
  - name: legacy-source
    path: /tmp/legacy
    mode: symlink
    added: 2026-01-01T00:00:00Z
    last_synced: 2026-01-02T00:00:00Z
`
	if err := os.WriteFile(manifestPath, []byte(legacy), 0644); err != nil {
		t.Fatalf("failed to write legacy manifest: %v", err)
	}

	if _, err := Load(repoPath); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoPath, ".metadata", "sources.json")); !os.IsNotExist(err) {
		t.Fatalf("Load() should not write source metadata on read-only path")
	}

	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest after Load(): %v", err)
	}

	if !strings.Contains(string(raw), "mode:") || !strings.Contains(string(raw), "added:") {
		t.Fatalf("Load() unexpectedly rewrote legacy manifest:\n%s", string(raw))
	}
}

func TestLoadForMutation_MigratesLegacyFormat(t *testing.T) {
	repoPath := t.TempDir()
	manifestPath := filepath.Join(repoPath, ManifestFileName)

	legacy := `version: 1
sources:
  - name: legacy-source
    path: /tmp/legacy
    mode: symlink
    added: 2026-01-01T00:00:00Z
    last_synced: 2026-01-02T00:00:00Z
`
	if err := os.WriteFile(manifestPath, []byte(legacy), 0644); err != nil {
		t.Fatalf("failed to write legacy manifest: %v", err)
	}

	if _, err := LoadForMutation(repoPath); err != nil {
		t.Fatalf("LoadForMutation() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoPath, ".metadata", "sources.json")); err != nil {
		t.Fatalf("expected migration to write source metadata: %v", err)
	}

	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest after LoadForMutation(): %v", err)
	}

	if strings.Contains(string(raw), "mode:") || strings.Contains(string(raw), "added:") || strings.Contains(string(raw), "last_synced:") {
		t.Fatalf("expected legacy fields removed after migration, got:\n%s", string(raw))
	}
}

func TestMigrateSourceIDs_OverrideSourceUsesOriginalRemoteIdentity(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{{
			Name:                    "team-tools",
			Path:                    "/tmp/local/tools",
			OverrideOriginalURL:     "https://github.com/example/tools.git",
			OverrideOriginalRef:     "main",
			OverrideOriginalSubpath: "resources",
		}},
	}

	if !m.migrateSourceIDs() {
		t.Fatalf("expected migrateSourceIDs() to generate ID")
	}

	got := m.Sources[0].ID
	want := GenerateSourceID(&Source{URL: "https://github.com/example/tools"})
	if got != want {
		t.Fatalf("expected override source ID to use original remote identity, got %q want %q", got, want)
	}
}
