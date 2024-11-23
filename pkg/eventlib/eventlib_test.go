package eventlib

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewEmitter(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantSize int
	}{
		{
			name:     "unbuffered channel",
			size:     0,
			wantSize: 0,
		},
		{
			name:     "buffered channel",
			size:     5,
			wantSize: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter := NewEmitter[int](context.Background(), tt.size)
			if cap(emitter.ch) != tt.wantSize {
				t.Errorf("NewEmitter() channel capacity = %v, want %v", cap(emitter.ch), tt.wantSize)
			}
		})
	}
}

func TestEmitter_Subscribe(t *testing.T) {
	emitter := NewEmitter[string](context.Background(), 0)

	onEvent := func(data string) error {
		return nil
	}

	onError := func(err error) {
		// Error handler
	}

	emitter.Subscribe(onEvent, onError)

	if len(emitter.fns) != 1 {
		t.Errorf("Subscribe() failed to add handler, got %v handlers, want 1", len(emitter.fns))
	}

	// Test nil handler
	emitter.Subscribe(nil, onError)
	if len(emitter.fns) != 1 {
		t.Errorf("Subscribe() should not add nil handler, got %v handlers, want 1", len(emitter.fns))
	}
}

func TestEmitter_Emit(t *testing.T) {
	ctx := context.Background()
	emitter := NewEmitter[string](ctx, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	var received string
	emitter.Subscribe(func(data string) error {
		received = data
		wg.Done()
		return nil
	}, nil)

	testData := "test message"
	emitter.Emit(testData)

	wg.Wait()

	if received != testData {
		t.Errorf("Emit() got = %v, want %v", received, testData)
	}
}

func TestEmitter_EmitWithError(t *testing.T) {
	ctx := context.Background()
	emitter := NewEmitter[string](ctx, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	expectedErr := errors.New("test error")
	var receivedErr error

	emitter.Subscribe(
		func(data string) error {
			return expectedErr
		},
		func(err error) {
			receivedErr = err
			wg.Done()
		},
	)

	emitter.Emit("test")
	wg.Wait()

	if receivedErr != expectedErr {
		t.Errorf("Error handler got = %v, want %v", receivedErr, expectedErr)
	}
}

func TestEmitter_End(t *testing.T) {
	ctx := context.Background()
	emitter := NewEmitter[string](ctx, 1)

	// Subscribe to verify no events are received after End()
	received := make(chan string, 1)
	emitter.Subscribe(func(data string) error {
		received <- data
		return nil
	}, nil)

	// First emit should work
	emitter.Emit("before-end")
	select {
	case msg := <-received:
		if msg != "before-end" {
			t.Errorf("Expected to receive 'before-end', got %s", msg)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for first emit")
	}

	emitter.End()

	// Try to emit after End()
	emitter.Emit("after-end")

	// Should not receive any more events
	select {
	case msg := <-received:
		t.Errorf("Should not receive events after End(), got %s", msg)
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	}

	if len(emitter.fns) != 0 {
		t.Errorf("End() didn't clear handlers, got %v handlers", len(emitter.fns))
	}
}

func TestEmitter_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	emitter := NewEmitter[int](ctx, 100)

	var wg sync.WaitGroup
	const numGoroutines = 10
	const numEmits = 100

	received := make(map[int]bool)
	var mu sync.Mutex

	emitter.Subscribe(func(data int) error {
		mu.Lock()
		received[data] = true
		mu.Unlock()
		return nil
	}, nil)

	// Start multiple goroutines emitting data
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numEmits; j++ {
				emitter.Emit(base*numEmits + j)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow time for processing

	mu.Lock()
	count := len(received)
	mu.Unlock()

	expected := numGoroutines * numEmits
	if count != expected {
		t.Errorf("ConcurrentAccess received %v events, want %v", count, expected)
	}
}

func TestEmitter_PanicRecovery(t *testing.T) {
	ctx := context.Background()
	emitter := NewEmitter[string](ctx, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedErr error
	emitter.Subscribe(
		func(data string) error {
			panic("test panic")
		},
		func(err error) {
			receivedErr = err
			wg.Done()
		},
	)

	emitter.Emit("test")
	wg.Wait()

	if receivedErr == nil {
		t.Error("PanicRecovery: error handler not called after panic")
	}

	if receivedErr.Error() != "panic in event handler: test panic" {
		t.Errorf("PanicRecovery: got unexpected error message: %v", receivedErr)
	}
}
