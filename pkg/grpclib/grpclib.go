package grpclib

import (
	"fmt"
	"net"

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
}

func CreateServer(opts ServerOptions) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", opts.Port))
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	s.RegisterService(opts.Desc, opts.Handler)
	reflection.Register(s)
	svc.Logger().Info("start gRPC server", "addr", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
