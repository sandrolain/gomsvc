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

// ClientTLSConfigBytes holds the configuration parameters for creating client TLS credentials from raw certificate data.
// All fields are required and validated using the validator package.
type ClientTLSConfigBytes struct {
	// Cert is the client's certificate (PEM encoded)
	Cert []byte `validate:"required"`

	// Key is the client's private key (PEM encoded)
	Key []byte `validate:"required"`

	// CA is the certificate authority's certificate (PEM encoded)
	CA []byte `validate:"required"`

	// ServerName is the expected server name for verification
	ServerName string `validate:"required"`
}

// ClientTLSConfigFiles holds the configuration parameters for creating client TLS credentials from files.
// All paths must be valid and accessible, and are validated using the validator package.
type ClientTLSConfigFiles struct {
	// CertFile is the path to the client's certificate (PEM encoded)
	CertFile string `validate:"required,filepath"`

	// KeyFile is the path to the client's private key (PEM encoded)
	KeyFile string `validate:"required,filepath"`

	// CAFile is the path to the certificate authority's certificate (PEM encoded)
	CAFile string `validate:"required,filepath"`

	// ServerName is the expected server name for verification
	ServerName string `validate:"required"`
}

// newClientTLSConfig creates a new TLS configuration for client connections.
// It configures the TLS settings with proper security defaults including TLS 1.2 minimum version.
// The function validates and loads the provided certificates and sets up the certificate pool.
func newClientTLSConfig(serverName string, cert, key, ca []byte) (*tls.Config, error) {
	var (
		clientCert tls.Certificate
		err        error
	)

	clientCert, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(ca) {
		return nil, errors.New("failed to add client CA's certificate")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certpool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   serverName,
	}, nil
}

// validate performs struct validation using the validator package.
// It checks if all required fields are present and meet the validation rules defined in struct tags.
func validate(args interface{}) error {
	validate := validator.New()
	if err := validate.Struct(args); err != nil {
		return fmt.Errorf("failed to validate input: %w", err)
	}
	return nil
}

// LoadClientTLSConfig creates TLS credentials by loading certificates from files.
// It validates the input configuration, reads the certificate files, and creates a TLS configuration
// suitable for client connections. The files must contain PEM encoded certificates and keys.
// Returns an error if any validation fails or if the files cannot be read.
func LoadClientTLSConfig(args ClientTLSConfigFiles) (*tls.Config, error) {
	if err := validate(args); err != nil {
		return nil, err
	}
	cert, err := os.ReadFile(args.CertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}
	key, err := os.ReadFile(args.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key: %w", err)
	}
	ca, err := os.ReadFile(args.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client CA certificate: %w", err)
	}
	return newClientTLSConfig(args.ServerName, cert, key, ca)
}

// CreateClientTLSConfig creates TLS credentials from raw certificate data.
// It validates the input configuration and creates a TLS configuration suitable for client connections.
// The certificates and key must be PEM encoded.
// Returns an error if any validation fails or if the certificates are invalid.
func CreateClientTLSConfig(args ClientTLSConfigBytes) (*tls.Config, error) {
	if err := validate(args); err != nil {
		return nil, err
	}
	return newClientTLSConfig(args.ServerName, args.Cert, args.Key, args.CA)
}
