package enclib

import "errors"

// Error represents an encoding/decoding error.
// It provides additional context about what went wrong during
// the encoding or decoding operation.
type Error struct {
	// Operation is the operation being performed (e.g., "encode", "decode")
	Operation string
	// Message is a human-readable description of the error
	Message string
	// Err is the underlying error that caused this error, if any
	Err error
}

var (
	// ErrInvalidQuotedString is returned when a quoted string is not properly formatted
	ErrInvalidQuotedString = errors.New("invalid quoted string format")
)
