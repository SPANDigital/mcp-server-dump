package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
)

// OAuthRoundTripper is an http.RoundTripper that automatically injects OAuth Bearer tokens
// into HTTP requests and handles token refresh when needed.
type OAuthRoundTripper struct {
	base   http.RoundTripper
	config *Config

	mu           sync.RWMutex
	initCond     *sync.Cond
	token        *oauth2.Token
	tokenSource  oauth2.TokenSource
	initializing bool
	initErr      error
}

// NewOAuthRoundTripper creates a new OAuthRoundTripper that wraps the base RoundTripper.
// It automatically loads cached tokens and performs OAuth flow if needed.
func NewOAuthRoundTripper(base http.RoundTripper, config *Config) (*OAuthRoundTripper, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth config cannot be nil")
	}

	if base == nil {
		base = http.DefaultTransport
	}

	rt := &OAuthRoundTripper{
		base:   base,
		config: config,
	}
	rt.initCond = sync.NewCond(&rt.mu)

	// Try to load cached token if caching is enabled
	if config.UseCache {
		cached, err := LoadToken(config.ResourceURI)
		if err == nil && cached != nil {
			rt.token = cached.ToOAuth2Token()

			// Create token source for auto-refresh if we have a valid token
			if rt.token != nil {
				rt.tokenSource = rt.createTokenSource(rt.token)
			}
		}
	}

	return rt, nil
}

// RoundTrip implements http.RoundTripper.
// It injects the Bearer token into the Authorization header and handles token refresh.
func (rt *OAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get valid token (may trigger OAuth flow or refresh)
	token, err := rt.getValidToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get valid OAuth token: %w", err)
	}

	// Clone request to avoid modifying original
	clonedReq := req.Clone(req.Context())

	// Inject Bearer token into Authorization header
	if token != nil && token.AccessToken != "" {
		clonedReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	}

	// Perform request
	resp, err := rt.base.RoundTrip(clonedReq)

	// If we get 401 Unauthorized, token may be invalid despite not being expired
	// Clear token and retry once
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		rt.mu.Lock()
		rt.token = nil
		rt.tokenSource = nil
		rt.mu.Unlock()

		// Try to get a fresh token
		newToken, tokenErr := rt.performOAuthFlow(req.Context())
		if tokenErr != nil {
			// Return original 401 response along with the error
			return resp, fmt.Errorf("failed to refresh token after 401: %w", tokenErr)
		}

		// Retry request with new token
		clonedReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newToken.AccessToken))
		return rt.base.RoundTrip(clonedReq)
	}

	return resp, err
}

// getValidToken returns a valid OAuth token, performing OAuth flow or token refresh if needed.
func (rt *OAuthRoundTripper) getValidToken(ctx context.Context) (*oauth2.Token, error) {
	rt.mu.RLock()
	token := rt.token
	tokenSource := rt.tokenSource
	rt.mu.RUnlock()

	// If we have a token source, use it to get a valid token (handles refresh automatically)
	if tokenSource != nil {
		newToken, err := tokenSource.Token()
		if err == nil {
			// Update stored token if it changed (check both access token and expiry)
			if token == nil || newToken.AccessToken != token.AccessToken || newToken.Expiry != token.Expiry {
				rt.mu.Lock()
				rt.token = newToken
				rt.mu.Unlock()

				// Save to cache if enabled
				if rt.config.UseCache {
					if saveErr := SaveToken(newToken, rt.config.ResourceURI, rt.config.Scopes); saveErr != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to save refreshed token to cache: %v\n", saveErr)
					}
				}
			}
			return newToken, nil
		}
		// If refresh failed, fall through to perform full OAuth flow
	}

	// Check if token is valid and not expired
	if token != nil && token.Valid() {
		return token, nil
	}

	// Need to perform OAuth flow
	return rt.performOAuthFlow(ctx)
}

// performOAuthFlow performs the full OAuth 2.1 authorization code flow with PKCE.
func (rt *OAuthRoundTripper) performOAuthFlow(ctx context.Context) (*oauth2.Token, error) {
	rt.mu.Lock()
	// Check if another goroutine is already performing OAuth flow
	if rt.initializing {
		// Wait for initialization to complete using condition variable
		for rt.initializing {
			rt.initCond.Wait()
		}
		initErr := rt.initErr
		token := rt.token
		rt.mu.Unlock()
		if initErr != nil {
			return nil, initErr
		}
		return token, nil
	}
	rt.initializing = true
	rt.mu.Unlock()

	// Perform OAuth flow
	token, err := Authorize(ctx, rt.config)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.initializing = false
	rt.initErr = err

	// Broadcast to all waiting goroutines
	rt.initCond.Broadcast()

	if err != nil {
		return nil, err
	}

	// Store token
	rt.token = token

	// Create token source for auto-refresh
	rt.tokenSource = rt.createTokenSource(token)

	// Save to cache if enabled
	if rt.config.UseCache {
		if saveErr := SaveToken(token, rt.config.ResourceURI, rt.config.Scopes); saveErr != nil {
			// Log error but don't fail the request
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to save token to cache: %v\n", saveErr)
		}
	}

	return token, nil
}

// createTokenSource creates an oauth2.TokenSource for automatic token refresh.
// Note: Uses context.Background() because the TokenSource is long-lived and outlives
// individual requests. The oauth2.TokenSource interface doesn't support per-call contexts,
// and the context is only used for HTTP client configuration (not request cancellation).
// Token refresh operations won't be cancelled by request context cancellation, which is
// acceptable for this use case as tokens are cached and reused across requests.
func (rt *OAuthRoundTripper) createTokenSource(token *oauth2.Token) oauth2.TokenSource {
	oauth2Config := &oauth2.Config{
		ClientID:     rt.config.ClientID,
		ClientSecret: rt.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: rt.config.TokenURL,
		},
	}

	// Create context with custom HTTP client that adds resource parameter
	// Uses Background() because TokenSource is long-lived and context is only for client config
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: &resourceParamTransport{
			base:     http.DefaultTransport,
			resource: rt.config.ResourceURI,
		},
	})

	// Create token source that auto-refreshes
	return oauth2.ReuseTokenSource(token, oauth2Config.TokenSource(ctx, token))
}
