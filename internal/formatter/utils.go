package formatter

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// anchorName converts a string to a URL-safe anchor name
func anchorName(s string) string {
	// Convert to lowercase and replace spaces with hyphens
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

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
			} else if value == "true" || value == "false" {
				// Boolean values
				custom[key] = value == "true"
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
