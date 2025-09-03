package app

import "embed"

// TemplateFS contains embedded template files for markdown generation
//
//go:embed templates/*.tmpl
var TemplateFS embed.FS
