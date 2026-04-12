package manifest

import (
	"os"
	"path/filepath"
	"reflect"
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
			name: "basic manifest with resources",
			content: `resources:
  - skill/pdf-processing
  - command/test
  - agent/reviewer`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Resources) != 3 {
					t.Errorf("expected 3 resources, got %d", len(m.Resources))
				}
				if m.Resources[0] != "skill/pdf-processing" {
					t.Errorf("unexpected resource[0]: %s", m.Resources[0])
				}
			},
		},
		{
			name: "manifest with targets",
			content: `resources:
  - skill/test
targets:
  - claude
  - opencode`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				// Old 'targets' field should be migrated to 'install.targets'
				if len(m.Install.Targets) != 2 {
					t.Errorf("expected 2 install.targets, got %d", len(m.Install.Targets))
				}
				if m.Install.Targets[0] != "claude" {
					t.Errorf("unexpected install.target[0]: %s", m.Install.Targets[0])
				}
				// Old field should be cleared after migration
				if len(m.Targets) != 0 {
					t.Errorf("expected old Targets field to be cleared after migration, got %d", len(m.Targets))
				}
			},
		},
		{
			name:    "empty resources array",
			content: `resources: []`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Resources) != 0 {
					t.Errorf("expected 0 resources, got %d", len(m.Resources))
				}
			},
		},
		{
			name: "minimal manifest",
			content: `resources:
  - skill/test`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Resources) != 1 {
					t.Errorf("expected 1 resource, got %d", len(m.Resources))
				}
			},
		},
		{
			name: "manifest with remote sources",
			content: `resources:
  - skill/test
sources:
  - url: https://github.com/example/catalog.git
    ref: main
    subpath: //skills/../skills/core//
    name: primary`,
			wantErr: false,
			checkFn: func(t *testing.T, m *Manifest) {
				if len(m.Sources) != 1 {
					t.Fatalf("expected 1 source, got %d", len(m.Sources))
				}
				s := m.Sources[0]
				if s.URL != "https://github.com/example/catalog.git" {
					t.Fatalf("source url = %q", s.URL)
				}
				if s.Ref != "main" || s.Name != "primary" || s.Subpath != "skills/core" {
					t.Fatalf("unexpected source: %+v", s)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "ai.package.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Load manifest
			m, err := Load(path)
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

func TestLoad_InvalidManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "invalid yaml",
			content: "resources: [invalid yaml",
		},
		{
			name: "invalid resource format",
			content: `resources:
  - invalid-format`,
		},
		{
			name: "empty resource type",
			content: `resources:
  - /name`,
		},
		{
			name: "empty resource name",
			content: `resources:
  - skill/`,
		},
		{
			name: "invalid resource type",
			content: `resources:
  - unknown/test`,
		},
		{
			name: "invalid target",
			content: `resources:
  - skill/test
targets:
  - invalid-tool`,
		},
		{
			name: "source missing url",
			content: `sources:
  - name: missing-url`,
		},
		{
			name: "source with disallowed path field",
			content: `sources:
  - url: https://github.com/example/catalog.git
    path: ./local/catalog`,
		},
		{
			name: "source with local path style url",
			content: `sources:
  - url: ./relative/path`,
		},
		{
			name: "source with absolute path style url",
			content: `sources:
  - url: /tmp/catalog`,
		},
		{
			name: "source with file scheme url",
			content: `sources:
  - url: file:///tmp/catalog`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "ai.package.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			_, err := Load(path)
			if err == nil {
				t.Errorf("Load() expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.yaml")

	_, err := Load(path)
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoadOrCreate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ai.package.yaml")

	// Test creating new manifest
	m, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}
	if m == nil {
		t.Fatal("LoadOrCreate() returned nil manifest")
	}
	if len(m.Resources) != 0 {
		t.Errorf("expected empty resources, got %d items", len(m.Resources))
	}

	// Create a file
	content := `resources:
  - skill/test`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test loading existing manifest
	m2, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}
	if len(m2.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(m2.Resources))
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ai.package.yaml")

	m := &Manifest{
		Resources: []string{"skill/test", "command/build"},
		Install: InstallConfig{
			Targets: []string{"claude"},
		},
		Sources: []ManifestSource{{
			URL:     "https://github.com/example/catalog.git",
			Ref:     "main",
			Subpath: "skills/core",
			Name:    "primary",
		}},
	}

	// Save manifest
	if err := m.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create file")
	}

	// Load it back and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if len(loaded.Resources) != len(m.Resources) {
		t.Errorf("resources count mismatch: got %d, want %d", len(loaded.Resources), len(m.Resources))
	}

	if len(loaded.Install.Targets) != len(m.Install.Targets) {
		t.Errorf("install.targets count mismatch: got %d, want %d", len(loaded.Install.Targets), len(m.Install.Targets))
	}

	if !reflect.DeepEqual(loaded.Sources, m.Sources) {
		t.Errorf("sources mismatch: got %+v, want %+v", loaded.Sources, m.Sources)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "ai.package.yaml")

	m := &Manifest{
		Resources: []string{"skill/test"},
	}

	// Save should create directory
	if err := m.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create file in subdirectory")
	}
}

func TestSave_NilManifest(t *testing.T) {
	var m *Manifest
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ai.package.yaml")

	err := m.Save(path)
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
				Resources: []string{"skill/test", "command/build"},
				Install: InstallConfig{
					Targets: []string{"claude"},
				},
				Sources: []ManifestSource{{
					URL:     "https://github.com/example/catalog.git",
					Ref:     "main",
					Subpath: "skills",
					Name:    "primary",
				}},
			},
			wantErr: false,
		},
		{
			name: "empty manifest",
			m: &Manifest{
				Resources: []string{},
			},
			wantErr: false,
		},
		{
			name: "invalid resource format",
			m: &Manifest{
				Resources: []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "invalid resource type",
			m: &Manifest{
				Resources: []string{"unknown/test"},
			},
			wantErr: true,
		},
		{
			name: "invalid target",
			m: &Manifest{
				Resources: []string{"skill/test"},
				Install: InstallConfig{
					Targets: []string{"invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid source missing url",
			m: &Manifest{
				Sources: []ManifestSource{{Name: "missing-url"}},
			},
			wantErr: true,
		},
		{
			name: "invalid source local path style url",
			m: &Manifest{
				Sources: []ManifestSource{{URL: "./catalog"}},
			},
			wantErr: true,
		},
		{
			name: "subpath is normalized",
			m: &Manifest{
				Sources: []ManifestSource{{
					URL:     "https://github.com/example/catalog.git",
					Subpath: "//skills/../skills/core//",
				}},
			},
			wantErr: false,
		},
		{
			name:    "nil manifest",
			m:       nil,
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

func TestAdd(t *testing.T) {
	m := &Manifest{
		Resources: []string{"skill/test"},
	}

	// Add new resource
	if err := m.Add("command/build"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if len(m.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(m.Resources))
	}

	if m.Resources[1] != "command/build" {
		t.Errorf("unexpected resource: %s", m.Resources[1])
	}

	// Add duplicate (should not add)
	if err := m.Add("skill/test"); err != nil {
		t.Fatalf("Add() duplicate error = %v", err)
	}

	if len(m.Resources) != 2 {
		t.Errorf("expected 2 resources after duplicate add, got %d", len(m.Resources))
	}

	// Add invalid resource
	if err := m.Add("invalid"); err == nil {
		t.Error("Add() expected error for invalid resource, got nil")
	}
}

func TestAdd_NilManifest(t *testing.T) {
	var m *Manifest
	err := m.Add("skill/test")
	if err == nil {
		t.Error("Add() expected error for nil manifest, got nil")
	}
}

func TestRemove(t *testing.T) {
	m := &Manifest{
		Resources: []string{"skill/test", "command/build", "agent/reviewer"},
	}

	// Remove existing resource
	if err := m.Remove("command/build"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if len(m.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(m.Resources))
	}

	// Verify removed
	if m.Has("command/build") {
		t.Error("resource still exists after removal")
	}

	// Remove non-existent (should not error)
	if err := m.Remove("nonexistent/resource"); err != nil {
		t.Fatalf("Remove() nonexistent error = %v", err)
	}

	if len(m.Resources) != 2 {
		t.Errorf("expected 2 resources after removing nonexistent, got %d", len(m.Resources))
	}
}

func TestRemove_NilManifest(t *testing.T) {
	var m *Manifest
	err := m.Remove("skill/test")
	if err == nil {
		t.Error("Remove() expected error for nil manifest, got nil")
	}
}

func TestHas(t *testing.T) {
	m := &Manifest{
		Resources: []string{"skill/test", "command/build"},
	}

	// Test existing resource
	if !m.Has("skill/test") {
		t.Error("Has() returned false for existing resource")
	}

	// Test non-existent resource
	if m.Has("agent/reviewer") {
		t.Error("Has() returned true for non-existent resource")
	}

	// Test nil manifest
	var nilManifest *Manifest
	if nilManifest.Has("skill/test") {
		t.Error("Has() returned true for nil manifest")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ai.package.yaml")

	// Test non-existent file
	if Exists(path) {
		t.Error("Exists() returned true for non-existent file")
	}

	// Create file
	if err := os.WriteFile(path, []byte("resources: []"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test existing file
	if !Exists(path) {
		t.Error("Exists() returned false for existing file")
	}
}

func TestValidateResourceReference(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{"valid skill", "skill/test", false},
		{"valid command", "command/build", false},
		{"valid agent", "agent/reviewer", false},
		{"valid package", "package/web-tools", false},
		{"valid nested command", "command/opencode-coder-dev/analyze-logs", false},
		{"valid nested with multiple slashes", "skill/test/extra", false},
		{"empty string", "", true},
		{"no slash", "invalid", true},
		{"empty type", "/name", true},
		{"empty name", "skill/", true},
		{"invalid type", "unknown/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceReference(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResourceReference(%q) error = %v, wantErr %v", tt.ref, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidTarget(t *testing.T) {
	tests := []struct {
		target string
		want   bool
	}{
		{"claude", true},
		{"opencode", true},
		{"copilot", true},
		{"invalid", false},
		{"", false},
		{"Claude", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got := isValidTarget(tt.target)
			if got != tt.want {
				t.Errorf("isValidTarget(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestMerge_OverlaySemantics(t *testing.T) {
	base := &Manifest{
		Resources: []string{"skill/base", "command/shared", "agent/base"},
		Install:   InstallConfig{Targets: []string{"claude", "opencode"}},
		Sources: []ManifestSource{
			{URL: "https://github.com/example/catalog.git", Name: "base-name", Ref: "main", Subpath: "skills/core"},
			{URL: "https://github.com/example/tools", Name: "tools", Ref: "stable", Subpath: "commands"},
		},
	}
	local := &Manifest{
		Resources: []string{"command/shared", "skill/local", "agent/local"},
		Install:   InstallConfig{Targets: []string{"opencode", "copilot"}},
		Sources: []ManifestSource{
			{URL: "https://github.com/example/catalog", Name: "local-alias", Ref: "dev", Subpath: "//skills/core//"},
			{URL: "https://github.com/example/tools.git", Name: "tools-other", Ref: "other", Subpath: "//commands//"},
			{URL: "https://github.com/example/local-only.git", Name: "local-only", Ref: "feature", Subpath: "extras"},
		},
	}

	merged := Merge(base, local)

	wantResources := []string{"skill/base", "command/shared", "agent/base", "skill/local", "agent/local"}
	if len(merged.Resources) != len(wantResources) {
		t.Fatalf("resources length = %d, want %d (%v)", len(merged.Resources), len(wantResources), merged.Resources)
	}
	for i, want := range wantResources {
		if merged.Resources[i] != want {
			t.Fatalf("resources[%d] = %q, want %q (all=%v)", i, merged.Resources[i], want, merged.Resources)
		}
	}

	wantTargets := []string{"claude", "opencode", "copilot"}
	if len(merged.Install.Targets) != len(wantTargets) {
		t.Fatalf("targets length = %d, want %d (%v)", len(merged.Install.Targets), len(wantTargets), merged.Install.Targets)
	}
	for i, want := range wantTargets {
		if merged.Install.Targets[i] != want {
			t.Fatalf("targets[%d] = %q, want %q (all=%v)", i, merged.Install.Targets[i], want, merged.Install.Targets)
		}
	}

	wantSources := []ManifestSource{
		{URL: "https://github.com/example/catalog.git", Name: "base-name", Ref: "main", Subpath: "skills/core"},
		{URL: "https://github.com/example/tools", Name: "tools", Ref: "stable", Subpath: "commands"},
		{URL: "https://github.com/example/local-only.git", Name: "local-only", Ref: "feature", Subpath: "extras"},
	}
	if !reflect.DeepEqual(merged.Sources, wantSources) {
		t.Fatalf("sources mismatch:\n got: %+v\nwant: %+v", merged.Sources, wantSources)
	}
}

func TestLoadProjectManifests(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		projectDir := t.TempDir()
		basePath := filepath.Join(projectDir, ManifestFileName)
		if err := os.WriteFile(basePath, []byte("resources:\n  - skill/base\n"), 0644); err != nil {
			t.Fatalf("write base manifest: %v", err)
		}

		loaded, err := LoadProjectManifests(projectDir)
		if err != nil {
			t.Fatalf("LoadProjectManifests() error = %v", err)
		}
		if loaded.Base == nil || loaded.Local != nil || loaded.Effective == nil {
			t.Fatalf("unexpected loaded manifests: base=%v local=%v effective=%v", loaded.Base != nil, loaded.Local != nil, loaded.Effective != nil)
		}
		if !loaded.Effective.Has("skill/base") {
			t.Fatalf("effective manifest missing base resource: %v", loaded.Effective.Resources)
		}
	})

	t.Run("base plus local overlay", func(t *testing.T) {
		projectDir := t.TempDir()
		basePath := filepath.Join(projectDir, ManifestFileName)
		localPath := filepath.Join(projectDir, LocalManifestFileName)

		baseContent := `resources:
  - skill/base
  - command/shared
install:
  targets:
    - claude
sources:
  - url: https://github.com/example/catalog.git
    ref: main
    subpath: skills/core
    name: base-name
`
		localContent := `resources:
  - command/shared
  - skill/local
install:
  targets:
    - opencode
sources:
  - url: https://github.com/example/catalog
    ref: dev
    subpath: //skills/core//
    name: local-alias
  - url: https://github.com/example/local-only.git
    ref: feature
    subpath: extras
    name: local-only
`

		if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
			t.Fatalf("write base manifest: %v", err)
		}
		if err := os.WriteFile(localPath, []byte(localContent), 0644); err != nil {
			t.Fatalf("write local manifest: %v", err)
		}

		loaded, err := LoadProjectManifests(projectDir)
		if err != nil {
			t.Fatalf("LoadProjectManifests() error = %v", err)
		}

		wantResources := []string{"skill/base", "command/shared", "skill/local"}
		if len(loaded.Effective.Resources) != len(wantResources) {
			t.Fatalf("effective resources len = %d, want %d (%v)", len(loaded.Effective.Resources), len(wantResources), loaded.Effective.Resources)
		}
		for i, want := range wantResources {
			if loaded.Effective.Resources[i] != want {
				t.Fatalf("effective resources[%d] = %q, want %q", i, loaded.Effective.Resources[i], want)
			}
		}

		wantTargets := []string{"claude", "opencode"}
		if len(loaded.Effective.Install.Targets) != len(wantTargets) {
			t.Fatalf("effective targets len = %d, want %d (%v)", len(loaded.Effective.Install.Targets), len(wantTargets), loaded.Effective.Install.Targets)
		}
		for i, want := range wantTargets {
			if loaded.Effective.Install.Targets[i] != want {
				t.Fatalf("effective targets[%d] = %q, want %q", i, loaded.Effective.Install.Targets[i], want)
			}
		}

		wantSources := []ManifestSource{
			{URL: "https://github.com/example/catalog.git", Ref: "main", Subpath: "skills/core", Name: "base-name"},
			{URL: "https://github.com/example/local-only.git", Ref: "feature", Subpath: "extras", Name: "local-only"},
		}
		if !reflect.DeepEqual(loaded.Effective.Sources, wantSources) {
			t.Fatalf("effective sources mismatch:\n got: %+v\nwant: %+v", loaded.Effective.Sources, wantSources)
		}
	})

	t.Run("local only", func(t *testing.T) {
		projectDir := t.TempDir()
		localPath := filepath.Join(projectDir, LocalManifestFileName)
		if err := os.WriteFile(localPath, []byte("resources:\n  - skill/local-only\n"), 0644); err != nil {
			t.Fatalf("write local manifest: %v", err)
		}

		loaded, err := LoadProjectManifests(projectDir)
		if err != nil {
			t.Fatalf("LoadProjectManifests() error = %v", err)
		}
		if loaded.Base != nil || loaded.Local == nil || loaded.Effective == nil {
			t.Fatalf("unexpected loaded manifests: base=%v local=%v effective=%v", loaded.Base != nil, loaded.Local != nil, loaded.Effective != nil)
		}
		if !loaded.Effective.Has("skill/local-only") {
			t.Fatalf("effective manifest missing local-only resource: %v", loaded.Effective.Resources)
		}
	})

	t.Run("none present", func(t *testing.T) {
		projectDir := t.TempDir()
		loaded, err := LoadProjectManifests(projectDir)
		if err != nil {
			t.Fatalf("LoadProjectManifests() error = %v", err)
		}
		if loaded.HasAny() {
			t.Fatalf("expected HasAny=false when no manifest files exist")
		}
		if loaded.Effective != nil {
			t.Fatalf("expected nil effective manifest when no files exist")
		}
	})
}
