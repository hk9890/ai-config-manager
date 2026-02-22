package frontmatter

import (
	"testing"
)

func TestHasFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "valid frontmatter",
			content:  "---\ntitle: test\n---\n# Content",
			expected: true,
		},
		{
			name:     "no frontmatter",
			content:  "# Just a heading\nSome content",
			expected: false,
		},
		{
			name:     "empty frontmatter",
			content:  "---\n---\n# Content",
			expected: true,
		},
		{
			name:     "delimiter later in document",
			content:  "# Heading\n---\nNot frontmatter",
			expected: false,
		},
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "just delimiter",
			content:  "---",
			expected: false,
		},
		{
			name:     "frontmatter with CRLF",
			content:  "---\r\ntitle: test\r\n---\r\n# Content",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasFrontmatter([]byte(tt.content))
			if result != tt.expected {
				t.Errorf("HasFrontmatter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectNil       bool
		expectError     bool
		expectedFields  map[string]interface{}
		expectedContent string
	}{
		{
			name:        "valid frontmatter with fields",
			content:     "---\nmodel: sonnet-4.5\ntemperature: 0.7\n---\n# Content here",
			expectNil:   false,
			expectError: false,
			expectedFields: map[string]interface{}{
				"model":       "sonnet-4.5",
				"temperature": 0.7,
			},
			expectedContent: "# Content here",
		},
		{
			name:            "no frontmatter returns nil",
			content:         "# Just a heading\nSome content",
			expectNil:       true,
			expectError:     false,
			expectedFields:  nil,
			expectedContent: "",
		},
		{
			name:            "empty frontmatter",
			content:         "---\n---\n# Content",
			expectNil:       false,
			expectError:     false,
			expectedFields:  map[string]interface{}{},
			expectedContent: "# Content",
		},
		{
			name:        "complex YAML with nested objects",
			content:     "---\nconfig:\n  nested: value\n  list:\n    - item1\n    - item2\n---\nBody text",
			expectNil:   false,
			expectError: false,
			expectedFields: map[string]interface{}{
				"config": map[string]interface{}{
					"nested": "value",
					"list":   []interface{}{"item1", "item2"},
				},
			},
			expectedContent: "Body text",
		},
		{
			name:            "content with --- later in document",
			content:         "---\ntitle: test\n---\n# Heading\n\n---\n\nThis is a horizontal rule",
			expectNil:       false,
			expectError:     false,
			expectedFields:  map[string]interface{}{"title": "test"},
			expectedContent: "# Heading\n\n---\n\nThis is a horizontal rule",
		},
		{
			name:            "frontmatter with CRLF line endings",
			content:         "---\r\ntitle: test\r\n---\r\n# Content",
			expectNil:       false,
			expectError:     false,
			expectedFields:  map[string]interface{}{"title": "test"},
			expectedContent: "# Content",
		},
		{
			name:            "frontmatter with no trailing content",
			content:         "---\ntitle: test\n---\n",
			expectNil:       false,
			expectError:     false,
			expectedFields:  map[string]interface{}{"title": "test"},
			expectedContent: "",
		},
		{
			name:        "invalid YAML returns error",
			content:     "---\n: invalid: yaml:\n---\n# Content",
			expectNil:   false,
			expectError: true,
		},
		{
			name:            "array values",
			content:         "---\ntags:\n  - go\n  - yaml\n---\n# Content",
			expectNil:       false,
			expectError:     false,
			expectedFields:  map[string]interface{}{"tags": []interface{}{"go", "yaml"}},
			expectedContent: "# Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse([]byte(tt.content))

			if tt.expectError {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if tt.expectNil {
				if result != nil {
					t.Errorf("Parse() expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Parse() returned nil, expected non-nil")
			}

			if result.Content != tt.expectedContent {
				t.Errorf("Parse() content = %q, want %q", result.Content, tt.expectedContent)
			}

			// Check expected fields
			for key, expectedVal := range tt.expectedFields {
				actualVal, ok := result.Fields[key]
				if !ok {
					t.Errorf("Parse() missing field %q", key)
					continue
				}
				// Simple comparison for primitive types
				if !deepEqual(actualVal, expectedVal) {
					t.Errorf("Parse() field %q = %v, want %v", key, actualVal, expectedVal)
				}
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter *Frontmatter
		key         string
		expected    string
	}{
		{
			name: "existing string key",
			frontmatter: &Frontmatter{
				Fields: map[string]interface{}{"model": "sonnet-4.5"},
			},
			key:      "model",
			expected: "sonnet-4.5",
		},
		{
			name: "missing key returns empty string",
			frontmatter: &Frontmatter{
				Fields: map[string]interface{}{"model": "sonnet-4.5"},
			},
			key:      "nonexistent",
			expected: "",
		},
		{
			name: "non-string value returns empty string",
			frontmatter: &Frontmatter{
				Fields: map[string]interface{}{"temperature": 0.7},
			},
			key:      "temperature",
			expected: "",
		},
		{
			name:        "nil frontmatter returns empty string",
			frontmatter: nil,
			key:         "anything",
			expected:    "",
		},
		{
			name: "nil fields returns empty string",
			frontmatter: &Frontmatter{
				Fields: nil,
			},
			key:      "anything",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.frontmatter.GetString(tt.key)
			if result != tt.expected {
				t.Errorf("GetString(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Frontmatter
		key    string
		value  interface{}
		verify func(*testing.T, *Frontmatter)
	}{
		{
			name: "set new field",
			setup: func() *Frontmatter {
				return &Frontmatter{Fields: map[string]interface{}{}}
			},
			key:   "model",
			value: "sonnet-4.5",
			verify: func(t *testing.T, f *Frontmatter) {
				if f.Fields["model"] != "sonnet-4.5" {
					t.Errorf("field not set correctly")
				}
			},
		},
		{
			name: "overwrite existing field",
			setup: func() *Frontmatter {
				return &Frontmatter{Fields: map[string]interface{}{"model": "old"}}
			},
			key:   "model",
			value: "new",
			verify: func(t *testing.T, f *Frontmatter) {
				if f.Fields["model"] != "new" {
					t.Errorf("field not overwritten correctly")
				}
			},
		},
		{
			name: "set field with nil fields map",
			setup: func() *Frontmatter {
				return &Frontmatter{Fields: nil}
			},
			key:   "model",
			value: "sonnet-4.5",
			verify: func(t *testing.T, f *Frontmatter) {
				if f.Fields == nil {
					t.Error("Fields map should be initialized")
					return
				}
				if f.Fields["model"] != "sonnet-4.5" {
					t.Errorf("field not set correctly")
				}
			},
		},
		{
			name: "set complex value",
			setup: func() *Frontmatter {
				return &Frontmatter{Fields: map[string]interface{}{}}
			},
			key:   "config",
			value: map[string]interface{}{"nested": "value"},
			verify: func(t *testing.T, f *Frontmatter) {
				config, ok := f.Fields["config"].(map[string]interface{})
				if !ok {
					t.Error("field should be a map")
					return
				}
				if config["nested"] != "value" {
					t.Errorf("nested value not set correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.setup()
			f.SetField(tt.key, tt.value)
			tt.verify(t, f)
		})
	}

	// Test nil frontmatter doesn't panic
	t.Run("nil frontmatter doesn't panic", func(t *testing.T) {
		var f *Frontmatter
		f.SetField("key", "value") // Should not panic
	})
}

func TestRender(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter *Frontmatter
		checkFunc   func(*testing.T, []byte)
	}{
		{
			name: "render with fields and content",
			frontmatter: &Frontmatter{
				Fields:  map[string]interface{}{"model": "sonnet-4.5"},
				Content: "# Content here",
			},
			checkFunc: func(t *testing.T, result []byte) {
				// Parse the result back to verify roundtrip
				parsed, err := Parse(result)
				if err != nil {
					t.Fatalf("failed to parse rendered output: %v", err)
				}
				if parsed.GetString("model") != "sonnet-4.5" {
					t.Errorf("model field not preserved")
				}
				if parsed.Content != "# Content here" {
					t.Errorf("content not preserved: got %q", parsed.Content)
				}
			},
		},
		{
			name: "render with empty fields",
			frontmatter: &Frontmatter{
				Fields:  map[string]interface{}{},
				Content: "# Content",
			},
			checkFunc: func(t *testing.T, result []byte) {
				s := string(result)
				if s != "---\n---\n# Content" {
					t.Errorf("unexpected output: %q", s)
				}
			},
		},
		{
			name:        "render nil frontmatter returns nil",
			frontmatter: nil,
			checkFunc: func(t *testing.T, result []byte) {
				if result != nil {
					t.Errorf("expected nil, got %q", string(result))
				}
			},
		},
		{
			name: "render preserves content exactly",
			frontmatter: &Frontmatter{
				Fields:  map[string]interface{}{"title": "test"},
				Content: "Line 1\n\nLine 3\n---\nLine 5",
			},
			checkFunc: func(t *testing.T, result []byte) {
				parsed, err := Parse(result)
				if err != nil {
					t.Fatalf("failed to parse: %v", err)
				}
				expected := "Line 1\n\nLine 3\n---\nLine 5"
				if parsed.Content != expected {
					t.Errorf("content not preserved: got %q, want %q", parsed.Content, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.frontmatter.Render()
			tt.checkFunc(t, result)
		})
	}
}

func TestSetFieldAndRenderRoundtrip(t *testing.T) {
	// Test that modifying fields and rendering produces valid output
	original := "---\nmodel: claude\ntemperature: 0.5\n---\n# My Document\n\nSome content here."

	fm, err := Parse([]byte(original))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Modify a field
	fm.SetField("model", "sonnet-4.5")
	fm.SetField("new_field", "new_value")

	// Render and parse again
	rendered := fm.Render()
	reparsed, err := Parse(rendered)
	if err != nil {
		t.Fatalf("Parse of rendered content failed: %v", err)
	}

	// Verify modifications persisted
	if reparsed.GetString("model") != "sonnet-4.5" {
		t.Errorf("model not updated: got %q", reparsed.GetString("model"))
	}
	if reparsed.GetString("new_field") != "new_value" {
		t.Errorf("new_field not added: got %q", reparsed.GetString("new_field"))
	}

	// Verify content preserved
	expectedContent := "# My Document\n\nSome content here."
	if reparsed.Content != expectedContent {
		t.Errorf("content not preserved: got %q, want %q", reparsed.Content, expectedContent)
	}
}

func TestContentPreservation(t *testing.T) {
	// Ensure that content body is preserved exactly, including special characters
	testCases := []string{
		"# Simple heading",
		"Code block:\n```go\nfunc main() {}\n```",
		"---\nHorizontal rule above",
		"Special chars: <>&\"'",
		"Unicode: 日本語 中文 한국어",
		"Whitespace preserved:   \n\t\t",
	}

	for _, content := range testCases {
		t.Run(content[:min(20, len(content))], func(t *testing.T) {
			input := "---\ntitle: test\n---\n" + content

			fm, err := Parse([]byte(input))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if fm.Content != content {
				t.Errorf("content not preserved:\ngot:  %q\nwant: %q", fm.Content, content)
			}
		})
	}
}

// deepEqual is a simple helper for comparing interface{} values
func deepEqual(a, b interface{}) bool {
	switch va := a.(type) {
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case int:
		vb, ok := b.(int)
		return ok && va == vb
	case float64:
		vb, ok := b.(float64)
		return ok && va == vb
	case bool:
		vb, ok := b.(bool)
		return ok && va == vb
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
