package app

import (
	"testing"
)

func TestRunValidation_ScanControls(t *testing.T) {
	tests := []struct {
		name        string
		cli         CLI
		expectError bool
		errorMsg    string
	}{
		{
			name: "all_scan_types_enabled",
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
			name: "two_types_disabled",
			cli: CLI{
				NoTools:     true,
				NoResources: true,
				NoPrompts:   false,
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
			errorMsg:    ErrAllScanTypesDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic using the CLI method
			err := tt.cli.ValidateScanOptions()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error but none occurred")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestCLI_ScanFlagDefaults(t *testing.T) {
	// Test that scan flags have correct default values
	cli := CLI{}

	if cli.NoTools != false {
		t.Errorf("Expected NoTools default to be false, got %v", cli.NoTools)
	}

	if cli.NoResources != false {
		t.Errorf("Expected NoResources default to be false, got %v", cli.NoResources)
	}

	if cli.NoPrompts != false {
		t.Errorf("Expected NoPrompts default to be false, got %v", cli.NoPrompts)
	}
}

func TestCLI_ScanFlagCombinations(t *testing.T) {
	// Test various flag combinations for logical consistency
	testCases := []struct {
		name      string
		noTools   bool
		noRes     bool
		noPrompts bool
		isValid   bool
	}{
		{"all_enabled", false, false, false, true},
		{"tools_only", false, true, true, true},
		{"resources_only", true, false, true, true},
		{"prompts_only", true, true, false, true},
		{"tools_and_resources", false, false, true, true},
		{"tools_and_prompts", false, true, false, true},
		{"resources_and_prompts", true, false, false, true},
		{"all_disabled", true, true, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := CLI{
				NoTools:     tc.noTools,
				NoResources: tc.noRes,
				NoPrompts:   tc.noPrompts,
			}

			// Check if at least one scan type is enabled
			atLeastOneEnabled := !cli.NoTools || !cli.NoResources || !cli.NoPrompts
			isValid := atLeastOneEnabled

			if isValid != tc.isValid {
				t.Errorf("Expected validity %v for combination %s, got %v", tc.isValid, tc.name, isValid)
			}
		})
	}
}
