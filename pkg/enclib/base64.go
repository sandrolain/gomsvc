// Package enclib provides encoding and decoding utilities for various formats
// including base64 (standard and URL-safe) and hexadecimal.
package enclib

import (
	"encoding/base64"
)

// EncodeBase64 encodes a byte slice using standard base64 encoding.
// It uses the standard base64 alphabet and includes padding characters ('=').
// The encoding preserves the exact bytes of the input and is commonly used
// for encoding binary data in text-based protocols.
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes a base64 encoded string back to its original bytes.
// It expects standard base64 encoding with padding characters.
// Returns an error if the input is not valid base64 encoded data.
//
// Common errors:
//   - Invalid base64 character
//   - Incorrect padding
//   - Invalid input length
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// EncodeBase64URL encodes a byte slice using URL-safe base64 encoding.
// The URL-safe encoding uses '-' and '_' instead of '+' and '/' and
// omits padding characters. This makes it safe to use in URLs and
// filenames without additional escaping.
func EncodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeBase64URL decodes a URL-safe base64 encoded string.
// It expects the URL-safe variant of base64 encoding without padding characters.
// Returns an error if the input is not valid base64 encoded data.
//
// Common errors:
//   - Invalid base64 character
//   - Invalid input length
func DecodeBase64URL(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}
