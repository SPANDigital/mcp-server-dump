package formatter

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// FormatHugo generates a Hugo documentation site structure with hierarchical content organization.
// It creates a content directory with subdirectories for tools, resources, and prompts,
// each containing individual markdown files and section index files (_index.md).
func FormatHugo(info *model.ServerInfo, outputDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, templateFS embed.FS) error {
	// Create the content directory structure
	contentDir := filepath.Join(outputDir, "content")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create content directory: %w", err)
	}

	// Generate root _index.md
	if err := generateRootIndex(info, contentDir, includeFrontmatter, frontmatterFormat, customFields); err != nil {
		return fmt.Errorf("failed to generate root index: %w", err)
	}

	// Generate tools section
	if len(info.Tools) > 0 {
		if err := generateToolsSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, templateFS); err != nil {
			return fmt.Errorf("failed to generate tools section: %w", err)
		}
	}

	// Generate resources section
	if len(info.Resources) > 0 {
		if err := generateResourcesSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, templateFS); err != nil {
			return fmt.Errorf("failed to generate resources section: %w", err)
		}
	}

	// Generate prompts section
	if len(info.Prompts) > 0 {
		if err := generatePromptsSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, templateFS); err != nil {
			return fmt.Errorf("failed to generate prompts section: %w", err)
		}
	}

	return nil
}

// generateRootIndex creates the root _index.md file with server information
func generateRootIndex(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = fmt.Sprintf("%s Documentation", info.Name)
	fields["date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = 1

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields)
		if err != nil {
			return fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		content.WriteString(frontmatter)
	}

	// Add content
	content.WriteString(fmt.Sprintf("# %s\n\n", info.Name))
	if info.Version != "" {
		content.WriteString(fmt.Sprintf("**Version:** %s\n\n", info.Version))
	}

	// Add capabilities overview
	content.WriteString("## Capabilities\n\n")
	if info.Capabilities.Tools {
		content.WriteString(fmt.Sprintf("- ✅ **Tools:** %d available\n", len(info.Tools)))
	}
	if info.Capabilities.Resources {
		content.WriteString(fmt.Sprintf("- ✅ **Resources:** %d available\n", len(info.Resources)))
	}
	if info.Capabilities.Prompts {
		content.WriteString(fmt.Sprintf("- ✅ **Prompts:** %d available\n", len(info.Prompts)))
	}

	// Add navigation
	content.WriteString("\n## Documentation Sections\n\n")
	if len(info.Tools) > 0 {
		content.WriteString("- [Tools]({{< ref \"tools\" >}}) - Available MCP tools\n")
	}
	if len(info.Resources) > 0 {
		content.WriteString("- [Resources]({{< ref \"resources\" >}}) - Available MCP resources\n")
	}
	if len(info.Prompts) > 0 {
		content.WriteString("- [Prompts]({{< ref \"prompts\" >}}) - Available MCP prompts\n")
	}

	// Write to file
	indexPath := filepath.Join(contentDir, "_index.md")
	return os.WriteFile(indexPath, content.Bytes(), 0o644)
}

// generateToolsSection creates the tools directory and all tool markdown files
func generateToolsSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, templateFS embed.FS) error {
	toolsDir := filepath.Join(contentDir, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// Generate tools section index
	if err := generateSectionIndex(toolsDir, "Tools", "Available MCP tools and their documentation", len(info.Tools), includeFrontmatter, frontmatterFormat, customFields, info); err != nil {
		return err
	}

	// Generate individual tool files
	for i, tool := range info.Tools {
		if err := generateToolFile(toolsDir, &tool, i+1, includeFrontmatter, frontmatterFormat, customFields, info, templateFS); err != nil {
			return fmt.Errorf("failed to generate tool file for %s: %w", tool.Name, err)
		}
	}

	return nil
}

// generateResourcesSection creates the resources directory and all resource markdown files
func generateResourcesSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, templateFS embed.FS) error {
	resourcesDir := filepath.Join(contentDir, "resources")
	if err := os.MkdirAll(resourcesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}

	// Generate resources section index
	if err := generateSectionIndex(resourcesDir, "Resources", "Available MCP resources and their documentation", len(info.Resources), includeFrontmatter, frontmatterFormat, customFields, info); err != nil {
		return err
	}

	// Generate individual resource files
	for i, resource := range info.Resources {
		if err := generateResourceFile(resourcesDir, &resource, i+1, includeFrontmatter, frontmatterFormat, customFields, info, templateFS); err != nil {
			return fmt.Errorf("failed to generate resource file for %s: %w", resource.Name, err)
		}
	}

	return nil
}

// generatePromptsSection creates the prompts directory and all prompt markdown files
func generatePromptsSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, templateFS embed.FS) error {
	promptsDir := filepath.Join(contentDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Generate prompts section index
	if err := generateSectionIndex(promptsDir, "Prompts", "Available MCP prompts and their documentation", len(info.Prompts), includeFrontmatter, frontmatterFormat, customFields, info); err != nil {
		return err
	}

	// Generate individual prompt files
	for i, prompt := range info.Prompts {
		if err := generatePromptFile(promptsDir, &prompt, i+1, includeFrontmatter, frontmatterFormat, customFields, info, templateFS); err != nil {
			return fmt.Errorf("failed to generate prompt file for %s: %w", prompt.Name, err)
		}
	}

	return nil
}

// generateSectionIndex creates a section _index.md file
func generateSectionIndex(dir, title, description string, itemCount int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = title
	fields["date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = getSectionWeight(title)

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields)
		if err != nil {
			return fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		content.WriteString(frontmatter)
	}

	// Add content
	content.WriteString(fmt.Sprintf("# %s\n\n", title))
	content.WriteString(fmt.Sprintf("%s\n\n", description))
	content.WriteString(fmt.Sprintf("**Total items:** %d\n\n", itemCount))

	// Write to file
	indexPath := filepath.Join(dir, "_index.md")
	return os.WriteFile(indexPath, content.Bytes(), 0o644)
}

// generateContentFile creates an individual content markdown file with the given template
func generateContentFile(dir string, data any, name, itemType string, weight int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, templateFS embed.FS) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = name
	fields["date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = weight
	fields["type"] = itemType

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields)
		if err != nil {
			return fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		content.WriteString(frontmatter)
	}

	// Create template with functions
	templateName := itemType + ".md.tmpl"
	tmpl := template.New(templateName).Funcs(template.FuncMap{
		"json": jsonIndent,
	})

	// Parse template from embedded filesystem - try test path first, then production path
	testPath := "test_templates/hugo/" + templateName
	prodPath := "templates/hugo/" + templateName

	tmpl, err := tmpl.ParseFS(templateFS, testPath)
	if err != nil {
		// Reset template for production path
		tmpl = template.New(templateName).Funcs(template.FuncMap{
			"json": jsonIndent,
		})
		tmpl, err = tmpl.ParseFS(templateFS, prodPath)
	}
	if err != nil {
		return fmt.Errorf("failed to parse %s template: %w", itemType, err)
	}

	// Execute template
	if err := tmpl.Execute(&content, data); err != nil {
		return fmt.Errorf("failed to execute %s template: %w", itemType, err)
	}

	// Generate filename
	filename := slugify(name) + ".md"
	filePath := filepath.Join(dir, filename)

	return os.WriteFile(filePath, content.Bytes(), 0o644)
}

// generateToolFile creates an individual tool markdown file
func generateToolFile(dir string, tool *model.Tool, weight int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, templateFS embed.FS) error {
	return generateContentFile(dir, tool, tool.Name, "tool", weight, includeFrontmatter, frontmatterFormat, customFields, info, templateFS)
}

// generateResourceFile creates an individual resource markdown file
func generateResourceFile(dir string, resource *model.Resource, weight int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, templateFS embed.FS) error {
	return generateContentFile(dir, resource, resource.Name, "resource", weight, includeFrontmatter, frontmatterFormat, customFields, info, templateFS)
}

// generatePromptFile creates an individual prompt markdown file
func generatePromptFile(dir string, prompt *model.Prompt, weight int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, templateFS embed.FS) error {
	return generateContentFile(dir, prompt, prompt.Name, "prompt", weight, includeFrontmatter, frontmatterFormat, customFields, info, templateFS)
}

// slugify converts a string to a URL-safe slug
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	s = reg.ReplaceAllString(s, "")

	// Remove duplicate hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	// If empty, return a default
	if s == "" {
		s = "unnamed"
	}

	return s
}

// getSectionWeight returns a weight value for ordering sections
func getSectionWeight(title string) int {
	switch strings.ToLower(title) {
	case "tools":
		return 10
	case "resources":
		return 20
	case "prompts":
		return 30
	default:
		return 100
	}
}
