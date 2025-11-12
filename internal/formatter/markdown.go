package formatter

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// FormatMarkdown formats server info as Markdown
func FormatMarkdown(info *model.ServerInfo, includeTOC, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any, customInitialisms []string, templateFS embed.FS) (string, error) {
	var result bytes.Buffer

	// Add frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := GenerateFrontmatter(info, frontmatterFormat, customFields, true) // Markdown format includes all server info like an index
		if err != nil {
			return "", fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		result.WriteString(frontmatter)
	}

	// Create template with custom functions
	tmpl := template.New("base.md.tmpl").Funcs(template.FuncMap{
		"anchor":     anchorName,
		"json":       jsonIndent,
		"jsonIndent": jsonIndent,
		"formatBool": formatBool,
		"contains":   strings.Contains,
		"humanizeKey": func(key string) string {
			return humanizeKeyWithCustomInitialisms(key, customInitialisms)
		},
	})

	// Parse all templates
	tmpl, err := tmpl.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse templates: %w", err)
	}

	// Execute the base template with the server info
	data := struct {
		*model.ServerInfo
		IncludeTOC bool
	}{
		ServerInfo: info,
		IncludeTOC: includeTOC,
	}

	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}
