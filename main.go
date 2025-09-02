package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "dev"     //nolint:unused // set by goreleaser
	commit  = "none"    //nolint:unused // set by goreleaser
	date    = "unknown" //nolint:unused // set by goreleaser
)

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
	defer session.Close()

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
		output = formatMarkdown(&info)
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

func formatMarkdown(info *ServerInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", info.Name))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", info.Version))

	sb.WriteString("## Capabilities\n\n")
	sb.WriteString(fmt.Sprintf("- Tools: %v\n", info.Capabilities.Tools))
	sb.WriteString(fmt.Sprintf("- Resources: %v\n", info.Capabilities.Resources))
	sb.WriteString(fmt.Sprintf("- Prompts: %v\n\n", info.Capabilities.Prompts))

	if len(info.Tools) > 0 {
		sb.WriteString("## Tools\n\n")
		for _, tool := range info.Tools {
			sb.WriteString(fmt.Sprintf("### %s\n\n", tool.Name))
			if tool.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", tool.Description))
			}
			if tool.InputSchema != nil {
				schemaJSON, _ := json.MarshalIndent(tool.InputSchema, "", "  ")
				sb.WriteString("**Input Schema:**\n```json\n")
				sb.WriteString(string(schemaJSON))
				sb.WriteString("\n```\n\n")
			}
		}
	}

	if len(info.Resources) > 0 {
		sb.WriteString("## Resources\n\n")
		for _, resource := range info.Resources {
			sb.WriteString(fmt.Sprintf("### %s\n\n", resource.Name))
			sb.WriteString(fmt.Sprintf("**URI:** `%s`\n\n", resource.URI))
			if resource.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", resource.Description))
			}
			if resource.MimeType != "" {
				sb.WriteString(fmt.Sprintf("**MIME Type:** %s\n\n", resource.MimeType))
			}
		}
	}

	if len(info.Prompts) > 0 {
		sb.WriteString("## Prompts\n\n")
		for _, prompt := range info.Prompts {
			sb.WriteString(fmt.Sprintf("### %s\n\n", prompt.Name))
			if prompt.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", prompt.Description))
			}
			if prompt.Arguments != nil {
				argsJSON, _ := json.MarshalIndent(prompt.Arguments, "", "  ")
				sb.WriteString("**Arguments:**\n```json\n")
				sb.WriteString(string(argsJSON))
				sb.WriteString("\n```\n\n")
			}
		}
	}

	return sb.String()
}
