package certlib

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"
)

func hashPublicKey(key *rsa.PublicKey) ([]byte, error) {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to hash key: %s", err)
	}

	h := sha1.New()
	h.Write(b)
	return h.Sum(nil), nil
}

// GenerateBasicCA creates a root CA with basic settings
func GenerateBasicCA(commonName string, organization string, country string, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
			Country:     []string{country},
		},
		Duration: duration,
	})
}

// GenerateBasicIntermediateCA creates an intermediate CA with basic settings
func GenerateBasicIntermediateCA(commonName string, organization string, country string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeIntermediateCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
			Country:     []string{country},
		},
		Issuer:   issuer,
		Duration: duration,
	})
}

// GenerateBasicServerCert creates a server certificate with basic settings
func GenerateBasicServerCert(commonName string, dnsNames []string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeServer, CertificateArgs{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames: dnsNames,
		Issuer:   issuer,
		Duration: duration,
	})
}

// GenerateBasicClientCert creates a client certificate with basic settings
func GenerateBasicClientCert(commonName string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeClient, CertificateArgs{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		Issuer:   issuer,
		Duration: duration,
	})
}

// CreateCertPool creates a new certificate pool from the given certificates
func CreateCertPool(certs ...*x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool
}

// CreateTLSConfig creates a basic TLS config for server or client
func CreateTLSConfig(cert CertKey, roots *x509.CertPool) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{*cert.TLSCertificate()},
		RootCAs:     roots,
	}
}
