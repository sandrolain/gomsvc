package enclib

import (
	"encoding/hex"
)

// EncodeHex encodes a byte slice as a hexadecimal string.
// Each byte is converted to its 2-character hexadecimal representation.
// The output uses lowercase letters (a-f) and is twice the length of the input.
//
// Example:
//
//	EncodeHex([]byte{255, 0, 171}) returns "ff00ab"
func EncodeHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeHex decodes a hexadecimal string to its original bytes.
// The input string must have an even length and contain only valid
// hexadecimal characters (0-9, a-f, A-F).
//
// Common errors:
//   - Invalid hex character
//   - Odd length input
//   - Empty input
//
// Example:
//
//	DecodeHex("ff00ab") returns []byte{255, 0, 171}, nil
func DecodeHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// EncodeHexUpper encodes a byte slice as a hexadecimal string using uppercase letters.
// Each byte is converted to its 2-character hexadecimal representation.
// The output uses uppercase letters (A-F) and is twice the length of the input.
//
// Example:
//
//	EncodeHexUpper([]byte{255, 0, 171}) returns "FF00AB"
func EncodeHexUpper(data []byte) string {
	dst := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(dst, data)
	for i := 0; i < len(dst); i++ {
		if dst[i] >= 'a' {
			dst[i] -= 32
		}
	}
	return string(dst)
}
