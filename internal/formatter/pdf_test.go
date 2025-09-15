package formatter

import (
	"testing"
)

func TestIsJSONStructuralChar(t *testing.T) {
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
		{'-', false},
	}

	for _, test := range tests {
		result := isJSONStructuralChar(test.char)
		if result != test.expected {
			t.Errorf("isJSONStructuralChar(%q) = %v, expected %v", test.char, result, test.expected)
		}
	}
}

func TestIsJSONNumber(t *testing.T) {
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
		{':', false},
	}

	for _, test := range tests {
		result := isJSONNumber(test.char)
		if result != test.expected {
			t.Errorf("isJSONNumber(%q) = %v, expected %v", test.char, result, test.expected)
		}
	}
}

func TestStartsWithBoolean(t *testing.T) {
	tests := []struct {
		line     string
		pos      int
		expected bool
	}{
		{"true", 0, true},
		{"false", 0, true},
		{"  true", 2, true},
		{"  false", 2, true},
		{"truthy", 0, false},     // not a valid JSON boolean
		{"falsehood", 0, false},  // not a valid JSON boolean
		{"true_value", 0, false}, // underscore makes it not a boolean
		{"false_flag", 0, false}, // underscore makes it not a boolean
		{"null", 0, false},
		{"string", 0, false},
		{"123", 0, false},
		{"", 0, false},
	}

	for _, test := range tests {
		result := startsWithBoolean(test.line, test.pos)
		if result != test.expected {
			t.Errorf("startsWithBoolean(%q, %d) = %v, expected %v", test.line, test.pos, result, test.expected)
		}
	}
}

func TestPDFEdgeCases(t *testing.T) {
	// Test maxJSONLineLength constant usage
	t.Run("MaxLineLengthConstant", func(t *testing.T) {
		// Verify constant is properly defined
		if maxJSONLineLength != 100 {
			t.Errorf("maxJSONLineLength should be 100, got %d", maxJSONLineLength)
		}
	})

	// Test very long lines (similar to PDF rendering limits)
	t.Run("VeryLongLines", func(t *testing.T) {
		longLine := make([]byte, maxJSONLineLength+10)
		for i := range longLine {
			longLine[i] = 'a'
		}
		longLineStr := string(longLine)

		// Test that our functions handle long strings without panicking
		if startsWithBoolean(longLineStr, 0) {
			t.Error("startsWithBoolean should return false for non-boolean long string")
		}

		// Test at various positions
		for pos := 0; pos < len(longLineStr) && pos < 10; pos++ {
			if isJSONStructuralChar(longLineStr[pos]) {
				t.Errorf("isJSONStructuralChar should return false for 'a' at position %d", pos)
			}
		}
	})

	// Test boundary conditions for startsWithBoolean
	t.Run("BooleanBoundaryConditions", func(t *testing.T) {
		// Test at end of string
		if startsWithBoolean("true", 4) {
			t.Error("startsWithBoolean('true', 4) should be false (out of bounds)")
		}

		// Test negative position
		if startsWithBoolean("true", -1) {
			t.Error("startsWithBoolean('true', -1) should be false")
		}

		// Test empty string
		if startsWithBoolean("", 0) {
			t.Error("startsWithBoolean('', 0) should be false")
		}
	})

	// Test unicode and special characters
	t.Run("UnicodeAndSpecialChars", func(t *testing.T) {
		specialChars := []byte{0xFF, 0x00, 0x80, 0xC0}
		for _, char := range specialChars {
			if isJSONStructuralChar(char) {
				t.Errorf("isJSONStructuralChar(%d) should be false for special char", char)
			}
			if isJSONNumber(char) {
				t.Errorf("isJSONNumber(%d) should be false for special char", char)
			}
		}
	})
}
