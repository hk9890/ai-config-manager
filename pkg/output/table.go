package output

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// TableData represents structured table data
type TableData struct {
	Headers []string     `json:"headers" yaml:"headers"`
	Rows    [][]string   `json:"rows" yaml:"rows"`
	Options TableOptions `json:"-" yaml:"-"`
}

// TableBuilder provides a fluent API for building tables
type TableBuilder struct {
	data *TableData
}

// NewTable creates a new TableBuilder with the given headers
func NewTable(headers ...string) *TableBuilder {
	return &TableBuilder{
		data: &TableData{
			Headers: headers,
			Rows:    [][]string{},
			Options: TableOptions{
				ShowBorders: true,
				AutoWrap:    true,
			},
		},
	}
}

// AddRow adds a row to the table
func (tb *TableBuilder) AddRow(cols ...string) *TableBuilder {
	tb.data.Rows = append(tb.data.Rows, cols)
	return tb
}

// AddSeparator adds an empty row for visual grouping
func (tb *TableBuilder) AddSeparator() *TableBuilder {
	tb.data.Rows = append(tb.data.Rows, make([]string, len(tb.data.Headers)))
	return tb
}

// WithOptions sets custom table options
func (tb *TableBuilder) WithOptions(opts TableOptions) *TableBuilder {
	tb.data.Options = opts
	return tb
}

// WithResponsive enables terminal-aware column sizing
func (tb *TableBuilder) WithResponsive() *TableBuilder {
	tb.data.Options.Responsive = true
	return tb
}

// WithTerminalWidth sets an explicit terminal width (mainly for testing)
// Pass 0 to use auto-detection (default behavior)
func (tb *TableBuilder) WithTerminalWidth(width int) *TableBuilder {
	tb.data.Options.TerminalWidth = width
	return tb
}

// WithDynamicColumn marks a specific column index to stretch and fill remaining space.
// Pass -1 to use the rightmost visible column (default behavior).
// The dynamic column will expand to use all remaining terminal width after
// allocating minimum widths to other columns.
//
// Example:
//
//	table := NewTable("Name", "Description", "Status").
//	    WithResponsive().
//	    WithDynamicColumn(1) // Description column fills remaining space
func (tb *TableBuilder) WithDynamicColumn(index int) *TableBuilder {
	tb.data.Options.DynamicColumn = index
	return tb
}

// WithMinColumnWidths sets minimum widths for each column.
// Columns will never shrink below these widths (except when hidden entirely).
// Pass fewer widths than columns to only constrain the first N columns.
//
// Example:
//
//	table := NewTable("Name", "Description", "Status").
//	    WithResponsive().
//	    WithMinColumnWidths(10, 15, 8) // Name: 10, Description: 15, Status: 8
func (tb *TableBuilder) WithMinColumnWidths(widths ...int) *TableBuilder {
	tb.data.Options.MinColumnWidths = widths
	return tb
}

// Format outputs the table in the specified format
func (tb *TableBuilder) Format(format Format) error {
	return FormatOutput(tb.data, format)
}

// formatTableData renders TableData in the requested format
func formatTableData(data *TableData, format Format) error {
	switch format {
	case Table:
		return renderTable(data)
	case JSON:
		return EncodeJSON(os.Stdout, data)
	case YAML:
		return EncodeYAML(os.Stdout, data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// determineVisibleColumns calculates which columns should be displayed based on
// available terminal width. Columns are hidden right-to-left when space is insufficient.
//
// Parameters:
//   - termWidth: Available terminal width in characters
//   - numCols: Total number of columns
//   - minWidths: Minimum width for each column
//
// Returns:
//   - Slice of column indices that should be visible (0-indexed)
//
// Algorithm:
//  1. Calculate border overhead: numCols + 1 (one | per column + final |)
//  2. Iterate left-to-right, accumulating column widths + borders
//  3. Stop when next column would exceed terminal width
//  4. Always show at least the first column (index 0)
func determineVisibleColumns(termWidth, numCols int, minWidths []int) []int {
	if numCols == 0 {
		return []int{}
	}

	visible := []int{}
	usedWidth := 0

	for i := 0; i < numCols; i++ {
		// Get minimum width for this column
		minWidth := getMinWidth(i, minWidths)

		// Calculate border overhead: one | per column + final | = numCols + 1
		// When checking if the NEXT column fits, we use len(visible)+2
		borders := len(visible) + 2

		// Check if this column fits
		if usedWidth+minWidth+borders <= termWidth {
			visible = append(visible, i)
			usedWidth += minWidth
		} else {
			// Column doesn't fit, stop here
			break
		}
	}

	// Ensure at least first column is visible
	if len(visible) == 0 {
		visible = []int{0}
	}

	return visible
}

// allocateColumnWidths distributes terminal width across visible columns.
// Fixed columns get their minimum width, and the dynamic column gets remaining space.
//
// Parameters:
//   - termWidth: Available terminal width
//   - visibleCols: Indices of visible columns
//   - minWidths: Minimum widths for all columns
//   - dynamicColIndex: Index of column to stretch (-1 = last visible)
//
// Returns:
//   - Map of column index to allocated width
//
// Algorithm:
//  1. Calculate border overhead: numVisibleCols + 1 (one | per column + final |)
//  2. Allocate minimum widths to all fixed columns (non-dynamic)
//  3. Give remaining space to dynamic column (minimum 15 chars)
func allocateColumnWidths(termWidth int, visibleCols []int, minWidths []int, dynamicColIndex int) map[int]int {
	if len(visibleCols) == 0 {
		return make(map[int]int)
	}

	widths := make(map[int]int)
	borders := len(visibleCols) + 1

	// Determine which column is dynamic
	// -1 means rightmost visible column
	// If the configured dynamic column is not visible, fall back to rightmost visible
	dynIdx := dynamicColIndex
	if dynIdx == -1 {
		dynIdx = visibleCols[len(visibleCols)-1]
	} else {
		// Check if the configured dynamic column is actually visible
		isVisible := false
		for _, colIdx := range visibleCols {
			if colIdx == dynIdx {
				isVisible = true
				break
			}
		}
		if !isVisible {
			// Fall back to rightmost visible column
			dynIdx = visibleCols[len(visibleCols)-1]
		}
	}

	// Allocate minimum widths to fixed columns (all except dynamic)
	fixedSpace := 0
	for _, colIdx := range visibleCols {
		if colIdx != dynIdx {
			minWidth := getMinWidth(colIdx, minWidths)
			widths[colIdx] = minWidth
			fixedSpace += minWidth
		}
	}

	// Allocate remaining space to dynamic column
	dynamicWidth := termWidth - fixedSpace - borders
	if dynamicWidth < 15 {
		dynamicWidth = 15 // Enforce minimum dynamic column width
	}
	widths[dynIdx] = dynamicWidth

	return widths
}

// getMinWidth returns the minimum width for a column, with fallback defaults.
// Returns the value from minWidths if available, otherwise returns a sensible default.
func getMinWidth(colIndex int, minWidths []int) int {
	// Default minimum widths if not specified
	const defaultMinWidth = 10

	if colIndex < len(minWidths) && minWidths[colIndex] > 0 {
		return minWidths[colIndex]
	}
	return defaultMinWidth
}

// filterColumnsData filters headers and rows to only include visible columns.
func filterColumnsData(headers []string, rows [][]string, visibleCols []int) ([]string, [][]string) {
	if len(visibleCols) == 0 {
		return []string{}, [][]string{}
	}

	// Filter headers
	filteredHeaders := make([]string, len(visibleCols))
	for i, colIdx := range visibleCols {
		if colIdx < len(headers) {
			filteredHeaders[i] = headers[colIdx]
		}
	}

	// Filter rows
	filteredRows := make([][]string, len(rows))
	for i, row := range rows {
		filteredRow := make([]string, len(visibleCols))
		for j, colIdx := range visibleCols {
			if colIdx < len(row) {
				filteredRow[j] = row[colIdx]
			}
		}
		filteredRows[i] = filteredRow
	}

	return filteredHeaders, filteredRows
}

// renderTable renders TableData as a human-readable table
//
// Terminal Width Detection:
// The terminal width is detected once at render time using NewTerminal().Width().
// This width is used by tablewriter to format the table with appropriate column widths
// and text wrapping.
//
// Responsive Column Sizing:
// When responsive mode is enabled (TableOptions.Responsive = true):
//  1. Determines which columns are visible based on terminal width
//  2. Hides columns right-to-left when space is insufficient
//  3. Allocates minimum widths to fixed columns
//  4. Expands dynamic column to fill remaining space
//  5. Truncates text with "..." when it exceeds column width
//
// Non-TTY Fallback:
// Responsive sizing only applies when output is a terminal (IsTTY() == true).
// When output is redirected/piped, tables use fixed widths.
//
// Terminal Resize Limitation:
// If the terminal is resized after the table has been rendered, the existing output
// will NOT automatically reflow or adjust. This is expected behavior for CLI tools:
// - The rendered output is immutable text in the shell buffer
// - Most CLI tools work this way (ls, git log, grep, etc.)
// - To see the table with new terminal dimensions, simply re-run the command
//
// This design is intentional and aligns with standard CLI tool behavior.
func renderTable(data *TableData) error {
	headers := data.Headers
	rows := data.Rows

	// Apply responsive sizing if enabled and terminal width is available
	if data.Options.Responsive {
		termWidth := data.Options.TerminalWidth
		if termWidth == 0 && IsTTY() {
			term, err := NewTerminal()
			if err == nil && term.Width() > 0 {
				termWidth = term.Width()
			}
		}

		// Only apply responsive sizing if we have a valid terminal width
		// and it meets the minimum threshold
		minTermWidth := data.Options.MinTerminalWidth
		if minTermWidth == 0 {
			minTermWidth = 15 // Default minimum
		}

		if termWidth >= minTermWidth {
			// Determine visible columns based on available width
			visibleCols := determineVisibleColumns(termWidth, len(headers), data.Options.MinColumnWidths)

			// Filter headers and rows to only include visible columns
			headers, rows = filterColumnsData(headers, rows, visibleCols)

			// Allocate widths for visible columns
			dynamicColIndex := data.Options.DynamicColumn
			columnWidths := allocateColumnWidths(termWidth, visibleCols, data.Options.MinColumnWidths, dynamicColIndex)

			// Create column width options for tablewriter
			// Map needs to be based on the NEW column indices after filtering
			twColumnWidths := make(map[int]int)
			for newIdx, origIdx := range visibleCols {
				if width, ok := columnWidths[origIdx]; ok {
					twColumnWidths[newIdx] = width
				}
			}

			// Create table with responsive width options (v1.1.3 API)
			table := tablewriter.NewTable(os.Stdout,
				tablewriter.WithMaxWidth(termWidth),
				tablewriter.WithColumnWidths(twColumnWidths),
				tablewriter.WithRowAutoWrap(tw.WrapTruncate), // Truncate text with ellipsis
			)

			// Convert headers to []any
			headerInterfaces := make([]any, len(headers))
			for i, h := range headers {
				headerInterfaces[i] = h
			}
			table.Header(headerInterfaces...)

			// Add rows
			for _, row := range rows {
				_ = table.Append(row)
			}

			_ = table.Render()
			return nil
		}
	}

	// Non-responsive mode or terminal too narrow - use default rendering
	table := tablewriter.NewWriter(os.Stdout)

	// Convert []string to []any for Header method
	headerInterfaces := make([]any, len(headers))
	for i, h := range headers {
		headerInterfaces[i] = h
	}
	table.Header(headerInterfaces...)

	for _, row := range rows {
		_ = table.Append(row)
	}

	_ = table.Render()
	return nil
}
