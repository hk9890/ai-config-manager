package frontmatter

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

const delimiter = "---"

// Frontmatter represents parsed YAML frontmatter from a markdown file
type Frontmatter struct {
	Fields  map[string]interface{} // Parsed YAML fields
	Content string                 // Markdown content after frontmatter
	Raw     string                 // Original frontmatter YAML string
}

// HasFrontmatter checks if content has YAML frontmatter.
// Frontmatter must start with "---" on the first line.
func HasFrontmatter(content []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeft(content, " \t"), []byte(delimiter+"\n")) ||
		bytes.HasPrefix(bytes.TrimLeft(content, " \t"), []byte(delimiter+"\r\n"))
}

// Parse extracts frontmatter from markdown content.
// Returns nil Frontmatter (not error) if no frontmatter present.
// Frontmatter must be delimited by "---" at the start and end.
func Parse(content []byte) (*Frontmatter, error) {
	if !HasFrontmatter(content) {
		return nil, nil
	}

	// Trim any leading whitespace
	trimmed := bytes.TrimLeft(content, " \t")

	// Find the opening delimiter
	start := bytes.Index(trimmed, []byte(delimiter))
	if start == -1 {
		return nil, nil
	}

	// Move past the opening delimiter and newline
	afterStart := start + len(delimiter)
	if afterStart >= len(trimmed) {
		return nil, nil
	}

	// Skip the newline after opening delimiter
	if trimmed[afterStart] == '\r' {
		afterStart++
	}
	if afterStart < len(trimmed) && trimmed[afterStart] == '\n' {
		afterStart++
	}

	// Find the closing delimiter
	rest := trimmed[afterStart:]
	end := findClosingDelimiter(rest)
	if end == -1 {
		return nil, nil
	}

	// Extract the raw YAML content between delimiters
	rawYAML := rest[:end]

	// Calculate where content starts (after closing delimiter and newline)
	contentStart := afterStart + end + len(delimiter)
	if contentStart < len(trimmed) {
		if trimmed[contentStart] == '\r' {
			contentStart++
		}
		if contentStart < len(trimmed) && trimmed[contentStart] == '\n' {
			contentStart++
		}
	}

	// Extract the markdown content
	var markdownContent string
	if contentStart < len(trimmed) {
		markdownContent = string(trimmed[contentStart:])
	}

	// Parse the YAML
	fields := make(map[string]interface{})
	if len(bytes.TrimSpace(rawYAML)) > 0 {
		if err := yaml.Unmarshal(rawYAML, &fields); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
		}
	}

	return &Frontmatter{
		Fields:  fields,
		Content: markdownContent,
		Raw:     string(rawYAML),
	}, nil
}

// findClosingDelimiter finds the closing "---" delimiter.
// It must be at the start of a line.
func findClosingDelimiter(content []byte) int {
	// Check if content starts with delimiter (empty frontmatter case)
	if bytes.HasPrefix(content, []byte(delimiter+"\n")) ||
		bytes.HasPrefix(content, []byte(delimiter+"\r\n")) ||
		bytes.Equal(content, []byte(delimiter)) {
		return 0
	}

	// Look for delimiter at start of subsequent lines
	search := content
	offset := 0
	for {
		// Find next newline
		nlIdx := bytes.IndexByte(search, '\n')
		if nlIdx == -1 {
			return -1
		}

		// Move past the newline
		lineStart := nlIdx + 1
		if lineStart >= len(search) {
			return -1
		}

		// Check if next line starts with delimiter
		remaining := search[lineStart:]
		if bytes.HasPrefix(remaining, []byte(delimiter+"\n")) ||
			bytes.HasPrefix(remaining, []byte(delimiter+"\r\n")) ||
			bytes.Equal(remaining, []byte(delimiter)) {
			return offset + lineStart
		}

		// Move search forward
		offset += lineStart
		search = remaining
	}
}

// GetString returns a string field value from the frontmatter.
// Returns empty string if the key is not found or the value is not a string.
func (f *Frontmatter) GetString(key string) string {
	if f == nil || f.Fields == nil {
		return ""
	}

	val, ok := f.Fields[key]
	if !ok {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// SetField sets a field value in the frontmatter.
// The value can be any type that is valid YAML.
func (f *Frontmatter) SetField(key string, value interface{}) {
	if f == nil {
		return
	}

	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	f.Fields[key] = value
}

// Render produces the full markdown with updated frontmatter.
// Returns the combined frontmatter and content as bytes.
func (f *Frontmatter) Render() []byte {
	if f == nil {
		return nil
	}

	var buf bytes.Buffer

	// Write opening delimiter
	buf.WriteString(delimiter)
	buf.WriteByte('\n')

	// Write YAML fields if any
	if len(f.Fields) > 0 {
		yamlBytes, err := yaml.Marshal(f.Fields)
		if err == nil {
			buf.Write(yamlBytes)
		}
	}

	// Write closing delimiter
	buf.WriteString(delimiter)
	buf.WriteByte('\n')

	// Write content
	buf.WriteString(f.Content)

	return buf.Bytes()
}
