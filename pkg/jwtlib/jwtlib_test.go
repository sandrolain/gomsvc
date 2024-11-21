package jwtlib

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
}
