package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairIntegration_MissingManifestFailsWithoutChanges(t *testing.T) {
	p := setupRepairTestProject(t)
	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	undeclared := filepath.Join(p.projectDir, ".claude", "skills", "manual.md")
	if err := os.WriteFile(undeclared, []byte("manual"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err == nil {
		t.Fatalf("expected missing manifest error, got success with output: %s", output)
	}
	assertFileExists(t, undeclared)
}

func TestRepairIntegration_ReconcileInstallsAndRemovesUndeclared(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "declared-skill", "declared")
	p.writeManifest(t, "skill/declared-skill")

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	undeclared := filepath.Join(skillsDir, "undeclared.md")
	if err := os.WriteFile(undeclared, []byte("manual"), 0644); err != nil {
		t.Fatalf("write undeclared: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	assertFileExists(t, filepath.Join(skillsDir, "declared-skill"))
	assertFileRemoved(t, undeclared)
}

func TestRepairIntegration_DryRunShowsPlanAndChangesNothing(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "declared-skill", "declared")
	p.writeManifest(t, "skill/declared-skill")

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	undeclared := filepath.Join(skillsDir, "undeclared.md")
	if err := os.WriteFile(undeclared, []byte("manual"), 0644); err != nil {
		t.Fatalf("write undeclared: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--dry-run")
	if err != nil {
		t.Fatalf("repair --dry-run failed: %v\nOutput: %s", err, output)
	}

	assertFileExists(t, undeclared)
	assertFileRemoved(t, filepath.Join(skillsDir, "declared-skill"))
	assertOutputContains(t, output, "dry-run")
	assertOutputContains(t, output, "Removals")
	assertOutputContains(t, output, "Installs")
}

func TestRepairIntegration_PackageExpansionInstallsMembers(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "pkg-member", "member")
	p.addPackageToRepo(t, "my-pkg", []string{"skill/pkg-member"})
	p.writeManifest(t, "package/my-pkg")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}
	assertFileExists(t, filepath.Join(p.projectDir, ".claude", "skills", "pkg-member"))
}

func TestRepairIntegration_ConflictingPathIsReplaced(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "conflict-skill", "member")
	p.writeManifest(t, "skill/conflict-skill")

	skillPath := filepath.Join(p.projectDir, ".claude", "skills", "conflict-skill")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(skillPath, []byte("manual file"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	info, err := os.Lstat(skillPath)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected conflicting file to be replaced by symlink")
	}
}

func TestRepairIntegration_PrunePackageDryRunAndApply(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t, "skill/missing", "skill/missing2")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--prune-package", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run prune failed: %v\nOutput: %s", err, output)
	}
	assertManifestContains(t, p.manifestPath, "skill/missing")
	assertManifestContains(t, p.manifestPath, "skill/missing2")

	output, err = runAimgr(t, "repair", "--project-path", p.projectDir, "--prune-package")
	if err != nil {
		t.Fatalf("apply prune failed: %v\nOutput: %s", err, output)
	}
	assertManifestNotContains(t, p.manifestPath, "skill/missing")
	assertManifestNotContains(t, p.manifestPath, "skill/missing2")
}

func TestRepairIntegration_JSONOutputIncludesPlanSchema(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "declared-skill", "declared")
	p.writeManifest(t, "skill/declared-skill")

	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	out, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--format=json", "--dry-run")
	if err != nil {
		t.Fatalf("repair json failed: %v\nOutput: %s", err, out)
	}

	var parsed struct {
		DryRun  bool `json:"dry_run"`
		Planned struct {
			Installs []any `json:"installs"`
			Fixes    []any `json:"fixes"`
			Removals []any `json:"removals"`
		} `json:"planned"`
		Applied any `json:"applied"`
		Summary any `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid json output: %v\n%s", err, out)
	}
	if !parsed.DryRun {
		t.Fatalf("expected dry_run=true")
	}
	if parsed.Applied == nil || parsed.Summary == nil {
		t.Fatalf("expected applied and summary in json output")
	}
}

func TestRepairIntegration_RemovedFlagsUnavailable(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t)

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--reset")
	if err == nil {
		t.Fatalf("expected --reset to fail, got success: %s", output)
	}
	if !strings.Contains(output, "unknown flag") || !strings.Contains(output, "--reset") {
		t.Fatalf("expected unknown flag error for --reset, got: %s", output)
	}

	output, err = runAimgr(t, "repair", "--project-path", p.projectDir, "--force")
	if err == nil {
		t.Fatalf("expected --force to fail, got success: %s", output)
	}
	if !strings.Contains(output, "unknown flag") || !strings.Contains(output, "--force") {
		t.Fatalf("expected unknown flag error for --force, got: %s", output)
	}
}
