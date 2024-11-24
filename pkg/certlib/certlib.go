// Package certlib provides functionality for generating and managing X.509 certificates
// and their corresponding private keys. It supports creating different types of certificates
// including root CAs, intermediate CAs, server certificates, and client certificates.
package certlib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"
)

// MinKeySize defines the minimum allowed RSA key size in bits
const MinKeySize = 2048

// DefaultKeySize defines the default RSA key size in bits used for certificate generation
const DefaultKeySize = 2048

// CertKey holds a certificate and its corresponding private key
type CertKey struct {
	Cert *x509.Certificate
	Key  *rsa.PrivateKey
}

// TLSCertificate converts the certificate and key pair into a tls.Certificate
func (c *CertKey) TLSCertificate() *tls.Certificate {
	return &tls.Certificate{
		Certificate: [][]byte{c.Cert.Raw},
		PrivateKey:  c.Key,
	}
}

// PublicKey returns the public key portion of the private key
func (c *CertKey) PublicKey() *rsa.PublicKey {
	return &c.Key.PublicKey
}

// CertificateType represents the type of certificate to be generated
type CertificateType int

const (
	// CertificateTypeRootCA represents a root Certificate Authority
	CertificateTypeRootCA CertificateType = iota
	// CertificateTypeIntermediateCA represents an intermediate Certificate Authority
	CertificateTypeIntermediateCA
	// CertificateTypeServer represents a server certificate
	CertificateTypeServer
	// CertificateTypeClient represents a client certificate
	CertificateTypeClient
)

// CertificateArgs contains the parameters needed to generate a certificate
type CertificateArgs struct {
	// Serial is the certificate's serial number
	Serial *big.Int
	// Subject contains the certificate subject information
	Subject pkix.Name
	// Extensions contains additional X.509 extensions
	Extensions []pkix.Extension
	// Issuer is the certificate and key that will sign this certificate
	Issuer CertKey
	// NotBefore is the time when the certificate becomes valid
	NotBefore time.Time
	// Duration specifies how long the certificate will be valid
	Duration time.Duration
	// EmailAddresses contains email addresses to include in the certificate
	EmailAddresses []string
	// DNSNames contains DNS names to include in the certificate
	DNSNames []string
	// IPAddresses contains IP addresses to include in the certificate
	IPAddresses []net.IP
	// KeySize specifies the size of the RSA key to generate
	KeySize int
}

// validateSubject checks if the subject information is valid for the given certificate type
func validateSubject(subject pkix.Name, certType CertificateType) error {
	// Validate subject fields based on certificate type
	if subject.CommonName == "" {
		return errors.New("CommonName is required")
	}

	// For CA certificates, require more strict validation
	if certType == CertificateTypeRootCA || certType == CertificateTypeIntermediateCA {
		if len(subject.Organization) == 0 {
			return errors.New("Organization is required for CA certificates")
		}
		if len(subject.Country) == 0 {
			return errors.New("Country is required for CA certificates")
		}
	}

	// Validate country code length if provided
	for _, country := range subject.Country {
		if len(country) != 2 {
			return errors.New("Country code must be exactly 2 characters (ISO 3166-1 alpha-2)")
		}
	}

	return nil
}

// validateServerIdentity checks if the server identity information is valid
func validateServerIdentity(args CertificateArgs) error {
	// Validate server identity fields
	if len(args.DNSNames) == 0 && len(args.IPAddresses) == 0 {
		return errors.New("at least one DNS name or IP address is required for server certificates")
	}

	// Validate DNS names
	for _, dns := range args.DNSNames {
		if dns == "" {
			return errors.New("empty DNS name is not allowed")
		}
		if len(dns) > 255 {
			return errors.New("DNS name exceeds maximum length of 255 characters")
		}
	}

	// Validate IP addresses
	for _, ip := range args.IPAddresses {
		if ip == nil {
			return errors.New("nil IP address is not allowed")
		}
		if ip.IsUnspecified() {
			return errors.New("unspecified IP address is not allowed")
		}
	}

	return nil
}

// GenerateCertificate generates a certificate and its corresponding private key based on the given parameters
func GenerateCertificate(certType CertificateType, args CertificateArgs) (res CertKey, err error) {
	// Validate subject fields
	if err = validateSubject(args.Subject, certType); err != nil {
		err = fmt.Errorf("invalid subject: %w", err)
		return
	}

	// For server certificates, validate DNS names and IP addresses
	if certType == CertificateTypeServer {
		if err = validateServerIdentity(args); err != nil {
			err = fmt.Errorf("invalid server identity: %w", err)
			return
		}
	}

	serialNumber := args.Serial
	if serialNumber == nil {
		serialNumber = big.NewInt(time.Now().UnixMilli())
	}

	notBefore := args.NotBefore
	if notBefore.IsZero() {
		notBefore = time.Now()
	}

	if args.Duration == 0 {
		err = errors.New("duration is required")
		return
	}

	notAfter := notBefore.Add(args.Duration)

	var keyUsage x509.KeyUsage
	var isCA bool
	var extKeyUsage []x509.ExtKeyUsage
	switch certType {
	case CertificateTypeRootCA, CertificateTypeIntermediateCA:
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature
		isCA = true
	case CertificateTypeServer:
		keyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		isCA = false
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case CertificateTypeClient:
		keyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		isCA = false
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               args.Subject,
		EmailAddresses:        args.EmailAddresses,
		DNSNames:              args.DNSNames,
		IPAddresses:           args.IPAddresses,
		ExtraExtensions:       args.Extensions,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: isCA,
		IsCA:                  isCA,
	}

	key, err := generateKey(args.KeySize)
	if err != nil {
		err = fmt.Errorf("unable to generate key: %s", err)
		return
	}

	issuerCert := args.Issuer.Cert
	issuerKey := args.Issuer.Key

	if certType != CertificateTypeRootCA {
		if issuerCert == nil {
			err = errors.New("issuer certificate is required")
			return
		}

		if issuerKey == nil {
			err = errors.New("issuer key is required")
			return
		}
	} else {
		issuerCert = cert
		issuerKey = key
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, issuerCert, &key.PublicKey, issuerKey)
	if err != nil {
		err = fmt.Errorf("unable to create certificate: %s", err)
		return
	}

	cert, err = x509.ParseCertificate(certBytes)
	if err != nil {
		err = fmt.Errorf("unable to parse certificate: %s", err)
		return
	}

	res.Cert = cert
	res.Key = key

	return
}

// generateKey generates a new RSA key pair with the given key size
func generateKey(keySize int) (*rsa.PrivateKey, error) {
	// Generate a new RSA key pair with the given key size
	if keySize == 0 {
		keySize = DefaultKeySize
	}

	if keySize < MinKeySize {
		return nil, fmt.Errorf("key size must be at least %d bits", MinKeySize)
	}

	return rsa.GenerateKey(rand.Reader, keySize)
}
