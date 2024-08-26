package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

type ClientTLSConfigArgs[T any] struct {
	Cert T
	Key  T
	CA   T
}

func CreateClientTLSCredentials(args ClientTLSConfigArgs[[]byte]) (cred credentials.TransportCredentials, err error) {
	clientCert, err := tls.X509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(args.CA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	})
	return
}

func LoadClientTLSCredentials(args ClientTLSConfigArgs[string]) (cred credentials.TransportCredentials, err error) {
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
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	})
	return
}
