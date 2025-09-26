package formatter

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

//go:embed test_templates/hugo/*.tmpl
var testHugoTemplateFS embed.FS

func TestFormatHugo(t *testing.T) {
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
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"param": map[string]any{"type": "string"},
					},
				},
			},
		},
		Resources: []model.Resource{
			{
				URI:         "test://resource",
				Name:        "test_resource",
				Description: "A test resource",
				MimeType:    "text/plain",
			},
		},
		Prompts: []model.Prompt{
			{
				Name:        "test_prompt",
				Description: "A test prompt",
				Arguments:   []any{"arg1", "arg2"},
			},
		},
	}

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "hugo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp dir: %v", err)
		}
	}()

	// Test without frontmatter
	t.Run("without frontmatter", func(t *testing.T) {
		hugoConfig := &HugoConfig{} // Default empty config
		err := FormatHugo(info, tempDir, false, "", nil, hugoConfig, nil, testHugoTemplateFS)
		if err != nil {
			t.Fatalf("FormatHugo failed: %v", err)
		}

		// Verify directory structure
		verifyDirExists(t, filepath.Join(tempDir, "content"))
		verifyDirExists(t, filepath.Join(tempDir, "content", "tools"))
		verifyDirExists(t, filepath.Join(tempDir, "content", "resources"))
		verifyDirExists(t, filepath.Join(tempDir, "content", "prompts"))

		// Verify files exist
		verifyFileExists(t, filepath.Join(tempDir, "content", "_index.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "tools", "_index.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "tools", "test-tool.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "resources", "_index.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "resources", "test-resource.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "prompts", "_index.md"))
		verifyFileExists(t, filepath.Join(tempDir, "content", "prompts", "test-prompt.md"))

		// Verify content of root index
		content, err := os.ReadFile(filepath.Join(tempDir, "content", "_index.md"))
		if err != nil {
			t.Fatalf("Failed to read _index.md: %v", err)
		}
		if !strings.Contains(string(content), "Test Server") {
			t.Errorf("Root index should contain server name")
		}
		if !strings.Contains(string(content), "1.0.0") {
			t.Errorf("Root index should contain version")
		}
	})

	// Test with frontmatter
	t.Run("with frontmatter", func(t *testing.T) {
		tempDir2, err := os.MkdirTemp("", "hugo_fm_test_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() {
			if cleanupErr := os.RemoveAll(tempDir2); cleanupErr != nil {
				t.Logf("Failed to clean up temp dir: %v", cleanupErr)
			}
		}()

		customFields := map[string]any{
			"author": "Test Author",
			"tags":   []string{"test", "mcp"},
		}

		hugoConfig := &HugoConfig{} // Default empty config
		err = FormatHugo(info, tempDir2, true, "yaml", customFields, hugoConfig, nil, testHugoTemplateFS)
		if err != nil {
			t.Fatalf("FormatHugo with frontmatter failed: %v", err)
		}

		// Verify frontmatter exists
		content, err := os.ReadFile(filepath.Join(tempDir2, "content", "_index.md"))
		if err != nil {
			t.Fatalf("Failed to read _index.md: %v", err)
		}
		if !strings.HasPrefix(string(content), "---") {
			t.Errorf("Content should start with YAML frontmatter delimiter")
		}
		if !strings.Contains(string(content), "author: Test Author") {
			t.Errorf("Frontmatter should contain custom author field")
		}
		if !strings.Contains(string(content), "title: Test Server Documentation") {
			t.Errorf("Frontmatter should contain title")
		}
	})
}

func TestFormatHugoErrorPaths(t *testing.T) {
	info := &model.ServerInfo{
		Name:    "Test Server",
		Version: "1.0.0",
		Capabilities: model.Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
	}

	t.Run("path traversal protection", func(t *testing.T) {
		hugoConfig := &HugoConfig{} // Default empty config
		err := FormatHugo(info, "../malicious", false, "", nil, hugoConfig, nil, testHugoTemplateFS)
		if err == nil {
			t.Error("Expected error for path traversal attempt")
		}
		if !strings.Contains(err.Error(), "path traversal") {
			t.Errorf("Expected path traversal error, got: %v", err)
		}
	})

	t.Run("system directory protection", func(t *testing.T) {
		hugoConfig := &HugoConfig{} // Default empty config
		err := FormatHugo(info, "/etc", false, "", nil, hugoConfig, nil, testHugoTemplateFS)
		if err == nil {
			t.Error("Expected error for system directory")
		}
		if !strings.Contains(err.Error(), "critical system directory") {
			t.Errorf("Expected system directory error, got: %v", err)
		}
	})

	t.Run("invalid output directory", func(t *testing.T) {
		hugoConfig := &HugoConfig{} // Default empty config
		err := FormatHugo(info, "/nonexistent/readonly/path", false, "", nil, hugoConfig, nil, testHugoTemplateFS)
		if err == nil {
			t.Error("Expected error for invalid directory")
		}
	})
}

func TestSlugifyEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"unicode characters", "测试文档", ""},
		{"only special chars", "@#$%^&*()", ""},
		{"very long name", strings.Repeat("a", 200), strings.Repeat("a", 200)},
		{"mixed unicode and ascii", "test-测试-doc", "test-doc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			if tt.expected == "" && result != "unnamed" {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, "unnamed")
			} else if tt.expected != "" && result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Name", "simple-name"},
		{"name_with_underscores", "name-with-underscores"},
		{"Name With  Multiple   Spaces", "name-with-multiple-spaces"},
		{"special@#$characters", "specialcharacters"},
		{"MiXeD CaSe NaMe", "mixed-case-name"},
		{"--leading-and-trailing--", "leading-and-trailing"},
		{"", "unnamed"},
		{"123numbers", "123numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetSectionWeight(t *testing.T) {
	tests := []struct {
		title    string
		expected int
	}{
		{"Tools", 10},
		{"tools", 10},
		{"TOOLS", 10},
		{"Resources", 20},
		{"resources", 20},
		{"Prompts", 30},
		{"prompts", 30},
		{"Other", 100},
		{"Custom Section", 100},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := getSectionWeight(tt.title)
			if result != tt.expected {
				t.Errorf("getSectionWeight(%q) = %d, want %d", tt.title, result, tt.expected)
			}
		})
	}
}

// Helper functions
func verifyDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("Directory %s does not exist: %v", path, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", path)
	}
}

func verifyFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("File %s does not exist: %v", path, err)
		return
	}
	if info.IsDir() {
		t.Errorf("%s is a directory, expected file", path)
	}
}
