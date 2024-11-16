package asynclib

import (
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
			<-ticker.C
			fn()
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
	Input   chan<- T
	Results <-chan WorketResult[T, R]
}

func StartWorkers[T any, R any](workersNum int, fn func(T) (R, error)) (res Workers[T, R]) {
	inputs := make(chan T, workersNum)
	results := make(chan WorketResult[T, R], workersNum)
	for n := 1; n <= workersNum; n++ {
		go func(n int) {
			for val := range inputs {
				res, err := fn(val)
				results <- WorketResult[T, R]{
					WorkerNum: n,
					Input:     val,
					Result:    res,
					Err:       err,
				}
			}
		}(n)
	}
	return Workers[T, R]{Input: inputs, Results: results}
}

func (j Workers[T, R]) Stop() {
	close(j.Input)
}
