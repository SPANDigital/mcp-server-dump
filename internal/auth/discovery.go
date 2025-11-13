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
		// Limit error body reads to 1KB to prevent memory exhaustion
		limitedReader := io.LimitReader(resp.Body, 1024)
		body, _ := io.ReadAll(limitedReader)
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
		// Limit error body reads to 1KB to prevent memory exhaustion
		limitedReader := io.LimitReader(resp.Body, 1024)
		body, _ := io.ReadAll(limitedReader)
		return nil, fmt.Errorf("failed to fetch metadata (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var metadata AuthServerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// discoverFromWellKnown attempts to discover OAuth endpoints by directly querying
// the .well-known/oauth-authorization-server endpoint (RFC 8414).
// This is a fallback when WWW-Authenticate header is not present.
func discoverFromWellKnown(endpoint string) (*Config, error) {
	// Parse endpoint URL
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Build .well-known URL
	wellKnownURL := &url.URL{
		Scheme: endpointURL.Scheme,
		Host:   endpointURL.Host,
		Path:   "/.well-known/oauth-authorization-server",
	}

	// Fetch authorization server metadata directly
	asMetadata, err := fetchAuthServerMetadata(wellKnownURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch .well-known metadata: %w", err)
	}

	// Build OAuth config from discovered metadata
	config := &Config{
		AuthURL:     asMetadata.AuthorizationEndpoint,
		TokenURL:    asMetadata.TokenEndpoint,
		ResourceURI: endpoint,
		Scopes:      asMetadata.ScopesSupported,
		UseCache:    true,
	}

	// Validate PKCE support if using authorization code flow
	if asMetadata.AuthorizationEndpoint != "" {
		if !contains(asMetadata.CodeChallengeMethodsSupported, "S256") {
			return nil, fmt.Errorf("authorization server does not support S256 PKCE (required by OAuth 2.1)")
		}
	}

	return config, nil
}

// parseDeviceFlowFromBody attempts to extract device flow endpoints from
// a non-standard JSON response body. This handles servers that advertise
// endpoints directly in the 401 response instead of using WWW-Authenticate.
func parseDeviceFlowFromBody(body []byte) (deviceAuthURL, tokenURL string, err error) {
	var response struct {
		DeviceFlow struct {
			Step1 string `json:"step_1"` // POST endpoint for device authorization
			Step3 string `json:"step_3"` // Poll endpoint (contains token URL)
		} `json:"device_flow"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract device authorization endpoint from step_1
	// Remove "POST " prefix if present (e.g., "POST http://...")
	deviceAuthURL = strings.TrimSpace(strings.TrimPrefix(response.DeviceFlow.Step1, "POST"))
	deviceAuthURL = strings.TrimSpace(deviceAuthURL)
	if deviceAuthURL == "" {
		return "", "", fmt.Errorf("no device_flow.step_1 found in response")
	}

	// Extract token endpoint from step_3 URL
	// step_3 format: "Poll http://host/oauth/device/poll with device_code"
	// Extract just the URL part
	step3Parts := strings.Fields(response.DeviceFlow.Step3)
	var step3URL string
	for _, part := range step3Parts {
		if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
			step3URL = part
			break
		}
	}

	if step3URL != "" {
		// Parse step3 URL and extract base (scheme + host + path up to /poll)
		parsedURL, parseErr := url.Parse(step3URL)
		if parseErr == nil {
			// Remove /poll or similar suffix to get token endpoint base
			tokenURL = fmt.Sprintf("%s://%s/oauth/token", parsedURL.Scheme, parsedURL.Host)
		}
	}

	return deviceAuthURL, tokenURL, nil
}

// DiscoverAndConfigure attempts to discover OAuth configuration by making a probe request
// to the MCP server endpoint. Returns discovered config or nil if server doesn't require OAuth.
// Uses a layered discovery approach: WWW-Authenticate → .well-known → JSON body parsing.
func DiscoverAndConfigure(ctx context.Context, endpoint string) (*Config, error) {
	// Make probe request to trigger 401 with WWW-Authenticate
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Request JSON response (some servers return HTML by default for browsers)
	req.Header.Set("Accept", "application/json")

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

	// Read response body first (needed for strategy 3, but don't consume from resp.Body yet)
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096)) // Limit to 4KB

	// Strategy 1: Try WWW-Authenticate header (RFC 9728 compliant)
	if resp.Header.Get("WWW-Authenticate") != "" {
		config, err := DiscoverFromResponse(resp)
		if err == nil && config != nil {
			return config, nil
		}
	}

	// Strategy 2: Try .well-known endpoint directly
	config, wellKnownErr := discoverFromWellKnown(endpoint)
	if wellKnownErr == nil && config != nil {
		return config, nil
	}

	// Strategy 3: Parse response body for non-standard device flow advertisement
	if readErr == nil && len(body) > 0 {
		deviceAuthURL, tokenURL, parseErr := parseDeviceFlowFromBody(body)
		if parseErr == nil && deviceAuthURL != "" {
			// Successfully parsed device flow endpoints
			return &Config{
				AuthURL:     deviceAuthURL, // Store device auth URL in AuthURL for now
				TokenURL:    tokenURL,
				ResourceURI: endpoint,
				Scopes:      DefaultScopes(),
				UseCache:    true,
			}, nil
		}
		// If parsing failed, include error in final message
		if parseErr != nil {
			return nil, fmt.Errorf("discovery failed: body parse error: %w", parseErr)
		}
	}

	// All strategies failed - return most informative error
	if wellKnownErr != nil {
		return nil, fmt.Errorf("discovery failed: WWW-Authenticate not found, .well-known failed (%w), body parsing failed", wellKnownErr)
	}
	return nil, fmt.Errorf("failed to discover OAuth endpoints using all strategies")
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
