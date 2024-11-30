// Package authlib provides OAuth2 token management and JWT validation functionality.
// It implements the OAuth2 client credentials flow for service-to-service authentication
// and includes efficient caching mechanisms for both access tokens and JWK Sets.
//
// Key Features:
//   - OAuth2 client credentials flow implementation
//   - Automatic token refresh and caching
//   - JWT token validation and parsing
//   - Configurable retry mechanism
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

	"github.com/eapache/go-resiliency/retrier"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

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

// TokenCache implements TokenProvider interface and handles caching of OAuth tokens.
// It provides automatic token refresh and retry logic collection.
// TokenCache is safe for concurrent use by multiple goroutines.
type TokenCache struct {
	// Token holds the current OAuth token
	Token string

	// ExpiresAt tracks the token's expiration time
	ExpiresAt time.Time

	// Config contains the OAuth configuration settings
	Config OAuthConfig

	// httpClient is used for making HTTP requests to the token endpoint
	httpClient *http.Client

	jwks         jwk.Set
	jwkCache     *jwk.Cache
	JWTExpiresAt time.Time
}

// OAuthConfig contains the configuration for OAuth client credentials flow.
// It includes all necessary parameters for token endpoint authentication.
type OAuthConfig struct {
	RetryConfig RetryConfig

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

	MaxAttempts int

	JWKURL string

	JWKExpirationTime time.Duration
}

// NewTokenCache creates a new TokenCache instance.
func NewTokenCache(config OAuthConfig) *TokenCache {
	if config.GrantType == "" {
		config.GrantType = "client_credentials"
	}
	return &TokenCache{
		Config:     config,
		httpClient: &http.Client{},
	}
}

// GetToken handles getting, validating, and caching the JWT token.
// It implements a cache-first strategy, only fetching a new token when
// the cached token is expired or invalid.
//
// The function is thread-safe and can be called concurrently. It uses
// the provided context for cancellation and timeout control.
func (cache *TokenCache) GetToken(ctx context.Context) (string, error) {
	// If the token is already in cache and it's not expired, return it
	if cache.Token != "" && time.Now().Before(cache.ExpiresAt) {
		log.Println("Token is valid and in cache")
		return cache.Token, nil
	}

	// If the token doesn't exist or is expired, get a new one
	log.Println("Fetching new OAuth token")
	token, err := cache.fetchNewTokenWithRetry(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}

	accessToken := token["access_token"].(string)

	jwkSet, err := cache.FetchJWK(ctx)
	if err != nil {
		log.Printf("Error fetching JWK: %v\n", err)
		return "", err
	}

	// Decode the JWT token to get the expiration time
	parsedToken, err := jwt.ParseString(accessToken, jwt.WithKeySet(jwkSet))
	if err != nil {
		return "", fmt.Errorf("error: unable to parse JWT token: %w", err)
	}

	// Get the expiration time from the JWT claims
	exp := parsedToken.Expiration()
	if exp.IsZero() {
		return "", fmt.Errorf("JWT token has no expiration time")
	}

	cache.ExpiresAt = exp
	cache.Token = accessToken

	return accessToken, nil
}

// fetchNewTokenWithRetry implements retry logic for token fetching.
// It attempts to fetch a new token up to MaxAttempts times, waiting
// WaitTime between attempts. The operation can be cancelled via context.
func (cache *TokenCache) fetchNewTokenWithRetry(ctx context.Context) (map[string]interface{}, error) {
	retryConfig := cache.Config.RetryConfig
	if retryConfig.MaxAttempts == 0 {
		retryConfig.MaxAttempts = 1
	}
	if retryConfig.WaitTime == 0 {
		retryConfig.WaitTime = time.Second
	}

	var token map[string]interface{}

	r := retrier.New(retrier.ConstantBackoff(retryConfig.MaxAttempts, retryConfig.WaitTime), nil)
	err := r.Run(func() error {
		var e error
		token, e = cache.fetchNewToken(ctx)
		return e
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch token after %d attempts: %w", retryConfig.MaxAttempts, err)
	}

	return token, nil
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

// FetchJWK fetches the JWK from Keycloak and caches it
func (cache *TokenCache) FetchJWK(ctx context.Context) (jwk.Set, error) {
	url := cache.Config.JWKURL

	if cache.jwkCache == nil {
		c := jwk.NewCache(ctx)

		err := c.Register(url, jwk.WithMinRefreshInterval(cache.Config.JWKExpirationTime))
		if err != nil {
			return nil, fmt.Errorf("error registering JWKS URL: %w", err)
		}

		jwkSet, err := c.Refresh(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("error refreshing JWKS: %w", err)
		}
		cache.jwkCache = c
		cache.jwks = jwkSet

		return cache.jwks, nil
	}

	return cache.jwkCache.Get(ctx, url)
}

// VerifyJWT validates a JWT token using the configured JWK set and returns the parsed token and claims.
//
// This method performs the following steps:
// 1. Fetches the JWK set from the configured JWKURL (with caching)
// 2. Parses and validates the JWT token using the JWK set
// 3. Extracts both standard and private claims from the token
//
// Standard claims extracted include:
//   - sub (Subject)
//   - iss (Issuer)
//   - aud (Audience)
//   - exp (Expiration Time)
//   - nbf (Not Before)
//   - iat (Issued At)
//   - jti (JWT ID)
//
// Parameters:
//   - ctx: Context for the operation, which can be used for cancellation
//   - jwtToken: The JWT token string to verify
//
// Returns:
//   - jwt.Token: The parsed JWT token object
//   - map[string]interface{}: Combined map of standard and private claims
//   - error: Any error that occurred during verification
//
// The method will return an error if:
//   - JWK set cannot be fetched
//   - JWT token is invalid or malformed
//   - Token signature verification fails
//   - Token validation fails (e.g., expired token)
func (cache *TokenCache) VerifyJWT(ctx context.Context, jwtToken string) (jwt.Token, map[string]interface{}, error) {
	jwkSet, err := cache.FetchJWK(ctx)
	if err != nil {
		log.Printf("Error fetching JWK: %v\n", err)
		return nil, nil, err
	}

	// Parse and verify the JWT
	token, err := jwt.Parse([]byte(jwtToken), jwt.WithKeySet(jwkSet))
	if err != nil {
		log.Printf("Error parsing JWT: %v\n", err)
		return nil, nil, err
	}

	err = jwt.Validate(token)
	if err != nil {
		log.Printf("JWT validation failed: %v\n", err)
		return nil, nil, err
	}

	claims := make(map[string]interface{})
	// Add standard claims
	if sub, ok := token.Get(jwt.SubjectKey); ok {
		claims["sub"] = sub
	}
	if iss, ok := token.Get(jwt.IssuerKey); ok {
		claims["iss"] = iss
	}
	if aud, ok := token.Get(jwt.AudienceKey); ok {
		claims["aud"] = aud
	}
	if exp, ok := token.Get(jwt.ExpirationKey); ok {
		claims["exp"] = exp
	}
	if nbf, ok := token.Get(jwt.NotBeforeKey); ok {
		claims["nbf"] = nbf
	}
	if iat, ok := token.Get(jwt.IssuedAtKey); ok {
		claims["iat"] = iat
	}
	if jti, ok := token.Get(jwt.JwtIDKey); ok {
		claims["jti"] = jti
	}

	// Add private claims
	for key, value := range token.PrivateClaims() {
		claims[key] = value
	}

	return token, claims, nil
}
