package grpclib

import (
	"fmt"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sandrolain/gomsvc/pkg/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientOptions struct {
	Url    string
	CAFile string
	Host   string
	Logger *slog.Logger
}

func CreateClient[T any](new func(grpc.ClientConnInterface) T, opts ClientOptions) (res T, err error) {
	// creds, err := credentials.NewClientTLSFromFile(opts.CAFile, opts.Host)
	// if err != nil {
	// 	err = fmt.Errorf("failed to create TLS credentials: %w", err)
	// 	return
	// }

	creds := insecure.NewCredentials()

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
