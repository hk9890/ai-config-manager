package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/hk9890/ai-config-manager/pkg/resource"
)

func TestRemovePackage(t *testing.T) {
	tests := []struct {
		name               string
		packageName        string
		packageDesc        string
		packageResources   []string
		setupResources     []string // resources to actually create in repo
		withResources      bool
		force              bool
		confirmInput       string
		wantError          bool
		wantPackageRemoved bool
		wantResourcesGone  []string // resources that should be removed
		wantResourcesKept  []string // resources that should remain
	}{
		{
			name:        "remove package only with force",
			packageName: "test-package",
			packageDesc: "Test package",
			packageResources: []string{
				"command/test-cmd",
				"skill/test-skill",
			},
			setupResources: []string{
				"command/test-cmd",
				"skill/test-skill",
			},
			withResources:      false,
			force:              true,
			wantError:          false,
			wantPackageRemoved: true,
			wantResourcesGone:  []string{},
			wantResourcesKept: []string{
				"command/test-cmd",
				"skill/test-skill",
			},
		},
		{
			name:        "remove package only with confirmation yes",
			packageName: "test-package",
			packageDesc: "Test package",
			packageResources: []string{
				"command/test-cmd",
			},
			setupResources: []string{
				"command/test-cmd",
			},
			withResources:      false,
			force:              false,
			confirmInput:       "y\n",
			wantError:          false,
			wantPackageRemoved: true,
			wantResourcesGone:  []string{},
			wantResourcesKept: []string{
				"command/test-cmd",
			},
		},
		{
			name:        "remove package only with confirmation no",
			packageName: "test-package",
			packageDesc: "Test package",
			packageResources: []string{
				"command/test-cmd",
			},
			setupResources: []string{
				"command/test-cmd",
			},
			withResources:      false,
			force:              false,
			confirmInput:       "n\n",
			wantError:          false,
			wantPackageRemoved: false,
			wantResourcesGone:  []string{},
			wantResourcesKept: []string{
				"command/test-cmd",
			},
		},
		{
			name:        "remove package with resources and force",
			packageName: "full-package",
			packageDesc: "Full package with resources",
			packageResources: []string{
				"command/cmd-one",
				"skill/skill-one",
				"agent/agent-one",
			},
			setupResources: []string{
				"command/cmd-one",
				"skill/skill-one",
				"agent/agent-one",
			},
			withResources:      true,
			force:              true,
			wantError:          false,
			wantPackageRemoved: true,
			wantResourcesGone: []string{
				"command/cmd-one",
				"skill/skill-one",
				"agent/agent-one",
			},
			wantResourcesKept: []string{},
		},
		{
			name:        "remove package with resources and confirmation yes",
			packageName: "full-package",
			packageDesc: "Full package with resources",
			packageResources: []string{
				"command/cmd-one",
				"skill/skill-one",
			},
			setupResources: []string{
				"command/cmd-one",
				"skill/skill-one",
			},
			withResources:      true,
			force:              false,
			confirmInput:       "yes\n",
			wantError:          false,
			wantPackageRemoved: true,
			wantResourcesGone: []string{
				"command/cmd-one",
				"skill/skill-one",
			},
			wantResourcesKept: []string{},
		},
		{
			name:        "remove package with resources and confirmation no",
			packageName: "full-package",
			packageDesc: "Full package with resources",
			packageResources: []string{
				"command/cmd-one",
			},
			setupResources: []string{
				"command/cmd-one",
			},
			withResources:      true,
			force:              false,
			confirmInput:       "n\n",
			wantError:          false,
			wantPackageRemoved: false,
			wantResourcesGone:  []string{},
			wantResourcesKept: []string{
				"command/cmd-one",
			},
		},
		{
			name:        "remove package with some missing resources",
			packageName: "partial-package",
			packageDesc: "Package with some missing resources",
			packageResources: []string{
				"command/exists",
				"command/missing",
				"skill/also-missing",
			},
			setupResources: []string{
				"command/exists",
			},
			withResources:      true,
			force:              true,
			wantError:          false,
			wantPackageRemoved: true,
			wantResourcesGone: []string{
				"command/exists",
			},
			wantResourcesKept: []string{},
		},
		{
			name:               "remove non-existent package",
			packageName:        "non-existent",
			packageDesc:        "Does not exist",
			packageResources:   []string{},
			setupResources:     []string{},
			withResources:      false,
			force:              true,
			wantError:          true,
			wantPackageRemoved: false,
			wantResourcesGone:  []string{},
			wantResourcesKept:  []string{},
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

			// Create packages directory
			packagesDir := filepath.Join(tmpDir, "packages")
			if err := os.MkdirAll(packagesDir, 0755); err != nil {
				t.Fatalf("failed to create packages dir: %v", err)
			}

			// Setup manager with temp repo
			mgr := repo.NewManagerWithPath(tmpDir)

			// Set environment variable for repo path
			oldRepoPath := os.Getenv("AIMGR_REPO_PATH")
			os.Setenv("AIMGR_REPO_PATH", tmpDir)
			defer os.Setenv("AIMGR_REPO_PATH", oldRepoPath)

			// Create package if not testing non-existent
			if tt.packageName != "non-existent" {
				pkg := &resource.Package{
					Name:        tt.packageName,
					Description: tt.packageDesc,
					Resources:   tt.packageResources,
				}
				if err := resource.SavePackage(pkg, tmpDir); err != nil {
					t.Fatalf("failed to save package: %v", err)
				}

				// Create metadata file
				metadataDir := filepath.Join(tmpDir, ".metadata", "packages")
				if err := os.MkdirAll(metadataDir, 0755); err != nil {
					t.Fatalf("failed to create metadata dir: %v", err)
				}
				metadata := map[string]interface{}{
					"name":           tt.packageName,
					"source_type":    "local",
					"resource_count": len(tt.packageResources),
				}
				metadataJSON, _ := json.Marshal(metadata)
				metadataPath := filepath.Join(metadataDir, tt.packageName+"-metadata.json")
				if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
					t.Fatalf("failed to write metadata: %v", err)
				}
			}

			// Setup resources
			for _, resRef := range tt.setupResources {
				resType, resName, err := resource.ParseResourceReference(resRef)
				if err != nil {
					t.Fatalf("failed to parse resource ref %s: %v", resRef, err)
				}

				var filePath string
				var content []byte

				switch resType {
				case resource.Skill:
					skillDir := filepath.Join(tmpDir, "skills", resName)
					if err := os.MkdirAll(skillDir, 0755); err != nil {
						t.Fatalf("failed to create skill dir: %v", err)
					}
					filePath = filepath.Join(skillDir, "SKILL.md")
					content = []byte("---\ndescription: Test skill\n---\n# Test Skill")
				case resource.Command:
					filePath = filepath.Join(tmpDir, "commands", resName+".md")
					content = []byte("---\ndescription: Test command\n---\n# Test Command")
				case resource.Agent:
					filePath = filepath.Join(tmpDir, "agents", resName+".md")
					content = []byte("---\ndescription: Test agent\n---\n# Test Agent")
				}

				if err := os.WriteFile(filePath, content, 0644); err != nil {
					t.Fatalf("failed to write resource: %v", err)
				}
			}

			// Setup command flags
			removeForceFlag = tt.force
			removeWithResourcesFlag = tt.withResources

			// Setup stdin for confirmation
			if tt.confirmInput != "" {
				oldStdin := os.Stdin
				r, w, _ := os.Pipe()
				os.Stdin = r
				w.Write([]byte(tt.confirmInput))
				w.Close()
				defer func() { os.Stdin = oldStdin }()
			}

			// Build args for full command path
			args := []string{"repo", "remove", "package/" + tt.packageName}
			if tt.force {
				args = append(args, "--force")
			}
			if tt.withResources {
				args = append(args, "--with-resources")
			}

			// Execute through root command
			rootCmd.SetArgs(args)

			// Capture output before execution
			var outBuf, errBuf bytes.Buffer
			rootCmd.SetOut(&outBuf)
			rootCmd.SetErr(&errBuf)

			err := rootCmd.Execute()

			// Check error expectation
			if tt.wantError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v\nOutput: %s\nError: %s", err, outBuf.String(), errBuf.String())
			}

			// Check package removal
			pkgPath := filepath.Join(tmpDir, "packages", tt.packageName+".package.json")
			_, pkgErr := os.Stat(pkgPath)
			packageExists := pkgErr == nil

			if tt.wantPackageRemoved && packageExists {
				t.Errorf("expected package to be removed, but it still exists")
			}
			if !tt.wantPackageRemoved && !packageExists && tt.packageName != "non-existent" {
				t.Errorf("expected package to remain, but it was removed")
			}

			// Check metadata removal
			if tt.wantPackageRemoved {
				metadataPath := filepath.Join(tmpDir, ".metadata", "packages", tt.packageName+"-metadata.json")
				if _, err := os.Stat(metadataPath); err == nil {
					t.Errorf("expected metadata to be removed, but it still exists")
				}
			}

			// Check removed resources
			for _, resRef := range tt.wantResourcesGone {
				resType, resName, err := resource.ParseResourceReference(resRef)
				if err != nil {
					t.Fatalf("failed to parse resource ref %s: %v", resRef, err)
				}
				_, err = mgr.Get(resName, resType)
				if err == nil {
					t.Errorf("expected resource %s to be removed, but it still exists", resRef)
				}
			}

			// Check kept resources
			for _, resRef := range tt.wantResourcesKept {
				resType, resName, err := resource.ParseResourceReference(resRef)
				if err != nil {
					t.Fatalf("failed to parse resource ref %s: %v", resRef, err)
				}
				_, err = mgr.Get(resName, resType)
				if err != nil {
					t.Errorf("expected resource %s to remain, but it was removed: %v", resRef, err)
				}
			}

			// Reset flags
			removeForceFlag = false
			removeWithResourcesFlag = false
		})
	}
}

// TestRemovePackageOutput is covered by the main TestRemovePackage tests
// The output validation is implicit in the successful removal checks
