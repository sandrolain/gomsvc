// Package certlib provides functionality for handling TLS certificates and credentials
// for both client and server applications. It supports loading certificates from files
// or raw bytes, and configuring TLS settings with proper security defaults.
package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

// ServerTLSConfigBytes holds the configuration parameters for creating server TLS credentials from raw certificate data.
// All fields are required and must contain valid PEM encoded certificates.
type ServerTLSConfigBytes struct {
	// Cert is the server's certificate (PEM encoded)
	Cert []byte `validate:"required"`
	// Key is the server's private key (PEM encoded)
	Key []byte `validate:"required"`
	// CA is the client CA certificate for client authentication (PEM encoded)
	CA []byte `validate:"required"`
}

// ServerTLSConfigFiles holds the configuration parameters for creating server TLS credentials from files.
// All paths must be valid and accessible. The files must contain valid PEM encoded certificates.
type ServerTLSConfigFiles struct {
	// CertFile is the server's certificate (PEM encoded)
	CertFile string `validate:"required,filepath"`
	// KeyFile is the server's private key (PEM encoded)
	KeyFile string `validate:"required,filepath"`
	// CAFile is the client CA certificate for client authentication (PEM encoded)
	CAFile string `validate:"required,filepath"`
}

// createServerTLSConfig creates a server TLS config from raw certificate data.
// It configures the TLS settings with proper security defaults including TLS 1.2 minimum version
// and requires client certificate verification. The function validates and loads the provided
// certificates and sets up the certificate pool for client authentication.
func createServerTLSConfig(cert []byte, key []byte, ca []byte) (*tls.Config, error) {
	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, errors.New("failed to add client CA's certificate")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// CreateServerTLSConfig creates TLS credentials from raw certificate data.
// The certificates and key must be PEM encoded. This function enforces client authentication,
// requiring clients to present valid certificates that can be verified against the provided CA.
// Returns an error if the certificates are invalid or cannot be loaded.
func CreateServerTLSConfig(args ServerTLSConfigBytes) (res *tls.Config, err error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	return createServerTLSConfig(args.Cert, args.Key, args.CA)
}

// LoadServerTLSConfig creates TLS credentials by loading certificates from files.
// The files must contain PEM encoded certificates and keys. This function enforces client
// authentication, requiring clients to present valid certificates that can be verified
// against the provided CA certificate.
// Returns an error if any of the files cannot be read or contain invalid certificates.
func LoadServerTLSConfig(args ServerTLSConfigFiles) (res *tls.Config, err error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	pemServerCert, err := os.ReadFile(args.CertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}
	pemServerKey, err := os.ReadFile(args.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server private key: %w", err)
	}
	pemClientCA, err := os.ReadFile(args.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client CA certificate: %w", err)
	}
	return createServerTLSConfig(pemServerCert, pemServerKey, pemClientCA)
}
