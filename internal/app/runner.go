package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/spandigital/mcp-server-dump/internal/formatter"
	"github.com/spandigital/mcp-server-dump/internal/model"
	"github.com/spandigital/mcp-server-dump/internal/transport"
)

// Run executes the main application logic
func Run(cli *CLI) error {
	// Create transport
	transportConfig := transport.Config{
		Transport:     cli.Transport,
		Endpoint:      cli.Endpoint,
		Timeout:       cli.Timeout,
		Headers:       cli.Headers,
		ServerCommand: cli.ServerCommand,
		Args:          cli.Args,
	}

	mcpTransport, err := transport.Create(&transportConfig)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Create MCP client
	mcpClient := mcp.NewClient(
		&mcp.Implementation{
			Name:    "mcp-server-dump",
			Version: GetVersion(),
		},
		nil,
	)

	// Connect to the server
	ctx := context.Background()
	session, err := mcpClient.Connect(ctx, mcpTransport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			log.Printf("Warning: failed to close session: %v", closeErr)
		}
	}()

	// Get server info from the initialization result
	initResult := session.InitializeResult()

	// Build our ServerInfo structure
	info := &model.ServerInfo{
		Name:    initResult.ServerInfo.Name,
		Version: initResult.ServerInfo.Version,
		Capabilities: model.Capabilities{
			Tools:     initResult.Capabilities.Tools != nil,
			Resources: initResult.Capabilities.Resources != nil,
			Prompts:   initResult.Capabilities.Prompts != nil,
		},
	}

	// Get tools if supported
	if initResult.Capabilities.Tools != nil {
		toolsList, toolsErr := session.ListTools(ctx, &mcp.ListToolsParams{})
		if toolsErr != nil {
			log.Printf("Warning: Failed to list tools: %v", toolsErr)
		} else {
			for _, tool := range toolsList.Tools {
				info.Tools = append(info.Tools, model.Tool{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: tool.InputSchema,
				})
			}
		}
	}

	// Get resources if supported
	if initResult.Capabilities.Resources != nil {
		resourcesList, resourcesErr := session.ListResources(ctx, &mcp.ListResourcesParams{})
		if resourcesErr != nil {
			log.Printf("Warning: Failed to list resources: %v", resourcesErr)
		} else {
			for _, resource := range resourcesList.Resources {
				info.Resources = append(info.Resources, model.Resource{
					URI:         resource.URI,
					Name:        resource.Name,
					Description: resource.Description,
					MimeType:    resource.MIMEType,
				})
			}
		}
	}

	// Get prompts if supported
	if initResult.Capabilities.Prompts != nil {
		promptsList, promptsErr := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
		if promptsErr != nil {
			log.Printf("Warning: Failed to list prompts: %v", promptsErr)
		} else {
			for _, prompt := range promptsList.Prompts {
				// Convert prompt arguments to []any
				var args []any
				for _, arg := range prompt.Arguments {
					args = append(args, arg)
				}
				info.Prompts = append(info.Prompts, model.Prompt{
					Name:        prompt.Name,
					Description: prompt.Description,
					Arguments:   args,
				})
			}
		}
	}

	// Apply contexts if context files are provided
	if len(cli.ContextFile) > 0 {
		contextConfig, contextErr := model.LoadContextConfig(cli.ContextFile)
		if contextErr != nil {
			return fmt.Errorf("failed to load context configuration: %w", contextErr)
		}

		// Apply contexts to tools, resources, and prompts
		for i := range info.Tools {
			contextConfig.ApplyToTool(&info.Tools[i])
		}
		for i := range info.Resources {
			contextConfig.ApplyToResource(&info.Resources[i])
		}
		for i := range info.Prompts {
			contextConfig.ApplyToPrompt(&info.Prompts[i])
		}
	}

	// Format output
	var output []byte
	switch cli.Format {
	case "json":
		output, err = formatter.FormatJSON(info)
	case "html":
		htmlStr, htmlErr := formatter.FormatHTML(info, !cli.NoTOC, TemplateFS)
		if htmlErr != nil {
			err = htmlErr
		} else {
			output = []byte(htmlStr)
		}
	case "pdf":
		if cli.Output == "" {
			return fmt.Errorf("PDF format requires --output flag")
		}
		output, err = formatter.FormatPDF(info, !cli.NoTOC)
	case "markdown":
		customFields := formatter.ParseCustomFields(cli.FrontmatterField)
		markdownStr, markdownErr := formatter.FormatMarkdown(info, !cli.NoTOC, cli.Frontmatter, cli.FrontmatterFormat, customFields, TemplateFS)
		if markdownErr != nil {
			err = markdownErr
		} else {
			output = []byte(markdownStr)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", cli.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write output
	if cli.Output != "" {
		return os.WriteFile(cli.Output, output, 0o600)
	}

	_, err = os.Stdout.Write(output)
	return err
}
