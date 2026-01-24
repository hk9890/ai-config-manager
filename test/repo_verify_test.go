package test

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCLIRepoVerifyBasic tests basic verify command execution
func TestCLIRepoVerifyBasic(t *testing.T) {
	// Just test that verify runs without crashing
	output, _ := runAimgr(t, "repo", "verify")

	// Should have the header
	if !strings.Contains(output, "Repository Verification") {
		t.Errorf("Expected 'Repository Verification' header, got: %s", output)
	}

	// Should have some status output
	validStatuses := []string{
		"No issues found",
		"ERRORS found",
		"Warnings only",
	}

	hasStatus := false
	for _, status := range validStatuses {
		if strings.Contains(output, status) {
			hasStatus = true
			break
		}
	}

	if !hasStatus {
		t.Errorf("Expected a status message in output, got: %s", output)
	}
}

// TestCLIRepoVerifyJSON tests JSON output format
func TestCLIRepoVerifyJSON(t *testing.T) {
	output, _ := runAimgr(t, "repo", "verify", "--json")

	// Parse JSON output
	var result struct {
		ResourcesWithoutMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"resources_without_metadata,omitempty"`
		OrphanedMetadata []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Path string `json:"path"`
		} `json:"orphaned_metadata,omitempty"`
		MissingSourcePaths []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Path       string `json:"path"`
			SourcePath string `json:"source_path"`
		} `json:"missing_source_paths,omitempty"`
		TypeMismatches []struct {
			Name         string `json:"name"`
			ResourceType string `json:"resource_type"`
			MetadataType string `json:"metadata_type"`
		} `json:"type_mismatches,omitempty"`
		HasErrors   bool `json:"has_errors"`
		HasWarnings bool `json:"has_warnings"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Just verify we got valid JSON with the right structure
	// We can't assert specific counts since tests share a repository
	t.Logf("Verify JSON output parsed successfully")
	t.Logf("  Resources without metadata: %d", len(result.ResourcesWithoutMetadata))
	t.Logf("  Orphaned metadata: %d", len(result.OrphanedMetadata))
	t.Logf("  Missing source paths: %d", len(result.MissingSourcePaths))
	t.Logf("  Type mismatches: %d", len(result.TypeMismatches))
	t.Logf("  Has errors: %v", result.HasErrors)
	t.Logf("  Has warnings: %v", result.HasWarnings)
}

// TestCLIRepoVerifyHelp tests the help output
func TestCLIRepoVerifyHelp(t *testing.T) {
	output, err := runAimgr(t, "repo", "verify", "--help")
	if err != nil {
		t.Fatalf("Failed to get help: %v", err)
	}

	// Check for key help sections
	expectedContent := []string{
		"Check for consistency issues",
		"--fix",
		"--json",
		"Resources without metadata",
		"Orphaned metadata",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output missing '%s'\nOutput: %s", expected, output)
		}
	}
}
