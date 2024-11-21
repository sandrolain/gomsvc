package certlib

import (
	"crypto/x509"
	"errors"
	"fmt"
	"time"
)

type VerifyCertificateArgs struct {
	Type          CertificateType
	Cert          *x509.Certificate
	DNSName       string
	Intermediates []*x509.Certificate
	Roots         []*x509.Certificate
}

func VerifyCertificate(args VerifyCertificateArgs) (err error) {
	cert := args.Cert
	if cert == nil {
		return errors.New("no certificate provided")
	}

	// Check certificate expiration
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	roots := x509.NewCertPool()
	for _, root := range args.Roots {
		roots.AddCert(root)
	}

	intermediates := x509.NewCertPool()
	for _, intermediate := range args.Intermediates {
		intermediates.AddCert(intermediate)
	}

	options := x509.VerifyOptions{
		DNSName:       args.DNSName,
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
	}

	switch args.Type {
	case CertificateTypeServer:
		options.KeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case CertificateTypeClient:
		options.KeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	// For client/server certs, require proper chain with intermediates
	if args.Type == CertificateTypeClient || args.Type == CertificateTypeServer {
		if len(args.Intermediates) == 0 {
			return errors.New("client/server certificates must be signed by an intermediate CA")
		}
	}

	chains, err := cert.Verify(options)
	if err != nil {
		err = fmt.Errorf("unable to verify certificate: %s", err)
		return
	}

	if len(chains) == 0 {
		err = errors.New("no certificate chain found")
		return
	}

	return
}
