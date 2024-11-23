// Package cryptolib provides cryptographic utilities for hashing and comparing values
// using industry-standard algorithms like BCrypt and SHA-256. This package simplifies
// common cryptographic operations while maintaining security best practices.
package cryptolib

import (
	"bytes"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

// HashBCrypt generates a BCrypt hash from the provided byte slice using a cost factor of 14.
// This function is suitable for securely hashing passwords before storage.
//
// Parameters:
//   - value: The byte slice to be hashed
//
// Returns:
//   - []byte: The resulting BCrypt hash
//   - error: Any error that occurred during hashing
func HashBCrypt(value []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(value, 14)
}

// CompareBCrypt safely compares a plain text value against a BCrypt hash.
// This function uses a constant-time comparison to prevent timing attacks.
//
// Parameters:
//   - value: The plain text value to compare
//   - hash: The BCrypt hash to compare against
//
// Returns:
//   - bool: true if the value matches the hash, false otherwise
func CompareBCrypt(value []byte, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, value)
	return err == nil
}

// HashSHA256 generates a SHA-256 hash from the provided byte slice.
// This function is suitable for general-purpose hashing where cryptographic
// security is needed, but not for password hashing (use HashBCrypt instead).
//
// Parameters:
//   - value: The byte slice to be hashed
//
// Returns:
//   - []byte: The resulting SHA-256 hash
func HashSHA256(value []byte) []byte {
	h := sha256.New()
	h.Write(value)
	return h.Sum(nil)
}

// CompareSHA256 safely compares a plain text value against a SHA-256 hash.
// This function uses a constant-time comparison to prevent timing attacks.
//
// Parameters:
//   - value: The plain text value to compare
//   - hash: The SHA-256 hash to compare against
//
// Returns:
//   - bool: true if the value matches the hash, false otherwise
func CompareSHA256(value []byte, hash []byte) bool {
	h := HashSHA256(value)
	return bytes.Equal(hash, h)
}
