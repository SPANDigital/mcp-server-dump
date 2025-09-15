package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testReadFileUsage   = "Read file contents"
	testReadOnlyAccess  = "Read-only access"
	testFilePermissions = 0o644
)

func TestLoadContextConfig(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid_yaml_file", func(t *testing.T) {
		testValidYAMLFile(t, tempDir)
	})

	t.Run("valid_json_file", func(t *testing.T) {
		testValidJSONFile(t, tempDir)
	})

	t.Run("multiple_files_merge", func(t *testing.T) {
		testMultipleFilesMerge(t, tempDir)
	})

	t.Run("invalid_yaml", func(t *testing.T) {
		testInvalidYAML(t, tempDir)
	})

	t.Run("unsupported_format", func(t *testing.T) {
		testUnsupportedFormat(t, tempDir)
	})

	t.Run("directory_traversal_attempt", func(t *testing.T) {
		testDirectoryTraversalAttempt(t)
	})
}

// testValidYAMLFile tests loading a valid YAML context file
func testValidYAMLFile(t *testing.T, tempDir string) {
	t.Helper()
	yamlContent := `
contexts:
  tools:
    read_file:
      usage: "Read file contents"
      security: "Read-only access"
  resources:
    "file://*":
      access: "File system access"
  prompts:
    analyze:
      purpose: "Code analysis"
`
	yamlPath := filepath.Join(tempDir, "context.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := LoadContextConfig([]string{yamlPath})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	validateYAMLResults(t, result)
}

// validateYAMLResults validates the results from loading a YAML context file
func validateYAMLResults(t *testing.T, result *ContextConfig) {
	t.Helper()
	// Check tools
	if result.Contexts.Tools["read_file"]["usage"] != testReadFileUsage {
		t.Errorf("Expected usage '%s', got '%s'", testReadFileUsage, result.Contexts.Tools["read_file"]["usage"])
	}
	if result.Contexts.Tools["read_file"]["security"] != testReadOnlyAccess {
		t.Errorf("Expected security '%s', got '%s'", testReadOnlyAccess, result.Contexts.Tools["read_file"]["security"])
	}

	// Check resources
	if result.Contexts.Resources["file://*"]["access"] != "File system access" {
		t.Errorf("Expected access 'File system access', got '%s'", result.Contexts.Resources["file://*"]["access"])
	}

	// Check prompts
	if result.Contexts.Prompts["analyze"]["purpose"] != "Code analysis" {
		t.Errorf("Expected purpose 'Code analysis', got '%s'", result.Contexts.Prompts["analyze"]["purpose"])
	}
}

// testValidJSONFile tests loading a valid JSON context file
func testValidJSONFile(t *testing.T, tempDir string) {
	t.Helper()
	jsonContent := `{
  "contexts": {
    "tools": {
      "write_file": {
        "usage": "Write file contents"
      }
    },
    "resources": {},
    "prompts": {}
  }
}`
	jsonPath := filepath.Join(tempDir, "context.json")
	if err := os.WriteFile(jsonPath, []byte(jsonContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := LoadContextConfig([]string{jsonPath})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Contexts.Tools["write_file"]["usage"] != "Write file contents" {
		t.Errorf("Expected usage 'Write file contents', got '%s'", result.Contexts.Tools["write_file"]["usage"])
	}
}

// testMultipleFilesMerge tests merging multiple context files
func testMultipleFilesMerge(t *testing.T, tempDir string) {
	t.Helper()
	basePath, overridePath := createMergeTestFiles(t, tempDir)

	result, err := LoadContextConfig([]string{basePath, overridePath})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	validateMergeResults(t, result)
}

// createMergeTestFiles creates test files for merge testing
func createMergeTestFiles(t *testing.T, tempDir string) (basePath, overridePath string) {
	t.Helper()
	baseContent := `
contexts:
  tools:
    read_file:
      usage: "Read file contents"
`
	basePath = filepath.Join(tempDir, "base.yaml")
	if err := os.WriteFile(basePath, []byte(baseContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create base file: %v", err)
	}

	overrideContent := `
contexts:
  tools:
    read_file:
      security: "Read-only access"
    write_file:
      usage: "Write file contents"
`
	overridePath = filepath.Join(tempDir, "override.yaml")
	if err := os.WriteFile(overridePath, []byte(overrideContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create override file: %v", err)
	}

	return basePath, overridePath
}

// validateMergeResults validates the results from merging multiple context files
func validateMergeResults(t *testing.T, result *ContextConfig) {
	t.Helper()
	// Check merged read_file tool
	if result.Contexts.Tools["read_file"]["usage"] != testReadFileUsage {
		t.Errorf("Expected usage '%s', got '%s'", testReadFileUsage, result.Contexts.Tools["read_file"]["usage"])
	}
	if result.Contexts.Tools["read_file"]["security"] != testReadOnlyAccess {
		t.Errorf("Expected security '%s', got '%s'", testReadOnlyAccess, result.Contexts.Tools["read_file"]["security"])
	}

	// Check write_file tool
	if result.Contexts.Tools["write_file"]["usage"] != "Write file contents" {
		t.Errorf("Expected usage 'Write file contents', got '%s'", result.Contexts.Tools["write_file"]["usage"])
	}
}

// testInvalidYAML tests handling of invalid YAML content
func testInvalidYAML(t *testing.T, tempDir string) {
	t.Helper()
	invalidContent := `invalid: yaml: content: [`
	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	_, err := LoadContextConfig([]string{invalidPath})
	if err == nil {
		t.Error("Expected error for invalid YAML but got none")
	}
}

// testUnsupportedFormat tests handling of unsupported file formats
func testUnsupportedFormat(t *testing.T, tempDir string) {
	t.Helper()
	txtContent := `some content`
	txtPath := filepath.Join(tempDir, "config.txt")
	if err := os.WriteFile(txtPath, []byte(txtContent), testFilePermissions); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	_, err := LoadContextConfig([]string{txtPath})
	if err == nil {
		t.Error("Expected error for unsupported format but got none")
	}
	if !strings.Contains(err.Error(), "unsupported file format") {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}
}

// testDirectoryTraversalAttempt tests protection against directory traversal attacks
func testDirectoryTraversalAttempt(t *testing.T) {
	t.Helper()
	_, err := LoadContextConfig([]string{"../malicious.yaml"})
	if err == nil {
		t.Error("Expected error for directory traversal but got none")
	}
	if !strings.Contains(err.Error(), "directory traversal not allowed") {
		t.Errorf("Expected directory traversal error, got: %v", err)
	}
}

func TestContextConfig_ApplyToTool(t *testing.T) {
	config := &ContextConfig{}
	config.Contexts.Tools = make(map[string]map[string]string)
	config.Contexts.Tools["read_file"] = map[string]string{
		"usage":    testReadFileUsage,
		"security": testReadOnlyAccess,
	}

	t.Run("matching_tool", func(t *testing.T) {
		tool := &Tool{Name: "read_file"}
		config.ApplyToTool(tool)

		if tool.Context["usage"] != testReadFileUsage {
			t.Errorf("Expected usage '%s', got '%s'", testReadFileUsage, tool.Context["usage"])
		}
		if tool.Context["security"] != testReadOnlyAccess {
			t.Errorf("Expected security '%s', got '%s'", testReadOnlyAccess, tool.Context["security"])
		}
	})

	t.Run("non_matching_tool", func(t *testing.T) {
		tool := &Tool{Name: "write_file"}
		config.ApplyToTool(tool)

		if tool.Context != nil {
			t.Errorf("Expected nil context for non-matching tool, got %+v", tool.Context)
		}
	})

	t.Run("tool_with_existing_context", func(t *testing.T) {
		tool := &Tool{
			Name: "read_file",
			Context: map[string]string{
				"existing": "value",
			},
		}
		config.ApplyToTool(tool)

		if tool.Context["existing"] != "value" {
			t.Errorf("Expected existing 'value', got '%s'", tool.Context["existing"])
		}
		if tool.Context["usage"] != testReadFileUsage {
			t.Errorf("Expected usage '%s', got '%s'", testReadFileUsage, tool.Context["usage"])
		}
	})
}

func TestContextConfig_ApplyToResource(t *testing.T) {
	config := &ContextConfig{}
	config.Contexts.Resources = make(map[string]map[string]string)
	config.Contexts.Resources["file://*"] = map[string]string{
		"access": "File system access",
	}
	config.Contexts.Resources["memory://state"] = map[string]string{
		"persistence": "Session-only",
	}

	t.Run("wildcard_pattern_match", func(t *testing.T) {
		resource := &Resource{URI: "file:///home/user/document.txt"}
		config.ApplyToResource(resource)

		if resource.Context["access"] != "File system access" {
			t.Errorf("Expected access 'File system access', got '%s'", resource.Context["access"])
		}
	})

	t.Run("exact_match", func(t *testing.T) {
		resource := &Resource{URI: "memory://state"}
		config.ApplyToResource(resource)

		if resource.Context["persistence"] != "Session-only" {
			t.Errorf("Expected persistence 'Session-only', got '%s'", resource.Context["persistence"])
		}
	})

	t.Run("no_match", func(t *testing.T) {
		resource := &Resource{URI: "http://example.com"}
		config.ApplyToResource(resource)

		if resource.Context != nil {
			t.Errorf("Expected nil context for non-matching resource, got %+v", resource.Context)
		}
	})
}

func TestContextConfig_ApplyToPrompt(t *testing.T) {
	config := &ContextConfig{}
	config.Contexts.Prompts = make(map[string]map[string]string)
	config.Contexts.Prompts["analyze_code"] = map[string]string{
		"purpose": "Code analysis",
		"output":  "Structured report",
	}

	t.Run("matching_prompt", func(t *testing.T) {
		prompt := &Prompt{Name: "analyze_code"}
		config.ApplyToPrompt(prompt)

		if prompt.Context["purpose"] != "Code analysis" {
			t.Errorf("Expected purpose 'Code analysis', got '%s'", prompt.Context["purpose"])
		}
		if prompt.Context["output"] != "Structured report" {
			t.Errorf("Expected output 'Structured report', got '%s'", prompt.Context["output"])
		}
	})

	t.Run("non_matching_prompt", func(t *testing.T) {
		prompt := &Prompt{Name: "generate_code"}
		config.ApplyToPrompt(prompt)

		if prompt.Context != nil {
			t.Errorf("Expected nil context for non-matching prompt, got %+v", prompt.Context)
		}
	})
}

func TestSecurityValidations(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("file_size_limit", func(t *testing.T) {
		// Create a file that exceeds the size limit
		largePath := filepath.Join(tempDir, "large.yaml")
		largeContent := strings.Repeat("x", maxContextFileSize+1)
		if err := os.WriteFile(largePath, []byte(largeContent), testFilePermissions); err != nil {
			t.Fatalf("Failed to create large file: %v", err)
		}

		_, err := LoadContextConfig([]string{largePath})
		if err == nil {
			t.Error("Expected error for oversized file but got none")
		}
		if !strings.Contains(err.Error(), "file size exceeds maximum") {
			t.Errorf("Expected file size error, got: %v", err)
		}
	})

	t.Run("directory_traversal_protection", func(t *testing.T) {
		// Attempt directory traversal
		_, err := LoadContextConfig([]string{"../../../etc/passwd"})
		if err == nil {
			t.Error("Expected error for directory traversal but got none")
		}
		if !strings.Contains(err.Error(), "directory traversal not allowed") {
			t.Errorf("Expected directory traversal error, got: %v", err)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := LoadContextConfig([]string{"/nonexistent/file.yaml"})
		if err == nil {
			t.Error("Expected error for nonexistent file but got none")
		}
	})
}

func TestMatchURIPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		uri      string
		expected bool
	}{
		{"file://*", "file:///home/user/document.txt", true},
		{"file://*", "memory://state", false},
		{"memory://*", "memory://state/current", true},
		{"*.txt", "document.txt", true},
		{"*.txt", "document.pdf", false},
		{"http://example.com", "http://example.com", true},
		{"http://example.com", "https://example.com", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.pattern, tt.uri), func(t *testing.T) {
			result := matchURIPattern(tt.pattern, tt.uri)
			if result != tt.expected {
				t.Errorf("matchURIPattern(%s, %s) = %v, expected %v", tt.pattern, tt.uri, result, tt.expected)
			}
		})
	}
}
