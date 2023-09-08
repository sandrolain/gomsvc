package grpclib

import (
	"context"
	"fmt"
	"net"

	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type EnvGrpcConfig struct {
	GrpcPort int `env:"GRPC_PORT" validate:"required"`
}

type ServerOptions struct {
	Port    int
	Desc    *grpc.ServiceDesc
	Handler interface{}
	Logger  *slog.Logger
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func CreateServer(opts ServerOptions) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", opts.Port))
	if err != nil {
		return err
	}
	logger := opts.Logger
	if logger == nil {
		logger = svc.Logger()
	}
	loggerOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(logger), loggerOpts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(logger), loggerOpts...),
		),
	)
	s.RegisterService(opts.Desc, opts.Handler)
	reflection.Register(s)
	svc.Logger().Info("start gRPC server", "addr", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
