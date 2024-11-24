package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"google.golang.org/grpc/credentials"
)

// ServerTLSConfigArgs holds the configuration parameters for creating server TLS credentials.
// Type parameter T can be either []byte for raw certificate data or string for file paths.
type ServerTLSConfigArgs[T any] struct {
	// Cert is the server's certificate (PEM encoded)
	Cert T
	// Key is the server's private key (PEM encoded)
	Key T
	// CA is the client CA certificate for client authentication (PEM encoded)
	CA T
}

// CreateServerTLSConfig creates gRPC transport credentials from raw certificate data.
// The certificates and key should be PEM encoded. Client authentication is required.
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

// LoadServerTLSCredentials creates gRPC transport credentials by loading certificates from files.
// The files should contain PEM encoded certificates and key. Client authentication is required.
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
