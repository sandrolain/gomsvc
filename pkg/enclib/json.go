package enclib

import (
	"bytes"
	"encoding/json"
)

// EncodeJSON encodes a value as a JSON string.
// The output is not indented or pretty-printed.
func EncodeJSON[T any](value T) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeJSON decodes a JSON string into a target value.
// T must be a pointer type to the desired target type.
func DecodeJSON[T any](data string) (T, error) {
	var target T
	err := json.Unmarshal([]byte(data), &target)
	if err != nil {
		return target, err
	}
	return target, nil
}

// EncodeJSONPretty encodes a value as an indented JSON string.
// The output is formatted with newlines and indentation for readability.
func EncodeJSONPretty[T any](value T) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CompactJSON removes whitespace from a JSON string.
// Returns an error if the input is not valid JSON.
func CompactJSON(data string) (string, error) {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(data))
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ValidateJSON checks if a string is valid JSON.
// Returns nil if the input is valid JSON, error otherwise.
func ValidateJSON(data string) error {
	return json.Unmarshal([]byte(data), &json.RawMessage{})
}
