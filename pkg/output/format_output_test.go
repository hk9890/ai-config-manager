//go:build unit

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatOutput_TableData(t *testing.T) {
	data := &TableData{
		Headers: []string{"Name", "Type"},
		Rows: [][]string{
			{"test-cmd", "command"},
			{"test-skill", "skill"},
		},
	}

	// Test JSON format
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(data, JSON)
	if err != nil {
		t.Fatalf("FormatOutput(TableData, JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result TableData
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(result.Headers))
	}
}

func TestFormatOutput_KeyValueData(t *testing.T) {
	data := &KeyValueData{
		Title: "Test Resource",
		Pairs: []KeyValue{
			{Key: "Name", Value: "test-cmd"},
			{Key: "Type", Value: "command"},
		},
	}

	// Test YAML format
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(data, YAML)
	if err != nil {
		t.Fatalf("FormatOutput(KeyValueData, YAML) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result KeyValueData
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	if result.Title != "Test Resource" {
		t.Errorf("Expected title 'Test Resource', got %q", result.Title)
	}
}

// MockRenderable implements the Renderable interface for testing
type MockRenderable struct {
	RenderCalled bool
	RenderFormat Format
	RenderError  error
}

func (m *MockRenderable) Render(format Format) error {
	m.RenderCalled = true
	m.RenderFormat = format
	return m.RenderError
}

func TestFormatOutput_Renderable(t *testing.T) {
	mock := &MockRenderable{}

	err := FormatOutput(mock, JSON)
	if err != nil {
		t.Fatalf("FormatOutput(Renderable) failed: %v", err)
	}

	if !mock.RenderCalled {
		t.Error("Expected Render() to be called")
	}
	if mock.RenderFormat != JSON {
		t.Errorf("Expected format JSON, got %v", mock.RenderFormat)
	}
}

func TestFormatOutput_Renderable_Error(t *testing.T) {
	expectedErr := fmt.Errorf("render error")
	mock := &MockRenderable{RenderError: expectedErr}

	err := FormatOutput(mock, Table)
	if err == nil {
		t.Fatal("Expected error from Renderable.Render()")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestFormatOutput_Generic_JSON(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"type": "command",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(data, JSON)
	if err != nil {
		t.Fatalf("FormatOutput(generic, JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name='test', got %v", result["name"])
	}
}

func TestFormatOutput_Generic_YAML(t *testing.T) {
	data := map[string]string{
		"name": "test",
		"type": "skill",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(data, YAML)
	if err != nil {
		t.Fatalf("FormatOutput(generic, YAML) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result map[string]string
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	if result["type"] != "skill" {
		t.Errorf("Expected type='skill', got %v", result["type"])
	}
}

func TestFormatOutput_Generic_Table_Error(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}

	err := FormatOutput(data, Table)
	if err == nil {
		t.Fatal("Expected error for generic type with table format")
	}
	if !strings.Contains(err.Error(), "table format not supported") {
		t.Errorf("Expected 'table format not supported' error, got: %v", err)
	}
}

func TestFormatGeneric_UnsupportedFormat(t *testing.T) {
	data := map[string]string{}

	err := formatGeneric(data, Format("invalid"))
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestFormatOutput_NilData(t *testing.T) {
	// Test with nil data - should encode as null
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(nil, JSON)
	if err != nil {
		t.Fatalf("FormatOutput(nil, JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := strings.TrimSpace(buf.String())
	if output != "null" {
		t.Errorf("Expected 'null' for nil data, got %q", output)
	}
}

func TestFormatOutput_ComplexStruct(t *testing.T) {
	type ComplexData struct {
		Name      string            `json:"name" yaml:"name"`
		Resources []string          `json:"resources" yaml:"resources"`
		Metadata  map[string]string `json:"metadata" yaml:"metadata"`
	}

	data := ComplexData{
		Name:      "complex-test",
		Resources: []string{"cmd1", "cmd2"},
		Metadata: map[string]string{
			"version": "1.0",
		},
	}

	// Test JSON
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := FormatOutput(data, JSON)
	if err != nil {
		t.Fatalf("FormatOutput(ComplexData, JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result ComplexData
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if result.Name != "complex-test" {
		t.Errorf("Expected name='complex-test', got %q", result.Name)
	}
	if len(result.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(result.Resources))
	}
}

func TestFormatOutput_TableBuilder_Integration(t *testing.T) {
	// Test integration with TableBuilder
	builder := NewTable("Name", "Status")
	builder.AddRow("resource1", "active")
	builder.AddRow("resource2", "inactive")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := builder.Format(JSON)
	if err != nil {
		t.Fatalf("TableBuilder.Format(JSON) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result TableData
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(result.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result.Rows))
	}
}

func TestFormatOutput_KeyValueBuilder_Integration(t *testing.T) {
	// Test integration with KeyValueBuilder
	builder := NewKeyValue("Test Title")
	builder.Add("Key1", "Value1")
	builder.Add("Key2", "Value2")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := builder.Format(YAML)
	if err != nil {
		t.Fatalf("KeyValueBuilder.Format(YAML) failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result KeyValueData
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse YAML output: %v", err)
	}

	if result.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %q", result.Title)
	}
	if len(result.Pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(result.Pairs))
	}
}
