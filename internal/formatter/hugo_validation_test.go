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
				BaseURL:         "https://example.com",
				LanguageCode:    "en-us",
				Theme:           "ananke",
				Github:          "user",
				Twitter:         "user",
				SiteLogo:        "images/logo.png",
				GoogleAnalytics: "G-ABCD123456",
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
			name: "invalid SiteLogo - path traversal",
			config: &HugoConfig{
				SiteLogo: "../../../etc/passwd",
			},
			expectError: true,
			errorMsg:    "invalid SiteLogo",
		},
		{
			name: "invalid SiteLogo - system directory",
			config: &HugoConfig{
				SiteLogo: "/etc/logo.png",
			},
			expectError: true,
			errorMsg:    "logo path cannot reference system directories",
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
		{
			name: "invalid GoogleAnalytics ID",
			config: &HugoConfig{
				GoogleAnalytics: "invalid-ga-id",
			},
			expectError: true,
			errorMsg:    "invalid GoogleAnalytics ID",
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

func TestValidateLogoPath(t *testing.T) {
	tests := []struct {
		name        string
		logoPath    string
		expectError bool
	}{
		{"empty path", "", false},
		{"valid relative path", "images/logo.png", false},
		{"valid absolute path", "/home/user/logo.png", false},
		{"invalid - path traversal", "../../../etc/passwd", true},
		{"invalid - system directory", "/etc/logo.png", true},
		{"invalid - bin directory", "/bin/logo.png", true},
		{"invalid - usr directory", "/usr/share/logo.png", true},
		{"valid - non-system absolute", "/var/www/logo.png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogoPath(tt.logoPath)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateSocialHandle(t *testing.T) {
	tests := []struct {
		name        string
		handle      string
		expectError bool
	}{
		{"empty handle", "", false},
		{"valid alphanumeric", "user123", false},
		{"valid with underscore", "user_name", false},
		{"valid with hyphen", "user-name", false},
		{"valid mixed", "My_User-123", false},
		{"invalid - contains @", "@username", true},
		{"invalid - contains space", "user name", true},
		{"invalid - contains special char", "user$name", true},
		{"invalid - contains dot", "user.name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSocialHandle(tt.handle)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateGoogleAnalyticsID(t *testing.T) {
	tests := []struct {
		name        string
		gaID        string
		expectError bool
	}{
		{"empty ID", "", false},
		{"valid GA4 ID", "G-ABCD123456", false},
		{"valid GA4 ID with numbers", "G-1234567890", false},
		{"valid GA4 ID mixed", "G-AB12CD34EF", false},
		{"invalid - too short", "G-ABC123", true},
		{"invalid - too long", "G-ABCD1234567", true},
		{"invalid - no G- prefix", "ABCD123456", true},
		{"invalid - lowercase letters", "G-abcd123456", true},
		{"invalid - contains special chars", "G-ABCD@12345", true},
		{"invalid - old UA format", "UA-12345678-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGoogleAnalyticsID(tt.gaID)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
