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
