package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestRemoveCmd(t *testing.T) {
	tests := []struct {
		name           string
		setupResources []struct {
			resType resource.ResourceType
			name    string
			content string
		}
		args          []string
		force         bool
		confirmInput  string // stdin input for confirmation
		wantError     bool
		wantRemoved   []string // resources that should be removed
		wantRemaining []string // resources that should remain
	}{
		{
			name: "remove single skill",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "test-skill", "---\ndescription: Test skill\n---\n# Test Skill"},
			},
			args:          []string{"skill/test-skill"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"skill/test-skill"},
			wantRemaining: []string{},
		},
		{
			name: "remove single command",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Command, "test-command", "---\ndescription: Test command\n---\n# Test Command"},
			},
			args:          []string{"command/test-command"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"command/test-command"},
			wantRemaining: []string{},
		},
		{
			name: "remove single agent",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Agent, "test-agent", "---\ndescription: Test agent\n---\n# Test Agent"},
			},
			args:          []string{"agent/test-agent"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"agent/test-agent"},
			wantRemaining: []string{},
		},
		{
			name: "remove multiple resources in one command",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "skill-one", "---\ndescription: Skill one\n---\n# Skill One"},
				{resource.Command, "command-one", "---\ndescription: Command one\n---\n# Command One"},
				{resource.Agent, "agent-one", "---\ndescription: Agent one\n---\n# Agent One"},
			},
			args:          []string{"skill/skill-one", "command/command-one", "agent/agent-one"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"skill/skill-one", "command/command-one", "agent/agent-one"},
			wantRemaining: []string{},
		},
		{
			name: "remove with wildcard pattern",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Command, "test-one", "---\ndescription: Test one\n---\n# Test One"},
				{resource.Command, "test-two", "---\ndescription: Test two\n---\n# Test Two"},
				{resource.Command, "keep-this", "---\ndescription: Keep this\n---\n# Keep This"},
			},
			args:          []string{"command/test-*"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"command/test-one", "command/test-two"},
			wantRemaining: []string{"command/keep-this"},
		},
		{
			name: "remove all types matching pattern",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "temp-skill", "---\ndescription: Temp skill\n---\n# Temp Skill"},
				{resource.Command, "temp-command", "---\ndescription: Temp command\n---\n# Temp Command"},
				{resource.Agent, "temp-agent", "---\ndescription: Temp agent\n---\n# Temp Agent"},
				{resource.Skill, "keep-skill", "---\ndescription: Keep skill\n---\n# Keep Skill"},
			},
			args:          []string{"*/temp-*"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"skill/temp-skill", "command/temp-command", "agent/temp-agent"},
			wantRemaining: []string{"skill/keep-skill"},
		},
		{
			name: "remove with confirmation yes",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "test-skill", "---\ndescription: Test skill\n---\n# Test Skill"},
			},
			args:          []string{"skill/test-skill"},
			force:         false,
			confirmInput:  "y\n",
			wantError:     false,
			wantRemoved:   []string{"skill/test-skill"},
			wantRemaining: []string{},
		},
		{
			name: "remove with confirmation no",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "test-skill", "---\ndescription: Test skill\n---\n# Test Skill"},
			},
			args:          []string{"skill/test-skill"},
			force:         false,
			confirmInput:  "n\n",
			wantError:     false,
			wantRemoved:   []string{},
			wantRemaining: []string{"skill/test-skill"},
		},
		{
			name:          "no patterns provided",
			args:          []string{},
			force:         true,
			wantError:     true,
			wantRemoved:   []string{},
			wantRemaining: []string{},
		},
		{
			name: "non-existent resource",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{},
			args:          []string{"skill/non-existent"},
			force:         true,
			wantError:     true,
			wantRemoved:   []string{},
			wantRemaining: []string{},
		},
		{
			name: "pattern with no matches",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "test-skill", "---\ndescription: Test skill\n---\n# Test Skill"},
			},
			args:          []string{"skill/no-match-*"},
			force:         true,
			wantError:     true,
			wantRemoved:   []string{},
			wantRemaining: []string{"skill/test-skill"},
		},
		{
			name: "duplicate patterns in args",
			setupResources: []struct {
				resType resource.ResourceType
				name    string
				content string
			}{
				{resource.Skill, "test-skill", "---\ndescription: Test skill\n---\n# Test Skill"},
			},
			args:          []string{"skill/test-skill", "skill/test-skill"},
			force:         true,
			wantError:     false,
			wantRemoved:   []string{"skill/test-skill"},
			wantRemaining: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp repo
			tmpDir := t.TempDir()

			// Create repo structure
			for _, resType := range []resource.ResourceType{resource.Command, resource.Skill, resource.Agent} {
				dir := filepath.Join(tmpDir, string(resType)+"s")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
			}

			// Setup manager with temp repo
			mgr := repo.NewManagerWithPath(tmpDir)

			// Setup resources
			for _, res := range tt.setupResources {
				// Write resource file
				var filePath string
				switch res.resType {
				case resource.Skill:
					skillDir := filepath.Join(tmpDir, "skills", res.name)
					if err := os.MkdirAll(skillDir, 0755); err != nil {
						t.Fatalf("failed to create skill dir: %v", err)
					}
					filePath = filepath.Join(skillDir, "SKILL.md")
				case resource.Command:
					filePath = filepath.Join(tmpDir, "commands", res.name+".md")
				case resource.Agent:
					filePath = filepath.Join(tmpDir, "agents", res.name+".md")
				}

				if err := os.WriteFile(filePath, []byte(res.content), 0644); err != nil {
					t.Fatalf("failed to write resource: %v", err)
				}
			}

			// Setup command
			removeCmd.SetArgs(tt.args)
			removeForceFlag = tt.force

			// Setup stdin for confirmation
			if tt.confirmInput != "" {
				oldStdin := os.Stdin
				r, w, _ := os.Pipe()
				os.Stdin = r
				w.Write([]byte(tt.confirmInput))
				w.Close()
				defer func() { os.Stdin = oldStdin }()
			}

			// Capture output
			var outBuf, errBuf bytes.Buffer
			removeCmd.SetOut(&outBuf)
			removeCmd.SetErr(&errBuf)

			// Execute command
			err := removeCmd.Execute()

			// Check error expectation
			if tt.wantError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check removed resources
			for _, resourceArg := range tt.wantRemoved {
				resType, name, err := ParseResourceArg(resourceArg)
				if err != nil {
					t.Fatalf("failed to parse resource arg %s: %v", resourceArg, err)
				}
				_, err = mgr.Get(name, resType)
				if err == nil {
					t.Errorf("expected resource %s to be removed, but it still exists", resourceArg)
				}
			}

			// Check remaining resources
			for _, resourceArg := range tt.wantRemaining {
				resType, name, err := ParseResourceArg(resourceArg)
				if err != nil {
					t.Fatalf("failed to parse resource arg %s: %v", resourceArg, err)
				}
				_, err = mgr.Get(name, resType)
				if err != nil {
					t.Errorf("expected resource %s to remain, but it was removed: %v", resourceArg, err)
				}
			}

			// Reset flags
			removeForceFlag = false
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no duplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "all duplicates",
			input: []string{"a", "a", "a"},
			want:  []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueStrings(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("uniqueStrings() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("uniqueStrings()[%d] = %s, want %s", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestRemoveCmdOutput(t *testing.T) {
	// Setup temp repo
	tmpDir := t.TempDir()

	// Create repo structure
	for _, resType := range []resource.ResourceType{resource.Command, resource.Skill, resource.Agent} {
		dir := filepath.Join(tmpDir, string(resType)+"s")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	// Create manager
	mgr := repo.NewManagerWithPath(tmpDir)

	// Setup skill
	skillDir := filepath.Join(tmpDir, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("---\ndescription: Test skill\n---\n# Test Skill"), 0644); err != nil {
		t.Fatalf("failed to write skill: %v", err)
	}

	// Execute with force flag
	removeCmd.SetArgs([]string{"skill/test-skill"})
	removeForceFlag = true

	var outBuf bytes.Buffer
	removeCmd.SetOut(&outBuf)
	removeCmd.SetErr(&outBuf)

	err := removeCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "✓ Removed skill/test-skill") {
		t.Errorf("expected success message in output, got: %s", output)
	}

	// Verify resource was removed
	_, err = mgr.Get("test-skill", resource.Skill)
	if err == nil {
		t.Errorf("expected resource to be removed")
	}

	// Reset flags
	removeForceFlag = false
}
