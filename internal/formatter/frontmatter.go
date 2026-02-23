package formatter

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// GenerateFrontmatter generates frontmatter for markdown output in YAML, TOML, or JSON format.
//
// The isIndexFile parameter controls which metadata fields are included in the generated frontmatter:
//   - When true: Includes comprehensive server-level metadata such as title, version, capabilities,
//     and item counts. Use this for main documentation index files (_index.md) or primary landing pages.
//   - When false: Generates minimal frontmatter with only basic metadata (generated_at, generator).
//     Use this for individual tool/resource/prompt pages to keep frontmatter clean and focused.
//
// Parameters:
//   - info: The MCP server information containing name, version, capabilities, and items
//   - format: The frontmatter format ("yaml", "toml", or "json")
//   - customFields: Optional map of custom key-value pairs to add/override in frontmatter
//   - isIndexFile: Controls metadata inclusion (see above)
//
// Returns the formatted frontmatter string with delimiters, ready to prepend to markdown content.
func GenerateFrontmatter(info *model.ServerInfo, format string, customFields map[string]any, isIndexFile bool) (string, error) {
	// Build frontmatter data
	frontmatter := make(map[string]any)

	// Always include basic metadata
	frontmatter["generated_at"] = time.Now().Format(time.RFC3339)
	frontmatter["generator"] = "mcp-server-dump"

	// Only include server-level metadata for index files
	if isIndexFile {
		frontmatter["title"] = info.Name + " Documentation"
		if info.Version != "" {
			frontmatter["version"] = info.Version
		}

		// Capabilities
		frontmatter["capabilities"] = map[string]bool{
			"tools":     info.Capabilities.Tools,
			"resources": info.Capabilities.Resources,
			"prompts":   info.Capabilities.Prompts,
		}

		// Counts
		frontmatter["tools_count"] = len(info.Tools)
		frontmatter["resources_count"] = len(info.Resources)
		frontmatter["prompts_count"] = len(info.Prompts)
	}

	// Add custom fields (these can override auto-generated ones)
	maps.Copy(frontmatter, customFields)

	// Format based on requested format
	switch format {
	case "yaml":
		return formatYAMLFrontmatter(frontmatter)
	case "toml":
		return formatTOMLFrontmatter(frontmatter)
	case "json":
		return formatJSONFrontmatter(frontmatter)
	default:
		return "", fmt.Errorf("unsupported frontmatter format: %s", format)
	}
}

func formatYAMLFrontmatter(data map[string]any) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML frontmatter: %w", err)
	}
	return "---\n" + string(yamlBytes) + "---\n\n", nil
}

func formatTOMLFrontmatter(data map[string]any) (string, error) {
	// Simple TOML formatting - for more complex cases, consider using a TOML library
	var tomlLines []string
	tomlLines = append(tomlLines, "+++")

	for key, value := range data {
		switch v := value.(type) {
		case string:
			tomlLines = append(tomlLines, fmt.Sprintf("%s = %q", key, v))
		case int:
			tomlLines = append(tomlLines, fmt.Sprintf("%s = %d", key, v))
		case float64:
			tomlLines = append(tomlLines, fmt.Sprintf("%s = %g", key, v))
		case bool:
			tomlLines = append(tomlLines, fmt.Sprintf("%s = %t", key, v))
		case []string:
			quotedItems := make([]string, len(v))
			for i, item := range v {
				quotedItems[i] = fmt.Sprintf("%q", item)
			}
			tomlLines = append(tomlLines, fmt.Sprintf("%s = [%s]", key, strings.Join(quotedItems, ", ")))
		case map[string]bool:
			// Handle capabilities map
			for subkey, subvalue := range v {
				tomlLines = append(tomlLines,
					fmt.Sprintf("[%s.%s]", key, subkey),
					fmt.Sprintf("value = %t", subvalue))
			}
		default:
			// Fallback to string representation
			tomlLines = append(tomlLines, fmt.Sprintf("%s = %q", key, fmt.Sprintf("%v", v)))
		}
	}

	tomlLines = append(tomlLines, "+++")
	return strings.Join(tomlLines, "\n") + "\n\n", nil
}

func formatJSONFrontmatter(data map[string]any) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON frontmatter: %w", err)
	}
	return "---\n" + string(jsonBytes) + "\n---\n\n", nil
}
