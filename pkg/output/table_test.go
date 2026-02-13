//go:build unit

package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTableBuilder_Simple(t *testing.T) {
	// Create a simple table
	builder := NewTable("Name", "Type", "Status")
	builder.AddRow("test-command", "command", "active")
	builder.AddRow("test-skill", "skill", "active")

	// Verify the data structure
	if len(builder.data.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(builder.data.Headers))
	}
	if len(builder.data.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(builder.data.Rows))
	}

	// Check chaining
	builder.AddRow("test-agent", "agent", "active")
	if len(builder.data.Rows) != 3 {
		t.Errorf("Expected 3 rows after chaining, got %d", len(builder.data.Rows))
	}
}

func TestTableBuilder_AddSeparator(t *testing.T) {
	builder := NewTable("Col1", "Col2")
	builder.AddRow("value1", "value2")
	builder.AddSeparator()
	builder.AddRow("value3", "value4")

	if len(builder.data.Rows) != 3 {
		t.Errorf("Expected 3 rows (including separator), got %d", len(builder.data.Rows))
	}

	// Check separator row is empty
	separatorRow := builder.data.Rows[1]
	if len(separatorRow) != 2 {
		t.Errorf("Expected separator row to have 2 columns, got %d", len(separatorRow))
	}
	for _, col := range separatorRow {
		if col != "" {
			t.Errorf("Expected empty separator column, got %q", col)
		}
	}
}

func TestTableBuilder_WithOptions(t *testing.T) {
	builder := NewTable("Header1")
	customOpts := TableOptions{
		ShowBorders: false,
		AutoWrap:    false,
	}
	builder.WithOptions(customOpts)

	if builder.data.Options.ShowBorders != false {
		t.Error("Expected ShowBorders to be false")
	}
	if builder.data.Options.AutoWrap != false {
		t.Error("Expected AutoWrap to be false")
	}
}

func TestTableData_FormatJSON(t *testing.T) {
	data := &TableData{
		Headers: []string{"Name", "Type"},
		Rows: [][]string{
			{"test-cmd", "command"},
			{"test-skill", "skill"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatTableData(data, JSON)
	if err != nil {
		t.Fatalf("formatTableData(JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Parse JSON to verify structure
	var result TableData
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result.Headers) != 2 {
		t.Errorf("Expected 2 headers in JSON, got %d", len(result.Headers))
	}
	if len(result.Rows) != 2 {
		t.Errorf("Expected 2 rows in JSON, got %d", len(result.Rows))
	}
}

func TestTableData_FormatYAML(t *testing.T) {
	data := &TableData{
		Headers: []string{"Name", "Type"},
		Rows: [][]string{
			{"test-cmd", "command"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatTableData(data, YAML)
	if err != nil {
		t.Fatalf("formatTableData(YAML) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Parse YAML to verify structure
	var result TableData
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	if len(result.Headers) != 2 {
		t.Errorf("Expected 2 headers in YAML, got %d", len(result.Headers))
	}
	if len(result.Rows) != 1 {
		t.Errorf("Expected 1 row in YAML, got %d", len(result.Rows))
	}
}

func TestTableData_FormatTable(t *testing.T) {
	data := &TableData{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"resource1", "active"},
			{"resource2", "inactive"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatTableData(data, Table)
	if err != nil {
		t.Fatalf("formatTableData(Table) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify table contains expected content
	// Note: Table may render with uppercase headers (NAME, STATUS)
	outputUpper := strings.ToUpper(output)
	if !strings.Contains(outputUpper, "NAME") {
		t.Errorf("Expected table output to contain 'NAME' header, got:\n%s", output)
	}
	if !strings.Contains(outputUpper, "STATUS") {
		t.Errorf("Expected table output to contain 'STATUS' header, got:\n%s", output)
	}
	if !strings.Contains(output, "resource1") {
		t.Errorf("Expected table output to contain 'resource1', got:\n%s", output)
	}
	if !strings.Contains(output, "resource2") {
		t.Errorf("Expected table output to contain 'resource2', got:\n%s", output)
	}
}

func TestTableData_FormatUnsupported(t *testing.T) {
	data := &TableData{
		Headers: []string{"Test"},
		Rows:    [][]string{},
	}

	err := formatTableData(data, Format("invalid"))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestNewTable_DefaultOptions(t *testing.T) {
	builder := NewTable("Header")

	// Check default options
	if !builder.data.Options.ShowBorders {
		t.Error("Expected ShowBorders to be true by default")
	}
	if !builder.data.Options.AutoWrap {
		t.Error("Expected AutoWrap to be true by default")
	}
}

func TestTableBuilder_Chaining(t *testing.T) {
	// Test fluent API chaining
	builder := NewTable("A", "B", "C").
		AddRow("1", "2", "3").
		AddSeparator().
		AddRow("4", "5", "6").
		WithOptions(TableOptions{ShowBorders: false, AutoWrap: false})

	if len(builder.data.Rows) != 3 {
		t.Errorf("Expected 3 rows after chaining, got %d", len(builder.data.Rows))
	}
	if builder.data.Options.ShowBorders {
		t.Error("Expected ShowBorders to be false after chaining")
	}
}

// ========================================
// Test Helpers for Responsive Table Testing
// ========================================

// newTestTable creates a TableBuilder with a fixed terminal width for testing.
// This allows deterministic testing of responsive table behavior without
// depending on the actual terminal size.
//
// Example:
//
//	table := newTestTable(100, "Name", "Description", "Status")
//	table.AddRow("test-command", "A long description...", "active")
//	output := captureTableOutput(table)
func newTestTable(termWidth int, headers ...string) *TableBuilder {
	return NewTable(headers...).
		WithResponsive().
		WithTerminalWidth(termWidth)
}

// captureTableOutput renders a table and captures its output as a string.
// This is useful for testing table rendering without polluting stdout.
//
// Example:
//
//	table := newTestTable(80, "Name", "Type")
//	table.AddRow("test", "command")
//	output := captureTableOutput(table)
//	if !strings.Contains(output, "test") {
//	    t.Error("Expected table to contain 'test'")
//	}
func captureTableOutput(table *TableBuilder) string {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Render table
	_ = table.Format(Table) // Ignore error for simplicity in tests

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// measureColumnWidths parses rendered table output and returns the visible
// width of each column (including borders and padding).
//
// This function analyzes the table's header row to determine column widths
// by locating the border characters (|) and measuring the distance between them.
//
// Returns a slice of column widths in the order they appear in the table.
// Returns nil if the table cannot be parsed.
//
// Example:
//
//	output := captureTableOutput(table)
//	widths := measureColumnWidths(output)
//	if widths[0] < 10 {
//	    t.Error("First column too narrow")
//	}
func measureColumnWidths(output string) []int {
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return nil // Table too short to analyze
	}

	// Find the border line (supports both ASCII and Unicode box-drawing)
	// ASCII format:  +------+------+
	// Unicode format: ┌──────┬──────┐ or ├──────┼──────┤
	var borderLine string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			firstChar := rune(trimmed[0])
			// Check for border characters (ASCII or Unicode box-drawing)
			if firstChar == '+' || firstChar == '┌' || firstChar == '├' || firstChar == '└' {
				borderLine = trimmed
				break
			}
		}
	}

	if borderLine == "" {
		return nil
	}

	// Count column widths by measuring distance between separator characters
	// Separators: + (ASCII), ┬ ┼ ┴ (Unicode)
	widths := []int{}
	lastPos := 0
	for i, ch := range borderLine {
		isSeparator := ch == '+' || ch == '┬' || ch == '┼' || ch == '┴' || ch == '┐' || ch == '┤' || ch == '┘'
		if isSeparator && i > 0 {
			widths = append(widths, i-lastPos)
			lastPos = i
		}
	}

	return widths
}

// measureTotalTableWidth returns the total width of a rendered table
// (including all borders and padding).
//
// Example:
//
//	output := captureTableOutput(table)
//	totalWidth := measureTotalTableWidth(output)
//	if totalWidth > 100 {
//	    t.Errorf("Table too wide: expected ≤100, got %d", totalWidth)
//	}
func measureTotalTableWidth(output string) int {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			firstChar := rune(trimmed[0])
			// Check for border characters (ASCII or Unicode box-drawing)
			if firstChar == '+' || firstChar == '┌' || firstChar == '├' || firstChar == '└' {
				// Measure visible width (character count)
				return len(trimmed)
			}
		}
	}
	return 0
}

// countVisibleColumns counts the number of columns in a rendered table.
//
// Example:
//
//	output := captureTableOutput(table)
//	numCols := countVisibleColumns(output)
//	if numCols != 3 {
//	    t.Errorf("Expected 3 columns, got %d", numCols)
//	}
func countVisibleColumns(output string) int {
	widths := measureColumnWidths(output)
	return len(widths)
}

// assertTextTruncated checks if text in the output contains truncation markers.
// Returns true if "..." or "…" (ellipsis) is found in the output.
//
// Example:
//
//	output := captureTableOutput(table)
//	if !assertTextTruncated(output) {
//	    t.Error("Expected text to be truncated with ellipsis")
//	}
func assertTextTruncated(output string) bool {
	return strings.Contains(output, "...") || strings.Contains(output, "…")
}

// ========================================
// Example Test Using Responsive Features
// ========================================

// TestTableBuilder_TerminalWidthControl demonstrates how to use the testing
// strategy for responsive table tests.
func TestTableBuilder_TerminalWidthControl(t *testing.T) {
	tests := []struct {
		name         string
		termWidth    int
		headers      []string
		rows         [][]string
		expectOutput string
	}{
		{
			name:      "Wide terminal",
			termWidth: 100,
			headers:   []string{"Name", "Description", "Status"},
			rows: [][]string{
				{"test-command", "A very long description that might need truncation", "active"},
			},
			expectOutput: "test-command",
		},
		{
			name:      "Narrow terminal",
			termWidth: 60,
			headers:   []string{"Name", "Description", "Status"},
			rows: [][]string{
				{"test-command", "A very long description that might need truncation", "active"},
			},
			expectOutput: "test-command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := newTestTable(tt.termWidth, tt.headers...)
			for _, row := range tt.rows {
				table.AddRow(row...)
			}

			output := captureTableOutput(table)

			// Verify content is present
			if !strings.Contains(output, tt.expectOutput) {
				t.Errorf("Expected output to contain %q, got:\n%s", tt.expectOutput, output)
			}

			// Verify terminal width constraint (with some tolerance for borders)
			totalWidth := measureTotalTableWidth(output)
			if totalWidth > tt.termWidth+5 {
				t.Errorf("Table width %d exceeds terminal width %d (with 5 char tolerance)", totalWidth, tt.termWidth)
			}

			// Log output for debugging
			t.Logf("Terminal width: %d, Table width: %d", tt.termWidth, totalWidth)
			t.Logf("Output:\n%s", output)
		})
	}
}

// TestTableBuilder_WithTerminalWidth verifies the WithTerminalWidth method works.
func TestTableBuilder_WithTerminalWidth(t *testing.T) {
	builder := NewTable("Header1", "Header2").
		WithResponsive().
		WithTerminalWidth(80)

	if !builder.data.Options.Responsive {
		t.Error("Expected Responsive to be enabled")
	}
	if builder.data.Options.TerminalWidth != 80 {
		t.Errorf("Expected TerminalWidth to be 80, got %d", builder.data.Options.TerminalWidth)
	}

	// Test chaining with AddRow
	builder.AddRow("value1", "value2")
	if len(builder.data.Rows) != 1 {
		t.Errorf("Expected 1 row after AddRow, got %d", len(builder.data.Rows))
	}
}

// TestTableBuilder_TerminalWidthZeroAutoDetect verifies that TerminalWidth=0
// triggers auto-detection (though in tests we can't test actual detection).
func TestTableBuilder_TerminalWidthZeroAutoDetect(t *testing.T) {
	builder := NewTable("Name", "Type").
		WithResponsive() // Don't set TerminalWidth, should default to 0

	if builder.data.Options.TerminalWidth != 0 {
		t.Errorf("Expected TerminalWidth to default to 0, got %d", builder.data.Options.TerminalWidth)
	}

	// Rendering should still work (may auto-detect or fall back to defaults)
	builder.AddRow("test", "command")
	output := captureTableOutput(builder)

	if !strings.Contains(output, "test") {
		t.Error("Expected table to render even with auto-detect width")
	}
}
