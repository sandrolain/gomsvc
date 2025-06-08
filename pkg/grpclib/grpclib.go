package grpclib

import (
	"context"
	"fmt"
	"net"

	"log/slog"

	"buf.build/go/protovalidate"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/sandrolain/gomsvc/pkg/certlib"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type EnvServerConfig struct {
	Port     int    `env:"GRPC_PORT" validate:"required,numeric"`
	CertFile string `env:"GRPC_CERT" validate:"omitempty,filepath"`
	KeyFile  string `env:"GRPC_KEY" validate:"omitempty,filepath"`
	CAFile   string `env:"GRPC_CA" validate:"omitempty,filepath"`
}

type ServerOptions struct {
	Port        int               `validate:"required,number"`
	ServiceDesc *grpc.ServiceDesc `validate:"required"`
	Handler     interface{}       `validate:"required"`
	Logger      *slog.Logger
	TLSConfig   *certlib.ServerTLSConfigFiles `validate:"omitempty"`
}

func ServerOptionsFromEnvConfig(cfg EnvServerConfig) ServerOptions {
	var creds *certlib.ServerTLSConfigFiles
	if cfg.CertFile != "" {
		creds = &certlib.ServerTLSConfigFiles{
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
			CAFile:   cfg.CAFile,
		}
	}
	return ServerOptions{
		Port:      cfg.Port,
		TLSConfig: creds,
	}
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
	if opts.TLSConfig != nil {
		cred, err := certlib.LoadServerTLSConfig(*opts.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
		serverOptions = append(serverOptions, grpc.Creds(
			credentials.NewTLS(cred),
		))
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
	s.RegisterService(opts.ServiceDesc, opts.Handler)
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
