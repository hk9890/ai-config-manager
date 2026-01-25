package marketplace

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseMarketplace_AnthropicsFormat tests parsing Anthropics official format
func TestParseMarketplace_AnthropicsFormat(t *testing.T) {
	// Anthropics format: description in metadata
	content := `{
  "name": "anthropic-agent-skills",
  "owner": {
    "name": "Keith Lazuka",
    "email": "klazuka@anthropic.com"
  },
  "metadata": {
    "description": "Anthropic example skills",
    "version": "1.0.0"
  },
  "plugins": [
    {
      "name": "document-skills",
      "description": "Collection of document processing suite",
      "source": "./"
    }
  ]
}`

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "marketplace.json")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config, err := ParseMarketplace(filePath)
	if err != nil {
		t.Fatalf("Failed to parse Anthropics format: %v", err)
	}

	if config.Name != "anthropic-agent-skills" {
		t.Errorf("Name = %v, want anthropic-agent-skills", config.Name)
	}

	// Test GetDescription() method
	desc := config.GetDescription()
	if desc != "Anthropic example skills" {
		t.Errorf("GetDescription() = %v, want 'Anthropic example skills'", desc)
	}

	// Test GetVersion() method
	version := config.GetVersion()
	if version != "1.0.0" {
		t.Errorf("GetVersion() = %v, want 1.0.0", version)
	}

	if len(config.Plugins) != 1 {
		t.Errorf("Plugin count = %d, want 1", len(config.Plugins))
	}
}

// TestParseMarketplace_TraditionalFormat tests parsing traditional format
func TestParseMarketplace_TraditionalFormat(t *testing.T) {
	// Traditional format: description at top level
	content := `{
  "name": "traditional-marketplace",
  "version": "1.0.0",
  "description": "Traditional format marketplace",
  "plugins": [
    {
      "name": "test-plugin",
      "description": "Test plugin",
      "source": "./plugin"
    }
  ]
}`

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "marketplace.json")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config, err := ParseMarketplace(filePath)
	if err != nil {
		t.Fatalf("Failed to parse traditional format: %v", err)
	}

	if config.Name != "traditional-marketplace" {
		t.Errorf("Name = %v, want traditional-marketplace", config.Name)
	}

	// Test GetDescription() method
	desc := config.GetDescription()
	if desc != "Traditional format marketplace" {
		t.Errorf("GetDescription() = %v, want 'Traditional format marketplace'", desc)
	}

	// Test GetVersion() method
	version := config.GetVersion()
	if version != "1.0.0" {
		t.Errorf("GetVersion() = %v, want 1.0.0", version)
	}
}

// TestParseMarketplace_MinimalFormat tests minimal format without description
func TestParseMarketplace_MinimalFormat(t *testing.T) {
	// Minimal format: no description at all
	content := `{
  "name": "minimal-marketplace",
  "plugins": [
    {
      "name": "test-plugin",
      "description": "Test plugin",
      "source": "./plugin"
    }
  ]
}`

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "marketplace.json")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	config, err := ParseMarketplace(filePath)
	if err != nil {
		t.Fatalf("Failed to parse minimal format: %v", err)
	}

	if config.Name != "minimal-marketplace" {
		t.Errorf("Name = %v, want minimal-marketplace", config.Name)
	}

	// Description should be empty but not cause error
	desc := config.GetDescription()
	if desc != "" {
		t.Errorf("GetDescription() = %v, want empty string", desc)
	}
}
