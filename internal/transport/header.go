package transport

import (
	"net/http"
)

// HeaderRoundTripper adds custom headers to HTTP requests
type HeaderRoundTripper struct {
	transport http.RoundTripper
	headers   map[string]string
}

// NewHeaderRoundTripper creates a new HeaderRoundTripper with the given headers
func NewHeaderRoundTripper(transport http.RoundTripper, headers map[string]string) *HeaderRoundTripper {
	return &HeaderRoundTripper{
		transport: transport,
		headers:   headers,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (h *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())

	// Add custom headers
	for key, value := range h.headers {
		newReq.Header.Set(key, value)
	}

	return h.transport.RoundTrip(newReq)
}
