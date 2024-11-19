package svc

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

func PanicWithError[T any](v T, e error) T {
	if e != nil {
		if logger != nil {
			panicLog(e)
		}
		panic(e)
	}
	return v
}

func PanicIfError(e error) {
	if e != nil {
		if logger != nil {
			panicLog(e)
		}
		panic(e)
	}
}

func panicLog(e error) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), slog.LevelError, "panic", pcs[0])
	r.Add("err", e)
	_ = logger.Handler().Handle(context.Background(), r)
	Exit(1)
}
