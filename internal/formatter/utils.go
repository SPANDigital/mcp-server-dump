package formatter

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	boolTrue  = "true"
	boolFalse = "false"
)

// extractJSONNumber extracts a complete number from a JSON string starting at index i
func extractJSONNumber(jsonStr string, i int) (numberStr string, nextIndex int) {
	start := i
	i = parseNumberSign(jsonStr, i)
	i = parseIntegerPart(jsonStr, i)
	i = parseDecimalPart(jsonStr, i)
	i = parseScientificNotation(jsonStr, i)
	return jsonStr[start:i], i
}

// parseNumberSign handles the optional negative sign at the beginning of a number
func parseNumberSign(jsonStr string, i int) int {
	if i < len(jsonStr) && jsonStr[i] == '-' {
		return i + 1
	}
	return i
}

// parseIntegerPart parses the integer portion of a JSON number
func parseIntegerPart(jsonStr string, i int) int {
	for i < len(jsonStr) && jsonStr[i] >= '0' && jsonStr[i] <= '9' {
		i++
	}
	return i
}

// parseDecimalPart parses the optional decimal portion of a JSON number
func parseDecimalPart(jsonStr string, i int) int {
	if i < len(jsonStr) && jsonStr[i] == '.' {
		i++                              // Skip the decimal point
		i = parseIntegerPart(jsonStr, i) // Parse decimal digits after decimal point
	}
	return i
}

// parseScientificNotation parses the optional scientific notation portion of a JSON number.
// Examples: "1.23e10", "5E-3", "2.5e+7" - handles 'e'/'E' followed by optional sign and digits
func parseScientificNotation(jsonStr string, i int) int {
	if i < len(jsonStr) && (jsonStr[i] == 'e' || jsonStr[i] == 'E') {
		i++ // Skip the 'e' or 'E'
		i = parseExponentSign(jsonStr, i)
		i = parseIntegerPart(jsonStr, i) // Reuse integer parsing for exponent digits
	}
	return i
}

// parseExponentSign handles the optional sign in scientific notation exponent
func parseExponentSign(jsonStr string, i int) int {
	if i < len(jsonStr) && (jsonStr[i] == '+' || jsonStr[i] == '-') {
		return i + 1
	}
	return i
}

// extractJSONQuotedString extracts a complete quoted string from JSON starting at index i
func extractJSONQuotedString(jsonStr string, i int) (quotedStr string, nextIndex int) {
	if i >= len(jsonStr) || jsonStr[i] != '"' {
		return "", i
	}

	start := i
	i++ // Skip opening quote

	for i < len(jsonStr) {
		if jsonStr[i] == '"' {
			// Check if it's escaped
			if i > 0 && jsonStr[i-1] == '\\' {
				i++
				continue
			}
			i++ // Include closing quote
			break
		}
		i++
	}

	return jsonStr[start:i], i
}

// anchorName converts a string to a URL-safe anchor name compatible with Goldmark's AutoHeadingID
func anchorName(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens (matching Goldmark behavior)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any characters that aren't letters, numbers, or hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// jsonIndent formats a value as indented JSON
func jsonIndent(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	return string(b), err
}

// ParseCustomFields parses custom frontmatter fields from key:value strings
func ParseCustomFields(fields []string) map[string]any {
	custom := make(map[string]any)
	for _, field := range fields {
		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Try to convert to appropriate type
			if strings.Contains(value, ",") {
				// Comma-separated values become arrays
				items := strings.Split(value, ",")
				for i, item := range items {
					items[i] = strings.TrimSpace(item)
				}
				custom[key] = items
			} else if value == boolTrue || value == boolFalse {
				// Boolean values
				custom[key] = value == boolTrue
			} else if intVal, err := strconv.Atoi(value); err == nil {
				// Integer values
				custom[key] = intVal
			} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				// Float values
				custom[key] = floatVal
			} else {
				// String values
				custom[key] = value
			}
		}
	}
	return custom
}

// formatBool formats a boolean as a checkmark or X
func formatBool(b bool) string {
	if b {
		return "✅ Supported"
	}
	return "❌ Not supported"
}

// isAlphaNumeric reports whether the character is alphanumeric or underscore.
// Used for word boundary checking in JSON parsing to handle identifiers like "true_value".
func isAlphaNumeric(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_'
}

// Package-level variables for performance optimization
var (
	// Pre-compiled regex for performance
	spacingRegex = regexp.MustCompile(`[_\s]+`)

	// English title case converter (reused across calls)
	titleCaser = cases.Title(language.English)

	// Common technical initialisms that should be uppercase (based on Go's golint)
	// Only add entries that are highly unlikely to be non-initialisms
	commonInitialisms = map[string]bool{
		"ACL":   true,
		"API":   true,
		"ASCII": true,
		"CDN":   true, // Addition: Content Delivery Network
		"CPU":   true,
		"CSS":   true,
		"DNS":   true,
		"EOF":   true,
		"GUID":  true,
		"HTML":  true,
		"HTTP":  true,
		"HTTPS": true,
		"ID":    true,
		"IP":    true,
		"JSON":  true,
		"JWT":   true, // Addition: JSON Web Token
		"LHS":   true,
		"QPS":   true,
		"RAM":   true,
		"RHS":   true,
		"RPC":   true,
		"SLA":   true,
		"SMTP":  true,
		"SQL":   true,
		"SSH":   true,
		"SSL":   true, // Addition: Secure Sockets Layer
		"TCP":   true,
		"TLS":   true,
		"TTL":   true,
		"UDP":   true,
		"UI":    true,
		"UID":   true,
		"URI":   true,
		"URL":   true,
		"UTF8":  true,
		"UUID":  true,
		"VM":    true,
		"XML":   true,
		"XMPP":  true,
		"XSRF":  true,
		"XSS":   true,
	}
)

// humanizeKey converts context keys with underscores or spaces to human-readable titles.
// It handles common technical acronyms properly and uses performance optimizations.
// Examples:
//   - "user_name" → "User Name"
//   - "api_key" → "API Key"
//   - "http_server_port" → "HTTP Server Port"
//   - "jwt_access_token" → "JWT Access Token"
//   - "database_url" → "Database URL"
//   - "ssl" → "SSL"
func humanizeKey(key string) string {
	if key == "" {
		return ""
	}

	// Replace underscores and multiple spaces with single spaces using pre-compiled regex
	spaced := spacingRegex.ReplaceAllString(key, " ")

	// Split into words for acronym processing
	words := strings.Fields(spaced)
	if len(words) == 0 {
		return ""
	}

	// Process each word
	for i, word := range words {
		upperWord := strings.ToUpper(word)

		// Check if this word is a known initialism/acronym
		if commonInitialisms[upperWord] {
			// Acronyms should always be uppercase
			words[i] = upperWord
		} else {
			// Apply title case to non-acronym words
			words[i] = titleCaser.String(strings.ToLower(word))
		}
	}

	return strings.Join(words, " ")
}
