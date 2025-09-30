package formatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// TestHugoModulesIntegration tests the complete Hugo modules workflow with Presidium layouts
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
		BaseURL:      "https://example.com",
		LanguageCode: "en-us",
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

	// Verify Hugo modules configuration for Presidium
	expectedConfigs := []string{
		"module:",
		"github.com/spandigital/presidium-styling-base",
		"github.com/spandigital/presidium-layouts-base",
		"menu:",
		"identifier:",
		"enterprise_key:",
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

// createTestSiteDir creates a temporary test site directory
func createTestSiteDir(t *testing.T) string {
	t.Helper()

	testSiteDir, err := os.MkdirTemp("", "hugo_modules_commands_test_*")
	if err != nil {
		t.Fatalf("Failed to create test site directory: %v", err)
	}
	t.Cleanup(func() {
		if removeErr := os.RemoveAll(testSiteDir); removeErr != nil {
			t.Logf("Failed to remove test site directory: %v", removeErr)
		}
	})

	return testSiteDir
}

// testHugoModInit tests Hugo module initialization
func testHugoModInit(t *testing.T, hugoBinary *HugoBinaryTestHelper, testSiteDir string) {
	t.Helper()

	err := hugoBinary.InitModule(testSiteDir, "example.com/test-site")
	if err != nil {
		t.Logf("Hugo mod init failed (this is expected in test environments): %v", err)
		t.Skip("Skipping remaining module tests due to mod init failure")
	}
	t.Log("Hugo mod init completed successfully")

	goModPath := filepath.Join(testSiteDir, "go.mod")
	if _, statErr := os.Stat(goModPath); os.IsNotExist(statErr) {
		t.Error("go.mod file was not created by hugo mod init")
	}
}

// testHugoModGetPresidium tests getting Presidium modules
func testHugoModGetPresidium(t *testing.T, hugoBinary *HugoBinaryTestHelper, testSiteDir string) {
	t.Helper()

	err := hugoBinary.GetModule(testSiteDir, "github.com/spandigital/presidium-styling-base")
	if err != nil {
		t.Logf("Hugo mod get styling failed (this is expected in test environments): %v", err)
		t.Skip("Skipping build test due to mod get failure")
	}

	err = hugoBinary.GetModule(testSiteDir, "github.com/spandigital/presidium-layouts-base")
	if err != nil {
		t.Logf("Hugo mod get layouts failed (this is expected in test environments): %v", err)
		t.Skip("Skipping build test due to mod get failure")
	}
	t.Log("Hugo mod get presidium completed successfully")

	goSumPath := filepath.Join(testSiteDir, "go.sum")
	if _, statErr := os.Stat(goSumPath); os.IsNotExist(statErr) {
		t.Log("go.sum file not found - modules may not have been fully downloaded")
	}
}

// testHugoVersion tests Hugo version command
func testHugoVersion(t *testing.T, hugoBinary *HugoBinaryTestHelper) {
	t.Helper()

	version, err := hugoBinary.GetVersion()
	if err != nil {
		t.Logf("Hugo version check failed: %v", err)
		return
	}

	t.Logf("Hugo version output: %s", strings.TrimSpace(version))

	versionLower := strings.ToLower(version)
	if !strings.Contains(versionLower, "hugo") || !strings.Contains(versionLower, "0.150") {
		t.Errorf("Unexpected version output format: %s", version)
	}
}

// testHugoModulesCommands tests Hugo modules commands with actual Hugo binary
func testHugoModulesCommands(t *testing.T) {
	t.Helper()

	hugoBinary := NewHugoBinaryTestHelper(t)
	hugoBinary.SkipIfDownloadFails()

	testSiteDir := createTestSiteDir(t)
	t.Logf("Testing Hugo modules commands in directory: %s", testSiteDir)

	t.Run("hugo mod init", func(t *testing.T) {
		testHugoModInit(t, hugoBinary, testSiteDir)
	})

	t.Run("hugo mod get presidium", func(t *testing.T) {
		testHugoModGetPresidium(t, hugoBinary, testSiteDir)
	})

	t.Run("hugo version", func(t *testing.T) {
		testHugoVersion(t, hugoBinary)
	})

	t.Run("hugo build with presidium", func(t *testing.T) {
		// First create minimal required content structure
		contentDir := filepath.Join(testSiteDir, "content")
		if mkdirErr := os.MkdirAll(contentDir, 0o755); mkdirErr != nil {
			t.Fatalf("Failed to create content directory: %v", mkdirErr)
		}

		// Create a basic index file with Presidium-required frontmatter
		indexContent := `---
title: "Test Site"
author: test@example.com
---

# Welcome to Test Site

This is a test site for Hugo modules integration testing with Presidium layouts.
`
		indexPath := filepath.Join(contentDir, "_index.md")
		if writeErr := os.WriteFile(indexPath, []byte(indexContent), 0o644); writeErr != nil {
			t.Fatalf("Failed to create index file: %v", writeErr)
		}

		// Create basic Hugo config that uses Presidium modules
		configContent := `baseURL: 'https://test.example.com'
languageCode: 'en-us'
title: 'Test Site'

module:
  imports:
    - path: github.com/spandigital/presidium-styling-base
    - path: github.com/spandigital/presidium-layouts-base

outputFormats:
  MenuIndex:
    baseName: menu
    mediaType: application/json
  SearchMap:
    baseName: searchmap
    mediaType: application/json

outputs:
  home:
    - HTML
    - RSS
    - MenuIndex
    - SearchMap

params:
  enterprise_key: test-site
  frontmatter:
    - key: author
      type: email
      strict: true
    - key: title
      type: text
      strict: true
`
		configPath := filepath.Join(testSiteDir, "hugo.yml")
		if writeErr := os.WriteFile(configPath, []byte(configContent), 0o644); writeErr != nil {
			t.Fatalf("Failed to create Hugo config: %v", writeErr)
		}

		// Attempt to build the site
		err := hugoBinary.BuildSite(testSiteDir)
		if err != nil {
			t.Logf("Hugo build failed (this may be expected in test environments): %v", err)
			// Don't fail the test - build failures are common in CI environments
			// due to network restrictions or missing dependencies
		} else {
			t.Log("Hugo build completed successfully with Presidium")

			// Verify public directory was created
			publicDir := filepath.Join(testSiteDir, "public")
			if _, statErr := os.Stat(publicDir); os.IsNotExist(statErr) {
				t.Error("public directory was not created by Hugo build")
			} else {
				// Verify Presidium-specific output files were created
				menuIndexPath := filepath.Join(publicDir, "menu.json")
				searchMapPath := filepath.Join(publicDir, "searchmap.json")

				if _, statErr := os.Stat(menuIndexPath); os.IsNotExist(statErr) {
					t.Log("MenuIndex output file not found - may not be generated in minimal site")
				} else {
					t.Log("MenuIndex output file created successfully")
				}

				if _, statErr := os.Stat(searchMapPath); os.IsNotExist(statErr) {
					t.Log("SearchMap output file not found - may not be generated in minimal site")
				} else {
					t.Log("SearchMap output file created successfully")
				}
			}
		}
	})

	t.Log("Hugo modules commands test workflow completed")
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
				BaseURL:      "https://docs.example.com",
				LanguageCode: "en-US",
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

// TestPresidiumSiteBuild tests building an actual Hugo site with Presidium from generated output
func TestPresidiumSiteBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Presidium site build test in short mode")
	}

	hugoBinary := NewHugoBinaryTestHelper(t)
	hugoBinary.SkipIfDownloadFails()

	siteDir, serverInfo, hugoConfig := setupPresidiumBuildTest(t)
	defer cleanupTestDir(t, siteDir)

	generatePresidiumSite(t, serverInfo, siteDir, hugoConfig)

	t.Run("Initialize Hugo modules", func(t *testing.T) {
		testInitializeHugoModules(t, hugoBinary, siteDir)
	})

	t.Run("Download Presidium modules", func(t *testing.T) {
		testDownloadPresidiumModules(t, hugoBinary, siteDir)
	})

	t.Run("Build Hugo site", func(t *testing.T) {
		testBuildPresidiumSite(t, hugoBinary, siteDir)
	})
}

// setupPresidiumBuildTest creates test environment for Presidium build
func setupPresidiumBuildTest(t *testing.T) (string, *model.ServerInfo, *HugoConfig) {
	t.Helper()

	siteDir, err := os.MkdirTemp("", "presidium_build_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp site directory: %v", err)
	}
	t.Logf("Testing Presidium site build in directory: %s", siteDir)

	serverInfo := createPresidiumTestServerInfo()
	hugoConfig := &HugoConfig{
		BaseURL:      "https://test-presidium.example.com",
		LanguageCode: "en",
	}

	return siteDir, serverInfo, hugoConfig
}

// createPresidiumTestServerInfo creates test MCP server info for Presidium tests
func createPresidiumTestServerInfo() *model.ServerInfo {
	return &model.ServerInfo{
		Name:    "test-mcp-server",
		Version: "1.0.0",
		Capabilities: model.Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
		Tools: []model.Tool{
			{
				Name:        "search_tool",
				Description: "A tool for searching content",
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
				Name:        "test_resource",
				Description: "A test resource",
				MimeType:    "application/json",
			},
		},
		Prompts: []model.Prompt{
			{
				Name:        "test_prompt",
				Description: "A test prompt",
				Arguments:   []any{"arg1"},
			},
		},
	}
}

// cleanupTestDir removes test directory
func cleanupTestDir(t *testing.T, siteDir string) {
	t.Helper()
	if removeErr := os.RemoveAll(siteDir); removeErr != nil {
		t.Logf("Failed to remove site directory: %v", removeErr)
	}
}

// generatePresidiumSite generates Hugo site with Presidium
func generatePresidiumSite(t *testing.T, serverInfo *model.ServerInfo, siteDir string, hugoConfig *HugoConfig) {
	t.Helper()

	err := FormatHugo(
		serverInfo,
		siteDir,
		true,
		"yaml",
		map[string]any{"author": "test@example.com"},
		hugoConfig,
		[]string{"MCP"},
		testHugoTemplateFS,
	)
	if err != nil {
		t.Fatalf("Failed to generate Hugo site: %v", err)
	}
	t.Log("Hugo site generated successfully")
}

// testInitializeHugoModules tests Hugo module initialization
func testInitializeHugoModules(t *testing.T, hugoBinary *HugoBinaryTestHelper, siteDir string) {
	t.Helper()

	err := hugoBinary.InitModule(siteDir, "example.com/test-mcp-server")
	if err != nil {
		t.Skipf("Hugo mod init failed: %v", err)
	}
	t.Log("Hugo modules initialized")
}

// testDownloadPresidiumModules tests downloading Presidium modules
func testDownloadPresidiumModules(t *testing.T, hugoBinary *HugoBinaryTestHelper, siteDir string) {
	t.Helper()

	if err := hugoBinary.GetModule(siteDir, "github.com/spandigital/presidium-styling-base"); err != nil {
		t.Skipf("Failed to download presidium-styling-base: %v", err)
	}

	if err := hugoBinary.GetModule(siteDir, "github.com/spandigital/presidium-layouts-base"); err != nil {
		t.Skipf("Failed to download presidium-layouts-base: %v", err)
	}
	t.Log("Presidium modules downloaded")
}

// testBuildPresidiumSite tests building the Hugo site and verifying outputs
func testBuildPresidiumSite(t *testing.T, hugoBinary *HugoBinaryTestHelper, siteDir string) {
	t.Helper()

	err := hugoBinary.BuildSite(siteDir)
	if err != nil {
		t.Logf("Hugo build failed: %v", err)
		t.Skip("Build failed - may be network or environment issue")
	}

	t.Log("Hugo site built successfully")

	publicDir := filepath.Join(siteDir, "public")
	verifyPublicDirectory(t, publicDir)
	verifyIndexFile(t, publicDir)
	verifySectionPages(t, publicDir)
	verifyPresidiumOutputs(t, publicDir)

	t.Log("All verifications passed - Presidium site built successfully!")
}

// verifyPublicDirectory verifies public directory exists
func verifyPublicDirectory(t *testing.T, publicDir string) {
	t.Helper()
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		t.Fatal("public directory was not created")
	}
}

// verifyIndexFile verifies index.html exists and contains expected content
func verifyIndexFile(t *testing.T, publicDir string) {
	t.Helper()

	indexPath := filepath.Join(publicDir, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Errorf("Failed to read index.html: %v", err)
		return
	}

	if !strings.Contains(string(indexContent), "test-mcp-server") {
		t.Error("index.html does not contain server name")
	} else {
		t.Log("index.html contains expected content")
	}
}

// verifySectionPages verifies section pages were created
func verifySectionPages(t *testing.T, publicDir string) {
	t.Helper()

	sections := []string{"tools", "resources", "prompts"}
	for _, section := range sections {
		sectionIndexPath := filepath.Join(publicDir, section, "index.html")
		if _, err := os.Stat(sectionIndexPath); os.IsNotExist(err) {
			t.Errorf("%s section index not created", section)
		} else {
			t.Logf("%s section created successfully", section)
		}
	}
}

// verifyPresidiumOutputs verifies Presidium-specific output files
func verifyPresidiumOutputs(t *testing.T, publicDir string) {
	t.Helper()

	checkPresidiumOutput(t, publicDir, "menu.json", "MenuIndex")
	checkPresidiumOutput(t, publicDir, "searchmap.json", "SearchMap")
}

// checkPresidiumOutput checks if a Presidium output file exists
func checkPresidiumOutput(t *testing.T, publicDir, filename, outputName string) {
	t.Helper()

	outputPath := filepath.Join(publicDir, filename)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Logf("%s not found (may be normal for simple sites)", filename)
	} else {
		t.Logf("Presidium %s output created", outputName)
	}
}

// TestPresidiumLayoutsFeatures tests that Presidium-specific features are properly configured
func TestPresidiumLayoutsFeatures(t *testing.T) {
	serverInfo := &model.ServerInfo{
		Name: "Presidium Feature Test",
		Capabilities: model.Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
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
		BaseURL:      "https://presidium-test.com",
		LanguageCode: "en",
	}

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "presidium_test_*")
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

	// Test Presidium-specific features
	presidiumFeatures := []string{
		"github.com/spandigital/presidium-layouts-base", // Module import
		"MenuIndex",                    // Output format
		"SearchMap",                    // Output format
		"enterprise_key:",              // Presidium config
		"sortByFilePath: true",         // Presidium param
		"lazyLoad: false",              // Presidium param
		"enableInlineShortcodes: true", // Feature flag
		"pluralizeListTitles: false",   // Feature flag
	}

	for _, feature := range presidiumFeatures {
		if !strings.Contains(configStr, feature) {
			t.Errorf("Hugo config should contain Presidium feature: %s", feature)
		}
	}

	// Verify navigation structure matches Presidium pattern
	navigationChecks := []string{
		"menu:",
		"identifier: overview",
		"name: Overview",
		"url: /",
		"identifier: tools",
		"name: Tools",
		"url: /tools/",
	}

	for _, check := range navigationChecks {
		if !strings.Contains(configStr, check) {
			t.Errorf("Navigation should contain: %s", check)
		}
	}

	t.Log("All Presidium layout features properly configured")
}
