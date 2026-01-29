//go:build integration

package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hk9890/ai-config-manager/pkg/repo"
)

// TestRepoSyncIdempotency tests that running repo sync twice on the same source
// should be idempotent - the second sync should succeed without errors.
//
// Current behavior (BUG):
// - ALL resource types FAIL on second sync with "already exists" error
// - This affects: flat commands, nested commands, skills, agents, and packages
//
// Expected behavior:
// - All resource types should succeed on second sync
// - Second sync should report "Skipped" or "Updated" (not "Added" or "Failed")
func TestRepoSyncIdempotency(t *testing.T) {
	// Create isolated repo and source directories
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Setup fixture ONCE with ALL resource types
	resources := setupAllResources(t, sourceDir)

	// Create repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Test each resource type separately
	tests := []struct {
		name         string
		resourcePath string
		expectFail   bool // Expected to fail on 2nd sync (documents current bug)
	}{
		{
			name:         "flat_command",
			resourcePath: resources["flat_command"],
			expectFail:   true, // BUG: Commands fail on second sync with "already exists"
		},
		{
			name:         "nested_command",
			resourcePath: resources["nested_command"],
			expectFail:   true, // BUG: Nested commands also fail on second sync with "already exists"
		},
		{
			name:         "skill",
			resourcePath: resources["skill"],
			expectFail:   true, // BUG: Skills also fail on second sync with "already exists"
		},
		{
			name:         "agent",
			resourcePath: resources["agent"],
			expectFail:   true, // BUG: Agents also fail on second sync with "already exists"
		},
		{
			name:         "package",
			resourcePath: resources["package"],
			expectFail:   true, // BUG: Packages also fail on second sync with "already exists"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIRST SYNC: Import the resource
			t.Logf("First sync: importing %s from %s", tt.name, tt.resourcePath)
			result1, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        false, // Don't force on first import
				SkipExisting: false,
				DryRun:       false,
			})

			if err != nil {
				t.Fatalf("First sync failed: %v", err)
			}

			// Verify first sync succeeded
			if len(result1.Added) != 1 {
				t.Errorf("First sync: expected 1 added, got %d", len(result1.Added))
			}
			if len(result1.Failed) > 0 {
				t.Errorf("First sync: expected 0 failed, got %d: %v", len(result1.Failed), result1.Failed)
			}
			if len(result1.Skipped) > 0 {
				t.Errorf("First sync: expected 0 skipped, got %d", len(result1.Skipped))
			}

			t.Logf("First sync result: %d added, %d skipped, %d failed",
				len(result1.Added), len(result1.Skipped), len(result1.Failed))

			// SECOND SYNC: Import the same resource again (should be idempotent)
			t.Logf("Second sync: importing same %s again", tt.name)
			result2, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        false, // Don't force - should handle existing resource gracefully
				SkipExisting: false, // Don't skip - default behavior should be idempotent
				DryRun:       false,
			})

			// Check if behavior matches expectations (documents current bug state)
			if tt.expectFail {
				// Commands currently FAIL on second sync (BUG)
				if err != nil || len(result2.Failed) > 0 {
					t.Logf("✗ Second sync FAILED as expected (BUG - should be idempotent)")
					if err != nil {
						t.Logf("  Error: %v", err)
					}
					for _, fail := range result2.Failed {
						t.Logf("  Failed: %s - %s", fail.Path, fail.Message)
					}
				} else {
					t.Errorf("Expected failure (bug not reproduced), but succeeded")
					t.Logf("  This means the bug may be fixed - update expectFail to false")
				}
			} else {
				// Skills/agents/packages succeed but with misleading "Added" message
				if err != nil || len(result2.Failed) > 0 {
					t.Errorf("Second sync failed unexpectedly: %v", err)
					for _, fail := range result2.Failed {
						t.Logf("  Failed: %s - %s", fail.Path, fail.Message)
					}
				} else {
					t.Logf("✓ Second sync succeeded (idempotent)")
					if len(result2.Added) > 0 {
						t.Logf("  ⚠ Got %d 'Added' (MISLEADING - should be 'Skipped' or 'Updated')", len(result2.Added))
					}
				}
			}

			t.Logf("Second sync result: %d added, %d skipped, %d failed",
				len(result2.Added), len(result2.Skipped), len(result2.Failed))
		})
	}
}

// TestRepoSyncIdempotencyWithSkipExisting tests that --skip-existing flag
// correctly skips resources on second sync
func TestRepoSyncIdempotencyWithSkipExisting(t *testing.T) {
	// Create isolated repo and source directories
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Setup fixture ONCE with ALL resource types
	resources := setupAllResources(t, sourceDir)

	// Create repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Test each resource type separately
	tests := []struct {
		name         string
		resourcePath string
		expectFail   bool // Expected to fail (currently all should work with --skip-existing)
	}{
		{
			name:         "flat_command_skip",
			resourcePath: resources["flat_command"],
			expectFail:   false, // --skip-existing should work for all types
		},
		{
			name:         "nested_command_skip",
			resourcePath: resources["nested_command"],
			expectFail:   false, // --skip-existing should work for all types
		},
		{
			name:         "skill_skip",
			resourcePath: resources["skill"],
			expectFail:   false, // --skip-existing should work for all types
		},
		{
			name:         "agent_skip",
			resourcePath: resources["agent"],
			expectFail:   false, // --skip-existing should work for all types
		},
		{
			name:         "package_skip",
			resourcePath: resources["package"],
			expectFail:   false, // --skip-existing should work for all types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First sync: import the resource
			t.Logf("First sync: importing %s", tt.name)
			result1, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        false,
				SkipExisting: false,
				DryRun:       false,
			})

			if err != nil {
				t.Fatalf("First sync failed: %v", err)
			}

			if len(result1.Added) != 1 {
				t.Errorf("First sync: expected 1 added, got %d", len(result1.Added))
			}

			// Second sync with SkipExisting: should skip the resource
			t.Logf("Second sync with --skip-existing: should skip %s", tt.name)
			result2, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        false,
				SkipExisting: true, // Skip existing resources
				DryRun:       false,
			})

			if tt.expectFail {
				// Currently not expected to fail for any type with --skip-existing
				if err != nil || len(result2.Failed) > 0 {
					t.Logf("✗ Second sync FAILED as expected")
				} else {
					t.Errorf("Expected failure, but succeeded")
				}
			} else {
				// Should succeed and skip
				if err != nil {
					t.Errorf("Second sync with --skip-existing failed: %v", err)
				}

				// Should be skipped
				if len(result2.Skipped) != 1 {
					t.Errorf("Second sync: expected 1 skipped, got %d", len(result2.Skipped))
				}
				if len(result2.Added) > 0 {
					t.Errorf("Second sync: expected 0 added, got %d", len(result2.Added))
				}
				if len(result2.Failed) > 0 {
					t.Errorf("Second sync: expected 0 failed, got %d", len(result2.Failed))
					for _, fail := range result2.Failed {
						t.Logf("  Failed: %s - %s", fail.Path, fail.Message)
					}
				} else {
					t.Logf("✓ Second sync with --skip-existing succeeded")
				}
			}

			t.Logf("Second sync result: %d added, %d skipped, %d failed",
				len(result2.Added), len(result2.Skipped), len(result2.Failed))
		})
	}
}

// Helper functions to set up test resources

// setupFlatCommand creates a flat command (commands/test.md) in the specified base directory
func setupFlatCommand(t *testing.T, baseDir string) string {
	commandsDir := filepath.Join(baseDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "flat-test.md")
	cmdContent := `---
description: Flat test command for sync idempotency
---
# flat-test
This is a flat test command for verifying sync idempotency.
`

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create flat command file: %v", err)
	}

	return cmdPath
}

// setupNestedCommand creates a nested command (commands/mydir/test.md) in the specified base directory
func setupNestedCommand(t *testing.T, baseDir string) string {
	commandsDir := filepath.Join(baseDir, "commands", "mydir")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("Failed to create nested commands directory: %v", err)
	}

	cmdPath := filepath.Join(commandsDir, "nested-test.md")
	cmdContent := `---
description: Nested test command for sync idempotency
---
# nested-test
This is a nested test command for verifying sync idempotency.
`

	if err := os.WriteFile(cmdPath, []byte(cmdContent), 0644); err != nil {
		t.Fatalf("Failed to create nested command file: %v", err)
	}

	return cmdPath
}

// setupTestCommand creates a test command in the specified base directory
// Deprecated: Use setupFlatCommand or setupNestedCommand instead
func setupTestCommand(t *testing.T, baseDir string) string {
	return setupFlatCommand(t, baseDir)
}

// setupTestSkill creates a test skill in the specified base directory
func setupTestSkill(t *testing.T, baseDir string) string {
	skillDir := filepath.Join(baseDir, "skills", "test-sync-skill")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	skillContent := `---
name: test-sync-skill
description: Test skill for sync idempotency
license: MIT
---
# test-sync-skill
This is a test skill for verifying sync idempotency.
`

	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("Failed to create SKILL.md: %v", err)
	}

	return skillDir
}

// setupTestAgent creates a test agent in the specified base directory
func setupTestAgent(t *testing.T, baseDir string) string {
	agentsDir := filepath.Join(baseDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "test-sync-agent.md")
	agentContent := `---
description: Test agent for sync idempotency
---
# test-sync-agent
This is a test agent for verifying sync idempotency.
`

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to create agent file: %v", err)
	}

	return agentPath
}

// setupTestPackage creates a test package in the specified base directory
// Package references a command that also needs to be created
func setupTestPackage(t *testing.T, baseDir string) string {
	// Create the command that the package will reference
	commandPath := setupTestCommand(t, baseDir)
	t.Logf("Created command for package: %s", commandPath)

	// Create package directory
	packagesDir := filepath.Join(baseDir, "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		t.Fatalf("Failed to create packages directory: %v", err)
	}

	packagePath := filepath.Join(packagesDir, "test-sync-package.package.json")
	packageContent := `{
  "name": "test-sync-package",
  "description": "Test package for sync idempotency",
  "resources": [
    "command/flat-test"
  ]
}
`

	if err := os.WriteFile(packagePath, []byte(packageContent), 0644); err != nil {
		t.Fatalf("Failed to create package file: %v", err)
	}

	return packagePath
}

// setupAllResources creates a complete test fixture with all resource types
// Returns a map of resource type names to their paths
func setupAllResources(t *testing.T, baseDir string) map[string]string {
	resources := make(map[string]string)

	// 1. Flat command: commands/flat-test.md
	resources["flat_command"] = setupFlatCommand(t, baseDir)
	t.Logf("Created flat command: %s", resources["flat_command"])

	// 2. Nested command: commands/mydir/nested-test.md
	resources["nested_command"] = setupNestedCommand(t, baseDir)
	t.Logf("Created nested command: %s", resources["nested_command"])

	// 3. Skill: skills/test-sync-skill/
	resources["skill"] = setupTestSkill(t, baseDir)
	t.Logf("Created skill: %s", resources["skill"])

	// 4. Agent: agents/test-sync-agent.md
	resources["agent"] = setupTestAgent(t, baseDir)
	t.Logf("Created agent: %s", resources["agent"])

	// 5. Package: packages/test-sync-package.package.json
	resources["package"] = setupTestPackage(t, baseDir)
	t.Logf("Created package: %s", resources["package"])

	return resources
}

// TestRepoSyncIdempotencyWithForce tests that --force flag correctly
// overwrites resources on second sync
func TestRepoSyncIdempotencyWithForce(t *testing.T) {
	// Create isolated repo and source directories
	repoDir := t.TempDir()
	sourceDir := t.TempDir()

	// Setup fixture ONCE with ALL resource types
	resources := setupAllResources(t, sourceDir)

	// Create repo manager
	manager := repo.NewManagerWithPath(repoDir)

	// Test each resource type separately
	tests := []struct {
		name         string
		resourcePath string
		modifyFunc   func(t *testing.T, resourcePath string)
		expectFail   bool // Expected to fail (currently all should work with --force)
	}{
		{
			name:         "flat_command_force",
			resourcePath: resources["flat_command"],
			modifyFunc:   modifyCommand,
			expectFail:   false, // --force should work for all types
		},
		{
			name:         "nested_command_force",
			resourcePath: resources["nested_command"],
			modifyFunc:   modifyCommand,
			expectFail:   false, // --force should work for all types
		},
		{
			name:         "skill_force",
			resourcePath: resources["skill"],
			modifyFunc:   modifySkill,
			expectFail:   false, // --force should work for all types
		},
		{
			name:         "agent_force",
			resourcePath: resources["agent"],
			modifyFunc:   modifyAgent,
			expectFail:   false, // --force should work for all types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First sync: import the resource
			t.Logf("First sync: importing %s", tt.name)
			result1, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        false,
				SkipExisting: false,
				DryRun:       false,
			})

			if err != nil {
				t.Fatalf("First sync failed: %v", err)
			}

			if len(result1.Added) != 1 {
				t.Errorf("First sync: expected 1 added, got %d", len(result1.Added))
			}

			// Modify the source resource
			t.Logf("Modifying source %s", tt.name)
			tt.modifyFunc(t, tt.resourcePath)

			// Second sync with Force: should overwrite the resource
			t.Logf("Second sync with --force: should overwrite %s", tt.name)
			result2, err := manager.AddBulk([]string{tt.resourcePath}, repo.BulkImportOptions{
				Force:        true, // Force overwrite
				SkipExisting: false,
				DryRun:       false,
			})

			if tt.expectFail {
				// Currently not expected to fail for any type with --force
				if err != nil || len(result2.Failed) > 0 {
					t.Logf("✗ Second sync FAILED as expected")
				} else {
					t.Errorf("Expected failure, but succeeded")
				}
			} else {
				// Should succeed and overwrite
				if err != nil {
					t.Errorf("Second sync with --force failed: %v", err)
				}

				// Should be added (overwritten)
				if len(result2.Added) != 1 {
					t.Errorf("Second sync: expected 1 added (overwritten), got %d", len(result2.Added))
				}
				if len(result2.Failed) > 0 {
					t.Errorf("Second sync: expected 0 failed, got %d", len(result2.Failed))
					for _, fail := range result2.Failed {
						t.Logf("  Failed: %s - %s", fail.Path, fail.Message)
					}
				} else {
					t.Logf("✓ Second sync with --force succeeded (overwritten)")
				}
			}

			t.Logf("Second sync result: %d added, %d skipped, %d failed",
				len(result2.Added), len(result2.Skipped), len(result2.Failed))
		})
	}
}

// Helper functions to modify resources

func modifyCommand(t *testing.T, commandPath string) {
	content, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatalf("Failed to read command: %v", err)
	}

	modified := string(content) + "\nModified content for force test.\n"
	if err := os.WriteFile(commandPath, []byte(modified), 0644); err != nil {
		t.Fatalf("Failed to modify command: %v", err)
	}
}

func modifySkill(t *testing.T, skillDir string) {
	skillPath := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("Failed to read skill: %v", err)
	}

	modified := string(content) + "\nModified content for force test.\n"
	if err := os.WriteFile(skillPath, []byte(modified), 0644); err != nil {
		t.Fatalf("Failed to modify skill: %v", err)
	}
}

func modifyAgent(t *testing.T, agentPath string) {
	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("Failed to read agent: %v", err)
	}

	modified := string(content) + "\nModified content for force test.\n"
	if err := os.WriteFile(agentPath, []byte(modified), 0644); err != nil {
		t.Fatalf("Failed to modify agent: %v", err)
	}
}
