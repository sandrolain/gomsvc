// Package authlib provides OAuth2 token management and JWT validation functionality.
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
// It includes both a descriptive message and the underlying cause of the error.
type ErrJWKFetch struct {
	Message string
	Cause   error
}

func (e *ErrJWKFetch) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// ErrTokenValidation represents an error that occurs during JWT token validation.
// It includes both a descriptive message and the underlying cause of the error.
type ErrTokenValidation struct {
	Message string
	Cause   error
}

func (e *ErrTokenValidation) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// JWKMetricsHook defines an interface for monitoring JWK operations.
// Implementations can track JWK fetches, cache hits, and cache misses.
type JWKMetricsHook interface {
	// OnJWKFetch is called after a JWK fetch attempt with the duration and any error
	OnJWKFetch(duration time.Duration, err error)
	// OnJWKCacheHit is called when a valid JWK Set is found in cache
	OnJWKCacheHit()
	// OnJWKCacheMiss is called when a JWK Set needs to be fetched
	OnJWKCacheMiss()
}

// KeyProvider defines the interface for JWK operations.
// Implementations should handle JWK Set retrieval and caching.
type KeyProvider interface {
	// FetchKeys retrieves a JWK Set, either from cache or by fetching from the JWKS endpoint
	FetchKeys(ctx context.Context) (jwk.Set, error)
}

// JWKCache implements KeyProvider interface and handles caching of JWK Sets.
// It automatically refreshes expired JWK Sets and implements retry logic for JWK fetching.
type JWKCache struct {
	jwkSet     jwk.Set          // The current JWK Set
	expiresAt  time.Time        // Expiration time of the current JWK Set
	config     JWKConfig        // JWK configuration settings
	httpClient *http.Client     // HTTP client for making requests
	metrics    JWKMetricsHook   // Optional metrics collection
	retryConf  RetryConfig      // Retry behavior configuration
}

// JWKConfig contains configuration for JWK fetching and validation.
// It includes all necessary parameters for JWK Set endpoint access.
type JWKConfig struct {
	JWKSURL        string            // URL to fetch JWK Set from
	Headers        map[string]string // Additional headers for JWK request
	ExpirationTime time.Duration     // How long to cache the JWK Set
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
// It supports various validation options and automatic key rotation.
type TokenValidator struct {
	keyProvider KeyProvider           // Provider for JWK Sets
	options     []jwt.ValidateOption // JWT validation options
}

// NewTokenValidator creates a new TokenValidator instance.
func NewTokenValidator(keyProvider KeyProvider, options ...jwt.ValidateOption) *TokenValidator {
	return &TokenValidator{
		keyProvider: keyProvider,
		options:     options,
	}
}

// ValidateToken verifies the JWT with the fetched JWK and returns the token and claims.
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
