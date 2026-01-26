package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hk9890/ai-config-manager/pkg/repo"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Format represents an output format type
type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
	YAML  Format = "yaml"
)

// ParseFormat parses a format string into a Format type
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "table":
		return Table, nil
	case "json":
		return JSON, nil
	case "yaml":
		return YAML, nil
	default:
		return "", fmt.Errorf("invalid format: %s (valid: table, json, yaml)", s)
	}
}

// BulkOperationResult represents the result of a bulk repository operation
type BulkOperationResult struct {
	Added        []ResourceResult `json:"added" yaml:"added"`
	Skipped      []ResourceResult `json:"skipped" yaml:"skipped"`
	Failed       []ResourceResult `json:"failed" yaml:"failed"`
	CommandCount int              `json:"command_count" yaml:"command_count"`
	SkillCount   int              `json:"skill_count" yaml:"skill_count"`
	AgentCount   int              `json:"agent_count" yaml:"agent_count"`
	PackageCount int              `json:"package_count" yaml:"package_count"`
}

// ResourceResult represents the result of a single resource operation
type ResourceResult struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
}

// FromBulkImportResult converts a repo.BulkImportResult to BulkOperationResult
func FromBulkImportResult(result *repo.BulkImportResult) *BulkOperationResult {
	bor := &BulkOperationResult{
		Added:        make([]ResourceResult, 0, len(result.Added)),
		Skipped:      make([]ResourceResult, 0, len(result.Skipped)),
		Failed:       make([]ResourceResult, 0, len(result.Failed)),
		CommandCount: result.CommandCount,
		SkillCount:   result.SkillCount,
		AgentCount:   result.AgentCount,
		PackageCount: result.PackageCount,
	}

	// Convert added resources
	for _, path := range result.Added {
		bor.Added = append(bor.Added, ResourceResult{
			Name: extractResourceName(path),
			Type: extractResourceType(path),
			Path: path,
		})
	}

	// Convert skipped resources
	for _, path := range result.Skipped {
		bor.Skipped = append(bor.Skipped, ResourceResult{
			Name:    extractResourceName(path),
			Type:    extractResourceType(path),
			Path:    path,
			Message: "already exists",
		})
	}

	// Convert failed resources
	for _, fail := range result.Failed {
		bor.Failed = append(bor.Failed, ResourceResult{
			Name:    extractResourceName(fail.Path),
			Type:    extractResourceType(fail.Path),
			Path:    fail.Path,
			Message: fail.Message,
		})
	}

	return bor
}

// extractResourceName extracts the resource name from a file path
func extractResourceName(path string) string {
	// Get the base filename
	name := path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		name = path[idx+1:]
	}
	if idx := strings.LastIndex(path, "\\"); idx != -1 {
		name = path[idx+1:]
	}

	// Remove file extensions
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimSuffix(name, ".package.json")

	return name
}

// extractResourceType extracts the resource type from a file path
func extractResourceType(path string) string {
	if strings.Contains(path, "/commands/") || strings.Contains(path, "\\commands\\") {
		return "command"
	}
	if strings.Contains(path, "/skills/") || strings.Contains(path, "\\skills\\") {
		return "skill"
	}
	if strings.Contains(path, "/agents/") || strings.Contains(path, "\\agents\\") {
		return "agent"
	}
	if strings.Contains(path, "/packages/") || strings.Contains(path, "\\packages\\") {
		return "package"
	}
	if strings.HasSuffix(path, ".package.json") {
		return "package"
	}
	if strings.HasSuffix(path, ".md") {
		return "command"
	}
	return "resource"
}

// FormatBulkResult formats a BulkOperationResult according to the specified format
func FormatBulkResult(result *BulkOperationResult, format Format) error {
	switch format {
	case Table:
		return formatAsTable(result)
	case JSON:
		return formatAsJSON(result)
	case YAML:
		return formatAsYAML(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// formatAsTable formats the result as a table (human-readable)
func formatAsTable(result *BulkOperationResult) error {
	// Create unified table with all resources
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("NAME", "STATUS", "MESSAGE")

	// Add all resources to table
	hasContent := false

	// Add successful resources
	for _, res := range result.Added {
		status := "SUCCESS"
		message := res.Message
		if message == "" {
			message = "Added to repository"
		}
		if err := table.Append(fmt.Sprintf("%s/%s", res.Type, res.Name), status, message); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
		hasContent = true
	}

	// Add skipped resources
	for _, res := range result.Skipped {
		status := "SKIPPED"
		message := res.Message
		if message == "" {
			message = "Already exists"
		}
		if err := table.Append(fmt.Sprintf("%s/%s", res.Type, res.Name), status, message); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
		hasContent = true
	}

	// Add failed resources
	for _, res := range result.Failed {
		status := "FAILED"
		message := res.Message
		if message == "" {
			message = "Unknown error"
		}
		if err := table.Append(fmt.Sprintf("%s/%s", res.Type, res.Name), status, message); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
		hasContent = true
	}

	// Render table if there's content
	if hasContent {
		if err := table.Render(); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
		fmt.Println()
	}

	// Print summary line
	totalAdded := len(result.Added)
	totalSkipped := len(result.Skipped)
	totalFailed := len(result.Failed)
	totalResources := totalAdded + totalSkipped + totalFailed

	if totalResources == 0 {
		fmt.Println("No resources to process")
	} else {
		fmt.Printf("Summary: %d added, %d failed, %d skipped (%d total)\n",
			totalAdded, totalFailed, totalSkipped, totalResources)
	}

	// Show tip about JSON format if there are failures
	if totalFailed > 0 {
		fmt.Println()
		fmt.Println("⚠ Use --format=json to see detailed error messages")
	}

	return nil
}

// formatAsJSON formats the result as JSON
func formatAsJSON(result *BulkOperationResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// formatAsYAML formats the result as YAML
func formatAsYAML(result *BulkOperationResult) error {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(result)
}
