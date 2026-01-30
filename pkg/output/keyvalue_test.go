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

func TestKeyValueBuilder_Simple(t *testing.T) {
	builder := NewKeyValue("Test Title")
	builder.Add("Key1", "Value1")
	builder.Add("Key2", "Value2")

	if builder.data.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %q", builder.data.Title)
	}
	if len(builder.data.Pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(builder.data.Pairs))
	}

	// Check chaining
	builder.Add("Key3", "Value3")
	if len(builder.data.Pairs) != 3 {
		t.Errorf("Expected 3 pairs after chaining, got %d", len(builder.data.Pairs))
	}
}

func TestKeyValueBuilder_AddSection(t *testing.T) {
	builder := NewKeyValue("Title")
	builder.Add("Key1", "Value1")
	builder.AddSection()
	builder.Add("Key2", "Value2")

	if len(builder.data.Pairs) != 3 {
		t.Errorf("Expected 3 pairs (including separator), got %d", len(builder.data.Pairs))
	}

	// Check separator is empty
	separator := builder.data.Pairs[1]
	if separator.Key != "" || separator.Value != "" {
		t.Errorf("Expected empty separator, got Key=%q Value=%q", separator.Key, separator.Value)
	}
}

func TestKeyValueData_FormatJSON(t *testing.T) {
	data := &KeyValueData{
		Title: "Test Data",
		Pairs: []KeyValue{
			{Key: "Name", Value: "test-resource"},
			{Key: "Type", Value: "command"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatKeyValueData(data, JSON)
	if err != nil {
		t.Fatalf("formatKeyValueData(JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Parse JSON to verify structure
	var result KeyValueData
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result.Title != "Test Data" {
		t.Errorf("Expected title 'Test Data' in JSON, got %q", result.Title)
	}
	if len(result.Pairs) != 2 {
		t.Errorf("Expected 2 pairs in JSON, got %d", len(result.Pairs))
	}
}

func TestKeyValueData_FormatYAML(t *testing.T) {
	data := &KeyValueData{
		Title: "Test Data",
		Pairs: []KeyValue{
			{Key: "Name", Value: "test-resource"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatKeyValueData(data, YAML)
	if err != nil {
		t.Fatalf("formatKeyValueData(YAML) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Parse YAML to verify structure
	var result KeyValueData
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	if result.Title != "Test Data" {
		t.Errorf("Expected title 'Test Data' in YAML, got %q", result.Title)
	}
	if len(result.Pairs) != 1 {
		t.Errorf("Expected 1 pair in YAML, got %d", len(result.Pairs))
	}
}

func TestKeyValueData_FormatTable(t *testing.T) {
	data := &KeyValueData{
		Title: "Resource Information",
		Pairs: []KeyValue{
			{Key: "Name", Value: "test-command"},
			{Key: "Type", Value: "command"},
			{Key: "Status", Value: "active"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatKeyValueData(data, Table)
	if err != nil {
		t.Fatalf("formatKeyValueData(Table) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected content
	if !strings.Contains(output, "Resource Information") {
		t.Error("Expected output to contain title 'Resource Information'")
	}
	if !strings.Contains(output, "====") {
		t.Error("Expected output to contain title underline")
	}
	if !strings.Contains(output, "Name: test-command") {
		t.Error("Expected output to contain 'Name: test-command'")
	}
	if !strings.Contains(output, "Type: command") {
		t.Error("Expected output to contain 'Type: command'")
	}
	if !strings.Contains(output, "Status: active") {
		t.Error("Expected output to contain 'Status: active'")
	}
}

func TestKeyValueData_FormatTable_WithSection(t *testing.T) {
	data := &KeyValueData{
		Title: "Test",
		Pairs: []KeyValue{
			{Key: "Key1", Value: "Value1"},
			{Key: "", Value: ""}, // Section separator
			{Key: "Key2", Value: "Value2"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatKeyValueData(data, Table)
	if err != nil {
		t.Fatalf("formatKeyValueData(Table) with section failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Count newlines to verify section separator added a blank line
	lines := strings.Split(output, "\n")
	hasBlankLine := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			hasBlankLine = true
			break
		}
	}

	if !hasBlankLine {
		t.Error("Expected output to contain blank line for section separator")
	}
}

func TestKeyValueData_FormatTable_NoTitle(t *testing.T) {
	data := &KeyValueData{
		Title: "", // No title
		Pairs: []KeyValue{
			{Key: "Key", Value: "Value"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatKeyValueData(data, Table)
	if err != nil {
		t.Fatalf("formatKeyValueData(Table) with no title failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not have title underline
	if strings.Contains(output, "====") {
		t.Error("Expected no title underline when title is empty")
	}
	if strings.Contains(output, "Key: Value") {
		// Good - has the key-value pair
	} else {
		t.Error("Expected output to contain 'Key: Value'")
	}
}

func TestKeyValueData_FormatUnsupported(t *testing.T) {
	data := &KeyValueData{
		Title: "Test",
		Pairs: []KeyValue{},
	}

	err := formatKeyValueData(data, Format("invalid"))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestKeyValueBuilder_Chaining(t *testing.T) {
	// Test fluent API chaining
	builder := NewKeyValue("Chained").
		Add("K1", "V1").
		AddSection().
		Add("K2", "V2").
		Add("K3", "V3")

	if len(builder.data.Pairs) != 4 {
		t.Errorf("Expected 4 pairs after chaining, got %d", len(builder.data.Pairs))
	}
}

func TestNewKeyValue_EmptyTitle(t *testing.T) {
	builder := NewKeyValue("")
	if builder.data.Title != "" {
		t.Errorf("Expected empty title, got %q", builder.data.Title)
	}
	if len(builder.data.Pairs) != 0 {
		t.Errorf("Expected 0 pairs initially, got %d", len(builder.data.Pairs))
	}
}
