package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Authorize performs the OAuth 2.1 authorization code flow with PKCE.
// It launches a browser for user authentication and returns the access token.
func Authorize(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	if cfg.AuthURL == "" || cfg.TokenURL == "" {
		return nil, fmt.Errorf("authorization and token URLs must be configured")
	}

	if cfg.ResourceURI == "" {
		return nil, fmt.Errorf("resource URI (MCP server endpoint) must be specified")
	}

	// Generate PKCE verifier and challenge
	verifier := oauth2.GenerateVerifier()

	// Determine scopes
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: "", // Will be set when we know the port
		Scopes:      scopes,
	}

	// Start loopback HTTP server for OAuth callback
	redirectURI, codeChan, errChan, shutdownFn, err := startCallbackServer(cfg.RedirectPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer shutdownFn()

	oauth2Config.RedirectURL = redirectURI

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL with PKCE and resource parameter
	authURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("resource", cfg.ResourceURI),
	)

	// Open browser for user authentication
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open automatically, visit:\n%s\n\n", authURL)

	if browserErr := openBrowser(authURL); browserErr != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", browserErr)
		fmt.Printf("Please manually visit the URL above.\n\n")
	}

	// Wait for authorization code or error
	var code string
	var receivedState string

	select {
	case result := <-codeChan:
		code = result.Code
		receivedState = result.State
	case authErr := <-errChan:
		return nil, fmt.Errorf("authorization failed: %w", authErr)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authorization timed out after 5 minutes")
	}

	// Validate state parameter
	if receivedState != state {
		return nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	// Exchange authorization code for token
	token, err := oauth2Config.Exchange(
		ctx,
		code,
		oauth2.VerifierOption(verifier),
		oauth2.SetAuthURLParam("resource", cfg.ResourceURI),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	fmt.Printf("✓ Authorization successful!\n\n")
	return token, nil
}

// callbackResult holds the authorization code and state from the OAuth callback.
type callbackResult struct {
	Code  string
	State string
}

// startCallbackServer starts a loopback HTTP server to receive the OAuth callback.
// It returns the redirect URI, channels for the code/error, and a shutdown function.
//
//nolint:gocritic // Multiple unnamed return values are intentional for channel types
func startCallbackServer(port int) (string, <-chan callbackResult, <-chan error, func(), error) {
	codeChan := make(chan callbackResult, 1)
	errChan := make(chan error, 1)

	// Listen on loopback address with specified or random port
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, nil, nil, err
	}

	// Get the actual port (important when port=0 for random)
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return "", nil, nil, nil, fmt.Errorf("failed to get TCP address from listener")
	}
	actualPort := tcpAddr.Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", actualPort)

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Extract code and state from query parameters
		query := r.URL.Query()
		code := query.Get("code")
		state := query.Get("state")
		errorParam := query.Get("error")
		errorDesc := query.Get("error_description")

		if errorParam != "" {
			msg := fmt.Sprintf("authorization error: %s", errorParam)
			if errorDesc != "" {
				msg = fmt.Sprintf("%s: %s", msg, errorDesc)
			}
			errChan <- fmt.Errorf("%s", msg)

			w.WriteHeader(http.StatusBadRequest)
			// Escape HTML to prevent XSS from malicious error parameters
			_, _ = fmt.Fprintf(w, "<html><body><h1>Authorization Failed</h1><p>%s</p><p>You can close this window.</p></body></html>", html.EscapeString(msg))
			return
		}

		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(w, "<html><body><h1>Authorization Failed</h1><p>No authorization code received.</p><p>You can close this window.</p></body></html>")
			return
		}

		// Send code and state to channel
		codeChan <- callbackResult{Code: code, State: state}

		// Send success response to browser
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "<html><body><h1>Authorization Successful!</h1><p>You can close this window and return to the terminal.</p></body></html>")
	})

	// Start server in goroutine
	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Shutdown function
	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}

	return redirectURI, codeChan, errChan, shutdownFn, nil
}

// generateState creates a cryptographically secure random state parameter for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// openBrowser opens the specified URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		// Try xdg-open first (standard), fallback to other common browsers
		for _, browser := range []string{"xdg-open", "x-www-browser", "www-browser", "firefox", "chrome", "chromium"} {
			if _, err := exec.LookPath(browser); err == nil {
				// #nosec G204 - browser name is from a hardcoded allowlist
				cmd = exec.Command(browser, url)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no suitable browser found")
		}
	case "windows":
		// Use 'start' command through cmd.exe
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// RefreshToken refreshes an access token using a refresh token.
func RefreshToken(ctx context.Context, cfg *Config, refreshToken string) (*oauth2.Token, error) {
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("token URL must be configured")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: cfg.TokenURL,
		},
	}

	// Create token source for refresh
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Add resource parameter to token refresh request
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &resourceParamTransport{
			base:     http.DefaultTransport,
			resource: cfg.ResourceURI,
		},
	})

	tokenSource := oauth2Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return newToken, nil
}

// resourceParamTransport is a custom RoundTripper that adds the resource parameter
// to token refresh requests (required by MCP OAuth specification).
type resourceParamTransport struct {
	base     http.RoundTripper
	resource string
}

// RoundTrip adds the resource parameter to OAuth token requests per MCP specification.
func (t *resourceParamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only modify token endpoint requests
	if strings.Contains(req.URL.Path, "token") && req.Method == http.MethodPost {
		// Parse existing form data
		if err := req.ParseForm(); err == nil {
			// Add resource parameter
			if t.resource != "" {
				values := req.Form
				values.Set("resource", t.resource)

				// Re-encode form data
				body := values.Encode()
				req.Body = io.NopCloser(strings.NewReader(body))
				req.ContentLength = int64(len(body))
			}
		}
	}

	return t.base.RoundTrip(req)
}
