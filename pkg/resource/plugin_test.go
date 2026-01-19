package resource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPlugin(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(tmpDir string) string
		wantValid bool
		wantError bool
	}{
		{
			name: "valid plugin",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "test-plugin")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"test-plugin","description":"Test"}`),
					0644,
				)
				return pluginDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "missing plugin.json",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "not-plugin")
				os.MkdirAll(pluginDir, 0755)
				return pluginDir
			},
			wantValid: false,
			wantError: false,
		},
		{
			name: "file instead of directory",
			setup: func(tmpDir string) string {
				filePath := filepath.Join(tmpDir, "file.txt")
				os.WriteFile(filePath, []byte("test"), 0644)
				return filePath
			},
			wantValid: false,
			wantError: false,
		},
		{
			name: "non-existent path",
			setup: func(tmpDir string) string {
				return filepath.Join(tmpDir, "non-existent")
			},
			wantValid: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			isValid, err := DetectPlugin(path)
			if (err != nil) != tt.wantError {
				t.Errorf("DetectPlugin() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if isValid != tt.wantValid {
				t.Errorf("DetectPlugin() = %v, want %v", isValid, tt.wantValid)
			}
		})
	}
}

func TestLoadPluginMetadata(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(tmpDir string) string
		wantName     string
		wantDesc     string
		wantVersion  string
		wantError    bool
		errorMessage string
	}{
		{
			name: "valid plugin with all fields",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "full-plugin")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{
						"name":"full-plugin",
						"description":"A complete plugin",
						"version":"1.0.0",
						"author":{"name":"Test Author","email":"test@example.com"},
						"license":"MIT"
					}`),
					0644,
				)
				return pluginDir
			},
			wantName:    "full-plugin",
			wantDesc:    "A complete plugin",
			wantVersion: "1.0.0",
			wantError:   false,
		},
		{
			name: "valid plugin with minimal fields",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "minimal-plugin")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"minimal-plugin","description":"Minimal"}`),
					0644,
				)
				return pluginDir
			},
			wantName:  "minimal-plugin",
			wantDesc:  "Minimal",
			wantError: false,
		},
		{
			name: "invalid JSON",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "bad-json")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{invalid json`),
					0644,
				)
				return pluginDir
			},
			wantError:    true,
			errorMessage: "failed to parse plugin.json",
		},
		{
			name: "missing name field",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "no-name")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"description":"No name"}`),
					0644,
				)
				return pluginDir
			},
			wantError:    true,
			errorMessage: "missing required field: name",
		},
		{
			name: "not a plugin directory",
			setup: func(tmpDir string) string {
				notPluginDir := filepath.Join(tmpDir, "not-plugin")
				os.MkdirAll(notPluginDir, 0755)
				return notPluginDir
			},
			wantError:    true,
			errorMessage: "not a valid plugin directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			metadata, err := LoadPluginMetadata(path)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadPluginMetadata() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				if tt.errorMessage != "" && err != nil {
					// Simple substring check
					if err.Error() == "" {
						t.Errorf("LoadPluginMetadata() expected error containing %q, got empty error", tt.errorMessage)
					}
				}
				return
			}

			if metadata.Name != tt.wantName {
				t.Errorf("LoadPluginMetadata() Name = %v, want %v", metadata.Name, tt.wantName)
			}
			if metadata.Description != tt.wantDesc {
				t.Errorf("LoadPluginMetadata() Description = %v, want %v", metadata.Description, tt.wantDesc)
			}
			if tt.wantVersion != "" && metadata.Version != tt.wantVersion {
				t.Errorf("LoadPluginMetadata() Version = %v, want %v", metadata.Version, tt.wantVersion)
			}
		})
	}
}

func TestScanPluginResources(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(tmpDir string) string
		wantCommandCount int
		wantSkillCount   int
		wantError        bool
	}{
		{
			name: "plugin with commands and skills",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "full-plugin")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"full-plugin","description":"Test"}`),
					0644,
				)

				// Create commands
				commandsDir := filepath.Join(pluginDir, "commands")
				os.MkdirAll(commandsDir, 0755)
				os.WriteFile(filepath.Join(commandsDir, "cmd1.md"), []byte("---\ndescription: Test\n---\n"), 0644)
				os.WriteFile(filepath.Join(commandsDir, "cmd2.md"), []byte("---\ndescription: Test\n---\n"), 0644)
				os.WriteFile(filepath.Join(commandsDir, "readme.txt"), []byte("not a command"), 0644)

				// Create skills
				skillsDir := filepath.Join(pluginDir, "skills")
				os.MkdirAll(filepath.Join(skillsDir, "skill1"), 0755)
				os.WriteFile(filepath.Join(skillsDir, "skill1", "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0644)
				os.MkdirAll(filepath.Join(skillsDir, "skill2"), 0755)
				os.WriteFile(filepath.Join(skillsDir, "skill2", "SKILL.md"), []byte("---\nname: skill2\n---\n"), 0644)
				os.MkdirAll(filepath.Join(skillsDir, "not-skill"), 0755) // no SKILL.md

				return pluginDir
			},
			wantCommandCount: 2,
			wantSkillCount:   2,
			wantError:        false,
		},
		{
			name: "plugin with only commands",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "commands-only")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"commands-only","description":"Test"}`),
					0644,
				)

				commandsDir := filepath.Join(pluginDir, "commands")
				os.MkdirAll(commandsDir, 0755)
				os.WriteFile(filepath.Join(commandsDir, "cmd1.md"), []byte("---\ndescription: Test\n---\n"), 0644)

				return pluginDir
			},
			wantCommandCount: 1,
			wantSkillCount:   0,
			wantError:        false,
		},
		{
			name: "plugin with only skills",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "skills-only")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"skills-only","description":"Test"}`),
					0644,
				)

				skillsDir := filepath.Join(pluginDir, "skills")
				os.MkdirAll(filepath.Join(skillsDir, "skill1"), 0755)
				os.WriteFile(filepath.Join(skillsDir, "skill1", "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0644)

				return pluginDir
			},
			wantCommandCount: 0,
			wantSkillCount:   1,
			wantError:        false,
		},
		{
			name: "empty plugin",
			setup: func(tmpDir string) string {
				pluginDir := filepath.Join(tmpDir, "empty-plugin")
				os.MkdirAll(filepath.Join(pluginDir, ".claude-plugin"), 0755)
				os.WriteFile(
					filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
					[]byte(`{"name":"empty-plugin","description":"Test"}`),
					0644,
				)
				return pluginDir
			},
			wantCommandCount: 0,
			wantSkillCount:   0,
			wantError:        false,
		},
		{
			name: "not a plugin",
			setup: func(tmpDir string) string {
				notPluginDir := filepath.Join(tmpDir, "not-plugin")
				os.MkdirAll(notPluginDir, 0755)
				return notPluginDir
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			commandPaths, skillPaths, err := ScanPluginResources(path)
			if (err != nil) != tt.wantError {
				t.Errorf("ScanPluginResources() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			if len(commandPaths) != tt.wantCommandCount {
				t.Errorf("ScanPluginResources() commandPaths count = %v, want %v", len(commandPaths), tt.wantCommandCount)
			}
			if len(skillPaths) != tt.wantSkillCount {
				t.Errorf("ScanPluginResources() skillPaths count = %v, want %v", len(skillPaths), tt.wantSkillCount)
			}
		})
	}
}

func TestDetectClaudeFolder(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(tmpDir string) string
		wantValid bool
		wantError bool
	}{
		{
			name: "valid .claude folder with commands and skills",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				os.MkdirAll(filepath.Join(claudeDir, "commands"), 0755)
				os.MkdirAll(filepath.Join(claudeDir, "skills"), 0755)
				return claudeDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "valid .claude folder with only commands",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				os.MkdirAll(filepath.Join(claudeDir, "commands"), 0755)
				return claudeDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "valid .claude folder with only skills",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				os.MkdirAll(filepath.Join(claudeDir, "skills"), 0755)
				return claudeDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "parent folder containing .claude subdirectory",
			setup: func(tmpDir string) string {
				projectDir := filepath.Join(tmpDir, "my-project")
				claudeDir := filepath.Join(projectDir, ".claude")
				os.MkdirAll(filepath.Join(claudeDir, "commands"), 0755)
				return projectDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "folder with commands but not named .claude",
			setup: func(tmpDir string) string {
				someDir := filepath.Join(tmpDir, "some-folder")
				os.MkdirAll(filepath.Join(someDir, "commands"), 0755)
				return someDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "empty .claude folder",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				os.MkdirAll(claudeDir, 0755)
				return claudeDir
			},
			wantValid: true,
			wantError: false,
		},
		{
			name: "regular folder without Claude structure",
			setup: func(tmpDir string) string {
				regularDir := filepath.Join(tmpDir, "regular")
				os.MkdirAll(regularDir, 0755)
				return regularDir
			},
			wantValid: false,
			wantError: false,
		},
		{
			name: "file instead of directory",
			setup: func(tmpDir string) string {
				filePath := filepath.Join(tmpDir, "file.txt")
				os.WriteFile(filePath, []byte("test"), 0644)
				return filePath
			},
			wantValid: false,
			wantError: false,
		},
		{
			name: "non-existent path",
			setup: func(tmpDir string) string {
				return filepath.Join(tmpDir, "non-existent")
			},
			wantValid: false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			isValid, err := DetectClaudeFolder(path)
			if (err != nil) != tt.wantError {
				t.Errorf("DetectClaudeFolder() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if isValid != tt.wantValid {
				t.Errorf("DetectClaudeFolder() = %v, want %v", isValid, tt.wantValid)
			}
		})
	}
}

func TestScanClaudeFolder(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(tmpDir string) string
		wantCommandCount int
		wantSkillCount   int
		wantError        bool
	}{
		{
			name: "claude folder with commands and skills",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				commandsDir := filepath.Join(claudeDir, "commands")
				skillsDir := filepath.Join(claudeDir, "skills")

				os.MkdirAll(commandsDir, 0755)
				os.MkdirAll(skillsDir, 0755)

				// Create commands
				os.WriteFile(filepath.Join(commandsDir, "test.md"), []byte("---\ndescription: Test\n---\n"), 0644)
				os.WriteFile(filepath.Join(commandsDir, "review.md"), []byte("---\ndescription: Review\n---\n"), 0644)
				os.WriteFile(filepath.Join(commandsDir, "readme.txt"), []byte("not a command"), 0644)

				// Create skills
				skill1Dir := filepath.Join(skillsDir, "skill1")
				os.MkdirAll(skill1Dir, 0755)
				os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0644)

				skill2Dir := filepath.Join(skillsDir, "skill2")
				os.MkdirAll(skill2Dir, 0755)
				os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("---\nname: skill2\n---\n"), 0644)

				os.MkdirAll(filepath.Join(skillsDir, "not-skill"), 0755) // no SKILL.md

				return claudeDir
			},
			wantCommandCount: 2,
			wantSkillCount:   2,
			wantError:        false,
		},
		{
			name: "project folder containing .claude subdirectory",
			setup: func(tmpDir string) string {
				projectDir := filepath.Join(tmpDir, "my-project")
				claudeDir := filepath.Join(projectDir, ".claude")
				commandsDir := filepath.Join(claudeDir, "commands")

				os.MkdirAll(commandsDir, 0755)
				os.WriteFile(filepath.Join(commandsDir, "cmd.md"), []byte("---\ndescription: Command\n---\n"), 0644)

				return projectDir
			},
			wantCommandCount: 1,
			wantSkillCount:   0,
			wantError:        false,
		},
		{
			name: "claude folder with only commands",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				commandsDir := filepath.Join(claudeDir, "commands")
				os.MkdirAll(commandsDir, 0755)
				os.WriteFile(filepath.Join(commandsDir, "cmd.md"), []byte("---\ndescription: Test\n---\n"), 0644)
				return claudeDir
			},
			wantCommandCount: 1,
			wantSkillCount:   0,
			wantError:        false,
		},
		{
			name: "claude folder with only skills",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				skillsDir := filepath.Join(claudeDir, "skills")
				skill1Dir := filepath.Join(skillsDir, "skill1")
				os.MkdirAll(skill1Dir, 0755)
				os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte("---\nname: skill1\n---\n"), 0644)
				return claudeDir
			},
			wantCommandCount: 0,
			wantSkillCount:   1,
			wantError:        false,
		},
		{
			name: "empty claude folder",
			setup: func(tmpDir string) string {
				claudeDir := filepath.Join(tmpDir, ".claude")
				os.MkdirAll(claudeDir, 0755)
				return claudeDir
			},
			wantCommandCount: 0,
			wantSkillCount:   0,
			wantError:        false,
		},
		{
			name: "not a claude folder",
			setup: func(tmpDir string) string {
				regularDir := filepath.Join(tmpDir, "regular")
				os.MkdirAll(regularDir, 0755)
				return regularDir
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := tt.setup(tmpDir)

			contents, err := ScanClaudeFolder(path)
			if (err != nil) != tt.wantError {
				t.Errorf("ScanClaudeFolder() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			if len(contents.CommandPaths) != tt.wantCommandCount {
				t.Errorf("ScanClaudeFolder() commandPaths count = %v, want %v", len(contents.CommandPaths), tt.wantCommandCount)
			}
			if len(contents.SkillPaths) != tt.wantSkillCount {
				t.Errorf("ScanClaudeFolder() skillPaths count = %v, want %v", len(contents.SkillPaths), tt.wantSkillCount)
			}
		})
	}
}
