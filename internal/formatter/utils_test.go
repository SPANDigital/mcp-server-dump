package formatter

import (
	"testing"
)

func TestHumanizeKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic underscore cases
		{"single underscore", "user_name", "User Name"},
		{"multiple underscores", "first_last_name", "First Last Name"},
		{"underscore at start", "_private_var", "Private Var"},
		{"underscore at end", "config_value_", "Config Value"},

		// Space cases
		{"single space", "user name", "User Name"},
		{"multiple spaces", "first  last   name", "First Last Name"},
		{"space at start", " private var", "Private Var"},
		{"space at end", "config value ", "Config Value"},

		// Mixed underscore and space cases
		{"mixed underscore and space", "user_name test", "User Name Test"},
		{"complex mixed", "api_key test_value", "Api Key Test Value"},

		// Edge cases
		{"empty string", "", ""},
		{"single character", "a", "A"},
		{"single underscore", "_", ""},
		{"multiple underscores only", "___", ""},
		{"multiple spaces only", "   ", ""},

		// Already formatted cases
		{"already title case", "User Name", "User Name"},
		{"mixed case", "userNAME", "Username"},

		// Special characters (should be preserved after underscores/spaces)
		{"with numbers", "user_id_123", "User Id 123"},
		{"with hyphen", "api-key_value", "Api-Key Value"},
		{"with dots", "file.ext_name", "File.ext Name"},

		// Common API/config key patterns
		{"database url", "database_url", "Database Url"},
		{"api key", "api_key", "Api Key"},
		{"max connections", "max_connections", "Max Connections"},
		{"timeout seconds", "timeout_seconds", "Timeout Seconds"},
		{"enable feature", "enable_feature_flag", "Enable Feature Flag"},

		// Technical terms that should remain readable
		{"ssl cert", "ssl_certificate_path", "Ssl Certificate Path"},
		{"http port", "http_server_port", "Http Server Port"},
		{"jwt token", "jwt_access_token", "Jwt Access Token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := humanizeKey(tt.input)
			if result != tt.expected {
				t.Errorf("humanizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHumanizeKeyConsistency(t *testing.T) {
	// Test that the function is consistent - same input should always produce same output
	testCases := []string{
		"user_name",
		"api_key_value",
		"database_connection_string",
		"enable_feature_flag",
	}

	for _, input := range testCases {
		first := humanizeKey(input)
		second := humanizeKey(input)
		if first != second {
			t.Errorf("humanizeKey(%q) is not consistent: first=%q, second=%q", input, first, second)
		}
	}
}

func TestHumanizeKeyWithRealWorldExamples(t *testing.T) {
	// Test with examples that might appear in actual MCP context
	realWorldTests := []struct {
		input    string
		expected string
	}{
		{"usage_instructions", "Usage Instructions"},
		{"security_requirements", "Security Requirements"},
		{"example_requests", "Example Requests"},
		{"rate_limit_info", "Rate Limit Info"},
		{"authentication_method", "Authentication Method"},
		{"response_format", "Response Format"},
		{"error_handling", "Error Handling"},
		{"best_practices", "Best Practices"},
		{"troubleshooting_guide", "Troubleshooting Guide"},
		{"version_compatibility", "Version Compatibility"},
	}

	for _, tt := range realWorldTests {
		t.Run(tt.input, func(t *testing.T) {
			result := humanizeKey(tt.input)
			if result != tt.expected {
				t.Errorf("humanizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}