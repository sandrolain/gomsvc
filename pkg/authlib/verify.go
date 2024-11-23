// Package authlib provides OAuth2 token management and JWT validation functionality.
// This file implements JSON Web Key (JWK) Set caching and JWT token validation,
// supporting automatic key rotation and various validation options.
//
// Key Features:
//   - JWK Set caching with automatic refresh
//   - JWT validation with configurable options
//   - Support for key rotation
//   - Metrics collection for monitoring
//   - Retry mechanisms for resilience
//
// Example Usage:
//
//	config := authlib.JWKConfig{
//	    JWKSURL:        "https://auth.example.com/.well-known/jwks.json",
//	    ExpirationTime: 24 * time.Hour,
//	}
//	
//	// Create JWK cache with automatic refresh
//	jwkCache := authlib.NewJWKCache(config)
//	
//	// Create validator with required options
//	validator := authlib.NewTokenValidator(jwkCache,
//	    jwt.WithIssuer("https://auth.example.com"),
//	    jwt.WithAudience("your-app"),
//	)
//	
//	// Validate a token
//	token, claims, err := validator.ValidateToken(context.Background(), "your-jwt-token")
package authlib

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

// ErrJWKFetch represents an error that occurs during JWK (JSON Web Key) fetching operations.
// It provides both a descriptive message and the underlying cause of the error,
// allowing for detailed error handling and logging.
type ErrJWKFetch struct {
	Message string // Human-readable error description
	Cause   error  // The underlying error that caused the failure
}

func (e *ErrJWKFetch) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// ErrTokenValidation represents an error that occurs during JWT token validation.
// It provides detailed information about validation failures, including specific
// validation rules that failed (e.g., expired token, invalid signature, wrong issuer).
type ErrTokenValidation struct {
	Message string // Human-readable error description
	Cause   error  // The underlying error that caused the validation failure
}

func (e *ErrTokenValidation) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// JWKMetricsHook defines an interface for monitoring JWK operations.
// Implementations can track various JWK-related events for monitoring,
// alerting, and performance analysis purposes.
type JWKMetricsHook interface {
	// OnJWKFetch is called after a JWK fetch attempt, providing the operation
	// duration and any error that occurred. This can be used to track latency
	// and error rates for JWK fetching operations.
	OnJWKFetch(duration time.Duration, err error)

	// OnJWKCacheHit is called when a valid JWK Set is found in the cache.
	// This can be used to monitor cache effectiveness and hit rates.
	OnJWKCacheHit()

	// OnJWKCacheMiss is called when a JWK Set needs to be fetched from
	// the authorization server. This can be used to monitor cache performance
	// and identify potential issues with expiration settings.
	OnJWKCacheMiss()
}

// KeyProvider defines the interface for JWK operations.
// This interface allows for different implementations of JWK Set management
// while maintaining a consistent API for token validation.
type KeyProvider interface {
	// FetchKeys retrieves a JWK Set, either from cache or by fetching from the JWKS endpoint.
	// The implementation should handle caching, refresh, and error recovery strategies.
	//
	// The context parameter can be used to cancel long-running operations
	// or set timeouts for key retrieval.
	FetchKeys(ctx context.Context) (jwk.Set, error)
}

// JWKCache implements KeyProvider interface and handles caching of JWK Sets.
// It provides automatic refresh of expired keys and implements retry logic
// for resilient key fetching. JWKCache is safe for concurrent use.
type JWKCache struct {
	// The current JWK Set
	jwkSet jwk.Set

	// Expiration time of the current JWK Set
	expiresAt time.Time

	// JWK configuration settings
	config JWKConfig

	// HTTP client for making requests
	httpClient *http.Client

	// Optional metrics collection
	metrics JWKMetricsHook

	// Retry behavior configuration
	retryConf RetryConfig
}

// JWKConfig contains configuration for JWK fetching and validation.
// It includes all necessary parameters for JWK Set endpoint access and caching behavior.
type JWKConfig struct {
	// JWKSURL is the URL of the JWKS (JSON Web Key Set) endpoint
	JWKSURL string

	// Headers contains additional headers to include in JWK requests
	Headers map[string]string

	// ExpirationTime specifies how long to cache the JWK Set
	// If not set, defaults to 24 hours
	ExpirationTime time.Duration
}

// NewJWKCache creates a new JWKCache instance.
func NewJWKCache(config JWKConfig) *JWKCache {
	if config.ExpirationTime == 0 {
		config.ExpirationTime = 24 * time.Hour // Default to 24 hours
	}
	return &JWKCache{
		config:     config,
		httpClient: &http.Client{},
		retryConf: RetryConfig{
			MaxAttempts: 3,
			WaitTime:    time.Second,
		},
	}
}

// SetHTTPClient allows setting a custom HTTP client.
func (cache *JWKCache) SetHTTPClient(client *http.Client) {
	cache.httpClient = client
}

// SetMetricsHook sets the metrics hook for monitoring.
func (cache *JWKCache) SetMetricsHook(hook JWKMetricsHook) {
	cache.metrics = hook
}

// SetRetryConfig configures retry behavior.
func (cache *JWKCache) SetRetryConfig(config RetryConfig) {
	cache.retryConf = config
}

// FetchKeys fetches the JWK Set and caches it.
func (cache *JWKCache) FetchKeys(ctx context.Context) (jwk.Set, error) {
	// If the JWK is in cache and it's still valid, return it
	if cache.jwkSet != nil && time.Now().Before(cache.expiresAt) {
		if cache.metrics != nil {
			cache.metrics.OnJWKCacheHit()
		}
		log.Println("JWK Set is valid and in cache")
		return cache.jwkSet, nil
	}

	if cache.metrics != nil {
		cache.metrics.OnJWKCacheMiss()
	}

	// If the JWK is not in cache or has expired, fetch it
	log.Println("Fetching JWK Set from authorization server")
	start := time.Now()
	jwkSet, err := cache.fetchJWKSetWithRetry(ctx)
	if cache.metrics != nil {
		cache.metrics.OnJWKFetch(time.Since(start), err)
	}
	if err != nil {
		return nil, &ErrJWKFetch{Message: "failed to fetch JWK Set", Cause: err}
	}

	// Set cache expiration time
	cache.expiresAt = time.Now().Add(cache.config.ExpirationTime)
	cache.jwkSet = jwkSet

	return cache.jwkSet, nil
}

// fetchJWKSetWithRetry implements retry logic for JWK fetching.
func (cache *JWKCache) fetchJWKSetWithRetry(ctx context.Context) (jwk.Set, error) {
	var lastErr error
	for attempt := 0; attempt < cache.retryConf.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if jwkSet, err := cache.fetchJWKSet(ctx); err == nil {
				return jwkSet, nil
			} else {
				lastErr = err
				// Wait before retry, unless it's the last attempt
				if attempt < cache.retryConf.MaxAttempts-1 {
					time.Sleep(cache.retryConf.WaitTime)
				}
			}
		}
	}
	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// fetchJWKSet retrieves the JWK Set from the authorization server.
func (cache *JWKCache) fetchJWKSet(ctx context.Context) (jwk.Set, error) {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", cache.config.JWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set additional headers if provided
	for key, value := range cache.config.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := cache.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWK Set: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read and parse response
	jwkSet, err := jwk.ParseReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWK Set: %w", err)
	}

	return jwkSet, nil
}

// TokenValidator handles JWT validation using JWK Sets.
// It supports various validation options and automatic key rotation through
// its KeyProvider implementation. The validator is safe for concurrent use.
type TokenValidator struct {
	// Provider for JWK Sets, handles key fetching and rotation
	keyProvider KeyProvider

	// JWT validation options (e.g., issuer, audience, time validation)
	options []jwt.ValidateOption
}

// NewTokenValidator creates a new TokenValidator instance.
// It accepts a KeyProvider for JWK management and optional JWT validation options.
//
// Common validation options include:
//   - jwt.WithIssuer("https://issuer")
//   - jwt.WithAudience("audience")
//   - jwt.WithTime(time.Now())
//
// Example:
//
//	validator := NewTokenValidator(jwkCache,
//	    jwt.WithIssuer("https://auth.example.com"),
//	    jwt.WithAudience("your-app"),
//	)
func NewTokenValidator(keyProvider KeyProvider, options ...jwt.ValidateOption) *TokenValidator {
	return &TokenValidator{
		keyProvider: keyProvider,
		options:     options,
	}
}

// ValidateToken verifies the JWT with the fetched JWK and returns the token and claims.
// It performs complete token validation including:
//   - Signature verification using JWK Set
//   - Token parsing and format validation
//   - Claims validation (exp, iat, nbf, iss, aud, etc.)
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - tokenString: The JWT token string to validate
//
// Returns:
//   - jwt.Token: The parsed and validated token
//   - map[string]interface{}: The token's claims
//   - error: Any validation error that occurred
//
// Security Considerations:
//   - Always verify the token signature before using claims
//   - Use appropriate validation options (issuer, audience, etc.)
//   - Consider token expiration and clock skew
//
// Example:
//
//	token, claims, err := validator.ValidateToken(ctx, "eyJhbGciOiJSUzI1...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if sub, ok := claims["sub"].(string); ok {
//	    fmt.Printf("Token subject: %s\n", sub)
//	}
func (v *TokenValidator) ValidateToken(ctx context.Context, tokenString string) (jwt.Token, map[string]interface{}, error) {
	jwkSet, err := v.keyProvider.FetchKeys(ctx)
	if err != nil {
		return nil, nil, &ErrTokenValidation{Message: "failed to fetch keys", Cause: err}
	}

	// Parse and verify the JWT
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(jwkSet),
	)
	if err != nil {
		return nil, nil, &ErrTokenValidation{Message: "failed to parse token", Cause: err}
	}

	// Validate the token with provided options
	if err := jwt.Validate(token, v.options...); err != nil {
		return nil, nil, &ErrTokenValidation{Message: "token validation failed", Cause: err}
	}

	// Extract claims
	claims := make(map[string]interface{})
	for iter := token.Iterate(ctx); iter.Next(ctx); {
		pair := iter.Pair()
		claims[pair.Key.(string)] = pair.Value
	}

	return token, claims, nil
}
