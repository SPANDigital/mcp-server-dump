package formatter

import (
	"bytes"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/language"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// HugoConfig holds Hugo-specific configuration options
type HugoConfig struct {
	BaseURL       string
	LanguageCode  string
	EnterpriseKey string // Optional Presidium enterprise key
	AuthorStrict  bool   // Whether to require author field in frontmatter (default: false)
}

// Validate validates the Hugo configuration and returns any errors found
func (hc *HugoConfig) Validate() error {
	if hc == nil {
		return nil // nil config is valid (uses defaults)
	}

	// Validate BaseURL if provided
	if hc.BaseURL != "" {
		if err := validateURL(hc.BaseURL); err != nil {
			return fmt.Errorf("invalid BaseURL: %w", err)
		}
	}

	// Validate LanguageCode format if provided
	if hc.LanguageCode != "" {
		if err := validateLanguageCode(hc.LanguageCode); err != nil {
			return fmt.Errorf("invalid LanguageCode: %w", err)
		}
	}

	return nil
}

// validateURL validates if the provided URL is well-formed using the standard library
func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	// Parse URL using standard library for robust validation
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	// Verify scheme is http or https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	// Verify host is present
	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// validateLanguageCode validates if the language code follows basic format
func validateLanguageCode(langCode string) error {
	if langCode == "" {
		return nil
	}

	// Use golang.org/x/text/language for robust validation
	// This properly handles:
	// - 2 and 3 letter language codes (e.g., "en", "fil")
	// - Script subtags (e.g., "zh-Hans" for Simplified Chinese)
	// - Region codes (e.g., "en-US", "pt-BR", "fr-CA")
	_, err := language.Parse(langCode)
	if err != nil {
		return fmt.Errorf("invalid language code %q: %w (examples: 'en', 'en-US', 'zh-Hans', 'pt-BR')", langCode, err)
	}

	return nil
}

// Compile regex patterns once at package level for performance
var (
	nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9-]+`)
	multiHyphenRegex = regexp.MustCompile(`-+`)
)

// FormatHugo generates a Hugo documentation site structure with hierarchical content organization.
// It creates a content directory with subdirectories for tools, resources, and prompts,
// each containing individual markdown files and section index files (_index.md).
func FormatHugo(info *model.ServerInfo, outputDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, hugoConfig *HugoConfig, customInitialisms []string, templateFS embed.FS) error {
	// Validate Hugo configuration first
	if err := hugoConfig.Validate(); err != nil {
		return fmt.Errorf("invalid Hugo configuration: %w", err)
	}

	// Validate and sanitize output directory to prevent path traversal
	outputDir = filepath.Clean(outputDir)
	if strings.Contains(outputDir, "..") {
		return fmt.Errorf("outputDir cannot contain path traversal sequences")
	}
	if filepath.IsAbs(outputDir) && strings.HasPrefix(outputDir, "/") {
		// Additional check for critical system directories (exact matches only)
		criticalPaths := []string{"/", "/bin", "/etc", "/usr", "/sys", "/proc", "/dev"}
		for _, criticalPath := range criticalPaths {
			if outputDir == criticalPath || outputDir == criticalPath+"/" {
				return fmt.Errorf("outputDir cannot be a critical system directory: %s", outputDir)
			}
		}
	}

	// Use a single timestamp for all files to ensure consistent ordering
	generationTime := time.Now()

	// Create the content directory structure
	contentDir := filepath.Join(outputDir, "content")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create content directory: %w", err)
	}

	// Generate hugo.yml configuration file
	if err := generateHugoConfig(info, outputDir, hugoConfig, templateFS); err != nil {
		return fmt.Errorf("failed to generate hugo.yml: %w", err)
	}

	// Generate root _index.md
	if err := generateRootIndex(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, generationTime); err != nil {
		return fmt.Errorf("failed to generate root index: %w", err)
	}

	// Generate all content sections
	return generateContentSections(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, generationTime, customInitialisms, templateFS)
}

// generateContentSections generates all content sections (tools, resources, prompts) if they exist
func generateContentSections(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, generationTime time.Time, customInitialisms []string, templateFS embed.FS) error {
	// Generate tools section
	if len(info.Tools) > 0 {
		if err := generateToolsSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate tools section: %w", err)
		}
	}

	// Generate resources section
	if len(info.Resources) > 0 {
		if err := generateResourcesSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate resources section: %w", err)
		}
	}

	// Generate prompts section
	if len(info.Prompts) > 0 {
		if err := generatePromptsSection(info, contentDir, includeFrontmatter, frontmatterFormat, customFields, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate prompts section: %w", err)
		}
	}

	return nil
}

// generateRootIndex creates the root _index.md file with server information
func generateRootIndex(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, generationTime time.Time) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = fmt.Sprintf("%s Documentation", info.Name)
	fields["date"] = generationTime.Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = 1

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields, true) // Root index is an index file
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
func generateToolsSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, generationTime time.Time, customInitialisms []string, templateFS embed.FS) error {
	toolsDir := filepath.Join(contentDir, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	// Generate tools section index
	if err := generateSectionIndex(toolsDir, "Tools", "Available MCP tools and their documentation", len(info.Tools), includeFrontmatter, frontmatterFormat, customFields, info, generationTime); err != nil {
		return err
	}

	// Generate individual tool files
	for i, tool := range info.Tools {
		if err := generateContentFile(toolsDir, &tool, tool.Name, "tool", i+1, includeFrontmatter, frontmatterFormat, customFields, info, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate tool file for %s: %w", tool.Name, err)
		}
	}

	return nil
}

// generateResourcesSection creates the resources directory and all resource markdown files
func generateResourcesSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, generationTime time.Time, customInitialisms []string, templateFS embed.FS) error {
	resourcesDir := filepath.Join(contentDir, "resources")
	if err := os.MkdirAll(resourcesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}

	// Generate resources section index
	if err := generateSectionIndex(resourcesDir, "Resources", "Available MCP resources and their documentation", len(info.Resources), includeFrontmatter, frontmatterFormat, customFields, info, generationTime); err != nil {
		return err
	}

	// Generate individual resource files
	for i, resource := range info.Resources {
		if err := generateContentFile(resourcesDir, &resource, resource.Name, "resource", i+1, includeFrontmatter, frontmatterFormat, customFields, info, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate resource file for %s: %w", resource.Name, err)
		}
	}

	return nil
}

// generatePromptsSection creates the prompts directory and all prompt markdown files
func generatePromptsSection(info *model.ServerInfo, contentDir string, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, generationTime time.Time, customInitialisms []string, templateFS embed.FS) error {
	promptsDir := filepath.Join(contentDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Generate prompts section index
	if err := generateSectionIndex(promptsDir, "Prompts", "Available MCP prompts and their documentation", len(info.Prompts), includeFrontmatter, frontmatterFormat, customFields, info, generationTime); err != nil {
		return err
	}

	// Generate individual prompt files
	for i, prompt := range info.Prompts {
		if err := generateContentFile(promptsDir, &prompt, prompt.Name, "prompt", i+1, includeFrontmatter, frontmatterFormat, customFields, info, generationTime, customInitialisms, templateFS); err != nil {
			return fmt.Errorf("failed to generate prompt file for %s: %w", prompt.Name, err)
		}
	}

	return nil
}

// generateSectionIndex creates a section _index.md file
func generateSectionIndex(dir, title, description string, itemCount int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, generationTime time.Time) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = title
	fields["date"] = generationTime.Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = getSectionWeight(title)

	// Add menu configuration for Presidium navigation
	// identifier: lowercase-with-hyphens version of title
	// name: Human-readable title
	identifier := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	menu := map[string]any{
		"main": map[string]any{
			"identifier": identifier,
			"name":       title,
			"weight":     getSectionWeight(title),
		},
	}
	fields["menu"] = menu

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields, true) // Section index is an index file
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
func generateContentFile(dir string, data any, name, itemType string, weight int, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, info *model.ServerInfo, generationTime time.Time, customInitialisms []string, templateFS embed.FS) error {
	var content bytes.Buffer

	// Prepare frontmatter fields
	fields := make(map[string]any)
	fields["title"] = humanizeKeyWithCustomInitialisms(name, customInitialisms)
	fields["date"] = generationTime.Format(time.RFC3339)
	fields["draft"] = false
	fields["weight"] = weight
	fields["type"] = itemType

	// Add custom fields
	for k, v := range customFields {
		fields[k] = v
	}

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, fields, false) // Individual content files are not index files
		if err != nil {
			return fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		content.WriteString(frontmatter)
	}

	// Create template with functions
	templateName := itemType + ".md.tmpl"
	tmpl := template.New(templateName).Funcs(template.FuncMap{
		"json":       jsonIndent,
		"contains":   strings.Contains,
		"sortedKeys": getSortedKeys,
		"humanizeKey": func(key string) string {
			return humanizeKeyWithCustomInitialisms(key, customInitialisms)
		},
		"humanize": func(name string) string {
			return humanizeKeyWithCustomInitialisms(name, customInitialisms)
		},
	})

	// Parse template from embedded filesystem - try test path first, then production path
	testPath := "test_templates/hugo/" + templateName
	prodPath := "templates/hugo/" + templateName

	tmpl, err := tmpl.ParseFS(templateFS, testPath)
	if err != nil {
		// Reset template for production path
		tmpl = template.New(templateName).Funcs(template.FuncMap{
			"json":       jsonIndent,
			"contains":   strings.Contains,
			"sortedKeys": getSortedKeys,
			"humanizeKey": func(key string) string {
				return humanizeKeyWithCustomInitialisms(key, customInitialisms)
			},
			"humanize": func(name string) string {
				return humanizeKeyWithCustomInitialisms(name, customInitialisms)
			},
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

// slugify converts a string to a URL-safe slug
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	s = nonAlphaNumRegex.ReplaceAllString(s, "")

	// Remove duplicate hyphens
	s = multiHyphenRegex.ReplaceAllString(s, "-")

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

// getSortedKeys returns sorted keys from a map[string]string for deterministic iteration
func getSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// generateHugoConfig creates a hugo.yml configuration file with sensible defaults
func generateHugoConfig(info *model.ServerInfo, outputDir string, hugoConfig *HugoConfig, templateFS embed.FS) error {
	// Create template
	tmpl := template.New("hugo.yml.tmpl")

	// Parse template from embedded filesystem - try test path first, then production path
	testPath := "test_templates/hugo/hugo.yml.tmpl"
	prodPath := "templates/hugo/hugo.yml.tmpl"

	tmpl, err := tmpl.ParseFS(templateFS, testPath)
	if err != nil {
		// Reset template for production path
		tmpl = template.New("hugo.yml.tmpl")
		tmpl, err = tmpl.ParseFS(templateFS, prodPath)
	}
	if err != nil {
		return fmt.Errorf("failed to parse hugo.yml template: %w", err)
	}

	// Execute template with combined data
	var content bytes.Buffer
	templateData := struct {
		*model.ServerInfo
		HugoConfig       *HugoConfig
		GeneratorVersion string
	}{
		ServerInfo:       info,
		HugoConfig:       hugoConfig,
		GeneratorVersion: "dev", // Version string for generated hugo.yml comment
	}
	if err := tmpl.Execute(&content, templateData); err != nil {
		return fmt.Errorf("failed to execute hugo.yml template: %w", err)
	}

	// Write to file
	configPath := filepath.Join(outputDir, "hugo.yml")
	return os.WriteFile(configPath, content.Bytes(), 0o644)
}
