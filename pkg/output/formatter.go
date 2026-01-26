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
	// Print added resources
	if len(result.Added) > 0 {
		fmt.Println("Added Resources:")
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Name", "Type")

		for _, res := range result.Added {
			if err := table.Append(res.Name, res.Type); err != nil {
				return fmt.Errorf("failed to append row: %w", err)
			}
		}
		if err := table.Render(); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
		fmt.Println()
	}

	// Print skipped resources
	if len(result.Skipped) > 0 {
		fmt.Println("Skipped Resources:")
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Name", "Type", "Reason")

		for _, res := range result.Skipped {
			if err := table.Append(res.Name, res.Type, res.Message); err != nil {
				return fmt.Errorf("failed to append row: %w", err)
			}
		}
		if err := table.Render(); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
		fmt.Println()
	}

	// Print failed resources
	if len(result.Failed) > 0 {
		fmt.Println("Failed Resources:")
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Name", "Type", "Error")

		for _, res := range result.Failed {
			if err := table.Append(res.Name, res.Type, res.Message); err != nil {
				return fmt.Errorf("failed to append row: %w", err)
			}
		}
		if err := table.Render(); err != nil {
			return fmt.Errorf("failed to render table: %w", err)
		}
		fmt.Println()
	}

	// Print summary
	totalResources := result.CommandCount + result.SkillCount + result.AgentCount + result.PackageCount
	fmt.Printf("Summary: %d resources (%d commands, %d skills, %d agents, %d packages)\n",
		totalResources, result.CommandCount, result.SkillCount, result.AgentCount, result.PackageCount)
	fmt.Printf("  Added: %d, Skipped: %d, Failed: %d\n",
		len(result.Added), len(result.Skipped), len(result.Failed))

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
