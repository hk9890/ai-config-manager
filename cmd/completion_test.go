package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectError  bool
		expectOutput string
	}{
		{
			name:         "bash completion",
			args:         []string{"bash"},
			expectError:  false,
			expectOutput: "bash",
		},
		{
			name:         "zsh completion",
			args:         []string{"zsh"},
			expectError:  false,
			expectOutput: "zsh",
		},
		{
			name:         "fish completion",
			args:         []string{"fish"},
			expectError:  false,
			expectOutput: "fish",
		},
		{
			name:         "powershell completion",
			args:         []string{"powershell"},
			expectError:  false,
			expectOutput: "powershell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run completion command
			completionCmd.Run(completionCmd, tt.args)

			// Restore stdout
			_ = w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Error("Expected completion script output, got empty string")
			}

			// For bash, verify it contains completion keywords
			if tt.args[0] == "bash" {
				if !strings.Contains(output, "completion") && !strings.Contains(output, "compgen") {
					t.Error("Expected bash completion keywords in output")
				}
			}
		})
	}
}

func TestCompletionCommandValidArgs(t *testing.T) {
	// Verify valid args are set correctly
	validArgs := []string{"bash", "zsh", "fish", "powershell"}

	if len(completionCmd.ValidArgs) != len(validArgs) {
		t.Errorf("Expected %d valid args, got %d", len(validArgs), len(completionCmd.ValidArgs))
	}

	for _, arg := range validArgs {
		found := false
		for _, validArg := range completionCmd.ValidArgs {
			if validArg == arg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be in ValidArgs", arg)
		}
	}
}

func TestCompletionCommandUsage(t *testing.T) {
	// Verify command has expected metadata
	if completionCmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("Unexpected Use string: %s", completionCmd.Use)
	}

	if completionCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if completionCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify long description contains instructions for all shells
	shells := []string{"Bash", "Zsh", "Fish", "PowerShell"}
	for _, shell := range shells {
		if !strings.Contains(completionCmd.Long, shell) {
			t.Errorf("Long description should contain instructions for %s", shell)
		}
	}
}

func TestCompletionCommandDisablesFlagsInUseLine(t *testing.T) {
	if !completionCmd.DisableFlagsInUseLine {
		t.Error("Expected DisableFlagsInUseLine to be true")
	}
}
