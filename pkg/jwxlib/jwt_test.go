package jwxlib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestCreateAndParseJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	now := time.Now()
	expiresAt := now.Add(time.Hour)
	testData := TestData{
		Field1: "test",
		Field2: 123,
	}

	params := JWTParams[TestData]{
		Subject:   "test-subject",
		Issuer:    "test-issuer",
		Secret:    secret,
		ExpiresAt: expiresAt,
		Data:      testData,
	}

	// Test JWT creation
	token, err := CreateJWT(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test JWT parsing
	claims, err := ParseJWT(token, params)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, params.Subject, claims.Subject)
	assert.Equal(t, params.Issuer, claims.Issuer)
	assert.Equal(t, testData, claims.Data)
	assert.Equal(t, expiresAt.Unix(), claims.ExpiresAt.Unix())

	// Test with invalid token
	_, err = ParseJWT("invalid-token", params)
	assert.Error(t, err)

	// Test with empty token
	_, err = ParseJWT("", params)
	assert.Error(t, err)
	assert.Equal(t, "the jwt string is empty", err.Error())

	// Test with wrong issuer
	wrongIssuerParams := params
	wrongIssuerParams.Issuer = "wrong-issuer"
	_, err = ParseJWT(token, wrongIssuerParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "\"iss\" not satisfied")

	// Test with expired token
	expiredParams := params
	expiredParams.ExpiresAt = now.Add(-time.Hour)
	expiredToken, err := CreateJWT(expiredParams)
	assert.NoError(t, err)
	_, err = ParseJWT(expiredToken, expiredParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "\"exp\" not satisfied")

	// Test with wrong secret
	wrongSecretParams := params
	wrongSecretParams.Secret = []byte("wrong-secret")
	_, err = ParseJWT(token, wrongSecretParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not verify message")
}

func TestExtractInfoFromJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	now := time.Now()
	expiresAt := now.Add(time.Hour)
	testData := TestData{
		Field1: "test",
		Field2: 123,
	}

	params := JWTParams[TestData]{
		Subject:   "test-subject",
		Issuer:    "test-issuer",
		Secret:    secret,
		ExpiresAt: expiresAt,
		Data:      testData,
	}

	// Create a token first
	token, err := CreateJWT(params)
	assert.NoError(t, err)

	// Test info extraction
	claims, err := ExtractInfoFromJWT[TestData](token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, params.Subject, claims.Subject)
	assert.Equal(t, params.Issuer, claims.Issuer)
	assert.Equal(t, testData, claims.Data)

	// Test with invalid token
	_, err = ExtractInfoFromJWT[TestData]("invalid-token")
	assert.Error(t, err)

	// Test with empty token
	_, err = ExtractInfoFromJWT[TestData]("")
	assert.Error(t, err)
	assert.Equal(t, "the jwt string is empty", err.Error())

	// Test with expired token - should still work since we're not validating
	expiredParams := params
	expiredParams.ExpiresAt = now.Add(-time.Hour)
	expiredToken, err := CreateJWT(expiredParams)
	assert.NoError(t, err)
	claims, err = ExtractInfoFromJWT[TestData](expiredToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, expiredParams.Subject, claims.Subject)
}

func TestNilData(t *testing.T) {
	secret := []byte("test-secret-key")
	now := time.Now()
	expiresAt := now.Add(time.Hour)

	params := JWTParams[*TestData]{
		Subject:   "test-subject",
		Issuer:    "test-issuer",
		Secret:    secret,
		ExpiresAt: expiresAt,
		Data:      nil,
	}

	// Test JWT creation with nil data
	token, err := CreateJWT(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test JWT parsing with nil data
	claims, err := ParseJWT(token, params)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, params.Subject, claims.Subject)
	assert.Equal(t, params.Issuer, claims.Issuer)
	assert.Nil(t, claims.Data)
}
