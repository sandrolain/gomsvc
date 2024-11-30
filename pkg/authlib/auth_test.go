package authlib

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupJWKServer(t *testing.T, key jwk.Key) *httptest.Server {
	set := jwk.NewSet()
	err := set.AddKey(key)
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(set)
	}))
}

func createTestJWT(t *testing.T, key jwk.Key) string {
	token := jwt.New()
	now := time.Now()
	token.Set(jwt.ExpirationKey, now.Add(time.Hour))
	token.Set(jwt.IssuedAtKey, now)
	token.Set(jwt.SubjectKey, "test-subject")

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	require.NoError(t, err)

	return string(signedToken)
}

type mockTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func setupMockServer(t *testing.T, response interface{}, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
}

func generateRSAKey(t *testing.T) jwk.Key {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	key, err := jwk.FromRaw(privateKey)
	require.NoError(t, err)

	err = key.Set(jwk.KeyIDKey, "test-key")
	require.NoError(t, err)

	err = key.Set(jwk.AlgorithmKey, jwa.RS256)
	require.NoError(t, err)

	return key
}

func TestNewTokenCache(t *testing.T) {
	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     "http://example.com/token",
		GrantType:    "client_credentials",
	}

	cache := NewTokenCache(config)
	assert.NotNil(t, cache)
	assert.Equal(t, config, cache.Config)
	assert.NotNil(t, cache.httpClient)
}

func TestGetToken(t *testing.T) {
	// Generate RSA key pair for signing
	raw := generateRSAKey(t)

	// Setup JWK server
	jwkServer := setupJWKServer(t, raw)
	defer jwkServer.Close()

	testJWT := createTestJWT(t, raw)

	tests := []struct {
		name          string
		mockResponse  interface{}
		mockStatus    int
		expectedError bool
	}{
		{
			name: "successful token fetch",
			mockResponse: mockTokenResponse{
				AccessToken: testJWT,
				ExpiresIn:   3600,
				TokenType:   "Bearer",
			},
			mockStatus:    http.StatusOK,
			expectedError: false,
		},
		{
			name:          "server error",
			mockResponse:  map[string]interface{}{},
			mockStatus:    http.StatusInternalServerError,
			expectedError: true,
		},
		{
			name: "missing access token",
			mockResponse: map[string]interface{}{
				"expires_in": 3600,
			},
			mockStatus:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t, tt.mockResponse, tt.mockStatus)
			defer server.Close()

			config := OAuthConfig{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				TokenURL:     server.URL,
				JWKURL:       jwkServer.URL,
				RetryConfig: RetryConfig{
					MaxAttempts: 1,
					WaitTime:    time.Millisecond,
				},
			}

			cache := NewTokenCache(config)
			cache.httpClient = &http.Client{
				Timeout: 1 * time.Second,
			}

			token, err := cache.GetToken(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			if resp, ok := tt.mockResponse.(mockTokenResponse); ok {
				assert.Equal(t, resp.AccessToken, token)
			}
		})
	}
}

func TestTokenCaching(t *testing.T) {
	// Generate RSA key pair for signing
	raw := generateRSAKey(t)

	// Setup JWK server
	jwkServer := setupJWKServer(t, raw)
	defer jwkServer.Close()

	testJWT := createTestJWT(t, raw)

	server := setupMockServer(t, mockTokenResponse{
		AccessToken: testJWT,
		ExpiresIn:   3600,
		TokenType:   "Bearer",
	}, http.StatusOK)
	defer server.Close()

	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		JWKURL:       jwkServer.URL,
		RetryConfig: RetryConfig{
			MaxAttempts: 1,
			WaitTime:    time.Millisecond,
		},
	}

	cache := NewTokenCache(config)
	cache.httpClient = &http.Client{
		Timeout: 1 * time.Second,
	}

	// First call should fetch new token
	token1, err := cache.GetToken(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	// Second call should use cached token
	token2, err := cache.GetToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, token1, token2)
}

func TestTokenExpiration(t *testing.T) {
	// Generate RSA key pair for signing
	raw := generateRSAKey(t)

	// Setup JWK server
	jwkServer := setupJWKServer(t, raw)
	defer jwkServer.Close()

	// Create two different JWT tokens with different expiration times
	token1 := jwt.New()
	token1.Set(jwt.ExpirationKey, time.Now().Add(time.Second))
	token1.Set(jwt.IssuedAtKey, time.Now())
	token1.Set(jwt.SubjectKey, "test-1")

	token2 := jwt.New()
	token2.Set(jwt.ExpirationKey, time.Now().Add(time.Hour))
	token2.Set(jwt.IssuedAtKey, time.Now())
	token2.Set(jwt.SubjectKey, "test-2")

	signedToken1, err := jwt.Sign(token1, jwt.WithKey(jwa.RS256, raw))
	require.NoError(t, err)

	signedToken2, err := jwt.Sign(token2, jwt.WithKey(jwa.RS256, raw))
	require.NoError(t, err)

	responses := []mockTokenResponse{
		{
			AccessToken: string(signedToken1),
			ExpiresIn:   1,
			TokenType:   "Bearer",
		},
		{
			AccessToken: string(signedToken2),
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		},
	}
	currentResponse := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses[currentResponse])
		if currentResponse < len(responses)-1 {
			currentResponse++
		}
	}))
	defer server.Close()

	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		JWKURL:       jwkServer.URL,
		RetryConfig: RetryConfig{
			MaxAttempts: 1,
			WaitTime:    time.Millisecond,
		},
	}

	cache := NewTokenCache(config)
	cache.httpClient = &http.Client{
		Timeout: 1 * time.Second,
	}

	// First call should fetch new token
	token1Str, err := cache.GetToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, string(signedToken1), token1Str)

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Second call should fetch new token due to expiration
	token2Str, err := cache.GetToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, string(signedToken2), token2Str)
	assert.NotEqual(t, token1Str, token2Str)
}

func TestRetryLogic(t *testing.T) {
	// Generate RSA key pair for signing
	raw := generateRSAKey(t)

	// Setup JWK server
	jwkServer := setupJWKServer(t, raw)
	defer jwkServer.Close()

	testJWT := createTestJWT(t, raw)
	failures := 2
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= failures {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(mockTokenResponse{
			AccessToken: testJWT,
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	config := OAuthConfig{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL,
		JWKURL:       jwkServer.URL,
		RetryConfig: RetryConfig{
			MaxAttempts: 3,
			WaitTime:    time.Millisecond,
		},
	}

	cache := NewTokenCache(config)
	cache.httpClient = &http.Client{
		Timeout: 1 * time.Second,
	}

	token, err := cache.GetToken(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, failures+1, callCount)
}

func TestVerifyJWT(t *testing.T) {
	// Generate RSA key pair for signing
	raw := generateRSAKey(t)

	// Setup JWK server
	jwkServer := setupJWKServer(t, raw)
	defer jwkServer.Close()

	cache := &TokenCache{
		Config: OAuthConfig{
			JWKURL: jwkServer.URL,
		},
		httpClient: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	t.Run("invalid token format", func(t *testing.T) {
		_, _, err := cache.VerifyJWT(context.Background(), "invalid-token")
		assert.Error(t, err)
	})

	t.Run("valid token", func(t *testing.T) {
		token := createTestJWT(t, raw)
		parsedToken, claims, err := cache.VerifyJWT(context.Background(), token)
		require.NoError(t, err)
		assert.NotNil(t, parsedToken)
		assert.NotNil(t, claims)
		assert.Equal(t, "test-subject", claims["sub"])
	})
}
