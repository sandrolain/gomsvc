// Package devlib provides development utilities for debugging and logging.
// It includes functions for pretty printing, JSON formatting, and source code location information.
// This package is intended for development use only and should not be used in production code.
package devlib

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/k0kubun/pp/v3"
)

// P pretty prints any value using the pp package.
// It's useful for debugging complex data structures.
func P(v any) {
	_, err := pp.Print(v)
	if err != nil {
		slog.Default().Error(err.Error())
	}
}

// PrintJSON formats and prints any value as indented JSON.
// It's useful for inspecting data structures in a JSON format.
func PrintJSON(val interface{}) {
	res, err := json.MarshalIndent(val, "> ", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(res))
}

// FileLine returns the current file name and line number.
// It uses runtime.Caller to get the caller's location in the source code.
// Returns an empty string if the caller information cannot be obtained.
func FileLine() string {
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		return fmt.Sprintf("%s:%d", fileName, fileLine)
	}
	return ""
}

// FileLineError creates an error with the current file name and line number prepended.
// It supports format strings and arguments like fmt.Errorf.
// The error message will be in the format: [file:line] message
func FileLineError(format string, args ...interface{}) error {
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		return fmt.Errorf("[%s:%d] %s", fileName, fileLine, fmt.Sprintf(format, args...))
	}
	return fmt.Errorf(format, args...)
}

// FileLinePrintf prints a message with the current file name and line number prepended.
// It supports format strings and arguments like fmt.Printf.
// The output will be in the format: [file:line] message
func FileLinePrintf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, fileName, fileLine, ok := runtime.Caller(1)
	if ok {
		fmt.Printf("[%s:%d] %s\n", fileName, fileLine, message)
		return
	}
	fmt.Println(message)
}
