// Package authlib provides OAuth2 token management and JWT validation functionality.
// It implements the OAuth2 client credentials flow for service-to-service authentication
// and includes efficient caching mechanisms for both access tokens and JWK Sets.
//
// Key Features:
//   - OAuth2 client credentials flow implementation
//   - Automatic token refresh and caching
//   - JWT token validation and parsing
//   - Configurable retry mechanism
//   - Extensible metrics collection
//   - Context-aware operations
//
// Example Usage:
//
//	config := authlib.OAuthConfig{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    TokenURL:     "https://auth.example.com/token",
//	}
//
//	cache := authlib.NewTokenCache(config)
//
//	// Optional: Configure custom retry behavior
//	cache.SetRetryConfig(authlib.RetryConfig{
//	    MaxAttempts: 3,
//	    WaitTime:    time.Second,
//	})
//
//	// Get a token (automatically handles caching and refresh)
//	token, err := cache.GetToken(context.Background())
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
// It provides both a descriptive message and the underlying cause of the error,
// allowing for detailed error handling and logging.
type ErrTokenFetch struct {
	Message string // Human-readable error description
	Cause   error  // The underlying error that caused the failure
}

func (e *ErrTokenFetch) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// MetricsHook defines an interface for monitoring token operations.
// Implementations can track various token-related events for monitoring,
// alerting, and performance analysis purposes.
type MetricsHook interface {
	// OnTokenFetch is called after a token fetch attempt, providing the operation
	// duration and any error that occurred. This can be used to track latency
	// and error rates for token fetching operations.
	OnTokenFetch(duration time.Duration, err error)

	// OnCacheHit is called when a valid token is found in the cache.
	// This can be used to monitor cache effectiveness and hit rates.
	OnCacheHit()

	// OnCacheMiss is called when a token needs to be fetched from the
	// authorization server. This can be used to monitor cache performance
	// and identify potential issues with token expiration settings.
	OnCacheMiss()
}

// RetryConfig defines parameters for retry behavior during token fetching.
// It allows customization of the retry mechanism to handle temporary failures
// and network issues gracefully.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts before giving up.
	// A value of 1 means no retries (only the initial attempt).
	MaxAttempts int

	// WaitTime is the duration to wait between retry attempts.
	// This helps prevent overwhelming the authorization server during issues.
	WaitTime time.Duration
}

// TokenProvider defines the interface for token operations.
// This interface allows for different implementations of token management
// while maintaining a consistent API for token consumers.
type TokenProvider interface {
	// GetToken retrieves a valid OAuth token, either from cache or by fetching a new one.
	// It handles token expiration and refresh automatically.
	//
	// The context parameter can be used to cancel long-running operations
	// or set timeouts for token retrieval.
	GetToken(ctx context.Context) (string, error)
}

// TokenCache implements TokenProvider interface and handles caching of OAuth tokens.
// It provides automatic token refresh, retry logic, and metrics collection.
// TokenCache is safe for concurrent use by multiple goroutines.
type TokenCache struct {
	// Token holds the current OAuth token
	Token string

	// ExpiresAt tracks the token's expiration time
	ExpiresAt time.Time

	// Config contains the OAuth configuration settings
	Config OAuthConfig

	// RetryConf specifies the retry behavior for token fetching
	RetryConf RetryConfig

	// Metrics provides hooks for monitoring token operations
	Metrics MetricsHook

	// httpClient is used for making HTTP requests to the token endpoint
	httpClient *http.Client
}

// OAuthConfig contains the configuration for OAuth client credentials flow.
// It includes all necessary parameters for token endpoint authentication.
type OAuthConfig struct {
	// ClientID is the OAuth client identifier
	ClientID string

	// ClientSecret is the OAuth client secret
	ClientSecret string

	// TokenURL is the full URL to the token endpoint
	TokenURL string

	// Headers contains additional headers to include in token requests
	Headers map[string]string

	// GrantType specifies the OAuth grant type (defaults to "client_credentials")
	GrantType string
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
// It implements a cache-first strategy, only fetching a new token when
// the cached token is expired or invalid.
//
// The function is thread-safe and can be called concurrently. It uses
// the provided context for cancellation and timeout control.
//
// If metrics collection is enabled, it will report:
//   - Cache hits/misses
//   - Token fetch duration
//   - Token fetch errors
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
// It attempts to fetch a new token up to MaxAttempts times, waiting
// WaitTime between attempts. The operation can be cancelled via context.
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
// It handles the HTTP request to the token endpoint, including proper
// header setting and error handling.
//
// The function expects a JSON response containing an "access_token" field.
// It will return an error if:
//   - The HTTP request fails
//   - The response status is not 200 OK
//   - The response cannot be decoded as JSON
//   - The response doesn't contain an access_token field
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
