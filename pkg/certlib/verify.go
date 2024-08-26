package certlib

import (
	"crypto/x509"
	"errors"
	"fmt"
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
	}

	switch args.Type {
	case CertificateTypeServer:
		options.KeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case CertificateTypeClient:
		options.KeyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
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
