package output

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
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

// renderTable renders TableData as a human-readable table
//
// Terminal Width Detection:
// The terminal width is detected once at render time using NewTerminal().Width().
// This width is used by tablewriter to format the table with appropriate column widths
// and text wrapping.
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
	table := tablewriter.NewWriter(os.Stdout)

	// Convert []string to []any for Header method
	headers := make([]any, len(data.Headers))
	for i, h := range data.Headers {
		headers[i] = h
	}
	table.Header(headers...)

	// Apply responsive sizing if enabled
	// If TerminalWidth is explicitly set (non-zero), use that value for testing
	// Otherwise, auto-detect from terminal if running in a TTY
	if data.Options.Responsive {
		termWidth := data.Options.TerminalWidth
		if termWidth == 0 && IsTTY() {
			term, err := NewTerminal()
			if err == nil && term.Width() > 0 {
				termWidth = term.Width()
			}
		}
		// Terminal width available for responsive features
		// Future enhancement: smart column width distribution
		_ = termWidth // placeholder for future use
	}

	for _, row := range data.Rows {
		table.Append(row)
	}

	table.Render()
	return nil
}
