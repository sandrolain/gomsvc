// Package cryptolib provides cryptographic utilities including secure random number generation.
package cryptolib

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// RandomBytes generates a slice of cryptographically secure random bytes.
// This function uses crypto/rand to ensure high-quality random numbers suitable
// for cryptographic use.
//
// Parameters:
//   - length: The number of random bytes to generate
//
// Returns:
//   - []byte: A slice containing the random bytes
//   - error: Any error that occurred during random number generation
//
// Example:
//
//	bytes, err := RandomBytes(16)
//	if err != nil {
//	    log.Fatal(err)
//	}
func RandomBytes(length int) (res []byte, err error) {
	res = make([]byte, length)
	_, err = rand.Read(res)
	return
}

// RandomBytesBase64 generates a base64-encoded string of random bytes.
// This function is useful when you need a random string that is safe to use
// in URLs, cookies, or other text-based contexts.
//
// Parameters:
//   - length: The number of random bytes to generate before base64 encoding
//
// Returns:
//   - string: The base64-encoded random bytes
//   - error: Any error that occurred during random number generation or encoding
//
// Note: The resulting string will be longer than the input length due to base64 encoding.
// The output length will be ceil(length * 4/3) due to base64 encoding characteristics.
//
// Example:
//
//	str, err := RandomBytesBase64(24)
//	if err != nil {
//	    log.Fatal(err)
//	}
func RandomBytesBase64(length int) (res string, err error) {
	b, err := RandomBytes(length)
	if err != nil {
		return
	}
	res = base64.StdEncoding.EncodeToString(b)
	return
}

// GenerateAES256Key generates a cryptographically secure 256-bit (32-byte) key
// suitable for use with AES-256 encryption. This function uses crypto/rand
// to ensure the generated key has sufficient entropy for cryptographic use.
//
// Returns:
//   - []byte: A 32-byte slice containing the random AES-256 key
//   - error: Any error that occurred during key generation
//
// Example:
//
//	key, err := GenerateAES256Key()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use key with AES-256 encryption
func GenerateAES256Key() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
