package grpclib

import (
	"context"
	"runtime"
	"time"

	"log/slog"

	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func log(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, Infof]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = svc.Logger().Handler().Handle(ctx, r)
}

func InvalidArgument(msg ...string) error {
	m := "Invalid Argument"
	if len(msg) > 0 && msg[0] != "" {
		m = msg[0]
	}
	e := status.Error(codes.InvalidArgument, m)
	log(context.Background(), slog.LevelWarn, m, "err", e)
	return e
}

func InternalError(msg ...string) error {
	m := "Internal Error"
	if len(msg) > 0 && msg[0] != "" {
		m = msg[0]
	}
	e := status.Error(codes.Internal, m)
	log(context.Background(), slog.LevelError, m, "err", e)
	return e
}
