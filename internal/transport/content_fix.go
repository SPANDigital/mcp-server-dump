package transport

import (
	"net/http"
	"strings"
)

// ContentTypeFixingTransport normalizes content types for MCP compatibility
type ContentTypeFixingTransport struct {
	transport http.RoundTripper
}

// NewContentTypeFixingTransport creates a new ContentTypeFixingTransport
func NewContentTypeFixingTransport(transport http.RoundTripper) *ContentTypeFixingTransport {
	return &ContentTypeFixingTransport{
		transport: transport,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (c *ContentTypeFixingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Fix content type by removing charset parameter from JSON responses
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		if strings.HasPrefix(contentType, "application/json;") {
			resp.Header.Set("Content-Type", "application/json")
		}
	}

	return resp, err
}
