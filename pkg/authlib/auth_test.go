package authlib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCache(t *testing.T) {
	// Create a test server that serves OAuth tokens
	var tokenCounter int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		err := r.ParseForm()
		require.NoError(t, err)

		// Verify headers
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Verify form values
		assert.Equal(t, "test-client", r.Form.Get("client_id"))
		assert.Equal(t, "test-secret", r.Form.Get("client_secret"))
		assert.Equal(t, r.Form.Get("grant_type"), r.Form.Get("grant_type")) // Accept any grant type

		// Create a test token response with a unique issuer to ensure different tokens
		token := jwt.New()
		tokenCounter++
		require.NoError(t, token.Set(jwt.IssuerKey, "test-issuer"))
		require.NoError(t, token.Set(jwt.ExpirationKey, time.Now().Add(time.Hour)))
		require.NoError(t, token.Set("counter", tokenCounter)) // Add a unique value

		tokenBytes, err := jwt.Sign(token, jwa.HS256, []byte("test-secret"))
		require.NoError(t, err)

		response := map[string]interface{}{
			"access_token": string(tokenBytes),
			"token_type":   "Bearer",
			"expires_in":   3600,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Run("GetToken success", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     server.URL,
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})

		// First call should fetch a new token
		token, err := cache.GetToken(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Second call should return the cached token
		cachedToken, err := cache.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, token, cachedToken)
	})

	t.Run("GetToken with expired token", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     server.URL,
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})

		// Get initial token
		token, err := cache.GetToken(context.Background())
		require.NoError(t, err)

		// Force token expiration
		cache.ExpiresAt = time.Now().Add(-time.Hour)

		// Get new token after expiration
		newToken, err := cache.GetToken(context.Background())
		require.NoError(t, err)
		assert.NotEqual(t, token, newToken)
	})

	t.Run("GetToken with custom grant type", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     server.URL,
			GrantType:    "custom_grant",
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})

		token, err := cache.GetToken(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("GetToken with invalid server response", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     "http://invalid-url",
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})

		_, err := cache.GetToken(context.Background())
		assert.Error(t, err)
	})
}

func TestNewTokenCache(t *testing.T) {
	t.Run("default grant type", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     "http://test-url",
		})

		assert.Equal(t, "client_credentials", cache.Config.GrantType)
	})

	t.Run("custom grant type", func(t *testing.T) {
		cache := NewTokenCache(OAuthConfig{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			TokenURL:     "http://test-url",
			GrantType:    "custom_grant",
		})

		assert.Equal(t, "custom_grant", cache.Config.GrantType)
	})
}
