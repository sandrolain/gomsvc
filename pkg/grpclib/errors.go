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

func getArgs(c codes.Code, def string, msg string, args []interface{}) (string, []interface{}, error) {
	if msg == "" {
		msg = def
	}
	e := status.Error(c, msg)
	args = append([]interface{}{"err", e}, args...)
	return msg, args, e
}

func InvalidArgument(msg string, args ...interface{}) error {
	m, args, e := getArgs(codes.InvalidArgument, "Invalid Argument", msg, args)
	log(context.Background(), slog.LevelWarn, m, args...)
	return e
}

func NotFound(msg string, args ...interface{}) error {
	m, args, e := getArgs(codes.NotFound, "Not Found", msg, args)
	log(context.Background(), slog.LevelWarn, m, args...)
	return e
}

func InternalError(msg string, args ...interface{}) error {
	m, args, e := getArgs(codes.Internal, "Internal Error", msg, args)
	log(context.Background(), slog.LevelError, m, args...)
	return e
}
