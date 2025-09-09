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
	bulletPoint = "•"  // bullet U+2022
	checkMark   = "✓"  // checkmark U+2713
	crossMark   = "✗"  // cross mark U+2717
)

// FormatPDF formats server info as PDF
func FormatPDF(info *model.ServerInfo, includeTOC bool) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Load UTF-8 font for Unicode support
	pdf.AddUTF8FontFromBytes("DejaVuSans", "", dejaVuSansFont)

	// Set font (start with regular style to test font loading)
	pdf.SetFont("DejaVuSans", "", 16)

	// Title
	title := info.Name
	if info.Version != "" {
		title += fmt.Sprintf(" (v%s)", info.Version)
	}
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Table of Contents
	if includeTOC {
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Cell(0, 10, "Table of Contents")
		pdf.Ln(10)

		pdf.SetFont("DejaVuSans", "", 12)
		pdf.Cell(0, 8, bulletPoint+" Capabilities")
		pdf.Ln(6)

		if info.Capabilities.Tools && len(info.Tools) > 0 {
			pdf.Cell(0, 8, bulletPoint+" Tools")
			pdf.Ln(6)
		}

		if info.Capabilities.Resources && len(info.Resources) > 0 {
			pdf.Cell(0, 8, bulletPoint+" Resources")
			pdf.Ln(6)
		}

		if info.Capabilities.Prompts && len(info.Prompts) > 0 {
			pdf.Cell(0, 8, bulletPoint+" Prompts")
			pdf.Ln(6)
		}

		pdf.Ln(10)
	}

	// Capabilities section
	pdf.SetFont("DejaVuSans", "", 14)
	pdf.Cell(0, 10, "Capabilities")
	pdf.Ln(10)

	pdf.SetFont("DejaVuSans", "", 12)
	toolsIcon := crossMark
	toolsStatus := notSupportedStatus
	if info.Capabilities.Tools {
		toolsIcon = checkMark
		toolsStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Tools: %s", toolsIcon, toolsStatus))
	pdf.Ln(6)

	resourcesIcon := crossMark
	resourcesStatus := notSupportedStatus
	if info.Capabilities.Resources {
		resourcesIcon = checkMark
		resourcesStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Resources: %s", resourcesIcon, resourcesStatus))
	pdf.Ln(6)

	promptsIcon := crossMark
	promptsStatus := notSupportedStatus
	if info.Capabilities.Prompts {
		promptsIcon = checkMark
		promptsStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("%s Prompts: %s", promptsIcon, promptsStatus))
	pdf.Ln(15)

	// Tools section
	if info.Capabilities.Tools && len(info.Tools) > 0 {
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Cell(0, 10, "Tools")
		pdf.Ln(10)

		for _, tool := range info.Tools {
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Cell(0, 8, tool.Name)
			pdf.Ln(8)

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

			pdf.Ln(8)
		}
	}

	// Resources section
	if info.Capabilities.Resources && len(info.Resources) > 0 {
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Cell(0, 10, "Resources")
		pdf.Ln(10)

		for _, resource := range info.Resources {
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Cell(0, 8, resource.Name)
			pdf.Ln(8)

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

			pdf.Ln(8)
		}
	}

	// Prompts section
	if info.Capabilities.Prompts && len(info.Prompts) > 0 {
		pdf.SetFont("DejaVuSans", "", 14)
		pdf.Cell(0, 10, "Prompts")
		pdf.Ln(10)

		for _, prompt := range info.Prompts {
			pdf.SetFont("DejaVuSans", "", 12)
			pdf.Cell(0, 8, prompt.Name)
			pdf.Ln(8)

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

			pdf.Ln(8)
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// renderJSONSchema renders a JSON schema in the PDF
func renderJSONSchema(pdf *fpdf.Fpdf, schema any) {
	pdf.SetFont("Courier", "", 9)

	// Convert to JSON string
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		pdf.Cell(0, 5, fmt.Sprintf("Error formatting schema: %v", err))
		pdf.Ln(5)
		return
	}

	// Split into lines and add to PDF
	lines := strings.Split(string(schemaJSON), "\n")
	for _, line := range lines {
		// Limit line length to prevent overflow
		if len(line) > 100 {
			line = line[:97] + "..."
		}
		pdf.Cell(0, 4, line)
		pdf.Ln(4)
	}
}
