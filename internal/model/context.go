package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	// maxContextFileSize limits context files to 10MB to prevent YAML/JSON bombs
	maxContextFileSize = 10 * 1024 * 1024
)

// ContextConfig represents the structure of context configuration files
type ContextConfig struct {
	Contexts struct {
		Tools     map[string]map[string]string `yaml:"tools" json:"tools"`
		Resources map[string]map[string]string `yaml:"resources" json:"resources"`
		Prompts   map[string]map[string]string `yaml:"prompts" json:"prompts"`
	} `yaml:"contexts" json:"contexts"`
}

// LoadContextConfig loads and merges multiple context files
func LoadContextConfig(files []string) (*ContextConfig, error) {
	config := &ContextConfig{}
	config.Contexts.Tools = make(map[string]map[string]string)
	config.Contexts.Resources = make(map[string]map[string]string)
	config.Contexts.Prompts = make(map[string]map[string]string)

	for _, file := range files {
		if err := config.mergeFile(file); err != nil {
			return nil, fmt.Errorf("failed to load context file %s: %w", file, err)
		}
	}

	return config, nil
}

// mergeFile loads a single context file and merges it with existing configuration
func (c *ContextConfig) mergeFile(filename string) error {
	cleanPath, err := validateFilePath(filename)
	if err != nil {
		return err
	}

	_, err = validateFileSize(cleanPath)
	if err != nil {
		return err
	}

	data, err := readContextFile(cleanPath)
	if err != nil {
		return err
	}

	tempConfig, err := parseContextFile(data, cleanPath)
	if err != nil {
		return err
	}

	c.mergeContextData(&tempConfig)
	return nil
}

// validateFilePath validates the file path to prevent directory traversal attacks
func validateFilePath(filename string) (string, error) {
	cleanPath := filepath.Clean(filename)
	if strings.Contains(cleanPath, "..") {
		return "", errors.New("invalid file path: directory traversal not allowed")
	}
	return cleanPath, nil
}

// validateFileSize checks the file size to prevent processing overly large files
func validateFileSize(cleanPath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() > maxContextFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes: %d", maxContextFileSize, fileInfo.Size())
	}
	return fileInfo, nil
}

// readContextFile safely reads the context file content with size limits
func readContextFile(cleanPath string) ([]byte, error) {
	file, err := os.Open(cleanPath) // #nosec G304 - path is validated and cleaned
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close context file: %v\n", closeErr)
		}
	}()

	limitedReader := io.LimitReader(file, maxContextFileSize)
	return io.ReadAll(limitedReader)
}

// parseContextFile parses the context file based on its extension
func parseContextFile(data []byte, cleanPath string) (config ContextConfig, err error) {
	ext := strings.ToLower(filepath.Ext(cleanPath))

	switch ext {
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, &config); err != nil {
			err = fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err = json.Unmarshal(data, &config); err != nil {
			err = fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		err = fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
	}
	return config, err
}

// mergeContextData merges the parsed context data into the current configuration
func (c *ContextConfig) mergeContextData(tempConfig *ContextConfig) {
	c.mergeTools(tempConfig.Contexts.Tools)
	c.mergeResources(tempConfig.Contexts.Resources)
	c.mergePrompts(tempConfig.Contexts.Prompts)
}

// mergeTools merges tool contexts from the temporary config
func (c *ContextConfig) mergeTools(tools map[string]map[string]string) {
	for toolName, contexts := range tools {
		if c.Contexts.Tools[toolName] == nil {
			c.Contexts.Tools[toolName] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Tools[toolName][key] = value
		}
	}
}

// mergeResources merges resource contexts from the temporary config
func (c *ContextConfig) mergeResources(resources map[string]map[string]string) {
	for resourcePattern, contexts := range resources {
		if c.Contexts.Resources[resourcePattern] == nil {
			c.Contexts.Resources[resourcePattern] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Resources[resourcePattern][key] = value
		}
	}
}

// mergePrompts merges prompt contexts from the temporary config
func (c *ContextConfig) mergePrompts(prompts map[string]map[string]string) {
	for promptName, contexts := range prompts {
		if c.Contexts.Prompts[promptName] == nil {
			c.Contexts.Prompts[promptName] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Prompts[promptName][key] = value
		}
	}
}

// ApplyToTool applies matching context to a tool
func (c *ContextConfig) ApplyToTool(tool *Tool) {
	if contexts, exists := c.Contexts.Tools[tool.Name]; exists {
		if tool.Context == nil {
			tool.Context = make(map[string]string)
		}
		for key, value := range contexts {
			tool.Context[key] = value
		}
	}
}

// ApplyToResource applies matching context to a resource using pattern matching
func (c *ContextConfig) ApplyToResource(resource *Resource) {
	for pattern, contexts := range c.Contexts.Resources {
		var matched bool

		// Try custom URI pattern matching first
		if strings.Contains(pattern, "*") {
			matched = matchURIPattern(pattern, resource.URI)
		} else {
			// Exact string match
			matched = pattern == resource.URI
		}

		if matched {
			if resource.Context == nil {
				resource.Context = make(map[string]string)
			}
			for key, value := range contexts {
				resource.Context[key] = value
			}
			break // Apply first matching pattern only
		}
	}
}

// matchURIPattern matches URI patterns like "file://*" against URIs like "file:///path/file.txt"
func matchURIPattern(pattern, uri string) bool {
	// Handle simple prefix patterns like "file://*"
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(uri, prefix)
	}

	// Handle suffix patterns like "*.txt"
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(uri, suffix)
	}

	// Handle middle wildcards - for now, fall back to exact match
	return pattern == uri
}

// ApplyToPrompt applies matching context to a prompt
func (c *ContextConfig) ApplyToPrompt(prompt *Prompt) {
	if contexts, exists := c.Contexts.Prompts[prompt.Name]; exists {
		if prompt.Context == nil {
			prompt.Context = make(map[string]string)
		}
		for key, value := range contexts {
			prompt.Context[key] = value
		}
	}
}
