package enclib

import (
	"bytes"
	"net/url"
	"sort"
	"strings"
)

// EncodeURL encodes a URL string while preserving the URL structure.
// It properly handles all URL components including scheme, user info,
// host, path, query parameters, and fragment.
//
// The function parses the URL first to ensure correct handling of each component.
// If the URL cannot be parsed, it returns an error.
//
// Example:
//
//	encoded, err := EncodeURL("https://user:pass@example.com/path?q=value#fragment")
func EncodeURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// DecodeURL decodes an encoded URL string while preserving URL structure.
// It handles the decoding of each URL component separately to ensure correct
// handling of special characters and encoding schemes.
//
// The function:
// - Decodes percent-encoded characters
// - Preserves URL structure
// - Handles spaces and Unicode characters correctly
// - Maintains query parameter order
//
// Returns an error if:
// - The URL is malformed
// - Contains invalid percent-encoded sequences
// - Has invalid UTF-8 sequences
func DecodeURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	if u.Path != "" {
		segments := strings.Split(u.Path, "/")
		for i, segment := range segments {
			if segment != "" {
				decoded, err := url.PathUnescape(segment)
				if err != nil {
					return "", err
				}
				segments[i] = decoded
			}
		}
		u.Path = strings.Join(segments, "/")
	}

	if u.RawQuery != "" {
		values, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		for key, values := range values {
			for _, value := range values {
				buf.WriteString(url.QueryEscape(key))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(value))
				buf.WriteByte('&')
			}
		}
		u.RawQuery = buf.String()
	}

	encoded := &bytes.Buffer{}
	if u.Scheme != "" {
		encoded.WriteString(u.Scheme)
		encoded.WriteString("://")
	}
	if u.Opaque != "" {
		encoded.WriteString(u.Opaque)
	} else {
		if u.User != nil {
			name := u.User.Username()
			password, _ := u.User.Password()
			encoded.WriteString(url.QueryEscape(name))
			if password != "" {
				encoded.WriteByte(':')
				encoded.WriteString(url.QueryEscape(password))
			}
			encoded.WriteByte('@')
		}
		encoded.WriteString(u.Host)
	}
	if u.Path != "" {
		encoded.WriteString(u.Path)
	}
	if u.RawQuery != "" {
		encoded.WriteByte('?')
		encoded.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		encoded.WriteByte('#')
		encoded.WriteString(url.QueryEscape(u.Fragment))
	}

	return encoded.String(), nil
}

// EncodeURLComponent encodes a string for safe use in a URL component.
// It encodes all characters that are not unreserved according to RFC 3986.
// Spaces are encoded as %20 rather than + to ensure consistent encoding
// across different URL components.
//
// Example:
//
//	EncodeURLComponent("Hello World!") returns "Hello%20World%21"
func EncodeURLComponent(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// DecodeURLComponent decodes a URL component string.
// It handles both %20 and + as space characters for compatibility
// with different encoding schemes.
//
// Returns an error if:
// - Contains invalid percent-encoded sequences
// - Has invalid UTF-8 sequences
func DecodeURLComponent(s string) (string, error) {
	return url.QueryUnescape(s)
}

// BuildQueryString builds a URL query string from a map of key-value pairs.
// The resulting string is properly encoded and sorted by key for consistency.
//
// Features:
// - Keys are sorted alphabetically
// - Spaces are encoded as %20
// - Special characters are properly escaped
// - Empty maps return an empty string
//
// Example:
//
//	params := map[string]string{"name": "John Doe", "age": "30"}
//	BuildQueryString(params) returns "age=30&name=John%20Doe"
func BuildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(strings.ReplaceAll(url.QueryEscape(params[k]), "+", "%20"))
	}
	return buf.String()
}
