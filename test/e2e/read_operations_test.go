//go:build e2e

package e2e

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// TestE2E_BasicReadOperations tests the basic read operations for the repository.
// This includes:
// - repo list (table format)
// - repo list --format=json (JSON format)
// - repo show <id> (detailed resource view)
func TestE2E_BasicReadOperations(t *testing.T) {
	// Setup: Load config and sync resources
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)
	t.Logf("Using config: %s", configPath)
	t.Logf("Using repo path: %s", repoPath)

	// Register cleanup
	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Sync resources first to populate the repository
	// Note: We need to set AIMGR_REPO_PATH because sync command doesn't read repo.path from config yet
	t.Log("Step 1: Syncing resources to populate repository...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	syncStdout, syncStderr, syncErr := runAimgrWithEnv(t, configPath, env, "repo", "sync")

	if syncErr != nil {
		t.Fatalf("Failed to sync resources: %v\nStdout: %s\nStderr: %s", syncErr, syncStdout, syncStderr)
	}

	t.Logf("Sync completed successfully")

	// Parse sync stats to verify we have resources
	stats := parseSyncStats(t, syncStdout)
	totalResources := stats.Added + stats.Updated
	if totalResources == 0 {
		t.Fatal("No resources were synced - cannot test read operations")
	}
	t.Logf("Synced %d resources (Added=%d, Updated=%d)", totalResources, stats.Added, stats.Updated)

	// Test 1: repo list (table format)
	t.Run("RepoList_TableFormat", func(t *testing.T) {
		t.Log("Testing 'repo list' with default table format...")
		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list")

		if err != nil {
			t.Fatalf("'repo list' failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify output is not empty
		if strings.TrimSpace(stdout) == "" {
			t.Error("'repo list' output is empty")
		}

		// Verify table format (should have header with columns)
		lines := strings.Split(stdout, "\n")
		if len(lines) < 2 {
			t.Errorf("Expected at least 2 lines (header + data), got %d", len(lines))
		}

		// Verify header contains expected columns
		// Note: Table format shows NAME and DESCRIPTION columns
		outputLower := strings.ToLower(stdout)
		expectedColumns := []string{"name", "description"}
		missingColumns := []string{}
		for _, col := range expectedColumns {
			if !strings.Contains(outputLower, col) {
				missingColumns = append(missingColumns, col)
			}
		}
		if len(missingColumns) > 0 {
			t.Errorf("Table output missing columns %v. First 5 lines:\n%s", missingColumns, strings.Join(lines[:min(5, len(lines))], "\n"))
		}

		// Verify we have resource rows (at least one data row)
		hasDataRows := false
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line != "" && !strings.HasPrefix(line, "-") { // Skip separator lines
				hasDataRows = true
				break
			}
		}
		if !hasDataRows {
			t.Error("No resource rows found in table output")
		}

		t.Logf("✓ Table format validated (%d lines)", len(lines))
	})

	// Test 2: repo list --format=json
	t.Run("RepoList_JSONFormat", func(t *testing.T) {
		t.Log("Testing 'repo list --format=json'...")
		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")

		if err != nil {
			t.Fatalf("'repo list --format=json' failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		// Verify output is not empty
		if strings.TrimSpace(stdout) == "" {
			t.Fatal("'repo list --format=json' output is empty")
		}

		// Verify JSON is valid and parseable
		items := parseRepoListJSON(t, stdout)

		// Verify we have at least one resource
		if len(items) == 0 {
			t.Error("No resources found in JSON output")
		}
		t.Logf("Found %d resources in JSON output", len(items))

		// Verify JSON structure has expected fields
		for i, item := range items {
			// Check for Name field (always present)
			if name, ok := item["Name"].(string); !ok || name == "" {
				t.Errorf("Resource %d has invalid 'Name' field: %v", i, item["Name"])
			}

			// Check if this is a package (has ResourceCount) or a resource (has Type)
			if resourceCount, isPackage := item["ResourceCount"]; isPackage {
				// This is a package
				if _, ok := resourceCount.(float64); !ok {
					t.Errorf("Package %d has invalid 'ResourceCount' field: %v", i, resourceCount)
				}
				// Packages should have Description
				if desc, ok := item["Description"].(string); !ok {
					t.Errorf("Package %d missing or invalid 'Description' field: %v", i, desc)
				}
			} else {
				// This is a resource (command, skill, agent)
				if resType, ok := item["Type"].(string); !ok || resType == "" {
					t.Errorf("Resource %d has invalid 'Type' field: %v", i, item["Type"])
				} else {
					// Verify type is one of the known types (including package from Type field)
					validTypes := map[string]bool{
						"command": true,
						"skill":   true,
						"agent":   true,
						"package": true,
					}
					if !validTypes[resType] {
						t.Errorf("Resource %d has unknown type '%s'", i, resType)
					}
				}

				// Path is optional for packages, required for other resources
				if resType, _ := item["Type"].(string); resType != "package" {
					if path, ok := item["Path"].(string); !ok || path == "" {
						t.Errorf("Resource %d has invalid 'Path' field: %v", i, item["Path"])
					}
				}
			}
		}

		t.Logf("✓ JSON format validated with %d resources", len(items))
	})

	// Test 3: repo show <id>
	t.Run("RepoShow", func(t *testing.T) {
		t.Log("Testing 'repo show <id>'...")

		// First, get a list of resources to find a valid ID
		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
		if err != nil {
			t.Fatalf("Failed to get resource list: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		items := parseRepoListJSON(t, stdout)
		if len(items) == 0 {
			t.Fatal("No resources available to test 'repo show'")
		}

		// Test with different resource types
		testedTypes := make(map[string]bool)
		for _, item := range items {
			resType, ok := item["Type"].(string)
			if !ok {
				// Skip packages (they don't have a Type field)
				continue
			}

			// Try to test at least one of each type
			if testedTypes[resType] {
				continue
			}

			// Get resource name
			resName, ok := item["Name"].(string)
			if !ok || resName == "" {
				continue
			}

			// Format as type/name for repo show command
			resID := resType + "/" + resName

			t.Logf("Testing 'repo show' with %s resource: %s", resType, resID)

			// Run repo show
			showStdout, showStderr, showErr := runAimgrWithEnv(t, configPath, env, "repo", "show", resID)

			if showErr != nil {
				t.Errorf("'repo show %s' failed: %v\nStdout: %s\nStderr: %s", resID, showErr, showStdout, showStderr)
				continue
			}

			// Verify output is not empty
			if strings.TrimSpace(showStdout) == "" {
				t.Errorf("'repo show %s' output is empty", resID)
				continue
			}

			// Verify output contains key information
			// Note: The output format is "Command:", "Skill:", or "Agent:" - not "Name:" or "Type:"
			outputLower := strings.ToLower(showStdout)

			// Check for the resource type label (e.g., "command:", "skill:", "agent:")
			hasTypeLabel := strings.Contains(outputLower, resType+":") ||
				strings.Contains(outputLower, strings.Title(resType)+":")
			if !hasTypeLabel {
				t.Errorf("'repo show %s' output missing type label '%s:'. Output:\n%s", resID, resType, showStdout)
			}

			// Check for description field
			if !strings.Contains(outputLower, "description") {
				t.Errorf("'repo show %s' output missing 'description' field. Output:\n%s", resID, showStdout)
			}

			// Verify resource name appears in output
			if !strings.Contains(showStdout, resName) {
				t.Errorf("'repo show %s' output does not contain resource name '%s'. Output:\n%s", resID, resName, showStdout)
			}

			// Verify resource type appears in output
			if !strings.Contains(outputLower, resType) {
				t.Errorf("'repo show %s' output does not contain type '%s'. Output:\n%s", resID, resType, showStdout)
			}

			testedTypes[resType] = true
			t.Logf("✓ Successfully tested 'repo show' for %s resource: %s", resType, resID)

			// Stop after testing 2-3 resources (or all types)
			if len(testedTypes) >= 3 {
				break
			}
		}

		if len(testedTypes) == 0 {
			t.Error("Failed to test 'repo show' with any resource")
		} else {
			t.Logf("✓ Tested 'repo show' with %d resource type(s): %v", len(testedTypes), getMapKeys(testedTypes))
		}
	})

	// Test 4: Verify different resource types are present
	t.Run("VerifyResourceTypes", func(t *testing.T) {
		t.Log("Verifying different resource types are present...")

		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
		if err != nil {
			t.Fatalf("Failed to get resource list: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		items := parseRepoListJSON(t, stdout)

		// Count resources by type
		typeCounts := make(map[string]int)
		for _, item := range items {
			if resType, ok := item["Type"].(string); ok {
				typeCounts[resType]++
			}
		}

		t.Logf("Resource type counts: %v", typeCounts)

		// Verify we have at least some resource types
		if len(typeCounts) == 0 {
			t.Error("No resource types found")
		}

		// Log statistics
		for resType, count := range typeCounts {
			t.Logf("  %s: %d", resType, count)
		}

		t.Logf("✓ Found %d different resource type(s)", len(typeCounts))
	})

	// Test 5: Validate output format consistency
	t.Run("OutputFormatConsistency", func(t *testing.T) {
		t.Log("Testing output format consistency...")

		// Run repo list multiple times and verify consistent output
		stdout1, _, err1 := runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
		if err1 != nil {
			t.Fatalf("First 'repo list' failed: %v", err1)
		}

		stdout2, _, err2 := runAimgrWithEnv(t, configPath, env, "repo", "list", "--format=json")
		if err2 != nil {
			t.Fatalf("Second 'repo list' failed: %v", err2)
		}

		// Parse both outputs
		items1 := parseRepoListJSON(t, stdout1)
		items2 := parseRepoListJSON(t, stdout2)

		// Verify same number of resources
		if len(items1) != len(items2) {
			t.Errorf("Inconsistent resource counts: first=%d, second=%d", len(items1), len(items2))
		}

		// Verify same resources are present (by name)
		names1 := make(map[string]bool)
		for _, item := range items1 {
			if name, ok := item["Name"].(string); ok {
				names1[name] = true
			}
		}

		names2 := make(map[string]bool)
		for _, item := range items2 {
			if name, ok := item["Name"].(string); ok {
				names2[name] = true
			}
		}

		// Check for differences
		for name := range names1 {
			if !names2[name] {
				t.Errorf("Resource '%s' present in first list but not second", name)
			}
		}

		for name := range names2 {
			if !names1[name] {
				t.Errorf("Resource '%s' present in second list but not first", name)
			}
		}

		t.Logf("✓ Output format is consistent across multiple calls")
	})
}

// TestE2E_RepoListFiltering tests repo list with various filtering options.
func TestE2E_RepoListFiltering(t *testing.T) {
	// Setup
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)

	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Sync resources
	t.Log("Syncing resources...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	syncStdout, syncStderr, syncErr := runAimgrWithEnv(t, configPath, env, "repo", "sync")

	if syncErr != nil {
		t.Fatalf("Failed to sync: %v\nStdout: %s\nStderr: %s", syncErr, syncStdout, syncStderr)
	}

	// Test: List only commands
	t.Run("ListCommands", func(t *testing.T) {
		t.Log("Testing 'repo list' with command filter...")

		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list", "command/*", "--format=json")
		if err != nil {
			t.Fatalf("'repo list command/*' failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if strings.TrimSpace(stdout) == "" {
			t.Log("No commands found (may be valid if sync didn't include commands)")
			return
		}

		items := parseRepoListJSON(t, stdout)

		// Verify all items are commands
		for i, item := range items {
			resType, ok := item["Type"].(string)
			if !ok {
				t.Errorf("Resource %d missing Type field", i)
				continue
			}

			if resType != "command" {
				t.Errorf("Expected only commands, but found %s: %v", resType, item)
			}
		}

		t.Logf("✓ Found %d command(s)", len(items))
	})

	// Test: List only skills
	t.Run("ListSkills", func(t *testing.T) {
		t.Log("Testing 'repo list' with skill filter...")

		stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "list", "skill/*", "--format=json")
		if err != nil {
			t.Fatalf("'repo list skill/*' failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}

		if strings.TrimSpace(stdout) == "" {
			t.Log("No skills found (may be valid if sync didn't include skills)")
			return
		}

		items := parseRepoListJSON(t, stdout)

		// Verify all items are skills
		for i, item := range items {
			resType, ok := item["Type"].(string)
			if !ok {
				t.Errorf("Resource %d missing Type field", i)
				continue
			}

			if resType != "skill" {
				t.Errorf("Expected only skills, but found %s: %v", resType, item)
			}
		}

		t.Logf("✓ Found %d skill(s)", len(items))
	})
}

// TestE2E_RepoShowInvalidID tests repo show with an invalid resource ID.
func TestE2E_RepoShowInvalidID(t *testing.T) {
	configPath := loadTestConfig(t, "e2e-test")
	repoPath := getRepoPathFromConfig(t, configPath)

	t.Cleanup(func() {
		cleanTestRepo(t, repoPath)
	})

	// Sync resources first
	t.Log("Syncing resources...")
	env := map[string]string{"AIMGR_REPO_PATH": repoPath}
	syncStdout, syncStderr, syncErr := runAimgrWithEnv(t, configPath, env, "repo", "sync")

	if syncErr != nil {
		t.Fatalf("Failed to sync: %v\nStdout: %s\nStderr: %s", syncErr, syncStdout, syncStderr)
	}

	// Test with invalid resource ID
	t.Log("Testing 'repo show' with invalid resource ID...")
	invalidID := "nonexistent-resource-12345"
	stdout, stderr, err := runAimgrWithEnv(t, configPath, env, "repo", "show", invalidID)

	// Expect error or appropriate message
	if err == nil {
		// Check if output indicates resource not found
		outputLower := strings.ToLower(stdout + stderr)
		if !strings.Contains(outputLower, "not found") && !strings.Contains(outputLower, "does not exist") {
			t.Errorf("Expected 'not found' error for invalid resource ID, but got success. Output: %s", stdout)
		}
	}

	t.Logf("✓ Invalid resource ID handled appropriately")
}

// parseRepoListJSON parses the structured JSON output from repo list --format=json
// The output has the format: {"packages": [...], "resources": [...]}
func parseRepoListJSON(t *testing.T, output string) []map[string]interface{} {
	t.Helper()

	var result struct {
		Packages  []map[string]interface{} `json:"packages"`
		Resources []map[string]interface{} `json:"resources"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Combine packages and resources into a single list
	allItems := make([]map[string]interface{}, 0, len(result.Packages)+len(result.Resources))
	allItems = append(allItems, result.Resources...)
	allItems = append(allItems, result.Packages...)

	return allItems
}

// Helper function to get map keys as a slice
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// assertJSONValid verifies that a string is valid JSON
func assertJSONValid(t *testing.T, jsonStr string) {
	t.Helper()

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Errorf("Invalid JSON: %v\nJSON: %s", err, jsonStr)
	}
}

// assertTableFormat verifies that output is in table format (has header and rows)
func assertTableFormat(t *testing.T, output string) {
	t.Helper()

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Errorf("Expected table format with header and data rows, got %d lines", len(lines))
		return
	}

	// Check for header (should contain column names)
	headerPattern := regexp.MustCompile(`(?i)name.*type.*source`)
	if !headerPattern.MatchString(lines[0]) {
		t.Errorf("Expected table header with 'Name', 'Type', 'Source'. Got: %s", lines[0])
	}
}
