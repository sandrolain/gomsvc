package grpclib

import (
	"fmt"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sandrolain/gomsvc/pkg/certlib"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
)

type ClientOptions struct {
	Url         string
	Logger      *slog.Logger
	Credentials Credentials
}

func CreateClient[T any](new func(grpc.ClientConnInterface) T, opts ClientOptions) (res T, err error) {
	creds, err := certlib.LoadClientTLSCredentials(certlib.ClientTLSConfigArgs[string]{
		Cert: opts.Credentials.CertPath,
		Key:  opts.Credentials.KeyPath,
		CA:   opts.Credentials.CAPath,
	})

	logger := opts.Logger
	if logger == nil {
		logger = svc.Logger()
	}
	loggerOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptorLogger(logger), loggerOpts...),
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(interceptorLogger(logger), loggerOpts...),
		),
	}

	conn, err := grpc.Dial(opts.Url, dialOpts...)
	if err != nil {
		err = fmt.Errorf("fail to dial: %w", err)
		return
	}

	res = new(conn)

	return
}
