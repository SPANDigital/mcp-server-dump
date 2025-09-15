package formatter

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

const (
	boolTrue  = "true"
	boolFalse = "false"
)

// extractJSONNumber extracts a complete number from a JSON string starting at index i
func extractJSONNumber(jsonStr string, i int) (numberStr string, nextIndex int) {
	start := i

	// Handle negative sign
	if i < len(jsonStr) && jsonStr[i] == '-' {
		i++
	}

	// Handle integer part
	for i < len(jsonStr) && jsonStr[i] >= '0' && jsonStr[i] <= '9' {
		i++
	}

	// Handle decimal part
	if i < len(jsonStr) && jsonStr[i] == '.' {
		i++
		for i < len(jsonStr) && jsonStr[i] >= '0' && jsonStr[i] <= '9' {
			i++
		}
	}

	// Handle scientific notation
	if i < len(jsonStr) && (jsonStr[i] == 'e' || jsonStr[i] == 'E') {
		i++
		if i < len(jsonStr) && (jsonStr[i] == '+' || jsonStr[i] == '-') {
			i++
		}
		for i < len(jsonStr) && jsonStr[i] >= '0' && jsonStr[i] <= '9' {
			i++
		}
	}

	return jsonStr[start:i], i
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
