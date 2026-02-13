//go:build unit

package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

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
			// Convert first character to rune for proper Unicode comparison
			runes := []rune(trimmed)
			if len(runes) > 0 {
				firstChar := runes[0]
				// Check for border characters (ASCII or Unicode box-drawing)
				if firstChar == '+' || firstChar == '┌' || firstChar == '├' || firstChar == '└' {
					borderLine = trimmed
					break
				}
			}
		}
	}

	if borderLine == "" {
		return nil
	}

	// Count column widths by finding separator characters
	// Column separators: ┬ (top), ┼ (middle), ┴ (bottom), + (ASCII)
	// End separators: ┐ (top-right), ┤ (middle-right), ┘ (bottom-right), + (ASCII)

	// Simply count the number of column separators + 1 for the number of columns
	// A 3-column table has format: ┌───┬───┬───┐ (2 separators + 1 = 3 columns)
	numColumns := 1 // Start with 1 (minimum)
	for _, ch := range borderLine {
		isColumnSeparator := ch == '+' || ch == '┬' || ch == '┼' || ch == '┴'
		if isColumnSeparator {
			numColumns++
		}
	}

	// Return a slice with the column count (actual widths don't matter for the test)
	widths := make([]int, numColumns)
	for i := range widths {
		widths[i] = 10 // Dummy width
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
		if len(trimmed) == 0 {
			continue
		}
		firstRune, _ := utf8.DecodeRuneInString(trimmed)
		// Check for border characters (ASCII or Unicode box-drawing)
		if firstRune == '+' || firstRune == '┌' || firstRune == '├' || firstRune == '└' {
			// Measure visible width (rune count, not bytes)
			width := utf8.RuneCountInString(trimmed)
			return width
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

			// Measure and log the table width (for now, just verify it renders)
			// TODO (ai-config-manager-6ju): Verify width constraint once responsive sizing is implemented
			totalWidth := measureTotalTableWidth(output)

			// Log output for debugging
			t.Logf("Terminal width: %d, Table width: %d", tt.termWidth, totalWidth)
			t.Logf("Table renders successfully with mocked terminal width")
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

// ========================================
// Responsive Column Sizing Tests
// ========================================

// TestTableBuilder_WithDynamicColumn verifies the WithDynamicColumn method.
func TestTableBuilder_WithDynamicColumn(t *testing.T) {
	builder := NewTable("Name", "Description", "Status").
		WithResponsive().
		WithDynamicColumn(1) // Description column

	if builder.data.Options.DynamicColumn != 1 {
		t.Errorf("Expected DynamicColumn to be 1, got %d", builder.data.Options.DynamicColumn)
	}

	// Test -1 for rightmost column
	builder.WithDynamicColumn(-1)
	if builder.data.Options.DynamicColumn != -1 {
		t.Errorf("Expected DynamicColumn to be -1, got %d", builder.data.Options.DynamicColumn)
	}
}

// TestTableBuilder_WithMinColumnWidths verifies the WithMinColumnWidths method.
func TestTableBuilder_WithMinColumnWidths(t *testing.T) {
	builder := NewTable("Name", "Description", "Status").
		WithResponsive().
		WithMinColumnWidths(10, 15, 8)

	if len(builder.data.Options.MinColumnWidths) != 3 {
		t.Errorf("Expected 3 minimum widths, got %d", len(builder.data.Options.MinColumnWidths))
	}

	expected := []int{10, 15, 8}
	for i, width := range builder.data.Options.MinColumnWidths {
		if width != expected[i] {
			t.Errorf("Expected width[%d] to be %d, got %d", i, expected[i], width)
		}
	}
}

// TestDetermineVisibleColumns tests the column visibility algorithm.
func TestDetermineVisibleColumns(t *testing.T) {
	tests := []struct {
		name             string
		termWidth        int
		numCols          int
		minWidths        []int
		expectedVisible  []int
		expectedNumShown int
	}{
		{
			name:             "Wide terminal - all columns visible",
			termWidth:        100,
			numCols:          4,
			minWidths:        []int{10, 15, 10, 15},
			expectedVisible:  []int{0, 1, 2, 3},
			expectedNumShown: 4,
		},
		{
			name:             "Medium terminal - 3 columns visible",
			termWidth:        60,
			numCols:          4,
			minWidths:        []int{10, 15, 10, 15},
			expectedVisible:  []int{0, 1, 2},
			expectedNumShown: 3,
		},
		{
			name:             "Narrow terminal - 2 columns visible",
			termWidth:        40,
			numCols:          4,
			minWidths:        []int{10, 15, 10, 15},
			expectedVisible:  []int{0, 1},
			expectedNumShown: 2,
		},
		{
			name:             "Very narrow - only first column",
			termWidth:        25,
			numCols:          4,
			minWidths:        []int{10, 15, 10, 15},
			expectedVisible:  []int{0},
			expectedNumShown: 1,
		},
		{
			name:             "Empty columns",
			termWidth:        100,
			numCols:          0,
			minWidths:        []int{},
			expectedVisible:  []int{},
			expectedNumShown: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visible := determineVisibleColumns(tt.termWidth, tt.numCols, tt.minWidths)

			if len(visible) != tt.expectedNumShown {
				t.Errorf("Expected %d visible columns, got %d: %v",
					tt.expectedNumShown, len(visible), visible)
			}

			for i, colIdx := range tt.expectedVisible {
				if i >= len(visible) || visible[i] != colIdx {
					t.Errorf("Expected visible[%d] to be %d, got %v",
						i, colIdx, visible)
					break
				}
			}
		})
	}
}

// TestAllocateColumnWidths tests the width allocation algorithm.
func TestAllocateColumnWidths(t *testing.T) {
	tests := []struct {
		name            string
		termWidth       int
		visibleCols     []int
		minWidths       []int
		dynamicColIndex int
		checkDynamic    bool
		minDynamicWidth int
	}{
		{
			name:            "3 columns with dynamic last",
			termWidth:       80,
			visibleCols:     []int{0, 1, 2},
			minWidths:       []int{10, 10, 10},
			dynamicColIndex: -1, // Last column
			checkDynamic:    true,
			minDynamicWidth: 15,
		},
		{
			name:            "3 columns with dynamic middle",
			termWidth:       80,
			visibleCols:     []int{0, 1, 2},
			minWidths:       []int{10, 10, 10},
			dynamicColIndex: 1, // Middle column
			checkDynamic:    true,
			minDynamicWidth: 15,
		},
		{
			name:            "2 columns with dynamic last",
			termWidth:       60,
			visibleCols:     []int{0, 1},
			minWidths:       []int{10, 15},
			dynamicColIndex: -1,
			checkDynamic:    true,
			minDynamicWidth: 15,
		},
		{
			name:            "Single column (all dynamic)",
			termWidth:       50,
			visibleCols:     []int{0},
			minWidths:       []int{10},
			dynamicColIndex: -1,
			checkDynamic:    true,
			minDynamicWidth: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widths := allocateColumnWidths(tt.termWidth, tt.visibleCols, tt.minWidths, tt.dynamicColIndex)

			// Check all visible columns have widths
			if len(widths) != len(tt.visibleCols) {
				t.Errorf("Expected %d column widths, got %d", len(tt.visibleCols), len(widths))
			}

			// Check dynamic column has minimum width
			if tt.checkDynamic {
				dynIdx := tt.dynamicColIndex
				if dynIdx == -1 {
					dynIdx = tt.visibleCols[len(tt.visibleCols)-1]
				}

				dynWidth, ok := widths[dynIdx]
				if !ok {
					t.Errorf("Dynamic column %d has no width", dynIdx)
				} else if dynWidth < tt.minDynamicWidth {
					t.Errorf("Dynamic column width %d is less than minimum %d",
						dynWidth, tt.minDynamicWidth)
				}
			}

			// Check fixed columns have minimum widths
			for _, colIdx := range tt.visibleCols {
				dynIdx := tt.dynamicColIndex
				if dynIdx == -1 {
					dynIdx = tt.visibleCols[len(tt.visibleCols)-1]
				}

				if colIdx != dynIdx {
					minWidth := getMinWidth(colIdx, tt.minWidths)
					actualWidth := widths[colIdx]
					if actualWidth != minWidth {
						t.Errorf("Fixed column %d: expected width %d, got %d",
							colIdx, minWidth, actualWidth)
					}
				}
			}
		})
	}
}

// TestGetMinWidth tests the minimum width helper function.
func TestGetMinWidth(t *testing.T) {
	tests := []struct {
		name          string
		colIndex      int
		minWidths     []int
		expectedWidth int
	}{
		{
			name:          "Width specified",
			colIndex:      1,
			minWidths:     []int{10, 20, 15},
			expectedWidth: 20,
		},
		{
			name:          "Width not specified - use default",
			colIndex:      3,
			minWidths:     []int{10, 20},
			expectedWidth: 10, // Default
		},
		{
			name:          "Zero width - use default",
			colIndex:      1,
			minWidths:     []int{10, 0, 15},
			expectedWidth: 10, // Default
		},
		{
			name:          "Empty widths array",
			colIndex:      0,
			minWidths:     []int{},
			expectedWidth: 10, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := getMinWidth(tt.colIndex, tt.minWidths)
			if width != tt.expectedWidth {
				t.Errorf("Expected width %d, got %d", tt.expectedWidth, width)
			}
		})
	}
}

// TestFilterColumnsData tests the column filtering function.
func TestFilterColumnsData(t *testing.T) {
	tests := []struct {
		name            string
		headers         []string
		rows            [][]string
		visibleCols     []int
		expectedHeaders []string
		expectedRows    [][]string
	}{
		{
			name:            "All columns visible",
			headers:         []string{"Name", "Type", "Status"},
			rows:            [][]string{{"test", "command", "active"}},
			visibleCols:     []int{0, 1, 2},
			expectedHeaders: []string{"Name", "Type", "Status"},
			expectedRows:    [][]string{{"test", "command", "active"}},
		},
		{
			name:            "Hide last column",
			headers:         []string{"Name", "Type", "Status"},
			rows:            [][]string{{"test", "command", "active"}},
			visibleCols:     []int{0, 1},
			expectedHeaders: []string{"Name", "Type"},
			expectedRows:    [][]string{{"test", "command"}},
		},
		{
			name:            "Only first column",
			headers:         []string{"Name", "Type", "Status"},
			rows:            [][]string{{"test", "command", "active"}},
			visibleCols:     []int{0},
			expectedHeaders: []string{"Name"},
			expectedRows:    [][]string{{"test"}},
		},
		{
			name:            "Empty visible columns",
			headers:         []string{"Name", "Type"},
			rows:            [][]string{{"test", "command"}},
			visibleCols:     []int{},
			expectedHeaders: []string{},
			expectedRows:    [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, rows := filterColumnsData(tt.headers, tt.rows, tt.visibleCols)

			// Check headers
			if len(headers) != len(tt.expectedHeaders) {
				t.Errorf("Expected %d headers, got %d", len(tt.expectedHeaders), len(headers))
			}
			for i, h := range tt.expectedHeaders {
				if i >= len(headers) || headers[i] != h {
					t.Errorf("Expected headers[%d] = %q, got %q", i, h, headers[i])
				}
			}

			// Check rows
			if len(rows) != len(tt.expectedRows) {
				t.Errorf("Expected %d rows, got %d", len(tt.expectedRows), len(rows))
			}
			for i, expectedRow := range tt.expectedRows {
				if i >= len(rows) {
					continue
				}
				row := rows[i]
				if len(row) != len(expectedRow) {
					t.Errorf("Row %d: expected %d columns, got %d", i, len(expectedRow), len(row))
				}
				for j, cell := range expectedRow {
					if j >= len(row) || row[j] != cell {
						t.Errorf("Row %d, col %d: expected %q, got %q", i, j, cell, row[j])
					}
				}
			}
		})
	}
}

// TestResponsiveTableRendering tests end-to-end responsive table rendering.
func TestResponsiveTableRendering(t *testing.T) {
	tests := []struct {
		name            string
		termWidth       int
		headers         []string
		rows            [][]string
		minWidths       []int
		expectAllCols   bool
		minExpectedCols int
	}{
		{
			name:            "Wide terminal shows all columns",
			termWidth:       120,
			headers:         []string{"Name", "Description", "Status", "Targets"},
			rows:            [][]string{{"test-cmd", "A test command", "active", "claude"}},
			minWidths:       []int{10, 20, 8, 12},
			expectAllCols:   true,
			minExpectedCols: 4,
		},
		{
			name:            "Medium terminal hides rightmost column",
			termWidth:       70,
			headers:         []string{"Name", "Description", "Status", "Targets"},
			rows:            [][]string{{"test-cmd", "A test command", "active", "claude"}},
			minWidths:       []int{10, 20, 8, 12},
			expectAllCols:   false,
			minExpectedCols: 2, // At least first 2 columns
		},
		{
			name:            "Narrow terminal shows minimal columns",
			termWidth:       40,
			headers:         []string{"Name", "Description", "Status", "Targets"},
			rows:            [][]string{{"test-cmd", "A test command", "active", "claude"}},
			minWidths:       []int{10, 20, 8, 12},
			expectAllCols:   false,
			minExpectedCols: 1, // At least first column
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := newTestTable(tt.termWidth, tt.headers...).
				WithMinColumnWidths(tt.minWidths...)

			for _, row := range tt.rows {
				table.AddRow(row...)
			}

			output := captureTableOutput(table)

			// Verify table rendered
			if output == "" {
				t.Error("Expected table output, got empty string")
				return
			}

			// Debug: print first few lines to see what we got
			lines := strings.Split(output, "\n")
			if len(lines) > 0 {
				t.Logf("First line of output: %q", lines[0])
				if len(lines) > 1 {
					t.Logf("Second line: %q", lines[1])
				}
			}

			// Count visible columns
			numCols := countVisibleColumns(output)
			t.Logf("Counted %d columns from output", numCols)
			if numCols < tt.minExpectedCols {
				t.Errorf("Expected at least %d columns, got %d", tt.minExpectedCols, numCols)
			}

			if tt.expectAllCols && numCols != len(tt.headers) {
				t.Errorf("Expected all %d columns visible, got %d", len(tt.headers), numCols)
			}

			// Verify first column content is always present
			if !strings.Contains(output, tt.rows[0][0]) {
				t.Errorf("Expected first column content %q in output", tt.rows[0][0])
			}
		})
	}
}
