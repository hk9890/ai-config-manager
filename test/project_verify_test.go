package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectVerifyIntegration_ReportsUndeclaredAndNotInstalled(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t, "skill/declared-skill")

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	undeclared := filepath.Join(skillsDir, "manual.md")
	if err := os.WriteFile(undeclared, []byte("manual"), 0644); err != nil {
		t.Fatalf("write undeclared: %v", err)
	}

	out, err := runAimgr(t, "verify", "--project-path", p.projectDir, "--format=json")
	if err != nil {
		t.Fatalf("verify failed: %v\nOutput: %s", err, out)
	}

	var issues []struct {
		IssueType string `json:"IssueType"`
		Resource  string `json:"Resource"`
	}
	if err := json.Unmarshal([]byte(out), &issues); err != nil {
		t.Fatalf("failed to parse verify json: %v\n%s", err, out)
	}

	foundNotInstalled := false
	foundUndeclared := false
	for _, issue := range issues {
		if issue.IssueType == "not-installed" && issue.Resource == "skill/declared-skill" {
			foundNotInstalled = true
		}
		if issue.IssueType == "undeclared" && issue.Resource == "manual.md" {
			foundUndeclared = true
		}
	}

	if !foundNotInstalled {
		t.Fatalf("expected not-installed issue for declared skill, got: %+v", issues)
	}
	if !foundUndeclared {
		t.Fatalf("expected undeclared issue for manual.md, got: %+v", issues)
	}
}

func TestProjectVerifyIntegration_DetectsBrokenAndWrongRepo(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t)

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	brokenPath := filepath.Join(skillsDir, "broken-skill")
	if err := os.Symlink(filepath.Join(p.repoDir, "skills", "does-not-exist"), brokenPath); err != nil {
		t.Fatalf("symlink broken: %v", err)
	}

	wrongTargetDir := filepath.Join(p.projectDir, "external", "skills", "wrong-skill")
	if err := os.MkdirAll(wrongTargetDir, 0755); err != nil {
		t.Fatalf("mkdir wrong target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wrongTargetDir, "SKILL.md"), []byte("wrong"), 0644); err != nil {
		t.Fatalf("write wrong target: %v", err)
	}
	wrongPath := filepath.Join(skillsDir, "wrong-skill")
	if err := os.Symlink(wrongTargetDir, wrongPath); err != nil {
		t.Fatalf("symlink wrong-repo: %v", err)
	}

	out, err := runAimgr(t, "verify", "--project-path", p.projectDir, "--format=json")
	if err != nil {
		t.Fatalf("verify failed: %v\nOutput: %s", err, out)
	}

	var issues []struct {
		IssueType string `json:"IssueType"`
	}
	if err := json.Unmarshal([]byte(out), &issues); err != nil {
		t.Fatalf("failed to parse verify json: %v\n%s", err, out)
	}

	foundBroken := false
	foundWrongRepo := false
	for _, issue := range issues {
		if issue.IssueType == "broken" {
			foundBroken = true
		}
		if issue.IssueType == "wrong-repo" {
			foundWrongRepo = true
		}
	}

	if !foundBroken || !foundWrongRepo {
		t.Fatalf("expected broken and wrong-repo issues, got: %+v", issues)
	}
}

func TestProjectVerifyIntegration_FixWrapsRepairReconciliation(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "declared-skill", "declared")
	p.writeManifest(t, "skill/declared-skill")

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	undeclared := filepath.Join(skillsDir, "manual.md")
	if err := os.WriteFile(undeclared, []byte("manual"), 0644); err != nil {
		t.Fatalf("write undeclared: %v", err)
	}

	out, err := runAimgr(t, "verify", "--project-path", p.projectDir, "--fix")
	if err != nil {
		t.Fatalf("verify --fix failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "deprecated") || !strings.Contains(out, "repair") {
		t.Fatalf("expected deprecation warning output, got: %s", out)
	}

	assertFileExists(t, filepath.Join(skillsDir, "declared-skill"))
	assertFileRemoved(t, undeclared)
}

func TestProjectVerifyIntegration_HelpAndOutputAvoidRemovedFlags(t *testing.T) {
	p := setupRepairTestProject(t)
	p.writeManifest(t)

	helpOut, err := runAimgr(t, "verify", "--help")
	if err != nil {
		t.Fatalf("verify --help failed: %v\nOutput: %s", err, helpOut)
	}

	if strings.Contains(helpOut, "--reset") || strings.Contains(helpOut, "--force") || strings.Contains(helpOut, "--yes") {
		t.Fatalf("verify help references removed flags: %s", helpOut)
	}

	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "manual.md"), []byte("manual"), 0644); err != nil {
		t.Fatalf("write undeclared: %v", err)
	}

	verifyOut, err := runAimgr(t, "verify", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("verify failed: %v\nOutput: %s", err, verifyOut)
	}

	if strings.Contains(verifyOut, "--reset") || strings.Contains(verifyOut, "--force") || strings.Contains(verifyOut, "--yes") {
		t.Fatalf("verify output references removed flags: %s", verifyOut)
	}
	if !strings.Contains(verifyOut, "aimgr repair") {
		t.Fatalf("verify output should guide to aimgr repair, got: %s", verifyOut)
	}
}
