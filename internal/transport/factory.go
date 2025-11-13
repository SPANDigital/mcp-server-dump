package transport

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/spandigital/mcp-server-dump/internal/auth"
)

// Config holds transport configuration
type Config struct {
	Transport     string
	Endpoint      string
	Timeout       time.Duration
	Headers       []string
	ServerCommand string
	Args          []string
}

// Create creates an MCP transport based on the configuration.
// The oauthConfig parameter is optional and only used for HTTP-based transports.
func Create(config *Config, oauthConfig *auth.Config) (mcp.Transport, error) {
	switch config.Transport {
	case "command":
		return createCommandTransport(config)
	case "sse":
		return createSSETransport(config, oauthConfig)
	case "streamable":
		return createStreamableTransport(config, oauthConfig)
	default:
		return nil, fmt.Errorf("unknown transport type: %s", config.Transport)
	}
}

func createCommandTransport(config *Config) (mcp.Transport, error) {
	var cmd *exec.Cmd

	switch {
	case config.ServerCommand != "":
		// Parse explicit server command
		parts := strings.Fields(config.ServerCommand)
		if len(parts) == 0 {
			return nil, fmt.Errorf("empty server command")
		}
		// #nosec G204 - Command is provided by user intentionally
		cmd = exec.Command(parts[0], parts[1:]...)
	case len(config.Args) > 0:
		// Use legacy args format
		// #nosec G204 - Command and args are provided by user intentionally
		cmd = exec.Command(config.Args[0], config.Args[1:]...)
	default:
		return nil, fmt.Errorf("command transport requires command or args")
	}

	return &mcp.CommandTransport{Command: cmd}, nil
}

func createSSETransport(config *Config, oauthConfig *auth.Config) (mcp.Transport, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("SSE transport requires --endpoint")
	}

	httpClient := &http.Client{Timeout: config.Timeout}

	// Build transport chain for SSE (OAuth layer added if configured)
	transport, err := buildHTTPTransportChain(http.DefaultTransport, config.Headers, false, oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build transport chain: %w", err)
	}
	httpClient.Transport = transport

	return &mcp.SSEClientTransport{
		Endpoint:   config.Endpoint,
		HTTPClient: httpClient,
	}, nil
}

func createStreamableTransport(config *Config, oauthConfig *auth.Config) (mcp.Transport, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("streamable transport requires --endpoint")
	}

	httpClient := &http.Client{Timeout: config.Timeout}

	// Build transport chain for streamable (includes content type fixing and OAuth if configured)
	transport, err := buildHTTPTransportChain(http.DefaultTransport, config.Headers, true, oauthConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build transport chain: %w", err)
	}
	httpClient.Transport = transport

	return &mcp.StreamableClientTransport{
		Endpoint:   config.Endpoint,
		HTTPClient: httpClient,
	}, nil
}

// buildHTTPTransportChain builds a chain of HTTP round trippers.
// The chain order is: base → OAuth (if configured) → custom headers → content type fix
func buildHTTPTransportChain(base http.RoundTripper, headerStrings []string, includeContentTypeFix bool, oauthConfig *auth.Config) (http.RoundTripper, error) {
	transport := base

	// Add OAuth layer first (if configured)
	// This allows OAuth to inject/update the Authorization header
	if oauthConfig != nil {
		oauthTransport, err := auth.NewOAuthRoundTripper(transport, oauthConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create OAuth transport: %w", err)
		}
		transport = oauthTransport
	}

	// Add custom headers if specified
	// These can override OAuth headers if needed (for backward compatibility)
	if len(headerStrings) > 0 {
		headers, err := parseHeaders(headerStrings)
		if err != nil {
			return nil, err
		}
		if len(headers) > 0 {
			transport = NewHeaderRoundTripper(transport, headers)
		}
	}

	// Add content type fixing for streamable transport
	if includeContentTypeFix {
		transport = NewContentTypeFixingTransport(transport)
	}

	return transport, nil
}

// parseHeaders parses header strings in format "Key:Value"
func parseHeaders(headerStrings []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, h := range headerStrings {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s (expected Key:Value)", h)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty header key in: %s", h)
		}
		headers[key] = value
	}
	return headers, nil
}
