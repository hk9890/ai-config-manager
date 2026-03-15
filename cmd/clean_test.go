package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/resource"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/tools"
)

func TestParseCleanFormat(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "default table", raw: ""},
		{name: "table", raw: "table"},
		{name: "json", raw: "json"},
		{name: "yaml rejected", raw: "yaml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCleanFormat(tt.raw)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCleanOwnedResourceDirs_RemovesAllEntryTypesAndKeepsRoots(t *testing.T) {
	projectDir := t.TempDir()

	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	skillsDir := filepath.Join(projectDir, ".claude", "skills")
	agentsDir := filepath.Join(projectDir, ".claude", "agents")
	for _, d := range []string{commandsDir, skillsDir, agentsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	regularFile := filepath.Join(commandsDir, "manual.md")
	if err := os.WriteFile(regularFile, []byte("manual"), 0644); err != nil {
		t.Fatalf("write regular file: %v", err)
	}

	symlinkTarget := filepath.Join(projectDir, "elsewhere.txt")
	if err := os.WriteFile(symlinkTarget, []byte("outside"), 0644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(symlinkTarget, filepath.Join(skillsDir, "wrong-repo-link")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	if err := os.Symlink(filepath.Join(projectDir, "missing-target"), filepath.Join(skillsDir, "broken-link")); err != nil {
		t.Fatalf("create broken symlink: %v", err)
	}

	nested := filepath.Join(agentsDir, "namespace", "inner.txt")
	if err := os.MkdirAll(filepath.Dir(nested), 0755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(nested, []byte("nested"), 0644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	owned := []OwnedResourceDir{
		{Path: commandsDir, Tool: tools.Claude, ResourceType: resource.Command},
		{Path: skillsDir, Tool: tools.Claude, ResourceType: resource.Skill},
		{Path: agentsDir, Tool: tools.Claude, ResourceType: resource.Agent},
	}

	removed, failed := cleanOwnedResourceDirs(owned)
	if len(failed) != 0 {
		t.Fatalf("expected no failures, got %v", failed)
	}
	if len(removed) != 4 {
		t.Fatalf("expected 4 removed top-level entries, got %d", len(removed))
	}

	for _, d := range []string{commandsDir, skillsDir, agentsDir} {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatalf("readdir %s: %v", d, err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected owned root dir %s to remain empty, found %d entries", d, len(entries))
		}
	}
}

func TestSummarizeCleanResult_CountsByType(t *testing.T) {
	projectDir := t.TempDir()
	ownedPath := filepath.Join(projectDir, ".claude", "skills")
	if err := os.MkdirAll(ownedPath, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	summary := summarizeCleanResult(
		[]OwnedResourceDir{{Path: ownedPath}, {Path: filepath.Join(projectDir, ".claude", "agents")}},
		[]CleanRemovedEntry{
			{EntryType: "file"},
			{EntryType: "symlink"},
			{EntryType: "directory"},
		},
		[]CleanFailedEntry{{}, {}},
	)

	if summary.OwnedDirsDetected != 2 || summary.OwnedDirsExisting != 1 {
		t.Fatalf("unexpected owned dir counts: %+v", summary)
	}
	if summary.Removed != 3 || summary.RemovedFiles != 1 || summary.RemovedSymlinks != 1 || summary.RemovedDirs != 1 {
		t.Fatalf("unexpected removed counts: %+v", summary)
	}
	if summary.Failed != 2 {
		t.Fatalf("unexpected failed count: %+v", summary)
	}
}

func TestDisplayCleanResult_JSONIncludesDetails(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	result := CleanResult{
		Warnings: []string{"warn"},
		Removed:  []CleanRemovedEntry{{Tool: "claude", ResourceType: "skill", Path: "/tmp/p", EntryType: "symlink"}},
		Failed:   []CleanFailedEntry{{Tool: "claude", ResourceType: "skill", Path: "/tmp/f", EntryType: "file", Error: "boom"}},
		Summary:  CleanSummary{Removed: 1, Failed: 1},
	}

	err = displayCleanResult(result, "json")
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("display json failed: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var parsed CleanResult
	if err := json.Unmarshal(buf[:n], &parsed); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, string(buf[:n]))
	}
	if len(parsed.Removed) != 1 || len(parsed.Failed) != 1 || parsed.Summary.Removed != 1 {
		t.Fatalf("unexpected parsed result: %+v", parsed)
	}
}

func TestCollectCleanWarnings_MissingManifest(t *testing.T) {
	projectDir := t.TempDir()
	warnings := collectCleanWarnings(projectDir)
	if len(warnings) != 1 {
		t.Fatalf("expected one warning, got %d", len(warnings))
	}
	if !strings.Contains(warnings[0], "ai.package.yaml") || !strings.Contains(warnings[0], "will not be able to restore") {
		t.Fatalf("unexpected warning: %s", warnings[0])
	}
}

func TestCleanCommand_HelpDoesNotExposeYesFlag(t *testing.T) {
	if cleanCmd.Flags().Lookup("yes") != nil {
		t.Fatalf("clean command should not expose --yes")
	}
}
