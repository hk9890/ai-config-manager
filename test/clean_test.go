package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanIntegration_WipesOwnedDirsAndPreservesManifestAndConfig(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t, "skill/declared")

	manifestBefore, err := os.ReadFile(p.manifestPath)
	if err != nil {
		t.Fatalf("read manifest before: %v", err)
	}

	commandsDir := filepath.Join(p.projectDir, ".claude", "commands")
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	agentsDir := filepath.Join(p.projectDir, ".claude", "agents")
	for _, d := range []string{commandsDir, skillsDir, agentsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(commandsDir, "manual.md"), []byte("manual"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Symlink(filepath.Join(p.projectDir, "missing-target"), filepath.Join(skillsDir, "broken-link")); err != nil {
		t.Fatalf("create broken symlink: %v", err)
	}
	wrongTarget := filepath.Join(p.projectDir, "outside.txt")
	if err := os.WriteFile(wrongTarget, []byte("outside"), 0644); err != nil {
		t.Fatalf("write wrong target: %v", err)
	}
	if err := os.Symlink(wrongTarget, filepath.Join(skillsDir, "wrong-repo-link")); err != nil {
		t.Fatalf("create wrong-repo symlink: %v", err)
	}
	nested := filepath.Join(agentsDir, "nested", "inner.txt")
	if err := os.MkdirAll(filepath.Dir(nested), 0755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(nested, []byte("nested"), 0644); err != nil {
		t.Fatalf("write nested: %v", err)
	}

	toolConfig := filepath.Join(p.projectDir, ".claude", "settings.json")
	if err := os.WriteFile(toolConfig, []byte(`{"keep":true}`), 0644); err != nil {
		t.Fatalf("write tool config: %v", err)
	}

	out, err := runAimgr(t, "clean", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("clean failed: %v\nOutput: %s", err, out)
	}

	for _, d := range []string{commandsDir, skillsDir, agentsDir} {
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatalf("readdir %s: %v", d, err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected %s to be empty, found %d entries", d, len(entries))
		}
	}

	assertFileExists(t, toolConfig)

	manifestAfter, err := os.ReadFile(p.manifestPath)
	if err != nil {
		t.Fatalf("read manifest after: %v", err)
	}
	if string(manifestBefore) != string(manifestAfter) {
		t.Fatalf("manifest changed after clean\nbefore:\n%s\nafter:\n%s", string(manifestBefore), string(manifestAfter))
	}

	assertOutputContains(t, out, "Summary:")
	assertOutputContains(t, out, "removed=")
}

func TestCleanIntegration_NoToolDirsNoOpMessage(t *testing.T) {
	projectDir := t.TempDir()
	repoDir := t.TempDir()
	configDir := t.TempDir()
	t.Setenv("AIMGR_REPO_PATH", repoDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", repoDir)

	out, err := runAimgr(t, "clean", "--project-path", projectDir)
	if err != nil {
		t.Fatalf("clean should succeed with no tool dirs: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(out, "No tool directories found in this project. Nothing to clean.") {
		t.Fatalf("expected no-op message, got: %s", out)
	}
}

func TestCleanIntegration_JSONAndMissingManifestWarning(t *testing.T) {
	p := setupRepairTestProject(t)

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "manual.md"), []byte("manual"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := runAimgr(t, "clean", "--project-path", p.projectDir, "--format=json")
	if err != nil {
		t.Fatalf("clean json failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "ai.package.yaml not found") {
		t.Fatalf("expected missing manifest warning, got: %s", out)
	}

	var parsed struct {
		Warnings []string `json:"warnings"`
		Removed  []any    `json:"removed"`
		Summary  struct {
			Removed int `json:"removed"`
		} `json:"summary"`
	}
	jsonStart := strings.Index(out, "{")
	if jsonStart == -1 {
		t.Fatalf("no json object found in output: %s", out)
	}
	if err := json.Unmarshal([]byte(out[jsonStart:]), &parsed); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(parsed.Warnings) == 0 {
		t.Fatalf("expected warnings in json output")
	}
	if parsed.Summary.Removed == 0 || len(parsed.Removed) == 0 {
		t.Fatalf("expected removed entries in json output: %+v", parsed)
	}
}

func TestCleanIntegration_YesFlagRemoved(t *testing.T) {
	p := setupRepairTestProject(t)
	out, err := runAimgr(t, "clean", "--project-path", p.projectDir, "--yes")
	if err == nil {
		t.Fatalf("expected --yes to fail, got success: %s", out)
	}
	if !strings.Contains(out, "unknown flag") || !strings.Contains(out, "--yes") {
		t.Fatalf("expected unknown flag error for --yes, got: %s", out)
	}
}
