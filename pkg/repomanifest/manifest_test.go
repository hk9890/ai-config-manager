package repomanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		checkFn func(*testing.T, *Manifest)
	}{
		{
			name: "basic manifest with local source",
			content: `version: 1
sources:
  - name: my-local-commands
    path: /home/user/my-resources`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if m.Version != 1 {
					t.Errorf("expected version 1, got %d", m.Version)
				}
				if len(m.Sources) != 1 {
					t.Fatalf("expected 1 source, got %d", len(m.Sources))
				}
				s := m.Sources[0]
				if s.Name != "my-local-commands" {
					t.Errorf("unexpected name: %s", s.Name)
				}
				if s.Path != "/home/user/my-resources" {
					t.Errorf("unexpected path: %s", s.Path)
				}
				if s.GetMode() != "symlink" {
					t.Errorf("unexpected mode: %s", s.GetMode())
				}
			},
		},
		{
			name: "manifest with git source",
			content: `version: 1
sources:
  - name: agentskills-catalog
    url: https://github.com/agentskills/catalog
    ref: main
    subpath: resources`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Sources) != 1 {
					t.Fatalf("expected 1 source, got %d", len(m.Sources))
				}
				s := m.Sources[0]
				if s.Name != "agentskills-catalog" {
					t.Errorf("unexpected name: %s", s.Name)
				}
				if s.URL != "https://github.com/agentskills/catalog" {
					t.Errorf("unexpected url: %s", s.URL)
				}
				if s.Ref != "main" {
					t.Errorf("unexpected ref: %s", s.Ref)
				}
				if s.Subpath != "resources" {
					t.Errorf("unexpected subpath: %s", s.Subpath)
				}
			},
		},
		{
			name: "manifest with multiple sources",
			content: `version: 1
sources:
  - name: source-a
    path: /path/a
  - name: source-b
    url: https://github.com/user/repo`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Sources) != 2 {
					t.Errorf("expected 2 sources, got %d", len(m.Sources))
				}
			},
		},
		{
			name: "empty sources array",
			content: `version: 1
sources: []`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Sources) != 0 {
					t.Errorf("expected 0 sources, got %d", len(m.Sources))
				}
			},
		},
		{
			name: "minimal manifest",
			content: `version: 1
sources:
  - name: test
    path: /test
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Sources) != 1 {
					t.Errorf("expected 1 source, got %d", len(m.Sources))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, ManifestFileName)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			m, err := Load(tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, m)
			}
		})
	}
}

func TestLoad_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Load should return empty manifest, not error
	m, err := Load(tmpDir)
	if err != nil {
		t.Errorf("Load() unexpected error for missing file: %v", err)
	}
	if m == nil {
		t.Fatal("Load() returned nil manifest for missing file")
	}
	if m.Version != 1 {
		t.Errorf("expected version 1, got %d", m.Version)
	}
	if len(m.Sources) != 0 {
		t.Errorf("expected empty sources, got %d items", len(m.Sources))
	}
}

func TestLoad_InvalidManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "invalid yaml",
			content: "version: [invalid yaml",
		},
		{
			name: "invalid version",
			content: `version: 2
sources: []`,
		},
		{
			name: "duplicate source names",
			content: `version: 1
sources:
  - name: duplicate
    path: /path1
    mode: symlink
    added: 2026-02-14T10:00:00Z
  - name: duplicate
    path: /path2
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
		},
		{
			name: "missing name",
			content: `version: 1
sources:
  - path: /path
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
		},
		{
			name: "missing path and url",
			content: `version: 1
sources:
  - name: test
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
		},
		{
			name: "both path and url",
			content: `version: 1
sources:
  - name: test
    path: /path
    url: https://github.com/user/repo
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
		},

		{
			name: "invalid source name format",
			content: `version: 1
sources:
  - name: Invalid_Name
    path: /path
    mode: symlink
    added: 2026-02-14T10:00:00Z`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, ManifestFileName)
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			_, err := Load(tmpDir)
			if err == nil {
				t.Errorf("Load() expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				Name: "test-source",
				Path: "/test/path",
			},
		},
	}

	// Save manifest
	if err := m.Save(tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, ManifestFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create file")
	}

	// Load it back and verify
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.Version != m.Version {
		t.Errorf("version mismatch: got %d, want %d", loaded.Version, m.Version)
	}

	if len(loaded.Sources) != len(m.Sources) {
		t.Errorf("sources count mismatch: got %d, want %d", len(loaded.Sources), len(m.Sources))
	}

	if loaded.Sources[0].Name != m.Sources[0].Name {
		t.Errorf("source name mismatch: got %s, want %s", loaded.Sources[0].Name, m.Sources[0].Name)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "subdir", "repo")

	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	// Save should create directory
	if err := m.Save(repoPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	path := filepath.Join(repoPath, ManifestFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create file in subdirectory")
	}
}

func TestSave_NilManifest(t *testing.T) {
	var m *Manifest
	tmpDir := t.TempDir()

	err := m.Save(tmpDir)
	if err == nil {
		t.Error("Save() expected error for nil manifest, got nil")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		m       *Manifest
		wantErr bool
	}{
		{
			name: "valid manifest",
			m: &Manifest{
				Version: 1,
				Sources: []*Source{
					{
						Name: "test",
						Path: "/path",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty sources",
			m: &Manifest{
				Version: 1,
				Sources: []*Source{},
			},
			wantErr: false,
		},
		{
			name:    "nil manifest",
			m:       nil,
			wantErr: true,
		},
		{
			name: "invalid version",
			m: &Manifest{
				Version: 2,
				Sources: []*Source{},
			},
			wantErr: true,
		},
		{
			name: "duplicate names",
			m: &Manifest{
				Version: 1,
				Sources: []*Source{
					{
						Name: "duplicate",
						Path: "/path1",
					},
					{
						Name: "duplicate",
						Path: "/path2",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddSource(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	// Add valid source
	source := &Source{
		Name: "test-source",
		Path: "/test/path",
	}

	if err := m.AddSource(source); err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}

	if len(m.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(m.Sources))
	}

	// Try to add duplicate name
	duplicate := &Source{
		Name: "test-source",
		Path: "/different/path",
	}

	if err := m.AddSource(duplicate); err == nil {
		t.Error("AddSource() expected error for duplicate name, got nil")
	}
}

func TestAddSource_AutoGenerateName(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	// Add source without name
	source := &Source{
		Path: "/home/user/my-resources",
	}

	if err := m.AddSource(source); err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}

	if source.Name == "" {
		t.Error("AddSource() did not auto-generate name")
	}

	t.Logf("Auto-generated name: %s", source.Name)
}

func TestAddSource_DuplicateIDDetection(t *testing.T) {
	t.Run("same path same name is duplicate by name", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		// Add first source
		s1 := &Source{
			Name: "tools",
			Path: "/tmp/foo/tools",
		}
		if err := m.AddSource(s1); err != nil {
			t.Fatalf("AddSource() first add error = %v", err)
		}

		// Add same source again with same name → caught by existing name check
		s2 := &Source{
			Name: "tools",
			Path: "/tmp/foo/tools",
		}
		err := m.AddSource(s2)
		if err == nil {
			t.Fatal("AddSource() expected error for duplicate name, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %s", err)
		}
	})

	t.Run("same absolute path different name detected by ID", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		// Use a temp dir so we can create paths that resolve to the same absolute path
		tmpDir := t.TempDir()
		toolsDir := filepath.Join(tmpDir, "foo", "tools")
		if err := os.MkdirAll(toolsDir, 0755); err != nil {
			t.Fatalf("failed to create tools dir: %v", err)
		}

		// Add first source with canonical path
		s1 := &Source{
			Name: "tools",
			Path: toolsDir,
		}
		if err := m.AddSource(s1); err != nil {
			t.Fatalf("AddSource() first add error = %v", err)
		}

		// Add same path via "../foo/tools" with a different name
		altPath := filepath.Join(tmpDir, "foo", "..", "foo", "tools")
		s2 := &Source{
			Name: "my-tools",
			Path: altPath,
		}
		err := m.AddSource(s2)
		if err == nil {
			t.Fatal("AddSource() expected error for duplicate ID (same path, different name), got nil")
		}
		if !strings.Contains(err.Error(), "same location already exists") {
			t.Errorf("expected 'same location already exists' in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "tools") {
			t.Errorf("expected existing source name 'tools' in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "src-") {
			t.Errorf("expected source ID in error, got: %s", err)
		}
	})

	t.Run("same URL with and without git suffix detected by ID", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		// Add first source with URL without .git
		s1 := &Source{
			Name: "my-repo",
			URL:  "https://github.com/org/repo",
		}
		if err := m.AddSource(s1); err != nil {
			t.Fatalf("AddSource() first add error = %v", err)
		}

		// Add same URL with .git suffix and different name → same ID
		s2 := &Source{
			Name: "repo-copy",
			URL:  "https://github.com/org/repo.git",
		}
		err := m.AddSource(s2)
		if err == nil {
			t.Fatal("AddSource() expected error for duplicate ID (URL with .git), got nil")
		}
		if !strings.Contains(err.Error(), "same location already exists") {
			t.Errorf("expected 'same location already exists' in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "my-repo") {
			t.Errorf("expected existing source name 'my-repo' in error, got: %s", err)
		}
	})

	t.Run("different sources succeed with different IDs", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		// Add first source
		s1 := &Source{
			Name: "source-a",
			Path: "/path/to/source-a",
		}
		if err := m.AddSource(s1); err != nil {
			t.Fatalf("AddSource() source-a error = %v", err)
		}

		// Add second, genuinely different source
		s2 := &Source{
			Name: "source-b",
			URL:  "https://github.com/org/different-repo",
		}
		if err := m.AddSource(s2); err != nil {
			t.Fatalf("AddSource() source-b error = %v", err)
		}

		if len(m.Sources) != 2 {
			t.Errorf("expected 2 sources, got %d", len(m.Sources))
		}

		// Verify they have different IDs
		if m.Sources[0].ID == m.Sources[1].ID {
			t.Errorf("expected different IDs, both got: %s", m.Sources[0].ID)
		}
	})

	t.Run("same name same ID is caught by name check not ID check", func(t *testing.T) {
		m := &Manifest{
			Version: 1,
			Sources: []*Source{},
		}

		s1 := &Source{
			Name: "my-repo",
			URL:  "https://github.com/org/repo",
		}
		if err := m.AddSource(s1); err != nil {
			t.Fatalf("AddSource() first add error = %v", err)
		}

		// Same name, same URL → same ID, but should be caught by name check
		s2 := &Source{
			Name: "my-repo",
			URL:  "https://github.com/org/repo",
		}
		err := m.AddSource(s2)
		if err == nil {
			t.Fatal("AddSource() expected error for duplicate, got nil")
		}
		// Should be the name-based error, not the ID-based error
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' in error, got: %s", err)
		}
	})
}

// TestAddSource_SetsAddedTimestamp removed - timestamp handling moved to sourcemetadata package

func TestAddSource_NilManifest(t *testing.T) {
	var m *Manifest
	source := &Source{
		Name: "test",
		Path: "/test",
	}

	err := m.AddSource(source)
	if err == nil {
		t.Error("AddSource() expected error for nil manifest, got nil")
	}
}

func TestAddSource_NilSource(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	err := m.AddSource(nil)
	if err == nil {
		t.Error("AddSource() expected error for nil source, got nil")
	}
}

func TestRemoveSource(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				Name: "source-a",
				Path: "/path/a",
			},
			{
				Name: "source-b",
				URL:  "https://github.com/user/repo",
			},
		},
	}

	// Remove by name
	removed, err := m.RemoveSource("source-a")
	if err != nil {
		t.Fatalf("RemoveSource() error = %v", err)
	}
	if removed == nil {
		t.Fatal("RemoveSource() returned nil source")
	}
	if removed.Name != "source-a" {
		t.Errorf("RemoveSource() returned wrong source: %s", removed.Name)
	}
	if len(m.Sources) != 1 {
		t.Errorf("expected 1 source after removal, got %d", len(m.Sources))
	}

	// Remove by URL
	removed, err = m.RemoveSource("https://github.com/user/repo")
	if err != nil {
		t.Fatalf("RemoveSource() error = %v", err)
	}
	if removed.Name != "source-b" {
		t.Errorf("RemoveSource() returned wrong source: %s", removed.Name)
	}
	if len(m.Sources) != 0 {
		t.Errorf("expected 0 sources after removal, got %d", len(m.Sources))
	}

	// Remove non-existent
	_, err = m.RemoveSource("nonexistent")
	if err == nil {
		t.Error("RemoveSource() expected error for nonexistent source, got nil")
	}
}

func TestRemoveSource_ByPath(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				Name: "test",
				Path: "/test/path",
			},
		},
	}

	removed, err := m.RemoveSource("/test/path")
	if err != nil {
		t.Fatalf("RemoveSource() error = %v", err)
	}
	if removed.Name != "test" {
		t.Errorf("RemoveSource() returned wrong source: %s", removed.Name)
	}
}

func TestRemoveSource_ByID(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				ID:   "src-abc123def456",
				Name: "source-a",
				Path: "/path/a",
			},
			{
				ID:   "src-789012345678",
				Name: "source-b",
				URL:  "https://github.com/user/repo",
			},
		},
	}

	// Remove by ID (highest priority)
	removed, err := m.RemoveSource("src-abc123def456")
	if err != nil {
		t.Fatalf("RemoveSource() by ID error = %v", err)
	}
	if removed.Name != "source-a" {
		t.Errorf("RemoveSource() returned wrong source: %s", removed.Name)
	}
	if len(m.Sources) != 1 {
		t.Errorf("expected 1 source after removal, got %d", len(m.Sources))
	}

	// Verify the remaining source is source-b
	if m.Sources[0].Name != "source-b" {
		t.Errorf("expected remaining source to be source-b, got %s", m.Sources[0].Name)
	}
}

func TestRemoveSource_IDTakesPrecedenceOverName(t *testing.T) {
	// Edge case: one source's ID happens to match another source's name.
	// ID matching should take priority.
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				ID:   "src-aaa111222333",
				Name: "src-bbb444555666",
				Path: "/path/first",
			},
			{
				ID:   "src-bbb444555666",
				Name: "second-source",
				Path: "/path/second",
			},
		},
	}

	// "src-bbb444555666" matches source-1 by name AND source-2 by ID.
	// ID should win → removes source-2.
	removed, err := m.RemoveSource("src-bbb444555666")
	if err != nil {
		t.Fatalf("RemoveSource() error = %v", err)
	}
	if removed.Name != "second-source" {
		t.Errorf("expected RemoveSource to match by ID (second-source), got: %s", removed.Name)
	}
	if len(m.Sources) != 1 {
		t.Errorf("expected 1 source remaining, got %d", len(m.Sources))
	}
	if m.Sources[0].Name != "src-bbb444555666" {
		t.Errorf("expected remaining source to be 'src-bbb444555666', got %s", m.Sources[0].Name)
	}
}

func TestRemoveSource_NilManifest(t *testing.T) {
	var m *Manifest
	_, err := m.RemoveSource("test")
	if err == nil {
		t.Error("RemoveSource() expected error for nil manifest, got nil")
	}
}

func TestRemoveSource_EmptyNameOrPath(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	_, err := m.RemoveSource("")
	if err == nil {
		t.Error("RemoveSource() expected error for empty nameOrPath, got nil")
	}
}

func TestGetSource(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				Name: "test",
				Path: "/test/path",
			},
			{
				Name: "git-source",
				URL:  "https://github.com/user/repo",
			},
		},
	}

	// Get by name
	source, found := m.GetSource("test")
	if !found {
		t.Error("GetSource() did not find source by name")
	}
	if source.Name != "test" {
		t.Errorf("GetSource() returned wrong source: %s", source.Name)
	}

	// Get by path
	source, found = m.GetSource("/test/path")
	if !found {
		t.Error("GetSource() did not find source by path")
	}
	if source.Name != "test" {
		t.Errorf("GetSource() returned wrong source: %s", source.Name)
	}

	// Get by URL
	source, found = m.GetSource("https://github.com/user/repo")
	if !found {
		t.Error("GetSource() did not find source by URL")
	}
	if source.Name != "git-source" {
		t.Errorf("GetSource() returned wrong source: %s", source.Name)
	}

	// Get non-existent
	_, found = m.GetSource("nonexistent")
	if found {
		t.Error("GetSource() found non-existent source")
	}
}

func TestGetSource_NilManifest(t *testing.T) {
	var m *Manifest
	_, found := m.GetSource("test")
	if found {
		t.Error("GetSource() returned true for nil manifest")
	}
}

func TestGetSource_EmptyNameOrPath(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{},
	}

	_, found := m.GetSource("")
	if found {
		t.Error("GetSource() returned true for empty nameOrPath")
	}
}

func TestHasSource(t *testing.T) {
	m := &Manifest{
		Version: 1,
		Sources: []*Source{
			{
				Name: "test",
				Path: "/test/path",
			},
		},
	}

	// Test existing
	if !m.HasSource("test") {
		t.Error("HasSource() returned false for existing source")
	}

	if !m.HasSource("/test/path") {
		t.Error("HasSource() returned false for existing path")
	}

	// Test non-existent
	if m.HasSource("nonexistent") {
		t.Error("HasSource() returned true for non-existent source")
	}

	// Test nil manifest
	var nilManifest *Manifest
	if nilManifest.HasSource("test") {
		t.Error("HasSource() returned true for nil manifest")
	}
}

func TestValidateSource(t *testing.T) {
	tests := []struct {
		name    string
		source  *Source
		wantErr bool
	}{
		{
			name: "valid local source",
			source: &Source{
				Name: "test",
				Path: "/test",
			},
			wantErr: false,
		},
		{
			name: "valid git source",
			source: &Source{
				Name: "test",
				URL:  "https://github.com/user/repo",
			},
			wantErr: false,
		},
		{
			name:    "nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name: "empty name",
			source: &Source{
				Path: "/test",
			},
			wantErr: true,
		},
		{
			name: "invalid name format",
			source: &Source{
				Name: "Invalid_Name",
				Path: "/test",
			},
			wantErr: true,
		},
		{
			name: "missing path and url",
			source: &Source{
				Name: "test",
			},
			wantErr: true,
		},
		{
			name: "both path and url",
			source: &Source{
				Name: "test",
				Path: "/test",
				URL:  "https://github.com/user/repo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSource(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidSourceName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid lowercase", "test", true},
		{"valid with hyphen", "test-source", true},
		{"valid with numbers", "test123", true},
		{"valid complex", "my-test-source-123", true},
		{"single char", "a", true},
		{"uppercase", "Test", false},
		{"underscore", "test_source", false},
		{"starts with hyphen", "-test", false},
		{"ends with hyphen", "test-", false},
		{"consecutive hyphens", "test--source", false},
		{"special chars", "test@source", false},
		{"spaces", "test source", false},
		{"empty", "", false},
		{"too long", "a123456789012345678901234567890123456789012345678901234567890123456", false},
		{"64 chars exactly", "a123456789012345678901234567890123456789012345678901234567890123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSourceName(tt.input)
			if got != tt.want {
				t.Errorf("isValidSourceName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateSourceName(t *testing.T) {
	tests := []struct {
		name   string
		source *Source
		want   string
	}{
		{
			name:   "local path",
			source: &Source{Path: "/home/user/my-resources"},
			want:   "my-resources",
		},
		{
			name:   "github https",
			source: &Source{URL: "https://github.com/agentskills/catalog"},
			want:   "catalog",
		},
		{
			name:   "github https with .git",
			source: &Source{URL: "https://github.com/user/repo.git"},
			want:   "repo",
		},
		{
			name:   "github ssh",
			source: &Source{URL: "git@github.com:user/repo.git"},
			want:   "repo",
		},
		{
			name:   "path with special chars",
			source: &Source{Path: "/home/user/My_Resources!"},
			want:   "my-resources",
		},
		{
			name:   "path with consecutive hyphens",
			source: &Source{Path: "/home/user/my___resources"},
			want:   "my-resources",
		},
		{
			name:   "empty path and url",
			source: &Source{},
			want:   "source",
		},
		{
			name:   "all special chars",
			source: &Source{Path: "/home/user/!!!"},
			want:   "source",
		},
		{
			name:   "starts with special",
			source: &Source{Path: "/home/user/_resources"},
			want:   "resources",
		},
		{
			name:   "ends with special",
			source: &Source{Path: "/home/user/resources_"},
			want:   "resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSourceName(tt.source)
			if got != tt.want {
				t.Errorf("generateSourceName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTimestampFormat removed - timestamp handling moved to sourcemetadata package
