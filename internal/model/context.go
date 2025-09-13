package model

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
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
	file, err := os.Open(filename) // #nosec G304 - filename is from user-controlled CLI parameter
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close() // #nosec G104 - file is read-only and error is not critical
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Determine file format by extension
	ext := strings.ToLower(filepath.Ext(filename))

	var tempConfig ContextConfig
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &tempConfig); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &tempConfig); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s (supported: .yaml, .yml, .json)", ext)
	}

	// Merge tools
	for toolName, contexts := range tempConfig.Contexts.Tools {
		if c.Contexts.Tools[toolName] == nil {
			c.Contexts.Tools[toolName] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Tools[toolName][key] = value
		}
	}

	// Merge resources
	for resourcePattern, contexts := range tempConfig.Contexts.Resources {
		if c.Contexts.Resources[resourcePattern] == nil {
			c.Contexts.Resources[resourcePattern] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Resources[resourcePattern][key] = value
		}
	}

	// Merge prompts
	for promptName, contexts := range tempConfig.Contexts.Prompts {
		if c.Contexts.Prompts[promptName] == nil {
			c.Contexts.Prompts[promptName] = make(map[string]string)
		}
		for key, value := range contexts {
			c.Contexts.Prompts[promptName][key] = value
		}
	}

	return nil
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
		matched, err := filepath.Match(pattern, resource.URI)
		if err != nil {
			// If pattern matching fails, try exact string match
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
