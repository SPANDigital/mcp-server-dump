package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// DeviceCodeResponse represents the response from the device authorization endpoint (RFC 8628).
type DeviceCodeResponse struct {
	// DeviceCode is the device verification code
	DeviceCode string `json:"device_code"`

	// UserCode is the end-user verification code
	UserCode string `json:"user_code"`

	// VerificationURI is the end-user verification URI
	VerificationURI string `json:"verification_uri"`

	// VerificationURIComplete is the complete verification URI including user_code (optional)
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`

	// ExpiresIn is the lifetime in seconds of the device_code and user_code
	ExpiresIn int `json:"expires_in"`

	// Interval is the minimum amount of time in seconds to wait between polling requests
	Interval int `json:"interval,omitempty"`
}

// AuthorizeWithDeviceFlow performs the OAuth 2.0 Device Authorization Grant flow (RFC 8628).
// It requests a device code, displays the user code to the user, and polls for authorization.
func AuthorizeWithDeviceFlow(ctx context.Context, cfg *Config) (*oauth2.Token, error) {
	deviceAuthURL := cfg.DeviceAuthURL
	if deviceAuthURL == "" {
		// Fallback to AuthURL for backward compatibility
		deviceAuthURL = cfg.AuthURL
	}
	if deviceAuthURL == "" {
		return nil, fmt.Errorf("device authorization endpoint must be configured")
	}
	if cfg.TokenURL == "" {
		return nil, fmt.Errorf("token endpoint must be configured")
	}
	if cfg.ResourceURI == "" {
		return nil, fmt.Errorf("resource URI (MCP server endpoint) must be specified")
	}

	// Step 1: Request device code
	deviceResp, err := requestDeviceCode(ctx, cfg, deviceAuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Display user code to user
	displayUserCode(deviceResp)

	// Step 3: Poll for authorization
	token, err := pollForToken(ctx, cfg, deviceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %w", err)
	}

	fmt.Printf("✓ Authorization successful!\n\n")
	return token, nil
}

// requestDeviceCode requests a device code from the device authorization endpoint.
func requestDeviceCode(ctx context.Context, cfg *Config, deviceAuthURL string) (*DeviceCodeResponse, error) {
	// Determine scopes
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	// Build request body
	data := url.Values{
		"client_id": {cfg.ClientID},
		"scope":     {strings.Join(scopes, " ")},
	}

	// Add resource parameter per MCP OAuth specification
	if cfg.ResourceURI != "" {
		data.Set("resource", cfg.ResourceURI)
	}

	// Add client secret if provided (for confidential clients)
	if cfg.ClientSecret != "" {
		data.Set("client_secret", cfg.ClientSecret)
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("device authorization failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}

	return &deviceResp, nil
}

// displayUserCode displays the user code and verification URI to the user.
func displayUserCode(deviceResp *DeviceCodeResponse) {
	fmt.Printf("\n")
	fmt.Printf("╔════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║          Device Authorization Required                ║\n")
	fmt.Printf("╠════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║                                                        ║\n")
	fmt.Printf("║  Please visit:                                         ║\n")
	fmt.Printf("║  %s%-46s%s  ║\n", "\033[1;36m", deviceResp.VerificationURI, "\033[0m")
	fmt.Printf("║                                                        ║\n")
	fmt.Printf("║  And enter the code:                                   ║\n")
	fmt.Printf("║  %s%-46s%s  ║\n", "\033[1;33m", deviceResp.UserCode, "\033[0m")
	fmt.Printf("║                                                        ║\n")

	if deviceResp.VerificationURIComplete != "" {
		fmt.Printf("║  Or visit this URL directly:                           ║\n")
		fmt.Printf("║  %s%-46s%s  ║\n", "\033[1;32m", deviceResp.VerificationURIComplete, "\033[0m")
		fmt.Printf("║                                                        ║\n")
	}

	expiresIn := time.Duration(deviceResp.ExpiresIn) * time.Second
	fmt.Printf("║  Code expires in: %s%-35v%s  ║\n", "\033[1m", expiresIn, "\033[0m")
	fmt.Printf("║                                                        ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")
	fmt.Printf("Waiting for authorization...\n")
}

// pollForToken polls the token endpoint until the user authorizes or an error occurs.
func pollForToken(ctx context.Context, cfg *Config, deviceResp *DeviceCodeResponse) (*oauth2.Token, error) {
	// Determine polling interval (default 5 seconds per RFC 8628)
	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Check if device code expired
			if time.Now().After(expiresAt) {
				return nil, fmt.Errorf("device code expired")
			}

			// Attempt to get token
			token, err := attemptTokenExchange(ctx, cfg, deviceResp.DeviceCode)
			if err != nil {
				// Check for specific error codes
				if strings.Contains(err.Error(), "authorization_pending") {
					// Still waiting for user authorization
					continue
				}
				if strings.Contains(err.Error(), "slow_down") {
					// Server requested slower polling
					interval += 5 * time.Second
					ticker.Reset(interval)
					continue
				}
				// Other errors are fatal
				return nil, err
			}

			// Success!
			return token, nil
		}
	}
}

// attemptTokenExchange attempts to exchange the device code for an access token.
func attemptTokenExchange(ctx context.Context, cfg *Config, deviceCode string) (*oauth2.Token, error) {
	// Build request body
	data := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
		"client_id":   {cfg.ClientID},
	}

	// Add resource parameter per MCP OAuth specification
	if cfg.ResourceURI != "" {
		data.Set("resource", cfg.ResourceURI)
	}

	// Add client secret if provided (for confidential clients)
	if cfg.ClientSecret != "" {
		data.Set("client_secret", cfg.ClientSecret)
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error response
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description,omitempty"`
		}
		if jsonErr := json.Unmarshal(body, &errorResp); jsonErr == nil {
			if errorResp.ErrorDescription != "" {
				return nil, fmt.Errorf("%s: %s", errorResp.Error, errorResp.ErrorDescription)
			}
			return nil, fmt.Errorf("%s", errorResp.Error)
		}
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse successful token response
	var token oauth2.Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}
