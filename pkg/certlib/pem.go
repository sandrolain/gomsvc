package certlib

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	crtPemType    = "CERTIFICATE"
	pubPemType    = "PUBLIC KEY"
	prvPemType    = "PRIVATE KEY"
	prvRsaPemType = "RSA PRIVATE KEY"
	pubRsaPemType = "RSA PUBLIC KEY"
)

// EncodeCertificateToPEM encodes an X.509 certificate to PEM format
// Returns the PEM-encoded certificate as a byte slice
func EncodeCertificateToPEM(cert *x509.Certificate) (certPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  crtPemType,
		Bytes: cert.Raw,
	}), nil
}

// EncodePrivateKeyToPEM encodes an RSA private key to PEM format using PKCS8
// Returns the PEM-encoded private key as a byte slice
func EncodePrivateKeyToPEM(key *rsa.PrivateKey) (keyPEMBytes []byte, err error) {
	data, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		err = fmt.Errorf("unable to marshal private key: %s", err)
		return
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  prvPemType,
		Bytes: data,
	}), nil
}

// EncodePublicKeyToPEM encodes an RSA public key to PEM format using PKIX
// Returns the PEM-encoded public key as a byte slice
func EncodePublicKeyToPEM(key *rsa.PublicKey) (keyPEMBytes []byte, err error) {
	data, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		err = fmt.Errorf("unable to marshal public key: %s", err)
		return
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  pubPemType,
		Bytes: data,
	}), nil
}

// EncodeRSAPrivateKeyToPEM encodes an RSA private key to PEM format using PKCS1
// Returns the PEM-encoded RSA private key as a byte slice
func EncodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) (keyPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  prvRsaPemType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), nil
}

// EncodeRSAPublicKeyToPEM encodes an RSA public key to PEM format using PKCS1
// Returns the PEM-encoded RSA public key as a byte slice
func EncodeRSAPublicKeyToPEM(key *rsa.PublicKey) (keyPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  pubRsaPemType,
		Bytes: x509.MarshalPKCS1PublicKey(key),
	}), nil
}

// ParseCertificateFromPEM decodes a PEM-encoded X.509 certificate
// Returns the parsed certificate or an error if the PEM block is invalid or parsing fails
func ParseCertificateFromPEM(certPEMBytes []byte) (cert *x509.Certificate, err error) {
	block, _ := pem.Decode(certPEMBytes)
	if block == nil {
		err = errors.New("failed to parse PEM block containing the certificate")
		return
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse certificate: %s", err)
		return
	}
	return
}

// ParsePrivateKeyFromPEM decodes a PEM-encoded private key in either PKCS1 or PKCS8 format
// Returns the parsed RSA private key or an error if the PEM block is invalid or parsing fails
func ParsePrivateKeyFromPEM(keyPEMBytes []byte) (key *rsa.PrivateKey, err error) {
	block, _ := pem.Decode(keyPEMBytes)
	if block == nil {
		err = errors.New("failed to parse PEM block containing the key")
		return
	}
	switch block.Type {
	case prvRsaPemType:
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			err = fmt.Errorf("failed to parse key: %s", err)
			return nil, err
		}
		return
	case prvPemType:
		var pKey any
		pKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = pKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("invalid private key")
		}
		return
	}
	err = errors.New("invalid key type")
	return
}

// validatePath checks if the given file path is valid and points to a regular file
// Returns an error if the path is invalid, points to a directory, or the file doesn't exist
func validatePath(path string) error {
	// Clean the path to remove any . or .. components
	cleanPath := filepath.Clean(path)

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return err
	}

	// Check if file exists and is regular
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("path is a directory")
	}

	return nil
}

// ParseCertificateFromFile reads and parses an X.509 certificate from a PEM file
// Returns the parsed certificate or an error if the file is invalid or parsing fails
func ParseCertificateFromFile(path string) (cert *x509.Certificate, err error) {
	if err = validatePath(path); err != nil {
		return nil, err
	}

	// #nosec G304 -- path has been validated by validatePath
	certPEMBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	return ParseCertificateFromPEM(certPEMBytes)
}

// ParsePrivateKeyFromFile reads and parses an RSA private key from a PEM file
// Returns the parsed private key or an error if the file is invalid or parsing fails
func ParsePrivateKeyFromFile(path string) (key *rsa.PrivateKey, err error) {
	if err = validatePath(path); err != nil {
		return nil, err
	}

	// #nosec G304 -- path has been validated by validatePath
	keyPEMBytes, err := os.ReadFile(path)
	if err != nil {
		return
	}
	return ParsePrivateKeyFromPEM(keyPEMBytes)
}
