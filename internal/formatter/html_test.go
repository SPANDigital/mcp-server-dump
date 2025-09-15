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
		{"truthy", 0, false},    // not a valid JSON boolean
		{"falsehood", 0, false}, // not a valid JSON boolean
		{"nullable", 0, false},  // not a valid JSON boolean
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
