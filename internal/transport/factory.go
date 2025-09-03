package transport

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

// Create creates an MCP transport based on the configuration
func Create(config *Config) (mcp.Transport, error) {
	switch config.Transport {
	case "command":
		return createCommandTransport(config)
	case "sse":
		return createSSETransport(config)
	case "streamable":
		return createStreamableTransport(config)
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

func createSSETransport(config *Config) (mcp.Transport, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("SSE transport requires --endpoint")
	}

	httpClient := &http.Client{Timeout: config.Timeout}

	// Build transport chain for SSE
	transport := buildHTTPTransportChain(http.DefaultTransport, config.Headers, false)
	httpClient.Transport = transport

	return &mcp.SSEClientTransport{
		Endpoint:   config.Endpoint,
		HTTPClient: httpClient,
	}, nil
}

func createStreamableTransport(config *Config) (mcp.Transport, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("streamable transport requires --endpoint")
	}

	httpClient := &http.Client{Timeout: config.Timeout}

	// Build transport chain for streamable (includes content type fixing)
	transport := buildHTTPTransportChain(http.DefaultTransport, config.Headers, true)
	httpClient.Transport = transport

	return &mcp.StreamableClientTransport{
		Endpoint:   config.Endpoint,
		HTTPClient: httpClient,
	}, nil
}

// buildHTTPTransportChain builds a chain of HTTP round trippers
func buildHTTPTransportChain(base http.RoundTripper, headerStrings []string, includeContentTypeFix bool) http.RoundTripper {
	transport := base

	// Add custom headers if specified
	if len(headerStrings) > 0 {
		headers, err := parseHeaders(headerStrings)
		if err == nil && len(headers) > 0 {
			transport = NewHeaderRoundTripper(transport, headers)
		}
	}

	// Add content type fixing for streamable transport
	if includeContentTypeFix {
		transport = NewContentTypeFixingTransport(transport)
	}

	return transport
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
