package eventlib

import (
	"context"
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
}

func (e *Emitter[T]) Subscribe(onEvent OnEventFn[T], onError OnErrorFn) {
	e.fns = append(e.fns, emitterFns[T]{
		onEvent: onEvent,
		onError: onError,
	})
}

func (e *Emitter[T]) Emit(data T) {
	e.ch <- data
}

func (e *Emitter[T]) GetEmitter() func(T) {
	return func(data T) {
		e.Emit(data)
	}
}

func (e *Emitter[T]) listen() {
	for {
		if e.ctx.Err() != nil {
			break
		}
		data := <-e.ch
		go func(data T) {
			for _, v := range e.fns {
				var err error
				if v.onEvent != nil {
					err = v.onEvent(data)
				}
				if err != nil && v.onError != nil {
					v.onError(err)
				}
			}
		}(data)
	}
}

func (e *Emitter[T]) End() {
	e.cancel()
}
