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

	// CamelCase boundary detection regexes
	// Pattern 1: lowercase/digits followed by uppercase (camelCase)
	camelCaseRegex = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	// Pattern 2: uppercase followed by uppercase then lowercase (HTTPServer -> HTTP Server)
	acronymRegex = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)

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

// buildInitialismsMap creates a merged map of built-in and custom initialisms
// Custom initialisms are converted to uppercase for consistent lookup
func buildInitialismsMap(customInitialisms []string) map[string]bool {
	// Start with built-in initialisms
	merged := make(map[string]bool, len(commonInitialisms)+len(customInitialisms))
	for initialism, value := range commonInitialisms {
		merged[initialism] = value
	}

	// Add custom initialisms (convert to uppercase for consistency)
	for _, custom := range customInitialisms {
		if custom != "" {
			upperCustom := strings.ToUpper(strings.TrimSpace(custom))
			if upperCustom != "" {
				merged[upperCustom] = true
			}
		}
	}

	return merged
}

// splitCamelCase splits a string on camelCase boundaries
// Examples: "XMLHttpRequest" → ["XML", "Http", "Request"], "getElementById" → ["get", "Element", "By", "Id"]
func splitCamelCase(word string) []string {
	if word == "" {
		return nil
	}

	// First, handle acronym boundaries: "XMLHttpRequest" → "XML HttpRequest"
	spaced := acronymRegex.ReplaceAllString(word, "$1 $2")

	// Then handle regular camelCase boundaries: "XML HttpRequest" → "XML Http Request"
	spaced = camelCaseRegex.ReplaceAllString(spaced, "$1 $2")

	// Split on the inserted spaces
	return strings.Fields(spaced)
}

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
	return humanizeKeyWithCustomInitialisms(key, nil)
}

// humanizeKeyWithCustomInitialisms converts context keys with underscores or spaces to human-readable titles,
// supporting custom initialisms in addition to built-in ones.
// It also handles camelCase boundaries properly.
// Examples:
//   - "user_name" → "User Name"
//   - "api_key" → "API Key"
//   - "XMLHttpRequest" → "XML HTTP Request"
//   - "getElementById" → "Get Element By ID"
//   - With custom initialisms ["CORP"]: "corp_api_key" → "CORP API Key"
func humanizeKeyWithCustomInitialisms(key string, customInitialisms []string) string {
	if key == "" {
		return ""
	}

	// Build the initialisms map (built-in + custom)
	initialisms := commonInitialisms
	if len(customInitialisms) > 0 {
		initialisms = buildInitialismsMap(customInitialisms)
	}

	// Phase 1: Replace underscores and multiple spaces with single spaces
	spaced := spacingRegex.ReplaceAllString(key, " ")

	// Phase 2: Split into initial words
	words := strings.Fields(spaced)
	if len(words) == 0 {
		return ""
	}

	// Phase 3: Further split each word on camelCase boundaries, but check for custom initialisms first
	var allWords []string
	for _, word := range words {
		// Check if the word (case-insensitive) is a custom initialism before splitting
		upperWord := strings.ToUpper(word)
		if initialisms[upperWord] {
			allWords = append(allWords, word)
		} else {
			camelWords := splitCamelCase(word)
			allWords = append(allWords, camelWords...)
		}
	}

	// Phase 4: Process each word for initialisms and title case
	for i, word := range allWords {
		upperWord := strings.ToUpper(word)

		// Check if this word is a known initialism/acronym
		if initialisms[upperWord] {
			// Acronyms should always be uppercase
			allWords[i] = upperWord
		} else {
			// Apply title case to non-acronym words
			allWords[i] = titleCaser.String(strings.ToLower(word))
		}
	}

	return strings.Join(allWords, " ")
}
