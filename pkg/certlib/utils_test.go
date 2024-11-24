package certlib

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"time"
)

// generateBasicCA creates a root CA with basic settings.
// The certificate will be self-signed and valid for the specified duration.
func generateBasicCA(commonName string, organization string, country string, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
			Country:      []string{country},
		},
		Duration: duration,
	})
}

// generateBasicIntermediateCA creates an intermediate CA with basic settings.
// The certificate will be signed by the provided issuer and valid for the specified duration.
func generateBasicIntermediateCA(commonName string, organization string, country string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeIntermediateCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
			Country:      []string{country},
		},
		Issuer:   issuer,
		Duration: duration,
	})
}

// generateBasicServerCert creates a server certificate with basic settings.
// The certificate will be signed by the provided issuer and valid for the specified duration.
// The certificate will include the provided DNS names in the Subject Alternative Names.
func generateBasicServerCert(commonName string, dnsNames []string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeServer, CertificateArgs{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames: dnsNames,
		Issuer:   issuer,
		Duration: duration,
	})
}

// generateBasicClientCert creates a client certificate with basic settings.
// The certificate will be signed by the provided issuer and valid for the specified duration.
func generateBasicClientCert(commonName string, issuer CertKey, duration time.Duration) (CertKey, error) {
	return GenerateCertificate(CertificateTypeClient, CertificateArgs{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		Issuer:   issuer,
		Duration: duration,
	})
}

// createCertPool creates a new certificate pool from the given certificates.
// This is useful for creating a pool of trusted certificates for TLS configuration.
func createCertPool(certs ...*x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool
}

// createTLSConfig creates a basic TLS config for server or client.
// The configuration uses TLS 1.2 or higher and requires certificates.
func createTLSConfig(cert CertKey, roots *x509.CertPool) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{*cert.TLSCertificate()},
		RootCAs:      roots,
		MinVersion:   tls.VersionTLS12,
	}
}
