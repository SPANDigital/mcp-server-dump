package formatter

import (
	"testing"
)

func TestIsWhitespace(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{
		{' ', true},
		{'\t', true},
		{'\n', true},
		{'\r', true},
		{'a', false},
		{'1', false},
		{'{', false},
		{0, false},
	}

	for _, test := range tests {
		result := isWhitespace(test.char)
		if result != test.expected {
			t.Errorf("isWhitespace(%q) = %v, expected %v", test.char, result, test.expected)
		}
	}
}

func TestIsPunctuation(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{
		{'{', true},
		{'}', true},
		{'[', true},
		{']', true},
		{',', true},
		{':', true},
		{' ', false},
		{'a', false},
		{'1', false},
		{'"', false},
	}

	for _, test := range tests {
		result := isPunctuation(test.char)
		if result != test.expected {
			t.Errorf("isPunctuation(%q) = %v, expected %v", test.char, result, test.expected)
		}
	}
}

func TestIsNumberStart(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{
		{'0', true},
		{'1', true},
		{'9', true},
		{'-', true},
		{'.', true},
		{'a', false},
		{' ', false},
		{'{', false},
		{'"', false},
	}

	for _, test := range tests {
		result := isNumberStart(test.char)
		if result != test.expected {
			t.Errorf("isNumberStart(%q) = %v, expected %v", test.char, result, test.expected)
		}
	}
}

func TestIsBooleanOrNull(t *testing.T) {
	tests := []struct {
		jsonStr  string
		pos      int
		expected bool
	}{
		{"true", 0, true},
		{"false", 0, true},
		{"null", 0, true},
		{"  true", 2, true},
		{"  false", 2, true},
		{"  null", 2, true},
		{"truthy", 0, false},     // not a valid JSON boolean
		{"falsehood", 0, false},  // not a valid JSON boolean
		{"nullable", 0, false},   // not a valid JSON boolean
		{"true_value", 0, false}, // underscore makes it not a boolean
		{"false_flag", 0, false}, // underscore makes it not a boolean
		{"null_check", 0, false}, // underscore makes it not a boolean
		{"string", 0, false},
		{"123", 0, false},
		{"", 0, false},
	}

	for _, test := range tests {
		result := isBooleanOrNull(test.jsonStr, test.pos)
		if result != test.expected {
			t.Errorf("isBooleanOrNull(%q, %d) = %v, expected %v", test.jsonStr, test.pos, result, test.expected)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	// Test empty strings
	t.Run("EmptyStrings", func(t *testing.T) {
		if isWhitespace(0) {
			t.Error("isWhitespace(0) should be false")
		}
		if isPunctuation(0) {
			t.Error("isPunctuation(0) should be false")
		}
		if isNumberStart(0) {
			t.Error("isNumberStart(0) should be false")
		}
		if isBooleanOrNull("", 0) {
			t.Error("isBooleanOrNull('', 0) should be false")
		}
	})

	// Test Unicode characters (should not match JSON tokens)
	t.Run("UnicodeCharacters", func(t *testing.T) {
		unicodeChars := []byte{0xC2, 0xA0, 0xE2, 0x80, 0x8A} // non-breaking space, hair space
		for _, char := range unicodeChars {
			if isWhitespace(char) {
				t.Errorf("isWhitespace(%d) should be false for unicode char", char)
			}
			if isPunctuation(char) {
				t.Errorf("isPunctuation(%d) should be false for unicode char", char)
			}
		}
	})

	// Test malformed JSON handling
	t.Run("MalformedJSON", func(t *testing.T) {
		malformedCases := []struct {
			input string
			pos   int
		}{
			{"tru", 0},   // incomplete "true"
			{"fals", 0},  // incomplete "false"
			{"nul", 0},   // incomplete "null"
			{"TRUE", 0},  // wrong case
			{"FALSE", 0}, // wrong case
			{"NULL", 0},  // wrong case
		}

		for _, tc := range malformedCases {
			if isBooleanOrNull(tc.input, tc.pos) {
				t.Errorf("isBooleanOrNull(%q, %d) should be false for malformed input", tc.input, tc.pos)
			}
		}
	})

	// Test boundary conditions
	t.Run("BoundaryConditions", func(t *testing.T) {
		// Test at string boundaries
		if isBooleanOrNull("true", 4) {
			t.Error("isBooleanOrNull('true', 4) should be false (out of bounds)")
		}
		if isBooleanOrNull("false", 5) {
			t.Error("isBooleanOrNull('false', 5) should be false (out of bounds)")
		}

		// Test negative position (should not panic)
		if isBooleanOrNull("true", -1) {
			t.Error("isBooleanOrNull('true', -1) should be false")
		}
	})
}
