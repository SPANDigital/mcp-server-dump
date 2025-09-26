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

// Error messages
const (
	ErrAllScanTypesDisabled = "cannot disable all scan types: at least one of tools, resources, or prompts must be enabled"
)

// Run executes the main application logic
func Run(cli *CLI) error {
	// Validate that at least one scan type is enabled
	if err := cli.ValidateScanOptions(); err != nil {
		return err
	}

	ctx := context.Background()
	session, err := createMCPSession(ctx, cli)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			log.Printf("Warning: failed to close session: %v", closeErr)
		}
	}()

	info := collectServerInfo(session, cli)

	if contextErr := applyContextConfig(info, cli.ContextFile); contextErr != nil {
		return contextErr
	}

	output, err := formatOutput(info, cli)
	if err != nil {
		return err
	}

	return writeOutput(output, cli.Output)
}

// createMCPSession establishes a connection to the MCP server using the configured transport.
// It returns a client session for communicating with the server, or an error if connection fails.
// The provided context allows for connection timeout and cancellation control.
func createMCPSession(ctx context.Context, cli *CLI) (*mcp.ClientSession, error) {
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

	session, err := mcpClient.Connect(ctx, mcpTransport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	return session, nil
}

// collectServerInfo gathers basic server information and capabilities from the MCP server.
// It initializes the server info structure with name, version, and capability flags.
// The CLI flags control which types of data are actually collected.
func collectServerInfo(session *mcp.ClientSession, cli *CLI) *model.ServerInfo {
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

	// Conditionally collect data based on CLI flags
	if !cli.NoTools {
		collectTools(session, ctx, initResult, info)
	} else {
		log.Printf("Skipping tools collection")
	}
	if !cli.NoResources {
		collectResources(session, ctx, initResult, info)
	} else {
		log.Printf("Skipping resources collection")
	}
	if !cli.NoPrompts {
		collectPrompts(session, ctx, initResult, info)
	} else {
		log.Printf("Skipping prompts collection")
	}

	return info
}

// collectTools retrieves and processes tools from the MCP server
func collectTools(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) {
	if initResult.Capabilities.Tools == nil {
		return
	}

	toolsList, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Printf("Warning: Failed to list tools: %v", err)
		return
	}

	for _, tool := range toolsList.Tools {
		info.Tools = append(info.Tools, model.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
}

// collectResources retrieves and processes resources from the MCP server
func collectResources(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) {
	if initResult.Capabilities.Resources == nil {
		return
	}

	resourcesList, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
	if err != nil {
		log.Printf("Warning: Failed to list resources: %v", err)
		return
	}

	for _, resource := range resourcesList.Resources {
		info.Resources = append(info.Resources, model.Resource{
			URI:         resource.URI,
			Name:        resource.Name,
			Description: resource.Description,
			MimeType:    resource.MIMEType,
		})
	}
}

// collectPrompts retrieves and processes prompts from the MCP server
func collectPrompts(session *mcp.ClientSession, ctx context.Context, initResult *mcp.InitializeResult, info *model.ServerInfo) {
	if initResult.Capabilities.Prompts == nil {
		return
	}

	promptsList, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		log.Printf("Warning: Failed to list prompts: %v", err)
		return
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
}

// applyContextConfig applies context enhancement configuration from external YAML/JSON files.
// It merges context data to enrich tool, resource, and prompt descriptions with additional content.
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

// formatOutput converts the server information into the requested output format (markdown, HTML, JSON, PDF, or Hugo).
// It uses the appropriate formatter based on the CLI format specification and configuration options.
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
	case "hugo":
		if cli.Output == "" {
			return nil, fmt.Errorf("hugo format requires --output flag (directory path)")
		}
		customFields := formatter.ParseCustomFields(cli.FrontmatterField)
		// Enable frontmatter by default for Hugo format
		enableFrontmatter := cli.Frontmatter || cli.Format == "hugo"

		// Create Hugo configuration from CLI parameters
		hugoConfig := &formatter.HugoConfig{
			BaseURL:         cli.HugoBaseURL,
			LanguageCode:    cli.HugoLanguageCode,
			Theme:           cli.HugoTheme, // Deprecated when using Hugo modules
			Github:          cli.HugoGithub,
			Twitter:         cli.HugoTwitter,
			SiteLogo:        cli.HugoSiteLogo,
			GoogleAnalytics: cli.HugoGoogleAnalytics,
		}

		// Warn user about deprecated HugoTheme field
		if cli.HugoTheme != "" {
			log.Printf("⚠️  WARNING: --hugo-theme is deprecated. Hugo now uses modules with Hextra theme by default. The theme parameter will be ignored.")
		}

		err := formatter.FormatHugo(info, cli.Output, enableFrontmatter, cli.FrontmatterFormat, customFields, hugoConfig, cli.CustomInitialisms, HugoTemplateFS)
		if err != nil {
			return nil, err
		}
		// Return empty bytes since Hugo writes directly to files
		return []byte{}, nil
	case "markdown":
		customFields := formatter.ParseCustomFields(cli.FrontmatterField)
		markdownStr, err := formatter.FormatMarkdown(info, !cli.NoTOC, cli.Frontmatter, cli.FrontmatterFormat, customFields, cli.CustomInitialisms, TemplateFS)
		if err != nil {
			return nil, err
		}
		return []byte(markdownStr), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", cli.Format)
	}
}

// writeOutput writes the formatted content to the specified output destination.
// If outputPath is empty, content is written to stdout; otherwise to the specified file.
// For Hugo format, the output is already written to files, so this function does nothing.
func writeOutput(output []byte, outputPath string) error {
	// Hugo format writes directly to files, so skip if output is empty
	if len(output) == 0 {
		return nil
	}

	if outputPath != "" {
		return os.WriteFile(outputPath, output, 0o600)
	}

	_, err := os.Stdout.Write(output)
	return err
}
