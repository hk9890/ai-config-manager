package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteFormatFlag(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}
	toComplete := ""

	completions, directive := completeFormatFlag(cmd, args, toComplete)

	// Verify expected format options are present
	expectedFormats := []string{"table", "json", "yaml"}
	if len(completions) != len(expectedFormats) {
		t.Errorf("Expected %d completions, got %d", len(expectedFormats), len(completions))
	}

	for _, format := range expectedFormats {
		found := false
		for _, completion := range completions {
			if completion == format {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s in completions", format)
		}
	}

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestCompleteResourceTypes(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}
	toComplete := ""

	completions, directive := completeResourceTypes(cmd, args, toComplete)

	// Verify expected resource types are present
	expectedTypes := []string{"command", "skill", "agent", "package"}
	if len(completions) < len(expectedTypes) {
		t.Errorf("Expected at least %d completions, got %d", len(expectedTypes), len(completions))
	}

	for _, resType := range expectedTypes {
		found := false
		for _, completion := range completions {
			if completion == resType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected resource type %s in completions", resType)
		}
	}

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestCompleteToolNames(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}
	toComplete := ""

	completions, directive := completeToolNames(cmd, args, toComplete)

	// Verify expected tool names are present
	expectedTools := []string{"claude", "opencode", "copilot", "windsurf"}

	for _, tool := range expectedTools {
		found := false
		for _, completion := range completions {
			if completion == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s in completions", tool)
		}
	}

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

// TestCompletionHelpersNoErrors verifies that completion helpers don't panic
func TestCompletionHelpersNoErrors(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}
	toComplete := ""

	// Test all completion helpers to ensure they don't panic
	helpers := []func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective){
		completeFormatFlag,
		completeResourceTypes,
		completeToolNames,
	}

	for _, helper := range helpers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Completion helper panicked: %v", r)
				}
			}()
			helper(cmd, args, toComplete)
		}()
	}
}

func TestCompleteConfigKeys(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}
	toComplete := ""

	completions, directive := completeConfigKeys(cmd, args, toComplete)

	// Verify expected config keys are present
	expectedKeys := []string{"repo.path"}
	found := false
	for _, key := range expectedKeys {
		for _, completion := range completions {
			if completion == key {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Expected at least repo.path in completions")
	}

	// Verify directive
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}
