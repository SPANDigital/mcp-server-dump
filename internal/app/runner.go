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
	session, err := createMCPSession(cli)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			log.Printf("Warning: failed to close session: %v", closeErr)
		}
	}()

	info, err := collectServerInfo(session)
	if err != nil {
		return err
	}

	if err := applyContextConfig(info, cli.ContextFile); err != nil {
		return err
	}

	output, err := formatOutput(info, cli)
	if err != nil {
		return err
	}

	return writeOutput(output, cli.Output)
}

// createMCPSession creates and connects an MCP session
func createMCPSession(cli *CLI) (*mcp.ClientSession, error) {
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
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	mcpClient := mcp.NewClient(
		&mcp.Implementation{
			Name:    "mcp-server-dump",
			Version: GetVersion(),
		},
		nil,
	)

	ctx := context.Background()
	session, err := mcpClient.Connect(ctx, mcpTransport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	return session, nil
}

// collectServerInfo gathers server information including tools, resources, and prompts
func collectServerInfo(session *mcp.ClientSession) (*model.ServerInfo, error) {
	ctx := context.Background()
	initResult := session.InitializeResult()

	info := &model.ServerInfo{
		Name:    initResult.ServerInfo.Name,
		Version: initResult.ServerInfo.Version,
		Capabilities: model.Capabilities{
			Tools:     initResult.Capabilities.Tools != nil,
			Resources: initResult.Capabilities.Resources != nil,
			Prompts:   initResult.Capabilities.Prompts != nil,
		},
	}

	if err := collectTools(session, ctx, initResult, info); err != nil {
		return nil, err
	}

	if err := collectResources(session, ctx, initResult, info); err != nil {
		return nil, err
	}

	if err := collectPrompts(session, ctx, initResult, info); err != nil {
		return nil, err
	}

	return info, nil
}

// collectTools retrieves and processes tools from the MCP server
func collectTools(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) error {
	if initResult.Capabilities.Tools == nil {
		return nil
	}

	toolsList, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Printf("Warning: Failed to list tools: %v", err)
		return nil
	}

	for _, tool := range toolsList.Tools {
		info.Tools = append(info.Tools, model.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}

	return nil
}

// collectResources retrieves and processes resources from the MCP server
func collectResources(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) error {
	if initResult.Capabilities.Resources == nil {
		return nil
	}

	resourcesList, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		log.Printf("Warning: Failed to list resources: %v", err)
		return nil
	}

	for _, resource := range resourcesList.Resources {
		info.Resources = append(info.Resources, model.Resource{
			URI:         resource.URI,
			Name:        resource.Name,
			Description: resource.Description,
			MimeType:    resource.MIMEType,
		})
	}

	return nil
}

// collectPrompts retrieves and processes prompts from the MCP server
func collectPrompts(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) error {
	if initResult.Capabilities.Prompts == nil {
		return nil
	}

	promptsList, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		log.Printf("Warning: Failed to list prompts: %v", err)
		return nil
	}

	for _, prompt := range promptsList.Prompts {
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

	return nil
}

// applyContextConfig applies context configuration to server info if context files are provided
func applyContextConfig(info *model.ServerInfo, contextFiles []string) error {
	if len(contextFiles) == 0 {
		return nil
	}

	contextConfig, err := model.LoadContextConfig(contextFiles)
	if err != nil {
		return fmt.Errorf("failed to load context configuration: %w", err)
	}

	for i := range info.Tools {
		contextConfig.ApplyToTool(&info.Tools[i])
	}
	for i := range info.Resources {
		contextConfig.ApplyToResource(&info.Resources[i])
	}
	for i := range info.Prompts {
		contextConfig.ApplyToPrompt(&info.Prompts[i])
	}

	return nil
}

// formatOutput formats the server info according to the specified format
func formatOutput(info *model.ServerInfo, cli *CLI) ([]byte, error) {
	switch cli.Format {
	case "json":
		return formatter.FormatJSON(info)
	case "html":
		htmlStr, err := formatter.FormatHTML(info, !cli.NoTOC, TemplateFS)
		if err != nil {
			return nil, err
		}
		return []byte(htmlStr), nil
	case "pdf":
		if cli.Output == "" {
			return nil, fmt.Errorf("PDF format requires --output flag")
		}
		return formatter.FormatPDF(info, !cli.NoTOC)
	case "markdown":
		customFields := formatter.ParseCustomFields(cli.FrontmatterField)
		markdownStr, err := formatter.FormatMarkdown(info, !cli.NoTOC, cli.Frontmatter, cli.FrontmatterFormat, customFields, TemplateFS)
		if err != nil {
			return nil, err
		}
		return []byte(markdownStr), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", cli.Format)
	}
}

// writeOutput writes the formatted output to file or stdout
func writeOutput(output []byte, outputPath string) error {
	if outputPath != "" {
		return os.WriteFile(outputPath, output, 0o600)
	}

	_, err := os.Stdout.Write(output)
	return err
}
