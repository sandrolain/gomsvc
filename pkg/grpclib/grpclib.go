package grpclib

import (
	"context"
	"fmt"
	"net"

	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sandrolain/gomsvc/pkg/certlib"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type EnvServerConfig struct {
	Port int    `env:"GRPC_PORT" validate:"required,numeric"`
	Cert string `env:"GRPC_CERT" validate:"required,filepath"`
	Key  string `env:"GRPC_KEY" validate:"required,filepath"`
	CA   string `env:"GRPC_CA" validate:"required,filepath"`
}

type EnvClientConfig struct {
	Cert string `env:"GRPC_CERT" validate:"required,file"`
	Key  string `env:"GRPC_KEY" validate:"required,file"`
	CA   string `env:"GRPC_CA" validate:"required,file"`
}

type Credentials struct {
	CertPath string `validate:"required,file"`
	KeyPath  string `validate:"required,file"`
	CAPath   string `validate:"required,file"`
}

type ServerOptions struct {
	Port        int               `validate:"required,number"`
	Desc        *grpc.ServiceDesc `validate:"required"`
	Handler     interface{}       `validate:"required"`
	Logger      *slog.Logger
	Credentials Credentials `validate:"required"`
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
	err = validator.New().Struct(opts)
	if err != nil {
		return err
	}
	cred, err := certlib.LoadServerTLSCredentials(certlib.ServerTLSConfigArgs[string]{
		Cert: opts.Credentials.CertPath,
		Key:  opts.Credentials.KeyPath,
		CA:   opts.Credentials.CAPath,
	})
	if err != nil {
		return err
	}
	s := grpc.NewServer(
		grpc.Creds(cred),
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
