package formatter

import (
	"bytes"
	"embed"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// FormatHTML formats server info as HTML
func FormatHTML(info *model.ServerInfo, includeTOC bool, templateFS embed.FS) (string, error) {
	// First generate markdown
	markdown, err := FormatMarkdown(info, includeTOC, false, "", nil, templateFS)
	if err != nil {
		return "", err
	}

	// Convert markdown to HTML using Goldmark with GitHub Flavored Markdown extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Linkify,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
