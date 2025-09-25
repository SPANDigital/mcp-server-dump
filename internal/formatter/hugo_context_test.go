package formatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

func TestFormatHugoWithContext(t *testing.T) {
	// Create test server info
	info := &model.ServerInfo{
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

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "hugo_context_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate Hugo output
	err = FormatHugo(info, tempDir, false, "", nil, testHugoTemplateFS)
	if err != nil {
		t.Fatalf("FormatHugo failed: %v", err)
	}

	// Test tool context rendering
	t.Run("tool context fields", func(t *testing.T) {
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
	})

	// Test resource context rendering
	t.Run("resource context fields", func(t *testing.T) {
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
	})

	// Test prompt context rendering
	t.Run("prompt context fields", func(t *testing.T) {
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
	})
}