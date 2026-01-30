package output

import (
	"fmt"
	"os"
	"strings"
)

// KeyValueData represents key-value pair output
type KeyValueData struct {
	Title string     `json:"title,omitempty" yaml:"title,omitempty"`
	Pairs []KeyValue `json:"pairs" yaml:"pairs"`
}

// KeyValue represents a single key-value pair
type KeyValue struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

// KeyValueBuilder provides a fluent API for building key-value output
type KeyValueBuilder struct {
	data *KeyValueData
}

// NewKeyValue creates a new KeyValueBuilder with the given title
func NewKeyValue(title string) *KeyValueBuilder {
	return &KeyValueBuilder{
		data: &KeyValueData{
			Title: title,
			Pairs: []KeyValue{},
		},
	}
}

// Add adds a key-value pair
func (kvb *KeyValueBuilder) Add(key, value string) *KeyValueBuilder {
	kvb.data.Pairs = append(kvb.data.Pairs, KeyValue{Key: key, Value: value})
	return kvb
}

// AddSection adds a blank line for visual grouping
func (kvb *KeyValueBuilder) AddSection() *KeyValueBuilder {
	kvb.data.Pairs = append(kvb.data.Pairs, KeyValue{Key: "", Value: ""})
	return kvb
}

// Format outputs the key-value data in the specified format
func (kvb *KeyValueBuilder) Format(format Format) error {
	return FormatOutput(kvb.data, format)
}

// formatKeyValueData renders KeyValueData in the requested format
func formatKeyValueData(data *KeyValueData, format Format) error {
	switch format {
	case Table:
		return renderKeyValue(data)
	case JSON:
		return EncodeJSON(os.Stdout, data)
	case YAML:
		return EncodeYAML(os.Stdout, data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderKeyValue renders KeyValueData as human-readable text
func renderKeyValue(data *KeyValueData) error {
	if data.Title != "" {
		fmt.Println(data.Title)
		fmt.Println(strings.Repeat("=", len(data.Title)))
		fmt.Println()
	}

	for _, kv := range data.Pairs {
		if kv.Key == "" {
			fmt.Println()
			continue
		}
		fmt.Printf("%s: %s\n", kv.Key, kv.Value)
	}

	return nil
}
