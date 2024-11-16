package grpclib

import (
	"context"
	"fmt"
	"net"

	"log/slog"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
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
	Credentials *Credentials
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

type GrpcServer struct {
	server *grpc.Server
	lis    net.Listener
	logger *slog.Logger
}

func NewGrpcServer(opts ServerOptions) (*GrpcServer, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", opts.Port))
	if err != nil {
		return nil, err
	}
	logger := opts.Logger
	if logger == nil {
		logger = svc.Logger()
	}

	err = validator.New().Struct(opts)
	if err != nil {
		return nil, err
	}

	serverOptions := []grpc.ServerOption{}
	if opts.Credentials != nil {
		cred, err := certlib.LoadServerTLSCredentials(certlib.ServerTLSConfigArgs[string]{
			Cert: opts.Credentials.CertPath,
			Key:  opts.Credentials.KeyPath,
			CA:   opts.Credentials.CAPath,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
		serverOptions = append(serverOptions, grpc.Creds(cred))
	}

	protovalidator, e := protovalidate.New()
	if e != nil {
		return nil, fmt.Errorf("failed to create validator: %w", e)
	}

	loggerOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	serverOptions = append(serverOptions,
		grpc.ChainUnaryInterceptor(
			protovalidate_middleware.UnaryServerInterceptor(protovalidator),
			logging.UnaryServerInterceptor(interceptorLogger(logger), loggerOpts...),
		),
		grpc.ChainStreamInterceptor(
			protovalidate_middleware.StreamServerInterceptor(protovalidator),
			logging.StreamServerInterceptor(interceptorLogger(logger), loggerOpts...),
		),
	)

	s := grpc.NewServer(serverOptions...)
	s.RegisterService(opts.Desc, opts.Handler)
	reflection.Register(s)

	return &GrpcServer{server: s, lis: lis, logger: logger}, nil
}

func (gs *GrpcServer) Start() error {
	gs.logger.Info("start gRPC server", "addr", gs.lis.Addr())
	if err := gs.server.Serve(gs.lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (gs *GrpcServer) Stop() {
	gs.server.GracefulStop()
	gs.logger.Info("gRPC server stopped")
}
