package formatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

func TestFormatHugoWithContext(t *testing.T) {
	info := createTestServerInfo()

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "hugo_context_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			t.Logf("Failed to clean up temp dir: %v", cleanupErr)
		}
	}()

	// Generate Hugo output
	err = FormatHugo(info, tempDir, false, "", nil, testHugoTemplateFS)
	if err != nil {
		t.Fatalf("FormatHugo failed: %v", err)
	}

	// Test individual components
	t.Run("tool context fields", func(t *testing.T) {
		testToolContextRendering(t, tempDir)
	})

	t.Run("resource context fields", func(t *testing.T) {
		testResourceContextRendering(t, tempDir)
	})

	t.Run("prompt context fields", func(t *testing.T) {
		testPromptContextRendering(t, tempDir)
	})
}

func createTestServerInfo() *model.ServerInfo {
	return &model.ServerInfo{
		Name:    "Test Server",
		Version: "1.0.0",
		Capabilities: model.Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
		Tools: []model.Tool{
			{
				Name:        "test_tool",
				Description: "A test tool",
				Context: map[string]string{
					"usage":    "Use this tool for testing",
					"security": "Only for authorized users",
					"examples": "```json\n{\"test\": true}\n```",
					"notes":    "Additional notes here",
				},
			},
		},
		Resources: []model.Resource{
			{
				URI:         "test://resource",
				Name:        "test_resource",
				Description: "A test resource",
				MimeType:    "text/plain",
				Context: map[string]string{
					"description": "Extended description",
					"access":      "Read-only access",
					"examples":    "## Examples\n\n- Example 1\n- Example 2",
				},
			},
		},
		Prompts: []model.Prompt{
			{
				Name:        "test_prompt",
				Description: "A test prompt",
				Context: map[string]string{
					"purpose":    "Testing purposes",
					"parameters": "- param1: string\n- param2: number",
					"output":     "JSON response",
				},
			},
		},
	}
}

func testToolContextRendering(t *testing.T, tempDir string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(tempDir, "content", "tools", "test-tool.md"))
	if err != nil {
		t.Fatalf("Failed to read tool file: %v", err)
	}

	contentStr := string(content)

	// Check that context section exists
	if !strings.Contains(contentStr, "## Additional Documentation") {
		t.Error("Additional Documentation section not found")
	}

	// Check individual context fields
	if !strings.Contains(contentStr, "usage") {
		t.Error("Context field 'usage' not rendered")
	}
	if !strings.Contains(contentStr, "Use this tool for testing") {
		t.Error("Context value for 'usage' not rendered")
	}

	if !strings.Contains(contentStr, "security") {
		t.Error("Context field 'security' not rendered")
	}
	if !strings.Contains(contentStr, "Only for authorized users") {
		t.Error("Context value for 'security' not rendered")
	}

	// Multi-line fields should be rendered as subsections
	if !strings.Contains(contentStr, "### examples") {
		t.Error("Multi-line context field 'examples' not rendered as subsection")
	}
}

func testResourceContextRendering(t *testing.T, tempDir string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(tempDir, "content", "resources", "test-resource.md"))
	if err != nil {
		t.Fatalf("Failed to read resource file: %v", err)
	}

	contentStr := string(content)

	// Check that context section exists
	if !strings.Contains(contentStr, "## Additional Documentation") {
		t.Error("Additional Documentation section not found")
	}

	// Check individual context fields
	if !strings.Contains(contentStr, "access") {
		t.Error("Context field 'access' not rendered")
	}
	if !strings.Contains(contentStr, "Read-only access") {
		t.Error("Context value for 'access' not rendered")
	}

	// Multi-line fields should be rendered as subsections
	if !strings.Contains(contentStr, "### examples") {
		t.Error("Multi-line context field 'examples' not rendered as subsection")
	}
}

func testPromptContextRendering(t *testing.T, tempDir string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(tempDir, "content", "prompts", "test-prompt.md"))
	if err != nil {
		t.Fatalf("Failed to read prompt file: %v", err)
	}

	contentStr := string(content)

	// Check that context section exists
	if !strings.Contains(contentStr, "## Additional Documentation") {
		t.Error("Additional Documentation section not found")
	}

	// Check individual context fields
	if !strings.Contains(contentStr, "purpose") {
		t.Error("Context field 'purpose' not rendered")
	}
	if !strings.Contains(contentStr, "Testing purposes") {
		t.Error("Context value for 'purpose' not rendered")
	}

	if !strings.Contains(contentStr, "output") {
		t.Error("Context field 'output' not rendered")
	}

	// Multi-line fields should be rendered as subsections
	if !strings.Contains(contentStr, "### parameters") {
		t.Error("Multi-line context field 'parameters' not rendered as subsection")
	}
}
