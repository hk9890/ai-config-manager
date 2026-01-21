package cmd

import (
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/tools"
)

func TestParseTargetFlag(t *testing.T) {
	tests := []struct {
		name        string
		targetFlag  string
		want        []tools.Tool
		wantErr     bool
		errContains string
	}{
		{
			name:       "empty flag returns nil",
			targetFlag: "",
			want:       nil,
			wantErr:    false,
		},
		{
			name:       "single valid target - claude",
			targetFlag: "claude",
			want:       []tools.Tool{tools.Claude},
			wantErr:    false,
		},
		{
			name:       "single valid target - opencode",
			targetFlag: "opencode",
			want:       []tools.Tool{tools.OpenCode},
			wantErr:    false,
		},
		{
			name:       "single valid target - copilot",
			targetFlag: "copilot",
			want:       []tools.Tool{tools.Copilot},
			wantErr:    false,
		},
		{
			name:       "multiple valid targets",
			targetFlag: "claude,opencode",
			want:       []tools.Tool{tools.Claude, tools.OpenCode},
			wantErr:    false,
		},
		{
			name:       "all three targets",
			targetFlag: "claude,opencode,copilot",
			want:       []tools.Tool{tools.Claude, tools.OpenCode, tools.Copilot},
			wantErr:    false,
		},
		{
			name:       "targets with spaces",
			targetFlag: "claude, opencode, copilot",
			want:       []tools.Tool{tools.Claude, tools.OpenCode, tools.Copilot},
			wantErr:    false,
		},
		{
			name:        "invalid target",
			targetFlag:  "invalid",
			want:        nil,
			wantErr:     true,
			errContains: "invalid target 'invalid'",
		},
		{
			name:        "mixed valid and invalid",
			targetFlag:  "claude,invalid,opencode",
			want:        nil,
			wantErr:     true,
			errContains: "invalid target 'invalid'",
		},
		{
			name:       "case insensitive",
			targetFlag: "CLAUDE,OpenCode,CoPilot",
			want:       []tools.Tool{tools.Claude, tools.OpenCode, tools.Copilot},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTargetFlag(tt.targetFlag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTargetFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("parseTargetFlag() error = %v, want error containing %q", err, tt.errContains)
					}
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseTargetFlag() got %d tools, want %d", len(got), len(tt.want))
				return
			}

			for i, tool := range tt.want {
				if got[i] != tool {
					t.Errorf("parseTargetFlag()[%d] = %v, want %v", i, got[i], tool)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
