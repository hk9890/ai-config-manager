//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/metadata"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repomanifest"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/resource"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/sourcemetadata"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/workspace"
)

func TestE2E_OverrideSource_RemoteBaselineOverrideClear(t *testing.T) {
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := t.TempDir()
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}

	remoteOrigin, remoteWorktree := createE2ERemoteGitSource(t)
	remoteURL := "https://example.com/e2e/team-tools.git"
	localOverrideDir := t.TempDir()

	writeE2ERemoteCommand(t, remoteWorktree, "source-shared", "remote baseline")
	seedE2ERemoteCache(t, repoPath, remoteURL, remoteOrigin)
	writeE2ECommandFixture(t, localOverrideDir, "source-shared", "local override")

	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{{
		Name: "team-tools",
		URL:  remoteURL,
	}}}
	if err := manifest.Save(repoPath); err != nil {
		t.Fatalf("failed to save baseline manifest: %v", err)
	}

	stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "sync")
	if err != nil {
		t.Fatalf("baseline sync failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertE2ECommandContains(t, repoPath, "source-shared", "remote baseline")

	stdout, stderr, err = runAimgrWithEnv(t, configPath, env, "repo", "override-source", "team-tools", "local:"+localOverrideDir)
	if err != nil {
		t.Fatalf("override-source failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "overridden") {
		t.Fatalf("expected override success output, got:\n%s", stdout)
	}

	assertE2ECommandContains(t, repoPath, "source-shared", "local override")

	meta, err := sourcemetadata.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load source metadata after override: %v", err)
	}
	state := meta.Get("team-tools")
	if state == nil || state.OverrideOriginalURL == "" {
		t.Fatalf("expected override breadcrumb metadata after override, state=%+v", state)
	}

	manifestOut, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "show-manifest")
	if err != nil {
		t.Fatalf("show-manifest failed: %v\nstdout:\n%s\nstderr:\n%s", err, manifestOut, stderr)
	}
	if strings.Contains(manifestOut, localOverrideDir) {
		t.Fatalf("show-manifest must not leak local override path, got:\n%s", manifestOut)
	}
	for _, expected := range []string{"name: team-tools", "url: " + remoteURL} {
		if !strings.Contains(manifestOut, expected) {
			t.Fatalf("expected show-manifest output to contain %q, got:\n%s", expected, manifestOut)
		}
	}

	infoJSON, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "info", "--format=json")
	if err != nil {
		t.Fatalf("repo info --format=json failed: %v\nstdout:\n%s\nstderr:\n%s", err, infoJSON, stderr)
	}
	if !strings.Contains(infoJSON, `"overridden": true`) {
		t.Fatalf("expected repo info json to show override state, got:\n%s", infoJSON)
	}
	if !strings.Contains(infoJSON, remoteURL) {
		t.Fatalf("expected repo info json restore target to include baseline location, got:\n%s", infoJSON)
	}

	stdout, stderr, err = runAimgrWithEnv(t, configPath, env, "repo", "override-source", "team-tools", "--clear")
	if err != nil {
		t.Fatalf("override-source --clear failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertE2ECommandContains(t, repoPath, "source-shared", "remote baseline")

	metaAfterClear, err := sourcemetadata.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load source metadata after clear: %v", err)
	}
	stateAfterClear := metaAfterClear.Get("team-tools")
	if stateAfterClear != nil && stateAfterClear.OverrideOriginalURL != "" {
		t.Fatalf("expected override breadcrumb metadata cleared, state=%+v", stateAfterClear)
	}
}

func TestE2E_OverrideSource_RemoveOverriddenSource(t *testing.T) {
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := t.TempDir()
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}

	remoteOrigin, remoteWorktree := createE2ERemoteGitSource(t)
	remoteURL := "https://example.com/e2e/team-tools.git"
	localOverrideDir := t.TempDir()

	writeE2ERemoteCommand(t, remoteWorktree, "source-shared", "remote baseline")
	seedE2ERemoteCache(t, repoPath, remoteURL, remoteOrigin)
	writeE2ECommandFixture(t, localOverrideDir, "source-shared", "local override")

	manifest := &repomanifest.Manifest{Version: 1, Sources: []*repomanifest.Source{{
		Name: "team-tools",
		URL:  remoteURL,
	}}}
	if err := manifest.Save(repoPath); err != nil {
		t.Fatalf("failed to save baseline manifest: %v", err)
	}

	if _, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "sync"); err != nil {
		t.Fatalf("baseline sync failed: %v\nstderr:\n%s", err, stderr)
	}
	if _, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "override-source", "team-tools", "local:"+localOverrideDir); err != nil {
		t.Fatalf("override failed: %v\nstderr:\n%s", err, stderr)
	}

	removeOut, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "remove", "team-tools")
	if err != nil {
		t.Fatalf("repo remove failed: %v\nstdout:\n%s\nstderr:\n%s", err, removeOut, stderr)
	}
	if !strings.Contains(removeOut, "both the active override and restore target") {
		t.Fatalf("expected override removal warning in output, got:\n%s", removeOut)
	}

	manifestAfter, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest after remove: %v", err)
	}
	if len(manifestAfter.Sources) != 0 {
		t.Fatalf("expected no sources after remove, got %d", len(manifestAfter.Sources))
	}

	resourcePath := filepath.Join(repoPath, "commands", "source-shared.md")
	if _, err := os.Stat(resourcePath); !os.IsNotExist(err) {
		t.Fatalf("expected source-owned resource to be removed with source, stat err=%v", err)
	}

	metaPath := metadata.GetMetadataPath("source-shared", resource.Command, repoPath)
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Fatalf("expected resource metadata removed, stat err=%v", err)
	}
}

func writeE2ECommandFixture(t *testing.T, sourceDir, name, description string) {
	t.Helper()
	commandsDir := filepath.Join(sourceDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}
	content := "---\ndescription: " + description + "\n---\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(commandsDir, name+".md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write command fixture: %v", err)
	}
}

func assertE2ECommandContains(t *testing.T, repoPath, commandName, expected string) {
	t.Helper()
	path := filepath.Join(repoPath, "commands", commandName+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read command %s: %v", commandName, err)
	}
	if !strings.Contains(string(data), expected) {
		t.Fatalf("expected command %s to contain %q, got:\n%s", commandName, expected, string(data))
	}
}

func runE2EGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func createE2ERemoteGitSource(t *testing.T) (remoteURL string, worktreePath string) {
	t.Helper()

	baseDir := t.TempDir()
	bareRepo := filepath.Join(baseDir, "remote.git")
	worktreePath = filepath.Join(baseDir, "worktree")

	runE2EGit(t, baseDir, "init", "--bare", bareRepo)
	runE2EGit(t, bareRepo, "symbolic-ref", "HEAD", "refs/heads/main")
	runE2EGit(t, baseDir, "clone", bareRepo, worktreePath)
	runE2EGit(t, worktreePath, "branch", "-M", "main")

	return bareRepo, worktreePath
}

func writeE2ERemoteCommand(t *testing.T, worktreePath, name, description string) {
	t.Helper()

	cmdDir := filepath.Join(worktreePath, "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create commands dir: %v", err)
	}

	content := "---\ndescription: " + description + "\n---\n# " + name + "\n"
	cmdPath := filepath.Join(cmdDir, name+".md")
	if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write remote command: %v", err)
	}

	runE2EGit(t, worktreePath, "add", ".")
	runE2EGit(t, worktreePath, "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", "update "+name)
	runE2EGit(t, worktreePath, "push", "origin", "main")
}

func seedE2ERemoteCache(t *testing.T, repoPath, remoteURL, remoteOrigin string) {
	t.Helper()

	cacheRepoPath := filepath.Join(repoPath, ".workspace", workspace.ComputeHash(remoteURL))
	if err := os.MkdirAll(filepath.Dir(cacheRepoPath), 0755); err != nil {
		t.Fatalf("failed to create workspace directory: %v", err)
	}
	runE2EGit(t, repoPath, "clone", "-b", "main", remoteOrigin, cacheRepoPath)
	runE2EGit(t, cacheRepoPath, "checkout", "main")
}
