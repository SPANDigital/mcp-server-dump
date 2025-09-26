package formatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// TestHugoModulesIntegration tests the complete Hugo modules workflow with Hextra theme
func TestHugoModulesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Hugo modules integration test in short mode")
	}

	// Create Hugo binary test helper
	hugoBinary := NewHugoBinaryTestHelper(t)
	defer hugoBinary.Cleanup()

	// Skip test if Hugo download fails (CI might not have internet)
	hugoBinary.SkipIfDownloadFails()

	// Create temporary site directory
	siteDir, err := os.MkdirTemp("", "hugo_modules_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp site directory: %v", err)
	}
	defer os.RemoveAll(siteDir)

	t.Logf("Testing Hugo modules in directory: %s", siteDir)

	// Test MCP server info for Hugo generation
	serverInfo := &model.ServerInfo{
		Name:    "Test MCP Server",
		Version: "1.0.0",
		Capabilities: model.Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
		Tools: []model.Tool{
			{
				Name:        "example_tool",
				Description: "An example tool for testing Hugo modules",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query",
						},
					},
				},
			},
		},
		Resources: []model.Resource{
			{
				URI:         "test://resource",
				Name:        "example_resource",
				Description: "An example resource for testing",
				MimeType:    "application/json",
			},
		},
		Prompts: []model.Prompt{
			{
				Name:        "example_prompt",
				Description: "An example prompt for testing",
				Arguments:   []any{"arg1", "arg2"},
			},
		},
	}

	// Create Hugo configuration with Hextra theme via modules
	hugoConfig := &HugoConfig{
		BaseURL:         "https://example.com",
		LanguageCode:    "en-us",
		Github:          "testuser",
		Twitter:         "testuser",
		SiteLogo:        "images/logo.png",
		GoogleAnalytics: "G-TEST123456",
	}

	// Test Hugo site generation with modules
	t.Run("Generate Hugo site with modules", func(t *testing.T) {
		// Generate Hugo documentation site
		err := FormatHugo(
			serverInfo,
			siteDir,
			true, // include frontmatter
			"yaml",
			map[string]any{"author": "Test Author"},
			hugoConfig,
			[]string{"MCP", "PROTO"}, // custom initialisms
			HugoTemplateFS, // Use production template filesystem
		)
		if err != nil {
			t.Fatalf("Failed to generate Hugo site: %v", err)
		}

		// Verify Hugo configuration file was created
		configPath := filepath.Join(siteDir, "hugo.yml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("Hugo configuration file not created: %s", configPath)
		}

		// Read and verify Hugo configuration contains modules
		configContent, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read Hugo config: %v", err)
		}

		configStr := string(configContent)

		// Verify Hugo modules configuration is present
		if !strings.Contains(configStr, "module:") {
			t.Error("Hugo config should contain module configuration")
		}
		if !strings.Contains(configStr, "github.com/imfing/hextra") {
			t.Error("Hugo config should import Hextra theme module")
		}

		// Verify Hextra-specific configuration
		if !strings.Contains(configStr, "navbar:") {
			t.Error("Hugo config should contain Hextra navbar configuration")
		}
		if !strings.Contains(configStr, "search:") {
			t.Error("Hugo config should contain Hextra search configuration")
		}
		if !strings.Contains(configStr, "sidebar:") {
			t.Error("Hugo config should contain Hextra sidebar configuration")
		}

		// Verify Google Analytics is configured
		if !strings.Contains(configStr, "G-TEST123456") {
			t.Error("Hugo config should contain Google Analytics ID")
		}

		t.Logf("Hugo configuration generated successfully with modules support")
	})

	t.Run("Verify content structure", func(t *testing.T) {
		// Verify content directory structure was created
		contentDir := filepath.Join(siteDir, "content")
		if _, err := os.Stat(contentDir); os.IsNotExist(err) {
			t.Fatalf("Content directory not created: %s", contentDir)
		}

		// Check for main index file
		indexPath := filepath.Join(contentDir, "_index.md")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Fatalf("Main index file not created: %s", indexPath)
		}

		// Check for section directories
		sectionsToCheck := []string{"tools", "resources", "prompts"}
		for _, section := range sectionsToCheck {
			sectionDir := filepath.Join(contentDir, section)
			if _, err := os.Stat(sectionDir); os.IsNotExist(err) {
				t.Fatalf("Section directory not created: %s", sectionDir)
			}

			// Check section index
			sectionIndex := filepath.Join(sectionDir, "_index.md")
			if _, err := os.Stat(sectionIndex); os.IsNotExist(err) {
				t.Fatalf("Section index not created: %s", sectionIndex)
			}
		}

		// Verify tool content file
		toolFile := filepath.Join(contentDir, "tools", "example-tool.md")
		if _, err := os.Stat(toolFile); os.IsNotExist(err) {
			t.Fatalf("Tool content file not created: %s", toolFile)
		}

		// Read tool file and verify it has frontmatter
		toolContent, err := os.ReadFile(toolFile)
		if err != nil {
			t.Fatalf("Failed to read tool file: %v", err)
		}

		toolStr := string(toolContent)
		if !strings.HasPrefix(toolStr, "---") {
			t.Error("Tool file should start with YAML frontmatter")
		}
		if !strings.Contains(toolStr, "title: example_tool") {
			t.Error("Tool file should contain title in frontmatter")
		}
		if !strings.Contains(toolStr, "weight:") {
			t.Error("Tool file should contain weight for Hextra navigation")
		}

		t.Logf("Content structure verified successfully")
	})

	// This test would be enabled if Hugo binary execution was fully implemented
	t.Run("Test Hugo modules commands (simulated)", func(t *testing.T) {
		t.Skip("Hugo binary execution not fully implemented - would test: mod init, mod get, build")

		// These commands would be tested if RunHugo was implemented:
		// 1. hugo mod init example.com/test-site
		// 2. hugo mod get github.com/imfing/hextra
		// 3. hugo --gc --minify (build site)
		// 4. Verify that public/ directory contains built site
		// 5. Verify that go.mod and go.sum are created
		// 6. Verify that _vendor/ directory is not needed (modules approach)

		t.Logf("Would test Hugo modules workflow: init -> get -> build")
	})
}

// TestHugoConfigValidationWithModules tests the enhanced configuration validation
func TestHugoConfigValidationWithModules(t *testing.T) {
	tests := []struct {
		name        string
		config      *HugoConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid modern config with modules",
			config: &HugoConfig{
				BaseURL:         "https://docs.example.com",
				LanguageCode:    "en-US",
				Theme:           "", // Empty when using modules
				Github:          "myorg",
				Twitter:         "myorg",
				SiteLogo:        "static/logo.svg",
				GoogleAnalytics: "G-1234567890",
			},
			expectError: false,
		},
		{
			name: "valid minimal config for modules",
			config: &HugoConfig{
				BaseURL: "https://localhost:1313",
			},
			expectError: false,
		},
		{
			name: "invalid config - bad Google Analytics",
			config: &HugoConfig{
				BaseURL:         "https://example.com",
				GoogleAnalytics: "UA-123456789-1", // Old Universal Analytics format
			},
			expectError: true,
			errorMsg:    "invalid GoogleAnalytics ID",
		},
		{
			name: "invalid config - bad language code",
			config: &HugoConfig{
				BaseURL:      "https://example.com",
				LanguageCode: "en-us-west-coast", // Too many segments
			},
			expectError: true,
			errorMsg:    "invalid LanguageCode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
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

// TestHextraThemeFeatures tests that Hextra-specific features are properly configured
func TestHextraThemeFeatures(t *testing.T) {
	serverInfo := &model.ServerInfo{
		Name: "Hextra Feature Test",
		Tools: []model.Tool{
			{Name: "tool1", Description: "Test tool 1"},
			{Name: "tool2", Description: "Test tool 2"},
		},
		Resources: []model.Resource{
			{Name: "resource1", Description: "Test resource 1", URI: "test://1"},
		},
		Prompts: []model.Prompt{
			{Name: "prompt1", Description: "Test prompt 1"},
		},
	}

	hugoConfig := &HugoConfig{
		BaseURL:         "https://hextra-test.com",
		LanguageCode:    "en",
		Github:          "hextrauser",
		TwitterAnalytics: "G-HEXTRA12345",
	}

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "hextra_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate Hugo site
	err = FormatHugo(
		serverInfo,
		tempDir,
		true,
		"yaml",
		nil,
		hugoConfig,
		nil,
		HugoTemplateFS,
	)
	if err != nil {
		t.Fatalf("Failed to format Hugo: %v", err)
	}

	// Read generated config
	configPath := filepath.Join(tempDir, "hugo.yml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	configStr := string(configContent)

	// Test Hextra-specific features
	hextraFeatures := []string{
		"github.com/imfing/hextra",  // Module import
		"displayTitle: true",        // Navbar config
		"displayLogo:",             // Logo config
		"type: search",             // Search integration
		"type: separator",          // Sidebar separator
		"flexsearch",               // Search type
		"displayCopyright: true",   // Footer config
		"displayToggle: true",      // Theme toggle
	}

	for _, feature := range hextraFeatures {
		if !strings.Contains(configStr, feature) {
			t.Errorf("Hugo config should contain Hextra feature: %s", feature)
		}
	}

	// Verify navigation structure optimized for documentation
	if !strings.Contains(configStr, `name: "Documentation"`) {
		t.Error("Should have Documentation as main navigation item")
	}
	if !strings.Contains(configStr, `pageRef: /tools`) {
		t.Error("Should use pageRef instead of url for internal links")
	}

	t.Log("All Hextra theme features properly configured")
}