package enclib

import (
	"github.com/bytedance/sonic"
)

// EncodeJSON encodes a value as a JSON string.
// The output is not indented or pretty-printed.
func EncodeJSON[T any](value T) (string, error) {
	data, err := sonic.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeJSON decodes a JSON string into a target value.
// T must be a pointer type to the desired target type.
func DecodeJSON[T any](data string) (T, error) {
	var target T
	err := sonic.Unmarshal([]byte(data), &target)
	if err != nil {
		return target, err
	}
	return target, nil
}

// EncodeJSONPretty encodes a value as an indented JSON string.
// The output is formatted with newlines and indentation for readability.
func EncodeJSONPretty[T any](value T) (string, error) {
	data, err := sonic.ConfigDefault.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CompactJSON removes whitespace from a JSON string.
// Returns an error if the input is not valid JSON.
func CompactJSON(data string) (string, error) {
	var value interface{}
	if err := sonic.Unmarshal([]byte(data), &value); err != nil {
		return "", err
	}
	compacted, err := sonic.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(compacted), nil
}

// ValidateJSON checks if a string is valid JSON.
// Returns nil if the input is valid JSON, error otherwise.
func ValidateJSON(data string) error {
	var value interface{}
	return sonic.Unmarshal([]byte(data), &value)
}
