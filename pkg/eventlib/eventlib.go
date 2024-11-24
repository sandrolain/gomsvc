package eventlib

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// NewEmitter creates a new event emitter with the specified context and channel buffer size.
// If size is 0 or negative, an unbuffered channel is created.
// The returned emitter must be cleaned up by calling End() when no longer needed.
func NewEmitter[T any](ctx context.Context, size int) *Emitter[T] {
	var ch chan T
	if size > 0 {
		ch = make(chan T, size)
	} else {
		ch = make(chan T)
	}
	ctx, cancel := context.WithCancel(ctx)

	emitter := Emitter[T]{
		ch:     ch,
		fns:    make([]emitterFns[T], 0),
		ctx:    ctx,
		cancel: cancel,
		mu:     &sync.RWMutex{},
	}

	go emitter.listen()
	return &emitter
}

// OnEventFn is a function type that handles events of type T.
// It returns an error if the event handling fails.
type OnEventFn[T any] func(T) error

// OnErrorFn is a function type that handles errors that occur during event processing.
type OnErrorFn func(error)

type emitterFns[T any] struct {
	onEvent OnEventFn[T]
	onError OnErrorFn
}

// Emitter is a generic event emitter that supports type-safe event handling.
// It provides concurrent-safe operations for emitting and handling events of type T.
type Emitter[T any] struct {
	ch     chan T
	fns    []emitterFns[T]
	ctx    context.Context
	cancel context.CancelFunc
	mu     *sync.RWMutex
}

// Subscribe adds a new event handler to the emitter.
// The onEvent function is called for each emitted event.
// The onError function is called when an error occurs during event handling.
// If onEvent is nil, the subscription is ignored.
// Multiple handlers can be subscribed to the same emitter.
func (e *Emitter[T]) Subscribe(onEvent OnEventFn[T], onError OnErrorFn) {
	if onEvent == nil {
		return // Don't add nil handlers
	}
	e.mu.Lock()
	e.fns = append(e.fns, emitterFns[T]{
		onEvent: onEvent,
		onError: onError,
	})
	e.mu.Unlock()
}

// Emit sends a new event to all subscribed handlers.
// If the emitter's context is cancelled or the channel is full,
// the event will be dropped.
func (e *Emitter[T]) Emit(data T) {
	select {
	case e.ch <- data:
	case <-e.ctx.Done():
		// Context cancelled, emitter is closed
	}
}

// GetEmitter returns a function that can be used to emit events.
// This is useful when you want to pass the emit capability without
// exposing the entire Emitter interface.
func (e *Emitter[T]) GetEmitter() func(T) {
	return func(data T) {
		e.Emit(data)
	}
}

func (e *Emitter[T]) listen() {
	defer func() {
		if r := recover(); r != nil {
			slog.Default().Error("panic recovered", "panic", r)
		}
	}()

	for {
		select {
		case <-e.ctx.Done():
			return
		case data, ok := <-e.ch:
			if !ok {
				return
			}
			e.handleEvent(data)
		}
	}
}

func (e *Emitter[T]) handleEvent(data T) {
	e.mu.RLock()
	handlers := make([]emitterFns[T], len(e.fns))
	copy(handlers, e.fns)
	e.mu.RUnlock()

	for _, v := range handlers {
		func(handler emitterFns[T]) {
			defer func() {
				if r := recover(); r != nil {
					if handler.onError != nil {
						handler.onError(fmt.Errorf("panic in event handler: %v", r))
					}
				}
			}()

			if handler.onEvent != nil {
				if err := handler.onEvent(data); err != nil && handler.onError != nil {
					handler.onError(err)
				}
			}
		}(v)
	}
}

func (e *Emitter[T]) End() {
	e.cancel() // Cancel context first

	// Clear handlers under lock to prevent new emissions
	e.mu.Lock()
	e.fns = nil
	e.mu.Unlock()
}
