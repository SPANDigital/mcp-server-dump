package formatter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

//go:embed DejaVuSans.ttf
var dejaVuSansFont []byte

const (
	supportedStatus    = "Supported"
	notSupportedStatus = "Not supported"
	// Unicode bullet characters
	bulletPoint = "•" // bullet U+2022
	checkMark   = "✓" // checkmark U+2713
	crossMark   = "✗" // cross mark U+2717
)

// Color constants (RGB values)
var (
	primaryBlue  = [3]int{37, 99, 235}   // #2563eb
	successGreen = [3]int{22, 163, 74}   // #16a34a
	warningRed   = [3]int{220, 38, 38}   // #dc2626
	textGray     = [3]int{100, 116, 139} // #64748b
	lightGray    = [3]int{241, 245, 249} // #f1f5f9

	// JSON syntax highlighting colors
	jsonKey     = [3]int{79, 70, 229}   // #4f46e5 - Purple for keys
	jsonString  = [3]int{22, 163, 74}   // #16a34a - Green for string values
	jsonNumber  = [3]int{245, 101, 101} // #f56565 - Orange-red for numbers
	jsonBoolean = [3]int{59, 130, 246}  // #3b82f6 - Blue for booleans
	jsonNull    = [3]int{156, 163, 175} // #9ca3af - Gray for null
	jsonBrace   = [3]int{75, 85, 99}    // #4b5563 - Dark gray for braces/brackets
)

// FormatPDF formats server info as PDF
func FormatPDF(info *model.ServerInfo, includeTOC bool) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Load UTF-8 font for Unicode support
	pdf.AddUTF8FontFromBytes("DejaVuSans", "", dejaVuSansFont)

	// Set title color and font
	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 20)

	// Title with colored background
	title := info.Name
	if info.Version != "" {
		title += fmt.Sprintf(" (v%s)", info.Version)
	}
	pdf.Cell(0, 12, title)
	pdf.Ln(20)

	// Add a subtle line under title
	pdf.SetDrawColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetLineWidth(0.8)
	pdf.Line(10, pdf.GetY()-5, 200, pdf.GetY()-5)
	pdf.Ln(5)

	// Table of Contents
	if includeTOC {
		// TOC header
		pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetFont("DejaVuSans", "", 16)
		pdf.Cell(0, 10, "Table of Contents")
		pdf.Ln(12)

		// TOC background
		pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
		pdf.Rect(10, pdf.GetY(), 190, 40, "F")
		pdf.Ln(5)

		pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
		pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
		pdf.SetFont("DejaVuSans", "", 12)
		pdf.Cell(0, 8, "  "+bulletPoint+" Capabilities")
		pdf.Ln(6)

		if info.Capabilities.Tools && len(info.Tools) > 0 {
			pdf.Cell(0, 8, "  "+bulletPoint+" Tools")
			pdf.Ln(6)
		}

		if info.Capabilities.Resources && len(info.Resources) > 0 {
			pdf.Cell(0, 8, "  "+bulletPoint+" Resources")
			pdf.Ln(6)
		}

		if info.Capabilities.Prompts && len(info.Prompts) > 0 {
			pdf.Cell(0, 8, "  "+bulletPoint+" Prompts")
			pdf.Ln(6)
		}

		pdf.Ln(10)
	}

	// Capabilities section
	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 16)
	pdf.Bookmark("Capabilities", 0, -1)
	pdf.Cell(0, 12, "Capabilities")
	pdf.Ln(15)

	// Add section separator line
	pdf.SetDrawColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY()-5, 200, pdf.GetY()-5)
	pdf.Ln(5)

	pdf.SetFont("DejaVuSans", "", 12)

	// Tools capability
	toolsIcon := crossMark
	toolsStatus := notSupportedStatus
	if info.Capabilities.Tools {
		toolsIcon = checkMark
		toolsStatus = supportedStatus
		pdf.SetTextColor(successGreen[0], successGreen[1], successGreen[2])
	} else {
		pdf.SetTextColor(warningRed[0], warningRed[1], warningRed[2])
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Tools: %s", toolsIcon, toolsStatus))
	pdf.Ln(8)

	// Resources capability
	resourcesIcon := crossMark
	resourcesStatus := notSupportedStatus
	if info.Capabilities.Resources {
		resourcesIcon = checkMark
		resourcesStatus = supportedStatus
		pdf.SetTextColor(successGreen[0], successGreen[1], successGreen[2])
	} else {
		pdf.SetTextColor(warningRed[0], warningRed[1], warningRed[2])
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Resources: %s", resourcesIcon, resourcesStatus))
	pdf.Ln(8)

	// Prompts capability
	promptsIcon := crossMark
	promptsStatus := notSupportedStatus
	if info.Capabilities.Prompts {
		promptsIcon = checkMark
		promptsStatus = supportedStatus
		pdf.SetTextColor(successGreen[0], successGreen[1], successGreen[2])
	} else {
		pdf.SetTextColor(warningRed[0], warningRed[1], warningRed[2])
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Prompts: %s", promptsIcon, promptsStatus))
	pdf.Ln(15)

	// Tools section
	if info.Capabilities.Tools && len(info.Tools) > 0 {
		pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Bookmark("Tools", 0, -1)
		pdf.Cell(0, 10, "Tools")
		pdf.Ln(10)

		for _, tool := range info.Tools {
			pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Bookmark(tool.Name, 1, -1)
			pdf.Cell(0, 8, tool.Name)
			pdf.Ln(8)

			pdf.SetTextColor(64, 64, 64)
			pdf.SetFont("DejaVuSans", "", 10)
			if tool.Description != "" {
				pdf.Cell(0, 6, tool.Description)
				pdf.Ln(6)
			}

			if tool.InputSchema != nil {
				pdf.Cell(0, 6, "Input Schema:")
				pdf.Ln(6)
				renderJSONSchema(pdf, tool.InputSchema)
			}

			if len(tool.Context) > 0 {
				pdf.Cell(0, 6, "Context:")
				pdf.Ln(6)
				renderContext(pdf, tool.Context)
			}

			pdf.Ln(8)
		}
	}

	// Resources section
	if info.Capabilities.Resources && len(info.Resources) > 0 {
		pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Bookmark("Resources", 0, 0)
		pdf.Cell(0, 10, "Resources")
		pdf.Ln(10)

		for _, resource := range info.Resources {
			pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Cell(0, 8, resource.Name)
			pdf.Ln(8)

			pdf.SetTextColor(64, 64, 64)
			pdf.SetFont("DejaVuSans", "", 10)
			pdf.Cell(0, 6, fmt.Sprintf("URI: %s", resource.URI))
			pdf.Ln(6)

			if resource.Description != "" {
				pdf.Cell(0, 6, resource.Description)
				pdf.Ln(6)
			}

			if resource.MimeType != "" {
				pdf.Cell(0, 6, fmt.Sprintf("MIME Type: %s", resource.MimeType))
				pdf.Ln(6)
			}

			if len(resource.Context) > 0 {
				pdf.Cell(0, 6, "Context:")
				pdf.Ln(6)
				renderContext(pdf, resource.Context)
			}

			pdf.Ln(8)
		}
	}

	// Prompts section
	if info.Capabilities.Prompts && len(info.Prompts) > 0 {
		pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Bookmark("Prompts", 0, 0)
		pdf.Cell(0, 10, "Prompts")
		pdf.Ln(10)

		for _, prompt := range info.Prompts {
			pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Cell(0, 8, prompt.Name)
			pdf.Ln(8)

			pdf.SetTextColor(64, 64, 64)
			pdf.SetFont("DejaVuSans", "", 10)
			if prompt.Description != "" {
				pdf.Cell(0, 6, prompt.Description)
				pdf.Ln(6)
			}

			if len(prompt.Arguments) > 0 {
				pdf.Cell(0, 6, "Arguments:")
				pdf.Ln(6)
				for _, arg := range prompt.Arguments {
					renderJSONSchema(pdf, arg)
				}
			}

			if len(prompt.Context) > 0 {
				pdf.Cell(0, 6, "Context:")
				pdf.Ln(6)
				renderContext(pdf, prompt.Context)
			}

			pdf.Ln(8)
		}
	}

	// Check for PDF generation errors before output
	if !pdf.Ok() {
		return nil, fmt.Errorf("PDF generation error: %w", pdf.Error())
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// renderJSONSchema renders a JSON schema with syntax highlighting in the PDF
func renderJSONSchema(pdf *fpdf.Fpdf, schema any) {
	pdf.SetFont("DejaVuSans", "", 9)

	// Convert to JSON string
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		pdf.Cell(0, 5, fmt.Sprintf("Error formatting schema: %v", err))
		pdf.Ln(5)
		return
	}

	// Process each line with syntax highlighting
	lines := strings.Split(string(schemaJSON), "\n")
	for _, line := range lines {
		// Limit line length to prevent overflow
		if len(line) > 100 {
			line = line[:97] + "..."
		}
		renderJSONLine(pdf, line)
		pdf.Ln(4)
	}
}

// renderContext renders context key-value pairs with proper formatting
func renderContext(pdf *fpdf.Fpdf, context map[string]string) {
	pdf.SetFont("DejaVuSans", "", 9)

	for key, value := range context {
		// Check if value contains newlines (block content)
		if strings.Contains(value, "\n") {
			// Render as block with key header
			pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
			pdf.Cell(0, 5, key+":")
			pdf.Ln(5)

			// Render block content with indentation
			pdf.SetTextColor(64, 64, 64)
			lines := strings.Split(value, "\n")
			inCodeBlock := false

			for _, line := range lines {
				// Handle different types of content
				trimmed := strings.TrimSpace(line)

				// Check for code block markers
				if strings.HasPrefix(trimmed, "```") {
					inCodeBlock = !inCodeBlock
					// Render code block marker with different formatting
					pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
					pdf.SetFont("DejaVuSans", "", 8)
					pdf.Cell(0, 4, "  "+line)
					pdf.Ln(4)
					continue
				}

				if inCodeBlock {
					// Inside code block - use monospace-like formatting
					pdf.SetTextColor(32, 32, 32) // Darker text for code
					pdf.SetFont("DejaVuSans", "", 8)
					// Add extra indentation for code content
					pdf.Cell(0, 4, "    "+line)
					pdf.Ln(4)
				} else {
					// Outside code block - normal formatting
					pdf.SetTextColor(64, 64, 64)
					pdf.SetFont("DejaVuSans", "", 9)

					switch {
					case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
						// List item - render with bullet
						pdf.Cell(0, 4, "  "+bulletPoint+" "+trimmed[2:])
					case strings.Contains(trimmed, ". ") && len(trimmed) > 2:
						// Numbered list item - find the ". " and skip past it
						if dotIndex := strings.Index(trimmed, ". "); dotIndex > 0 {
							listContent := trimmed[dotIndex+2:]
							pdf.Cell(0, 4, "  "+bulletPoint+" "+listContent)
						} else {
							pdf.Cell(0, 4, "  "+line)
						}
					case trimmed == "":
						// Empty line
						pdf.Cell(0, 4, "")
					default:
						// Regular content line
						pdf.Cell(0, 4, "  "+line)
					}
					pdf.Ln(4)
				}
			}
		} else {
			// Single line - render as bullet point
			pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
			pdf.SetFont("DejaVuSans", "", 9)
			pdf.Cell(0, 5, fmt.Sprintf("  %s %s: %s", bulletPoint, key, value))
			pdf.Ln(5)
		}
	}
}

// renderJSONLine renders a single line of JSON with syntax highlighting
func renderJSONLine(pdf *fpdf.Fpdf, line string) {
	lineHeight := 4.0

	i := 0
	for i < len(line) {
		char := line[i]

		// Handle whitespace - preserve indentation
		if char == ' ' || char == '\t' {
			spaces := ""
			for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
				spaces += string(line[i])
				i++
			}
			pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
			width := pdf.GetStringWidth(spaces)
			pdf.CellFormat(width, lineHeight, spaces, "", 0, "L", false, 0, "")
			continue
		}

		// Handle structural characters
		if char == '{' || char == '}' || char == '[' || char == ']' || char == ',' || char == ':' {
			pdf.SetTextColor(jsonBrace[0], jsonBrace[1], jsonBrace[2])
			width := pdf.GetStringWidth(string(char))
			pdf.CellFormat(width, lineHeight, string(char), "", 0, "L", false, 0, "")
			i++
			continue
		}

		// Handle quoted strings (keys and string values)
		if char == '"' {
			quote, nextIndex := extractJSONQuotedString(line, i)

			// Determine if this is a key or value by checking what follows
			isKey := false
			for j := nextIndex; j < len(line); j++ {
				if line[j] == ':' {
					isKey = true
					break
				}
				if line[j] != ' ' && line[j] != '\t' {
					break
				}
			}

			if isKey {
				pdf.SetTextColor(jsonKey[0], jsonKey[1], jsonKey[2])
			} else {
				pdf.SetTextColor(jsonString[0], jsonString[1], jsonString[2])
			}

			width := pdf.GetStringWidth(quote)
			pdf.CellFormat(width, lineHeight, quote, "", 0, "L", false, 0, "")
			i = nextIndex
			continue
		}

		// Handle numbers, booleans, and null
		if char >= '0' && char <= '9' || char == '-' || char == '.' {
			number, nextIndex := extractJSONNumber(line, i)
			pdf.SetTextColor(jsonNumber[0], jsonNumber[1], jsonNumber[2])
			width := pdf.GetStringWidth(number)
			pdf.CellFormat(width, lineHeight, number, "", 0, "L", false, 0, "")
			i = nextIndex
			continue
		}

		if strings.HasPrefix(line[i:], "true") || strings.HasPrefix(line[i:], "false") {
			var word string
			if strings.HasPrefix(line[i:], "true") {
				word = "true"
			} else {
				word = "false"
			}
			pdf.SetTextColor(jsonBoolean[0], jsonBoolean[1], jsonBoolean[2])
			width := pdf.GetStringWidth(word)
			pdf.CellFormat(width, lineHeight, word, "", 0, "L", false, 0, "")
			i += len(word)
			continue
		}

		if strings.HasPrefix(line[i:], "null") {
			pdf.SetTextColor(jsonNull[0], jsonNull[1], jsonNull[2])
			width := pdf.GetStringWidth("null")
			pdf.CellFormat(width, lineHeight, "null", "", 0, "L", false, 0, "")
			i += 4
			continue
		}

		// Default: render as regular text
		pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
		width := pdf.GetStringWidth(string(char))
		pdf.CellFormat(width, lineHeight, string(char), "", 0, "L", false, 0, "")
		i++
	}
}
