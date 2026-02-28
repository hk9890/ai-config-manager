package output

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

// EncodeJSON encodes data as JSON to the writer
func EncodeJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// EncodeYAML encodes data as YAML to the writer
func EncodeYAML(w io.Writer, data interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(data)
}
