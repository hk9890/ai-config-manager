//go:build unit

package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEncodeJSON_Simple(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test-resource",
		"type":   "command",
		"active": true,
	}

	var buf bytes.Buffer
	err := EncodeJSON(&buf, data)
	if err != nil {
		t.Fatalf("EncodeJSON failed: %v", err)
	}

	// Parse to verify valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse encoded JSON: %v", err)
	}

	if result["name"] != "test-resource" {
		t.Errorf("Expected name='test-resource', got %v", result["name"])
	}
	if result["type"] != "command" {
		t.Errorf("Expected type='command', got %v", result["type"])
	}
}

func TestEncodeJSON_Indented(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}

	var buf bytes.Buffer
	err := EncodeJSON(&buf, data)
	if err != nil {
		t.Fatalf("EncodeJSON failed: %v", err)
	}

	output := buf.String()
	// Check for indentation (2 spaces)
	if !strings.Contains(output, "  \"key\"") {
		t.Error("Expected JSON to be indented with 2 spaces")
	}
}

func TestEncodeJSON_NestedStructure(t *testing.T) {
	type Resource struct {
		Name     string            `json:"name"`
		Type     string            `json:"type"`
		Metadata map[string]string `json:"metadata"`
	}

	data := Resource{
		Name: "test",
		Type: "command",
		Metadata: map[string]string{
			"version": "1.0",
			"author":  "test-author",
		},
	}

	var buf bytes.Buffer
	err := EncodeJSON(&buf, data)
	if err != nil {
		t.Fatalf("EncodeJSON failed: %v", err)
	}

	// Parse to verify structure
	var result Resource
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse encoded JSON: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected name='test', got %q", result.Name)
	}
	if result.Metadata["version"] != "1.0" {
		t.Errorf("Expected metadata version='1.0', got %q", result.Metadata["version"])
	}
}

func TestEncodeYAML_Simple(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test-resource",
		"type":   "command",
		"active": true,
	}

	var buf bytes.Buffer
	err := EncodeYAML(&buf, data)
	if err != nil {
		t.Fatalf("EncodeYAML failed: %v", err)
	}

	// Parse to verify valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse encoded YAML: %v", err)
	}

	if result["name"] != "test-resource" {
		t.Errorf("Expected name='test-resource', got %v", result["name"])
	}
	if result["type"] != "command" {
		t.Errorf("Expected type='command', got %v", result["type"])
	}
}

func TestEncodeYAML_Indented(t *testing.T) {
	data := map[string]interface{}{
		"parent": map[string]string{
			"child": "value",
		},
	}

	var buf bytes.Buffer
	err := EncodeYAML(&buf, data)
	if err != nil {
		t.Fatalf("EncodeYAML failed: %v", err)
	}

	output := buf.String()
	// Check for indentation (2 spaces)
	if !strings.Contains(output, "  child:") {
		t.Errorf("Expected YAML to be indented with 2 spaces, got:\n%s", output)
	}
}

func TestEncodeYAML_NestedStructure(t *testing.T) {
	type Resource struct {
		Name     string            `yaml:"name"`
		Type     string            `yaml:"type"`
		Metadata map[string]string `yaml:"metadata"`
	}

	data := Resource{
		Name: "test",
		Type: "command",
		Metadata: map[string]string{
			"version": "1.0",
			"author":  "test-author",
		},
	}

	var buf bytes.Buffer
	err := EncodeYAML(&buf, data)
	if err != nil {
		t.Fatalf("EncodeYAML failed: %v", err)
	}

	// Parse to verify structure
	var result Resource
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse encoded YAML: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected name='test', got %q", result.Name)
	}
	if result.Metadata["version"] != "1.0" {
		t.Errorf("Expected metadata version='1.0', got %q", result.Metadata["version"])
	}
}

func TestEncodeJSON_EmptyData(t *testing.T) {
	data := map[string]interface{}{}

	var buf bytes.Buffer
	err := EncodeJSON(&buf, data)
	if err != nil {
		t.Fatalf("EncodeJSON with empty data failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "{}" {
		t.Errorf("Expected '{}' for empty data, got %q", output)
	}
}

func TestEncodeYAML_EmptyData(t *testing.T) {
	data := map[string]interface{}{}

	var buf bytes.Buffer
	err := EncodeYAML(&buf, data)
	if err != nil {
		t.Fatalf("EncodeYAML with empty data failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "{}" {
		t.Errorf("Expected '{}' for empty data, got %q", output)
	}
}
