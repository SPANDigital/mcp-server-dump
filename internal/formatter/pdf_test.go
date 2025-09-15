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
		{"truthy", 0, false},    // not a valid JSON boolean
		{"falsehood", 0, false}, // not a valid JSON boolean
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
