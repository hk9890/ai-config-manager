//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/repomanifest"
	"github.com/dynatrace-oss/ai-config-manager/v3/pkg/sourcemetadata"
)

type asyncCLIProcess struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	done   chan error
}

func startAimgrAsync(t *testing.T, repoPath string, extraEnv map[string]string, args ...string) *asyncCLIProcess {
	t.Helper()

	binPath := buildTestBinary(t)
	cmd := exec.Command(binPath, args...)

	env := append(os.Environ(), "AIMGR_REPO_PATH="+repoPath)
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start aimgr %v: %v", args, err)
	}

	p := &asyncCLIProcess{cmd: cmd, stdout: stdout, stderr: stderr, done: make(chan error, 1)}
	go func() { p.done <- cmd.Wait() }()
	return p
}

func (p *asyncCLIProcess) wait(t *testing.T, timeout time.Duration) (string, string, error) {
	t.Helper()
	select {
	case err := <-p.done:
		return p.stdout.String(), p.stderr.String(), err
	case <-time.After(timeout):
		_ = p.cmd.Process.Kill()
		_, _, err := p.wait(t, 2*time.Second)
		return p.stdout.String(), p.stderr.String(), fmt.Errorf("timed out waiting for process: %w", err)
	}
}

func (p *asyncCLIProcess) assertStillRunning(t *testing.T) {
	t.Helper()
	select {
	case err := <-p.done:
		t.Fatalf("expected process to be blocked and still running, exited early: %v\nstdout:\n%s\nstderr:\n%s", err, p.stdout.String(), p.stderr.String())
	default:
	}
}

func waitForMarker(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for marker: %s", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func releaseMarker(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("release"), 0644); err != nil {
		t.Fatalf("failed to write release marker %s: %v", path, err)
	}
}

func requireCLIProcessSuccess(t *testing.T, stdout, stderr string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("process failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
}

func createSingleSourceTree(t *testing.T, baseDir, commandName string) string {
	t.Helper()
	sourceDir := filepath.Join(baseDir, commandName+"-source")
	commandsDir := filepath.Join(sourceDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("failed to create commands directory: %v", err)
	}
	content := fmt.Sprintf("---\ndescription: %s\n---\n# %s\n", commandName, commandName)
	if err := os.WriteFile(filepath.Join(commandsDir, commandName+".md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write command file: %v", err)
	}
	return sourceDir
}

func requireManifestHasSourcePath(t *testing.T, repoPath, sourcePath string) {
	t.Helper()
	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}
	abs, _ := filepath.Abs(sourcePath)
	for _, src := range manifest.Sources {
		if src.Path == abs {
			return
		}
	}
	t.Fatalf("manifest missing expected source path: %s", abs)
}

func TestE2E_ConcurrentRepoInitSerializesAndRemainsConsistent(t *testing.T) {
	repoPath := t.TempDir()
	signalDir := t.TempDir()

	proc1 := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "init",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "init")

	waitForMarker(t, filepath.Join(signalDir, "init.ready"), 5*time.Second)
	proc2 := startAimgrAsync(t, repoPath, nil, "repo", "init")
	proc2.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "init.release"))

	stdout1, stderr1, err1 := proc1.wait(t, 10*time.Second)
	requireCLIProcessSuccess(t, stdout1, stderr1, err1)
	stdout2, stderr2, err2 := proc2.wait(t, 10*time.Second)
	requireCLIProcessSuccess(t, stdout2, stderr2, err2)

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
		t.Fatalf("expected .git to exist after concurrent init: %v", err)
	}
	if _, err := repomanifest.Load(repoPath); err != nil {
		t.Fatalf("manifest corrupted after concurrent init: %v", err)
	}
}

func TestE2E_ConcurrentRepoSyncSerializesAndPreservesMetadata(t *testing.T) {
	repoPath := t.TempDir()
	sourceDir := createSingleSourceTree(t, t.TempDir(), "sync-concurrency-cmd")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)
	stdout, stderr, err = runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "add", "local:"+sourceDir)
	requireCLIProcessSuccess(t, stdout, stderr, err)

	signalDir := t.TempDir()
	proc1 := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "sync",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "sync")

	waitForMarker(t, filepath.Join(signalDir, "sync.ready"), 5*time.Second)
	proc2 := startAimgrAsync(t, repoPath, nil, "repo", "sync")
	proc2.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "sync.release"))

	stdout1, stderr1, err1 := proc1.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout1, stderr1, err1)
	stdout2, stderr2, err2 := proc2.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout2, stderr2, err2)

	if _, err := os.Stat(filepath.Join(repoPath, "commands", "sync-concurrency-cmd.md")); err != nil {
		t.Fatalf("expected synced command to exist: %v", err)
	}
	meta, err := sourcemetadata.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load source metadata: %v", err)
	}
	if len(meta.Sources) == 0 {
		t.Fatalf("expected source metadata entries after sync")
	}
}

func TestE2E_RepoVerifyWaitsBehindSyncAndStateRemainsConsistent(t *testing.T) {
	repoPath := t.TempDir()
	sourceDir := createSingleSourceTree(t, t.TempDir(), "verify-vs-sync-cmd")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)
	stdout, stderr, err = runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "add", "local:"+sourceDir)
	requireCLIProcessSuccess(t, stdout, stderr, err)

	signalDir := t.TempDir()
	syncProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "sync",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "sync")

	waitForMarker(t, filepath.Join(signalDir, "sync.ready"), 5*time.Second)
	verifyProc := startAimgrAsync(t, repoPath, nil, "repo", "verify", "--format=json")
	verifyProc.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "sync.release"))

	stdoutSync, stderrSync, errSync := syncProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdoutSync, stderrSync, errSync)
	stdoutVerify, stderrVerify, errVerify := verifyProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdoutVerify, stderrVerify, errVerify)

	if _, err := os.Stat(filepath.Join(repoPath, "commands", "verify-vs-sync-cmd.md")); err != nil {
		t.Fatalf("expected synced command to exist after contention: %v", err)
	}
	if _, err := repomanifest.Load(repoPath); err != nil {
		t.Fatalf("manifest corrupted after sync/verify contention: %v", err)
	}
	if _, err := sourcemetadata.Load(repoPath); err != nil {
		t.Fatalf("source metadata corrupted after sync/verify contention: %v", err)
	}
}

func TestE2E_ConcurrentRepoReadersRunTogether(t *testing.T) {
	repoPath := t.TempDir()
	sourceDir := createSingleSourceTree(t, t.TempDir(), "reader-concurrency-cmd")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)
	stdout, stderr, err = runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "add", "local:"+sourceDir)
	requireCLIProcessSuccess(t, stdout, stderr, err)

	reader1Signals := t.TempDir()
	verifyProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":     "verify-read",
		"AIMGR_TEST_REPO_SIGNAL_DIR":  reader1Signals,
		"AIMGR_TEST_REPO_SIGNAL_NAME": "verify-reader",
	}, "repo", "verify", "--format=json")

	waitForMarker(t, filepath.Join(reader1Signals, "verify-reader.ready"), 5*time.Second)

	reader2Signals := t.TempDir()
	listProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":     "list",
		"AIMGR_TEST_REPO_SIGNAL_DIR":  reader2Signals,
		"AIMGR_TEST_REPO_SIGNAL_NAME": "list-reader",
	}, "repo", "list", "--format=json")

	// If list is blocked by verify's read-lock, this marker would not appear.
	waitForMarker(t, filepath.Join(reader2Signals, "list-reader.ready"), 5*time.Second)

	releaseMarker(t, filepath.Join(reader2Signals, "list-reader.release"))
	releaseMarker(t, filepath.Join(reader1Signals, "verify-reader.release"))

	stdoutList, stderrList, errList := listProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdoutList, stderrList, errList)
	stdoutVerify, stderrVerify, errVerify := verifyProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdoutVerify, stderrVerify, errVerify)
}

func TestE2E_ConcurrentAddAndRemoveSerializeWithoutLosingState(t *testing.T) {
	repoPath := t.TempDir()
	sourceA := createSingleSourceTree(t, t.TempDir(), "remove-me")
	sourceB := createSingleSourceTree(t, t.TempDir(), "keep-me")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)
	stdout, stderr, err = runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "add", "local:"+sourceA)
	requireCLIProcessSuccess(t, stdout, stderr, err)

	manifest, err := repomanifest.Load(repoPath)
	if err != nil || len(manifest.Sources) != 1 {
		t.Fatalf("expected exactly one source before concurrent add/remove: err=%v count=%d", err, len(manifest.Sources))
	}
	sourceToRemove := manifest.Sources[0].Name

	signalDir := t.TempDir()
	addProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "add",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "add", "local:"+sourceB)

	waitForMarker(t, filepath.Join(signalDir, "add.ready"), 5*time.Second)
	removeProc := startAimgrAsync(t, repoPath, nil, "repo", "remove", sourceToRemove)
	removeProc.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "add.release"))

	stdout1, stderr1, err1 := addProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout1, stderr1, err1)
	stdout2, stderr2, err2 := removeProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout2, stderr2, err2)

	if _, err := os.Stat(filepath.Join(repoPath, "commands", "remove-me.md")); !os.IsNotExist(err) {
		t.Fatalf("expected removed command to be gone, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoPath, "commands", "keep-me.md")); err != nil {
		t.Fatalf("expected added command to exist, stat err=%v", err)
	}
	requireManifestHasSourcePath(t, repoPath, sourceB)
}

func TestE2E_ConcurrentApplyManifestAndAddSerializeAndKeepBothSources(t *testing.T) {
	repoPath := t.TempDir()
	sourceFromManifest := createSingleSourceTree(t, t.TempDir(), "from-manifest")
	sourceFromAdd := createSingleSourceTree(t, t.TempDir(), "from-add")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)

	manifestPath := filepath.Join(t.TempDir(), repomanifest.ManifestFileName)
	manifestBody := fmt.Sprintf("version: 1\nsources:\n  - name: manifest-source\n    path: %s\n", sourceFromManifest)
	if err := os.WriteFile(manifestPath, []byte(manifestBody), 0644); err != nil {
		t.Fatalf("failed to write input manifest: %v", err)
	}

	signalDir := t.TempDir()
	applyProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "apply-manifest",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "apply-manifest", manifestPath)

	waitForMarker(t, filepath.Join(signalDir, "apply-manifest.ready"), 5*time.Second)
	addProc := startAimgrAsync(t, repoPath, nil, "repo", "add", "local:"+sourceFromAdd)
	addProc.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "apply-manifest.release"))

	stdout1, stderr1, err1 := applyProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout1, stderr1, err1)
	stdout2, stderr2, err2 := addProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout2, stderr2, err2)

	requireManifestHasSourcePath(t, repoPath, sourceFromManifest)
	requireManifestHasSourcePath(t, repoPath, sourceFromAdd)
	if _, err := os.Stat(filepath.Join(repoPath, "commands", "from-add.md")); err != nil {
		t.Fatalf("expected added command to exist: %v", err)
	}
}

func TestE2E_ConcurrentDropAndAddSerializeWithConsistentFinalRepo(t *testing.T) {
	repoPath := t.TempDir()
	newSource := createSingleSourceTree(t, t.TempDir(), "new-drop-cmd")

	stdout, stderr, err := runAimgrWithEnv(t, "", map[string]string{"AIMGR_REPO_PATH": repoPath}, "repo", "init")
	requireCLIProcessSuccess(t, stdout, stderr, err)

	signalDir := t.TempDir()
	addProc := startAimgrAsync(t, repoPath, map[string]string{
		"AIMGR_TEST_REPO_HOLD_OP":    "add",
		"AIMGR_TEST_REPO_SIGNAL_DIR": signalDir,
	}, "repo", "add", "local:"+newSource)

	waitForMarker(t, filepath.Join(signalDir, "add.ready"), 5*time.Second)
	dropProc := startAimgrAsync(t, repoPath, nil, "repo", "drop")
	dropProc.assertStillRunning(t)

	releaseMarker(t, filepath.Join(signalDir, "add.release"))

	stdout1, stderr1, err1 := addProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout1, stderr1, err1)
	stdout2, stderr2, err2 := dropProc.wait(t, 20*time.Second)
	requireCLIProcessSuccess(t, stdout2, stderr2, err2)

	if _, err := os.Stat(filepath.Join(repoPath, "commands", "new-drop-cmd.md")); !os.IsNotExist(err) {
		t.Fatalf("expected command added before drop to be removed, stat err=%v", err)
	}

	manifest, err := repomanifest.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load manifest after drop+add: %v", err)
	}
	if len(manifest.Sources) != 1 {
		t.Fatalf("expected one preserved source after drop+add, got %d", len(manifest.Sources))
	}
	requireManifestHasSourcePath(t, repoPath, newSource)

	meta, err := sourcemetadata.Load(repoPath)
	if err != nil {
		t.Fatalf("failed to load source metadata after drop+add: %v", err)
	}
	if len(meta.Sources) != 0 {
		t.Fatalf("expected cleared source metadata after drop+add, got %d entries", len(meta.Sources))
	}
}
