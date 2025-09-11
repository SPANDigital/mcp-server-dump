package formatter

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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

	// Add JSON syntax highlighting to code blocks
	highlightedHTML := addJSONSyntaxHighlighting(buf.String())

	// Wrap HTML with professional styling
	styledHTML := wrapWithCSS(highlightedHTML, info.Name)
	return styledHTML, nil
}

// addJSONSyntaxHighlighting finds JSON code blocks and adds syntax highlighting
func addJSONSyntaxHighlighting(htmlContent string) string {
	// Regular expression to find code blocks that might contain JSON
	codeBlockRegex := regexp.MustCompile(`<pre><code>([^<]*)</code></pre>`)

	return codeBlockRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the code content
		content := codeBlockRegex.FindStringSubmatch(match)
		if len(content) < 2 {
			return match
		}

		codeContent := content[1]

		// Try to determine if this is JSON by attempting to parse it
		var jsonData any
		if err := json.Unmarshal([]byte(codeContent), &jsonData); err != nil {
			// Not valid JSON, return original
			return match
		}

		// Apply syntax highlighting
		highlightedCode := highlightJSONHTML(codeContent)
		return fmt.Sprintf(`<pre><code class="json-highlighted">%s</code></pre>`, highlightedCode)
	})
}

// highlightJSONHTML applies syntax highlighting to JSON string for HTML output
func highlightJSONHTML(jsonStr string) string {
	var result strings.Builder

	i := 0
	for i < len(jsonStr) {
		char := jsonStr[i]

		// Handle whitespace - preserve formatting
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			result.WriteByte(char)
			i++
			continue
		}

		// Handle structural characters
		if char == '{' || char == '}' || char == '[' || char == ']' || char == ',' || char == ':' {
			result.WriteString(fmt.Sprintf(`<span class="json-punctuation">%c</span>`, char))
			i++
			continue
		}

		// Handle quoted strings (keys and string values)
		if char == '"' {
			quote, nextIndex := extractJSONQuotedString(jsonStr, i)

			// Determine if this is a key or value
			isKey := false
			for j := nextIndex; j < len(jsonStr); j++ {
				if jsonStr[j] == ':' {
					isKey = true
					break
				}
				if jsonStr[j] != ' ' && jsonStr[j] != '\t' && jsonStr[j] != '\n' && jsonStr[j] != '\r' {
					break
				}
			}

			if isKey {
				result.WriteString(fmt.Sprintf(`<span class="json-key">%s</span>`, quote))
			} else {
				result.WriteString(fmt.Sprintf(`<span class="json-string">%s</span>`, quote))
			}
			i = nextIndex
			continue
		}

		// Handle numbers
		if (char >= '0' && char <= '9') || char == '-' || char == '.' {
			number, nextIndex := extractJSONNumber(jsonStr, i)
			result.WriteString(fmt.Sprintf(`<span class="json-number">%s</span>`, number))
			i = nextIndex
			continue
		}

		// Handle booleans and null
		if strings.HasPrefix(jsonStr[i:], boolTrue) || strings.HasPrefix(jsonStr[i:], boolFalse) {
			var word string
			if strings.HasPrefix(jsonStr[i:], boolTrue) {
				word = boolTrue
			} else {
				word = boolFalse
			}
			result.WriteString(fmt.Sprintf(`<span class="json-boolean">%s</span>`, word))
			i += len(word)
			continue
		}

		if strings.HasPrefix(jsonStr[i:], "null") {
			result.WriteString(`<span class="json-null">null</span>`)
			i += 4
			continue
		}

		// Default: output character as-is
		result.WriteByte(char)
		i++
	}

	return result.String()
}

// wrapWithCSS wraps HTML content with professional styling
func wrapWithCSS(htmlContent, title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - MCP Server Documentation</title>
    <style>
        :root {
            --primary-blue: #2563eb;
            --success-green: #16a34a;
            --warning-red: #dc2626;
            --text-gray: #64748b;
            --light-gray: #f1f5f9;
            --border-gray: #e2e8f0;
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            line-height: 1.6;
            color: #1e293b;
            background-color: #ffffff;
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        h1 {
            color: var(--primary-blue);
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
            border-bottom: 3px solid var(--primary-blue);
            padding-bottom: 0.5rem;
        }
        
        h2 {
            color: var(--primary-blue);
            font-size: 1.75rem;
            font-weight: 600;
            margin: 2rem 0 1rem 0;
            padding-left: 0.5rem;
            border-left: 4px solid var(--primary-blue);
        }
        
        h3 {
            color: #334155;
            font-size: 1.25rem;
            font-weight: 600;
            margin: 1.5rem 0 0.75rem 0;
        }
        
        p {
            margin-bottom: 1rem;
            color: #475569;
        }
        
        ul {
            margin: 1rem 0;
            padding-left: 1.5rem;
        }
        
        li {
            margin-bottom: 0.5rem;
            color: #475569;
        }
        
        /* Status indicators with colors */
        li:contains("✓") {
            color: var(--success-green);
            font-weight: 500;
        }
        
        li:contains("✗") {
            color: var(--warning-red);
            font-weight: 500;
        }
        
        /* Code blocks */
        code {
            background-color: var(--light-gray);
            padding: 0.125rem 0.375rem;
            border-radius: 0.25rem;
            font-size: 0.875rem;
            color: #e11d48;
        }
        
        pre {
            background-color: var(--light-gray);
            padding: 1rem;
            border-radius: 0.5rem;
            overflow-x: auto;
            margin: 1rem 0;
            border: 1px solid var(--border-gray);
        }
        
        pre code {
            background: none;
            color: #334155;
            padding: 0;
        }
        
        /* JSON syntax highlighting */
        .json-highlighted {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .json-key {
            color: #4f46e5;  /* Purple for keys */
            font-weight: 500;
        }
        
        .json-string {
            color: #16a34a;  /* Green for string values */
        }
        
        .json-number {
            color: #dc2626;  /* Red for numbers */
            font-weight: 500;
        }
        
        .json-boolean {
            color: #3b82f6;  /* Blue for booleans */
            font-weight: 600;
        }
        
        .json-null {
            color: #9ca3af;  /* Gray for null */
            font-style: italic;
        }
        
        .json-punctuation {
            color: #4b5563;  /* Dark gray for punctuation */
            font-weight: 500;
        }
        
        /* Table styling */
        table {
            width: 100%%;
            border-collapse: collapse;
            margin: 1rem 0;
            background-color: white;
            border-radius: 0.5rem;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        
        th, td {
            padding: 0.75rem;
            text-align: left;
            border-bottom: 1px solid var(--border-gray);
        }
        
        th {
            background-color: var(--primary-blue);
            color: white;
            font-weight: 600;
        }
        
        tr:nth-child(even) {
            background-color: var(--light-gray);
        }
        
        /* Capability status styling */
        .capability-supported {
            color: var(--success-green);
            font-weight: 600;
        }
        
        .capability-unsupported {
            color: var(--warning-red);
            font-weight: 600;
        }
        
        /* Section styling */
        .section {
            margin: 2rem 0;
            padding: 1.5rem;
            background-color: #fafbfc;
            border-radius: 0.75rem;
            border: 1px solid var(--border-gray);
        }
        
        /* Table of contents */
        .toc {
            background-color: var(--light-gray);
            padding: 1.5rem;
            border-radius: 0.5rem;
            margin: 2rem 0;
            border-left: 4px solid var(--primary-blue);
        }
        
        .toc h2 {
            margin-top: 0;
            color: var(--primary-blue);
        }
        
        /* Responsive design */
        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }
            
            h1 {
                font-size: 2rem;
            }
            
            h2 {
                font-size: 1.5rem;
            }
        }
    </style>
</head>
<body>
    %s
</body>
</html>`, title, htmlContent)
}
