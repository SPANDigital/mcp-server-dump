package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DiscoverFromResponse attempts to discover OAuth endpoints from a 401 Unauthorized response.
// It parses the WWW-Authenticate header according to RFC 9728 and fetches metadata.
func DiscoverFromResponse(resp *http.Response) (*Config, error) {
	if resp.StatusCode != http.StatusUnauthorized {
		return nil, fmt.Errorf("expected 401 Unauthorized response, got %d", resp.StatusCode)
	}

	// Parse WWW-Authenticate header
	authHeader := resp.Header.Get("WWW-Authenticate")
	if authHeader == "" {
		return nil, fmt.Errorf("no WWW-Authenticate header in 401 response")
	}

	// Extract metadata URL from WWW-Authenticate header
	// Format: Bearer realm="https://example.com", as_uri="https://as.example.com"
	metadataURL, err := parseWWWAuthenticate(authHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WWW-Authenticate header: %w", err)
	}

	if metadataURL == "" {
		return nil, fmt.Errorf("no metadata URL found in WWW-Authenticate header")
	}

	// Fetch protected resource metadata
	prMetadata, err := fetchProtectedResourceMetadata(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch protected resource metadata: %w", err)
	}

	// Get authorization server URL from metadata
	if len(prMetadata.AuthorizationServers) == 0 {
		return nil, fmt.Errorf("no authorization servers found in protected resource metadata")
	}

	asURL := prMetadata.AuthorizationServers[0] // Use first authorization server

	// Fetch authorization server metadata
	asMetadata, err := fetchAuthServerMetadata(asURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch authorization server metadata: %w", err)
	}

	// Build OAuth config from discovered metadata
	config := &Config{
		AuthURL:     asMetadata.AuthorizationEndpoint,
		TokenURL:    asMetadata.TokenEndpoint,
		ResourceURI: prMetadata.Resource,
		Scopes:      prMetadata.ScopesSupported,
		UseCache:    true, // Enable caching by default
	}

	// Validate PKCE support
	if !contains(asMetadata.CodeChallengeMethodsSupported, "S256") {
		return nil, fmt.Errorf("authorization server does not support S256 PKCE (required by OAuth 2.1)")
	}

	return config, nil
}

// parseWWWAuthenticate parses the WWW-Authenticate header to extract the metadata URL.
// RFC 9728 format: Bearer realm="https://resource.example.com"
// Some servers may use as_uri parameter for authorization server URL.
func parseWWWAuthenticate(header string) (string, error) {
	// Check if it's a Bearer challenge
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return "", fmt.Errorf("not a Bearer challenge")
	}

	// Remove "Bearer " prefix
	params := strings.TrimPrefix(header, "Bearer ")
	params = strings.TrimPrefix(params, "bearer ")

	// Parse parameters (simplified parser for key="value" pairs)
	// In production, should use more robust parsing
	realm := extractParam(params, "realm")
	if realm == "" {
		return "", fmt.Errorf("no realm parameter found")
	}

	// The realm is typically the protected resource metadata URL
	// Format: https://resource.example.com/.well-known/oauth-protected-resource
	return realm, nil
}

// extractParam extracts a parameter value from a WWW-Authenticate header.
// Format: key="value"
func extractParam(params, key string) string {
	// Look for key="value" or key='value'
	for _, quote := range []string{`"`, `'`} {
		prefix := key + "=" + quote
		idx := strings.Index(params, prefix)
		if idx == -1 {
			continue
		}

		// Find closing quote
		start := idx + len(prefix)
		end := strings.Index(params[start:], quote)
		if end == -1 {
			continue
		}

		return params[start : start+end]
	}

	return ""
}

// fetchProtectedResourceMetadata fetches RFC 9728 protected resource metadata.
func fetchProtectedResourceMetadata(metadataURL string) (*ProtectedResourceMetadata, error) {
	// If URL doesn't include the well-known path, append it
	if !strings.Contains(metadataURL, "/.well-known/") {
		u, err := url.Parse(metadataURL)
		if err != nil {
			return nil, err
		}
		u.Path = "/.well-known/oauth-protected-resource"
		metadataURL = u.String()
	}

	// #nosec G107 - URL is constructed from server-provided metadata or user input
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch metadata (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var metadata ProtectedResourceMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// fetchAuthServerMetadata fetches RFC 8414 authorization server metadata.
func fetchAuthServerMetadata(issuerURL string) (*AuthServerMetadata, error) {
	// Build well-known URL according to RFC 8414
	// Format: <issuer>/.well-known/oauth-authorization-server
	u, err := url.Parse(issuerURL)
	if err != nil {
		return nil, err
	}

	// Append well-known path
	u.Path = strings.TrimSuffix(u.Path, "/") + "/.well-known/oauth-authorization-server"
	metadataURL := u.String()

	// #nosec G107 - URL is constructed from server-provided metadata or user input
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch metadata (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var metadata AuthServerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// DiscoverAndConfigure attempts to discover OAuth configuration by making a probe request
// to the MCP server endpoint. Returns discovered config or nil if server doesn't require OAuth.
func DiscoverAndConfigure(ctx context.Context, endpoint string) (*Config, error) {
	// Make probe request to trigger 401 with WWW-Authenticate
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// If not 401, server doesn't require OAuth
	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}

	// Discover OAuth configuration from response
	return DiscoverFromResponse(resp)
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
