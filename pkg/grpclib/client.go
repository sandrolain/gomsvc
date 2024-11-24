package grpclib

import (
	"fmt"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sandrolain/gomsvc/pkg/certlib"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ClientOptions struct {
	Url         string `validate:"required,url"`
	Logger      *slog.Logger
	Credentials *certlib.ClientTLSConfigFiles
	ServerName  string // Added for TLS verification
}

func CreateClient[T any](new func(grpc.ClientConnInterface) T, opts ClientOptions) (res T, err error) {
	dialOptions := []grpc.DialOption{}

	if opts.Credentials != nil {
		if opts.ServerName == "" {
			err = fmt.Errorf("ServerName is required when using TLS credentials")
			return
		}

		creds, e := certlib.LoadClientTLSConfig(*opts.Credentials)

		if e != nil {
			err = fmt.Errorf("failed to load credentials: %w", e)
			return
		}

		dialOptions = append(dialOptions, grpc.WithTransportCredentials(
			credentials.NewTLS(creds),
		))
	}

	logger := opts.Logger
	if logger == nil {
		logger = svc.Logger()
	}

	loggerOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	dialOptions = append(dialOptions,
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptorLogger(logger), loggerOpts...),
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(interceptorLogger(logger), loggerOpts...),
		),
	)

	conn, err := grpc.NewClient(opts.Url, dialOptions...)
	if err != nil {
		err = fmt.Errorf("fail to dial: %w", err)
		return
	}

	res = new(conn)

	return
}
