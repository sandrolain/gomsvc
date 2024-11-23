package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

type ServerTLSConfigArgs[T any] struct {
	Cert T
	Key  T
	CA   T
}

func CreateServerTLSConfig(args ServerTLSConfigArgs[[]byte]) (cred credentials.TransportCredentials, err error) {
	serverCert, err := tls.X509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(args.CA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}
	cred = credentials.NewTLS(cfg)
	return
}

func LoadServerTLSCredentials(args ServerTLSConfigArgs[string]) (cred credentials.TransportCredentials, err error) {
	serverCert, err := tls.LoadX509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	pemClientCA, err := os.ReadFile(args.CA)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	})
	return
}
