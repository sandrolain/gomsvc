/*
Package enclib provides a comprehensive set of encoding utilities for handling various data formats
and transformations. It includes functions for base64, URL, JSON, hex, and string encoding/decoding
operations with a focus on security and correctness.

Key features:

  - Base64 encoding with standard and URL-safe variants
  - URL encoding with proper handling of components and special characters
  - JSON encoding with type safety and error handling
  - Hex encoding for binary data
  - String encoding with support for quoted strings and control characters

Example usage:

	// Base64 encoding
	encoded := enclib.EncodeBase64([]byte("Hello, World!"))
	decoded, err := enclib.DecodeBase64(encoded)

	// URL encoding
	urlEncoded, err := enclib.EncodeURL("https://example.com/path with spaces")
	urlDecoded, err := enclib.DecodeURL(urlEncoded)

	// JSON encoding
	jsonStr, err := enclib.EncodeJSON(map[string]interface{}{"key": "value"})
	var result map[string]interface{}
	err = enclib.DecodeJSON(jsonStr, &result)

	// Hex encoding
	hexStr := enclib.EncodeHex([]byte{0xFF, 0x00, 0xAB})
	hexBytes, err := enclib.DecodeHex(hexStr)

	// String encoding
	quotedStr := enclib.EncodeQuotedString("String with \"quotes\"")
	unquotedStr, err := enclib.DecodeQuotedString(quotedStr)

The package is designed to be:

  - Secure: Uses standard library functions and follows security best practices
  - Correct: Extensive test coverage ensures correct handling of edge cases
  - Efficient: Minimizes allocations and uses efficient algorithms
  - Easy to use: Consistent API design across all encoding functions

For more information about specific encoding functions, see the documentation
for individual functions and types.
*/
package enclib
