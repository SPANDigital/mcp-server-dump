package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-pdf/fpdf"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

var (
	version = "dev"     // set by goreleaser
	commit  = "none"    //nolint:unused // set by goreleaser
	date    = "unknown" //nolint:unused // set by goreleaser
)

// getVersion returns the version string, attempting to get it from VCS info if available
func getVersion() string {
	// If version was set by goreleaser, use it
	if version != "dev" {
		return version
	}

	// Try to get version from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				// Use short commit hash (first 8 characters)
				if len(setting.Value) >= 8 {
					return "dev-" + setting.Value[:8]
				}
				return "dev-" + setting.Value
			}
		}
	}

	// Fallback to original behavior
	return version
}

//go:embed templates/*.tmpl
var templateFS embed.FS

type CLI struct {
	// Version flag
	Version kong.VersionFlag `kong:"short='v',help='Show version information'"`

	// Output options
	Output string `kong:"short='o',help='Output file for documentation (defaults to stdout)'"`
	Format string `kong:"short='f',default='markdown',enum='markdown,json,html,pdf',help='Output format'"`
	NoTOC  bool   `kong:"help='Disable table of contents in markdown output'"`

	// Frontmatter options
	Frontmatter       bool     `kong:"short='F',help='Include frontmatter in markdown output'"`
	FrontmatterField  []string `kong:"short='M',help='Add custom frontmatter field (format: key:value), can be used multiple times'"`
	FrontmatterFormat string   `kong:"default='yaml',enum='yaml,toml,json',help='Frontmatter format'"`

	// Transport selection
	Transport string `kong:"short='t',default='command',enum='command,sse,streamable',help='Transport type'"`

	// Transport-specific options
	Endpoint      string        `kong:"help='HTTP endpoint for SSE/Streamable transports'"`
	Timeout       time.Duration `kong:"default='30s',help='HTTP timeout for SSE/Streamable transports'"`
	Headers       []string      `kong:"short='H',help='HTTP headers for SSE/Streamable transports (format: Key:Value)'"`
	ServerCommand string        `kong:"help='Server command for explicit command transport'"`

	// Legacy command format (backward compatibility)
	Args []string `kong:"arg,optional,help='Command and arguments (legacy format for backward compatibility)'"`
}

type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
	Tools        []Tool       `json:"tools"`
	Resources    []Resource   `json:"resources"`
	Prompts      []Prompt     `json:"prompts"`
}

type Capabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   any    `json:"arguments,omitempty"`
}

// HeaderRoundTripper wraps an http.RoundTripper to add custom headers
type HeaderRoundTripper struct {
	transport http.RoundTripper
	headers   map[string]string
}

// RoundTrip implements http.RoundTripper
func (h *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// Add custom headers
	for key, value := range h.headers {
		newReq.Header.Set(key, value)
	}

	return h.transport.RoundTrip(newReq)
}

// ContentTypeFixingTransport normalizes content types for MCP compatibility
type ContentTypeFixingTransport struct {
	transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (c *ContentTypeFixingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Fix content type by removing charset parameter from JSON responses
	// This works around a bug in the MCP Go SDK that doesn't properly parse
	// content types with charset parameters
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		if strings.HasPrefix(contentType, "application/json;") {
			resp.Header.Set("Content-Type", "application/json")
		}
	}

	return resp, err
}

// parseHeaders converts header strings in "Key:Value" format to a map
func parseHeaders(headerStrings []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, header := range headerStrings {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s (expected Key:Value)", header)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty header key in: %s", header)
		}
		headers[key] = value
	}
	return headers, nil
}

// createTransport creates an appropriate MCP transport based on CLI options
func createTransport(cli *CLI) (mcp.Transport, error) {
	// Handle backward compatibility: if Args provided but no explicit transport, use command
	if len(cli.Args) > 0 && cli.Transport == "command" {
		if len(cli.Args) == 0 {
			return nil, fmt.Errorf("no command specified")
		}
		// #nosec G204 - Command and args are provided by user intentionally
		cmd := exec.Command(cli.Args[0], cli.Args[1:]...)
		return &mcp.CommandTransport{Command: cmd}, nil
	}

	switch cli.Transport {
	case "command":
		if cli.ServerCommand == "" && len(cli.Args) == 0 {
			return nil, fmt.Errorf("command transport requires either --server-command or legacy args")
		}

		var cmd *exec.Cmd
		if cli.ServerCommand != "" {
			// Parse the command string - simple space splitting for now
			parts := strings.Fields(cli.ServerCommand)
			if len(parts) == 0 {
				return nil, fmt.Errorf("empty server command")
			}
			// #nosec G204 - Command is provided by user intentionally
			cmd = exec.Command(parts[0], parts[1:]...)
		} else {
			// Use legacy args format
			// #nosec G204 - Command and args are provided by user intentionally
			cmd = exec.Command(cli.Args[0], cli.Args[1:]...)
		}
		return &mcp.CommandTransport{Command: cmd}, nil

	case "sse":
		if cli.Endpoint == "" {
			return nil, fmt.Errorf("SSE transport requires --endpoint")
		}
		httpClient := &http.Client{Timeout: cli.Timeout}

		// Add custom headers if specified
		if len(cli.Headers) > 0 {
			headers, err := parseHeaders(cli.Headers)
			if err != nil {
				return nil, fmt.Errorf("failed to parse headers: %w", err)
			}
			httpClient.Transport = &HeaderRoundTripper{
				transport: http.DefaultTransport,
				headers:   headers,
			}
		}

		return &mcp.SSEClientTransport{
			Endpoint:   cli.Endpoint,
			HTTPClient: httpClient,
		}, nil

	case "streamable":
		if cli.Endpoint == "" {
			return nil, fmt.Errorf("streamable transport requires --endpoint")
		}
		httpClient := &http.Client{Timeout: cli.Timeout}

		// Build transport chain: base -> headers -> content type fix
		var transport http.RoundTripper = http.DefaultTransport

		// Add custom headers if specified
		if len(cli.Headers) > 0 {
			headers, err := parseHeaders(cli.Headers)
			if err != nil {
				return nil, fmt.Errorf("failed to parse headers: %w", err)
			}
			transport = &HeaderRoundTripper{
				transport: transport,
				headers:   headers,
			}
		}

		// Always add content type fixing for streamable transport to work around
		// MCP Go SDK bug with charset parameters
		transport = &ContentTypeFixingTransport{
			transport: transport,
		}

		httpClient.Transport = transport

		return &mcp.StreamableClientTransport{
			Endpoint:   cli.Endpoint,
			HTTPClient: httpClient,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cli.Transport)
	}
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("mcp-server-dump"),
		kong.Description("Dump MCP server capabilities and documentation"),
		kong.UsageOnError(),
		kong.Vars{
			"version": getVersion(),
		},
	)

	if err := run(&cli); err != nil {
		ctx.FatalIfErrorf(err)
	}
}

func run(cli *CLI) error {
	// Create transport based on CLI options
	transport, err := createTransport(cli)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Create MCP client
	mcpClient := mcp.NewClient(
		&mcp.Implementation{
			Name:    "mcp-server-dump",
			Version: "0.1.0",
		},
		nil,
	)

	// Connect to the server
	ctx := context.Background()
	session, err := mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close session: %v\n", err)
		}
	}()

	// Get server info from the initialization result
	initResult := session.InitializeResult()

	// Gather server information
	info := ServerInfo{
		Name:    initResult.ServerInfo.Name,
		Version: initResult.ServerInfo.Version,
		Capabilities: Capabilities{
			Tools:     initResult.Capabilities.Tools != nil,
			Resources: initResult.Capabilities.Resources != nil,
			Prompts:   initResult.Capabilities.Prompts != nil,
		},
	}

	// List tools if supported
	if initResult.Capabilities.Tools != nil {
		toolsList, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			log.Printf("Warning: Failed to list tools: %v", err)
		} else {
			for _, tool := range toolsList.Tools {
				info.Tools = append(info.Tools, Tool{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: tool.InputSchema,
				})
			}
		}
	}

	// List resources if supported
	if initResult.Capabilities.Resources != nil {
		resourcesList, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
		if err != nil {
			log.Printf("Warning: Failed to list resources: %v", err)
		} else {
			for _, resource := range resourcesList.Resources {
				info.Resources = append(info.Resources, Resource{
					URI:         resource.URI,
					Name:        resource.Name,
					Description: resource.Description,
					MimeType:    resource.MIMEType,
				})
			}
		}
	}

	// List prompts if supported
	if initResult.Capabilities.Prompts != nil {
		promptsList, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
		if err != nil {
			log.Printf("Warning: Failed to list prompts: %v", err)
		} else {
			for _, prompt := range promptsList.Prompts {
				info.Prompts = append(info.Prompts, Prompt{
					Name:        prompt.Name,
					Description: prompt.Description,
					Arguments:   prompt.Arguments,
				})
			}
		}
	}

	// Format output
	var output string
	var outputBytes []byte
	switch cli.Format {
	case "json":
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(data)
	case "markdown":
		// Parse custom fields if frontmatter is enabled
		var customFields map[string]any
		if cli.Frontmatter {
			customFields = parseCustomFields(cli.FrontmatterField)
		}
		formatted, err := formatMarkdown(&info, !cli.NoTOC, cli.Frontmatter, cli.FrontmatterFormat, customFields)
		if err != nil {
			return fmt.Errorf("failed to format markdown: %w", err)
		}
		output = formatted
	case "html":
		formatted, err := formatHTML(&info, !cli.NoTOC)
		if err != nil {
			return fmt.Errorf("failed to format HTML: %w", err)
		}
		output = formatted
	case "pdf":
		pdfBytes, err := formatPDF(&info, !cli.NoTOC)
		if err != nil {
			return fmt.Errorf("failed to format PDF: %w", err)
		}
		outputBytes = pdfBytes
	default:
		return fmt.Errorf("unknown format: %s", cli.Format)
	}

	// Write output
	if cli.Output != "" {
		var dataToWrite []byte
		if outputBytes != nil {
			dataToWrite = outputBytes
		} else {
			dataToWrite = []byte(output)
		}
		if err := os.WriteFile(cli.Output, dataToWrite, 0o600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Documentation written to %s\n", cli.Output)
	} else {
		if outputBytes != nil {
			return fmt.Errorf("PDF output requires an output file (-o flag)")
		}
		fmt.Print(output)
	}

	return nil
}

// anchorName converts a string to a URL-safe anchor name
func anchorName(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	s = reg.ReplaceAllString(s, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")
	// Trim hyphens from start and end
	s = strings.Trim(s, "-")
	return s
}

// jsonIndent formats any value as indented JSON
func jsonIndent(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseCustomFields parses key:value pairs from CLI arguments into a map
func parseCustomFields(fields []string) map[string]any {
	custom := make(map[string]any)
	for _, field := range fields {
		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Check if value contains comma-separated items (for arrays)
			if strings.Contains(value, ",") {
				items := strings.Split(value, ",")
				trimmedItems := make([]string, len(items))
				for i, item := range items {
					trimmedItems[i] = strings.TrimSpace(item)
				}
				custom[key] = trimmedItems
			} else {
				// Auto-detect types for single values
				if v, err := strconv.ParseBool(value); err == nil {
					custom[key] = v
				} else if v, err := strconv.Atoi(value); err == nil {
					custom[key] = v
				} else if v, err := strconv.ParseFloat(value, 64); err == nil {
					custom[key] = v
				} else {
					custom[key] = value // Keep as string
				}
			}
		}
	}
	return custom
}

// generateFrontmatter generates frontmatter in the specified format
func generateFrontmatter(info *ServerInfo, format string, customFields map[string]any) (string, error) {
	// Build frontmatter data
	data := map[string]any{
		"title":        fmt.Sprintf("%s Documentation", info.Name),
		"generated_at": time.Now().Format(time.RFC3339),
		"generator":    "mcp-server-dump",
		"capabilities": map[string]bool{
			"tools":     info.Capabilities.Tools,
			"resources": info.Capabilities.Resources,
			"prompts":   info.Capabilities.Prompts,
		},
		"tools_count":     len(info.Tools),
		"resources_count": len(info.Resources),
		"prompts_count":   len(info.Prompts),
	}

	// Add version if present
	if info.Version != "" {
		data["version"] = info.Version
	}

	// Merge custom fields
	for k, v := range customFields {
		data[k] = v
	}

	switch format {
	case "yaml":
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal yaml frontmatter: %w", err)
		}
		return fmt.Sprintf("---\n%s---\n\n", string(yamlData)), nil

	case "toml":
		// Simple TOML generation for common types
		var buf bytes.Buffer
		buf.WriteString("+++\n")

		// Write simple fields first
		for k, v := range data {
			switch val := v.(type) {
			case string:
				fmt.Fprintf(&buf, "%s = %q\n", k, val)
			case int, int64, float64:
				fmt.Fprintf(&buf, "%s = %v\n", k, val)
			case bool:
				fmt.Fprintf(&buf, "%s = %v\n", k, val)
			case map[string]bool:
				if k == "capabilities" {
					buf.WriteString("\n[capabilities]\n")
					for ck, cv := range val {
						fmt.Fprintf(&buf, "%s = %v\n", ck, cv)
					}
				}
			}
		}
		buf.WriteString("+++\n\n")
		return buf.String(), nil

	case "json":
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal json frontmatter: %w", err)
		}
		return fmt.Sprintf("%s\n\n", string(jsonData)), nil

	default:
		return "", fmt.Errorf("unsupported frontmatter format: %s", format)
	}
}

func formatHTML(info *ServerInfo, includeTOC bool) (string, error) {
	// First generate markdown (no frontmatter for HTML output)
	markdown, err := formatMarkdown(info, includeTOC, false, "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to format markdown: %w", err)
	}

	// Configure goldmark with GitHub Flavored Markdown extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Footnote,    // Footnotes
			extension.Typographer, // Typographic substitutions
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Generate heading IDs
			parser.WithAttribute(),     // Support heading attributes
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // Convert line breaks to <br>
			html.WithUnsafe(),    // Allow raw HTML (needed for our templates)
		),
	)

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	return buf.String(), nil
}

func formatMarkdown(info *ServerInfo, includeTOC, includeFrontmatter bool, frontmatterFormat string, customFields map[string]any) (string, error) {
	// Create template data with TOC flag
	templateData := struct {
		*ServerInfo
		IncludeTOC bool
	}{
		ServerInfo: info,
		IncludeTOC: includeTOC,
	}

	// Define template functions
	funcMap := template.FuncMap{
		"anchor": anchorName,
		"json":   jsonIndent,
	}

	// Parse all templates
	tmpl, err := template.New("base.md.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse templates: %w", err)
	}

	// Execute the base template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	markdownContent := buf.String()

	// Prepend frontmatter if requested
	if includeFrontmatter {
		frontmatter, err := generateFrontmatter(info, frontmatterFormat, customFields)
		if err != nil {
			return "", fmt.Errorf("failed to generate frontmatter: %w", err)
		}
		markdownContent = frontmatter + markdownContent
	}

	return markdownContent, nil
}

// renderJSONSchema renders a JSON schema in the PDF with proper formatting and page breaks
func renderJSONSchema(pdf *fpdf.Fpdf, schema any, title string) {
	if schema == nil {
		return
	}

	if pdf.GetY() > 220 {
		pdf.AddPage()
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, title, "", 1, "L", false, 0, "")
	pdf.Ln(1)

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err == nil {
		pdf.SetFont("Courier", "", 8)
		// Split into lines and handle page breaks
		schemaLines := strings.Split(string(schemaJSON), "\n")
		for _, line := range schemaLines {
			if pdf.GetY() > 275 {
				pdf.AddPage()
			}
			if strings.TrimSpace(line) != "" {
				pdf.CellFormat(0, 4, line, "", 1, "L", false, 0, "")
			}
		}
	}
}

func formatPDF(info *ServerInfo, includeTOC bool) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)

	// Add first page
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 20, info.Name, "", 1, "C", false, 0, "")

	// Version
	if info.Version != "" {
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(0, 10, fmt.Sprintf("Version: %s", info.Version), "", 1, "C", false, 0, "")
	}
	pdf.Ln(10)

	// Table of Contents (if enabled)
	if includeTOC {
		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 12, "Table of Contents", "", 1, "L", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont("Arial", "", 11)

		// Capabilities
		pdf.CellFormat(20, 6, "* Capabilities", "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, "{cap_page}", "", 1, "R", false, 0, "")

		// Tools
		if len(info.Tools) > 0 {
			pdf.CellFormat(20, 6, "* Tools", "", 0, "L", false, 0, "")
			pdf.CellFormat(0, 6, "{tools_page}", "", 1, "R", false, 0, "")
		}

		// Resources
		if len(info.Resources) > 0 {
			pdf.CellFormat(20, 6, "* Resources", "", 0, "L", false, 0, "")
			pdf.CellFormat(0, 6, "{resources_page}", "", 1, "R", false, 0, "")
		}

		// Prompts
		if len(info.Prompts) > 0 {
			pdf.CellFormat(20, 6, "* Prompts", "", 0, "L", false, 0, "")
			pdf.CellFormat(0, 6, "{prompts_page}", "", 1, "R", false, 0, "")
		}

		pdf.Ln(15)
	}

	// Capabilities section
	if pdf.GetY() > 250 {
		pdf.AddPage()
	}

	// Register the page for TOC and add bookmark
	if includeTOC {
		pdf.RegisterAlias("{cap_page}", fmt.Sprintf("%d", pdf.PageNo()))
	}
	pdf.Bookmark("Capabilities", 0, 0)

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 12, "Capabilities", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetFont("Arial", "", 11)
	capabilities := []string{
		fmt.Sprintf("* Tools: %s", formatBool(info.Capabilities.Tools)),
		fmt.Sprintf("* Resources: %s", formatBool(info.Capabilities.Resources)),
		fmt.Sprintf("* Prompts: %s", formatBool(info.Capabilities.Prompts)),
	}

	for _, cap := range capabilities {
		pdf.CellFormat(0, 6, cap, "", 1, "L", false, 0, "")
	}
	pdf.Ln(10)

	// Tools section
	if len(info.Tools) > 0 {
		if pdf.GetY() > 240 {
			pdf.AddPage()
		}

		// Register the page for TOC and add bookmark
		if includeTOC {
			pdf.RegisterAlias("{tools_page}", fmt.Sprintf("%d", pdf.PageNo()))
		}
		pdf.Bookmark("Tools", 0, 0)

		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 12, "Tools", "", 1, "L", false, 0, "")
		pdf.Ln(5)

		for i, tool := range info.Tools {
			// Check if we need a new page
			if pdf.GetY() > 240 {
				pdf.AddPage()
			}

			// Tool name with bookmark
			pdf.Bookmark(tool.Name, 1, 0)
			pdf.SetFont("Arial", "B", 14)
			pdf.CellFormat(0, 10, tool.Name, "", 1, "L", false, 0, "")
			pdf.Ln(2)

			// Tool description
			if tool.Description != "" {
				pdf.SetFont("Arial", "", 10)
				pdf.MultiCell(0, 5, tool.Description, "", "L", false)
				pdf.Ln(3)
			}

			// Input schema
			renderJSONSchema(pdf, tool.InputSchema, "Input Schema:")

			// Add spacing between tools
			if i < len(info.Tools)-1 {
				pdf.Ln(8)
			}
		}
		pdf.Ln(10)
	}

	// Resources section
	if len(info.Resources) > 0 {
		if pdf.GetY() > 240 {
			pdf.AddPage()
		}

		// Register the page for TOC and add bookmark
		if includeTOC {
			pdf.RegisterAlias("{resources_page}", fmt.Sprintf("%d", pdf.PageNo()))
		}
		pdf.Bookmark("Resources", 0, 0)

		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 12, "Resources", "", 1, "L", false, 0, "")
		pdf.Ln(5)

		for i, resource := range info.Resources {
			if pdf.GetY() > 240 {
				pdf.AddPage()
			}

			// Resource name with bookmark
			pdf.Bookmark(resource.Name, 1, 0)
			pdf.SetFont("Arial", "B", 14)
			pdf.CellFormat(0, 10, resource.Name, "", 1, "L", false, 0, "")
			pdf.Ln(2)

			// URI
			if resource.URI != "" {
				pdf.SetFont("Arial", "B", 9)
				pdf.CellFormat(0, 5, "URI:", "", 1, "L", false, 0, "")
				pdf.SetFont("Courier", "", 9)
				pdf.MultiCell(0, 4, resource.URI, "", "L", false)
				pdf.Ln(2)
			}

			// Description
			if resource.Description != "" {
				pdf.SetFont("Arial", "", 10)
				pdf.MultiCell(0, 5, resource.Description, "", "L", false)
				pdf.Ln(2)
			}

			// MIME Type
			if resource.MimeType != "" {
				pdf.SetFont("Arial", "", 9)
				pdf.CellFormat(0, 5, fmt.Sprintf("MIME Type: %s", resource.MimeType), "", 1, "L", false, 0, "")
			}

			// Add spacing between resources
			if i < len(info.Resources)-1 {
				pdf.Ln(8)
			}
		}
		pdf.Ln(10)
	}

	// Prompts section
	if len(info.Prompts) > 0 {
		if pdf.GetY() > 240 {
			pdf.AddPage()
		}

		// Register the page for TOC and add bookmark
		if includeTOC {
			pdf.RegisterAlias("{prompts_page}", fmt.Sprintf("%d", pdf.PageNo()))
		}
		pdf.Bookmark("Prompts", 0, 0)

		pdf.SetFont("Arial", "B", 16)
		pdf.CellFormat(0, 12, "Prompts", "", 1, "L", false, 0, "")
		pdf.Ln(5)

		for i, prompt := range info.Prompts {
			if pdf.GetY() > 240 {
				pdf.AddPage()
			}

			// Prompt name with bookmark
			pdf.Bookmark(prompt.Name, 1, 0)
			pdf.SetFont("Arial", "B", 14)
			pdf.CellFormat(0, 10, prompt.Name, "", 1, "L", false, 0, "")
			pdf.Ln(2)

			// Description
			if prompt.Description != "" {
				pdf.SetFont("Arial", "", 10)
				pdf.MultiCell(0, 5, prompt.Description, "", "L", false)
				pdf.Ln(3)
			}

			// Arguments
			renderJSONSchema(pdf, prompt.Arguments, "Arguments:")

			// Add spacing between prompts
			if i < len(info.Prompts)-1 {
				pdf.Ln(8)
			}
		}
	}

	// Check for PDF generation errors
	if pdf.Error() != nil {
		return nil, fmt.Errorf("PDF generation error: %w", pdf.Error())
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// formatBool returns ✅ for true and ❌ for false
func formatBool(b bool) string {
	if b {
		return "✅ Supported"
	}
	return "❌ Not supported"
}
