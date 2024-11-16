package grpclib

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	g "github.com/sandrolain/gomsvc/pkg/grpclib/test"
	"github.com/sandrolain/gomsvc/pkg/netlib"
	"google.golang.org/grpc"
)

type testServer struct {
	g.UnimplementedUnitTestServiceServer
}

func (s *testServer) RunTest(ctx context.Context, in *g.UnitTestRequest) (*g.UnitTestResponse, error) {
	return &g.UnitTestResponse{
		Success: true,
	}, nil
}

func TestNewGrpcServer(t *testing.T) {
	opts := ServerOptions{
		Port:    8080,
		Desc:    &g.UnitTestService_ServiceDesc,
		Handler: &testServer{},
		Logger:  slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})),
	}

	_, err := NewGrpcServer(opts)
	if err != nil {
		t.Fatalf("NewGrpcServer returned error: %v", err)
	}
}

func TestNewGrpcServer_FailToLoadCredentials(t *testing.T) {
	opts := ServerOptions{
		Port:    8080,
		Desc:    &g.UnitTestService_ServiceDesc,
		Handler: &testServer{},
		Credentials: &Credentials{
			CertPath: "non-existent-cert",
			KeyPath:  "non-existent-key",
			CAPath:   "non-existent-ca",
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})),
	}

	_, err := NewGrpcServer(opts)
	if err == nil {
		t.Fatalf("NewGrpcServer did not return error when loading credentials failed")
	}
}

func TestGrpcServer_Start(t *testing.T) {
	port, err := netlib.GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort returned error: %v", err.Error())
	}

	opts := ServerOptions{
		Port:    port,
		Desc:    &g.UnitTestService_ServiceDesc,
		Handler: &testServer{},
		Logger:  slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})),
	}

	srv, err := NewGrpcServer(opts)
	if err != nil {
		t.Fatalf("NewGrpcServer returned error: %v", err)
	}

	go func() {
		srv.Start()
	}()

	time.Sleep(500 * time.Millisecond)

	conn, err := grpc.Dial(fmt.Sprintf(":%v", opts.Port), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("grpc.Dial returned error: %v", err)
	}
	defer conn.Close()

	client := g.NewUnitTestServiceClient(conn)

	resp, err := client.RunTest(context.Background(), &g.UnitTestRequest{})
	if err != nil {
		t.Fatalf("RunTest returned error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("RunTest returned failure")
	}

	srv.Stop()
}
