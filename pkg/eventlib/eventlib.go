package eventlib

import (
	"context"
	"fmt"
	"sync"
)

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

type OnEventFn[T any] func(T) error
type OnErrorFn func(error)

type emitterFns[T any] struct {
	onEvent OnEventFn[T]
	onError OnErrorFn
}

type Emitter[T any] struct {
	ch     chan T
	fns    []emitterFns[T]
	ctx    context.Context
	cancel context.CancelFunc
	mu     *sync.RWMutex
}

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

func (e *Emitter[T]) Emit(data T) {
	select {
	case e.ch <- data:
	case <-e.ctx.Done():
		// Context cancelled, emitter is closed
	}
}

func (e *Emitter[T]) GetEmitter() func(T) {
	return func(data T) {
		e.Emit(data)
	}
}

func (e *Emitter[T]) listen() {
	defer func() {
		if r := recover(); r != nil {
			// Log or handle panic if needed
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
