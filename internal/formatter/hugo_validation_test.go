package formatter

import (
	"strings"
	"testing"
)

func TestHugoConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *HugoConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config is valid",
			config:      nil,
			expectError: false,
		},
		{
			name:        "empty config is valid",
			config:      &HugoConfig{},
			expectError: false,
		},
		{
			name: "valid config",
			config: &HugoConfig{
				BaseURL:      "https://example.com",
				LanguageCode: "en-us",
			},
			expectError: false,
		},
		{
			name: "invalid BaseURL - no protocol",
			config: &HugoConfig{
				BaseURL: "example.com",
			},
			expectError: true,
			errorMsg:    "invalid BaseURL",
		},
		{
			name: "invalid BaseURL - contains spaces",
			config: &HugoConfig{
				BaseURL: "https://example .com",
			},
			expectError: true,
			errorMsg:    "invalid character",
		},
		{
			name: "invalid LanguageCode - wrong format",
			config: &HugoConfig{
				LanguageCode: "english",
			},
			expectError: true,
			errorMsg:    "invalid LanguageCode",
		},
		{
			name: "invalid LanguageCode - not well formed",
			config: &HugoConfig{
				LanguageCode: "en-us-ca",
			},
			expectError: true,
			errorMsg:    "invalid language code",
		},
		{
			name: "valid LanguageCode variations",
			config: &HugoConfig{
				LanguageCode: "fr",
			},
			expectError: false,
		},
		{
			name: "valid LanguageCode with region",
			config: &HugoConfig{
				LanguageCode: "fr-ca",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"empty URL", "", false},
		{"valid HTTPS", "https://example.com", false},
		{"valid HTTP", "http://example.com", false},
		{"valid with port", "https://example.com:8080", false},
		{"valid with path", "https://example.com/path", false},
		{"invalid - no protocol", "example.com", true},
		{"invalid - wrong protocol", "ftp://example.com", true},
		{"invalid - contains spaces", "https://example .com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateLanguageCode(t *testing.T) {
	tests := []struct {
		name        string
		langCode    string
		expectError bool
	}{
		{"empty language code", "", false},
		{"valid - en", "en", false},
		{"valid - fr", "fr", false},
		{"valid - en-us", "en-us", false},
		{"valid - fr-ca", "fr-ca", false},
		{"valid - uppercase converted", "EN-US", false},
		{"valid - 3-letter code", "fil", false},
		{"valid - script subtag", "zh-Hans", false},
		{"valid - 3-letter region", "en-USA", false},
		{"invalid - not well formed", "english", true},
		{"invalid - not well formed tag", "en-us-ca", true},
		{"invalid - numbers in language", "e1", true},
		{"invalid - special chars", "en@us", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLanguageCode(tt.langCode)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

