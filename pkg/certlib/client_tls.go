package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

type ClientTLSConfigArgs[T any] struct {
	Cert       T
	Key        T
	CA         T
	ServerName string
}

func CreateClientTLSCredentials(args ClientTLSConfigArgs[[]byte]) (cred credentials.TransportCredentials, err error) {
	if args.ServerName == "" {
		return nil, errors.New("ServerName is required")
	}

	clientCert, err := tls.X509KeyPair(args.Cert, args.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(args.CA) {
		return nil, errors.New("failed to add client CA's certificate")
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certpool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   args.ServerName,
	})
	return
}

func LoadClientTLSCredentials(args ClientTLSConfigArgs[string]) (cred credentials.TransportCredentials, err error) {
	if args.ServerName == "" {
		err = errors.New("ServerName is required")
		return
	}

	caCertBytes, err := os.ReadFile(args.CA)
	if err != nil {
		return
	}
	clientCert, err := tls.LoadX509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(caCertBytes) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certpool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   args.ServerName,
	})
	return
}
