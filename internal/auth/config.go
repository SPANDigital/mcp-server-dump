package auth

import (
	"time"

	"golang.org/x/oauth2"
)

// FlowType represents the OAuth flow type to use.
type FlowType string

const (
	// FlowTypeAuto automatically selects the best flow based on available endpoints
	FlowTypeAuto FlowType = "auto"

	// FlowTypeAuthorizationCode uses authorization code flow with PKCE (RFC 6749 + RFC 7636)
	FlowTypeAuthorizationCode FlowType = "authorization-code"

	// FlowTypeDeviceFlow uses device authorization grant flow (RFC 8628)
	FlowTypeDeviceFlow FlowType = "device"

	// FlowTypeClientCredentials uses client credentials grant (RFC 6749)
	FlowTypeClientCredentials FlowType = "client-credentials"
)

// Config holds OAuth 2.1 configuration for authenticating with MCP servers.
type Config struct {
	// ClientID is the OAuth client identifier (required unless using DCR)
	ClientID string

	// ClientSecret is the OAuth client secret (optional, for confidential clients)
	ClientSecret string

	// Scopes are the OAuth scopes to request (e.g., "mcp:tools", "mcp:resources", "mcp:prompts")
	Scopes []string

	// RedirectPort is the port for the loopback redirect server (0 = random ephemeral port)
	RedirectPort int

	// ResourceURI is the MCP server endpoint URI (used for RFC 8707 resource parameter)
	ResourceURI string

	// UseCache enables token caching to disk
	UseCache bool

	// AuthURL is the authorization endpoint (normally discovered via metadata)
	// For device flow, this is the device authorization endpoint
	AuthURL string

	// TokenURL is the token endpoint (normally discovered via metadata)
	TokenURL string

	// RegistrationEndpoint is the dynamic client registration endpoint (RFC 7591)
	RegistrationEndpoint string

	// FlowType specifies which OAuth flow to use
	FlowType FlowType

	// UseDCR enables Dynamic Client Registration (RFC 7591)
	UseDCR bool
}

// TokenCache represents a cached OAuth token for a specific MCP server.
type TokenCache struct {
	// ResourceURI is the MCP server endpoint this token is for
	ResourceURI string `json:"resource_uri"`

	// AccessToken is the OAuth access token
	AccessToken string `json:"access_token"`

	// RefreshToken is the OAuth refresh token (if provided)
	RefreshToken string `json:"refresh_token,omitempty"`

	// TokenType is the token type (typically "Bearer")
	TokenType string `json:"token_type"`

	// Expiry is when the access token expires
	Expiry time.Time `json:"expiry"`

	// Scopes are the scopes granted for this token
	Scopes []string `json:"scopes,omitempty"`
}

// ToOAuth2Token converts TokenCache to oauth2.Token for use with oauth2 library.
func (tc *TokenCache) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  tc.AccessToken,
		RefreshToken: tc.RefreshToken,
		TokenType:    tc.TokenType,
		Expiry:       tc.Expiry,
	}
}

// FromOAuth2Token creates a TokenCache from an oauth2.Token.
func FromOAuth2Token(token *oauth2.Token, resourceURI string, scopes []string) *TokenCache {
	return &TokenCache{
		ResourceURI:  resourceURI,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		Scopes:       scopes,
	}
}

// ProtectedResourceMetadata represents RFC 9728 protected resource metadata.
type ProtectedResourceMetadata struct {
	// Resource is the protected resource identifier
	Resource string `json:"resource"`

	// AuthorizationServers lists the authorization servers that can issue tokens
	AuthorizationServers []string `json:"authorization_servers"`

	// ScopesSupported are the scopes this resource understands
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// BearerMethodsSupported are the ways Bearer tokens can be sent
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
}

// AuthServerMetadata represents RFC 8414 authorization server metadata.
// Note: Named AuthServerMetadata (not ServerMetadata) to avoid confusion with other server types.
type AuthServerMetadata struct { //nolint:revive // AuthServerMetadata is intentionally prefixed for clarity
	// Issuer is the authorization server's issuer identifier
	Issuer string `json:"issuer"`

	// AuthorizationEndpoint is the URL for the authorization endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint"`

	// TokenEndpoint is the URL for the token endpoint
	TokenEndpoint string `json:"token_endpoint"`

	// DeviceAuthorizationEndpoint is the URL for device authorization (RFC 8628)
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitempty"`

	// RegistrationEndpoint is the URL for dynamic client registration (optional)
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// ScopesSupported are the scopes the server supports
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// ResponseTypesSupported are the OAuth response types supported
	ResponseTypesSupported []string `json:"response_types_supported"`

	// CodeChallengeMethodsSupported are the PKCE challenge methods (must include "S256")
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`

	// GrantTypesSupported are the grant types supported
	GrantTypesSupported []string `json:"grant_types_supported,omitempty"`

	// TokenEndpointAuthMethodsSupported are the client authentication methods
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
}

// ClientRegistration represents a cached client registration for a specific MCP server.
type ClientRegistration struct {
	// ResourceURI is the MCP server endpoint this registration is for
	ResourceURI string `json:"resource_uri"`

	// ClientID is the registered OAuth client identifier
	ClientID string `json:"client_id"`

	// ClientSecret is the registered OAuth client secret (if provided)
	ClientSecret string `json:"client_secret,omitempty"`

	// RegistrationAccessToken is the token for managing this registration
	RegistrationAccessToken string `json:"registration_access_token,omitempty"`

	// RegisteredAt is when this registration was created
	RegisteredAt time.Time `json:"registered_at"`
}

// ClientRegistrationRequest represents a dynamic client registration request (RFC 7591).
type ClientRegistrationRequest struct {
	// ClientName is a human-readable name for the client
	ClientName string `json:"client_name"`

	// RedirectURIs are the redirect URIs for authorization code flow
	RedirectURIs []string `json:"redirect_uris,omitempty"`

	// GrantTypes are the OAuth grant types the client will use
	GrantTypes []string `json:"grant_types"`

	// TokenEndpointAuthMethod is how the client authenticates to the token endpoint
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`

	// Scope is the space-separated list of scopes
	Scope string `json:"scope,omitempty"`
}

// ClientRegistrationResponse represents the response from dynamic client registration.
type ClientRegistrationResponse struct {
	// ClientID is the registered client identifier
	ClientID string `json:"client_id"`

	// ClientSecret is the client secret (for confidential clients)
	ClientSecret string `json:"client_secret,omitempty"`

	// ClientIDIssuedAt is when the client_id was issued (Unix timestamp)
	ClientIDIssuedAt int64 `json:"client_id_issued_at,omitempty"`

	// ClientSecretExpiresAt is when the client_secret expires (Unix timestamp, 0=never)
	ClientSecretExpiresAt int64 `json:"client_secret_expires_at,omitempty"`

	// RegistrationAccessToken is the token for managing this registration
	RegistrationAccessToken string `json:"registration_access_token,omitempty"`

	// RegistrationClientURI is the URI for managing this registration
	RegistrationClientURI string `json:"registration_client_uri,omitempty"`
}

// DefaultScopes returns the default MCP scopes to request.
func DefaultScopes() []string {
	return []string{"mcp:tools", "mcp:resources", "mcp:prompts"}
}
