package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "dev"     //nolint:unused // set by goreleaser
	commit  = "none"    //nolint:unused // set by goreleaser
	date    = "unknown" //nolint:unused // set by goreleaser
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type CLI struct {
	Output string `kong:"short='o',help='Output file for documentation (defaults to stdout)'"`
	Format string `kong:"short='f',default='markdown',enum='markdown,json',help='Output format'"`

	Command string   `kong:"arg,required,help='Command to execute the MCP server'"`
	Args    []string `kong:"arg,optional,help='Arguments to pass to the MCP server command'"`
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
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

type Prompt struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Arguments   interface{} `json:"arguments,omitempty"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("mcp-server-dump"),
		kong.Description("Dump MCP server capabilities and documentation"),
		kong.UsageOnError(),
	)

	if err := run(cli); err != nil {
		ctx.FatalIfErrorf(err)
	}
}

func run(cli CLI) error {
	// Create the command
	// #nosec G204 - Command and args are provided by user intentionally
	cmd := exec.Command(cli.Command, cli.Args...)

	// Create command transport
	transport := &mcp.CommandTransport{Command: cmd}

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
	switch cli.Format {
	case "json":
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(data)
	case "markdown":
		formatted, err := formatMarkdown(&info)
		if err != nil {
			return fmt.Errorf("failed to format markdown: %w", err)
		}
		output = formatted
	default:
		return fmt.Errorf("unknown format: %s", cli.Format)
	}

	// Write output
	if cli.Output != "" {
		if err := os.WriteFile(cli.Output, []byte(output), 0o600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Documentation written to %s\n", cli.Output)
	} else {
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

// jsonIndent formats an interface{} as indented JSON
func jsonIndent(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func formatMarkdown(info *ServerInfo) (string, error) {
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
	if err := tmpl.Execute(&buf, info); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
