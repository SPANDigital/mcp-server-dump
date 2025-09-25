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
		{"complex mixed", "api_key test_value", "API Key Test Value"},

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
		{"with numbers", "user_id_123", "User ID 123"},
		{"with hyphen", "api-key_value", "Api-Key Value"},
		{"with dots", "file.ext_name", "File.ext Name"},

		// Common API/config key patterns (with proper acronym handling)
		{"database url", "database_url", "Database URL"},
		{"api key", "api_key", "API Key"},
		{"max connections", "max_connections", "Max Connections"},
		{"timeout seconds", "timeout_seconds", "Timeout Seconds"},
		{"enable feature", "enable_feature_flag", "Enable Feature Flag"},

		// Technical terms with proper acronym handling
		{"ssl cert", "ssl_certificate_path", "SSL Certificate Path"},
		{"http port", "http_server_port", "HTTP Server Port"},
		{"jwt token", "jwt_access_token", "JWT Access Token"},

		// Additional acronym test cases (single words should be uppercase)
		{"api", "api", "API"},
		{"http", "http", "HTTP"},
		{"jwt", "jwt", "JWT"},
		{"ssl", "ssl", "SSL"},
		{"url", "url", "URL"},
		{"json", "json", "JSON"},
		{"xml", "xml", "XML"},
		{"sql", "sql", "SQL"},
		{"tcp", "tcp", "TCP"},
		{"udp", "udp", "UDP"},
		{"html", "html", "HTML"},
		{"css", "css", "CSS"},

		// Multi-word with acronyms (all acronyms should be uppercase)
		{"user api key", "user_api_key", "User API Key"},
		{"server http port", "server_http_port", "Server HTTP Port"},
		{"auth jwt token", "auth_jwt_token", "Auth JWT Token"},
		{"database sql query", "database_sql_query", "Database SQL Query"},
		{"web html content", "web_html_content", "Web HTML Content"},
		{"api json response", "api_json_response", "API JSON Response"},
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
