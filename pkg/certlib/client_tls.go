package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// ClientTLSConfigArgs holds the configuration parameters for creating client TLS credentials.
// Type parameter T can be either []byte for raw certificate data or string for file paths.
type ClientTLSConfigArgs[T any] struct {
	// Cert is the client's certificate (PEM encoded)
	Cert T
	// Key is the client's private key (PEM encoded)
	Key T
	// CA is the certificate authority's certificate (PEM encoded)
	CA T
	// ServerName is the expected server name for verification
	ServerName string
}

// CreateClientTLSCredentials creates gRPC transport credentials from raw certificate data.
// The certificates and key should be PEM encoded.
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

// LoadClientTLSCredentials creates gRPC transport credentials by loading certificates from files.
// The files should contain PEM encoded certificates and key.
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
