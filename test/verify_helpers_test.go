package test

import (
	"encoding/json"
	"testing"
)

// AssertVerifyClean checks that repo verify reports no errors.
// This helper should be called after successful import/update operations
// to ensure the repository state is valid.
func AssertVerifyClean(t *testing.T) {
	t.Helper()

	// Run verify --format=json to get structured output
	output, _ := runAimgr(t, "repo", "verify", "--format=json")

	var result struct {
		OrphanedMetadata         []map[string]interface{} `json:"orphaned_metadata"`
		ResourcesWithoutMetadata []map[string]interface{} `json:"resources_without_metadata"`
		TypeMismatches           []map[string]interface{} `json:"type_mismatches"`
		PackagesWithMissingRefs  []map[string]interface{} `json:"packages_with_missing_refs"`
		HasErrors                bool                     `json:"has_errors"`
		HasWarnings              bool                     `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse verify JSON output: %v\nOutput: %s", err, output)
	}

	// Check for errors
	if result.HasErrors {
		t.Errorf("Verify reported errors (should be clean)")
	}

	if len(result.OrphanedMetadata) > 0 {
		t.Errorf("Verify found %d orphaned metadata entries:", len(result.OrphanedMetadata))
		for _, m := range result.OrphanedMetadata {
			t.Errorf("  - %v", m)
		}
	}

	if len(result.TypeMismatches) > 0 {
		t.Errorf("Verify found %d type mismatches:", len(result.TypeMismatches))
		for _, m := range result.TypeMismatches {
			t.Errorf("  - %v", m)
		}
	}

	if len(result.PackagesWithMissingRefs) > 0 {
		t.Errorf("Verify found %d packages with missing refs:", len(result.PackagesWithMissingRefs))
		for _, p := range result.PackagesWithMissingRefs {
			t.Errorf("  - %v", p)
		}
	}

	// Warnings are OK (e.g., resources without metadata), but log them
	if result.HasWarnings && len(result.ResourcesWithoutMetadata) > 0 {
		t.Logf("Note: %d resources without metadata (warnings, not errors)", len(result.ResourcesWithoutMetadata))
	}
}

// AssertVerifyHasErrors checks that verify reports expected errors.
// Use this for negative tests where you expect verify to catch issues.
func AssertVerifyHasErrors(t *testing.T, expectOrphaned, expectMismatches, expectMissingRefs bool) {
	t.Helper()

	// Run verify --format=json to get structured output
	output, _ := runAimgr(t, "repo", "verify", "--format=json")

	var result struct {
		OrphanedMetadata        []map[string]interface{} `json:"orphaned_metadata"`
		TypeMismatches          []map[string]interface{} `json:"type_mismatches"`
		PackagesWithMissingRefs []map[string]interface{} `json:"packages_with_missing_refs"`
		HasErrors               bool                     `json:"has_errors"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse verify JSON output: %v\nOutput: %s", err, output)
	}

	if !result.HasErrors {
		t.Errorf("Expected verify to report errors, but it passed")
	}

	if expectOrphaned && len(result.OrphanedMetadata) == 0 {
		t.Errorf("Expected orphaned metadata, but found none")
	}

	if expectMismatches && len(result.TypeMismatches) == 0 {
		t.Errorf("Expected type mismatches, but found none")
	}

	if expectMissingRefs && len(result.PackagesWithMissingRefs) == 0 {
		t.Errorf("Expected packages with missing refs, but found none")
	}
}
