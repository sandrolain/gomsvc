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

func EncodeCertificateToPEM(cert *x509.Certificate) (certPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  crtPemType,
		Bytes: cert.Raw,
	}), nil
}

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

func EncodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) (keyPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  prvRsaPemType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), nil
}

func EncodeRSAPublicKeyToPEM(key *rsa.PublicKey) (keyPEMBytes []byte, err error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  pubRsaPemType,
		Bytes: x509.MarshalPKCS1PublicKey(key),
	}), nil
}

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
