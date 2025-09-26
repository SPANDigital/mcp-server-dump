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
		{"mixed case", "userNAME", "User Name"},

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
		{"cdn", "cdn", "CDN"},
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

		// CamelCase boundary test cases (new functionality)
		{"XMLHttpRequest", "XMLHttpRequest", "XML HTTP Request"},
		{"getElementById", "getElementById", "Get Element By ID"},
		{"innerHTML", "innerHTML", "Inner HTML"},
		{"XMLParser", "XMLParser", "XML Parser"},
		{"httpRequest", "httpRequest", "HTTP Request"},
		{"parseJSON", "parseJSON", "Parse JSON"},
		{"createXMLNode", "createXMLNode", "Create XML Node"},

		// Multi-word with acronyms (all acronyms should be uppercase)
		{"user api key", "user_api_key", "User API Key"},
		{"cdn endpoint url", "cdn_endpoint_url", "CDN Endpoint URL"},
		{"server http port", "server_http_port", "Server HTTP Port"},
		{"auth jwt token", "auth_jwt_token", "Auth JWT Token"},
		{"database sql query", "database_sql_query", "Database SQL Query"},
		{"web html content", "web_html_content", "Web HTML Content"},
		{"api json response", "api_json_response", "API JSON Response"},

		// Mixed underscore and camelCase test cases (complex scenarios)
		{"api_XMLHttpRequest", "api_XMLHttpRequest", "API XML HTTP Request"},
		{"user_getElementById", "user_getElementById", "User Get Element By ID"},
		{"httpRequest_data", "httpRequest_data", "HTTP Request Data"},
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

func TestHumanizeKeyWithCustomInitialisms(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		customInitialisms []string
		expected          string
	}{
		// Basic custom initialism tests
		{"custom initialism CORP", "corp_api_key", []string{"CORP"}, "CORP API Key"},
		{"custom initialism ACME", "acme_system", []string{"ACME"}, "ACME System"},
		{"multiple custom initialisms", "corp_acme_api", []string{"CORP", "ACME"}, "CORP ACME API"},

		// Case insensitive custom initialisms (should work with lowercase input)
		{"case insensitive custom", "myorg_api", []string{"myorg"}, "MYORG API"},
		{"mixed case custom", "MyOrg_system", []string{"MyOrg"}, "MYORG System"},

		// Custom initialisms with camelCase
		{"custom with camelCase", "corpRequest", []string{"CORP"}, "CORP Request"},
		{"camelCase with custom initialism", "parseCorpData", []string{"CORP"}, "Parse CORP Data"},

		// Duplicates should be handled gracefully
		{"duplicate custom initialisms", "corp_api", []string{"CORP", "corp", "CORP"}, "CORP API"},

		// Empty custom initialisms should fall back to built-in
		{"empty custom list", "api_key", []string{}, "API Key"},
		{"nil custom list", "http_server", nil, "HTTP Server"},

		// Custom initialisms should not override built-in ones
		{"custom should not override builtin", "api_corp_system", []string{"CORP"}, "API CORP System"},

		// Whitespace in custom initialisms should be handled
		{"custom with whitespace", "corp_api", []string{" CORP ", "  API  "}, "CORP API"},

		// Complex mixed scenarios
		{"complex mixed scenario", "corp_XMLHttpRequest_api", []string{"CORP"}, "CORP XML HTTP Request API"},

		// Edge case: CamelCase with custom initialism interaction
		{"corpAPIHandler with custom CORP", "corpAPIHandler", []string{"CORP"}, "CORP API Handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := humanizeKeyWithCustomInitialisms(tt.input, tt.customInitialisms)
			if result != tt.expected {
				t.Errorf("humanizeKeyWithCustomInitialisms(%q, %v) = %q, want %q",
					tt.input, tt.customInitialisms, result, tt.expected)
			}
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", nil},
		{"single word", "hello", []string{"hello"}},
		{"simple camelCase", "helloWorld", []string{"hello", "World"}},
		{"XMLHttpRequest", "XMLHttpRequest", []string{"XML", "Http", "Request"}},
		{"getElementById", "getElementById", []string{"get", "Element", "By", "Id"}},
		{"innerHTML", "innerHTML", []string{"inner", "HTML"}},
		{"all lowercase", "lowercase", []string{"lowercase"}},
		{"all uppercase", "UPPERCASE", []string{"UPPERCASE"}},
		{"numbers", "test123", []string{"test123"}},
		{"mixed with numbers", "test123ABC", []string{"test123", "ABC"}},
		{"consecutive caps HTTPAPI", "HTTPAPI", []string{"HTTPAPI"}},
		{"consecutive caps mixed HTTPSServer", "HTTPSServer", []string{"HTTPS", "Server"}},
		{"API prefix APIHandler", "APIHandler", []string{"API", "Handler"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitCamelCase(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitCamelCase(%q) = %v, want %v (length mismatch)",
					tt.input, result, tt.expected)
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("splitCamelCase(%q) = %v, want %v (at index %d)",
						tt.input, result, tt.expected, i)
					break
				}
			}
		})
	}
}

func TestBuildInitialismsMap(t *testing.T) {
	tests := []struct {
		name              string
		customInitialisms []string
		expectedKeys      []string
	}{
		{
			name:              "empty custom list",
			customInitialisms: []string{},
			expectedKeys:      []string{"API", "HTTP", "JSON"}, // Should contain built-in initialisms
		},
		{
			name:              "nil custom list",
			customInitialisms: nil,
			expectedKeys:      []string{"API", "HTTP", "JSON"}, // Should contain built-in initialisms
		},
		{
			name:              "custom initialisms added",
			customInitialisms: []string{"CORP", "ACME"},
			expectedKeys:      []string{"API", "HTTP", "JSON", "CORP", "ACME"}, // Should contain both
		},
		{
			name:              "lowercase custom initialisms converted",
			customInitialisms: []string{"corp", "acme"},
			expectedKeys:      []string{"API", "HTTP", "JSON", "CORP", "ACME"}, // Should be uppercase
		},
		{
			name:              "whitespace handled",
			customInitialisms: []string{" CORP ", "  ACME  "},
			expectedKeys:      []string{"API", "HTTP", "JSON", "CORP", "ACME"}, // Should be trimmed
		},
		{
			name:              "empty strings ignored",
			customInitialisms: []string{"CORP", "", " ", "ACME"},
			expectedKeys:      []string{"API", "HTTP", "JSON", "CORP", "ACME"}, // Empty strings ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildInitialismsMap(tt.customInitialisms)

			// Check that all expected keys are present
			for _, expectedKey := range tt.expectedKeys {
				if !result[expectedKey] {
					t.Errorf("buildInitialismsMap(%v) missing expected key %q",
						tt.customInitialisms, expectedKey)
				}
			}

			// Check that result contains built-in initialisms count + unique custom count
			expectedMinSize := len(commonInitialisms)
			if len(result) < expectedMinSize {
				t.Errorf("buildInitialismsMap(%v) result size %d is less than built-in size %d",
					tt.customInitialisms, len(result), expectedMinSize)
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
