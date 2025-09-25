package app

import "embed"

// TemplateFS contains embedded template files for markdown generation
//
//go:embed templates/*.tmpl
var TemplateFS embed.FS

// HugoTemplateFS contains embedded template files for Hugo format generation
//
//go:embed templates/hugo/*.tmpl
var HugoTemplateFS embed.FS
