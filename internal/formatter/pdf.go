package formatter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

//go:embed DejaVuSans.ttf
var dejaVuSansFont []byte

// Regex for numbered lists (1. 2. 10. etc.)
var numberedListRegex = regexp.MustCompile(`^\s*\d+\.\s+(.*)`)

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
	pdf := initializePDF()

	addPDFTitle(pdf, info)

	if includeTOC {
		addTableOfContents(pdf, info)
	}

	addCapabilitiesSection(pdf, info)
	addToolsSection(pdf, info)
	addResourcesSection(pdf, info)
	addPromptsSection(pdf, info)

	return finalizePDF(pdf)
}

// initializePDF creates and initializes a new PDF document
func initializePDF() *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8FontFromBytes("DejaVuSans", "", dejaVuSansFont)
	return pdf
}

// addPDFTitle adds the title section to the PDF
func addPDFTitle(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 20)

	title := info.Name
	if info.Version != "" {
		title += fmt.Sprintf(" (v%s)", info.Version)
	}
	pdf.Cell(0, 12, title)
	pdf.Ln(20)

	pdf.SetDrawColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetLineWidth(0.8)
	pdf.Line(10, pdf.GetY()-5, 200, pdf.GetY()-5)
	pdf.Ln(5)
}

// addTableOfContents adds the table of contents section
func addTableOfContents(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 16)
	pdf.Cell(0, 10, "Table of Contents")
	pdf.Ln(12)

	pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.Rect(10, pdf.GetY(), 190, 40, "F")
	pdf.Ln(5)

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

// addCapabilitiesSection adds the capabilities section to the PDF
func addCapabilitiesSection(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 16)
	pdf.Bookmark("Capabilities", 0, -1)
	pdf.Cell(0, 12, "Capabilities")
	pdf.Ln(15)

	pdf.SetDrawColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY()-5, 200, pdf.GetY()-5)
	pdf.Ln(5)

	pdf.SetFont("DejaVuSans", "", 12)

	addCapabilityLine(pdf, "Tools", info.Capabilities.Tools)
	addCapabilityLine(pdf, "Resources", info.Capabilities.Resources)
	addCapabilityLine(pdf, "Prompts", info.Capabilities.Prompts)

	pdf.Ln(15)
}

// addCapabilityLine adds a single capability line with status
func addCapabilityLine(pdf *fpdf.Fpdf, name string, supported bool) {
	var icon, status string
	if supported {
		icon = checkMark
		status = supportedStatus
		pdf.SetTextColor(successGreen[0], successGreen[1], successGreen[2])
	} else {
		icon = crossMark
		status = notSupportedStatus
		pdf.SetTextColor(warningRed[0], warningRed[1], warningRed[2])
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s %s: %s", icon, name, status))
	pdf.Ln(8)
}

// addToolsSection adds the tools section to the PDF
func addToolsSection(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	if !info.Capabilities.Tools || len(info.Tools) == 0 {
		return
	}

	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 14)
	pdf.Bookmark("Tools", 0, -1)
	pdf.Cell(0, 10, "Tools")
	pdf.Ln(10)

	for _, tool := range info.Tools {
		renderTool(pdf, tool)
	}
}

// renderTool renders a single tool in the PDF
func renderTool(pdf *fpdf.Fpdf, tool model.Tool) {
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

// addResourcesSection adds the resources section to the PDF
func addResourcesSection(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	if !info.Capabilities.Resources || len(info.Resources) == 0 {
		return
	}

	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 14)
	pdf.Bookmark("Resources", 0, 0)
	pdf.Cell(0, 10, "Resources")
	pdf.Ln(10)

	for _, resource := range info.Resources {
		renderResource(pdf, resource)
	}
}

// renderResource renders a single resource in the PDF
func renderResource(pdf *fpdf.Fpdf, resource model.Resource) {
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

// addPromptsSection adds the prompts section to the PDF
func addPromptsSection(pdf *fpdf.Fpdf, info *model.ServerInfo) {
	if !info.Capabilities.Prompts || len(info.Prompts) == 0 {
		return
	}

	pdf.SetTextColor(primaryBlue[0], primaryBlue[1], primaryBlue[2])
	pdf.SetFont("DejaVuSans", "", 14)
	pdf.Bookmark("Prompts", 0, 0)
	pdf.Cell(0, 10, "Prompts")
	pdf.Ln(10)

	for _, prompt := range info.Prompts {
		renderPrompt(pdf, prompt)
	}
}

// renderPrompt renders a single prompt in the PDF
func renderPrompt(pdf *fpdf.Fpdf, prompt model.Prompt) {
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

// finalizePDF completes the PDF generation and returns the bytes
func finalizePDF(pdf *fpdf.Fpdf) ([]byte, error) {
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

	// Sort keys to ensure deterministic output
	keys := make([]string, 0, len(context))
	for key := range context {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := context[key]
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
						if len(trimmed) > 2 {
							pdf.Cell(0, 4, "  "+bulletPoint+" "+trimmed[2:])
						} else {
							pdf.Cell(0, 4, "  "+bulletPoint)
						}
					case numberedListRegex.MatchString(trimmed):
						// Numbered list item - extract content after number
						if matches := numberedListRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
							listContent := matches[1]
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
	const lineHeight = 4.0
	i := 0

	for i < len(line) {
		char := line[i]

		switch {
		case char == ' ' || char == '\t':
			i = renderWhitespace(pdf, line, i, lineHeight)
		case isJSONStructuralChar(char):
			i = renderStructuralChar(pdf, char, lineHeight)
		case char == '"':
			i = renderQuotedStringPDF(pdf, line, i, lineHeight)
		case isJSONNumber(char):
			i = renderNumberPDF(pdf, line, i, lineHeight)
		case startsWithBoolean(line, i):
			i = renderBooleanPDF(pdf, line, i, lineHeight)
		case strings.HasPrefix(line[i:], "null"):
			i = renderNullPDF(pdf, lineHeight)
		default:
			i = renderDefaultChar(pdf, char, lineHeight)
		}
	}
}

// renderWhitespace renders whitespace characters and returns next index
func renderWhitespace(pdf *fpdf.Fpdf, line string, i int, lineHeight float64) int {
	spaces := ""
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		spaces += string(line[i])
		i++
	}
	pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
	width := pdf.GetStringWidth(spaces)
	pdf.CellFormat(width, lineHeight, spaces, "", 0, "L", false, 0, "")
	return i
}

// isJSONStructuralChar checks if character is a JSON structural character
func isJSONStructuralChar(char byte) bool {
	return char == '{' || char == '}' || char == '[' || char == ']' || char == ',' || char == ':'
}

// renderStructuralChar renders JSON structural characters
func renderStructuralChar(pdf *fpdf.Fpdf, char byte, lineHeight float64) int {
	pdf.SetTextColor(jsonBrace[0], jsonBrace[1], jsonBrace[2])
	width := pdf.GetStringWidth(string(char))
	pdf.CellFormat(width, lineHeight, string(char), "", 0, "L", false, 0, "")
	return 1
}

// renderQuotedStringPDF renders quoted strings with appropriate coloring
func renderQuotedStringPDF(pdf *fpdf.Fpdf, line string, i int, lineHeight float64) int {
	quote, nextIndex := extractJSONQuotedString(line, i)
	isKey := isPDFJSONKey(line, nextIndex)

	if isKey {
		pdf.SetTextColor(jsonKey[0], jsonKey[1], jsonKey[2])
	} else {
		pdf.SetTextColor(jsonString[0], jsonString[1], jsonString[2])
	}

	width := pdf.GetStringWidth(quote)
	pdf.CellFormat(width, lineHeight, quote, "", 0, "L", false, 0, "")
	return nextIndex
}

// isPDFJSONKey determines if a quoted string is a JSON key
func isPDFJSONKey(line string, startIndex int) bool {
	for j := startIndex; j < len(line); j++ {
		if line[j] == ':' {
			return true
		}
		if line[j] != ' ' && line[j] != '\t' {
			break
		}
	}
	return false
}

// isJSONNumber checks if character can start a JSON number
func isJSONNumber(char byte) bool {
	return (char >= '0' && char <= '9') || char == '-' || char == '.'
}

// renderNumberPDF renders JSON numbers
func renderNumberPDF(pdf *fpdf.Fpdf, line string, i int, lineHeight float64) int {
	number, nextIndex := extractJSONNumber(line, i)
	pdf.SetTextColor(jsonNumber[0], jsonNumber[1], jsonNumber[2])
	width := pdf.GetStringWidth(number)
	pdf.CellFormat(width, lineHeight, number, "", 0, "L", false, 0, "")
	return nextIndex
}

// startsWithBoolean checks if the current position starts with a boolean
func startsWithBoolean(line string, i int) bool {
	return strings.HasPrefix(line[i:], "true") || strings.HasPrefix(line[i:], "false")
}

// renderBooleanPDF renders boolean values
func renderBooleanPDF(pdf *fpdf.Fpdf, line string, i int, lineHeight float64) int {
	var word string
	if strings.HasPrefix(line[i:], "true") {
		word = "true"
	} else {
		word = "false"
	}
	pdf.SetTextColor(jsonBoolean[0], jsonBoolean[1], jsonBoolean[2])
	width := pdf.GetStringWidth(word)
	pdf.CellFormat(width, lineHeight, word, "", 0, "L", false, 0, "")
	return i + len(word)
}

// renderNullPDF renders null values
func renderNullPDF(pdf *fpdf.Fpdf, lineHeight float64) int {
	pdf.SetTextColor(jsonNull[0], jsonNull[1], jsonNull[2])
	width := pdf.GetStringWidth("null")
	pdf.CellFormat(width, lineHeight, "null", "", 0, "L", false, 0, "")
	return 4
}

// renderDefaultChar renders default characters
func renderDefaultChar(pdf *fpdf.Fpdf, char byte, lineHeight float64) int {
	pdf.SetTextColor(textGray[0], textGray[1], textGray[2])
	width := pdf.GetStringWidth(string(char))
	pdf.CellFormat(width, lineHeight, string(char), "", 0, "L", false, 0, "")
	return 1
}
