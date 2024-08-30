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

const MinKeySize = 2048
const DefaultKeySize = 2048

type CertKey struct {
	Cert *x509.Certificate
	Key  *rsa.PrivateKey
}

func (c *CertKey) TLSCertificate() *tls.Certificate {
	return &tls.Certificate{
		Certificate: [][]byte{c.Cert.Raw},
		PrivateKey:  c.Key,
	}
}

type CertificateType int

const (
	CertificateTypeRootCA CertificateType = iota
	CertificateTypeIntermediateCA
	CertificateTypeServer
	CertificateTypeClient
)

type CertificateArgs struct {
	Serial         *big.Int
	Subject        pkix.Name
	Extensions     []pkix.Extension
	Issuer         CertKey
	NotBefore      time.Time
	Duration       time.Duration
	EmailAddresses []string
	DNSNames       []string
	IPAddresses    []net.IP
	KeySize        int
}

func GenerateCertificate(certType CertificateType, args CertificateArgs) (res CertKey, err error) {
	serialNumber := args.Serial
	if serialNumber == nil {
		serialNumber = big.NewInt(time.Now().UnixMilli())
	}

	notBefore := time.Time{}
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
		Subject:               args.Subject, // TODO: validation
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

func generateKey(keySize int) (*rsa.PrivateKey, error) {
	if keySize == 0 {
		keySize = DefaultKeySize
	}

	if keySize < MinKeySize {
		return nil, fmt.Errorf("key size must be at least %d", MinKeySize)
	}
	return rsa.GenerateKey(rand.Reader, keySize)
}
