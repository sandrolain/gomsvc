// Package jwxlib provides JWT (JSON Web Token) creation and validation functionality
// with support for generic data types. It uses the lestrrat-go/jwx/v2 library for
// robust JWT handling and security.
//
// Key Features:
//   - Type-safe JWT creation and parsing
//   - Generic support for custom claim data
//   - Comprehensive validation
//   - Thread-safe operations
//
// Example Usage:
//
//	type UserData struct {
//	    Role string `json:"role"`
//	}
//
//	params := jwxlib.JWTParams[UserData]{
//	    Subject:   "user123",
//	    Issuer:    "myapp",
//	    Secret:    []byte("your-secret"),
//	    ExpiresAt: time.Now().Add(24 * time.Hour),
//	    Data:      UserData{Role: "admin"},
//	}
//
//	// Create JWT
//	token, err := jwxlib.CreateJWT(params)
//
//	// Parse and validate JWT
//	claims, err := jwxlib.ParseJWT(token, params)
package jwxlib

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// JWTParams holds the parameters needed for JWT creation and validation.
// It supports generic type T for custom claim data.
type JWTParams[T any] struct {
	// Subject is the principal subject of the JWT (usually a user ID)
	Subject string

	// Issuer identifies the principal that issued the JWT
	Issuer string

	// Secret is the key used for signing and validating the JWT
	Secret []byte

	// ExpiresAt specifies when the JWT will expire
	ExpiresAt time.Time

	// Data contains the custom claims of type T
	Data T
}

// Claims represents the JWT claims structure with generic support for custom data.
// It embeds standard JWT claims and adds a generic Data field.
type Claims[T any] struct {
	// Standard JWT claims
	Subject   string    `json:"sub,omitempty"`
	Issuer    string    `json:"iss,omitempty"`
	IssuedAt  time.Time `json:"iat,omitempty"`
	ExpiresAt time.Time `json:"exp,omitempty"`

	// Custom data of type T
	Data T `json:"dat,omitempty"`
}

// CreateJWT creates a new JWT with the provided parameters and custom data.
// It signs the token using HS256 (HMAC with SHA-256).
//
// Parameters:
//   - params: JWTParams containing all necessary JWT fields and custom data
//
// Returns:
//   - string: The signed JWT string
//   - error: Any error that occurred during token creation
//
// Security Considerations:
//   - The secret should be at least 32 bytes long for HS256
//   - Store secrets securely and never expose them
func CreateJWT[T any](params JWTParams[T]) (string, error) {
	// Create a new token
	builder := jwt.NewBuilder()
	token, err := builder.
		Subject(params.Subject).
		Issuer(params.Issuer).
		IssuedAt(time.Now()).
		Expiration(params.ExpiresAt).
		Claim("dat", params.Data).
		Build()

	if err != nil {
		return "", err
	}

	// Sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, params.Secret))
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// ParseJWT parses and validates a JWT string using the provided parameters.
// It performs full validation including signature, expiration, and issuer checks.
//
// Parameters:
//   - jwtString: The JWT string to parse and validate
//   - params: JWTParams containing validation parameters
//
// Returns:
//   - *Claims[T]: Parsed and validated claims including custom data
//   - error: Any error that occurred during parsing or validation
//
// Security Considerations:
//   - Always validate tokens before trusting their contents
//   - Check both signature and claims validity
func ParseJWT[T any](jwtString string, params JWTParams[T]) (*Claims[T], error) {
	if jwtString == "" {
		return nil, errors.New("the jwt string is empty")
	}

	// Parse and validate the token
	token, err := jwt.Parse([]byte(jwtString),
		jwt.WithKey(jwa.HS256, params.Secret),
		jwt.WithValidate(true),
		jwt.WithIssuer(params.Issuer),
	)
	if err != nil {
		return nil, err
	}

	// Extract claims
	claims := &Claims[T]{
		Subject:   token.Subject(),
		Issuer:    token.Issuer(),
		IssuedAt:  token.IssuedAt(),
		ExpiresAt: token.Expiration(),
	}

	// Extract custom data
	var data T
	if v, ok := token.Get("dat"); !ok {
		return nil, errors.New("cannot obtain JWT custom data")
	} else {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, errors.New("cannot marshal JWT custom data")
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, errors.New("cannot unmarshal JWT custom data")
		}
	}
	claims.Data = data

	return claims, nil
}

// ExtractInfoFromJWT extracts claims from a JWT without validating its signature.
// This is useful when you need to inspect the token contents without verification.
//
// WARNING: This function does not verify the token's signature. Never trust data
// from an unverified token in a security-critical context.
//
// Parameters:
//   - jwtString: The JWT string to parse
//
// Returns:
//   - *Claims[T]: Parsed claims including custom data
//   - error: Any error that occurred during parsing
func ExtractInfoFromJWT[T any](jwtString string) (*Claims[T], error) {
	if jwtString == "" {
		return nil, errors.New("the jwt string is empty")
	}

	// Parse without verification
	token, err := jwt.Parse([]byte(jwtString),
		jwt.WithValidate(false),
		jwt.WithVerify(false),
	)
	if err != nil {
		return nil, err
	}

	// Extract claims
	claims := &Claims[T]{
		Subject:   token.Subject(),
		Issuer:    token.Issuer(),
		IssuedAt:  token.IssuedAt(),
		ExpiresAt: token.Expiration(),
	}

	// Extract custom data
	var data T
	if v, ok := token.Get("dat"); !ok {
		return nil, errors.New("cannot obtain JWT custom data")
	} else {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, errors.New("cannot marshal JWT custom data")
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, errors.New("cannot unmarshal JWT custom data")
		}
	}
	claims.Data = data

	return claims, nil
}
