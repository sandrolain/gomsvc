package authlib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWKCache(t *testing.T) {
	// Create a test server that serves JWK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are passed correctly
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		// Create a test key set
		key, err := jwk.New([]byte("test-secret"))
		require.NoError(t, err)
		require.NoError(t, key.Set(jwk.KeyIDKey, "test-kid"))
		require.NoError(t, key.Set(jwk.AlgorithmKey, jwa.HS256))

		keySet := jwk.NewSet()
		keySet.Add(key)

		// Serve the key set
		json.NewEncoder(w).Encode(keySet)
	}))
	defer server.Close()

	t.Run("FetchKeys with custom headers", func(t *testing.T) {
		cache := NewJWKCache(JWKConfig{
			JWKSURL: server.URL,
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
			ExpirationTime: time.Hour,
		})

		// First fetch should get from server
		keys, err := cache.FetchKeys(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, keys)
		assert.Equal(t, 1, keys.Len())

		// Second fetch should get from cache
		keys2, err := cache.FetchKeys(context.Background())
		require.NoError(t, err)
		assert.Equal(t, keys, keys2)
	})

	t.Run("FetchKeys with expired cache", func(t *testing.T) {
		cache := NewJWKCache(JWKConfig{
			JWKSURL:        server.URL,
			ExpirationTime: -time.Hour, // Expired
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})

		// First fetch
		keys1, err := cache.FetchKeys(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, keys1)

		// Second fetch should get new keys due to expiration
		keys2, err := cache.FetchKeys(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, keys2)
		// Keys should be different objects due to refetch
		assert.NotSame(t, keys1, keys2)
	})

	t.Run("FetchKeys with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{
			Timeout: time.Second * 5,
		}

		cache := NewJWKCache(JWKConfig{
			JWKSURL: server.URL,
			Headers: map[string]string{
				"X-Test-Header": "test-value",
			},
		})
		cache.SetHTTPClient(customClient)

		keys, err := cache.FetchKeys(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, keys)
	})
}

func TestTokenValidator(t *testing.T) {
	// Create a mock KeyProvider for testing
	mockKeyProvider := &mockKeyProvider{
		keys: createTestJWKSet(t),
	}

	t.Run("ValidateToken success", func(t *testing.T) {
		validator := NewTokenValidator(mockKeyProvider)

		// Create a test token
		token := createTestToken(t)

		// Create headers with key ID
		headers := jws.NewHeaders()
		require.NoError(t, headers.Set(jws.KeyIDKey, "test-kid"))

		tokenBytes, err := jwt.Sign(token, jwa.HS256, []byte("test-secret"),
			jwt.WithHeaders(headers))
		require.NoError(t, err)

		// Validate the token
		validToken, claims, err := validator.ValidateToken(context.Background(), string(tokenBytes))
		require.NoError(t, err)
		assert.NotNil(t, validToken)
		assert.Equal(t, "test-subject", claims["sub"])
		assert.Equal(t, "test-issuer", claims["iss"])
	})

	t.Run("ValidateToken with invalid token", func(t *testing.T) {
		validator := NewTokenValidator(mockKeyProvider)

		// Test with invalid token
		_, _, err := validator.ValidateToken(context.Background(), "invalid-token")
		assert.Error(t, err)
	})

	t.Run("ValidateToken with custom options", func(t *testing.T) {
		validator := NewTokenValidator(
			mockKeyProvider,
			jwt.WithIssuer("test-issuer"),
			jwt.WithSubject("test-subject"),
		)

		// Create a test token
		token := createTestToken(t)

		// Create headers with key ID
		headers := jws.NewHeaders()
		require.NoError(t, headers.Set(jws.KeyIDKey, "test-kid"))

		tokenBytes, err := jwt.Sign(token, jwa.HS256, []byte("test-secret"),
			jwt.WithHeaders(headers))
		require.NoError(t, err)

		// Validate the token
		validToken, claims, err := validator.ValidateToken(context.Background(), string(tokenBytes))
		require.NoError(t, err)
		assert.NotNil(t, validToken)
		assert.Equal(t, "test-subject", claims["sub"])
	})
}

// Helper types and functions

type mockKeyProvider struct {
	keys jwk.Set
}

func (m *mockKeyProvider) FetchKeys(ctx context.Context) (jwk.Set, error) {
	return m.keys, nil
}

func createTestJWKSet(t *testing.T) jwk.Set {
	key, err := jwk.New([]byte("test-secret"))
	require.NoError(t, err)
	require.NoError(t, key.Set(jwk.KeyIDKey, "test-kid"))
	require.NoError(t, key.Set(jwk.AlgorithmKey, jwa.HS256))

	keySet := jwk.NewSet()
	keySet.Add(key)
	return keySet
}

func createTestToken(t *testing.T) jwt.Token {
	token := jwt.New()
	require.NoError(t, token.Set(jwt.SubjectKey, "test-subject"))
	require.NoError(t, token.Set(jwt.IssuerKey, "test-issuer"))
	require.NoError(t, token.Set(jwt.IssuedAtKey, time.Now()))
	require.NoError(t, token.Set(jwt.ExpirationKey, time.Now().Add(time.Hour)))
	return token
}
