package app

import (
	"errors"
	"time"

	"github.com/alecthomas/kong"
)

// CLI represents the command line interface configuration
type CLI struct {
	// Version flag
	Version kong.VersionFlag `kong:"short='v',help='Show version information'"`

	// Output options
	Output string `kong:"short='o',help='Output file for documentation (defaults to stdout, required for hugo format as directory)'"`
	Format string `kong:"short='f',default='markdown',enum='markdown,json,html,pdf,hugo',help='Output format'"`
	NoTOC  bool   `kong:"help='Disable table of contents in markdown output'"`

	// Frontmatter options
	Frontmatter       bool     `kong:"short='F',help='Include frontmatter in markdown output (enabled by default for Hugo format)'"`
	FrontmatterField  []string `kong:"short='M',help='Add custom frontmatter field (format: key:value), can be used multiple times'"`
	FrontmatterFormat string   `kong:"default='yaml',enum='yaml,toml,json',help='Frontmatter format'"`

	// Transport selection
	Transport string `kong:"short='t',default='command',enum='command,sse,streamable',help='Transport type'"`

	// Transport-specific options
	Endpoint string        `kong:"help='HTTP endpoint for SSE/Streamable transports'"`
	Timeout  time.Duration `kong:"default='30s',help='HTTP timeout for SSE/Streamable transports'"`
	Headers  []string      `kong:"short='H',help='HTTP headers for SSE/Streamable transports (format: Key:Value)'"`

	// Context configuration
	ContextFile   []string `kong:"help='Path to context configuration files (YAML/JSON), can be used multiple times'"`
	ServerCommand string   `kong:"help='Server command for explicit command transport'"`

	// OAuth 2.1 authentication options
	OAuthClientID     string   `kong:"name='oauth-client-id',help='OAuth 2.1 client ID for authenticated MCP server access'"`
	OAuthClientSecret string   `kong:"name='oauth-client-secret',help='OAuth 2.1 client secret (for confidential clients, use with caution)'"`
	OAuthScopes       []string `kong:"name='oauth-scopes',help='OAuth scopes to request (comma-separated, e.g., mcp:tools,mcp:resources,mcp:prompts)'"`
	OAuthAuthURL      string   `kong:"name='oauth-auth-url',help='OAuth authorization endpoint URL (normally discovered automatically)'"`
	OAuthTokenURL     string   `kong:"name='oauth-token-url',help='OAuth token endpoint URL (normally discovered automatically)'"`
	OAuthRedirectPort int      `kong:"name='oauth-redirect-port',default='0',help='Port for OAuth loopback redirect (0=random ephemeral port)'"`
	OAuthNoCache      bool     `kong:"name='oauth-no-cache',help='Disable OAuth token caching (always require fresh authentication)'"`
	OAuthFlow         string   `kong:"name='oauth-flow',default='auto',enum='auto,authorization-code,device,client-credentials',help='OAuth flow type (auto-detects by default)'"`

	// Scanning options
	NoTools     bool `kong:"help='Skip scanning tools from the MCP server'"`
	NoResources bool `kong:"help='Skip scanning resources from the MCP server'"`
	NoPrompts   bool `kong:"help='Skip scanning prompts from the MCP server'"`

	// Tool calling options
	CallTool     []string `kong:"help='Call specific tool(s) by name, can be used multiple times'"`
	ToolArgs     string   `kong:"help='JSON arguments for tool calls (applies to all --call-tool invocations)'"`
	CallAllTools bool     `kong:"help='Call all available tools with empty arguments for testing'"`

	// Hugo-specific options (only used when format=hugo)
	// Uses Hugo Modules with Presidium layouts
	HugoBaseURL       string `kong:"help='Base URL for Hugo site (e.g., https://example.com)'"`
	HugoLanguageCode  string `kong:"help='Language code for Hugo site (default: en-us)'"`
	HugoEnterpriseKey string `kong:"help='Enterprise key for Presidium configuration (optional, uses server name if not specified)'"`
	HugoAuthorStrict  bool   `kong:"help='Require author field in frontmatter (default: false)'"`

	// Deprecated Hugo flags (maintained for backward compatibility)
	HugoTheme           string `kong:"help='[DEPRECATED] No longer supported with Presidium layouts',hidden"`
	HugoGithub          string `kong:"help='[DEPRECATED] No longer supported with Presidium layouts',hidden"`
	HugoTwitter         string `kong:"help='[DEPRECATED] No longer supported with Presidium layouts',hidden"`
	HugoSiteLogo        string `kong:"help='[DEPRECATED] No longer supported with Presidium layouts',hidden"`
	HugoGoogleAnalytics string `kong:"help='[DEPRECATED] No longer supported with Presidium layouts',hidden"`

	// Context formatting options
	CustomInitialisms []string `kong:"help='Additional technical initialisms to recognize for human-readable headings (comma-separated, e.g., API,CDN,JWT)'"`

	// Legacy command format (backward compatibility)
	Args []string `kong:"arg,optional,help='Command and arguments (legacy format for backward compatibility)'"`
}

// ValidateScanOptions validates that at least one scan type is enabled
func (cli *CLI) ValidateScanOptions() error {
	if cli.NoTools && cli.NoResources && cli.NoPrompts {
		return errors.New(ErrAllScanTypesDisabled)
	}
	return nil
}
