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

	siteDir, serverInfo, hugoConfig := setupHugoModulesTest(t)
	defer func() {
		if err := os.RemoveAll(siteDir); err != nil {
			t.Logf("Failed to remove site directory: %v", err)
		}
	}()

	t.Run("Generate Hugo site with modules", func(t *testing.T) {
		testHugoSiteGeneration(t, serverInfo, siteDir, hugoConfig)
	})

	t.Run("Verify content structure", func(t *testing.T) {
		testContentStructure(t, siteDir)
	})

	t.Run("Test Hugo modules commands (simulated)", func(t *testing.T) {
		testHugoModulesCommands(t)
	})
}

// setupHugoModulesTest sets up the test environment and returns necessary components
func setupHugoModulesTest(t *testing.T) (string, *model.ServerInfo, *HugoConfig) {
	t.Helper()

	// Create Hugo binary test helper
	hugoBinary := NewHugoBinaryTestHelper(t)
	hugoBinary.SkipIfDownloadFails()

	// Create temporary site directory
	siteDir, err := os.MkdirTemp("", "hugo_modules_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp site directory: %v", err)
	}
	t.Logf("Testing Hugo modules in directory: %s", siteDir)

	// Create test server info
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

	// Create Hugo configuration
	hugoConfig := &HugoConfig{
		BaseURL:         "https://example.com",
		LanguageCode:    "en-us",
		Github:          "testuser",
		Twitter:         "testuser",
		SiteLogo:        "images/logo.png",
		GoogleAnalytics: "G-TEST123456",
	}

	return siteDir, serverInfo, hugoConfig
}

// testHugoSiteGeneration tests Hugo site generation with modules
func testHugoSiteGeneration(t *testing.T, serverInfo *model.ServerInfo, siteDir string, hugoConfig *HugoConfig) {
	t.Helper()

	// Generate Hugo documentation site
	err := FormatHugo(
		serverInfo,
		siteDir,
		true, // include frontmatter
		"yaml",
		map[string]any{"author": "Test Author"},
		hugoConfig,
		[]string{"MCP", "PROTO"}, // custom initialisms
		testHugoTemplateFS,       // Use test template filesystem
	)
	if err != nil {
		t.Fatalf("Failed to generate Hugo site: %v", err)
	}

	// Verify Hugo configuration
	verifyHugoConfig(t, siteDir)
	t.Logf("Hugo configuration generated successfully with modules support")
}

// verifyHugoConfig verifies the generated Hugo configuration
func verifyHugoConfig(t *testing.T, siteDir string) {
	t.Helper()

	configPath := filepath.Join(siteDir, "hugo.yml")
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		t.Fatalf("Hugo configuration file not created: %s", configPath)
	}

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Hugo config: %v", err)
	}

	configStr := string(configContent)

	// Verify Hugo modules configuration
	expectedConfigs := []string{
		"module:",
		"github.com/imfing/hextra",
		"navbar:",
		"search:",
		"sidebar:",
		"G-TEST123456",
	}

	for _, config := range expectedConfigs {
		if !strings.Contains(configStr, config) {
			t.Errorf("Hugo config should contain: %s", config)
		}
	}
}

// testContentStructure tests the generated content directory structure
func testContentStructure(t *testing.T, siteDir string) {
	t.Helper()

	contentDir := filepath.Join(siteDir, "content")
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		t.Fatalf("Content directory not created: %s", contentDir)
	}

	// Check main index file
	indexPath := filepath.Join(contentDir, "_index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatalf("Main index file not created: %s", indexPath)
	}

	// Check section directories
	sectionsToCheck := []string{"tools", "resources", "prompts"}
	for _, section := range sectionsToCheck {
		verifySectionDirectory(t, contentDir, section)
	}

	// Verify tool content file
	verifyToolContent(t, contentDir)
	t.Logf("Content structure verified successfully")
}

// verifySectionDirectory verifies a section directory exists with proper structure
func verifySectionDirectory(t *testing.T, contentDir, section string) {
	t.Helper()

	sectionDir := filepath.Join(contentDir, section)
	if _, err := os.Stat(sectionDir); os.IsNotExist(err) {
		t.Fatalf("Section directory not created: %s", sectionDir)
	}

	sectionIndex := filepath.Join(sectionDir, "_index.md")
	if _, err := os.Stat(sectionIndex); os.IsNotExist(err) {
		t.Fatalf("Section index not created: %s", sectionIndex)
	}
}

// verifyToolContent verifies tool content file has proper structure
func verifyToolContent(t *testing.T, contentDir string) {
	t.Helper()

	toolFile := filepath.Join(contentDir, "tools", "example-tool.md")
	if _, err := os.Stat(toolFile); os.IsNotExist(err) {
		t.Fatalf("Tool content file not created: %s", toolFile)
	}

	toolContent, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("Failed to read tool file: %v", err)
	}

	toolStr := string(toolContent)
	expectedContent := []string{
		"---",
		"title: example_tool",
		"weight:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(toolStr, content) {
			t.Errorf("Tool file should contain: %s", content)
		}
	}
}

// testHugoModulesCommands tests Hugo modules commands (currently simulated)
func testHugoModulesCommands(t *testing.T) {
	t.Helper()
	t.Skip("Hugo binary execution not fully implemented - would test: mod init, mod get, build")
	t.Logf("Would test Hugo modules workflow: init -> get -> build")
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
		GoogleAnalytics: "G-1234567890",
	}

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "hextra_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	// Generate Hugo site
	err = FormatHugo(
		serverInfo,
		tempDir,
		true,
		"yaml",
		nil,
		hugoConfig,
		nil,
		testHugoTemplateFS,
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
		"github.com/imfing/hextra", // Module import
		"displayTitle: true",       // Navbar config
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
