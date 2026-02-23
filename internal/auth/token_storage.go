package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// tokenCacheDir returns the directory path for token cache files.
// Uses XDG Base Directory specification: ~/.config/mcp-server-dump/tokens/
func tokenCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use XDG Base Directory: ~/.config/mcp-server-dump/tokens/
	cacheDir := filepath.Join(homeDir, ".config", "mcp-server-dump", "tokens")
	return cacheDir, nil
}

// tokenCachePath returns the file path for a cached token based on the resource URI.
// Tokens are stored as: ~/.config/mcp-server-dump/tokens/<hash>.json
func tokenCachePath(resourceURI string) (string, error) {
	dir, err := tokenCacheDir()
	if err != nil {
		return "", err
	}

	// Create hash of resource URI for filename
	hash := sha256.Sum256([]byte(resourceURI))
	filename := hex.EncodeToString(hash[:8]) + ".json" // Use first 8 bytes of hash

	return filepath.Join(dir, filename), nil
}

// LoadToken loads a cached OAuth token for the specified MCP server endpoint.
// Returns nil if no cached token exists or if the token file is invalid.
func LoadToken(resourceURI string) (*TokenCache, error) {
	cachePath, err := tokenCachePath(resourceURI)
	if err != nil {
		return nil, err
	}

	// Check if cache file exists
	if _, statErr := os.Stat(cachePath); os.IsNotExist(statErr) {
		return nil, nil // No cached token
	}

	// Read cache file
	// #nosec G304 - cachePath is generated from user-controlled resourceURI which is intentional
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token cache: %w", err)
	}

	// Parse JSON
	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse token cache: %w", err)
	}

	// Validate resource URI matches
	if cache.ResourceURI != resourceURI {
		return nil, fmt.Errorf("token cache resource URI mismatch")
	}

	return &cache, nil
}

// SaveToken saves an OAuth token to the cache for the specified MCP server endpoint.
// Creates the cache directory if it doesn't exist and sets appropriate file permissions (0600).
func SaveToken(token *oauth2.Token, resourceURI string, scopes []string) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	// Create cache directory if it doesn't exist
	cacheDir, err := tokenCacheDir()
	if err != nil {
		return err
	}

	if mkdirErr := os.MkdirAll(cacheDir, 0o700); mkdirErr != nil {
		return fmt.Errorf("failed to create token cache directory: %w", mkdirErr)
	}

	// Convert to TokenCache
	cache := FromOAuth2Token(token, resourceURI, scopes)

	// Marshal to JSON
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}

	// Get cache file path
	cachePath, err := tokenCachePath(resourceURI)
	if err != nil {
		return err
	}

	// Write to file with restricted permissions (owner read/write only)
	if err := os.WriteFile(cachePath, data, 0o600); err != nil { //nolint:gosec // G703: path is constructed from SHA256 hash of resource URI, not directly from user input
		return fmt.Errorf("failed to write token cache: %w", err)
	}

	return nil
}

// ClearToken removes the cached token for the specified MCP server endpoint.
func ClearToken(resourceURI string) error {
	cachePath, err := tokenCachePath(resourceURI)
	if err != nil {
		return err
	}

	// Remove cache file (ignore if doesn't exist)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token cache: %w", err)
	}

	return nil
}

// ClearAllTokens removes all cached tokens.
func ClearAllTokens() error {
	cacheDir, err := tokenCacheDir()
	if err != nil {
		return err
	}

	// Remove entire tokens directory
	if err := os.RemoveAll(cacheDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token cache directory: %w", err)
	}

	return nil
}

// ListCachedServers returns a list of resource URIs that have cached tokens.
func ListCachedServers() ([]string, error) {
	cacheDir, err := tokenCacheDir()
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, statErr := os.Stat(cacheDir); os.IsNotExist(statErr) {
		return []string{}, nil // No cache directory means no tokens
	}

	// Read directory
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read token cache directory: %w", err)
	}

	servers := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Read token file to get resource URI
		cachePath := filepath.Join(cacheDir, entry.Name())
		// #nosec G304 - cachePath is from directory listing, controlled by filesystem
		data, fileErr := os.ReadFile(cachePath)
		if fileErr != nil {
			continue // Skip invalid files
		}

		var cache TokenCache
		if err := json.Unmarshal(data, &cache); err != nil {
			continue // Skip invalid JSON
		}

		servers = append(servers, cache.ResourceURI)
	}

	return servers, nil
}
