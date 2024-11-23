// Package authlib provides OAuth2 token management and JWT validation functionality.
// It implements client credentials flow for OAuth2 authentication and includes
// caching mechanisms for both access tokens and JWK Sets.
package authlib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
)

// ErrTokenFetch represents an error that occurs during token fetching operations.
// It includes both a descriptive message and the underlying cause of the error.
type ErrTokenFetch struct {
	Message string
	Cause   error
}

func (e *ErrTokenFetch) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// MetricsHook defines an interface for monitoring token operations.
// Implementations can track token fetches, cache hits, and cache misses.
type MetricsHook interface {
	// OnTokenFetch is called after a token fetch attempt with the duration and any error
	OnTokenFetch(duration time.Duration, err error)
	// OnCacheHit is called when a valid token is found in cache
	OnCacheHit()
	// OnCacheMiss is called when a token needs to be fetched
	OnCacheMiss()
}

// RetryConfig defines parameters for retry behavior during token fetching.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int
	// WaitTime is the duration to wait between retry attempts
	WaitTime    time.Duration
}

// TokenProvider defines the interface for token operations.
// Implementations should handle token retrieval and caching.
type TokenProvider interface {
	// GetToken retrieves a valid OAuth token, either from cache or by fetching a new one
	GetToken(ctx context.Context) (string, error)
}

// TokenCache implements TokenProvider interface and handles caching of OAuth tokens.
// It automatically refreshes expired tokens and implements retry logic for token fetching.
type TokenCache struct {
	Token      string        // The current OAuth token
	ExpiresAt  time.Time    // Expiration time of the current token
	Config     OAuthConfig  // OAuth configuration settings
	RetryConf  RetryConfig // Retry behavior configuration
	Metrics    MetricsHook // Optional metrics collection
	httpClient *http.Client
}

// OAuthConfig contains the configuration for OAuth client credentials flow.
// It includes all necessary parameters for token endpoint authentication.
type OAuthConfig struct {
	ClientID     string            // OAuth client identifier
	ClientSecret string            // OAuth client secret
	TokenURL     string            // Full URL to the token endpoint
	Headers      map[string]string // Additional headers for token request
	GrantType    string           // OAuth grant type, defaults to "client_credentials"
}

// NewTokenCache creates a new TokenCache instance.
func NewTokenCache(config OAuthConfig) *TokenCache {
	if config.GrantType == "" {
		config.GrantType = "client_credentials"
	}
	return &TokenCache{
		Config: config,
		RetryConf: RetryConfig{
			MaxAttempts: 3,
			WaitTime:    time.Second,
		},
		httpClient: &http.Client{},
	}
}

// SetHTTPClient allows setting a custom HTTP client.
func (cache *TokenCache) SetHTTPClient(client *http.Client) {
	cache.httpClient = client
}

// SetMetricsHook sets the metrics hook for monitoring.
func (cache *TokenCache) SetMetricsHook(hook MetricsHook) {
	cache.Metrics = hook
}

// SetRetryConfig configures retry behavior.
func (cache *TokenCache) SetRetryConfig(config RetryConfig) {
	cache.RetryConf = config
}

// GetToken handles getting, validating, and caching the JWT token.
func (cache *TokenCache) GetToken(ctx context.Context) (string, error) {
	// If the token is already in cache and it's not expired, return it
	if cache.Token != "" && time.Now().Before(cache.ExpiresAt) {
		if cache.Metrics != nil {
			cache.Metrics.OnCacheHit()
		}
		log.Println("Token is valid and in cache")
		return cache.Token, nil
	}

	if cache.Metrics != nil {
		cache.Metrics.OnCacheMiss()
	}

	// If the token doesn't exist or is expired, get a new one
	log.Println("Fetching new OAuth token")
	start := time.Now()
	token, err := cache.fetchNewTokenWithRetry(ctx)
	if cache.Metrics != nil {
		cache.Metrics.OnTokenFetch(time.Since(start), err)
	}
	if err != nil {
		return "", &ErrTokenFetch{Message: "failed to fetch new token", Cause: err}
	}

	// Decode the JWT token to get the expiration time
	parsedToken, err := jwt.ParseString(token["access_token"].(string))
	if err != nil {
		return "", &ErrTokenFetch{Message: "failed to parse JWT token", Cause: err}
	}

	// Get the expiration time from the JWT claims
	exp := parsedToken.Expiration()
	if exp.IsZero() {
		return "", &ErrTokenFetch{Message: "JWT token has no expiration time"}
	}

	cache.ExpiresAt = exp
	cache.Token = token["access_token"].(string)

	return cache.Token, nil
}

// fetchNewTokenWithRetry implements retry logic for token fetching.
func (cache *TokenCache) fetchNewTokenWithRetry(ctx context.Context) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < cache.RetryConf.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if token, err := cache.fetchNewToken(ctx); err == nil {
				return token, nil
			} else {
				lastErr = err
				// Wait before retry, unless it's the last attempt
				if attempt < cache.RetryConf.MaxAttempts-1 {
					time.Sleep(cache.RetryConf.WaitTime)
				}
			}
		}
	}
	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// fetchNewToken retrieves a new token from the authorization server.
func (cache *TokenCache) fetchNewToken(ctx context.Context) (map[string]interface{}, error) {
	config := cache.Config
	data := fmt.Sprintf("grant_type=%s&client_id=%s&client_secret=%s",
		config.GrantType, config.ClientID, config.ClientSecret)

	// Create the POST request with context
	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, bytes.NewBufferString(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default Content-Type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Set additional headers if provided
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := cache.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read and decode the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result["access_token"] == nil {
		return nil, fmt.Errorf("response does not contain access_token")
	}

	return result, nil
}
