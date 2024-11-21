package asynclib

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetTimeout(t *testing.T) {
	var callbackExecuted atomic.Bool
	callback := func() {
		callbackExecuted.Store(true)
	}

	timeout := SetTimeout(callback, 100) // 100 milliseconds

	time.Sleep(150 * time.Millisecond)

	timeout.Cancel()

	require.True(t, callbackExecuted.Load(), "The callback should have been executed")
}

func TestSetTimeoutCancel(t *testing.T) {
	var callbackExecuted atomic.Bool
	callback := func() {
		callbackExecuted.Store(true)
	}

	timeout := SetTimeout(callback, 200) // 200 milliseconds

	timeout.Cancel() // Cancel the timeout

	time.Sleep(250 * time.Millisecond)

	require.False(t, callbackExecuted.Load(), "The callback should not have been executed after cancel")
}

func TestSetInterval(t *testing.T) {
	var callbackExecuted atomic.Int32
	callback := func() {
		callbackExecuted.Add(1)
	}

	interval := SetInterval(callback, 100) // 100 milliseconds

	time.Sleep(350 * time.Millisecond)

	interval.Stop()

	require.Equal(t, int32(3), callbackExecuted.Load(), "The callback should have been executed 3 times")
}

func TestSetIntervalStop(t *testing.T) {
	var callbackExecuted atomic.Int32
	callback := func() {
		callbackExecuted.Add(1)
	}

	interval := SetInterval(callback, 100) // 100 milliseconds

	time.Sleep(150 * time.Millisecond)

	require.Equal(t, int32(1), callbackExecuted.Load(), "The callback should have been executed only once before stop")

	interval.Stop() // Stop the interval

	time.Sleep(200 * time.Millisecond)

	require.Equal(t, int32(1), callbackExecuted.Load(), "The callback should have been executed only once after stop")
}

func TestStartWorkers(t *testing.T) {
	workers := StartWorkers[int, int](2, func(val int) (int, error) {
		if val < 0 {
			return 0, fmt.Errorf("negative value: %d", val)
		}
		return val * 2, nil
	})
	defer workers.Stop()

	workers.Input <- 1
	workers.Input <- 2
	workers.Input <- -3

	var results []int
	for i := 0; i < 3; i++ {
		select {
		case res := <-workers.Results:
			if res.Input < 0 {
				if res.Err == nil {
					t.Fatalf("Worker should have returned an error for negative value: %v", res.Input)
				}
				continue
			}
			if res.Err != nil {
				t.Fatalf("Worker returned an error: %v", res.Err)
			}
			results = append(results, res.Result)
		case <-time.After(time.Second):
			t.Fatal("Timed out waiting for worker results")
		}
	}

	require.NotContains(t, results, -6, "Results should not contain -6")
	require.Contains(t, results, 2, "Results should contain 2")
	require.Contains(t, results, 4, "Results should contain 4")
}
