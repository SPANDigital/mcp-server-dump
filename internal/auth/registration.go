package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RegisterClient performs Dynamic Client Registration (RFC 7591) with the authorization server.
// It registers the client and returns the client_id and optional client_secret.
func RegisterClient(ctx context.Context, registrationEndpoint, resourceURI string, scopes []string) (*ClientRegistration, error) {
	if registrationEndpoint == "" {
		return nil, fmt.Errorf("registration endpoint not provided")
	}

	// Determine grant types based on what we support
	grantTypes := []string{
		"urn:ietf:params:oauth:grant-type:device_code",
		"refresh_token",
	}

	// Build registration request
	// Some servers require redirect_uris even for device flow, so provide a standard loopback URI
	// Using http://localhost (without port) as per OAuth 2.1 native app best practices
	regRequest := ClientRegistrationRequest{
		ClientName:              "mcp-server-dump",
		RedirectURIs:            []string{"http://localhost"}, // Standard loopback URI for native apps
		GrantTypes:              grantTypes,
		TokenEndpointAuthMethod: "none", // Public client (no client secret required)
		Scope:                   strings.Join(scopes, " "),
	}

	// Marshal request
	reqBody, err := json.Marshal(regRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	// Make registration request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registrationEndpoint, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, fmt.Errorf("failed to read registration response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var regResponse ClientRegistrationResponse
	if err := json.Unmarshal(body, &regResponse); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	if regResponse.ClientID == "" {
		return nil, fmt.Errorf("registration response missing client_id")
	}

	// Create client registration record
	registration := &ClientRegistration{
		ResourceURI:             resourceURI,
		ClientID:                regResponse.ClientID,
		ClientSecret:            regResponse.ClientSecret,
		RegistrationAccessToken: regResponse.RegistrationAccessToken,
		RegisteredAt:            time.Now(),
	}

	return registration, nil
}

// LoadClientRegistration loads a cached client registration for the given resource URI.
func LoadClientRegistration(resourceURI string) (*ClientRegistration, error) {
	cacheDir, err := getRegistrationCacheDir()
	if err != nil {
		return nil, err
	}

	// Hash the resource URI to create filename
	hash := sha256.Sum256([]byte(resourceURI))
	filename := hex.EncodeToString(hash[:]) + ".json"
	filePath := filepath.Join(cacheDir, filename)

	// Check if file exists
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		return nil, nil // No cached registration
	}

	// Read file
	data, err := os.ReadFile(filePath) // #nosec G304 - file path is derived from hashed resource URI
	if err != nil {
		return nil, fmt.Errorf("failed to read registration cache: %w", err)
	}

	// Parse registration
	var registration ClientRegistration
	if err := json.Unmarshal(data, &registration); err != nil {
		return nil, fmt.Errorf("failed to parse registration cache: %w", err)
	}

	return &registration, nil
}

// SaveClientRegistration saves a client registration to the cache.
func SaveClientRegistration(registration *ClientRegistration) error {
	if registration == nil || registration.ResourceURI == "" {
		return fmt.Errorf("invalid registration")
	}

	cacheDir, err := getRegistrationCacheDir()
	if err != nil {
		return err
	}

	// Create cache directory if it doesn't exist
	if mkdirErr := os.MkdirAll(cacheDir, 0o700); mkdirErr != nil {
		return fmt.Errorf("failed to create registration cache directory: %w", mkdirErr)
	}

	// Hash the resource URI to create filename
	hash := sha256.Sum256([]byte(registration.ResourceURI))
	filename := hex.EncodeToString(hash[:]) + ".json"
	filePath := filepath.Join(cacheDir, filename)

	// Marshal registration
	data, err := json.MarshalIndent(registration, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	// Write file with secure permissions
	if writeErr := os.WriteFile(filePath, data, 0o600); writeErr != nil {
		return fmt.Errorf("failed to write registration cache: %w", writeErr)
	}

	return nil
}

// ClearClientRegistration removes a cached client registration for the given resource URI.
func ClearClientRegistration(resourceURI string) error {
	cacheDir, err := getRegistrationCacheDir()
	if err != nil {
		return err
	}

	// Hash the resource URI to create filename
	hash := sha256.Sum256([]byte(resourceURI))
	filename := hex.EncodeToString(hash[:]) + ".json"
	filePath := filepath.Join(cacheDir, filename)

	// Remove file (ignore error if file doesn't exist)
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove registration cache: %w", err)
	}

	return nil
}

// getRegistrationCacheDir returns the directory for cached client registrations.
func getRegistrationCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use ~/.config/mcp-server-dump/registrations/
	cacheDir := filepath.Join(homeDir, ".config", "mcp-server-dump", "registrations")
	return cacheDir, nil
}

// GetOrRegisterClient gets a cached client registration or performs DCR if needed.
// This is the main entry point for automatic client registration.
func GetOrRegisterClient(ctx context.Context, resourceURI, registrationEndpoint string, scopes []string) (*ClientRegistration, error) {
	// Try to load from cache first
	cached, err := LoadClientRegistration(resourceURI)
	if err != nil {
		// Log error but continue with registration
		fmt.Fprintf(os.Stderr, "Warning: failed to load cached registration: %v\n", err)
	}

	if cached != nil && cached.ClientID != "" {
		// Found cached registration
		return cached, nil
	}

	// No cached registration - perform DCR
	if registrationEndpoint == "" {
		return nil, fmt.Errorf("no cached registration and server does not support Dynamic Client Registration")
	}

	fmt.Printf("Registering client with authorization server...\n")
	registration, err := RegisterClient(ctx, registrationEndpoint, resourceURI, scopes)
	if err != nil {
		return nil, fmt.Errorf("dynamic client registration failed: %w", err)
	}

	fmt.Printf("✓ Client registered successfully\n")
	fmt.Printf("  Client ID: %s\n", registration.ClientID)

	// Save to cache
	if saveErr := SaveClientRegistration(registration); saveErr != nil {
		// Log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to cache registration: %v\n", saveErr)
	}

	return registration, nil
}
