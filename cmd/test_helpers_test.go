package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	stdoutCh := make(chan string, 1)
	stderrCh := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		stdoutCh <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderrR)
		stderrCh <- buf.String()
	}()

	fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout := <-stdoutCh
	stderr := <-stderrCh

	_ = stdoutR.Close()
	_ = stderrR.Close()

	return stdout, stderr
}
