package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repo"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repomanifest"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/sourcemetadata"
	"gopkg.in/yaml.v3"
)

func TestFormatInclude(t *testing.T) {
	tests := []struct {
		name     string
		include  []string
		expected string
	}{
		{
			name:     "nil include (no filtering)",
			include:  nil,
			expected: "all",
		},
		{
			name:     "empty slice (no filtering)",
			include:  []string{},
			expected: "all",
		},
		{
			name:     "single short pattern",
			include:  []string{"skills/*"},
			expected: "skills/*",
		},
		{
			name:     "two short patterns",
			include:  []string{"skills/*", "commands/*"},
			expected: "skills/*, commands/*",
		},
		{
			name:     "combined text exactly at limit",
			include:  []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, // 30 chars
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:     "combined text over limit shows count",
			include:  []string{"some-very-long-pattern/*", "another-very-long-pattern/*"},
			expected: "2 filters",
		},
		{
			name:     "many patterns shows count",
			include:  []string{"a", "b", "c", "d"},
			expected: "4 filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInclude(tt.include)
			if result != tt.expected {
				t.Errorf("formatInclude(%v) = %q, want %q", tt.include, result, tt.expected)
			}
		})
	}
}

func TestRenderSourcesTableIncludeColumn(t *testing.T) {
	// Create metadata (no sync times needed for this test)
	metadata := &sourcemetadata.SourceMetadata{
		Version: 1,
		Sources: make(map[string]*sourcemetadata.SourceState),
	}

	sources := []*repomanifest.Source{
		{
			Name: "no-filter",
			URL:  "https://github.com/user/repo",
			// Include is nil — should show "all"
		},
		{
			Name:    "with-filter",
			URL:     "https://github.com/user/other",
			Include: []string{"skills/*", "commands/*"},
		},
	}

	// renderSourcesTable writes to stdout; wrap in a capture to assert content
	// Since renderSourcesTable writes directly to os.Stdout via output.Table,
	// we just confirm it doesn't error and that formatInclude returns correct values
	// (the detailed table rendering is already tested by the output package).
	for i, src := range sources {
		got := formatInclude(src.Include)
		switch i {
		case 0:
			if got != "all" {
				t.Errorf("source without include: formatInclude() = %q, want %q", got, "all")
			}
		case 1:
			if !strings.Contains(got, "skills/*") {
				t.Errorf("source with include: formatInclude() = %q, should contain %q", got, "skills/*")
			}
		}
	}

	output := captureOutput(t, func() {
		if err := renderSourcesTable(sources, metadata); err != nil {
			t.Fatalf("renderSourcesTable() error = %v", err)
		}
	})

	for _, expected := range []string{"INCLUDE", "all", "skills/*"} {
		if !strings.Contains(output.Stdout, expected) {
			t.Fatalf("expected sources table output to contain %q, got:\n%s", expected, output.Stdout)
		}
	}
}

func TestFormatTimeSince(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "zero time",
			time:     time.Time{},
			expected: "never",
		},
		{
			name:     "just now",
			time:     time.Now().Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "minutes ago",
			time:     time.Now().Add(-5 * time.Minute),
			expected: "5m ago",
		},
		{
			name:     "hours ago",
			time:     time.Now().Add(-2 * time.Hour),
			expected: "2h ago",
		},
		{
			name:     "days ago",
			time:     time.Now().Add(-3 * 24 * time.Hour),
			expected: "3d ago",
		},
		{
			name:     "weeks ago",
			time:     time.Now().Add(-14 * 24 * time.Hour),
			expected: "2w ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeSince(tt.time)
			if result != tt.expected {
				t.Errorf("formatTimeSince() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckSourceHealth(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	existingPath := filepath.Join(tempDir, "existing")
	if err := os.Mkdir(existingPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name     string
		source   *repomanifest.Source
		expected bool
	}{
		{
			name: "existing local path",
			source: &repomanifest.Source{
				Name: "test-local",
				Path: existingPath,
			},
			expected: true,
		},
		{
			name: "non-existing local path",
			source: &repomanifest.Source{
				Name: "test-missing",
				Path: filepath.Join(tempDir, "nonexistent"),
			},
			expected: false,
		},
		{
			name: "remote URL (always healthy)",
			source: &repomanifest.Source{
				Name: "test-remote",
				URL:  "https://github.com/user/repo",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkSourceHealth(tt.source)
			if result != tt.expected {
				t.Errorf("checkSourceHealth() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatSource(t *testing.T) {
	now := time.Now()
	twoHoursAgo := now.Add(-2 * time.Hour)

	// Create metadata with sync time for first test
	metadataWithSync := &sourcemetadata.SourceMetadata{
		Version: 1,
		Sources: map[string]*sourcemetadata.SourceState{
			"my-local-commands": {
				LastSynced: twoHoursAgo,
			},
		},
	}

	// Create empty metadata for second test
	emptyMetadata := &sourcemetadata.SourceMetadata{
		Version: 1,
		Sources: make(map[string]*sourcemetadata.SourceState),
	}

	tests := []struct {
		name     string
		source   *repomanifest.Source
		metadata *sourcemetadata.SourceMetadata
		contains []string // Substrings that should be present
	}{
		{
			name: "local source with sync time",
			source: &repomanifest.Source{
				Name: "my-local-commands",
				Path: "/home/user/resources",
			},
			metadata: metadataWithSync,
			contains: []string{
				"my-local-commands",
				"local:",
				"/home/user/resources",
				"[symlink]",
				"ago",
			},
		},
		{
			name: "remote source never synced",
			source: &repomanifest.Source{
				Name: "agentskills-catalog",
				URL:  "https://github.com/agentskills/catalog",
			},
			metadata: emptyMetadata,
			contains: []string{
				"agentskills-catalog",
				"remote:",
				"https://github.com/agentskills/catalog",
				"[copy]",
				"never",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSource(tt.source, tt.metadata)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("formatSource() result missing %q\nGot: %s", substr, result)
				}
			}
		})
	}
}

func TestBuildRepoInfoOutput_OverriddenSourceIncludesRestoreFields(t *testing.T) {
	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{{
		Name:                "team-tools",
		Path:                "/tmp/local/team-tools",
		OverrideOriginalURL: "https://github.com/example/tools",
		OverrideOriginalRef: "main",
	}}}

	metadata := &sourcemetadata.SourceMetadata{Version: 1, Sources: map[string]*sourcemetadata.SourceState{}}

	out := buildRepoInfoOutput("/tmp/repo", 0, 0, 0, 0, 0, manifest, metadata)
	if len(out.Sources) != 1 {
		t.Fatalf("expected one source, got %d", len(out.Sources))
	}

	got := out.Sources[0]
	if !got.Overridden {
		t.Fatalf("expected overridden=true")
	}
	if got.RestoreTo == "" || !strings.Contains(got.RestoreTo, "https://github.com/example/tools") {
		t.Fatalf("expected restore target in structured output, got %q", got.RestoreTo)
	}
}

func TestBuildRepoInfoOutput_IncludesSourceIdentityFieldsForStructuredOutput(t *testing.T) {
	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{
		{
			Name:    "remote-root",
			URL:     "https://github.com/example/root",
			Ref:     "main",
			Subpath: "",
		},
		{
			Name:    "remote-subpath",
			URL:     "https://github.com/example/tools",
			Ref:     "release-2026",
			Subpath: "skills/core",
		},
		{
			Name: "remote-default",
			URL:  "https://github.com/example/default",
		},
	}}

	metadata := &sourcemetadata.SourceMetadata{Version: 1, Sources: map[string]*sourcemetadata.SourceState{}}
	out := buildRepoInfoOutput("/tmp/repo", 0, 0, 0, 0, 0, manifest, metadata)

	if len(out.Sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(out.Sources))
	}

	root := out.Sources[0]
	if root.Location != "https://github.com/example/root" {
		t.Fatalf("expected root source location to stay URL-only, got %q", root.Location)
	}
	if root.Ref != "main" {
		t.Fatalf("expected ref=main, got %q", root.Ref)
	}
	if root.Subpath != "" {
		t.Fatalf("expected empty subpath for repo root source, got %q", root.Subpath)
	}

	subpath := out.Sources[1]
	if subpath.Ref != "release-2026" {
		t.Fatalf("expected ref=release-2026, got %q", subpath.Ref)
	}
	if subpath.Subpath != "skills/core" {
		t.Fatalf("expected subpath=skills/core, got %q", subpath.Subpath)
	}

	defaultSource := out.Sources[2]
	if defaultSource.Ref != "" {
		t.Fatalf("expected empty ref for default source, got %q", defaultSource.Ref)
	}
	if defaultSource.Subpath != "" {
		t.Fatalf("expected empty subpath for default source, got %q", defaultSource.Subpath)
	}
}

func TestBuildRepoInfoOutput_JSONAndYAMLOmitUnsetSourceIdentityFields(t *testing.T) {
	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{
		{
			Name: "remote-default",
			URL:  "https://github.com/example/default",
		},
		{
			Name:    "remote-pinned",
			URL:     "https://github.com/example/pinned",
			Ref:     "stable",
			Subpath: "agents",
		},
	}}

	metadata := &sourcemetadata.SourceMetadata{Version: 1, Sources: map[string]*sourcemetadata.SourceState{}}
	out := buildRepoInfoOutput("/tmp/repo", 0, 0, 0, 0, 0, manifest, metadata)

	jsonBytes, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	var jsonParsed map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonParsed); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	jsonSources, ok := jsonParsed["sources"].([]any)
	if !ok || len(jsonSources) != 2 {
		t.Fatalf("expected 2 json sources, got %#v", jsonParsed["sources"])
	}

	jsonDefault, ok := jsonSources[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first json source object, got %#v", jsonSources[0])
	}
	if _, exists := jsonDefault["ref"]; exists {
		t.Fatalf("expected json default source to omit ref, got %#v", jsonDefault["ref"])
	}
	if _, exists := jsonDefault["subpath"]; exists {
		t.Fatalf("expected json default source to omit subpath, got %#v", jsonDefault["subpath"])
	}

	jsonPinned, ok := jsonSources[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second json source object, got %#v", jsonSources[1])
	}
	if jsonPinned["ref"] != "stable" {
		t.Fatalf("expected json ref=stable, got %#v", jsonPinned["ref"])
	}
	if jsonPinned["subpath"] != "agents" {
		t.Fatalf("expected json subpath=agents, got %#v", jsonPinned["subpath"])
	}

	yamlBytes, err := yaml.Marshal(out)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %v", err)
	}

	var yamlParsed map[string]any
	if err := yaml.Unmarshal(yamlBytes, &yamlParsed); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	yamlSources, ok := yamlParsed["sources"].([]any)
	if !ok || len(yamlSources) != 2 {
		t.Fatalf("expected 2 yaml sources, got %#v", yamlParsed["sources"])
	}

	yamlDefault, ok := yamlSources[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first yaml source object, got %#v", yamlSources[0])
	}
	if _, exists := yamlDefault["ref"]; exists {
		t.Fatalf("expected yaml default source to omit ref, got %#v", yamlDefault["ref"])
	}
	if _, exists := yamlDefault["subpath"]; exists {
		t.Fatalf("expected yaml default source to omit subpath, got %#v", yamlDefault["subpath"])
	}

	yamlPinned, ok := yamlSources[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second yaml source object, got %#v", yamlSources[1])
	}
	if yamlPinned["ref"] != "stable" {
		t.Fatalf("expected yaml ref=stable, got %#v", yamlPinned["ref"])
	}
	if yamlPinned["subpath"] != "agents" {
		t.Fatalf("expected yaml subpath=agents, got %#v", yamlPinned["subpath"])
	}
}

func TestRenderSourcesTable_ShowsRefAndSubpathIdentity(t *testing.T) {
	sources := []*repomanifest.Source{
		{
			Name: "repo-root",
			URL:  "https://github.com/example/catalog",
		},
		{
			Name:    "pinned-subpath",
			URL:     "https://github.com/example/catalog",
			Ref:     "release-2026",
			Subpath: "skills/platform",
		},
	}
	metadata := &sourcemetadata.SourceMetadata{Version: 1, Sources: map[string]*sourcemetadata.SourceState{}}

	output := captureOutput(t, func() {
		if err := renderSourcesTable(sources, metadata); err != nil {
			t.Fatalf("renderSourcesTable() failed: %v", err)
		}
	})

	for _, expected := range []string{"repo root", "ref \"release-2026\"", "subpath \"skills/platform\""} {
		if !strings.Contains(output.Stdout, expected) {
			t.Fatalf("expected table output to contain %q, got:\n%s", expected, output.Stdout)
		}
	}
}

func TestRenderSourcesTable_ShowsOverrideRestoreTarget(t *testing.T) {
	sources := []*repomanifest.Source{{
		Name:                    "team-tools",
		Path:                    "/tmp/local/team-tools",
		OverrideOriginalURL:     "https://github.com/example/tools",
		OverrideOriginalRef:     "main",
		OverrideOriginalSubpath: "resources",
	}}
	metadata := &sourcemetadata.SourceMetadata{Version: 1, Sources: map[string]*sourcemetadata.SourceState{}}

	output := captureOutput(t, func() {
		if err := renderSourcesTable(sources, metadata); err != nil {
			t.Fatalf("renderSourcesTable() failed: %v", err)
		}
	})

	for _, expected := range []string{"OVERRIDE", "url \"https://github.com/example/tools\"", "ref \"main\"", "subpath \"resources\""} {
		if !strings.Contains(output.Stdout, expected) {
			t.Fatalf("expected table output to contain %q, got:\n%s", expected, output.Stdout)
		}
	}
}

func TestRepoInfo_JSON_IncludesOverrideState(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)

	mgr := repo.NewManagerWithPath(repoDir)
	if err := mgr.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{{
		Name:                "team-tools",
		Path:                "/tmp/local/team-tools",
		OverrideOriginalURL: "https://github.com/example/tools",
		OverrideOriginalRef: "main",
	}}}
	if err := manifest.Save(repoDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	originalFormat := repoInfoFormatFlag
	repoInfoFormatFlag = "json"
	defer func() { repoInfoFormatFlag = originalFormat }()
	output := captureOutput(t, func() {
		if err := repoInfoCmd.RunE(repoInfoCmd, nil); err != nil {
			t.Fatalf("repo info failed: %v", err)
		}
	})

	var parsed struct {
		Sources []struct {
			Name       string `json:"name"`
			Overridden bool   `json:"overridden"`
			RestoreTo  string `json:"restore_to"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(output.Stdout), &parsed); err != nil {
		t.Fatalf("failed to parse repo info json: %v\n%s", err, output.Stdout)
	}
	if len(parsed.Sources) != 1 {
		t.Fatalf("expected one source, got %d", len(parsed.Sources))
	}
	if !parsed.Sources[0].Overridden {
		t.Fatalf("expected overridden source in json output")
	}
	if !strings.Contains(parsed.Sources[0].RestoreTo, "https://github.com/example/tools") {
		t.Fatalf("expected restore_to to include remote url, got %q", parsed.Sources[0].RestoreTo)
	}
}

func TestRepoInfo_MissingRepoDoesNotCreateLockState(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)
	if err := os.RemoveAll(repoDir); err != nil {
		t.Fatalf("failed to remove repo dir: %v", err)
	}

	output := captureOutput(t, func() {
		if err := repoInfoCmd.RunE(repoInfoCmd, nil); err != nil {
			t.Fatalf("repo info failed: %v", err)
		}
	})

	if !strings.Contains(output.Stdout, "Repository not initialized") {
		t.Fatalf("expected missing repo message, got:\n%s", output.Stdout)
	}
	if _, statErr := os.Stat(filepath.Join(repoDir, ".workspace")); !os.IsNotExist(statErr) {
		t.Fatalf("expected missing repo path to remain untouched, stat err: %v", statErr)
	}
}
