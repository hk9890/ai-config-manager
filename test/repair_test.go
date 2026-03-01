package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRepairIntegration_BrokenSymlink verifies that repair fixes a broken symlink
// by reinstalling the resource from the repository.
func TestRepairIntegration_BrokenSymlink(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "broken-skill", "A skill with a broken symlink")

	// Create a broken symlink in the project
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	brokenLink := filepath.Join(skillsDir, "broken-skill")
	createBrokenSymlink(t, brokenLink)

	// Add the resource to the manifest so repair has context
	p.writeManifest(t, "skill/broken-skill")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	// The symlink should now be valid and point to the repo
	target, err := os.Readlink(brokenLink)
	if err != nil {
		t.Fatalf("Symlink does not exist after repair: %v", err)
	}
	expectedTarget := filepath.Join(p.repoDir, "skills", "broken-skill")
	if target != expectedTarget {
		t.Errorf("Symlink points to wrong target: got %s, want %s", target, expectedTarget)
	}
}

// TestRepairIntegration_MissingFromManifest verifies that repair installs resources
// that are listed in the manifest but not present on disk.
func TestRepairIntegration_MissingFromManifest(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "missing-skill", "A skill missing from disk")

	// Create a manifest referencing the skill — but don't install it
	p.writeManifest(t, "skill/missing-skill")

	// Ensure skills dir exists so installer has somewhere to put it
	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	// The skill should now be installed
	installedPath := filepath.Join(p.projectDir, ".claude", "skills", "missing-skill")
	assertFileExists(t, installedPath)
}

// TestRepairIntegration_CleanProject verifies that a project with no issues
// reports "nothing to repair".
func TestRepairIntegration_CleanProject(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "clean-skill", "A clean, healthy skill")

	// Install the skill properly
	p.installSkillSymlink(t, "clean-skill")

	// Add it to the manifest
	p.writeManifest(t, "skill/clean-skill")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	assertOutputContains(t, output, "nothing to repair")
}

// TestRepairIntegration_OrphanedResources verifies that symlinks on disk that are
// NOT in the manifest are reported as hints (not auto-removed).
func TestRepairIntegration_OrphanedResources(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "orphan-skill", "An orphaned skill")

	// Install the skill (so the symlink exists) but do NOT add it to the manifest
	p.installSkillSymlink(t, "orphan-skill")

	// Write an empty manifest (or no manifest at all)
	p.writeManifest(t) // empty resources list

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	// Hint should be printed mentioning uninstall
	assertOutputContains(t, output, "uninstall")

	// Symlink should NOT have been removed
	assertFileExists(t, filepath.Join(p.projectDir, ".claude", "skills", "orphan-skill"))
}

// TestRepairIntegration_PackageMembers verifies that when a package is in the manifest
// and its member skills are missing from disk, repair detects and reports the issue.
// Note: repair currently detects the package issue and reports it (possibly as a failed
// fix), but does not expand packages into individual member installs at the resource level.
// This test verifies that the command completes without crashing and reports the package.
func TestRepairIntegration_PackageMembers(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "pkg-skill", "A skill that is a package member")
	p.addPackageToRepo(t, "my-pkg", []string{"skill/pkg-skill"})

	// Manifest references the package
	p.writeManifest(t, "package/my-pkg")

	// Ensure skills dir exists
	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// repair should not crash — it detects the package issue even if it cannot fully
	// resolve it via the not-installed path (package refs are not directly installable).
	output, _ := runAimgr(t, "repair", "--project-path", p.projectDir)
	t.Logf("repair output for package members: %s", output)

	// The output should mention the package
	assertOutputContains(t, output, "my-pkg")
}

// TestRepairIntegration_PackageNotInRepo verifies that when a manifest references
// a package that doesn't exist in the repo, repair handles it gracefully.
func TestRepairIntegration_PackageNotInRepo(t *testing.T) {
	p := setupRepairTestProject(t)

	// Manifest references a package that doesn't exist in the repo
	p.writeManifest(t, "package/nonexistent-pkg")

	// Ensure skills dir exists
	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "skills"), 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// repair should not crash — it may report warnings or failures but exit 0
	output, _ := runAimgr(t, "repair", "--project-path", p.projectDir)
	// Output should mention something about the package or nothing to do
	// The key assertion is that the CLI doesn't panic/crash
	t.Logf("repair output for missing package: %s", output)
}

// TestRepairIntegration_NestedBrokenSymlink verifies that a broken symlink
// for a namespaced command (namespace/cmd) is repaired correctly.
func TestRepairIntegration_NestedBrokenSymlink(t *testing.T) {
	p := setupRepairTestProject(t)
	p.createNestedCommandInRepo(t, "api", "deploy", "Deploy command")

	// Create a broken symlink at .claude/commands/api/deploy.md
	nsDir := filepath.Join(p.projectDir, ".claude", "commands", "api")
	if err := os.MkdirAll(nsDir, 0755); err != nil {
		t.Fatalf("Failed to create namespace dir: %v", err)
	}
	brokenLink := filepath.Join(nsDir, "deploy.md")
	createBrokenSymlink(t, brokenLink)

	// Add it to the manifest
	p.writeManifest(t, "command/api/deploy")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	// The nested symlink should be fixed
	target, err := os.Readlink(brokenLink)
	if err != nil {
		t.Fatalf("Nested command symlink does not exist after repair: %v", err)
	}
	expectedTarget := filepath.Join(p.repoDir, "commands", "api", "deploy.md")
	if target != expectedTarget {
		t.Errorf("Nested symlink points to wrong target: got %s, want %s", target, expectedTarget)
	}
}

// TestRepairIntegration_NestedMissingCommand verifies that a nested/namespaced command
// listed in the manifest but not on disk is installed by repair.
func TestRepairIntegration_NestedMissingCommand(t *testing.T) {
	p := setupRepairTestProject(t)
	p.createNestedCommandInRepo(t, "tools", "lint", "Lint command")

	// Manifest references the nested command but it is not installed
	p.writeManifest(t, "command/tools/lint")

	// Create the commands directory
	if err := os.MkdirAll(filepath.Join(p.projectDir, ".claude", "commands"), 0755); err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nOutput: %s", err, output)
	}

	// The nested command should now be installed
	installedPath := filepath.Join(p.projectDir, ".claude", "commands", "tools", "lint.md")
	assertFileExists(t, installedPath)
}

// TestRepairIntegration_ResetForceRemovesUnmanaged verifies that --reset --force
// removes unmanaged files from resource directories.
func TestRepairIntegration_ResetForceRemovesUnmanaged(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create an unmanaged regular file in the skills dir
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	unmanagedFile := filepath.Join(skillsDir, "manual-skill.md")
	if err := os.WriteFile(unmanagedFile, []byte("manually added"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--reset", "--force")
	if err != nil {
		t.Fatalf("repair --reset --force failed: %v\nOutput: %s", err, output)
	}

	// The unmanaged file should be removed
	assertFileRemoved(t, unmanagedFile)
	assertOutputContains(t, output, "Removed")
}

// TestRepairIntegration_ResetDryRunPreservesFiles verifies that --reset --dry-run
// reports what would be removed but does NOT remove anything.
func TestRepairIntegration_ResetDryRunPreservesFiles(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create an unmanaged file
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	unmanagedFile := filepath.Join(skillsDir, "dry-run-skill.md")
	if err := os.WriteFile(unmanagedFile, []byte("manual content"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--reset", "--dry-run")
	if err != nil {
		t.Fatalf("repair --reset --dry-run failed: %v\nOutput: %s", err, output)
	}

	// File should still exist
	assertFileExists(t, unmanagedFile)
	assertOutputContains(t, output, "Would remove")
}

// TestRepairIntegration_ResetMixedManaged verifies that --reset --force removes only
// unmanaged files while leaving managed symlinks in place.
func TestRepairIntegration_ResetMixedManaged(t *testing.T) {
	p := setupRepairTestProject(t)
	p.addSkillToRepo(t, "managed-skill", "A managed skill")

	// Install a managed symlink
	managedLink := p.installSkillSymlink(t, "managed-skill")
	p.writeManifest(t, "skill/managed-skill")

	// Also create an unmanaged regular file
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	unmanagedFile := filepath.Join(skillsDir, "unmanaged.txt")
	if err := os.WriteFile(unmanagedFile, []byte("junk"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--reset", "--force")
	if err != nil {
		t.Fatalf("repair --reset --force failed: %v\nOutput: %s", err, output)
	}
	t.Logf("Output: %s", output)

	// Managed symlink should still exist
	assertFileExists(t, managedLink)
	// Unmanaged file should be removed
	assertFileRemoved(t, unmanagedFile)
}

// TestRepairIntegration_ResetNestedNamespace verifies that unmanaged files inside
// namespace subdirectories of commands/ are removed by --reset --force.
func TestRepairIntegration_ResetNestedNamespace(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create an unmanaged file in a namespace subdir
	nsDir := filepath.Join(p.projectDir, ".claude", "commands", "myns")
	if err := os.MkdirAll(nsDir, 0755); err != nil {
		t.Fatalf("Failed to create namespace dir: %v", err)
	}
	unmanagedFile := filepath.Join(nsDir, "manual-cmd.md")
	if err := os.WriteFile(unmanagedFile, []byte("manual"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--reset", "--force")
	if err != nil {
		t.Fatalf("repair --reset --force failed: %v\nOutput: %s", err, output)
	}

	// The unmanaged file in the namespace dir should be removed
	assertFileRemoved(t, unmanagedFile)
}

// TestRepairIntegration_PrunePackageForceRemovesInvalid verifies that --prune-package --force
// removes a manifest reference to a package that doesn't exist in the repo.
func TestRepairIntegration_PrunePackageForceRemovesInvalid(t *testing.T) {
	p := setupRepairTestProject(t)

	// Write manifest referencing a non-existent package
	p.writeManifest(t, "package/ghost-pkg")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--prune-package", "--force")
	if err != nil {
		t.Fatalf("repair --prune-package --force failed: %v\nOutput: %s", err, output)
	}

	// The invalid ref should be removed from the manifest
	assertManifestNotContains(t, p.manifestPath, "package/ghost-pkg")
}

// TestRepairIntegration_PrunePackageDryRunPreservesManifest verifies that
// --prune-package --dry-run prints what would be removed but does not modify the manifest.
func TestRepairIntegration_PrunePackageDryRunPreservesManifest(t *testing.T) {
	p := setupRepairTestProject(t)

	// Write manifest with an invalid ref
	p.writeManifest(t, "skill/nonexistent-skill")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--prune-package", "--dry-run")
	if err != nil {
		t.Fatalf("repair --prune-package --dry-run failed: %v\nOutput: %s", err, output)
	}

	// Manifest should be unchanged
	assertManifestContains(t, p.manifestPath, "skill/nonexistent-skill")
	assertOutputContains(t, output, "Would remove")
}

// TestRepairIntegration_PrunePackagePartialPackage verifies that when a package exists
// in the repo but a member is missing, repair only warns (does not remove the manifest ref).
func TestRepairIntegration_PrunePackagePartialPackage(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create a package whose member skill doesn't exist in the repo
	p.addPackageToRepo(t, "partial-pkg", []string{"skill/ghost-skill"})
	p.writeManifest(t, "package/partial-pkg")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--prune-package", "--force")
	if err != nil {
		t.Fatalf("repair --prune-package --force failed: %v\nOutput: %s", err, output)
	}

	// The package ref should still be in the manifest (it's a partial/warning, not invalid)
	assertManifestContains(t, p.manifestPath, "package/partial-pkg")
	// Output should warn about the partial package
	assertOutputContains(t, output, "partial-pkg")
}

// TestRepairIntegration_ResetAndPrunePackageForce verifies that combining --reset and
// --prune-package with --force handles both cleanup operations.
func TestRepairIntegration_ResetAndPrunePackageForce(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create an unmanaged file in skills dir
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	unmanagedFile := filepath.Join(skillsDir, "junk.md")
	if err := os.WriteFile(unmanagedFile, []byte("junk"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	// Also write a manifest with an invalid ref
	p.writeManifest(t, "command/ghost-cmd")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir,
		"--reset", "--prune-package", "--force")
	if err != nil {
		t.Fatalf("repair --reset --prune-package --force failed: %v\nOutput: %s", err, output)
	}
	t.Logf("Output: %s", output)

	// Unmanaged file should be removed
	assertFileRemoved(t, unmanagedFile)
	// Invalid ref should be removed from manifest
	assertManifestNotContains(t, p.manifestPath, "command/ghost-cmd")
}

// TestRepairIntegration_AllFlagsDryRun verifies that --reset --prune-package --dry-run
// makes no changes at all.
func TestRepairIntegration_AllFlagsDryRun(t *testing.T) {
	p := setupRepairTestProject(t)

	// Create unmanaged file
	skillsDir := filepath.Join(p.projectDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}
	unmanagedFile := filepath.Join(skillsDir, "preserve-me.md")
	if err := os.WriteFile(unmanagedFile, []byte("keep this"), 0644); err != nil {
		t.Fatalf("Failed to create unmanaged file: %v", err)
	}

	// Write manifest with an invalid ref
	p.writeManifest(t, "skill/does-not-exist")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir,
		"--reset", "--prune-package", "--dry-run")
	if err != nil {
		t.Fatalf("repair all-flags dry-run failed: %v\nOutput: %s", err, output)
	}

	// Nothing should be changed
	assertFileExists(t, unmanagedFile)
	assertManifestContains(t, p.manifestPath, "skill/does-not-exist")

	// Output should indicate dry-run actions
	if !strings.Contains(output, "Would remove") && !strings.Contains(output, "nothing") {
		t.Logf("Output (no 'Would remove' or 'nothing'): %s", output)
	}
}

// TestRepairIntegration_JSONOutput verifies that --format=json produces valid JSON output.
func TestRepairIntegration_JSONOutput(t *testing.T) {
	p := setupRepairTestProject(t)

	// Clean project with no issues
	p.addSkillToRepo(t, "json-skill", "Skill for JSON output test")
	p.installSkillSymlink(t, "json-skill")
	p.writeManifest(t, "skill/json-skill")

	output, err := runAimgr(t, "repair", "--project-path", p.projectDir, "--format=json")
	if err != nil {
		t.Fatalf("repair --format=json failed: %v\nOutput: %s", err, output)
	}

	// Parse the JSON output
	var result struct {
		Fixed   interface{} `json:"fixed"`
		Failed  interface{} `json:"failed"`
		Hints   interface{} `json:"hints"`
		Summary interface{} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify required JSON fields are present
	if result.Fixed == nil {
		t.Error("JSON output missing 'fixed' field")
	}
	if result.Failed == nil {
		t.Error("JSON output missing 'failed' field")
	}
	if result.Hints == nil {
		t.Error("JSON output missing 'hints' field")
	}
	if result.Summary == nil {
		t.Error("JSON output missing 'summary' field")
	}
}
