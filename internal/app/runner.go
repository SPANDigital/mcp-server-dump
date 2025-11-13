package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/spandigital/mcp-server-dump/internal/auth"
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

	// Call tools if requested
	if toolErr := callTools(session, ctx, info, cli); toolErr != nil {
		return toolErr
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
//
//nolint:gocyclo // OAuth configuration logic requires multiple conditional branches
func createMCPSession(ctx context.Context, cli *CLI) (*mcp.ClientSession, error) {
	transportConfig := transport.Config{
		Transport:     cli.Transport,
		Endpoint:      cli.Endpoint,
		Timeout:       cli.Timeout,
		Headers:       cli.Headers,
		ServerCommand: cli.ServerCommand,
		Args:          cli.Args,
	}

	// Create OAuth config if client ID is provided or if endpoint requires OAuth
	var oauthConfig *auth.Config
	if cli.OAuthClientID != "" {
		// Validate that both auth and token URLs are provided if either is specified
		if (cli.OAuthAuthURL != "" || cli.OAuthTokenURL != "") &&
			(cli.OAuthAuthURL == "" || cli.OAuthTokenURL == "") {
			return nil, fmt.Errorf("both --oauth-auth-url and --oauth-token-url must be provided together")
		}

		// If auth/token URLs not provided, discover them automatically
		var authURL, tokenURL string
		if cli.OAuthAuthURL == "" && cli.OAuthTokenURL == "" {
			fmt.Printf("Discovering OAuth endpoints from %s...\n", cli.Endpoint)
			discoveredConfig, err := auth.DiscoverAndConfigure(ctx, cli.Endpoint)
			if err != nil {
				return nil, fmt.Errorf("failed to discover OAuth endpoints: %w", err)
			}
			if discoveredConfig == nil {
				return nil, fmt.Errorf("server does not advertise OAuth endpoints")
			}
			authURL = discoveredConfig.AuthURL
			tokenURL = discoveredConfig.TokenURL
			fmt.Printf("✓ Discovered OAuth endpoints\n")
			fmt.Printf("  Authorization URL: %s\n", authURL)
			fmt.Printf("  Token URL: %s\n", tokenURL)
		} else {
			// Use explicitly provided URLs
			authURL = cli.OAuthAuthURL
			tokenURL = cli.OAuthTokenURL
		}

		// Build OAuth configuration
		oauthConfig = &auth.Config{
			ClientID:     cli.OAuthClientID,
			ClientSecret: cli.OAuthClientSecret,
			Scopes:       cli.OAuthScopes,
			RedirectPort: cli.OAuthRedirectPort,
			ResourceURI:  cli.Endpoint, // MCP server endpoint is the resource URI
			UseCache:     !cli.OAuthNoCache,
			AuthURL:      authURL,
			TokenURL:     tokenURL,
		}

		// If scopes not specified, use defaults
		if len(oauthConfig.Scopes) == 0 {
			oauthConfig.Scopes = auth.DefaultScopes()
		}
	} else if cli.Transport != "command" && cli.Endpoint != "" {
		// For HTTP transports without explicit OAuth config, try discovery
		// This is a best-effort attempt - if it fails, we'll proceed without OAuth
		discoveredConfig, err := auth.DiscoverAndConfigure(ctx, cli.Endpoint)
		if err == nil && discoveredConfig != nil {
			fmt.Printf("OAuth required by server. Please provide --oauth-client-id to authenticate.\n")
			fmt.Printf("Discovered endpoints:\n")
			fmt.Printf("  Authorization URL: %s\n", discoveredConfig.AuthURL)
			fmt.Printf("  Token URL: %s\n", discoveredConfig.TokenURL)
			return nil, fmt.Errorf("OAuth authentication required but no client ID provided (use --oauth-client-id)")
		}
		// If discovery fails or returns nil, proceed without OAuth
	}

	mcpTransport, err := transport.Create(&transportConfig, oauthConfig)
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
		return formatHTML(info, cli)
	case "pdf":
		return formatPDF(info, cli)
	case "hugo":
		return formatHugo(info, cli)
	case "markdown":
		return formatMarkdown(info, cli)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", cli.Format)
	}
}

// formatHTML generates HTML output from server information
func formatHTML(info *model.ServerInfo, cli *CLI) ([]byte, error) {
	htmlStr, err := formatter.FormatHTML(info, !cli.NoTOC, TemplateFS)
	if err != nil {
		return nil, err
	}
	return []byte(htmlStr), nil
}

// formatPDF generates PDF output from server information
func formatPDF(info *model.ServerInfo, cli *CLI) ([]byte, error) {
	if cli.Output == "" {
		return nil, fmt.Errorf("PDF format requires --output flag")
	}
	return formatter.FormatPDF(info, !cli.NoTOC)
}

// formatHugo generates Hugo site from server information
func formatHugo(info *model.ServerInfo, cli *CLI) ([]byte, error) {
	if cli.Output == "" {
		return nil, fmt.Errorf("hugo format requires --output flag (directory path)")
	}

	customFields := formatter.ParseCustomFields(cli.FrontmatterField)
	enableFrontmatter := cli.Frontmatter || cli.Format == "hugo"

	hugoConfig := &formatter.HugoConfig{
		BaseURL:       cli.HugoBaseURL,
		LanguageCode:  cli.HugoLanguageCode,
		EnterpriseKey: cli.HugoEnterpriseKey,
		AuthorStrict:  cli.HugoAuthorStrict,
	}

	warnDeprecatedHugoFlags(cli)

	err := formatter.FormatHugo(info, cli.Output, enableFrontmatter, cli.FrontmatterFormat, customFields, hugoConfig, cli.CustomInitialisms, HugoTemplateFS)
	if err != nil {
		return nil, err
	}
	// Return empty bytes since Hugo writes directly to files
	return []byte{}, nil
}

// warnDeprecatedHugoFlags logs warnings for deprecated Hugo configuration flags
func warnDeprecatedHugoFlags(cli *CLI) {
	if cli.HugoTheme != "" {
		log.Printf("⚠️  WARNING: --hugo-theme is no longer supported. Hugo now uses Presidium layouts via Hugo modules.")
	}
	if cli.HugoGithub != "" || cli.HugoTwitter != "" || cli.HugoSiteLogo != "" || cli.HugoGoogleAnalytics != "" {
		log.Printf("⚠️  WARNING: Social media, logo, and analytics flags are no longer supported for Presidium layouts.")
	}
}

// formatMarkdown generates markdown output from server information
func formatMarkdown(info *model.ServerInfo, cli *CLI) ([]byte, error) {
	customFields := formatter.ParseCustomFields(cli.FrontmatterField)
	markdownStr, err := formatter.FormatMarkdown(info, !cli.NoTOC, cli.Frontmatter, cli.FrontmatterFormat, customFields, cli.CustomInitialisms, TemplateFS)
	if err != nil {
		return nil, err
	}
	return []byte(markdownStr), nil
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

// callTools calls MCP tools based on CLI configuration and stores results in ServerInfo.
// It supports calling specific tools by name or all available tools.
func callTools(session *mcp.ClientSession, ctx context.Context, info *model.ServerInfo, cli *CLI) error {
	// Skip if no tool calling is requested
	if len(cli.CallTool) == 0 && !cli.CallAllTools {
		return nil
	}

	// Skip if tools capability is not available
	if !info.Capabilities.Tools || len(info.Tools) == 0 {
		if cli.CallAllTools || len(cli.CallTool) > 0 {
			log.Printf("Warning: Tool calling requested but server has no tools capability or no tools available")
		}
		return nil
	}

	// Parse tool arguments if provided
	var args any
	if cli.ToolArgs != "" {
		if err := json.Unmarshal([]byte(cli.ToolArgs), &args); err != nil {
			return fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	// Determine which tools to call
	var toolsToCall []string
	if cli.CallAllTools {
		// Call all available tools
		for _, tool := range info.Tools {
			toolsToCall = append(toolsToCall, tool.Name)
		}
		log.Printf("Calling all %d available tools", len(toolsToCall))
	} else {
		// Call specific tools
		toolsToCall = cli.CallTool
	}

	// Call each tool and collect results
	for _, toolName := range toolsToCall {
		result := callSingleTool(session, ctx, toolName, args)
		info.ToolCalls = append(info.ToolCalls, result)
	}

	return nil
}

// callSingleTool calls a single tool and returns the result.
// It handles errors gracefully by storing them in the ToolCall result.
func callSingleTool(session *mcp.ClientSession, ctx context.Context, toolName string, args any) model.ToolCall {
	log.Printf("Calling tool: %s", toolName)

	result := model.ToolCall{
		ToolName:  toolName,
		Arguments: args,
	}

	callResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		result.Error = err.Error()
		log.Printf("Warning: Tool call failed for %s: %v", toolName, err)
		return result
	}

	// Store the content and structured content from the result
	if callResult != nil {
		for _, content := range callResult.Content {
			result.Content = append(result.Content, content)
		}
		result.StructuredContent = callResult.StructuredContent
	}

	return result
}
