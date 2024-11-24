// Package eventlib provides a generic event emitter implementation for Go applications.
//
// The package implements a type-safe event emitter pattern using Go generics, allowing
// for strongly-typed event handling with support for concurrent operations. It provides
// a simple pub/sub (publish/subscribe) mechanism that can be used to implement event-driven
// architectures in Go applications.
//
// Basic usage:
//
//	// Create a new emitter for string events with a context and buffer size
//	emitter := eventlib.NewEmitter[string](context.Background(), 10)
//
//	// Subscribe to events
//	emitter.Subscribe(
//	    func(data string) error {
//	        fmt.Println("Received:", data)
//	        return nil
//	    },
//	    func(err error) {
//	        fmt.Println("Error:", err)
//	    },
//	)
//
//	// Emit events
//	emitter.Emit("Hello, World!")
//
//	// Clean up when done
//	defer emitter.End()
//
// Features:
//   - Generic type support for type-safe event handling
//   - Buffered or unbuffered event channels
//   - Concurrent-safe operation with mutex protection
//   - Context-based cancellation
//   - Error handling support
//   - Panic recovery in event handlers
package eventlib
