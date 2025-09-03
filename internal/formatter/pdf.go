package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

const (
	supportedStatus    = "Supported"
	notSupportedStatus = "Not supported"
)

// FormatPDF formats server info as PDF
func FormatPDF(info *model.ServerInfo, includeTOC bool) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)

	// Title
	title := info.Name
	if info.Version != "" {
		title += fmt.Sprintf(" (v%s)", info.Version)
	}
	pdf.Cell(0, 10, title)
	pdf.Ln(15)

	// Table of Contents
	if includeTOC {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Table of Contents")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 8, "• Capabilities")
		pdf.Ln(6)

		if info.Capabilities.Tools && len(info.Tools) > 0 {
			pdf.Cell(0, 8, "• Tools")
			pdf.Ln(6)
		}

		if info.Capabilities.Resources && len(info.Resources) > 0 {
			pdf.Cell(0, 8, "• Resources")
			pdf.Ln(6)
		}

		if info.Capabilities.Prompts && len(info.Prompts) > 0 {
			pdf.Cell(0, 8, "• Prompts")
			pdf.Ln(6)
		}

		pdf.Ln(10)
	}

	// Capabilities section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Capabilities")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	toolsStatus := notSupportedStatus
	if info.Capabilities.Tools {
		toolsStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("Tools: %s", toolsStatus))
	pdf.Ln(6)

	resourcesStatus := notSupportedStatus
	if info.Capabilities.Resources {
		resourcesStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("Resources: %s", resourcesStatus))
	pdf.Ln(6)

	promptsStatus := notSupportedStatus
	if info.Capabilities.Prompts {
		promptsStatus = supportedStatus
	}
	pdf.Cell(0, 8, fmt.Sprintf("Prompts: %s", promptsStatus))
	pdf.Ln(15)

	// Tools section
	if info.Capabilities.Tools && len(info.Tools) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Tools")
		pdf.Ln(10)

		for _, tool := range info.Tools {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, tool.Name)
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 10)
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
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Resources")
		pdf.Ln(10)

		for _, resource := range info.Resources {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, resource.Name)
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 10)
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
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Prompts")
		pdf.Ln(10)

		for _, prompt := range info.Prompts {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, prompt.Name)
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 10)
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
