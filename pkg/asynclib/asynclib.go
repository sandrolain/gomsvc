// Package asynclib provides utilities for asynchronous operations in Go, including
// timeouts, intervals, and worker pools. It implements familiar JavaScript-like
// patterns such as setTimeout and setInterval, along with a generic worker pool
// for parallel processing.
package asynclib

import (
	"sync"
	"time"
)

// Timeout represents a scheduled function execution that can be cancelled.
// It provides channels to monitor completion and cancellation states.
type Timeout struct {
	// Done receives true when the scheduled function completes
	Done <-chan bool
	// Cancelled receives true when the timeout is cancelled
	Cancelled <-chan bool
	// s is the internal channel for signaling cancellation
	s chan bool
	// Timer is the underlying timer that triggers the function
	Timer *time.Timer
}

// SetTimeout schedules a function to be executed after a specified delay.
// It mimics JavaScript's setTimeout behavior.
//
// Parameters:
//   - fn: The function to be executed after the delay
//   - millis: The delay in milliseconds before executing the function
//
// Returns:
//   - Timeout: A Timeout instance that can be used to cancel the scheduled execution
func SetTimeout(fn func(), millis int64) Timeout {
	done := make(chan bool, 1)
	cancel := make(chan bool, 1)
	timer := time.NewTimer(time.Duration(time.Millisecond * time.Duration(millis)))
	go func() {
		<-timer.C
		fn()
		done <- true
	}()
	return Timeout{Done: done, Cancelled: cancel, s: cancel, Timer: timer}
}

// Cancel stops the timer and cancels the scheduled function execution.
// It sends a signal through the Cancelled channel.
func (t Timeout) Cancel() {
	t.Timer.Stop()
	t.s <- true
}

// Interval represents a recurring function execution that can be stopped.
// It provides a way to execute a function at regular intervals.
type Interval struct {
	// Ticker is the underlying ticker that triggers the function at regular intervals
	Ticker *time.Ticker
	// Stopped receives true when the interval is stopped
	Stopped <-chan bool
	// s is the internal channel for signaling stop
	s chan bool
}

// SetInterval schedules a function to be executed repeatedly at specified intervals.
// It mimics JavaScript's setInterval behavior.
//
// Parameters:
//   - fn: The function to be executed at each interval
//   - millis: The interval duration in milliseconds between executions
//
// Returns:
//   - Interval: An Interval instance that can be used to stop the recurring execution
func SetInterval(fn func(), millis int64) Interval {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(time.Duration(millis * int64(time.Millisecond)))
	go func() {
		for {
			select {
			case <-ticker.C:
				fn()
			case <-stopped:
				ticker.Stop()
				return
			}
		}
	}()
	return Interval{Ticker: ticker, Stopped: stopped, s: stopped}
}

// Stop halts the recurring execution of the interval.
// It stops the ticker and sends a signal through the Stopped channel.
func (i Interval) Stop() {
	i.Ticker.Stop()
	i.s <- true
}

// WorketResult represents the result of a single worker's execution.
// It contains information about the worker, input, result, and any error that occurred.
type WorketResult[T any, R any] struct {
	// WorkerNum is the identifier of the worker that processed this result
	WorkerNum int
	// Input is the original input value that was processed
	Input T
	// Result is the processed output value
	Result R
	// Err contains any error that occurred during processing
	Err error
}

// Workers manages a pool of worker goroutines that process inputs in parallel.
// It provides channels for sending inputs and receiving results.
type Workers[T any, R any] struct {
	// Input is the channel for sending values to be processed by workers
	Input chan T
	// Results is the channel for receiving processed results from workers
	Results chan WorketResult[T, R]
	// done is the internal channel for signaling worker termination
	done chan struct{}
	// stopOnce ensures Stop() is called only once
	stopOnce sync.Once
}

// StartWorkers creates and starts a pool of worker goroutines.
// Each worker processes inputs using the provided function.
//
// Parameters:
//   - workersNum: The number of worker goroutines to start
//   - fn: The function that each worker will use to process inputs
//
// Returns:
//   - Workers[T, R]: A Workers instance that can be used to send inputs and receive results
func StartWorkers[T any, R any](workersNum int, fn func(T) (R, error)) (res Workers[T, R]) {
	inputs := make(chan T, workersNum)
	results := make(chan WorketResult[T, R], workersNum)
	done := make(chan struct{})

	workers := Workers[T, R]{
		Input:   inputs,
		Results: results,
		done:    done,
	}

	for n := 1; n <= workersNum; n++ {
		go func(n int) {
			for {
				select {
				case val, ok := <-inputs:
					if !ok {
						return
					}
					res, err := fn(val)
					results <- WorketResult[T, R]{
						WorkerNum: n,
						Input:     val,
						Result:    res,
						Err:       err,
					}
				case <-done:
					return
				}
			}
		}(n)
	}
	return workers
}

// Stop gracefully shuts down the worker pool.
// It closes all channels and terminates all worker goroutines.
// This method is safe to call multiple times.
func (j *Workers[T, R]) Stop() {
	j.stopOnce.Do(func() {
		close(j.done)
		close(j.Input)
		close(j.Results)
	})
}
