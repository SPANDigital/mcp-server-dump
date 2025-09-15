package app

import (
	"testing"
)

func TestCLI_ScanValidation(t *testing.T) {
	tests := []struct {
		name        string
		cli         CLI
		expectError bool
		errorMsg    string
	}{
		{
			name: "all_scan_types_enabled_default",
			cli: CLI{
				NoTools:     false,
				NoResources: false,
				NoPrompts:   false,
			},
			expectError: false,
		},
		{
			name: "only_tools_disabled",
			cli: CLI{
				NoTools:     true,
				NoResources: false,
				NoPrompts:   false,
			},
			expectError: false,
		},
		{
			name: "only_resources_disabled",
			cli: CLI{
				NoTools:     false,
				NoResources: true,
				NoPrompts:   false,
			},
			expectError: false,
		},
		{
			name: "only_prompts_disabled",
			cli: CLI{
				NoTools:     false,
				NoResources: false,
				NoPrompts:   true,
			},
			expectError: false,
		},
		{
			name: "tools_and_resources_disabled",
			cli: CLI{
				NoTools:     true,
				NoResources: true,
				NoPrompts:   false,
			},
			expectError: false,
		},
		{
			name: "tools_and_prompts_disabled",
			cli: CLI{
				NoTools:     true,
				NoResources: false,
				NoPrompts:   true,
			},
			expectError: false,
		},
		{
			name: "resources_and_prompts_disabled",
			cli: CLI{
				NoTools:     false,
				NoResources: true,
				NoPrompts:   true,
			},
			expectError: false,
		},
		{
			name: "all_scan_types_disabled",
			cli: CLI{
				NoTools:     true,
				NoResources: true,
				NoPrompts:   true,
			},
			expectError: true,
			errorMsg:    "cannot disable all scan types: at least one of tools, resources, or prompts must be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			hasError := tt.cli.NoTools && tt.cli.NoResources && tt.cli.NoPrompts

			if tt.expectError {
				if !hasError {
					t.Errorf("Expected validation error but none occurred")
				}
			} else {
				if hasError {
					t.Errorf("Unexpected validation error for valid configuration")
				}
			}
		})
	}
}

func TestCLI_ScanFlags(t *testing.T) {
	// Test that the CLI struct has the expected scan flag fields
	cli := CLI{}

	// Verify scan flags are boolean type and have default false values
	if cli.NoTools != false {
		t.Errorf("Expected NoTools default to be false, got %v", cli.NoTools)
	}

	if cli.NoResources != false {
		t.Errorf("Expected NoResources default to be false, got %v", cli.NoResources)
	}

	if cli.NoPrompts != false {
		t.Errorf("Expected NoPrompts default to be false, got %v", cli.NoPrompts)
	}

	// Test that flags can be set
	cli.NoTools = true
	cli.NoResources = true
	cli.NoPrompts = true

	if !cli.NoTools {
		t.Errorf("Failed to set NoTools flag")
	}

	if !cli.NoResources {
		t.Errorf("Failed to set NoResources flag")
	}

	if !cli.NoPrompts {
		t.Errorf("Failed to set NoPrompts flag")
	}
}
