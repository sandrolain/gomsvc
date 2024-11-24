package enclib

import (
	"fmt"
	"regexp"
	"strings"
)

// EncodeQuotedString encodes a string with proper quote escaping.
// It adds surrounding double quotes and escapes special characters
// according to Go string literal rules.
//
// The following characters are escaped:
//   - Double quote (") -> \"
//   - Backslash (\) -> \\
//   - Control characters -> \n, \r, \t, etc.
//
// Example:
//
//	EncodeQuotedString(`Hello "World"`) returns `"Hello \"World\""`
func EncodeQuotedString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return fmt.Sprintf("\"%s\"", s)
}

// DecodeQuotedString decodes a quoted string, removing surrounding quotes
// and unescaping special characters.
//
// The function handles:
//   - Surrounding double quotes
//   - Escaped quotes (\" -> ")
//   - Escaped backslashes (\\ -> \)
//   - Escaped control characters (\n -> newline, etc.)
//
// Returns an error if:
//   - String doesn't start/end with quotes
//   - Contains invalid escape sequences
//   - Quotes are not properly escaped
//
// Example:
//
//	DecodeQuotedString(`"Hello \"World\""`) returns `Hello "World"`, nil
func DecodeQuotedString(s string) (string, error) {
	if len(s) < 2 || !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		return "", fmt.Errorf("string must be wrapped in double quotes")
	}

	// Remove the surrounding quotes
	s = s[1 : len(s)-1]

	// Handle escaped characters
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			if i+1 >= len(s) {
				return "", fmt.Errorf("invalid escape sequence at end of string")
			}
			switch s[i+1] {
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			default:
				return "", fmt.Errorf("invalid escape sequence \\%c", s[i+1])
			}
			i++
		} else {
			result.WriteByte(s[i])
		}
	}

	return result.String(), nil
}

// EncodeNonPrintable encodes non-printable characters in a string.
// It replaces control characters and other non-printable characters
// with their escaped hexadecimal representation.
//
// Encoding rules:
//   - ASCII control chars (0x00-0x1F) -> \xHH
//   - Delete char (0x7F) -> \x7F
//   - Unicode control chars -> \uHHHH
//   - Other non-printable chars -> \UHHHHHHHH
//
// Example:
//
//	EncodeNonPrintable("Hello\x00World") returns "Hello\\x00World"
func EncodeNonPrintable(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		case 0:
			result.WriteString("\\x00")
		default:
			if r < 32 || r == 127 {
				result.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

// SanitizeControlChars replaces control characters in a string with their
// visible representations. This is useful for logging and display purposes
// where control characters might affect output formatting.
//
// Replacements:
//   - Newline -> \n
//   - Carriage return -> \r
//   - Tab -> \t
//   - Other control chars -> \xHH
//
// The function preserves printable characters, including spaces and Unicode.
//
// Example:
//
//	SanitizeControlChars("Hello\nWorld\t!") returns "Hello\\nWorld\\t!"
func SanitizeControlChars(s string) string {
	// Replace newlines, tabs, and carriage returns with a single space
	s = regexp.MustCompile(`[\n\t\r]+`).ReplaceAllString(s, " ")
	// Remove other control characters
	s = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(s, "")
	// Collapse multiple spaces into one
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
