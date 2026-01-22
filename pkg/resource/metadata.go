package resource

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents parsed YAML frontmatter from a markdown file
type Frontmatter map[string]interface{}

// ParseFrontmatter parses YAML frontmatter from a markdown file
// Frontmatter must be delimited by "---" at the beginning and end
func ParseFrontmatter(filePath string) (Frontmatter, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseFrontmatterReader(file)
}

// ParseFrontmatterReader parses YAML frontmatter from a reader
func ParseFrontmatterReader(r io.Reader) (Frontmatter, string, error) {
	scanner := bufio.NewScanner(r)

	// Check for opening delimiter
	if !scanner.Scan() {
		return nil, "", fmt.Errorf("empty file")
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		return nil, "", fmt.Errorf("no frontmatter found (must start with '---')")
	}

	// Read frontmatter content until closing delimiter
	var frontmatterLines []string
	foundClosing := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClosing = true
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	if !foundClosing {
		return nil, "", fmt.Errorf("no closing frontmatter delimiter found")
	}

	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	// Parse YAML
	frontmatterYAML := strings.Join(frontmatterLines, "\n")
	var frontmatter Frontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	// Read remaining content (the markdown body)
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	content := strings.Join(contentLines, "\n")

	return frontmatter, content, nil
}

// GetString safely extracts a string value from frontmatter
func (f Frontmatter) GetString(key string) string {
	if val, ok := f[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetStringSlice safely extracts a []string from frontmatter
func (f Frontmatter) GetStringSlice(key string) []string {
	if val, ok := f[key]; ok {
		// Try []interface{} (from YAML unmarshaling)
		if slice, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		// Try []string directly
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return nil
}

// GetMap safely extracts a map[string]string from frontmatter
func (f Frontmatter) GetMap(key string) map[string]string {
	result := make(map[string]string)

	if val, ok := f[key]; ok {
		// Try direct map[string]interface{} first (for manually constructed maps)
		if m, ok := val.(map[string]interface{}); ok {
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
		} else if m, ok := val.(Frontmatter); ok {
			// Also handle nested Frontmatter type (from YAML unmarshaling)
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
		}
	}

	return result
}

// WriteFrontmatter writes frontmatter and content to a file
func WriteFrontmatter(filePath string, frontmatter Frontmatter, content string) error {
	var buf bytes.Buffer

	// Write opening delimiter
	buf.WriteString("---\n")

	// Write YAML
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}
	buf.Write(yamlBytes)

	// Write closing delimiter
	buf.WriteString("---\n")

	// Write content
	if content != "" {
		buf.WriteString("\n")
		buf.WriteString(content)
	}

	// Write to file
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
