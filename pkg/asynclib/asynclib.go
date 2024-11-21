package asynclib

import (
	"sync"
	"time"
)

type Timeout struct {
	Done      <-chan bool
	Cancelled <-chan bool
	s         chan bool
	Timer     *time.Timer
}

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

func (t Timeout) Cancel() {
	t.Timer.Stop()
	t.s <- true
}

type Interval struct {
	Ticker  *time.Ticker
	Stopped <-chan bool
	s       chan bool
}

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

func (i Interval) Stop() {
	i.Ticker.Stop()
	i.s <- true
}

type WorketResult[T any, R any] struct {
	WorkerNum int
	Input     T
	Result    R
	Err       error
}

type Workers[T any, R any] struct {
	Input    chan T
	Results  chan WorketResult[T, R]
	done     chan struct{}
	stopOnce sync.Once
}

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

func (j *Workers[T, R]) Stop() {
	j.stopOnce.Do(func() {
		close(j.done)
		close(j.Input)
		close(j.Results)
	})
}
