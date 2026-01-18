package resource

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid simple", "test", false},
		{"valid with hyphen", "test-command", false},
		{"valid alphanumeric", "test123", false},
		{"valid long", "this-is-a-very-long-but-still-valid-name-for-testing", false},
		{"empty", "", true},
		{"too long", "this-is-a-very-long-name-that-exceeds-the-maximum-allowed-length-of-sixty-four-characters", true},
		{"uppercase", "TestCommand", true},
		{"leading hyphen", "-test", true},
		{"trailing hyphen", "test-", true},
		{"consecutive hyphens", "test--command", true},
		{"with underscore", "test_command", true},
		{"with space", "test command", true},
		{"with special char", "test@command", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name      string
		desc      string
		resType   ResourceType
		wantError bool
	}{
		{"valid command desc", "A simple command", Command, false},
		{"valid skill desc", "A simple skill", Skill, false},
		{"empty desc", "", Command, true},
		{"long skill desc", string(make([]byte, 1025)), Skill, true},
		{"long command desc", string(make([]byte, 2000)), Command, false}, // Commands are flexible
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.desc, tt.resType)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateDescription() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestResourceValidate(t *testing.T) {
	tests := []struct {
		name      string
		resource  Resource
		wantError bool
	}{
		{
			name: "valid command",
			resource: Resource{
				Name:        "test-command",
				Type:        Command,
				Description: "A test command",
			},
			wantError: false,
		},
		{
			name: "valid skill",
			resource: Resource{
				Name:        "test-skill",
				Type:        Skill,
				Description: "A test skill",
			},
			wantError: false,
		},
		{
			name: "invalid name",
			resource: Resource{
				Name:        "Invalid-Name",
				Type:        Command,
				Description: "A test command",
			},
			wantError: true,
		},
		{
			name: "invalid type",
			resource: Resource{
				Name:        "test",
				Type:        "invalid",
				Description: "A test",
			},
			wantError: true,
		},
		{
			name: "empty description",
			resource: Resource{
				Name: "test",
				Type: Command,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resource.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Resource.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
